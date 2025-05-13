# Fresh ideas to move forward

## 8 proposals

1. Acknowledge the fact that JSON schema is perhaps not the best choice to accurately specify a serialization format.
   > JSON schema was designed as a way to express validation constraints, not to specify explicitly a data format.
   > 
   > The approach folowed by JSON schema is to derive a data format _implicitly_ from a series of validation constraints that are applied.
   > 
   > In practice, constraints are most of the time strong enough to imply a given data type such as a `struct` type.
   > And this is why it is possible to generate code for our APIs. Most of the time. But not always.
   >
   > The more I think about this design choice by Swagger authors, the more I think it is unappropriate.
   > 
   > All the more as JSON schema has evolved a lot, toward a meta-vocabulary to describe many things,
   > but not how to serialize DATA TYPES. And this is by design, so there is no point in complaining about it...
   > 
   > Think about how you would convert a JSON schema into a `protobuf` spec for example. It is harder than it seems.
   >
   > In practice, this means that we should allow much more "intent metadata/desired target" input to flow from the developer,
   > which are not part of the spec.
   > 
   > This could be configured as rules or as ad-hoc `x-go-*` extensions.
   
2. Abandon the "dynamic JSON" go way of doing things
   > Work with opaque JSON documents that you can navigate (similar to what `gopkg.in/yaml.v3` does).
   * Stop exposing data objects with exported fields.
   > This has proved to be difficult (or impossible) to maintain over the long run. Example: `spec.Spec`
   * Stop using go maps to represent objects.
   > The unordered, non-deterministic property of go maps makes them inappropriate for code, spec, or doc generation
   
   * Stop using external libraries for the inner workings.
   > easyjson is just fine, but it is not so hard to work out a similar lexing/writing concept,
   > more suited to our JSON spec and schema representation.
   > 
   > At the core of the new design, we have a fast JSON parser and document node hierarchy with "verbatim" reproducibility as targets.
   > 
   > Documents may share an in-memory "document storage" to function as a cache for remote documents and quick `$ref` resolution.
   > 
   > These base components are highly reusable in very different contexts, even for other things than APIs
   > (e.g. from json linter to json stream parsing).
   
3. Abandon the json standard library for all internal codegen/validation/spec gen use cases
   > use a bespoke json parser to build such JSON document objects
   
4. Separate entirely JSON schema support from OpenAPI spec
   * JSON schema parsing, analysis and validation support independent use cases (beyond APIs): they should be maintainable separately
   * Similarly, the jsonschema models codegen feature should be usable independently (i.e. no need for an OAI spec)
   * JSON schema support is focused on supporting the various versions and flavors of json schema
     
5. Since we no longer need to represent a spec or a JSON schema with native go types
> We are confident that we can model these with full accuracy, including nulls, huge numbers, etc.
> 
> Speed is not really an issue at this stage. Accuracy is.
> 
> Spec analysis should produce a structure akin to an AST, which focuses on describing the source JSON schema structure well, never mixing with requirements from the target codegen language.
> Example: when analyzing, we should not express a situation like "needs a pointer" or "need to be exported", but rather "can be null", or "can't be null".

6. Abandon recursive-like, self-described JSON validation.
> Elegance has been kind of delusional here and ended up with a memory-hungry, slow JSON validator that could only be used
> for spec validation and not (realistically) when serving untyped API.

   * Proposed design: after schema analysis ("compile" time) we may produce a closure that more or less wires the validation path efficiently.
   * Proposed design (meta-schema or spec validation)

   > Most of the validation work is done during the analysis stage, which would also raise
   > warnings about legit but unworkable json schema constructs

7. Analysis should distinguish 2 very different objectives:
    * analysis with the intent to produce a validator
      * for instance, this analysis doesn't need to retain metadata (descriptions, `$comment`), it doesn't need to retain the `$ref` structure, ...
      * this analysis is focused on producing the most efficient path for validation: it may short-circuit redundant validations, pick the best fail-fast path etc
      * at this stage there is no validator yet: one outcome could be a validator closure, and another one could be a generated validation function
    * analysis with the intent to produce code (requires a much deeper understanding to capture the intent of the schema)
      * at this stage, we may introduce a target-specific analysis. For example, it may be a good idea to know if the "zero value" (a very go-ish idea) is valid to take the right decision.
      * it becomes also interesting to infer the _intent_ that comes with some spec patterns
8. Codegen should be very explicit about separating concerns
   > More specifically, be explicit about what pertains to the analysis of the source spec (e.g. "this is a polymorphic type"),
   > and what pertains to the solution in the target language (e.g. "this is an interface type", "this is a pointer").
   > This would make it easier to support alternative solutions that represent the same specification
   > (in this example, the developer might find our generated interface types awkward to use and might prefer a composition of concrete types)

## [A mono-repo layout](./mono-repo.md)

## New components to support JSON

The main design goal are:

* to build immutable data structure that clone efficiently
* to be able to unmarshal and marshal JSON verbatim
* to support JSON even where native types don't, e.g. non-UTF-8 strings, numbers that overflow float64, null value
* to be able (on option) to keep an accurate track of the context of an error
* to drastically reduce the memory needed to process large JSON documents (i.e. ~< 4 GB)
* to limit garbage collection to a minimum viable level.
* to access a document with JSON pointers

Core components to be brought to the table:

* JSON lexer / JSON writer
* JSON Document, backed by a JSON store
* JSON Schema extends the JSON Document and so does the OpenAPI spec.

Since these structures do not use golang native types (everything remains pretty much raw `[]byte` under the hood),
and only resolve lazily into go types, we can reason with complete JSON semantics without paying much attention
to null vakues, zero values, pointers and the like.

In this implementation, pointers are never used to store values internally.

### [JSON lexer / parser](./json-lexer.md)

### [JSON document and node](./json-document.md)

A document can be walked over and navigated through:

* object keys and array elements can be iterated over
* json pointers can be resolved (json schema `$ref` semantics are not known at this level) efficiently (e.g. constant time or o(log n))
* desirable but not required: should lookup a key efficiently (e.g. constant time or o(log n))

> NOTE: this supersedes the features previously provided by `jsonpointer`, `swag`

Options:

* a document may keep the parsing context of its nodes, to report higher-level errors
* a document may support various tuning options regarding how best to store things (e.g. reduce numbers as soon as possible, compress large strings, etc)
* ...

#### How about YAML documents?

We keep the current concept that any YAML should be translated to JSON before being processed.

What about "verbatim" then? Well, we just have to punt it: the promise is JSON-verbatim and yes, reconstructed YAML
would lose some of its initial structure (like doc refs, tags, etc).

YAML is just too complex to promise anything like that anytime soon.

#### Possible extensions

* the basic requirement on a JSON document is to support MarshalJSON and UnmarshalJSON, provide a builder side-type to support mutation
* more advanced use-case may be supported by additional (possibly composed) types
* examples:
  * JSON document with JSONPath query support
  * JSON document with other Marshal or Unmarshal targets, such as MarshalBSON (to store in MongoDB), MarshalJSONB (to store directly into postgres), ...

To avoid undue propagation of dependencies to external stuff like DB drivers etc, they should come as independant modules.

There is a contrib module to absorb novel ideas and experimentations without breaking anything.

#### [JSON document store](./json-store.md)

A document store organizes a few blocks of memory to store (i) interned strings and (ii) JSON scalar values packed in memory.

#### [JSON schema](./json-schema.md)

The JSON schema type extends the JSON document.
This type supports all published JSON schema versions (from draft 4 to draft 2020).

Hooks are a mechanism to customize how a schema is built. This is used for example to derive the OpenAPI definition of a schema from the standard JSON schema.

```go
type Hook func(*schemaHooks)

type schemaHookFunc func(s *Schema) error
type schemaHooks struct {
  beforeKey,afterKey,beforeElem,afterElem,beforeValidate,afterValidate  schemaHookFunc
  setErr func(s *Schema, error) // allows to hook an error on the parent schema 
}
```

#### Schema analyzers

* analyzer for serialization & code gen
  * analyze namespaces : signals potential naming conflicts
  * analyze allOf patterns: serialization vs validation only
  * allOf/anyOf/oneOf: inspect overlapping members
  * allOf/anyOf/oneOf: inspect primitive only vs 
  * analyze $ref with additional keys (supported in recent JSON schema drafts)
  * detect cyclical $ref
  * is null valid?
  * has default value?
  * is default value valid?
  * named vs anonymous schema?
  * is schema used ?
  * golang-specific analyzer
    * inspect naming issues (e.g. ToGoName / ToVarName would drastically change the original naming, or might hurt linter - e.g. non-ascii-)
    * inspect enums: primitive vs complex types
    
* |analyzer for validation](./json-validation-analyzer.md)
  * array defines tuple?
  * enum values valid (prune) ?
  * validation result is always false?
  * validation spec is useless (e.g. doesn't apply to appropriate type)
  * golang specific:
    * is zero-value valid?
    * has additionalProperties?
    * has additionalItems?
   
  * derive canonical representation: a unique semantic representation of a JSON schema (the order of keys notwithstanding)
    
## Swagger spec schema

```go
type Schema struct {
  json.Schema // with hooks to restrict and extend JSON schema
}

type SimpleSchema struct {
  json.Schema  // more hooks to restrict 
}

type pathItem struct {
  method http.Method
  path []string
  securityDefinition ...
  tags []string
  operation *Operation
}

type Spec struct {
  json.Document

  Metadata

  parameters []SimpleSchema
  responses []Schema

  operations []Operation
  pathItems []pathItem
  definitions []Schema
}

// Builder may construct an OpenAPI specification programmatically.
type Builder struct {
  Spec
}

func (b Builder) Spec() (Spec,error) {}
```

Spec analyzer:
* find factorizations (e.g. parameters validations vs schema validations)
* find factorizations (e.g. common parameters, common responses)

Spec linter:
* based on linting rules
 
## [Model generation](./modelgen.md)

Ideas:

* self-sustained library
* provided with a CLI for testing, experimenting, or focused usage (same style as go-swagger, just more focused)
* ships with templates
* reuses templates-repo
* primary support is for golang, but I'd like to add protobuf just to prove the concept of multi-target code generation
* keep the original objectives: clean code, readable, linter-friendly and godoc-friendly
* inject much more customization with `x-go-*` tags
* focused primarily on supporting JSON schema draft 4 to 2020, plus swagger v2 and OAI v3.x idioms

Features outline:
* support null, use of pointers
  * by default doesn't use pointers but "nullable types" to remain 100% safe regarding JSON semantics
  * customizable rendering available locally or globally (e.g. with pattern matching rules) with the following options:
    * favor native types, assuming that zero-value vs null doesn't affect semantics (recommended approach)
    * use pointers with parsimony, whenever explicitly told to (e.g. x-go-nullable)
    * generated structs are all "nullable" without resorting to a pointer (i.e. embed an "isNuDefined bool" field)

```go
type Nullable[T any] struct {
  isDefined bool
  T // doesn't work in golang yet
}

func (n Nullable[T]) IsNull() bool {
  return !n.isDefined
}

func (n Nullable[T]) Value() T {
  if n.isDefined {
    return n.T
  }

  var zero T
  return zero
}
```

* polymorphic types (swagger v2), aka "inheritance" can be rendered in several ways
    * are supported for any kind of container, not just plain vanilla
    * with an interface type (default, like go-swagger currently does)
    * with embedded types on demand
    * other designs? (contributed)

* custom types
    * either wrapped (embedded structs) or imported as is (like now)
    * may use internal json document as a legit type (dev wants an opaque structure)
    * OrderedMaps or custom map implementations

* marshaling/unmarshaling
    * option to support different JSON libraries - default remains the standard lib
    * option to generate code for our internal json parser, easyjson and perhaps a few others
    * built-in option to support streaming (internal json parser)
 
* validation
    * simplified interface: Validate(context.Context) error <- no strfmt specification (injected in context)
    * option not to generate validation functions (like now)
    * option to inject runtime validators (closure functions produced by schema analysis)
    * option to generate UnmarshalValidate and MarshalValidate (has impact on the choice for pointers etc)
    * implementation of "required": either delegate to UnmarshalValidate/MarshalValidate or post-unmarshaling by Validate()

 * oneOf, anyOf
    * analysis determines if oneOf holds some mutually exclusive cases
    * several available rendering designs: composition, private members with accessors
    
 * allOf
    * keep type composition for most cases (e.g. identified as "serialization" schemas)
    * propose the option to inline allOf members rather than composing types ("lifting")
    * support "validation only allOf" idiom
    * support parts with common property names (i.e. inline & merge validations)
  
* contributed extensions
    * Pluggable extensions to the model generation logic (i.e. routes processing logic to plugin whenever some `x-*`extension is found
    * Contributed templates to support pluggable extensions

Out of this scope: generating schema from JSON data (the classical nodeJS basic tool)?

## [API generation](./apigen.md)

Same modular and extensible approach as the one described for models.

Features:
* handlers generation
* generation of supporting files
  * main server
    * Here we can't take some inspiration from others, who support common frameworks such as gin, fiber, etc
  * API initialization

* client generation 

Need a lot of additions to support the new concepts in OAI v3.

## Spec generation

I have to reflect on that one. I am out of ideas for now.

