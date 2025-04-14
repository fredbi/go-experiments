# JSON document

Detailed design

A hierarchy of nodes that represent JSON elements organized into arrays and objects in a memory-efficient way.
All these are represented by go slices, not maps: ordering is maintained, and shallow cloning of slice value doesn't require allocating new memory.
All scalar content is stored in the document store (see below). Memory storage is shared among documents with similar content. In particular, by default, the document will intern strings for keys.
A document is immutable. Since cloning a document is essentially a copy-on-write operation, it is mandatory to mutate stuff with clones. A documen comes with a "builder" interface to mutate stuff and return a clone, e.g. add/remove things.

A document knows how to decode/encode and unmarshal/marshal JSON (essentially the same API as encoding/json).

A document can be walked over and navigated through:

* object keys and array elements can be iterated over
* json pointers can be resolved (json schema $ref semantics are not known at this level) efficiently (e.g. constant time or o(log n))
* desirable but not required: should lookup a key efficiently (e.g. constant time or o(log n))

> NOTE: this supersedes features previously provided by jsonpointer, swag

A JSON document uses string interning for keys (is that something that can be disabled?)

Options:

* a document may keep the parsing context of its nodes, to report higher-level errors
* a document may support various tuning options regarding how best to store things (e.g. reduce numbers as soon as possible, compress large strings, etc)
* ...

### How about YAML documents?

We keep the current concept that any YAML should be translated to JSON before being processed.

What about "verbatim" then? YAML is just too complex for me to drift into such details right now.

### Possible extensions

    the basic requirement on a JSON document is to support MarshalJSON and UnmarshalJSON, provide a builder side-type to support mutation
    more advanced use-case may be supported by additional (possibly composed) types
    examples:
        JSON document with JSONPath query support
        JSON document with other Marshal or Unmarshal targets, such as MarshalBSON (to store in MongoDB), MarshalJSONB (to store directly into postgres), ...

To avoid undue propagation of dependencies to external stuff like DB drivers etc, these should come as an independant module.

## Design

A `Document` holds a root `Node` and knows which `Store` backs its values.

`Nodes` describe the hierarchy of the JSON document.

We need a `Context` to (optionally) keep the full parsing context and point accurately to errors or warnings.

JSON document node

```go
type Document struct {
  store interfaces.Store
  root  Node
  err   error
}
```

```go
type Context struct {...} // text-context to report errors etc

func makeDocument(store Store, node Node) Document

func (d *Document) Kind() types.NodeKind { return d.root.Kind() }

func (d *Document) MarshalJSON() ([]byte, error) {}
func (d *Document) UnmarshalJSON([]byte) error {}
func (d *Document) Decode(r io.Reader) error {}
func (d *Document) Encode(w io.Writer) error {}

// Key returns the document located under key k, or false if this key is not present.
func (d *Document) Key(k string) (Document, bool)
// Elem returns the i-th element in an array, or false if i is equal or larger than the size of the array.
func (d *Document) Elem(i int) (Document, bool)
// KeysIterator returns an iterator over keys and documents in an object document
func (d *Document) KeysIterator() iter.Seq2[string,Document]
// ElemsIterator returns an iterator over the elements of an array document
func (d *Document) ElemsIterator() iter.Seq[Document]

// Get a JSON [Pointer] inside this [Document], or false if the pointer cannot be resolved
func (d Document) Get(p Pointer) (Document, bool)

func (d *Document) Reset() {}
func (d Document) Ok() bool { return d.err == nil }
func (d Document) Err() error { return d.err }
 
// Pointer implements a JSON pointer
type Pointer struct {
  path []string // TODO: interned?
  Resettable
}

func (p Pointer) Path() string { return strings.Join(p.path,"/") }

// NodeKind describes the kind of node in a JSON document
type NodeKind uint8
const (
  NodeKindNull NodeKind = iota
  NodeKindScalar
  NodeKindObject
  NodeKindArray
)

// ValueKind describe the kind of JSON value held by a node of kind NodeKindScalar
type ValueKind uint8
const (
  ValueKindNull ValueKind = iota
  ValueKindString
  ValueKindNumber
  ValueKindInteger
  ValueKindBool
}

// Node describes a node in the hierarchical structure of a JSON document.
// It can be any valid JSON value or construct.
type Node struct {
  kind NodeKind
  value Value // ValueKind for objects and arrays is NullValue
  children []Node // objects and array have children, nil for scalar nodes
  context Context // the error context (maybe use pointer here?) 
  keysIndex map[string]int // lookup index for objects, so Key() finds a key in constant time
}


func (n Node) Key(k string) (Document, bool) // always false if kind != NodeKindObject
func (n Node) Elem(i int) (Document, bool) // always false if kind != NodeKindArray
func (n Node) KeysIterator() iter.Seq2[string,Document] // nil if kind != NodeKindObject
func (n Node) ElemsIterator() iter.Seq[Document] // nil if kind != NodeKindArray
func (n Node) Value() (Value, bool) {} // always false if kind != NodeKindScalar

type Value struct {
  kind ValueKind

  // ScalarValue
  s StringValue
  n NumberValue  // TODO: add IntegerValue 
  b BoolValue
}

func (v BoolValue) Bool() (bool,bool) {}
func (v StringValue) String() (string,bool) {}

func (v NumberValue) Number() (string,bool) {}
func (v NumberValue) IsInteger() bool {}
func (v NumberValue) IsNegative() bool {}
func (v NumberValue) Int64() (int64,bool) {}
func (v NumberValue) Uint64() (uint64,bool) {}
func (v NumberValue) Float64() (float64,bool) {}
func (v NumberValue) BigInt() (big.Int,bool) {}
func (v NumberValue) BigRat() big.Rat {} 

type StringValue []byte
type NumberValue []byte
type BoolValue bool
```

Builder is the way to construct `Document`s programmatically or modify existing documents into new instances (a document is immutable).

```go
type Builder struct {
  Document
}

func (b Builder) WithObject() Builder {}
func (b Builder) AddElement() Builder {}
func (b Builder) AddKey(string, Node) Builder {}
func (b Builder) Document() (Document, error) {}
...
```

## TODO

There is a contrib module to absorb novel ideas and experimentations without breaking anything.
