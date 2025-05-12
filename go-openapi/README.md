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
It requires careful design and planning. The difference is that we've 10 years of accumulated feedback and have figured out different designs.

I've taken a look at [a few existing alternatives](alternatives.md) in the goland space. There are a lot of brilliant ideas. And also a few failures.
I am very proud that our work inspired many others. But I didn't see any major improvements, especially regarding JSON schema conformance.

In the sections below, I am giving a more comprehensive analysis. Even though OAIv3 is the ultimate goal, it is not the only one and many micro-designs need
to be questioned, reviewed, and implemented along the way.

## [go-swagger and the go-openapi toolkit status in a nutshell](./go-openapi-repos.md)

A quick keep/change-like analysis of the various tools that have been introduced.

## [Core design analysis](./existing-design-analysis.md)

## repo by repo analysis

Trying to be honest here: a little self-criticism doesn't hurt :).

### [go-swagger](./go-swagger.md)

### [analysis](./analysis.md)

### errors

I think the _idea_ of this repo is pretty much reusable.
It has remained maintained to keep up with go advances in error handling.

I like the idea of self-serving errors with `func (e Error) ServeHTTP()`.
Perhaps a few things should be handed over to different repos, though.
The way I see that one is more as generic error type than it is now (handling validation errors etc). Not sure yet.

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

### [runtime](./runtime.md)

### [spec & spec3](./spec.md)

### strfmt

### [stubs](./stubs.md)

### swag

### validate

This repo mixes 3 different use cases:
1. JSON schema validation
2. swagger v2 spec validation (which relies on JSON schema validation, plus a number of rules)
3. validation helpers to be used by generated code

This repo is deeply tied to the dynamic JSON approach. We may reuse tests, validation helpers, but not much more.

## [Fresh ideas to move forward](./ideas.md)
