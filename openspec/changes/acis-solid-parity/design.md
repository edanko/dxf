# Design

## Context

DXF solid entities currently retain SAT/SAB payloads and expose shallow metadata such as format, version, record count, and body count. There is no package that parses records into topology, no mesh extraction, and no way to create or export ACIS bodies. The `acisbrowser` and `meshex` add-ons therefore expose payloads rather than geometry. See `proposal.md` for the gap analysis.

Constraints:

- Original payloads must survive load/save, particularly for unsupported records.
- The parser must be safe on malformed and adversarial input.
- No external ACIS kernel or native dependency may be introduced.
- Existing entity APIs and rendering remain usable when parsing fails.

## Goals / Non-Goals

**Goals:**

- Decode supported SAT and SAB streams into a shared typed record and topology model.
- Expose bodies, faces, loops, coedges, edges, vertices, transforms, bounds, and validation.
- Create simple bodies from triangle meshes and export them as SAT/SAB.
- Mesh supported bodies into indexed triangles for rendering and export.
- Integrate decoded geometry with 3DSOLID, REGION, BODY, path conversion, scene rendering, and add-ons.

**Non-Goals:**

- A general ACIS modeling kernel or arbitrary history-tree editing.
- Boolean operations between ACIS bodies.
- Exact reproduction of unsupported analytic surfaces.
- Editing arbitrary existing ACIS structures.

## Decisions

### Independent `acis` package

The package owns stream parsing, topology, creation, export, meshing, and diagnostics. It depends only on `math`; DXF entities use it through a narrow adapter. This keeps malformed ACIS failures out of the document model.

Alternative: parse inside `entity/body.go`. Rejected because the parser is large and needs independent fuzzing and tests.

### SAT and SAB front ends, shared record graph

SAT text and SAB binary each produce the same `Record` sequence with type, id, references, and ordered attributes. A normalization pass resolves pointer syntax, named references, and end markers into typed topology. Unknown record types remain as `RawRecord` nodes with their attributes.

### Topology with explicit reference resolution

`Body`, `Lump`, `Shell`, `Face`, `Loop`, `Coedge`, `Edge`, and `Vertex` are concrete types. References are stored as IDs and resolved during load; unresolved references produce diagnostics but do not abort the whole body. Each face exposes its surface form and loops, and each loop exposes ordered coedges.

### On-demand cache keyed by payload

Entities parse lazily when decoded geometry is requested. The cache key is a hash of the raw payload plus format, so repeated calls are cheap and an edited payload invalidates the cache. The cache stores the parsed bodies and diagnostics.

### Payload authority and re-export policy

For payloads with no unsupported records, transforming decoded geometry marks the payload stale and saving re-exports the decoded bodies. For payloads with unsupported records, the original payload is authoritative: transforms update the in-memory geometry but saving keeps the original payload and reports that the geometry change was not serialized. This avoids destroying data the parser cannot reproduce.

### Mesh extraction via planar triangulation

Faces expose their boundary loops. Convex or simple planar loops are triangulated with the existing earcut implementation in `math`, after projecting to the face plane. Triangle orientation is corrected against the face normal and shell orientation. Curved or unsupported surfaces produce diagnostics and contribute no triangles.

### Simple body creation from meshes

Creation accepts an indexed triangle mesh with optional welded vertices. It builds one body, one lump, one shell, one planar face per triangle, and shared edge/vertex records. Export writes a minimal supported SAT or SAB stream with ACIS 700-compatible headers. The caller receives an error when indices are invalid, faces are degenerate, or the topology cannot be represented.

### Diagnostics as data

Parsing and meshing return a diagnostic list with entity ID, record type, reason, and severity. Unsupported history records and surface forms are warnings; malformed references and truncated streams are errors. Diagnostics are exposed through the browser and rendering fallback reports.

## Risks / Trade-offs

- [ACIS dialects vary across authoring systems] → Support the documented subset, keep unknown records raw, and expand via fixtures rather than guessing.
- [Re-exporting transformed supported bodies can drop unmodeled semantics] → Track unsupported records and keep the original payload when any exist.
- [Meshing curved faces is approximate or unsupported] → Report unsupported surfaces and return only verified triangles.
- [Malformed input can cause pathological allocation] → Bound record counts, reference fan-out, and payload sizes; fuzz the parsers.
- [Lazy parsing adds synchronization complexity] → Keep caches per entity or guard shared caches with a mutex and immutable results.

## Migration Plan

New APIs are additive. Existing metadata and raw-payload access remain. Entities with unsupported ACIS content continue to render through their previous fallback behavior.
