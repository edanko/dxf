# Design

## Context

`drawing.Drawing` already stores entities and has an `EntityDatabase`, spatial
index, and batch processor, but lacks space/layout/block facades. `factory/`
has a registry. `layouts/` exists. Handles are managed by `handle/`. The parser
is the only significant authoring entry point today. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- ezdxf-like ergonomics without breaking existing constructors.
- Automatic handle/owner management.
- One consistent document graph for spaces, blocks, and layouts.

**Non-Goals:**

- A full undo stack for every mutation (transaction support is scoped to
  authoring operations).
- Rewriting the spatial index.
- GUI or rendering integration.

## Decisions

### Spaces are block records

Model model space and paper space as block records, matching DXF. The facade's
`ModelSpace` and `PaperSpace` return entity containers backed by those records.
This keeps save output correct and avoids a parallel representation.

### Factory as a thin typed layer

Each `Add<Type>` constructs the entity, applies options, registers it with the
database, assigns handle/owner, and appends to the target container. Shared
logic lives in one generic `add` helper. Alternative: methods on each entity.
Rejected because ownership and registration would be duplicated.

### Layout object owns its block record

A layout holds its name, block record handle, plot settings, and tab order. The
layout manager keeps the model layout and paper layouts in one ordered list.

### Transactions as an operation log

A transaction records inverse operations (added handles, prior field values)
and applies them in reverse on rollback. This is simpler than snapshotting the
whole document and sufficient for authoring. Alternative: copy-on-write document.
Rejected for memory cost.

### Selection delegates to the query engine

Selection helpers build query expressions and run them against the entity
database rather than duplicating filtering logic.

## Risks / Trade-offs

- **API surface growth** → Generate `Add<Type>` methods from the entity registry
  where possible; keep options functional.
- **Rollback correctness** → Test rollback for add, modify, and delete.
- **Owner bugs** → Centralize assignment in the add helper and assert in tests.
- **Layout/block desync** → Create and delete them together in one operation.
