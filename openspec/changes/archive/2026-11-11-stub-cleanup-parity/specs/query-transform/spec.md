# Spec Delta

## ADDED Requirements

### Requirement: Block purge removes only unused blocks
The system SHALL resolve INSERT block names, retain blocks reachable through nested inserts and layout ownership, and report each removed block.

#### Scenario: Remove unused block
- **WHEN** a block is not referenced by any insert or layout
- **THEN** purge removes it and reports its name

#### Scenario: Nested dependency retained
- **WHEN** a used block contains an insert of another block
- **THEN** purge retains the nested block even if it has no top-level insert

#### Scenario: Layout block retained
- **WHEN** a paper-space block is owned by a layout
- **THEN** purge retains it

### Requirement: Spatial queries honor entity geometry
The system SHALL answer drawing-level spatial queries from entity bounding boxes and apply the requested range, polygon, or nearest filter.

#### Scenario: Polygon query
- **WHEN** a polygon query is run over entities outside and inside the polygon
- **THEN** only entities whose bounds intersect the polygon are returned

### Requirement: MPolygon copies are independent
The system SHALL deep-copy an MPolygon's embedded entity state, paths, patterns, gradients, and seeds so mutating the copy leaves the original unchanged.

#### Scenario: Mutate copied MPolygon
- **WHEN** a copied MPolygon changes its layer, pattern, or path data
- **THEN** the original MPolygon retains its previous values
