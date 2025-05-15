package geojson

import (
	"testing"

	"github.com/twpayne/go-geom"

	"github.com/stretchr/testify/assert"
)

func mkBBox(args []float64) *BBox {
	gg := geom.NewBounds(geom.Layout(len(args) / 2)).Set(args...)
	return &BBox{gg}
}

func TestBBoxValid(t *testing.T) {
	cases := []struct {
		name    string
		args    []float64
		success *BBox
		err     string
		valid   bool
	}{
		{
			name:    "true for 4 length array",
			args:    []float64{1, 2, 3, 4},
			success: mkBBox([]float64{1, 2, 3, 4}),
			valid:   true,
		},
		{
			name:    "true for 3d box",
			args:    []float64{1, 2, 3, 4, 5, 6},
			success: mkBBox([]float64{1, 2, 3, 4, 5, 6}),
			valid:   true,
		},
		{
			name:    "false for nil box",
			success: &BBox{geom.NewBounds(geom.NoLayout)},
			valid:   true,
		},
		{
			name:  "false for short array",
			args:  []float64{1, 2, 3},
			err:   "geojson: bbox even number of arguments required: 3",
			valid: false,
		},
		{
			name:  "false for incorrect length array",
			args:  []float64{1, 2, 3, 4, 5},
			err:   "geojson: bbox even number of arguments required: 5",
			valid: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bb, err := NewBBox(tc.args)
			if tc.valid {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.success, bb)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
