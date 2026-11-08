# Design

## Context

Existing related code: `csg` (bbox-approximation booleans), `text2path`
(placeholder glyphs), `export/svg`, `render`, `patterns`, `path`. There is no
add-on framework or streaming I/O. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- Add-ons as isolated packages with clear entry points.
- Streaming paths that avoid full-document memory.
- Native Go implementations where feasible.

**Non-Goals:**

- Reproducing Python-only integrations that require CPython or matplotlib.
- Full CAD kernel fidelity for CSG.
- Bit-identical output with ezdxf add-ons.

## Decisions

### Package-per-add-on under `addons/`

Each add-on is its own package with a small public API and no cross-add-on
imports. Shared helpers live in internal packages. This keeps dependency weight
opt-in. Alternative: one large `addons` package. Rejected for dependency bloat.

### Registry only for discoverable capabilities

Use a registry for add-ons that expose a named operation (exporters, backends),
but keep library-style helpers as plain functions. Avoid over-abstracting.

### Streaming via tag-level callbacks

`iterdxf` and `r12writer` operate on the tag stream directly, reusing the tagger
and writer primitives rather than the document model. This bounds memory.

### CSG via a mesh kernel

Implement primitives, half-edge mesh cleanup, and boolean operations with a
BSP or robust triangle-intersection approach. Replace the current bbox
approximation. Document numerical limits.

### Font and image dependencies are optional

Text-to-path and drawing add-ons isolate font/image decoding behind interfaces so
core users do not pay the dependency cost.

### External tools behind explicit interfaces

ODA and HPGL integrations shell out through an injectable command runner so tests
can substitute a fake and users get clear missing-tool errors.

## Risks / Trade-offs

- **Scope explosion** → Ship add-ons in independent groups; each is useful alone.
- **CSG robustness** → Start with convex primitives and document non-convex
  limitations; add fuzzing.
- **Optional dependency conflicts** → Keep add-ons in separate modules or build
  tags where needed.
- **External tool variance** → Validate output and surface tool stderr.
