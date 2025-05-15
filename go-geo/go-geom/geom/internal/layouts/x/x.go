package x

import (
	"github.com/fredbi/go-geom/geom"
	"github.com/fredbi/go-geom/geom/internal/layouts/stub"
)

var (
	_ geom.T          = &T{}
	_ geom.Point      = &Point{}
	_ geom.Line       = &Line{}
	_ geom.Ring       = &Ring{}
	_ geom.LineString = &LineString{}
	_ geom.Polygon    = &Polygon{}
	_ geom.Rectangle  = &Rectangle{}
	_ geom.Square     = &Square{}
	_ geom.Hexagon    = &Hexagon{}
	_ geom.Triangle   = &Triangle{}
)

type coords struct {
	c [1]float64
}

func (c coords) Coords() []float64 {
	return c[:]
}

func (c *coords) SetFlatCoords(in [][]float64) {
}

func (c *coords) FlatCoords() [][]float64 {
	return [][]float64{}
}

// T defines a single dimension geometry (layout X)
type T struct {
	stub.NotImplementedGeometry
}

type Point struct {
	stub.NotImplementedGeometry
	coords
}

func (p *Point) SetCoords(in []float64)               {}
func (p *Point) WithFlatCoords(in [][]float64) *Point { return nil }
func (p *Point) WithCoords(in []float64) *Point       { return nil }

type Line struct {
	stub.NotImplementedGeometry
}

type Ring struct {
	stub.NotImplementedGeometry
}

type LineString struct {
	stub.NotImplementedGeometry
}

type Polygon struct {
	stub.NotImplementedGeometry
}

type Rectangle struct {
	stub.NotImplementedGeometry
}

type Square struct {
	stub.NotImplementedGeometry
}

type Triangle struct {
	stub.NotImplementedGeometry
}

type Hexagon struct {
	stub.NotImplementedGeometry
}

func NewPoint(...geom.LayoutOption) *Point {
	return nil
}

func NewLine(p1, p2 *Point, _ ...geom.LayoutOption) *Line {
	return nil
}

func NewRing(pt []geom.Point, _ geom.LayoutOption) *Ring {
	return nil
}

func NewLineString(pt []geom.Point, _ geom.LayoutOption) *Ring {
	return nil
}

func NewPolygon(pt []geom.Point, _ geom.LayoutOption) *Polygon {
	return nil
}

func NewRectangle(a, b geom.Point, _ geom.LayoutOption) *Rectangle {
	return nil
}

func NewSquare(o geom.Point, c float64, _ geom.LayoutOption) *Square {
	return nil
}

func NewTriangle(a, b, c geom.Point, c float64, _ geom.LayoutOption) *Triangle {
	return nil
}

func NewHexagon(o geom.Point, c float64, _ geom.LayoutOption) *Hexagon {
	return nil
}
