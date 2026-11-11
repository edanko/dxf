# Proposal

## Why

The implementation grew through thirteen parity changes and now contains dead helpers, duplicate adapters, and inline interface assertions. These increase maintenance cost and obscure the public model without adding behavior.

## What Changes

- Remove or replace dead legacy `parser.go` helpers.
- Remove or wire `render/simple_dimension.go` and `render/entity_renderer.go`.
- Deduplicate `blockEntityAdapter` in `binder.go` and `drawing/space.go`.
- Replace inline interface assertions with named interfaces.
- Apply SOLID/DRY/YAGNI and readability fixes across packages touched by parity work.
- **No public behavior changes**; all existing tests and the ezdxf cross-check must continue to pass.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. This is a pure refactor and uses `skip_specs: true`.

## Impact

Affected packages: root, `drawing`, `render`, `object`, `cmd/dxf`, and any package with duplicate adapters or dead code. Public APIs remain compatible.
