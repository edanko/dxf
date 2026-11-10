# Spec Delta

## ADDED Requirements

### Requirement: Analyze command reports statistics

The system SHALL provide `dxf analyze` that reports tag, entity, table, block, and object statistics in text or JSON.

#### Scenario: Analyze a fixture

- **WHEN** `dxf analyze FILE --json` runs
- **THEN** it exits 0 and emits counts and section summaries as valid JSON.

### Requirement: DiffTags compares structures

The system SHALL provide `dxf difftags` that reports added, removed, and changed tag paths between two DXF files.

#### Scenario: Differing files

- **WHEN** two files differ in an entity layer
- **THEN** the diff reports the entity path and the old/new layer values
- **AND** exits non-zero when differences exist.

### Requirement: Debug command reports diagnostics

The system SHALL provide `dxf debug` that reports detected version, encoding, section sizes, loader fallbacks, and recovery issues.

#### Scenario: Debug a damaged file

- **WHEN** `dxf debug` runs on a file with a truncated section
- **THEN** the report includes the recovery issue and still exits successfully.

### Requirement: Standards command validates rules

The system SHALL provide `dxf standards` with configurable rule sets and non-zero exit on violations.

#### Scenario: Layer name violation

- **WHEN** a drawing violates a configured layer-name rule
- **THEN** the command reports the entity/table path and exits non-zero.

### Requirement: ZIP commands package drawings

The system SHALL provide `dxf zip` and `dxf unzip` for DXF archives with atomic output behavior.

#### Scenario: Round trip an archive

- **WHEN** a DXF is zipped and then unzipped
- **THEN** the extracted file is byte-identical to the input.
