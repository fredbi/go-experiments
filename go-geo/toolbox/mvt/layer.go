package mvt

//.*NO TEST
import (
	"log"
	"time"

	"github.com/fredbi/geo/internal/triangle"
	"github.com/fredbi/geo/pkg/bound"
	"github.com/fredbi/geo/pkg/clip"
	"github.com/fredbi/geo/pkg/maptile"
	"github.com/fredbi/geo/pkg/project"

	"fmt"

	"github.com/fredbi/geo/utils/geojson"
	"github.com/twpayne/go-geom"
)

const (
	// DefaultExtent for mapbox vector tiles. (https://www.mapbox.com/vector-tiles/specification/)
	DefaultExtent       = 4096
	MimeType            = "application/vnd.mapbox-vector-tile"
	TriangleOverlapZoom = 14
)

// Layer is intermediate MVT layer to be encoded/decoded or projected.
type Layer struct {
	Name     string
	Version  uint32
	Extent   uint32
	Features []*geojson.Feature
}

// Layers is a set of layers.
type Layers []*Layer

// NewLayer is a helper to create a Layer from a feature collection
// and a name, it sets the default extent and version to 1.
func NewLayer(name string, fc *geojson.FeatureCollection) *Layer {
	return &Layer{
		Name:     name,
		Version:  2,
		Extent:   DefaultExtent,
		Features: fc.Features,
	}
}

// NewLayers creates a set of layers given a set of feature collections.
func NewLayers(layers map[string]*geojson.FeatureCollection) Layers {
	result := make(Layers, 0, len(layers))
	for name, fc := range layers {
		result = append(result, NewLayer(name, fc))
	}

	return result
}

// ProjectToTile will project all the geometries in the layer
// to tile coordinates based on the extent and the mercator projection.
func (l *Layer) ProjectToTile(tile maptile.Tile) {
	p := newProjection(tile, l.Extent)

	for _, f := range l.Features {
		f.Geometry, _ = project.Geometry(f.Geometry, p.ToTile)
	}
}

// If data is outside the tile and buffer area, then we need to abandon it
func (l *Layer) RemoveDataOutsideBuffer(buffs []float64) {

	newFt := l.Features[:0]
	extend := maptile.DefaultExtent
	var buff float64
	if len(buffs) > 0 {
		buff = buffs[0]
	} else {
		buff = maptile.DefaultTileBuffer
	}
	for i, f := range l.Features {
		if f.Geometry.Bounds().Min(0) > -buff && f.Geometry.Bounds().Max(0) < (float64(extend)+buff) && f.Geometry.Bounds().Min(1) > -buff && f.Geometry.Bounds().Max(1) < float64(extend)+buff {
			newFt = append(newFt, l.Features[i])
		}
	}
	l.Features = newFt
}

// If data is outside the tile and buffer area, then we need to abandon it
func (ls Layers) RemoveDataOutsideBuffer(buffs ...float64) {

	for _, l := range ls {
		l.RemoveDataOutsideBuffer(buffs)
	}
}

// ProjectToWGS84 will project all the geometries backed to WGS84 from
// the extent and mercator projection.
func (l *Layer) ProjectToWGS84(tile maptile.Tile) {
	p := newProjection(tile, l.Extent)
	for _, f := range l.Features {
		f.Geometry, _ = project.Geometry(f.Geometry, p.ToWGS84)
	}
}

// ProjectToTile will project all the geometries in all layers
// to tile coordinates based on the extent and the mercator projection.
func (ls Layers) ProjectToTile(tile maptile.Tile) {
	for _, l := range ls {
		l.ProjectToTile(tile)
	}
}

// ProjectToWGS84 will project all the geometries in all the layers backed
// to WGS84 from the extent and mercator projection.
func (ls Layers) ProjectToWGS84(tile maptile.Tile) {
	for _, l := range ls {
		l.ProjectToWGS84(tile)
	}
}

// ToFeatureCollections converts the layers to sets of geojson
// feature collections.
func (ls Layers) ToFeatureCollections() map[string]*geojson.FeatureCollection {
	result := make(map[string]*geojson.FeatureCollection, len(ls))
	for _, l := range ls {
		result[l.Name] = &geojson.FeatureCollection{
			Features: l.Features,
		}
	}

	return result
}

func (ls Layers) Clip(bound bound.Bound) {
	for _, l := range ls {
		l.Clip(bound)
	}
}

func (l *Layer) Clip(bound bound.Bound) {
	newFt := l.Features[:0]
	for i, f := range l.Features {
		clipped := clip.Geometry(bound, f.Geometry)
		if clipped != nil && len(clipped.FlatCoords()) > 0 {
			l.Features[i].Geometry = clipped
			newFt = append(newFt, l.Features[i])
		}
	}
	l.Features = newFt
}

func (ls Layers) Empty() bool {

	for _, l := range ls {
		if len(l.Features) > 0 {
			return false
		}
	}
	return true
}

func (ls Layers) RemoveOverlapTriangle(mapName string, zoom int) {
	if zoom < TriangleOverlapZoom {
		return
	}
	for _, l := range ls {
		l.RemoveOverlapTriangle(mapName)
	}
}

func (l *Layer) RemoveOverlapTriangle(mapName string) {
	if l.Name != mapName {
		return
	}

	newFt := l.Features[:0]
	//fmt.Println("here start RemoveOverlapTriangle, total len", len(l.Features))
	start := time.Now()
	for i, f := range l.Features {
		if i == 0 {
			newFt = append(newFt, l.Features[i])
		} else {
			overlap := false
			for _, newf := range newFt {
				t1, err := triangleFromGeom(newf.Geometry)
				if err != nil {
					continue
				}
				t2, err := triangleFromGeom(f.Geometry)
				if err != nil {
					continue
				}
				res := triangle.TriTri2D(t1, t2, 0.0, true, false)
				overlap = res || overlap
				if overlap {
					log.Println("found overlap", t1, t2)
					break
				}
			}
			if !overlap {
				newFt = append(newFt, l.Features[i])
			}
		}
	}
	elapsed := time.Since(start)
	log.Printf("RemoveOverlapTriangle took %s, before %d, after %d", elapsed, len(l.Features), len(newFt))
	l.Features = newFt
}

func triangleFromGeom(g geom.T) (*triangle.Triangle, error) {
	coords := g.FlatCoords()
	if len(coords) < 6 {
		return nil, fmt.Errorf("coords length too short %v", coords)
	}
	t := triangle.NewTriangle(coords[0], coords[1], coords[2], coords[3], coords[4], coords[5])
	return t, nil
}
