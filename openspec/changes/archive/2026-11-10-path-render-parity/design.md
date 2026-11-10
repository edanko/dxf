# Design

## Context

`path/` converts only line, lwpolyline, polyline, circle, arc, and ellipse.
`render/` draws only line/circle/arc/lwpolyline and has SVG/PDF backends.
`export/svg` covers a similar subset. `text2path` glyphs are placeholders.
`patterns/` has lookup tables. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- A single entity-to-path pipeline feeding all backends.
- Real text outlines and hatch fills.
- Layout-aware rendering.

**Non-Goals:**

- Pixel-perfect AutoCAD plotting styles (CTB/STB) in this change.
- Full 3D rendering of ACIS solids.
- Animation or interactive viewers.

## Decisions

### Path as the intermediate representation

All entities lower to a path of move/line/curve/close segments plus fill rules.
Backends consume paths only. This isolates entity complexity from backends.
Alternative: each backend walks entities. Rejected because styling and
tessellation would be duplicated.

### Renderer interface with a style struct

Define `Renderer` with `Begin`, `DrawPath`, `SetStyle`, `Clip`, and `End`.
Styles carry resolved color, lineweight, linetype, transparency, and fill.
Backends translate styles to their native model.

### Tessellation policy per backend

Paths keep curves; backends tessellate according to their own tolerance or
native curve support. SVG/PDF keep curves where possible; raster tessellates.
This preserves vector quality.

### Font loading for text2path

Load TrueType/OpenType outlines through a font parser and cache glyphs by
font/postscript name. Fall back to a built-in stroke font only when no font is
found, and report the fallback. Alternative: ship pre-extracted glyph tables.
Rejected for size and licensing.

### Hatch rendering via boundary plus pattern

Use the hatch boundary for fill and generate pattern lines from `patterns/`
definitions, clipped to the boundary. Gradients map to backend gradient
primitives where available.

### Layout rendering as a scene graph

Build a scene of layout paper space plus viewport views, then render. Viewport
content is transformed into viewport coordinates.

## Risks / Trade-offs

- **Font availability** → Configurable font search paths and explicit fallback
  reporting.
- **Hatch performance** → Clip pattern lines early and cache compiled patterns.
- **Backend divergence** → Shared conformance tests across backends.
- **Large drawings** → Stream paths and support spatial culling.
