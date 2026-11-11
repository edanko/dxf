# Design

## Context

`table/` has appid, block_record, dimstyle, layer, linetype, lineweight, style,
symboltable, table, tables, table_type, ucs, view, and viewport. `object/` has
only acdbdictionarywdflt, acdbplaceholder, dictionary, group, and xrecord.
Loading discards tables and objects. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**

- Typed, uniform table and object APIs.
- Correct handle/owner graph.
- Round-trip preservation for every standard table and object.

**Non-Goals:**

- Full evaluation of fields, materials, or visual styles.
- Rendering table styles.
- Rewriting the handle allocator.

## Decisions

### Generic table container with typed entries

Keep a generic `Table[T]` container and implement typed entries with a shared
entry base (name, handle, owner, flags, xdata, extension dictionary). This
avoids ten near-identical containers. Alternative: one bespoke container per
table. Rejected as duplication.

### Object registry

Mirror the entity registry: map DXF object type to constructor and loader.
Unknown objects fall back to retained raw objects. This keeps the object model
open and preserves third-party objects.

### Handle/owner wiring as a post-load pass

After all sections are bound, run a resolution pass that sets owners, links
dictionary entries, connects reactors, and validates references. This is simpler
than resolving during parse, where referenced objects may not exist yet.

### Dictionary as the canonical lookup

Model the root dictionary explicitly and resolve keys through it rather than
maintaining parallel maps. This matches the DXF object graph and keeps save
output consistent.

## Risks / Trade-offs

- **Reference cycles** → Resolve iteratively with a visited set; store handles,
  not pointers, in persisted structures.
- **Typed entry drift** → Shared entry base and table-driven tests per table.
- **Unknown object loss** → Raw object fallback plus round-trip tests.
