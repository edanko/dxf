# Spec Delta

## ADDED Requirements

### Requirement: Unhandled tags survive on typed records
Typed table entries, blocks, and objects SHALL retain every tag they do not model and SHALL re-emit those tags on save in their original relative order.

#### Scenario: Unknown table entry code survives
- **WHEN** a LAYER entry containing an unmodelled group code is loaded and saved
- **THEN** the saved entry contains that code with its original value

#### Scenario: Unknown object code survives
- **WHEN** a DICTIONARY object containing an unmodelled group code is loaded and saved
- **THEN** the saved object contains that code with its original value

### Requirement: Comments retain position
Comment records SHALL be preserved at their original position relative to surrounding tags within a section.

#### Scenario: Comment between entities keeps order
- **WHEN** a comment appears between two entities and the document is saved
- **THEN** the comment appears between the same two entities in the output

#### Scenario: Stripping removes all comments
- **WHEN** comments are stripped during load
- **THEN** no comment records appear in the saved document
