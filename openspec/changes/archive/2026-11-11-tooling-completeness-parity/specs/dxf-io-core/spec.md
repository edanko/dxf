# Spec Delta

## ADDED Requirements

### Requirement: Raw section loading exposes grouped tags
The system SHALL load ASCII or binary DXF content into named sections of ordered raw tags without constructing or binding a document.

#### Scenario: Raw sections
- **WHEN** a DXF file containing HEADER, TABLES, and ENTITIES sections is loaded raw
- **THEN** each section is returned with its tags in source order

#### Scenario: Truncated raw section
- **WHEN** the input ends before an ENDSEC
- **THEN** raw loading returns the tags read so far with an error naming the open section
