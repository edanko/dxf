# Tasks

## 1. Framework and streaming

- [x] 1.1 Add the add-on package layout and capability registry; verify discovery
- [x] 1.2 Implement the R12 streaming writer; verify a 10k-entity file parses
- [x] 1.3 Implement the streaming reader; verify bounded-memory iteration

## 2. Drawing and plot styles

- [x] 2.1 Implement the drawing add-on over the core renderer; verify SVG/PDF/PNG output
- [x] 2.2 Implement CTB/STB parsing and application; verify color and lineweight rules

## 3. Conversion

- [x] 3.1 Implement the importer for supported formats; verify geometry equivalence
- [x] 3.2 Implement dxf2code generation; verify generated code reproduces a document
- [x] 3.3 Implement mtxpl helpers; verify documented transformations

## 4. CSG and text

- [x] 4.1 Replace bbox CSG with mesh primitives and booleans; verify volume and face-orientation tests
- [x] 4.2 Implement MTextExplode; verify line positioning
- [x] 4.3 Implement real text-to-path; verify glyph outlines

## 5. Geospatial, mesh, and external tools

- [x] 5.1 Implement geo export; verify valid GeoJSON
- [x] 5.2 Implement mesh and OpenSCAD export; verify vertices/faces
- [x] 5.3 Implement ODA and HPGL/2 integration with an injectable runner; verify missing-tool errors

## 6. Remaining helpers

- [x] 6.1 Implement xplayer animation export; verify frame output
- [x] 6.2 Implement binpacking; verify non-overlap and bounds
- [x] 6.3 Implement tablepainter; verify table rendering
- [x] 6.4 Implement gerber helpers; verify output validity

## 7. Verification

- [x] 7.1 Add per-add-on fixture tests and verify the full suite
- [x] 7.2 Update `docs/parity/README.md` add-on matrix

## Workflow follow-up

- Archive the change after review.
