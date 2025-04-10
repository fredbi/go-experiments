# Musings with go-openapi (still very draftish)

This year (2025), the `go-openapi` initiative is now 10 years old...

As a contributor to this project over the past, ahem, 8 years or so, I believe it is time for a retrospective and honest criticism.

Contributing to this vast project has been an exhilarating experience. So many jewels accumulated, such a vast collection of feedback and
developper's experience accumulated... It would be such a waste just to throw away or archive for good such an accumulation of knowledge.

On the other hand, it is just a fact that any major refactoring effort would require a great deal of development, testing, etc.
Since the days of the project gathering dozens of developers are long gone, deciding to dive in and start such a daunting task is not easy.

My point here is to try and understand how to move forward, somehow reinvigorate the project with fresh views, and perhaps, deliver to the
golang community better tooling to produce nice APIs.

## go-swagger and the go-openapi toolkit

A quick keep / change analysis of the various tools that have been introduced.

| Repo          | What did we want?   | Current status | Keep ? | Change ? |
|---------------|---------------------|----------------|--------|----------|
|go-swagger     | A CLI to gen code   |                |  Y     | split    |  
|               | A CLI to gen spec   |                |  Y     | split    |
|               | A CLI to do whatever   |                |  Y     | split    |
| analysis      | A higher level spec analzer | barely changed |  Y     | split / change concept  |
| errors        | A common error type | barely changed |  Y     | improve  |
| inflect       | Pluralize words when generating doc | not used (but in 1 place) | N | could be part of the name mangling stuff in swag |
| jsonpointer   | A way to deal with json pointers in a go struct | barely changed |  need to change concept  | keep for backward compat.  |
| jsonreference | A way to deal with json ref in a go struct | barely changed |  need to change concept  | keep for backward compat.  |
|github-release | A fork to build go-swagger releases | no longer in use | N | adopt modern tooling (e.g. goreleaser) |
| kvstore       | An example to illustrate dependency injection | not maintained | N | new repo to collect examples |
| loads         | A convenient spec loader | | Y | trim down API |
| runtime       | The runtime for client and server | barely changed | Y | major split |
| spec          | A type for swagger specs          | only technical changes recently | N | change concept |
| spec3         | An attempt to support OAI v3      | not complete - given up | N | change concept | 
| strfmt        | Types to support swagger "format"s | issue with dependencies (e.g. mongodb) | Y | major split |
| stubs         | A test cases/examples generator    | never completed | Y | rewrite |
| swag          | A bag of tools           | being split | move mangling away, reduce API surface |
| swaggersocket | An example demonstrating streaming | not maintained | N | new repo to collect examples |
| validate      | A json schema and spec validator  | only technical changes recently | N | major split, change concept | 

## Core design analysis

### json

Since `swagger` instates JSON and JSON schema as the lingua franca to describe data serialization (API messages), number one achievement was to deal with those.

At the core of the project, we thus find an approach to deal with JSON in `go`.

> Back in 2015, it was most likely the right approach to start with.
> Only a few third-party tools were available and nobody could guess in which direction the core golang team would move regarding json.

A major design goal was to rely as much as possible on the go standard library. Unfortunately, the `encoding/json` library has been lagging on so many
subtleties of JSON, not to mention performance issues. 

The main idea is to consider that JSON data can be mapped into some "dynamic JSON" golang structure, which is basically what you get when the standard
library `encoding/json` unmarshals JSON data into an untyped receiver. Like in:

```go
var r interface{}
_ = json.Unmarshal(jsonBytes, &r)
```

This transforms json containers into go maps and slices and maps scalar values to native types (e.g. string, float64).
You don't get any `struct` with this.

This approach has been used to support the schema schema v4 spec.
The `validate` repo contains a json schema validator (draft v4), which is based on this idea.

This approach required a few add-ons to deal with json schema:
* $ref resolution was the most complex (even for draft v4) and is dealt with by the `spec` repo (not a good idea in retrospect)
* other aspects of $ref were handed over to `jsonpointer` and `jsonreference`. Essentially, the intent is to support `struct`s

A nice outcome of this approach was to be able to rapidly deliver a working API runtime for "untyped" APIs, i.e. stuff is validated and served
dynamically based on the JSON spec, without any generated code.

Real-world issues we faced with this are intimately bound to the design choices of the go language regarding data structures vs json.

* handling `null` values
* checking for json schema `required`
* forced trade-offs when deciding to use pointers rather than values in receiver go types
* couldn't keep up with the creativity of the json schema committee, who have introduced many major evolutions since v4
* couldn't keep up with generalized usage of $ref's in specs, untyped things in unexpected places (such as extension tags, ...)

Rest reassured: the toolkit works and cover most "sound" use cases.

## repo by repo analysis

Trying to be honest here: a little self-criticism doesn't hurt :).

### go-swagger

### analysis
### errors
### inflect
### jsonpointer & jsonreference
### loads
### runtime
### spec & spec3
### strfmt
### stubs
### swag
### validate

