# Proposal

## Why

Only 29 entity types are reconstructed from tags; every other type falls back to
a lossy `GenericEntity`. Even for supported types there is no common per-entity
load contract, and operations that ezdxf exposes on entities (transform, copy,
bounding box, explode, subentity access) are missing or inconsistent. This makes
the Go model unable to represent or manipulate most real drawings.

## What Changes

- Add typed models and tag loaders for all standard DXF entities, including:
  3DSOLID, REGION, BODY, surfaces (EXTRUDED, LOFTED, PLANES, REVOLVED, SWEPT),
  MESH, HELIX, LIGHT, SUN, MULTILEADER, MLEADERSTYLE, MLINESTYLE, ACAD_TABLE,
  TABLESTYLE, OLE2FRAME, ACAD_PROXY_ENTITY, ARC_DIMENSION, LARGE_RADIAL_DIMENSION,
  MPOLYGON, SHAPE, XLINE, RAY, WIPEOUT, TOLERANCE, FIELD, and all dimension
  variants.
- Introduce a common entity contract: handle, owner, layer, linetype, color,
  true color, lineweight, linetype scale, visibility, transparency, extrusion,
  and extended data.
- Add per-entity load from tags and round-trip format tests.
- Add transform, copy, bounding box, explode, and subentity access with
  per-type implementations.
- Add entity validation hooks used by audit.
- Keep unknown/proxy entities as retained raw tags rather than losing them.

## Capabilities

### New Capabilities

- `entity-model`: typed DXF entity models, common attributes, loading, and
  entity operations.

### Modified Capabilities

None.

## Impact

Affected code: `entity/*` (new files per type and a common base), root
`parser.go`/`dxf_main.go` load dispatch, `math/*` for transforms, `drawing/*`
for entity storage. Large test surface: per-type load/format tests plus
transform, bbox, and explode tests. The generic fallback remains for third-party
types.
