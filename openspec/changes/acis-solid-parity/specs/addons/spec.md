# Spec Delta

## ADDED Requirements

### Requirement: ACIS browser exposes decoded geometry
The ACIS browser SHALL expose decoded body counts, bounds, face counts, and mesh data in addition to raw SAT/SAB payload and metadata.

#### Scenario: Browser entry for a supported solid
- **WHEN** the browser inspects a drawing containing a supported 3DSOLID
- **THEN** its entry reports body, face, and triangle counts plus the decoded bounding box

#### Scenario: Browser entry for unsupported payload
- **WHEN** the browser inspects an entity whose payload cannot be decoded
- **THEN** its entry retains the payload metadata and reports the decoding diagnostic

### Requirement: Mesh export includes ACIS bodies
Mesh export SHALL include meshes generated from supported ACIS bodies alongside native MESH and polyface entities.

#### Scenario: Export a drawing with a solid
- **WHEN** a drawing containing a supported 3DSOLID is exported as OBJ
- **THEN** the output contains the solid triangles and the exported vertex count matches the decoded mesh
