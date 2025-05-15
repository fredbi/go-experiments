package geojson

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-geom"
)

func TestFeatureCollectionFromJSON(t *testing.T) {
	rawJSON := `
	  { "type": "FeatureCollection",
	    "features": [
	      { "type": "Feature",
	        "geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
	        "properties": {"prop0": "value0"}
	      },
	      { "type": "Feature",
	        "geometry": {
	          "type": "LineString",
	          "coordinates": [
	            [102.0, 0.0], [103.0, 1.0], [104.0, 0.0], [105.0, 1.0]
	            ]
	          },
	        "properties": {
	          "prop0": "value0",
	          "prop1": 0.0
	        }
	      },
	      { "type": "Feature",
	         "geometry": {
	           "type": "Polygon",
	           "coordinates": [
	             [ [100.0, 0.0], [101.0, 0.0], [101.0, 1.0],
	               [100.0, 1.0], [100.0, 0.0] ]
	             ]
	         },
	         "properties": {
	           "prop0": "value0",
	           "prop1": {"this": "that"}
	         }
	       }
	     ]
	  }`

	fc, err := FeatureCollectionFromJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("should unmarshal feature collection without issue, err %v", err)
	}

	if len(fc.Features) != 3 {
		t.Errorf("should have 3 features but got %d", len(fc.Features))
	}

	f := fc.Features[0]
	if gt, ok := f.Geometry.(*geom.Point); !ok {
		t.Errorf("incorrect feature type: %v != %v", gt, "Point")
	}

	f = fc.Features[1]
	if gt, ok := f.Geometry.(*geom.LineString); !ok {
		t.Errorf("incorrect feature type: %v != %v", gt, "LineString")
	}

	f = fc.Features[2]
	if gt, ok := f.Geometry.(*geom.Polygon); !ok {
		t.Errorf("incorrect feature type: %v != %v", gt, "Polygon")
	}

	// check unmarshal/marshal loop
	var expected interface{}
	err = json.Unmarshal([]byte(rawJSON), &expected)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	data, err := json.MarshalIndent(fc, "", " ")
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	var raw interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !assert.EqualValues(t, raw, expected) {
		t.Errorf("invalid marshalling: \n%v", string(data))
	}

	// not a feature collection
	data, _ = NewFeature(geom.NewPoint(geom.XY)).MarshalJSON()
	_, err = FeatureCollectionFromJSON(data)
	if err == nil {
		t.Error("should return error if not a feature collection")
	}

	if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("incorrect error: %v", err)
	}

	// invalid json
	_, err = FeatureCollectionFromJSON([]byte(`{"type": "FeatureCollection",`)) // truncated
	if err == nil {
		t.Errorf("should return error for invalid json")
	}
}

func TestFeatureCollectionMarshalJSON(t *testing.T) {
	fc := NewFeatureCollection()
	blob, err := fc.MarshalJSON()

	if err != nil {
		t.Fatalf("should marshal to json just fine but got %v", err)
	}

	if !bytes.Contains(blob, []byte(`"features":[]`)) {
		t.Errorf("json should set features object to at least empty array")
	}
}

func TestFeatureCollectionMarshal(t *testing.T) {
	fc := NewFeatureCollection()
	fc.Features = nil
	blob, err := json.Marshal(fc)

	if err != nil {
		t.Fatalf("should marshal to json just fine but got %v", err)
	}

	if !bytes.Contains(blob, []byte(`"features":[]`)) {
		t.Errorf("json should set features object to at least empty array")
	}
}

func TestFeatureCollectionMarshalValue(t *testing.T) {
	fc := NewFeatureCollection()
	fc.Features = nil
	blob, err := json.Marshal(*fc)

	if err != nil {
		t.Fatalf("should marshal to json just fine but got %v", err)
	}

	if !bytes.Contains(blob, []byte(`"features":[]`)) {
		t.Errorf("json should set features object to at least empty array")
	}
}
