# JSON schema validation analyzer

This analyzes a JSON schema document specifically with the intent to validate JSON data against this schema.

```
type DocValidatorFunc func(d json.Document) error

// Schema is an analyzed schema, with intent to validate data
type Schema struct {
  schema jsonschema.Schema
}

type SchemaHas [32]byte

func New(s jsonschema.Schema, opts ...Option) Schema {
  return Schema {
    schema: s,
  }
}

type Option func(*options)

func WithMetadataStripped(enabled bool) Option { ... }

// Analyze the input JSON  schema.
func (a *Analyzer) Analyze() error {
}

// DocValidator returns a closure that can validate a JSON Document against the input schema.
func (a *Analyzer) DocValidator() DocValidatorFunc { }

// ValidationAST returns an AST to construct validators.
func (a *Analyzer) ValidationAST() AST { ... }

// CanonicalSchema returns the canonicalized JSON schema with its unique 32-bits hash
func (a *Analyzer) CanonicalSchema() (SchemaHash, jsonschema.Schema) {}


```

## Canonical form of a JSON schema

We define a "canonical form" of a JSON schema in order to:
* recognize any given validation pattern in some unique way
* factorize validators whenever possible
* be able to refactor and optimize a validator, so it becomes relatively immune to schema authoring style

Rules for a canonical schema:
* `type`
  * `type` is required. If missing from the original schema, this means `type`: [ array, object, string, number, bool, null]
  * `type` arrays are not sensitive to the order of types. We reorder such an array from most specific to less specific: [ null, bool, number, string, object, array ]
  * `type` is transformed into a single value, `type` arrays are rearranged in a child `oneOf` clause (since type clauses are mutually exclusive)
  * Outcome: a canonical schema always have one and only on `type` 
* Keys
  * the order of keys in a schema object does not contribute to the determination of the unique hash.
  * The hash is not sensitive to key ordering, but the canonicalized version of the schema keeps the original order of keys
  * Outcome: schemas with the same hash (same validation rules) may have a different string representation
* Metadata: title, comments, descriptions do not contribute to the hash. These are kept in the canonicalized schema, but irrelevant for validation. Use WithMetadataStripped(true) to remove them.
* $ref
  * All $ref are rewritten in full to remove all need for "syntaxic sugar" such as "$id", "$dynamicAnchor", etc. These keywords do not contribute to the computation of
    the unique hash: only the hash of the resolved schema matter.
  * All schemas in $ref are recursively canonicalized. Two different $ref that yield the same Hash are consider equivalent.
  * In the resulting canonical schema, the first explored $ref with a given hash is used (deterministic output)

## Validation AST

This builds the validation tree to apply on some input JSON data - it describes validations with abstract operators.
