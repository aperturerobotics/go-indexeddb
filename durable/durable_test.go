//go:build js && wasm
// +build js,wasm

package durable

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/aperturerobotics/go-indexeddb/idb"
	"github.com/hack-pad/safejs"
)

func TestDurableTransaction(t *testing.T) {
	ctx := context.Background()

	dbReq, err := idb.Global().Open(ctx, "test_db", 1, func(db *idb.Database, oldVersion, newVersion uint) error {
		store, err := db.CreateObjectStore("test_store", idb.ObjectStoreOptions{})
		if err != nil {
			return err
		}
		_, err = store.CreateIndex("test_index", safejs.Safe(js.ValueOf("key")), idb.IndexOptions{})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	db, err := dbReq.Await(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new durable transaction
	dt, err := NewDurableTransaction(db, idb.TransactionReadWrite, "test_store")
	if err != nil {
		t.Fatal(err)
	}

	// Get the object store
	store, err := dt.GetObjectStore("test_store")
	if err != nil {
		t.Fatal(err)
	}

	// Add an item
	key := safejs.Safe(js.ValueOf("key"))
	item := safejs.Safe(js.ValueOf("foo"))
	if err := store.AddKey(ctx, key, item); err != nil {
		t.Fatal(err)
	}

	// Get the item
	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	want := item
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// Update the item
	item = safejs.Safe(js.ValueOf("baz"))
	if err := store.PutKey(ctx, key, item); err != nil {
		t.Fatal(err)
	}

	// Get the updated item
	got, err = store.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	want = item
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// Delete the item
	if err := store.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}

	// Verify the item was deleted
	delVal, err := store.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if !delVal.IsUndefined() {
		t.Errorf("expected undefined, got %v", delVal.Type().String())
	}
}
