# analysis


* spec flattening (i.e. bundling remote schema documents into a single root document) is very complex. Much more than it should be at least
> spec flattening started by introducing a lot of other transforms (like renaming things, and reorganizing complex things),
>  which were unrelated to _just_ bundling remote `$ref` in a single document.
>  So we introduced the concept of "minimal flattening" to do just that (yeah it makes things more complicated to explain)...
>  Minimal flattening is the essential feature that enables code generation in the presence of remote $ref.

* the `diff` feature is very useful and should be kept. Perhaps we could make it less verbose and spec-version dependent. Not sure yet how to improve that part.
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


