package mvt

import (
	"math"
	"math/bits"

	"github.com/fredbi/geo/pkg/maptile"
	"github.com/fredbi/geo/pkg/project"

	"github.com/fredbi/geo/internal/mercator"
	"github.com/twpayne/go-geom"
)

type projection struct {
	ToTile  project.Projection
	ToWGS84 project.Projection
}

func newProjection(tile maptile.Tile, extent uint32) *projection {
	if isPowerOfTwo(extent) {
		// powers of two extents allows for some more simplicity
		n := uint32(bits.TrailingZeros32(extent))
		z := uint32(tile.Z) + n

		minx := float64(tile.X << n)
		miny := float64(tile.Y << n)
		return &projection{
			ToTile: func(p geom.Coord) geom.Coord {
				x, y := mercator.ToPlanar(p[0], p[1], z)
				return geom.Coord{
					math.Floor(x - minx),
					math.Floor(y - miny),
				}
			},
			ToWGS84: func(p geom.Coord) geom.Coord {
				lon, lat := mercator.ToGeo(p[0]+minx+0.5, p[1]+miny+0.5, z)
				return geom.Coord{lon, lat}
			},
		}
	}

	return nonPowerOfTwoProjection(tile, extent)
}

func nonPowerOfTwoProjection(tile maptile.Tile, extent uint32) *projection {
	// I really don't know why anyone would use a non-power of two extent,
	// but technically it is supported.
	e := float64(extent)
	z := uint32(tile.Z)

	minx := float64(tile.X)
	miny := float64(tile.Y)
	return &projection{
		ToTile: func(p geom.Coord) geom.Coord {
			x, y := mercator.ToPlanar(p[0], p[1], z)
			return geom.Coord{
				math.Floor((x - minx) * e),
				math.Floor((y - miny) * e),
			}
		},
		ToWGS84: func(p geom.Coord) geom.Coord {
			lon, lat := mercator.ToGeo((p[0]/e)+minx, (p[1]/e)+miny, z)
			return geom.Coord{lon, lat}
		},
	}
}

func isPowerOfTwo(n uint32) bool {
	return (n & (n - 1)) == 0
}
