# spec and spec3


`spec` exposes the go data types that build up an OpenAPI v2 specification. Essentially, you can marshal and unmarshal JSON bytes.

The hidden complexity here lies in the built-in support for `$ref`. Spec expansion and `$ref` resolution offer a poor interface and little flexibility.
In particular, it is difficult to extend this feature and support the more advanced `$ref` specification that comes with more recent json schema drafts.

Hence:
* **`$ref` should be its own type, with the capability to `Resolve` the location pointed to, and possibly to `Expand` it.**
* **all the base path, ID etc that is currently handled by the ref resolver type should be captured by a context.Context**

So we would have something like:
```go
Ref.Resolve(context.Context) (targetType, error)
```

* the resolution of cycles with `$ref` should not produce a random result

json pointers and `$ref` are tricky to resolve with a strongly typed language (it gets even worse in the `analysis` package).
What we do now is that we know exactly which places accept `$ref`s in a swagger specification or a JSON schema so we anticipate the kind of thing
we can find in the target. It works in most cases, 

 #### A note on `spec3`

`spec3` has been an unfinished attempt to support OpenAPI v3. 
It wanted to solve a few design issues of the original `spec`, namely the key ordering issue.
It relied on easyjson to do this efficiently, and support for `$ref` resolution (`Expand`, `Resolve` etc...) disappeared.

So the new object model is able to marshal and unmarshal an OpenAPIv3 spec, full stop. No validation for jsonschema v5, no `$ref`, no codegen.

The benefit was that it gave a good insight about differences between versions, and found solutions to some problems.

The problem I see with that approach is that it missed the *main* problem, which is, in my opinion, long-term maintainability and evolutivity across versions.

We know now for a fact that the OpenAPI committee has no problem with issuing breaking changes. Even the JSON schema committee may do so from time to time.
What we take for granted as go developers, i.e. the "forever backward compatibility assurance", is a unique and highly distinctive feature of golang.
When dealing with stuff that comes from outside go-land, the ability to absorb external changes while remaining compatible to your users is essential.

Hence:

**data types without any exported field. Only methods.**

So `spec3` is cool, but will it shorten our next cycle when `3.1`, `3.2`, and `4.0` are out? How about json schema (now Draft 2020: we missed five major releases!).
I don't think so.

The `OrderedMap` layout is cool. However, we now have in golang iterators to replace that part.

Removing `$ref` processing and expansion from the scope was a good decision (it is the most complex part of the `spec` package). However, the design of this feature remains an unanswered question there.

Support for JSON pointers (formerly, using the `jsonpointer` package) has been removed too. 

Hence:
**`$ref` resolution (and possibly expansion) should be a feature of the type supporting a JSON schema**.

> It has nothing specific to the OpenAPI spec schema, which inherits from it.

**JSON pointer should be a feature of any type supporting some sort of JSON document**

> JSON pointers are defined by RFC 6901 and are more general than JSON schema.

Still no clear solution for JSON stuff that doesn't fit well with go native types, i.e. `null`, overflowing numerical types, the decimal representation of numerical types ...
Even if these problems are overall minor, we build a situation in which we know in advance that none of these may ever be solved.

The dependency on easyjson is a minor problem after all: the relied-upon components may easily be re-insourced to internal packages.

Likely eventual fate of these repos: github archive.
