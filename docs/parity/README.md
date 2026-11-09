# ezdxf Feature Parity Roadmap (Go `dxf`)

Status: research baseline, living document.
Last updated: 2026-11-09.

This document is the durable entry point for continuing the "Go `dxf` should match
Python `ezdxf`" effort across sessions and agents. Read this first, then read the
OpenSpec changes listed in the roadmap.

## Reference inputs

| Input | Value |
|---|---|
| Go repository | `github.com/edanko/dxf` at `/home/egor/src/dxf` |
| Go module | `github.com/edanko/dxf`, Go 1.25 |
| ezdxf reference | `/home/egor/src/ezdxf` |
| ezdxf version | `1.4.6b0` (`src/ezdxf/version.py`) |
| ezdxf commit | `8a593d329` |
| Research date | 2026-11-08 |

The ezdxf source tree is the authority for behaviour. Its documentation is under
`/home/egor/src/ezdxf/docs/source/` (especially `reference.rst`, `concepts/`,
`tasks/`, `dxfentities/`, `dxfobjects/`, `sections/`, `tables/`, `math/`,
`path.rst`, `render/`, `tools/`, `addons/`).

## Goal and definition of done

Goal: after all roadmap proposals are implemented, the Go library must be able to
do everything the ezdxf core package and its bundled add-ons can do for DXF
documents, with equivalent observable behaviour:

1. Read/write ASCII and binary DXF for R12, R2000, R2004, R2007, R2010, R2013,
   R2018; read R13/R14 and pre-R12 with upgrade (same policy as ezdxf).
2. Preserve unknown/third-party tags across a load/save round trip.
3. Model every documented DXF entity, table entry, and object with its documented
   attributes, including sub-entities and ACIS data.
4. Provide the high-level document/layout/graphics-factory API (create, query,
   modify, delete, copy, transform, explode, import, export).
5. Provide equivalent geometry, path, bounding-box, query, audit, recover, xref,
   clipping, rendering, and add-on functionality.
6. Pass parity tests derived from ezdxf test data and behaviour, not only Go unit
   tests.

"Parity" means the same input produces the same semantic result; byte-for-byte
output is not required, but round-trip stability and documented tag values are.

## How to continue in a new session

1. Read this file, then `openspec list --json` and `openspec list --specs --json`.
2. Pick the next change from the roadmap. Changes are ordered by dependency.
3. Read its `proposal.md`, `specs/`, `design.md`, and `tasks.md`.
4. Use `/opsx-apply` (or the `openspec-apply-change` skill) to implement it.
5. Keep this document current: update the status matrix and the "Progress log"
   whenever a change is implemented, partially implemented, or re-scoped.
6. Add regression fixtures under `testdata/` and tests next to the package under
   change. Prefer fixtures copied from `/home/egor/src/ezdxf/tests/` or generated
   with ezdxf itself.
7. Never claim parity for an area without an executable test that exercises the
   documented ezdxf behaviour.

Useful commands:

```sh
cd /home/egor/src/dxf
go build ./...
go test ./...
openspec list --json
openspec status --change "<name>" --json
openspec instructions <artifact> --change "<name>" --json
```


## Current architecture (Go)

| Package | Role | Baseline assessment |
|---|---|---|
| root (`dxf_main.go`, `binder.go`) | public entry points, ASCII/binary load, document binding, entity dispatch | `FromReader`/`FromBinaryReader` bind all six core sections through `documentBinder`; `parser.go` legacy helpers remain unused |
| `lldxf` | tagger, tags, encoding, loader, validator, binary tagger | section grouping, comments per section, section order, recovery issues |
| `format` | ASCII and binary formatter | ASCII works; binary write exists but coverage is partial |
| `drawing` | document, sections, entity database, serialization | binds and re-emits header/classes/tables/blocks/entities/objects plus opaque sections; comments and section order retained |
| `header` | header variables | typed core variables plus `Extra` retention for every other `$...` |
| `table` | APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE, STYLE, UCS, VIEW, VIEWPORT | all entries parsed; unknown entry types retained as `RawSymbolTable` |
| `block` | block definitions | names, handles, layer names, contained entities |
| `entity` | entity structs and `Format` methods | broad type inventory; raw-tag retention on known types; `GenericEntity` fallback |
| `object` | OBJECTS section | dictionary, group, xrecord, placeholder parsed; everything else retained as `RawObject` |
| `class` | CLASSES section | full class record fields |
| `math` | vectors, matrices, bbox, Bezier, BSpline/NURBS, bulge, OCS/UCS, offset, earcut, delaunay, construct | substantial but not full ezdxf math surface |
| `path` | path commands, converter, tools | line/arc/circle/ellipse/lwpolyline/polyline only; no spline/hatch/mtext conversion |
| `render` | simple renderer, SVG/PDF backends, forms, arrows | only line/circle/arc/lwpolyline rendered; no fill/hatch/dimension/mtext/mesh |
| `export` | SVG exporter | only line/circle/arc/lwpolyline/polyline/point |
| `text2path` | text outline conversion | placeholder glyph geometry, no real font/shx parsing |
| `query` | entity query and parser | useful selector/expression subset |
| `audit` | auditor, recovery, block cycles | partial; some placeholder loading |
| `spatial` | R-tree and entity spatial index | usable |
| `layouts` | model/paper layouts and viewports | partial |
| `xref` | xref manager and loader | partial; no real document binding |
| `factory` | entity factory/registry | usable |
| `csg` | mesh boolean CSG | simplified/placeholder booleans |
| `patterns` | hatch pattern definitions | lookup tables only |
| `xdata`, `class`, `color`, `insunit`, `handle`, `geometry` | support packages | usable subsets |

## Status legend

- `ABSENT` — no implementation.
- `STUB` — API exists but returns placeholder/empty/error, or silently drops data.
- `PARTIAL` — works for a subset of documented behaviour.
- `DONE` — covered by executable parity tests (must state test location).
- `N/A` — Python-only concern with no meaningful Go equivalent.

The baseline below is deliberately conservative. Items marked `PARTIAL`/`STUB`
must be verified before implementation starts.

## Feature matrix

### 1. Document model and DXF I/O

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| New document, version selection | `document.py`, `sections/header.py` | `DONE` — R12–R2018 profiles (`format/version.go`), `$ACADVER` detection, normalization, target-version save (`TestVersionProfiles`, `TestSaveToTargetVersion`) |
| HEADER variables (all `$...`) | `lldxf/hdrvars.py`, `sections/headervars.py` | `PARTIAL` — typed core set; other variables retained in `Header.Extra` and re-emitted (`dxf_io_core_test.go`) |
| CLASSES section | `sections/classes.py` | `DONE` — parsed and re-emitted (`TestClassesRoundTrip`) |
| TABLES section + all entries | `sections/tables.py`, `tables/*` | `PARTIAL` — all entry types typed with shared base; unknown entries retained raw |
| BLOCKS section + block records | `sections/blocks.py`, `entities/block.py` | `DONE` — names, handles, layer names, contained entities (`TestLoadBindsAllDocumentSections`, `TestCorpusRoundTrip`) |
| ENTITIES section | `sections/entities.py` | `PARTIAL` — many types via `GenericEntity` fallback |
| OBJECTS section | `sections/objects.py`, `dxfobjects/*` | `PARTIAL` — standard object types typed, unknown retained raw, retrieval by handle (`TestObjectsRetrievableByHandle`) |
| THUMBNAILIMAGE section | `document.py` | `DONE` — retained opaque and re-emitted (`TestOpaqueSectionsAreRetained`) |
| ACDSDATA section | `sections/acdsdata.py` | `DONE` — retained opaque and re-emitted |
| Unknown tag retention | `entities/dxfentity.py` | `PARTIAL` — known entities keep unhandled codes; unknown table entries/blocks/typed objects still drop some tags |
| Load/save round trip | `document.py` | `DONE` — all 51 `testdata/*.dxf` round-trip with matching inventories (`TestCorpusRoundTrip`) |
| Recover mode | `recover.py` | `PARTIAL` — missing ENDSEC/EOF and tagger errors recovered with a report (`TestRecoveryLoadsTruncatedDocument`) |
| Comments (`999`) | `comments.py` | `PARTIAL` — preserved per section and re-emitted at section start; original position is not retained (`TestCommentsArePreservedUnlessStripped`) |
| R12 strict mode | `r12strict.py` | `DONE` — sanitizer drops non-R12 tables/entities/layers, true-color header variables, records conversion report (`TestR12Downgrade`, `drawing/r12strict.go`) |
| Binary DXF read/write | `lldxf/tagger.py`, `tagwriter.py` | `DONE` — binary writer emits all sections + EOF with per-code int16/int32/int64/float/raw sizing; ASCII/binary tag streams are value-identical (`TestBinaryASCIIEquivalence`) |
| Encoding/codepage handling | `lldxf/encoding.py`, `tools/codepage.py` | `DONE` — codepage-aware decode on load (ANSI_1252 etc.), UTF-8 for R2007+, codepage output for older targets (`TestCodepageDecoding`, `TestUTF8AndCodepageOutput`) |
| File management (temp/backup/zip) | `filemanagement.py`, `tools/zipmanager.py` | `ABSENT` |

### 2. Entities

ezdxf registers 99 `DXFTYPE` values in `src/ezdxf/entities/`. The Go loader
registry (root `loadEntity`, registered via `entity.SetDefaultLoader` +
`registerEntityLoaders`) now constructs typed entities for every standard
drawing entity and retains every other record verbatim.

Typed load dispatch (root loaders, verified by `TestRegistryCompleteness`):
POINT, LINE, CIRCLE, ARC, TEXT, MTEXT, SOLID, TRACE, 3DFACE, LWPOLYLINE,
POLYLINE, VERTEX, SEQEND, INSERT, ATTRIB, ATTDEF, ELLIPSE, SPLINE, LEADER,
MLeader/MULTILEADER (alias), HATCH, MPOLYGON (typed hatch with record-name
override), IMAGE, DIMENSION (+ ARC_DIMENSION, LARGE_RADIAL_DIMENSION aliases),
SHAPE, XLINE, RAY, WIPEOUT, MLINE, TOLERANCE, 3DSOLID, REGION, BODY, MESH,
HELIX, EXTRUDED_SURFACE, LOFTED_SURFACE, REVOLVED_SURFACE, SWEPT_SURFACE,
PLANESURFACE.

Whole-record retention (`entity.RetainedRecordType`, default loader; verified
by `TestRetainedRecordSurvival`, `TestLightSunRetainedRoundTrip`,
`TestFieldRetainedRoundTrip`, `TestAcadTableRetainedRoundTrip`): ACAD_TABLE,
ACAD_PROXY_ENTITY, OLE2FRAME, LIGHT, SUN, FIELD, LAYOUT-as-entity records,
third-party/app-defined types, and any unregistered standard type. Retained
records re-emit every tag verbatim (code 0/handle from the base), so
unknown/third-party entities survive load/save unchanged.

Entity operations (`entity` package): `Transform` for every geometry category
(line/curve/text/polyline/insert/hatch/dimension/mleader/leader/solid3d) with
matrix-composition tests; `BBox` including INSERT block content
(`TestInsertBBoxContent`); `Explode` for INSERT, POLYLINE, HATCH (boundary
polylines), DIMENSION, LEADER, MLEADER (`explode_roundtrip_test.go`).

ACIS payload limitations: text SAT payloads (code 90 count + code 10 lines) on
3DSOLID/REGION/BODY are retained whole-record and round-trip verbatim
(`acis_roundtrip_test.go`); binary SAB payloads (codes 90/310) survive through
the same whole-record retention. Go does not parse ACIS bodies, history trees,
or SAT version semantics; typed accessors (volume/area/center) are only
available on payload-free records. HATCH edge-path boundaries and ELLIPSE
records with embedded splines fall back to whole-record retention.

Missing or only partially modelled (non-exhaustive):

- Tables: `ACAD_TABLE` typed content rows/columns, `TABLESTYLE`.
- Multileaders: `MLEADERSTYLE`, `MLEADERSTYLECOLLECTION`.
- MLine: `MLINESTYLE`, `MLINESTYLECOLLECTION`.
- Images/underlays: `IMAGEDEF`, `IMAGEDEF_REACTOR`, `PDFDEFINITION`,
  `PDFREFERENCE`, `PDFUNDERLAY`, `DGNUNDERLAY`, `DWFUNDERLAY`,
  `UNDERLAYDEFINITION`.
- Objects-as-entities: `LAYOUT`, `PLOTSETTINGS`, `RASTERVARIABLES`,
  `WIPEOUTVARIABLES`, `IDBUFFER`, `FIELDLIST`, `SPATIAL_FILTER`, `GEODATA`
  (retained, not typed).
- Sub-entities: `VERTEX` variants (`POLYFACE`, `POLYMESH`),
  `MESHVERTEXCACHE`.
- ACIS: body history parsing, SAT/SAB binary decoding.

Per-entity requirements to verify for every type: all documented group codes,
defaults, version gating, OCS/extrusion handling, subclass markers, XDATA/APPID,
reactors, extension dictionaries, copy/clone semantics, and `Format` output.

### 3. Tables and objects

| Area | ezdxf reference | Go baseline |
|---|---|---|
| APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE, STYLE, UCS, VIEW, VPORT | `tables/*`, `entities/*` | `PARTIAL` — shared `EntryBase` (handle/owner/flags/xdata/extdict/reactors), generic lookup/add/remove/rename with duplicate-name errors, typed parse/format (`TestParseTablesEntryCounts`, `TestTableLookupAddRemoveRename`, `TestTableHandlesOwnersUnique`) |
| Dictionaries, XRECORD, DICTIONARYVAR, DICTIONARYWITHDEFAULT | `entities/dictionary.py` | `PARTIAL` — typed round trip and key lookup (`TestDictionary*`) |
| GROUP | `entities/dxfgroups.py` | `PARTIAL` — typed round trip (`TestGroup*`) |
| LAYOUT, PLOTSETTINGS | `entities/layout.py`, `dxfobjects/plotsettings.py` | `PARTIAL` — typed name/block record/flags, retained round trip (`TestLayoutNameAndBlockRecord`) |
| MLINESTYLE, MLEADERSTYLE, TABLESTYLE | `entities/mlinestyle.py`, `mleader.py`, `table.py` | `PARTIAL` — typed name/description/flags, retained round trip (`TestStandardObjectTypedModels`) |
| MATERIAL, VISUALSTYLE, SUN, LIGHT, GEODATA | `entities/material.py`, `visualstyle.py`, `sun.py`, `light.py`, `geodata.py` | `PARTIAL` — typed name/description/flags, retained round trip |
| SPATIAL_FILTER, IDBUFFER, RASTERVARIABLES, WIPEOUTVARIABLES | `entities/spatial_filter.py`, `idbuffer.py` | `PARTIAL` — typed models, retained round trip |
| SORTENTSTABLE, FIELDLIST, IMAGEDEF, IMAGEDEF_REACTOR, UNDERLAYDEFINITION, PDF/DWF/DGN DEFINITION | `entities/sortentstable.py`, `field.py`, `image.py`, `underlay.py` | `PARTIAL` — typed sort order, entity handles, filename/size (`TestImageDefPathAndSize`, `TestSortEntsTableOrder`) |
| ACAD_PROXY_ENTITY, ACAD_XREC_ROUNDTRIP | `entities/acad_proxy_entity.py` | `PARTIAL` |
| Extension dictionary/reactors on every entity | `xdict.py`, `appdata.py` | `PARTIAL` — preserved on typed table entries and objects (`TestObjectReactorsAndExtDict`) |

### 4. High-level document and layout API

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| `doc.modelspace()`, paper space layouts | `layouts/*` | `PARTIAL` |
| `layout.add_*` factory methods for every entity | `graphicsfactory.py` | `PARTIAL` — root has helpers for a subset |
| `dxfattribs` handling | `gfxattribs.py` | `PARTIAL` |
| Entity linking (INSERT/ATTRIB/SEQEND, POLYLINE/VERTEX) | `entitydb.py`, `entities/subentity.py` | `PARTIAL` |
| Entity copy/transform/explode | `entities/copy.py`, `transform.py`, `explode.py` | `PARTIAL` |
| Blocks: create, insert, explode, purge, rename | `blkrefs.py`, `entities/block.py` | `PARTIAL` |
| Header variable API | `document.py` | `PARTIAL` |
| Query/select API | `query.py`, `select.py` | `PARTIAL` |
| Transaction/undo model | `entitydb.py` | `STUB` |

### 5. Math and geometry

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Vec2/Vec3, Matrix44, transformations | `math/_vector.py`, `_matrix44.py`, `transformtools.py` | `PARTIAL` |
| Bounding boxes (2D/3D, extents) | `math/bbox.py`, `bbox.py` | `PARTIAL` |
| Bezier (3P/4P, derivatives, flattening) | `math/_bezier3p.py`, `_bezier4p.py` | `PARTIAL` |
| B-spline/NURBS (knots, derivatives, interpolation) | `math/_bspline.py`, `bspline.py`, `cspline.py` | `PARTIAL` |
| Bulge/arc helpers | `math/bulge.py`, `arc.py` | `PARTIAL` |
| OCS/UCS | `math/ucs.py`, `concepts/ocs.rst` | `PARTIAL` |
| Offset 2D | `math/offset2d.py` | `PARTIAL` |
| Triangulation (earcut, Delaunay) | `math/_mapbox_earcut.py`, `triangulation.py` | `PARTIAL` |
| Clipping (polygon, line, portal) | `math/clipping.py`, `tools/clipping_portal.py` | `ABSENT`/`STUB` |
| Shape/construct2d/construct3d | `math/shape.py`, `construct2d.py`, `construct3d.py` | `PARTIAL` |
| R-tree | `math/rtree.py` | `DONE`-ish (`spatial`) |
| Perlin noise, clustering, parametrization | `math/perlin.py`, `clustering.py`, `parametrize.py` | `ABSENT`/`STUB` |
| Intersections, curve tools, legacy math | `math/curvetools.py`, `legacy.py` | `PARTIAL` |

### 6. Path, rendering, visualization

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Path commands and converter for all entities | `path/*` | `PARTIAL` — no spline/hatch/mtext/image/leader conversion |
| Path tools: flatten, transform, offset, nesting | `path/tools.py`, `path/nesting.py` | `PARTIAL` — no nesting |
| Drawing add-on / render pipeline | `addons/drawing/*`, `render/*` | `PARTIAL` — line/circle/arc/lwpolyline only |
| SVG/PDF backends | `addons/drawing/backend.py`, `pymupdf.py` | `PARTIAL` — basic backends |
| matplotlib/PyQt/PyMuPDF backends | `addons/drawing/*` | `N/A` in Go; provide native backends and image output |
| Linetype rendering | `render/_linetypes.py`, `linetypes.py` | `PARTIAL` |
| Hatch rendering | `render/hatching.py` | `STUB` |
| Dimension rendering | `render/dimension.py`, `dim_*.py` | `STUB` |
| MTEXT rendering | `render/abstract_mtext_renderer.py`, `tools/text_layout.py` | `STUB` |
| Mesh rendering | `render/mesh.py`, `mesh_builder.py` | `PARTIAL` |
| Arrow heads | `render/arrows.py` | `PARTIAL` |
| MLeader/MLine rendering | `render/mleader.py`, `mline.py` | `ABSENT` |

### 7. Audit, recover, validation, repair

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Audit protocol and fixers | `audit.py` | `PARTIAL` |
| Recover damaged files | `recover.py` | `ABSENT` |
| R12 strict checker | `r12strict.py` | `ABSENT` |
| Tag/validator checks | `lldxf/validator.py`, `lldxf/repair.py` | `PARTIAL` |
| Orphan/cycle repair | `audit.py` | `PARTIAL` |
| Header/table/block/object consistency | `audit.py` | `PARTIAL` |

### 8. Query, selection, transformation, decomposition

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Query language | `query.py`, `queryparser.py` | `PARTIAL` |
| `select` filters | `select.py` | `PARTIAL` |
| `groupby` | `groupby.py` | `ABSENT` |
| Transform entities/layouts | `transform.py` | `PARTIAL` |
| Explode INSERT/POLYLINE/MTEXT/etc. | `explode.py`, `addons/mtextsurrogate` | `PARTIAL` |
| Bounding box of entities/layouts | `bbox.py` | `PARTIAL` |
| Disassemble to primitives | `disassemble.py` | `ABSENT` |
| Upright/rotate-to-view | `upright.py` | `ABSENT` |
| Reorder/seqend repair | `reorder.py` | `ABSENT` |
| Block reference traversal | `blkrefs.py` | `PARTIAL` |

### 9. XRef, images, underlays, clipping

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| XREF load/attach/overlay | `xref.py` | `PARTIAL` |
| Image definitions and rendering | `entities/image.py` | `PARTIAL` |
| PDF/DGN/DWF underlays | `entities/underlay.py` | `PARTIAL` |
| XCLIP boundaries | `xclip.py` | `PARTIAL` |
| Nested xref resolution | `xref.py` | `PARTIAL` |

### 10. Add-ons

| Add-on | ezdxf reference | Go baseline |
|---|---|---|
| `drawing` | `addons/drawing` | `PARTIAL` (`render`, `export`) |
| `r12writer` | `addons/r12writer.py` | `ABSENT` |
| `iterdxf` | `addons/iterdxf.py` | `ABSENT` |
| `importer` | `addons/importer.py` | `ABSENT` |
| `dxf2code` | `addons/dxf2code` | `ABSENT` |
| `acadctb` | `addons/acadctb.py` | `ABSENT` |
| `pycsg` | `addons/pycsg.py` | `STUB` (`csg`) |
| `MTextExplode` | `addons/mtextsurrogate` | `ABSENT` |
| `text2path` | `addons/text2path.py` | `STUB` |
| `geo` | `addons/geo.py` | `ABSENT` |
| `meshex` | `addons/meshex.py` | `ABSENT` |
| `openscad` | `addons/openscad.py` | `ABSENT` |
| `odafc` | `addons/odafc.py` | `ABSENT` |
| `hpgl2` | `addons/hpgl2` | `ABSENT` |
| `xplayer` | `addons/xplayer.py` | `ABSENT` |
| `binpacking`, `tablepainter`, `gerber_D6673`, `mtxpl` | `addons/*` | `ABSENT` |

### 11. CLI and tooling

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| `ezdxf` CLI (`info`, `audit`, `draw`, `browse`, `convert`, `strip`, `view`, `pp`) | `commands.py`, `__main__.py` | `ABSENT` |
| App settings / config | `appsettings.py`, `_options.py` | `ABSENT` |
| Comments | `comments.py` | `PARTIAL` |
| Fonts, SHX/TTF, text metrics | `fonts/*`, `tools/text*.py` | `STUB` |
| ISO hatch patterns / standards | `tools/_iso_pattern.py`, `standards.py` | `PARTIAL` |
| Revcloud | `revcloud.py` | `ABSENT` |
| Zoom/extents helpers | `zoom.py` | `ABSENT` |
| Analyze/difftags/debug tools | `tools/analyze.py`, `difftags.py` | `ABSENT` |

### 12. Colors, enums, units, standards

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| ACI colors, true color, transparency | `colors.py`, `concepts/true_color.rst` | `PARTIAL` |
| Enums (text alignment, hatch, etc.) | `enums.py` | `PARTIAL` |
| Units and insunits | `units.py`, `concepts/units.rst` | `PARTIAL` |
| Standards checks | `tools/standards.py` | `ABSENT` |
| Lineweights | `concepts/lineweights.rst` | `PARTIAL` |
| Linetype patterns | `tools/pattern.py`, `_linetypes.py` | `PARTIAL` |

## Roadmap: ordered OpenSpec changes

Each row becomes one OpenSpec change. Implement top to bottom. Later changes
assume the earlier ones.

| # | Change name | Scope | Depends on |
|---|---|---|---|
| 1 | `dxf-io-core-parity` | full section parse/bind/write, unknown tag retention, round trip, recover hooks, comments, thumbnail/ACDSDATA scaffolding | — |
| 2 | `dxf-version-encoding-parity` | R12–R2018 read/write/new, R13/R14 + pre-R12 upgrade, ASCII/binary, codepages, full header variables | 1 |
| 3 | `entity-model-parity` | every ezdxf entity type, full attributes, sub-entities, ACIS, proxy, per-entity tests | 1, 2 |
| 4 | `table-object-model-parity` | all table entries, all OBJECTS types, dictionaries, layouts, groups, extension dictionaries/reactors | 1, 3 |
| 5 | `graphics-factory-api-parity` | `add_*` factory for all entities, `dxfattribs`, entity linking, block/insert/explode API, transactions | 3, 4 |
| 6 | `geometry-math-parity` | vectors/matrices, bbox/extents, Bezier/BSpline/NURBS, bulge, OCS/UCS, offset, triangulation, clipping, intersections, construct2d/3d | — |
| 7 | `path-render-parity` | path conversion for all entities, path tools/nesting, drawing pipeline, SVG/PDF/raster backends, linetype/hatch/dimension/mtext/mesh/mleader rendering | 3, 6 |
| 8 | `audit-recover-parity` | audit protocol/fixers, recover, r12strict, validator/repair, consistency checks | 1–4 |
| 9 | `query-transform-parity` | query language, select, groupby, transform, explode, bbox, disassemble, upright, reorder, blkrefs | 3, 6 |
| 10 | `xref-image-underlay-parity` | xref load/bind, images, PDF/DGN/DWF underlays, xclip | 3, 4, 9 |
| 11 | `addons-parity` | importer, r12writer, iterdxf, dxf2code, acadctb, csg, meshex, geo, openscad, odafc, hpgl2, mtext explode, xplayer, binpacking, tablepainter, gerber | 1–10 |
| 12 | `cli-tooling-parity` | CLI commands, appsettings/options, filemanagement/zip, fonts/text metrics, revcloud, zoom, analysis tools, standards | 1–11 |
| 13 | `color-enum-units-parity` | ACI/true color/transparency, enums, units/insunits, lineweights, linetype patterns, ISO patterns | — |

Rationale for ordering: I/O correctness and version support are prerequisites for
trustworthy round-trip tests; entity/table/object modelling is the data substrate;
the high-level API depends on it; math and rendering can proceed in parallel but
are placed after the data model to avoid rework; audit/query/transform/xref depend
on complete data; add-ons and CLI depend on everything.

## Verification strategy

1. **Round-trip fixtures.** For every `testdata/*.dxf` and every fixture copied
   from `ezdxf/tests/`, assert load → save → load produces an equivalent document
   (entities, attributes, tables, blocks, objects, unknown tags).
2. **Golden tag tests.** For each entity, assert exact tag code/value pairs for
   representative instances, including version-specific tags.
3. **Cross-check with ezdxf.** Generate expected tag dumps with the Python library
   in a throwaway venv when semantics are unclear. Never add a runtime Python
   dependency to the Go library.
4. **Behavioural tests.** Port ezdxf tests where practical, especially
   `tests/test_*.py` for math, path, query, audit, transform, xref, rendering.
5. **Fuzzing.** Fuzz the tagger/loader/parser with truncated and malformed files;
   recover mode must not panic.
6. **Performance smoke tests.** `iterdxf`-style streaming must keep memory bounded.

## Known WIP and hazards

- `parser.go` currently contains unused `Parse*` helpers with placeholder
  implementations (`ParseObjects`, `ParseDimension`, `ParseHatch`, `ParseLeader`,
  `ParseMLeader`, `ParsePolyline`, `ParseVertex`, `ParseImage`, `ParseSeqEnd`)
  and incorrect normal-vector indices in `ParseSpline`/`ParseEllipse`. Do not
  build on them; replace with the entity-model work in change 3.
- `drawing.FromReader` only binds ENTITIES. HEADER/TABLES/BLOCKS/OBJECTS loaded
  by `lldxf.Loader` are discarded.
- `drawing.FromJSON` is a no-op.
- `audit` has placeholder file loading and incomplete repair.
- `csg` boolean operations are bounding-box approximations, not real CSG.
- `text2path` glyphs are placeholders, not font-derived outlines.
- `render`/`export` silently ignore unsupported entity types.
- `spatial` comment admits placeholder bbox handling; verify before relying on it.
- The repository has a large untracked `.agents/`, `.dsh/`, `.zcode/` skill
  configuration. Treat as tooling, not library code.

## Progress log

| Date | Change | Status | Notes |
|---|---|---|---|
| 2026-11-08 | research baseline | done | This document; staged test/CSG work committed as `test: add comprehensive test coverage and CSG module` |
| 2026-11-08 | proposals 1–13 | created | All 13 changes scaffolded with proposal, delta spec, design, and tasks; `openspec validate --changes --strict` passes (13/13) |
| 2026-11-08 | implementation | not started | Begin with change 1 (`dxf-io-core-parity`) via `/opsx-apply`; do not build on the WIP `parser.go` helpers |
| 2026-11-08 | change 1 `dxf-io-core-parity` | done | Two-phase load/bind: `documentBinder` binds HEADER/CLASSES/TABLES/BLOCKS/ENTITIES/OBJECTS; raw-tag retention; opaque sections; comments; recovery; writer section order and unique handles. All 51 fixtures round-trip. Remaining differences: unknown tags on known table entries/blocks/typed objects are partially dropped, comments emit at section start, binary writer unchanged until change 2. Tests: `dxf_io_core_test.go`, `corpus_roundtrip_test.go`, `fuzz_core_test.go`. |
| 2026-11-09 | change 3 `entity-model-parity` | done | Loader registry covers every standard drawing entity with typed structs (37 registered types; `TestRegistryCompleteness`); MPOLYGON typed via hatch, dimension variants, MLEADER/MULTILEADER alias. Whole-record retention (`RetainedRecordType`) for ACAD_TABLE/ACAD_PROXY_ENTITY/OLE2FRAME/LIGHT/SUN/FIELD/third-party records; point-code tagger cast preserves unparseable strings so ACIS text payloads survive. Operations: Transform (matrix composition tests), BBox incl. INSERT content, Explode for INSERT/POLYLINE/HATCH/DIMENSION/LEADER/MLEADER. Tests: `geom_roundtrip_test.go`, `text_roundtrip_test.go`, `polyline_roundtrip_test.go`, `insert_roundtrip_test.go`, `hatch_roundtrip_test.go`, `mpolygon_roundtrip_test.go`, `leader_roundtrip_test.go`, `mline_roundtrip_test.go`, `dimension_roundtrip_test.go`, `explode_roundtrip_test.go`, `transform_compose_test.go`, `transform_more_test.go`, `bbox_insert_test.go`, `acis_roundtrip_test.go`, `surface_roundtrip_test.go`, `misc_roundtrip_test.go`, `registry_completeness_test.go`. ACIS bodies are retained, not parsed. |
| 2026-11-09 | change 2 `dxf-version-encoding-parity` | done | Version profiles R12–R2018 (`format/version.go`); `$ACADVER` detect/normalize; target-version save with conversion report; R12 sanitizer incl. true-color header variables; codepage-aware load + UTF-8/codepage output; binary writer completed (all sections, EOF, signature, per-code int16/int32/int64 sizing, bool codes as int16 flags) — ASCII and binary tag streams now value-identical on `columns_R2018.dxf` (7734 tags). Fixed code-class lists in `lldxf/tags.go` (string codes 100–109, bool 290–299, int64 160–169) and the binary tagger read path. Tests: `version_encoding_test.go`. |

Append a row whenever a change moves state. Keep the matrix above in sync.
