# Proposal

## Why

Geometry underpins entities, transforms, bounding boxes, rendering, and queries.
The Go `math` package already has vectors, matrices, bounding boxes, Bezier,
BSpline/NURBS, bulge, earcut, Delaunay, offset, and OCS/UCS, but the surface is
incomplete and inconsistent compared with ezdxf's `math` and `acc` modules.
Missing operations force callers to reimplement geometry and produce divergent
results.

## What Changes

- Standardize vector, matrix, and quaternion APIs with documented conventions.
- Add matrix decomposition (translation, rotation, scale, shear) and inversion
  error handling.
- Add WCS/UCS/OCS conversion for points, vectors, and directions.
- Complete Bezier, BSpline, and NURBS evaluation, derivatives, knot insertion,
  and parameterization.
- Add curve-curve, curve-line, line-line, and ray-plane intersections.
- Add offset and parallel curve operations with join handling.
- Add polygon area, centroid, winding, containment, convex hull, and clipping.
- Harden triangulation (earcut and Delaunay) for degenerate input.
- Add bulge/arc conversions in both directions.
- Add interpolation and curve fitting (least squares) where ezdxf provides it.
- Define a shared epsilon/precision policy and document tolerances.

## Capabilities

### New Capabilities

- `geometry-math`: vectors, matrices, coordinate systems, curves, polygons,
  intersections, offsets, triangulation, and numeric tolerance behavior.

### Modified Capabilities

None.

## Impact

Affected code: `math/*` and all packages that consume it (`entity`, `drawing`,
`path`, `render`, `query`, `spatial`, `csg`). Public API additions and
documented conventions. Tests use analytic cases plus property-based checks.
