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


//DocumentShareable allows [stores.Store] objects to provide more services to be shared across [json.Document]s.
type DocumentShareable interface {
  // Loader function to grab JSON from a remote or local file location.
  Loader() func(string) ([]byte,error)
}
```
