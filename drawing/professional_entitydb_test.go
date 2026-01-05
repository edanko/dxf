package drawing

import (
	"testing"
	"time"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

// MockEntity implements entity.Entity interface for testing
type MockEntity struct {
	handle     string
	entityType string
	layer      *table.Layer
	attrs      map[string]interface{}
}

func (m *MockEntity) IsEntity() bool                  { return true }
func (m *MockEntity) Format(f format.Formatter)       { /* mock implementation */ }
func (m *MockEntity) SetBlockRecord(h handle.Handler) { /* mock implementation */ }
func (m *MockEntity) Layer() *table.Layer             { return m.layer }
func (m *MockEntity) SetLayer(l *table.Layer)         { m.layer = l }
func (m *MockEntity) SetLtscale(float64)              { /* mock implementation */ }
func (m *MockEntity) BBox() ([]float64, []float64) {
	return []float64{0, 0, 10, 10}, []float64{5, 5, 15, 15}
}
func (m *MockEntity) SetColor(c color.ColorNumber) { /* mock implementation */ }
func (m *MockEntity) Handle() string               { return m.handle }
func (m *MockEntity) SetHandle(h *handle.HandleGenerator) {
	m.handle = h.Next()
}
func (m *MockEntity) DXFType() string                       { return m.entityType }
func (m *MockEntity) GetAttributes() map[string]interface{} { return m.attrs }
func (m *MockEntity) LoadAttributes(attribs map[string]interface{}) error {
	m.attrs = attribs
	return nil
}

func NewMockEntity(entityType string) *MockEntity {
	return &MockEntity{
		entityType: entityType,
		attrs:      make(map[string]interface{}),
	}
}

func TestProfessionalEntityDB_BasicCRUD(t *testing.T) {
	db := NewProfessionalEntityDB()

	// Test AddEntity
	entity1 := NewMockEntity("TEST")
	err := db.AddEntity(entity1)
	if err != nil {
		t.Fatalf("Failed to add entity: %v", err)
	}

	// Test GetEntity
	retrieved, err := db.GetEntity(entity1.Handle())
	if err != nil {
		t.Fatalf("Failed to get entity: %v", err)
	}
	if retrieved.Handle() != entity1.Handle() {
		t.Fatal("Retrieved entity is not the same as added entity")
	}

	// Test UpdateEntity
	entity1.entityType = "UPDATED"
	err = db.UpdateEntity(entity1)
	if err != nil {
		t.Fatalf("Failed to update entity: %v", err)
	}

	// Test DeleteEntity
	err = db.DeleteEntity(entity1.Handle())
	if err != nil {
		t.Fatalf("Failed to delete entity: %v", err)
	}

	// Entity should no longer be alive
	if db.IsEntityAlive(entity1.Handle()) {
		t.Fatal("Entity should be marked as dead")
	}
}

func TestProfessionalEntityDB_QueryOperations(t *testing.T) {
	db := NewProfessionalEntityDB()

	// Add test entities
	entity1 := NewMockEntity("LINE")
	entity2 := NewMockEntity("CIRCLE")
	entity3 := NewMockEntity("LINE")

	db.AddEntity(entity1)
	db.AddEntity(entity2)
	db.AddEntity(entity3)

	// Test GetEntitiesByType
	entities := db.GetEntitiesByType("LINE")
	if len(entities) != 2 {
		t.Fatalf("Expected 2 LINE entities, got %d", len(entities))
	}

	// Test GetEntitiesByType for non-existent type
	entities = db.GetEntitiesByType("NONEXISTENT")
	if len(entities) != 0 {
		t.Fatalf("Expected 0 NONEXISTENT entities, got %d", len(entities))
	}

	// Test FilterEntities
	entities = db.FilterEntities(func(e entity.Entity) bool {
		return e.DXFType() == "LINE"
	})
	if len(entities) != 2 {
		t.Fatalf("Expected 2 filtered entities, got %d", len(entities))
	}
}

func TestProfessionalEntityDB_LifecycleManagement(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("TEST")
	db.AddEntity(entity1)
	db.DeleteEntity(entity1.Handle())

	// Test GetLiveEntities
	liveEntities := db.GetLiveEntities()
	if len(liveEntities) != 0 {
		t.Fatalf("Expected 0 live entities, got %d", len(liveEntities))
	}

	// Test GetTrashCan
	trash := db.GetTrashCan()
	if len(trash) != 1 {
		t.Fatalf("Expected 1 item in trash can, got %d", len(trash))
	}
	if trash[0] != entity1.Handle() {
		t.Fatal("Trash can contains wrong handle")
	}

	// Test EmptyTrashCan
	err := db.EmptyTrashCan()
	if err != nil {
		t.Fatalf("Failed to empty trash can: %v", err)
	}

	trash = db.GetTrashCan()
	if len(trash) != 0 {
		t.Fatal("Trash can should be empty after EmptyTrashCan")
	}

	// Entity should be completely gone
	_, err = db.GetEntity(entity1.Handle())
	if err == nil {
		t.Fatal("Should not be able to retrieve entity after emptying trash")
	}
}

func TestProfessionalEntityDB_TransactionSupport(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("TEST")

	// Begin transaction
	tx, err := db.BeginTransaction("test-tx")
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Add entity within transaction
	err = db.AddEntity(entity1)
	if err != nil {
		t.Fatalf("Failed to add entity in transaction: %v", err)
	}

	// Entity should not be visible outside transaction
	_, err = db.GetEntity(entity1.Handle())
	if err == nil {
		t.Fatal("Entity should not be visible before transaction commit")
	}

	// Commit transaction
	err = db.CommitTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Entity should now be visible
	_, err = db.GetEntity(entity1.Handle())
	if err != nil {
		t.Fatalf("Failed to get entity after commit: %v", err)
	}

	// Test rollback
	entity2 := NewMockEntity("TEST2")
	tx2, err := db.BeginTransaction("test-tx-rollback")
	if err != nil {
		t.Fatalf("Failed to begin second transaction: %v", err)
	}

	err = db.AddEntity(entity2)
	if err != nil {
		t.Fatalf("Failed to add entity in second transaction: %v", err)
	}

	// Rollback transaction
	err = db.RollbackTransaction(tx2)
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Entity2 should not exist
	_, err = db.GetEntity(entity2.Handle())
	if err == nil {
		t.Fatal("Entity should not exist after rollback")
	}
}

func TestProfessionalEntityDB_BulkOperations(t *testing.T) {
	db := NewProfessionalEntityDB()

	// Create multiple entities
	entities := make([]entity.Entity, 3)
	for i := 0; i < 3; i++ {
		entities[i] = NewMockEntity("BULK_TEST")
	}

	// Test BulkAddEntities
	err := db.BulkAddEntities(entities)
	if err != nil {
		t.Fatalf("Failed to bulk add entities: %v", err)
	}

	// Verify all entities were added
	stats := db.GetStatistics()
	if stats.TotalEntities != 3 {
		t.Fatalf("Expected 3 entities after bulk add, got %d", stats.TotalEntities)
	}

	// Test BulkDeleteEntities
	handles := make([]string, 3)
	for i, e := range entities {
		handles[i] = e.Handle()
	}

	err = db.BulkDeleteEntities(handles)
	if err != nil {
		t.Fatalf("Failed to bulk delete entities: %v", err)
	}

	// Verify all entities are in trash
	trash := db.GetTrashCan()
	if len(trash) != 3 {
		t.Fatalf("Expected 3 items in trash after bulk delete, got %d", len(trash))
	}
}

func TestProfessionalEntityDB_Indexing(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("LINE")
	entity2 := NewMockEntity("CIRCLE")
	entity3 := NewMockEntity("ARC")

	// Add entities
	db.AddEntity(entity1)
	db.AddEntity(entity2)
	db.AddEntity(entity3)

	// Test GetStatistics
	stats := db.GetStatistics()
	if stats.TotalEntities != 3 {
		t.Fatalf("Expected 3 total entities, got %d", stats.TotalEntities)
	}
	if stats.LiveEntities != 3 {
		t.Fatalf("Expected 3 live entities, got %d", stats.LiveEntities)
	}
	if stats.TypeCounts["LINE"] != 1 {
		t.Fatalf("Expected 1 LINE entity, got %d", stats.TypeCounts["LINE"])
	}
	if stats.TypeCounts["CIRCLE"] != 1 {
		t.Fatalf("Expected 1 CIRCLE entity, got %d", stats.TypeCounts["CIRCLE"])
	}
	if stats.TypeCounts["ARC"] != 1 {
		t.Fatalf("Expected 1 ARC entity, got %d", stats.TypeCounts["ARC"])
	}
}

func TestProfessionalEntityDB_Validation(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("TEST")
	db.AddEntity(entity1)

	// Test Validate
	err := db.Validate()
	if err != nil {
		t.Fatalf("Database validation failed: %v", err)
	}

	// Test RebuildIndexes
	err = db.RebuildIndexes()
	if err != nil {
		t.Fatalf("Failed to rebuild indexes: %v", err)
	}

	// Test Optimize
	err = db.Optimize()
	if err != nil {
		t.Fatalf("Failed to optimize database: %v", err)
	}
}

func TestProfessionalEntityDB_Snapshots(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("TEST")
	db.AddEntity(entity1)

	// Create snapshot
	snapshot := db.CreateSnapshot()
	if snapshot == nil {
		t.Fatal("Failed to create snapshot")
	}

	// Modify database
	db.DeleteEntity(entity1.Handle())

	// Restore from snapshot
	err := db.RestoreSnapshot(snapshot)
	if err != nil {
		t.Fatalf("Failed to restore snapshot: %v", err)
	}

	// Entity should be restored
	_, err = db.GetEntity(entity1.Handle())
	if err != nil {
		t.Fatal("Entity should exist after snapshot restore")
	}
}

func TestProfessionalEntityDB_ConcurrentAccess(t *testing.T) {
	db := NewProfessionalEntityDB()

	// Test concurrent operations (basic test)
	done := make(chan bool, 2)

	// Goroutine 1: Add entities
	go func() {
		for i := 0; i < 10; i++ {
			entity := NewMockEntity("CONCURRENT_TEST")
			db.AddEntity(entity)
		}
		done <- true
	}()

	// Goroutine 2: Query entities
	go func() {
		for i := 0; i < 10; i++ {
			stats := db.GetStatistics()
			_ = stats
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Final validation
	err := db.Validate()
	if err != nil {
		t.Fatalf("Database validation failed after concurrent access: %v", err)
	}
}

func TestProfessionalEntityDB_CacheManagement(t *testing.T) {
	db := NewProfessionalEntityDB()

	entity1 := NewMockEntity("TEST")
	db.AddEntity(entity1)

	// Get cache stats
	stats := db.GetCacheStats()
	if stats["size"] == nil {
		t.Fatal("Cache stats should include size")
	}

	// Clear cache
	err := db.ClearCache()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Cache should be empty
	stats = db.GetCacheStats()
	if size, ok := stats["size"].(int); ok && size != 0 {
		t.Fatalf("Expected empty cache after clear, got size %d", size)
	}
}
