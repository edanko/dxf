# Proposal

## Why

ACIS-backed entities are currently opaque payloads with shallow metadata. Callers cannot inspect solids, regions, or bodies as geometry, generate a mesh for rendering/export, or create simple ACIS bodies, which leaves a major ezdxf capability gap for 3D drawings.

## What Changes

- Add an `acis` package that decodes and encodes SAT text and SAB binary streams: header, records, typed references, attributes, and end markers.
- Model ACIS bodies, lumps, shells, faces, loops, coedges, edges, and vertices with transforms, bounding boxes, and validation.
- Load ACIS geometry from `3DSOLID`, `REGION`, and `BODY` entities on demand while keeping the original payload for lossless round trip.
- Create and export simple ACIS bodies from supported mesh input, matching ezdxf's documented "simple known structures" scope rather than implementing a solid-modeling kernel.
- Generate triangle meshes from ACIS bodies for rendering, export, and the acisbrowser add-on.
- Render ACIS-backed solids in scene output, using faces from `mesh_from_body`, with explicit diagnostics for unsupported constructs.
- Report unsupported ACIS entity types, invalid references, and history constructs without panicking; unsupported content remains retained verbatim.

## Capabilities

### New Capabilities

- `acis-geometry`: parsing, modeling, meshing, creation, and export of simple ACIS SAT/SAB bodies used by DXF solid entities.

### Modified Capabilities

- `entity-model`: ACIS payloads gain decoded body geometry while preserving the original payload for round trip.
- `path-render`: ACIS-backed solid entities render as face geometry.
- `addons`: the ACIS browser and mesh export add-ons expose body geometry and generated meshes.

## Impact

Affected code: new `acis` package; `entity/body.go`, `entity/solid3d.go`, `entity/acis_metadata.go`; `path`, `render`, `addons/acisbrowser`, `addons/meshex`, and tests. New parsing code carries malformed-input risk and must be covered by fuzzing. Existing payload-only behavior remains available, so files with unsupported ACIS content still round-trip unchanged. No public API removal is planned.
