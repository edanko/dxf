# Design

## Context

The loader now builds typed entities, tables, blocks, and objects, and the corpus round-trip suite covers all fixtures. Fidelity still depends on what each typed model chooses to emit: some table entries, blocks, and typed objects drop unhandled tags, comments are re-emitted at section starts, duplicate handles are only reported, and block-cycle repair is absent. Attribute coverage is asserted by type inventory and selected field tests, not by a per-record conformance matrix. See `proposal.md` for the motivating gaps.

Constraints:

- Existing public APIs and saved-file compatibility must remain stable.
- Unknown/third-party records must continue to round-trip.
- Repair must be explicit and reported; no silent data loss.
- Tests must remain runnable without Python; the ezdxf cross-check stays opt-in.

## Goals / Non-Goals

**Goals:**

- Make every typed record preserve tags it does not model and emit them deterministically.
- Preserve comment order relative to section content.
- Define and test a per-entity attribute, default, version-gate, copy, and bounding-box matrix.
- Add tag-level validation and repair plus duplicate-handle and block-cycle repair.
- Extend the cross-check to compare typed attributes, not only inventories.

**Non-Goals:**

- ACIS geometry parsing, new rendering features, new add-ons, and API redesign.
- Byte-for-byte output equality with ezdxf.
- Repairing arbitrary semantic corruption without a documented rule.

## Decisions

### Record-level retention ledger

Each typed record keeps an ordered ledger of source tags and a set of codes it emitted. Formatting emits canonical modeled tags, then unhandled source tags in their original relative order. The ledger is shared by entities, table entries, blocks, and objects.

Alternative: append only unknown codes. Rejected because repeated codes and their order convey meaning for some records.

### Comments as positioned tokens

The section loader records comments with the index of the preceding content token instead of a section-local list. The writer replays comments before the token they preceded, and strip mode removes them. This preserves order without changing the public comments map, which becomes a derived view.

### Attribute conformance matrix

A table in the package docs lists, per entity type, the ezdxf reference file, modeled group codes, defaults, version gates, and tests. A generated golden-tag test per type asserts wire output for representative instances and reload equality. A missing row or missing test is a test failure, not an undocumented omission.

Alternative: rely on corpus round-trip. Rejected because corpus files do not exercise every field or version gate.

### Validation before binding

Tag-level validation runs on the grouped section tags before typed binding: it checks structural terminators, numeric ranges, required group codes, and duplicate handles. Repair applies deterministic transformations and records each action. Binding then sees repaired or original tags and reports skipped records.

### Duplicate handles and block cycles

Duplicate handles are resolved by assigning a fresh handle to later duplicates. References are rewritten only when exactly one referrer resolves to the duplicate; ambiguous references remain reported. Block cycles are found with a depth-first traversal of INSERT definitions; the edge that closes a cycle is removed from the block definition and reported. The block itself is retained.

### Bounding-box completeness

Every entity provides a bounding box through its concrete geometry. Proxy entities use decoded proxy geometry when available. Entities with no drawable geometry return the documented empty representation; query and extents treat that representation as no geometry rather than a zero-origin box.

### Cross-check typed fields

The opt-in harness compares version, inventory, and a stable typed-field projection per entity, table entry, and object. The projection is defined next to each model and avoids handles, generated defaults, and ordering that ezdxf does not guarantee.

## Risks / Trade-offs

- [Raw-tag ledgers increase memory for large files] → Store only unhandled tags per record and share immutable slices where possible.
- [Changed tag order can alter downstream consumers] → Preserve relative order of unhandled tags and add golden tests for order-sensitive records.
- [Repair can hide corruption or corrupt references] → Report every action and refuse ambiguous rewrites.
- [Attribute matrix becomes stale as ezdxf evolves] → Generate the matrix from the pinned reference source and fail tests when a type lacks a row.
- [Extended cross-check is slow] → Keep it opt-in and run it on a representative corpus in CI.

## Migration Plan

No public API removal is planned. Existing tolerated losses become preserved tags, which may change saved output but not semantic content. Repair remains opt-in through the existing fix paths.
