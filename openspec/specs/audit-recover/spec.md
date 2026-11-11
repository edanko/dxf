# audit-recover Specification

## Purpose
Defines how the library validates a document, reports integrity problems, repairs
safe defects, and recovers usable content from damaged files.

## Requirements

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

### Requirement: Tag-level validation and repair

The audit SHALL validate required tags, value ranges, and structural terminators, and fix mode SHALL repair only defects with a deterministic safe repair.

#### Scenario: Missing terminator is repaired

- **WHEN** a block record lacks its required terminator and audit runs with fix enabled
- **THEN** the terminator is inserted at the correct position and a re-audit reports no remaining terminator issue

#### Scenario: Unsafe tag value is reported

- **WHEN** a numeric field contains a value outside its documented range
- **THEN** the audit reports the record handle, code, and offending value without changing it

### Requirement: Duplicate handles are repaired

Fix mode SHALL assign unique handles to duplicate records deterministically and SHALL update references that unambiguously point to the repaired record.

#### Scenario: Duplicate handle becomes unique

- **WHEN** two entities share a handle and audit runs with fix enabled
- **THEN** one entity receives a new unique handle and the report records the change

#### Scenario: Ambiguous references stay reported

- **WHEN** a duplicate handle is referenced by a record whose intended target is ambiguous
- **THEN** the reference is left unchanged and reported as unresolved

### Requirement: Block cycles are repaired

Fix mode SHALL break block-reference cycles without deleting the referenced block definitions.

#### Scenario: Self-reference is broken

- **WHEN** a block contains an INSERT of itself and audit runs with fix enabled
- **THEN** the self-reference is removed and the block definition remains present

#### Scenario: Nested cycle is reported and broken

- **WHEN** block A references B and B references A
- **THEN** the cycle is reported and at least one reference in the cycle is removed so traversal terminates

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
