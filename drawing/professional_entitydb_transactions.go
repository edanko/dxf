package drawing

import (
	"fmt"
)

// BeginTransaction starts a new database transaction
func (db *ProfessionalEntityDB) BeginTransaction(id string) (*Transaction, error) {
	// Create snapshot before acquiring lock to avoid deadlock
	// (CreateSnapshot tries to acquire RLock which would conflict with our Lock)
	snapshot := db.CreateSnapshot()

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.currentTx != nil {
		return nil, fmt.Errorf("transaction already in progress")
	}

	tx := NewTransaction(id)
	tx.Snapshot = snapshot
	db.currentTx = tx
	db.transactions = append(db.transactions, tx)

	return tx, nil
}

// CommitTransaction commits the current transaction
func (db *ProfessionalEntityDB) CommitTransaction(tx *Transaction) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.currentTx == nil {
		return fmt.Errorf("no transaction in progress")
	}

	if db.currentTx.ID != tx.ID {
		return fmt.Errorf("transaction mismatch")
	}

	if !tx.Active {
		return fmt.Errorf("transaction is not active")
	}

	// Execute all operations in transaction
	for _, op := range tx.Operations {
		if err := op.Execute(db); err != nil {
			// If any operation fails, rollback the entire transaction
			db.rollbackTransaction(tx)
			return fmt.Errorf("transaction failed during commit: %v", err)
		}
	}

	// Mark transaction as completed
	tx.Active = false
	db.currentTx = nil
	db.version++

	return nil
}

// RollbackTransaction rolls back the current transaction
func (db *ProfessionalEntityDB) RollbackTransaction(tx *Transaction) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.currentTx == nil {
		return fmt.Errorf("no transaction in progress")
	}

	if db.currentTx.ID != tx.ID {
		return fmt.Errorf("transaction mismatch")
	}

	if !tx.Active {
		return fmt.Errorf("transaction is not active")
	}

	return db.rollbackTransaction(tx)
}

// rollbackTransaction performs the actual rollback (internal method)
func (db *ProfessionalEntityDB) rollbackTransaction(tx *Transaction) error {
	// Rollback operations in reverse order
	for i := len(tx.Operations) - 1; i >= 0; i-- {
		op := tx.Operations[i]
		if err := op.Rollback(db); err != nil {
			return fmt.Errorf("rollback failed for operation %s: %v", op.Description(), err)
		}
	}

	// Restore snapshot (lock already held by caller)
	if tx.Snapshot != nil {
		if err := db.restoreSnapshotInternal(tx.Snapshot); err != nil {
			return fmt.Errorf("failed to restore snapshot: %v", err)
		}
	}

	// Mark transaction as inactive
	tx.Active = false
	db.currentTx = nil

	return nil
}

// GetCurrentTransaction returns the current active transaction
func (db *ProfessionalEntityDB) GetCurrentTransaction() *Transaction {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.currentTx
}

// GetTransactionHistory returns all completed transactions
func (db *ProfessionalEntityDB) GetTransactionHistory() []*Transaction {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]*Transaction, len(db.transactions))
	copy(history, db.transactions)
	return history
}

// ClearTransactionHistory clears all transaction history
func (db *ProfessionalEntityDB) ClearTransactionHistory() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Only allow clearing if no active transaction
	if db.currentTx != nil {
		return fmt.Errorf("cannot clear history with active transaction")
	}

	db.transactions = make([]*Transaction, 0)
	return nil
}
