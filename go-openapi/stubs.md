# stubs


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

Not sure yet about (5), (6)
