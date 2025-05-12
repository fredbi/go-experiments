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
  * we don't need `pascalize` here (decided upstream) but we do in some other context (docstring?)
  * don't need `IncludeModel`. We decide upstream if we want to include a model or not
* `header.gotmpl`:
  * `Package` is needed
  * We don't want `DefaultImports` any longer, just `Imports` 
  * Don't want `strfmt` to be included by default. It would be used only with specific options, like "generate a standalone validator"
 
* `schema.gotmpl`
  * we won't use `IsBaseType`, `IsSubType`: we would rather use `IsInterface`, `IsStruct`
  * `IsEmbedded` is useful
  * We need `pascalize` and `humanize` to build comments
  * need `ReceiverName`
  * `SuperAlias` should be `IsAliased`
  * shouldn't tinker with `AdditionalProperties` or `AdditionalItems`here
  * `IncludeValidator` is a legit codegen option
 
  ```go
  type GenSchema struct {
	//resolvedType
        IsAnonymous       bool  // useful when struct { A struct {...}} (uncommon but possible)
	IsArray           bool  // => IsSlice
	IsMap             bool  // => ok. Alternative maps should be possible (e.g. json/types/OrderedMap)
	IsInterface       bool  // => IsInterface means something different. IsAny replaces the old meaning of IsInterface
	IsPrimitive       bool  // ok
	IsCustomFormatter bool  // => new name?
	IsAliased         bool  // ok means we have type X Y declaration
	IsNullable        bool  // part of the source. The meaning in the target should be IsPointer
	IsStream          bool  // ok
	IsEmptyOmitted    bool  // ok: define the tag
	IsJSONString      bool  // ?
	IsEnumCI          bool  // no. IsEnum is sufficient, then build up an EnumStrategy
	IsBase64          bool  // no
	IsExternal        bool  // yes

	// A tuple gets rendered as an anonymous struct with P{index} as property name
	IsTuple            bool  // yes
	HasAdditionalItems bool  // no. Part of the tuple strategy
        // TODO: IsStruct

	// A complex object gets rendered as a struct
	IsComplexObject bool   // no. Use IsStruct

	// A polymorphic type
	IsBaseType       bool // no. Use either IsStruct or IsInterface
	HasDiscriminator bool // no sure?

	GoType        string  // yes
	Pkg           string  // yes
	PkgAlias      string  // no sure we need this
	AliasedType   string  // The Y in type X Y
	SwaggerType   string  // part of the source, interesting for docstring
	SwaggerFormat string  // part of the source, interesting for docstring
	Extensions    spec.Extensions // perhaps for extensions

	// The type of the element in a slice or map
	ElemType *resolvedType // 

	// IsMapNullOverride indicates that a nullable object is used within an
	// aliased map. In this case, the reference is not rendered with a pointer
	IsMapNullOverride bool // no. Was a hack

	// IsSuperAlias indicates that the aliased type is really the same type,
	// e.g. in golang, this translates to: type A = B
	IsSuperAlias bool  // should be IsRealiased

	// IsEmbedded applies to externally defined types. When embedded, a type
	// is generated in models that embeds the external type, with the Validate
	// method.
	IsEmbedded bool // ok

	SkipExternalValidation bool  // don't remember
	//sharedValidations - reserved for generating a standalone validator. TBD
        spec.SchemaValidations

	HasValidations        bool
	HasContextValidations bool
	Required              bool
	HasSliceValidations   bool
	ItemsEnum             []interface{}
 
	Example                    string  // should be examples now
	OriginalName               string  // can't remember
	Name                       string  // ok
	Suffix                     string  // can' remember
	Path                       string  // ok
	ValueExpression            string  // for validator
	IndexVar                   string
	KeyVar                     string
	Title                      string  // yes
	Description                string  // yes
	Location                   string  // yes
	ReceiverName               string  // yes
	Items                      *GenSchema
	AllowsAdditionalItems      bool // no
	HasAdditionalItems         bool  // no
	AdditionalItems            *GenSchema
	Object                     *GenSchema
	XMLName                    string  // yes. Improve this
	CustomTag                  string // yes. Improve this
	Properties                 GenSchemaList // no => Fields when IsStruct = true
	AllOf                      GenSchemaList // no
	HasAdditionalProperties    bool // no
	IsAdditionalProperties     bool // no
	AdditionalProperties       *GenSchema // no
	StrictAdditionalProperties bool // no
	ReadOnly                   bool // for validator
	IsVirtual                  bool  // no
	IsBaseType                 bool // no
	HasBaseType                bool  // no
	IsSubType                  bool  // no
	IsExported                 bool // no
	IsElem                     bool // IsElem gives some context when the schema is part of an array or a map
	IsProperty                 bool // IsProperty gives some context when the schema is a property of an object
	DiscriminatorField         string // no
	DiscriminatorValue         string // not sure
	Discriminates              map[string]string // not sure
	Parents                    []string
	IncludeValidator           bool // yes
	IncludeModel               bool // no
	Default                    interface{} // yes
	WantsMarshalBinary         bool // do we generate MarshalBinary interface? // yes
	StructTags                 []string // yes
	ExtraImports               map[string]string // non-standard imports detected when using external types // ah yes
	ExternalDocs               *spec.ExternalDocumentation // yes
	WantsRootedErrorPath       bool // not sure
}
```

Validation-only pieces of data. May be skipped to generate the struct only.

```go
TODO
```

## Generated situations

```go
type TargetSchema struct {
  GenOptions
  TypeDefinition
  TypeValidation // TODO
  MarshalOptions // TODO: options with/without JSON marshal, YAML marshal, Binary marshal
  Source SourceSchema // retrieve original spec
}

type Metadata struct {
  Name string
  Title string
  Definition string
  Comment string
  Path string
  Examples []any
}

type ContainerContextFlags struct {
  IsPointer bool
  IsAnonymous bool
  IsEmbedded bool
}

type ContainerContext struct {
  ContainerContextFlags
  ReceiverName string
  Element *TypeDefinition
}

type NameContainerContext struct {
  ContainerContext

  Name string
  Tags []string // struct tags
}

type ContainerFlags struct {
  IsStruct bool     // type A struct { <<Fields>> }
  IsMap bool        // type A map[K]V, see KeyType<K> and ElementType<V>
  IsSlice bool      // type A []<ElementType>
  IsTuple bool      // special case for a struct
  IsExternal bool   // from import
  IsPrimitive bool  // type A <string|int|...>
  IsAny bool        // type A any
  IsInterface bool  // type A interface {}
  IsAliased bool    // type A = B
  IsRedefined bool  // type A B
  IsStream bool     // type A io.Writer/io.Reader
  //IsFile bool       // type A io.Reader (different than IsStream?)
}

type TypeDefinition struct {
  Metadata
  GoType string     // A
  ContainerFlags

  KeyType     *ContainerContext // type of a map key, e.g. string
  ElementType *ContainerContext
  Fields      []NamedContainerContext // fields in a struct or tuple
  Default      any
 }
```

* Container types:
  * struct, slice, map, tuple

* primitive types: string, int64 ... IsPrimitive
* `struct` of fields, some may be unexported. IsStruct
  * Elements of a struct are fields 
* slice of elements: IsSlice
* The type of an element may be either IsAnonymous, IsPointer, ... in the context of this container (ContainerContext)
* map of elements: IsMap
* `any`
* `struct`, which is an interface, e.g., `struct X interface { ... }`
* `struct`, which is a tuple. IsTuple
* anonymous `struct`, e.g. `[]struct{A int}, struct{A struct{...} ...`. IsAnonymous
* json/types, including OrderedMap for maps
* external types
* external map type (anything that has an iterator)
* external slice type (anything that has an iterator)
* io.Reader / io.Writer (IsStream, IsFile)

Supported constructs:
* type aliasing: `type A = B` (IsAliased)
* type "definition" (terminology from go spec): `type A B` (IsRedefined)
* embedded type (HasEmbedded):
  ```go
  struct A {
     B
     ...
  }
```

* private field (e.g., isDefined bool)
* struct tags (json, yaml, XML, custom)
* pointers, but not on slices, maps, or `any`. IsPointer
* Composed types that use embedding. HasEmbedded? (e.g. AllOf)
```go
  struct A {
     B
     C
     D
     ...
  }
```

Unless wrapped in some external type, there is no situation in which we generate:

* a channel type
* a function type
* a parameterized (generic) type

Pointers should never be mandatory, but in the case of circular references.

* Using a pointer to convey the "is defined or null" semantics is an option (it may be the default)
* Whenever the property is not required, zero is indistinguishable from absent.
* Whenever the zero value is not valid and the null value is not valid (the default for OpenAPI), a pointer is not necessary
* Whenever the zero value is valid and not the null value, a pointer is functionally correct
* When null is a possible value, a pointer is not correct, but may be acceptable (e.g., when required set to null is not functionally different from zero)
* Possible alternatives are:
  * using MarshalValidate / UnmarshalValidate
  * using json/types
  * using external types that mimic json/types

## Simple schemas

Simple schemas should not use a different set of templates. These should just be schemas.

Why do we have a special processing in go-swagger?
* For required, it works the other way around: pointer is when _not required_ here
* Defaults are initialized differently (with global vars, should be `MakeX()`).
* The whole validation logic is built differently (server-side)

## Namespace

* package
* may pepper out the generated models over several sub-packages (e.g., according to tag, $ref path, x-go-package extension, spec-transform rule ...)
* may opt in for codegen enums in a separate package

Enums are not just there for validation: we should generate constants as they are useful and idiomatic.
The problem is to generate proper constant names.

Be wary of name conflicts.
Be wary of conflicts produced by generated names (e.g. xxxItems, AllOffXYZ, ...).

## Defaults

* Type constructor with defaults (`MakeX() X`, `NewX() *X`)
* -

## Validator

Standalone validation should be entirely optional, as we know we can produce an "UnmarshalValidate([]byte) (T,error)" or "DecodeValidate(io.Reader) (T, error)" 
function dynamically (this could be some `init()` option to produce this at runtime).

Standalone validation comes with extra responsibilities:
* "standalone" for `required` needs pointers OR `json/types` values
* -
