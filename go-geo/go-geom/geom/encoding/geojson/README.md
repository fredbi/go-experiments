# GeoJSON

This package **encodes and decodes** [GeoJSON](http://geojson.org/) into Go structs
using the geometries in the [go-geom](https://github.com/twpayne/go-geom) package.
Supports both the [json.Marshaler](http://golang.org/pkg/encoding/json/#Marshaler) and
[json.Unmarshaler](http://golang.org/pkg/encoding/json/#Unmarshaler) interfaces.
The package also provides helper functions such as `FeatureCollectionFromJSON` and `FeatureFromJSON`.

## Examples

### Unmarshalling  (JSON -> Go)

```go
rawJSON := []byte(`
{ 
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
      "properties": {"prop0": "value0"}
    }
  ]
}`)

fc, _ := geojson.UnmarshalFeatureCollection(rawJSON)

// or

fc := geojson.NewFeatureCollection()
err := json.Unmarshal(rawJSON, &fc)

// Geometry unmarshals into the correct geo.Geometry type.
point := geom.Must(fc.Features[0].Geometry()).(*geom.Point)
```

### Marshalling (Go -> JSON)

```go
fc := geojson.NewFeatureCollection()
fc.Append(geojson.NewFeature(geom.NewPointFlat(geom.XY, []float64{1, 2})))

rawJSON, _ := fc.MarshalJSON()

// or
blob, _ := json.Marshal(fc)
```

## Feature Properties

GeoJSON features can have properties of any type. This can cause issues in a statically typed
language such as Go. Included is a `Properties` type with some helper methods that will try to
force convert a property. An optional default, will be used if the property is missing or the wrong
type.

```go
f.Properties.MustBool(key string, def ...bool) bool
f.Properties.MustFloat64(key string, def ...float64) float64
f.Properties.MustInt(key string, def ...int) int
f.Properties.MustString(key string, def ...string) string
```
