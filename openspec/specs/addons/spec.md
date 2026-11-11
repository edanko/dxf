# addons Specification

## Purpose

Defines the optional add-on packages that build on the core library, from
streaming I/O and drawing to CSG, text, geospatial, mesh, and tool integrations.

## Requirements

### Requirement: Add-on framework

Add-ons SHALL live in separate packages, depend only on the core library and
their documented dependencies, and SHALL be discoverable through a registry where
they expose a named capability.

#### Scenario: Register an add-on

- **WHEN** an add-on registers itself under a name
- **THEN** callers can discover and invoke it without importing every add-on

### Requirement: Drawing add-on

The drawing add-on SHALL render a document to vector and raster backends using
the core renderer, with configurable size, background, and layout selection.

#### Scenario: Render a document

- **WHEN** a document is rendered through the drawing add-on to SVG
- **THEN** the output contains the document's visible geometry

### Requirement: Streaming writer

The R12 streaming writer SHALL write entities incrementally without building a
full document, producing a valid R12 file.

#### Scenario: Stream many entities

- **WHEN** ten thousand entities are written through the streaming writer
- **THEN** the output is a valid R12 file containing all entities

### Requirement: Streaming reader

The streaming reader SHALL iterate entities from a large file without loading the
whole document into memory.

#### Scenario: Iterate entities

- **WHEN** a large file is iterated
- **THEN** entities are yielded in file order with bounded memory

### Requirement: Importer

The importer SHALL convert supported external geometry formats into DXF
entities.

#### Scenario: Import supported format

- **WHEN** a supported external file is imported
- **THEN** the resulting document contains equivalent geometry

### Requirement: Code generation

The code generator SHALL emit source code that recreates a document's entities
and structure.

#### Scenario: Generate and run

- **WHEN** code is generated for a sample document and executed
- **THEN** it produces a document with the same entity inventory

### Requirement: Plot style tables

The plot-style add-on SHALL parse CTB/STB tables and apply their color, lineweight,
and screening rules during rendering.

#### Scenario: Apply a CTB

- **WHEN** a CTB table is applied to a rendered drawing
- **THEN** entity colors and lineweights follow the table's rules

### Requirement: CSG

The CSG add-on SHALL provide mesh primitives and boolean union, difference, and
intersection that produce valid meshes.

#### Scenario: Boolean difference

- **WHEN** a box is subtracted from a larger box
- **THEN** the result mesh encloses the expected volume and has no inverted faces

### Requirement: MTextExplode

The MTextExplode add-on SHALL convert MTEXT content into positioned TEXT entities
honoring line breaks, alignment, and basic formatting.

#### Scenario: Explode MTEXT

- **WHEN** a multi-line MTEXT is exploded
- **THEN** each line becomes a TEXT entity at the correct position

### Requirement: Text to path

The text-to-path add-on SHALL convert text entities to path outlines using real
font data.

#### Scenario: Convert text

- **WHEN** a TEXT entity is converted to paths
- **THEN** the paths form the glyph outlines

### Requirement: Geospatial export

The geo add-on SHALL export supported entities to GeoJSON and related coordinate
formats with a configurable CRS transform.

#### Scenario: Export GeoJSON

- **WHEN** a drawing with lines and points is exported to GeoJSON
- **THEN** the output is valid GeoJSON containing those features

### Requirement: Mesh and OpenSCAD export

The mesh and OpenSCAD add-ons SHALL export supported 3D entities and meshes to
standard mesh formats and OpenSCAD source.

#### Scenario: Export a mesh

- **WHEN** a MESH is exported to a mesh format
- **THEN** the output contains the mesh vertices and faces

### Requirement: External tool integration

The ODA converter and HPGL/2 add-ons SHALL invoke their external tools when
available and SHALL return a clear error when they are not.

#### Scenario: Missing external tool

- **WHEN** an add-on requiring an external executable is used and the executable
  is absent
- **THEN** the call fails with an actionable error and does not corrupt output

### Requirement: Animation, packing, table, and PCB helpers

The xplayer, binpacking, tablepainter, gerber, and mtxpl add-ons SHALL provide
their documented transformations and SHALL preserve document validity.

#### Scenario: Pack rectangles

- **WHEN** a set of rectangles is packed into a bin
- **THEN** no rectangle overlaps and all fit within the bin bounds

### Requirement: Standards rule engine is reusable

The system SHALL expose a standards rule engine independent of the CLI, with structured violation reports and configurable severities.

#### Scenario: Programmatic rule run

- **WHEN** a caller runs the standards engine on a drawing
- **THEN** it receives violations with code, severity, and object handle.

### Requirement: ZIP manager preserves entries

The system SHALL create and extract ZIP archives while preserving entry names and file contents.

#### Scenario: Archive extraction

- **WHEN** an archive is extracted
- **THEN** each entry is written at its recorded path
- **AND** path traversal outside the target directory is rejected.

### Requirement: Browser data extraction is available

The system SHALL expose document, entity, and ACIS inspection data in a form usable by a GUI or CLI browser without requiring a GUI dependency.

#### Scenario: Entity browser data

- **WHEN** a caller requests browser data for a document
- **THEN** entities are grouped by type and layout
- **AND** each entry exposes handle, type, layer, and bounding box.

### Requirement: DWG conversion delegates to an external tool

The system SHALL convert DWG to DXF through a configurable external converter and SHALL return an actionable error when the tool is missing.

#### Scenario: Missing converter

- **WHEN** DWG conversion is requested without a configured converter
- **THEN** the operation fails with an error naming the required tool
- **AND** no partial output file is left behind.

### Requirement: Geometry generator add-ons are available

The system SHALL provide dimension-line helpers, r12export, genetic algorithm, Menger sponge, and Sierpinski pyramid add-ons with documented inputs and outputs.

#### Scenario: Menger sponge generation

- **WHEN** a Menger sponge is generated at a given depth
- **THEN** the expected number of cube/face entities is produced
- **AND** the result can be added to a drawing.

### Requirement: Mixins match ezdxf behavior

The system SHALL provide the documented mixin behaviors for add-on entities and objects.

#### Scenario: Mixin application

- **WHEN** a mixin is applied to a compatible entity
- **THEN** the documented attributes and methods are available
- **AND** existing entity data is unchanged.

### Requirement: Plot styles apply to rendering

The system SHALL parse STB text and binary CTB plot style tables and apply color, lineweight, and screening overrides during rendering.

#### Scenario: Binary CTB

- **WHEN** a binary CTB file is loaded
- **THEN** its plot styles are available
- **AND** rendering applies the configured color and lineweight.

### Requirement: SHX and TTF text metrics are available

The system SHALL provide text metrics for SHX and TTF fonts and SHALL use them for layout and rendering.

#### Scenario: SHX metrics

- **WHEN** a SHX font is loaded
- **THEN** callers can measure text width and height
- **AND** text rendered with that font uses the measured metrics.

### Requirement: ISO hatch patterns are available

The system SHALL provide the ISO hatch pattern library and resolve patterns by name.

#### Scenario: ISO pattern lookup

- **WHEN** a hatch references an ISO pattern name
- **THEN** the pattern definition is found
- **AND** its line segments are emitted by the hatch renderer.

### Requirement: Text size APIs report measured extents
The system SHALL measure TEXT, single-line MTEXT, and laid-out MTEXT content in drawing units using the configured font metrics and layout rules.

#### Scenario: Single-line text size
- **WHEN** a TEXT entity with height and string content is measured
- **THEN** the result reports width, height, ascent, descent, and total advance in drawing units

#### Scenario: Laid-out MTEXT
- **WHEN** an MTEXT entity with wrapping width and multiple paragraphs is measured
- **THEN** the result reports the laid-out width and height including paragraph spacing

### Requirement: Native DWG metadata is readable
The system SHALL read DWG version and release metadata and partially parse supported header, class, and object records without invoking an external converter.

#### Scenario: DWG header metadata
- **WHEN** a DWG file with a supported version signature is inspected
- **THEN** the result reports the version, release name, and parsed section inventory

#### Scenario: Unsupported DWG record
- **WHEN** the DWG parser meets an unsupported record encoding
- **THEN** inspection reports the record and offset while returning metadata decoded before it

### Requirement: Export add-ons use decoded geometry
The system SHALL make ACIS body meshes and linked-edge loops available to mesh and edge export add-ons without duplicating geometry logic.

#### Scenario: Export ACIS body mesh
- **WHEN** a drawing containing a supported 3DSOLID is exported to a mesh format
- **THEN** the export includes the triangles produced from the decoded ACIS body

#### Scenario: Export linked-edge loop
- **WHEN** an edge network forms a closed loop and is exported to a polyline
- **THEN** the exported polyline follows the loop order and orientation
