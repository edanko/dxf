package spatial

import (
	"math"
)

// SpatialIndex is a simple spatial index for fast entity queries
type SpatialIndex struct {
	entities []interface{}
	bbox     *BoundingBox
}

// NewSpatialIndex creates a new spatial index
func NewSpatialIndex() *SpatialIndex {
	return &SpatialIndex{
		entities: make([]interface{}, 0),
		bbox:     nil,
	}
}

// Insert adds an entity to the spatial index
func (si *SpatialIndex) Insert(entity interface{}, x, y float64) {
	// Create bounding box for entity (simplified as 1x1 square)
	entityBBox := NewBoundingBox(x-0.5, y-0.5, x+0.5, y+0.5)

	si.entities = append(si.entities, entity)
	si.bbox = si.bbox.Extend(entityBBox)
}

// Query finds entities within the given bounding box
func (si *SpatialIndex) Query(bbox *BoundingBox) []interface{} {
	var results []interface{}

	if si.bbox == nil || si.bbox.Intersects(bbox) {
		for _, entity := range si.entities {
			results = append(results, entity)
		}
	} else {
		// Only check entities that might intersect
		for _, entity := range si.entities {
			// Simple check: entity center in query box
			// In real implementation, this would use entity.BBox()
			entityCenter := getEntityCenter(entity)
			if bbox.Contains(entityCenter.X, entityCenter.Y) {
				results = append(results, entity)
			}
		}
	}

	return results
}

// QueryNearest finds the nearest entity to a given point
func (si *SpatialIndex) QueryNearest(x, y float64) (interface{}, float64) {
	if len(si.entities) == 0 {
		return nil, math.Inf(1)
	}

	var nearest interface{}
	minDist := math.Inf(1)

	for _, entity := range si.entities {
		entityCenter := getEntityCenter(entity)
		dist := math.Sqrt((x-entityCenter.X)*(x-entityCenter.X) + (y-entityCenter.Y)*(y-entityCenter.Y))

		if dist < minDist {
			nearest = entity
			minDist = dist
		}
	}

	return nearest, minDist
}

// QueryRange finds entities within a rectangular range
func (si *SpatialIndex) QueryRange(minX, minY, maxX, maxY float64) []interface{} {
	var results []interface{}
	queryBBox := NewBoundingBox(minX, minY, maxX, maxY)

	for _, entity := range si.entities {
		entityCenter := getEntityCenter(entity)
		if queryBBox.Contains(entityCenter.X, entityCenter.Y) {
			results = append(results, entity)
		}
	}

	return results
}

// Count returns the number of entities in the spatial index
func (si *SpatialIndex) Count() int {
	return len(si.entities)
}

// UpdateBounds recomputes the overall bounding box
func (si *SpatialIndex) UpdateBounds() {
	if len(si.entities) == 0 {
		return
	}

	// Initialize with first entity
	firstCenter := getEntityCenter(si.entities[0])
	si.bbox = NewBoundingBox(
		firstCenter.X-0.5, firstCenter.Y-0.5,
		firstCenter.X+0.5, firstCenter.Y+0.5,
	)

	// Expand to include all entities
	for _, entity := range si.entities[1:] {
		entityCenter := getEntityCenter(entity)
		entityBBox := NewBoundingBox(
			entityCenter.X-0.5, entityCenter.Y-0.5,
			entityCenter.X+0.5, entityCenter.Y+0.5,
		)
		si.bbox = si.bbox.Extend(entityBBox)
	}
}

// Point2D represents a 2D point
type Point2D struct {
	X, Y float64
}

// getEntityCenter extracts a center point from an entity
// This is a placeholder - in real implementation, entities would have BBox() methods
func getEntityCenter(entity interface{}) Point2D {
	// Default implementation - return origin
	// Real implementation would call entity.BBox().Center()
	return Point2D{X: 0.0, Y: 0.0}
}
