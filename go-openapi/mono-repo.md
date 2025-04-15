# A mono-repo layout

One git repo, many modules.

> Yes I am aware: the go team is not really a fan of mono-repos.
> 
> However, many large or very large projects (such as go-opentelemetry) have adopted this kind of layout.
>
> From my personal experience on this project, I believe this could be a move in the right direction.

## Why?

I've been trying to maintain this fragile mesh of cross-dependencies for a while now.
Managing a simple bug fix that traverses many dependent repos is truly a hell.

Mono-repo would simplify the workflow that currently is:

1. pick an issue reported in go-swagger
2. find the impacted go-openapi repo
3. PR to fix it, unit test, mention "contribute to go-swagger/go-swager#123)
4. PR to update the dependency in go-swagger, with an integration test to prove the fix, mention "fixes #123"

The process gets even longer with indirect dependencies across several go-openapi repos.

Into:

1. Pick an issue in "core"
2. PR to fix it, with unit and integration test (in different modules), mention "fixes #123"

A structure that would look something like with a new go-openapi/core github repository:

## go-openapi core

* `github.com/go-openapi/core` - Core go-openapi components
* `github.com/go-openapi/core/docs` - Source of the go-openapi documentation site (design documents, benchmarks, ...)

## JSON

A memory-efficient implementation of JSON fully compliant with the standard.

* **`github.com/go-openapi/core/json`** [go.mod] An immutable, verbatim JSON document
* `github.com/go-openapi/core/json/lexers/default-lexer` - A fast JSON lexer
* **`github.com/go-openapi/core/json/lexers/default-lexer/examples`** [go.mod] - Sample usage of the lexer
* **`github.com/go-openapi/core/json/lexers/default-lexer/benchmarks`** [go.mod] - Comparative benchmarks for lexing JSON
* `github.com/go-openapi/core/json/interfaces` - public type definitions to support JSON
* `github.com/go-openapi/core/json/internal/nodes` - JSON document internal node
* ~~`github.com/go-openapi/core/json/documents/contrib` [go.mod] - Contributed alternative implementations of the json document~~
* `github.com/go-openapi/core/json/documents` - additional features pluggable to a JSON document
*  `github.com/go-openapi/core/json/documents/jsonpath` - JSONPath expressions for a JSON document
*  `github.com/go-openapi/core/json/documents/hooks` - Hooks to customize the behavior of a JSON document
* **`github.com/go-openapi/core/json/documents/examples`** [go.mod] - Examples on how to work with JSON documents
* `github.com/go-openapi/core/json/stores/default-store` - A memory-efficient store for JSON documents
* **`github.com/go-openapi/core/json/stores/default-store/benchmarks`** [go.mod]  - Comparative benchmarks for packing and storing JSON
* `github.com/go-openapi/core/json/writers/default-writer` - A fast JSON document writer
* **`github.com/go-openapi/core/json/writers/default-writer/benchmarks`** [go.mod] - Comparative benchmarks for writing JSON

* `github.com/go-openapi/core/json/lexers/contrib` [go.mod] - Contributed alternative parser implementations
* `github.com/go-openapi/core/json/writers/contrib` [go.mod] - Contributed alternative implementations of the json document writer
* `github.com/go-openapi/core/json/stores/contrib` [go.mod] - Contributed alternative implementations of the json document store

## JSON schema

An advanced JSON schema analyzer to build efficient validators and generated targets.

* **`github.com/go-openapi/core/jsonschema`** [go.mod] - JSON schema implementation based on a json document (draft 4 to draft 2020)
> not sure it wouldn't be better side by side with json.Document
* `github.com/go-openapi/core/jsonschema/analyzer` - JSON schema specialized analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/common` - Schema analysis techniques common to all analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/common/ast` - Abstract Syntax Tree for schema analysis
* `github.com/go-openapi/core/jsonschema/analyzer/common/canonical` - A unique canonical representation of a JSON schema

* `github.com/go-openapi/core/jsonschema/analyzer/validation` - An analyzer dedicated to producing efficient validators

* `github.com/go-openapi/core/jsonschema/analyzer/generation` - An analyzer dedicated to producing idiomatic generated targets
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets - Target-specific analyzers
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets/golang` - An analyzer dedicated to producing clean go
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets/protobuf` [go.mod] - An analyzer dedicated to producing correct and efficient protobuf
* `github.com/go-openapi/core/jsonschema/analyzer/generation/targets/doc` - An analyzer dedicated to producing documentation
* `github.com/go-openapi/core/jsonschema/validator` - A runtime validator based on analysis, as a closure
* **`github.com/go-openapi/core/jsonschema/validator/benchmarks`** [go.mod] - Comparative benchmarks for JSON schema validators
* `github.com/go-openapi/core/jsonschema/converter` - JSON schema version converter

## Model code generation

A versatile type generation system based on JSON schema.

* **`github.com/go-openapi/core/genmodels`** [go.mod]  - Data structures code generation from a JSON schema specification
* `github.com/go-openapi/core/genmodels/cmd/genmodels` - A minimal CLI to generate models from JSON schema
* `github.com/go-openapi/core/genmodels/generator` - Models generator
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates` - Model templates
* `github.com/go-openapi/core/genmodels/generator/targets/golang/settings` - Language-specific generation settings for golang
* `github.com/go-openapi/core/genmodels/generator/targets/protobuf/templates` - protobuf templates
* `github.com/go-openapi/core/genmodels/generator/targets/protobuf/settings` - protobuf settings
* **`github.com/go-openapi/core/genmodels/generator/contrib`** [go.mod] - Pluggable supporting code for contributed models generator
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates/contrib` - Contributed templates for models

## API code generation

A versatile API generation system based on OpenAPI (v2, v3.x)

* **`github.com/go-openapi/core/genapi`** [go.mod] - API code generation for operation handlers, client SDK
* `github.com/go-openapi/core/genapi/cmd/genapi` - A minimal CLI to generate API components, excluding models
* `github.com/go-openapi/core/genapi/generator` - API generator package
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/client` - Client SDK templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/server` - Server templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/grpc-client` - grpc client templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/grpc-server` - grpc server templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/settings` - Language-specific generation settings for golang
* **`github.com/go-openapi/core/genapi/generator/contrib`** [go.mod] - Pluggable supporting code for contributed API components
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/contrib` - Contributed templates for API components
 
## Spec generation from code

Reverse-engineering of your API code to produce an OpenAPI spec or JSON schema.

* **`github.com/go-openapi/core/genspec`** [go.mod] - OpenAPI spec generation from source code
* `github.com/go-openapi/core/genspec/cmd/genspec` - A minimal CLI to generate spec from source
* `github.com/go-openapi/core/genspec/scanner` - The code scanner and implementation of source annotations
* `github.com/go-openapi/core/genspec/scanner/annotations` - Definition of supported source annotations

## Documentation generation

Static documentation generator from JSON schema and OpenAPI spec.

* **`github.com/go-openapi/core/gendoc`** [go.mod] - OpenAPI doc generation from spec
* `github.com/go-openapi/core/gendoc/cmd/gendoc` - A minimal CLI to generate documentation from spec (or schema)
* `github.com/go-openapi/core/gendoc/generator` - The doc generator package
* **`github.com/go-openapi/core/gendoc/generator/contrib`** [go.mod] - Contributed doc generator pluggable functionality
* `github.com/go-openapi/core/gendoc/generator/targets/markdown/templates` - markdown templates
* `github.com/go-openapi/core/gendoc/generator/targets/markdown/settings` - markdown settings
* `github.com/go-openapi/core/gendoc/generator/targets/hugo/templates` - hugo templates
* `github.com/go-openapi/core/gendoc/generator/targets/hugo/settings` - hugo settings
* `github.com/go-openapi/core/gendoc/generator/targets/jekyll/templates` - jekyll templates
* `github.com/go-openapi/core/gendoc/generator/targets/jekyll/settings` - jekyll settings
* `github.com/go-openapi/core/gendoc/generator/targets/contrib/templates` - contributed documentation templates
* `github.com/go-openapi/core/gendoc/generator/targets/contrib/settings` - contributed documentation settings

## Stubs

Examples and test case generator for JSON schema and OpenAPI spec.

Ships with a python tool to play with against your favorite LLM provider.

* **`github.com/go-openapi/core/stubs`** [go.mod] - A faker utility for JSON schema and OpenAPI spec
* `github.com/go-openapi/core/stubs/cmd/genstubs` - A minimal CLI to generate JSON examples and test cases
* `github.com/go-openapi/core/stubs/generator` - The stubs generator package
* `github.com/go-openapi/core/stubs/generator/contrib`

## Serve spec

* `github.com/go-openapi/core/serve` [go.mod] -- A simple http server to serve specs with UI

## OpenAPI spec (v2, v3.x)

* **`github.com/go-openapi/core/spec`** [go.mod] - JSON document for OpenAPI v2, v3.x implementation
* `github.com/go-openapi/core/spec/analyzer` - OpenAPI spec analyzer for code generation
* `github.com/go-openapi/core/spec/analyzer/ast` -- Abstract Syntax Tree definitions for the spec analyzer
* `github.com/go-openapi/core/spec/validator` - OpenAPI spec validator
* `github.com/go-openapi/core/spec/linter` - OpenAPI spec linter
* `github.com/go-openapi/core/spec/mixer` - OpenAPI spec merger (mixin)
* `github.com/go-openapi/core/spec/differ` - OpenAPI spec diff
* `github.com/go-openapi/core/spec/amender` - OpenAPI rule-based spec rewrite
* `github.com/go-openapi/core/spec/bundler` - OpenAPI spec bundling
* `github.com/go-openapi/core/spec/converter` - OpenAPI spec version converter
* `github.com/go-openapi/core/spec/cmd/oaispec` - A minimal CLI frontend for the OpenAPI spec tool: validate, lint, mixin, diff, convert, bundle, amend

## errors

* `github.com/go-openapi/core/errors` [go.mod] - A common error type for go-openapi repositories

## API runtime

A pluggable runtime environment to run API clients and servers.

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

## templates repo

* `github.com/go-openapi/core/templates-repo` [go.mod] - Templates repository for code generation

## string formats

* **`github.com/go-openapi/core/strfmt`** [go.mod] - Types to support standard string formats
* **`github.com/go-openapi/core/strfmt/bson-formats`** [go.mod] - Standard string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/goswagger-formats`** [go.mod] - Extra formats provided with go-swagger
* **`github.com/go-openapi/core/strfmt/goswagger-bson-formats`** [go.mod] - Extra string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/contrib-formats`** [go.mod] - Extra contributed formats
* **`github.com/go-openapi/core/example`** [go.mod] - Code examples to build a custom format and use it with go-swagger and your API
* 
## swag: a bag of utilities

* `github.com/go-openapi/core/swag` - A bag of utilities for swagger
* `github.com/go-openapi/core/swag/conv`
* `github.com/go-openapi/core/swag/mangling` - Name mangling to support code generation
* `github.com/go-openapi/core/swag/stringutils`
* `github.com/go-openapi/core/swag/yamlutils` - Utilities to deal with YAML
* `github.com/go-openapi/core/swag/cli` - Utilities to deal with command line interfaces
* `github.com/go-openapi/core/swag/jsonutils/adapters` - JSON adapters to plug in different JSON libraries at runtime

## validation helpers
* **`github.com/go-openapi/core/validate`** [go.mod] - Data validation helpers

## plugin support
* **`github.com/go-openapi/core/plugin`** [go.mod] - Pluggable feature facility based on hashicorp plugin

Perhaps move the following one to go-openapi? For now, let's keep the tradition alive.

## go-swagger CLI

A CLI tool that reunites all go-openapi features within a single tool with binary and docker distributions.

* **`github.com/go-swagger/go-swagger/`** [go.mod] -- CLI front-end to go-openapi features
* `github.com/go-swagger/go-swagger/cmd/swagger`
* `github.com/go-swagger/go-swagger/cmd/swaggerui` - A graphical client for go-swagger
* `github.com/go-swagger/go-swagger/plugins` - Pluggable functionality for go-swagger
* **`github.com/go-swagger/go-swagger/examples`** [go.mod] - Examples
