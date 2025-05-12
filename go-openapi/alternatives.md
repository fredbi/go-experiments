# Alternatives to go-openapi and go-swagger

Updated: May 12th 2025

## [kin-openapi](https://github.com/getkin/kin-openapi)

A library to serialize/deserialize OpenAPI v2/v3 specs. Mostly equivalent (and inspired from) go-openapi/spec,
with the additional support of OIA v3, but no real support for recent JSON schema versions.
It uses go-openapi/jsonpointer to resolve internal JSON pointers.

It is rather close to our good old go-openapi/spec work. Too close indeed, to my taste, as I think we need to build something
different, since that one doesn't solve the deep design issues we hit with go-openapi.


https://github.com/oapi-codegen/oapi-codegen
This is built on top of kin-openapi.
Interestingly, it supports a variety of external routers. This is an idea I'd like to keep.

https://github.com/oasdiff/oasdiff

https://github.com/a-h/rest
It is a declarative spec builder, without reliance on code comments or tags, but on explicit declarations alongside your API.
Interesting.

https://github.com/goadesign/goa

## [libopenapi](https://github.com/pb33f/libopenapi)


## Interesting findings along the way

github.com/speakeasy-api/jsonpath


I like this a lot too: https://github.com/speakeasy-api/openapi-overlay

Worth taking a look at too: https://github.com/speakeasy-api/easytemplate

linter: https://github.com/speakeasy-api/vacuum

https://github.com/ugorji/go/tree/master/codec : compare to the lexer I've built. Interesting.
