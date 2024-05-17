//go:build js && wasm
// +build js,wasm

package durable

import (
	"context"
	"errors"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
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

// runWithRetry retries if we get a Transaction Finished error.
func (t *DurableTransaction) runWithRetry(fn func(txn *idb.Transaction) error) error {
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

// DurableObjectStore represents an object store that automatically retries on failure.
type DurableObjectStore struct {
	dt    *DurableTransaction
	name  string
	store *idb.ObjectStore
}

// Add creates a structured clone of the value, and stores the cloned value in the object store. This is for adding new records to an object store.
func (d *DurableObjectStore) Add(ctx context.Context, value safejs.Value) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Add(value)
		if err != nil {
			return err
		}
		return req.Await(ctx)
	})
}

// Clear clears the entire object store. This is for deleting all current records out of an object store.
func (d *DurableObjectStore) Clear(ctx context.Context) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Clear()
		if err != nil {
			return err
		}
		return req.Await(ctx)
	})
}

// Count returns the total number of records in the store.
func (d *DurableObjectStore) Count(ctx context.Context) (uint, error) {
	var cnt uint
	rerr := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Count()
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cnt = resp
		return nil
	})
	return cnt, rerr
}

// Delete deletes the store object selected by the specified key. This is for deleting individual records out of an object store.
func (d *DurableObjectStore) Delete(ctx context.Context, key safejs.Value) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Delete(key)
		if err != nil {
			return err
		}
		return req.Await(ctx)
	})
}

// Get returns the objects selected by the specified key. This is for retrieving specific records from an object store.
func (d *DurableObjectStore) Get(ctx context.Context, key safejs.Value) (safejs.Value, error) {
	var value safejs.Value
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Get(key)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		value = resp
		return nil
	})
	return value, err
}

// Put creates a structured clone of the value, and stores the cloned value in the object store. This is for updating existing records in an object store when the transaction's mode is readwrite.
func (d *DurableObjectStore) Put(ctx context.Context, value safejs.Value) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.Put(value)
		if err != nil {
			return err
		}
		_, err = req.Await(ctx)
		return err
	})
}

// PutKey is the same as Put, but includes the key to use to identify the record.
func (d *DurableObjectStore) PutKey(ctx context.Context, key, value safejs.Value) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.PutKey(key, value)
		if err != nil {
			return err
		}
		_, err = req.Await(ctx)
		return err
	})
}

// AddKey is the same as Add, but includes the key to use to identify the record.
func (d *DurableObjectStore) AddKey(ctx context.Context, key, value safejs.Value) error {
	return d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.AddKey(key, value)
		if err != nil {
			return err
		}
		return req.Await(ctx)
	})
}

// GetKey retrieves and returns the record key for the object matching the specified parameter.
func (d *DurableObjectStore) GetKey(ctx context.Context, value safejs.Value) (safejs.Value, error) {
	var key safejs.Value
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.GetKey(value)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		key = resp
		return nil
	})
	return key, err
}

// CountKey returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided key.
func (d *DurableObjectStore) CountKey(ctx context.Context, key safejs.Value) (uint, error) {
	var cnt uint
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.CountKey(key)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cnt = resp
		return nil
	})
	return cnt, err
}

// CountRange returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided KeyRange.
func (d *DurableObjectStore) CountRange(ctx context.Context, keyRange *idb.KeyRange) (uint, error) {
	var cnt uint
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.CountRange(keyRange)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cnt = resp
		return nil
	})
	return cnt, err
}

// GetAllKeys returns an ArrayRequest that retrieves record keys for all objects in the object store.
func (d *DurableObjectStore) GetAllKeys(ctx context.Context) ([]safejs.Value, error) {
	var keys []safejs.Value
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.GetAllKeys()
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		keys = resp
		return nil
	})
	return keys, err
}

// GetAllKeysRange returns an ArrayRequest that retrieves record keys for all objects in the object store matching the specified query. If maxCount is 0, retrieves all objects matching the query.
func (d *DurableObjectStore) GetAllKeysRange(ctx context.Context, query *idb.KeyRange, maxCount uint) ([]safejs.Value, error) {
	var keys []safejs.Value
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.GetAllKeysRange(query, maxCount)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		keys = resp
		return nil
	})
	return keys, err
}

// OpenCursor returns a CursorWithValueRequest, and, in a separate thread, returns a new CursorWithValue. Used for iterating through an object store by primary key with a cursor.
func (d *DurableObjectStore) OpenCursor(ctx context.Context, direction idb.CursorDirection) (*idb.CursorWithValue, error) {
	var cursor *idb.CursorWithValue
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenCursor(direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}

// OpenCursorKey is the same as OpenCursor, but opens a cursor over the given key instead.
func (d *DurableObjectStore) OpenCursorKey(ctx context.Context, key safejs.Value, direction idb.CursorDirection) (*idb.CursorWithValue, error) {
	var cursor *idb.CursorWithValue
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenCursorKey(key, direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}

// OpenCursorRange is the same as OpenCursor, but opens a cursor over the given range instead.
func (d *DurableObjectStore) OpenCursorRange(ctx context.Context, keyRange *idb.KeyRange, direction idb.CursorDirection) (*idb.CursorWithValue, error) {
	var cursor *idb.CursorWithValue
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenCursorRange(keyRange, direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}

// OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a new Cursor. Used for iterating through all keys in an object store.
func (d *DurableObjectStore) OpenKeyCursor(ctx context.Context, direction idb.CursorDirection) (*idb.Cursor, error) {
	var cursor *idb.Cursor
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenKeyCursor(direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}

// OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the given key instead.
func (d *DurableObjectStore) OpenKeyCursorKey(ctx context.Context, key safejs.Value, direction idb.CursorDirection) (*idb.Cursor, error) {
	var cursor *idb.Cursor
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenKeyCursorKey(key, direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}

// OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the given key range instead.
func (d *DurableObjectStore) OpenKeyCursorRange(ctx context.Context, keyRange *idb.KeyRange, direction idb.CursorDirection) (*idb.Cursor, error) {
	var cursor *idb.Cursor
	err := d.dt.runWithRetry(func(txn *idb.Transaction) error {
		req, err := d.store.OpenKeyCursorRange(keyRange, direction)
		if err != nil {
			return err
		}
		resp, err := req.Await(ctx)
		if err != nil {
			return err
		}
		cursor = resp
		return nil
	})
	return cursor, err
}
