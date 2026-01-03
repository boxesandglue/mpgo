package mp

import "sort"

// Pen mirrors MetaPost's pen objects: a closed knot list describing the pen
// shape. MetaPost stores this as pen_p on stroke/fill nodes (mp.c:564,1056ff).
// Here we keep a minimal container with the pen's knot head.
type Pen struct {
	Head       *Knot
	Elliptical bool // mirrors pen_is_elliptical macro (mp.c:444)
}

// NewPenFromPath builds a Pen from an existing path (the path should be a closed
// convex outline, similar to the expectation in mp_make_pen, mp.c:10964ff).
func NewPenFromPath(p *Path) *Pen {
	if p == nil {
		return &Pen{}
	}
	return &Pen{Head: p.Head}
}

// PenCircle constructs an elliptical pen with diameter d, matching MetaPost's
// pencircle (mp.w:10440-10452, mp_get_pen_circle).
//
// MetaPost stores elliptical pens as a single knot where:
//   - (x_coord, y_coord) = center (translation)
//   - (left_x, left_y) = where (1,0) transforms to (first basis vector)
//   - (right_x, right_y) = where (0,1) transforms to (second basis vector)
//
// For an untransformed pencircle of diameter d:
//   - center = (0, 0)
//   - (1,0) -> (d, 0)   i.e. left_x=d, left_y=0
//   - (0,1) -> (0, d)   i.e. right_x=0, right_y=d
//
// The transformation matrix is thus:
//
//	| left_x   right_x |   | d  0 |
//	| left_y   right_y | = | 0  d |
func PenCircle(d Number) *Pen {
	k := NewKnot()
	// mp.w:10448-10453: center at origin, transformation stores diameter
	k.XCoord = 0
	k.YCoord = 0
	k.LeftX = d  // where (1,0) maps to: x-component
	k.LeftY = 0  // where (1,0) maps to: y-component
	k.RightX = 0 // where (0,1) maps to: x-component
	k.RightY = d // where (0,1) maps to: y-component
	k.LType = KnotExplicit
	k.RType = KnotExplicit
	k.Next = k
	k.Prev = k
	return &Pen{Head: k, Elliptical: true}
}

// PenRazor constructs a razor pen (horizontal line segment), matching
// makepen((-.5,0)--(.5,0)--cycle) (penrazor) in MetaPost.
// The size parameter scales the pen (default penrazor has length 1).
// Use PenRazorRotated for calligraphic effects at an angle.
func PenRazor(size Number) *Pen {
	half := size / 2
	p := NewPath()
	p.Append(&Knot{XCoord: -half, YCoord: 0})
	p.Append(&Knot{XCoord: half, YCoord: 0})
	return NewPenFromPath(p)
}

// PenRazorRotated constructs a razor pen rotated by the given angle (in degrees).
// This is equivalent to penrazor scaled size rotated angle in MetaPost.
func PenRazorRotated(size Number, angleDegrees Number) *Pen {
	half := size / 2
	// Convert degrees to radians
	angleRad := angleDegrees * 3.14159265358979323846 / 180.0
	cos := cosNumber(angleRad)
	sin := sinNumber(angleRad)

	// Rotate (-half, 0) and (half, 0) by angle
	x1 := -half * cos
	y1 := -half * sin
	x2 := half * cos
	y2 := half * sin

	p := NewPath()
	p.Append(&Knot{XCoord: x1, YCoord: y1})
	p.Append(&Knot{XCoord: x2, YCoord: y2})
	return NewPenFromPath(p)
}

// cosNumber computes cosine using Taylor series.
func cosNumber(x Number) Number {
	// Normalize to [-pi, pi]
	for x > 3.14159265358979323846 {
		x -= 2 * 3.14159265358979323846
	}
	for x < -3.14159265358979323846 {
		x += 2 * 3.14159265358979323846
	}
	// Taylor series: cos(x) = 1 - x²/2! + x⁴/4! - ...
	result := Number(1.0)
	term := Number(1.0)
	x2 := x * x
	for i := 1; i <= 10; i++ {
		term *= -x2 / Number((2*i-1)*(2*i))
		result += term
	}
	return result
}

// sinNumber computes sine using Taylor series.
func sinNumber(x Number) Number {
	// Normalize to [-pi, pi]
	for x > 3.14159265358979323846 {
		x -= 2 * 3.14159265358979323846
	}
	for x < -3.14159265358979323846 {
		x += 2 * 3.14159265358979323846
	}
	// Taylor series: sin(x) = x - x³/3! + x⁵/5! - ...
	result := x
	term := x
	x2 := x * x
	for i := 1; i <= 10; i++ {
		term *= -x2 / Number((2*i)*(2*i+1))
		result += term
	}
	return result
}

// Eps is MetaPost's epsilon value - a very small positive number.
// Used for penspeck and other near-zero comparisons.
const Eps Number = 0.00049

// PenSpeck constructs a nearly invisible point pen, matching
// penspeck = pensquare scaled eps in MetaPost.
// Useful for drawing paths without visible stroke.
func PenSpeck() *Pen {
	return PenSquare(Eps)
}

// PenSquare constructs a square pen of side length size, matching
// makepen(unitsquare shifted -(.5,.5)) (pensquare) in MetaPost. For now this
// returns a 4-knot closed path (axis-aligned).
func PenSquare(size Number) *Pen {
	half := size / 2
	p := NewPath()
	p.Append(&Knot{XCoord: -half, YCoord: -half})
	p.Append(&Knot{XCoord: half, YCoord: -half})
	p.Append(&Knot{XCoord: half, YCoord: half})
	p.Append(&Knot{XCoord: -half, YCoord: half})
	return NewPenFromPath(p)
}

// MakePen builds a pen from an arbitrary path by taking its convex hull,
// akin to mp_make_pen(mp.c:9290ff). We ignore Bezier controls and use knot
// coordinates only, mirroring mp_convex_hull call when need_hull=true.
func MakePen(path *Path) *Pen {
	points := make([][2]Number, 0)
	if path == nil || path.Head == nil {
		return &Pen{}
	}
	cur := path.Head
	for {
		points = append(points, [2]Number{cur.XCoord, cur.YCoord})
		cur = cur.Next
		if cur == path.Head {
			break
		}
	}
	hull := convexHull(points)
	if len(hull) == 0 {
		return &Pen{}
	}
	if len(hull) == 1 {
		// Degenerate hull -> treat like an elliptical pen centered at point.
		pen := PenCircle(0)
		pen.Head.XCoord = hull[0][0]
		pen.Head.YCoord = hull[0][1]
		return pen
	}
	p := NewPath()
	for _, pt := range hull {
		p.Append(&Knot{XCoord: pt[0], YCoord: pt[1]})
	}
	return &Pen{Head: p.Head, Elliptical: false}
}

// convexHull computes the 2D convex hull using the monotonic chain algorithm.
func convexHull(pts [][2]Number) [][2]Number {
	if len(pts) < 3 {
		return pts
	}
	// sort by x,y
	sort.Slice(pts, func(i, j int) bool {
		if pts[i][0] == pts[j][0] {
			return pts[i][1] < pts[j][1]
		}
		return pts[i][0] < pts[j][0]
	})
	cross := func(o, a, b [2]Number) Number {
		return (a[0]-o[0])*(b[1]-o[1]) - (a[1]-o[1])*(b[0]-o[0])
	}
	lower := make([][2]Number, 0)
	for _, p := range pts {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}
	upper := make([][2]Number, 0)
	for i := len(pts) - 1; i >= 0; i-- {
		p := pts[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}
	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	return hull
}

// penPoints returns the raw points of the pen knot loop (ignores controls).
func penPoints(pen *Pen) [][2]Number {
	if pen == nil || pen.Head == nil {
		return nil
	}
	var pts [][2]Number
	cur := pen.Head
	for {
		pts = append(pts, [2]Number{cur.XCoord, cur.YCoord})
		cur = cur.Next
		if cur == pen.Head || cur == nil {
			break
		}
	}
	return pts
}

// PenEnvelopeHull builds a convex hull over the pen translated to each knot
// position of the given path. This is a coarse approximation of the swept
// area MetaPost computes in the offset phase for non-elliptical pens
// (mp_offset_prep/mp_apply_offset, mp.c ~15800ff). For now this serves to
// emit a fill outline instead of stroking.
func PenEnvelopeHull(path *Path, pen *Pen) [][2]Number {
	if path == nil || path.Head == nil || pen == nil || pen.Head == nil {
		return nil
	}
	base := penPoints(pen)
	if len(base) == 0 {
		return nil
	}
	var cloud [][2]Number
	cur := path.Head
	for {
		for _, pt := range base {
			cloud = append(cloud, [2]Number{pt[0] + cur.XCoord, pt[1] + cur.YCoord})
		}
		cur = cur.Next
		if cur == path.Head || cur == nil {
			break
		}
	}
	return convexHull(cloud)
}

// GetPenScale computes the scale factor of an elliptical pen, matching
// mp_get_pen_scale (mp.w:11529-11547).
//
// For elliptical pens, this returns sqrt(|det(M)|) where M is the
// transformation matrix:
//
//	| a  b |   | left_x - x_coord    right_x - x_coord |
//	| c  d | = | left_y - y_coord    right_y - y_coord |
//
// For an untransformed pencircle with diameter d, this returns d.
// For polygonal pens, this returns 0 (use PenBBox instead).
func GetPenScale(pen *Pen) Number {
	if pen == nil || pen.Head == nil {
		return 0
	}
	if !pen.Elliptical {
		return 0
	}
	p := pen.Head
	// mp.w:11538-11541: extract transformation matrix
	a := p.LeftX - p.XCoord  // where (1,0) maps to: x
	b := p.RightX - p.XCoord // where (0,1) maps to: x
	c := p.LeftY - p.YCoord  // where (1,0) maps to: y
	d := p.RightY - p.YCoord // where (0,1) maps to: y

	// mp.w:11542: sqrt_det computes sqrt(|a*d - b*c|)
	det := a*d - b*c
	if det < 0 {
		det = -det
	}
	return sqrtNumber(det)
}

// sqrtNumber computes the square root of a Number.
func sqrtNumber(x Number) Number {
	if x <= 0 {
		return 0
	}
	// Newton-Raphson iteration for sqrt
	r := x
	for i := 0; i < 20; i++ {
		r = (r + x/r) / 2
	}
	return r
}
