# Design

## Approach

Keep new math utilities pure and independent of the DXF document model. `Shape` owns `[]Path`-style rings and reuses existing polygon and area helpers. Portal clipping builds on the existing segment clipping primitives but returns parameter ranges so curves can be split.

Triangulation fixes target the hole-bridge step: choose a bridge vertex that does not cross existing edges, then verify the triangulated area before returning. If verification fails, fall back to a deterministic bridge search.

## Testing

Add property-style tests for noise range/repeatability, clustering invariants, Shape area/containment, portal endpoint cases, and area-preserving triangulation. Reuse the existing triangulation diagnostic.
