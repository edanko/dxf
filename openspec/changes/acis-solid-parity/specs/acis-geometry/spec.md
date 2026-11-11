# Spec Delta

## Purpose

Defines ACIS SAT and SAB parsing, topology modeling, meshing, and export so DXF solid entities can be inspected, rendered, and exchanged as geometry.

## ADDED Requirements

### Requirement: ACIS streams decode into typed records
The system SHALL decode supported SAT text and SAB binary streams into typed header, record, reference, attribute, and end-marker data.

#### Scenario: Valid SAT stream
- **WHEN** a SAT stream with a supported header and body records is parsed
- **THEN** the result exposes the version, record count, body count, typed records, and resolved references without error

#### Scenario: Valid SAB stream
- **WHEN** a SAB stream with a supported header and body records is parsed
- **THEN** the result exposes the same typed topology as the equivalent SAT stream

#### Scenario: Malformed stream
- **WHEN** a stream contains a truncated record or an invalid reference
- **THEN** parsing returns an error naming the record or reference and does not panic

### Requirement: ACIS topology exposes bodies and references
The system SHALL expose bodies, lumps, shells, faces, loops, coedges, edges, and vertices with resolved topology references, bounding boxes, and validation.

#### Scenario: Simple body tree
- **WHEN** a body containing one lump, one shell, and planar faces is loaded
- **THEN** callers can traverse from the body to its faces, loops, coedges, edges, and vertices, and obtain a finite bounding box

#### Scenario: Broken reference
- **WHEN** a face refers to a missing loop
- **THEN** validation reports the broken reference and the affected entity without panicking

### Requirement: Simple ACIS bodies can be created and exported
The system SHALL create simple ACIS bodies from supported triangle mesh input and export them as SAT or SAB data that the same parser can read back.

#### Scenario: Export a closed box
- **WHEN** a closed box mesh is converted to an ACIS body and exported as SAT
- **THEN** parsing the exported stream reproduces one body with the same face and vertex counts

#### Scenario: Unsupported mesh topology
- **WHEN** mesh faces cannot form a closed supported body
- **THEN** creation returns an explicit error identifying the unsupported topology

### Requirement: ACIS bodies produce triangle meshes
The system SHALL triangulate supported ACIS faces into an indexed triangle mesh with consistent outward orientation.

#### Scenario: Cube body meshing
- **WHEN** a cube body with six planar quadrilateral faces is meshed
- **THEN** the mesh contains twelve triangles and every triangle has a finite normal

#### Scenario: Unsupported face
- **WHEN** a face uses a surface form that cannot be triangulated
- **THEN** meshing reports the face and reason while returning the supported triangles

### Requirement: Unsupported ACIS content is reported without data loss
The system SHALL report unsupported ACIS entity types, history constructs, and attributes through structured diagnostics while retaining the original payload verbatim.

#### Scenario: History construct retained
- **WHEN** a body payload contains an unsupported history record
- **THEN** the diagnostic names the unsupported record and saving the entity re-emits the original payload unchanged
