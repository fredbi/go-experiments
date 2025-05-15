package project

import (
	"fmt"
	"math"

	"github.com/twpayne/go-geom"
)

// EarthRadius is the radius of the earth in meters. It is used in geo distance calculations.
// To keep things consistent, this value matches WGS84 Web Mercator (EPSG:3857).
const EarthRadius = 6378137.0 // meters
const earthRadiusPi = EarthRadius * math.Pi

type Projection func(geom.Coord) geom.Coord

// Mercator performs the Spherical Pseudo-Mercator projection used by most web maps.
var Mercator = struct {
	ToWGS84 Projection
}{
	ToWGS84: func(p geom.Coord) geom.Coord {
		return geom.Coord{
			180.0 * p[0] / earthRadiusPi,
			180.0 / math.Pi * (2*math.Atan(math.Exp(p[1]/EarthRadius)) - math.Pi/2.0),
		}
	},
}

// WGS84 is what common uses lon/lat projection.
var WGS84 = struct {
	// ToMercator projections from WGS to Mercator, used by most web maps
	ToMercator Projection
}{
	ToMercator: func(g geom.Coord) geom.Coord {
		y := math.Log(math.Tan((90.0+g[1])*math.Pi/360.0)) * EarthRadius
		return geom.Coord{
			earthRadiusPi / 180.0 * g[0],
			math.Max(-earthRadiusPi, math.Min(y, earthRadiusPi)),
		}
	},
}

// MercatorScaleFactor returns the mercator scaling factor for a given degree latitude.
func MercatorScaleFactor(g geom.Coord) float64 {
	if g[1] < -90.0 || g[1] > 90.0 {
		panic(fmt.Sprintf("project: latitude out of range, given %f", g[1]))
	}

	return 1.0 / math.Cos(g[1]/180.0*math.Pi)
}

// Geometry is a helper to project any geomtry.
func Geometry(g geom.T, proj Projection) (geom.T, error) {
	if g == nil {
		return nil, nil
	}

	switch g := g.(type) {
	case *geom.Point:
		return Point(g, proj)
	case *geom.MultiPoint:
		return MultiPoint(g, proj)
	case *geom.LineString:
		return LineString(g, proj)
	case *geom.MultiLineString:
		return MultiLineString(g, proj)
	case *geom.LinearRing:
		return Ring(g, proj)
	case *geom.Polygon:
		return Polygon(g, proj)
	case *geom.MultiPolygon:
		return MultiPolygon(g, proj)
	case *geom.GeometryCollection:
		return Collection(g, proj)
	}

	panic("geometry type not supported")
}

// Point is a helper to project an a point
func Point(p *geom.Point, proj Projection) (*geom.Point, error) {
	coords := proj(p.Coords())
	return p.SetCoords(coords)
}

// MultiPoint is a helper to project an entire multi point.
func MultiPoint(mp *geom.MultiPoint, proj Projection) (*geom.MultiPoint, error) {
	converted := make([]geom.Coord, mp.NumCoords())
	for i := 0; i < mp.NumCoords(); i++ {
		converted[i] = proj(mp.Coord(i))
	}
	return mp.SetCoords(converted)
}

// LineString is a helper to project an entire line string.
func LineString(ls *geom.LineString, proj Projection) (*geom.LineString, error) {
	n := ls.NumCoords()
	converted := make([]geom.Coord, n)
	for i := 0; i < n; i++ {
		converted[i] = proj(ls.Coord(i))
	}
	return ls.SetCoords(converted)
}

// MultiLineString is a helper to project an entire multi linestring.
func MultiLineString(mls *geom.MultiLineString, proj Projection) (*geom.MultiLineString, error) {
	converted := make([][]geom.Coord, mls.NumLineStrings())
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := mls.LineString(i)
		lsc := make([]geom.Coord, ls.NumCoords())
		for j := 0; j < ls.NumCoords(); j++ {
			lsc[j] = proj(ls.Coord(j))
		}
		converted[i] = lsc
	}

	return mls.SetCoords(converted)
}

// Ring is a helper to project an entire ring.
func Ring(r *geom.LinearRing, proj Projection) (*geom.LinearRing, error) {
	n := r.NumCoords()
	converted := make([]geom.Coord, n)
	for i := 0; i < n; i++ {
		converted[i] = proj(r.Coord(i))
	}
	return r.SetCoords(converted)
}

// Polygon is a helper to project an entire polygon.
func Polygon(p *geom.Polygon, proj Projection) (*geom.Polygon, error) {
	converted := make([][]geom.Coord, p.NumLinearRings())
	for i := 0; i < p.NumLinearRings(); i++ {
		r := p.LinearRing(i)
		rc := make([]geom.Coord, r.NumCoords())
		for j := 0; j < r.NumCoords(); j++ {
			rc[j] = proj(r.Coord(j))
		}
		converted[i] = rc
	}

	return p.SetCoords(converted)
}

// MultiPolygon is a helper to project an entire multi polygon.
func MultiPolygon(mp *geom.MultiPolygon, proj Projection) (*geom.MultiPolygon, error) {
	n := mp.NumPolygons()
	converted := make([][][]geom.Coord, n)

	for i := 0; i < n; i++ {
		p := mp.Polygon(i)
		pc := make([][]geom.Coord, p.NumLinearRings())

		for j := 0; j < p.NumLinearRings(); j++ {
			r := p.LinearRing(j)
			rc := make([]geom.Coord, r.NumCoords())

			for k := 0; k < r.NumCoords(); k++ {
				rc[k] = proj(r.Coord(k))
			}
			pc[j] = rc
		}
		converted[i] = pc
	}
	mp, err := mp.SetCoords(converted)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

// Collection is a helper to project a rectangle.
func Collection(c *geom.GeometryCollection, proj Projection) (*geom.GeometryCollection, error) {
	gc := geom.NewGeometryCollection()
	for i := 0; i < c.NumGeoms(); i++ {
		g, err := Geometry(gc.Geom(i), proj)
		if err != nil {
			return nil, err
		}
		if err = gc.Push(g); err != nil {
			return nil, err
		}
	}
	*c = *gc
	return c, nil
}
