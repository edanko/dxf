# Spec Delta

## ADDED Requirements

### Requirement: Collection objects are typed

The system SHALL provide typed models for MLEADERSTYLECOLLECTION, MLINESTYLECOLLECTION, LAYER_FILTER, and VBA_PROJECT, including their documented membership and attributes.

#### Scenario: MLeader style collection

- **WHEN** a drawing contains an MLEADERSTYLECOLLECTION
- **THEN** the collection exposes its member object handles
- **AND** the members resolve through the object database.

#### Scenario: Layer filter

- **WHEN** a drawing contains a LAYER_FILTER
- **THEN** the filter exposes its names and filter expressions
- **AND** unknown filter tags are retained.

### Requirement: Object lookup remains stable

The system SHALL resolve newly typed collection objects by handle and by dictionary key without changing existing handle or owner values.

#### Scenario: Resolve by handle

- **WHEN** a caller requests a collection object by its handle
- **THEN** the typed object is returned
- **AND** its owner handle is unchanged.
