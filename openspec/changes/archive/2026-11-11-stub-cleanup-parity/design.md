# Design

## Context

Several exposed APIs are stubs or simplified to the point of returning wrong results: `Drawing.FromJSON` is a no-op, `ComprehensiveAuditor.AuditFile` always fails, `math.PerformBoolean` returns the first mesh for intersection/difference/XOR, `spatial.SpatialIndex` uses 1x1 point boxes and zero centers, `RTree.Remove` is skipped by tests, `PurgeUnusedBlocks` cannot resolve INSERT block names, and `MPolygon.Copy` shares its embedded entity. Additional geometry helpers are documented as simplified. See `proposal.md` for the complete list.

Constraints:

- Do not leave APIs that silently report success while doing nothing.
- Prefer correct implementations; where that is not feasible, return an explicit unsupported error.
- Preserve existing callers where a safe wrapper is possible.
- Document any intentional breaking change and its replacement.

## Goals / Non-Goals

**Goals:**

- Make JSON serialization round-trip supported document state or fail explicitly.
- Make file auditing load and audit the real file.
- Replace boolean, spatial, purge, copy, and geometry stubs with correct behavior or explicit errors.
- Refresh the parity documentation so it matches the code.

**Non-Goals:**

- New DXF features beyond the stubs listed here.
- Byte-for-byte JSON compatibility with any prior format; `FromJSON` currently accepts nothing.
- Full ACIS or DWG geometry; those belong to other changes.

## Decisions

### JSON round-trip via a versioned envelope

`ToJSON` emits a versioned envelope containing metadata plus the document's ASCII DXF payload encoded as a string. `FromJSON` validates the version and payload, loads the payload with the normal DXF reader, and replaces the receiver only after a successful load. Unsupported envelopes return a descriptive error and leave the receiver unchanged.

Alternative: a typed JSON projection of every entity and table. Rejected because it duplicates the DXF codec, risks divergence, and cannot represent unknown tags without redesigning it as a second wire format.

The existing summary fields can remain alongside the payload for human inspection, but they are not the source of truth.

### File audit uses the normal loader

`ComprehensiveAuditor.AuditFile` calls `dxf.FromFile`. Missing, unreadable, or structurally invalid files return an error. Recovery remains a separate explicit path, not an implicit fallback.

### Boolean operations delegate to CSG

`PerformBooleanChecked` converts both meshes to `csg` volumes, applies the requested operation, and converts the result back to an indexed mesh. Invalid indices, non-manifold input, or empty results return errors. The old `PerformBoolean` becomes a deprecated wrapper that returns an empty mesh and records no success state; callers needing errors use the checked API.

Alternative: implement a second BSP boolean engine in `math`. Rejected because `csg` already owns tested boolean geometry.

### Spatial indexing backed by entity bounds

`EntitySpatialIndex` remains the entity-level API and is backed by the R-tree with each entity's bounding box. `Query` returns only intersecting boxes, `QueryNearest` uses box distance, `Remove` removes the stored box/data pair, and polygon queries use the precise entity box or geometry.

The legacy point-based `SpatialIndex` is either removed with a migration note or reimplemented as a thin box-based adapter. No API may claim geometry-aware queries while using zero centers.

### Block purge by resolved references

`PurgeUnusedBlocks` walks INSERT references recursively through block definitions, starting from model and paper layouts, and preserves blocks that are referenced, layout-owned, xref-owned, or part of standard model/paper space. A report lists removed block names and reasons. The integer-returning wrapper remains for compatibility.

### MPolygon deep copy

`MPolygon.Copy` deep-copies its embedded entity, base attributes, boundary paths, pattern, gradient, and seeds. No pointer is shared with the source.

### Correct or reject simplified geometry

- ELLIPSE Bezier conversion uses the documented arc-to-cubic construction and matches sampled ellipse points within tolerance.
- SOLID3D transforms apply the matrix to all points and preserve direction/extrusion semantics; singular matrices return errors.
- Boundary path arc and bulge conversions use the same bulge/arc helpers as the path package.
- SUN position uses a documented solar-position algorithm or returns an unsupported error for invalid dates.

No helper may keep a "simplified" path that returns visibly wrong geometry without a diagnostic.

### Documentation refresh

`docs/parity/README.md` is updated in the same change: stale "not applied" notes, block purge, spatial indexing, disassembly scope, DWG behavior, and known hazards must match the code.

## Risks / Trade-offs

- [JSON payload duplicates DXF text and is larger than a typed projection] → Accept size for correctness; document the envelope and keep summaries optional.
- [Changing boolean behavior may surprise callers relying on the old first-mesh result] → Provide the checked API, deprecate the old wrapper, and document the migration.
- [Removing point-based spatial APIs is breaking] → Keep a compatible adapter where possible and document the replacement.
- [Correct solar and solid geometry expands scope] → Prefer explicit unsupported errors over approximate results where the documented algorithm is unavailable.
- [Purge can delete blocks needed by unloaded xrefs or layouts] → Treat xref-owned and layout-owned blocks as roots and report every removal.

## Migration Plan

`PerformBoolean` and legacy spatial APIs may be deprecated or removed with documented replacements. JSON has no existing behavior to migrate. All other changes are additive or fix previously failing behavior. The parity README records the migration notes.
