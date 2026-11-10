# color-enum-units Specification

## Purpose

Defines the shared value types for colors, transparency, DXF enumerations, and
units, and the rules for resolving and converting them.

## Requirements

### Requirement: ACI color table

The library SHALL provide the full AutoCAD Color Index table with RGB values for
indices 1 through 255 and SHALL look up the nearest index for an RGB color.

#### Scenario: Known index

- **WHEN** ACI index 1 is resolved
- **THEN** the result is the documented red RGB value

#### Scenario: Nearest index

- **WHEN** an arbitrary RGB color is mapped to ACI
- **THEN** the returned index has the smallest documented color distance

### Requirement: Color kinds

A color value SHALL represent ACI, true color, BYLAYER, and BYBLOCK, and SHALL
report which kind it is.

#### Scenario: BYLAYER

- **WHEN** an entity's color is BYLAYER
- **THEN** the value reports the BYLAYER kind and resolves from its layer

#### Scenario: True color round trip

- **WHEN** a true-color value is saved and loaded
- **THEN** the same RGB value is returned

### Requirement: Color resolution

The library SHALL resolve effective colors using entity, block, and layer context,
including nested blocks.

#### Scenario: Resolve BYBLOCK

- **WHEN** an entity inside a block uses BYBLOCK and the insert has a color
- **THEN** the resolved color is the insert's color

#### Scenario: Resolve BYLAYER

- **WHEN** an entity uses BYLAYER and its layer has a color
- **THEN** the resolved color is the layer's color

### Requirement: Transparency

The library SHALL represent transparency as either an alpha value or a
transparency method and SHALL preserve it on round trip.

#### Scenario: Alpha transparency

- **WHEN** an entity with 50 percent alpha is saved and loaded
- **THEN** the alpha value is preserved

### Requirement: DXF enumerations

The library SHALL provide typed enumerations for the DXF values used by entities,
tables, and objects, with string and numeric conversion.

#### Scenario: Enum round trip

- **WHEN** an enum value is written and read back
- **THEN** the same typed value is returned

#### Scenario: Unknown enum value

- **WHEN** an unrecognized numeric enum value is loaded
- **THEN** it is retained without error and reported as unknown

### Requirement: Units

The library SHALL define the supported drawing units, map them to `$INSUNITS`,
and convert lengths between them.

#### Scenario: Convert millimeters to meters

- **WHEN** a length in millimeters is converted to meters
- **THEN** the result is the length divided by 1000 within tolerance

#### Scenario: Unitless drawing

- **WHEN** a drawing declares unitless
- **THEN** conversions are rejected or treated as identity according to the
  documented policy

### Requirement: Linetype and lineweight defaults

The library SHALL provide standard linetype names, lineweight values, and their
defaults, and SHALL validate them on load.

#### Scenario: Default lineweight

- **WHEN** an entity omits lineweight
- **THEN** the documented default is returned

### Requirement: Validation

Loading and saving SHALL validate color, enum, and unit values and SHALL report
invalid values without silently substituting them.

#### Scenario: Invalid ACI

- **WHEN** an entity declares an out-of-range ACI index
- **THEN** the loader reports the invalid value and preserves the raw value
