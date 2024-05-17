package idb // import "github.com/aperturerobotics/go-indexeddb/idb"

Package idb is a low-level driver that provides type-safe bindings to IndexedDB
in Wasm programs. The primary focus is to align with the IndexedDB spec,
followed by ease of use.

To get started, get the global indexedDB instance with idb.Global(). See below
for examples.

VARIABLES

var (
	// ErrCursorStopIter stops iteration when returned from a CursorRequest.Iter() handler
	ErrCursorStopIter = errors.New("stop cursor iteration")
)

FUNCTIONS

func IsTxnFinishedErr(err error) bool
    IsTxnFinishedErr checks if an error corresponds to a transaction finishing.
    see RetryTxn for details

func RetryTxn(
	ctx context.Context,
	db *Database,
	txnMode TransactionMode,
	fn func(txn *Transaction) error,
	objectStoreName string,
	objectStoreNames ...string,
) error
    *
      - RetryTxn retries the function with a new transaction if the txn finishes
        prematurely. *
      - IndexedDB transactions automatically commit when all outstanding
        requests
      - have been satisfied. Go WebAssembly will background a Goroutine leading
        to
      - IndexedDB transactions completing earlier than expected, leading to
        errors
      - with a suffix "The transaction has finished." when further operations
        are
      - attempted on the completed transaction. *
      - See: https://github.com/w3c/IndexedDB/issues/34 for more details. *
      - RetryTxn is a mechanism that automatically re-creates the transaction
        and
      - retries the operation whenever we encounter this specific error. This
      - ensures that operations can continue even if the transaction has been
      - automatically committed. We will wrap all transaction operations within
        a
      - retry logic that detects the "The transaction has finished." error and
      - retries the operation with a new transaction.


TYPES

type AckRequest struct {
	*Request
}
    AckRequest is a Request that doesn't retrieve a value, only used to detect
    errors.

func (a *AckRequest) Await(ctx context.Context) error
    Await waits for success or failure, then returns the results.

func (a *AckRequest) Result()
    Result is a no-op. This kind of request does not retrieve any data in the
    result.

type ArrayRequest struct {
	*Request
}
    ArrayRequest is a Request that retrieves an array of js.Values

func (a *ArrayRequest) Await(ctx context.Context) ([]js.Value, error)
    Await waits for success or failure, then returns the results.

func (a *ArrayRequest) Result() ([]js.Value, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

type Cursor struct {
	// Has unexported fields.
}
    Cursor represents a cursor for traversing or iterating over multiple records
    in a Database

func (c *Cursor) Advance(count uint) error
    Advance sets the number of times a cursor should move its position forward.

func (c *Cursor) Continue() error
    Continue advances the cursor to the next position along its direction.

func (c *Cursor) ContinueKey(key safejs.Value) error
    ContinueKey advances the cursor to the next position along its direction.

func (c *Cursor) ContinuePrimaryKey(key, primaryKey safejs.Value) error
    ContinuePrimaryKey sets the cursor to the given index key and primary key
    given as arguments. Returns an error if the source is not an index.

func (c *Cursor) Delete() (*AckRequest, error)
    Delete returns an AckRequest, and, in a separate thread, deletes the record
    at the cursor's position, without changing the cursor's position. This can
    be used to delete specific records.

func (c *Cursor) Direction() (CursorDirection, error)
    Direction returns the direction of traversal of the cursor

func (c *Cursor) Key() (safejs.Value, error)
    Key returns the key for the record at the cursor's position. If the cursor
    is outside its range, this is set to undefined.

func (c *Cursor) PrimaryKey() (safejs.Value, error)
    PrimaryKey returns the cursor's current effective primary key. If the cursor
    is currently being iterated or has iterated outside its range, this is set
    to undefined.

func (c *Cursor) Request() (*Request, error)
    Request returns the Request that was used to obtain the cursor.

func (c *Cursor) Source() (objectStore *ObjectStore, index *Index, err error)
    Source returns the ObjectStore or Index that the cursor is iterating

func (c *Cursor) Update(value safejs.Value) (*Request, error)
    Update returns a Request, and, in a separate thread, updates the value at
    the current position of the cursor in the object store. This can be used to
    update specific records.

type CursorDirection int
    CursorDirection is the direction of traversal of the cursor

const (
	// CursorNext direction causes the cursor to be opened at the start of the source.
	CursorNext CursorDirection = iota
	// CursorNextUnique direction causes the cursor to be opened at the start of the source. For every key with duplicate values, only the first record is yielded.
	CursorNextUnique
	// CursorPrevious direction causes the cursor to be opened at the end of the source.
	CursorPrevious
	// CursorPreviousUnique direction causes the cursor to be opened at the end of the source. For every key with duplicate values, only the first record is yielded.
	CursorPreviousUnique
)
func (d CursorDirection) String() string

type CursorRequest struct {
	*Request
}
    CursorRequest is a Request that retrieves a Cursor

func (c *CursorRequest) Await(ctx context.Context) (*Cursor, error)
    Await waits for success or failure, then returns the results.

func (c *CursorRequest) Iter(ctx context.Context, iter func(*Cursor) error) error
    Iter invokes the callback when the request succeeds for each cursor
    iteration

func (c *CursorRequest) Result() (*Cursor, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

type CursorWithValue struct {
	*Cursor
}
    CursorWithValue represents a cursor for traversing or iterating over
    multiple records in a database. It is the same as the Cursor, except that it
    includes the value property.

func (c *CursorWithValue) Value() (safejs.Value, error)
    Value returns the value of the current cursor

type CursorWithValueRequest struct {
	*Request
}
    CursorWithValueRequest is a Request that retrieves a CursorWithValue

func (c *CursorWithValueRequest) Await(ctx context.Context) (*CursorWithValue, error)
    Await waits for success or failure, then returns the results.

func (c *CursorWithValueRequest) Iter(ctx context.Context, iter func(*CursorWithValue) error) error
    Iter invokes the callback when the request succeeds for each cursor
    iteration

func (c *CursorWithValueRequest) Result() (*CursorWithValue, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

type DOMException struct {
	// Has unexported fields.
}
    DOMException is a JavaScript DOMException with a standard name. Use
    errors.Is() to compare by name.

func NewDOMException(name string) DOMException
    NewDOMException returns a new DOMException with the given name. Only useful
    for errors.Is() comparisons with errors returned from idb.

func (e DOMException) Error() string

func (e DOMException) Is(target error) bool
    Is returns true target is a DOMException and matches this DOMException's
    name. Use 'errors.Is()' to call it.

type Database struct {
	// Has unexported fields.
}
    Database provides a connection to a database. You can use a Database object
    to open a transaction on your database then create, manipulate, and delete
    objects (data) in that database.

func (db *Database) Close() error
    Close closes the connection to a database.

func (db *Database) CreateObjectStore(name string, options ObjectStoreOptions) (*ObjectStore, error)
    CreateObjectStore creates and returns a new object store or index.

func (db *Database) DeleteObjectStore(name string) error
    DeleteObjectStore destroys the object store with the given name in the
    connected database, along with any indexes that reference it.

func (db *Database) Name() (string, error)
    Name returns the name of the connected database.

func (db *Database) ObjectStoreNames() ([]string, error)
    ObjectStoreNames returns a list of the names of the object stores currently
    in the connected database.

func (db *Database) Transaction(mode TransactionMode, objectStoreName string, objectStoreNames ...string) (_ *Transaction, err error)
    Transaction returns a transaction object containing the
    Transaction.ObjectStore() method, which you can use to access your object
    store.

func (db *Database) TransactionWithOptions(options TransactionOptions, objectStoreName string, objectStoreNames ...string) (*Transaction, error)
    TransactionWithOptions returns a transaction object containing the
    Transaction.ObjectStore() method, which you can use to access your object
    store.

func (db *Database) Version() (uint, error)
    Version returns the version of the connected database.

type Factory struct {
	// Has unexported fields.
}
    Factory lets applications asynchronously access the indexed databases.
    A typical program will call Global() to access window.indexedDB.

func Global() *Factory
    Global returns the global IndexedDB instance. Can be called multiple times,
    will always return the same result (or error if one occurs).

func WrapFactory(jsFactory js.Value) (*Factory, error)
    WrapFactory wraps the given IDBFactory object

func (f *Factory) CompareKeys(a, b js.Value) (int, error)
    CompareKeys compares two keys and returns a result indicating which one is
    greater in value.

func (f *Factory) DeleteDatabase(name string) (*AckRequest, error)
    DeleteDatabase requests the deletion of a database.

func (f *Factory) Open(upgradeCtx context.Context, name string, version uint, upgrader Upgrader) (*OpenDBRequest, error)
    Open requests to open a connection to a database.

type Index struct {
	// Has unexported fields.
}
    Index provides asynchronous access to an index in a database. An index
    is a kind of object store for looking up records in another object store,
    called the referenced object store. You use this to retrieve data.

func (i *Index) Count() (*UintRequest, error)
    Count returns a UintRequest, and, in a separate thread, returns the total
    number of records in the index.

func (i *Index) CountKey(key js.Value) (*UintRequest, error)
    CountKey returns a UintRequest, and, in a separate thread, returns the total
    number of records that match the provided key.

func (i *Index) CountRange(keyRange *KeyRange) (*UintRequest, error)
    CountRange returns a UintRequest, and, in a separate thread, returns the
    total number of records that match the provided KeyRange.

func (i *Index) Get(key js.Value) (*Request, error)
    Get returns a Request, and, in a separate thread, returns objects selected
    by the specified key. This is for retrieving specific records from an index.

func (i *Index) GetAllKeys() (*ArrayRequest, error)
    GetAllKeys returns an ArrayRequest that retrieves record keys for all
    objects in the index.

func (i *Index) GetAllKeysRange(query *KeyRange, maxCount uint) (*ArrayRequest, error)
    GetAllKeysRange returns an ArrayRequest that retrieves record keys for
    all objects in the index matching the specified query. If maxCount is 0,
    retrieves all objects matching the query.

func (i *Index) GetKey(value js.Value) (*Request, error)
    GetKey returns a Request, and, in a separate thread retrieves and returns
    the record key for the object matching the specified parameter.

func (i *Index) KeyPath() (js.Value, error)
    KeyPath returns the key path of this index. If js.Null(), this index is not
    auto-populated.

func (i *Index) MultiEntry() (bool, error)
    MultiEntry affects how the index behaves when the result of evaluating the
    index's key path yields an array. If true, there is one record in the index
    for each item in an array of keys. If false, then there is one record for
    each key that is an array.

func (i *Index) Name() (string, error)
    Name returns the name of this index

func (i *Index) ObjectStore() (*ObjectStore, error)
    ObjectStore returns the object store referenced by this index.

func (i *Index) OpenCursor(direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursor returns a CursorWithValueRequest, and, in a separate thread,
    returns a new CursorWithValue. Used for iterating through an index by
    primary key with a cursor.

func (i *Index) OpenCursorKey(key js.Value, direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursorKey is the same as OpenCursor, but opens a cursor over the given
    key instead.

func (i *Index) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursorRange is the same as OpenCursor, but opens a cursor over the given
    range instead.

func (i *Index) OpenKeyCursor(direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a
    new Cursor. Used for iterating through all keys in an object store.

func (i *Index) OpenKeyCursorKey(key js.Value, direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the
    given key instead.

func (i *Index) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the
    given key range instead.

func (i *Index) Unique() (bool, error)
    Unique indicates this index does not allow duplicate values for a key.

type IndexOptions struct {
	// Unique disallows duplicate values for a single key.
	Unique bool
	// MultiEntry adds an entry in the index for each array element when the keyPath resolves to an Array. If false, adds one single entry containing the Array.
	MultiEntry bool
}
    IndexOptions contains all options used to create an Index

type KeyRange struct {
	// Has unexported fields.
}
    KeyRange represents a continuous interval over some data type that is used
    for keys. Records can be retrieved from ObjectStore and Index objects using
    keys or a range of keys.

func NewKeyRangeBound(lower, upper safejs.Value, lowerOpen, upperOpen bool) (*KeyRange, error)
    NewKeyRangeBound creates a new key range with the specified upper and lower
    bounds. The bounds can be open (that is, the bounds exclude the endpoint
    values) or closed (that is, the bounds include the endpoint values).

func NewKeyRangeLowerBound(lower safejs.Value, open bool) (*KeyRange, error)
    NewKeyRangeLowerBound creates a new key range with only a lower bound.

func NewKeyRangeOnly(only safejs.Value) (*KeyRange, error)
    NewKeyRangeOnly creates a new key range containing a single value.

func NewKeyRangeUpperBound(upper safejs.Value, open bool) (*KeyRange, error)
    NewKeyRangeUpperBound creates a new key range with only an upper bound.

func (k *KeyRange) Includes(key safejs.Value) (bool, error)
    Includes returns a boolean indicating whether a specified key is inside the
    key range.

func (k *KeyRange) Lower() (safejs.Value, error)
    Lower returns the lower bound of the key range.

func (k *KeyRange) LowerOpen() (bool, error)
    LowerOpen returns false if the lower-bound value is included in the key
    range.

func (k *KeyRange) Upper() (safejs.Value, error)
    Upper returns the upper bound of the key range.

func (k *KeyRange) UpperOpen() (bool, error)
    UpperOpen returns false if the upper-bound value is included in the key
    range.

type ObjectStore struct {
	// Has unexported fields.
}
    ObjectStore represents an object store in a database. Records within an
    object store are sorted according to their keys. This sorting enables fast
    insertion, look-up, and ordered retrieval.

func (o *ObjectStore) Add(value js.Value) (*AckRequest, error)
    Add returns an AckRequest, and, in a separate thread, creates a structured
    clone of the value, and stores the cloned value in the object store. This is
    for adding new records to an object store.

func (o *ObjectStore) AddKey(key, value js.Value) (*AckRequest, error)
    AddKey is the same as Add, but includes the key to use to identify the
    record.

func (o *ObjectStore) AutoIncrement() (bool, error)
    AutoIncrement returns the value of the auto increment flag for this object
    store.

func (o *ObjectStore) Clear() (*AckRequest, error)
    Clear returns an AckRequest, then clears this object store in a separate
    thread. This is for deleting all current records out of an object store.

func (o *ObjectStore) Count() (*UintRequest, error)
    Count returns a UintRequest, and, in a separate thread, returns the total
    number of records in the store.

func (o *ObjectStore) CountKey(key safejs.Value) (*UintRequest, error)
    CountKey returns a UintRequest, and, in a separate thread, returns the total
    number of records that match the provided key.

func (o *ObjectStore) CountRange(keyRange *KeyRange) (*UintRequest, error)
    CountRange returns a UintRequest, and, in a separate thread, returns the
    total number of records that match the provided KeyRange.

func (o *ObjectStore) CreateIndex(name string, keyPath safejs.Value, options IndexOptions) (*Index, error)
    CreateIndex creates a new index during a version upgrade, returning a new
    Index object in the connected database.

func (o *ObjectStore) Delete(key js.Value) (*AckRequest, error)
    Delete returns an AckRequest, and, in a separate thread, deletes the store
    object selected by the specified key. This is for deleting individual
    records out of an object store.

func (o *ObjectStore) DeleteIndex(name string) error
    DeleteIndex destroys the specified index in the connected database, used
    during a version upgrade.

func (o *ObjectStore) Get(key safejs.Value) (*Request, error)
    Get returns a Request, and, in a separate thread, returns the objects
    selected by the specified key. This is for retrieving specific records from
    an object store.

func (o *ObjectStore) GetAllKeys() (*ArrayRequest, error)
    GetAllKeys returns an ArrayRequest that retrieves record keys for all
    objects in the object store.

func (o *ObjectStore) GetAllKeysRange(query *KeyRange, maxCount uint) (*ArrayRequest, error)
    GetAllKeysRange returns an ArrayRequest that retrieves record keys for all
    objects in the object store matching the specified query. If maxCount is 0,
    retrieves all objects matching the query.

func (o *ObjectStore) GetKey(value safejs.Value) (*Request, error)
    GetKey returns a Request, and, in a separate thread retrieves and returns
    the record key for the object matching the specified parameter.

func (o *ObjectStore) Index(name string) (*Index, error)
    Index opens an index from this object store after which it can, for example,
    be used to return a sequence of records sorted by that index using a cursor.

func (o *ObjectStore) IndexNames() ([]string, error)
    IndexNames returns a list of the names of indexes on objects in this object
    store.

func (o *ObjectStore) KeyPath() (safejs.Value, error)
    KeyPath returns the key path of this object store. If this returns
    js.Null(), the application must provide a key for each modification
    operation.

func (o *ObjectStore) Name() (string, error)
    Name returns the name of this object store.

func (o *ObjectStore) OpenCursor(direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursor returns a CursorWithValueRequest, and, in a separate thread,
    returns a new CursorWithValue. Used for iterating through an object store by
    primary key with a cursor.

func (o *ObjectStore) OpenCursorKey(key safejs.Value, direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursorKey is the same as OpenCursor, but opens a cursor over the given
    key instead.

func (o *ObjectStore) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorWithValueRequest, error)
    OpenCursorRange is the same as OpenCursor, but opens a cursor over the given
    range instead.

func (o *ObjectStore) OpenKeyCursor(direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a
    new Cursor. Used for iterating through all keys in an object store.

func (o *ObjectStore) OpenKeyCursorKey(key safejs.Value, direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the
    given key instead.

func (o *ObjectStore) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorRequest, error)
    OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the
    given key range instead.

func (o *ObjectStore) Put(value safejs.Value) (*Request, error)
    Put returns a Request, and, in a separate thread, creates a structured clone
    of the value, and stores the cloned value in the object store. This is for
    updating existing records in an object store when the transaction's mode is
    readwrite.

func (o *ObjectStore) PutKey(key, value safejs.Value) (*Request, error)
    PutKey is the same as Put, but includes the key to use to identify the
    record.

func (o *ObjectStore) Transaction() (*Transaction, error)
    Transaction returns the Transaction object to which this object store
    belongs.

type ObjectStoreOptions struct {
	KeyPath       js.Value
	AutoIncrement bool
}
    ObjectStoreOptions contains all available options for creating an
    ObjectStore

type OpenDBRequest struct {
	*Request
}
    OpenDBRequest provides access to the results of requests to open or delete
    databases (performed using Factory.open and Factory.DeleteDatabase).

func (o *OpenDBRequest) Await(ctx context.Context) (*Database, error)
    Await waits for success or failure, then returns the results.

func (o *OpenDBRequest) Result() (*Database, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

type Request struct {
	// Has unexported fields.
}
    Request provides access to results of asynchronous requests to databases and
    database objects using event listeners. Each reading and writing operation
    on a database is done using a request.

func (r *Request) Await(ctx context.Context) (safejs.Value, error)
    Await waits for success or failure, then returns the results.

func (r *Request) AwaitCursor(ctx context.Context) (*Cursor, error)
    AwaitCursor awaits the iterator cursor and returns the value.

    returns nil if there are no more results.

func (r *Request) Err() (err error)
    Err returns an error in the event of an unsuccessful request, indicating
    what went wrong.

func (r *Request) Listen(ctx context.Context, success, failed func()) error
    Listen invokes the success callback when the request succeeds and failed
    when it fails.

func (r *Request) ListenError(ctx context.Context, failed func()) error
    ListenError invokes the callback when the request fails

func (r *Request) ListenSuccess(ctx context.Context, success func()) error
    ListenSuccess invokes the callback when the request succeeds

func (r *Request) ReadyState() (string, error)
    ReadyState returns the state of the request. Every request starts in
    the pending state. The state changes to done when the request completes
    successfully or when an error occurs.

func (r *Request) Result() (safejs.Value, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

func (r *Request) Source() (objectStore *ObjectStore, index *Index, err error)
    Source returns the source of the request, such as an Index or an
    ObjectStore. If no source exists (such as when calling Factory.Open),
    it returns nil for both.

func (r *Request) Transaction() (*Transaction, error)
    Transaction returns the transaction for the request. This can return nil
    for certain requests, for example those returned from Factory.Open unless
    an upgrade is needed. (You're just connecting to a database, so there is no
    transaction to return).

type Transaction struct {
	// Has unexported fields.
}
    Transaction provides a static, asynchronous transaction on a database. All
    reading and writing of data is done within transactions. You use Database
    to start transactions, Transaction to set the mode of the transaction (e.g.
    is it TransactionReadOnly or TransactionReadWrite), and you access an
    ObjectStore to make a request. You can also use a Transaction object to
    abort transactions.

func (t *Transaction) Abort() error
    Abort rolls back all the changes to objects in the database associated with
    this transaction.

func (t *Transaction) Await(ctx context.Context) error
    Await waits for success or failure, then returns the results.

func (t *Transaction) Commit() error
    Commit for an active transaction, commits the transaction. Note that this
    doesn't normally have to be called — a transaction will automatically commit
    when all outstanding requests have been satisfied and no new requests have
    been made. Commit() can be used to start the commit process without waiting
    for events from outstanding requests to be dispatched.

func (t *Transaction) Database() (*Database, error)
    Database returns the database connection with which this transaction is
    associated.

func (t *Transaction) Durability() (TransactionDurability, error)
    Durability returns the durability hint the transaction was created with.

func (t *Transaction) Err() error
    Err returns an error indicating the type of error that occurred when
    there is an unsuccessful transaction. Returns nil if the transaction is
    not finished, is finished and successfully committed, or was aborted with
    Transaction.Abort().

func (t *Transaction) Mode() (TransactionMode, error)
    Mode returns the mode for isolating access to data in the object
    stores that are in the scope of the transaction. The default value is
    TransactionReadOnly.

func (t *Transaction) ObjectStore(name string) (*ObjectStore, error)
    ObjectStore returns an ObjectStore representing an object store that is part
    of the scope of this transaction.

func (t *Transaction) ObjectStoreNames() ([]string, error)
    ObjectStoreNames returns a list of the names of ObjectStores associated with
    the transaction.

type TransactionDurability int
    TransactionDurability is a hint to the user agent of whether to prioritize
    performance or durability when committing a transaction.

const (
	// DurabilityDefault indicates the user agent should use its default durability behavior for the storage bucket. This is the default for transactions if not otherwise specified.
	DurabilityDefault TransactionDurability = iota
	// DurabilityRelaxed indicates the user agent may consider that the transaction has successfully committed as soon as all outstanding changes have been written to the operating system, without subsequent verification.
	DurabilityRelaxed
	// DurabilityStrict indicates the user agent may consider that the transaction has successfully committed only after verifying all outstanding changes have been successfully written to a persistent storage medium.
	DurabilityStrict
)
func (d TransactionDurability) String() string

type TransactionMode int
    TransactionMode defines the mode for isolating access to data in the
    transaction's current object stores.

const (
	// TransactionReadOnly allows data to be read but not changed.
	TransactionReadOnly TransactionMode = iota
	// TransactionReadWrite allows reading and writing of data in existing data stores to be changed.
	TransactionReadWrite
)
func (m TransactionMode) String() string

type TransactionOptions struct {
	Mode       TransactionMode
	Durability TransactionDurability
}
    TransactionOptions contains all available options for creating and starting
    a Transaction

type UintRequest struct {
	*Request
}
    UintRequest is a Request that retrieves a uint result

func (u *UintRequest) Await(ctx context.Context) (uint, error)
    Await waits for success or failure, then returns the results.

func (u *UintRequest) Result() (uint, error)
    Result returns the result of the request. If the request failed and the
    result is not available, an error is returned.

type Upgrader func(db *Database, oldVersion, newVersion uint) error
    Upgrader is a function that can upgrade the given database from an old
    version to a new one.

