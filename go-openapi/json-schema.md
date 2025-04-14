# JSON schema

Detailed design

JSON schema

The JSON schema type extends the JSON document. This type supports all published JSON schema versions (from draft 4 to draft 2020).

```go
type Option func(*options)

func WithStrictVersion(enabled bool) Option {}
func WithRequiredVersion(JSONSchemaVersion) Option {}
finc WithHooks(hooks ...Hook) Option {}

type NamedSchema {
  Key string
  Schema
}

type Schema struct {
  json.Document

  ref Reference // TODO: Reference stands for JSON Reference ($ref) - fix the misnomer in stores.Reference -> stores.Handle
  dynamicRef Reference
  dynamicAnchor Reference

  types json.StringOrArray // document type exposed by json
  properties []NamedSchema
  allOf []Schema
  anyOf []Schema
  oneOf []Schema
  not Schema // may be null
  dependentSchemas []Schema // may be null
  dependentRequired []Schema

  definitions []Schema // or $defs

  // schema metadata
  Version  // version, minCompatibleVersion, maxCompatibleVersion
  Metadata // ID, $id, description, $comment...

  // validation clauses
  StringValidations
  NumberValidations
  ArrayValidations
  ObjectValidations
  TupleValidations
}

func (s *Schema) UnmarshalJSON([]byte) error {}
func (s *Schema) Decode(io.Reader) error {}

func (s *Schema) decode() error {
// Lexer etc -> validate on the go the JSON schema structure

// schema is object or bool
...
s.err = s.check()

return s.err
}

func (s Schema) Properties() []Schema { return s.properties }
...

type Metadata struct {
  schema json.StringValue
  id json.StringValue
  title json.StringValue
  description json.StringValue
  comment json.StringValue
  default json.Document
  example json.Document
  vocabulary []URI
}

// StringsValidations describe all schema validations available for the string type
type StringValidations struct {
  minLength json.NonNegativeIntegerValue
  maxLength json.NonNegativeIntegerValue // >=0
  pattern json.StringValue
}

type NumberValidations struct {
  minimum json.NumberValue
  exclusiveMinimum json.BoolValue
  maximum json.NumberValue
  exclusiveMaximum json.BoolValue
  multipleOf json.PositiveNumberValue // > 0
}

type ObjectValidations struct {
  minProperties json.IntegerValue
  maxProperties json.IntegerValue
  patternProperties json.Document // Object
  additionalProperties SchemaOrBool
  requiredProperties []json.StringValue
}

type ArrayValidations struct {
  items SchemaOrArrayOfSchemas // Document type exposed by json
  minItems json.NonNegativeInteger
  maxItems json.NonNegativeInteger
  uniqueItems json.BoolValue
}

type TupleValidations struct {
  additionalItems SchemaOrBool
}

type SchemaOrBool struct {
  json.Document
}

type SchemaOrArrayOfSchemas struct {
  json.Document
}

// Reference implements the JSON schema `$ref` applicator.
type Reference struct {
  sync.RWLock

  pointer json.Pointer

  cached *Schema
}

func (r *Reference) Resolve(context.Context) (Schema, error) {}
func (r *Reference) Expand(context.Context) (Schema, error) {}
```

## Builder

```go
type Builder {
  Schema
}

func (b *Builder) WithProperties(...NamedSchema) Builder {}
func (b *Builder) WithAllOf(...Schema) Builder {}
func (b *Builder) Schema() (Schema, error) {}
```

## Hooks
Hooks are a mechanism to customize how a schema is built. This is used for example to derive the OpenAPI definition of a schema from the standard JSON schema.

```go
type Hook func(*schemaHooks)

type schemaHookFunc func(s *Schema) error
type schemaHooks struct {
  beforeKey,afterKey,beforeElem,afterElem,beforeValidate,afterValidate  schemaHookFunc
  setErr func(s *Schema, error) // allows to hook an error on the parent schema 
}
```

## Schema analyzers

    analyzer for serialization & code gen
        analyze namespaces : signals potential naming conflicts
        analyze allOf patterns: serialization vs validation only
        allOf/anyOf/oneOf: inspect overlapping members
        allOf/anyOf/oneOf: inspect primitive only vs
        analyze $ref with additional keys (supported in recent JSON schema drafts)
        detect cyclical $ref
        is null valid?
        has default value?
        is default value valid?
        named vs anonymous schema?
        is schema used ?
        golang-specific analyzer
            inspect naming issues (e.g. ToGoName / ToVarName would drastically change the original naming, or might hurt linter - e.g. non-ascii-)
            inspect enums: primitive vs complex types

    analyzer for validation

        array defines tuple?

        enum values valid (prune) ?

        validation result is always false?

        validation spec is useless (e.g. doesn't apply to appropriate type)

        golang specific:
            is zero-value valid?
            has additionalProperties?
            has additionalItems?

        derive canonical representation: a unique semantic representation of a JSON schema (the order of keys notwithstanding)
