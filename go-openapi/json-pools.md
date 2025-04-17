# pools


```go
// TODO: move to package pools at a higher level (core? swag?)

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
