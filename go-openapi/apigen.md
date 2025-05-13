# API generation

* hands over:
  * template repo
  * spec deserialization
  * spec validation
  * spec analysis
  * schema code gen (including simple schemas)
  * authorization stuff deferred to runtime

That leaves us with the following concepts to support:

  * Operation, same as Swagger
    * requestBody
  * Servers
    * Server variables
  * Link Objects
  * contentMediaType (from JSONSchema)
  * jsonSchemaDialect
  * webhooks
  * license: now explicitly a SPDX identifier - see https://github.com/spdx/tools-golang
  * additional HTTP methods (options, patch, trace)
  * servers in path item
  * callbacks
  * deprecated: true|false
  * requestBody (replaces `parameters: [ in: body ]`)
  * parameters:
    * in cookie
    * style (simple, matrix, label...)
    * explode
    * allowReserved
    * example or examples
    * content
  * responses:
    * content
