package mp

import "math"

// Predefined paths matching MetaPost's plain.mp definitions.
// All circles have diameter 1 (radius 0.5) centered at origin.

// circleAngles are the 8 points of MetaPost's fullcircle (0° through 315°).
var circleAngles = []float64{0, 45, 90, 135, 180, 225, 270, 315}

// makeCircleKnot creates a knot on a circle of given radius at the specified
// angle (degrees), with KnotGiven direction and tension 1. This mirrors
// MetaPost's makepath pencircle (mp.c:mp_make_path), which sets each knot's
// direction to the tangent of the circle at that point.
func makeCircleKnot(deg, r float64) *Knot {
	rad := deg * math.Pi / 180
	x := r * math.Cos(rad)
	y := r * math.Sin(rad)

	knot := NewKnot()
	knot.XCoord = Number(x)
	knot.YCoord = Number(y)

	// Tangent direction at angle θ on a circle: (-sin θ, cos θ)
	// nArg converts this to MetaPost's internal angle representation.
	tanX := -math.Sin(rad)
	tanY := math.Cos(rad)
	dirAngle := nArg(Number(tanX), Number(tanY))

	knot.LType = KnotGiven
	knot.RType = KnotGiven
	knot.LeftX = dirAngle  // left given angle
	knot.RightX = dirAngle // right given angle
	knot.LeftY = unity     // left tension = 1
	knot.RightY = unity    // right tension = 1

	return knot
}

// FullCircle returns a unit circle (diameter 1) centered at the origin.
// Equivalent to MetaPost's `fullcircle` (= makepath pencircle).
// The path starts at (0.5, 0) and goes counterclockwise with 8 knots.
// Control points are computed by the Hobby-Knuth solver, matching MetaPost.
func FullCircle() *Path {
	r := 0.5
	p := NewPath()

	var knots []*Knot
	for _, deg := range circleAngles {
		knots = append(knots, makeCircleKnot(deg, r))
	}

	// Link knots into a cycle
	for i, knot := range knots {
		p.Append(knot)
		if i > 0 {
			knots[i-1].Next = knot
			knot.Prev = knots[i-1]
		}
	}
	knots[len(knots)-1].Next = knots[0]
	knots[0].Prev = knots[len(knots)-1]

	// Solve to compute control points (like MetaPost's makepath pencircle)
	e := NewEngine()
	e.AddPath(p)
	e.Solve()

	return p
}

// HalfCircle returns the upper half of a unit circle.
// Equivalent to MetaPost's `halfcircle` (= subpath (0,4) of fullcircle).
// The path starts at (0.5, 0), goes through (0, 0.5), and ends at (-0.5, 0).
func HalfCircle() *Path {
	r := 0.5
	p := NewPath()

	angles := []float64{0, 45, 90, 135, 180}

	var knots []*Knot
	for i, deg := range angles {
		knot := makeCircleKnot(deg, r)

		if i == 0 {
			knot.LType = KnotEndpoint
			// RType stays KnotGiven
		} else if i == len(angles)-1 {
			// LType stays KnotGiven
			knot.RType = KnotEndpoint
		}

		knots = append(knots, knot)
	}

	// Link knots
	for i, knot := range knots {
		p.Append(knot)
		if i > 0 {
			knots[i-1].Next = knot
			knot.Prev = knots[i-1]
		}
	}

	e := NewEngine()
	e.AddPath(p)
	e.Solve()

	return p
}

// QuarterCircle returns the first quadrant arc of a unit circle.
// Equivalent to MetaPost's `quartercircle` (= subpath (0,2) of fullcircle).
// The path starts at (0.5, 0) and ends at (0, 0.5).
func QuarterCircle() *Path {
	r := 0.5
	p := NewPath()

	angles := []float64{0, 45, 90}

	var knots []*Knot
	for i, deg := range angles {
		knot := makeCircleKnot(deg, r)

		if i == 0 {
			knot.LType = KnotEndpoint
		} else if i == len(angles)-1 {
			knot.RType = KnotEndpoint
		}

		knots = append(knots, knot)
	}

	for i, knot := range knots {
		p.Append(knot)
		if i > 0 {
			knots[i-1].Next = knot
			knot.Prev = knots[i-1]
		}
	}

	e := NewEngine()
	e.AddPath(p)
	e.Solve()

	return p
}

// UnitSquare returns a unit square from (0,0) to (1,1).
// Equivalent to MetaPost's `unitsquare`.
// The path is (0,0)--(1,0)--(1,1)--(0,1)--cycle.
func UnitSquare() *Path {
	p := NewPath()

	coords := [][2]float64{
		{0, 0},
		{1, 0},
		{1, 1},
		{0, 1},
	}

	var knots []*Knot
	for _, c := range coords {
		knot := NewKnot()
		knot.XCoord = Number(c[0])
		knot.YCoord = Number(c[1])
		// Straight lines: control points equal to knot coordinates
		knot.LeftX = knot.XCoord
		knot.LeftY = knot.YCoord
		knot.RightX = knot.XCoord
		knot.RightY = knot.YCoord
		knot.LType = KnotExplicit
		knot.RType = KnotExplicit
		knots = append(knots, knot)
	}

	// Link knots into a cycle
	for i, knot := range knots {
		p.Append(knot)
		if i > 0 {
			knots[i-1].Next = knot
			knot.Prev = knots[i-1]
		}
	}
	// Close the cycle
	knots[len(knots)-1].Next = knots[0]
	knots[0].Prev = knots[len(knots)-1]

	return p
}
