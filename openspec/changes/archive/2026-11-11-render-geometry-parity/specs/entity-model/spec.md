# Spec Delta

## ADDED Requirements

### Requirement: Boundary edge variants are typed and transformable
HATCH and MPOLYGON boundary paths SHALL preserve each edge variant as typed data and transform it with the owning entity.

#### Scenario: Transform mixed boundary
- **WHEN** a HATCH with line, arc, ellipse, and spline edges is transformed
- **THEN** each edge retains its type and the transformed boundary matches the transformed entity

#### Scenario: Round trip boundary variants
- **WHEN** a HATCH with all supported edge variants is loaded and saved
- **THEN** every edge variant and its typed parameters round-trip unchanged

### Requirement: Ellipse embedded spline data is typed and preserved
The ELLIPSE model SHALL expose embedded spline control data when present and SHALL preserve it on save.

#### Scenario: Load and save embedded spline
- **WHEN** an ELLIPSE record with embedded spline data is loaded
- **THEN** the spline data is available through the typed model and is re-emitted unchanged

### Requirement: Complex linetype elements preserve shape and text data
LINETYPE table elements SHALL retain shape number, text value, rotation, scale, and offsets alongside dash lengths.

#### Scenario: Round trip complex linetype
- **WHEN** a complex linetype with shape and text elements is loaded and saved
- **THEN** every element and its parameters round-trip unchanged

#### Scenario: Transform placement
- **WHEN** a complex linetype is applied to a transformed path
- **THEN** shape and text placements follow the transformed path orientation
