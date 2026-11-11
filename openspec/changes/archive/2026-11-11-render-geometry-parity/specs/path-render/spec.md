# Spec Delta

## ADDED Requirements

### Requirement: HATCH and MPOLYGON boundaries render all supported edge types
The path and scene pipeline SHALL convert polyline, line, arc, ellipse, spline, and bulge boundary edges, including holes and transformed boundaries, into closed fill paths.

#### Scenario: Mixed edge boundary
- **WHEN** a HATCH has an outer boundary with line and arc edges and an inner hole with a spline edge
- **THEN** the scene contains the outer and hole paths with the hole excluded from the fill

#### Scenario: Unsupported boundary edge
- **WHEN** a boundary contains an edge type that cannot be converted
- **THEN** rendering reports the entity handle and edge reason without dropping the supported edges

### Requirement: ELLIPSE embedded splines render through the spline pipeline
An ELLIPSE carrying embedded spline control data SHALL render from that spline data instead of falling back to the retained record.

#### Scenario: Embedded spline ellipse
- **WHEN** an ELLIPSE with embedded spline data is converted to a path
- **THEN** the path follows the spline control points and the scene emits the expected curve geometry

### Requirement: Complex linetypes render shape and text elements
The renderer SHALL place embedded shape and text linetype elements along the path with their rotation, scale, and offsets applied, in addition to dash and gap elements.

#### Scenario: Shape and text elements
- **WHEN** a complex linetype contains a shape element and a text element
- **THEN** rendering emits the shape and text geometry at the documented pattern positions

#### Scenario: Missing shape resource
- **WHEN** a complex linetype references a shape that cannot be resolved
- **THEN** rendering emits the remaining dash and text elements and reports the missing shape and linetype name

### Requirement: PDF gradients render full color stops
The PDF backend SHALL emit axial and radial shadings with all configured color stops instead of a single first color.

#### Scenario: Multi-stop axial gradient
- **WHEN** a scene contains an axial gradient with three color stops
- **THEN** the PDF content includes a shading object with all three stops in order

#### Scenario: Radial gradient
- **WHEN** a scene contains a radial gradient
- **THEN** the PDF content includes a radial shading with the configured center and radius

### Requirement: Render fallbacks are deterministic
When an entity cannot be rendered with supported geometry, the scene fallback report SHALL include the entity handle and a stable reason string.

#### Scenario: Repeat build
- **WHEN** the same drawing is rendered twice with the same options
- **THEN** both fallback reports contain the same entries in the same order

### Requirement: Visual conformance detects regressions
The render test harness SHALL compare a curated fixture set against golden raster and SVG output with documented tolerances for anti-aliasing and font differences.

#### Scenario: Changed output
- **WHEN** a fixture's rendered output differs beyond the documented tolerance
- **THEN** the conformance test fails and reports the fixture, backend, and differing region
