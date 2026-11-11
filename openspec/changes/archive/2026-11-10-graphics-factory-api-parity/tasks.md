# Tasks

## 1. Document facade

- [x] 1.1 Add `ModelSpace`, `PaperSpace`, `Blocks`, `Layouts`, `Groups`, `Objects`, and `EntityDB` accessors; verify with a new drawing
- [x] 1.2 Model spaces as block records and verify owner handles on save
- [x] 1.3 Implement automatic handle allocation and registration; verify uniqueness across 100 entities

## 2. Factory

- [x] 2.1 Add a generic add helper with options and verify handle/owner assignment
- [x] 2.2 Add `Add<Type>` methods for all supported entities and verify each creates the right type
- [x] 2.3 Add attribute handling for INSERT and verify values are accessible

## 3. Blocks and layouts

- [x] 3.1 Implement block definition creation and entity addition; verify block content round trip
- [x] 3.2 Implement INSERT creation with position, scale, rotation, and attributes; verify bbox and explode
- [x] 3.3 Implement layout create/rename/delete/activate with block-record linkage; verify save/reload

## 4. Groups, selection, extents

- [x] 4.1 Implement group creation and membership; verify both-sided references
- [x] 4.2 Implement selection by type, layer, and extent using the query engine; verify result sets
- [x] 4.3 Implement model-space and per-layout extents; verify recalculation after edits

## 5. Transactions

- [x] 5.1 Implement transaction begin/commit/rollback for add, modify, and delete; verify rollback restores state

## 6. Verification

- [x] 6.1 Add an end-to-end authoring test (new drawing, block, layout, entities, save, reload) and verify equivalence
- [x] 6.2 Update `docs/parity/README.md` authoring API matrix

## Workflow follow-up

- Archive the change after review.
