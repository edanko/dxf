# ezdxf Feature Parity Roadmap (Go `dxf`)

Status: implementation complete for roadmap changes 1–18 plus the `core-fidelity-parity` and `tooling-completeness-parity` follow-ups; remaining gaps are tracked as follow-up OpenSpec changes.
Last updated: 2026-11-11.

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
| root (`dxf_main.go`, `binder.go`) | public entry points, ASCII/binary load, document binding, entity dispatch | `FromReader`/`FromBinaryReader` bind all six core sections through `documentBinder`; legacy `parser.go` helpers removed |
| `lldxf` | tagger, tags, encoding, loader, validator, binary tagger | section grouping, comments per section, section order, recovery issues |
| `format` | ASCII and binary formatter | `DONE` — complete binary writer (all sections, EOF, per-code sizing) |
| `drawing` | `addons/drawing` | `DONE` — `addons/drawing.RenderSVG`/`Render` over the core `render` backends (`TestRenderSVG`) |
| `header` | header variables | typed core variables plus `Extra` retention for every other `$...` |
| `table` | APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE, STYLE, UCS, VIEW, VIEWPORT | all entries parsed; unknown entry types retained as `RawSymbolTable` |
| `block` | block definitions | names, handles, layer names, contained entities |
| `entity` | entity structs and `Format` methods | broad typed inventory (54 registered types, including objects-as-entities and MESHVERTEXCACHE) plus whole-record retention for unmodelled records |
| `object` | OBJECTS section | typed dictionaries, groups, xrecords, layouts, plot settings, styles, image/underlay definitions; unknown objects retained as `RawObject` |
| `class` | CLASSES section | full class record fields |
| `math` | vectors, matrices, bbox, Bezier, BSpline/NURBS, bulge, OCS/UCS, offset, earcut, delaunay, construct, Perlin, clustering, Shape, portal clipping | substantial; math parity for changes 5 and 15 complete |
| `path` | path commands, converter, tools, nesting | converters for all supported entities; offset and nesting implemented |
| `render` | scene builder, SVG/PDF/raster backends, forms, arrows | `DONE` — core entity rendering including DIMENSION block content/measured text, mesh and polyface faces, and all arrow styles (`render/entities_parity_test.go`, `TestArrowHeadGeometryStyles`) |
| `export` | SVG exporter | legacy exporter subset; use `render`/`addons/drawing` for full output |
| `text2path` | `addons/text2path.py` | `DONE` — TTF glyph outlines (`addons/text2path`) plus built-in vector fallback |
| `query` | entity query and parser | expression language, selections, grouping, transforms, explode, traversal |
| `audit` | auditor, recovery, block cycles | `DONE` — audit/fix/recover/r12strict, deterministic duplicate-handle repair, and block-cycle repair (`audit/duplicate_handle_fix.go`, `audit/block_cycle_check.go`) |
| `spatial` | R-tree and entity spatial index | usable |
| `layouts` | model/paper layouts and viewports | create/rename/delete/activate; block purge absent |
| `xref` | xref manager and loader | `DONE` — attach/detach/bind/reload, relative paths + search paths, nested loading with cycle detection, BLOCK_RECORD xref flags (`xref_parity_e2e_test.go`) |
| `factory` | entity factory/registry | usable |
| `csg` | mesh boolean CSG | `DONE` — triangle-level booleans, primitives, volume (`csg/boolean_test.go`) |
| `patterns` | hatch pattern definitions | `DONE` — predefined ANSI/ISO tables, `.pat` parser with recoverable file/line errors, case-insensitive custom registry preserving source names, non-mutating scaling and measurement conversion (`patterns/pat_test.go`) |
| `edges` | edge topology and hatch boundary construction | `DONE` — immutable edges, tolerance spatial-hash graph (degree/leaves/components), closed-loop and open-chain analysis preserving source orientation, ordered LWPOLYLINE/POLYLINE creation, hatch boundary from closed loops with `ErrOpenChain` (`edges/edges_test.go`) |
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
| Unknown tag retention | `entities/dxfentity.py` | `DONE` — typed entities, table entries, blocks, and objects keep an ordered per-record ledger of unhandled source tags and re-emit them in source order (`TestUnknownEntityCodeRetained`, `TestUnknownTableAndObjectCodesRetained`) |
| Load/save round trip | `document.py` | `DONE` — all 54 `testdata/*.dxf` round-trip with matching inventories (`TestCorpusRoundTrip`) |
| Recover mode | `recover.py` | `DONE` — missing ENDSEC/EOF and tagger errors recovered with a report; malformed entity records are skipped and counted; `RecoverFile`/`RecoverReader` add a merged audit pass (`TestRecoveryLoadsTruncatedDocument`, `TestRecoverMalformedEntity`, `audit/recover.go`) |
| Comments (`999`) | `comments.py` | `DONE` — stored as positioned tokens anchored to the preceding code-0 record count and replayed at the same boundary; `StripComments` clears both maps and positions (`TestCommentsKeepInterEntityPosition`, `TestCommentsArePreservedUnlessStripped`) |
| R12 strict mode | `r12strict.py` | `DONE` — sanitizer drops non-R12 tables/entities/layers, true-color header variables, records conversion report (`TestR12Downgrade`, `drawing/r12strict.go`) |
| Binary DXF read/write | `lldxf/tagger.py`, `tagwriter.py` | `DONE` — binary writer emits all sections + EOF with per-code int16/int32/int64/float/raw sizing; ASCII/binary tag streams are value-identical (`TestBinaryASCIIEquivalence`) |
| Encoding/codepage handling | `lldxf/encoding.py`, `tools/codepage.py` | `DONE` — codepage-aware decode on load (ANSI_1252 etc.), UTF-8 for R2007+, codepage output for older targets (`TestCodepageDecoding`, `TestUTF8AndCodepageOutput`) |
| File management (temp/backup/zip) | `filemanagement.py`, `tools/zipmanager.py` | `DONE` — atomic write + optional backup (`filemanagement`); ZIP manager creates/reads/extracts archives with entry validation, traversal/symlink rejection, and atomic output (`TestZipRoundTripPreservesEntries`, `TestExtractZipRejectsTraversal`; `filemanagement/zip.go`) |

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
PLANESURFACE, ACAD_TABLE (typed rows/cells/styles/merges), MESHVERTEXCACHE
(also `MESH_VERTEX_CACHE` alias), and objects-as-entities LAYOUT,
PLOTSETTINGS, RASTERVARIABLES, WIPEOUTVARIABLES, IDBUFFER, FIELDLIST,
SPATIAL_FILTER, GEODATA.

Whole-record retention (`entity.RetainedRecordType`, default loader; verified
by `TestRetainedRecordSurvival`, `TestLightSunRetainedRoundTrip`,
`TestFieldRetainedRoundTrip`): ACAD_PROXY_ENTITY, OLE2FRAME, LIGHT, SUN,
FIELD, third-party/app-defined types, and any unregistered standard type.
Retained records re-emit every tag verbatim (code 0/handle from the base), so
unknown/third-party entities survive load/save unchanged.

Entity operations (`entity` package): `Transform` for every geometry category
(line/curve/text/polyline/insert/hatch/dimension/mleader/leader/solid3d) with
matrix-composition tests; `BBox` including INSERT block content
(`TestInsertBBoxContent`); `Explode` for INSERT, POLYLINE, HATCH (boundary
polylines), DIMENSION, LEADER, MLEADER (`explode_roundtrip_test.go`).

ACIS payloads: text SAT payloads on 3DSOLID/REGION/BODY are retained and
round-trip verbatim (`acis_roundtrip_test.go`); the loader reports SAT format,
ACIS version, record count, body count, and history/end markers
(`entity.ParseSATMetadata`). Binary SAB payloads (codes 90/310, including
ACDSDATA records linked by handle) are preserved byte-for-byte and report
binary ACIS presence (`entity.ParseSABMetadata`, `acis_metadata_test.go`).
Decoding is intentionally shallow: full ACIS body geometry and history-tree
reconstruction are out of scope. HATCH edge-path boundaries and ELLIPSE
records with embedded splines fall back to whole-record retention.

Missing or only partially modelled (non-exhaustive):

- Tables: `TABLESTYLE` is typed but not fully wired into table creation/editing.
- Images/underlays: `PDFREFERENCE`, `PDFUNDERLAY`, `DGNUNDERLAY`,
  `DWFUNDERLAY` are retained; definition objects are typed.
- ACIS: full body geometry/history-tree decoding; typed metadata is shallow by
  design.

Per-entity requirements to verify for every type: all documented group codes,
defaults, version gating, OCS/extrusion handling, subclass markers, XDATA/APPID,
reactors, extension dictionaries, copy/clone semantics, and `Format` output.

### 3. Tables and objects

| Area | ezdxf reference | Go baseline |
|---|---|---|
| APPID, BLOCK_RECORD, DIMSTYLE, LAYER, LTYPE, STYLE, UCS, VIEW, VPORT | `tables/*`, `entities/*` | `PARTIAL` — shared `EntryBase` (handle/owner/flags/xdata/extdict/reactors), generic lookup/add/remove/rename with duplicate-name errors, typed parse/format (`TestParseTablesEntryCounts`, `TestTableLookupAddRemoveRename`, `TestTableHandlesOwnersUnique`) |
| Dictionaries, XRECORD, DICTIONARYVAR, DICTIONARYWITHDEFAULT | `entities/dictionary.py` | `PARTIAL` — typed round trip, key lookup, and typed MLEADERSTYLE/MLINESTYLE collections reclassified from root dictionary keys (`TestDictionary*`, `TestMLeaderStyleCollectionMembers`, `TestMLineStyleCollectionFromRootDictionary`) |
| GROUP | `entities/dxfgroups.py` | `PARTIAL` — typed round trip (`TestGroup*`) |
| LAYOUT, PLOTSETTINGS | `entities/layout.py`, `dxfobjects/plotsettings.py` | `PARTIAL` — typed name/block record/flags, retained round trip; typed when loaded from ENTITIES (`TestLayoutObjectInEntitiesTyped`) |
| MLINESTYLE, MLEADERSTYLE, TABLESTYLE | `entities/mlinestyle.py`, `mleader.py`, `table.py` | `PARTIAL` — typed name/description/flags, retained round trip; MLEADERSTYLECOLLECTION/MLINESTYLECOLLECTION typed (`TestStandardObjectTypedModels`) |
| MATERIAL, VISUALSTYLE, SUN, LIGHT, GEODATA | `entities/material.py`, `visualstyle.py`, `sun.py`, `light.py`, `geodata.py` | `PARTIAL` — typed name/description/flags, retained round trip |
| SPATIAL_FILTER, IDBUFFER, RASTERVARIABLES, WIPEOUTVARIABLES | `entities/spatial_filter.py`, `idbuffer.py` | `PARTIAL` — typed models, retained round trip; RASTERVARIABLES constructor (`object.NewRasterVariables`) |
| SORTENTSTABLE, FIELDLIST, IMAGEDEF, IMAGEDEF_REACTOR, UNDERLAYDEFINITION, PDF/DWF/DGN DEFINITION | `entities/sortentstable.py`, `field.py`, `image.py`, `underlay.py` | `DONE` — typed sort order, entity handles, filename/size plus constructors and the Page accessor (`object.NewImageDef`, `object.NewUnderlayDef`, `TestImageDefPathAndSize`, `TestSortEntsTableOrder`, `TestUnderlayEntityRoundTrip`) |
| ACAD_PROXY_ENTITY, ACAD_XREC_ROUNDTRIP | `entities/acad_proxy_entity.py` | `DONE` — embedded entity records extract through `entity.ExtractProxyGraphics`; `entity.DecodeProxyPrimitives` decodes circle/arc/polyline/text/colour/transform opcodes, applies the matrix stack, and reports unknown opcodes with their byte offset while keeping earlier primitives (`TestDecodeProxyPrimitives`) |
| Dynamic block read helpers | `dynblkhelper.py` | `DONE` — `Drawing.ResolveDynamicBlock` resolves a direct dynamic INSERT or follows `AcDbBlockRepBTag` from an anonymous block to the named definition, returning GUID, true name, record handle, and XDATA metadata without modifying the drawing; ordinary blocks return no definition (`drawing/dynamic_block_test.go`) |
| User and binary XRECORD data | `entities/xrecord.py` | `DONE` — `entity.UserRecord` compiles and parses strings, integers, floats, points, nested lists, and dictionaries with explicit errors for newlines and overlong strings; `entity.BinaryRecord` hex-encodes arbitrary payloads in code-310 chunks; both preserve unrelated XRECORD tags via `ApplyTo` (`entity/user_record_test.go`) |
| Extension dictionary/reactors on every entity | `xdict.py`, `appdata.py` | `DONE` — typed on entities, table entries, and objects; unknown groups retained verbatim (`TestEntityExtensionDictionaryAndReactors`, `TestObjectExtensionDictionaryAndReactors`, `TestCollectionExtensionDictionary`) |
| LAYER_FILTER, VBA_PROJECT | `entities/layer.py`, `dxfobjects/vba.py` | `DONE` — typed layer handles/names/expressions and VBA payload with retained unknown tags (`TestLayerFilterTyped`, `TestVBAProjectTyped`) |

### 4. High-level document and layout API

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| `doc.modelspace()`, paper space layouts | `layouts/*` | `DONE` — `d.Model()`, `d.BlockSpace`, `d.SpaceByName`, layout create/rename/delete/activate with block+BLOCK_RECORD linkage (`TestFacadeLayoutLifecycle`, `TestFacadeLayoutSaveReload`) |
| `layout.add_*` factory methods for every entity | `graphicsfactory.py` | `DONE` — typed `Add<Type>` methods on every space for all supported entity types plus `AddGeneric` by DXF type (`TestFacadeAddEntityTypes`) |
| `dxfattribs` handling | `gfxattribs.py` | `DONE` — functional entity options (`WithLayer`, `WithColor`, `WithLinetype`, `WithLineweight`, `WithInvisible`, `WithLTScale`, `WithOwner`, `WithAttribs`) applied at add time, layers auto-created (`TestFacadeEntityOptions`) |
| Entity linking (INSERT/ATTRIB/SEQEND, POLYLINE/VERTEX) | `entitydb.py`, `entities/subentity.py` | `DONE` — loader links ATTRIB/SEQEND to INSERT, `GetAttribute` access, options attach attribs at insert time (`TestFacadeInsertAttribsReload`, `TestFacadeInsertAttribsExplode`) |
| Entity copy/transform/explode | `entities/copy.py`, `transform.py`, `explode.py` | `DONE` — `Copy()` deep clones (basis of transactions), per-entity matrix `Transform`, `Insert.Explode` (`TestFacadeInsertAttribsExplode`, `TestFacadeTransactions`) |
| Blocks: create, insert, explode, purge, rename | `blkrefs.py`, `entities/block.py` | `PARTIAL` — create/insert/explode and layout rename (block+BLOCK_RECORD); purge `ABSENT` (`TestFacadeBlockRoundTrip`) |
| Header variable API | `document.py` | `PARTIAL` — `d.Header()` accessor with typed `Variable(name)`, `RecalcExtents` writes EXTMIN/EXTMAX (`TestFacadeSelectionAndExtents`) |
| Query/select API | `query.py`, `select.py` | `DONE` — `SelectByType`/`SelectByLayer`/`SelectInExtents` over model space, `Extents`/`LayoutExtents` (`TestFacadeSelectionAndExtents`) |
| Transaction/undo model | `entitydb.py` | `DONE` — `BeginTransaction`/`Commit`/`Rollback` with snapshot restore of entities and layout ownership (`TestFacadeTransactions`) |

### 5. Math and geometry

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Vec2/Vec3, Matrix44, transformations | `math/_vector.py`, `_matrix44.py`, `transformtools.py` | `DONE` — zero-safe normalization, compose/invert/decompose with singular-matrix errors, WCS/UCS/OCS point/vector/direction round trips (`TestZeroVectorNormalizationIsSafe`, `TestMatrixDecompositionRoundTrip`, `TestSingularMatrixErrors`, `TestUCSPointVectorDirectionRoundTrips`) |
| Bounding boxes (2D/3D, extents) | `math/bbox.py`, `bbox.py` | `DONE` — union, intersection, containment, size/volume, 2D/3D transforms (`TestBoundingBox2DTransform`, `TestBBox3DOperations`, `TestBBox3DContainsPoint3D`, `TestBBox3DCenter`) |
| Bezier (3P/4P, derivatives, flattening) | `math/_bezier3p.py`, `_bezier4p.py` | `DONE` — evaluation, derivatives, split, bbox, flattening, arc length (`TestBezier`, `TestBezierTangent`, `TestBezierControlPoints`) |
| B-spline/NURBS (knots, derivatives, interpolation) | `math/_bspline.py`, `bspline.py`, `cspline.py` | `DONE` — clamped/periodic forms, knot insertion, derivatives, rational circle exactness, Catmull-Rom/BSpline interpolation (`TestClampedBSplineForm`, `TestPeriodicBSplineForm`, `TestInsertKnotPreservesShape`, `TestRationalCircleIsExact`, `TestRationalCircleArcIsExact`, `TestInterpolateBSplinePassesThroughPoints`) |
| Curve fitting | `math/fit.py`, `cspline.py` | `DONE` — polynomial fit, least-squares circle, interpolation in `math/fit` (`TestFitPolynomialLine`, `TestFitPolynomialParabola`, `TestFitCircleLSQOnSampledCircle`) |
| Bulge/arc helpers | `math/bulge.py`, `arc.py` | `DONE` — bidirectional bulge/arc/segment conversion; center sign matches ezdxf (positive bulge puts center left of chord travel) (`TestBulgeArcRoundTrip`, `TestBulgeCenterLocation`, `TestBulgeSegmentApproximate`) |
| OCS/UCS | `math/ucs.py`, `concepts/ocs.rst` | `DONE` — arbitrary-axis rule and angle conversions verified against ezdxf 1.4.6 probes (`TestOCSAxisMatchesEzdxfProbe`, `TestOCSAngleToWCSMatchesEzdxfProbe`, `TestUCSFromEulerAnglesIsOrthogonalAndRightHanded`) |
| Offset 2D | `math/offset2d.py` | `DONE` — polyline offset with round joins plus Bezier/spline offsets (`TestOffsetPolylineParallelAndJoins`, `TestOffsetBezierAndSplineParallel`) |
| Triangulation (earcut, Delaunay) | `math/_mapbox_earcut.py`, `triangulation.py` | `DONE` — ear-clip with holes, deterministic hole-bridge candidate search, post-triangulation area verification, Delaunay, degenerate-input handling, diagnostics (`TestTriangulatePolygonWithHole`, `TestTriangulateCenteredHoleNoExteriorBay`, `TestTriangulateDisjointHolesAreaParity`, `TestTriangulateDegenerateInput`) |
| Clipping (polygon, line, portal) | `math/clipping.py`, `tools/clipping_portal.py` | `DONE` — polygon clipping to rect/convex windows; portal clipping for line segments and quadratic/cubic Bezier curves returning visible sub-segments (`TestClipPolygonToRect`, `TestClipPolygonToConvex`, `TestClipSegmentOutsideToInside`, `TestClipQuadraticToPortal`, `TestClipCubicToPortal`) |
| Shape/construct2d/3d | `math/shape.py`, `construct2d.py`, `construct3d.py` | `DONE` — construct2d/construct3d helpers incl. 3-point circles, line intersections, point-line distances (`TestConstruct2DCircleFrom3Points`, `TestConstruct3DLineLineIntersection`, `TestConstruct3DTriangleArea`); `Shape` container stores an outer boundary and holes, preserves orientation, computes area/containment/centroid/bbox, and normalizes winding (`TestShapeAreaWithHole`, `TestShapeContainment`, `TestShapeNormalizedOrientation`) |
| R-tree | `math/rtree.py` | `DONE`-ish (`spatial`) |
| Perlin noise, clustering, parametrization | `math/perlin.py`, `clustering.py`, `parametrize.py` | `DONE` — seeded deterministic Perlin 2D/3D with octave variants (`TestPerlinDeterministic2D`, `TestPerlinRange2D`), point clustering with membership and centroids (`TestClusterPointsTwoGroups`, `TestClusterMembershipPartition`), arc-length parametrization (`TestParameterizeMonotonic`) |
| Intersections, curve tools, legacy math | `math/curvetools.py`, `legacy.py` | `DONE` — typed line/segment/Bezier/ray-plane results with parallel, coincident, and tangent handling (`TestIntersectSegmentSegment2D`, `TestIntersectLineBezier2D`, `TestIntersectBezierBezier2D`, `TestIntersectRayPlane`, `TestIntersectLineLine3D`) |

### 6. Path, rendering, visualization

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Path commands and converter for all entities | `path/*` | `DONE` — converters for line/arc/circle/ellipse/spline/polylines (bulge-aware), HATCH, SOLID/3DFACE/TRACE, IMAGE/WIPEOUT frames, LEADER/MLEADER/MLINE, POINT, DIMENSION (block content), RAY/XLINE, TEXT/MTEXT outlines; arcs flattened to line segments (`TestRenderConformance`, `path/converter_ext_test.go`) |
| Path tools: flatten, transform, offset, nesting | `path/tools.py`, `path/nesting.py` | `DONE` — flatten + transform (now transforms commands, not only vertices, `TestPathTransformCommands`); offset implemented (`Tools.Offset`); nesting classifies outer boundaries and holes into a normalized tree with CCW outer / CW holes (`BuildNestingTree`, `ClassifyNestedPaths`; `TestNestingConcentricRings`, `TestNestingDisjointRings`, `TestNestingIslandInsideHole`) |
| Drawing add-on / render pipeline | `addons/drawing/*`, `render/*` | `DONE` — `render.BuildScene` resolves styles (ByLayer/ByBlock, ACI/truecolor/alpha, lineweight, linetype scale), fits model extents (5% margin) or header extents, projects an active VIEWPORT, supports explicit extents and returns a fallback report (`TestViewportScene`, `TestSceneExtentsOverride`) |
| SVG/PDF backends | `addons/drawing/backend.py`, `pymupdf.py` | `DONE` — unified `Renderer` interface (Begin/SetStyle/DrawPath/Clip/FillGradient/End) with registry; compact SVG path data, PDF content streams with xref tables (`TestSVGPathGeometryAndStyle`, `TestPDFStructure`) |
| matplotlib/PyQt/PyMuPDF backends | `addons/drawing/*` | `N/A` in Go; provided native raster backend instead: PNG with supersampled anti-aliasing, background, clipping, dash patterns (`TestPNGRenderedViaRegistry`, `TestRasterStyleColorWidthDash`) |
| Linetype rendering | `render/_linetypes.py`, `linetypes.py` | `DONE` — dash patterns baked from layer linetype x linetype scale, device-scaled, stroked by all backends (`TestRasterStyleColorWidthDash`) |
| Hatch rendering | `render/hatching.py` | `DONE` — solid fill, predefined ANSI31–38 and ISO `ACAD_ISO02W100`–`ACAD_ISO15W100`, and custom `.pat` patterns expanded to clipped segments (`HatchFillSegments`); unknown named patterns keep the boundary stroke and report handle plus name in scene fallbacks; gradients: SVG linearGradient, PNG axis lerp, PDF first color (`TestSceneHatchFill`, `TestHatchPatternExpanded`, `TestISOHatchPatternExpanded`, `TestParsedPatternExpandsToClippedSegments`, `TestUnknownHatchPatternKeepsBoundaryAndReports`, `TestFillGradientBackends`) |
| Edge topology and boundaries | `explode.py`, `boundary.py` equivalents | `DONE` — `edges.BuildGraph` merges endpoints within tolerance, reports degree/leaves/components, extracts loops and chains, builds ordered LWPOLYLINE/POLYLINE and hatch boundaries, and preserves/reverses source orientation (`edges/edges_test.go`) |
| Dimension rendering | `render/dimension.py`, `dim_*.py` | `DONE` — DIMENSION entities render referenced dimension block content through the path converter, falling back to exploded geometry with measured text; arrowheads use the complete arrow-style set (`TestDimensionRendersReferencedBlock`, `TestDimensionFallbackUsesMeasurementText`) |
| MTEXT rendering | `render/abstract_mtext_renderer.py`, `tools/text_layout.py` | `DONE` — inline codes stripped (`StripCodes`), greedy word wrap at measured font advances (`LayoutMText`/`MeasureAdvance`), paragraph pitch 1.6714x height, alignment offsets from measured advances, TrueType glyph outlines via system fonts with marching-squares contours and built-in vector font fallback (`TestTextToPathsTTF`, `TestLayoutMTextWrap`, `TestTextAlignmentUsesMeasuredAdvance`, `TestTextOutlines`) |
| Mesh rendering | `render/mesh.py`, `mesh_builder.py` | `DONE` — MESH and POLYFACE/POLYMESH faces emitted as closed paths (filled or wireframe); invalid/degenerate faces skipped (`TestMeshRendersClosedFaces`, `TestPolyfaceRendersClosedFaces`) |
| Arrow heads | `render/arrows.py` | `DONE` — closed/open, arch tick, oblique, box, box-filled, diamond, circle, dot and dot-small geometries with closed/filled flags (`TestArrowHeadGeometryStyles`) |
| MLeader/MLine rendering | `render/mleader.py`, `mline.py` | `DONE` — converted to path geometry (leader arrows via path, mline style lines) and rendered by all backends (converter tests in `path/converter_ext_test.go`, corpus conformance) |

Plot style tables are modeled by `addons/acadctb`: text STB and binary CTB/STB files parse into a shared style table with color, true color, lineweight and screening overrides; `render.SceneOptions` applies color-based CTB or named STB overrides during scene building. Print/plot scaling (PS/PDF paper layout) is out of scope. Text rendering uses system TrueType fonts (Noto/Liberation/DejaVu/Vera) via `golang.org/x/image/font/opentype`; when none is available the built-in single-stroke vector font is used and `StyledScene.Fallbacks` records the substitution. SHX fonts can be loaded for glyph metadata and text measurement, and TTF advances/ascent/descent are exposed by the metrics API.

### 7. Audit, recover, validation, repair

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Audit protocol and fixers | `audit.py` | `DONE` — pluggable check registry (handles, owners, tables, references, xdata, reactors, extension dictionaries, extents, entity invariant hooks), safe fixers with re-audit, bounded report with severity counts and JSON (`TestRegistryCodes`, `TestCleanDocumentZeroErrors`, `TestReferenceIntegrity`, `TestTableIntegrity`, `audit/parity_test.go`) |
| Recover damaged files | `recover.py` | `DONE` — `RecoverFile`/`RecoverReader` merge loader recovery (truncated sections, skipped malformed records) with a full audit pass; report attached to the document (`TestRecoverTruncatedFile`, `TestRecoverMalformedEntity`, `audit/recover.go`) |
| R12 strict checker | `r12strict.py` | `DONE` — strict R12 audit flags OBJECTS/CLASSES sections and opaque object sections against the version profile (`TestStrictR12RejectsObjects`); the sanitizing mode is `drawing/r12strict.go` |
| Tag/validator checks | `lldxf/validator.py`, `lldxf/repair.py` | `DONE` — pre-bind validation reports section terminators, required handles, numeric ranges, and duplicate handles with code/value/handle; recover mode inserts missing ENDBLK/ENDSEC and assigns fresh handles, then re-validates (`TestTagValidationReportsCodeValueHandle`, `TestRepairInsertsMissingBlockTerminator`) |
| Orphan/cycle repair | `audit.py` | `DONE` — missing/dangling owners and missing table entries fixed; duplicate handles repaired deterministically with references left untouched when ambiguous; INSERT cycles broken by removing the closing edge while retaining block definitions (`TestDuplicateHandleNamesBothObjects`, `TestBlockCycleRepair`) |
| Header/table/block/object consistency | `audit.py` | `DONE` — required default entries, duplicate/empty entry names, owner mismatches, block record/layout links, dictionary entries, header extents (`TestTableIntegrity`, `TestDictionaryEntryUnresolved`, `TestExtentsAndSplineHook`) |

### 8. Query, selection, transformation, decomposition

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Query language | `query.py`, `queryparser.py` | `DONE` — expression lexer plus function predicates (`bbox_intersects`, `bbox_within`, `near`, `has_xdata`, `xdata_value`, `attr`), compiled predicates with position-carrying errors, spatial pre-filtering (`query.ExecuteWithFilter`) and document-level `Drawing.QueryExpression`; tests in `query/parity_test.go` |
| `select` filters | `select.py` | `DONE` — `query.EntitySelector` filters plus handle-based `query.Selection` sets with union/intersection/difference and named sets on the drawing (`SetSelectionSet`); `TestQuerySelectionsAndTransformsE2E` |
| `groupby` | `groupby.py` | `DONE` — `query.GroupByAttribute`/`GroupHandles` over layer, type, color, linetype, text and `attr:TAG` INSERT attributes; grouping tests in `query/parity_test.go` |
| Transform entities/layouts | `transform.py` | `DONE` — matrix builders (`RotationMatrix`, `ScaleMatrix`, `NonUniformScaleMatrix`) plus `Selection.ApplyM` and `Drawing.TransformSelection`/`TransformBlock`/`TransformDrawing`; `TestTransformBlockAndDrawingE2E` |
| Explode INSERT/POLYLINE/MTEXT/etc. | `explode.py`, `addons/mtextsurrogate` | `DONE` — INSERT, POLYLINE, HATCH, LEADER, MLEADER and DIMENSION (extension lines, dimension line, arrow ticks, text); MText surrogate formatting is not modelled (MText explodes to itself); `TestExplodeDimension` and the dimension test in `query/parity_test.go` |
| Bounding box of entities/layouts | `bbox.py` | `DONE` — per-entity `BBox()` feeds `query.EntityContext`/`Selection.Bounds` and layout extents (`Drawing.Extents`, `LayoutExtents`); decoded proxy geometry contributes its real box, and geometry-less records report no extents rather than zero-origin bounds (`TestProxyBBoxFromDecodedGraphics`, `TestEmptySelectionBounds`) |
| Disassemble to primitives | `disassemble.py` | `DONE` — `query.Disassemble` flattens nested inserts through composed insert matrices, clears handles and reports cycles; `Drawing.DisassembleModel`; nested/cycle tests in `query/parity_test.go` |
| Primitive decomposition | `disassemble.py`, `explode.py` | `DONE` — `query.MakePrimitive` maps LINE/ARC/CIRCLE/ELLIPSE/SPLINE/POLYLINE/MESH/TEXT/MTEXT/SOLID/INSERT/POINT to kind, handle, vertices, optional path or mesh, and diagnostics; `query.RecursiveDecompose` expands nested INSERT and HATCH paths with depth/cycle guards; `ToVertices`/`ToPaths`/`ToMeshes`/`ToControlVertices` convert the stream (`query/primitive_test.go`) |
| Upright/rotate-to-view | `upright.py` | `DONE` — `entity.MakeUpright` for TEXT, MTEXT (attachment corner mirroring) and ATTRIB, `Drawing.MakeUprightAll`; `TestQuerySelectionsAndTransformsE2E` |
| Reorder/seqend repair | `reorder.py` | `DONE` — `Drawing.DrawOrder`/`SendToBack`/`BringToFront`/`SetDrawOrder` persisted as `SORTENTSTABLE` (`object.NewSortEntsTable`, applied on load by `Drawing.ApplySortEntsTables`); pre-R13 drawings fall back to physical entity order; `TestDrawOrderRoundTripE2E`, `TestDrawOrderR12Fallback`; SEQEND repair is not applicable (no SEQEND model) |
| Block reference traversal | `blkrefs.py` | `DONE` — `query.TraverseBlockRefs` depth-first insert walk with cycle reporting, exposed as `Drawing.TraverseBlockRefs`; traversal/cycle tests in `query/parity_test.go` |
| Zoom/extents fitting | view fitting (CLI `zoom`) | `DONE` — `drawing.FitExtentsToViewport`, `Drawing.FitView`/`FitLayoutView` (aspect ratio + margin, zoom factor); `TestQuerySelectionsAndTransformsE2E` |

### 9. XRef, images, underlays, clipping

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| XREF load/attach/overlay | `xref.py` | `DONE` — AttachXRef/DetachXRef/BindXRef/ReloadXRef, relative paths + search paths, missing-file reporting (`xref_parity_e2e_test.go`) |
| Image definitions and rendering | `entities/image.py` | `DONE` — IMAGEDEF/IMAGEDEF_REACTOR objects, full attribute load (fade/contrast/brightness/clipping/boundary), scene rasters with SVG/PNG rendering (`TestImageEntityRoundTrip`, `TestImageRenderingScene`, `TestRasterImageDraw`) |
| PDF/DGN/DWF underlays | `entities/underlay.py` | `DONE` — definition objects plus UNDERLAY entity with typed loader, fade/scale/rotation round trip (`TestUnderlayEntityRoundTrip`) |
| XCLIP boundaries | `xclip.py` | `DONE` — rectangular and polyline boundaries on INSERT, save/reload and SVG clipPath rendering (`TestXClipRoundTrip`) |
| Nested xref resolution | `xref.py` | `DONE` — recursive loading with canonical-path cycle detection (`TestNestedXRefsLoad`, `TestXRefCycleTerminates`) |

### 10. Add-ons

| Add-on | ezdxf reference | Go baseline |
|---|---|---|
| `drawing` | `addons/drawing` | `DONE` — native SVG/PDF/PNG backends via core `render`; plot style overrides applied during scene building |
| `r12writer` | `addons/r12writer.py` | `DONE` — `addons/r12writer` incremental R12 writer (header/tables + entities, no full document; `TestR12WriterStreaming` 10k entities) |
| `iterdxf` | `addons/iterdxf.py` | `DONE` — `addons/iterdxf` bounded-memory record iterator in file order (`TestIterDXFStreamingRecords`) |
| `importer` | `addons/importer.py` | `DONE` — `addons/importer.ImportSVG` (line/rect/circle/polyline/path; `TestImportSVGElements`) |
| `dxf2code` | `addons/dxf2code` | `DONE` — `addons/dxf2code.Generate` emits Go source (parsed with go/parser, inventory counts) |
| `acadctb` | `addons/acadctb.py` | `DONE` — text STB and binary CTB/STB parse into a shared plot-style model with color, true color, lineweight, screening and linetype overrides; `Apply` and render-scene integration (`TestParseBinaryCTB`, `TestParseBinarySTBNamedStyles`, `TestNamedPlotStyleOverridesRendering`) |
| `pycsg` | `addons/pycsg.py` | `DONE` — real triangle-level booleans + Box/Sphere primitives + `csg.Volume` (`csg/boolean_test.go`) |
| `MTextExplode` | `addons/mtextsurrogate` | `DONE` — `addons/mtxpl` Lines/Explode → positioned TEXT entities |
| `text2path` | `addons/text2path.py` | `DONE` — `addons/text2path` TTF/OTF outlines via system fonts plus built-in vector fallback; `Size`/`MTextSize` report measured width, height, ascent, descent, advance, wrapped line count and rotated bounds with an explicit fallback reason; SHX font loader with glyph metadata and text metrics; TTF metrics API (`text2path/shx.go`, `text2path/metrics.go`, `addons/text2path/size.go`, `addons/text2path/size_test.go`) |
| `geo` | `addons/geo.py` | `DONE` — `addons/geo.ExportGeoJSON` + `CRSFunc` transform |
| `meshex` | `addons/meshex.py` | `DONE` — `addons/meshex.ExportOBJ`/`ExportOpenSCAD` route direct meshes and decoded ACIS SAT bodies through `Geometry`; `ExportOBJWithLoops`/`ExportOpenSCADWithLoops` append closed linked-edge loops in source order (`addons/meshex/geometry_test.go`) |
| `openscad` | `addons/openscad.py` | `DONE` — `addons/meshex.ExportOpenSCAD` polyhedron source, with decoded ACIS and linked-edge loop geometry |
| `odafc` | `addons/odafc.py` | `DONE` — `addons/odafc` CommandRunner + actionable missing-tool error |
| `browser` | `addons/browser` | `DONE` — `addons/browser.Extract` exposes document, layout and entity inspection data (handle, type, layer, bbox) grouped by layout and type (`addons/browser/browser_test.go`) |
| `acisbrowser` | `addons/acisbrowser` | `DONE` — `addons/acisbrowser.Extract` exposes ACIS SAT/SAB payloads and metadata for REGION/BODY/SOLID3D without a GUI (`addons/acisbrowser/acisbrowser_test.go`) |
| `dwg` | `addons/dwg` | `DONE` — `addons/dwg.Converter` delegates DWG→DXF conversion to a configurable external `CommandRunner`, names the missing tool on failure and leaves no partial output (`addons/dwg/dwg_test.go`); native `addons/dwg.ReadMetadata`/`ReadMetadataFile` additionally read version, release, maintenance/codepage, and the R13–R2000 section locator without writing or converting (`addons/dwg/metadata_test.go`) |
| `r12export` | `addons/r12export.py` | `DONE` — streaming R12 exporter converts SPLINE/ELLIPSE/LWPOLYLINE to POLYLINE, MTEXT to TEXT, MESH to POLYFACE and reports conversions (`addons/r12export/r12export_test.go`) |
| `dimlines` | `addons/dimlines.py` | `DONE` — `addons/dimlines` builds linear and angular dimension lines from LINE/TEXT primitives with configurable style (`addons/dimlines/dimlines_test.go`) |
| `genetic_algorithm` | `addons/genetic_algorithm.py` | `DONE` — `addons/genetic` DNA types, mutation/crossover strategies, selections and maximizing optimizer with deterministic RNG support (`addons/genetic/genetic_test.go`) |
| `menger_sponge`, `sierpinski_pyramid` | `addons/menger_sponge.py`, `addons/sierpinski_pyramid.py` | `DONE` — `addons/fractals` generates Menger sponge variants and 3/4-sided Sierpinski pyramids, with cube/pyramid and face counts, merged or per-cell MESH output and drawing helpers (`addons/fractals/fractals_test.go`) |
| `mixins` | `addons/mixins.py` | `DONE` — `addons/mixins.SubscriptAttributes` exposes Has/Get/Set semantics for existing keys, typed access, cloning, and leaves entity data unchanged (`addons/mixins/mixins_test.go`) |
| `hpgl2` | `addons/hpgl2` | `DONE` — `addons/hpgl2.ExportHPGL2`/`PlotWithRunner` |
| `xplayer` | `addons/xplayer.py` | `DONE` — `addons/xplayer.ExportAnimationSVG` per-frame SVG |
| `binpacking`, `tablepainter`, `gerber_D6673`, `mtxpl` | `addons/*` | `DONE` — `addons/binpacking.PackRectangles` (shelf pack), `addons/tablepainter.PaintTable`, `addons/gerber.ExportGerber` (RS-274X), `addons/mtxpl` |

### 11. CLI and tooling

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| `ezdxf` CLI (`info`, `audit`, `draw`, `browse`, `convert`, `strip`, `view`, `pp`) | `commands.py`, `__main__.py` | `DONE` — `cmd/dxf` provides info/analyze/difftags/debug/audit/convert/draw/query/explode/fix/strip/header/layers/browse/standards/zip/unzip/version/completion, text+JSON, exit codes 0/1/2/3 (`cmd/dxf/main_test.go`, `cmd/dxf/devtools_test.go`, `cmd/dxf/zip_test.go`, `docs/cli.md`) |
| App settings / config | `appsettings.py`, `_options.py` | `DONE` — `appsettings` with defaults < file < `DXF_*` env precedence |
| Comments | `comments.py` | `DONE` — preserve per section and strip via `comments.StripComments` |
| Native DWG metadata | LibreDWG R13–R2018 header equivalents | `DONE` — read-only version/release detection, maintenance and codepage fields, R13–R2000 section locator names/addresses/sizes, and diagnostics for unsupported sections; R2004+ reports the encrypted/compressed section map as unsupported instead of guessing. The external ODA `Converter` is unchanged, and native parsing never writes DWG output (`addons/dwg/metadata_test.go`) |
| Raw section loading | `lldxf/loader.py` equivalents | `DONE` — `lldxf.ParseRawSections` tags ASCII or binary DXF into ordered, named section bodies without binding; truncated input returns the tags read so far and an error naming the open section (`lldxf/raw_sections_test.go`) |
| Julian date conversion | `tools/julian.py` equivalents | `DONE` — `math.JulianDate`, `TimeFromJulian`, `CivilToJulian`, and `JulianToCivil`; J2000 is `2451545.0` at 2000-01-01T12:00Z and round trips within one second (`math/julian_test.go`) |
| Fonts, SHX/TTF, text metrics | `fonts/*`, `tools/text*.py` | `DONE` — TTF/OTF outlines via system fonts; SHX shape/unifont loader with glyph metrics (`text2path/shx.go`); TTF advance/ascent/descent metrics and TEXT/MTEXT measured size APIs (`text2path/metrics.go`, `addons/text2path/size.go`, `TestParseSHXShapesAndMeasure`, `TestMeasureTTF`, `TestTextSizeReportsMeasuredExtents`) |
| ISO hatch patterns / standards | `tools/_iso_pattern.py`, `standards.py` | `DONE` — ANSI31–38 and ISO `ACAD_ISO02W100`–`ACAD_ISO15W100` patterns with category lookup and render expansion (`TestISOPatterns`, `TestISOHatchPatternExpanded`); standards rule engine and `dxf standards` implemented (`standards` package, `standards_test.go`) |
| Revcloud | `revcloud.py` | `DONE` — `query.RevCloudRectangle`/`RevCloudArcs` (closed bulged LWPOLYLINE, configurable arc chord and bulge) plus `Space.AddRevCloud`/`Drawing.AddRevCloud`; revcloud tests in `query/parity_test.go` and `TestQuerySelectionsAndTransformsE2E` |
| Zoom/extents helpers | `zoom.py` | `DONE` — `drawing.FitExtentsToViewport` and `Drawing.FitView`/`FitLayoutView`; `TestQuerySelectionsAndTransformsE2E` |
| Analyze/difftags/debug tools | `tools/analyze.py`, `difftags.py` | `DONE` — `analyze` tag/entity/table/block/object statistics, structural `difftags` with changed/added/removed tag paths, `debug` loader/recovery diagnostics; text+JSON (`cmd/dxf/analyze.go`, `cmd/dxf/difftags.go`, `cmd/dxf/debug.go`, `cmd/dxf/devtools_test.go`) |

### 12. Colors, enums, units, standards

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| ACI colors, true color, transparency | `colors.py`, `concepts/true_color.rst` | `DONE` — tagged color value (ACI/true color/BYLAYER/BYBLOCK) with full ACI table, nearest-index lookup and validation (`color/colorvalue.go`), alpha transparency (`color/transparency.go`), entity round trips (`color/roundtrip_test.go`, `color/colorvalue_test.go`) |
| Enums (text alignment, hatch, etc.) | `enums.py` | `DONE` — typed enum sets with name/numeric conversion and unknown-value reporting (`enums/enums.go`, `enums/enums_test.go`) |
| Units and insunits | `units.py`, `concepts/units.rst` | `DONE` — unit definitions 0-20 with meter factors, `$INSUNITS` mapping and length conversion (`units/units.go`, `units/units_test.go`) |
| Standards checks | `tools/standards.py` | `DONE` — reusable rule engine with structured violations (code/severity/path/handle), configurable severities and disable list, layer-name rule, and `dxf standards` CLI (`standards/standards.go`, `standards/layer.go`, `standards/standards_test.go`, `cmd/dxf/standards.go`) |
| Lineweights | `concepts/lineweights.rst` | `DONE` — lineweight default (-3 ByLayer) and validation (`enums/enums.go` `DefaultLineWeight`/`ValidateLineWeight`) |
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

### Follow-up parity changes

These changes cover remaining gaps found after the first thirteen changes were
implemented. `entity-object-completion-parity`, `math-shape-clipping-parity`,
`addons-remaining-parity`, `dev-tools-parity`, and `library-refactor` are
complete; changes 16 and 19 are not yet applied.

| # | Change name | Scope | Depends on |
|---|---|---|---|
| 14 | `entity-object-completion-parity` (done) | ACAD_TABLE typed content, collection objects, POLYFACE/POLYMESH vertices, MESHVERTEXCACHE, ACIS SAT/SAB metadata, objects-as-entities | 3, 4 |
| 15 | `math-shape-clipping-parity` (done) | Perlin noise, clustering, Shape container, portal clipping, triangulation hole bridge, path nesting | 6, 7 |
| 16 | `render-plot-font-parity` (done) | full DIMENSION renderer, mesh/polyface faces, arrow heads, STB/binary CTB plot styles, SHX/TTF text metrics, ISO hatch patterns | 7, 11 |
| 17 | `addons-remaining-parity` (done) | browser/acisbrowser extraction, DWG via CommandRunner, r12export, dimlines, genetic algorithm, Menger sponge, Sierpinski pyramid, mixins | 11 |
| 18 | `dev-tools-parity` (done) | `dxf analyze`, `difftags`, `debug`, `standards`, ZIP manager and CLI zip/unzip | 12 |
| 19 | `library-refactor` | dead code, duplicate adapters, named interfaces, SOLID/DRY/YAGNI review | 1–13 |

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
3. **Cross-check with ezdxf.** Generate expected dumps with the Python library
   in a throwaway venv when semantics are unclear. Never add a runtime Python
   dependency to the Go library. The executable harness is
   `ezdxf_crosscheck_test.go` + `scripts/ezdxf_dump.py`:

   ```sh
   DXF_EZDXF_CROSSCHECK=1 \
   DXF_EZDXF_PYTHON=/path/to/venv/bin/python \
   go test -run TestEzdxfCrossCheck .
   ```

   It compares version, entity/block/subentity inventory and a stable typed-field
   projection (entity geometry/text, table flags, dictionary/layout/group
   fields) for every `testdata/*.dxf`, in addition to source TABLES entries and
   typed OBJECTS records. Object records expose both the source wire type and
   the typed model, including collection member handles; `MESHVERTEXCACHE`
   records are read from the source stream because
   ezdxf does not expose them as graphics. Table/object expectations come from
   raw source records so ezdxf auto-created defaults (layers, dictionaries,
   layouts) do not create false diffs. Set
   `DXF_EZDXF_SOURCE=/home/egor/src/ezdxf` to use a local source checkout; the
   harness creates a temporary venv with `uv` or `python3 -m venv` when
   `DXF_EZDXF_PYTHON` is unset.
4. **Behavioural tests.** Port ezdxf tests where practical, especially
   `tests/test_*.py` for math, path, query, audit, transform, xref, rendering.
5. **Fuzzing.** Fuzz the tagger/loader/parser with truncated and malformed files;
   recover mode must not panic. Dedicated fuzz targets cover native DWG metadata
   (`FuzzReadMetadata`), PAT parsing (`FuzzParsePAT`) and proxy primitive
   decoding (`FuzzDecodeProxyPrimitives`), in addition to the tagger/loader and
   audit targets. `tooling_fixture_test.go` exercises decomposition, measured
   text size, custom patterns, edge loops and binding-free raw section loading
   on `testdata/tooling_fixture.dxf`.
6. **Performance smoke tests.** `iterdxf`-style streaming must keep memory bounded.

## Known WIP and hazards

- `drawing.FromJSON` is a no-op.
- No known math gaps remain from changes 5 and 15.
- Rendering gaps closed by change 16: DIMENSION block content with measured
  text, mesh/polyface face rendering, all arrow heads, CTB/STB plot-style
  overrides, SHX loading/metrics, and ISO hatch patterns.
- Audit remaining limitation: ambiguous references to a repaired duplicate
  handle are left unchanged (the target is ambiguous by construction); the
  original duplicate issue and any unresolved owner/reactor issues stay in the
  report for manual follow-up.
- Tooling scope completed by `tooling-completeness-parity`: primitive
  decomposition, dynamic-block read helpers, user/binary XRECORD data, proxy
  primitive decoding, measured TEXT/MTEXT sizing, `.pat` parsing, linked-edge
  topology, read-only native DWG metadata, Julian conversion, raw section
  loading, and ACIS SAT mesh export. Remaining limitations: ACIS SAT decoding
  handles flat polygonal faces only (SAB payloads and curved surfaces are
  retained but not triangulated); DWG R2004+ section maps are diagnostic-only
  and conversion still requires the external ODA converter; proxy decoding
  covers the documented opcode subset and reports unknown opcodes with offsets;
  `.pat` parsing supports the documented simple pattern-line subset.
- Add-on/tool gaps: `xqt` is intentionally not ported because the Go package
  has no Qt dependency.
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
| 2026-11-10 | change 4 `table-object-model-parity` | done | All nine tables typed with lookup/add/remove/rename and duplicate-name errors; dictionaries, groups, xrecords, layouts, plot settings, styles, image/underlay definitions, and remaining standard objects typed or retained; extension dictionaries/reactors preserved. |
| 2026-11-10 | change 5 `geometry-math-parity` | done | Vectors/matrices, bbox, Bezier/BSpline/NURBS, bulge, OCS/UCS, offset, earcut/Delaunay, intersections, construct2d/3d. Initially documented gaps: Perlin/clustering, Shape, portal clipping, triangulation hole bridge; closed by change 15. |
| 2026-11-10 | change 6 `graphics-factory-api-parity` | done | `Add<Type>` factory methods for every supported entity, block/insert/attrib linking, layouts, groups, selection/extents, transactions with rollback. |
| 2026-11-10 | change 7 `path-render-parity` | done | Path converters for all supported entities; scene style resolution; SVG/PDF/raster backends; linetype/hatch/MTEXT/MLeader/MLine rendering. Dimension/mesh/arrow rendering remains partial. |
| 2026-11-10 | change 8 `audit-recover-parity` | done | Pluggable audit checks and fixers, recovery mode, R12 strict checker, table/reference/dictionary/extents consistency. Duplicate-handle and cycle repair remain partial. |
| 2026-11-10 | change 9 `query-transform-parity` | done | Query expressions, selections, grouping, transforms, explode/disassemble, upright text, draw order, block traversal, zoom/extents, revcloud. |
| 2026-11-10 | change 10 `xref-image-underlay-parity` | done | XREF attach/detach/bind/reload, nested loading with cycle detection, images, PDF/DGN/DWF underlays, XCLIP. |
| 2026-11-10 | change 11 `addons-parity` | done | r12writer, iterdxf, SVG importer, dxf2code, STB plot styles, CSG, MText explode, text2path, geo, mesh/OpenSCAD, ODA, HPGL2, xplayer, binpacking, tablepainter, gerber. Binary CTB and browser/acisbrowser/dwg/dimlines/r12export/genetic/menger/sierpinski/mixins remain absent (closed by change 17). `xqt` is intentionally not ported because the Go package has no Qt dependency. |
| 2026-11-10 | change 12 `cli-tooling-parity` | done | `cmd/dxf` CLI with info/audit/convert/draw/query/explode/fix/strip/header/layers/browse/version/completion, text+JSON output, exit codes, config precedence, atomic writes/backups, shell completion, integration tests. `docs/cli.md` added. |
| 2026-11-10 | change 13 `color-enum-units-parity` | done | Full ACI table/nearest lookup/transparency, typed enums, units/$INSUNITS conversion, lineweight validation. Linetype patterns and ISO hatch patterns remain partial. |
| 2026-11-10 | ezdxf cross-check harness | done | `scripts/ezdxf_dump.py` + opt-in `TestEzdxfCrossCheck` compare version, entity/block/subentity inventory, source tables, and source objects across all 52 fixtures. Passing against ezdxf 1.4.6b0; `go test ./...` passes. |
| 2026-11-10 | `library-refactor` | done | Removed dead `parser.go` and unused render helpers; consolidated block adapters in `internal/blockadapter`; replaced inline interface assertions with named interfaces across the public packages; `go test ./...` and the ezdxf cross-check still pass. |
| 2026-11-10 | change 14 `entity-object-completion-parity` | done | ACAD_TABLE typed rows/cells/styles/merges with typed loader and round trips; POLYFACE/POLYMESH vertex variants and face records; MESHVERTEXCACHE typed and linked to MESH; SAT/SAB metadata parsing with byte-preserving payloads and ACDSDATA linkage; typed MLEADERSTYLE/MLINESTYLE collections, LAYER_FILTER, VBA_PROJECT; objects-as-entities loaded through typed object models without moving sections; typed reactors/extension dictionaries on entities and objects. Cross-check now compares typed object inventories and passes across 54 fixtures. Tests: `acad_table_roundtrip_test.go`, `polyface_mesh_roundtrip_test.go`, `acis_metadata_test.go`, `object_collection_roundtrip_test.go`, `object_entity_roundtrip_test.go`, `extension_links_test.go`, plus package unit tests. |
| 2026-11-10 | change 15 `math-shape-clipping-parity` | done | Seeded deterministic Perlin 2D/3D with octave variants; point clustering with membership and centroids; `Shape` container with holes, area/containment/centroid/bbox and winding normalization; portal clipping for line segments and quadratic/cubic Bezier curves; earcut hole bridging via deterministic visible-vertex candidate search plus post-triangulation area verification (centered hole now has no exterior bay and exact area); path nesting tree with normalized CCW outer / CW holes. Tests: `math/perlin_test.go`, `math/clustering_test.go`, `math/shape_test.go`, `math/clipping_test.go`, `math/parity_ext_test.go`, `path/nesting_test.go`. |
| 2026-11-10 | change 17 `addons-remaining-parity` | done | GUI-free browser and ACIS browser extraction; external DWG converter with actionable missing-tool errors and atomic output; streaming R12 exporter converting splines/ellipses/LWPolylines/MTEXT/meshes; linear and angular dimension-line generators; genetic algorithm engine; Menger sponge and Sierpinski pyramid generators; `SubscriptAttributes` mixin. Tests: `addons/browser/browser_test.go`, `addons/acisbrowser/acisbrowser_test.go`, `addons/dwg/dwg_test.go`, `addons/r12export/r12export_test.go`, `addons/dimlines/dimlines_test.go`, `addons/genetic/genetic_test.go`, `addons/fractals/fractals_test.go`, `addons/mixins/mixins_test.go`. |
| 2026-11-10 | change 18 `dev-tools-parity` | done | `dxf analyze` tag/entity/table/block/object statistics; `dxf difftags` structural added/removed/changed tag paths with non-zero exit on differences; `dxf debug` version/encoding/section/loader/recovery diagnostics; reusable `standards` rule engine with structured violations, configurable severities and `dxf standards`; ZIP manager with atomic creation/extraction, per-entry path traversal and symlink rejection, plus `dxf zip`/`dxf unzip`. Tests: `cmd/dxf/devtools_test.go`, `cmd/dxf/zip_test.go`, `standards/standards_test.go`, `filemanagement/zip_test.go`. |
| 2026-11-11 | change 16 `render-plot-font-parity` | done | DIMENSION entities render referenced dimension blocks fully with measured text fallback; MESH and POLYFACE/POLYMESH faces emit closed filled or wireframe paths; all arrow-head styles (closed/open, arch tick, oblique, box, diamond, circle, dot) implemented; shared STB/binary CTB plot-style model with color, true color, lineweight and screening overrides applied during scene building; SHX shape/unifont loader with glyph metrics and TTF advance/ascent/descent metrics API; ISO `ACAD_ISO02W100`–`ACAD_ISO15W100` hatch patterns with lookup and render expansion. Tests: `render/entities_parity_test.go`, `render/render_test.go`, `addons/acadctb/binary_test.go`, `addons/acadctb/stb_test.go`, `text2path/shx_test.go`, `text2path/metrics_test.go`, `patterns/patterns_test.go`. |
| 2026-11-11 | change `core-fidelity-parity` | done | Ordered per-record retention for entities/tables/blocks/objects; positioned comments; per-type entity/table/object conformance matrices and golden round trips; deep-copy isolation for base tags/XDATA, ACIS bodies/regions, and MPolygon state; decoded proxy bounding boxes and empty-selection bounds; pre-bind tag validation and recover repair (terminators, handles, ranges); deterministic duplicate-handle and block-cycle repair. Remaining limitation: ambiguous references to a repaired duplicate handle are left unresolved and reported. Tests: `retention_comments_test.go`, `conformance_test.go`, `table_object_conformance_test.go`, `entity/copy_isolation_test.go`, `entity/proxy_bbox_test.go`, `query/empty_bounds_test.go`, `tag_validation_test.go`, `audit/block_cycle_fix_test.go`. |
| 2026-11-11 | change `tooling-completeness-parity` | done | `query.MakePrimitive`/`RecursiveDecompose` with path, mesh and diagnostic conversions; dynamic-block resolution; `entity.UserRecord`/`BinaryRecord`; proxy graphics primitive opcodes; measured TEXT/MTEXT `Size`/`MTextSize` and anchor layout; PAT parser/registry/scaling/measurement conversion and clipped hatch expansion; immutable `edges` graph with loops/chains/entity boundaries; native DWG version/section metadata; Julian conversion; raw section loading; ACIS SAT mesh decoding and meshex geometry/loop export. Fuzz targets cover DWG metadata, PAT and proxy decoding; `testdata/tooling_fixture.dxf` exercises the combined scope. Tests: `query/primitive_test.go`, `drawing/dynamic_block_test.go`, `entity/user_record_test.go`, `entity/proxy_primitives_test.go`, `addons/text2path/size_test.go`, `patterns/pat_test.go`, `edges/edges_test.go`, `addons/dwg/metadata_test.go`, `math/julian_test.go`, `lldxf/raw_sections_test.go`, `entity/acis_mesh_test.go`, `addons/meshex/geometry_test.go`, `tooling_fixture_test.go`. |

Append a row whenever a change moves state. Keep the matrix above in sync.
