# Proposal

## Why

ezdxf ships a command-line interface for inspecting, auditing, converting,
rendering, and querying drawings. The Go library has no CLI, so users must write
throwaway programs for routine operations. A first-class CLI also provides a
stable integration surface for scripts and CI.

## What Changes

- Add a CLI binary with subcommands for audit, info, convert, draw, query,
  explode, fix, strip, header, layers, and browse.
- Support text and JSON output, stable exit codes, and quiet/verbose modes.
- Add configuration and application settings with file and environment
  overrides.
- Add file management helpers: path resolution, backups, and atomic writes.
- Add comment handling options.
- Add shell completion and a version command.
- Add CLI integration tests that run the binary against fixtures.

## Capabilities

### New Capabilities

- `cli-tooling`: the command-line interface, its commands, output contracts,
  configuration, and exit codes.

### Modified Capabilities

None.

## Impact

Affected code: a new `cmd/dxf` (or equivalent) package plus small public helpers
in `appsettings`, `filemanagement`, and `comments`. No breaking changes to the
library. Tests execute the built binary. Depends on the audit, query, render, and
version changes.
