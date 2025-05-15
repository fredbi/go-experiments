package geom

type (
	// T exposes full geometry capabilities
	T interface {
		BaseGeometry

		Sorter
		Clusterizer

		Surveyor
		Projector
		Operator

		Topologist
		Featurist
	}

	// BaseGeometry knows how to perform basic operations on all geometries
	BaseGeometry interface {
		// Layout yields the layout of the current geometry
		Layout() Layout

		// IsEmpty tells whether the current geometry is actually empty
		IsEmpty() bool

		// Clone as a new geometry
		Clone() T

		// Equals another geometry, with some tolerance according to the precision on coordinates
		Equals(T, ...EqualityOption) bool

		// Round all coordinates that define T, defaults to the layout precision (default is 6 decimals)
		Round(...RoundingOption)

		// Bounds returns the bounding box covering T
		Bounds() Bounds

		FlatCoords() [][]float64
		SetFlatCoords([][]float64) error

		// Centroid yields the centroid of the current geometry
		Centroid() Point

		// Vertices returns all vertices that constitute the geometry
		Vertices() []Point

		// Edges returns all edges as Lines. Single Points return an empty list.
		Edges() []Line
	}

	// Topologist knows about the topological properties of a geometry
	Topologist interface {
		Interior() T
		Border() T

		Intersects(T, ...TopologyOption) bool

		// The Point of the current geometry closest to T
		PointClosestTo(T) Point

		// The shortest Line from the current geometry that touches T
		ShortestLineTo(T) Line

		IsInside(T, ...TopologyOption) bool
		IsOutside(T, ...TopologyOption) bool

		// On returns true when T lies on the border of the current geometry
		IsOn(T, ...TopologyOption) bool

		// Intersection of a geometry with some other geometries
		Intersection(T, ...TopologyOption) T
		IntersectionWith([]T, ...TopologyOption) T

		// Union of a geometry with some other geometries
		Union(T, ...TopologyOption) T
		UnionWith([]T, ...TopologyOption) T
	}

	// Sorter knows how to sort vertices on a geometry. TODO: Less/Swap/Len
	Sorter interface {
		Sort(...SortStrategy)
	}

	// Surveyor knows how to measure a geometry, such as comuting
	// area, signed area (taking into account the orientation of rings),
	// distance (with various distance strategies) and angle.
	Surveyor interface {
		Area() float64
		SignedArea() float64
		Length() float64
		DistanceTo(T, ...DistanceStrategy) float64
		Angle(T) float64
		Volume() float64
		SignedVolume() float64
	}

	// Operator knows how to proceed to various operationns and transforms on geometries
	Operator interface {
		ConvexHull() T
		Simplify(...SimplificationStrategy) T
		Clip(T) T
		Buffer(float64) T

		Affine(float64, float64, Line) T
		Rotate(float64) T
		Translate(Line) T
		Scale(float64) T
		Symmetrical(T, ...SymmetryStrategy) T

		// Tesselate a geometry using a base Tesselator. This returns a collection
		// of geometries of the Tesselator type that cover the original geometry.
		Tesselate(Tesselator, ...TesselateOption) PolygonCollection
	}

	// Projector knows how to project a geometry onto another one.
	//
	// ProjectionStrategy is orthogonal projection.
	Projector interface {
		ProjectOn(ProjectionStrategy) T
	}

	// Clusterizer knows how to clusterize a geometry
	Clusterizer interface {
		Clusterize(ClusteringStrategy) T
	}

	// Featurist knows how to embed data as features within geometries.
	// Features may be used by encoders (e.g. geojson encoding).
	Featurist interface {
		Features() interface{}
		SetFeatures(interface{})
	}
)

type (
	// SimplificationStrategy knows how to simplify geometries.
	SimplificationStrategy interface {
		simplify(T) T
	}

	// ClusteringStrategy knows how to clusterize geometries
	ClusteringStrategy interface {
		clusterize(T) T
	}

	// DistanceStrategy knows how to compute a distance between geometries
	DistanceStrategy interface {
		distance(T, T) float64
	}

	// SortStrategy knows how to sort geometries.
	//
	// There are two kinds of sorting:
	// - sorting the points of one geometry
	// - sorting a collection of geometries
	SortStrategy interface {
		sortSingle(T)
		sortMany(...T)
	}

	// SymmetryStrategy knows about symmetry operator
	SymmetryStrategy interface {
		symmetrical(T, T) T
	}

	// ProjectionStrategy knows how to project a geometry onto another one
	ProjectionStrategy interface {
		project(T, T) T
	}

	// Tesselator is a base geometry that may be used to tesselate other geometries.
	//
	// Supported Tesselators are:
	//   * Square
	//   * Hexagon
	//   * Triangle
	Tesselator interface {
		tesselate(T) T
	}
)
