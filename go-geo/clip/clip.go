package clip

import (
	bound "github.com/fredbi/geo/pkg/bound"
	"github.com/twpayne/go-geom"
)

// Code based on https://github.com/mapbox/lineclip
// line will clip a line into a set of lines
// along the bounding box boundary.
func line(box bound.Bound, in *geom.LineString, open bool) *geom.MultiLineString {
	if len(in.Coords()) == 0 {
		return nil
	}

	out := geom.NewMultiLineString(in.Layout())
	line := 0

	var codeA int
	if open {
		codeA = bitCodeOpen(box, in.Coord(0))
	} else {
		codeA = bitCode(box, in.Coord(0))
	}

	loopTo := in.NumCoords()
	for i := 1; i < loopTo; i++ {
		a := in.Coord(i - 1)
		b := in.Coord(i)

		var codeB int
		if open {
			codeB = bitCodeOpen(box, b)
		} else {
			codeB = bitCode(box, b)
		}
		endCode := codeB

		// loops through all the intersection of the line and box.
		// eg. across a corner could have two intersections.
		for {
			if codeA|codeB == 0 { //nolint:gocritic
				// both points are in the box, accept
				out = push(out, line, a)
				if codeB != endCode { // segment went outside
					out = push(out, line, b)
					if i < loopTo-1 {
						line++
					}
				} else if i == loopTo-1 {
					out = push(out, line, b)
				}
				break
			} else if codeA&codeB != 0 {
				// both on one side of the box.
				// segment not part of the final result.
				break
			} else if codeA != 0 {
				// A is outside, B is inside, clip edge
				a = intersect(box, codeA, a, b)
				codeA = bitCode(box, a)
			} else {
				// B is outside, A is inside, clip edge
				b = intersect(box, codeB, a, b)
				codeB = bitCode(box, b)
			}
		}

		codeA = endCode // new start is the old end
	}
	if out.NumLineStrings() == 0 {
		return nil
	}
	return out
}

func push(out *geom.MultiLineString, i int, p geom.Coord) *geom.MultiLineString {
	if i >= out.NumLineStrings() {
		ls := geom.NewLineString(out.Layout())
		err := out.Push(ls.Clone())
		if err != nil {
			return nil
		}
	}

	//pt := geom.NewPoint(out.Layout())
	coords := out.Coords()
	coords[i] = append(coords[i], p)
	out.MustSetCoords(coords)
	//out[i] = append(out[i], p)
	return out
}

// ring will clip the Ring into a smaller ring around the bounding box boundary.
func ring(box bound.Bound, in *geom.LinearRing) *geom.LinearRing {
	var out *geom.LinearRing
	//var out orb.Ring
	if in.NumCoords() == 0 {
		return in
	}
	coord := in.Coords()
	length := len(coord)
	f := coord[0]
	l := coord[length-1]

	initClosed := false
	if f.Equal(in.Layout(), l) {
		initClosed = true
	}

	for edge := 1; edge <= 8; edge <<= 1 {
		out = geom.NewLinearRing(in.Layout())

		loopTo := in.NumCoords()

		// if we're not a nice closed ring, don't implicitly close it.
		prev := in.Coord(loopTo - 1)
		if !initClosed {
			prev = in.Coord(0)
		}

		prevInside := bitCode(box, prev)&edge == 0

		for i := 0; i < loopTo; i++ {
			p := in.Coord(i)
			inside := bitCode(box, p)&edge == 0

			// if segment goes through the clip window, add an intersection
			if inside != prevInside {
				i := intersect(box, edge, prev, p)
				cd := append(out.Coords(), i)
				//pt := geom.NewPoint(in.Layout()).MustSetCoords(p)
				out.MustSetCoords(cd)
			}
			if inside {
				cd := append(out.Coords(), p)
				out.MustSetCoords(cd)
			}

			prev = p
			prevInside = inside
		}

		if out.NumCoords() == 0 {
			return nil
		}

		in, out = out, in
	}
	out = in // swap back

	if initClosed {
		// need to make sure our output is also closed.
		if l := out.NumCoords(); l != 0 {
			f := out.Coord(0)
			l := out.Coord(l - 1)

			if !f.Equal(out.Layout(), l) {
				cd := append(out.Coords(), f)
				out.MustSetCoords(cd)
				//out = append(out, f)
			}
		}
	}

	return out
}

// bitCode returns the point position relative to the bbox:
//         left  mid  right
//    top  1001  1000  1010
//    mid  0001  0000  0010
// bottom  0101  0100  0110
func bitCode(b bound.Bound, p geom.Coord) int {
	code := 0
	if p[0] < b.Min[0] {
		code |= 1
	} else if p[0] > b.Max[0] {
		code |= 2
	}

	if p[1] < b.Min[1] {
		code |= 4
	} else if p[1] > b.Max[1] {
		code |= 8
	}

	return code
}

func bitCodeOpen(b bound.Bound, p geom.Coord) int {
	code := 0
	if p[0] <= b.Min[0] {
		code |= 1
	} else if p[0] >= b.Max[0] {
		code |= 2
	}

	if p[1] <= b.Min[1] {
		code |= 4
	} else if p[1] >= b.Max[1] {
		code |= 8
	}

	return code
}

// intersect a segment against one of the 4 lines that make up the bbox
func intersect(box bound.Bound, edge int, a, b geom.Coord) geom.Coord {
	if edge&8 != 0 { //nolint:gocritic
		// top
		return geom.Coord{a[0] + (b[0]-a[0])*(box.Max[1]-a[1])/(b[1]-a[1]), box.Max[1]}
	} else if edge&4 != 0 {
		// bottom
		return geom.Coord{a[0] + (b[0]-a[0])*(box.Min[1]-a[1])/(b[1]-a[1]), box.Min[1]}
	} else if edge&2 != 0 {
		// right
		return geom.Coord{box.Max[0], a[1] + (b[1]-a[1])*(box.Max[0]-a[0])/(b[0]-a[0])}
	} else if edge&1 != 0 {
		// left
		return geom.Coord{box.Min[0], a[1] + (b[1]-a[1])*(box.Min[0]-a[0])/(b[0]-a[0])}
	}

	panic("no edge??")
}
