# Spec Delta

## ADDED Requirements

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
