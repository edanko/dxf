# Spec Delta

## Purpose

Defines the command-line interface for inspecting, validating, converting,
rendering, and querying DXF files, including its output and exit-code contracts.

## ADDED Requirements

### Requirement: Command structure

The CLI SHALL provide subcommands for audit, info, convert, draw, query, explode,
fix, strip, header, layers, browse, and version, and SHALL provide help for each.

#### Scenario: Help

- **WHEN** the CLI is invoked with `--help`
- **THEN** it lists the available subcommands and exits successfully

#### Scenario: Unknown command

- **WHEN** an unknown subcommand is invoked
- **THEN** the CLI prints an error and exits with a usage error code

### Requirement: Info command

The info command SHALL report the DXF version, encoding, entity counts by type,
table counts, block count, and object count.

#### Scenario: Info output

- **WHEN** info is run on a fixture
- **THEN** the output includes the version and accurate counts

### Requirement: Audit command

The audit command SHALL run the document audit, print issues, support fix mode,
and exit non-zero when errors remain.

#### Scenario: Audit clean file

- **WHEN** audit runs on a valid file
- **THEN** it reports no errors and exits zero

#### Scenario: Audit with fix

- **WHEN** audit `--fix` runs on a repairable file
- **THEN** it writes the repaired file and reports the fixes

### Requirement: Convert command

The convert command SHALL save a document to a requested version and encoding.

#### Scenario: Convert version

- **WHEN** convert is run with a target version
- **THEN** the output file declares that version and loads successfully

### Requirement: Draw command

The draw command SHALL render a document or layout to SVG, PDF, or PNG with
configurable size and background.

#### Scenario: Render to PNG

- **WHEN** draw is run with PNG output
- **THEN** the output file is a valid image of the requested size

### Requirement: Query command

The query command SHALL evaluate a query expression and print matching entities
in text or JSON.

#### Scenario: JSON query

- **WHEN** query is run with JSON output
- **THEN** the output is valid JSON containing the matched entities

### Requirement: Explode and strip commands

The explode command SHALL write a document with composite entities exploded; the
strip command SHALL write a document without comments or optionally without
proxy content.

#### Scenario: Strip comments

- **WHEN** strip is run on a file with comments
- **THEN** the output contains no code 999 comments

### Requirement: Header and layers commands

The header command SHALL list and set header variables; the layers command SHALL
list layers and their properties.

#### Scenario: List layers

- **WHEN** layers is run on a fixture
- **THEN** each layer name and its color and linetype are printed

### Requirement: Output formats and exit codes

Commands SHALL support text and JSON output where applicable and SHALL use
documented exit codes: zero for success, one for data errors, two for usage
errors, and three for missing external resources.

#### Scenario: Data error exit code

- **WHEN** a command encounters an invalid input file
- **THEN** it exits with code one and prints an actionable message

### Requirement: Configuration

The CLI SHALL read configuration from a config file and environment variables,
with command-line flags taking precedence.

#### Scenario: Precedence

- **WHEN** a value is set in the config file and overridden by a flag
- **THEN** the flag value is used

### Requirement: Atomic writes and backups

Commands that modify files SHALL write atomically and SHALL create a backup when
configured.

#### Scenario: Backup on fix

- **WHEN** a modifying command runs with backups enabled
- **THEN** the original file remains available as a backup

### Requirement: Shell completion and version

The CLI SHALL provide shell completion generation and a version command that
reports the library and build versions.

#### Scenario: Version output

- **WHEN** version is run
- **THEN** it prints a parseable version string
