# Design

## Context

`lldxf.Loader` already groups tags by section and exposes entities, tables,
blocks, and objects. `dxf.FromReader` uses the loader only to build ENTITIES;
the rest is discarded. `drawing.Drawing` has section containers, but no binder
connects loaded tags to them. `parser.go` contains unused per-section helpers
and a no-op OBJECTS parser. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**

- One load path that produces a complete `drawing.Drawing`.
- Lossless retention of unknown content.
- Deterministic, testable round trips.
- A recovery mode that degrades gracefully.

**Non-Goals:**

- Full semantic modelling of every entity/object (that is the entity and
  table/object changes).
- Version conversion (separate change).
- Byte-for-byte output equality with ezdxf.

## Decisions

### Two-phase load: parse then bind

Keep `lldxf` as the tag-level parser and add a document binder that walks the
loader output and populates `drawing.Drawing`. This preserves the existing
low-level API and keeps section parsing testable in isolation. Alternative:
make each entity parse itself from tags. Rejected for now because it would
duplicate tag handling across dozens of types before the entity model is ready.

### Raw tag retention on entities

Extend the common entity base with an ordered list of unhandled tags and a raw
object-tag store for unmodelled objects. `Format` emits known tags first and raw
tags afterwards, preserving order within each group. This matches ezdxf's
"ignore but retain" behavior. Alternative: store the entire original tag stream
per entity. Rejected because it makes edits and duplicate-tag handling
ambiguous.

### Section registry with opaque fallback

Represent sections through a small registry. Known sections bind to typed
containers; unknown or opaque sections (THUMBNAILIMAGE, ACDSDATA) store their
raw tags and are re-emitted. This allows new sections without changing the
writer.

### Recovery as an option, not a separate parser

Add `LoadOptions{Recover bool}` to the loader and reuse the same tagger and
binder. Recovery synthesizes missing ENDSEC/EOF markers, skips malformed
entities, and records issues in a report. Alternative: separate tolerant parser.
Rejected because two parsers drift.

### Comments as first-class tags

Keep code 999 tags attached to their section and emit them unless
`StripComments` is set. This avoids reordering comments relative to entities.

## Risks / Trade-offs

- **Raw tag duplication** → Keep a single source of truth: parsed fields are
  authoritative, raw tags are only for codes the model does not consume.
- **Round-trip ordering changes** → Compare semantic equality in tests, not
  bytes; document any unavoidable ordering differences.
- **Recovery masking corruption** → Always return a report and never silently
  drop content without an entry.
- **Performance regression from retaining tags** → Measure on the largest
  fixture; raw tags are bounded by input size.
