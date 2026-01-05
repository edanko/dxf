package drawing

import (
	"fmt"
	"sync"
	"time"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/spatial"
)

// EntityEntry represents an entity with additional metadata
type EntityEntry struct {
	Entity     entity.Entity
	IsAlive    bool
	CreatedAt  time.Time
	ModifiedAt time.Time
	Version    int
	Tags       map[string]interface{} // Additional metadata
}

// DatabaseSnapshot represents state of database at a point in time
type DatabaseSnapshot struct {
	Timestamp     time.Time
	Entities      map[string]*EntityEntry
	HandleCounter int
	Version       int
}

// OperationType represents the type of database operation
type OperationType int

const (
	OperationAdd OperationType = iota
	OperationUpdate
	OperationDelete
	OperationBulkAdd
	OperationBulkDelete
)

// Operation represents a database operation for transaction support
type Operation interface {
	Execute(db *ProfessionalEntityDB) error
	Rollback(db *ProfessionalEntityDB) error
	Type() OperationType
	EntityHandle() string
	Description() string
}

// AddOperation represents adding an entity
type AddOperation struct {
	Entity   entity.Entity
	OldEntry *EntityEntry
	Desc     string
}

func (op *AddOperation) Execute(db *ProfessionalEntityDB) error {
	return db.addEntityInternal(op.Entity)
}

func (op *AddOperation) Rollback(db *ProfessionalEntityDB) error {
	if op.OldEntry == nil {
		// Entity was never committed, nothing to rollback
		return nil
	}
	return db.restoreEntityEntry(op.Entity.Handle(), op.OldEntry)
}

func (op *AddOperation) Type() OperationType  { return OperationAdd }
func (op *AddOperation) EntityHandle() string { return op.Entity.Handle() }
func (op *AddOperation) Description() string  { return op.Desc }

// UpdateOperation represents updating an entity
type UpdateOperation struct {
	Entity   entity.Entity
	OldEntry *EntityEntry
	Desc     string
}

func (op *UpdateOperation) Execute(db *ProfessionalEntityDB) error {
	return db.updateEntityInternal(op.Entity)
}

func (op *UpdateOperation) Rollback(db *ProfessionalEntityDB) error {
	if op.OldEntry != nil {
		return db.restoreEntityEntry(op.Entity.Handle(), op.OldEntry)
	}
	return fmt.Errorf("cannot rollback update: no old entry stored")
}

func (op *UpdateOperation) Type() OperationType  { return OperationUpdate }
func (op *UpdateOperation) EntityHandle() string { return op.Entity.Handle() }
func (op *UpdateOperation) Description() string  { return op.Desc }

// DeleteOperation represents deleting an entity
type DeleteOperation struct {
	HandleStr     string
	OldEntry      *EntityEntry
	OperationDesc string
}

func (op *DeleteOperation) Execute(db *ProfessionalEntityDB) error {
	return db.deleteEntityInternal(op.HandleStr)
}

func (op *DeleteOperation) Rollback(db *ProfessionalEntityDB) error {
	return db.restoreEntityEntry(op.HandleStr, op.OldEntry)
}

func (op *DeleteOperation) Type() OperationType  { return OperationDelete }
func (op *DeleteOperation) EntityHandle() string { return op.HandleStr }
func (op *DeleteOperation) Description() string  { return op.OperationDesc }

// Transaction represents a database transaction
type Transaction struct {
	ID         string
	StartTime  time.Time
	Operations []Operation
	Snapshot   *DatabaseSnapshot
	Active     bool
}

// NewTransaction creates a new transaction
func NewTransaction(id string) *Transaction {
	return &Transaction{
		ID:         id,
		StartTime:  time.Now(),
		Operations: make([]Operation, 0),
		Active:     true,
	}
}

// DatabaseStats provides comprehensive database statistics
type DatabaseStats struct {
	TotalEntities      int
	LiveEntities       int
	DeadEntities       int
	Transactions       int
	ModifiedCount      int
	LastModified       time.Time
	TypeCounts         map[string]int
	LayerCounts        map[string]int
	SpatialIndexHits   int
	SpatialIndexMisses int
}

// ProfessionalEntityDB provides professional-grade entity database management
type ProfessionalEntityDB struct {
	mu sync.RWMutex

	// Core storage
	entities map[string]*EntityEntry
	handles  *handle.HandleGenerator

	// Indexing
	spatialIndex *spatial.RTree
	typeIndex    map[string]map[string]bool // type -> handles
	layerIndex   map[string]map[string]bool // layer -> handles

	// Lifecycle management
	liveEntities map[string]bool
	trashCan     []string
	version      int

	// Transaction support
	transactions []*Transaction
	currentTx    *Transaction

	// Performance and caching
	cache       map[string]*EntityEntry
	stats       *DatabaseStats
	lastCleanup time.Time

	// Configuration
	enableSpatialIndex bool
	enableTypeIndex    bool
	enableLayerIndex   bool
	cacheSize          int
}

// NewProfessionalEntityDB creates a new professional entity database
func NewProfessionalEntityDB() *ProfessionalEntityDB {
	return &ProfessionalEntityDB{
		entities:     make(map[string]*EntityEntry),
		handles:      handle.NewHandleGenerator(),
		spatialIndex: spatial.NewRTree(spatial.DefaultRTreeOptions()),
		typeIndex:    make(map[string]map[string]bool),
		layerIndex:   make(map[string]map[string]bool),
		liveEntities: make(map[string]bool),
		trashCan:     make([]string, 0),
		transactions: make([]*Transaction, 0),
		cache:        make(map[string]*EntityEntry),
		stats: &DatabaseStats{
			TypeCounts:  make(map[string]int),
			LayerCounts: make(map[string]int),
		},
		lastCleanup:        time.Now(),
		enableSpatialIndex: true,
		enableTypeIndex:    true,
		enableLayerIndex:   true,
		cacheSize:          1000,
	}
}

// NewProfessionalEntityDBWithConfig creates a new professional entity database with custom configuration
func NewProfessionalEntityDBWithConfig(config ProfessionalDBConfig) *ProfessionalEntityDB {
	db := NewProfessionalEntityDB()
	db.enableSpatialIndex = config.EnableSpatialIndex
	db.enableTypeIndex = config.EnableTypeIndex
	db.enableLayerIndex = config.EnableLayerIndex
	db.cacheSize = config.CacheSize

	return db
}

// ProfessionalDBConfig provides configuration options for professional entity database
type ProfessionalDBConfig struct {
	EnableSpatialIndex bool
	EnableTypeIndex    bool
	EnableLayerIndex   bool
	CacheSize          int
}

// DatabaseInterface defines the interface for entity database operations
type DatabaseInterface interface {
	// Basic CRUD operations
	AddEntity(e entity.Entity) error
	GetEntity(handle string) (entity.Entity, error)
	UpdateEntity(e entity.Entity) error
	DeleteEntity(handle string) error

	// Query operations
	GetEntitiesByType(entityType string) []entity.Entity
	GetEntitiesByLayer(layerName string) []entity.Entity
	GetEntitiesInBounds(min, max math.Vec2) []entity.Entity
	FilterEntities(predicate func(entity.Entity) bool) []entity.Entity

	// Lifecycle operations
	IsEntityAlive(handle string) bool
	GetLiveEntities() []entity.Entity
	GetTrashCan() []string
	EmptyTrashCan() error

	// Transaction operations
	BeginTransaction(id string) (*Transaction, error)
	CommitTransaction(tx *Transaction) error
	RollbackTransaction(tx *Transaction) error
	GetCurrentTransaction() *Transaction

	// Advanced operations
	CloneEntity(handle string, newHandle string) (entity.Entity, error)
	BulkAddEntities(entities []entity.Entity) error
	BulkDeleteEntities(handles []string) error

	// Indexing and performance
	RebuildIndexes() error
	GetStatistics() *DatabaseStats
	Optimize() error

	// Validation and integrity
	Validate() error
	CreateSnapshot() *DatabaseSnapshot
	RestoreSnapshot(snapshot *DatabaseSnapshot) error

	// Cache management
	ClearCache() error
	GetCacheStats() map[string]interface{}
}
