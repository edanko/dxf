# Proposal

## Why

Loading a DXF file currently reconstructs only the ENTITIES section. HEADER
variables, TABLES, BLOCKS, CLASSES, and OBJECTS are parsed into an intermediate
loader but never applied to the document, and unknown tags are dropped. This
makes it impossible to read and re-save a real-world drawing without data loss,
and it blocks every higher-level feature that depends on a complete document
model.

## What Changes

- Bind all parsed sections into `drawing.Drawing` on load: HEADER, CLASSES,
  TABLES, BLOCKS, ENTITIES, OBJECTS.
- Preserve raw tags for unknown entities, unknown object types, and unknown
  group codes so they survive load/save round trips.
- Implement the OBJECTS section binding instead of the current no-op
  `ParseObjects`.
- Preserve comments (code 999) and section ordering metadata.
- Preserve THUMBNAILIMAGE and ACDSDATA sections as opaque tag blocks.
- Add a recovery-oriented load mode that tolerates truncated or inconsistent
  files and reports recovered structures.
- Make load → save → load semantically stable for the existing `testdata/`
  corpus.
- Add round-trip and golden-tag tests for the core document path.

## Capabilities

### New Capabilities

- `dxf-io-core`: complete DXF document loading, binding, preservation, and
  round-trip behavior across all sections.

### Modified Capabilities

None. The project has no existing main specs.

## Impact

Affected code: `dxf_main.go`, `parser.go`, `lldxf/*`, `drawing/*`, `header/*`,
`table/*`, `block/*`, `object/*`, `class/*`, and `entity/*`. Public entry points
`dxf.FromReader`, `dxf.FromFile`, `dxf.FromStringData` change behavior from
"entities only" to "complete document". Existing callers that rely on entities
only continue to work. New tests and fixtures are added under the package test
files and `testdata/`.
