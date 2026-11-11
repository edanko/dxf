# Spec Delta

## ADDED Requirements

### Requirement: Dynamic block references resolve to definitions
The system SHALL resolve an INSERT that references a dynamic block, directly or through an anonymous block, to its dynamic block definition and expose read-only dynamic metadata.

#### Scenario: Direct dynamic reference
- **WHEN** an INSERT references a dynamic block definition
- **THEN** lookup returns that definition and its dynamic metadata

#### Scenario: Anonymous intermediate block
- **WHEN** an INSERT references an anonymous block that references a dynamic block
- **THEN** lookup follows the anonymous block and returns the dynamic definition

#### Scenario: Ordinary block
- **WHEN** an INSERT references a non-dynamic block
- **THEN** lookup returns no dynamic definition without error

### Requirement: User records round-trip structured XRECORD data
The system SHALL compile and parse named `UserRecord` data supporting strings, integers, floats, points, lists, and dictionaries stored in XRECORD tags.

#### Scenario: Nested record round trip
- **WHEN** a user record containing a nested list and dictionary is compiled and parsed
- **THEN** the parsed value equals the original structure

#### Scenario: Invalid string
- **WHEN** a user record string contains a newline or exceeds the documented limit
- **THEN** compilation returns an explicit error

### Requirement: Binary records round-trip arbitrary bytes
The system SHALL compile and parse named `BinaryRecord` payloads as arbitrary bytes in XRECORD data.

#### Scenario: Binary round trip
- **WHEN** a binary record containing all byte values is compiled and parsed
- **THEN** the parsed bytes equal the input

### Requirement: Proxy graphics decode documented primitive opcodes
The system SHALL decode proxy graphics into primitives for the documented line, arc, circle, polyline, text, color, and transform opcodes, in addition to embedded entity records.

#### Scenario: Primitive proxy graphics
- **WHEN** a proxy entity contains a line, circle, and text primitive
- **THEN** extraction returns those primitives with their decoded positions and attributes

#### Scenario: Unknown opcode
- **WHEN** a proxy stream contains an unknown opcode
- **THEN** extraction reports the opcode and byte offset and returns the primitives decoded before it
