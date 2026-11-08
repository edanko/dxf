# Tasks

## 1. Audit foundation

- [ ] 1.1 Define issue, severity, and report types; verify serialization and counts
- [ ] 1.2 Build a document-wide handle map and verify duplicate/dangling detection
- [ ] 1.3 Add the check registry and run checks over a clean document; verify zero errors

## 2. Checks

- [ ] 2.1 Implement handle, owner, and orphan checks; verify corrupted fixtures
- [ ] 2.2 Implement table, block-record, layer, linetype, style, and layout reference checks; verify missing-reference fixtures
- [ ] 2.3 Implement xdata, reactor, and extension-dictionary checks; verify malformed links
- [ ] 2.4 Implement extents and entity invariant hooks; verify invalid spline fixture

## 3. Fix and recover

- [ ] 3.1 Implement safe fixers with re-audit; verify missing-owner and missing-default-entry repairs
- [ ] 3.2 Merge loader recovery with audit reporting; verify truncated and malformed fixtures
- [ ] 3.3 Implement strict R12 audit using the version profile; verify incompatible content is reported

## 4. Verification

- [ ] 4.1 Add corrupted fixtures and fuzz tests; verify no panic and bounded reports
- [ ] 4.2 Update `docs/parity/README.md` audit/recover matrix

## Workflow follow-up

- Archive the change after review.
