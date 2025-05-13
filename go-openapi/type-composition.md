# Type composition

This corresponds to JSON schema constructs:

* allOf
* anyOf
* oneOf

* As well as the special polymorphic construct supported by OpenAPI: allOf + discriminator

In OpenAPI v2, only `allOf` is supported, together with the polymorphic form of it.

In OpenAPI v3, we add `anyOf` and `oneOf`. These may specify a discriminator, too. The discriminator seems to be optional there.

## allOf (not polymorphic)

Inputs from the analyzer:
* canonical "allOf", ready for serialization: compatible, non-redundant types,
* determine if member schemas are named (as with `"$ref": "#/name"` or anonymous
  * in the case of an anonymous schema, the analyzer can emit propositions   
* prevent incompatible types as in
```yaml
allOf:
  - type: object
  - type: array
```
* lift trivial forms with only one member
* reduce redundant types as in
```yaml
allOf:
  - type: number
  - type: integer
```
which may be simplified into:
```yaml
  type: integer
```

* duplicate field? Subschemas stripped from duplicates as in (example with anonymous schema):
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

Is simplified into:
```yaml
  type: object
  properties:
    a:
      type: object
      minProperties: 3
      maxProperties: 5
    b:
      type: number     
```
* duplicate field? Subschemas stripped from duplicates as in (example with named schema):
```yaml
definitions:
  A1:
    type: object
    properties:
      a:
        type: object
        maxProperties: 5
      b:
        type: number
  A2:
    type: object
    properties:
      a:
        type: object
        minProperties: 3
  A3:
    type: object
    properties:
      a:
        type: object
        patternProperties: 'x-.*'
      c:
        type: number
  M:
    allOf:
    - $ref: '#/A1'
    - $ref: '#/A2'
    - $ref: '#/A3'
```
should be transformed into (if A1, A2 and A3 prove to be unused later on, they might be pruned):
```yaml
  A1:
   ...
  A2:
   ...
  A3:
   ...
  A1WithoutA:
    type: object
    properties:
      b:
        type: number
  A3WithoutA:
    type: object
    properties:
      c:
        type: number
  M:
    allOf:
      - type: object
        properties:
          a:
            type: object
            minProperties: 3  # merged validations for property "a"
            maxProperties: 5
            patternProperties: 'x-.*'
      - $ref: '#/A1WithoutA'
      # "- $ref: '#/A2WithoutA' is removed as it ends up being an empty schema
      - $ref: '#/A3WithoutA'
```

* metadata only as in
```yaml
allOf:
  - $ref: '#/part1'
  - title: 'Override'  # expand title in $ref then override with that one
```

* validation only as in
```yaml
allOf:
  - $ref: '#/part1'
  - required: [prop1]
```
may not be lifted, but it is clear that the generated type may be so.

Go target for:
```yaml
allOf:
  - $ref: '#/A1'
  - $ref: '#/A2'
```

```go
type A struct {
  A1
  A2
}
```

## oneOf

Inputs from the analyzer:
* lift trivial oneOf with only one member
* reduce redundant oneOf members
* lift / merge common properties? (not sure)
* anyOf that requalify formally as oneOf as in:
```yaml
anyOf: # is actually oneOf
  - type: object
  - type: array
```
Options:
* by default, the outcome of the validation should instruct about which branch is valid
* for OpenAPI (not JSONSchema), option to use the discriminator when unmarshaling. (should be cross-checked with the outcome of the validation)

```yaml
oneOf:
  - type: object
  - type: array
```

Go target:
```go
type KindOfA uint
const (
  AIsObject KindOfA = iota
  AIsArray
)

type A struct {
  aDiscriminator KindOfA // a discriminator is always created even if not present in the go-openapi schema (this works for pure JSON schema)
  aObject map[string]any
  aArray []any
}

func (a A) KindOfA() KindOfA { return a.aDiscriminator }

func (a A) IsObject() bool { return a.aDiscriminator == AIsObject }
func (a A) IsArray() bool { return a.aDiscriminator == AIsArray }

func (a A) Object() map[string]any { return a.aObject}
func (a A) Array() []any { return a.aArray}

func (a *A) SetObject(p map[string]any) {
  a.aObject = p
  aDiscriminator = AIsObject
  aArray = nil
}
func (a *A) SetArray(p []any) {
  a.aArray = p
  aDiscriminator = AIsArray
  aObject = nil
}
```

Example with named types:
```yaml
definitions:
  A1:
    type: object
    properties:
      a:
        type: string
  A2:
    type: object
    properties:
      b:
        type: string
  Merged:
    oneOf:
      $ref: '#/A1'
      $ref: '#/A2'
```

Go target:
```go
type KindOfMerged uint
const (
  MergedIsA1 KindOfA = iota
  MergedIsA2
)

type Merged struct {
  mergedDiscriminator KindOfMerged
  mergedA1 A1
  mergedA2 A2
}
```

Example:
```yaml
  color:
    oneOf:
      - const: RGB
        title: Red, Green, Blue
        description: Specify colors with the red, green, and blue additive color model
      - const: CMYK
        title: Cyan, Magenta, Yellow, Black
        description: Specify colors with the cyan, magenta, yellow, and black subtractive color model
```

Should be categorized with members of type "string".
The analyzer should do a smart job to be able to infer that the correct go target is a follows, and
that we don't need the Getter/Setters:

```go
type Color string

const (
  // RGB (means) Red, Green, Blue.
  //
  // Specify colors with the red, green, and blue additive color model.
  RGB Color = "RGB"   // notice that this should be some kind of initialism right there
  // CMYK (means) Cyan, Magenta, Yellow, Black.
  //
  // Specify colors with the cyan, magenta, yellow, and black subtractive color model.
  CMYK Color = "CMYK"
)

func (o *Color) UnmarshalJSON(data []byte) error { ... }
...
```

## anyOf

Inputs from the analyzer:
* removes anyOf that are in fact oneOf

Options:
* may keep only one version (the first that matches validation, "prioritized anyOf")
* may keep all valid versions (-> need to instruct the validator not to short circuit validations)
* use the provided discriminator when unmarshaling

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
