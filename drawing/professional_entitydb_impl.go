package drawing

import (
	"fmt"
	"time"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/spatial"
)

// addEntityInternal adds an entity without transaction support (internal method)
func (db *ProfessionalEntityDB) addEntityInternal(e entity.Entity) error {
	if e == nil {
		return fmt.Errorf("cannot add nil entity")
	}

	handle := e.Handle()
	if handle == "" {
		e.SetHandle(db.handles)
		handle = e.Handle()
	}

	entry := &EntityEntry{
		Entity:     e,
		IsAlive:    true,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Version:    0,
		Tags:       make(map[string]interface{}),
	}

	db.entities[handle] = entry
	db.liveEntities[handle] = true
	db.updateIndexes(handle, entry, true)
	db.updateStats()
	return nil
}

// updateEntityInternal updates an entity without transaction support (internal method)
func (db *ProfessionalEntityDB) updateEntityInternal(e entity.Entity) error {
	if e == nil {
		return fmt.Errorf("cannot update with nil entity")
	}

	handle := e.Handle()
	if handle == "" {
		return fmt.Errorf("entity must have a handle for update")
	}

	entry, exists := db.entities[handle]
	if !exists {
		return fmt.Errorf("entity with handle %s not found", handle)
	}

	// Update entry
	entry.Entity = e
	entry.ModifiedAt = time.Now()
	entry.Version++

	// Update indexes
	db.updateIndexes(handle, entry, false)
	db.updateStats()
	return nil
}

// deleteEntityInternal deletes an entity without transaction support (internal method)
func (db *ProfessionalEntityDB) deleteEntityInternal(handle string) error {
	if handle == "" {
		return fmt.Errorf("handle cannot be empty")
	}

	entry, exists := db.entities[handle]
	if !exists {
		return fmt.Errorf("entity with handle %s not found", handle)
	}

	// Mark as dead but keep in trash can
	entry.IsAlive = false
	delete(db.liveEntities, handle)
	db.trashCan = append(db.trashCan, handle)

	// Update indexes
	db.removeFromIndexes(handle, entry)
	db.updateStats()
	return nil
}

// restoreEntityEntry restores an entity entry (internal method for rollback)
func (db *ProfessionalEntityDB) restoreEntityEntry(handle string, entry *EntityEntry) error {
	if handle == "" || entry == nil {
		return fmt.Errorf("invalid restore parameters")
	}

	db.entities[handle] = entry
	if entry.IsAlive {
		db.liveEntities[handle] = true
	} else {
		delete(db.liveEntities, handle)
	}

	db.updateIndexes(handle, entry, entry.IsAlive)
	db.updateStats()
	return nil
}

// AddEntity adds an entity to the database with transaction support
func (db *ProfessionalEntityDB) AddEntity(e entity.Entity) error {
	// Assign handle if not already assigned
	if e.Handle() == "" {
		e.SetHandle(db.handles)
	}

	// Check if we're already in a transaction (mutex already held)
	if db.currentTx != nil {
		op := &AddOperation{
			Entity: e,
			Desc:   fmt.Sprintf("Add entity %s", e.DXFType()),
		}
		db.currentTx.Operations = append(db.currentTx.Operations, op)
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.addEntityInternal(e)
}

// GetEntity retrieves an entity by handle
func (db *ProfessionalEntityDB) GetEntity(handle string) (entity.Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if entry, exists := db.cache[handle]; exists {
		return entry.Entity, nil
	}

	entry, exists := db.entities[handle]
	if !exists || !entry.IsAlive {
		return nil, fmt.Errorf("entity with handle %s not found", handle)
	}

	// Update cache
	if len(db.cache) < db.cacheSize {
		db.cache[handle] = entry
	}

	return entry.Entity, nil
}

// UpdateEntity updates an existing entity with transaction support
func (db *ProfessionalEntityDB) UpdateEntity(e entity.Entity) error {
	// Check if we're already in a transaction (mutex already held)
	if db.currentTx != nil {
		oldEntry, _ := db.entities[e.Handle()]
		op := &UpdateOperation{
			Entity:   e,
			OldEntry: oldEntry,
			Desc:     fmt.Sprintf("Update entity %s", e.DXFType()),
		}
		db.currentTx.Operations = append(db.currentTx.Operations, op)
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.updateEntityInternal(e)
}

// DeleteEntity removes an entity from the database with transaction support
func (db *ProfessionalEntityDB) DeleteEntity(handle string) error {
	// Check if we're already in a transaction (mutex already held)
	if db.currentTx != nil {
		oldEntry, _ := db.entities[handle]
		op := &DeleteOperation{
			HandleStr:     handle,
			OldEntry:      oldEntry,
			OperationDesc: fmt.Sprintf("Delete entity %s", handle),
		}
		db.currentTx.Operations = append(db.currentTx.Operations, op)
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.deleteEntityInternal(handle)
}

// GetEntitiesByType returns all entities of a specific type
func (db *ProfessionalEntityDB) GetEntitiesByType(entityType string) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, entry := range db.entities {
		if entry.IsAlive && entry.Entity.DXFType() == entityType {
			entities = append(entities, entry.Entity)
		}
	}
	return entities
}

// GetEntitiesByLayer returns all entities on a specific layer
func (db *ProfessionalEntityDB) GetEntitiesByLayer(layerName string) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, entry := range db.entities {
		if entry.IsAlive {
			if layer := entry.Entity.Layer(); layer != nil && layer.Name() == layerName {
				entities = append(entities, entry.Entity)
			}
		}
	}
	return entities
}

// GetEntitiesInBounds returns all entities within specified bounds
func (db *ProfessionalEntityDB) GetEntitiesInBounds(min, max math.Vec2) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if !db.enableSpatialIndex {
		// Fallback to linear search
		return db.getEntitiesInBoundsLinear(min, max)
	}

	// Use spatial index
	bounds := spatial.NewBoundingBox(min.X(), min.Y(), max.X(), max.Y())
	results := db.spatialIndex.Query(bounds)
	var entities []entity.Entity
	for _, result := range results {
		if handle, ok := result.(string); ok {
			if entry, exists := db.entities[handle]; exists && entry.IsAlive {
				entities = append(entities, entry.Entity)
			}
		}
	}
	return entities
}

// getEntitiesInBoundsLinear performs linear search for entities in bounds
func (db *ProfessionalEntityDB) getEntitiesInBoundsLinear(min, max math.Vec2) []entity.Entity {
	var entities []entity.Entity
	for _, entry := range db.entities {
		if !entry.IsAlive {
			continue
		}

		// Get entity bounds
		minBbox, maxBbox := entry.Entity.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			entityMin := math.NewVec2(minBbox[0], minBbox[1])
			entityMax := math.NewVec2(maxBbox[0], maxBbox[1])

			// Check intersection
			if entityMin.X() <= max.X() && entityMax.X() >= min.X() &&
				entityMin.Y() <= max.Y() && entityMax.Y() >= min.Y() {
				entities = append(entities, entry.Entity)
			}
		}
	}
	return entities
}

// FilterEntities returns entities that match a predicate function
func (db *ProfessionalEntityDB) FilterEntities(predicate func(entity.Entity) bool) []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for _, entry := range db.entities {
		if entry.IsAlive && predicate(entry.Entity) {
			entities = append(entities, entry.Entity)
		}
	}
	return entities
}

// IsEntityAlive checks if an entity is alive
func (db *ProfessionalEntityDB) IsEntityAlive(handle string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	alive, exists := db.liveEntities[handle]
	return exists && alive
}

// GetLiveEntities returns all live entities
func (db *ProfessionalEntityDB) GetLiveEntities() []entity.Entity {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var entities []entity.Entity
	for handle, alive := range db.liveEntities {
		if alive {
			if entry, exists := db.entities[handle]; exists {
				entities = append(entities, entry.Entity)
			}
		}
	}
	return entities
}

// GetTrashCan returns handles of deleted entities
func (db *ProfessionalEntityDB) GetTrashCan() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return append([]string{}, db.trashCan...)
}

// EmptyTrashCan permanently removes all deleted entities
func (db *ProfessionalEntityDB) EmptyTrashCan() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, handle := range db.trashCan {
		delete(db.entities, handle)
		db.removeFromIndexesByHandle(handle)
	}

	db.trashCan = db.trashCan[:0]
	db.updateStats()
	return nil
}

// CloneEntity creates a copy of an entity with new handle
func (db *ProfessionalEntityDB) CloneEntity(handle string, newHandle string) (entity.Entity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	entry, exists := db.entities[handle]
	if !exists || !entry.IsAlive {
		return nil, fmt.Errorf("entity with handle %s not found", handle)
	}

	// Create a deep copy (this would need to be implemented per entity type)
	// For now, create a shallow copy with new handle
	cloned := entry.Entity

	// Assign new handle
	if newHandle != "" {
		// This would need to be implemented per entity type
		// For now, just return the original entity
		return cloned, fmt.Errorf("entity cloning not fully implemented")
	}

	return cloned, nil
}

// BulkAddEntities adds multiple entities efficiently
func (db *ProfessionalEntityDB) BulkAddEntities(entities []entity.Entity) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.currentTx != nil {
		// Add each as individual operation for transaction support
		for _, e := range entities {
			op := &AddOperation{
				Entity: e,
				Desc:   fmt.Sprintf("Bulk add entity %s", e.DXFType()),
			}
			db.currentTx.Operations = append(db.currentTx.Operations, op)
		}
		return nil
	}

	// Bulk add without transaction
	for _, e := range entities {
		if err := db.addEntityInternal(e); err != nil {
			return err
		}
	}
	return nil
}

// BulkDeleteEntities deletes multiple entities efficiently
func (db *ProfessionalEntityDB) BulkDeleteEntities(handles []string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.currentTx != nil {
		// Add each as individual operation for transaction support
		for _, handle := range handles {
			oldEntry, _ := db.entities[handle]
			op := &DeleteOperation{
				HandleStr:     handle,
				OldEntry:      oldEntry,
				OperationDesc: fmt.Sprintf("Bulk delete entity %s", handle),
			}
			db.currentTx.Operations = append(db.currentTx.Operations, op)
		}
		return nil
	}

	// Bulk delete without transaction
	for _, handle := range handles {
		if err := db.deleteEntityInternal(handle); err != nil {
			return err
		}
	}
	return nil
}
