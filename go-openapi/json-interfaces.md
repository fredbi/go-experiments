# JSON interfaces

```go
package interfaces

// Resettable can reset an object to its default or zero value.
// This is useful when recycling objects made available from a [sync.Pool].
type Resettable interface{
  Reset()
}

// WithErrState is the common interface for all types that manage an internal error state.
type WithErrState interface {
  Ok() bool
  Err() error
}

```
