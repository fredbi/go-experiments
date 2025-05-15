package geojson

import (
	"fmt"
	"io"

	"github.com/twpayne/go-geom"
)

// FeatureFromJSON decodes the data into a GeoJSON feature.
// Alternately one can call json.Unmarshal(f) directly for the same result.
func FeatureFromJSON(data []byte) (*Feature, error) {
	f := &Feature{}
	err := json.Unmarshal(data, f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// NewFeature creates and initializes a GeoJSON feature given the required attributes.
func NewFeature(geometry geom.T) *Feature {
	return &Feature{
		Geometry:   geometry,
		Properties: make(map[string]interface{}),
	}
}

// A Feature corresponds to GeoJSON feature object
type Feature struct {
	ID         interface{} `json:"id,omitempty"`
	BBox       *BBox       `json:"bbox,omitempty"`
	Geometry   geom.T      `json:"geometry"`
	Properties Properties  `json:"properties"`
}

// MarshalJSON converts the feature object into the proper JSON.
// It will handle the encoding of all the child geometries.
// Alternately one can call json.Marshal(f) directly for the same result.
func (f Feature) MarshalJSON() ([]byte, error) {
	g, err := NewGeometry(f.Geometry)
	if err != nil {
		return nil, err
	}

	jf := &jsonFeature{
		ID:         f.ID,
		BBox:       f.BBox,
		Type:       TypeFeature,
		Properties: f.Properties,
		Geometry:   g,
	}

	if len(jf.Properties) == 0 {
		jf.Properties = nil
	}

	return json.Marshal(jf)
}

// UnmarshalJSON handles the correct unmarshalling of the data
// into the geom.T types.
func (f *Feature) UnmarshalJSON(data []byte) error {
	jf := &jsonFeature{}
	err := json.Unmarshal(data, &jf)
	if err != nil {
		return err
	}

	if jf.Type != TypeFeature {
		return ErrUnsupportedType(jf.Type)
	}

	g, err := jf.Geometry.Geometry()
	if err != nil {
		return err
	}

	*f = Feature{
		ID:         jf.ID,
		BBox:       jf.BBox,
		Properties: jf.Properties,
		Geometry:   g,
	}

	return nil
}

func (Feature) IsGeoJSONInterface() {}
func (f *Feature) UnmarshalGQL(v interface{}) error {
	props, ok := v.(string)
	if !ok {
		return fmt.Errorf("geojson feature must be strings, got %T", v)
	}
	if props == "" {
		return nil
	}

	return json.UnmarshalFromString(props, f)
}

func (f Feature) MarshalGQL(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(f)
}

type jsonFeature struct {
	ID         interface{} `json:"id,omitempty"`
	BBox       *BBox       `json:"bbox,omitempty"`
	Type       ObjectType  `json:"type"`
	Geometry   *Geometry   `json:"geometry"`
	Properties Properties  `json:"properties"`
}
