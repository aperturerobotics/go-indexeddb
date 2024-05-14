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

[idb-pkg]: https://pkg.go.dev/github.com/hack-pad/go-indexeddb/idb

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

## Upstream

This package is a fork of [github.com/hack-pad/go-indexeddb](https://github.com/hack-pad/go-indexeddb).

## License

MIT
