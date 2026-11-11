#!/usr/bin/env python3
"""Emit canonical semantic dumps for DXF files using ezdxf.

The output is intentionally limited to stable, semantic facts that both ezdxf
and the Go library expose: DXF version, all graphical entities, user-defined
block contents, table entries, and object handles/types. Raw tag streams are out
of scope; the Go corpus round-trip tests cover tag preservation separately.

Known ezdxf normalizations are applied on both sides:
- R12 (AC1006/AC1009) files have generated handles, so handles are omitted.
- ezdxf creates missing default tables and required objects. Table and object
  inventories therefore come from the source tag stream, not from ezdxf's
  auto-populated collections.
- SEQEND is a structural marker, not a graphical entity.

Usage:
    python ezdxf_dump.py testdata/a.dxf testdata/b.dxf > expected.json
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import ezdxf
from ezdxf.entities import DXFGraphic

PRE_R13 = {"AC1006", "AC1009"}

TABLE_ATTRS = {
    "APPID": "appids",
    "BLOCK_RECORD": "block_records",
    "DIMSTYLE": "dimstyles",
    "LAYER": "layers",
    "LTYPE": "linetypes",
    "STYLE": "styles",
    "UCS": "ucs",
    "VIEW": "views",
    "VPORT": "viewports",
}


def _read_tags(path: str) -> list[tuple[int, str]]:
    """Read an ASCII DXF tag stream without interpreting it."""
    with open(path, "r", encoding="utf-8", errors="replace") as stream:
        lines = stream.read().splitlines()

    if len(lines) % 2:
        raise ValueError(f"{path}: odd number of lines in ASCII DXF tag stream")

    tags: list[tuple[int, str]] = []
    for index in range(0, len(lines), 2):
        code_text = lines[index].strip()
        if not code_text:
            raise ValueError(f"{path}:{index + 1}: empty group code")
        tags.append((int(code_text), lines[index + 1]))
    return tags


def _has_section(tags: list[tuple[int, str]], name: str) -> bool:
    return any(
        tags[index] == (0, "SECTION") and tags[index + 1] == (2, name)
        for index in range(len(tags) - 1)
    )


def _section(tags: list[tuple[int, str]], name: str) -> list[tuple[int, str]]:
    """Return the tags inside SECTION/name, excluding its delimiters."""
    for index in range(len(tags) - 1):
        if tags[index] == (0, "SECTION") and tags[index + 1] == (2, name):
            start = index + 2
            break
    else:
        return []

    for end in range(start, len(tags)):
        if tags[end] == (0, "ENDSEC"):
            return tags[start:end]
    return tags[start:]


def _records(tags: list[tuple[int, str]]) -> list[tuple[str, list[tuple[int, str]]]]:
    """Split a tag stream at group code 0 records."""
    records: list[tuple[str, list[tuple[int, str]]]] = []
    index = 0
    while index < len(tags):
        if tags[index][0] != 0:
            index += 1
            continue

        record_type = tags[index][1].strip()
        index += 1
        body: list[tuple[int, str]] = []
        while index < len(tags) and tags[index][0] != 0:
            body.append(tags[index])
            index += 1
        records.append((record_type, body))
    return records


def _after_marker(
    record: list[tuple[int, str]], marker: str, code: int
) -> str:
    """Return the first value with `code` after the given 100 subclass marker."""
    seen = False
    for tag_code, value in record:
        if tag_code == 100:
            seen = value.strip().lower() == marker.lower()
            continue
        if seen and tag_code == code:
            return value.strip()
    return ""


def _first(record: list[tuple[int, str]], code: int) -> str:
    for tag_code, value in record:
        if tag_code == code:
            return value.strip()
    return ""


def _source_tables(tags: list[tuple[int, str]], version: str) -> dict[str, list[dict[str, object]]]:
    """Extract table entries from the source TABLES section."""
    tables: dict[str, list[dict[str, object]]] = {}
    current: str | None = None

    for record_type, body in _records(_section(tags, "TABLES")):
        if record_type == "TABLE":
            current = _first(body, 2)
            if current:
                tables[current] = []
            continue
        if record_type == "ENDTAB":
            current = None
            continue
        if current is None:
            continue

        name = _first(body, 2)
        tables[current].append(
            {
                "name": name,
                "handle": _normalize_handle(version, _first(body, 5)),
                "fields": {"flags": _first(body, 70) or "0"},
            }
        )

    for entries in tables.values():
        entries.sort(key=lambda item: (item["name"], item["handle"]))
    return tables


def _tables_with_ezdxf_handles(
    doc: object,
    source: dict[str, list[dict[str, object]]],
    version: str,
) -> dict[str, list[dict[str, object]]]:
    """Use ezdxf handles where available, filtered to source table entries.

    ezdxf auto-creates missing tables and default entries; the source names keep
    those out. ezdxf also assigns handles to entries that omit code 5, which is
    the same behavior the Go loader has.
    """
    tables: dict[str, list[dict[str, object]]] = {}
    for table_name, source_entries in source.items():
        attr = TABLE_ATTRS.get(table_name)
        table = getattr(doc.tables, attr, None) if attr else None
        by_name = {}
        if table is not None:
            by_name = {entry.dxf.name: entry for entry in table}

        entries = []
        for source_entry in source_entries:
            name = source_entry["name"]
            entry = by_name.get(name)
            handle = _handle(entry, version) if entry is not None else source_entry["handle"]
            entries.append(
                {
                    "name": name,
                    "handle": handle,
                    "fields": source_entry["fields"],
                }
            )
        entries.sort(key=lambda item: (item["name"], item["handle"]))
        tables[table_name] = entries
    return tables


COLLECTION_KEYS = {
    "ACAD_MLEADERSTYLE": "MLEADERSTYLECOLLECTION",
    "ACAD_MLINESTYLE": "MLINESTYLECOLLECTION",
}
COLLECTION_TYPES = {"MLEADERSTYLECOLLECTION", "MLINESTYLECOLLECTION"}


def _dictionary_entries(body: list[tuple[int, str]]) -> dict[str, str]:
    """Return key -> value handle pairs from a DICTIONARY record."""
    entries: dict[str, str] = {}
    key = ""
    for code, value in body:
        if code == 3:
            key = value.strip()
        elif code in (350, 360) and key:
            entries[key] = value.strip()
            key = ""
    return entries


def _root_dictionary_handle(doc: object) -> str:
    try:
        return doc.rootdict.dxf.handle or ""
    except Exception:
        return ""


def _object_fields(record_type: str, body: list[tuple[int, str]]) -> dict[str, str]:
    """Stable typed-field projection. Keys mirror the Go test helper."""
    if record_type in {"DICTIONARY", "DICTIONARYWDFLT", "ACDBDICTIONARYWDFLT"}:
        return {"entry_count": str(sum(1 for code, _ in body if code == 3))}
    if record_type == "LAYOUT":
        return {"flags": _after_marker(body, "AcDbLayout", 70) or "0"}
    if record_type == "MLINESTYLE":
        return {"flags": _first(body, 70) or "0"}
    if record_type == "MLEADERSTYLE":
        return {"flags": _first(body, 179) or "0"}
    if record_type == "GROUP":
        return {"entities": str(sum(1 for code, _ in body if code == 340))}
    return {}


def _source_objects(
    tags: list[tuple[int, str]], version: str, doc: object
) -> list[dict[str, object]]:
    """Extract source object types, typed collection inventory, and handles.

    Dictionaries referenced by the root dictionary's ACAD_MLEADERSTYLE and
    ACAD_MLINESTYLE entries are reported as typed collections without changing
    their source wire type or handle.
    """
    records = [
        (record_type, body)
        for record_type, body in _records(_section(tags, "OBJECTS"))
        if record_type not in {"ENDSEC", "EOF"}
    ]
    by_handle = {_first(body, 5): (record_type, body) for record_type, body in records}

    collection_for: dict[str, str] = {}
    root = by_handle.get(_root_dictionary_handle(doc))
    if root is not None and root[0] == "DICTIONARY":
        for key, target in _dictionary_entries(root[1]).items():
            collection = COLLECTION_KEYS.get(key)
            if collection:
                collection_for[target] = collection

    objects: list[dict[str, object]] = []
    for record_type, body in records:
        raw_handle = _first(body, 5)
        typed_type = record_type
        members: list[str] = []
        if record_type == "DICTIONARY" and raw_handle in collection_for:
            typed_type = collection_for[raw_handle]
            members = sorted(
                _normalize_handle(version, handle)
                for handle in _dictionary_entries(body).values()
            )
        elif record_type in COLLECTION_TYPES:
            members = sorted(
                _normalize_handle(version, handle)
                for handle in _dictionary_entries(body).values()
            )
        objects.append(
            {
                "type": record_type,
                "handle": _normalize_handle(version, raw_handle),
                "typed_type": typed_type,
                "members": members,
                "fields": (
                    {}
                    if typed_type in COLLECTION_TYPES
                    else _object_fields(record_type, body)
                ),
            }
        )
    objects.sort(key=lambda item: (item["type"], item["handle"]))
    return objects


def _normalize_handle(version: str, handle: str) -> str:
    return "" if version in PRE_R13 else handle


def _handle(entity: object, version: str) -> str:
    if version in PRE_R13:
        return ""
    try:
        return entity.dxf.handle or ""
    except Exception:
        return ""


def _layer(entity: object) -> str:
    try:
        return entity.dxf.layer or "0"
    except Exception:
        return "0"


def _float(value: object) -> str:
    try:
        return f"{float(value):.4f}"
    except (TypeError, ValueError):
        return "0.0000"


def _point(value: object) -> str:
    parts = list(value) if value is not None else []
    while len(parts) < 3:
        parts.append(0.0)
    return ",".join(_float(v) for v in parts[:3])


def _entity_fields(entity: object) -> dict[str, str]:
    """Stable typed-field projection. Keys mirror the Go test helper."""
    kind = entity.dxftype()
    dxf = entity.dxf
    if kind == "LINE":
        return {"start": _point(dxf.start), "end": _point(dxf.end)}
    if kind == "CIRCLE":
        return {"center": _point(dxf.center), "radius": _float(dxf.radius)}
    if kind == "ARC":
        return {
            "center": _point(dxf.center),
            "radius": _float(dxf.radius),
            "start_angle": _float(dxf.start_angle),
            "end_angle": _float(dxf.end_angle),
        }
    if kind == "POINT":
        return {"location": _point(dxf.location)}
    if kind == "TEXT":
        return {
            "text": dxf.text or "",
            "height": _float(dxf.height),
            "rotation": _float(dxf.rotation),
        }
    if kind == "MTEXT":
        # MTEXT text assembly differs for column fragments and leading-space
        # normalization; inventory already covers the record.
        return {}
    if kind == "INSERT":
        return {"block_name": dxf.name or "", "insert": _point(dxf.insert)}
    if kind == "LWPOLYLINE":
        return {"vertex_count": str(len(list(entity.get_points())))}
    if kind == "SPLINE":
        return {
            "degree": str(int(dxf.degree)),
            "control_count": str(len(entity.control_points)),
            "knot_count": str(len(entity.knots)),
        }
    if kind == "ELLIPSE":
        return {"center": _point(dxf.center)}
    if kind == "HATCH":
        return {"pattern_name": dxf.pattern_name or ""}
    return {}


def _entity_dump(entity: object, version: str) -> dict[str, object]:
    return {
        "type": entity.dxftype(),
        "handle": _handle(entity, version),
        "layer": _layer(entity),
        "fields": _entity_fields(entity),
    }


def _entities_dump(entities: object, version: str) -> list[dict[str, object]]:
    return sorted(
        (
            _entity_dump(entity, version)
            for entity in entities
            if entity.dxftype() != "SEQEND"
        ),
        key=lambda item: (item["type"], item["handle"], item["layer"]),
    )


def _block_dump(block: object, version: str) -> dict[str, object]:
    return {"name": block.name, "entities": _entities_dump(block, version)}


def _is_layout_block(name: str) -> bool:
    # Layout blocks are represented differently in Go (ENTITIES holds the
    # layout entities), so compare only user-defined block definitions.
    return name.startswith("*")


def _source_only_entities(tags: list[tuple[int, str]], version: str) -> list[dict[str, str]]:
    """Include source records ezdxf does not expose as DXFGraphic entities."""
    out = []
    for record_type, body in _records(_section(tags, "ENTITIES")):
        if record_type == "MESHVERTEXCACHE":
            out.append(
                {
                    "type": "MESHVERTEXCACHE",
                    "handle": _normalize_handle(version, _first(body, 5)),
                    "layer": _first(body, 8) or "0",
                    "fields": {},
                }
            )
    return out


def dump(path: str) -> dict[str, object]:
    source_tags = _read_tags(path)
    doc = ezdxf.readfile(path)
    version = doc.dxfversion

    # ACAD_TABLE is a tag-storage entity rather than a DXFGraphic in ezdxf.
    # Include it explicitly so the inventory matches the source ENTITIES stream.
    entities = [
        entity
        for entity in doc.entitydb.values()
        if (isinstance(entity, DXFGraphic) or entity.dxftype() == "ACAD_TABLE")
        and entity.dxftype() != "SEQEND"
    ]
    entity_dump = _entities_dump(entities, version)
    entity_dump.extend(_source_only_entities(source_tags, version))
    entity_dump.sort(key=lambda item: (item["type"], item["handle"], item["layer"]))

    blocks = sorted(
        (
            _block_dump(block, version)
            for block in doc.blocks
            if not _is_layout_block(block.name)
        ),
        key=lambda item: item["name"],
    )

    source_tables = _source_tables(source_tags, version)

    return {
        "version": version,
        "entities": entity_dump,
        "blocks": blocks,
        # A missing TABLES section means the loader may synthesize the standard
        # defaults; there is no source inventory to compare against.
        "tables": (
            _tables_with_ezdxf_handles(doc, source_tables, version)
            if _has_section(source_tags, "TABLES")
            else None
        ),
        "objects": _source_objects(source_tags, version, doc),
    }


def main(argv: list[str]) -> int:
    if not argv:
        print("usage: ezdxf_dump.py <file.dxf>...", file=sys.stderr)
        return 2

    result = {str(Path(path)): dump(path) for path in argv}
    json.dump(result, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
