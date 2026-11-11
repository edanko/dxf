# Tasks

## 1. Lossless records and comments

- [x] 1.1 Add an ordered per-record retention ledger and route entity load/format through it; verify unknown entity codes survive `load → save → load` in a golden tag test.
- [x] 1.2 Apply the retention ledger to typed table entries, blocks, and objects; verify unknown codes on LAYER, BLOCK, and DICTIONARY records survive a round trip.
- [x] 1.3 Store comments as positioned tokens and replay them at their original record boundary; verify inter-entity comments keep order and `StripComments` removes every comment.
- [x] 1.4 Document the retention and comment guarantees in `docs/parity/README.md`; verify the document no longer claims these areas are partial.

## 2. Attribute conformance and bounds

- [x] 2.1 Add a per-type attribute conformance matrix and a generated test that fails when a registered type lacks a matrix row or golden fixture.
- [x] 2.2 Fill the entity matrix with documented codes, defaults, and version gates; verify golden LINE, SPLINE, HATCH, and ACAD_TABLE round trips plus one version-downgrade omission test.
- [x] 2.3 Fill the table and object matrices; verify golden DIMSTYLE, LAYOUT, and dictionary round trips preserve typed fields and unhandled codes.
- [x] 2.4 Verify copy isolation for every mutable entity family, including ACIS payloads and MPolygon state; add regression tests where a copied object still shares state.
- [x] 2.5 Derive bounding boxes from concrete geometry for every entity category, including decoded proxy geometry, and treat empty geometry as no extents; verify query bounds return found=false for empty selections.

## 3. Validation and repair

- [x] 3.1 Add tag-level validation before binding for section terminators, required codes, numeric ranges, and duplicate handles; verify malformed fixtures report code, value, and handle.
- [x] 3.2 Add a repair pass that records every action and re-validates afterward; verify a missing block terminator is inserted and the repaired file audits clean.
- [x] 3.3 Implement duplicate-handle repair by assigning a fresh handle to later duplicates and rewriting only unambiguous references; verify ambiguous references remain reported.
- [x] 3.4 Implement block-cycle repair that removes the closing INSERT edge, reports the cycle, and retains the block; verify self-cycle and two-block cycle fixtures.
- [x] 3.5 Document the validator and repair behavior in `docs/parity/README.md`; verify the docs list the remaining limitations explicitly.

## 4. Semantic cross-check and integration

- [x] 4.1 Extend the opt-in ezdxf cross-check to compare a stable typed-field projection per entity, table entry, and object; verify a deliberate field difference fails the harness.
- [x] 4.2 Run `go test ./...` and the opt-in cross-check over `testdata/`; verify every fixture passes or is listed with a documented exclusion.
