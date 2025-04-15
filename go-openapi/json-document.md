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

### Other possible extensions

The basic requirement on a JSON document is to support MarshalJSON and UnmarshalJSON, JSON pointer and to provide a builder side-type to support mutation.
    
More advanced use cases may be supported by additional (possibly composed) types.
    
Examples:
* JSON document with JSONPath query support - I am not sure about why I would need that to process OpenAPI specs, but it might come in handy at some point.
> here is a possible source of inspiration or reuse: https://pkg.go.dev/k8s.io/client-go/util/jsonpath

* JSON document with other Marshal or Unmarshal targets, such as MarshalBSON (to store in MongoDB), MarshalJSONB (to store directly into postgres), ...

To avoid undue propagation of dependencies to external stuff like DB drivers etc, these should come as an independent module.

## Design

A `Document` holds a root `Node` and knows which `Store` backs its values.

`Nodes` describe the hierarchy of the JSON document.

Further, we need a `Context` to (optionally) keep the full parsing context and point accurately to errors or warnings.

The `Document` acts lazily: strings and number remain [stores.Handle]`s or `[]byte` until explicitly used.

The `Document` interns strings used as keys (perhaps provide this as an option that may be disabled).

```go
// DocumentFactory receives the desired settings to produce [Document]s in an efficient way.
type DocumentFactory struct {
    documentOptions
}

NewDocumentFactory(opts ...Option) *DocumentFactory {}

// Document produces a Document with the factory settings
func (f *DocumentFactory) Document() Document  { ... }


type documentOptions struct {
  decodeOptions
  encodeOptions
}

type decodeOptions struct {
  lexerFactory func() (lexers.Lexer,func())
  lexerFromReaderFactory func(io.Reader) (lexers.Lexer,func()) // no pool for now
  decodeHooks []hooks.Hook
  captureContext bool
}

type encodeOptions struct {
  writerFactory func() (writers.Writer, func())
  writerToWriterFactory func(io.Writer) (writers.Writer,func()) // no pool for now
  encodeHooks []hooks.Hook
  indentation string
  indent bool
}

// Option to tune a JSON document to your liking
type Option func(*documentOptions)

func WithKeepContext(enabled bool) Option
func WithLexerFactory(func() lexers.Lexer) Option
func WithWriterFactory(func() writers.Writer) Option
func WithLexerFromPool(interfaces.Pool[lexers.Lexer]) Option
func WithWriterFromPool(interfaces.Pool[writers.Writer]) Option
func WithDecodeHooks(callbacks ...hooks.Hook) Option
func WithEncodeHooks(callbacks ...hools.Hook) Option

func noRedeem() func() {}
```

```go
// examples of preset factories -> perhaps in a different package, e.g. "presets", or "documents"
func ObjectDocument() Object
func ArrayDocument() Array
func PositiveIntegerDocument() // ??
func PositiveNumberDocument() // ??
func NonNegativeIntegerDocument() // ?? 
```

The bottom line: DocumentFactory -> Document -> JSON document Node

```go
type Document struct {
  store interfaces.Store
  document
  _ struct{}
}

// document is the internal representation, without the extra pointer to the store.
type document {
  err   error // useful? or node err should be enough?  
  root  Node
  documentOptions // should we use a pointer here instead?
}
```

```go
type Context struct { // text-context to report errors etc
  // see lexer context
}
```

```go
func makeDocument(store Store, node Node) Document

func (d *Document) Kind() types.NodeKind { return d.root.Kind() }

func (d *Document) MarshalJSON() ([]byte, error) {
  jw,redeem := d.writerFactory()
  defer redeem()

  root.encode(w,d.documentEncodeOptions)

  return jw.Bytes(), jw.Error()
}

func (d *Document) UnmarshalJSON([]byte) error {
  lex, redeem := d.lexerFactory()
  defer redeem()

  d.err = nil
  d.root.Reset()

  // consume tokens and build nodes in the document
  root.decode(lex,d.documentDecodeOptions)

  return lex.Error()
}

func (d *Document) Decode(r io.Reader) error {
  lex := d.lexerFromReaderFactory(r)
  d.err = nil
  d.root.Reset()

  root.decode(lex,d.documentDecodeOptions)

  return lex.Error()
}

func (d *Document) Encode(w io.Writer) error {
  jw, redeem := d.writerToWriterFactory(w)
  defer redeem()

  root.encode(jw,d.documentEncodeOptions)

  return jw.Error()
}


// Key returns the document located under key k, or false if this key is not present.
func (d *Document) Key(k string) (Document, bool) {
  return d.root.Key(k)
}

// Elem returns the i-th element in an array, or false if i is equal or larger than the size of the array.
func (d *Document) Elem(i int) (Document, bool) {
  return d.root.Elem(i)
}

// Pairs returns an iterator over keys and documents in an object document
func (d *Document) Pairs() iter.Seq2[string,Document] {
  nodePairs := d.root.Pairs()
  if nodePairs == nil {
    return nil
  }

  return func(yield func(string, Document) bool) {
    for key, node := range nodePairs {
      doc := makeDocument(d.store, node)
      if !yield(key, doc) {
		return
	  }
	}
  }
}

// Elems returns an iterator over the elements of an array document
func (d *Document) Elems() iter.Seq[Document] {
  nodeElems := d.root.Elems()
  if nodeElems == nil {
    return nil
  }

  return func(yield func(Document) bool) {
    for _, node := range nodeElems {
      doc := makeDocument(d.store, node)
      if !yield(doc) {
        return
      }
    }
  }
}

// Get a JSON [Pointer] inside this [Document], or false if the pointer cannot be resolved
func (d Document) Get(p Pointer) (Document, bool) {
}

func (d *Document) Reset() {
  d.root.Reset()
  d.err = nil
  d.Context.Reset()
}

// String representation of a document, like MarshalJSON, but returns a string.
//
// Not specifically optimized: avoid using it for very large documents.
func (d Document) String() string {
  /*  */
  b, _ := d.MarshalJSON()

  return string(b)
}

func (d Document) Ok() bool { return d.err == nil }
func (d Document) Err() error { return d.err }
 
// Pointer implements a JSON pointer.
//
// As specified by :
// * https://www.rfc-editor.org/rfc/rfc6901
// * https://datatracker.ietf.org/doc/html/draft-ietf-appsawg-json-pointer-07
type Pointer struct {
  path []string // TODO: interned?
  Resettable
}

// String representation of a JSON pointer.
func (p Pointer) String() string {
  // TODO: escape stuff
  // see github.com/go-openapi/jsonpointer, but we don't have to deal with all the reflect stuff
  return strings.Join(p.path,"/")
}

// NodeKind describes the kind of node in a JSON document
type NodeKind uint8
const (
  NodeKindNull NodeKind = iota
  NodeKindScalar
  NodeKindObject
  NodeKindArray
)

// ValueKind describes the kind of JSON value held by a node of kind NodeKindScalar
type ValueKind uint8
const (
  ValueKindNull ValueKind = iota
  ValueKindString
  ValueKindNumber
  ValueKindInteger
  ValueKindBool
}

// why do we need to export Node? Why not keeping it private?

// Node describes a node in the hierarchical structure of a JSON document.
// It can be any valid JSON value or construct.
//
// Note: duplicate keys in objects are not allowed.
//
// Perhaps we should make this generic and accept either unique.Handle[string] or string as param type.
type Node struct {
  kind NodeKind
  key unique.Handle[string]  // key for objects properties
  value Value // ValueKind for objects and arrays is NullValue
  children []Node // objects and array have children, nil for scalar nodes
  ctx Context // the error context (maybe use a pointer here?) 
  keysIndex map[unique.Handle[string]]int // lookup index for objects, so Key() finds a key in constant time. Refers to the index of the Node in children.
}

func (n node) Context() Context { // node context }

func (n Node) decode(l lexers.Lexer, opts decodeOptions) (nodes []Node) { // replace by func defaultDecode(l lexer.Lexer) (nodes []Node)
  // TODO: split in parts that may interrupted with hooks
  // if hooks are defined, replace the default decode by a closure that injects the desired hooks and options
  // same for encode()

  // this way, options (incl. hooks) are included at factory build time and code paths remain optimized when options allow that.

  // recursively build nodes
  // very much like easyjon unmarshaling operates

  // inject hooks
  switch n.kind {
    ...
  }
  if !l.Ok() {
    return nodes
  }
  // object
  ...
  // array
  for _, child := range n.children {
    elements = append(elements, child.decode(l, opts)...)
    if !l.Ok() {
      return nodes
    }
    ...
  }
  // errors are trapped in the context

  // Q: do we want to trap only lexing errors in the context? what about higher-level errors?
  // A: nah. We need to capture the context systematically, but only when told to do so.
}

func (n Node) encode(w writers.Writer, opts encodeOptions) []Node {
  // again very much like easyjson node traversal
}

var emptyDocument = Document{}

func (n Node) Key(k string) (Document, bool) {  // always false if kind != NodeKindObject
  if n.kind != NodeKindObject {
    return emptyDocument, false
  }

  return n.children[n.keysIndex[unique.Make(k)]]
}

func (n Node) Elem(i int) (Document, bool) { // always false if kind != NodeKindArray
  if n.kind != NodeKindArray {
    return emptyDocument, false
  }
  if i >= len(n.children) || i < 0 {
    return emptyDocument, false
  }

  return n.children[i]
}

// Pairs return all (key,Node) pairs inside an object.
func (n Node) Pairs() iter.Seq2[string,Node] // nil if kind != NodeKindObject
  if n.kind != NodeKindObject {
    return nil
  }

  return func(yield func(string, Node) bool) {
    for _, pair := range n.children {
      if !yield(pair.key, pair) {
		return
	  }
	}
  }
}

func (n Node) Elems() iter.Seq[Node] { // nil if kind != NodeKindArray
  if n.kind != NodeKindArray {
    return nil
  }

  return func(yield func(Node) bool) {
    for _, node := range n.children {
      if !yield(node) {
        return
      }
    }
  }
}

nullValue = Value{}

func (n Node) Value() (Value, bool) {  // always false if kind != NodeKindScalar
  if n.kind != NodeKindScalar {
    return nullValue, false
  }

  return n.value
}

type Value struct {
  kind ValueKind

  // Maybe we need only one type here which is ScalarValue?
  // The problem with that design is that the value is no longer self-sufficient
  // and needs to know which relevant Store is being used...
  //
  // another problem here is that if we want values to remain lazy for as long as possible, we
  // should keep compressed strings compressed until consumed
  ScalarValue
  // s StringValue
  // n NumberValue  // TODO: add IntegerValue 
  // b BoolValue
}

// alternate design: 100% lazy, but requires a pointer to be embedded.Could be an acceptable trade-off
type ScalarValue struct {
  store  stores.Store
  scalar stores.Handle
}

/*
// ScalarValue methods: on-the-fly resolve of values, i.e. call the corresponding methods from Reference + beef-up processing for numerical types

func (v BoolValue) Bool() (bool,bool) {}
func (v StringValue) String() (string,bool) {}

func (v NumberValue) Number() (string,bool) {}

// checks if supported by native types (limited to 64 bits, let the caller determine if small receivers are needed)
func (v NumberValue) IsInteger() bool {}
func (v NumberValue) IsNegative() bool {}
func (v NumberValue) IsInt64() bool {}
func (v NumberValue) IsUint64() bool {}
func (v NumberValue) IsFloat64() bool {}
*/

// conversions
func (v NumberValue) Int64() (int64,bool) {} // should be preferred when supported
func (v NumberValue) Uint64() (uint64,bool) {} // should be preferred when IsInteger() && !IsNegative() and !IsInt64()
func (v NumberValue) Float64() (float64,bool) {} // should be preferred when !IsInteger() && IsFloat64()
func (v NumberValue) BigInt() (big.Int,bool) {}  // should be preferred whenever IsInteger() && !IsInt64() && !IsUint64()
func (v NumberValue) BigRat() big.Rat {} // should be preferred whenever: !IsInteger() && !isFloat64()
func (v NumberValue) Preferred() any {} // render the preferred go value to represent a number, either int64, uint64, float64, bit.Int or big.Rat

/*
type StringValue []byte
type NumberValue []byte
type BoolValue bool
*/

```
## Typed documents

The `json` package may expose a few restricted types of documents. At the moment, there is:

```go
// Object is a JSON document of type object
type Object struct {
  Document
}

func (o Object) xx // TODO: how to restrict that this unmarshals only objects?
// add Hooks?

// Array is a JSON document of type array
type Array struct {
  Document
}

func (o Object) xx // TODO: how to restrict that this unmarshals only objects?
// add Hooks?

type PositiveInteger struct {
  Document
}

type NonNegativeInteger struct {
  Document
}

type StringOrArray struct {
  Document
}
```

## Builder

Builder is the way to construct `Document`s programmatically or modify existing documents into new instances (a document is immutable).

```go
type Builder struct {
  Document
}

func (b Builder) WithObject() Builder {}
func (b Builder) AddElemWithNode(Node) Builder {}
func (b Builder) AddElemWithDocument(Document) Builder {}
func (b Builder) AddKeyWithNode(string, Node) Builder {}
func (b Builder) AddKeyWithDocument(string, Document) Builder {}
func (b Builder) Document() (Document, error) {}
...
```

Example:
```go
var doc,subDoc Document
_ := stdjson.Unmarshal([]byte(`{}`), &doc)
_ := stdjson.Unmarshal([]byte(`{"a":1}`), &subDoc)

newDoc := doc.Builder().
  AddKeyWithDocument("model", subDoc).
  Document()

if !newDoc.Ok() {
  ...
}

log.Println(newDoc.String())
{"a":{1}}
```
## Hooks

The construction of a JSON document may be customized with hooks (i.e. callbacks).

This allows to reuse much of the document building and rendering. The objective is to extend this infrastructure to more specialized document types, e.g. JSON schemas.

Package `hooks`:

```go
// Hook to customize the behavior of a JSON document
type Hook func(*documentHooks)

type HookFunc func(*HookContext) error // TODO: should add more context so the callback knows better about what's going on.

type documentHooks struct {
  decodeHooks
  encodeHooks
}

// HookContext provides callbacks with more context, possibly customized
type HookContext struct {
  Node Node // current Node
  Key []byte // current key
  Index int // current iteration (element, key)
  Context Context // current parsing context
  Custom any
}

type decodeHooks struct {
  start, beforeObject, afterObject, beforeArray, afterArray, beforeScalar, afterScalar, beforeKey, afterKey, finalize HookFunc
}

type encodeHooks struct {
  start, beforeObject, afterObject, beforeArray, afterArray, beforeScalar, afterScalar, beforeKey, afterKey, finalize HookFunc
}
```

## TODO

There is a contrib module to absorb novel ideas and experimentations without breaking anything.
