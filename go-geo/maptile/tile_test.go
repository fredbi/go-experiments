package maptile_test

import (
	"testing"

	b "github.com/fredbi/geo/pkg/bound"
	"github.com/fredbi/geo/pkg/maptile"
	"github.com/twpayne/go-geom"
)

func TestTileNum2Deg(t *testing.T) {
	testcases := []struct {
		tile             maptile.Tile
		expectedZepislon float64
	}{
		{
			tile:             maptile.New(2, 1, 1),
			expectedZepislon: 0.001220703125,
		},
		{
			tile:             maptile.New(2, 2, 1),
			expectedZepislon: 0.001220703125,
		},
		{
			tile:             maptile.New(2, 2, 2),
			expectedZepislon: 0.0006103515625,
		},
		{
			tile:             maptile.New(2, 2, 10),
			expectedZepislon: 2.384185791015625e-06,
		},
	}

	for i, test := range testcases {
		zepislon := test.tile.ZEpislon()
		if zepislon != test.expectedZepislon {
			t.Errorf("Failed test %v. Expected zepislon (%v), got (%v)", i, test.expectedZepislon, zepislon)
		}
	}
}

func TestTileBound(t *testing.T) {
	testcases := []struct {
		tile          maptile.Tile
		expectedBound b.Bound
		buff          float64
	}{
		{
			tile: maptile.New(1, 1, 2),
			expectedBound: b.Bound{
				Min: geom.Coord{-90, 0},
				Max: geom.Coord{0, 66.51326044311185},
			},
			buff: 0,
		},
		{
			tile: maptile.New(8, 8, 5),
			expectedBound: b.Bound{
				Min: geom.Coord{-90, 61.60639637138627},
				Max: geom.Coord{-78.75, 66.51326044311185},
			},
			buff: 0,
		},
		{
			tile: maptile.New(16, 16, 6),
			expectedBound: b.Bound{
				Min: geom.Coord{-90, 64.16810689799152},
				Max: geom.Coord{-84.375, 66.51326044311185},
			},
			buff: 0,
		},
		{
			tile: maptile.New(16, 16, 6),
			expectedBound: b.Bound{
				Min: geom.Coord{-92.8125, 62.91523303947611},
				Max: geom.Coord{-81.5625, 67.60922060496384},
			},
			buff: 0.5,
		},
		{
			tile: maptile.New(1573, 3342, 13),
			expectedBound: b.Bound{
				Min: geom.Coord{-110.8740234375, 31.3911575228247},
				Max: geom.Coord{-110.830078125, 31.42866311735861},
			},
			buff: 0,
		},
		{
			tile: maptile.New(8956, 12223, 15),
			expectedBound: b.Bound{
				Min: geom.Coord{-81.6064453125, 41.50857729743933},
				Max: geom.Coord{-81.595458984375, 41.51680395810115},
			},
			buff: 0,
		},
		{
			tile: maptile.New(8956, 12223, 15),
			expectedBound: b.Bound{
				Min: geom.Coord{-81.6119384765625, 41.50446357504802},
				Max: geom.Coord{-81.5899658203125, 41.52091689636248},
			},
			buff: 0.5,
		},
	}

	for i, test := range testcases {
		bound := test.tile.Bound(test.buff)
		if bound.Min[0] != test.expectedBound.Min[0] || bound.Min[1] != test.expectedBound.Min[1] || bound.Max[0] != test.expectedBound.Max[0] ||
			bound.Max[1] != test.expectedBound.Max[1] {
			t.Error(i, "th test Failed test. Expected ,", test.expectedBound, "got ", bound)
		}
	}
}
