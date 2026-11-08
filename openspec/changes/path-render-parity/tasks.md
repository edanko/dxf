# Tasks

## 1. Path pipeline

- [ ] 1.1 Define the path segment model and style struct; verify with unit tests
- [ ] 1.2 Add converters for line, arc, circle, ellipse, spline, and polylines; verify geometry tolerance tests
- [ ] 1.3 Add converters for HATCH, TEXT, MTEXT, dimensions, leaders, solids, images, and wipeouts; verify each produces paths

## 2. Backends

- [ ] 2.1 Refactor SVG to the renderer interface with stroke/fill/color/lineweight/transparency; verify golden SVG
- [ ] 2.2 Refactor PDF to the renderer interface; verify the output opens and contains geometry
- [ ] 2.3 Add a raster PNG backend with size, background, and anti-aliasing; verify dimensions and pixel coverage
- [ ] 2.4 Add the backend registry; verify a custom backend dispatches

## 3. Text and hatch

- [ ] 3.1 Implement real font outline loading and caching for text2path; verify glyphs render
- [ ] 3.2 Implement MTEXT line layout and formatting; verify multi-line output
- [ ] 3.3 Implement solid, pattern, and gradient hatch rendering; verify fill and pattern cases

## 4. Layouts and clipping

- [ ] 4.1 Implement layout and viewport rendering with transforms; verify a layout scene
- [ ] 4.2 Implement rectangle/boundary clipping and explicit extents; verify no geometry outside the clip

## 5. Verification

- [ ] 5.1 Add cross-backend conformance tests for a styled sample drawing; verify all backends agree on visible geometry
- [ ] 5.2 Update `docs/parity/README.md` render/path matrix and record unsupported plot styles

## Workflow follow-up

- Archive the change after review.
