package clip

import (
	"reflect"
	"testing"

	"github.com/fredbi/geo/pkg/bound"
	"github.com/paulmach/orb"
	"github.com/twpayne/go-geom"
)

func TestInternalLine(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LineString
		output *geom.MultiLineString
	}{
		{
			name:  "clip line",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{30, 30}}),
			input: toLineString(orb.LineString{
				{-10, 10}, {10, 10}, {10, -10}, {20, -10}, {20, 10}, {40, 10},
				{40, 20}, {20, 20}, {20, 40}, {10, 40}, {10, 20}, {5, 20}, {-10, 20},
			}),
			output: toMls(orb.MultiLineString{
				{{0, 10}, {10, 10}, {10, 0}},
				{{20, 0}, {20, 10}, {30, 10}},
				{{30, 20}, {20, 20}, {20, 30}},
				{{10, 30}, {10, 20}, {5, 20}, {0, 20}},
			}),
		},
		{
			name:  "clips line crossing many times",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{20, 20}}),
			input: toLineString(orb.LineString{
				{10, -10}, {10, 30}, {20, 30}, {20, -10},
			}),
			output: toMls(orb.MultiLineString{
				{{10, 0}, {10, 20}},
				{{20, 20}, {20, 0}},
			}),
		},
		{
			name:  "no changes if all inside",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{20, 20}}),
			input: toLineString(orb.LineString{
				{1, 1}, {2, 2}, {3, 3},
			}),
			output: toMls(orb.MultiLineString{
				{{1, 1}, {2, 2}, {3, 3}},
			}),
		},
		{
			name:  "empty if nothing in bound",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{2, 2}}),
			input: toLineString(orb.LineString{
				{10, 10}, {20, 20}, {30, 30},
			}),
			output: nil,
		},
		{
			name:  "floating point example",
			bound: toBound(orb.Bound{Min: orb.Point{-91.93359375, 42.29356419217009}, Max: orb.Point{-91.7578125, 42.42345651793831}}),
			input: toLineString(orb.LineString{
				{-86.66015624999999, 42.22851735620852}, {-81.474609375, 38.51378825951165},
				{-85.517578125, 37.125286284966776}, {-85.8251953125, 38.95940879245423},
				{-90.087890625, 39.53793974517628}, {-91.93359375, 42.32606244456202},
				{-86.66015624999999, 42.22851735620852},
			}),
			output: toMls(orb.MultiLineString{
				{
					{-91.91208030440808, 42.29356419217009},
					{-91.93359375, 42.32606244456202},
					{-91.7578125, 42.3228109416169},
				},
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := line(tc.bound, tc.input, false)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect clip %v", tc.name)
				t.Logf("%v %v", result, result.Coords())
				t.Logf("%v", tc.output)
			}
		})
	}
}

func toBound(b orb.Bound) bound.Bound {
	return bound.NewBound(b.Min[0], b.Min[1], b.Max[0], b.Max[1])
}

func multiLineStringToCoords(mls orb.MultiLineString) [][]geom.Coord {
	var coordss [][]geom.Coord
	for _, line := range mls {
		var coords []geom.Coord
		for _, p := range line {
			coords = append(coords, geom.Coord{p[0], p[1]})
		}
		coordss = append(coordss, coords)
	}
	return coordss
}

func lineToCoords(s orb.LineString) []geom.Coord {
	var coords []geom.Coord
	for _, point := range s {
		coords = append(coords, geom.Coord{point[0], point[1]})
	}
	return coords
}

func ringToCoords(s orb.Ring) []geom.Coord {
	var coords []geom.Coord
	for _, point := range s {
		coords = append(coords, geom.Coord{point[0], point[1]})
	}
	return coords
}

func toMls(mls orb.MultiLineString) *geom.MultiLineString {
	newMls := geom.NewMultiLineString(geom.XY)
	coords := multiLineStringToCoords(mls)
	newMls.MustSetCoords(coords)
	return newMls
}

func toRing(r orb.Ring) *geom.LinearRing {
	ring := geom.NewLinearRing(geom.XY)
	coords := ringToCoords(r)
	ring.MustSetCoords(coords)
	return ring
}

func toLineString(s orb.LineString) *geom.LineString {
	ls := geom.NewLineString(geom.XY)
	coords := lineToCoords(s)
	ls.MustSetCoords(coords)
	return ls
}

func toPoint(p orb.Point) *geom.Point {
	return geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{p[0], p[1]})
}

func toMPoint(mp orb.MultiPoint) *geom.MultiPoint {
	newMp := geom.NewMultiPoint(geom.XY)
	var coords []geom.Coord
	for _, p := range mp {
		coords = append(coords, geom.Coord{p[0], p[1]})
	}
	newMp.MustSetCoords(coords)
	return newMp
}

func toPolygon(p orb.Polygon) *geom.Polygon {
	newp := geom.NewPolygon(geom.XY)
	for _, r := range p {
		ring := toRing(r)
		err := newp.Push(ring)
		if err != nil {
			return nil
		}
	}
	return newp
}

func toMultiPolygon(mp orb.MultiPolygon) *geom.MultiPolygon {
	newMp := geom.NewMultiPolygon(geom.XY)
	for _, p := range mp {
		newp := toPolygon(p)
		err := newMp.Push(newp)
		if err != nil {
			return nil
		}
	}
	return newMp
}

func TestInternalRing(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LinearRing
		output *geom.LinearRing
	}{
		{
			name:  "clips polygon",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{30, 30}}),
			input: toRing(orb.Ring{
				{-10, 10}, {0, 10}, {10, 10}, {10, 5}, {10, -5},
				{10, -10}, {20, -10}, {20, 10}, {40, 10}, {40, 20},
				{20, 20}, {20, 40}, {10, 40}, {10, 20}, {5, 20},
				{-10, 20}}),
			// note: we allow duplicate points if polygon endpoints are
			// on the box boundary.
			output: toRing(orb.Ring{
				{0, 10}, {0, 10}, {10, 10}, {10, 5}, {10, 0},
				{20, 0}, {20, 10}, {30, 10}, {30, 20}, {20, 20},
				{20, 30}, {10, 30}, {10, 20}, {5, 20}, {0, 20},
			}),
		},
		{
			name:  "completely inside bound",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{10, 10}}),
			input: toRing(orb.Ring{
				{3, 3}, {5, 3}, {5, 5}, {3, 5}, {3, 3},
			}),
			output: toRing(orb.Ring{
				{3, 3}, {5, 3}, {5, 5}, {3, 5}, {3, 3},
			}),
		},
		{
			name:  "completely around bound",
			bound: toBound(orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{2, 2}}),
			input: toRing(orb.Ring{
				{0, 0}, {3, 0}, {3, 3}, {0, 3}, {0, 0},
			}),
			output: toRing(orb.Ring{{1, 2}, {1, 1}, {2, 1}, {2, 2}, {1, 2}}),
		},
		{
			name:  "completely around touching corners",
			bound: toBound(orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{3, 3}}),
			input: toRing(orb.Ring{
				{0, 2}, {2, 0}, {4, 2}, {2, 4}, {0, 2},
			}),
			output: toRing(orb.Ring{{1, 1}, {1, 1}, {3, 1}, {3, 1}, {3, 3}, {3, 3}, {1, 3}, {1, 3}, {1, 1}}),
		},
		{
			name:  "around but cut corners",
			bound: toBound(orb.Bound{Min: orb.Point{0.5, 0.5}, Max: orb.Point{3.5, 3.5}}),
			input: toRing(orb.Ring{
				{0, 2}, {2, 4}, {4, 2}, {2, 0}, {0, 2},
			}),
			output: toRing(orb.Ring{{0.5, 2.5}, {1.5, 3.5}, {2.5, 3.5}, {3.5, 2.5}, {3.5, 1.5}, {2.5, 0.5}, {1.5, 0.5}, {0.5, 1.5}, {0.5, 2.5}}),
		},
		{
			name:  "unclosed ring",
			bound: toBound(orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{4, 4}}),
			input: toRing(orb.Ring{
				{2, 0}, {3, 0}, {3, 5}, {2, 5},
			}),
			output: toRing(orb.Ring{{3, 1}, {3, 4}}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ring(tc.bound, tc.input)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect clip")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}
}

func TestLineString(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LineString
		output *geom.MultiLineString
	}{
		{
			name:  "clips line crossing many times",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{20, 20}}),
			input: toLineString(orb.LineString{
				{10, -10}, {10, 30}, {20, 30}, {20, -10},
			}),
			output: toMls(orb.MultiLineString{
				{{10, 0}, {10, 20}},
				{{20, 20}, {20, 0}},
			}),
		},
		{
			name:  "touches the sides a bunch of times",
			bound: toBound(orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{6, 6}}),
			input: toLineString(orb.LineString{{2, 3}, {1, 4}, {2, 5}, {2, 6}, {3, 5}, {4, 6}, {5, 5}, {5, 7}, {0, 7}, {0, 3}, {2, 3}}),
			output: toMls(orb.MultiLineString{
				{{2, 3}, {1, 4}, {2, 5}, {2, 6}, {3, 5}, {4, 6}, {5, 5}, {5, 6}},
				{{1, 3}, {2, 3}},
			}),
		},
		{
			name:  "floating point example",
			bound: toBound(orb.Bound{Min: orb.Point{-91.93359375, 42.29356419217009}, Max: orb.Point{-91.7578125, 42.42345651793831}}),
			input: toLineString(orb.LineString{
				{-86.66015624999999, 42.22851735620852}, {-81.474609375, 38.51378825951165},
				{-85.517578125, 37.125286284966776}, {-85.8251953125, 38.95940879245423},
				{-90.087890625, 39.53793974517628}, {-91.93359375, 42.32606244456202},
				{-86.66015624999999, 42.22851735620852},
			}),
			output: toMls(orb.MultiLineString{
				{
					{-91.91208030440808, 42.29356419217009},
					{-91.93359375, 42.32606244456202},
					{-91.7578125, 42.3228109416169},
				},
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := LineString(tc.bound, tc.input)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect clip")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}
}

func TestLineString_OpenBound(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LineString
		output *geom.MultiLineString
	}{
		{
			name:  "clips line crossing many times",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{20, 20}}),
			input: toLineString(orb.LineString{
				{10, -10}, {10, 30}, {20, 30}, {20, -10},
			}),
			output: toMls(orb.MultiLineString{
				{{10, 0}, {10, 20}},
			}),
		},
		{
			name:  "touches the sides a bunch of times",
			bound: toBound(orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{6, 6}}),
			input: toLineString(orb.LineString{{2, 3}, {1, 4}, {2, 5}, {2, 6}, {3, 5}, {4, 6}, {5, 5}, {5, 7}, {0, 7}, {0, 3}, {2, 3}}),
			output: toMls(orb.MultiLineString{
				{{2, 3}, {1, 4}},
				{{1, 4}, {2, 5}, {2, 6}},
				{{2, 6}, {3, 5}, {4, 6}},
				{{4, 6}, {5, 5}, {5, 6}},
				{{1, 3}, {2, 3}},
			}),
		},
		{
			name:  "floating point example",
			bound: toBound(orb.Bound{Min: orb.Point{-91.93359375, 42.29356419217009}, Max: orb.Point{-91.7578125, 42.42345651793831}}),
			input: toLineString(orb.LineString{
				{-86.66015624999999, 42.22851735620852}, {-81.474609375, 38.51378825951165},
				{-85.517578125, 37.125286284966776}, {-85.8251953125, 38.95940879245423},
				{-90.087890625, 39.53793974517628}, {-91.93359375, 42.32606244456202},
				{-86.66015624999999, 42.22851735620852},
			}),
			output: toMls(orb.MultiLineString{
				{
					{-91.91208030440808, 42.29356419217009},
					{-91.93359375, 42.32606244456202},
				}, {
					{-91.93359375, 42.32606244456202},
					{-91.7578125, 42.3228109416169},
				},
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := LineString(tc.bound, tc.input, OpenBound(true))
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect clip")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}
}

func TestRing_CompletelyOutside(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LinearRing
		output *geom.LinearRing
	}{
		{
			name:  "bound in lower left",
			bound: toBound(orb.Bound{Min: orb.Point{-1, -1}, Max: orb.Point{0, 0}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound in lower right",
			bound: toBound(orb.Bound{Min: orb.Point{3, -1}, Max: orb.Point{4, 0}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound in upper right",
			bound: toBound(orb.Bound{Min: orb.Point{3, 3}, Max: orb.Point{4, 4}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound in upper left",
			bound: toBound(orb.Bound{Min: orb.Point{-1, 3}, Max: orb.Point{0, 4}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound to the left",
			bound: toBound(orb.Bound{Min: orb.Point{-1, -1}, Max: orb.Point{0, 3}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound to the right",
			bound: toBound(orb.Bound{Min: orb.Point{3, -1}, Max: orb.Point{4, 3}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound to the top",
			bound: toBound(orb.Bound{Min: orb.Point{-1, 3}, Max: orb.Point{3, 3}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound to the bottom",
			bound: toBound(orb.Bound{Min: orb.Point{-1, -1}, Max: orb.Point{3, 0}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Ring(tc.bound, tc.input)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect clip")
				t.Logf("%v %+v", result == nil, result)
				t.Logf("%v %+v", tc.output == nil, tc.output)
			}
		})
	}
}
