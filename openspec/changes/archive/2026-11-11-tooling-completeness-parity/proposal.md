# Proposal

## Why

Several ezdxf utilities and advanced record workflows are still missing. Callers cannot decompose composite geometry into primitives/paths/meshes, read dynamic-block definitions, store structured user data in XRECORDs, measure laid-out text, parse custom hatch patterns, recover full proxy graphics, inspect DWG metadata natively, or build linked-edge topology.

## What Changes

- Add primitive decomposition: `make_primitive`, recursive decompose, and conversions to vertices, paths, meshes, and control vertices, including MTEXT and other composite sources.
- Add dynamic-block read helpers that resolve a block reference to its dynamic block definition and expose dynamic-block metadata; creation/modification stays out of scope.
- Add `UserRecord` and `BinaryRecord` helpers for compiling, parsing, and round-tripping structured and binary data in XRECORD entities.
- Add text-size APIs for TEXT, MTEXT, and laid-out MTEXT extents using the existing font metrics and layout pipeline.
- Add a custom hatch-pattern parser for `.pat` files with definition lookup, scaling, measurement-system conversion, and render expansion.
- Complete proxy-graphics decoding for the documented primitive opcode set, not only embedded entity records.
- Add native DWG metadata and partial record parsing (header, classes, objects) alongside the existing external ODA conversion path; conversion continues to use the external tool.
- Add EdgeMiner/EdgeSmith-style linked-edge detection and entity creation (polylines, hatch boundary paths) from edge networks.
- Add the remaining small public utilities where they have observable behavior: Julian-date conversion, ascending/descending entity reorder, and raw section loading.

## Capabilities

### New Capabilities

- `edge-topology`: linked-edge detection, vertex/degree analysis, loop extraction, and creation of polylines and hatch boundaries from edge networks.
- `hatch-patterns`: custom PAT parsing, pattern definition lookup, scaling, measurement-system conversion, and expansion to line segments.

### Modified Capabilities

- `dxf-io-core`: raw section loading without document binding.
- `geometry-math`: Julian-date conversion.
- `query-transform`: primitive/path/mesh disassembly and ascending/descending reorder helpers.
- `entity-model`: dynamic-block read helpers, structured XRECORD user/binary records, and complete proxy-graphics primitives.
- `addons`: text-size APIs, native DWG metadata parsing, and ACIS/edge-backed export integration.
- `path-render`: text-size inputs and hatch expansion use the completed pattern and metric APIs.

## Impact

Affected code: `query`, `entity`, `object`, `drawing`, `text2path`, `patterns`, `render`, `addons/dwg`, new `addons/edgetools` or equivalent, and tests. DWG parsing is read-only and must never replace or weaken the external converter. Pattern parsing and edge extraction need malformed-input coverage. No public API removal is planned.
