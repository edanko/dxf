# Spec Delta

## ADDED Requirements

### Requirement: Perlin noise is deterministic

The system SHALL generate Perlin noise for 2D and 3D coordinates from a caller-provided seed, producing identical output for identical inputs.

#### Scenario: Repeatable output

- **WHEN** the same seed and coordinate are used twice
- **THEN** both noise values are identical
- **AND** values stay within the documented range.

### Requirement: Point clustering is supported

The system SHALL group points into clusters using a documented distance criterion and SHALL expose cluster membership and centroids.

#### Scenario: Two distant groups

- **WHEN** points form two well-separated groups
- **THEN** clustering returns two clusters
- **AND** each point belongs to exactly one cluster.

### Requirement: Shape container represents closed boundaries

The system SHALL provide a Shape container that stores an outer boundary and zero or more holes, preserves orientation, and computes area and containment.

#### Scenario: Polygon with hole

- **WHEN** a shape is built from an outer ring and one inner ring
- **THEN** its area is the outer area minus the hole area
- **AND** a point inside the hole is outside the shape.

### Requirement: Portal clipping clips segments

The system SHALL clip line and curve segments against a rectangular portal and return only the visible segments.

#### Scenario: Segment crossing a portal

- **WHEN** a line starts outside a portal and ends inside it
- **THEN** clipping returns the portion inside the portal
- **AND** segments entirely outside are removed.

### Requirement: Triangulation preserves polygon area

The system SHALL triangulate polygons with holes so that the sum of triangle areas equals the input area within the documented tolerance.

#### Scenario: Centered hole

- **WHEN** a polygon contains a centered hole
- **THEN** triangulation covers the polygon without a bay outside the boundary
- **AND** the area check passes.
