# json store for documents

Detailed design

A document store organizes a few blocks of memory to store JSON scalar values packed in memory.

The primary design objective is to minimize the memory impact of dealing with multiple JSON documents.

Speed comes as a secondary objective only.

Stores act as a companion object to [json.Document] to store JSON scalar values. A [Store] doesn't represent the structure of a JSON document, but
merely serves as a sink to store values.

It serves as a cache when loading several JSON documents.

There is no garbage collection inside a Store: its purpose is precisely to relieve the runtime from a lot of memory management on thousands of small objects.
When the caller is done with a Store, it may relinquish it.

Stores may be pooled to recycle memory.

Security advisory: even though a store supports concurrent access by multiple go routines, it may not be a good idea to share the data it contains across several http requests.

The Store pool optionally rewrites the entire arena with zeros when recycling a previously allocated pool.

It serves internal "slots" as References to a document to keep track of its values, in a more efficient way than a pointer.
All the smart lies in how the reference (e.g. a uint32 or uint64) is used to locate a value of a given type in the memory block. The A document store may be rather limited in size, e.g. 4 GB or so: we don't need an unlimited address space.
Values with a small size (e.g. null, true, false, small numbers, etc) may not actually need to consume a slot in the arena, as reference
may be large enough to encode the value.

NOTE: This supersedes the document cache currently managed in repo `github.com/go-openapi/spec`

Inspirations:
* https://github.com/rgcl/jsonpack
* https://sqlite.org/json1.html
* JSONB (postgresql)
* BSON (mongodb)

## Fact sheet

* Module: github.com/go-openapi/core/json
* Interface definition: github.com/go-openapi/core/interfaces
* Default implementation: github.com/go-openapi/core/json/stores/default-store
* Alternate implementations: github.com/go-openapi/core/json/stores/big-store  -- A store that supports 1 TB of values
* Contributed implementations: github.com/go-openapi/core/json/stores/contrib [go.mod] - Contributed alternative store implementations
* Examples: github.com/go-openapi/core/json/lexers/examples [go.mod]
* Poolable: Yes

Possible extensions:

The interface of a document store is simple, yet extensible. Essentially, this is a cache:

## Interfaces

Package `interfaces`:
```go
// DocumentSharable allows [stores.Store] objects to provide more services to be shared across [json.Document]s.
type DocumentSharable interface {
  // Loader function to grab JSON from a remote or local file location.
  Loader() func(string) ([]byte,error)
}

type InternedKey unique.Handle[string]

func (k InternedKey) String() string {
  return k.Handle.Value()
}
```

Package `stores`:
```go
type Store interface {
  Put(value Value) Reference
  Get(Reference) Value

  // Stuff sharable by different documents (TODO: optional)
  // interfaces.DocumentSharable

  interfaces.Resettable
}

// uint64 version (TODO uint32 version).
//
// Reference packs a JSON value as a uint64 plus a reference to []byte located in the arena.
//
//  |0-3     | 4-7  | 8-23   | 24-63                       |
//  | header | length        | offset: 0 - 2^40-1 = 1024 GB|
//  |        | small| small payload                        |
//
// Header: 4 bits - 15 available distinct behaviors.
//
// 0: null value
// 1: false value
// 2: true value
// 3: small number (=<7 bytes): only numbers longer than 7 require an arena slot
//  small size is encoded in bits 4-6 (0-7)
//  negative small numbers get bit 7 set
//  a small number with size 0 is 0 (-0 is equivalent to 0).
//  small numbers are BCD-encoded bytes in small payload section [8-63] (7 bytes -> may encode integers up to 14 digits and floats up to 13 significant digits, the leading 0 is stripped) 
// 4: small string (=<7 bytes): only strings larger than 7 digits require an arena slot
//  small size is encoded in bits 4-7 (used range: 0-7)
//  a small string with size 0 is "".
// 5: small ascii string (=<8 bytes): ascii strings that may be encoded on 7 bits only are compressed from up to 8 bytes
//  small size is same as above
//  the "" case is captured above
//  only captures the case of 8 bytes strings: smaller strings are captured by (4), and larger strings need an arena slot.
//  the payload section packs up to 8 7-bit characters into up to 7 bytes
// 6: large number: offset points to arena slot for BCD-encoded []byte
// 7: large string: offset points to arena slot for []byte (possibly compressed)
//
// length [4-23]: 0-2^20 = single value up to 1GB
// offset [24-63]: 0-2^40 = 1024 GB
//
// Encoding
// numbers: are encoded as BCD, cf. https://pkg.go.dev/github.com/yerden/go-util/bcd
type Reference uint64

func (r Reference) Value() types.Value {}
func makeReference(v Value) Reference {}

type ReferenceConstraint interface {
  ~uint32
  ~uint64
}

type Store[T ReferenceConstraint] struct {
  sync.RWLock
  next uint64 // the offset of the next readily available arena slot

  // The arena is used to capture large numbers and strings, which can't be covered by standalone [Reference]s.
  // for numbers, this corresponds to integers with more than 14 digits and floats with more than 13 significant digits.
  // for strings, this corresponds to strings larger than 8 bytes, non-ascii strings larger than 7 bytes.
  arena []byte // could be arena [][]byte for multi arenas Stores - not sure we want to get there
  // zlib dictionary
  dict []byte
}

// New store for JSON documents
func New(opts ...Option) *Store { }

type Option func(*options)

type options struct {
  arenaOptions
  documentSharableOptions
  enableStringCompression bool
  // internedValues map[unique.Handle[string]]struct{} // not sure this is a good idea
}

type documentSharableOptions struct {
  loaderFunc func(string) ([]byte,error)
}

type arenaOptions struct {
  maxArenaSize uint64
  preallocatedSize uint32
  maxArenasCount uint8 // this is when we want to daisy-chain multiple memory arenas (case of a reference of small size: so we have another indirection here...)
  zeroesOnReset bool // ensure that recycling a store doesn't leak any data
}

type CompressionLevel int

const (
 CompressionLevelOne CompressLevel = iota + 1
 CompressionLevelTwo
  ...
)

const (
  CompressionLevelFastest = CompressionLevelOne
  CompressionLevelHighest = CompressionLevelNine
)

// WithCompress enables compression of strings longer than threshold using zlib, with a compression level (from 1 to 9).
//
// By default, compression is disabled.
//
// Example:
// WithCompress(1024, zlib.BestSpeed) will compress all strings of 1024 or more bytes to be compressed internally in the Store.
//
// For best compression, the zlib dictionary is shared among all documents held by the Store.
//
// Note: at this moment, we don't compress numbers.
func WithCompress(threshold uint, level CompressionLeve) Option {}

func WithPreallocatedArena(size uint64) Option {}

// not sure this is a good idea. Documents already use interning for object keys.
// this could be valuable for JSON schemas, for which many values are repeated such as schema types, but since there are short enough
// to be covered as "short string" inlined references, I am not so sure we have a use case for value interning.
//
// WithValueInterning enables interning of some frequent values across documents in the Store.
//
// An interned value doesn't consume any slot in the arena.
//
// Interning is supported for strings and numbers. Boolean values don't need interning.
// The empty string or the 0 numerical value don't need interning.
//
// By default, no values are interned.
//func WithValueInterning(values ...json.ScalarValue) Option {}

// WithLoader specifies a document loader for the Store.
//
// The default is [loading.JSONDoc].
func WithLoader(func(url string) ([]byte, error)
```

Pooling:
```go
var pool pools.Pool[Store] = pools.New()

func BorrowStore() *Store { return pool.Borrow() }
func RedeemStore(s *Store) { pool.Redeem(s) }
```

It is possible to implement (possibly experiment in the contrib module) stores with different properties.

Examples:

* using different internal packing methods. Some could be inspired by other systems (e.g. JSON packing in sqllite, jsonpack, bson or some other representation that may be more suitable to convert into a specific backend)
* extending to storage-backed cache and relieving memory limitations
    ...
