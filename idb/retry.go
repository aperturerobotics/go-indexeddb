//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"strings"
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
	switch {
	case err == nil:
		return false
	case strings.HasSuffix(err.Error(), "The transaction has finished."):
		return true
	case strings.HasSuffix(err.Error(), "The database connection is closing."):
		return true
	default:
		return false
	}
}
