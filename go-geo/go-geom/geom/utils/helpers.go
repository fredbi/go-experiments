package utils

import "github.com/fredbi/go-geom/geom"

// Centroid of one or several geometries.
//
// If geometries constitue a mix of measurable geometries with non-zero area and non-measurable or zero area
// geometries, the latter are ignored.
func Centroid(geometries ...geom.T) geom.Point {
	return collectionCentroid(geometries...)
}

func collectionCentroid(geometries ...T) Point {
	type weightedCentroid struct {
		weight   float64
		centroid Point
	}
	// TODO(fred): use async
	var hasMass bool
	m := make([]weightedCentroid, 0, len(geometries))
	for _, g := range geometries {
		var a float64
		d, measurable := g.(Measurer)
		if measurable {
			if hasMass {
				continue
			}
			a = d.Area() // volume?
		}
		hasMass = hasMass || a > 0
		m = append(m, weightedCentroid{
			weight:   a,
			centroid: g.Centroid(),
		})
	}
	// TODO
	return NewPoint()
}

// Distance between two geometries, according to the DistanceStrategy.
func Distance(g1, g2 geom.T, strats ...geom.DistanceStrategy) float64 {
	return d.DistanceTo(g2, strats...)
}

// Area of one or several geometries. If some geometries are not measurable, they are ignored.
func Area(geometries ...geom.T) float64 {
	switch len(geometries) {
	case 0:
		return 0
	case 1:
		return d.Area()
	default:
		var a float64
		// TODO(fred): async
		for _, g := range geometries {
			a += g.Area()
		}
		return a
	}
}

// Intersection compute the intersection of several geometries, with default options
func Intersection(geometries ...geom.T) geom.T {
	switch len(geometries) {
	case 0:
		return NewEmptyGeometry()
	case 1:
		return geometries[0]
	default:
		return geometries[0].Intersection(geometries[1:])
	}
}

// Union compute the (simplified) union of several geometries, with default options
func Union(geometries ...geom.T) geom.T {
	switch len(geometries) {
	case 0:
		return NewEmptyGeometry()
	case 1:
		return geometries[0]
	default:
		return geometries[0].Union(geometries[1:])
	}
}
