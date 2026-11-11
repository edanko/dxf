# query-transform Specification

## Purpose

Defines how callers query, group, select, transform, explode, and reorder
entities in a loaded or authored document.

## Requirements

### Requirement: Query expressions

The library SHALL evaluate query expressions that filter entities by type,
layer, handle, color, linetype, attribute values, xdata, and spatial predicates.

#### Scenario: Filter by type and layer

- **WHEN** the query `type == 'LINE' and layer == 'Walls'` is evaluated
- **THEN** only LINE entities on layer `Walls` are returned

#### Scenario: Attribute filter

- **WHEN** a query filters on an attribute value
- **THEN** entities whose attribute matches are returned and others excluded

#### Scenario: Spatial filter

- **WHEN** a query selects entities intersecting a rectangle
- **THEN** only entities whose bounding boxes intersect it are returned

### Requirement: Query errors are reported

Invalid query syntax or unknown attributes SHALL produce a clear error without
panicking.

#### Scenario: Syntax error

- **WHEN** an invalid expression is parsed
- **THEN** an error identifies the offending position

### Requirement: Grouping

The library SHALL group query results by a chosen attribute and return the groups
with their members.

#### Scenario: Group by layer

- **WHEN** entities are grouped by layer
- **THEN** each group contains exactly the entities on that layer

### Requirement: Selection sets

The library SHALL support named or value selection sets that can be combined
with union, intersection, and difference, and reused across operations.

#### Scenario: Combine selections

- **WHEN** two selections are intersected
- **THEN** the result contains only entities present in both

### Requirement: Transform operations

The library SHALL transform individual entities, selections, blocks, and whole
documents by translation, rotation, scaling, and arbitrary matrices.

#### Scenario: Transform a selection

- **WHEN** a selection is rotated about a point
- **THEN** every selected entity's geometry reflects the rotation and unselected
  entities are unchanged

#### Scenario: Transform a block definition

- **WHEN** a block definition is transformed
- **THEN** all inserts of that block reflect the new definition

### Requirement: Explode and disassemble

The library SHALL explode composite entities into their constituent entities and
SHALL disassemble nested inserts into a flat list.

#### Scenario: Explode a dimension

- **WHEN** a DIMENSION is exploded
- **THEN** its lines, arrows, and text are returned as separate entities

#### Scenario: Disassemble nested inserts

- **WHEN** a document with nested inserts is disassembled
- **THEN** the result contains the fully transformed leaf entities

### Requirement: Draw order

The library SHALL support reading and changing entity draw order and SHALL
preserve it on save.

#### Scenario: Send to back

- **WHEN** an entity is sent to the back of the draw order
- **THEN** it sorts before the other entities and the order survives a round trip

### Requirement: Upright text

The library SHALL normalize text entities so their readable direction is upright
without changing their insertion point.

#### Scenario: Upside-down text

- **WHEN** an upside-down text entity is made upright
- **THEN** its rotation changes by 180 degrees and its position is preserved

### Requirement: Zoom and extents

The library SHALL compute zoom factors and view centers for a given extents and
aspect ratio.

#### Scenario: Fit extents

- **WHEN** extents are fitted into a viewport of known aspect ratio
- **THEN** the resulting view contains all extents with the expected margin

### Requirement: Block reference traversal

The library SHALL traverse block references, including nested references, and
report cycles without infinite recursion.

#### Scenario: Nested references

- **WHEN** a block contains an insert of another block
- **THEN** traversal reports both levels in order

#### Scenario: Reference cycle

- **WHEN** blocks reference each other cyclically
- **THEN** traversal terminates and reports the cycle

### Requirement: Revcloud

The library SHALL create revision-cloud polylines around a boundary or set of
points.

#### Scenario: Revcloud from rectangle

- **WHEN** a revcloud is created from a rectangle
- **THEN** the result is a closed polyline following the boundary with the
  configured arc style

### Requirement: Proxy graphics

The library SHALL extract proxy graphics from proxy entities when present.

#### Scenario: Proxy with graphics

- **WHEN** a proxy entity contains proxy graphics
- **THEN** the extracted entities represent that graphics data

### Requirement: Query and bounds cover all typed entities

Spatial queries, selection bounds, and layout extents SHALL use the bounding box of every supported entity category and SHALL report empty bounds when no entity contributes geometry.

#### Scenario: Query finds proxy geometry

- **WHEN** a proxy entity with decodable graphics lies inside the query rectangle
- **THEN** the spatial query returns it based on its actual bounding box

#### Scenario: Empty selection reports empty bounds

- **WHEN** a selection contains only entities without drawable geometry
- **THEN** bounds are reported as empty and no zero-origin extent is fabricated

### Requirement: Composite entities decompose to primitives
The query API SHALL expose primitive decomposition that recursively expands INSERT, MTEXT, POLYLINE, HATCH, and other supported composites into renderable primitives.

#### Scenario: Nested insert decomposition
- **WHEN** an INSERT contains a nested INSERT with a LINE and TEXT
- **THEN** recursive decomposition yields the transformed LINE and TEXT primitives

#### Scenario: Unsupported primitive source
- **WHEN** an entity has no primitive representation
- **THEN** decomposition reports the entity type and handle instead of silently dropping it

### Requirement: Primitives convert to vertices paths and meshes
The query API SHALL provide vertex, path, mesh, and control-vertex conversion for primitives that support each representation.

#### Scenario: Curve to path
- **WHEN** an ARC primitive is converted to a path
- **THEN** the result is a curved path with the arc endpoints

#### Scenario: Mesh primitive
- **WHEN** a MESH or polyface primitive is converted to a mesh
- **THEN** the result preserves vertex positions and face indices

#### Scenario: No path representation
- **WHEN** a POINT primitive is converted to a path
- **THEN** conversion returns an empty result rather than fabricated geometry

### Requirement: Reorder helpers sort entities by handle
The query API SHALL provide ascending and descending reorder helpers that use source order for equal sort handles and accept an optional handle mapping.

#### Scenario: Ascending order
- **WHEN** entities with handles 20, 10, and 30 are reordered ascending
- **THEN** the result order is 10, 20, 30

#### Scenario: Equal sort handles
- **WHEN** two entities map to the same sort handle
- **THEN** they retain their original relative order
