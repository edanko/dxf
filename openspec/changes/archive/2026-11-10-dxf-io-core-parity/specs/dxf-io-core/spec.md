# Spec Delta

## Purpose

Defines how the library loads, binds, preserves, and writes a complete DXF
document so that real-world drawings can be read and modified without data loss.

## ADDED Requirements

### Requirement: Load binds every document section

The loader SHALL parse and bind HEADER, CLASSES, TABLES, BLOCKS, ENTITIES, and
OBJECTS sections into the drawing document, not only ENTITIES.

#### Scenario: Loading a document with tables and blocks

- **WHEN** a DXF file containing layers, linetypes, blocks, and model-space
  entities is loaded
- **THEN** the returned drawing exposes the same layers, linetypes, block
  definitions, and entities that the file declares

#### Scenario: Loading a document with objects

- **WHEN** a DXF file containing a DICTIONARY, XRECORD, or LAYOUT object is
  loaded
- **THEN** those objects are present in the drawing's OBJECTS section and are
  retrievable by handle

### Requirement: Unknown content is preserved

The loader and writer SHALL preserve unknown entities, unknown object types, and
unknown group codes as raw tags so they survive a load/save round trip.

#### Scenario: Unknown entity survives round trip

- **WHEN** a document contains an entity type the library does not model and it
  is loaded then saved
- **THEN** the saved document still contains that entity with its original group
  codes and values

#### Scenario: Unknown group code on a known entity survives

- **WHEN** a known entity contains an unrecognized group code and it is loaded
  then saved
- **THEN** the saved entity still contains that group code and value

### Requirement: Round trip is semantically stable

Loading a valid document, saving it, and loading the saved document SHALL
produce an equivalent document: same entity types and handles, same table
entries, same block names, same object handles, and same unknown tags.

#### Scenario: Round trip of the test corpus

- **WHEN** each fixture in `testdata/` is loaded, saved, and loaded again
- **THEN** the two loaded documents report equal entity counts, table entry
  names, block names, and object handles

### Requirement: Comments are preserved

The loader SHALL retain code 999 comment tags and the writer SHALL emit them in
their original sections unless comment stripping is explicitly requested.

#### Scenario: Comment survives save

- **WHEN** a document containing code 999 comments is loaded and saved without
  stripping
- **THEN** the saved document contains the same comment text

### Requirement: Opaque sections are retained

THUMBNAILIMAGE and ACDSDATA sections SHALL be retained as opaque tag blocks and
written back unchanged when present.

#### Scenario: Thumbnail round trip

- **WHEN** a document with a THUMBNAILIMAGE section is loaded and saved
- **THEN** the saved document contains an equivalent THUMBNAILIMAGE section

### Requirement: Recovery mode tolerates damaged input

A recovery load mode SHALL load as much valid content as possible from a
truncated or inconsistent file, SHALL NOT panic, and SHALL report the recovered
and skipped structures.

#### Scenario: Truncated file

- **WHEN** a DXF file is truncated in the middle of an entity
- **THEN** recovery mode returns a drawing with the preceding complete content
  and a non-empty recovery report

#### Scenario: Missing ENDSEC

- **WHEN** a section is missing its ENDSEC marker
- **THEN** recovery mode closes the section, keeps its content, and reports the
  repair

### Requirement: Section ordering and handles are maintained

The writer SHALL emit sections in DXF order (HEADER, CLASSES, TABLES, BLOCKS,
ENTITIES, OBJECTS, optional THUMBNAILIMAGE/ACDSDATA) and SHALL assign unique
handles to all newly created objects.

#### Scenario: Saved file has valid section order

- **WHEN** a new document is created and saved
- **THEN** section markers appear in DXF order and every handle in the file is
  unique
