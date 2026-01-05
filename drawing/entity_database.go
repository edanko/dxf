// Package drawing provides advanced document management for DXF drawings
package drawing

import (
	"fmt"
	"math"
	"sync"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/spatial"
)

// EntityDatabase provides robust entity management with handle generation
type EntityDatabase struct {
	mu       sync.RWMutex
	entities map[string]entity.Entity
	handles  *handle.HandleGenerator
	modified bool
}

// NewEntityDatabase creates a new entity database
func NewEntityDatabase() *EntityDatabase {
	return &EntityDatabase{
		entities: make(map[string]entity.Entity),
		handles:  handle.NewHandleGenerator(),
		modified: false,
	}
}

// AddEntity adds an entity to the database with automatic handle assignment
func (db *EntityDatabase) AddEntity(e entity.Entity) error {
	if e == nil {
		return fmt.Errorf("cannot add nil entity")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// Assign handle if not already set
	if e.Handle() == "" {
		e.SetHandle(db.handles)
	}

	// Validate unique handle
	if _, exists := db.entities[e.Handle()]; exists {
		return fmt.Errorf("entity with handle %s already exists", e.Handle())
	}

	db.entities[e.Handle()] = e
	db.modified = true
	return nil
}

// GetEntity retrieves an entity by handle
func (db *EntityDatabase) GetEntity(handle string) (entity.Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ent, exists := db.entities[handle]
	if !exists {
		return nil, fmt.Errorf("entity with handle %s not found", handle)
	}
	return ent, nil
}

// GetEntityByHandle returns entity and existence flag (compatible with existing code)
func (db *EntityDatabase) GetEntityByHandle(handle string) (entity.Entity, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ent, exists := db.entities[handle]
	return ent, exists
}

// DeleteEntity removes an entity from the database
func (db *EntityDatabase) DeleteEntity(handle string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.entities[handle]; !exists {
		return fmt.Errorf("entity with handle %s not found", handle)
	}

	delete(db.entities, handle)
	db.modified = true
	return nil
}

// UpdateEntity replaces an existing entity with a new version
func (db *EntityDatabase) UpdateEntity(e entity.Entity) error {
	if e == nil {
		return fmt.Errorf("cannot update with nil entity")
	}

	if e.Handle() == "" {
		return fmt.Errorf("entity must have a handle for update")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.entities[e.Handle()]; !exists {
		return fmt.Errorf("entity with handle %s not found", e.Handle())
	}

	db.entities[e.Handle()] = e
	db.modified = true
	return nil
}

// GetAllEntities returns all entities in the database
func (db *EntityDatabase) GetAllEntities() []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entities := make([]entity.Entity, 0, len(db.entities))
	for _, ent := range db.entities {
		entities = append(entities, ent)
	}
	return entities
}

// GetEntitiesByType returns all entities of a specific type
func (db *EntityDatabase) GetEntitiesByType(entityType string) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, ent := range db.entities {
		if ent.DXFType() == entityType {
			entities = append(entities, ent)
		}
	}
	return entities
}

// GetEntitiesByLayer returns all entities on a specific layer
func (db *EntityDatabase) GetEntitiesByLayer(layerName string) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, ent := range db.entities {
		if layer := ent.Layer(); layer != nil && layer.Name() == layerName {
			entities = append(entities, ent)
		}
	}
	return entities
}

// GetEntityCount returns the total number of entities
func (db *EntityDatabase) GetEntityCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.entities)
}

// GetEntityCountByType returns the count of entities of a specific type
func (db *EntityDatabase) GetEntityCountByType(entityType string) int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	count := 0
	for _, ent := range db.entities {
		if ent.DXFType() == entityType {
			count++
		}
	}
	return count
}

// FindEntitiesByAttribute finds entities with specific attribute values
func (db *EntityDatabase) FindEntitiesByAttribute(attrName, attrValue string) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, ent := range db.entities {
		// Use GetAttributes method from entity interface
		attrs := ent.GetAttributes()
		if val, exists := attrs[attrName]; exists && fmt.Sprintf("%v", val) == attrValue {
			entities = append(entities, ent)
		}
	}
	return entities
}

// FilterEntities returns entities that match a predicate function
func (db *EntityDatabase) FilterEntities(predicate func(entity.Entity) bool) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, ent := range db.entities {
		if predicate(ent) {
			entities = append(entities, ent)
		}
	}
	return entities
}

// GroupEntitiesBy groups entities by a key function
func (db *EntityDatabase) GroupEntitiesBy(keyFunc func(entity.Entity) string) map[string][]entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	groups := make(map[string][]entity.Entity)
	for _, ent := range db.entities {
		key := keyFunc(ent)
		groups[key] = append(groups[key], ent)
	}
	return groups
}

// Validate checks the integrity of the entity database
func (db *EntityDatabase) Validate() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var errors []string

	// Check for invalid handles
	for handle, ent := range db.entities {
		if handle == "" {
			errors = append(errors, "found entity with empty handle")
		}
		if ent == nil {
			errors = append(errors, fmt.Sprintf("found nil entity at handle %s", handle))
		}
	}

	// Check for handle duplicates (shouldn't happen with current implementation)
	handleCount := make(map[string]int)
	for handle := range db.entities {
		handleCount[handle]++
		if handleCount[handle] > 1 {
			errors = append(errors, fmt.Sprintf("duplicate handle found: %s", handle))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("entity database validation failed: %v", errors)
	}

	return nil
}

// IsModified returns whether the database has been modified since last save
func (db *EntityDatabase) IsModified() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.modified
}

// SetModified sets the modified flag
func (db *EntityDatabase) SetModified(modified bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.modified = modified
}

// Clear removes all entities from the database
func (db *EntityDatabase) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.entities = make(map[string]entity.Entity)
	db.modified = true
}

// GetHandles returns all entity handles in the database
func (db *EntityDatabase) GetHandles() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	handles := make([]string, 0, len(db.entities))
	for handle := range db.entities {
		handles = append(handles, handle)
	}
	return handles
}

// Optimize reorganizes the database for better performance
func (db *EntityDatabase) Optimize() {
	db.mu.Lock()
	defer db.mu.Unlock()

	// For now, just trigger garbage collection
	// Could implement more sophisticated optimizations later
	db.modified = true
}

// Clone creates a copy of the entity database
func (db *EntityDatabase) Clone() *EntityDatabase {
	db.mu.RLock()
	defer db.mu.RUnlock()

	clone := &EntityDatabase{
		entities: make(map[string]entity.Entity),
		handles:  handle.NewHandleGenerator(),
		modified: false,
	}

	// Copy entities (shallow copy - entities themselves are not cloned)
	for handle, ent := range db.entities {
		clone.entities[handle] = ent
	}

	return clone
}

// Statistics provides information about the database
func (db *EntityDatabase) Statistics() map[string]int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := make(map[string]int)
	stats["total"] = len(db.entities)

	typeCounts := make(map[string]int)
	for _, ent := range db.entities {
		entityType := ent.DXFType()
		typeCounts[entityType]++
	}

	for entityType, count := range typeCounts {
		stats[entityType] = count
	}

	return stats
}

// Spatial Query Methods

// BoundingBox represents a 2D bounding box
type BoundingBox struct {
	MinX, MinY, MaxX, MaxY float64
}

// NewBoundingBox creates a new bounding box
func NewBoundingBox(minX, minY, maxX, maxY float64) *BoundingBox {
	return &BoundingBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

// ContainsPoint checks if a point is inside the bounding box
func (b *BoundingBox) ContainsPoint(x, y float64) bool {
	return x >= b.MinX && x <= b.MaxX && y >= b.MinY && y <= b.MaxY
}

// Intersects checks if two bounding boxes intersect
func (b *BoundingBox) Intersects(other *BoundingBox) bool {
	return !(b.MaxX < other.MinX || b.MinX > other.MaxX ||
		b.MaxY < other.MinY || b.MinY > other.MaxY)
}

// GetEntitiesInBox returns all entities within the specified 2D bounding box
func (db *EntityDatabase) GetEntitiesInBox(minX, minY, maxX, maxY float64) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	box := NewBoundingBox(minX, minY, maxX, maxY)
	var entities []entity.Entity
	for _, ent := range db.entities {
		minBbox, maxBbox := ent.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			entBox := NewBoundingBox(minBbox[0], minBbox[1], maxBbox[0], maxBbox[1])
			if box.Intersects(entBox) {
				entities = append(entities, ent)
			}
		}
	}
	return entities
}

// GetEntitiesInRadius returns all entities within the specified radius from a center point
func (db *EntityDatabase) GetEntitiesInRadius(cx, cy, radius float64) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	radiusSq := radius * radius
	var entities []entity.Entity
	for _, ent := range db.entities {
		minBbox, maxBbox := ent.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			centerX := (minBbox[0] + maxBbox[0]) / 2
			centerY := (minBbox[1] + maxBbox[1]) / 2
			dx := centerX - cx
			dy := centerY - cy
			if dx*dx+dy*dy <= radiusSq {
				entities = append(entities, ent)
			}
		}
	}
	return entities
}

// PointInPolygon checks if a point is inside a polygon
func PointInPolygon(px, py float64, polygon []struct{ X, Y float64 }) bool {
	inside := false
	n := len(polygon)
	for i, j := 0, n-1; i < n; j = i {
		pi := polygon[i]
		pj := polygon[j]
		if ((pi.Y > py) != (pj.Y > py)) &&
			(px < (pj.X-pi.X)*(py-pi.Y)/(pj.Y-pi.Y)+pi.X) {
			inside = !inside
		}
	}
	return inside
}

// GetEntitiesInPolygon returns all entities within the specified polygon
func (db *EntityDatabase) GetEntitiesInPolygon(polygon []struct{ X, Y float64 }) []entity.Entity {
	if len(polygon) < 3 {
		return nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var minX, minY, maxX, maxY float64 = math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, p := range polygon {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	var entities []entity.Entity
	for _, ent := range db.entities {
		minBbox, maxBbox := ent.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			cx := (minBbox[0] + maxBbox[0]) / 2
			cy := (minBbox[1] + maxBbox[1]) / 2
			if PointInPolygon(cx, cy, polygon) {
				entities = append(entities, ent)
			}
		}
	}
	return entities
}

// GetSpatialIndex returns the spatial index for external use
func (db *EntityDatabase) GetSpatialIndex() *spatial.RTree {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Build R-tree from current entities
	rt := spatial.NewRTree(spatial.DefaultRTreeOptions())
	for _, ent := range db.entities {
		minBbox, maxBbox := ent.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			bbox := spatial.NewBoundingBox(minBbox[0], minBbox[1], maxBbox[0], maxBbox[1])
			rt.Insert(bbox, ent.Handle())
		}
	}
	return rt
}

// RebuildSpatialIndex rebuilds the spatial index from all entities
func (db *EntityDatabase) RebuildSpatialIndex() *spatial.RTree {
	return db.GetSpatialIndex()
}
