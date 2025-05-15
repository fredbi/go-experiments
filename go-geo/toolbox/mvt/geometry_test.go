package mvt

import (
	"fmt"
	"reflect"
	"testing"

	vectortile "github.com/fredbi/geo/pkg/mvt/vtile"
	"github.com/twpayne/go-geom"
)

func TestGeometry_Point(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.Point
	}{
		{
			name:   "basic point",
			input:  []uint32{9, 50, 34},
			output: geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{25, 17}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_POINT, tc.input, tc.output)
		})
	}
}

func TestGeometry_MultiPoint(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.MultiPoint
	}{
		{
			name:   "basic multi point",
			input:  []uint32{17, 10, 14, 3, 9},
			output: geom.NewMultiPoint(geom.XY).MustSetCoords([]geom.Coord{geom.Coord{5, 7}, geom.Coord{3, 2}}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_POINT, tc.input, tc.output)
		})
	}
}

func TestGeometry_LineString(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.LineString
	}{
		{
			name:  "basic line string",
			input: []uint32{9, 4, 4, 18, 0, 16, 16, 0},
			output: geom.NewLineString(geom.XY).MustSetCoords(
				[]geom.Coord{
					geom.Coord{2, 2}, geom.Coord{2, 10}, geom.Coord{10, 10},
				},
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_LINESTRING, tc.input, tc.output)
		})
	}
}

func TestGeometry_MultiLineString(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.MultiLineString
	}{
		{
			name:  "basic multi line string",
			input: []uint32{9, 4, 4, 18, 0, 16, 16, 0, 9, 17, 17, 10, 4, 8},
			output: geom.NewMultiLineString(geom.XY).MustSetCoords(
				[][]geom.Coord{
					[]geom.Coord{
						geom.Coord{2, 2}, geom.Coord{2, 10}, geom.Coord{10, 10},
					},
					[]geom.Coord{
						geom.Coord{1, 1}, geom.Coord{3, 5},
					},
				},
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_LINESTRING, tc.input, tc.output)
		})
	}
}

func TestGeometry_Polygon(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.Polygon
	}{
		{
			name:  "basic polygon",
			input: []uint32{9, 6, 12, 18, 10, 12, 24, 44, 15},
			output: geom.NewPolygon(geom.XY).MustSetCoords(
				[][]geom.Coord{
					[]geom.Coord{
						geom.Coord{3, 6}, geom.Coord{8, 12}, geom.Coord{20, 34}, geom.Coord{3, 6},
					},
				},
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_POLYGON, tc.input, tc.output)
		})
	}

	// should encode a ring as a polygon
	ring := geom.NewLinearRing(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{3, 6}, geom.Coord{8, 12}, geom.Coord{20, 34}, geom.Coord{3, 6},
		},
	)

	gt, data, err := encodeGeometry(ring)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if gt != vectortile.Tile_POLYGON {
		t.Errorf("should be polygon type: %v", gt)
	}

	if !reflect.DeepEqual(data, []uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}) {
		t.Errorf("incorrect data: %v", data)
	}

	// should leave work for unclosed rings

	ring = geom.NewLinearRing(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{3, 6}, geom.Coord{8, 12}, geom.Coord{20, 34},
		},
	)
	_, data, _ = encodeGeometry(ring)
	if !reflect.DeepEqual(data, []uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}) {
		t.Errorf("incorrect data: %v", data)
	}

	p := geom.NewPolygon(geom.XY)
	_ = p.Push(ring)
	_, data, _ = encodeGeometry(p)
	if !reflect.DeepEqual(data, []uint32{9, 6, 12, 18, 10, 12, 24, 44, 15}) {
		t.Errorf("incorrect data: %v", data)
	}
}

func TestGeometry_MultiPolygon(t *testing.T) {
	cases := []struct {
		name   string
		input  []uint32
		output *geom.MultiPolygon
	}{
		{
			name: "multi polygon",
			input: []uint32{9, 0, 0, 26, 20, 0, 0, 20, 19, 0, 15, 9, 22, 2, 26,
				18, 0, 0, 18, 17, 0, 15, 9, 4, 13, 26, 0, 8, 8, 0, 0, 7, 15},
			output: geom.NewMultiPolygon(geom.XY).MustSetCoords(
				[][][]geom.Coord{
					[][]geom.Coord{
						[]geom.Coord{
							geom.Coord{0, 0}, geom.Coord{10, 0}, geom.Coord{10, 10}, geom.Coord{0, 10}, geom.Coord{0, 0},
						},
					},
					[][]geom.Coord{
						[]geom.Coord{
							geom.Coord{11, 11}, geom.Coord{20, 11}, geom.Coord{20, 20}, geom.Coord{11, 20}, geom.Coord{11, 11},
						},
						[]geom.Coord{
							geom.Coord{13, 13}, geom.Coord{13, 17}, geom.Coord{17, 17}, geom.Coord{17, 13}, geom.Coord{13, 13},
						},
					},
				},
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			compareGeometry(t, vectortile.Tile_POLYGON, tc.input, tc.output)
		})
	}
}

func TestKeyValueEncoder_JSON(t *testing.T) {
	kve := newKeyValueEncoder()
	i, err := kve.Value([]int{1, 2, 3})
	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}

	value := decodeValue(kve.Values[i])
	if value != "[1,2,3]" {
		t.Errorf("should encode non standard types as json")
	}
}

type stringer int

func (s stringer) String() string {
	return fmt.Sprintf("%d", s)
}

func TestEncodeValue(t *testing.T) {
	cases := []struct {
		name   string
		input  interface{}
		output interface{}
	}{
		{
			name:   "string",
			input:  "abc",
			output: "abc",
		},
		{
			name:   "stringer",
			input:  stringer(10),
			output: "10",
		},
		{
			name:   "int",
			input:  int(1),
			output: float64(1),
		},
		{
			name:   "int8",
			input:  int8(2),
			output: float64(2),
		},
		{
			name:   "int16",
			input:  int16(3),
			output: float64(3),
		},
		{
			name:   "int32",
			input:  int32(4),
			output: float64(4),
		},
		{
			name:   "int64",
			input:  int64(5),
			output: float64(5),
		},
		{
			name:   "uint",
			input:  int(1),
			output: float64(1),
		},
		{
			name:   "uint8",
			input:  int8(2),
			output: float64(2),
		},
		{
			name:   "uint16",
			input:  int16(3),
			output: float64(3),
		},
		{
			name:   "uint32",
			input:  int32(4),
			output: float64(4),
		},
		{
			name:   "uint64",
			input:  int64(5),
			output: float64(5),
		},
		{
			name:   "float32",
			input:  float32(6),
			output: float64(6),
		},
		{
			name:   "float64",
			input:  float64(7),
			output: float64(7),
		},
		{
			name:   "bool",
			input:  true,
			output: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := encodeValue(tc.input)
			if err != nil {
				t.Fatalf("encode failure: %v", err)
			}

			result := decodeValue(val)
			if !reflect.DeepEqual(result, tc.output) {
				t.Errorf("incorrect value: %[1]T != %[2]T, %[1]v != %[2]v", result, tc.output)
			}
		})
	}

	// error if a weird type, but typical json decode result
	input := map[string]interface{}{
		"a": 1,
		"b": 2,
	}

	_, err := encodeValue(input)
	if err == nil {
		t.Errorf("expecting error: %v", err)
	}
}

func compareGeometry(
	t testing.TB,
	geomType vectortile.Tile_GeomType,
	input []uint32,
	expected geom.T,
) {
	t.Helper()
	// test encoding
	gt, encoded, err := encodeGeometry(expected)
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	if gt != geomType {
		t.Errorf("type mismatch: %v != %v", gt, geomType)
	}
	if !reflect.DeepEqual(encoded, input) {
		t.Logf("%v", encoded)
		t.Logf("%v", input)
		t.Errorf("different encoding")
	}
	result, err := decodeGeometry(geomType, input)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// if result.GeoJSONType() != expected.GeoJSONType() {
	// 	t.Errorf("types different: %s != %s", result.GeoJSONType(), expected.GeoJSONType())
	// }

	if !reflect.DeepEqual(*result.Bounds(), *expected.Bounds()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry bounds not equal")
	}
	if !reflect.DeepEqual(result.Ends(), expected.Ends()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry ends not equal")
	}
	if !reflect.DeepEqual(result.Endss(), expected.Endss()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry endss not equal")
	}
	if !reflect.DeepEqual(result.FlatCoords(), expected.FlatCoords()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry FlatCoords not equal")
	}
	if !reflect.DeepEqual(result.Stride(), expected.Stride()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry stride not equal")
	}
	if !reflect.DeepEqual(result.SRID(), expected.SRID()) {
		t.Logf("%v", result)
		t.Logf("%v", expected)
		t.Errorf("geometry SRID not equal")
	}

}
