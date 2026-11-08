# Tasks

## 1. Query engine

- [ ] 1.1 Define the entity query context and compile AST to predicates; verify type/layer/handle filters
- [ ] 1.2 Add attribute, xdata, and spatial predicates; verify each filter
- [ ] 1.3 Add query error reporting with positions; verify invalid syntax cases
- [ ] 1.4 Execute queries over the entity database with spatial pre-filtering; verify result sets

## 2. Selections and grouping

- [ ] 2.1 Implement handle-based selection sets and union/intersection/difference; verify set operations
- [ ] 2.2 Implement grouping by attribute; verify group membership

## 3. Transform

- [ ] 3.1 Implement entity, selection, block, and document transforms; verify rotation/scale tests
- [ ] 3.2 Implement explode and disassemble with cycle detection; verify nested and cyclic blocks

## 4. Document operations

- [ ] 4.1 Implement draw-order read/write with SORTENTSTABLE and version fallback; verify round trip
- [ ] 4.2 Implement upright text normalization; verify rotation and position
- [ ] 4.3 Implement zoom/extents helpers; verify fit calculations
- [ ] 4.4 Implement block reference traversal and cycle reporting; verify nested cases
- [ ] 4.5 Implement revcloud creation; verify boundary output
- [ ] 4.6 Implement proxy-graphics extraction; verify a proxy fixture

## 5. Verification

- [ ] - [ ] 5.1 Add end-to-end query/transform/explode tests and verify save/reload stability
- [ ] 5.2 Update `docs/parity/README.md` query/transform matrix

## Workflow follow-up

- Archive the change after review.
