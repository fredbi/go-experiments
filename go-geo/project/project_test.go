package project

import (
	"math"
	"reflect"
	"testing"

	"github.com/twpayne/go-geom"
)

var (
	Epsilon = 1e-6

	Cities = [][2]float64{
		{57.09700, 9.85000}, {49.03000, -122.32000}, {39.23500, -76.17490},
		{57.20000, -2.20000}, {16.75000, -99.76700}, {5.60000, -0.16700},
		{51.66700, -176.46700}, {9.00000, 38.73330}, {-34.7666, 138.53670},
		{12.80000, 45.00000}, {42.70000, -110.86700}, {13.48167, 144.79330},
		{33.53300, -81.71700}, {42.53300, -99.85000}, {26.01670, 50.55000},
		{35.75000, -84.00000}, {51.11933, -1.15543}, {82.52000, -62.28000},
		{32.91700, -85.91700}, {31.19000, 29.95000}, {36.70000, 3.21700},
		{34.14000, -118.10700}, {32.50370, -116.45100}, {47.83400, 10.86800},
		{28.25000, 129.70000}, {16.75000, -22.95000}, {31.95000, 35.95000},
		{52.35000, 4.86660}, {13.58670, 144.93670}, {6.90000, 134.15000},
		{40.03000, 32.90000}, {33.65000, -85.78300}, {49.33000, 10.59700},
		{17.13330, -61.78330}, {-23.4333, -70.60000}, {51.21670, 4.40000},
		{29.60000, 35.01000}, {38.58330, -121.48300}, {34.16700, -97.13300},
		{45.60000, 9.15000}, {-18.3500, -70.33330}, {-7.88000, -14.42000},
		{15.28330, 38.90000}, {-25.2333, -57.51670}, {23.96500, 32.82000},
		{-36.8832, 174.75000}, {-38.0333, 144.46670}, {46.03300, 12.60000},
		{41.66700, -72.83300}, {35.45000, 139.45000}}
)

func TestMercator(t *testing.T) {
	for _, city := range Cities {
		g := geom.Coord{
			city[1],
			city[0],
		}
		p := WGS84.ToMercator(g)
		g = Mercator.ToWGS84(p)
		if math.Abs(g[1]-city[0]) > Epsilon {
			t.Errorf("latitude miss match: %f != %f", g[1], city[0])
		}

		if math.Abs(g[0]-city[1]) > Epsilon {
			t.Errorf("longitude miss match: %f != %f", g[0], city[1])
		}
	}
}

func TestMercatorScaleFactor(t *testing.T) {
	cases := []struct {
		name   string
		point  geom.Coord
		factor float64
	}{
		{
			name:   "30 deg",
			point:  geom.Coord{0, 30.0},
			factor: 1.154701,
		},
		{
			name:   "45 deg",
			point:  geom.Coord{0, 45.0},
			factor: 1.414214,
		},
		{
			name:   "60 deg",
			point:  geom.Coord{0, 60.0},
			factor: 2,
		},
		{
			name:   "80 deg",
			point:  geom.Coord{0, 80.0},
			factor: 5.758770,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if f := MercatorScaleFactor(tc.point); math.Abs(tc.factor-f) > Epsilon {
				t.Errorf("incorrect factor: %v != %v", f, tc.factor)
			}
		})
	}
}

var (
	p = geom.NewPoint(geom.XY).MustSetCoords(
		geom.Coord{5, 7},
	)
	multip = geom.NewMultiPoint(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{5, 7}, geom.Coord{3, 2},
		},
	)

	mls = geom.NewMultiLineString(geom.XY).MustSetCoords(
		[][]geom.Coord{
			[]geom.Coord{
				geom.Coord{2, 2}, geom.Coord{2, 10}, geom.Coord{10, 10},
			},
			[]geom.Coord{
				geom.Coord{1, 1}, geom.Coord{3, 5},
			},
		},
	)
	ls = geom.NewLineString(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{-81.60346275, 41.50998572},
			geom.Coord{-81.6033669, 41.50991259},
			geom.Coord{-81.60355599, 41.50976036},
			geom.Coord{-81.6040648, 41.50936811},
			geom.Coord{-81.60404411, 41.50935405},
		},
	)
	plg = geom.NewPolygon(geom.XY).MustSetCoords(
		[][]geom.Coord{
			[]geom.Coord{
				geom.Coord{3, 6}, geom.Coord{8, 12}, geom.Coord{20, 34}, geom.Coord{3, 6},
			},
		},
	)

	ring = geom.NewLinearRing(geom.XY).MustSetCoords(
		[]geom.Coord{
			geom.Coord{3, 6}, geom.Coord{8, 12}, geom.Coord{20, 34}, geom.Coord{3, 6},
		},
	)

	mplg = geom.NewMultiPolygon(geom.XY).MustSetCoords(
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
	)

	geoms = []geom.T{p, multip, ls, mls, plg, ring, mplg}
)

func FakeProjection(g geom.Coord) geom.Coord {
	return g
}

func TestGeomtry(t *testing.T) {
	for _, g := range geoms {
		rtg, err := Geometry(g, FakeProjection)
		if err != nil {
			t.Errorf("error not expected %s", err)
		}
		if !reflect.DeepEqual(g, rtg) {
			t.Errorf("geom not equal")
		}
	}
}
