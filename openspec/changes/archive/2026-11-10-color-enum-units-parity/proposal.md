# Proposal

## Why

Colors, enumerations, and units are cross-cutting: every entity, table entry,
renderer, and exporter depends on them. The Go library has a `color` package, an
`insunit` package, and scattered constants, but no complete ACI table, no
true-color/transparency model, no consolidated enum set, and no unit conversion
matching ezdxf's `colors`, `enums`, and `units` modules. Inconsistent handling
causes wrong rendering and lossy round trips.

## What Changes

- Add the full AutoCAD Color Index table with RGB values and nearest-index
  lookup.
- Model ACI, true color, BYLAYER, BYBLOCK, and transparency consistently.
- Add color resolution against layer and block context.
- Add a consolidated enum package for DXF enumerations used by entities, tables,
  and objects.
- Add unit definitions, `$INSUNITS` handling, and conversion between units.
- Add linetype, lineweight, and text-style default helpers.
- Validate color, enum, and unit values on load and save.

## Capabilities

### New Capabilities

- `color-enum-units`: colors, transparency, DXF enumerations, units, and the
  shared value types used across the model.

### Modified Capabilities

None.

## Impact

Affected code: `color/*`, `insunit/*`, new `enums/*` and `units/*`, plus
`entity/*`, `table/*`, `render/*`, and `export/*` consumers. Public API additions
and normalization of existing constants. Tests cover the ACI table, color
resolution, enum round trips, and unit conversions.
