# runtime

## Analysis

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


## What would I like to do?

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

## Detailed design (draftish)

* keep UntypedAPI
* build UntypedClient
* trim down client down to BindParams()
* trim down server down to BindRequest()

* server: auth, content nego, etc -> pluggable directly to the router (this is actually the case, clarify the interface)
* client: TLS, tracing, etc -> pluggable directly to Transport (more or less like now, need to clarify the interface)

* server: add more auth middleware
* client: add credential providers

* major change: pluggable router, possibly static (e.g.generated code)
* major change:
  
  client: (http.Client build-up, TLS, tracing, etc) ->
  client: (middleware stack incl. credentials provider / transport hooks/content nego) -> BindParams() -> Request -> <<< wire >> ->
  server: <<< request >> -> Consume() -> (route) -> (middleware stack incl. auth, content nego) -> BindRequest() -> UnmarshalValidate() (body) -> HandlerFunc() Responder
  server: Responder -> Produce()
  
