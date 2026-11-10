package drawing

import (
	"errors"
	"maps"

	"github.com/edanko/dxf/entity"
)

// DrawingTransaction is an all-or-nothing editing session over the model space,
// mirroring ezdxf's doc.transaction() context manager. All entities added
// or removed while the transaction is open become visible only on Commit;
// Rollback restores the previous model space completely.
//
// Only one transaction may be active per drawing at a time.
// copyEntity returns a snapshot copy of an entity when the concrete type
// supports it; otherwise the snapshot shares the entity (mutations of an
// existing entity are not rolled back, matching ezdxf transaction scope
// over add/remove).
func copyEntity(e entity.Entity) entity.Entity {
	if c, ok := e.(entity.Copier); ok {
		return c.Copy()
	}
	return e
}

type DrawingTransaction struct {
	drawing        *Drawing
	active         bool
	snapshot       []entity.Entity
	snapshotLayout map[string]string
	added          int
	removed        int
}

// BeginTransaction starts a transaction over the model space entities.
// Returns an error if a transaction is already active.
func (d *Drawing) BeginTransaction() (*DrawingTransaction, error) {
	if d.transaction != nil && d.transaction.active {
		return nil, errors.New("a transaction is already active")
	}
	txn := &DrawingTransaction{
		drawing:        d,
		active:         true,
		snapshot:       make([]entity.Entity, 0, len(d.entitiesSection)),
		snapshotLayout: make(map[string]string, len(d.entityLayout)),
	}
	for _, e := range d.entitiesSection {
		txn.snapshot = append(txn.snapshot, copyEntity(e))
	}
	maps.Copy(txn.snapshotLayout, d.entityLayout)
	d.transaction = txn
	return txn, nil
}

// InTransaction reports whether a transaction is currently active.
func (d *Drawing) InTransaction() bool {
	return d.transaction != nil && d.transaction.active
}

// Commit makes all changes performed during the transaction permanent.
func (t *DrawingTransaction) Commit() error {
	if !t.active {
		return errors.New("transaction is not active")
	}
	t.active = false
	t.snapshot = nil
	t.snapshotLayout = nil
	if t.drawing.transaction == t {
		t.drawing.transaction = nil
	}
	return nil
}

// Rollback discards all changes performed during the transaction and
// restores the model space to the state captured at BeginTransaction.
func (t *DrawingTransaction) Rollback() error {
	if !t.active {
		return errors.New("transaction is not active")
	}
	restored := make([]entity.Entity, 0, len(t.snapshot))
	for _, e := range t.snapshot {
		restored = append(restored, copyEntity(e))
	}
	t.drawing.entitiesSection = restored
	t.drawing.entityLayout = make(map[string]string, len(t.snapshotLayout))
	maps.Copy(t.drawing.entityLayout, t.snapshotLayout)
	t.active = false
	t.snapshot = nil
	t.snapshotLayout = nil
	if t.drawing.transaction == t {
		t.drawing.transaction = nil
	}
	return nil
}

// Stats returns the number of entities added and removed since the
// transaction began (informational).
func (t *DrawingTransaction) Stats() (added int, removed int) {
	return t.added, t.removed
}

// recordTransactionActivity updates transaction counters from Space.Add
// and DeleteEntity.
func (d *Drawing) recordTransactionActivity(added, removed int) {
	if d.InTransaction() {
		d.transaction.added += added
		d.transaction.removed += removed
	}
}
