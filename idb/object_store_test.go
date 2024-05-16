//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/aperturerobotics/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

func TestObjectStoreIndexNames(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", safejs.Safe(js.ValueOf("indexKey")), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	names, err := store.IndexNames()
	assert.NoError(t, err)
	assert.Equal(t, []string{"myindex"}, names)
}

func TestObjectStoreKeyPath(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("primary"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	keyPath, err := store.KeyPath()
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf("primary")), keyPath)
}

func TestObjectStoreName(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	name, err := store.Name()
	assert.NoError(t, err)
	assert.Equal(t, "mystore", name)
}

func TestObjectStoreAutoIncrement(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			AutoIncrement: true,
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	autoIncrement, err := store.AutoIncrement()
	assert.NoError(t, err)
	assert.Equal(t, true, autoIncrement)
}

func TestObjectStoreTransaction(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("primary"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	txnGet, err := store.Transaction()
	assert.NoError(t, err)
	assert.Equal(t, txn.jsTransaction, txnGet.jsTransaction)
}

func TestObjectStoreAdd(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("id"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	addReq, err := store.Add(js.ValueOf(map[string]interface{}{
		"id": "some id",
	}))
	assert.NoError(t, err)
	getVal, err := safejs.ValueOf("some id")
	assert.NoError(t, err)
	getReq, err := store.GetKey(getVal)
	assert.NoError(t, err)

	assert.NoError(t, addReq.Await(ctx))
	result, err := getReq.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf("some id")), result)
}

func TestObjectStoreClear(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	{
		txn, err := db.Transaction(TransactionReadWrite, "mystore")
		assert.NoError(t, err)
		store, err := txn.ObjectStore("mystore")
		assert.NoError(t, err)
		_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
		assert.NoError(t, err)
		assert.NoError(t, txn.Await(ctx))
	}

	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	clearReq, err := store.Clear()
	assert.NoError(t, err)
	getReq, err := store.GetAllKeys()
	assert.NoError(t, err)

	assert.NoError(t, clearReq.Await(ctx))
	result, err := getReq.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []js.Value(nil), result)
}

func TestObjectStoreCount(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		countFn func(*ObjectStore) (*UintRequest, error)
	}{
		{
			name: "count",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				return store.Count()
			},
		},
		{
			name: "count key",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				val, err := safejs.ValueOf("some key")
				if err != nil {
					return nil, err
				}
				return store.CountKey(val)
			},
		},
		{
			name: "count range",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				keyRange, err := NewKeyRangeOnly(safejs.Safe(js.ValueOf("some key")))
				assert.NoError(t, err)
				return store.CountRange(keyRange)
			},
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)

			_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
			assert.NoError(t, err)

			ctx := context.Background()
			req, err := tc.countFn(store)
			assert.NoError(t, err)
			count, err := req.Await(ctx)
			assert.NoError(t, err)

			assert.Equal(t, uint(1), count)
		})
	}
}

func TestObjectStoreCreateIndex(t *testing.T) {
	t.Parallel()
	testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		index, err := store.CreateIndex("myindex", safejs.Safe(js.ValueOf("primary")), IndexOptions{
			Unique:     true,
			MultiEntry: true,
		})
		assert.NoError(t, err)

		unique, err := index.Unique()
		assert.NoError(t, err)
		assert.Equal(t, true, unique)
		multiEntry, err := index.MultiEntry()
		assert.NoError(t, err)
		assert.Equal(t, true, multiEntry)
	})
}

func TestObjectStoreDelete(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
	assert.NoError(t, err)

	_, err = store.Delete(js.ValueOf("some key"))
	assert.NoError(t, err)
	getKeyVal, err := safejs.ValueOf("some key")
	assert.NoError(t, err)
	req, err := store.GetKey(getKeyVal)
	assert.NoError(t, err)
	result, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.Undefined()), result)
}

func TestObjectStoreDeleteIndex(t *testing.T) {
	t.Parallel()
	testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", safejs.Safe(js.ValueOf("primary")), IndexOptions{})
		assert.NoError(t, err)
		err = store.DeleteIndex("myindex")
		assert.NoError(t, err)
		names, err := store.IndexNames()
		assert.NoError(t, err)
		assert.Equal(t, []string(nil), names)
	})
}

func TestObjectStoreGet(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		keys         map[string]interface{}
		getFn        func(*ObjectStore) (interface{}, error)
		expectResult interface{}
	}{
		{
			name: "get all keys",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				return store.GetAllKeys()
			},
			expectResult: []js.Value{js.ValueOf("some id"), js.ValueOf("some other id")},
		},
		{
			name: "get all keys query",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeOnly(safejs.Safe(js.ValueOf("some id")))
				assert.NoError(t, err)
				return store.GetAllKeysRange(keyRange, 10)
			},
			expectResult: []js.Value{js.ValueOf("some id")},
		},
		{
			name: "get",
			keys: map[string]interface{}{
				"some id": "some value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				getKeyVal, err := safejs.ValueOf("some id")
				if err != nil {
					return nil, err
				}
				return store.Get(getKeyVal)
			},
			expectResult: safejs.Safe(js.ValueOf("some value")),
		},
		{
			name: "get key",
			keys: map[string]interface{}{
				"some id": "some value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				getKeyVal, err := safejs.ValueOf("some id")
				if err != nil {
					return nil, err
				}
				return store.GetKey(getKeyVal)
			},
			expectResult: safejs.Safe(js.ValueOf("some id")),
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for key, value := range tc.keys {
				_, err := store.AddKey(js.ValueOf(key), js.ValueOf(value))
				assert.NoError(t, err)
			}
			req, err := tc.getFn(store)
			assert.NoError(t, err)
			ctx := context.Background()
			var result interface{}
			switch req := req.(type) {
			case *ArrayRequest:
				result, err = req.Await(ctx)
			case *Request:
				result, err = req.Await(ctx)
			default:
				t.Fatalf("Invalid return type: %T", req)
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, result)
		})
	}
}

func TestObjectStoreIndex(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", safejs.Safe(js.ValueOf("indexKey")), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	index, err := store.Index("myindex")
	assert.NoError(t, err)
	assert.NotZero(t, index)
}

func TestObjectStorePut(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("id"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	ctx := context.Background()
	req, err := store.Put(safejs.Safe(js.ValueOf(map[string]interface{}{
		"id":    "some id",
		"value": "some value",
	})))
	assert.NoError(t, err)
	resultKey, err := req.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf("some id")), resultKey)
}

func TestObjectStorePutKey(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	ctx := context.Background()
	req, err := store.PutKey(safejs.Safe(js.ValueOf("some id")), safejs.Safe(js.ValueOf("some value")))
	assert.NoError(t, err)
	resultKey, err := req.Await(ctx)
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf("some id")), resultKey)
}

func TestObjectStoreOpenCursor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		keys          map[string]interface{}
		cursorFn      func(*ObjectStore) (interface{}, error)
		expectResults []js.Value
	}{
		{
			name: "open cursor next",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursor(CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some value"),
				js.ValueOf("some other value"),
			},
		},
		{
			name: "open cursor previous",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursor(CursorPrevious)
			},
			expectResults: []js.Value{
				js.ValueOf("some other value"),
				js.ValueOf("some value"),
			},
		},
		{
			name: "open cursor over key",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursorKey(safejs.Safe(js.ValueOf("some id")), CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some value"),
			},
		},
		{
			name: "open cursor over key range",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeLowerBound(safejs.Safe(js.ValueOf("some more")), true)
				assert.NoError(t, err)
				return store.OpenCursorRange(keyRange, CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some other value"),
			},
		},
		{
			name: "open key cursor",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenKeyCursor(CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some id"),
				js.ValueOf("some other id"),
			},
		},
		{
			name: "open key cursor key",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenKeyCursorKey(safejs.Safe(js.ValueOf("some id")), CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some id"),
			},
		},
		{
			name: "open key cursor range",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeLowerBound(safejs.Safe(js.ValueOf("some more")), true)
				assert.NoError(t, err)
				return store.OpenKeyCursorRange(keyRange, CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some other id"),
			},
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for key, value := range tc.keys {
				_, err = store.AddKey(js.ValueOf(key), js.ValueOf(value))
				assert.NoError(t, err)
			}

			req, err := tc.cursorFn(store)
			assert.NoError(t, err)
			var results []js.Value
			ctx := context.Background()
			switch req := req.(type) {
			case *CursorWithValueRequest:
				err := req.Iter(ctx, func(cursor *CursorWithValue) error {
					value, err := cursor.Value()
					if assert.NoError(t, err) {
						results = append(results, safejs.Unsafe(value))
						err = cursor.Continue()
						return err
					}
					return err
				})
				assert.NoError(t, err)
			case *CursorRequest:
				err := req.Iter(ctx, func(cursor *Cursor) error {
					key, err := cursor.Key()
					if assert.NoError(t, err) {
						results = append(results, safejs.Unsafe(key))
						return cursor.Continue()
					}
					return err
				})
				assert.NoError(t, err)
			default:
				t.Fatalf("Invalid cursor type: %T", req)
			}
			assert.Equal(t, tc.expectResults, results)
		})
	}
}
