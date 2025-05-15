package mvt

import (
	"reflect"
	"testing"

	"github.com/fredbi/geo/pkg/simplify"
	"github.com/fredbi/geo/utils/geojson"
	"github.com/twpayne/go-geom"
)

func TestLayerSimplify(t *testing.T) {
	// should remove feature that are empty.
	ls := Layers{&Layer{
		Features: []*geojson.Feature{
			geojson.NewFeature(geom.NewLineString(geom.XY)),
			geojson.NewFeature(geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 1}},
			)),
		},
	}}

	simplifier := simplify.DouglasPeucker(10)
	ls.Simplify(simplifier)

	if len(ls[0].Features) != 1 {
		t.Errorf("should remove empty feature")
	}

	if v := ls[0].Features[0].Geometry.FlatCoords(); !reflect.DeepEqual(v, []float64{0, 0, 1, 1}) {
		t.Errorf("incorrect type: %v", v)
	}
}
