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

### Requirement: ACAD_TABLE content is typed

The system SHALL expose ACAD_TABLE rows, columns, cells, table style, and merge
information through typed accessors while preserving unknown tags for round-trip
output.

#### Scenario: Read an ACAD_TABLE with cells

- **WHEN** a drawing containing an ACAD_TABLE with text cells is loaded
- **THEN** the table exposes its row and column count, cell text, and cell
  coordinates
- **AND** saving the drawing preserves the original table record.

#### Scenario: Unknown table tags survive

- **WHEN** an ACAD_TABLE contains a group code the typed model does not interpret
- **THEN** the tag is retained and emitted unchanged on save.

### Requirement: Subentity variants are typed

The system SHALL model POLYFACE and POLYMESH VERTEX variants and MESHVERTEXCACHE
records with their documented attributes and link them to their owning entity.

#### Scenario: Polyface mesh vertices

- **WHEN** a POLYLINE with the polyface flag is loaded
- **THEN** its vertices expose face records and vertex indices
- **AND** the parent polyline reports them as subentities.

#### Scenario: Mesh vertex cache

- **WHEN** a MESH with a vertex cache is loaded
- **THEN** the cache is available as a typed object linked to the mesh.

### Requirement: ACIS payloads are decoded

The system SHALL decode SAT text and SAB binary payloads sufficiently to report
ACIS version, record count, and body presence, and SHALL preserve the complete
payload for round-trip output.

#### Scenario: SAT body metadata

- **WHEN** a 3DSOLID with a SAT payload is loaded
- **THEN** the entity reports the ACIS version and a non-zero body count
- **AND** the original payload is emitted unchanged on save.

#### Scenario: SAB binary payload

- **WHEN** a 3DSOLID with a SAB payload is loaded
- **THEN** the payload is retained byte-for-byte and the entity reports that
  binary ACIS data is present.

### Requirement: Objects in ENTITIES are typed

The system SHALL load object-like records that appear in the ENTITIES section as
typed entities when a typed model exists, and SHALL retain them unchanged
otherwise.

#### Scenario: Layout record in ENTITIES

- **WHEN** a LAYOUT record is found in the ENTITIES section
- **THEN** it is exposed through the typed layout entity model
- **AND** it still round-trips as an ENTITIES record.

### Requirement: Extension dictionaries and reactors cover typed records

The system SHALL preserve extension dictionaries and persistent reactors on all
newly typed entity and object records.

#### Scenario: Round trip extension data

- **WHEN** a newly typed record carries an extension dictionary or reactor handles
- **THEN** those handles are readable through the typed API
- **AND** they are emitted unchanged on save.

### Requirement: Typed attributes match documented group codes

Every typed entity SHALL expose named access for its documented group codes, apply documented defaults, and gate version-specific fields by the target DXF version.

#### Scenario: Golden attribute round trip

- **WHEN** a representative LINE, SPLINE, HATCH, and ACAD_TABLE record is loaded, modified through typed access, and saved
- **THEN** every documented field appears under its correct group code with the expected default or modified value

#### Scenario: Version-gated field is omitted

- **WHEN** an entity field introduced after the save target is set and the document is saved to an older version
- **THEN** the field is omitted or converted according to the version profile and the conversion is reported

### Requirement: Copy isolates mutable state

A deep copy of an entity SHALL duplicate owned subentities, payloads, and embedded records so that changing the copy does not change the source.

#### Scenario: Embedded entity copy is independent

- **WHEN** an entity with an embedded entity or subentity collection is copied and the copy is modified
- **THEN** the original entity and its embedded state remain unchanged

#### Scenario: ACIS payload copy is independent

- **WHEN** an ACIS-backed entity is copied and the copy's payload is edited
- **THEN** the original payload bytes and metadata remain unchanged

### Requirement: Bounding boxes cover every entity category

Every supported entity category, including placeholders, proxies, underlays, tables, and object-like records, SHALL return a bounding box that reflects its drawable geometry or an explicitly documented empty result.

#### Scenario: Proxy bounding box is derived

- **WHEN** a proxy entity with decodable graphics is queried for its bounding box
- **THEN** the result encloses the decoded geometry instead of a zero-size box

#### Scenario: Empty entity reports empty box

- **WHEN** an entity has no drawable geometry
- **THEN** its bounding box is reported as empty through the documented empty representation
