package geojson

import (
	"fmt"
	"io"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/twpayne/go-geom"
)

// DefaultLayout is the default layout for empty geometries.
// FIXME This should be Codec-specific, not global
var DefaultLayout = geom.XY

// ErrDimensionalityTooLow is returned when the dimensionality is too low.
type ErrDimensionalityTooLow int

func (e ErrDimensionalityTooLow) Error() string {
	return fmt.Sprintf("geojson: dimensionality too low (%d)", int(e))
}

// ErrUnsupportedType is returned when the type is unsupported.
type ErrUnsupportedType string

func (e ErrUnsupportedType) Error() string {
	return fmt.Sprintf("geojson: unsupported type: %s", string(e))
}

// this package was largely copied from twpayne/go-geom but it replaces the json serializer
// the result is a significant speed increase about 2 for marshalling and about 5x for unmarshalling
//
// benchmarks from the original:
//
//   goos: linux
//   goarch: amd64
//   pkg: github.com/twpayne/go-geom/encoding/geojson
//   BenchmarkGeometryMarshalJSON-8     	    5000	    350704 ns/op	   24774 B/op	       2 allocs/op
//   BenchmarkGeometryUnmarshalJSON-8   	    2000	    927454 ns/op	   84138 B/op	    2041 allocs/op
//
// benchmarks from this version:
//
//   goos: linux
//   goarch: amd64
//   pkg: github.com/oneconcern/ocpkg/geojson
//   BenchmarkGeometryMarshalJSON-8     	   20000	     86430 ns/op	       4 B/op	       0 allocs/op
//   BenchmarkGeometryUnmarshalJSON-8   	   10000	    174813 ns/op	   51734 B/op	      29 allocs/op

// An ObjectType for GeoJSON objects
type ObjectType string

const (
	TypePoint              ObjectType = "Point"
	TypeMultiPoint         ObjectType = "MultiPoint"
	TypeLineString         ObjectType = "LineString"
	TypeMultiLineString    ObjectType = "MultiLineString"
	TypePolygon            ObjectType = "Polygon"
	TypeMultiPolygon       ObjectType = "MultiPolygon"
	TypeGeometryCollection ObjectType = "GeometryCollection"
	TypeFeature            ObjectType = "Feature"
	TypeFeatureCollection  ObjectType = "FeatureCollection"
)

func (e ObjectType) IsValid() bool {
	switch e {
	case TypePoint, TypeMultiPoint, TypeLineString, TypeMultiLineString, TypePolygon, TypeMultiPolygon, TypeGeometryCollection, TypeFeature, TypeFeatureCollection:
		return true
	}
	return false
}

func (e ObjectType) String() string {
	return string(e)
}

func (e *ObjectType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ObjectType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid GeoJSONType", str)
	}
	return nil
}

func (e ObjectType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

var (
	jsonNil = []byte("null")
	json    jsoniter.API
)

func init() {
	json = jsoniter.ConfigFastest
}

type NoCopyRawMessage jsoniter.RawMessage

func (m *NoCopyRawMessage) UnmarshalJSON(data []byte) error {
	*m = data
	return nil
}

func (m NoCopyRawMessage) MarshalJSON() ([]byte, error) {
	bb := jsoniter.RawMessage(m)
	return []byte(bb), nil
}

// NewGeometry will create a Geometry object but will convert
// the input into a GoeJSON geometry. For example, it will convert
// Rings and Bounds into Polygons.
func NewGeometry(g geom.T) (*Geometry, error) {
	if g == nil {
		return nil, nil
	}
	switch g := g.(type) {
	case *geom.Point:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypePoint,
			Coordinates: &rwm,
		}, nil
	case *geom.LineString:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypeLineString,
			Coordinates: &rwm,
		}, nil
	case *geom.Polygon:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypePolygon,
			Coordinates: &rwm,
		}, nil
	case *geom.MultiPoint:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypeMultiPoint,
			Coordinates: &rwm,
		}, nil
	case *geom.MultiLineString:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypeMultiLineString,
			Coordinates: &rwm,
		}, nil
	case *geom.MultiPolygon:
		var coords jsoniter.RawMessage
		coords, err := json.Marshal(g.Coords())
		if err != nil {
			return nil, err
		}
		rwm := NoCopyRawMessage(coords)
		return &Geometry{
			Type:        TypeMultiPolygon,
			Coordinates: &rwm,
		}, nil
	case *geom.GeometryCollection:
		geometries := make([]*Geometry, len(g.Geoms()))
		for i, subGeometry := range g.Geoms() {
			var err error
			geometries[i], err = NewGeometry(subGeometry)
			if err != nil {
				return nil, err
			}
		}
		return &Geometry{
			Type:       TypeGeometryCollection,
			Geometries: geometries,
		}, nil
	default:
		return nil, geom.ErrUnsupportedType{Value: g}
	}
}

// A Geometry matches the structure of a GeoJSON Geometry.
type Geometry struct {
	Type        ObjectType        `json:"type"`
	BBox        *BBox             `json:"bbox,omitempty"`
	Coordinates *NoCopyRawMessage `json:"coordinates,omitempty"`
	Geometries  []*Geometry       `json:"geometries,omitempty"`
}

func (Geometry) IsGeoJSONInterface() {}

func guessLayout0(coords0 []float64) (geom.Layout, error) {
	switch n := len(coords0); n {
	case 0, 1:
		return geom.NoLayout, ErrDimensionalityTooLow(len(coords0))
	case 2:
		return geom.XY, nil
	case 3:
		return geom.XYZ, nil
	case 4:
		return geom.XYZM, nil
	default:
		return geom.Layout(n), nil
	}
}

func guessLayout1(coords1 []geom.Coord) (geom.Layout, error) {
	if len(coords1) == 0 {
		return DefaultLayout, nil
	}
	return guessLayout0(coords1[0])
}

func guessLayout2(coords2 [][]geom.Coord) (geom.Layout, error) {
	if len(coords2) == 0 {
		return DefaultLayout, nil
	}
	return guessLayout1(coords2[0])
}

func guessLayout3(coords3 [][][]geom.Coord) (geom.Layout, error) {
	if len(coords3) == 0 {
		return DefaultLayout, nil
	}
	return guessLayout2(coords3[0])
}

// Geometry returns the geom.T for the geojson Geometry.
// This will convert the "Geometries" into a geom.GeometryCollection if applicable.
func (g Geometry) Geometry() (geom.T, error) {
	switch g.Type {
	case TypePoint:
		if g.Coordinates == nil {
			return geom.NewPoint(geom.NoLayout), nil
		}
		var coords geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout0(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewPoint(layout).SetCoords(coords)
	case TypeLineString:
		if g.Coordinates == nil {
			return geom.NewLineString(geom.NoLayout), nil
		}
		var coords []geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout1(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewLineString(layout).SetCoords(coords)
	case TypePolygon:
		if g.Coordinates == nil {
			return geom.NewPolygon(geom.NoLayout), nil
		}
		var coords [][]geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout2(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewPolygon(layout).SetCoords(coords)
	case TypeMultiPoint:
		if g.Coordinates == nil {
			return geom.NewMultiPoint(geom.NoLayout), nil
		}
		var coords []geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout1(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewMultiPoint(layout).SetCoords(coords)
	case TypeMultiLineString:
		if g.Coordinates == nil {
			return geom.NewMultiLineString(geom.NoLayout), nil
		}
		var coords [][]geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout2(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewMultiLineString(layout).SetCoords(coords)
	case TypeMultiPolygon:
		if g.Coordinates == nil {
			return geom.NewMultiPolygon(geom.NoLayout), nil
		}
		var coords [][][]geom.Coord
		if err := json.Unmarshal(*g.Coordinates, &coords); err != nil {
			return nil, err
		}
		layout, err := guessLayout3(coords)
		if err != nil {
			return nil, err
		}
		return geom.NewMultiPolygon(layout).SetCoords(coords)
	case TypeGeometryCollection:
		geoms := make([]geom.T, len(g.Geometries))
		for i, subGeometry := range g.Geometries {
			var err error
			geoms[i], err = subGeometry.Geometry()
			if err != nil {
				return nil, err
			}
		}
		gc := geom.NewGeometryCollection()
		if err := gc.Push(geoms...); err != nil {
			return nil, err
		}
		return gc, nil
	default:
		return nil, ErrUnsupportedType(g.Type)
	}
}

func Marshal(g geom.T) ([]byte, error) {
	ng, err := NewGeometry(g)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ng)
}

// Unmarshal unmarshalls a []byte to an arbitrary geometry.
func Unmarshal(data []byte, g *geom.T) error {
	gg := &Geometry{}
	if err := json.Unmarshal(data, gg); err != nil {
		return err
	}
	var err error
	*g, err = gg.Geometry()
	return err
}

func (g *Geometry) UnmarshalGQL(v interface{}) error {
	props, ok := v.(string)
	if !ok {
		return fmt.Errorf("geojson geometries must be strings, got %T", v)
	}
	if props == "" {
		return nil
	}

	return json.UnmarshalFromString(props, g)
}

func (g Geometry) MarshalGQL(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(g)
}
