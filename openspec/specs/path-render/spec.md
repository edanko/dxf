# path-render Specification

## Purpose
Defines how entities become drawable paths and how those paths are rendered to
vector and raster backends with correct styling and layout placement.

## Requirements

### Requirement: Paths from every renderable entity

The library SHALL produce paths for LINE, ARC, CIRCLE, ELLIPSE, SPLINE,
LWPOLYLINE, POLYLINE, HATCH, TEXT, MTEXT, DIMENSION, LEADER, MLEADER, SOLID,
TRACE, 3DFACE, IMAGE, WIPEOUT, and supported surfaces.

#### Scenario: Ellipse to path

- **WHEN** an ELLIPSE is converted to a path
- **THEN** the path is a closed curve matching the ellipse's center, axes, and
  rotation within tolerance

#### Scenario: Spline to path

- **WHEN** a SPLINE is converted to a path
- **THEN** the path samples the curve within the configured tessellation
  tolerance

### Requirement: Vector backends

The library SHALL provide SVG and PDF backends that render paths with stroke,
fill, color, lineweight, and transparency.

#### Scenario: SVG output

- **WHEN** a drawing with styled entities is rendered to SVG
- **THEN** the output contains path elements with the corresponding stroke and
  fill attributes

#### Scenario: PDF output

- **WHEN** a drawing is rendered to PDF
- **THEN** the output is a valid PDF whose page content contains the geometry

### Requirement: Raster backend

The library SHALL provide a raster backend that renders to an image with a
configurable size, background, and anti-aliasing.

#### Scenario: PNG dimensions

- **WHEN** a drawing is rendered to PNG at 800x600
- **THEN** the resulting image has those dimensions and non-background pixels
  where geometry exists

### Requirement: Text renders as outlines

Text SHALL be converted to path outlines using real font glyph data, including
alignment, rotation, and width factor.

#### Scenario: Text glyphs

- **WHEN** a TEXT entity is rendered
- **THEN** its glyph outlines are present in the output rather than placeholder
  boxes

#### Scenario: MTEXT formatting

- **WHEN** an MTEXT entity with multiple lines and formatting is rendered
- **THEN** line breaks and supported formatting affect the output layout

### Requirement: Hatch fills

HATCH entities SHALL render solid fills, pattern fills, and gradient fills
according to their definition.

#### Scenario: Solid hatch

- **WHEN** a solid HATCH is rendered
- **THEN** its boundary is filled with the hatch color

#### Scenario: Pattern hatch

- **WHEN** a pattern HATCH is rendered
- **THEN** the pattern lines are drawn within the boundary at the declared scale
  and angle

### Requirement: Layouts and viewports

Rendering SHALL support paper-space layouts and viewports, applying each
viewport's view center, height, and transform.

#### Scenario: Layout render

- **WHEN** a layout with a viewport is rendered
- **THEN** the viewport shows the model-space content at the declared view

### Requirement: Clipping and extents

The renderer SHALL support clipping to a rectangle or boundary and SHALL fit
content to explicit extents.

#### Scenario: Clip rectangle

- **WHEN** a clip rectangle excludes part of the drawing
- **THEN** no rendered geometry appears outside it

### Requirement: Backends are pluggable

Renderers SHALL be registerable by name and SHALL share a common interface for
paths, styles, and output.

#### Scenario: Custom backend

- **WHEN** a backend is registered under a name and selected
- **THEN** rendering dispatches to it without changing caller code

### Requirement: Path nesting is available

The system SHALL classify closed paths into outer boundaries and nested holes and SHALL expose the nesting tree.

#### Scenario: Nested rings

- **WHEN** two closed paths form concentric rings
- **THEN** nesting reports one outer path with one hole
- **AND** the hole orientation is normalized for filling.

### Requirement: Dimensions render from block content

The system SHALL render DIMENSION entities using their referenced dimension block content with the entity transform, measured text, and configured text placement.

#### Scenario: Linear dimension

- **WHEN** a linear dimension is rendered
- **THEN** extension lines, dimension line, arrowheads, and measured text appear at the dimension's location
- **AND** the text matches the dimension measurement format.

### Requirement: Meshes render as faces

The system SHALL render MESH and polyface mesh entities as filled or wireframe faces according to scene style.

#### Scenario: Mesh with faces

- **WHEN** a MESH with faces is rendered
- **THEN** each face is emitted as a closed path
- **AND** hidden or invalid faces are skipped.

### Requirement: Arrow heads are complete

The system SHALL render closed, open, dot, architectural tick, and oblique arrow heads for dimensions and leaders.

#### Scenario: Closed arrow

- **WHEN** a dimension uses a closed filled arrow
- **THEN** the rendered output contains a filled arrow polygon at each endpoint.

### Requirement: Text layout uses measured metrics
The system SHALL use the text-size APIs for line width, wrapping decisions, alignment, and extents when converting TEXT and MTEXT to paths and scene geometry.

#### Scenario: Wrapped MTEXT width
- **WHEN** MTEXT wraps because the measured line exceeds its column width
- **THEN** each rendered line fits the column and the paragraph height matches the measured extents

#### Scenario: Alignment uses measured width
- **WHEN** center- or right-aligned text is laid out
- **THEN** its anchor is offset using the measured advance rather than a fixed character estimate

### Requirement: Scene hatching expands parsed pattern definitions
The system SHALL expand predefined and parsed custom hatch pattern definitions into clipped line segments and report a named pattern that cannot be resolved.

#### Scenario: Custom pattern expansion
- **WHEN** a hatch references a parsed custom pattern and a closed boundary
- **THEN** the scene contains clipped pattern segments inside the boundary

#### Scenario: Unknown pattern name
- **WHEN** a hatch references a pattern name that is not defined
- **THEN** the scene reports the entity handle and missing pattern name and keeps the boundary stroke

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
