//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"errors"
)

/*
RetryTxn retries the function with a new transaction if the txn finishes prematurely.

IndexedDB transactions automatically commit when all outstanding requests have
been satisfied. When a goroutine yields to the JavaScript event loop, for
example in a select statement, the transaction commits automatically and the
next request fails with a TransactionInactiveError.

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
		if IsTxnEndedErr(err) {
			// txn committed automatically already
			err = nil
		}

		return err
	}
}

var (
	errTransactionInactive = NewDOMException("TransactionInactiveError")
	errInvalidState        = NewDOMException("InvalidStateError")
)

// IsTxnFinishedErr reports whether err is a request placed against a
// transaction that is no longer active, such as one that committed
// automatically. See RetryTxn for details.
func IsTxnFinishedErr(err error) bool {
	return errors.Is(err, errTransactionInactive)
}

// IsTxnEndedErr reports whether err is a Commit or Abort call rejected because
// the transaction already committed or aborted.
func IsTxnEndedErr(err error) bool {
	return errors.Is(err, errInvalidState)
}
