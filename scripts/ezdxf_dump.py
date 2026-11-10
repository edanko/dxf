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


def _first(record: list[tuple[int, str]], code: int) -> str:
    for tag_code, value in record:
        if tag_code == code:
            return value.strip()
    return ""


def _source_tables(tags: list[tuple[int, str]], version: str) -> dict[str, list[dict[str, str]]]:
    """Extract table entries from the source TABLES section."""
    tables: dict[str, list[dict[str, str]]] = {}
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
            {"name": name, "handle": _normalize_handle(version, _first(body, 5))}
        )

    for entries in tables.values():
        entries.sort(key=lambda item: (item["name"], item["handle"]))
    return tables


def _tables_with_ezdxf_handles(
    doc: object,
    source: dict[str, list[dict[str, str]]],
    version: str,
) -> dict[str, list[dict[str, str]]]:
    """Use ezdxf handles where available, filtered to source table entries.

    ezdxf auto-creates missing tables and default entries; the source names keep
    those out. ezdxf also assigns handles to entries that omit code 5, which is
    the same behavior the Go loader has.
    """
    tables: dict[str, list[dict[str, str]]] = {}
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
            entries.append({"name": name, "handle": handle})
        entries.sort(key=lambda item: (item["name"], item["handle"]))
        tables[table_name] = entries
    return tables


def _source_objects(tags: list[tuple[int, str]], version: str) -> list[dict[str, str]]:
    """Extract object types and handles from the source OBJECTS section."""
    objects = [
        {
            "type": record_type,
            "handle": _normalize_handle(version, _first(body, 5)),
        }
        for record_type, body in _records(_section(tags, "OBJECTS"))
        if record_type not in {"ENDSEC", "EOF"}
    ]
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


def _entity_dump(entity: object, version: str) -> dict[str, str]:
    return {
        "type": entity.dxftype(),
        "handle": _handle(entity, version),
        "layer": _layer(entity),
    }


def _entities_dump(entities: object, version: str) -> list[dict[str, str]]:
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
        "entities": _entities_dump(entities, version),
        "blocks": blocks,
        # A missing TABLES section means the loader may synthesize the standard
        # defaults; there is no source inventory to compare against.
        "tables": (
            _tables_with_ezdxf_handles(doc, source_tables, version)
            if _has_section(source_tags, "TABLES")
            else None
        ),
        "objects": _source_objects(source_tags, version),
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
