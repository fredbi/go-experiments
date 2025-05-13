# Type composition

This corresponds to JSON schema constructs:

* allOf
* anyOf
* oneOf

* As well as the special polymorphic construct supported by OpenAPI: allOf + discriminator

## allOf (not polymorphic)

Inputs from the analyzer:
* canonical "allOf", ready for serialization: compatible, non-redundant types,
* prevent incompatible types as in
```yaml
allOf:
  - type: object
  - type: array
```
* reduce redundant types as in
```yaml
allOf:
  - type: number
  - type: integer
```

* duplicate field? Subschemas stripped from duplicates as in:
```yaml
allOf:
  - type: object
    properties:
      a:
        type: object
        maxProperties: 5
      b:
        type: number
  - type: object
    properties:
      a:
        type: object
        minProperties: 3
```

* metadata only as in
```yaml
allOf:
  - $ref: '#/part1'
  - title: 'Override'
```

* validation only as in
```yaml
allOf:
  - $ref: '#/part1'
  - required: [prop1]
```

Go target:
```go
type A struct {
  AllofA1
  AllOfA2
}
```

## oneOf

Inputs from the analyzer:
* anyOf that requalify formally as oneOf as in:
```yaml
anyOf: # is actually oneOf
  - type: object
  - type: array
```
Options:
* enforces discriminator

Go target:
```go
type A struct {
  oneOfADiscriminator int
  oneOfA1 OneOfA1
  oneOfA2 OneOfA2
}

func (a A) KindOfA() string

func (a A) IsOneOfA1() bool
func (a A) IsOneOfA2() bool

func (a A) AsOneOfA1() OneOfA1 {}
func (a A) AsOneOfA2() OneOfA2 {}
```

## anyOf

Inputs from the analyzer:
* removes anyOf that are in fact oneOf

Options:
* may keep only one version (the first that matches validation, "prioritized anyOf")
* may keep all valid versions
* enforces discriminator

Go target:
```go
type A struct {
  // anyOfADiscriminator int
  isAnyOfA1 bool
  isAnyOfA2 bool
  anyOfA1 AnyOfA1
  anyOfA2 AnyOfA2
}

func (a A) KindsOfA() []string

func (a A) IsAnyOfA1() bool
func (a A) IsAnyOfA2() bool

func (a A) AsAnyOfA1() AnyOfA1 {}
func (a A) AsAnyOfA2() AnyOfA2 {}
```

## allOf (polymorphic)

TODO
