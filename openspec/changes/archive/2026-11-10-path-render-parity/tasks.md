# Tasks

## 1. Path pipeline

- [x] 1.1 Define the path segment model and style struct; verify with unit tests
- [x] 1.2 Add converters for line, arc, circle, ellipse, spline, and polylines; verify geometry tolerance tests
- [x] 1.3 Add converters for HATCH, TEXT, MTEXT, dimensions, leaders, solids, images, and wipeouts; verify each produces paths

## 2. Backends

- [x] 2.1 Refactor SVG to the renderer interface with stroke/fill/color/lineweight/transparency; verify golden SVG
- [x] 2.2 Refactor PDF to the renderer interface; verify the output opens and contains geometry
- [x] 2.3 Add a raster PNG backend with size, background, and anti-aliasing; verify dimensions and pixel coverage
- [x] 2.4 Add the backend registry; verify a custom backend dispatches

## 3. Text and hatch

- [x] 3.1 Implement real font outline loading and caching for text2path; verify glyphs render
- [x] 3.2 Implement MTEXT line layout and formatting; verify multi-line output
- [x] 3.3 Implement solid, pattern, and gradient hatch rendering; verify fill and pattern cases

## 4. Layouts and clipping

- [x] 4.1 Implement layout and viewport rendering with transforms; verify a layout scene
- [x] 4.2 Implement rectangle/boundary clipping and explicit extents; verify no geometry outside the clip

## 5. Verification

- [x] 5.1 Add cross-backend conformance tests for a styled sample drawing; verify all backends agree on visible geometry
- [x] 5.2 Update `docs/parity/README.md` render/path matrix and record unsupported plot styles

## Workflow follow-up

- Archive the change after review.
