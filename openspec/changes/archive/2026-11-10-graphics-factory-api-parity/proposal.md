# Proposal

## Why

The current API is mostly a parser plus per-package constructors. Users cannot
open a document and add entities the way ezdxf allows (`msp.add_line(...)`,
`blocks.new(...)`, `layouts.new(...)`), and there is no coherent object for
model space, paper space, blocks, layouts, groups, or the entity database. This
ergonomic gap is the main reason the Go library cannot replace ezdxf in
authoring workflows.

## What Changes

- Add a high-level `Drawing` facade with `ModelSpace`, `PaperSpace`, `Blocks`,
  `Layouts`, `Groups`, `Objects`, and `EntityDB` accessors.
- Add a graphics factory with `Add<Type>` methods for every supported entity,
  returning the created entity.
- Add block definition creation, block reference insertion, and attribute
  handling.
- Add layout creation, rename, delete, and active-layout selection.
- Add handle allocation, owner assignment, and entity registration in the
  database.
- Add group creation and membership management.
- Add selection helpers by type, layer, and spatial extent.
- Add document extents and bounding-box recalculation.
- Add optional transaction/undo support for authoring operations.

## Capabilities

### New Capabilities

- `graphics-factory`: high-level document authoring API, factories, layouts,
  blocks, groups, and entity database behavior.

### Modified Capabilities

None.

## Impact

Affected code: `drawing/*`, `factory/*`, `layouts/*`, `block/*`, `query/*`,
`handle/*`, `entity/*`. New public API; existing constructors remain. Tests
cover authoring workflows, handle allocation, block/layout creation, and
save/reload. Depends on the document I/O and entity model changes.
