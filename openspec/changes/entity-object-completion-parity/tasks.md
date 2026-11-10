# Tasks

## 1. ACAD_TABLE

- [ ] 1.1 Add typed row/column/cell model and loader; verify with `acad_table_simple.dxf`
- [ ] 1.2 Add table style and merge handling; verify round trip with `acad_table_with_blk_ref.dxf`
- [ ] 1.3 Register ACAD_TABLE in the entity loader and add unit tests

## 2. Subentities

- [ ] 2.1 Add POLYFACE/POLYMESH vertex variants and face records
- [ ] 2.2 Add MESHVERTEXCACHE typed model and linkage
- [ ] 2.3 Add round-trip tests for polyface and mesh fixtures

## 3. ACIS

- [ ] 3.1 Add SAT header/body metadata parser
- [ ] 3.2 Add SAB presence and payload preservation checks
- [ ] 3.3 Add ACIS tests that assert payload byte equality

## 4. Objects and collections

- [ ] 4.1 Add typed collection objects and layer filter
- [ ] 4.2 Add VBA_PROJECT model
- [ ] 4.3 Load objects-as-entities through typed models while preserving section placement
- [ ] 4.4 Extend extension dictionary/reactor coverage and tests

## 5. Verification

- [ ] 5.1 Extend the ezdxf cross-check harness for typed inventories
- [ ] 5.2 Run `go test ./...` and the opt-in cross-check
- [ ] 5.3 Update `docs/parity/README.md`
