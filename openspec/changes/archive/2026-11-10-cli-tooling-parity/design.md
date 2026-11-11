# Design

## Context

No CLI exists. The library exposes loading, audit, query, render, and save APIs
that can back commands. `appsettings`, `filemanagement`, and `comments` modules
from ezdxf have no Go equivalent. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- A scriptable, stable CLI with machine-readable output.
- Thin commands over library APIs.
- Safe file modification.

**Non-Goals:**

- A GUI or interactive editor.
- Reproducing every ezdxf CLI flag.
- Long-running server mode.

## Decisions

### Use a standard CLI framework

Use a mature Go CLI library (cobra or equivalent) for subcommands, flags, help,
and completion. This avoids hand-rolled parsing and gives consistent UX.
Alternative: `flag` only. Rejected because nested subcommands and completion are
painful.

### Commands are thin adapters

Each command parses flags, calls a library function, and formats output. Business
logic stays in packages so it is testable without the binary. Alternative: logic
in commands. Rejected for reuse.

### Output writer abstraction

Commands write through an output interface with text and JSON implementations.
Errors go to stderr; data to stdout. Exit codes map to typed errors.

### Config precedence: flags > env > file > defaults

Implement a single settings resolver and test precedence explicitly. Use a
documented config path and `DXF_` environment prefix.

### Atomic writes with optional backup

Write to a temp file in the destination directory, fsync, rename, and optionally
copy the original to a backup path first. This prevents partial output.

## Risks / Trade-offs

- **Flag surface growth** → Keep commands focused; document flags in help.
- **JSON schema stability** → Version the JSON output and test it.
- **Cross-platform paths** → Use `filepath` and test on Linux; document Windows
  caveats.
- **Destructive commands** → Require explicit output paths and default to no
  in-place modification.
