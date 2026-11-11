# Design

## Context

`entity/` has ~75 files and many `Format` methods, but loading is centralized in
a 29-case switch in the root package, and unsupported types become
`GenericEntity`. There is no shared base carrying common DXF attributes and no
uniform transform/bbox/explode contract. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**

- A uniform, testable entity contract.
- Typed coverage of all standard entities.
- Operations implemented per type with explicit unsupported errors.

**Non-Goals:**

- Full geometric fidelity for ACIS solids (3DSOLID, REGION, BODY); these are
  modelled structurally and retained, not evaluated.
- Rendering-quality tessellation of every surface.
- Changing existing public type names where avoidable.

## Decisions

### Common embedded base

Add an `EntityBase` embedded by every entity, holding shared attributes, the
handle, owner, raw tags, and xdata. Provide default implementations of the
common interface. Alternative: an interface with duplicated fields per type.
Rejected because it multiplies boilerplate and invites inconsistency.

### Registry-based load dispatch

Replace the root switch with a registry mapping DXF type name to a constructor
and loader. Each entity package file registers itself. This removes the central
switch and lets new types be added without editing the root. Alternative:
generated switch. Rejected for build complexity.

### Capability interfaces for optional operations

Define small interfaces (`Transformer`, `Bounder`, `Exploder`, `SubentityOwner`)
and implement them per type. Callers type-assert or use helpers. Alternative:
one large interface with no-op methods. Rejected because no-ops hide unsupported
behavior.

### Transform policy

Apply affine transforms directly where exact (points, lines, polylines, text
insertion). For circles and arcs under non-uniform scale, convert to ellipse or
return a documented error according to the configured policy. This mirrors
ezdxf's explicit behavior.

### Structural ACIS representation

Store ACIS SAT/SAB payloads and surface control data without evaluating booleans
or tessellating. Geometry evaluation belongs to the CSG add-on change.

## Risks / Trade-offs

- **Large surface area** → Land types in batches by category with per-batch
  tests; keep the registry so partial coverage is safe.
- **Transform edge cases** → Centralize in `math` and test each entity category.
- **API churn** → Keep existing constructors; add the base and interfaces
  additively.
- **Raw-tag retention memory** → Store only unhandled tags, not the full stream.
