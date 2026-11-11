# Spec Delta

## ADDED Requirements

### Requirement: ACIS solids render as face geometry
The renderer SHALL emit ACIS body faces as closed paths, using fill or wireframe according to the active scene style.

#### Scenario: Render a supported solid
- **WHEN** a scene contains a 3DSOLID with supported decoded geometry
- **THEN** the scene contains one closed path per mesh triangle with the entity style applied

#### Scenario: Unsupported ACIS content
- **WHEN** a solid contains an unsupported face or history construct
- **THEN** the supported faces still render and the scene fallback report names the entity and unsupported record

### Requirement: ACIS path conversion uses decoded geometry
Path conversion for 3DSOLID, REGION, and BODY SHALL use decoded ACIS bodies when available and SHALL report when only retained payload is available.

#### Scenario: Path from a solid
- **WHEN** path conversion is requested for a supported solid
- **THEN** the result contains the solid face outlines instead of an empty path

#### Scenario: Payload-only solid
- **WHEN** path conversion is requested for a solid whose payload cannot be decoded
- **THEN** conversion reports an explicit unsupported-geometry error and does not emit fabricated paths
