# Spec Delta

## ADDED Requirements

### Requirement: Tag-level validation and repair
The audit SHALL validate required tags, value ranges, and structural terminators, and fix mode SHALL repair only defects with a deterministic safe repair.

#### Scenario: Missing terminator is repaired
- **WHEN** a block record lacks its required terminator and audit runs with fix enabled
- **THEN** the terminator is inserted at the correct position and a re-audit reports no remaining terminator issue

#### Scenario: Unsafe tag value is reported
- **WHEN** a numeric field contains a value outside its documented range
- **THEN** the audit reports the record handle, code, and offending value without changing it

### Requirement: Duplicate handles are repaired
Fix mode SHALL assign unique handles to duplicate records deterministically and SHALL update references that unambiguously point to the repaired record.

#### Scenario: Duplicate handle becomes unique
- **WHEN** two entities share a handle and audit runs with fix enabled
- **THEN** one entity receives a new unique handle and the report records the change

#### Scenario: Ambiguous references stay reported
- **WHEN** a duplicate handle is referenced by a record whose intended target is ambiguous
- **THEN** the reference is left unchanged and reported as unresolved

### Requirement: Block cycles are repaired
Fix mode SHALL break block-reference cycles without deleting the referenced block definitions.

#### Scenario: Self-reference is broken
- **WHEN** a block contains an INSERT of itself and audit runs with fix enabled
- **THEN** the self-reference is removed and the block definition remains present

#### Scenario: Nested cycle is reported and broken
- **WHEN** block A references B and B references A
- **THEN** the cycle is reported and at least one reference in the cycle is removed so traversal terminates
