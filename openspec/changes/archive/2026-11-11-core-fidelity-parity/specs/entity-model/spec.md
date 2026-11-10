# Spec Delta

## ADDED Requirements

### Requirement: Typed attributes match documented group codes
Every typed entity SHALL expose named access for its documented group codes, apply documented defaults, and gate version-specific fields by the target DXF version.

#### Scenario: Golden attribute round trip
- **WHEN** a representative LINE, SPLINE, HATCH, and ACAD_TABLE record is loaded, modified through typed access, and saved
- **THEN** every documented field appears under its correct group code with the expected default or modified value

#### Scenario: Version-gated field is omitted
- **WHEN** an entity field introduced after the save target is set and the document is saved to an older version
- **THEN** the field is omitted or converted according to the version profile and the conversion is reported

### Requirement: Copy isolates mutable state
A deep copy of an entity SHALL duplicate owned subentities, payloads, and embedded records so that changing the copy does not change the source.

#### Scenario: Embedded entity copy is independent
- **WHEN** an entity with an embedded entity or subentity collection is copied and the copy is modified
- **THEN** the original entity and its embedded state remain unchanged

#### Scenario: ACIS payload copy is independent
- **WHEN** an ACIS-backed entity is copied and the copy's payload is edited
- **THEN** the original payload bytes and metadata remain unchanged

### Requirement: Bounding boxes cover every entity category
Every supported entity category, including placeholders, proxies, underlays, tables, and object-like records, SHALL return a bounding box that reflects its drawable geometry or an explicitly documented empty result.

#### Scenario: Proxy bounding box is derived
- **WHEN** a proxy entity with decodable graphics is queried for its bounding box
- **THEN** the result encloses the decoded geometry instead of a zero-size box

#### Scenario: Empty entity reports empty box
- **WHEN** an entity has no drawable geometry
- **THEN** its bounding box is reported as empty through the documented empty representation
