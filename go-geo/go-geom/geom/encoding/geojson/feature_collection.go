/*
Package geojson is a library for encoding and decoding GeoJSON into Go structs using
the geometries in the go-geom package. Supports both the json.Marshaler and json.Unmarshaler
interfaces as well as helper functions such as `FeatureCollectionFromJSON` and `UnmarshalFeature`.
*/
package geojson

import (
	"fmt"
	"io"
)

// FeatureCollectionFromJSON decodes the data into a GeoJSON feature collection.
// Alternately one can call json.Unmarshal(fc) directly for the same result.
func FeatureCollectionFromJSON(data []byte) (*FeatureCollection, error) {
	fc := &FeatureCollection{}
	err := json.Unmarshal(data, fc)
	if err != nil {
		return nil, err
	}

	return fc, nil
}

// NewFeatureCollection creates and initializes a new feature collection.
func NewFeatureCollection() *FeatureCollection {
	return &FeatureCollection{
		Features: []*Feature{},
	}
}

// A FeatureCollection correlates to a GeoJSON feature collection.
type FeatureCollection struct {
	BBox     *BBox      `json:"bbox,omitempty"`
	Features []*Feature `json:"features"`
}

// Append appends a feature to the collection.
func (fc *FeatureCollection) Append(feature *Feature) *FeatureCollection {
	fc.Features = append(fc.Features, feature)
	return fc
}

// MarshalJSON converts the feature collection object into the proper JSON.
// It will handle the encoding of all the child features and geometries.
// Alternately one can call json.Marshal(fc) directly for the same result.
func (fc FeatureCollection) MarshalJSON() ([]byte, error) {
	c := jsonFeatureCollection{
		Type:     TypeFeatureCollection,
		Features: fc.Features,
	}

	if c.Features == nil {
		c.Features = []*Feature{}
	}
	return json.Marshal(c)
}

// UnmarshalJSON implements json.Unmarshaler
func (fc *FeatureCollection) UnmarshalJSON(data []byte) error {
	var gfc jsonFeatureCollection
	if err := json.Unmarshal(data, &gfc); err != nil {
		return err
	}
	if gfc.Type != TypeFeatureCollection {
		return ErrUnsupportedType(gfc.Type)
	}
	fc.Features = gfc.Features
	return nil
}

func (FeatureCollection) IsGeoJSONInterface() {}
func (fc *FeatureCollection) UnmarshalGQL(v interface{}) error {
	props, ok := v.(string)
	if !ok {
		return fmt.Errorf("geojson feature collection must be strings, got %T", v)
	}
	if props == "" {
		return nil
	}
	return json.UnmarshalFromString(props, fc)
}

func (fc FeatureCollection) MarshalGQL(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(fc)
}

type jsonFeatureCollection struct {
	Type     ObjectType `json:"type"`
	Features []*Feature `json:"features"`
}
