# Tasks

## 1. Framework and streaming

- [ ] 1.1 Add the add-on package layout and capability registry; verify discovery
- [ ] 1.2 Implement the R12 streaming writer; verify a 10k-entity file parses
- [ ] 1.3 Implement the streaming reader; verify bounded-memory iteration

## 2. Drawing and plot styles

- [ ] 2.1 Implement the drawing add-on over the core renderer; verify SVG/PDF/PNG output
- [ ] 2.2 Implement CTB/STB parsing and application; verify color and lineweight rules

## 3. Conversion

- [ ] 3.1 Implement the importer for supported formats; verify geometry equivalence
- [ ] 3.2 Implement dxf2code generation; verify generated code reproduces a document
- [ ] 3.3 Implement mtxpl helpers; verify documented transformations

## 4. CSG and text

- [ ] 4.1 Replace bbox CSG with mesh primitives and booleans; verify volume and face-orientation tests
- [ ] 4.2 Implement MTextExplode; verify line positioning
- [ ] 4.3 Implement real text-to-path; verify glyph outlines

## 5. Geospatial, mesh, and external tools

- [ ] 5.1 Implement geo export; verify valid GeoJSON
- [ ] 5.2 Implement mesh and OpenSCAD export; verify vertices/faces
- [ ] 5.3 Implement ODA and HPGL/2 integration with an injectable runner; verify missing-tool errors

## 6. Remaining helpers

- [ ] 6.1 Implement xplayer animation export; verify frame output
- [ ] 6.2 Implement binpacking; verify non-overlap and bounds
- [ ] 6.3 Implement tablepainter; verify table rendering
- [ ] 6.4 Implement gerber helpers; verify output validity

## 7. Verification

- [ ] 7.1 Add per-add-on fixture tests and verify the full suite
- [ ] 7.2 Update `docs/parity/README.md` add-on matrix

## Workflow follow-up

- Archive the change after review.
