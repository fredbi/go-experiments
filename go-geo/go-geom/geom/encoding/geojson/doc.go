// Package geojson encodes and decodes [GeoJSON](http://geojson.org/) into Go structs
// using the geometries in the [go-geom](https://github.com/twpayne/go-geom) package.
// Supports both the [json.Marshaler](http://golang.org/pkg/encoding/json/#Marshaler) and
// [json.Unmarshaler](http://golang.org/pkg/encoding/json/#Unmarshaler) interfaces.
// The package also provides helper functions such as `FeatureCollectionFromJSON` and `FeatureFromJSON`.
//
// ## Feature Properties
//
// GeoJSON features can have properties of any type. This can cause issues in a statically typed
// language such as Go. Included is a `Properties` type with some helper methods that will try to
// force convert a property. An optional default, will be used if the property is missing or the wrong
// type.
//
// ## GraphQL
//
// This package also includes the necessary additions to use with graphql gqlgen
//
// This code was originally in https://github.com/twpayne/go-geom/tree/master/encoding/geojson
package geojson

//go:generate go run ../../hack/gen-map-getters.go --name Properties --description "defines the feature properties with some helper methods"
