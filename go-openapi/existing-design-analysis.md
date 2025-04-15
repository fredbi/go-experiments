# Analysis of the existing design (v0.x)

## toolkit not framework

The `go-openapi` project was designed as a toolkit rather than a framework.

The project provides basic tooling, holding as little opinion as possible, which allows scaffolding custom solutions based on the OpenAPI standard.
The `go-openapi` repos have never been intended to expose a turnkey, out-of-the-box universal framework.
Well, overall we have kept this compass, even though there are a few departures from this concept, such as:

* the `runtime` component has grown too much into a hard-to-maintain monolithic middleware (see below [runtime](#runtime))
* customizing `go-swagger` is hard
* using custom `format`s (e.g. custom `strfmt` types) is hard

The main problem I see with this approach, which I've largely supported so far, is that it tends to result in a project where experts talk to experts only.
Many less experienced developers have been struggling with our code base. Sadly, we missed a lot of contributions because of this.

Simplification has become an acute challenge. This is true for our codebase in general, our package APIs as well as our user interfaces (CLI, documentation).

## Why would I want to keep the same compass?

Simply because API land is too vast to be covered by a framework.
There are so many use cases, so many opinions, so many ways to achieve similar things.
Even though OpenAPI was created to unify the practice of building a REST-like API, the field remains largely open.

I like the idea of a toolkit that simplifies a few things (boring things, highly specialized or technical things) but leaves it up
to developers to decide about their design.

## Approach to refactoring

There are a lot of jewels here and there. What I'd like to do is to split them apart.
I am currently working at "salvaging" a few such nice bits of code by refactoring actions.
It is a humble improvement, but hopefully it could help in a future more ambitious endeavor.

Since keeping backward compatibility is important, this is primarily achieved by moving various sub-packages into modules.
This would eventually lead to either new go-openapi repos or muti-modules mono-repos with the existing ones (ex: `swag` or `strfmt`).

This is a non-destructive action to preserve and maintain some features. This is no silver bullet and nothing major will ever come from this.

## Dealing with json in go

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

As a nice outcome, this approach rapidly delivered a working API runtime for "untyped" APIs, i.e. stuff is validated and served
dynamically based on the JSON spec, without any generated code. More explanations about this use case may be found [here](https://goswagger.io/go-swagger/tutorial/dynamic/).

Real-world issues we faced with this are intimately bound to the design choices of the go language regarding data structures vs json.

The central issues in a nutshell:

* handling `null` JSON values
* checking for json schema `required`
* forced trade-offs when deciding to use pointers rather than values in receiver go types
* problems with keeping the order of keys when generating specs, docs, and basically everything where the author intended to see things appear in a given order
* couldn't keep up with the creativity of the json schema committee, who have introduced many major evolutions since v4
* couldn't keep up with generalized usage of `$ref`'s in specs, untyped things in unexpected places (such as extension tags, ...)
* dealing with union types (e.g. JSON schema declarations such as `type: [...]`, `anyOf`, `oneOf`, `allOf`, or the swagger v2 peculiar way
 of supporting inheritance.
* degraded performances when doing anything (including running the generated API) with the standard lib

Rest assured: the toolkit works and covers most "sound" use cases. The issue with this design approach is that it sets a "glass ceiling"
and some niche use cases or edge cases are tough

## Handling a swagger specification

This is essentially supported by the `spec` package.

`spec` exposes the `Spec` type, which knows how to unmarshal and marshal a JSON swagger v2 spec. That's it.
The `$ref` resolution feature is inside, so the package knows how to fetch remote documents over the network if needed.

Alongside `$ref` resolution comes the `Expand` feature, which expands all `$ref`. A valid use case is the validation of a spec or schema.

## Unfinished jobs / unsupported use cases

In no particular order:

* OAI v3.x support
* spec linting
* test cases and examples generation
* test generation
* language-agnostic / support other target languages
* XML support
* ready-to-use streaming support
* workable multiple MIME-type support
* full support for OAIv2 "polymorphic types" (now deprecated in OAIv3)
* ready-to-use authentication middleware
* OAI to grpc
* pluggable features (e.g for string formats)
* ...


