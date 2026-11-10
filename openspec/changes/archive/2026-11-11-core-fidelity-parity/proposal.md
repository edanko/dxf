# Proposal

## Why

The roadmap changes delivered broad feature coverage, but parity is still asserted mainly through inventory-level checks. Attribute-level fidelity, unknown-tag retention, repair coverage, and bounding boxes remain unproven, so a document can pass the current cross-check while losing data or reporting incorrect geometry.

## What Changes

- Add an executable per-record conformance matrix covering every typed entity, table entry, and object: documented group codes, defaults, subclass markers, version gating, XDATA/reactors, extension dictionaries, copy/clone semantics, and save/reload stability.
- Preserve all unhandled tags on typed table entries, blocks, and objects, including their original order relative to emitted tags.
- Preserve comment records at their original section position instead of re-emitting them at the start of the section.
- Complete tag-level validation and repair: malformed value ranges, missing required terminators, invalid owners, and recoverable structural defects.
- Repair duplicate handles deterministically and add block-reference cycle repair, leaving unsafe cases reported.
- Make bounding boxes complete for placeholder, proxy, underlay, table, and other currently degenerate entity types.
- Extend the ezdxf cross-check from inventories to semantic attributes and round-trip equivalence, with opt-in golden dumps for representative records.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dxf-io-core`: unknown-tag retention and comment-position fidelity across load/save.
- `entity-model`: per-entity attribute completeness, defaults, copy semantics, proxy/placeholder bounding boxes.
- `table-object-model`: complete tag retention and typed attribute coverage for table entries and objects.
- `audit-recover`: tag-level validator/repair, duplicate-handle repair, and block-cycle repair.
- `query-transform`: bounding boxes for every supported entity category, including placeholder and proxy records.

## Impact

Affected code: `entity`, `table`, `object`, `block`, `lldxf`, `audit`, `query`, `drawing`, and their tests. The change may alter saved output where previously dropped tags are now retained and where comments move to their original positions. The semantic cross-check may fail documents that currently pass inventory-only checks; those failures are the intended outcome until fixed. No public API removal is planned.
