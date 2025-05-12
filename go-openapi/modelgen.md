# Model generation

The objective is to generate idiomatic go data structures from a JSON schema.
It should support all JSON schema versions, from Draft 4 to 2020.

When comparing this approach to go-swagger, I'd like:
* to simplify templates a lot: templates should just be there to iterate over the model structure, not to take big design decisions
* to simplify model.go a lot, by deferring most of the complex analysis to a `structural-analyzer`
* to clearly split in the resulting data structure what comes from this analysis (the source) and comes from design decisions/dev-options to map this as a legit go structure


## Inputs

Model generation requires an `core/jsonschema/analyzers/structural-analyzer/Schema` to be prepared from the JSON schema source.

This result is mapped as `Source`.

The "model.go" equivalent handles generation options and design choices to produce a `Target` that is handed over to the templates engine.

The template repo is assumed to be a standalone utility, e.g. `core/template-repo`.

I don't think that we really need a super-smart template engine, like the one used at XX. We want to make it simpler, not to add javascript in it...

## Method

* start from go-swagger templates for models
* sketch how the desired template should look like
* sketch which pieces of data the `Target` should contain
* we may unit test this

Example:

* `model.gotmpl` -> do we need `ExtraSchemas`? Can't remember exactly what this was used for
  * do we want to still support annotations? May be deferred
  * we don't need `IsExported` (decided upstream)
  * we don't need `pascalize` (decided upstream)
* `header.gotmpl`
