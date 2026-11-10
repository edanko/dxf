# Design

## Context

`query/` has `EntityQuery`, a lexer, and a parser for expressions but no
document-wide execution API. `spatial/` has a placeholder bounding-box comment.
Transform, explode, and reorder exist only partially in `entity/` and `drawing/`.
See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- One query engine used by selection, grouping, and CLI.
- Reusable selections and composable operations.
- Safe traversal of nested and cyclic references.

**Non-Goals:**

- SQL or a general-purpose scripting language.
- Undo for every transform (transactions from the graphics factory cover
  authoring).
- Rendering proxy graphics beyond extracting their entities.

## Decisions

### AST plus compiled predicate

Parse expressions into an AST and compile to a predicate over an entity context
(type, layer, attributes, xdata, bbox). This allows both evaluation and
introspection. Alternative: evaluate directly during parse. Rejected because
grouping and diagnostics need the tree.

### Query execution over the entity database

The document builds the candidate set from its entity database and applies the
predicate, optionally pre-filtering with the spatial index. This keeps query
independent of storage.

### Selections are handle sets

A selection stores entity handles, not pointers, and resolves on use. This keeps
selections valid across edits and serializable. Set operations are handle-set
operations.

### Transform pipeline

Implement transform as a visitor over entities using the entity capability
interfaces. Document and block transforms reuse the same visitor. This avoids
duplicating per-container logic.

### Disassemble with a visited stack

Traverse inserts recursively, applying the accumulated transform, and track the
block stack to detect cycles. Report cycles instead of recursing forever.

### Draw order via sort handles

Store an explicit order list and emit `SORTENTSTABLE` where the target version
supports it; otherwise order entities in the ENTITIES section. This matches DXF
semantics per version.

## Risks / Trade-offs

- **Query performance** → Pre-filter with the spatial index and short-circuit
  predicates; cache compiled queries.
- **Handle-based selections after deletion** → Resolution skips missing handles
  and reports them.
- **Transform of non-representable types** → Return explicit errors per entity
  capability.
- **Order portability** → Document which versions use SORTENTSTABLE.
