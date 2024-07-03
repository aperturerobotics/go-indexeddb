# go-indexeddb

[![Go Reference][reference-badge]][reference]
[![CI][ci-badge]][ci-workflow]

[reference-badge]: https://pkg.go.dev/badge/github.com/aperturerobotics/go-indexeddb/idb.svg
[ci-badge]: https://github.com/aperturerobotics/go-indexeddb/actions/workflows/tests.yml/badge.svg
[reference]: https://pkg.go.dev/github.com/aperturerobotics/go-indexeddb/idb
[ci-workflow]: https://github.com/aperturerobotics/go-indexeddb/actions/workflows/tests.yml

**go-indexeddb** is a low-level Go driver that provides type-safe bindings to IndexedDB in Wasm programs. The primary focus is to align with the IndexedDB spec, followed by ease of use.

[IndexedDB][] is a transactional database system, like an SQL-based RDBMS. However, unlike SQL-based RDBMSes, which use fixed-column tables, IndexedDB is an object-oriented database. IndexedDB lets you store and retrieve objects that are indexed with a key; any objects supported by the structured clone algorithm can be stored.

[IndexedDB]: https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API

See the [reference][] for full documentation and examples.


This package is available at **github.com/aperturerobotics/go-indexeddb**.

## Package index

Summary of the packages provided by this module:

- [`idb`][idb-pkg]: Package `idb` provides a low-level Go driver with type-safe bindings to IndexedDB in Wasm programs.
- [`durable`][durable-pkg]: Package `durable` provides a workaround for [transactions expiring].

[idb-pkg]: https://pkg.go.dev/github.com/aperturerobotics/go-indexeddb/idb?GOOS=js
[durable-pkg]: https://pkg.go.dev/github.com/aperturerobotics/go-indexeddb/durable?GOOS=js
[transactions expiring]: #Transactions-Expiring

## Usage

1. Get the package:

   ```bash
   go get github.com/aperturerobotics/go-indexeddb@latest
   ```

2. Import it in your code:

   ```go
   import "github.com/aperturerobotics/go-indexeddb/idb"
   ```

3. To get started, get the global indexedDB instance:

   ```go
   db, err := idb.Global().Open(ctx, "MyDatabase", nil)
   ```

Check out the [reference][] for more details and examples!

## Testing

This package can be tested in a browser environment using [`wasmbrowsertest`](https://github.com/agnivade/wasmbrowsertest).

1. Install `wasmbrowsertest`:
   ```bash
   go install github.com/agnivade/wasmbrowsertest@latest
   ```

2. Rename the `wasmbrowsertest` binary to `go_js_wasm_exec`:
   ```bash
   mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec
   ```

3. Run the tests with the `js` GOOS and `wasm` GOARCH:
   ```bash
   GOOS=js GOARCH=wasm go test -v ./...
   ```

This will compile the tests to WebAssembly and run them in a headless browser environment.

## Transactions Expiring

IndexedDB transactions automatically commit when all outstanding requests have
been satisfied. When a Goroutine is suspended due to a select statement or other
context switching, the IndexedDB transation commits automatically, leading to
errors with a suffix "The transaction has finished."

`RetryTxn` automatically re-creates the transaction and retries the operation
whenever we encounter this specific error. This ensures that operations can
continue even if the transaction has been automatically committed.

When a transaction becomes inactive it will also commit the changes made up to
that point. Calling the "abort" method will attempt to "roll back" the changes
made by the transaction. However, this is a relatively weak transaction
mechanism and should not be relied upon in the same way as traditional
transaction systems (such as those in BoltDB or similar databases).

Reference:

- https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API/Using_IndexedDB
- https://github.com/w3c/IndexedDB/issues/34


## Upstream

This package is a fork of [github.com/hack-pad/go-indexeddb](https://github.com/hack-pad/go-indexeddb).

## License

MIT
