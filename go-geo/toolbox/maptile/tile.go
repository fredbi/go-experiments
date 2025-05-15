// Package maptile defines a Tile type and methods to work with
// web map projected tile data.
package maptile

import (
	"math"

	"github.com/fredbi/geo/internal/mercator"
	b "github.com/fredbi/geo/pkg/bound"
	"github.com/twpayne/go-geom"
)

const (
	DefaultEpislon    = 10.0
	DefaultExtent     = 4096
	DefaultTileBuffer = 4096.0
	MaxZ              = 22
)

type Tiles []Tile

func NewTileBound() b.Bound {
	return b.Bound{Min: geom.Coord{0, 0}, Max: geom.Coord{DefaultExtent, DefaultExtent}}
}

func NewTileBoundBuff(buffs ...float64) b.Bound {
	var buff float64
	if len(buffs) > 0 {
		buff = buffs[0]
	} else {
		buff = DefaultTileBuffer
	}
	return b.Bound{
		Min: geom.Coord{-buff, -buff},
		Max: geom.Coord{DefaultExtent + DefaultTileBuffer, DefaultExtent + DefaultTileBuffer},
	}
}

// Tile is an x, y, z web mercator tile.
type Tile struct {
	X, Y      uint32
	Z         Zoom
	Tolerance float64
	Extent    float64
	Buffer    float64
}

// A Zoom is a strict type for a tile zoom level.
type Zoom uint32

// New creates a new tile with the given coordinates.
func New(x, y uint32, z Zoom) Tile {
	return Tile{
		Z:         z,
		X:         x,
		Y:         y,
		Buffer:    DefaultTileBuffer,
		Extent:    DefaultExtent,
		Tolerance: DefaultEpislon,
	}
}

// Bound returns the geo bound for the tile.
// An optional tileBuffer parameter can be passes to create a buffer
// around the bound in tile dimension. e.g. a tileBuffer of 1 would create
// a bound 9x the size of the tile, centered around the provided tile.
func (t Tile) Bound(tileBuffer ...float64) b.Bound {
	buffer := 0.0
	if len(tileBuffer) > 0 {
		buffer = tileBuffer[0]
	}

	x := float64(t.X)
	y := float64(t.Y)

	minx := x - buffer

	miny := y - buffer
	if miny < 0 {
		miny = 0
	}
	lon1, lat1 := mercator.ToGeo(minx, miny, uint32(t.Z))

	maxx := x + 1 + buffer

	maxtiles := float64(uint32(1 << t.Z))
	maxy := y + 1 + buffer
	if maxy > maxtiles {
		maxy = maxtiles
	}

	lon2, lat2 := mercator.ToGeo(maxx, maxy, uint32(t.Z))
	return b.Bound{
		Min: geom.Coord{lon1, lat2},
		Max: geom.Coord{lon2, lat1},
	}
}

//ZEpision is used for simplify lib, the smaller the zoom level, the largeer zepision value
// we get
func (t Tile) ZEpislon() float64 {

	if t.Z == MaxZ {
		return 0
	}
	epi := t.Tolerance
	if epi <= 0 {
		return 0
	}
	ext := t.Extent

	denom := (math.Exp2(float64(t.Z)) * ext)

	e := epi / denom
	return e
}
