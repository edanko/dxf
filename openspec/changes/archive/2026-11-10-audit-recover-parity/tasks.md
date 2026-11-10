# Tasks

## 1. Audit foundation

- [x] 1.1 Define issue, severity, and report types; verify serialization and counts
- [x] 1.2 Build a document-wide handle map and verify duplicate/dangling detection
- [x] 1.3 Add the check registry and run checks over a clean document; verify zero errors

## 2. Checks

- [x] 2.1 Implement handle, owner, and orphan checks; verify corrupted fixtures
- [x] 2.2 Implement table, block-record, layer, linetype, style, and layout reference checks; verify missing-reference fixtures
- [x] 2.3 Implement xdata, reactor, and extension-dictionary checks; verify malformed links
- [x] 2.4 Implement extents and entity invariant hooks; verify invalid spline fixture

## 3. Fix and recover

- [x] 3.1 Implement safe fixers with re-audit; verify missing-owner and missing-default-entry repairs
- [x] 3.2 Merge loader recovery with audit reporting; verify truncated and malformed fixtures
- [x] 3.3 Implement strict R12 audit using the version profile; verify incompatible content is reported

## 4. Verification

- [x] 4.1 Add corrupted fixtures and fuzz tests; verify no panic and bounded reports
- [x] 4.2 Update `docs/parity/README.md` audit/recover matrix

## Workflow follow-up

- Archive the change after review.
