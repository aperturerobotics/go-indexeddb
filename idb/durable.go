//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"errors"
	"strings"

	"github.com/hack-pad/safejs"
)

/*
RetryTxn retries the function with a new transaction if the txn finishes prematurely.

IndexedDB transactions automatically commit when all outstanding requests have
been satisfied. When a Goroutine is suspended due to a select statement or other
context switching, the IndexedDB transation commits automatically, leading to
errors with a suffix "The transaction has finished."

See: https://github.com/w3c/IndexedDB/issues/34 for more details.

RetryTxn is a mechanism that automatically re-creates the transaction and
retries the operation whenever we encounter this specific error. This
ensures that operations can continue even if the transaction has been
automatically committed.
*/
func RetryTxn(
	ctx context.Context,
	db *Database,
	txnMode TransactionMode,
	fn func(txn *Transaction) error,
	objectStoreName string,
	objectStoreNames ...string,
) error {
	for {
		txn, err := db.Transaction(txnMode, objectStoreName, objectStoreNames...)
		if err != nil {
			return err
		}

		// call the fn
		err = fn(txn)

		// if the fn returns txn finished, retry.
		if IsTxnFinishedErr(err) {
			continue
		}

		// check for error performing the operation
		if err != nil {
			_ = txn.Abort()
			return err
		}

		// commit the txn
		err = txn.Commit()
		if IsTxnFinishedErr(err) {
			// txn committed automatically already
			err = nil
		}

		return err
	}
}

// IsTxnFinishedErr checks if an error corresponds to a transaction finishing.
// see RetryTxn for details
func IsTxnFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasSuffix(err.Error(), "The transaction has finished.")
}

// DurableTransaction represents a transaction that automatically retries on failure.
type DurableTransaction struct {
	db               *Database
	txnMode          TransactionMode
	objectStoreNames []string
	txn              *Transaction
	objectStores     map[string]*DurableObjectStore
}

// NewDurableTransaction creates a new DurableTransaction.
func NewDurableTransaction(db *Database, txnMode TransactionMode, objectStoreNames ...string) (*DurableTransaction, error) {
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
func (dt *DurableTransaction) GetObjectStore(name string) (*DurableObjectStore, error) {
	store, ok := dt.objectStores[name]
	if !ok {
		return nil, errors.New("object store not available in this txn")
	}
	return store, nil
}

// ensureTransaction ensures dt.txn is not nil.
func (dt *DurableTransaction) ensureTransaction() error {
	if dt.txn != nil {
		return nil
	}

	txn, err := dt.db.Transaction(dt.txnMode, dt.objectStoreNames[0], dt.objectStoreNames[1:]...)
	if err != nil {
		return err
	}
	dt.txn = txn

	for name, durableStore := range dt.objectStores {
		store, err := dt.txn.ObjectStore(name)
		if err != nil {
			return err
		}
		durableStore.store = store
	}

	return nil
}

// runWithRetry retries if we get a Transaction Finished error.
func (dt *DurableTransaction) runWithRetry(fn func(txn *Transaction) error) error {
	for {
		if err := dt.ensureTransaction(); err != nil {
			return err
		}

		err := fn(dt.txn)
		if err == nil {
			return nil
		}

		if !IsTxnFinishedErr(err) {
			return err
		}

		// mark txn as nil and retry
		dt.txn = nil
	}
}

// DurableObjectStore represents an object store that automatically retries on failure.
type DurableObjectStore struct {
	dt    *DurableTransaction
	name  string
	store *ObjectStore
}

// Add returns an AckRequest, and, in a separate thread, creates a structured clone of the value, and stores the cloned value in the object store. This is for adding new records to an object store.
func (dos *DurableObjectStore) Add(value safejs.Value) (*AckRequest, error) {
	var req *AckRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Add(value)
		return err
	})
	return req, rerr
}

// Clear returns an AckRequest, then clears this object store in a separate thread. This is for deleting all current records out of an object store.
func (dos *DurableObjectStore) Clear() (*AckRequest, error) {
	var req *AckRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Clear()
		return err
	})
	return req, rerr
}

// Count returns a UintRequest, and, in a separate thread, returns the total number of records in the store.
func (dos *DurableObjectStore) Count() (*UintRequest, error) {
	var req *UintRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Count()
		return err
	})
	return req, rerr
}

// Delete returns an AckRequest, and, in a separate thread, deletes the store object selected by the specified key. This is for deleting individual records out of an object store.
func (dos *DurableObjectStore) Delete(key safejs.Value) (*AckRequest, error) {
	var req *AckRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Delete(key)
		return err
	})
	return req, rerr
}

// Get returns a Request, and, in a separate thread, returns the objects selected by the specified key. This is for retrieving specific records from an object store.
func (dos *DurableObjectStore) Get(key safejs.Value) (*Request, error) {
	var req *Request
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Get(key)
		return err
	})
	return req, rerr
}

// Put returns a Request, and, in a separate thread, creates a structured clone of the value, and stores the cloned value in the object store. This is for updating existing records in an object store when the transaction's mode is readwrite.
func (dos *DurableObjectStore) Put(value safejs.Value) (*Request, error) {
	var req *Request
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.Put(value)
		return err
	})
	return req, rerr
}

// PutKey is the same as Put, but includes the key to use to identify the record.
func (dos *DurableObjectStore) PutKey(key, value safejs.Value) (*Request, error) {
	var req *Request
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.PutKey(key, value)
		return err
	})
	return req, rerr
}

// AddKey is the same as Add, but includes the key to use to identify the record.
func (dos *DurableObjectStore) AddKey(key, value safejs.Value) (*AckRequest, error) {
	var req *AckRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.AddKey(key, value)
		return err
	})
	return req, rerr
}

// GetKey returns a Request, and, in a separate thread retrieves and returns the record key for the object matching the specified parameter.
func (dos *DurableObjectStore) GetKey(value safejs.Value) (*Request, error) {
	var req *Request
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.GetKey(value)
		return err
	})
	return req, rerr
}

// CountKey returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided key.
func (dos *DurableObjectStore) CountKey(key safejs.Value) (*UintRequest, error) {
	var req *UintRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.CountKey(key)
		return err
	})
	return req, rerr
}

// CountRange returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided KeyRange.
func (dos *DurableObjectStore) CountRange(keyRange *KeyRange) (*UintRequest, error) {
	var req *UintRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.CountRange(keyRange)
		return err
	})
	return req, rerr
}

// GetAllKeys returns an ArrayRequest that retrieves record keys for all objects in the object store.
func (dos *DurableObjectStore) GetAllKeys() (*ArrayRequest, error) {
	var req *ArrayRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.GetAllKeys()
		return err
	})
	return req, rerr
}

// GetAllKeysRange returns an ArrayRequest that retrieves record keys for all objects in the object store matching the specified query. If maxCount is 0, retrieves all objects matching the query.
func (dos *DurableObjectStore) GetAllKeysRange(query *KeyRange, maxCount uint) (*ArrayRequest, error) {
	var req *ArrayRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.GetAllKeysRange(query, maxCount)
		return err
	})
	return req, rerr
}

// OpenCursor returns a CursorWithValueRequest, and, in a separate thread, returns a new CursorWithValue. Used for iterating through an object store by primary key with a cursor.
func (dos *DurableObjectStore) OpenCursor(direction CursorDirection) (*CursorWithValueRequest, error) {
	var req *CursorWithValueRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenCursor(direction)
		return err
	})
	return req, rerr
}

// OpenCursorKey is the same as OpenCursor, but opens a cursor over the given key instead.
func (dos *DurableObjectStore) OpenCursorKey(key safejs.Value, direction CursorDirection) (*CursorWithValueRequest, error) {
	var req *CursorWithValueRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenCursorKey(key, direction)
		return err
	})
	return req, rerr
}

// OpenCursorRange is the same as OpenCursor, but opens a cursor over the given range instead.
func (dos *DurableObjectStore) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorWithValueRequest, error) {
	var req *CursorWithValueRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenCursorRange(keyRange, direction)
		return err
	})
	return req, rerr
}

// OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a new Cursor. Used for iterating through all keys in an object store.
func (dos *DurableObjectStore) OpenKeyCursor(direction CursorDirection) (*CursorRequest, error) {
	var req *CursorRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenKeyCursor(direction)
		return err
	})
	return req, rerr
}

// OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the given key instead.
func (dos *DurableObjectStore) OpenKeyCursorKey(key safejs.Value, direction CursorDirection) (*CursorRequest, error) {
	var req *CursorRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenKeyCursorKey(key, direction)
		return err
	})
	return req, rerr
}

// OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the given key range instead.
func (dos *DurableObjectStore) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorRequest, error) {
	var req *CursorRequest
	rerr := dos.dt.runWithRetry(func(txn *Transaction) (err error) {
		req, err = dos.store.OpenKeyCursorRange(keyRange, direction)
		return err
	})
	return req, rerr
}
