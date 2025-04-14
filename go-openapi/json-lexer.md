# JSON lexer

Detailed design

JSON lexer / parser

* fast, zero-allocation, zero-garbage, limited configurable buffering - but not zero-copy
* produce accurate tokens with verbatim reproducibility in mind (see also a similar project at golang.org/exp)
* maintain the context of syntax errors, and the current parsing context to report higher level errors
* can be used for other things than building documents: parsing or filtering streams, linting json
* ensure strictness about JSON (parsing numbers, escaping unicode, etc), but may optionally be more lax
* the primary use case (that defines the optimized path) is buffered JSON, with no assumption about the lifecycle of the buffer (hence copy is needed).


This is similar in design to `easyjson/jlexer/Lexer`, but with a super reduced API instead: all the information is in the delivered token, so you only need to consume tokens to figure out what to do.

## Fact sheet

* Module: `github.com/go-openapi/core/json`
* Interface definition: `github.com/go-openapi/core/interfaces`
* Default implementation: `github.com/go-openapi/core/json/lexers/default-lexer`
* Alternate implementations: `github.com/go-openapi/core/json/lexers/ld-lexer`
* Contributed implementations: `github.com/go-openapi/core/json/lexers/contrib` [go.mod] - Contributed alternative parser implementations
* Examples: `github.com/go-openapi/core/json/lexers/examples` [go.mod]
* Poolable: Yes

## Interfaces

Package `interfaces`
```go
// Resettable is capable of resetting an object to its default or zero value.
// This is useful when recycling objects made available from a [sync.Pool].
type Resettable interface{
  Reset()
}

// WithErrState is the common interface for all types which manage an internal error state.
type WithErrState interface {
  Ok() bool
  Err() error
}

// TODO: move to package pools at a high level (swag?)

// Pool wraps a [sync.Pool]
type Pool[T any] struct {
  sync.Pool

  Borrow() *T
  Redeem(*T)
}

func (p *Pool[T any]) Borrow() *T {...}
func (p *Pool[T any]) Redeem(*T) {...}

func NewPool[T any]() Pool[T] {
  ...
}

// PoolSlice knows how to pool slices based on their underlying array.
//
// Note: redeeming a slice invalidates all potential slices based on the same underlying array.
type Poolslice[T any] struct {
  Borrow() []T
  Redeem([]T)
}
```

Package `lexers`

```go
// Lexer decomposes a stream or buffer of json bytes into JSON tokens.
type Lexer interface{
  // NextToken retrieves the next token.
  // When the input is exhausted, an empty token is returned and the lexer's state is in error, with Err() returning [io.EOF].
  //
  // The returned [Token] cannot be assumed to contain valid data when [NextToken] is called again. It is up to the caller to
  // make sure that any token content is properly copied in whatever target storage for future use before calling again [NextToken].
  NextToken() Token

  interfaces.Resettable
  interfaces.WithErrState
}

// Example: TODO

// Token represents an elementary JSON token.
type Token struct {}

func (t Token) Kind() TokenKind

type TokenKind uint8

const (
  TokenKindNull TokenKind = iota
  TokenKindNumber
  ...
)
```

## Default implementation

Package `defaut-lexer`:
```go
// L implements a json Lexer.
type L struct {
  _ struct{}
}

func (l *L) NextToken() lexers.Token {
}

// Option to tune the behavior of the lexer to your liking.
type Option func(*options)

// WithVerbatimStrings tells the lexer not to alter input strings in any way.
//
// This means that non-UTF-8 JSON or the JSON way of encoding UTF-8 is left unaltered.
//
// By default, this is disabled and quoted UTF-8 lexemes inside JSON are checked and transformed into UTF-8 runes.
func WithVerbatimStrings(enabled bool) Option {}

// WithStrictNumbers enforces strict validation of the rules defining JSON numbers.
// This is enabled by default.
//
// Whenever disabled, tokens such as "01" or "1e-10" are considered valid numbers.
func WithStrictNumbers(enabled bool) Option {}

// WithMaxContainersStack enables a limit to be checked on the maximum nesting of containers.
//
// This is intended as a security guard against possible DOS attacks with streams such as "{{{{{.." or "[[[[...".
//
// This option is disabled by default when lexing from a buffer,
// and set to 1024 when lexing from a stream of bytes.
func WithMaxNestedContainers(limit uint32) Option {}

// WithMaxTokenLength enables a limit to be checked on the maximum length of strings and numbers.
//
// This is intended as a security guard against possible DOS attacks with streams such as "{"a": "xxxxxxxxxxxxxxxxxxxxxxxxx...".
//
// This option is disabled by default when lexing from a buffer,
// and set to 32767 (2^15-1) when lexing from a stream of bytes.
func WithMaxTokenLength(limit uint32) Option {}

// New lexer from a buffer.
func New(buf []byte, opt ...Option) *L { ...}

// New lexer from an [io.Reader].
//
// Notice that the passed [io.Reader] is wrapped in an internal buffered reader.
func NewFromReader(r io.Reader, opt ...Option) *L { ...}
```

Pooling

```go
var pool pools.Pool[L] = pools.NewPool[L]

func BorrowLexer() *L { return pool.Borrow() }
func RedeemLexer(l *L) { pool.Redeem(l) }
```

## TODO

`ld-lexer`: a JSON lexer that parses newline delimited JSON (e.g. for streams).

Numbers remain strings. Tokens may be converted to numerical types if needed, this is the token consumer's call.

A Number token knows whether it encodes an integer value or a signed (negative) value.

Similarly, a json writer (again along the same lines as easyjson jwriter) is produced to marshal a stream of tokens into a stream of bytes.

Possible extensions:

    the lexer / parser expose a small interface so many different implementations may coexist
    examples: lexer for JSON line-delimited (jsonl), text-delimited JSON, or other stream-oriented formatting of JSON
    there is a contrib sub-modules to more easily introduce novel or experimental features without breaking anything else


