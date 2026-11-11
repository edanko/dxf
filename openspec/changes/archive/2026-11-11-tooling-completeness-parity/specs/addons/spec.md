# Spec Delta

## ADDED Requirements

### Requirement: Text size APIs report measured extents
The system SHALL measure TEXT, single-line MTEXT, and laid-out MTEXT content in drawing units using the configured font metrics and layout rules.

#### Scenario: Single-line text size
- **WHEN** a TEXT entity with height and string content is measured
- **THEN** the result reports width, height, ascent, descent, and total advance in drawing units

#### Scenario: Laid-out MTEXT
- **WHEN** an MTEXT entity with wrapping width and multiple paragraphs is measured
- **THEN** the result reports the laid-out width and height including paragraph spacing

### Requirement: Native DWG metadata is readable
The system SHALL read DWG version and release metadata and partially parse supported header, class, and object records without invoking an external converter.

#### Scenario: DWG header metadata
- **WHEN** a DWG file with a supported version signature is inspected
- **THEN** the result reports the version, release name, and parsed section inventory

#### Scenario: Unsupported DWG record
- **WHEN** the DWG parser meets an unsupported record encoding
- **THEN** inspection reports the record and offset while returning metadata decoded before it

### Requirement: Export add-ons use decoded geometry
The system SHALL make ACIS body meshes and linked-edge loops available to mesh and edge export add-ons without duplicating geometry logic.

#### Scenario: Export ACIS body mesh
- **WHEN** a drawing containing a supported 3DSOLID is exported to a mesh format
- **THEN** the export includes the triangles produced from the decoded ACIS body

#### Scenario: Export linked-edge loop
- **WHEN** an edge network forms a closed loop and is exported to a polyline
- **THEN** the exported polyline follows the loop order and orientation
