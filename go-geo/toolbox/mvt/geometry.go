package mvt

import (
	"encoding/json"
	"fmt"
	"reflect"

	vectortile "github.com/fredbi/geo/pkg/mvt/vtile"
	"github.com/pkg/errors"
	"github.com/twpayne/go-geom"
)

const (
	moveTo    = 1
	lineTo    = 2
	closePath = 7
)

// Orientation defines the order of the points in a polygon
// or closed ring.
type Orientation int8

// Constants to define orientation.
// They follow the right hand rule for orientation.
const (
	// CCW stands for Counter Clock Wise
	CCW Orientation = 1

	// CW stands for Clock Wise
	CW Orientation = -1
)

type geomEncoder struct {
	prevX, prevY int32
	Data         []uint32
}

// A geomDecoder holds state for geometry decoding.
type geomDecoder struct {
	geom []uint32
	i    int

	prev geom.Coord
}

func encodeGeometry(g geom.T) (vectortile.Tile_GeomType, []uint32, error) {
	switch g := g.(type) {
	case *geom.Point:
		e := newGeomEncoder(3)
		e.MoveTo([]geom.Coord{g.Coords()})

		return vectortile.Tile_POINT, e.Data, nil
	case *geom.MultiPoint:
		coords2d := g.Coords()
		e := newGeomEncoder(1 + 2*len(coords2d))
		e.MoveTo(coords2d)

		return vectortile.Tile_POINT, e.Data, nil
	case *geom.LineString:
		coords2d := g.Coords()
		e := newGeomEncoder(2 + 2*len(coords2d))
		e.MoveTo([]geom.Coord{coords2d[0]})
		e.LineTo(coords2d[1:])

		return vectortile.Tile_LINESTRING, e.Data, nil
	case *geom.MultiLineString:
		coords3d := g.Coords()

		e := newGeomEncoder(elMLS(coords3d))
		for _, coords2d := range coords3d {
			e.MoveTo([]geom.Coord{coords2d[0]})
			e.LineTo(coords2d[1:])
		}

		return vectortile.Tile_LINESTRING, e.Data, nil
	case *geom.LinearRing:
		coords2d := g.Coords()

		e := newGeomEncoder(3 + 2*len(coords2d))
		e.MoveTo([]geom.Coord{coords2d[0]})
		if closed(coords2d) {
			e.LineTo(coords2d[1 : len(coords2d)-1])
		} else {
			e.LineTo(coords2d[1:])
		}
		e.ClosePath()

		return vectortile.Tile_POLYGON, e.Data, nil
	case *geom.Polygon:
		coords3d := g.Coords()
		e := newGeomEncoder(elP(coords3d))
		for _, coords2d := range coords3d {
			e.MoveTo([]geom.Coord{coords2d[0]})
			if closed(coords2d) {
				e.LineTo(coords2d[1 : len(coords2d)-1])
			} else {
				e.LineTo(coords2d[1:])
			}
			e.ClosePath()
		}

		return vectortile.Tile_POLYGON, e.Data, nil
	case *geom.MultiPolygon:
		coords4d := g.Coords()
		e := newGeomEncoder(elMP(coords4d))
		for _, coords3d := range coords4d {
			for _, coords2d := range coords3d {
				e.MoveTo([]geom.Coord{coords2d[0]})
				if closed(coords2d) {
					e.LineTo(coords2d[1 : len(coords2d)-1])
				} else {
					e.LineTo(coords2d[1:])
				}
				e.ClosePath()
			}
		}

		return vectortile.Tile_POLYGON, e.Data, nil
	case *geom.GeometryCollection:
		return 0, nil, errors.New("geometry collections are not supported")

	}
	panic(fmt.Sprintf("geometry type not supported: %T", g))
}

func newGeomEncoder(l int) *geomEncoder {
	return &geomEncoder{
		Data: make([]uint32, 0, l),
	}
}

func (ge *geomEncoder) MoveTo(points []geom.Coord) {
	l := uint32(len(points))
	ge.Data = append(ge.Data, (l<<3)|moveTo)
	ge.addPoints(points)
}

func (ge *geomEncoder) addPoints(points []geom.Coord) {
	for i := range points {
		x := int32(points[i][0]) - ge.prevX
		y := int32(points[i][1]) - ge.prevY

		ge.prevX = int32(points[i][0])
		ge.prevY = int32(points[i][1])

		ge.Data = append(ge.Data,
			uint32((x<<1)^(x>>31)),
			uint32((y<<1)^(y>>31)),
		)
	}
}

func (ge *geomEncoder) ClosePath() {
	ge.Data = append(ge.Data, (1<<3)|closePath)
}

func (ge *geomEncoder) LineTo(points []geom.Coord) {
	l := uint32(len(points))
	ge.Data = append(ge.Data, (l<<3)|lineTo)
	ge.addPoints(points)
}

func newKeyValueEncoder() *keyValueEncoder {
	return &keyValueEncoder{
		keyMap:   make(map[string]uint32),
		valueMap: make(map[interface{}]uint32),
	}
}

type keyValueEncoder struct {
	Keys   []string
	keyMap map[string]uint32

	Values   []*vectortile.Tile_Value
	valueMap map[interface{}]uint32
}

func (kve *keyValueEncoder) Key(s string) uint32 {
	if i, ok := kve.keyMap[s]; ok {
		return i
	}

	i := uint32(len(kve.Keys))
	kve.Keys = append(kve.Keys, s)
	kve.keyMap[s] = i

	return i
}

func (kve *keyValueEncoder) Value(v interface{}) (uint32, error) {
	// If a type is not comparable we can't figure out uniqueness in the hash,
	// we also can't encode it into a vectortile.Tile_Value.
	// So we encoded it as a json string, which is what other encoders
	// also do.
	if !reflect.TypeOf(v).Comparable() {
		data, err := json.Marshal(v)
		if err != nil {
			return 0, errors.Errorf("uncomparable: %T", v)
		}

		v = string(data)
	}

	if i, ok := kve.valueMap[v]; ok {
		return i, nil
	}

	tv, err := encodeValue(v)
	if err != nil {
		return 0, err
	}

	i := uint32(len(kve.Values))
	kve.Values = append(kve.Values, tv)
	kve.valueMap[v] = i

	return i, nil
}

func encodeValue(v interface{}) (*vectortile.Tile_Value, error) {
	tv := &vectortile.Tile_Value{}
	switch t := v.(type) {
	case string:
		tv.StringValue = &t
	case fmt.Stringer:
		s := t.String()
		tv.StringValue = &s
	case int:
		i := int64(t)
		tv.SintValue = &i
	case int8:
		i := int64(t)
		tv.SintValue = &i
	case int16:
		i := int64(t)
		tv.SintValue = &i
	case int32:
		i := int64(t)
		tv.SintValue = &i
	case int64:
		i := t
		tv.SintValue = &i
	case uint:
		i := uint64(t)
		tv.UintValue = &i
	case uint8:
		i := uint64(t)
		tv.UintValue = &i
	case uint16:
		i := uint64(t)
		tv.UintValue = &i
	case uint32:
		i := uint64(t)
		tv.UintValue = &i
	case uint64:
		i := t
		tv.UintValue = &i
	case float32:
		tv.FloatValue = &t
	case float64:
		tv.DoubleValue = &t
	case bool:
		tv.BoolValue = &t
	default:
		return nil, errors.Errorf("unable to encode value of type %T: %v", v, v)
	}

	return tv, nil
}

// functions to estimate encoded length

func elMLS(mls [][]geom.Coord) int {
	c := 0
	for _, ls := range mls {
		c += 2 + 2*len(ls)
	}

	return c
}

func elP(p [][]geom.Coord) int {
	c := 0
	for _, r := range p {
		c += 3 + 2*len(r)
	}

	return c
}

func elMP(mp [][][]geom.Coord) int {
	c := 0
	for _, p := range mp {
		c += elP(p)
	}

	return c
}

// Closed will return true if the ring is a real ring.
// ie. 4+ points and the first and last points match.
// NOTE: this will not check for self-intersection.
func closed(ring []geom.Coord) bool {
	first := ring[0]
	last := ring[len(ring)-1]
	if len(first) != len(last) {
		return false
	}
	for i, val := range first {
		if val != last[i] {
			return false
		}
	}
	return true
}

func decodeGeometry(geomType vectortile.Tile_GeomType, geombytes []uint32) (geom.T, error) {
	if len(geombytes) < 2 {
		return nil, errors.Errorf("geom is not long enough: %v", geombytes)
	}

	// we set prev deault 2 values hard coded, but may be changed by the stride in the future
	gd := &geomDecoder{geom: geombytes, prev: geom.Coord{0, 0}}

	switch geomType {
	case vectortile.Tile_POINT:
		return gd.decodePoint()
	case vectortile.Tile_LINESTRING:
		return gd.decodeLineString()
	case vectortile.Tile_POLYGON:
		return gd.decodePolygon()
	}

	return nil, errors.Errorf("unknown geometry type: %v", geomType)
}

func (gd *geomDecoder) decodePoint() (geom.T, error) {
	_, count, err := gd.cmdAndCount()
	if err != nil {
		return nil, err
	}

	if count == 1 {
		point := geom.NewPoint(geom.XY)
		coord := gd.NextPoint()
		point, err = point.SetCoords(coord)
		if err != nil {
			return nil, err
		}
		return point, nil
	}
	coord2d := make([]geom.Coord, 0, count)
	for i := uint32(0); i < count; i++ {
		nextp := gd.NextPoint()
		coord2d = append(coord2d, nextp)
	}
	mp := geom.NewMultiPoint(geom.XY)
	mp, err = mp.SetCoords(coord2d)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

func (gd *geomDecoder) decodeLine() ([]geom.Coord, error) {
	cmd, count, err := gd.cmdAndCount()
	if err != nil {
		return nil, err
	}

	if cmd != moveTo || count != 1 {
		return nil, errors.New("first command not one moveTo")
	}

	first := gd.NextPoint()
	cmd, count, err = gd.cmdAndCount()
	if err != nil {
		return nil, err
	}

	if cmd != lineTo {
		return nil, errors.New("second command not a lineTo")
	}

	coord2d := make([]geom.Coord, 0, count+1)
	coord2d = append(coord2d, first)

	for i := uint32(0); i < count; i++ {
		coord2d = append(coord2d, gd.NextPoint())
	}

	return coord2d, nil
}

func (gd *geomDecoder) decodeLineString() (geom.T, error) {
	var coord3d [][]geom.Coord
	for !gd.done() {
		coord2d, err := gd.decodeLine()
		if err != nil {
			return nil, err
		}

		if gd.done() && len(coord3d) == 0 {
			ls := geom.NewLineString(geom.XY)
			ls, err = ls.SetCoords(coord2d)
			if err != nil {
				return nil, err
			}
			return ls, nil
		}
		coord3d = append(coord3d, coord2d)
	}
	mls := geom.NewMultiLineString(geom.XY)

	mls, err := mls.SetCoords(coord3d)
	if err != nil {
		return nil, err
	}
	return mls, nil
}

func (gd *geomDecoder) decodePolygon() (geom.T, error) {
	// var mp orb.MultiPolygon
	// var p orb.Polygon
	var coord3d [][]geom.Coord
	var coord4d [][][]geom.Coord

	for !gd.done() {
		coord2d, err := gd.decodeLine()
		if err != nil {
			return nil, err
		}

		// r := orb.Ring(ls)

		cmd, _, err := gd.cmdAndCount()
		if err != nil {
			return nil, err
		}

		if cmd == closePath && !closed(coord2d) {
			coord2d = append(coord2d, coord2d[0])
		}

		// figure out if new polygon
		if len(coord4d) == 0 && len(coord3d) == 0 {
			coord3d = append(coord3d, coord2d)
		} else {
			if orientation(coord2d) == CCW {
				coord4d = append(coord4d, coord3d)
				coord3d = [][]geom.Coord{coord2d}
			} else {
				coord3d = append(coord3d, coord2d)
				//p = append(p, r)
			}
		}
	}

	if len(coord4d) == 0 {
		p := geom.NewPolygon(geom.XY)
		p, err := p.SetCoords(coord3d)
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	coord4d = append(coord4d, coord3d)
	mp := geom.NewMultiPolygon(geom.XY)
	mp, err := mp.SetCoords(coord4d)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

func (gd *geomDecoder) cmdAndCount() (uint32, uint32, error) {
	if gd.i >= len(gd.geom) {
		return 0, 0, errors.New("no more data")
	}

	v := gd.geom[gd.i]

	cmd := v & 0x07
	count := v >> 3
	gd.i++

	if cmd != closePath {
		if v := gd.i + int(2*count); len(gd.geom) < v {
			return 0, 0, errors.Errorf("data cut short: needed %d, have %d", v, len(gd.geom))
		}
	}

	return cmd, count, nil
}

func (gd *geomDecoder) NextPoint() geom.Coord {
	gd.i += 2
	gd.prev[0] += unzigzag(gd.geom[gd.i-2])
	gd.prev[1] += unzigzag(gd.geom[gd.i-1])
	// it has to be a clone because it is slice
	return gd.prev.Clone()
}

func (gd *geomDecoder) done() bool {
	return gd.i >= len(gd.geom)
}

func decodeValue(v *vectortile.Tile_Value) interface{} {
	if v == nil {
		return nil
	}
	switch {
	case v.StringValue != nil:
		rt := *v.StringValue
		return rt
	case v.FloatValue != nil:
		rt := float64(*v.FloatValue)
		return rt
	case v.DoubleValue != nil:
		rt := *v.DoubleValue
		return rt
	case v.IntValue != nil:
		rt := float64(*v.IntValue)
		return rt
	case v.UintValue != nil:
		rt := float64(*v.UintValue)
		return rt
	case v.SintValue != nil:
		rt := float64(*v.SintValue)
		return rt
	case v.BoolValue != nil:
		rt := *v.BoolValue
		return rt
	default:
		return nil
	}
}

func unzigzag(v uint32) float64 {
	return float64(int32(((v >> 1) & ((1 << 32) - 1)) ^ -(v & 1)))
}

// Orientation returns 1 if the the ring is in couter-clockwise order,
// return -1 if the ring is the clockwise order and 0 if the ring is
// degenerate and had no area.
func orientation(r []geom.Coord) Orientation {
	area := 0.0

	// This is a fast planar area computation, which is okay for this use.
	// implicitly move everything to near the origin to help with roundoff
	offsetX := r[0][0]
	offsetY := r[0][1]
	for i := 1; i < len(r)-1; i++ {
		area += (r[i][0]-offsetX)*(r[i+1][1]-offsetY) -
			(r[i+1][0]-offsetX)*(r[i][1]-offsetY)
	}

	if area > 0 {
		return CCW
	}

	if area < 0 {
		return CW
	}

	// degenerate case, no area
	return 0
}
