# Proposal

## Why

The math package covers the common geometry surface but lacks several ezdxf utilities used by generators, nesting, and exact boundary workflows. Triangulation also has a known hole-bridge area shortfall and portal clipping is missing.

## What Changes

- Add Perlin noise with repeatable seeded output.
- Add clustering helpers for point sets.
- Add a `Shape` container for closed boundaries with holes and orientation.
- Add portal clipping for line/curve segments against a rectangular portal.
- Fix triangulation hole bridging so clipped area matches the input polygon.
- Add path nesting helpers built on offset and containment.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `geometry-math`: Perlin, clustering, Shape, portal clipping, triangulation correctness.
- `path-render`: path nesting.

## Impact

Affected packages: `math`, `path`, tests and fixtures. No public breaking changes; new APIs are additive.
