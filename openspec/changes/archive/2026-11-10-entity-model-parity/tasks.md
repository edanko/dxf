# Tasks

## 1. Entity foundation

- [x] 1.1 Add `EntityBase` with shared attributes, raw tags, and xdata; verify with a round-trip test on LINE
- [x] 1.2 Add the type registry and move the existing 29 loaders onto it; verify all existing entity tests still pass
- [x] 1.3 Add capability interfaces for transform, bbox, explode, and subentities; verify compile-time checks per type

## 2. Core geometry entities

- [x] 2.1 Complete load/format for POINT, LINE, CIRCLE, ARC, ELLIPSE, SPLINE, and LWPOLYLINE; verify per-type round trips
- [x] 2.2 Complete POLYLINE/VERTEX/SEQEND including subentity access and copy; verify a 3D polyline round trip
- [x] 2.3 Complete TEXT, MTEXT, ATTRIB, ATTDEF, and TOLERANCE; verify formatting and unicode content

## 3. Composite and annotation entities

- [x] 3.1 Complete INSERT with attributes, transform, bbox, and explode; verify nested block content
- [x] 3.2 Complete HATCH, MPOLYGON, and boundary paths; verify pattern and boundary round trips
- [x] 3.3 Complete DIMENSION and all variants including ARC_DIMENSION and LARGE_RADIAL_DIMENSION; verify each loads typed
- [x] 3.4 Complete LEADER, MLEADER, MULTILEADER, MLINE, and their styles; verify subentity access

## 4. Solids, surfaces, and remaining types

- [x] 4.1 Add 3DSOLID, REGION, BODY with retained ACIS payloads; verify payload round trip
- [x] 4.2 Add EXTRUDEDSURFACE, LOFTEDSURFACE, PLANESURFACE, REVOLVEDSURFACE, SWEPTSURFACE, MESH, and HELIX; verify typed fields
- [x] 4.3 Add SHAPE, XLINE, RAY, WIPEOUT, LIGHT, SUN, OLE2FRAME, ACAD_TABLE, and ACAD_PROXY_ENTITY; verify each loads and re-saves
- [x] 4.4 Add FIELD and unknown/proxy retention; verify third-party entities survive

## 5. Operations

- [x] 5.1 Implement transform for every geometry category and verify matrix composition tests
- [x] 5.2 Implement bounding boxes including inserts and verify against known fixtures
- [x] 5.3 Implement explode for INSERT, POLYLINE, HATCH, DIMENSION, and MLEADER; verify output entity types

## 6. Validation

- [x] 6.1 Add per-type load/format golden tests and verify the registry has no untyped standard types
- [x] 6.2 Update `docs/parity/README.md` entity matrix and record ACIS limitations

## Workflow follow-up

- Archive the change after review.
