# Spec Delta

## ADDED Requirements

### Requirement: Text layout uses measured metrics
The system SHALL use the text-size APIs for line width, wrapping decisions, alignment, and extents when converting TEXT and MTEXT to paths and scene geometry.

#### Scenario: Wrapped MTEXT width
- **WHEN** MTEXT wraps because the measured line exceeds its column width
- **THEN** each rendered line fits the column and the paragraph height matches the measured extents

#### Scenario: Alignment uses measured width
- **WHEN** center- or right-aligned text is laid out
- **THEN** its anchor is offset using the measured advance rather than a fixed character estimate

### Requirement: Scene hatching expands parsed pattern definitions
The system SHALL expand predefined and parsed custom hatch pattern definitions into clipped line segments and report a named pattern that cannot be resolved.

#### Scenario: Custom pattern expansion
- **WHEN** a hatch references a parsed custom pattern and a closed boundary
- **THEN** the scene contains clipped pattern segments inside the boundary

#### Scenario: Unknown pattern name
- **WHEN** a hatch references a pattern name that is not defined
- **THEN** the scene reports the entity handle and missing pattern name and keeps the boundary stroke
