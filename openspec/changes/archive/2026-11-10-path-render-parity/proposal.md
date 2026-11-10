# Proposal

## Why

The render and export layers currently handle only lines, circles, arcs, and
polylines. HATCH, text, splines, ellipses, dimensions, images, and most other
entities are not drawn, text-to-path produces placeholder glyphs, and PDF/SVG
output cannot represent lineweights, transparency, or layouts. This prevents
headless preview, plotting, and any ezdxf-drawing-like workflow.

## What Changes

- Build paths from every renderable entity: lines, arcs, circles, ellipses,
  splines, polylines, hatches, text, dimensions, leaders, solids, surfaces,
  images, and underlays.
- Add a renderer interface with SVG, PDF, and raster (PNG) backends.
- Support stroke and fill styles: color, true color, lineweight, linetype,
  transparency, and layer state.
- Render text as paths with real glyph outlines from font files.
- Render hatch patterns and gradient fills.
- Render layouts, viewports, and paper space with correct transforms.
- Add clipping and view extents.
- Make backends pluggable through a registry.

## Capabilities

### New Capabilities

- `path-render`: entity-to-path conversion, rendering backends, styling, text
  outlines, and layout plotting.

### Modified Capabilities

None.

## Impact

Affected code: `path/*`, `render/*`, `export/*`, `text2path/*`, `patterns/*`,
`layouts/*`, `entity/*`. New public renderer API. Tests compare SVG/PDF output
against golden files and raster output against pixel tolerances. Depends on the
entity model and geometry changes.
