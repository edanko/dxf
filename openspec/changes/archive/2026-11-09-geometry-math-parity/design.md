# Design

## Context

`math/` already contains Vec2/Vec3, Matrix44, BoundingBox2D/3D, Bezier, BSpline,
Bulge, Construct2D, earcut, Delaunay, Offset2D, and OCS/UCS. APIs and tolerance
behavior are inconsistent, and intersections, decomposition, hull, clipping, and
curve fitting are missing. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- A coherent, documented geometry API.
- Robust behavior for degenerate input.
- Shared tolerance policy.

**Non-Goals:**

- Exact symbolic geometry.
- Full 3D solid boolean evaluation (CSG add-on).
- Replacing existing types with a new naming scheme.

## Decisions

### Keep concrete Vec2/Vec3 and Matrix44

Extend the existing concrete types rather than introducing generics. Go generics
do not simplify mixed 2D/3D math and would churn call sites.

### Row-major 4x4 with documented convention

Standardize on row-major storage and column-vector multiplication, and document
it in the package. Add explicit helpers for translation/rotation/scale to avoid
caller confusion. Decomposition returns a struct with an error for
non-decomposable matrices.

### Tolerance as an explicit parameter

Provide package defaults (`Epsilon`) and accept a tolerance in comparison and
intersection APIs. Avoid hidden global state. Alternative: global settable
epsilon. Rejected because it makes tests order-dependent.

### Intersections return typed results

Return a result struct that distinguishes none, one, two, coincident, and
overlapping cases rather than a bare slice. This forces callers to handle
parallel/coincident cases.

### Triangulation hardening

Wrap earcut and Delaunay with input sanitation (deduplicate, remove collinear
runs, detect self-intersection) and return diagnostics. This prevents panics and
silent empty output.

### Curve fitting in `math/fit`

Add least-squares fitting and interpolation as a separate subpackage so the core
curve types stay dependency-free.

## Risks / Trade-offs

- **Convention mismatch** → Document and test row-major/column-vector behavior
  once; add conversion helpers.
- **Robustness cost** → Sanitization adds overhead; keep a fast path for clean
  input.
- **API duplication** → Deprecate, do not delete, older helpers.
- **Floating-point drift** → Use relative tolerances for large coordinates and
  test with translated geometry.
