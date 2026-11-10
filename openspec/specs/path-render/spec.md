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
