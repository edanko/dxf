# Proposal

## Why

Rendering covers the common cases, but several documented geometry forms still fall back to retained records or simplified output. HATCH edge paths, embedded ellipse splines, complex linetypes, and PDF gradients can therefore disappear or render incorrectly, and the current conformance tests do not detect visual regressions.

## What Changes

- Render HATCH and MPOLYGON boundaries made of line, arc, ellipse, spline, and bulge edges, including transformed boundaries and hole paths, instead of falling back to retained records.
- Render ELLIPSE records that carry embedded spline control data through the spline pipeline.
- Render complex linetypes fully: dash/gap elements plus embedded shape and text elements, with rotation, scale, and offsets applied along the path.
- Implement real PDF axial and radial gradient shadings instead of emitting only the first color.
- Add a visual conformance harness that renders a curated fixture set to raster and SVG and compares against golden images with documented tolerances.
- Keep fallback reporting deterministic when an entity still cannot be rendered, including the entity handle and reason.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `path-render`: complete HATCH/MPOLYGON boundary conversion, embedded-spline ellipses, complex linetype geometry, PDF gradients, and visual conformance.
- `entity-model`: typed access and transformation for boundary edge variants and embedded ellipse spline data.

## Impact

Affected code: `entity/boundary_paths.go`, `entity/hatch*.go`, `entity/mpolygon.go`, `entity/ellipse.go`, `path`, `render` backends, `render/linetypes`, and test fixtures. Output changes for drawings that previously relied on fallbacks. Golden-image tests add platform-sensitive artifacts; the harness must tolerate documented anti-aliasing and font differences.
