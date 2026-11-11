# Spec Delta

## Purpose

Defines the high-level document authoring API: spaces, factories, blocks,
layouts, groups, the entity database, and selection helpers.

## ADDED Requirements

### Requirement: Drawing exposes document spaces

A loaded or new drawing SHALL expose model space, paper space, blocks, layouts,
groups, objects, and the entity database as first-class accessors.

#### Scenario: Model space entities

- **WHEN** a new drawing is created and an entity is added to model space
- **THEN** model space contains that entity and the entity's owner references
  the model-space block record

#### Scenario: Paper space

- **WHEN** a layout is created and an entity is added to its paper space
- **THEN** the entity belongs to that layout's block record, not model space

### Requirement: Factory adds entities

The graphics factory SHALL provide an add operation for every supported entity
type, SHALL assign a handle and owner, and SHALL return the created entity.

#### Scenario: Add a line

- **WHEN** `AddLine` is called with two points
- **THEN** a LINE with a unique handle is added to the target space and returned

#### Scenario: Add with attributes

- **WHEN** an entity is added with layer, color, and linetype options
- **THEN** the created entity carries those attributes

### Requirement: Block definitions and references

The drawing SHALL support creating block definitions, adding entities to them,
and inserting references with position, scale, rotation, and attributes.

#### Scenario: Create and insert a block

- **WHEN** a block named `Door` is created with a line and inserted at a point
- **THEN** the INSERT references `Door` and its bounding box includes the
  transformed line

#### Scenario: Attribute values

- **WHEN** an insert of a block with attributes is created with values
- **THEN** the ATTRIB entities carry those values and are accessible from the
  insert

### Requirement: Layout management

The drawing SHALL support creating, renaming, deleting, and activating layouts,
and SHALL keep their block records consistent.

#### Scenario: New layout

- **WHEN** a layout named `Sheet1` is created
- **THEN** a corresponding block record and LAYOUT object exist and are linked

#### Scenario: Delete layout

- **WHEN** a layout is deleted
- **THEN** its entities and block record are removed and references are cleared

### Requirement: Handles and ownership are assigned automatically

Every authored entity, table entry, and object SHALL receive a unique handle and
the correct owner without manual bookkeeping.

#### Scenario: Many entities

- **WHEN** one hundred entities are added
- **THEN** all handles are unique and every owner resolves to an existing object

### Requirement: Groups

The drawing SHALL support creating groups and adding or removing entities while
maintaining group membership on both sides.

#### Scenario: Group membership

- **WHEN** an entity is added to a group
- **THEN** the group lists the entity and the entity reports the group

### Requirement: Selection helpers

The drawing SHALL support selecting entities by type, layer, and spatial extent,
and SHALL return stable results.

#### Scenario: Select by layer

- **WHEN** entities on layers `A` and `B` exist and a query selects layer `A`
- **THEN** only entities on layer `A` are returned

### Requirement: Extents

The drawing SHALL compute model-space and per-layout extents from entity
bounding boxes and SHALL allow explicit recalculation.

#### Scenario: Recalculate extents

- **WHEN** geometry changes and extents are recalculated
- **THEN** the reported extents enclose all model-space geometry

### Requirement: Transactions

Authoring operations SHALL be groupable in a transaction that can be committed
or rolled back, restoring the prior document state on rollback.

#### Scenario: Rollback

- **WHEN** entities are added inside a transaction and the transaction is rolled
  back
- **THEN** the drawing contains none of those entities and handles are reusable
