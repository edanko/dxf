# Design

## Context

Several ezdxf capabilities are still absent or narrower in Go: disassembly only flattens INSERTs, dynamic blocks have no read helpers, XRECORD user/binary records are not modelled, text measurement is limited to raw font metrics, hatch patterns are hardcoded, proxy graphics cover embedded entities but not primitive opcodes, DWG support is external-only, and linked-edge tools do not exist. See `proposal.md` for the full inventory.

Constraints:

- Additive APIs unless a missing capability requires a new package.
- Read-only helpers must not mutate source entities.
- Large files must remain streamable; no full DWG reconstruction.
- Unsupported data returns structured diagnostics rather than fake geometry.

## Goals / Non-Goals

**Goals:**

- Provide primitive, path, mesh, and control-vertex disassembly for all supported composite entities.
- Add dynamic-block read helpers, XRECORD user/binary records, text-size APIs, custom PAT parsing, proxy primitive decoding, DWG metadata/partial parsing, and linked-edge topology.
- Add small utilities that ezdxf exposes: Julian date, handle reorder, raw section loading.
- Integrate new geometry with rendering and export add-ons.

**Non-Goals:**

- Dynamic block creation or modification.
- Full DWG entity geometry parsing or native DWG writing.
- A full ACIS kernel; ACIS meshing is consumed from `acis-solid-parity`.
- General-purpose graph algorithms beyond linked-edge use cases.

## Decisions

### Primitive decomposition model

Define a `Primitive` value with kind, source handle, vertices, optional path, optional mesh, and diagnostics. `make_primitive` maps each supported entity to one or more primitives. `recursive_decompose` expands INSERT, MTEXT, POLYLINE, HATCH, and other composites through their explode or path representations, with cycle and depth guards. `to_vertices`, `to_paths`, `to_meshes`, and `to_control_vertices` consume the primitive stream.

Unsupported entities yield a diagnostic and no primitive. `POINT` yields a vertex primitive with no path or mesh.

### Dynamic block resolution

Read `AcDbDynamicBlockGUID` and `AcDbBlockRepBTag` from BLOCK_RECORD XDATA/extension data. Follow an anonymous block's block-record reference to its parent definition. Return a read-only descriptor with GUID, block record handle, and definition name; never synthesize or edit dynamic parameters.

### XRECORD user and binary records

`UserRecord` compiles typed values into XRECORD tags using code 302 structure markers, code 1 strings, code 90 integers, code 40 floats, and point codes. Parsing is recursive for lists and dictionaries and rejects invalid nesting, newlines, and overlong strings. `BinaryRecord` stores arbitrary bytes in documented binary tags with a name. Both preserve unrelated XRECORD tags when updating an existing record.

### Text size measurement

Compute single-line and multi-line extents from `text2path` metrics, accounting for height, width factor, rotation, alignment, line spacing, and paragraph wrapping. The API returns width, height, ascent, descent, and the transformed bounding box. Missing fonts fall back to the built-in vector metrics and report the substitution.

### Custom hatch pattern parsing

Parse `.pat` text into pattern definitions: comment and blank handling, name/description lines, angle, base point, offset, and dash items. Errors include file and line. Parsing returns successfully parsed patterns plus diagnostics for bad definitions. Lookup is case-insensitive while preserving original names. Scaling creates a new definition and never mutates the registry. The renderer accepts both predefined and parsed patterns and reports unknown names.

### Proxy graphics primitive decoder

Parse the documented binary stream as bits and bytes, maintaining current color, layer, linetype, and transform. Implement line, arc, circle, polyline, text, color, and transform opcodes. Unknown opcodes report their value and offset and stop or skip according to the documented length rules. Embedded entity records remain a separate fallback path.

### Native DWG metadata and partial parser

Read the DWG version header and section locator first. When the section layout is recognized, parse header variables, classes, and object summaries needed for inspection. Unsupported sections and compressed/encrypted payloads report diagnostics. Entity geometry and writing remain out of scope; the ODA converter remains the supported DWG-to-DXF path.

### Linked-edge topology

Add an `edges` package with an immutable edge type and a graph builder. Endpoints merge by a configurable tolerance using quantized spatial hashing. The graph reports vertex degree, leaves, connected components, open chains, and closed loops. Loop extraction returns ordered edges; entity creation builds LWPOLYLINE/POLYLINE and hatch boundary paths. Open boundaries return an explicit error. Orientation can be reversed per loop.

### Utility additions

Expose Julian-date conversion independent of SUN, stable ascending/descending handle reorder over entity slices, and a raw section loader that returns grouped tags without typed binding.

## Risks / Trade-offs

- [Primitive recursion can explode on large nested blocks] → Add depth, node, and cycle limits and report truncation.
- [Dynamic block metadata varies by authoring system] → Prefer documented keys and return not-found rather than guessing.
- [PAT syntax has vendor extensions] → Support the documented subset, report unsupported tokens, and keep partial results.
- [DWG files are complex and versioned] → Scope to metadata and recognized sections; never claim full geometry support.
- [Text metrics differ from AutoCAD] → Document metric assumptions and expose fallback reasons.
- [Linked-edge tolerance changes topology] → Make tolerance explicit, deterministic, and covered by near-miss tests.

## Migration Plan

New APIs are additive. The DWG add-on keeps the existing external converter; native parsing is a separate inspection path. Hatch pattern lookup gains a registry overlay without removing predefined constants.
