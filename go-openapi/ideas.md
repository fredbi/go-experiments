# Fresh ideas to move forward

## 8 proposals

1. Acknowledge the fact that json schema is perhaps not the best choice to accurately specify a serialization format.
   > JSON schema was designed as a way to express validation constraints, not to specify explicitly a data format.
   > The json schema approach is to derive a data format _implicitly_ from the series of validation constraints that are applied.
   > In practice, constraints are most of the time strong enough to imply a given data type such as a `struct` type. But not always.
   > Think about how you would convert a json schema into a proto buf spec for example. It is harder than it seems.
   > In practice, this means that we should allow much more "intent metadata/desired target" input to flow from the developer, which are not part of the spec.
   > This could be configured as rules or as ad-hoc `x-go-*` extensions.
   
2. Abandon the "dynamic JSON" go way of doing things: work with opaque JSON documents that you can navigate (similar to what `gopkg.in/yaml.v3` does)
   * Stop exposing data objects with exported fields. This has proved to be difficult (or impossible) to maintain over the long run
   * Stop using go maps to represent objects. The unordered, non-deterministic property of go maps makes them inappropriate for code, spec, or doc generation
   * Stop using external libraries for the inner workings. easyjson is just fine, but it is not so hard to work out a similar lexing/writing concept, more suited to our
     JSON spec and schema representation.
     At the core of the new design, we have a fast JSON parser and document node hierarchy with "verbatim" reproducibility as targets.
     Documents may share an in-memory "document storage" to function as a cache for remote documents and quick `$ref` resolution.
     These base components are highly reusable in very different contexts than APIs (e.g. from json linter to json stream parsing).
3. Abandon the json standard library for all internal codegen/validation/spec gen use cases: use a bespoke json parser to build such JSON document objects
4. Separate entirely JSON schema support from OpenAPI spec
   * JSON schema parsing, analysis and validation support independent use cases (beyond APIs): they should be maintainable separately
   * Similarly, the jsonschema models codegen feature should be usable independently (i.e. no need for an OAI spec)
   * JSON schema support is focused on supporting the various versions and flavors of json schema
5. Since we no longer need to represent a spec or a JSON schema with native go types, we are confident that we can model these with full accuracy, including nulls, huge numbers, etc.
   Speed is not really an issue at this stage. Accuracy is.
   Spec analysis should produce a structure akin to an AST, which focuses on describing the source JSON schema structure well, never mixing with requirements from the target codegen language.
   Example: when analyzing, we should not express a situation like "needs a pointer" or "need to be exported", but rather "can be null", or "can't be null".
6. Abandon recursive-like, self-described JSON validation. Elegance has been kind of delusional here and ended up with a memory-hungry, slow JSON validator that could only be used
   for spec validation and not (realistically) when serving untyped API.
   * Proposed design: after schema analysis ("compile" time) we may produce a closure that more or less wires the validation path efficiently.
   * Proposed design (meta-schema or spec validation): most of the validation work is done during the analysis stage, which would also raise warnings about legit but unworkable json schema constructs
7. Analysis should distinguish 2 very different objectives:
    * analysis with the intent to produce a validator
      * for instance, this analysis doesn't need to retain metadata (descriptions, `$comment`), it doesn't need to retain the `$ref` structure, ...
      * this analysis is focused on producing the most efficient path for validation: it may short-circuit redundant validations, pick the best fail-fast path etc
      * at this stage there is no validator yet: one outcome could be a validator closure, and another one could be a generated validation function
    * analysis with the intent to produce code (requires a much deeper understanding to capture the intent of the schema)
      * at this stage, we may introduce a target-specific analysis. For example, it may be a good idea to know if the "zero value" (a very go-ish idea) is valid to take the right decision.
      * it becomes also interesting to infer the _intent_ that comes with some spec patterns
8. Codegen should be very explicit about what pertains to the analysis of the source spec (e.g. "this is a polymorphic type", and what pertains to the solution in the target language
    (e.g. "this is an interface type"). This would make it easier to support alternative solutions to represent the same specification (in this example, the developer might find generated
    interface types awkward to use and would prefer a composition of concrete types)

## A mono-repo layout

One git repo, many modules.

Simplify the workflow that currently is:

1. pick an issue reported in go-swagger
2. find the impacted go-openapi repo
3. PR to fix it, unit test, mention "contribute to go-swagger/go-swager#123)
4. PR to update the dependency in go-swagger, with an integration test to prove the fix, mention "fixes #123"

The process gets even longer with indirect dependencies across several go-openapi repos.

Into:

1. Pick an issue in "core"
2. PR to fix it, with unit and integration test (in different modules), mention "fixes #123"


A structure that would look something like with a new go-openapi/core github repository:

* `github.com/go-openapi/core` - Core go-openapi components
* `github.com/go-openapi/core/docs` - Source of the go-openapi documentation site
 
* **`github.com/go-openapi/core/json`** [go.mod] An immutable, verbatim JSON document and JSON schema
* `github.com/go-openapi/core/json/lexers/default-lexer` - A fast JSON lexer
* `github.com/go-openapi/core/json/types` - type definitions to support JSON
* `github.com/go-openapi/core/json/lexers/contrib` [go.mod] - Contributed alternative parser implementations
* `github.com/go-openapi/core/json/documents/contrib` [go.mod] - Contributed alternative implementations of the json document
* `github.com/go-openapi/core/json/documents/jsonpath` - JSONPath expressions for a JSON document
* `github.com/go-openapi/core/json/stores/default-store` - A memory-efficient store for JSON documents
* `github.com/go-openapi/core/json/stores/contrib` [go.mod] - Contributed alternative implementations of the json document store
* `github.com/go-openapi/core/json/writers/default-writer` - A fast JSON document writer
* `github.com/go-openapi/core/json/writers/contrib` [go.mod] - Contributed alternative implementations of the json document writer

* **`github.com/go-openapi/core/jsonschema`** [go.mod] - JSON schema implementation based on a json document (draft 4 to draft 2020)
* `github.com/go-openapi/core/jsonschema/analyzer` - JSON schema specialized analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/common` - Schema analysis techniques common to all analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/common/canonical` - A unique canonical representation of a JSON schema
* `github.com/go-openapi/core/jsonschema/analyzer/validation` - An analyzer dedicated to producing efficient validators
* `github.com/go-openapi/core/jsonschema/analyzer/generation` - An analyzer dedicated to producing idiomatic generated targets
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets - Target-specific analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets/golang` - An analyzer dedicated to producing clean go
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets/protobuf` [go.mod] - An analyzer dedicated to producing correct and efficient protobuf
* `github.com/go-openapi/core/jsonschema/validator` - A runtime validator based on analysis, as closure
  
* **`github.com/go-openapi/core/genmodels`** [go.mod]  - Data structures code generation from a JSON schema specification
* `github.com/go-openapi/core/genmodels/cmd/genmodels` - A minimal CLI to generate models from JSON schema
* `github.com/go-openapi/core/genmodels/generator` - Models generator
* `github.com/go-openapi/core/genmodels/generator/contrib` - Pluggable supporting code for contributed models generator
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates` - Model templates
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates/contrib` - Contributed templates for models
* `github.com/go-openapi/core/genmodels/generator/targets/golang/settings` - Language-specific generation settings for golang
  
* **`github.com/go-openapi/core/genapi`** [go.mod] - API code generation for operation handlers, client SDK
* `github.com/go-openapi/core/genapi/cmd/genapi` - A minimal CLI to generate API components, excluding models
* `github.com/go-openapi/core/genapi/generator` - API generator
* **`github.com/go-openapi/core/genapi/generator/contrib`** [go.mod] - Pluggable supporting code for contributed API components
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/client` - Client SDK templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/server` - Server templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/contrib` - Contributed templates for API components
* `github.com/go-openapi/core/genapi/generator/targets/golang/settings` - Language-specific generation settings for golang

* **`github.com/go-openapi/core/genspec`** [go.mod] - OpenAPI spec generation from source code
* `github.com/go-openapi/core/genspec/cmd/genspec` - A minimal CLI to generate spec from source
* `github.com/go-openapi/core/genspec/cmd/scanner` - The code scanner and implementation of source annotations

* **`github.com/go-openapi/core/spec`** [go.mod] - JSON document for OpenAPI v2, v3.x implementation
* `github.com/go-openapi/core/spec/analyzer` - OpenAPI spec analyzer for code generation
* `github.com/go-openapi/core/spec/validator` - OpenAPI spec validator
* `github.com/go-openapi/core/spec/linter` - OpenAPI spec linter
* `github.com/go-openapi/core/spec/mixer` - OpenAPI spec merger (mixin)
* `github.com/go-openapi/core/spec/differ` - OpenAPI spec diff
* `github.com/go-openapi/core/spec/cmd/linter` - A minimal CLI frontend for the OpenAPI spec linter

* `github.com/go-openapi/core/errors` [go.mod] - A common error type for go-openapi repositories
 
* **`github.com/go-openapi/core/runtime`** [go.mod] - Runtime components to run untyped or generated APIs
* `github.com/go-openapi/core/runtime/client` [go.mod] - Runtime API client library
* `github.com/go-openapi/core/runtime/server` [go.mod] - Runtime API server library
* **`github.com/go-openapi/core/runtime/producers`** [go.mod] - Response media producers
* **`github.com/go-openapi/core/runtime/producers/contrib`** [go.mod] - Contributed extra response media producers
* **`github.com/go-openapi/core/runtime/consumers`** [go.mod] - Request media consumers
* **`github.com/go-openapi/core/runtime/consumers/contrib`** [go.mod] - Contributed extra request media consumers
* `github.com/go-openapi/core/runtime/middleware` [go.mod] - Collection of middleware
* `github.com/go-openapi/core/runtime/middleware/server` [go.mod] - Collection of middleware for an API server 
* `github.com/go-openapi/core/runtime/middleware/client` [go.mod] - Collection of middleware for an API client
 
* `github.com/go-openapi/core/templates-repo` [go.mod] - Templates repository for code generation
  
* **`github.com/go-openapi/core/strfmt`** [go.mod] - Types to support string formats
* **`github.com/go-openapi/core/strfmt/goswagger-formats`** [go.mod] - Extra formats provided with go-swagger
* **`github.com/go-openapi/core/strfmt/bson-formats`** [go.mod] - Standard string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/goswagger-bson-formats`** [go.mod] - Extra string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/contrib-formats`** [go.mod] - Extra formats contributed
  
* `github.com/go-openapi/core/swag` - A bag of utilities for swagger
* `github.com/go-openapi/core/swag/conv`
* `github.com/go-openapi/core/swag/mangling` - Name mangling to support code generation
* `github.com/go-openapi/core/swag/stringutils`
* `github.com/go-openapi/core/swag/yamlutils` - Utilities to deal with YAML
* `github.com/go-openapi/core/swag/cli` - Utilities to deal with command line interfaces
* `github.com/go-openapi/core/swag/jsonutils/adapters` - JSON adapters to plug in different JSON libraries at runtime

* **`github.com/go-openapi/core/validate`** [go.mod] - Data validation helpers

* **`github.com/go-openapi/core/plugin`** [go.mod] - Pluggable feature facility based on hashicorp plugin

Perhaps move that one to go-openapi?

* **`github.com/go-swagger/go-swagger/`** [go.mod] -- CLI front-end to go-openapi features
* `github.com/go-swagger/go-swagger/cmd/swagger`
* `github.com/go-swagger/go-swagger/cmd/swaggerui` - A graphical client for go-swagger
* **`github.com/go-swagger/go-swagger/examples`** [go.mod] - Examples

## New components to support JSON

The main design goal are:

* to be able to unmarshal and marshal JSON verbatim
* to support JSON even where native types don't, e.g. non-UTF-8 strings, numbers that overflow float64, null value
* to be able (on option) to keep an accurate track of the context of an error
* to drastically reduce the memory needed to process large JSON documents (i.e. ~< 4 GB)
* to limit garbage collection to a minimum viable level.

### [JSON lexer / parser](./json-lexer.md)


### [JSON document and node](./json-document.md)

A document can be walked over and navigated through:

* object keys and array elements can be iterated over
* json pointers can be resolved (json schema `$ref` semantics are not known at this level) efficiently (e.g. constant time or o(log n))
* desirable but not required: should lookup a key efficiently (e.g. constant time or o(log n))

> NOTE: this supersedes features previously provided by `jsonpointer`, `swag`

Options:
* a document may keep the parsing context of its nodes, to report higher-level errors
* a document may support various tuning options regarding how best to store things (e.g. reduce numbers as soon as possible, compress large strings, etc)
* ...

#### How about YAML documents?

We keep the current concept that any YAML should be translated to JSON before being processed.

What about "verbatim" then? YAML is just too complex for me to drift into such details right now.

Possible extensions:

* the basic requirement on a JSON document is to support MarshalJSON and UnmarshalJSON, provide a builder side-type to support mutation
* more advanced use-case may be supported by additional (possibly composed) types
* examples:
  * JSON document with JSONPath query support
  * JSON document with other Marshal or Unmarshal targets, such as MarshalBSON (to store in MongoDB), MarshalJSONB (to store directly into postgres), ...

To avoid undue propagation of dependencies to external stuff like DB drivers etc, these should come as an independant module.

There is a contrib module to absorb novel ideas and experimentations without breaking anything.

#### [JSON document store](./json-store.md)

A document store organizes a few blocks of memory to store (i) interned strings and (ii) JSON scalar values packed in memory.

#### [JSON schema](./json-schema.md)

The JSON schema type extends the JSON document. This type supports all published JSON schema versions (from draft 4 to draft 2020).

Hooks are a mechanism to customize how a schema is built. This is used for example to derive the OpenAPI definition of a schema from the standard JSON schema.

```go
type Hook func(*schemaHooks)

type schemaHookFunc func(s *Schema) error
type schemaHooks struct {
  beforeKey,afterKey,beforeElem,afterElem,beforeValidate,afterValidate  schemaHookFunc
  setErr func(s *Schema, error) // allows to hook an error on the parent schema 
}
```

Schema analyzers:
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
 
## Model generation

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
  T
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
    * pluggable extensions to the model generation logic (i.e. routes processing logic to plugin whenever some `x-*`extension is found
    * contributed templates to support pluggable extensions

Out of this scope: generating schema from JSON data (the classical nodeJS basic tool)?

## API generation

Same modular and extensible approach as the one described for models.

Features:
* handlers generation
* supporting files generation
  * main server
  * initialization

* client generation 

## Spec generation

I have to reflect on that one. I am out of ideas for now.

