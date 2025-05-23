package mvt

import (
	"testing"

	"github.com/fredbi/geo/pkg/maptile"
	"github.com/fredbi/geo/pkg/project"
)

func TestNonPowerOfTwoProjection(t *testing.T) {
	tile := maptile.New(8956, 12223, 15)
	regProj := newProjection(tile, 4096)
	nonProj := nonPowerOfTwoProjection(tile, 4096)

	expected := loadGeoJSON(t, tile)

	//return
	layers := NewLayers(loadGeoJSON(t, tile))
	// loopy de loop of projections
	for _, l := range layers {
		for _, f := range l.Features {
			f.Geometry, _ = project.Geometry(f.Geometry, regProj.ToTile)
		}
	}

	for _, l := range layers {
		for _, f := range l.Features {
			f.Geometry, _ = project.Geometry(f.Geometry, nonProj.ToWGS84)
		}
	}

	for _, l := range layers {
		for _, f := range l.Features {
			f.Geometry, _ = project.Geometry(f.Geometry, nonProj.ToTile)
		}
	}

	for _, l := range layers {
		for _, f := range l.Features {
			f.Geometry, _ = project.Geometry(f.Geometry, regProj.ToWGS84)
		}
	}

	result := layers.ToFeatureCollections()

	xEpsilon, yEpsilon := tileEpsilon(tile)
	for key := range expected {
		for i := range expected[key].Features {
			r := result[key].Features[i]
			e := expected[key].Features[i]

			compareGeomGeometry(t, r.Geometry, e.Geometry, xEpsilon, yEpsilon)
		}
	}
}
