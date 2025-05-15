package utils

import (
	"github.com/fredbi/go-geom/geom/codes"
	//"github.com/fredbi/go-geom/geom/internal/layouts/s2"
	//"github.com/fredbi/go-geom/geom/internal/layouts/s3"
	"github.com/fredbi/go-geom/geom/internal/layouts/stub"
	"github.com/fredbi/go-geom/geom/internal/layouts/x"
	//"github.com/fredbi/go-geom/geom/internal/layouts/x"
	//"github.com/fredbi/go-geom/geom/internal/layouts/xy"
	//"github.com/fredbi/go-geom/geom/internal/layouts/xyz"
	"github.com/fredbi/go-geom/geom"
)

var factory Factory

func init() {
	factory = defaultFactory()
}

type Factory struct {
	defaults int
}

func (f factory) NewEmptyGeometry(opts ...geom.LayoutOption) geom.EmptyGeometry {
	return stub.EmptyGeometry(opts...)
}

func (f factory) NewPoint(opts ...geom.LayoutOption) geom.Point {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewPoint(opts...)
	case XY, XYEarth:
		return xy.NewPoint(opts...)
	case XYZ, XYZEarth:
		return xyz.NewPoint(opts...)
	case S2:
		return s2.NewPoint(opts...)
	case S3:
		return s3.NewPoint(opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewLine(p1, p2 geom.Point, opts ...geom.LayoutOption) Line {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	//if p1.Layout() != cfg.Layout || p2.Layout() != cfg.Layout {
	//return NewEmptyGeometry(append(opts, WithCause(codes.ErrInconsistentLayout))...)
	//}
	switch cfg.layout {
	case X:
		return x.NewLine(p1, p2, opts...)
	case XY, XYEarth:
		return xy.NewLine(p1, p2, opts...)
	case XYZ, XYZEarth:
		return xyz.NewLine(p1, p2, opts...)
	case S2:
		return s2.NewLine(p1, p2, opts...)
	case S3:
		return s3.NewLine(p1, p2, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewRing(pt []geom.Point, opts ...geom.LayoutOption) geom.Ring {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewRing(pt, opts...)
	case XY, XYEarth:
		return xy.NewRing(pt, opts...)
	case XYZ, XYZEarth:
		return xyz.NewRing(pt, opts...)
	case S2:
		return s2.NewRing(pt, opts...)
	case S3:
		return s3.NewRing(pt, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewLineString(pt []geom.Point, opts ...geom.LayoutOption) geom.LineString {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewLineString(pt, opts...)
	case XY, XYEarth:
		return xy.NewLineString(pt, opts...)
	case XYZ, XYZEarth:
		return xyz.NewLineString(pt, opts...)
	case S2:
		return s2.NewLineString(pt, opts...)
	case S3:
		return s3.NewLineString(pt, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewPolygon(pt []geom.Point, opts ...geom.LayoutOption) geom.Polygon {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewPolygon(pt, opts...)
	case XY, XYEarth:
		return xy.NewPolygon(pt, opts...)
	case XYZ, XYZEarth:
		return xyz.NewPolygon(pt, opts...)
	case S2:
		return s2.NewPolygon(pt, opts...)
	case S3:
		return s3.NewPolygon(pt, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewRectangle(a, b geom.Point, opts ...geom.LayoutOption) geom.Rectangle {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewRectangle(a, b, opts...)
	case XY, XYEarth:
		return xy.NewRectangle(a, b, opts...)
	case XYZ, XYZEarth:
		return xyz.NewRectangle(a, b, opts...)
	case S2:
		return s2.NewRectangle(a, b, opts...)
	case S3:
		return s3.NewRectangle(a, b, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewSquare(o geom.Point, c float64, opts ...geom.LayoutOption) geom.Square {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewSquare(o, c, opts...)
	case XY, XYEarth:
		return xy.NewSquare(o, c, opts...)
	case XYZ, XYZEarth:
		return xyz.NewSquare(o, c, opts...)
	case S2:
		return s2.NewSquare(o, c, opts...)
	case S3:
		return s3.NewSquare(o, c, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewTriangle(a, b, c geom.Point, opts ...geom.LayoutOption) geom.Triangle {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewTriangle(a, b, c, opts...)
	case XY, XYEarth:
		return xy.NewTriangle(a, b, c, opts...)
	case XYZ, XYZEarth:
		return xyz.NewTriangle(a, b, c, opts...)
	case S2:
		return s2.NewTriangle(a, b, c, opts...)
	case S3:
		return s3.NewTriangle(a, b, c, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func (f factory) NewHexagon(o geom.Point, c float64, opts ...geom.LayoutOption) geom.Hexagon {
	cfg := defautLayoutConfig()
	for _, apply := range opts {
		apply(cfg)
	}
	switch cfg.layout {
	case X:
		return x.NewHexagon(o, c, opts...)
	case XY, XYEarth:
		return xy.NewHexagon(o, c, opts...)
	case XYZ, XYZEarth:
		return xyz.NewHexagon(o, c, opts...)
	case S2:
		return s2.NewHexagon(o, c, opts...)
	case S3:
		return s3.NewHexagon(o, c, opts...)
	default:
		return stub.NewEmptyGeometry(opts...).WithCause(codes.ErrUnsupportedLayout)
	}
}

func NewEmptyGeometry(opts ...geom.LayoutOption) geom.EmptyGeometry {
	return factory.NewEmptyGeometry(opts...)
}
