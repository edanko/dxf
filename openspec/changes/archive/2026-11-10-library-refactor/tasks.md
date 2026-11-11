# Tasks

## 1. Dead code

- [x] 1.1 Audit `parser.go` helpers and remove or wire them
- [x] 1.2 Remove or wire `render/simple_dimension.go`
- [x] 1.3 Remove or wire `render/entity_renderer.go`

## 2. Duplication and interfaces

- [x] 2.1 Consolidate block entity adapters
- [x] 2.2 Replace inline interface assertions with named interfaces
- [x] 2.3 Review object/load.go, cmd/dxf/main.go, and render for duplication

## 3. Quality review

- [x] 3.1 Run Golang skills review for SOLID/DRY/YAGNI and readability
- [x] 3.2 Address findings with behavior-preserving edits
- [x] 3.3 Run `go test ./...` and the ezdxf cross-check
- [x] 3.4 Update `docs/parity/README.md` hazards section
