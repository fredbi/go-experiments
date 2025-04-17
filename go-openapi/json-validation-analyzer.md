# JSON schema validation analyzer

This analyzes a JSON schema document to validate JSON data against this schema.

```
type DocValidatorFunc func(d json.Document) error

// Schema is an analyzed schema, with the intent to validate data
type Schema struct {
  schema jsonschema.Schema
  root ast.Tree
}

type SchemaHash [32]byte

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
// TODO: partial analysis is interesting to share some of the processing with other analyzers

// DocValidator returns a closure that can validate a JSON Document against the input schema.
func (a *Analyzer) DocValidator() DocValidatorFunc { }

// ValidationAST returns an AST to construct validators.
func (a *Analyzer) ValidationAST() ast.Tree { ... }

// CanonicalSchema returns the canonicalized JSON schema with its unique 32-bits hash
func (a *Analyzer) CanonicalSchema() (SchemaHash, jsonschema.Schema) {}
```

## Validation AST for a JSON schema

We build an AST of the schema to reason about it without the constraints of its version or changes in semantics.

Package `ast`:

```go
// Tree represents an Abstract Syntax Tree for a single schema
type Tree struct {
  root Node
}

// Walk an AST tree in a depth-first, left-to-right manner
func (t Tree) Walk(apply func(*Node) error)

// Schema returns the canonical JSON schema that results from the AST.
//
// The JSONSchema representation may differ depending on the rendering version selected.
func (t Tree) Schema(json.SchemaVersion) jsonschema.Schema { ... } // TODO: replace param by ...Option

// Forest represents an Abstract Syntax Tree for a collection of schema,
//
// Schema factorization is performed over the whole collection.
type Forest struct {
  roots []Node
}

// Node is an AST node representing an elementary element of syntax of the JSON schema.
type Node struct {
  children []Node
  kind Kind
  hash SchemaHash
  path json.Pointer // holds the path of a schema
  ...
}

func (n Node) Kind() NodeKind { ... }

type Kind uint32

const (
  Empty Kind = iota
  TypeValidations
  StringValidation // minLength, maxLength, pattern, format
  NumberValidation // minimum, exclusiveMinimum, maximum, exclusiveMaximum, multipleOf, format
  ObjectValidation // properties, additionalProperties, required, minProperties, maxProperties
  ArrayValidation  // items (array case), maxItems, minItems
  TupleValidation  // items (tuple case)
  Enum // applies to all types
  Null
  Const
  Ref // children with resolved $ref
  Compositions // allOf, anyOf, oneOf
  // operators
  True
  False
  And // And(children...)
  Or // Or(children...)
  Xor
  Not
  Metadata // keep metadata such as description, title, $id, $comment, default, example, etc
  Annotation
  ... // dependency, if else, media, ...
)


type SubKind uint32

const (
  None SubKind = iota
  StringValidationMinLength
  StringValidationMaxLength
  ...
  NumberValidationMinimum
  NumberValidationMaximum
  ...
)
```

Building the AST:

* walk the schema and convert document nodes into AST node

* the first pass is relatively straightforward:

  * schemas are walked over depth-first, and the JSON path of each item is stored in an AST Node
  * empty schema is True
  * `type` array resolves as an OR group with one element in types and the rest of the schema cloned
  * missing type resolves as "type": ["string","number","bool","array","object","null"]
  * $ref are resolved under a Ref node
  * cyclical $ref are kept (not obvious what to do)
  * validations are arranged by their kind, e.g. TypeValidations, StringValidations, compositions etc.
    Keys in a schema constitute an AND node
  * allOf are expressed as an AND node
  * anyOf => OR
  * oneOf => XOR
  * metadata is stripped into a Metadata node (just leave the portion of the original schema there)
  * Other keys (unrecognized by the JSON schema grammar) are stripped into an Annotations node (just leave the portion of the original schema there)

Notice that we cannot reconstruct the original schema (not unaltered at least) back from the AST.

* Stage 2:
  * at this stage, we no longer have multiple types: prune inapplicable validations
 

## Validation expressions

```go 
package ast

// Expression is a validation expression
type Expression struct {
  nodes []Node
}

// Filter out the expression of all validations that do not apply to a given type.
//
// If the type is null, all validations are stripped.
//
// Example [Maximum, MinLength].FilterForType("string") = MinLength
func (e Expression) FilterForType(string) Expression {
  ...
}

// Evaluate nodes in an expression recursively and replace always true or always false statements by nodes True or False.
//
// Examples:
// (Minimum(4), Maximum(2) <=> Range(4,2) <=> False
// (Enum([-1,1], Maximum(3)) => Enum([-1,1]) // TODO: Enum([a,b]) becomes OR(CONST(a),CONST(b))
func (e Expression) Evaluate() (Expression,Node) {
  ...
}

// Reduce the expression recursively by simplifying away redundant nodes and nodes that are true or false.
//
// It returns the reduced expression and a Node of kind Annotations or Empty annotated with remarks.
//
// Examples:
//
//   AND(x,True) = x, AND(x, False) = Falsen, OR(x,True) = x, XOR(x,True) = NOT(x), etc
//   AND(x,x) = x, NOT(NOT(x)) = x ...
func (e Expression) Reduce() (Expression,Node) {
  ...
}

// Factorize the expression:
//
// e.g. AND(OR(a,b),OR(a,c)) becomes AND(a, OR(b,c))
// NOT(OR(a,b)) becomes AND(NOT(a),NOT(b))
func (e Expression) Factorize() (Expression,Node) {
  ...
} 
```

## Canonical form of a JSON schema

We define a "canonical form" of a JSON schema in order to:
* recognize any given validation pattern in some unique way
* factorize validators whenever possible
* be able to refactor and optimize a validator, so it becomes relatively immune to schema authoring style

This canonical form is built while from the AST (see below). It provides a JSON schema-compatible view of the AST.

The canonical form is required to build fast validators and is desirable (perhaps not required) to contribute to the accuracy of other aspects of model code gen.

Rules for a canonical schema:
* warning annotations: x-go-analysis-warn: ["{warn code}", "{description}"]
* path annotations: all data items in schema are annotated with their expected hierarchical path (JSON pointer) in JSON data, e.g. "x-go-analysis-path: /items/1"
* empty schemas are always converted with a boolean true  {} => true
* Likewise, the structure 'not: {}' becomes: false
* `type`
  * `type` is required. If missing from the original schema, this means `type`: [ array, object, string, number, bool, null]
  * `type` arrays are not sensitive to the order of types. We reorder such an array from most specific to less specific: [ null, bool, number, string, object, array ]
  * `type` is transformed into a single value, `type` arrays are rearranged in a child `oneOf` clause (since type clauses are mutually exclusive)
  * Outcome: a canonical schema always have one and only one `type` 
* Keys
  * the order of keys in a schema object does not contribute to the determination of the unique hash.
  * The hash is not sensitive to key ordering, but the canonicalized version of the schema keeps the original order of keys
  * Outcome: schemas with the same hash (same validation rules) may have a different string representation
* Metadata: title, comments, and descriptions do not contribute to the hash. These are kept in the canonicalized schema, but irrelevant for validation. Use WithMetadataStripped(true) to remove them.
* $ref
  * All $ref's are rewritten in full to remove all need for "syntactic sugar" such as "$id", "$dynamicAnchor", etc. These keywords do not contribute to the computation of
    the unique hash: only the hash of the resolved schema matters.
  * All schemas in $ref are recursively canonicalized. Two different $ref that yield the same Hash are considered equivalent.
  * In the resulting canonical schema, the first explored $ref with a given hash is used (deterministic output). If anonymous schemas compete with named schemas, the first named schema found takes over.
* redundant validations
  * validations that do not apply to the given type of schema are removed (raise warning)
* type-specific validations
  * a schema that mixes validations for different types is subject to a rewrite: if the schema supports multiple types (see `type` above) it becomes split into `oneOf` schemas of one and only one type
  * validations that do not apply to each type are pruned
  * validations that apply to several types are kept
  * validations that apply to all members in the `oneOf` clause are lifted
  * if null is allowed, the member in the oneOf with type: null cannot have any validation attached. This is annotated as "x-go-analysis-may-be-null: true"
* enum: enum values that do not pass other validations at the same level are pruned (raise warning)
  * if all enum values are pruned, the schema resolves as false.
* jsonschema version-specific behavior (i.e. deprecated stuff):
  * the canonical schema is always represented in the latest supported version of JSON schema (i.e. draft 2020)
  * deprecated features are rewritten
  * an annotation is left to mark the change, x-go-canonical-deprecation: "old construct". These annotations are ignored for hash computation.
* validation expressions
  * any validation expression returns either true or false
  * validation clauses constitute an expression with an implied "AND" logical
  * validations inside allOf imply an expression like ( member1 AND member2 AND ...)
  * validations inside anyOf imply an expression like ( member1 OR member2 OR ...)
  * validations inside oneOf imply ( member1 XOR member2 XOR ...)
  * "not" implies NOT ( member )
  * for any given schema, we can find the equivalent expression with the operators: (, ), AND, OR, XOR, NOT  
  * expressions that are always false or always true are simplified away:
    * x AND false = false, x AND true = x, x OR false = x, x OR true = true, x XOR false = x, x XOR true = NOT x; not false = true, not true = false 
* inline schemas in allOf, oneOf, anyOf are all replaced by $ref. The name of the $ref is generated from the path of the parent. Such $ref's are annotated as anonymous: x-go-analysis-anonymous-ref: true.
  and might disappear after simplification by unique hash
* mutually exclusive members in allOf result in the allOf to be always false
* mutually exclusive members in anyOf are transformed into oneOf clauses. Exemple anyOf: [ member1, member2, member3] where member1 AND member2 = false becomes:
  * anyOf: [ oneOf: [member1, member2], member3 ] 
* mutually exclusive members in oneOf are annotated to keep that information.
  * The following json schema annotation is added to mark that fact (as of draft 2020, annotations are allowed with $ref):
  * Ex: oneOf: [ {type: string}, {type: number}, {type: integer} ]  => string AND number = false, string AND integer = false, number AND integer != false
  * TODO: wrong example!! $ref rewrite should only apply to composite members (objects or arrays, not primitive types)
  * => oneOf: [ {$ref: '#/xyz', x-go-analysis-mutually-exclusive: [0,2]}, {$ref: '#/abc', x-go-analysis-mutually-exclusive:[0,1]}, {$ref: '#/bcd', x-go-analysis-mutually-exclusive: [0,2]]
  * Ex: oneOf: [ string, number ]  => string AND number = false => oneOf: [ string, number ], x-go-mutually-exclusive: [ 0,1 ]
  * These annotations are relevant and used for hash computation
* order of validations: principle most specific (or fastest check) first (fail early principle). E.g. type always comes first. If we have enum, then start with that.
* check numeric validations:
  * minimum <[=] value <[=] maximum is evaluated for compatibility (may be simplified to false if incompatible and raise warning)
  * whenever valid, such ranges are annotated with the tuple: 'x-go-analysis-numeric-range: [minimum, exlusiveMinimum[true|false], maximum,exclusiveMaximum]
  * range and implied type: if maximum < (range of a type), add the "format" clause to the numerical type to reduce the type (i.e. "format": "int32", ...). 
  * check compatibility with numerical format / possibly truncate range to the bound implied by format
  * if the min/max range is redundant with format, replace min/max by format (keep the range annotation, add the "x-go-analysis-implied-format: true" annotation)
  * multipleOf:
    * type implied by multipleOf:
      * a number multipleOf an integer is reduced to type integer
      * integer multipleOf a non-integer implies that the multipleOf may be rewritten to use an integer multiple (raise warning)
        * Ex: {type: integer, multipleOf: 1.1} => {type: integer, multipleOf: 11}
        * Ex: {type: integer, multipleOf: 0.5} => {type: integer}
        * Ex: {type: integer, multipleOf: 0.8} => {type: integer, multipleOf: 4} TODO: find the explicit rule, not just examples
* check string validations
  * minLength, maxLength checked for compatibility, if minLength = maxLength add annotation "x-go-analysis-string-length-equals: {minLength}"
  * recognize simple idioms: minLength > 0 only is annotated "x-go-analysis-string-not-empty: true"
  * pattern:
    * check regexp idiom compatibility: if compatible with RE2, annotate "x-go-analysis-pattern-compatibility: [re2]", if pure ECMA regexp, add ecma in the compatible list
    * if doesn't compile, raise error (?) or warn
    * recognize simple regexp patterns: e.g '/^.+$/' (not empty), plus some other (alpha chars, word chars, no blank space...). For each recognized pattern,
      we add the corresponding annotation, to instruct implementations to pick a faster alternative than applying a regexp
  * string formats:
    * for each format, we may check redundant validations (e.g. format: "date-time" and minLength>0 is redundant: format is sufficient)
    * TODO: for each supported type, the strfmt registry should export a list of short-circuit validations, e.g. fixed length for uuid, date etc.
* array validations: maxItems, minItems (if both, check range validity and annotate with "x-go-analysis-array-len-range: [minLen,maxLen]"
* tuples: explicit tuples with x-go-analysis-tuple: true, if additional items are supported annotate this
* object validations
  * check min/maxProperties, if range "x-go-analysis-object-len-range: [minProperties,maxProperties]"
  * "properties" and "required"
    * properties in the require array are reordered to reflect the order of their appearance in the object definition
    * if the require array contains undefined properties, they are appended to the properties object with schema: true
    * a key in properties means: "if the property shows up, it must validate this schema" =
        * (property missing) OR (property present and validate schema) 
      * so the validation of NOT {schema property} (not required) is:
        *      NOT ((property missing) OR (property present and validate schema))
        *  <=> (property present) AND NOT (property present and validate schema)
        *  <=> (property present) AND (NOT (property present) OR NOT validate schema)
        *  <=> (property present) AND (property absent OR NOT validate schema)
        *  <=> (property present AND property absent) OR (property present AND NOT validate schema))
        *  <=> false OR (property present AND NOT validate schema)
        *  <=> property present AND NOT validate schema
        *  <=> required: [{property}, {property}: {not: {schema}}
    *        NOT {schema property} {required} translates as:
      *  <=> NOT (property present and validate schema)
      *  <=> (property absent} OR NOT validate schema
      *  {property}: {not {schema}} and removed from the required array
  * if addditionalProperties are supported annotate this
    * not additionalProperties is complicated ...
    * not additionalItems is complicated ...
* overlapping properties in allOf/anyOf/oneOf constructs : TODO
* default value:
  * raise warning if default doesn't validate the schema
* hash: schemas are annotated with their unique hash: x-go-analysis-hash: "hex hash string"
* expression development by path: for each data path, expressions are simplified (factorized with AND/developed with OR, NOT simplified)
  * raise warning for data paths that resolve to "false" (data cannot be validated)
  * raise warning for data paths that resolve to "false" unless "null" (non-null data cannot be validated): simplify as schema "type: null" (keep nested annotations)
* if / then / else: TODO
* dependentSchema / dependentRequired: TODO


