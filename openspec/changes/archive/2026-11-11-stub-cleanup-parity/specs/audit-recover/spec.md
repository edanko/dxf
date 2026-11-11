# Spec Delta

## ADDED Requirements

### Requirement: File audit loads the real document
The system SHALL audit a file through the same loader used by normal file input and report loading failures as errors.

#### Scenario: Audit a valid file
- **WHEN** an existing readable DXF file is audited
- **THEN** the result reflects the loaded document's entities, tables, blocks, and objects

#### Scenario: Missing file
- **WHEN** the file path does not exist
- **THEN** the audit returns an error identifying the path and does not return a successful empty report

#### Scenario: Malformed file
- **WHEN** the file cannot be loaded in strict mode
- **THEN** the audit returns the loader error with the recovery report when recovery data is available
