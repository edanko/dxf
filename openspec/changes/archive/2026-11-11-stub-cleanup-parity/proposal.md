# Proposal

## Why

Several exported APIs are placeholders that return success without doing the work or silently return the wrong geometry. This is worse than an absent feature because callers cannot detect the failure, and the parity document currently overstates the affected areas.

## What Changes

- Implement JSON document round trip: `FromJSON` loads the data emitted by `ToJSON`, and `ToJSON` includes enough entity, table, block, and object data to reconstruct the document; otherwise mark the API unsupported with an explicit error.
- Implement `AuditFileComprehensive` using the real file loader, or return a documented unsupported error instead of a misleading placeholder.
- Replace `math.PerformBoolean` placeholders with real boolean results or remove the function in favor of the existing `csg` package, with migration guidance.
- Replace the legacy `spatial.SpatialIndex` placeholder with real entity bounding boxes and correct query behavior, and fix the skipped `RTree.Remove` case.
- Fix `Drawing.PurgeUnusedBlocks` so it resolves INSERT block names correctly, preserves nested dependencies, and reports what was removed.
- Deep-copy `MPolygon` embedded entity state instead of sharing it.
- Audit and correct simplified implementations in ellipse Bezier conversion, solid transforms, boundary paths, and sun position, or document each limitation and return an explicit error where the result would be misleading.
- Refresh `docs/parity/README.md` so completed changes, remaining gaps, and known hazards match the code.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dxf-io-core`: JSON document serialization round-trips supported document state or fails explicitly.
- `audit-recover`: file-level audit entry points behave correctly or fail explicitly.
- `geometry-math`: boolean mesh operations and spatial indexing return correct geometry or are removed in favor of the supported API.
- `query-transform`: block purge, spatial queries, and MPolygon copy semantics are correct.
- `entity-model`: simplified geometry conversions either produce correct results or report unsupported cases.

## Impact

Affected code: `drawing/serialization.go`, `audit/auditor_comprehensive.go`, `math/construct3d_mesh.go`, `spatial`, `drawing`, `entity/ellipse.go`, `entity/solid3d.go`, `entity/boundary_paths.go`, `entity/sun.go`, `entity/mpolygon.go`, and docs. The boolean and spatial changes may be **BREAKING** if a stub is removed; any removal must include a documented replacement and migration note. JSON output may change and loses compatibility with the previous incomplete format.
