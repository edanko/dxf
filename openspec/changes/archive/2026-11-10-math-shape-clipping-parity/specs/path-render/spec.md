# Spec Delta

## ADDED Requirements

### Requirement: Path nesting is available

The system SHALL classify closed paths into outer boundaries and nested holes and SHALL expose the nesting tree.

#### Scenario: Nested rings

- **WHEN** two closed paths form concentric rings
- **THEN** nesting reports one outer path with one hole
- **AND** the hole orientation is normalized for filling.
