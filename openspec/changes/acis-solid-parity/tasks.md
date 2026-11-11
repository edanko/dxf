# Tasks

## 1. SAT and SAB decoding

- [ ] 1.1 Create the `acis` package with shared record, reference, attribute, and diagnostic types; verify `go test ./acis` compiles and unit tests cover record construction.
- [ ] 1.2 Implement the SAT text decoder for headers, records, references, attributes, and end markers; verify valid and malformed SAT fixtures.
- [ ] 1.3 Implement the SAB binary decoder producing the same record graph as SAT; verify equivalent SAT/SAB fixtures produce identical topology.
- [ ] 1.4 Add fuzz tests for truncated and malformed SAT/SAB streams; verify no panic and every error names the failing record or offset.

## 2. Topology and entity integration

- [ ] 2.1 Build typed body, lump, shell, face, loop, coedge, edge, and vertex objects with ID-based reference resolution; verify traversal and broken-reference diagnostics.
- [ ] 2.2 Add bounding boxes and structural validation for bodies and faces; verify finite bounds for a valid body and a diagnostic for a missing loop.
- [ ] 2.3 Add lazy decoded-body caching on 3DSOLID, REGION, and BODY while preserving the original payload; verify repeated access parses once and unchanged entities re-emit bytes verbatim.
- [ ] 2.4 Transform decoded ACIS geometry with the host entity and invalidate the cache; verify translated bodies and the supported-payload versus unsupported-payload save rules.

## 3. Meshing

- [ ] 3.1 Triangulate supported planar face loops and orient triangles from the face normal and shell orientation; verify a cube body produces twelve finite-normal triangles.
- [ ] 3.2 Report unsupported surfaces and malformed loops while returning supported triangles; verify partial meshing diagnostics name the face and reason.
- [ ] 3.3 Document the ACIS decoding and meshing limits in `docs/parity/README.md`; verify the docs distinguish supported planar bodies from unsupported surfaces and history.

## 4. Creation and export

- [ ] 4.1 Convert an indexed closed triangle mesh into one body/lump/shell with shared vertices and edges; verify a box mesh produces the expected face and vertex counts.
- [ ] 4.2 Export supported bodies to SAT and SAB and reparse the result; verify the parser recovers the same body, face, and vertex counts.
- [ ] 4.3 Return explicit errors for open, degenerate, or non-manifold input meshes; verify each rejection case.

## 5. Rendering and add-on integration

- [ ] 5.1 Render ACIS solids as triangle face paths in `BuildScene` and path conversion, with a diagnostic fallback for unsupported payloads; verify a 3DSOLID fixture emits twelve closed paths.
- [ ] 5.2 Expose body, face, triangle, and bounding-box data through `addons/acisbrowser`; verify supported and unsupported fixtures.
- [ ] 5.3 Include ACIS body triangles in `addons/meshex` export; verify exported vertex and face counts for a cube.
- [ ] 5.4 Run `go test ./...` and a fuzz smoke run over ACIS fixtures; verify no regressions and no panics.
