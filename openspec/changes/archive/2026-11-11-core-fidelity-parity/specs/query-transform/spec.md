# Spec Delta

## ADDED Requirements

### Requirement: Query and bounds cover all typed entities
Spatial queries, selection bounds, and layout extents SHALL use the bounding box of every supported entity category and SHALL report empty bounds when no entity contributes geometry.

#### Scenario: Query finds proxy geometry
- **WHEN** a proxy entity with decodable graphics lies inside the query rectangle
- **THEN** the spatial query returns it based on its actual bounding box

#### Scenario: Empty selection reports empty bounds
- **WHEN** a selection contains only entities without drawable geometry
- **THEN** bounds are reported as empty and no zero-origin extent is fabricated
