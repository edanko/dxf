# Proposal

## Why

The CLI covers the main document workflows but lacks the developer tools that make ezdxf practical for debugging and standards enforcement: analyze, difftags, debug, standards, and ZIP packaging.

## What Changes

- Add `dxf analyze` for tag/entity/object statistics.
- Add `dxf difftags` for structural tag diffs between two files.
- Add `dxf debug` for loader diagnostics.
- Add standards validation with configurable rule sets.
- Add ZIP manager and `dxf zip`/`dxf unzip` commands.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-tooling`: analyze, difftags, debug, standards, zip commands.
- `addons`: standards rule engine and ZIP manager.

## Impact

Affected packages: `cmd/dxf`, new `standards`, `filemanagement` ZIP support. CLI output formats remain text+JSON.
