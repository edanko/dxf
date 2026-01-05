package drawing

import (
	"fmt"
	"time"

	"github.com/edanko/dxf/spatial"
)

// updateIndexes updates all indexes when entity is added or modified
func (db *ProfessionalEntityDB) updateIndexes(handle string, entry *EntityEntry, isAlive bool) {
	if !isAlive {
		db.removeFromIndexesByHandle(handle)
		return
	}

	entity := entry.Entity

	// Update type index
	if db.enableTypeIndex {
		entityType := entity.DXFType()
		if db.typeIndex[entityType] == nil {
			db.typeIndex[entityType] = make(map[string]bool)
		}
		db.typeIndex[entityType][handle] = true
	}

	// Update layer index
	if db.enableLayerIndex {
		if layer := entity.Layer(); layer != nil {
			layerName := layer.Name()
			if db.layerIndex[layerName] == nil {
				db.layerIndex[layerName] = make(map[string]bool)
			}
			db.layerIndex[layerName][handle] = true
		}
	}

	// Update spatial index
	if db.enableSpatialIndex {
		minBbox, maxBbox := entity.BBox()
		if len(minBbox) >= 2 && len(maxBbox) >= 2 {
			bounds := spatial.NewBoundingBox(minBbox[0], minBbox[1], maxBbox[0], maxBbox[1])
			db.spatialIndex.Insert(bounds, handle)
		}
	}
}

// removeFromIndexes removes an entity from all indexes
func (db *ProfessionalEntityDB) removeFromIndexes(handle string, entry *EntityEntry) {
	entity := entry.Entity

	// Remove from type index
	if db.enableTypeIndex {
		entityType := entity.DXFType()
		if typeMap := db.typeIndex[entityType]; typeMap != nil {
			delete(typeMap, handle)
		}
	}

	// Remove from layer index
	if db.enableLayerIndex {
		if layer := entity.Layer(); layer != nil {
			layerName := layer.Name()
			if layerMap := db.layerIndex[layerName]; layerMap != nil {
				delete(layerMap, handle)
			}
		}
	}

	// Remove from spatial index (note: RTree doesn't have Remove method yet)
	// For now, we'll rebuild the spatial index when needed
	if db.enableSpatialIndex {
		// TODO: Implement Remove in RTree or rebuild index
	}
}

// removeFromIndexesByHandle removes an entity from all indexes by handle
func (db *ProfessionalEntityDB) removeFromIndexesByHandle(handle string) {
	entry, exists := db.entities[handle]
	if !exists {
		return
	}

	db.removeFromIndexes(handle, entry)
}

// updateStats updates database statistics
func (db *ProfessionalEntityDB) updateStats() {
	db.stats.TotalEntities = len(db.entities)
	db.stats.LiveEntities = len(db.liveEntities)
	db.stats.DeadEntities = len(db.trashCan)
	db.stats.Transactions = len(db.transactions)
	db.stats.LastModified = time.Now()

	// Update type and layer counts
	db.stats.TypeCounts = make(map[string]int)
	db.stats.LayerCounts = make(map[string]int)

	for _, entry := range db.entities {
		if !entry.IsAlive {
			continue
		}

		// Count by type
		entityType := entry.Entity.DXFType()
		db.stats.TypeCounts[entityType]++

		// Count by layer
		if layer := entry.Entity.Layer(); layer != nil {
			layerName := layer.Name()
			db.stats.LayerCounts[layerName]++
		}
	}
}

// RebuildIndexes rebuilds all indexes from scratch
func (db *ProfessionalEntityDB) RebuildIndexes() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.rebuildIndexesInternal()
}

// rebuildIndexesInternal rebuilds indexes without acquiring lock (internal method)
func (db *ProfessionalEntityDB) rebuildIndexesInternal() error {
	// Clear existing indexes
	db.typeIndex = make(map[string]map[string]bool)
	db.layerIndex = make(map[string]map[string]bool)
	db.spatialIndex = spatial.NewRTree(spatial.DefaultRTreeOptions())

	// Rebuild indexes
	for handle, entry := range db.entities {
		if entry.IsAlive {
			db.updateIndexes(handle, entry, true)
		}
	}

	return nil
}

// GetStatistics returns current database statistics
func (db *ProfessionalEntityDB) GetStatistics() *DatabaseStats {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Return a copy to prevent external modification
	stats := *db.stats
	return &stats
}

// Optimize optimizes the database for better performance
func (db *ProfessionalEntityDB) Optimize() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Rebuild indexes (lock already held)
	if err := db.rebuildIndexesInternal(); err != nil {
		return err
	}

	// Clear cache
	db.cache = make(map[string]*EntityEntry)

	// Trigger garbage collection
	// Note: In Go, we can suggest GC but not force it
	db.lastCleanup = time.Now()

	return nil
}

// Validate checks the integrity of the entity database
func (db *ProfessionalEntityDB) Validate() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var errors []string

	// Check for invalid handles
	for handle, entry := range db.entities {
		if handle == "" {
			errors = append(errors, "found entity with empty handle")
		}
		if entry == nil {
			errors = append(errors, "found nil entity entry")
		}
		if entry.Entity == nil {
			errors = append(errors, "found nil entity in entry")
		}
	}

	// Check for handle consistency
	liveCount := 0
	for handle, entry := range db.entities {
		if entry.IsAlive {
			liveCount++
			if _, isLive := db.liveEntities[handle]; !isLive {
				errors = append(errors, "live entity not in liveEntities map")
			}
		}
	}

	if liveCount != len(db.liveEntities) {
		errors = append(errors, "live entity count mismatch")
	}

	// Check trash can consistency
	for _, handle := range db.trashCan {
		if entry, exists := db.entities[handle]; !exists {
			errors = append(errors, "trash can contains handle not in entities")
		} else if entry.IsAlive {
			errors = append(errors, "trash can contains live entity")
		}
	}

	// Check index consistency
	for entityType, handles := range db.typeIndex {
		for handle := range handles {
			if entry, exists := db.entities[handle]; !exists {
				errors = append(errors, "type index contains handle not in entities")
			} else if !entry.IsAlive {
				errors = append(errors, "type index contains dead entity")
			} else if entry.Entity.DXFType() != entityType {
				errors = append(errors, "type index contains wrong entity type")
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("database validation failed: %v", errors)
	}

	return nil
}

// CreateSnapshot creates a snapshot of current database state
func (db *ProfessionalEntityDB) CreateSnapshot() *DatabaseSnapshot {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Deep copy entities
	entities := make(map[string]*EntityEntry)
	for handle, entry := range db.entities {
		entityCopy := *entry
		entities[handle] = &entityCopy
	}

	return &DatabaseSnapshot{
		Timestamp:     time.Now(),
		Entities:      entities,
		HandleCounter: 0, // TODO: Implement proper handle counter access
		Version:       db.version,
	}
}

// RestoreSnapshot restores database to a previous state
func (db *ProfessionalEntityDB) RestoreSnapshot(snapshot *DatabaseSnapshot) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.restoreSnapshotInternal(snapshot)
}

// restoreSnapshotInternal restores database state without acquiring lock (internal method)
func (db *ProfessionalEntityDB) restoreSnapshotInternal(snapshot *DatabaseSnapshot) error {
	// Clear current state
	db.entities = make(map[string]*EntityEntry)
	db.liveEntities = make(map[string]bool)
	db.trashCan = make([]string, 0)
	db.typeIndex = make(map[string]map[string]bool)
	db.layerIndex = make(map[string]map[string]bool)
	db.spatialIndex = spatial.NewRTree(spatial.DefaultRTreeOptions())

	// Restore entities
	for handle, entry := range snapshot.Entities {
		db.entities[handle] = entry
		if entry.IsAlive {
			db.liveEntities[handle] = true
		}
	}

	// Restore handle counter (this needs proper implementation)
	db.version = snapshot.Version

	// Rebuild indexes
	for handle, entry := range db.entities {
		if entry.IsAlive {
			db.updateIndexes(handle, entry, true)
		}
	}

	db.updateStats()
	return nil
}

// ClearCache clears the entity cache
func (db *ProfessionalEntityDB) ClearCache() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.cache = make(map[string]*EntityEntry)
	return nil
}

// GetCacheStats returns cache statistics
func (db *ProfessionalEntityDB) GetCacheStats() map[string]interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return map[string]interface{}{
		"size":         len(db.cache),
		"max_size":     db.cacheSize,
		"usage":        float64(len(db.cache)) / float64(db.cacheSize),
		"last_cleanup": db.lastCleanup,
	}
}
