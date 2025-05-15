package tileprovider

import (
	"math"

	"github.com/go-spatial/geom"
)

func NewTile(z, x, y uint) *Tile {
	return newTile(z, x, y, float64(TileBuffer), WebMercator)
}

func newTile(z, x, y uint, buffer float64, srid uint64) *Tile {
	return &Tile{z: z, x: x, y: y, buffer: buffer, srid: srid}
}

// Tile2Lon will return the west most longitude
func Tile2Lon(x, z uint) float64 { return float64(x)/math.Exp2(float64(z))*360.0 - 180.0 }

// Tile2Lat will return the east most Latitude
func Tile2Lat(y, z uint) float64 {
	n := math.Pi
	if y != 0 {
		n = math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	}

	return 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
}

type Tile struct {
	z, x, y uint
	buffer  float64
	srid    uint64
}

func (t *Tile) Bounds() [4]float64 {
	east := Tile2Lon(t.x, t.z)
	west := Tile2Lon(t.x+1, t.z)
	north := Tile2Lat(t.y, t.z)
	south := Tile2Lat(t.y+1, t.z)

	return [4]float64{east, north, west, south}
}

func (t *Tile) ZXY() (uint, uint, uint) { return t.z, t.x, t.y }

// Extent returns the extent of the tile excluding any buffer
func (t *Tile) Extent() (*geom.Extent, uint64) {
	max := 20037508.34

	// resolution
	res := (max * 2) / math.Exp2(float64(t.z))

	// unbuffered extent
	return geom.NewExtent(
		[2]float64{
			-max + (float64(t.x) * res), // MinX
			max - (float64(t.y) * res),  // Miny
		},
		[2]float64{
			-max + (float64(t.x) * res) + res, // MaxX
			max - (float64(t.y) * res) - res,  // MaxY
		},
	), t.srid
}

// BufferedExtent returns the extent of the tile including any buffer
func (t *Tile) BufferedExtent() (*geom.Extent, uint64) {
	extent, _ := t.Extent()

	mvtTileWidthHeight := float64(MapboxExtent)
	// the bounds / extent
	mvtTileExtent := [4]float64{
		0 - t.buffer, 0 - t.buffer,
		mvtTileWidthHeight + t.buffer, mvtTileWidthHeight + t.buffer,
	}

	xspan := extent.MaxX() - extent.MinX()
	yspan := extent.MaxY() - extent.MinY()

	bufferedExtent := geom.NewExtent(
		[2]float64{
			(mvtTileExtent[0] * xspan / mvtTileWidthHeight) + extent.MinX(),
			(mvtTileExtent[1] * yspan / mvtTileWidthHeight) + extent.MinY(),
		},
		[2]float64{
			(mvtTileExtent[2] * xspan / mvtTileWidthHeight) + extent.MinX(),
			(mvtTileExtent[3] * yspan / mvtTileWidthHeight) + extent.MinY(),
		},
	)
	return bufferedExtent, t.srid
}
