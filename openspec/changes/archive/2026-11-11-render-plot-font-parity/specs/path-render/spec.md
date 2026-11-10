# Spec Delta

## ADDED Requirements

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
