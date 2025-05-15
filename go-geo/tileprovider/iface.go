package tileprovider

import (
	"github.com/go-spatial/tegola"
	"github.com/go-spatial/tegola/provider"
)

const (
	WebMercator  = 3857
	WGS84        = 4326
	MapboxExtent = 4096
	MaxZoom      = tegola.MaxZ
)

var (
	WebMercatorBounds        = [4]float64{-20026376.39, -20048966.10, 20026376.39, 20048966.10}
	WGS84Bounds              = [4]float64{-180.0, -85.0511, 180.0, 85.0511}
	TileBuffer        uint64 = 64
)

type Tiler interface {
	provider.Tiler
	NewLayer(string) LayerBuilder
}

type LayerBuilder interface {
	WithGeomType(string) LayerBuilder
	WithSRID(uint64) LayerBuilder
	WithGeomField(string) LayerBuilder
	WithQuery(string) LayerBuilder
	WithIDField(string) LayerBuilder
	Save() error
}
