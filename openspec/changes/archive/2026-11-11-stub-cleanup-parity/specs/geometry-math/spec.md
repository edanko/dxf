# Spec Delta

## ADDED Requirements

### Requirement: Boolean mesh operations produce operation-specific geometry
The system SHALL return correct union, intersection, difference, and symmetric-difference geometry for supported closed triangle meshes, either directly or through the supported CSG implementation.

#### Scenario: Intersection of overlapping cubes
- **WHEN** two overlapping closed cube meshes are intersected
- **THEN** the resulting mesh covers only the shared volume and has a finite positive volume

#### Scenario: Difference removes the second mesh
- **WHEN** one cube is subtracted from another overlapping cube
- **THEN** the result excludes the second cube's volume

#### Scenario: Unsupported input
- **WHEN** an input mesh is not closed or has duplicate or degenerate faces
- **THEN** the operation returns an error naming the invalid mesh instead of returning an input mesh unchanged

### Requirement: Spatial indexing uses entity bounds
The system SHALL index each entity by its actual bounding box and answer range, intersection, nearest, update, and removal queries from those bounds.

#### Scenario: Range query excludes distant entity
- **WHEN** two entities are indexed far apart and a range around one is queried
- **THEN** only the entity whose bounds intersect the range is returned

#### Scenario: Nearest uses geometry
- **WHEN** nearest is queried from a point
- **THEN** the distance is measured to the entity bounds and the nearest entity is returned

#### Scenario: Remove and reinsert
- **WHEN** an entity is removed and the tree is queried again
- **THEN** it is absent, and a later reinsert makes it queryable again
