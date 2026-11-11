# Spec Delta

## ADDED Requirements

### Requirement: JSON serialization round-trips supported document state
The system SHALL serialize and reconstruct supported entities, tables, blocks, and objects through `ToJSON` and `FromJSON`, or return an explicit unsupported error instead of reporting success.

#### Scenario: Document round trip
- **WHEN** a drawing with entities, layers, blocks, and typed objects is serialized and loaded back
- **THEN** the reconstructed drawing has the same entity types, handles, layers, block names, and object handles

#### Scenario: Unsupported JSON shape
- **WHEN** `FromJSON` receives data without the required document fields
- **THEN** it returns an error naming the missing or unsupported field and does not modify the drawing

#### Scenario: Explicitly unsupported state
- **WHEN** the document contains state the JSON format cannot represent
- **THEN** `ToJSON` returns an error identifying that state rather than emitting data that loses it silently
