package geom

import (
	"github.com/fredbi/go-geom/geom/internal/options"
)

type (
	// Options to build and use geometries

	// LayoutOption configures a Layout
	LayoutOption func(options.Layout)

	// EqualityOption configures the behavior of the Equal operator
	EqualityOption func(options.Equality)

	// RoundingOption configures the precision for geometry coordinates and operators
	RoundingOption func(options.Rounding)

	// TopologyOption configures the behavior the topological operators
	TopologyOption func(options.Topology)

	// TesselateOption configures a Tesselator
	TesselateOption func(options.Tesselator)
)

func WithSRID(srid uint32) LayoutOption {
	return options.WithSRID(srid)
}

func WithPrecision(precision uint32) RoundingOption {
	return options.WithPrecision(precision)
}
