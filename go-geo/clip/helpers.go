// Package clip is a library for clipping geometry to a bounding box.
package clip

import (
	bound "github.com/fredbi/geo/pkg/bound"
	"github.com/twpayne/go-geom"
)

// Geometry will clip the geometry to the bounding box using the
// correct functions for the type.
// This operation will modify the input of '1d or 2d geometry' by using as a
// scratch space so clone if necessary.
func Geometry(b bound.Bound, g geom.T) geom.T {
	if g == nil {
		return nil
	}

	// if !b.Intersects(g.Bound()) {
	// 	return nil
	// }

	switch g := g.(type) {
	case *geom.Point:
		return g // Intersect check above
	case *geom.MultiPoint:
		mp := MultiPoint(b, g)

		if mp == nil {
			return nil
		}
		if mp.NumPoints() == 1 {
			return mp.Point(0)
		}

		return mp
	case *geom.LineString:
		mls := LineString(b, g)

		if mls == nil {
			return nil
		}
		if mls.NumLineStrings() == 1 {
			return mls.LineString(0)
		}
		return mls
	case *geom.MultiLineString:
		mls := MultiLineString(b, g)

		if mls == nil {
			return nil
		}
		if mls.NumLineStrings() == 1 {
			return mls.LineString(0)
		}
		return mls
	case *geom.LinearRing:
		r := Ring(b, g)
		if r == nil {
			return nil
		}

		return r
	case *geom.Polygon:
		p := Polygon(b, g)
		if p == nil {
			return nil
		}

		return p
	case *geom.MultiPolygon:
		mp := MultiPolygon(b, g)

		if mp == nil {
			return nil
		}
		if mp.NumPolygons() == 1 {
			return mp.Polygon(0)
		}
		return mp
	case *geom.GeometryCollection:
		c := Collection(b, g)

		if c == nil {
			return nil
		}
		if c.NumGeoms() == 1 {
			return c.Geom(0)
		}
		// case orb.Bound:
		// 	b = Bound(b, g)
		// 	if b.IsEmpty() {
		// 		return nil
		// 	}

		// 	return b
		// }

		//panic(fmt.Sprintf("geometry type not supported: %T", g))
		return nil

	}
	return nil
}

// MultiPoint returns a new set with the points outside the bound removed.
func MultiPoint(b bound.Bound, mp *geom.MultiPoint) *geom.MultiPoint {
	newMps := geom.NewMultiPoint(mp.Layout())
	coords := mp.Coords()
	for _, coord := range coords {
		if b.Contains(coord) {
			p := geom.NewPoint(mp.Layout()).MustSetCoords(coord)
			err := newMps.Push(p.Clone())
			if err != nil {
				return nil
			}
		}
	}
	return newMps
}

// LineString clips the linestring to the bounding box.
func LineString(b bound.Bound, ls *geom.LineString, opts ...Option) *geom.MultiLineString {
	open := false
	if len(opts) > 0 {
		o := &options{}
		for _, opt := range opts {
			opt(o)
		}

		open = o.openBound
	}

	result := line(b, ls, open)
	if result == nil || result.NumCoords() == 0 {
		return nil
	}

	return result
}

// MultiLineString clips the linestrings to the bounding box
// and returns a linestring union.
func MultiLineString(b bound.Bound, mls *geom.MultiLineString, opts ...Option) *geom.MultiLineString {
	open := false
	if len(opts) > 0 {
		o := &options{}
		for _, opt := range opts {
			opt(o)
		}

		open = o.openBound
	}
	result := geom.NewMultiLineString(mls.Layout())
	var coords [][]geom.Coord
	for i := 0; i < mls.NumLineStrings(); i++ {
		r := line(b, mls.LineString(i), open)

		if r != nil && r.NumCoords() != 0 {
			coords = append(coords, r.Coords()...)
		}

	}
	result.MustSetCoords(coords)
	// for _, ls := range mls {
	// 	r := line(b, ls, open)
	// 	if len(r) != 0 {
	// 		result = append(result, r...)
	// 	}
	// }
	return result
}

// Ring clips the ring to the bounding box and returns another ring.
// This operation will modify the input by using as a scratch space
// so clone if necessary.
func Ring(b bound.Bound, r *geom.LinearRing) *geom.LinearRing {
	result := ring(b, r)
	if result == nil || result.NumCoords() == 0 {
		return nil
	}

	return result
}

// Polygon clips the polygon to the bounding box excluding the inner rings
// if they do not intersect the bounding box.
// This operation will modify the input by using as a scratch space
// so clone if necessary.
func Polygon(b bound.Bound, p *geom.Polygon) *geom.Polygon {
	if p.NumLinearRings() == 0 {
		return nil
	}

	r := Ring(b, p.LinearRing(0))
	if r == nil {
		return nil
	}
	result := geom.NewPolygon(p.Layout())
	err := result.Push(r.Clone())
	if err != nil {
		return nil
	}
	//result := orb.Polygon{r}
	for i := 1; i < p.NumLinearRings(); i++ {
		r := Ring(b, p.LinearRing(i))
		if r != nil {
			err = result.Push(r.Clone())
			if err != nil {
				return nil
			}
			//result = append(result, r)
		}
	}

	return result
}

// MultiPolygon clips the multi polygon to the bounding box excluding
// any polygons if they don't intersect the bounding box.
// This operation will modify the input by using as a scratch space
// so clone if necessary.
func MultiPolygon(b bound.Bound, mp *geom.MultiPolygon) *geom.MultiPolygon {
	result := geom.NewMultiPolygon(mp.Layout())

	for i := 0; i < mp.NumPolygons(); i++ {
		polygon := mp.Polygon(i)
		p := Polygon(b, polygon)
		if p != nil {
			err := result.Push(p.Clone())
			if err != nil {
				return nil
			}
		}
	}
	return result
}

// Collection clips each element in the collection to the bounding box.
// It will exclude elements if they don't intersect the bounding box.
// This operation will modify the input of '2d geometry' by using as a
// scratch space so clone if necessary.
func Collection(b bound.Bound, c *geom.GeometryCollection) *geom.GeometryCollection {
	result := geom.NewGeometryCollection()
	for i := 0; i < c.NumGeoms(); i++ {
		g := c.Geom(i)
		clipped := Geometry(b, g)
		if clipped != nil {
			result.MustPush(clipped)
		}
	}
	// var result orb.Collection
	// for _, g := range c {
	// 	clipped := Geometry(b, g)
	// 	if clipped != nil {
	// 		result = append(result, clipped)
	// 	}
	// }

	return result
}

// Bound intersects the two bounds. May result in an
// empty/degenerate bound.
// func Bound(b, bound orb.Bound) orb.Bound {
// 	if b.IsEmpty() && bound.IsEmpty() {
// 		return bound
// 	}

// 	if b.IsEmpty() {
// 		return bound
// 	} else if bound.IsEmpty() {
// 		return b
// 	}

// 	return orb.Bound{
// 		Min: orb.Point{
// 			math.Max(b.Min[0], bound.Min[0]),
// 			math.Max(b.Min[1], bound.Min[1]),
// 		},
// 		Max: orb.Point{
// 			math.Min(b.Max[0], bound.Max[0]),
// 			math.Min(b.Max[1], bound.Max[1]),
// 		},
// 	}
// }
