package simplify

import (
	"reflect"
	"testing"

	"github.com/twpayne/go-geom"
)

func TestDouglasPeucker(t *testing.T) {
	cases := []struct {
		name      string
		threshold float64
		ls        *geom.LineString
		expected  *geom.LineString
		indexMap  []int
	}{
		{
			name:      "no reduction",
			threshold: 0.1,
			ls: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0.5, 0.2}, geom.Coord{1, 0}},
			),
			expected: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0.5, 0.2}, geom.Coord{1, 0}},
			),
			indexMap: []int{0, 1, 2},
		},
		{
			name:      "reduction",
			threshold: 1.1,
			ls: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0.5, 0.2}, geom.Coord{1, 0}},
			),
			expected: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{1, 0}},
			),
			indexMap: []int{0, 2},
		},
		{
			name:      "removes coplanar points",
			threshold: 0,
			ls: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0, 1}, geom.Coord{0, 2}},
			),
			expected: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{geom.Coord{0, 0}, geom.Coord{0, 2}},
			),
			indexMap: []int{0, 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, im := DouglasPeucker(tc.threshold).simplify(tc.ls, true)
			if !reflect.DeepEqual(v, tc.expected) {
				t.Log(v)
				t.Log(tc.expected)
				t.Errorf("incorrect line")
			}

			if !reflect.DeepEqual(im, tc.indexMap) {
				t.Log(im)
				t.Log(tc.indexMap)
				t.Errorf("incorrect index map")
			}
		})
	}
}
