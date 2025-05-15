package simplify

import (
	"reflect"
	"testing"

	"github.com/twpayne/go-geom"
)

var allGeom = []geom.T{
	geom.NewPoint(geom.XY),
	geom.NewPolygon(geom.XY),
	geom.NewLineString(geom.XY),
	geom.NewLinearRing(geom.XY),
	geom.NewMultiLineString(geom.XY),
	geom.NewMultiPoint(geom.XY),
	geom.NewMultiPolygon(geom.XY),
	geom.NewGeometryCollection(),
}

func TestSimplify(t *testing.T) {
	r := DouglasPeucker(10)
	for _, g := range allGeom {
		simplify(r, g)
	}
}

func TestPolygon(t *testing.T) {
	p := geom.NewPolygon(geom.XY).MustSetCoords(
		[][]geom.Coord{
			[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 0}, geom.Coord{1, 1}, geom.Coord{0, 0}},
			[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0, 0}},
		},
	)

	ex := [][]geom.Coord{
		[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 0}, geom.Coord{1, 1}, geom.Coord{0, 0}},
	}
	simp := DouglasPeucker(0).Polygon(p)
	if !reflect.DeepEqual(ex, simp.Coords()) {
		t.Errorf("should remove empty ring")
	}
}

func TestMultiPolygon(t *testing.T) {
	mp := geom.NewMultiPolygon(geom.XY).MustSetCoords(
		[][][]geom.Coord{
			[][]geom.Coord{
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 0}, geom.Coord{1, 1}, geom.Coord{0, 0}},
			},
			[][]geom.Coord{
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0, 0}},
			},
		},
	)

	ex := [][][]geom.Coord{
		[][]geom.Coord{
			[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 0}, geom.Coord{1, 1}, geom.Coord{0, 0}},
		},
	}
	simpMultiP := DouglasPeucker(0).MultiPolygon(mp)
	if !reflect.DeepEqual(ex, simpMultiP.Coords()) {
		t.Errorf("should remove empty polygon")
	}
}
