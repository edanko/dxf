# Design

## Context

`audit/` has placeholder file loading and no integrity checks. The document I/O
change adds complete binding and a recovery load mode. Handles, tables, objects,
and entities exist but are not cross-validated. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- Comprehensive, structured, testable validation.
- Safe automatic repair for well-understood defects.
- Recovery that never panics and always reports.

**Non-Goals:**

- Reconstructing arbitrary lost geometry.
- Semantic validation of ACIS payloads.
- Auto-correcting ambiguous references.

## Decisions

### Check registry

Implement each check as a function registered by code and severity. The auditor
runs checks over the document and collects issues. This makes checks independent
and easy to test. Alternative: one monolithic audit function. Rejected for
maintainability.

### Handles as the reference graph

Validate references by resolving handles through a document-wide handle map built
once. This catches duplicates and dangling links in one pass.

### Fixes are explicit and conservative

Each check may register a fixer that returns whether it changed the document.
Fix mode applies only fixers marked safe, then re-runs the audit. Unsafe issues
remain reported. This avoids destructive guessing.

### Recovery reuses the tolerant loader

Recovery is the loader's `Recover` mode plus an audit pass. The loader records
structural repairs; the audit records semantic issues. One report merges both.

### Entity hooks via an interface

Entities that can validate themselves implement `Audit(ctx) []Issue`. The
document audit invokes the hook and adds document-level checks.

### Strict R12 as a profile check

Reuse the version profile from the version change to flag unsupported sections,
tables, and fields.

## Risks / Trade-offs

- **False positives** → Severity levels and documented codes; tests for valid
  documents must be clean.
- **Fix side effects** → Safe fixers only, with re-audit and tests.
- **Performance on large files** → Build the handle map once and avoid repeated
  scans.
- **Report noise** → Aggregate repeated issues by code and object.
