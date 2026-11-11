# Spec Delta

## ADDED Requirements

### Requirement: ACIS entities expose decoded geometry
An entity carrying a supported SAT or SAB payload SHALL expose decoded ACIS bodies while retaining the original payload byte-for-byte.

#### Scenario: Load a supported 3DSOLID
- **WHEN** a 3DSOLID with a supported payload is loaded
- **THEN** callers can obtain its decoded bodies, bounding box, and face count, and serializing it reproduces the original payload

#### Scenario: Unsupported payload
- **WHEN** an entity carries an unsupported ACIS construct
- **THEN** decoded geometry reports the unsupported record and the entity still round-trips its payload unchanged

### Requirement: ACIS entity transforms update decoded geometry
Transforming an ACIS entity SHALL apply the same transform to its decoded bodies and keep the payload consistent with the transformed geometry.

#### Scenario: Translate a solid
- **WHEN** a 3DSOLID with decoded geometry is translated
- **THEN** the decoded body bounding box moves by the same translation and saving produces parseable ACIS data at the new location

### Requirement: ACIS payload copies are independent
Copying an ACIS entity SHALL deep-copy decoded geometry and payload buffers so mutations to the copy do not affect the original.

#### Scenario: Mutate copied solid
- **WHEN** a copied 3DSOLID has its decoded geometry or payload changed
- **THEN** the original entity retains its prior geometry and bytes
