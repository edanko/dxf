# Tasks

## 1. JSON round trip and file audit

- [x] 1.1 Define a versioned JSON envelope containing document metadata and an embedded ASCII DXF payload; make `ToJSON` emit it and `FromJSON` load it into the receiver only after successful validation.
- [x] 1.2 Verify entity, table, block, object, layer, and layout state survive a JSON round trip, malformed JSON leaves the receiver unchanged, and an unsupported document state returns an explicit error.
- [x] 1.3 Implement `AuditFileComprehensive` with the real `dxf.FromFile` loader; verify a clean file audits its actual entities, a missing file errors, and a malformed strict load reports the load error.
- [x] 1.4 Update `docs/cli.md`/`docs/parity/README.md` for the new JSON and audit behavior.

## 2. Boolean meshes and spatial indexing

- [x] 2.1 Add a checked boolean API backed by `csg` for union, intersection, difference, and XOR, returning explicit errors for invalid, open, non-manifold, or empty results.
- [x] 2.2 Verify overlapping cubes produce operation-specific volumes (intersection, difference, union, XOR) and the invalid-input error path.
- [x] 2.3 Deprecate `math.PerformBoolean` as a wrapper with documented migration to the checked API, or remove it; verify no caller relies on the old placeholder behavior.
- [x] 2.4 Back `EntitySpatialIndex` with real entity bounding boxes and the R-tree; verify range queries exclude distant entities, nearest uses box distance, and remove/reinsert works.
- [x] 2.5 Fix `RTree.Remove`, unskip `TestRTreeRemove`, and verify shrinking node bounds and item counts.
- [x] 2.6 Remove or convert the legacy point-based `SpatialIndex` to a documented box adapter; verify no zero-center or 1x1 placeholder behavior remains.
- [x] 2.7 Update `docs/parity/README.md` math/spatial rows.

## 3. Purge, spatial query, and MPolygon copy

- [x] 3.1 Rewrite `PurgeUnusedBlocks` with recursive reachability from model and paper layouts, preserving layout, xref, and standard blocks and resolving INSERT block names through the typed API.
- [x] 3.2 Add a purge report naming removed blocks and reasons while retaining the integer-returning wrapper; verify nested dependencies, layout blocks, xref blocks, and an unused nested block.
- [x] 3.3 Make spatial selection honor real entity geometry and polygon boundaries; verify a window query excludes an entity outside the polygon and includes one crossing the window.
- [x] 3.4 Deep-copy `MPolygon` embedded entity state, boundary paths, pattern, gradient, and seeds; verify mutating a copy does not change the source.
- [x] 3.5 Update `docs/parity/README.md` purge/spatial/copy rows.

## 4. Simplified geometry corrections

- [x] 4.1 Replace `Ellipse.ToBezier` with documented arc-to-cubic conversion; verify sampled points match the ellipse within tolerance and a degenerate ellipse errors.
- [x] 4.2 Correct SOLID3D transforms to transform every defining point and direction while preserving the ACIS payload and rejecting singular matrices; verify geometry and payload state.
- [x] 4.3 Correct boundary-path arc and bulge round trips; verify each edge type survives load, copy, transform, and format without changing its curve.
- [x] 4.4 Replace the simplified sun position calculation with the documented algorithm or return an unsupported error for invalid dates; verify a known date/location and an invalid-date error.
- [x] 4.5 Update `docs/parity/README.md` entity geometry limitations.

## 5. Documentation and integration

- [x] 5.1 Refresh `docs/parity/README.md` status, change table, remaining gaps, and hazards so they match the implemented code and this change set.
- [x] 5.2 Run `go test ./...` and `go vet ./...`; verify the unskipped spatial test and the deliberate failure tests for JSON, purge, boolean, and copy.
- [x] 5.3 Search for remaining silent placeholders in exported paths and either implement them or return explicit unsupported errors with tests.
