# Spec Delta

## ADDED Requirements

### Requirement: Standards rule engine is reusable

The system SHALL expose a standards rule engine independent of the CLI, with structured violation reports and configurable severities.

#### Scenario: Programmatic rule run

- **WHEN** a caller runs the standards engine on a drawing
- **THEN** it receives violations with code, severity, and object handle.

### Requirement: ZIP manager preserves entries

The system SHALL create and extract ZIP archives while preserving entry names and file contents.

#### Scenario: Archive extraction

- **WHEN** an archive is extracted
- **THEN** each entry is written at its recorded path
- **AND** path traversal outside the target directory is rejected.
