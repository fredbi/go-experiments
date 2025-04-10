# Musings with go-openapi (still very draftish)

The `go-openapi` initiative is now 10 years old...

As a contributor to this project over the past, ahem, 8 years or so, I believe it is time for a retrospective and honest criticism.

Contributing to this vast project has been an exhilarating experience. So many jewels accumulated, such a vast collection of feedbacks and
developper's experience accumulated... It would be such a waste to just throw away or archive for good such an accumulation of knowledge.

On the other hand, it is just a fact that any major refactoring effort would require a great deal of development, testing etc.
Since the days of the project gathering dozens of developers are long gone, deciding to dive in and start such a daunting task is not easy.

My point here is to try and understand how to move forward, somehow reinvigorate the project with fresh views and perhaps, deliver to the
golang community better tooling to produce nice APIs.

## go-swagger and the go-openapi toolkit

A quick keep / change analysis of the various tools that have been introduced.

| Repo          | What did we want?   | Current status | Keep ? | Change ? |
|---------------|---------------------|----------------|--------|----------|
|go-swagger     | A CLI to gen code   |                |  Y     | split    |  
|               | A CLI to gen spec   |                |  Y     | split    |
|               | A CLI to do whatever   |                |  Y     | split    |
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

At the core of the project, we find an approach to deal with JSON in `go`.
Back in 2015, I still believe it was the right approach to start with. Nobody could guess in which direction the core golang team would move regarding json.

The main idea is to consider that JSON data can be mapped into some "dynamic JSON" golang structure, which is basically what you get when the standard
library `encoding/json` unmarshals JSON data into an untyped receiver. Like in:
```go
var r interface{}
_ = json.Unmarshal(jsonBytes, &r)
```

However, after 10 years we 
## repo by repo analysis

Trying to be honest here: a little self-criticism doesn't hurt :).
### 
### errors

