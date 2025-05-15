package geojson

import (
	"fmt"

	//"github.com/twpayne/go-geom"
	"github.com/fredbi/go-geom/geom"
	"github.com/fredbi/go-geom/geom/utils"
)

func NewBBox(args []float64) (*BBox, error) {
	if len(args) == 0 {
		return &BBox{utils.NewBounds(geom.NoLayout)}, nil
	}
	if len(args)&1 != 0 {
		return nil, fmt.Errorf("geojson: bbox even number of arguments required: %d", len(args))
	}

	gg := utils.NewBounds(geom.Layout(len(args) / 2)).Set(args...)
	return &BBox{gg}, nil
}

// BBox is for the geojson bbox attribute which is an array with all axes
// of the most southwesterly point followed by all axes of the most northeasterly point.
type BBox struct {
	geom.Bounds
}

func (bb BBox) MarshalJSON() ([]byte, error) {
	b := bb.Bounds
	var mins []float64
	var maxs []float64
	for i := 0; i < b.Layout().Stride(); i++ {
		mins = append(mins, b.Min(i))
		maxs = append(maxs, b.Max(i))
	}
	return json.Marshal(append(mins, maxs...))
}

func (bb *BBox) UnmarshalJSON(data []byte) error {
	var coords []float64
	if err := json.Unmarshal(data, &coords); err != nil {
		return err
	}
	if len(coords)&1 != 0 {
		return fmt.Errorf("geojson: bbox even number of arguments required: %d", len(coords))
	}
	gg := utils.NewBounds(geom.Layout(len(coords) / 2)).Set(coords...)
	*bb = BBox{gg}
	return nil
}

func (bb *BBox) Center() geom.Point {
	coords := make([]float64, bb.Layout().Stride())
	for i := 0; i < bb.Layout().Stride(); i++ {
		coords[i] = (bb.Min(i) + bb.Max(i)) / 2.0
	}
	return utils.NewPoint(bb.Layout(), coords)
}
