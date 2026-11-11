# Spec Delta

## ADDED Requirements

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
