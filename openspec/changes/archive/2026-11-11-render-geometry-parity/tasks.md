# Tasks

## 1. Typed boundary edges

- [x] 1.1 Define typed boundary edge variants for polyline points, lines, arcs, ellipses, splines, and bulges with load, format, and transform support; verify mixed-edge round-trip and transform tests.
- [x] 1.2 Convert HATCH and MPOLYGON boundaries, including holes, into closed scene paths after entity transforms; verify a mixed outer boundary and spline hole fixture.
- [x] 1.3 Report unsupported boundary edges with entity handle and edge reason while retaining supported edges; verify the fallback report is stable across repeated builds.
- [x] 1.4 Document boundary edge coverage and fallbacks in `docs/parity/README.md`; verify the docs list every supported edge type.

## 2. Embedded ellipse splines

- [x] 2.1 Parse embedded ELLIPSE spline data into the typed spline model and re-emit it unchanged; verify load/save and typed-field tests.
- [x] 2.2 Convert embedded-spline ellipses through the spline path pipeline; verify the rendered path follows the control data rather than the retained record.
- [x] 2.3 Report malformed embedded spline data with an entity handle and retain the original record; verify the malformed fixture.

## 3. Complex linetypes

- [x] 3.1 Preserve ordered complex linetype elements with shape, text, rotation, scale, and offsets through load and format; verify a shape-and-text round trip.
- [x] 3.2 Render shape and text elements along paths using SHX and text2path resources and apply element placement; verify geometry positions and transformed path orientation.
- [x] 3.3 Emit remaining dash and text elements when a shape cannot be resolved and report the missing shape and linetype; verify the fallback test.
- [x] 3.4 Document complex linetype support and resource requirements in `docs/parity/README.md`; verify the docs name the SHX requirement.

## 4. PDF gradients

- [x] 4.1 Emit axial PDF shadings with all configured color stops in order; verify a three-stop fixture contains one shading with three stops.
- [x] 4.2 Emit radial PDF shadings with the configured center and radius; verify a radial fixture contains the expected shading dictionary.
- [x] 4.3 Fall back deterministically for unsupported color spaces with a diagnostic; verify PDF output remains valid.
- [x] 4.4 Document PDF gradient support and fallbacks in `docs/parity/README.md`; verify the docs no longer claim first-color rendering.

## 5. Visual conformance

- [x] 5.1 Add a golden-image harness with fixed viewport, background, line width, and scale plus per-channel and mismatch-ratio tolerances; verify a deliberate pixel change fails with fixture, backend, and region.
- [x] 5.2 Add curated raster and SVG fixtures for boundary edges, embedded splines, complex linetypes, and PDF gradients; verify the documented tolerances pass on the reference platform.
- [x] 5.3 Run `go test ./...` and the visual conformance suite; verify no regressions and record the supported platform tolerance.
