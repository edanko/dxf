# Tasks

## 1. Core algebra

- [ ] 1.1 Document and complete Vec2/Vec3 operations including zero-safe normalization; verify with unit tests
- [ ] 1.2 Add matrix decomposition and singular-inverse errors; verify compose/invert tests
- [ ] 1.3 Standardize WCS/UCS/OCS point, vector, and direction conversion; verify round trips

## 2. Bounding boxes

- [ ] 2.1 Complete BoundingBox2D/3D union, intersection, containment, and transform; verify analytic cases

## 3. Curves

- [ ] 3.1 Complete Bezier evaluation, derivatives, split, bbox, and arc length; verify endpoint and derivative tests
- [ ] 3.2 Complete BSpline/NURBS evaluation, derivatives, knot insertion, clamped/periodic forms; verify rational circle test
- [ ] 3.3 Add curve fitting and interpolation in `math/fit`; verify against known point sets

## 4. Intersections and offsets

- [ ] 4.1 Add line-line, line-curve, curve-curve, and ray-plane intersections with typed results; verify parallel/coincident/tangent cases
- [ ] 4.2 Add curve offset with join handling; verify polyline and spline offsets

## 5. Polygons and triangulation

- [ ] 5.1 Add polygon area, centroid, orientation, containment, convex hull, and clipping; verify unit square and containment tests
- [ ] 5.2 Harden earcut and Delaunay for degenerate input and add diagnostics; verify hole and collinear cases

## 6. Conversions and tolerance

- [ ] 6.1 Add bulge/arc bidirectional conversion; verify round trips
- [ ] 6.2 Document and apply the tolerance policy across comparisons and intersections; verify custom-tolerance tests

## 7. Verification

- [ ] 7.1 Add property-based tests for transforms, offsets, and triangulation; verify no panics and invariant preservation
- [ ] 7.2 Update `docs/parity/README.md` math matrix

## Workflow follow-up

- Archive the change after review.
