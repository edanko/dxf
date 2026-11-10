# geometry-math Specification

## Purpose
Defines the geometric and linear-algebra primitives the library exposes, their
coordinate conventions, and the numeric tolerance guarantees callers can rely on.

## Requirements

### Requirement: Vector algebra

The library SHALL provide 2D and 3D vectors with addition, subtraction, scaling,
dot product, cross product, length, normalization, distance, angle, and
component-wise operations.

#### Scenario: 3D cross product

- **WHEN** the cross product of two non-parallel 3D vectors is computed
- **THEN** the result is perpendicular to both input vectors

#### Scenario: Normalize zero vector

- **WHEN** a zero-length vector is normalized
- **THEN** the operation returns a documented zero or error result, never NaN

### Requirement: Matrix transforms

The library SHALL provide 4x4 matrices with composition, inversion, point and
vector transformation, and decomposition into translation, rotation, scale, and
shear.

#### Scenario: Compose and invert

- **WHEN** a transform is composed with its inverse and applied to a point
- **THEN** the original point is recovered within tolerance

#### Scenario: Singular matrix

- **WHEN** a singular matrix is inverted
- **THEN** the operation reports failure rather than returning invalid numbers

### Requirement: Coordinate systems

The library SHALL convert points, vectors, and directions between WCS, UCS, and
OCS using an explicit axis and elevation.

#### Scenario: OCS round trip

- **WHEN** a point is converted from WCS to OCS and back
- **THEN** the original point is recovered within tolerance

#### Scenario: Direction versus point

- **WHEN** a direction is transformed between coordinate systems
- **THEN** translation is not applied to it

### Requirement: Bounding boxes

The library SHALL provide 2D and 3D bounding boxes with expansion, union,
intersection, containment, center, size, and transformation.

#### Scenario: Union

- **WHEN** two boxes are unioned
- **THEN** the result contains both inputs

#### Scenario: Transform

- **WHEN** a box is transformed
- **THEN** the resulting axis-aligned box encloses all transformed corners

### Requirement: Bezier curves

The library SHALL evaluate Bezier curves, compute derivatives, split them, and
report their bounding box and arc length approximation.

#### Scenario: Endpoint interpolation

- **WHEN** a cubic Bezier is evaluated at parameters 0 and 1
- **THEN** the results equal its first and last control points

### Requirement: BSpline and NURBS

The library SHALL evaluate rational and non-rational BSpline curves and surfaces,
compute derivatives, insert knots, and support clamped and periodic forms.

#### Scenario: Clamped curve endpoints

- **WHEN** a clamped BSpline is evaluated at its parameter limits
- **THEN** the results equal its first and last control points

#### Scenario: Rational circle

- **WHEN** a NURBS representation of a circle is evaluated
- **THEN** all sampled points lie on the circle within tolerance

### Requirement: Curve intersections

The library SHALL compute line-line, line-curve, curve-curve, and ray-plane
intersections with explicit handling for parallel, coincident, and tangent cases.

#### Scenario: Line-line intersection

- **WHEN** two non-parallel lines intersect
- **THEN** the reported point lies on both lines within tolerance

#### Scenario: Parallel lines

- **WHEN** two parallel non-coincident lines are intersected
- **THEN** the result reports no intersection rather than a spurious point

### Requirement: Curve offset

The library SHALL offset lines, arcs, circles, polylines, and splines by a signed
distance, with configurable join handling for corners.

#### Scenario: Offset a polyline

- **WHEN** an open polyline is offset by a positive distance
- **THEN** the result is parallel at that distance and preserves endpoint
  orientation according to the documented join rule

### Requirement: Polygon operations

The library SHALL compute polygon area, centroid, orientation, point containment,
convex hull, and clipping against a rectangle or convex polygon.

#### Scenario: Area of a square

- **WHEN** the area of a unit square is computed
- **THEN** the result is 1 within tolerance regardless of vertex order

#### Scenario: Point in polygon

- **WHEN** a point inside a simple polygon is tested
- **THEN** containment returns true; a point outside returns false

### Requirement: Triangulation

The library SHALL triangulate simple polygons with holes and point sets, and SHALL
handle degenerate and collinear input without panicking.

#### Scenario: Polygon with hole

- **WHEN** a polygon with one hole is triangulated
- **THEN** the triangles cover the polygon area and exclude the hole

#### Scenario: Degenerate input

- **WHEN** collinear or duplicate points are triangulated
- **THEN** the operation returns a valid, possibly empty, result

### Requirement: Bulge and arc conversion

The library SHALL convert between bulge values, arcs, and center/radius/angle
representations in both directions.

#### Scenario: Bulge round trip

- **WHEN** a bulge value is converted to an arc and back
- **THEN** the original bulge is recovered within tolerance

### Requirement: Numeric tolerance is documented

The library SHALL define and document a default absolute tolerance and SHALL
accept caller-supplied tolerances for comparisons and intersections.

#### Scenario: Custom tolerance

- **WHEN** a comparison is made with a caller-supplied tolerance
- **THEN** the supplied value governs the result
