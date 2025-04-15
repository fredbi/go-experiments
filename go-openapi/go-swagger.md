# Analysis of go-swagger

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
