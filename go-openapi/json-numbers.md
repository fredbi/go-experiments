# JSON numbers

We have a very general way to resolve numbers lazily.

````go
type NumberValue struct {
  store *stores.Store
  value stores.Handle // fka Reference
}
```

This type stores numerical values and resolves lazily the handle.

We resolve the value to a concrete go type only when we need to.
It can be either of: `int64`, `uint64`, `float64`, `big.Int` or `big.Rat` (JSON floats are rationals).

Most of the time, we don't need to know what is the appropriate representation as a go type.

For instance, for validation, we only need to compare `NumberValue`s, know the min or max value in a collection of NumberValues.

The `MultipleOf` operator may also be used by some JSON schema validation.

Package `numbers` (`core/json/numbers` - Operators on JSON numbers)

```go
// Compare return 1 if a > b, -1 if b < a and 0 if a=b (similar to [bytes.Compare]
func Compare(a json.NumberValue,b json.NumberValue) int {
  // simplistic, unoptimized version
  // likely path to optimization: check sign, check uint64, check float64, and event. check Rat 
  // another way is to amortize the allocation of big.Rat thanks to a pool, since this is just temporary garbage.
  x,y := a.BigRat(), b.BigRat()
  return x.Cpm(y)
}

func Min(numbers ...json.NumberValue) json.NumberValue {
  if len(numbers) == 0 { return json.NumberValue{} }
  min := numbers[0]

  for _, x := range numbers[1:]
    if Compare(min, x) < 0 {
      min = x
    }
  }

  return min
}

func Max(numbers ...json.NumberValue) json.NumberValue {
  ...
}

func IsMultipleOf(a json.NumberValue, multiple json.NumberValue) bool {
  // simplistic implementation
  x,y := a.BigRat(), multiple.BigRat()
  num := y.Num()
  num = num.Abs()
  if num.Cmp(big.NewInt(0)) == 0 {
    return false
  }

  z := a.Quo(a, multiple)

  return z.IsInt()
}
```
