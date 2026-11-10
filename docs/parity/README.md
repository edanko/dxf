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
| `drawing` | `addons/drawing` | `DONE` — `addons/drawing.RenderSVG`/`Render` over the core `render` backends (`TestRenderSVG`) |
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
| `text2path` | `addons/text2path.py` | `DONE` — TTF glyph outlines (`text2path.TextToPathsTTF`) + `addons/text2path.ConvertEntity` |
| `query` | entity query and parser | useful selector/expression subset |
| `audit` | auditor, recovery, block cycles | partial; some placeholder loading |
| `spatial` | R-tree and entity spatial index | usable |
| `layouts` | model/paper layouts and viewports | partial |
| `xref` | xref manager and loader | `DONE` — attach/detach/bind/reload, relative paths + search paths, nested loading with cycle detection, BLOCK_RECORD xref flags (`xref_parity_e2e_test.go`) |
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
| Recover mode | `recover.py` | `DONE` — missing ENDSEC/EOF and tagger errors recovered with a report; malformed entity records are skipped and counted; `RecoverFile`/`RecoverReader` add a merged audit pass (`TestRecoveryLoadsTruncatedDocument`, `TestRecoverMalformedEntity`, `audit/recover.go`) |
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
| SPATIAL_FILTER, IDBUFFER, RASTERVARIABLES, WIPEOUTVARIABLES | `entities/spatial_filter.py`, `idbuffer.py` | `PARTIAL` — typed models, retained round trip; RASTERVARIABLES constructor (`object.NewRasterVariables`) |
| SORTENTSTABLE, FIELDLIST, IMAGEDEF, IMAGEDEF_REACTOR, UNDERLAYDEFINITION, PDF/DWF/DGN DEFINITION | `entities/sortentstable.py`, `field.py`, `image.py`, `underlay.py` | `DONE` — typed sort order, entity handles, filename/size plus constructors and the Page accessor (`object.NewImageDef`, `object.NewUnderlayDef`, `TestImageDefPathAndSize`, `TestSortEntsTableOrder`, `TestUnderlayEntityRoundTrip`) |
| ACAD_PROXY_ENTITY, ACAD_XREC_ROUNDTRIP | `entities/acad_proxy_entity.py` | `PARTIAL` |
| Extension dictionary/reactors on every entity | `xdict.py`, `appdata.py` | `PARTIAL` — preserved on typed table entries and objects (`TestObjectReactorsAndExtDict`) |

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
| Triangulation (earcut, Delaunay) | `math/_mapbox_earcut.py`, `triangulation.py` | `PARTIAL` — ear-clip with holes, Delaunay, degenerate-input handling, diagnostics (`TestTriangulatePolygonWithHole`, `TestTriangulateDegenerateInput`). Known gap: a hole floating centered in a polygon merges through a diagonal needle bridge that winds a thin bay outside the clipped ring, leaving an area shortfall that `DiagnoseTriangulation` reports as an area mismatch |
| Clipping (polygon, line, portal) | `math/clipping.py`, `tools/clipping_portal.py` | `PARTIAL` — polygon clipping to rect/convex windows (`TestClipPolygonToRect`, `TestClipPolygonToConvex`); portal clipping `ABSENT` |
| Shape/construct2d/3d | `math/shape.py`, `construct2d.py`, `construct3d.py` | `PARTIAL` — construct2d/construct3d helpers incl. 3-point circles, line intersections, point-line distances (`TestConstruct2DCircleFrom3Points`, `TestConstruct3DLineLineIntersection`, `TestConstruct3DTriangleArea`); `Shape` container `ABSENT` |
| R-tree | `math/rtree.py` | `DONE`-ish (`spatial`) |
| Perlin noise, clustering, parametrization | `math/perlin.py`, `clustering.py`, `parametrize.py` | `PARTIAL` — arc-length parametrization `DONE` (`TestParameterizeMonotonic`); Perlin/clustering `ABSENT` |
| Intersections, curve tools, legacy math | `math/curvetools.py`, `legacy.py` | `DONE` — typed line/segment/Bezier/ray-plane results with parallel, coincident, and tangent handling (`TestIntersectSegmentSegment2D`, `TestIntersectLineBezier2D`, `TestIntersectBezierBezier2D`, `TestIntersectRayPlane`, `TestIntersectLineLine3D`) |

### 6. Path, rendering, visualization

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Path commands and converter for all entities | `path/*` | `DONE` — converters for line/arc/circle/ellipse/spline/polylines (bulge-aware), HATCH, SOLID/3DFACE/TRACE, IMAGE/WIPEOUT frames, LEADER/MLEADER/MLINE, POINT, DIMENSION (block content), RAY/XLINE, TEXT/MTEXT outlines; arcs flattened to line segments (`TestRenderConformance`, `path/converter_ext_test.go`) |
| Path tools: flatten, transform, offset, nesting | `path/tools.py`, `path/nesting.py` | `PARTIAL` — flatten + transform (now transforms commands, not only vertices, `TestPathTransformCommands`); offset/nesting `ABSENT` |
| Drawing add-on / render pipeline | `addons/drawing/*`, `render/*` | `DONE` — `render.BuildScene` resolves styles (ByLayer/ByBlock, ACI/truecolor/alpha, lineweight, linetype scale), fits model extents (5% margin) or header extents, projects an active VIEWPORT, supports explicit extents and returns a fallback report (`TestViewportScene`, `TestSceneExtentsOverride`) |
| SVG/PDF backends | `addons/drawing/backend.py`, `pymupdf.py` | `DONE` — unified `Renderer` interface (Begin/SetStyle/DrawPath/Clip/FillGradient/End) with registry; compact SVG path data, PDF content streams with xref tables (`TestSVGPathGeometryAndStyle`, `TestPDFStructure`) |
| matplotlib/PyQt/PyMuPDF backends | `addons/drawing/*` | `N/A` in Go; provided native raster backend instead: PNG with supersampled anti-aliasing, background, clipping, dash patterns (`TestPNGRenderedViaRegistry`, `TestRasterStyleColorWidthDash`) |
| Linetype rendering | `render/_linetypes.py`, `linetypes.py` | `DONE` — dash patterns baked from layer linetype x linetype scale, device-scaled, stroked by all backends (`TestRasterStyleColorWidthDash`) |
| Hatch rendering | `render/hatching.py` | `DONE` — solid fill, predefined ANSI31–38 patterns expanded to clipped segments (`HatchFillSegments`), gradients: SVG linearGradient, PNG axis lerp, PDF first color (`TestSceneHatchFill`, `TestHatchPatternExpanded`, `TestFillGradientBackends`) |
| Dimension rendering | `render/dimension.py`, `dim_*.py` | `PARTIAL` — dimension block content lowered to paths by the converter; full block renderer (transforms, text placement at paper scale) not modeled |
| MTEXT rendering | `render/abstract_mtext_renderer.py`, `tools/text_layout.py` | `DONE` — inline codes stripped (`StripCodes`), greedy word wrap at column width (`LayoutMText`), paragraph pitch 1.6714x height, TrueType glyph outlines via system fonts with marching-squares contours and built-in vector font fallback (`TestTextToPathsTTF`, `TestLayoutMTextWrap`, `TestTextOutlines`) |
| Mesh rendering | `render/mesh.py`, `mesh_builder.py` | `PARTIAL` |
| Arrow heads | `render/arrows.py` | `PARTIAL` |
| MLeader/MLine rendering | `render/mleader.py`, `mline.py` | `DONE` — converted to path geometry (leader arrows via path, mline style lines) and rendered by all backends (converter tests in `path/converter_ext_test.go`, corpus conformance) |

Plot style tables (CTB/STB) are not modeled: entities render with their screen colors. Print/plot scaling (PS/PDF paper layout) is out of scope. Text rendering uses system TrueType fonts (Noto/Liberation/DejaVu/Vera) via `golang.org/x/image/font/opentype`; when none is available the built-in single-stroke vector font is used and `StyledScene.Fallbacks` records the substitution.

### 7. Audit, recover, validation, repair

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Audit protocol and fixers | `audit.py` | `DONE` — pluggable check registry (handles, owners, tables, references, xdata, reactors, extension dictionaries, extents, entity invariant hooks), safe fixers with re-audit, bounded report with severity counts and JSON (`TestRegistryCodes`, `TestCleanDocumentZeroErrors`, `TestReferenceIntegrity`, `TestTableIntegrity`, `audit/parity_test.go`) |
| Recover damaged files | `recover.py` | `DONE` — `RecoverFile`/`RecoverReader` merge loader recovery (truncated sections, skipped malformed records) with a full audit pass; report attached to the document (`TestRecoverTruncatedFile`, `TestRecoverMalformedEntity`, `audit/recover.go`) |
| R12 strict checker | `r12strict.py` | `DONE` — strict R12 audit flags OBJECTS/CLASSES sections and opaque object sections against the version profile (`TestStrictR12RejectsObjects`); the sanitizing mode is `drawing/r12strict.go` |
| Tag/validator checks | `lldxf/validator.py`, `lldxf/repair.py` | `PARTIAL` — loader recovers truncated sections and skips malformed records with a report (`TestRecoverTruncatedFile`); tag-level rewriting is not modeled |
| Orphan/cycle repair | `audit.py` | `PARTIAL` — missing/dangling entity owners resolved to *Model_Space block record, missing table entries created (`TestMissingOwnerReportedAndFixed`, `TestMissingLayerZeroReportedAndFixed`); duplicate handles are reported, not repaired; cycle detection is not modeled |
| Header/table/block/object consistency | `audit.py` | `DONE` — required default entries, duplicate/empty entry names, owner mismatches, block record/layout links, dictionary entries, header extents (`TestTableIntegrity`, `TestDictionaryEntryUnresolved`, `TestExtentsAndSplineHook`) |

### 8. Query, selection, transformation, decomposition

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| Query language | `query.py`, `queryparser.py` | `DONE` — expression lexer plus function predicates (`bbox_intersects`, `bbox_within`, `near`, `has_xdata`, `xdata_value`, `attr`), compiled predicates with position-carrying errors, spatial pre-filtering (`query.ExecuteWithFilter`) and document-level `Drawing.QueryExpression`; tests in `query/parity_test.go` |
| `select` filters | `select.py` | `DONE` — `query.EntitySelector` filters plus handle-based `query.Selection` sets with union/intersection/difference and named sets on the drawing (`SetSelectionSet`); `TestQuerySelectionsAndTransformsE2E` |
| `groupby` | `groupby.py` | `DONE` — `query.GroupByAttribute`/`GroupHandles` over layer, type, color, linetype, text and `attr:TAG` INSERT attributes; grouping tests in `query/parity_test.go` |
| Transform entities/layouts | `transform.py` | `DONE` — matrix builders (`RotationMatrix`, `ScaleMatrix`, `NonUniformScaleMatrix`) plus `Selection.ApplyM` and `Drawing.TransformSelection`/`TransformBlock`/`TransformDrawing`; `TestTransformBlockAndDrawingE2E` |
| Explode INSERT/POLYLINE/MTEXT/etc. | `explode.py`, `addons/mtextsurrogate` | `DONE` — INSERT, POLYLINE, HATCH, LEADER, MLEADER and DIMENSION (extension lines, dimension line, arrow ticks, text); MText surrogate formatting is not modelled (MText explodes to itself); `TestExplodeDimension` and the dimension test in `query/parity_test.go` |
| Bounding box of entities/layouts | `bbox.py` | `PARTIAL` — per-entity `BBox()` feeds `query.EntityContext`/`Selection.Bounds` and layout extents (`Drawing.Extents`, `LayoutExtents`); exotic placeholder entities keep degenerate boxes |
| Disassemble to primitives | `disassemble.py` | `DONE` — `query.Disassemble` flattens nested inserts through composed insert matrices, clears handles and reports cycles; `Drawing.DisassembleModel`; nested/cycle tests in `query/parity_test.go` |
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
| `drawing` | `addons/drawing` | `PARTIAL` (`render`, `export`) |
| `r12writer` | `addons/r12writer.py` | `DONE` — `addons/r12writer` incremental R12 writer (header/tables + entities, no full document; `TestR12WriterStreaming` 10k entities) |
| `iterdxf` | `addons/iterdxf.py` | `DONE` — `addons/iterdxf` bounded-memory record iterator in file order (`TestIterDXFStreamingRecords`) |
| `importer` | `addons/importer.py` | `DONE` — `addons/importer.ImportSVG` (line/rect/circle/polyline/path; `TestImportSVGElements`) |
| `dxf2code` | `addons/dxf2code` | `DONE` — `addons/dxf2code.Generate` emits Go source (parsed with go/parser, inventory counts) |
| `acadctb` | `addons/acadctb.py` | `PARTIAL` — `addons/acadctb` STB text parse + Apply (color/lineweight); binary CTB unsupported with descriptive error (`TestParseAndApply`) |
| `pycsg` | `addons/pycsg.py` | `DONE` — real triangle-level booleans + Box/Sphere primitives + `csg.Volume` (`csg/boolean_test.go`) |
| `MTextExplode` | `addons/mtextsurrogate` | `DONE` — `addons/mtxpl` Lines/Explode → positioned TEXT entities |
| `text2path` | `addons/text2path.py` | `STUB` |
| `geo` | `addons/geo.py` | `DONE` — `addons/geo.ExportGeoJSON` + `CRSFunc` transform |
| `meshex` | `addons/meshex.py` | `DONE` — `addons/meshex.ExportOBJ` |
| `openscad` | `addons/openscad.py` | `DONE` — `addons/meshex.ExportOpenSCAD` polyhedron source |
| `odafc` | `addons/odafc.py` | `DONE` — `addons/odafc` CommandRunner + actionable missing-tool error |
| `hpgl2` | `addons/hpgl2` | `DONE` — `addons/hpgl2.ExportHPGL2`/`PlotWithRunner` |
| `xplayer` | `addons/xplayer.py` | `DONE` — `addons/xplayer.ExportAnimationSVG` per-frame SVG |
| `binpacking`, `tablepainter`, `gerber_D6673`, `mtxpl` | `addons/*` | `DONE` — `addons/binpacking.PackRectangles` (shelf pack), `addons/tablepainter.PaintTable`, `addons/gerber.ExportGerber` (RS-274X), `addons/mtxpl` |

### 11. CLI and tooling

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| `ezdxf` CLI (`info`, `audit`, `draw`, `browse`, `convert`, `strip`, `view`, `pp`) | `commands.py`, `__main__.py` | `ABSENT` |
| App settings / config | `appsettings.py`, `_options.py` | `ABSENT` |
| Comments | `comments.py` | `PARTIAL` |
| Fonts, SHX/TTF, text metrics | `fonts/*`, `tools/text*.py` | `STUB` |
| ISO hatch patterns / standards | `tools/_iso_pattern.py`, `standards.py` | `PARTIAL` |
| Revcloud | `revcloud.py` | `DONE` — `query.RevCloudRectangle`/`RevCloudArcs` (closed bulged LWPOLYLINE, configurable arc chord and bulge) plus `Space.AddRevCloud`/`Drawing.AddRevCloud`; revcloud tests in `query/parity_test.go` and `TestQuerySelectionsAndTransformsE2E` |
| Zoom/extents helpers | `zoom.py` | `DONE` — `drawing.FitExtentsToViewport` and `Drawing.FitView`/`FitLayoutView`; `TestQuerySelectionsAndTransformsE2E` |
| Analyze/difftags/debug tools | `tools/analyze.py`, `difftags.py` | `ABSENT` |

### 12. Colors, enums, units, standards

| Capability | ezdxf reference | Go baseline |
|---|---|---|
| ACI colors, true color, transparency | `colors.py`, `concepts/true_color.rst` | `DONE` — tagged color value (ACI/true color/BYLAYER/BYBLOCK) with full ACI table, nearest-index lookup and validation (`color/colorvalue.go`), alpha transparency (`color/transparency.go`), entity round trips (`color/roundtrip_test.go`, `color/colorvalue_test.go`) |
| Enums (text alignment, hatch, etc.) | `enums.py` | `DONE` — typed enum sets with name/numeric conversion and unknown-value reporting (`enums/enums.go`, `enums/enums_test.go`) |
| Units and insunits | `units.py`, `concepts/units.rst` | `DONE` — unit definitions 0-20 with meter factors, `$INSUNITS` mapping and length conversion (`units/units.go`, `units/units_test.go`) |
| Standards checks | `tools/standards.py` | `ABSENT` |
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
