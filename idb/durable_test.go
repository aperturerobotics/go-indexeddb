//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"errors"
	"syscall/js"
	"testing"

	"github.com/aperturerobotics/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

func TestRetryTxn(t *testing.T) {
	t.Parallel()

	const storeName = "mystore"
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{})
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		var callCount int
		err := RetryTxn(context.Background(), db, TransactionReadWrite, func(txn *Transaction) error {
			callCount++
			store, err := txn.ObjectStore(storeName)
			assert.NoError(t, err)
			_, err = store.PutKey(safejs.Safe(js.ValueOf("key")), safejs.Safe(js.ValueOf("some value")))
			return err
		}, storeName)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("retry on txn finished", func(t *testing.T) {
		t.Parallel()
		var callCount int
		err := RetryTxn(context.Background(), db, TransactionReadWrite, func(txn *Transaction) error {
			callCount++
			store, err := txn.ObjectStore(storeName)
			assert.NoError(t, err)
			_, err = store.PutKey(safejs.Safe(js.ValueOf("key")), safejs.Safe(js.ValueOf("some value")))
			assert.NoError(t, err)
			if callCount == 1 {
				return errors.New("The transaction has finished.")
			}
			return nil
		}, storeName)
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("return other error", func(t *testing.T) {
		t.Parallel()
		err := RetryTxn(context.Background(), db, TransactionReadWrite, func(txn *Transaction) error {
			return errors.New("some error")
		}, storeName)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "some error")
	})
}

func TestIsTxnFinishedErr(t *testing.T) {
	t.Parallel()
	assert.Equal(t, false, IsTxnFinishedErr(nil))
	assert.Equal(t, false, IsTxnFinishedErr(errors.New("some error")))
	assert.Equal(t, true, IsTxnFinishedErr(errors.New("The transaction has finished.")))
}
func TestDurableTransaction(t *testing.T) {
	t.Parallel()
	const storeName = "mystore"
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{})
		assert.NoError(t, err)
	})

	txn, err := NewDurableTransaction(db, TransactionReadWrite, storeName)
	assert.NoError(t, err)

	store, err := txn.GetObjectStore(storeName)
	assert.NoError(t, err)
	assert.NotEqual(t, store, nil)

	_, err = txn.GetObjectStore("invalid")
	assert.Error(t, err)
}

func TestDurableObjectStore(t *testing.T) {
	t.Parallel()
	const storeName = "mystore"
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{})
		assert.NoError(t, err)
	})

	txn, err := NewDurableTransaction(db, TransactionReadWrite, storeName)
	assert.NoError(t, err)

	store, err := txn.GetObjectStore(storeName)
	assert.NoError(t, err)

	t.Run("AddKey", func(t *testing.T) {
		req, err := store.AddKey(safejs.Safe(js.ValueOf("key")), safejs.Safe(js.ValueOf("value")))
		assert.NoError(t, err)
		assert.NoError(t, req.Await(context.Background()))
	})

	t.Run("Clear", func(t *testing.T) {
		req, err := store.Clear()
		assert.NoError(t, err)
		assert.NoError(t, req.Await(context.Background()))
	})

	t.Run("Count", func(t *testing.T) {
		req, err := store.Count()
		assert.NoError(t, err)
		count, err := req.Await(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, uint(0), count)
	})

	t.Run("Delete", func(t *testing.T) {
		req, err := store.Delete(safejs.Safe(js.ValueOf("key")))
		assert.NoError(t, err)
		assert.NoError(t, req.Await(context.Background()))
	})

	t.Run("Get", func(t *testing.T) {
		req, err := store.Get(safejs.Safe(js.ValueOf("key")))
		assert.NoError(t, err)
		_, err = req.Await(context.Background())
		assert.NoError(t, err)
	})

	t.Run("PutKey", func(t *testing.T) {
		req, err := store.PutKey(safejs.Safe(js.ValueOf("key")), safejs.Safe(js.ValueOf("value")))
		assert.NoError(t, err)
		_, err = req.Await(context.Background())
		assert.NoError(t, err)
	})
}
