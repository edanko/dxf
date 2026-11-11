# Design

## Context

`xref/` has manager and loader stubs. `entity/` has partial IMAGE and WIPEOUT
definitions, and `object/` lacks IMAGEDEF, RASTERVARIABLES, and underlay
definitions. Rendering ignores images. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- Reliable attach/bind/reload with explicit unresolved state.
- Correct clipping in both model and render.
- Round-trip preservation of all reference metadata.

**Non-Goals:**

- Decoding DGN/DWF content for geometry extraction.
- Editing raster pixel data.
- Full AutoCAD path-search semantics beyond configured search paths.

## Decisions

### References are block records plus metadata

An xref is represented as a block record with an external path and an insert,
matching DXF. Binding rewrites the block record to local content and clears the
path. This reuses the block model and save path.

### Resolution is lazy and explicit

Loading records references without requiring files. Resolution happens on demand
(reload, bind, render, audit) and returns a typed unresolved result. This lets
documents with missing references load and save safely.

### Clip as a boundary object

Store clip boundaries as polyline/rectangle data on the reference and apply them
in the entity-to-path stage. Rendering clips paths rather than pixels. Images
clip at the sampling stage.

### Image/underlay definitions as objects

IMAGEDEF and underlay definitions are objects referenced by handle from entities,
following the DXF object graph. This keeps entity data small and shareable.

### Cycle detection by canonical path

Resolve paths to canonical absolute paths and track a visited set during nested
loads. Report cycles with the participating paths.

## Risks / Trade-offs

- **Path portability** → Store both original and resolved paths; write the
  original form.
- **Missing files in tests** → Use fixtures with intentionally missing paths and
  assert unresolved reporting.
- **Clip complexity** → Start with rectangular and polyline clips; document
  unsupported inverted or self-intersecting clips.
- **Large images** → Decode lazily and cache by path and modification time.
