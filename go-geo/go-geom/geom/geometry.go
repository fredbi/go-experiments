package geom

type (
	// EmptyGeometry is an empty geometry
	EmptyGeometry interface {
		T

		// whenever a method returns an EmptyGeometry, an error may be
		// the cause for it
		Cause() error
	}

	Point interface {
		T

		Coords() []float64
		SetCoords([]float64) error

		WithFlatCoords([][]float64, ...func(error)) Point
		// WithCoords sets the coordinates of a point, with some possible error callback
		// (e.g. when Layouts don't match, coordinates don't match the layout constraints, ...).
		//
		// Panics if the callback is not defined and an error occurs.
		WithCoords([]float64, ...func(error)) Point
	}

	Line interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Line

		Ends() [2][]float64
		SetEnds([]float64, []float64) error
		WithEnds([]float64, []float64, ...func(error)) Line
	}

	// LineString is a continuous set of Lines
	LineString interface {
		T

		WithFlatCoords([][]float64, ...func(error)) LineString

		// AddPoints adds up new lines defined by their new end point
		AddPoints(...Point)
		WithPoints(...Point) LineString

		IsRing() bool

		// AsRing closes the LineString and makes it a Ring
		AsRing() Ring
	}

	// Ring determines a simple polygon with no holes
	Ring interface {
		T

		// AsPolygon makes a polygon from a single Ring
		AsPolygon() Polygon
	}

	Polygon interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Polygon
		ExteriorRing() Ring
		InteriorRings() []Ring
		InteriorRing(int) Ring
	}

	// Bounds represent a bounding box
	Bounds interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Bounds
		Extends(T) Bounds
		SetMinMax([]float64, []float64) error
		WithMinMax([]float64, []float64, ...func(error)) Bounds

		AsRectangle() Rectangle
	}

	Triangle interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Triangle

		tesselate(T)
	}

	Rectangle interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Rectangle
		AsBounds() Bounds

		tesselate(T)
	}

	Square interface {
		T

		WithFlatCoords([][]float64, ...func(error)) Square
		AsBounds() Bounds

		tesselate(T)
	}

	Hexagon interface {
		T
		WithFlatCoords([][]float64, ...func(error)) Hexagon

		tesselate(T)
	}

	// Arc interface{}
	//Ellipse   interface{}
	//Circle    interface{}
	// Cap interface{}
)
