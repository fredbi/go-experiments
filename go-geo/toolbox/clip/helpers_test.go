package clip

import (
	"reflect"
	"testing"

	bound "github.com/fredbi/geo/pkg/bound"
	"github.com/paulmach/orb"
	"github.com/twpayne/go-geom"
)

func TestGeometry(t *testing.T) {
	bound := toBound(orb.Bound{Min: orb.Point{-1, -1}, Max: orb.Point{1, 1}})

	cases := []struct {
		name   string
		input  geom.T
		output geom.T
	}{
		{
			name:   "only one multipoint in bound",
			input:  toMPoint(orb.MultiPoint{{0, 0}, {5, 5}}),
			output: toPoint(orb.Point{0, 0}),
		},
		{
			name: "only one multilinestring in bound",
			input: toMls(orb.MultiLineString{
				{{0, 0}, {5, 5}},
				{{6, 6}, {7, 7}},
			}),
			output: toLineString(orb.LineString{{0, 0}, {1, 1}}),
		},
		{
			name: "only one multipolygon in bound",
			input: toMultiPolygon(orb.MultiPolygon{
				{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}},
				{{{2, 2}, {3, 2}, {3, 3}, {2, 3}, {2, 2}}},
			}),
			output: toPolygon(orb.Polygon{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Geometry(bound, tc.input)

			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("not equal")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}
}

func TestRing(t *testing.T) {
	cases := []struct {
		name   string
		bound  bound.Bound
		input  *geom.LinearRing
		output *geom.LinearRing
	}{
		{
			name:  "regular clip",
			bound: toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{1.5, 1.5}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: toRing(orb.Ring{
				{1, 1}, {1.5, 1}, {1.5, 1.5}, {1, 1.5}, {1, 1},
			}),
		},
		{
			name:  "bound to the top",
			bound: toBound(orb.Bound{Min: orb.Point{-1, 3}, Max: orb.Point{3, 4}}),
			input: toRing(orb.Ring{
				{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1},
			}),
			output: nil,
		},
		{
			name:  "bound in lower left",
			bound: toBound(orb.Bound{Min: orb.Point{-1, -1}, Max: orb.Point{0, 0}}),
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
				t.Errorf("not equal")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}
}

func TestMultiLineString(t *testing.T) {
	bound := toBound(orb.Bound{Min: orb.Point{0, 0}, Max: orb.Point{2, 2}})
	cases := []struct {
		name   string
		open   bool
		input  *geom.MultiLineString
		output *geom.MultiLineString
	}{
		{
			name: "regular closed bound clip",
			input: toMls(orb.MultiLineString{
				{{1, 1}, {2, 1}, {2, 2}, {3, 3}},
			}),
			output: toMls(orb.MultiLineString{
				{{1, 1}, {2, 1}, {2, 2}, {2, 2}},
			}),
		},
		{
			name: "open bound clip",
			open: true,
			input: toMls(orb.MultiLineString{
				{{1, 1}, {2, 1}, {2, 2}, {3, 3}},
			}),
			output: toMls(orb.MultiLineString{
				{{1, 1}, {2, 1}},
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := MultiLineString(bound, tc.input, OpenBound(tc.open))

			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("not equal")
				t.Logf("%v", result)
				t.Logf("%v", tc.output)
			}
		})
	}

}

// func TestBound(t *testing.T) {
// 	cases := []struct {
// 		name string
// 		b1   orb.Bound
// 		b2   orb.Bound
// 		rs   orb.Bound
// 	}{
// 		{
// 			name: "normal intersection",
// 			b1:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			b2:   orb.Bound{Min: orb.Point{1, 2}, Max: orb.Point{4, 5}},
// 			rs:   orb.Bound{Min: orb.Point{1, 2}, Max: orb.Point{3, 4}},
// 		},
// 		{
// 			name: "1 contains 2",
// 			b1:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			b2:   orb.Bound{Min: orb.Point{1, 2}, Max: orb.Point{2, 3}},
// 			rs:   orb.Bound{Min: orb.Point{1, 2}, Max: orb.Point{2, 3}},
// 		},
// 		{
// 			name: "no overlap",
// 			b1:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			b2:   orb.Bound{Min: orb.Point{4, 5}, Max: orb.Point{5, 6}},
// 			rs:   orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{0, 0}}, // empty
// 		},
// 		{
// 			name: "same bound",
// 			b1:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			b2:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			rs:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 		},
// 		{
// 			name: "1 is empty",
// 			b1:   orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{0, 0}},
// 			b2:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 			rs:   orb.Bound{Min: orb.Point{0, 1}, Max: orb.Point{3, 4}},
// 		},
// 		{
// 			name: "both are empty",
// 			b1:   orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{0, 0}},
// 			b2:   orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{0, 0}},
// 			rs:   orb.Bound{Min: orb.Point{1, 1}, Max: orb.Point{0, 0}},
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			r1 := Bound(tc.b1, tc.b2)
// 			r2 := Bound(tc.b1, tc.b2)

// 			if tc.rs.IsEmpty() && (!r1.IsEmpty() || !r2.IsEmpty()) {
// 				t.Errorf("should be empty")
// 				t.Logf("%v", r1)
// 				t.Logf("%v", r2)
// 			}

// 			if !tc.rs.IsEmpty() {
// 				if !r1.Equal(tc.rs) {
// 					t.Errorf("r1 not equal")
// 					t.Logf("%v", r1)
// 					t.Logf("%v", tc.rs)
// 				}
// 				if !r2.Equal(tc.rs) {
// 					t.Errorf("r2 not equal")
// 					t.Logf("%v", r2)
// 					t.Logf("%v", tc.rs)
// 				}
// 			}
// 		})
// 	}
// }
