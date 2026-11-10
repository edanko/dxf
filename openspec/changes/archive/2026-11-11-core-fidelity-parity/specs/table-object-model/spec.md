# Spec Delta

## ADDED Requirements

### Requirement: Table entries preserve unknown content
Every typed table entry SHALL retain unmodelled tags and SHALL expose the documented fields for its record type through typed accessors.

#### Scenario: Unknown DIMSTYLE code survives
- **WHEN** a DIMSTYLE entry containing an unmodelled group code is loaded and saved
- **THEN** the saved entry retains the code and value

#### Scenario: Documented field is accessible
- **WHEN** a typed table entry contains a documented field
- **THEN** the field is readable and writable through the typed API and is emitted under its documented group code

### Requirement: Objects preserve unknown content
Every typed object SHALL retain unmodelled tags and SHALL expose its documented fields, including collection membership and owner references.

#### Scenario: LAYOUT unknown code survives
- **WHEN** a LAYOUT object with an unmodelled group code is loaded and saved
- **THEN** the saved object retains the code and value

#### Scenario: Collection membership round trips
- **WHEN** a collection object is loaded with member handles
- **THEN** the members are exposed through the typed collection and the saved object preserves the same membership
