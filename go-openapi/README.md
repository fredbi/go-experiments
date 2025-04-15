# Musings with go-openapi

> TODO(fredbi): organize this doc in smaller pages and publish it on github pages.

This year (2025), the `go-openapi` initiative is 10 years old...

This paper is about the go-openapi and go-swagger projects, which have been painstakingly maintained over this period by a bunch of golang enthusiasts.
It reflects my personal views and these only, not the views of the community of maintainers.

* https://github.com/go-openapi
* https://github.com/go-swagger/go-swagger

As a contributor to this project over the past, ahem, 8 years or so, it is time for a retrospective and honest criticism.
Contributing to this vast project has been an exhilarating experience. So many jewels collected, such a vast collection of feedback, questions,
and developer's experience accumulated... It would be a pity just to throw it away or archive for good this trove of knowledge.

> On the other hand, it is just a fact that any major refactoring effort would require a great deal of development, testing, etc.
> Since the days of the project gathering dozens of developers are long gone, deciding to dive in and start such a daunting task is not easy.
>
> My point here is to try and understand how to move forward, somehow reinvigorate the project with fresh views, and perhaps, deliver to the
> golang community better tooling to produce nice APIs.

The main and most awaited feature is support for OpenAPI v3 (aka OAIv3) or even v4.
But many Bothans died trying to do so... it is a _huge_ amount of work ahead. Few people realize how large it is.
It requires careful design and planning. The difference is that we've 10 years of accumulated feedback and figuring out different designs.

In the sections below, I am giving a more comprehensive analysis. Even though OAIv3 is the ultimate goal, it is not the only one and many micro-designs need
to be questioned, reviewed, and implemented along the way.

## go-swagger and the go-openapi toolkit status in a nutshell

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
| strfmt        | Types to support swagger "format"s | issue with dependencies (e.g. MongoDB) | Y | major split |
| stubs         | A test cases/examples generator    | never completed | Y | rewrite |
| swag          | A bag of tools           | being split | move mangling away, reduce API surface |
| swaggersocket | An example demonstrating streaming | not maintained | N | new repo to collect examples |
| validate      | A json schema and spec validator  | only technical changes recently | N | major split, change concept | 

## Core design analysis

### toolkit not framework

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

#### Why would I want to keep the same kompass?

Simply because API land is too vast to be covered by a framework.
There are so many use cases, so many opinions, so many ways to achieve similar things.
Even though OpenAPI was created to unify the practice of building a REST-like API, the field remains largely open.

I like the idea of a toolkit that simplifies a few things (boring things, highly specialized or technical things) but leaves it up
to developers to decide about their design.

#### Approach to refactoring

There are a lot of jewels here and there. What I'd like to do is to split them apart.
I am currently working at "salvaging" a few such nice bits of code by refactoring actions.
It is a humble improvement, but hopefully it could help in a future more ambitious endeavor.

Since keeping backward compatibility is important, this is primarily achieved by moving various sub-packages into modules.
This would eventually lead to either new go-openapi repos or muti-modules mono-repos with the existing ones (ex: `swag` or `strfmt`).

This is a non-destructive action to preserve and maintain some features. This is no silver bullet and nothing major will ever come from this.

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

### Handling a swagger specification

This is essentially supported by the `spec` package.

`spec` exposes the `Spec` type, which knows how to unmarshal and marshal a JSON swagger v2 spec. That's it.
The `$ref` resolution feature is inside, so the package knows how to fetch remote documents over the network if needed.

Alongside `$ref` resolution comes the `Expand` feature, which expands all `$ref`. A valid use case is the validation of a spec or schema.

### Unfinished jobs / unsupported use cases

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

## repo by repo analysis

Trying to be honest here: a little self-criticism doesn't hurt :).

### go-swagger

> My opinion is very much biased by my personal use of the tool and my favorite areas.
> As a matter of fact, I've been mostly active with improving the schema to model generation,
> and by and large I've essentially worked on the code generation side of this tool.
> Not so much on the other way around: generating spec from code, which I have never personally used
> and have barely contributed to.

Let's start with some positive assessment. I see many major achievements:

* Supports _almost_ everything dealing with OAIv2, only fringe use cases remain poorly supported
* Distributed for many platforms as binary releases and docker images
* Generated code is not too bad, it is way better than anything you may get from swagger-codegen
* Spec generation is still a popular use case, and is unique (afaik) in the golang space
* Once well understood the tool is highly customizable
* A lot of examples are provided
* There are a lot of useful features like spec validation, mixin, diffing, flattening, ...
* Model generation brings a lot of features like embedding custom types, etc
* Code is well-covered by a lot of unit tests, there are also a lot of integration tests
* A large documentation set is available

The main issue is the lack of OAIv3 support. **That's a killer one.**

Unfortunately, my response to that one is kind of diluted all over this note.
To move up to that stage, there are so many things that need to be reviewed and improved.

As a maintainer and user, I've hit the following issues and may raise quite a few (constructive) criticisms:
* CLI
  1. the main intent of the tool is to expose a CLI user interface, but there are too many built-in features
  2. the tool is highly customizable, but documentation about how to do so is scarce
  3. should improve the dockerized use case
  4. bloated/undocumented options
 
* codegen
  1. Many smart things that have been polished after countless bug fixes and edge case reports should be factorized,
     e.g. name mangling subtleties reinjected into `swag/mangling`
  2. The template repo is another smart thing that should be factored out (and deserves to be maintained/improved/enriched in its own right)
  3. More `contrib` codegen alternatives should be added. The `Stratoscale` approach was great but came out as a single shot, unfortunately.
     Perhaps it is just too hard currently to deal with custom templating.
  4. The analysis of swagger types is a bit confusing: using `analysis` was the goal, but instead we get a type resolver that is super-difficult to follow
  5. Model construction is super hard to follow. Almost nobody dares to mess with that part nowadays.
  6. generated code is very difficult to test and test code coverage is not very significant (regarding templates).
     We test "expected generated code" and only on a few occasions do we test the actual behavior of the generated program.
     All this testing is fine but eventually, it made the product more and more rigid as testing against expected generated statements generates a lot
     of impacts on tests even with minor changes in templates
  8. Hit hard limits of the design when it comes to addressing fringe/edge cases such as:
       * polymorphism (only works with vanilla models, would require heavy rework to cover all cases),
       * using fewer pointers (you currently don't really have a choice)
       * distinguish empty vs null vs zero value (which is not a natural way of approaching things in go, and rightly so)
       * advance use cases with `allOf`
       * inability to evolve (with a reasonable effort) toward supporting `anyOf` or `oneOf`
  9. Confusing pointer/not pointer conventions as it depends on the type: parameters are nullable whenever not required (so you may check whether they were present or not)
     but schema types are nullable whenever required (because validation is carried out after unmarshaling).
     It gets even more complicated when playing around with a few available options like omitempty and x-nullable.
 10. the mutability of schemas produces subtle bugs and prevents the parallelizing spec processing.

* code scan
  1. The main hypothesis is that code compiles so that we may analyze an AST. So far so good, but said AST is internally highly dependent on the go version used.
     Even though the introduction of "go toolchain" versioning significantly improved things, this is still fragile.
  2. The other main design of that part is that a lot of information is passed through formatted comments. Maybe too much as a matter of fact.
  3. Comment parsing relies on regexp'es and is (very) difficult to follow.
  4. It is very difficult to test
  5. It is poorly tested and test code coverage is not significant at all
 
* releases
  1. The pace of releases has slowed down to almost a halt, perhaps once a year.

#### What would I like to do?

* Move the tool to just a CLI that coordinates stuff: most features should be externalized to independent modules that could be used as standalone libraries
* Modernize config management (e.g using `koanf`)
* Modernize CLI (e.g. using `cobra` on top of `koanf`)
* Externalize the templates repo feature
* Externalize the code scan feature
* Externalize the model generation feature, with its templates
* Externalize the client/server feature, with its templates
* Support pluggable features using the portable plugin design from `hashicorp`

* New features that I'd like
  * Support OAI v3.x (see below)
  * Still support OAI v2
  * Stop being shy and expose (propose?) extensions to the standard
  * Generated server supporting streams out-of-the-box (likely with an opinionated view)
  * Produce a GUI with a more interactive experience like "try and generate a small bit so I can see if I like the result" (e.g. using `fyne-io`)
  * Introduce rule-based configuration to inject custom tags based on rules (e.g. pattern matching)
  * Add the ability to generate protobuf from jsonschema (I know this is already covered somewhere else, I'd just like to make it better if I can)
  * Add the ability to generate `grpc` and `nats` servers or clients based on this protobuf
  * Expose new useful tools like extracting schemas from a database, or better API documentation tools

* Doc: support versioned documentation
* Releases: modernize with `goreleaser`

### analysis

* spec flattening (i.e. bundling remote schema documents into a single root document) is very complex. Much more than it should be at least
> spec flattening started by introducing a lot of other transforms (like renaming things, and reorganizing complex things),
>  which were unrelated to _just_ bundling remote `$ref` in a single document.
>  So we introduced the concept of "minimal flattening" to do just that (yeah it makes things more complicated to explain)...

* the analyzer is actually not used a lot by go-swagger, not as much as it should at least

> In particular go-swagger largely resorts to its internal type resolver rather than to schema analysis, so we have that kind
> of more or less duplicated features.

#### What would I like to do?

I think the entire feature is too dependent on many historical design choices which I'd like to change (handling JSON, handling $ref's, etc).
I don't think we will ever push a "v2" of that one.

Therefore, the eventual fate of this repo will likely be a "github archive".

The intent (that is, to analyze swagger stuff and make it more palatable to other tools) remains totally valid.
However, I would like to reason about analysis in a more focused way. For example, I'd make different, more specialized _analyzers_ to support different intents:

* analyzing a source spec (or schema) to describe its content in a detailed way (e.g. more like an AST)
  * examples: listing the properties of an object, listing the endpoints of an API etc, like this package does now
* analyzing a source schema in a detailed way with the intent to produce a _validator_ (dynamically or as generated code)
  * need a more thorough inspection (not currently existing) about json schema, such as redundant validations, impossible validations, naturally mutually exclusive `anyOf`, incompatible `allOf`, etc.
  * find an optimized path for validations (e.g. prioritizing things that are likely to fail earlier, e.g. enum check before others, stricter or faster checks come first, etc.
  * find factorizations for validations (e.g. all schemas that run the same validations, even if formally distinct entities, should be processed with just one function)
* analyzing a source spec (or schema) with the intent to generate code for a given target
  * this is specialized by target: generating a protobuf doesn't require the same level of understanding of a schema as generating for golang (or another language)
  * bring more details about the source, pick additional meta-data (e.g. `x-go-*` annotations or other ways of annotating a json schema),
    possibly injected by developers, so as to deliver the most accurate possible description of what should be generated
  * examples: this is where it becomes important to know if a given data item is nullable, if the zero value is valid etc. These checks are currently mostly handled in `go-swagger` and not by `analysis`

The `diff` feature is very useful and should be kept. Perhaps we could make it less verbose and spec-version dependent. Not sure yet how to improve that part.

### errors

I think this repo is pretty much reusable. I has remained maintained to keep up with go advances in error handling.

Perhaps a few things should be handed over to different repos, though. I see that one more as generic error type than it is now (handling validation errors etc). Not sure yet.

Likely a "v2" with a reduced scope and API surface.

### inflect

Likely a github archive.

If it eventually proves useful (but it didn't make it over the past decade ...), re-insource with the name mangling feature (`swag/mangling`).

### jsonpointer & jsonreference

Not in line with the new design, which predates entirely the feature set.

Likely a github archive.

### loads

Likely a "v2" with a reduced API surface. I just want a flexible loader to load JSON or YAML documents. The provided one does the job.

I don't want all the features currently onboarded in the `loads.Spec` struct, such as `Pristine`, `Analyzed` etc.

They are mostly confusing users and not really used appropriately.

### runtime

`runtime` exposes a lot of features. Way too many, in fact.
So we have:

* a client runtime that holds most of the common stuff used by generated SDK clients
  * it's loaded with features and the `request` object is super complex
  * options to configure TLS, and a few other things
* a few interesting server middlewares, such as serving spec or a spec UI (redoc, swagger UI)
* a (monolithic) server middleware that takes care of a lot of things when serving a swagger API:
  * the server is aware of the swagger specs and takes decisions based on the swagger document. There is no generated code in there.
  * content negotiation
  * routing (the runtime brings in its own bespoke `denco` router, very efficient btw)
  * parsing request parameters
  * request pre-preprocessing before calling API handlers (e.g. from generated code or using the "untyped API" feature)
  * response handling to produce the final output (e.g. text, json, yaml...)
* probably a few other things that I forgot

The central issue in my opinion is the **monolithic** aspect of this component. Middleware should assemble as a stack of small focused features.

Ideally, developers should be able to select which ones they prefer.
But this was difficult to achieve with this architecture, so the "API" interface is a big middleware and it may be slightly amended by injecting additional middlewares.

It all stems from decisions about what feature is handed over to generated code (e.g. data validation) and what feature is handled dynamically at runtime (e.g. security, routing).

I think we should make it so that developers have more options to make their own decisions about this.

In an ideal design, we could imagine for instance that:

* validation could be something that is carried out dynamically, or not.
* routing could be something that comes from generated code (using a developer's favorite framework), or not
* middlewares could be added or replaced in the usual idiomatic way of scaffolding an http server, not using some out-of-band configuration of the API middleware

A few issues:

* dealing with TLS config was initially a good idea, but the configuration stuff remains taxing
* bringing in tracing (OTEL support) was a good idea, but it causes problems with all the required dependencies
* can't choose the router (standard library, or other stuff widely available)
* can't choose to code gen routes over dynamically resolving them (like it is now): we should leave it to developer whether they prefer to have their route building stuff exposed
  using their favorite framework (possibly generated code) or to have this part hidden by a smart middleware that knows about the spec
* dependency injection should be made simpler (e.g. using request context). At this moment, the way to inject, say, a database connection pool is kind of awkward.


#### What would I like to do?

In short, I'd like to make this masterpiece of engineering simpler to use and to give more visibility to the many nice features that are built inside.

I'd also like to be more modular and its parts reusable, letting people take their own decisions about how they implement their API.

Well, to revive the toolkit spirit here.

Start with some heavy refactoring to split a few things as their own modules like:

* middlewares (UI, serve doc, etc)
* content negotiation and keep-alive as their own middleware
* separately defined producers and consumers, so many more could be contributed
* tracing as a pluggable feature, no longer imposing dependencies
* figure a way to be able to inject a router or already prepared routes
* client-side upload file feature
* modernized options setting for client (including TLS)
* modernized methods to use `context.Context`, deprecate or make optional client timeout settings
* figure a way to be able to inject client middleware
* figure a way to inject the outcome of the content negotiation more workable by API handlers (it's really a pain to implement as for now)
* find a way to split the security layer, so as to make it pluggable
* modernize logging, with the possibility to inject external loggers (such as `zap`, `zerolog` etc)

### spec & spec3

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

### strfmt
### stubs

A test fixtures and examples generator, driven by a spec.

Never really worked.

**What I would like to do?**
1. Generating valid JSON data from a schema
2. Generating _invalid_ JSON data from a schema, with every validation subject to a failing case, so I can test validators
3. Generating valid examples from a schema. Examples differ from test because they should be somehow representative of the content (test-only fixtures may be quite formal validations)
4. Generating parameter and response, response header examples in a OAI spec, possibly offering to inject these in the spec (by way of mixin)
5. Constructing fake random JSON schemas that are grammatically valid (use to test schema processing & generation - with a kind of fuzzing approach)
6. Same for swagger specs
   


What do we need to do that?

* the original stubs repo relied on `github.com/Pallinder/go-randomdata` (not updated in 6y) and `github.com/zach-klippenstein/goregen` (10 y)
* however goregen remains super interesting in its approach to simplify regexps. Faking output that validates a "pattern" or "patternproperties" clause remains challenging
* go-randomdata is probably deprecated.
  * more maintained contenders are `https://github.com/go-faker/faker` (I used that one), https://github.com/jaswdr/faker, 
  * https://github.com/brianvoe/gofakeit (I've used this one): that's probably the most complete fake available for go (and it supports the "generate from regexp" pattern)
 
For (1) and (2):
* analyze the schema for validation
* iterate over paths just like for validating JSON data
* in lieu of "validating" some input data, drill down to elementary types (including null)
   * the current JSON node (hence a unitary property) gets a collection of validations to pass (including type)
   * and we know from the analysis whether they are possible or not

Now for (1) :
* if string, if number, etc we may generate an appropriate type
* if validation is pattern we have the fakeit.Regex() to help
* iterate over elements or objects, now considering validations such as minProperties, maxItems, unique etc.
* for objects, keys are picked among properties, if not required toss a coin to add them
* if additionalProperties, toss a coin to add some
* patternProperties: use Regexp to generate keys

 Now for (2) :
 * plan to generate exactly one failing case per validation, so we may have a lot of fixtures
 * so basically we start with a valid test case from (1) (will all properties, none missing), then we iterate the json object
   retrieve the associated validations for this path from the analyzed schema
   iterate validations, building one failure at a time

Examples (3) is most challenging, in the absence of any guidance from the dev (some packages out there use tags to do that, e.g. x-example: internet.email)

Is it possible to guess something using a language model? I mean, given a schema, a field and it's associated description, title, whatever context we may find,
* couldn't we just guess that a field called "email" or with "format": "email" should in fact contain an email?
* I am not sure this applies to numbers, though. But ok, field "age" shouldn't be examplified with something like "999999"...
  
Here are a few experimental works that I've found:
* https://github.com/vblagoje/openapi-llm (python)- not really what I am looking for, but the bridging with LLM is interesting
* I believe that this python lib should be able to classify a schema or a spec https://mirascope.com/tutorials/more_advanced/named_entity_recognition/

Once this is done, you have say a "class" associated to each field successfully matched and with this class we may find a good faker.

### swag
### validate

This repo mixes 3 different use cases:
1. JSON schema validation
2. swagger v2 spec validation (which relies on JSON schema validation, plus a number of rules)
3. validation helpers to be used by generated code

This repo is deeply tied to the dynamic JSON approach. We may reuse tests, validation helpers, but not much more.

## Fresh ideas to move forward

1. Acknowledge the fact that json schema is perhaps not the best choice to accurately specify a serialization format.
   json schema was designed as a way to express validation constraints, not to specify explicitly a data format.
   The json schema approach is to derive a data format _implicitly_ from the series of validation constraints that are applied.
   In practice, constraints are most of the time strong enough to imply a given data type such as a `struct` type. But not always.
   Think about how you would convert a json schema into a proto buf spec for example. It is harder than it seems.
   In practice, this means that we should allow much more "intent metadata/desired target" input to flow from the developer, which are not part of the spec.
   This could be configured as rules or as ad-hoc `x-go-*` extensions.
3. Abandon the "dynamic JSON" go way of doing things: work with opaque JSON documents that you can navigate (similar to what `gopkg.in/yaml.v3` does)
   * Stop exposing data objects with exported fields. This has proved to be difficult (or impossible) to maintain over the long run
   * Stop using go maps to represent objects. The unordered, non-deterministic property of go maps makes them inappropriate for code, spec, or doc generation
   * Stop using external libraries for the inner workings. easyjson is just fine, but it is not so hard to work out a similar lexing/writing concept, more suited to our
     JSON spec and schema representation.
     At the core of the new design, we have a fast JSON parser and document node hierarchy with "verbatim" reproducibility as targets.
     Documents may share an in-memory "document storage" to function as a cache for remote documents and quick `$ref` resolution.
     These base components are highly reusable in very different contexts than APIs (e.g. from json linter to json stream parsing).
4. Abandon the json standard library for all internal codegen/validation/spec gen use cases: use a bespoke json parser to build such JSON document objects
5. Separate entirely JSON schema support from OpenAPI spec
   * JSON schema parsing, analysis and validation support independent use cases (beyond APIs): they should be maintainable separately
   * Similarly, the jsonschema models codegen feature should be usable independently (i.e. no need for an OAI spec)
   * JSON schema support is focused on supporting the various versions and flavors of json schema
6. Since we no longer need to represent a spec or a JSON schema with native go types, we are confident that we can model these with full accuracy, including nulls, huge numbers, etc.
   Speed is not really an issue at this stage. Accuracy is.
   Spec analysis should produce a structure akin to an AST, which focuses on describing the source JSON schema structure well, never mixing with requirements from the target codegen language.
   Example: when analyzing, we should not express a situation like "needs a pointer" or "need to be exported", but rather "can be null", or "can't be null".
8. Abandon recursive-like, self-described JSON validation. Elegance has been kind of delusional here and ended up with a memory-hungry, slow JSON validator that could only be used
   for spec validation and not (realistically) when serving untyped API.
   * Proposed design: after schema analysis ("compile" time) we may produce a closure that more or less wires the validation path efficiently.
   * Proposed design (meta-schema or spec validation): most of the validation work is done during the analysis stage, which would also raise warnings about legit but unworkable json schema constructs
9. Analysis should distinguish 2 very different objectives:
    * analysis with the intent to produce a validator
      * for instance, this analysis doesn't need to retain metadata (descriptions, `$comment`), it doesn't need to retain the `$ref` structure, ...
      * this analysis is focused on producing the most efficient path for validation: it may short-circuit redundant validations, pick the best fail-fast path etc
      * at this stage there is no validator yet: one outcome could be a validator closure, and another one could be a generated validation function
    * analysis with the intent to produce code (requires a much deeper understanding to capture the intent of the schema)
      * at this stage, we may introduce a target-specific analysis. For example, it may be a good idea to know if the "zero value" (a very go-ish idea) is valid to take the right decision.
      * it becomes also interesting to infer the _intent_ that comes with some spec patterns
10. Codegen should be very explicit about what pertains to the analysis of the source spec (e.g. "this is a polymorphic type", and what pertains to the solution in the target language
    (e.g. "this is an interface type"). This would make it easier to support alternative solutions to represent the same specification (in this example, the developer might find generated
    interface types awkward to use and would prefer a composition of concrete types)

### Mono-repo layout

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

### New components to support JSON

The main design goal are:

* to be able to unmarshal and marshal JSON verbatim
* to support JSON even where native types don't, e.g. non-UTF-8 strings, numbers that overflow float64, null value
* to be able (on option) to keep an accurate track of the context of an error
* to drastically reduce the memory needed to process large JSON documents (i.e. ~< 4 GB)
* to limit garbage collection to a minimum viable level.

#### [JSON lexer / parser](./json-lexer.md)


#### [JSON document and node](./json-document.md)

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
    
#### Swagger spec schema

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
 
### Model generation

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
### API generation

Same modular and extensible approach as the one described for models.

Features:
* handlers generation
* supporting files generation
  * main server
  * initialization

* client generation 

