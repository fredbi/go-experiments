package stub

import (
	"errors"

	"github.com/fredbi/go-geom/geom"
)

var (
	_ geom.EmptyGeometry = &EmptyGeometry{}
	_ geom.T             = &NotImplementedGeometry{}

	ErrNotImplemented = errors.New("feature not implemented")
)

type (
	EmptyGeometry struct {
		cause error

		EmptySorter
		EmptySurveyor
		EmptyFeaturist
		EmptyTopologist
		EmptyOperator
		EmptyProjector
		EmptyClusterizer
	}

	EmptyFeaturist   struct{}
	EmptySorter      struct{}
	EmptySurveyor    struct{}
	EmptyTopologist  struct{}
	EmptyProjector   struct{}
	EmptyOperator    struct{}
	EmptyClusterizer struct{}

	NotImplementedGeometry struct {
		NotImplementedSorter
		NotImplementedSurveyor
		NotImplementedFeaturist
		NotImplementedTopologist
		NotImplementedOperator
		NotImplementedProjector
		NotImplementedClusterizer
	}

	NotImplementedFeaturist   struct{}
	NotImplementedSorter      struct{}
	NotImplementedSurveyor    struct{}
	NotImplementedTopologist  struct{}
	NotImplementedProjector   struct{}
	NotImplementedOperator    struct{}
	NotImplementedClusterizer struct{}
)

func NewEmptyGeometry(_ geom.LayoutOption) *EmptyGeometry {
	return &EmptyGeometry{}
}

// WithCause is used internally to catch errors
func (e *EmptyGeometry) WithCause(err error) *EmptyGeometry {
	e.cause = err
	return e
}

func (e EmptyGeometry) Cause() error {
	return e.cause
}

func (e *EmptyGeometry) Layout() geom.Layout                        { return geom.NoLayout }
func (e *EmptyGeometry) IsEmpty() bool                              { return true }
func (e *EmptyGeometry) Clone() geom.T                              { return nil }
func (e *EmptyGeometry) Round(...geom.RoundingOption)               {}
func (e *EmptyGeometry) Bounds() geom.Bounds                        { return nil }
func (e *EmptyGeometry) Equals(geom.T, ...geom.EqualityOption) bool { return false }
func (e *EmptyGeometry) Centroid() geom.Point                       { return nil }
func (e *EmptyGeometry) FlatCoords() [][]float64                    { return nil }
func (e *EmptyGeometry) Vertices() []geom.Point                     { return nil }
func (e *EmptyGeometry) Edges() []geom.Line                         { return nil }
func (e *EmptyGeometry) SetFlatCoords([][]float64) error            { return nil }

func (e *EmptyFeaturist) Features() interface{}     { return nil }
func (e *EmptyFeaturist) SetFeatures(_ interface{}) {}

func (e *EmptySorter) Sort(_ ...geom.SortStrategy) {}

func (e *EmptySurveyor) Area() float64                                           { return 0 }
func (e *EmptySurveyor) SignedArea() float64                                     { return 0 }
func (e *EmptySurveyor) Volume() float64                                         { return 0 }
func (e *EmptySurveyor) SignedVolume() float64                                   { return 0 }
func (e *EmptySurveyor) Length() float64                                         { return 0 }
func (e *EmptySurveyor) DistanceTo(_ geom.T, _ ...geom.DistanceStrategy) float64 { return 0 }
func (e *EmptySurveyor) Angle(_ geom.T) float64                                  { return 0 }

func (e *EmptyTopologist) Interior() geom.T                                         { return nil }
func (e *EmptyTopologist) Border() geom.T                                           { return nil }
func (e *EmptyTopologist) Intersects(geom.T, ...geom.TopologyOption) bool           { return false }
func (e *EmptyTopologist) PointClosestTo(geom.T) geom.Point                         { return nil }
func (e *EmptyTopologist) ShortestLineTo(geom.T) geom.Line                          { return nil }
func (e *EmptyTopologist) IsInside(geom.T, ...geom.TopologyOption) bool             { return false }
func (e *EmptyTopologist) IsOutside(geom.T, ...geom.TopologyOption) bool            { return false }
func (e *EmptyTopologist) IsOn(geom.T, ...geom.TopologyOption) bool                 { return false }
func (e *EmptyTopologist) Intersection(geom.T, ...geom.TopologyOption) geom.T       { return nil }
func (e *EmptyTopologist) IntersectionWith([]geom.T, ...geom.TopologyOption) geom.T { return nil }
func (e *EmptyTopologist) Union(geom.T, ...geom.TopologyOption) geom.T              { return nil }
func (e *EmptyTopologist) UnionWith([]geom.T, ...geom.TopologyOption) geom.T        { return nil }

func (e *EmptyOperator) ConvexHull() geom.T                                  { return nil }
func (e *EmptyOperator) Simplify(...geom.SimplificationStrategy) geom.T      { return nil }
func (e *EmptyOperator) Clip(geom.T) geom.T                                  { return nil }
func (e *EmptyOperator) Buffer(float64) geom.T                               { return nil }
func (e *EmptyOperator) Affine(float64, float64, geom.Line) geom.T           { return nil }
func (e *EmptyOperator) Rotate(float64) geom.T                               { return nil }
func (e *EmptyOperator) Translate(geom.Line) geom.T                          { return nil }
func (e *EmptyOperator) Scale(float64) geom.T                                { return nil }
func (e *EmptyOperator) Symmetrical(geom.T, ...geom.SymmetryStrategy) geom.T { return nil }
func (e *EmptyOperator) Tesselate(geom.Tesselator, ...geom.TesselateOption) geom.PolygonCollection {
	return nil
}

func (e *EmptyProjector) ProjectOn(geom.ProjectionStrategy) geom.T { return nil }

func (e *EmptyClusterizer) Clusterize(geom.ClusteringStrategy) geom.T { return nil }

func (e *NotImplementedGeometry) Layout() geom.Layout                        { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) IsEmpty() bool                              { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Clone() geom.T                              { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Round(...geom.RoundingOption)               {}
func (e *NotImplementedGeometry) Bounds() geom.Bounds                        { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Equals(geom.T, ...geom.EqualityOption) bool { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Centroid() geom.Point                       { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) FlatCoords() [][]float64                    { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Vertices() []geom.Point                     { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) Edges() []geom.Line                         { panic(ErrNotImplemented) }
func (e *NotImplementedGeometry) SetFlatCoords([][]float64) error            { panic(ErrNotImplemented) }

func (e *NotImplementedFeaturist) Features() interface{}     { panic(ErrNotImplemented) }
func (e *NotImplementedFeaturist) SetFeatures(_ interface{}) {}

func (e *NotImplementedSorter) Sort(_ ...geom.SortStrategy) {}

func (e *NotImplementedSurveyor) Area() float64         { panic(ErrNotImplemented) }
func (e *NotImplementedSurveyor) SignedArea() float64   { panic(ErrNotImplemented) }
func (e *NotImplementedSurveyor) Volume() float64       { panic(ErrNotImplemented) }
func (e *NotImplementedSurveyor) SignedVolume() float64 { panic(ErrNotImplemented) }
func (e *NotImplementedSurveyor) Length() float64       { panic(ErrNotImplemented) }
func (e *NotImplementedSurveyor) DistanceTo(_ geom.T, _ ...geom.DistanceStrategy) float64 {
	panic(ErrNotImplemented)
}
func (e *NotImplementedSurveyor) Angle(_ geom.T) float64 { panic(ErrNotImplemented) }

func (e *NotImplementedTopologist) Interior() geom.T { panic(ErrNotImplemented) }
func (e *NotImplementedTopologist) Border() geom.T   { panic(ErrNotImplemented) }
func (e *NotImplementedTopologist) Intersects(geom.T, ...geom.TopologyOption) bool {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) PointClosestTo(geom.T) geom.Point { panic(ErrNotImplemented) }
func (e *NotImplementedTopologist) ShortestLineTo(geom.T) geom.Line  { panic(ErrNotImplemented) }
func (e *NotImplementedTopologist) IsInside(geom.T, ...geom.TopologyOption) bool {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) IsOutside(geom.T, ...geom.TopologyOption) bool {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) IsOn(geom.T, ...geom.TopologyOption) bool { panic(ErrNotImplemented) }
func (e *NotImplementedTopologist) Intersection(geom.T, ...geom.TopologyOption) geom.T {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) IntersectionWith([]geom.T, ...geom.TopologyOption) geom.T {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) Union(geom.T, ...geom.TopologyOption) geom.T {
	panic(ErrNotImplemented)
}
func (e *NotImplementedTopologist) UnionWith([]geom.T, ...geom.TopologyOption) geom.T {
	panic(ErrNotImplemented)
}

func (e *NotImplementedOperator) ConvexHull() geom.T { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Simplify(...geom.SimplificationStrategy) geom.T {
	panic(ErrNotImplemented)
}
func (e *NotImplementedOperator) Clip(geom.T) geom.T                        { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Buffer(float64) geom.T                     { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Affine(float64, float64, geom.Line) geom.T { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Rotate(float64) geom.T                     { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Translate(geom.Line) geom.T                { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Scale(float64) geom.T                      { panic(ErrNotImplemented) }
func (e *NotImplementedOperator) Symmetrical(geom.T, ...geom.SymmetryStrategy) geom.T {
	panic(ErrNotImplemented)
}
func (e *NotImplementedOperator) Tesselate(geom.Tesselator, ...geom.TesselateOption) geom.PolygonCollection {
	panic(ErrNotImplemented)
}

func (e *NotImplementedProjector) ProjectOn(geom.ProjectionStrategy) geom.T { panic(ErrNotImplemented) }

func (e *NotImplementedClusterizer) Clusterize(geom.ClusteringStrategy) geom.T {
	panic(ErrNotImplemented)
}
