# edge-topology Specification

## Purpose

Defines linked-edge graph construction from drawing geometry and the creation of ordered polylines and hatch boundaries from that topology.

## Requirements

### Requirement: Linked edges build a vertex graph
The system SHALL build a graph from line segments and supported curves, merging endpoints within a caller-provided tolerance and recording vertex degree, leaves, and connected components.

#### Scenario: Shared endpoints merge
- **WHEN** two segments share an endpoint within tolerance
- **THEN** the graph contains one vertex of degree two and two edges

#### Scenario: Distinct endpoints
- **WHEN** two segments do not share endpoints within tolerance
- **THEN** the graph contains two disconnected components

### Requirement: Edge analysis reports loops and open chains
The system SHALL report closed loops, open chains, leaves, and edge counts for each connected component.

#### Scenario: Closed rectangle
- **WHEN** four edges form a closed rectangle
- **THEN** analysis reports one loop with four edges and no leaves

#### Scenario: Open chain
- **WHEN** three edges form an open chain
- **THEN** analysis reports one open chain with two leaves

### Requirement: Edge networks create drawing entities
The system SHALL create ordered LWPOLYLINE or POLYLINE entities and hatch boundary paths from selected loops and chains.

#### Scenario: Rectangle to polyline
- **WHEN** a closed four-edge loop is converted to a polyline
- **THEN** the result is a closed four-vertex polyline following edge order

#### Scenario: Open chain to boundary
- **WHEN** an open chain is requested as a closed hatch boundary
- **THEN** creation returns an explicit error because the chain is not closed

### Requirement: Edge creation preserves source orientation
Created polylines and boundaries SHALL preserve the selected edge orientation and SHALL expose a documented way to reverse a loop.

#### Scenario: Reversed loop
- **WHEN** a loop is created with reverse enabled
- **THEN** the resulting vertex order is the reverse of the source loop order
