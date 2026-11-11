# Proposal

## Why

The `audit` package currently has placeholder file loading and does not verify
document integrity. Real DXF files often contain dangling handles, orphaned
entities, missing table entries, and broken references; ezdxf's audit and recover
facilities are essential for trusting input and repairing output. Without them,
load failures surface as opaque parse errors and damaged files cannot be salvaged.

## What Changes

- Implement a document audit that checks handles, ownership, tables, blocks,
  references, xdata, reactors, extension dictionaries, and extents.
- Return a structured report with severity, location, and message.
- Add a fix mode that repairs safe issues and re-audits.
- Add a recover mode that loads damaged files and reports repairs.
- Add per-entity audit hooks.
- Add strict R12 conformance checks.
- Add a CLI-facing audit entry point (implementation belongs to the CLI change).
- Add corrupted-fixture tests and fuzz tests.

## Capabilities

### New Capabilities

- `audit-recover`: document and entity validation, structured reports, repair,
  and recovery behavior.

### Modified Capabilities

None.

## Impact

Affected code: `audit/*`, `drawing/*`, `handle/*`, `table/*`, `object/*`,
`entity/*`, and load entry points. New public audit API and report types. Tests
use deliberately corrupted fixtures. Depends on document I/O and the table/object
model.
