package simplify

import (
	"github.com/twpayne/go-geom"
)

type Simplifier interface {
	Simplify(g geom.T) geom.T
	LineString(ls *geom.LineString) *geom.LineString
	MultiLineString(mls *geom.MultiLineString) *geom.MultiLineString
	Ring(r *geom.LinearRing) *geom.LinearRing
	Polygon(p *geom.Polygon) *geom.Polygon
	MultiPolygon(mp *geom.MultiPolygon) *geom.MultiPolygon
	Collection(c *geom.GeometryCollection) *geom.GeometryCollection
}
type simplifier interface {
	simplify(*geom.LineString, bool) (*geom.LineString, []int)
}

func simplify(s simplifier, ge geom.T) geom.T {
	if ge == nil {
		return nil
	}

	switch g := ge.(type) {
	case *geom.Point:
		return g
	case *geom.MultiPoint:
		if g == nil {
			return nil
		}
		return g
	case *geom.LineString:
		g = lineString(s, g)
		if len(g.Coords()) == 0 {
			return nil
		}
		return g
	case *geom.MultiLineString:
		g = multiLineString(s, g)
		if len(g.Coords()) == 0 {
			return nil
		}
		return g
	case *geom.LinearRing:
		g = ring(s, g)
		if len(g.Coords()) == 0 {
			return nil
		}
		return g
	case *geom.Polygon:
		g = polygon(s, g)
		if len(g.Coords()) == 0 {
			return nil
		}
		return g
	case *geom.MultiPolygon:
		g = multiPolygon(s, g)
		if len(g.Coords()) == 0 {
			return nil
		}
		return g
	case *geom.GeometryCollection:
		g = collection(s, g)
		if g.NumGeoms() == 0 {
			return nil
		}
		return g

	}

	panic("unsupported type")
}

func lineString(s simplifier, ls *geom.LineString) *geom.LineString {
	return runSimplify(s, ls)
}

func multiLineString(s simplifier, mls *geom.MultiLineString) *geom.MultiLineString {
	newMls := geom.NewMultiLineString(mls.Layout())
	for i := 0; i < mls.NumLineStrings(); i++ {
		ls := runSimplify(s, mls.LineString(i))
		err := newMls.Push(ls)
		if err != nil {
			panic("error when push linestring")
		}
	}
	return mls
}

func ring(s simplifier, r *geom.LinearRing) *geom.LinearRing {
	coords := r.Coords()
	ls := geom.NewLineString(r.Layout()).MustSetCoords(coords)
	ls = runSimplify(s, ls)
	r.MustSetCoords(ls.Coords())
	return r
}

func polygon(s simplifier, p *geom.Polygon) *geom.Polygon {
	newp := geom.NewPolygon(p.Layout())
	for i := 0; i < p.NumLinearRings(); i++ {
		r := ring(s, p.LinearRing(i))
		if i != 0 && r.NumCoords() <= 2 {
			continue
		}

		err := newp.Push(r)
		if err != nil {
			panic("error when push linearRing")
		}
	}
	return newp
}

func multiPolygon(s simplifier, mp *geom.MultiPolygon) *geom.MultiPolygon {
	newMp := geom.NewMultiPolygon(mp.Layout())
	for i := 0; i < mp.NumPolygons(); i++ {
		p := polygon(s, mp.Polygon(i))
		if p.NumCoords() <= 2 {
			continue
		}
		err := newMp.Push(p)
		if err != nil {
			panic("error when push polygon")
		}
	}
	return newMp
}

func collection(s simplifier, c *geom.GeometryCollection) *geom.GeometryCollection {
	newC := geom.NewGeometryCollection()
	newC.SetSRID(c.SRID())
	for i := 0; i < c.NumGeoms(); i++ {
		g := simplify(s, c.Geom(i))
		err := newC.Push(g)
		if err != nil {
			panic("error when push Geom")
		}
	}
	return newC
}

func runSimplify(s simplifier, ls *geom.LineString) *geom.LineString {
	if ls.NumCoords() <= 2 {
		return ls
	}
	ls, _ = s.simplify(ls, false)
	return ls
}
