# Spec Delta

## ADDED Requirements

### Requirement: Ellipse Bezier conversion matches ellipse geometry
The system SHALL convert supported elliptical arcs and full ellipses to Bezier segments whose sampled points follow the ellipse within the documented tolerance, or return an explicit unsupported error.

#### Scenario: Full ellipse conversion
- **WHEN** a full ellipse is converted to Bezier segments
- **THEN** every sampled segment point lies on the ellipse within tolerance and the curve closes

#### Scenario: Unsupported embedded spline
- **WHEN** an ellipse record uses an embedded spline representation that cannot be converted
- **THEN** conversion returns an error naming the unsupported representation

### Requirement: Solid transforms preserve geometry
The system SHALL transform 3DSOLID and REGION points, directions, and ACIS payload consistently, and report an error when the inverse transform is singular.

#### Scenario: Transform round trip
- **WHEN** a solid is translated, rotated, and scaled, then inverse-transformed
- **THEN** its decoded points return to the original coordinates within tolerance

#### Scenario: Singular transform
- **WHEN** a transform with a zero determinant is applied
- **THEN** the operation returns a singular-matrix error

### Requirement: Boundary path edges round-trip exactly
The system SHALL preserve and transform line, arc, ellipse, spline, and bulge boundary path edges with their documented parameters.

#### Scenario: Arc boundary
- **WHEN** a hatch boundary contains an arc edge
- **THEN** its center, radius, start angle, end angle, and counterclockwise flag survive a load and save round trip

### Requirement: Sun position reports invalid input
The system SHALL compute solar position from documented date, time, and location inputs and return an error for out-of-range inputs.

#### Scenario: Valid solar position
- **WHEN** a valid date, time, and geographic location are supplied
- **THEN** the result reports finite azimuth and altitude values

#### Scenario: Invalid date
- **WHEN** an invalid month or day is supplied
- **THEN** the calculation returns an error naming the invalid field
