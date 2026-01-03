package mp

import "math"

// Predefined paths matching MetaPost's plain.mp definitions.
// All circles have diameter 1 (radius 0.5) centered at origin.

// For Bézier circle approximation with 8 segments, MetaPost uses
// control point distance = radius * 4 * (sqrt(2) - 1) / 3 ≈ 0.2761423749
// This gives very accurate circular arcs.
const circleKappa = 0.2761423749153966

// FullCircle returns a unit circle (diameter 1) centered at the origin.
// Equivalent to MetaPost's `fullcircle`.
// The path starts at (0.5, 0) and goes counterclockwise.
func FullCircle() *Path {
	r := 0.5 // radius
	k := circleKappa

	p := NewPath()

	// 8 points around the circle, starting at (r, 0) = 0°
	// Each arc spans 45°
	angles := []float64{0, 45, 90, 135, 180, 225, 270, 315}

	var knots []*Knot
	for _, deg := range angles {
		rad := deg * math.Pi / 180
		x := r * math.Cos(rad)
		y := r * math.Sin(rad)

		knot := NewKnot()
		knot.XCoord = Number(x)
		knot.YCoord = Number(y)
		knot.LType = KnotExplicit
		knot.RType = KnotExplicit

		// Control points are perpendicular to the radius
		// Left control: rotate tangent by -90° (clockwise from previous)
		// Right control: rotate tangent by +90° (counterclockwise to next)
		tanX := -math.Sin(rad) // tangent direction (perpendicular to radius)
		tanY := math.Cos(rad)

		knot.LeftX = Number(x - k*tanX)
		knot.LeftY = Number(y - k*tanY)
		knot.RightX = Number(x + k*tanX)
		knot.RightY = Number(y + k*tanY)

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

// HalfCircle returns the upper half of a unit circle.
// Equivalent to MetaPost's `halfcircle`.
// The path starts at (0.5, 0), goes through (0, 0.5), and ends at (-0.5, 0).
func HalfCircle() *Path {
	r := 0.5
	k := circleKappa

	p := NewPath()

	// 5 points: 0°, 45°, 90°, 135°, 180°
	angles := []float64{0, 45, 90, 135, 180}

	var prev *Knot
	for i, deg := range angles {
		rad := deg * math.Pi / 180
		x := r * math.Cos(rad)
		y := r * math.Sin(rad)

		knot := NewKnot()
		knot.XCoord = Number(x)
		knot.YCoord = Number(y)

		tanX := -math.Sin(rad)
		tanY := math.Cos(rad)

		knot.LeftX = Number(x - k*tanX)
		knot.LeftY = Number(y - k*tanY)
		knot.RightX = Number(x + k*tanX)
		knot.RightY = Number(y + k*tanY)

		if i == 0 {
			knot.LType = KnotEndpoint
			knot.RType = KnotExplicit
		} else if i == len(angles)-1 {
			knot.LType = KnotExplicit
			knot.RType = KnotEndpoint
		} else {
			knot.LType = KnotExplicit
			knot.RType = KnotExplicit
		}

		p.Append(knot)
		if prev != nil {
			prev.Next = knot
			knot.Prev = prev
		}
		prev = knot
	}

	return p
}

// QuarterCircle returns the first quadrant arc of a unit circle.
// Equivalent to MetaPost's `quartercircle`.
// The path starts at (0.5, 0) and ends at (0, 0.5).
func QuarterCircle() *Path {
	r := 0.5
	k := circleKappa

	p := NewPath()

	// 3 points: 0°, 45°, 90°
	angles := []float64{0, 45, 90}

	var prev *Knot
	for i, deg := range angles {
		rad := deg * math.Pi / 180
		x := r * math.Cos(rad)
		y := r * math.Sin(rad)

		knot := NewKnot()
		knot.XCoord = Number(x)
		knot.YCoord = Number(y)

		tanX := -math.Sin(rad)
		tanY := math.Cos(rad)

		knot.LeftX = Number(x - k*tanX)
		knot.LeftY = Number(y - k*tanY)
		knot.RightX = Number(x + k*tanX)
		knot.RightY = Number(y + k*tanY)

		if i == 0 {
			knot.LType = KnotEndpoint
			knot.RType = KnotExplicit
		} else if i == len(angles)-1 {
			knot.LType = KnotExplicit
			knot.RType = KnotEndpoint
		} else {
			knot.LType = KnotExplicit
			knot.RType = KnotExplicit
		}

		p.Append(knot)
		if prev != nil {
			prev.Next = knot
			knot.Prev = prev
		}
		prev = knot
	}

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
