# Design

## Context

The scene renderer covers styles, paths, hatches, dimensions, meshes, arrows, plot styles, SHX metrics, and ISO patterns. Remaining geometry gaps are concentrated in entity edge cases: HATCH and MPOLYGON edge-path boundaries are retained instead of drawn, ELLIPSE records with embedded splines are retained, complex linetypes are reduced to dash lengths, and PDF gradients use only the first color. Visual conformance is structural rather than image-based. See `proposal.md` for the gap inventory.

Constraints:

- Rendering output must remain deterministic for the same scene and backend.
- Existing path and renderer interfaces should not change for simple entities.
- Unsupported content must be reported, not silently approximated.
- Golden visual tests must be stable across supported platforms.

## Goals / Non-Goals

**Goals:**

- Render every documented HATCH/MPOLYGON boundary edge type, holes, and transforms.
- Convert ELLIPSE embedded splines through the spline path pipeline.
- Render complex linetype text and shape elements with placement and orientation.
- Produce complete axial and radial PDF gradients.
- Add golden-image and structural visual conformance tests.
- Report unsupported geometry in a deterministic fallback report.

**Non-Goals:**

- New backends, interactive viewers, or plot/paper-space scaling.
- Exact PDF feature parity beyond gradient rendering and existing path fills.
- Rasterizing text identically to AutoCAD fonts.

## Decisions

### Canonical boundary edge model

Introduce a typed boundary edge union for line, arc, ellipse, spline, and bulge segments. Each edge owns its control geometry, closed flag, and source flags. HATCH and MPOLYGON parsers build these edges directly, and a normalizer converts legacy point-only boundaries where unambiguous.

Holes remain separate loops. The converter flattens each loop after applying the owning entity transform, then emits path subpaths in loop order. Winding is normalized for even-odd filling but preserved where the source record exposes orientation.

Alternative: convert edges to polylines during parsing. Rejected because it loses curve fidelity and prevents later exact formatting.

### Embedded ellipse splines as first-class data

ELLIPSE records with embedded spline data parse into the existing spline model. Formatting writes the original embedded form; path conversion and rendering use the decoded spline. If the embedded spline is malformed, the record is retained and a diagnostic is reported.

### Complex linetype element renderer

Extend the dash pipeline from lengths to pattern elements. A linetype pattern becomes an ordered sequence of dash, gap, text, and shape elements. Text uses `text2path` glyphs; shapes use the loaded SHX font registry. Each element applies its rotation, scale, and x/y offset at its position along the segment. Missing fonts or shapes produce a warning and fall back to the dash geometry for that element.

The existing dash-only path remains for simple linetypes and for callers that only need lengths.

### PDF gradient shading

PDF axial gradients use a type 2 shading dictionary and radial gradients use type 3. Multi-stop gradients are represented with a stitching function over exponential sub-functions so every stop participates. Extend is set from the gradient geometry. Unsupported color spaces fall back to the first color with a diagnostic, preserving current behavior.

### Visual conformance

Golden fixtures render at a fixed viewport, background, line width, and scale. Raster comparisons use a per-channel tolerance and a bounded mismatch ratio; SVG comparisons assert path presence, ordering, and key style attributes. Fixtures that need text use the built-in vector font unless the test explicitly requires a system font.

Fallbacks are collected with entity handle, feature, and reason, then sorted by handle and feature so reports are stable.

## Risks / Trade-offs

- [Boundary edge conversion can change hatch output for existing files] → Add golden fixtures for each edge type and compare area and containment before rendering.
- [Complex linetype rendering is font-dependent] → Use deterministic built-in fonts by default and treat missing fonts as reported fallbacks.
- [PDF shading functions are easy to malform] → Validate generated PDFs with an independent parser in tests and compare rasterized output.
- [Golden images are platform-sensitive] → Fix renderer settings, avoid antialiasing differences, and use tolerances with documented bounds.
- [More geometry increases scene build time] → Cache flattened edges per entity transform and avoid repeated tessellation.

## Migration Plan

Output changes intentionally for previously unrendered boundary and linetype content. No API removal is planned. Callers that need the old dash-only behavior can continue using the length-based linetype API.
