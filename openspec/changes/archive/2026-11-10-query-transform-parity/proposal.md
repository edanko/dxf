# Proposal

## Why

Users need to find and manipulate entities after loading: filter by type, layer,
attributes, or location; transform selections; explode composites; and reorder
draw order. The `query` package has an expression lexer/parser, but it is not
integrated with a document-wide selection API, and transform/explode/reorder
operations are incomplete. This blocks scripted editing workflows that ezdxf
supports through `query`, `select`, `transform`, `explode`, `reorder`, and
related modules.

## What Changes

- Define a query expression language over entity attributes, xdata, and geometry.
- Add a document query API returning entities and handles.
- Add grouping and aggregation by attribute.
- Add selection sets that can be stored, combined, and transformed.
- Add document and per-entity transform operations.
- Add explode and disassemble operations.
- Add draw-order manipulation.
- Add upright text normalization.
- Add zoom/extents helpers.
- Add block-reference traversal.
- Add revcloud creation.
- Add proxy-graphic extraction.

## Capabilities

### New Capabilities

- `query-transform`: querying, selecting, grouping, transforming, exploding, and
  reordering document content.

### Modified Capabilities

None.

## Impact

Affected code: `query/*`, `drawing/*`, `entity/*`, `math/*`, `block/*`,
`spatial/*`, `render/*`. New public query/selection API. Tests cover expression
evaluation, spatial filters, transform pipelines, explode output, and ordering.
Depends on the entity model and graphics factory changes.
