package bound

//NO TEST
import (
	"github.com/twpayne/go-geom"
)

type Bound struct {
	Min, Max geom.Coord
}

func (b Bound) Contains(coord geom.Coord) bool {
	if coord[1] < b.Min[1] || b.Max[1] < coord[1] {
		return false
	}

	if coord[0] < b.Min[0] || b.Max[0] < coord[0] {
		return false
	}

	return true
}

func NewBound(minx, miny, maxx, maxy float64) Bound {
	return Bound{Min: geom.Coord{minx, miny}, Max: geom.Coord{maxx, maxy}}
}
