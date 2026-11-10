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
