# Tasks

## 1. Baseline and scaffolding

- [x] 1.1 Add a round-trip test helper that loads, saves, reloads, and compares entity/table/block/object inventories; verify it fails on the current entities-only load path
- [x] 1.2 Add `LoadOptions` with `Recover` and `StripComments` and verify option plumbing with unit tests

## 2. Section binding

- [x] 2.1 Implement HEADER binding into `drawing.Drawing` and verify header variables survive a round trip
- [x] 2.2 Implement TABLES binding for existing table entry types and verify layer/linetype/style/block-record counts match fixtures
- [x] 2.3 Implement BLOCKS binding and verify block names and contained entities match fixtures
- [x] 2.4 Implement CLASSES binding and verify class records survive a round trip
- [x] 2.5 Implement OBJECTS binding for dictionary, group, xrecord, and placeholder types; verify retrieval by handle
- [x] 2.6 Retain THUMBNAILIMAGE and ACDSDATA as opaque sections and verify they are re-emitted

## 3. Preservation

- [x] 3.1 Add raw-tag storage to the common entity base and emit unhandled tags in `Format`; verify unknown entity and unknown group-code round trips
- [x] 3.2 Add raw-tag storage for unmodelled objects and verify unknown object round trips
- [x] 3.3 Preserve code 999 comments per section and verify save keeps them unless stripping is enabled

## 4. Recovery

- [x] 4.1 Implement recovery for missing ENDSEC/EOF and truncated entities with a structured report; verify with malformed fixtures
- [x] 4.2 Add fuzz tests for the tagger and loader and verify no panic or unbounded memory on malformed input

## 5. Writer correctness

- [x] 5.1 Enforce DXF section order and unique handles in the writer; verify with a newly created document
- [x] 5.2 Run the full round-trip suite over `testdata/` and record any remaining semantic differences in `docs/parity/README.md`

## Workflow follow-up

- Update `docs/parity/README.md` status matrix for document I/O.
- Archive the change after review.
