package geom

// Layout defines the type of geometry used.
//
// The layout mainly determines the dimensions used by this geometry.
// There are some other settings, such as the associated SRID and the
// precision to be used for coordinates.
type Layout uint8

const (
	// NoLayout defined. All geometries will be EmptyGeometry
	NoLayout Layout = iota

	// X is a single dimension geometry (points on a line)
	X

	// XY is a planar 2-dimensional euclidian geometry
	XY

	// XYEarth is a planar geoemtry, like XY, but X are longitudes in [-180,180] and Y are latitudes in [-90,90]
	XYEarth

	// XYSpherical is a spherical geometry, where X are longitudes in [-180,180] and Y are latitudes in [-90,90]
	XYSpherical

	// XYZ is a 3-dimensional euclidian geometry
	XYZ

	// XYZSpherical is a spherical geometry, where X are longitudes in [-180,180], Y are latitudes in [-90,90] and unbounded Z altitudes
	XYZSpherical

	// XYZEarth is a planar geometry,like XYZ, but X are longitudes in [-180,180] and Y are latitudes in [-90,90]. The "altitude" Z is not bounded.
	XYZEarth

	// S2 is a 2-dimensional spherical geometry
	S2

	// S3 is a 3-dimensional spherical geometry, adding a Z dimension to S2 definitions
	S3
)

func (l Layout) String() string {
	switch l {
	case NoLayout:
		return "none"
	case X:
		return "X"
	case XY:
		return "XY"
	case XYSpherical:
		return "XYSpherical"
	case XYEarth:
		return "XYEarth"
	case XYZ:
		return "XYZ"
	case XYZSpherical:
		return "XYZSpherical"
	case XYZEarth:
		return "XYZEarth"
	case S2:
		return "S2"
	case S3:
		return "S3"
	default:
		panic("dev error: invalid layout")
	}
}

// Dimensions yields the dimensionality for a layout
func (l Layout) Dimensions() int {
	switch l {
	case NoLayout:
		return 0
	case X:
		return 1
	case XY, XYSpherical, XYEarth, S2:
		return 2
	case XYZ, XYZSpherical, XYZEarth, S3:
		return 3
	default:
		panic("dev error: invalid layout")
	}
}
