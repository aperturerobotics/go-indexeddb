//go:build js && wasm
// +build js,wasm

package durable

import (
	"errors"

	"github.com/aperturerobotics/go-indexeddb/idb"
)

// DurableTransaction represents a transaction that automatically retries on
// failure due to the transaction finishing prematurely.
//
// See: ../../README.md#Transactions-Expiring
type DurableTransaction struct {
	db               *idb.Database
	txnMode          idb.TransactionMode
	objectStoreNames []string
	txn              *idb.Transaction
	objectStores     map[string]*DurableObjectStore
}

// NewDurableTransaction creates a new DurableTransaction.
func NewDurableTransaction(db *idb.Database, txnMode idb.TransactionMode, objectStoreNames ...string) (*DurableTransaction, error) {
	if len(objectStoreNames) == 0 {
		return nil, errors.New("transaction must have at least one object store")
	}

	dt := &DurableTransaction{
		db:               db,
		txnMode:          txnMode,
		objectStoreNames: objectStoreNames,
		objectStores:     make(map[string]*DurableObjectStore),
	}

	if err := dt.ensureTransaction(); err != nil {
		return nil, err
	}

	for _, name := range objectStoreNames {
		store, err := dt.txn.ObjectStore(name)
		if err != nil {
			return nil, err
		}
		dt.objectStores[name] = &DurableObjectStore{
			dt:    dt,
			name:  name,
			store: store,
		}
	}

	return dt, nil
}

// GetObjectStore returns the DurableObjectStore for the given name.
func (t *DurableTransaction) GetObjectStore(name string) (*DurableObjectStore, error) {
	store, ok := t.objectStores[name]
	if !ok {
		return nil, errors.New("object store not available in this txn")
	}
	return store, nil
}

// Abort attempts to abort the transaction (undoing the ops) if one is active.
// no-op if the transaction was already committed
// Returns if the abort request did anything and any error.
// NOTE: the transaction will commit automatically if the goroutine is backgrounded.
func (t *DurableTransaction) Abort() (bool, error) {
	if t.txn == nil {
		return false, nil
	}

	err := t.txn.Abort()
	t.txn = nil
	if err == nil {
		return true, nil
	}
	if idb.IsTxnFinishedErr(err) {
		return false, nil
	}
	return false, err
}

// Commit attempts to commit the transaction if one is active.
// no-op if the transaction was already committed
// NOTE: the transaction will commit automatically if the goroutine is backgrounded.
func (t *DurableTransaction) Commit() error {
	if t.txn == nil {
		return nil
	}

	err := t.txn.Commit()
	t.txn = nil
	if idb.IsTxnFinishedErr(err) {
		err = nil
	}
	return err
}

// ensureTransaction ensures dt.txn is not nil.
func (t *DurableTransaction) ensureTransaction() error {
	if t.txn != nil {
		return nil
	}

	txn, err := t.db.Transaction(t.txnMode, t.objectStoreNames[0], t.objectStoreNames[1:]...)
	if err != nil {
		return err
	}
	t.txn = txn

	for name, durableStore := range t.objectStores {
		store, err := t.txn.ObjectStore(name)
		if err != nil {
			return err
		}
		durableStore.store = store
	}

	return nil
}

// TxnWithRetry retries if we get a Transaction Finished error.
func (t *DurableTransaction) TxnWithRetry(fn func(txn *idb.Transaction) error) error {
	for {
		if err := t.ensureTransaction(); err != nil {
			return err
		}

		err := fn(t.txn)
		if err == nil {
			return nil
		}

		if !idb.IsTxnFinishedErr(err) {
			return err
		}

		// mark txn as nil and retry
		t.txn = nil
	}
}
