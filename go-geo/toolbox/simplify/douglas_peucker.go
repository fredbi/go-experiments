package simplify

import (
	"github.com/twpayne/go-geom"
)

//var _ orb.Simplifier = &DouglasPeuckerSimplifier{}

// A DouglasPeuckerSimplifier wraps the DouglasPeucker function.
type DouglasPeuckerSimplifier struct {
	Threshold float64
}

// DouglasPeucker creates a new DouglasPeuckerSimplifier.
func DouglasPeucker(threshold float64) *DouglasPeuckerSimplifier {
	return &DouglasPeuckerSimplifier{
		Threshold: threshold,
	}
}

func (s *DouglasPeuckerSimplifier) simplify(ls *geom.LineString, wim bool) (*geom.LineString, []int) {
	coords := ls.Coords()
	mask := make([]byte, len(coords))
	mask[0] = 1
	mask[len(mask)-1] = 1

	found := dpWorker(coords, s.Threshold, mask)
	var indexMap []int
	if wim {
		indexMap = make([]int, 0, found)
	}

	count := 0
	for i, v := range mask {
		if v == 1 {
			coords[count] = coords[i]
			count++
			if wim {
				indexMap = append(indexMap, i)
			}
		}
	}
	ls.MustSetCoords(coords[:count])
	return ls, indexMap
}

// dpWorker does the recursive threshold checks.
// Using a stack array with a stackLength variable resulted in
// 4x speed improvement over calling the function recursively.
func dpWorker(coord []geom.Coord, threshold float64, mask []byte) int {
	found := 2

	var stack []int
	stack = append(stack, 0, len(coord)-1)

	for len(stack) > 0 {
		start := stack[len(stack)-2]
		end := stack[len(stack)-1]

		// modify the line in place
		maxDist := 0.0
		maxIndex := 0

		for i := start + 1; i < end; i++ {
			dist := DistanceFromSegmentSquared(coord[start], coord[end], coord[i])
			if dist > maxDist {
				maxDist = dist
				maxIndex = i
			}
		}

		if maxDist > threshold*threshold {
			found++
			mask[maxIndex] = 1

			stack[len(stack)-1] = maxIndex
			stack = append(stack, maxIndex, end)
		} else {
			stack = stack[:len(stack)-2]
		}
	}

	return found
}

// DistanceFromSegmentSquared returns point's squared distance from the segement [a, b].
func DistanceFromSegmentSquared(a, b, point geom.Coord) float64 {
	x := a[0]
	y := a[1]
	dx := b[0] - x
	dy := b[1] - y

	if dx != 0 || dy != 0 {
		t := ((point[0]-x)*dx + (point[1]-y)*dy) / (dx*dx + dy*dy)

		if t > 1 {
			x = b[0]
			y = b[1]
		} else if t > 0 {
			x += dx * t
			y += dy * t
		}
	}

	dx = point[0] - x
	dy = point[1] - y

	return dx*dx + dy*dy
}

// Simplify will run the simplification for any geometry type.
func (s *DouglasPeuckerSimplifier) Simplify(g geom.T) geom.T {
	return simplify(s, g)
}

// LineString will simplify the linestring using this simplifier.
func (s *DouglasPeuckerSimplifier) LineString(ls *geom.LineString) *geom.LineString {
	return lineString(s, ls)
}

// MultiLineString will simplify the multi-linestring using this simplifier.
func (s *DouglasPeuckerSimplifier) MultiLineString(mls *geom.MultiLineString) *geom.MultiLineString {
	return multiLineString(s, mls)
}

// Ring will simplify the ring using this simplifier.
func (s *DouglasPeuckerSimplifier) Ring(r *geom.LinearRing) *geom.LinearRing {
	return ring(s, r)
}

// Polygon will simplify the polygon using this simplifier.
func (s *DouglasPeuckerSimplifier) Polygon(p *geom.Polygon) *geom.Polygon {
	return polygon(s, p)
}

// MultiPolygon will simplify the multi-polygon using this simplifier.
func (s *DouglasPeuckerSimplifier) MultiPolygon(mp *geom.MultiPolygon) *geom.MultiPolygon {
	return multiPolygon(s, mp)
}

// Collection will simplify the collection using this simplifier.
func (s *DouglasPeuckerSimplifier) Collection(c *geom.GeometryCollection) *geom.GeometryCollection {
	return collection(s, c)
}
