package geom

type (
	// Collection of geometries, which may be manipulated as a geometry itself.
	//
	// Collection is Sortable.
	Collection      []T
	LineCollection  []Line
	PointCollection []Point
	RingCollection  []Ring

	// PolygonCollection is a MultiPolygon geometry
	PolygonCollection []Polygon
)

func (c Collection) Area() LineString { return nil }
func (c Collection) Layout() Layout   { return XY }
func (c Collection) IsEmpty() bool    { return len(c) == 0 }
func (c Collection) Clone() T         { return nil }

// Round all coordinates that define T, defaults to 6 decimals
func (c Collection) Round(...RoundingOption) {}

// Bounds returns the bounding box covering T
func (c Collection) Bounds() *Bounds { return nil }
func (c Collection) Equals(T) bool   { return false }

// Centroid yields the centroid of any geometry
func (c Collection) Centroid() Point {
	return nil
}

func (c Collection) SRID() int   { return 0 }
func (c Collection) SetSRID(int) {}

func (pc PointCollection) LineString() LineString   { return nil }
func (pc PointCollection) Vertices() []Point        { return nil }
func (pc PointCollection) Edges() LineString        { return nil }
func (pc PointCollection) AsLineString() LineString { return nil }
