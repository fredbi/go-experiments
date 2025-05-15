package mvt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"reflect"
	"testing"

	"github.com/fredbi/geo/pkg/maptile"
	vectortile "github.com/fredbi/geo/pkg/mvt/vtile"
	"github.com/fredbi/geo/utils/geojson"

	"github.com/twpayne/go-geom"
)

func TestMarshalMarshalGzipped_Full(t *testing.T) {
	tile := maptile.New(8956, 12223, 15)

	ls := geom.NewLineString(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{-81.60346275, 41.50998572},
			geom.Coord{-81.6033669, 41.50991259},
			geom.Coord{-81.60355599, 41.50976036},
			geom.Coord{-81.6040648, 41.50936811},
			geom.Coord{-81.60404411, 41.50935405},
		},
	)
	expected := ls.Clone()

	f := geojson.NewFeature(ls)
	f.Properties = geojson.Properties{
		"source":       "openstreetmap.org",
		"kind":         "path",
		"name":         "Uptown Alley",
		"landuse_kind": "retail",
		"sort_rank":    float64(354),
		"kind_detail":  "pedestrian",
		"min_zoom":     float64(13),
		"id":           float64(246698394),
	}

	fc := geojson.NewFeatureCollection()
	fc.Append(f)

	layers := Layers{NewLayer("roads", fc)}

	// project to the tile coords
	layers.ProjectToTile(tile)

	// marshal
	encoded, err := MarshalGzipped(layers)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// unmarshal
	decoded, err := UnmarshalGzipped(encoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// project back
	decoded.ProjectToWGS84(tile)

	// compare the results
	result := decoded[0].Features[0]
	compareProperties(t, result.Properties, f.Properties)

	// compare geometry
	xe, ye := tileEpsilon(tile)
	compareGeomGeometry(t, result.Geometry, expected, xe, ye)
}

func TestMarshalUnmarshal(t *testing.T) {
	cases := []struct {
		name string
		tile maptile.Tile
	}{
		{
			name: "15-8956-12223",
			tile: maptile.New(8956, 12223, 15),
		},
		{
			name: "16-17896-24449",
			tile: maptile.New(17896, 24449, 16),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expected := loadGeoJSON(t, tc.tile)
			layers := NewLayers(loadGeoJSON(t, tc.tile))
			layers.ProjectToTile(tc.tile)
			data, err := Marshal(layers)
			if err != nil {
				t.Errorf("marshal error: %v", err)
			}

			layers, err = Unmarshal(data)
			if err != nil {
				t.Errorf("unmarshal error: %v", err)
			}

			layers.ProjectToWGS84(tc.tile)
			result := layers.ToFeatureCollections()

			xEpsilon, yEpsilon := tileEpsilon(tc.tile)
			for key := range expected {
				for i := range expected[key].Features {
					r := result[key].Features[i]
					e := expected[key].Features[i]

					// t.Logf("checking layer %s: feature %d", key, i)
					compareProperties(t, r.Properties, e.Properties)
					compareGeomGeometry(t, r.Geometry, e.Geometry, xEpsilon, yEpsilon)
				}
			}
		})
	}
}

func TestMarshalUnmarshalMiniMesh(t *testing.T) {
	cases := []struct {
		name string
		tile maptile.Tile
	}{
		{
			name: "15-8956-12223",
			tile: maptile.New(8956, 12223, 15),
		},
		{
			name: "16-17896-24449",
			tile: maptile.New(17896, 24449, 16),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expected := loadMeshMiniData(t)
			fmt.Println("expected", expected)
			layers := NewLayers(loadMeshMiniData(t))
			layers.ProjectToTile(tc.tile)
			data, err := Marshal(layers)
			if err != nil {
				t.Errorf("marshal error: %v", err)
			}

			layers, err = Unmarshal(data)
			if err != nil {
				t.Errorf("unmarshal error: %v", err)
			}

			layers.ProjectToWGS84(tc.tile)
			result := layers.ToFeatureCollections()

			xEpsilon, yEpsilon := tileEpsilon(tc.tile)
			for key := range expected {
				for i := range expected[key].Features {
					r := result[key].Features[i]
					e := expected[key].Features[i]

					// t.Logf("checking layer %s: feature %d", key, i)
					compareProperties(t, r.Properties, e.Properties)
					compareGeomGeometry(t, r.Geometry, e.Geometry, xEpsilon, yEpsilon)
				}
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	cases := []struct {
		name string
		tile maptile.Tile
	}{
		{
			name: "15-8956-12223",
			tile: maptile.New(8956, 12223, 15),
		},
		{
			name: "16-17896-24449",
			tile: maptile.New(17896, 24449, 16),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expected := loadGeoJSON(t, tc.tile)
			layers, err := UnmarshalGzipped(loadMVT(t, tc.tile))
			if err != nil {
				t.Fatalf("error unmarshalling: %v", err)
			}

			layers.ProjectToWGS84(tc.tile)
			result := layers.ToFeatureCollections()

			xEpsilon, yEpsilon := tileEpsilon(tc.tile)
			for key := range expected {
				for i := range expected[key].Features {
					r := result[key].Features[i]
					e := expected[key].Features[i]

					t.Logf("checking layer %s: feature %d", key, i)
					compareProperties(t, r.Properties, e.Properties)
					compareGeomGeometry(t, r.Geometry, e.Geometry, xEpsilon, yEpsilon)
				}
			}
		})
	}
}

func compareProperties(t testing.TB, result, expected geojson.Properties) {
	t.Helper()

	// properties
	fr := map[string]interface{}(result)
	fe := map[string]interface{}(expected)

	for k, v := range fe {
		if _, ok := v.([]interface{}); ok {
			// arrays are not included
			delete(fr, k)
			delete(fe, k)
		}

		// https: //github.com/tilezen/mapbox-vector-tile/pull/97
		// mapzen error where a 1 is encoded a boolean true
		// just ignore all known cases of that.
		if k == "scale_rank" || k == "layer" {
			if v == 1.0 {
				delete(fr, k)
				delete(fe, k)
			}
		}
	}

	if !reflect.DeepEqual(fr, fe) {
		t.Errorf("properties not equal")
		if len(fr) != len(fe) {
			t.Errorf("properties length not equal: %v != %v", len(fr), len(fe))
		}

		for k := range fr {
			t.Logf("%s: %T %v -- %T %v", k, fr[k], fr[k], fe[k], fe[k])
		}
	}
}

func loadMVT(t testing.TB, tile maptile.Tile) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%d-%d-%d.mvt", tile.Z, tile.X, tile.Y))
	if err != nil {
		t.Fatalf("failed to load mvt file: %v", err)
	}

	return data
}

func TestMarshal_ID(t *testing.T) {
	cases := []struct {
		name string
		id   interface{}
		val  float64
	}{
		{
			name: "int",
			id:   int(86427531),
			val:  86427531,
		},
		{
			name: "int8",
			id:   int8(123),
			val:  123,
		},
		{
			name: "int16",
			id:   int16(6884),
			val:  6884,
		},
		{
			name: "int32",
			id:   int32(123),
			val:  123,
		},
		{
			name: "int64",
			id:   int64(12345678),
			val:  12345678,
		},
		{
			name: "uint",
			id:   uint(86427531),
			val:  86427531,
		},
		{
			name: "uint8",
			id:   uint8(123),
			val:  123,
		},
		{
			name: "uint16",
			id:   uint16(6884),
			val:  6884,
		},
		{
			name: "uint32",
			id:   uint32(123),
			val:  123,
		},
		{
			name: "uint64",
			id:   uint64(12345678),
			val:  12345678,
		},
		{
			name: "float32",
			id:   float32(123.45),
			val:  123,
		},
		{
			name: "float64",
			id:   float64(123.45),
			val:  123,
		},
		{
			name: "string",
			id:   "123456",
			val:  123456,
		},

		// negatives
		{
			name: "negative string",
			id:   "-123456",
			val:  0, // nil
		},
		{
			name: "negative int",
			id:   int(-123456),
			val:  0, // nil
		},
		{
			name: "negative int64",
			id:   int64(-123456),
			val:  0, // nil
		},
		{
			name: "negative float64",
			id:   float64(-123456),
			val:  0, // nil
		},
	}

	f := geojson.NewFeature(geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{1, 2}))
	f.Properties["type"] = "point"
	fc := geojson.NewFeatureCollection().Append(f)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f.ID = tc.id

			data, err := Marshal(NewLayers(map[string]*geojson.FeatureCollection{"roads": fc}))
			if err != nil {
				t.Errorf("marshal error: %v", err)
			}

			ls, err := Unmarshal(data)
			if err != nil {
				t.Errorf("unmarshal error: %v", err)
			}

			id := ls.ToFeatureCollections()["roads"].Features[0].ID
			if tc.val > 0 {
				if id.(float64) != tc.val {
					t.Errorf("incorrect id: %v != %v", id, tc.val)
				}
			} else {
				if id != nil {
					t.Errorf("id should be nil: %v", id)
				}
			}
		})
	}

	t.Run("unmarshaled int from json", func(t *testing.T) {
		f.ID = 123

		data, err := fc.MarshalJSON()
		if err != nil {
			t.Fatalf("json marshal error: %v", err)
		}

		fc, err = geojson.FeatureCollectionFromJSON(data)
		if err != nil {
			t.Fatalf("unmarshal json error: %v", err)
		}

		if _, ok := fc.Features[0].ID.(float64); !ok {
			t.Errorf("json should unmarshal numbers to float64: %T", fc.Features[0].ID)
		}

		data, err = Marshal(NewLayers(map[string]*geojson.FeatureCollection{"roads": fc}))
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		ls, err := Unmarshal(data)
		if err != nil {
			t.Errorf("unmarshal error: %v", err)
		}

		id := ls.ToFeatureCollections()["roads"].Features[0].ID
		if _, ok := id.(float64); !ok {
			// this is to be consistent with json decoding
			t.Errorf("should unmarshal id to float64: %T", id)
		}
	})
}

func BenchmarkMarshal(b *testing.B) {
	layers := NewLayers(loadGeoJSON(b, maptile.New(17896, 24449, 16)))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Marshal(layers)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	layers := NewLayers(loadGeoJSON(b, maptile.New(17896, 24449, 16)))
	data, err := Marshal(layers)
	if err != nil {
		b.Fatalf("marshal error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Unmarshal(data)
	}
}

func BenchmarkDecode(b *testing.B) {
	layers := NewLayers(loadGeoJSON(b, maptile.New(17896, 24449, 16)))
	data, err := Marshal(layers)
	if err != nil {
		b.Fatalf("marshal error: %v", err)
	}

	vt := &vectortile.Tile{}
	err = vt.Unmarshal(data)
	if err != nil {
		b.Fatalf("unmarshal error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = decode(vt)
	}
}

func BenchmarkProjectToTile(b *testing.B) {
	tile := maptile.New(17896, 24449, 16)
	layers := NewLayers(loadGeoJSON(b, maptile.New(17896, 24449, 16)))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		layers.ProjectToTile(tile)
	}
}

func BenchmarkProjectToGeo(b *testing.B) {
	tile := maptile.New(17896, 24449, 16)
	layers := NewLayers(loadGeoJSON(b, maptile.New(17896, 24449, 16)))

	layers.ProjectToTile(tile)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		layers.ProjectToWGS84(tile)
	}
}

func loadGeoJSON(t testing.TB, tile maptile.Tile) map[string]*geojson.FeatureCollection {
	data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%d-%d-%d.json", tile.Z, tile.X, tile.Y))
	if err != nil {
		t.Fatalf("failed to load mvt file: %v", err)
	}

	r := make(map[string]*geojson.FeatureCollection)
	err = json.Unmarshal(data, &r)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	return r
}

func tileEpsilon(tile maptile.Tile) (float64, float64) {
	b := tile.Bound()
	xEpsilon := (b.Max[0] - b.Min[0]) / 4096 * 2 // allowed error
	yEpsilon := (b.Max[1] - b.Min[1]) / 4096 * 2

	return xEpsilon, yEpsilon
}

func compareGeomGeometry(
	t testing.TB,
	result, expected geom.T,
	xEpsilon, yEpsilon float64,
) {
	t.Helper()

	if len(result.FlatCoords()) != len(expected.FlatCoords()) {
		fmt.Println("FlatCoords not equal")
		t.Errorf("geometry FlatCoords length not equal: %v != %v", len(result.FlatCoords()), len(expected.FlatCoords()))
	}

	if !reflect.DeepEqual(result.Ends(), expected.Ends()) {
		fmt.Println("Ends not equal")
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry ends not equal")
	}
	if !reflect.DeepEqual(result.Endss(), expected.Endss()) {
		fmt.Println("Endss not equal")
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry endss not equal")
	}

	if !reflect.DeepEqual(result.Stride(), expected.Stride()) {
		fmt.Println("Stride not equal")

		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry stride not equal")
	}
	if !reflect.DeepEqual(result.SRID(), expected.SRID()) {
		fmt.Println("SRID not equal")

		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry SRID not equal")
	}
	eCoords := expected.FlatCoords()
	rCoords := result.FlatCoords()
	for i := range eCoords {
		diff := math.Abs(eCoords[i] - rCoords[i])

		if diff > xEpsilon || diff > yEpsilon {
			t.Errorf("%d x: %f != %f    %f", i, eCoords[i], rCoords[i], diff)
		}
	}

	// rBounds := result.Bounds()
	// eBounds := expected.Bounds()
	// for i := range eBounds.min {
	// 	diff := math.Abs(eBounds[i] - rBounds[i])

	// 	if diff > xEpsilon || diff > yEpsilon {
	// 		t.Errorf("Bound not equal %d x: %f != %f    %f", i, eBounds[i], rBounds[i], diff)
	// 	}
	// }

}

func loadMeshMiniData(t testing.TB) map[string]*geojson.FeatureCollection {
	s := json.RawMessage([]byte(`
	{
		"meshmap": {
		  "type": "FeatureCollection",
		  "features": [
			{
			  "type": "Feature",
			  "geometry": {
				"type": "Polygon",
				"coordinates": [
				  [
					[
					  -110.910501181,
					  31.3703422151
					],
					[
					  -110.910647811,
					  31.3704183139
					],
					[
					  -110.910649376,
					  31.370273069
					],
					[
					  -110.910501181,
					  31.3703422151
					]
				  ]
				]
			  },
			  "properties": {
				"angle": 3.33,
				"kind": "building"
			  }
			},
			{
			  "type": "Feature",
			  "geometry": {
				"type": "Polygon",
				"coordinates": [
				  [
					[
					  -110.910649376,
					  31.370273069
					],
					[
					  -110.910834194,
					  31.3702683656
					],
					[
					  -110.910696854,
					  31.3701421487
					],
					[
					  -110.910649376,
					  31.370273069
					]
				  ]
				]
			  },
			  "properties": {
				"angle": 3.33,
				"kind": "building"
			  }
			},
			{
			  "type": "Feature",
			  "geometry": {
				"type": "Polygon",
				"coordinates": [
				  [
					[
					  -110.910647811,
					  31.3704183139
					],
					[
					  -110.910749199,
					  31.370369153
					],
					[
					  -110.910649376,
					  31.370273069
					],
					[
					  -110.910647811,
					  31.3704183139
					]
				  ]
				]
			  },
			  "properties": {
				"angle": 3.33
			  }
			}
		  ]
		}
	  }`))
	r := make(map[string]*geojson.FeatureCollection)
	err := json.Unmarshal(s, &r)
	fmt.Println(r["meshmap"].Features[0].Geometry, r["meshmap"].Features[0].Properties)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	return r
}
