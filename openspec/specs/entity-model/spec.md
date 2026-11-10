# entity-model Specification

## Purpose
Defines the typed entity model, the attributes every entity shares, and the
operations the library provides on entities regardless of their concrete type.

## Requirements

### Requirement: Standard entities have typed models

Every standard DXF entity type that the library reads SHALL have a typed model
with named fields for its documented group codes, rather than a generic
key/value fallback.

#### Scenario: Load a surface entity

- **WHEN** a drawing containing an EXTRUDEDSURFACE is loaded
- **THEN** the entity is represented by a typed surface model exposing its
  control points and surface properties

#### Scenario: Load a mesh entity

- **WHEN** a drawing containing a MESH is loaded
- **THEN** the entity exposes its vertex and face arrays

#### Scenario: Load a dimension variant

- **WHEN** a drawing containing an ARC_DIMENSION is loaded
- **THEN** the entity is represented by the arc-dimension model, not a generic
  dimension

### Requirement: Common entity attributes

Every entity SHALL expose handle, owner, layer, linetype, color, true color,
lineweight, linetype scale, visibility, transparency, extrusion direction, and
extended data.

#### Scenario: Shared attributes round trip

- **WHEN** an entity with a layer, true color, lineweight, and xdata is loaded
  and saved
- **THEN** all shared attributes are present with their original values

#### Scenario: Defaults are applied

- **WHEN** an entity omits an optional shared attribute
- **THEN** the model returns the documented default without mutating the file

### Requirement: Every entity loads from tags

Each typed entity SHALL provide a load path from parsed tags and SHALL ignore
unknown tags while retaining them for output.

#### Scenario: Load and re-save

- **WHEN** a typed entity is loaded and saved
- **THEN** known fields round trip and unknown tags remain in the output

### Requirement: Entities support affine transforms

Entities SHALL support translation, rotation, scaling, and general 4x4
transformation where the type permits it, and SHALL report when a transform is
not representable.

#### Scenario: Transform a line

- **WHEN** a line is translated and rotated
- **THEN** its start and end points reflect the composed transform

#### Scenario: Non-uniform scale of a circle

- **WHEN** a circle is non-uniformly scaled
- **THEN** the operation returns a documented error or converts to an ellipse
  according to the stated policy

### Requirement: Entities report bounding boxes

Entities SHALL report an axis-aligned bounding box in world coordinates, and
inserts SHALL include transformed block content.

#### Scenario: Insert bounding box

- **WHEN** an INSERT references a block containing geometry
- **THEN** its bounding box encloses the transformed block content

### Requirement: Entities can be copied

Entities SHALL support deep copy that duplicates owned subentities and
associations without sharing mutable state.

#### Scenario: Copy a polyline

- **WHEN** a polyline with vertices is copied and the copy is modified
- **THEN** the original polyline and its vertices are unchanged

### Requirement: Subentities are accessible

Container entities SHALL expose their subentities: polyline vertices, insert
attributes and seqend, and multileader content.

#### Scenario: Polyline vertices

- **WHEN** a POLYLINE is loaded
- **THEN** its VERTEX entities are accessible in order and its SEQEND is
  represented

### Requirement: Entities can be exploded

Entities that decompose into simpler geometry SHALL provide an explode
operation; types that cannot SHALL return a documented error.

#### Scenario: Explode an insert

- **WHEN** an INSERT is exploded
- **THEN** the returned entities are the block content transformed into the
  insert's coordinate system

### Requirement: Unknown entities are retained

Entity types the library does not model SHALL be preserved as raw-tag entities
that can be written back unchanged.

#### Scenario: Third-party entity

- **WHEN** a drawing contains an unmodelled third-party entity
- **THEN** it loads as a retained raw entity and is written back with its tags
