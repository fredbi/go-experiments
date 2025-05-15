package simplify

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/twpayne/go-geom"
)

func TestDouglasPeucker_BenchmarkData(t *testing.T) {
	cases := []struct {
		threshold float64
		length    int
	}{
		{0.1, 1118},
		{0.5, 257},
		{1.0, 144},
		{1.5, 95},
		{2.0, 71},
		{3.0, 46},
		{4.0, 39},
		{5.0, 33},
	}

	ls := benchmarkData()

	for i, tc := range cases {
		r := DouglasPeucker(tc.threshold).LineString(ls.Clone())
		if r.NumCoords() != tc.length {
			t.Errorf("%d: reduced poorly, %d != %d", i, r.NumCoords(), tc.length)
		}
	}
}

func BenchmarkDouglasPeucker(b *testing.B) {
	ls := benchmarkData()

	var data []*geom.LineString
	for i := 0; i < b.N; i++ {
		data = append(data, ls.Clone())
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DouglasPeucker(0.1).LineString(data[i])
	}
}

func benchmarkData() *geom.LineString {
	// Data taken from the simplify-js example at http://mourner.github.io/simplify-js/
	f, err := os.Open("testdata/lisbon2portugal.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var points []float64
	err = json.NewDecoder(f).Decode(&points)
	if err != nil {
		panic(err)
	}
	ls := geom.NewLineString(geom.XY)
	coords := []geom.Coord{}
	for i := 0; i < len(points); i += 2 {
		coords = append(coords, geom.Coord{points[i], points[i+1]})
	}
	ls.MustSetCoords(coords)
	return ls
}
