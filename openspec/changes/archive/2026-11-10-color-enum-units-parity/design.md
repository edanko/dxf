# Design

## Context

`color/` exists but lacks a complete ACI table and true-color/transparency model.
`insunit/` exists with limited units. Constants are scattered across entity and
table packages. See `proposal.md`.

## Goals / Non-Goals

**Goals:**

- One canonical representation for colors and units.
- Typed enums with lossless unknown handling.
- Consistent resolution across rendering and export.

**Non-Goals:**

- ICC color management.
- Plot-style-dependent color mapping (belongs to the add-on change).
- Unit-aware geometry arithmetic beyond length conversion.

## Decisions

### Color as a tagged value

Represent color as a struct with a kind (ACI, true color, BYLAYER, BYBLOCK) and
the corresponding payload. Provide constructors and resolution helpers. Avoid
interfaces, which complicate comparison and serialization. Alternative: separate
types per kind. Rejected for API friction.

### Generated ACI table

Store the 256-entry ACI table as generated data with a test that checks a sample
of documented values and monotonic nearest lookup. Avoid hand-typed values.

### Enums as named integer types with metadata

Define each enum as a named integer type plus maps for name and numeric
conversion. Unknown values remain representable because the underlying type is an
integer. Alternative: string enums. Rejected because DXF stores numbers and
round-trip fidelity matters.

### Units as a table with conversion factors

Define units with a canonical meter factor and convert through it. `$INSUNITS`
maps to the same table. Unitless is explicit and never silently treated as
meters.

### Validation at boundaries

Validate in load and save, not in constructors, so raw invalid values can be
preserved and reported. Constructors validate to catch programmer errors.

## Risks / Trade-offs

- **ACI approximation** → Document that nearest lookup is approximate and expose
  the distance used.
- **Enum drift** → Generate enum definitions from the documented DXF value lists
  where possible.
- **Floating-point conversion** → Use float64 and document tolerance.
- **API duplication** → Deprecate old constants, keep aliases during migration.
