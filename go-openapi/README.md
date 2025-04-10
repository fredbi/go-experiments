# Musings with go-openapi (still very draftish)

This year (2025), the `go-openapi` initiative is now 10 years old...

As a contributor to this project over the past, ahem, 8 years or so, I believe it is time for a retrospective and honest criticism.

Contributing to this vast project has been an exhilarating experience. So many jewels accumulated, such a vast collection of feedback and
developer's experience accumulated... It would be such a waste just to throw away or archive for good such an accumulation of knowledge.

On the other hand, it is just a fact that any major refactoring effort would require a great deal of development, testing, etc.
Since the days of the project gathering dozens of developers are long gone, deciding to dive in and start such a daunting task is not easy.

My point here is to try and understand how to move forward, somehow reinvigorate the project with fresh views, and perhaps, deliver to the
golang community better tooling to produce nice APIs.

The main and most awaited feature is support for OpenAPI v3 (aka OAIv3) or even v4. But many Bothans died trying to do so... it is a _huge_ amount of work ahead.

In the sections below, I am giving a more comprehensive analysis. Even though OAIv3 is the ultimate goal, it is not the only one and many micro-designs need
to be questioned, reviewed, and implemented along the way.

## go-swagger and the go-openapi toolkit

A quick keep/change-like analysis of the various tools that have been introduced.

| Repo          | What did we want?   | Current status | Keep? | Change? |
|---------------|---------------------|----------------|--------|----------|
|go-swagger     | A CLI to gen code   |                |  Y     | split    |  
|               | A CLI to gen spec   |                |  Y     | split    |
|               | A CLI to do whatever   |                |  Y     | split    |
| analysis      | A higher level spec analyzer | barely changed |  Y     | split / change concept  |
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

### toolkit not framework

The `go-openapi` project was designed as a toolkit rather than a framework.

The project provides basic tooling, holding as little opinion as possible,
which allows scaffolding custom solutions based on the OpenAPI standard.

The `go-openapi` repos have never been intended to expose a turnkey, out-of-the-box universal framework.

Well, overall we have kept this compass, even though there are a few departures from this concept, such as:
* the `runtime` component has grown too much into a hard-to-maintain monolithic middleware (see below [runtime](#runtime)
* customizing `go-swagger` is hard
* using custom `format`s (e.g. custom `strfmt` types) is hard

Essentially, there are a lot of jewels here and there. What I'd like to do (and am currently doing on some parts) is to split apart
a lot of those. Since keeping backward compatibility is important, this is primarily achieved by moving various sub-packages into modules.

This would lead to either new go-openapi repos or muti-modules mono-repos with the existing ones (ex: `swag` or `strfmt`).

I am currently working at "salvaging" a few such nice bits of code by refactoring actions. A humble improvement, but hopefully it could help some future,
more ambitious endeavor.

### json

Since `swagger` instates JSON and JSON schema as the lingua franca to describe data serialization (API messages),
the number one challenge was to deal with those.

At the core of the project, we thus find an approach to deal with JSON in `go`.

A major design goal was to rely as much as possible on the go standard library. Unfortunately, the `encoding/json` library has been lagging on so many
subtleties of JSON, not to mention performance issues. 

> Back in 2015, it was most likely the right approach to start with.
> Only a few third-party tools were available and nobody could guess in which direction the core golang team would move regarding json.

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

This approach requires a few add-ons to deal with json schema:
* $ref resolution was the most complex (even for draft v4) and is dealt with by the `spec` repo (not a good idea in retrospect)
* other aspects of $ref were handed over to `jsonpointer` and `jsonreference`. Essentially, the intent is to support `struct`s

A nice outcome of this approach was to be able to rapidly deliver a working API runtime for "untyped" APIs, i.e. stuff is validated and served
dynamically based on the JSON spec, without any generated code. More explanations about this use case may be found [here](https://goswagger.io/go-swagger/tutorial/dynamic/).

Real-world issues we faced with this are intimately bound to the design choices of the go language regarding data structures vs json.

In a nutshell:

* handling `null` JSON values
* checking for json schema `required`
* forced trade-offs when deciding to use pointers rather than values in receiver go types
* couldn't keep up with the creativity of the json schema committee, who have introduced many major evolutions since v4
* couldn't keep up with generalized usage of $ref's in specs, untyped things in unexpected places (such as extension tags, ...)
* dealing with union types (e.g. jsonschema declarations such as `type: [...]`, `anyOf`, `oneOf`, `allOf`, or the swagger v2 peculiar way
 of supporting inheritance.

Rest assured: the toolkit works and covers most "sound" use cases. The issue with this design approach is that it sets a "glass ceiling"
and some niche use cases or edge cases are very hard

### Handling a swagger specification

This is essentially brought by the `spec` package.

### Unfinished jobs / unsupported use cases

In no particular order:

* OAI v3.x support
* spec linting
* test cases and examples generation
* test generation
* language-agnostic / support other target languages
* XML support
* ready-to-use streaming support
* workable multiple MIME type support
* full support for OAIv2 "polymorphic types" (now deprecated in OAIv3)
* ready-to-use authentication middleware
* ...

## repo by repo analysis

Trying to be honest here: a little self-criticism doesn't hurt :).

### go-swagger

> My opinion is very much biased by my personal use of the tool and favorite areas.
> As a matter of fact, I've been mostly active with improving the schema to model generation,
> and by and large very much in the code generation side of this tool. Not so much on
> the other way around: generating spec from code, which I have never personally used and barely contributed to.

Let's start with some positive assessment. I see many major successes:
* ...

The main issue is the lack of OAIv3 support. That's a killer one.
Unfortunately, my response to that one is kind of diluted all over this note.
In order to move there, there are so many things that need to be reviewed and improved.
As a maintainer and user, I've hit the following issues / identified the following criticisms:
* CLI
  1. the main intent of the tool is to provide a CLI, but there are too many built-in features
  2. the tool is highly customizable, but documentation about how to do so is scarce
  3. should improve the dockerized use case
  4. bloated/undocumented options
 
* codegen
  1. Many smart things that have been polished after countless bug fixes and edge case reports should be factorized,
     e.g. name mangling subtleties reinjected into `swag/mangling`
  2. The template repo is another smart thing that should be factored out (and deserves to be maintained/improved/enriched in its own right)
  3. More `contrib` codegen alternatives should be added. The `stratoscale` approach was great, but came out as a single shot, unfortunately.
     Perhaps it is just too hard currently to deal with custom templating.
  4. The analysis of swagger types

* code scan
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

