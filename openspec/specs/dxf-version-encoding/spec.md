# dxf-version-encoding Specification

## Purpose
Defines how the library detects, normalizes, converts, and writes DXF releases
and how it encodes text in ASCII and binary files across those releases.

## Requirements

### Requirement: Source version detection

The loader SHALL determine the DXF version from the header `$ACADVER` value and
SHALL expose it on the loaded document.

#### Scenario: Detect R2018

- **WHEN** a file whose header declares `AC1032` is loaded
- **THEN** the document reports the R2018 version

#### Scenario: Missing version

- **WHEN** a file has no `$ACADVER` value
- **THEN** the loader applies the documented fallback version and reports that
  it was inferred

### Requirement: Version normalization

The loader SHALL normalize versions the library does not write: pre-R12 input
to R12 and R13/R14 input to R2000, and SHALL report the normalization.

#### Scenario: R14 input

- **WHEN** an R14 (`AC1014`) file is loaded
- **THEN** the document version is R2000 and the load report records the upgrade

### Requirement: Write a requested version

The writer SHALL accept a target version from the supported set (R12, R2000,
R2004, R2007, R2010, R2013, R2018) and emit a file that declares that version.

#### Scenario: Write R12

- **WHEN** a document is saved with target version R12
- **THEN** the output header declares `AC1009` and the file parses as R12

#### Scenario: Unsupported target

- **WHEN** a caller requests an unsupported target version
- **THEN** the save fails with a clear error and does not write a partial file

### Requirement: ASCII and binary equivalence

Saving the same document as ASCII and as binary SHALL produce documents that
load to equivalent content.

#### Scenario: ASCII versus binary

- **WHEN** one document is saved twice, once ASCII and once binary, and both are
  loaded
- **THEN** entity counts, table entries, block names, and object handles match

### Requirement: Binary output is complete

The binary writer SHALL emit HEADER, CLASSES, TABLES, BLOCKS, ENTITIES,
OBJECTS when the version supports them, and the EOF marker.

#### Scenario: Binary object section

- **WHEN** a document containing OBJECTS is saved as binary
- **THEN** the binary output contains an OBJECTS section and ends with EOF

### Requirement: Version-specific text encoding

The writer SHALL encode text as UTF-8 for R2007 and later and using the declared
codepage for earlier versions, and the loader SHALL decode accordingly.

#### Scenario: Unicode text in R2018

- **WHEN** an entity contains non-ASCII text and is saved as R2018
- **THEN** the text is stored as UTF-8 and loads back unchanged

#### Scenario: Codepage text in R2000

- **WHEN** a document declares a codepage and contains characters in that
  codepage and is saved as R2000
- **THEN** the output uses that codepage and loads back unchanged

### Requirement: R12 restrictions

When writing R12 the writer SHALL omit sections and fields that R12 does not
support and SHALL record each omission in a conversion report.

#### Scenario: R12 downgrade

- **WHEN** an R2018 document with OBJECTS and CLASSES is saved as R12
- **THEN** the output has no OBJECTS or CLASSES section and the report lists the
  dropped content

### Requirement: Version conversion is reversible when possible

Upgrading a document to a newer version and back SHALL preserve content that
both versions support.

#### Scenario: Upgrade and downgrade

- **WHEN** an R2000 document is converted to R2018 and back to R2000
- **THEN** supported entities, tables, and blocks are unchanged
