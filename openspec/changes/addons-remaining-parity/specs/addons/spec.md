# Spec Delta

## ADDED Requirements

### Requirement: Browser data extraction is available

The system SHALL expose document, entity, and ACIS inspection data in a form usable by a GUI or CLI browser without requiring a GUI dependency.

#### Scenario: Entity browser data

- **WHEN** a caller requests browser data for a document
- **THEN** entities are grouped by type and layout
- **AND** each entry exposes handle, type, layer, and bounding box.

### Requirement: DWG conversion delegates to an external tool

The system SHALL convert DWG to DXF through a configurable external converter and SHALL return an actionable error when the tool is missing.

#### Scenario: Missing converter

- **WHEN** DWG conversion is requested without a configured converter
- **THEN** the operation fails with an error naming the required tool
- **AND** no partial output file is left behind.

### Requirement: Geometry generator add-ons are available

The system SHALL provide dimension-line helpers, r12export, xqt, genetic algorithm, Menger sponge, and Sierpinski pyramid add-ons with documented inputs and outputs.

#### Scenario: Menger sponge generation

- **WHEN** a Menger sponge is generated at a given depth
- **THEN** the expected number of cube/face entities is produced
- **AND** the result can be added to a drawing.

### Requirement: Mixins match ezdxf behavior

The system SHALL provide the documented mixin behaviors for add-on entities and objects.

#### Scenario: Mixin application

- **WHEN** a mixin is applied to a compatible entity
- **THEN** the documented attributes and methods are available
- **AND** existing entity data is unchanged.
