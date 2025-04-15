# A mono-repo layout

One git repo, many modules.

> Yes I am aware: the go team is not really fan of mono-repos.
> 
> However, many large or very large projects (such as go-opentelemetry) have adopted this kind of layout.

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
* `github.com/go-openapi/core/docs` - Source of the go-openapi documentation site

## JSON

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

## JSON schema

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
* `github.com/go-openapi/core/jsonschema/converter` - JSON schema version converters

## Model code generation

* **`github.com/go-openapi/core/genmodels`** [go.mod]  - Data structures code generation from a JSON schema specification
* `github.com/go-openapi/core/genmodels/cmd/genmodels` - A minimal CLI to generate models from JSON schema
* `github.com/go-openapi/core/genmodels/generator` - Models generator
* `github.com/go-openapi/core/genmodels/generator/contrib` - Pluggable supporting code for contributed models generator
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates` - Model templates
* `github.com/go-openapi/core/genmodels/generator/targets/golang/templates/contrib` - Contributed templates for models
* `github.com/go-openapi/core/genmodels/generator/targets/golang/settings` - Language-specific generation settings for golang

## API code generation

* **`github.com/go-openapi/core/genapi`** [go.mod] - API code generation for operation handlers, client SDK
* `github.com/go-openapi/core/genapi/cmd/genapi` - A minimal CLI to generate API components, excluding models
* `github.com/go-openapi/core/genapi/generator` - API generator
* **`github.com/go-openapi/core/genapi/generator/contrib`** [go.mod] - Pluggable supporting code for contributed API components
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/client` - Client SDK templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/server` - Server templates
* `github.com/go-openapi/core/genapi/generator/targets/golang/templates/contrib` - Contributed templates for API components
* `github.com/go-openapi/core/genapi/generator/targets/golang/settings` - Language-specific generation settings for golang

## Spec generation from code

* **`github.com/go-openapi/core/genspec`** [go.mod] - OpenAPI spec generation from source code
* `github.com/go-openapi/core/genspec/cmd/genspec` - A minimal CLI to generate spec from source
* `github.com/go-openapi/core/genspec/cmd/scanner` - The code scanner and implementation of source annotations

## OpenAPI spec (v2, v3.x)

* **`github.com/go-openapi/core/spec`** [go.mod] - JSON document for OpenAPI v2, v3.x implementation
* `github.com/go-openapi/core/spec/analyzer` - OpenAPI spec analyzer for code generation
* `github.com/go-openapi/core/spec/validator` - OpenAPI spec validator
* `github.com/go-openapi/core/spec/linter` - OpenAPI spec linter
* `github.com/go-openapi/core/spec/mixer` - OpenAPI spec merger (mixin)
* `github.com/go-openapi/core/spec/differ` - OpenAPI spec diff
* `github.com/go-openapi/core/spec/cmd/linter` - A minimal CLI frontend for the OpenAPI spec linter

## errors

* `github.com/go-openapi/core/errors` [go.mod] - A common error type for go-openapi repositories

## API runtime

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

* **`github.com/go-openapi/core/strfmt`** [go.mod] - Types to support string formats
* **`github.com/go-openapi/core/strfmt/goswagger-formats`** [go.mod] - Extra formats provided with go-swagger
* **`github.com/go-openapi/core/strfmt/bson-formats`** [go.mod] - Standard string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/goswagger-bson-formats`** [go.mod] - Extra string format types extended with BSON support
* **`github.com/go-openapi/core/strfmt/contrib-formats`** [go.mod] - Extra formats contributed

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

Perhaps move that one to go-openapi?

## go-swagger CLI

* **`github.com/go-swagger/go-swagger/`** [go.mod] -- CLI front-end to go-openapi features
* `github.com/go-swagger/go-swagger/cmd/swagger`
* `github.com/go-swagger/go-swagger/cmd/swaggerui` - A graphical client for go-swagger
* **`github.com/go-swagger/go-swagger/examples`** [go.mod] - Examples
