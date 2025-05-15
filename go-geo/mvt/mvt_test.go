package mvt

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/fredbi/geo/pkg/maptile"
	"github.com/fredbi/geo/utils/geojson"
)

func TestMeshData(t *testing.T) {
	cases := []struct {
		name   string
		layer  string
		tile   maptile.Tile
		expect int
	}{
		{
			name:   "7-24-52",
			layer:  "mesh",
			tile:   maptile.New(24, 52, 7),
			expect: 15565,
		},
		{
			name:   "7-25-52",
			layer:  "mesh",
			tile:   maptile.New(25, 52, 7),
			expect: 0,
		},
		{
			name:   "8-49-104",
			layer:  "mesh",
			tile:   maptile.New(49, 104, 8),
			expect: 15565,
		},
		{
			name:   "10-196-417",
			layer:  "mesh",
			tile:   maptile.New(196, 417, 10),
			expect: 15565,
		},
		{
			name:   "10-196-418",
			layer:  "mesh",
			tile:   maptile.New(196, 418, 10),
			expect: 15429,
		},
		{
			name:   "10-197-418",
			layer:  "mesh",
			tile:   maptile.New(197, 418, 10),
			expect: 0,
		},
		{
			name:   "11-392-835",
			layer:  "mesh",
			tile:   maptile.New(392, 835, 11),
			expect: 15565,
		},
		{
			name:   "12-785-1671",
			layer:  "mesh",
			tile:   maptile.New(785, 1671, 12),
			expect: 12832,
		},
		{
			name:   "14-3142-6687",
			layer:  "mesh",
			tile:   maptile.New(3142, 6687, 14),
			expect: 2365,
		},
		{
			name:   "15-8956-12223",
			layer:  "mesh",
			tile:   maptile.New(8956, 12223, 15),
			expect: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			data, err := ioutil.ReadFile("testdata/2018082412_0_mesh.geojson")
			if err != nil {
				t.Fatalf("failed to load mvt file: %v", err)
			}

			fc := make(map[string]*geojson.FeatureCollection)
			err = json.Unmarshal(data, &fc)
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			layers := NewLayers(fc)
			layers.ProjectToTile(tc.tile) // x, y, z
			layers.RemoveDataOutsideBuffer(640.0)
			result := len(layers[0].Features)
			if tc.expect != result {
				t.Errorf("expect feature number %d, get %d", tc.expect, result)
			}

			if len(layers[0].Features) > 0 {
				numProperties := len(layers[0].Features[0].Properties)
				if numProperties != 5 {
					t.Errorf("expect property num 5, get %d", numProperties)
				}
			}

		})
	}
}

func TestMeshDataClip(t *testing.T) {
	cases := []struct {
		name   string
		layer  string
		tile   maptile.Tile
		expect int
	}{
		{
			name:   "7-24-52",
			layer:  "mesh",
			tile:   maptile.New(24, 52, 7),
			expect: 15565,
		},
		{
			name:   "7-25-52",
			layer:  "mesh",
			tile:   maptile.New(25, 52, 7),
			expect: 0,
		},
		{
			name:   "8-49-104",
			layer:  "mesh",
			tile:   maptile.New(49, 104, 8),
			expect: 15565,
		},
		{
			name:   "10-196-417",
			layer:  "mesh",
			tile:   maptile.New(196, 417, 10),
			expect: 15565,
		},
		{
			name:   "10-196-418",
			layer:  "mesh",
			tile:   maptile.New(196, 418, 10),
			expect: 6770,
		},
		{
			name:   "10-197-418",
			layer:  "mesh",
			tile:   maptile.New(197, 418, 10),
			expect: 0,
		},
		{
			name:   "11-392-835",
			layer:  "mesh",
			tile:   maptile.New(392, 835, 11),
			expect: 15565,
		},
		{
			name:   "12-785-1671",
			layer:  "mesh",
			tile:   maptile.New(785, 1671, 12),
			expect: 15565,
		},
		{
			name:   "14-3142-6687",
			layer:  "mesh",
			tile:   maptile.New(3142, 6687, 14),
			expect: 7883,
		},
		{
			name:   "15-6284-13374",
			layer:  "mesh",
			tile:   maptile.New(6284, 13374, 15),
			expect: 2001,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			data, err := ioutil.ReadFile("testdata/2018082412_0_mesh.geojson")
			if err != nil {
				t.Fatalf("failed to load mvt file: %v", err)
			}

			fc := make(map[string]*geojson.FeatureCollection)
			err = json.Unmarshal(data, &fc)
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			layers := NewLayers(fc)
			layers.ProjectToTile(tc.tile) // x, y, z
			layers.Clip(maptile.NewTileBoundBuff(0.0))
			if len(layers[0].Features) > 0 {
				numProperties := len(layers[0].Features[0].Properties)
				if numProperties != 5 {
					t.Errorf("expect property num 5, get %d", numProperties)
				}
			}
			result := len(layers[0].Features)
			if tc.expect != result {
				t.Errorf("expect feature number %d, get %d", tc.expect, result)
			}

		})
	}
}
