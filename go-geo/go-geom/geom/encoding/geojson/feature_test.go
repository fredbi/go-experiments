package geojson

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"
)

func TestFeatureMarshalJSON(t *testing.T) {
	f := NewFeature(geom.NewPointFlat(geom.XY, []float64{1, 2}))
	blob, err := f.MarshalJSON()
	if err != nil {
		t.Fatalf("error marshalling to json: %v", err)
	}

	if !bytes.Contains(blob, []byte(`"properties":null`)) {
		t.Errorf("json should set properties to null if there are none")
	}
}

func TestFeatureMarshalJSON_BBox(t *testing.T) {
	f := &Feature{
		BBox:       &BBox{geom.NewBounds(geom.XY).Set(1, 1, 2, 2)},
		Geometry:   geom.NewPointFlat(geom.XY, []float64{1, 2}),
		Properties: make(map[string]interface{}),
	}

	// bbox empty
	f.BBox = nil
	blob, err := f.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, blob, []byte("bbox"))

	// some bbox
	f.BBox = &BBox{geom.NewBounds(geom.XY).Set(1, 2, 3, 4)}
	blob, err = f.MarshalJSON()
	require.NoError(t, err)
	require.True(t, bytes.Contains(blob, []byte(`"bbox":[1,2,3,4]`)))
}

func TestFeatureMarshalJSON_Bound(t *testing.T) {
	bb := geom.NewBounds(geom.XY).SetCoords(geom.Coord{1, 1}, geom.Coord{2, 2})
	f := &Feature{
		BBox:       &BBox{bb},
		Geometry:   bb.Polygon(),
		Properties: make(map[string]interface{}),
	}

	blob, err := f.MarshalJSON()

	if err != nil {
		t.Fatalf("error marshalling to json: %v", err)
	}

	if !bytes.Contains(blob, []byte(`"type":"Polygon"`)) {
		t.Errorf("should set type to polygon")
	}

	if !bytes.Contains(blob, []byte(`"coordinates":[[[1,1],[1,2],[2,2],[2,1],[1,1]]]`)) {
		t.Errorf("should set type to polygon coords: %s", blob)
	}
}

func TestFeatureMarshal(t *testing.T) {
	f := NewFeature(geom.NewPointFlat(geom.XY, []float64{1, 2}))
	blob, err := json.Marshal(f)

	if err != nil {
		t.Fatalf("should marshal to json just fine but got %v", err)
	}

	if !bytes.Contains(blob, []byte(`"properties":null`)) {
		t.Errorf("json should set properties to null if there are none")
	}
	if !bytes.Contains(blob, []byte(`"type":"Feature"`)) {
		t.Errorf("json should set properties to null if there are none")
	}
}

func TestFeatureMarshalValue(t *testing.T) {
	f := NewFeature(geom.NewPointFlat(geom.XY, []float64{1, 2}))
	blob, err := json.Marshal(*f)

	if err != nil {
		t.Fatalf("should marshal to json just fine but got %v", err)
	}

	if !bytes.Contains(blob, []byte(`"properties":null`)) {
		t.Errorf("json should set properties to null if there are none")
	}
}

func TestFeatureFromJSON(t *testing.T) {
	rawJSON := `
	  { "type": "Feature",
	    "geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
	    "properties": {"prop0": "value0"}
	  }`

	f, err := FeatureFromJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(f.Properties) != 1 {
		t.Errorf("should have 1 property but got: %v", f.Properties)
	}

	// not a feature
	data, _ := NewFeatureCollection().MarshalJSON()
	_, err = FeatureFromJSON(data)
	if err == nil {
		t.Error("should return error if not a feature")
	}

	if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("incorrect error: %v", err)
	}

	// invalid json
	_, err = FeatureFromJSON([]byte(`{"type": "Feature",`)) // truncated
	if err == nil {
		t.Errorf("should return error for invalid json")
	}

	f = &Feature{}
	err = f.UnmarshalJSON([]byte(`{"type": "Feature",`)) // truncated
	if err == nil {
		t.Errorf("should return error for invalid json")
	}
}

func TestFeatureFromJSON_BBox(t *testing.T) {
	rawJSON := `
	  { "type": "Feature",
	    "geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
			"bbox": [1,2,3,4],
	    "properties": {"prop0": "value0"}
	  }`

	f, err := FeatureFromJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	require.EqualValues(t, 1, f.BBox.Min(0))
	require.EqualValues(t, 2, f.BBox.Min(1))
	require.EqualValues(t, 3, f.BBox.Max(0))
	require.EqualValues(t, 4, f.BBox.Max(1))
}

func TestMarshalFeatureID(t *testing.T) {
	f := &Feature{
		ID: "asdf",
	}

	data, err := f.MarshalJSON()
	if err != nil {
		t.Fatalf("should marshal, %v", err)
	}

	fmt.Println(string(data))
	if !bytes.Equal(data, []byte(`{"id":"asdf","type":"Feature","geometry":null,"properties":null}`)) {
		t.Errorf("data not correct")
		t.Logf("%v", string(data))
	}

	f.ID = 123
	data, err = f.MarshalJSON()
	if err != nil {
		t.Fatalf("should marshal, %v", err)

	}

	if !bytes.Equal(data, []byte(`{"id":123,"type":"Feature","geometry":null,"properties":null}`)) {
		t.Errorf("data not correct")
		t.Logf("%v", string(data))
	}
}

func TestFeatureFromJSONID(t *testing.T) {
	rawJSON := `
	  { "type": "Feature",
	    "id": 123,
	    "geometry": {"type": "Point", "coordinates": [102.0, 0.5]}
	  }`

	f, err := FeatureFromJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("should unmarshal feature without issue, err %v", err)
	}

	if v, ok := f.ID.(float64); !ok || v != 123 {
		t.Errorf("should parse id as number, got %T %f", f.ID, v)
	}

	rawJSON = `
	  { "type": "Feature",
	    "id": "abcd",
	    "geometry": {"type": "Point", "coordinates": [102.0, 0.5]}
	  }`

	f, err = FeatureFromJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("should unmarshal feature without issue, err %v", err)
	}

	if v, ok := f.ID.(string); !ok || v != "abcd" {
		t.Errorf("should parse id as string, got %T %s", f.ID, v)
	}
}
