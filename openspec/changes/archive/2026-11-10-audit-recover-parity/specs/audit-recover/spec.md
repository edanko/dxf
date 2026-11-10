# Spec Delta

## Purpose

Defines how the library validates a document, reports integrity problems, repairs
safe defects, and recovers usable content from damaged files.

## ADDED Requirements

### Requirement: Document audit

The library SHALL audit a loaded document and report problems with handles,
ownership, tables, blocks, references, xdata, reactors, extension dictionaries,
and extents.

#### Scenario: Duplicate handle

- **WHEN** two objects share a handle
- **THEN** the audit reports a duplicate-handle error naming both objects

#### Scenario: Missing owner

- **WHEN** an entity's owner handle does not resolve
- **THEN** the audit reports an invalid-owner error

#### Scenario: Clean document

- **WHEN** a valid document is audited
- **THEN** the report contains no errors and the document is unchanged

### Requirement: Structured report

The audit SHALL return a report of issues, each with severity, code, object
handle where available, and a human-readable message.

#### Scenario: Report contents

- **WHEN** an audit finds problems
- **THEN** every issue has a severity and message and the report exposes counts
  by severity

### Requirement: Fix mode

The audit SHALL support a fix mode that repairs safe issues, leaves unsafe issues
reported, and re-audits after repair.

#### Scenario: Fix missing owner

- **WHEN** fix mode encounters an entity with a missing owner
- **THEN** it assigns the documented default owner and the re-audit no longer
  reports that issue

#### Scenario: Unsafe issue is not fixed

- **WHEN** fix mode encounters an issue it cannot repair safely
- **THEN** the issue remains in the report and the document is otherwise
  unchanged

### Requirement: Reference integrity

The audit SHALL verify that layer, linetype, text style, dimension style, block
record, and layout references resolve, and SHALL report unresolved references.

#### Scenario: Missing layer

- **WHEN** an entity references a layer that does not exist
- **THEN** the audit reports a missing-layer reference

### Requirement: Table integrity

The audit SHALL verify table entry names, uniqueness, required default entries,
and owner handles.

#### Scenario: Missing required table entry

- **WHEN** a required default entry such as layer `0` is absent
- **THEN** the audit reports the missing entry

### Requirement: Recovery mode

Recovery SHALL load usable content from damaged files without panicking and
SHALL report every repair or skipped structure.

#### Scenario: Corrupt entity

- **WHEN** a file contains a malformed entity record
- **THEN** recovery skips that entity, loads the rest, and reports the skip

#### Scenario: Truncated section

- **WHEN** a file ends before a section terminator
- **THEN** recovery closes the section, retains prior content, and reports it

### Requirement: Entity audit hooks

Entity types SHALL be able to validate their own invariants, and the document
audit SHALL invoke those hooks.

#### Scenario: Invalid spline

- **WHEN** a spline has inconsistent degree and control-point counts
- **THEN** its entity audit reports the inconsistency

### Requirement: Strict R12 audit

The audit SHALL support a strict R12 mode that reports content not representable
in R12.

#### Scenario: R12 incompatibility

- **WHEN** a document containing OBJECTS is audited in strict R12 mode
- **THEN** the report lists the incompatible content
