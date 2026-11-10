# Design

## Approach

Refactor in small, test-backed steps:

1. Identify dead code with `grep`, `go vet`, and test coverage.
2. Delete unused helpers; if a helper is still useful, move it behind a named interface.
3. Consolidate block adapters into one exported or internal type with `Unwrap`, `LayerName`, and entity forwarding.
4. Replace inline `interface{ Method() }` assertions with named interfaces in the owning package.
5. Review for single-responsibility violations, duplicated parsing, and speculative abstractions.

Each step runs `go test ./...` and the opt-in cross-check before moving on.

## Risks

Broad refactors can accidentally change tag output. Keep changes mechanical and rely on round-trip/cross-check tests.
