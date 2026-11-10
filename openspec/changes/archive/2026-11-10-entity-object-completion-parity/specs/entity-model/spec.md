# Spec Delta

## ADDED Requirements

### Requirement: ACAD_TABLE content is typed

The system SHALL expose ACAD_TABLE rows, columns, cells, table style, and merge information through typed accessors while preserving unknown tags for round-trip output.

#### Scenario: Read an ACAD_TABLE with cells

- **WHEN** a drawing containing an ACAD_TABLE with text cells is loaded
- **THEN** the table exposes its row and column count, cell text, and cell coordinates
- **AND** saving the drawing preserves the original table record.

#### Scenario: Unknown table tags survive

- **WHEN** an ACAD_TABLE contains a group code the typed model does not interpret
- **THEN** the tag is retained and emitted unchanged on save.

### Requirement: Subentity variants are typed

The system SHALL model POLYFACE and POLYMESH VERTEX variants and MESHVERTEXCACHE records with their documented attributes and link them to their owning entity.

#### Scenario: Polyface mesh vertices

- **WHEN** a POLYLINE with the polyface flag is loaded
- **THEN** its vertices expose face records and vertex indices
- **AND** the parent polyline reports them as subentities.

#### Scenario: Mesh vertex cache

- **WHEN** a MESH with a vertex cache is loaded
- **THEN** the cache is available as a typed object linked to the mesh.

### Requirement: ACIS payloads are decoded

The system SHALL decode SAT text and SAB binary payloads sufficiently to report ACIS version, record count, and body presence, and SHALL preserve the complete payload for round-trip output.

#### Scenario: SAT body metadata

- **WHEN** a 3DSOLID with a SAT payload is loaded
- **THEN** the entity reports the ACIS version and a non-zero body count
- **AND** the original payload is emitted unchanged on save.

#### Scenario: SAB binary payload

- **WHEN** a 3DSOLID with a SAB payload is loaded
- **THEN** the payload is retained byte-for-byte and the entity reports that binary ACIS data is present.

### Requirement: Objects in ENTITIES are typed

The system SHALL load object-like records that appear in the ENTITIES section as typed entities when a typed model exists, and SHALL retain them unchanged otherwise.

#### Scenario: Layout record in ENTITIES

- **WHEN** a LAYOUT record is found in the ENTITIES section
- **THEN** it is exposed through the typed layout entity model
- **AND** it still round-trips as an ENTITIES record.

## ADDED Requirements

### Requirement: Extension dictionaries and reactors cover typed records

The system SHALL preserve extension dictionaries and persistent reactors on all newly typed entity and object records.

#### Scenario: Round trip extension data

- **WHEN** a newly typed record carries an extension dictionary or reactor handles
- **THEN** those handles are readable through the typed API
- **AND** they are emitted unchanged on save.
