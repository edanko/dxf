# Spec Delta

## Purpose

Defines parsing, lookup, scaling, and expansion of custom hatch pattern definition files so drawings can use patterns beyond the built-in libraries.

## ADDED Requirements

### Requirement: PAT files parse into pattern definitions
The system SHALL parse `.pat` text into named pattern definitions containing description, angle, base point, offset, and dash items.

#### Scenario: Multiple patterns
- **WHEN** a file contains two pattern blocks
- **THEN** lookup by name returns two definitions with their parsed line data

#### Scenario: Malformed line
- **WHEN** a pattern line has an invalid number of fields
- **THEN** parsing reports the file and line number and returns the patterns that parsed successfully

### Requirement: Pattern lookup supports case-insensitive names
Pattern lookup SHALL match names case-insensitively while preserving the definition's original name.

#### Scenario: Case-insensitive lookup
- **WHEN** a definition is stored as `Brick` and lookup uses `brick`
- **THEN** the definition is returned with its original name

### Requirement: Patterns scale by factor and measurement system
The system SHALL scale pattern geometry by a caller factor and convert between imperial and metric base scales without mutating the stored definition.

#### Scenario: Scale by factor
- **WHEN** a pattern is requested with scale 2
- **THEN** every base point, offset, and dash length is doubled in the returned definition

#### Scenario: Original remains unchanged
- **WHEN** a scaled definition is returned
- **THEN** the stored definition retains its original values

### Requirement: Patterns expand to hatch line segments
The hatch renderer SHALL expand a selected pattern definition into clipped line segments using the hatches scale, angle, and boundary.

#### Scenario: Pattern inside a boundary
- **WHEN** a rectangular hatch uses a custom pattern with known line spacing
- **THEN** the emitted segments lie inside the boundary and repeat at the scaled spacing

#### Scenario: Unknown pattern name
- **WHEN** a hatch references a pattern name that is not defined
- **THEN** rendering reports the hatch handle and missing pattern name and emits the boundary only
