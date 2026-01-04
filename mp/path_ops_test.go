package mp

import (
	"math"
	"testing"
)

// Helper to create a simple solved path: (0,0)..(100,0)
func makeSimplePath() *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = 0
	k1.YCoord = 0
	k1.RightX = 33.33333 // Typical control point for a straight-ish curve
	k1.RightY = 0
	k1.LeftX = 0
	k1.LeftY = 0
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = 100
	k2.YCoord = 0
	k2.LeftX = 66.66667
	k2.LeftY = 0
	k2.RightX = 100
	k2.RightY = 0
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	p.Append(k2)

	return p
}

// Helper to create a line path: (0,0)--(100,0)
// Control points are placed on the line to ensure linear interpolation.
func makeLinePath() *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = 0
	k1.YCoord = 0
	// For a straight line, control points should be on the line itself
	// RightX at 1/3 of the way (typical for straight MetaPost lines)
	k1.RightX = 33.33333
	k1.RightY = 0
	k1.LeftX = 0
	k1.LeftY = 0
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = 100
	k2.YCoord = 0
	// LeftX at 2/3 of the way
	k2.LeftX = 66.66667
	k2.LeftY = 0
	k2.RightX = 100
	k2.RightY = 0
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	p.Append(k2)

	return p
}

// Helper to create a multi-segment path: (0,0)--(100,0)--(100,100)
func makeMultiSegmentPath() *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = 0
	k1.YCoord = 0
	k1.RightX = 0
	k1.RightY = 0
	k1.LeftX = 0
	k1.LeftY = 0
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = 100
	k2.YCoord = 0
	k2.LeftX = 100
	k2.LeftY = 0
	k2.RightX = 100
	k2.RightY = 0
	k2.LType = KnotExplicit
	k2.RType = KnotExplicit
	p.Append(k2)

	k3 := NewKnot()
	k3.XCoord = 100
	k3.YCoord = 100
	k3.LeftX = 100
	k3.LeftY = 100
	k3.RightX = 100
	k3.RightY = 100
	k3.LType = KnotExplicit
	k3.RType = KnotEndpoint
	p.Append(k3)

	return p
}

// Helper to create a square cycle: (0,0)--(100,0)--(100,100)--(0,100)--cycle
func makeSquareCycle() *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = 0
	k1.YCoord = 0
	k1.RightX = 0
	k1.RightY = 0
	k1.LeftX = 0
	k1.LeftY = 0
	k1.LType = KnotExplicit
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = 100
	k2.YCoord = 0
	k2.LeftX = 100
	k2.LeftY = 0
	k2.RightX = 100
	k2.RightY = 0
	k2.LType = KnotExplicit
	k2.RType = KnotExplicit
	p.Append(k2)

	k3 := NewKnot()
	k3.XCoord = 100
	k3.YCoord = 100
	k3.LeftX = 100
	k3.LeftY = 100
	k3.RightX = 100
	k3.RightY = 100
	k3.LType = KnotExplicit
	k3.RType = KnotExplicit
	p.Append(k3)

	k4 := NewKnot()
	k4.XCoord = 0
	k4.YCoord = 100
	k4.LeftX = 0
	k4.LeftY = 100
	k4.RightX = 0
	k4.RightY = 100
	k4.LType = KnotExplicit
	k4.RType = KnotExplicit
	p.Append(k4)

	return p
}

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

func TestPathLength(t *testing.T) {
	tests := []struct {
		name     string
		path     *Path
		expected int
	}{
		{"nil path", nil, 0},
		{"empty path", NewPath(), 0},
		{"single segment", makeLinePath(), 1},
		{"two segments", makeMultiSegmentPath(), 2},
		{"square cycle", makeSquareCycle(), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.path.PathLength()
			if got != tt.expected {
				t.Errorf("PathLength() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPointOf_Line(t *testing.T) {
	p := makeLinePath()

	tests := []struct {
		t         Number
		expectedX Number
		expectedY Number
	}{
		{0, 0, 0},
		{0.5, 50, 0},
		{1, 100, 0},
		{0.25, 25, 0},
		{0.75, 75, 0},
	}

	for _, tt := range tests {
		x, y := p.PointOf(tt.t)
		if !approxEqual(x, tt.expectedX, 0.001) || !approxEqual(y, tt.expectedY, 0.001) {
			t.Errorf("PointOf(%g) = (%g, %g), want (%g, %g)", tt.t, x, y, tt.expectedX, tt.expectedY)
		}
	}
}

func TestPointOf_MultiSegment(t *testing.T) {
	p := makeMultiSegmentPath()

	tests := []struct {
		t         Number
		expectedX Number
		expectedY Number
	}{
		{0, 0, 0},
		{1, 100, 0},
		{2, 100, 100},
		{0.5, 50, 0},
		{1.5, 100, 50},
	}

	for _, tt := range tests {
		x, y := p.PointOf(tt.t)
		if !approxEqual(x, tt.expectedX, 0.001) || !approxEqual(y, tt.expectedY, 0.001) {
			t.Errorf("PointOf(%g) = (%g, %g), want (%g, %g)", tt.t, x, y, tt.expectedX, tt.expectedY)
		}
	}
}

func TestPointOf_Cycle(t *testing.T) {
	p := makeSquareCycle()

	tests := []struct {
		t         Number
		expectedX Number
		expectedY Number
	}{
		{0, 0, 0},
		{1, 100, 0},
		{2, 100, 100},
		{3, 0, 100},
		{4, 0, 0}, // wraps around
	}

	for _, tt := range tests {
		x, y := p.PointOf(tt.t)
		if !approxEqual(x, tt.expectedX, 0.001) || !approxEqual(y, tt.expectedY, 0.001) {
			t.Errorf("PointOf(%g) = (%g, %g), want (%g, %g)", tt.t, x, y, tt.expectedX, tt.expectedY)
		}
	}
}

func TestDirectionOf_Line(t *testing.T) {
	p := makeLinePath()

	// For a straight line (0,0) to (100,0), direction should be (positive, 0)
	dx, dy := p.DirectionOf(0.5)
	if dx <= 0 {
		t.Errorf("DirectionOf(0.5) dx = %g, want positive", dx)
	}
	if !approxEqual(dy, 0, 0.001) {
		t.Errorf("DirectionOf(0.5) dy = %g, want 0", dy)
	}

	// Direction should be consistent along the line
	dx0, dy0 := p.DirectionOf(0)
	dx1, dy1 := p.DirectionOf(1)
	if !approxEqual(dx0, dx1, 0.001) || !approxEqual(dy0, dy1, 0.001) {
		t.Errorf("Direction not consistent: (%.2g,%.2g) vs (%.2g,%.2g)", dx0, dy0, dx1, dy1)
	}
}

func TestDirectionOf_MultiSegment(t *testing.T) {
	p := makeMultiSegmentPath() // (0,0)--(100,0)--(100,100)

	// At t=0.5, direction should be along x-axis
	dx, dy := p.DirectionOf(0.5)
	if dx <= 0 {
		t.Errorf("DirectionOf(0.5) dx = %g, want positive", dx)
	}
	if !approxEqual(dy, 0, 0.001) {
		t.Errorf("DirectionOf(0.5) dy = %g, want 0", dy)
	}

	// At t=1.5, direction should be along y-axis
	dx, dy = p.DirectionOf(1.5)
	if !approxEqual(dx, 0, 0.001) {
		t.Errorf("DirectionOf(1.5) dx = %g, want 0", dx)
	}
	if dy <= 0 {
		t.Errorf("DirectionOf(1.5) dy = %g, want positive", dy)
	}
}

func TestSubpath_SingleSegment(t *testing.T) {
	p := makeLinePath() // (0,0)--(100,0)

	// subpath (0.25, 0.75) should give (25,0)--(75,0)
	sub := p.Subpath(0.25, 0.75)

	if sub.PathLength() != 1 {
		t.Errorf("Subpath should have 1 segment, got %d", sub.PathLength())
	}

	x0, y0 := sub.PointOf(0)
	if !approxEqual(x0, 25, 0.001) || !approxEqual(y0, 0, 0.001) {
		t.Errorf("Subpath start = (%g, %g), want (25, 0)", x0, y0)
	}

	x1, y1 := sub.PointOf(1)
	if !approxEqual(x1, 75, 0.001) || !approxEqual(y1, 0, 0.001) {
		t.Errorf("Subpath end = (%g, %g), want (75, 0)", x1, y1)
	}
}

func TestSubpath_MultiSegment(t *testing.T) {
	p := makeMultiSegmentPath() // (0,0)--(100,0)--(100,100)

	// subpath (0.5, 1.5) should span from (50,0) to (100,50)
	sub := p.Subpath(0.5, 1.5)

	x0, y0 := sub.PointOf(0)
	if !approxEqual(x0, 50, 0.001) || !approxEqual(y0, 0, 0.001) {
		t.Errorf("Subpath start = (%g, %g), want (50, 0)", x0, y0)
	}

	// End point
	n := sub.PathLength()
	xn, yn := sub.PointOf(Number(n))
	if !approxEqual(xn, 100, 0.001) || !approxEqual(yn, 50, 0.001) {
		t.Errorf("Subpath end = (%g, %g), want (100, 50)", xn, yn)
	}
}

func TestSubpath_Reversed(t *testing.T) {
	p := makeLinePath() // (0,0)--(100,0)

	// subpath (0.75, 0.25) should give reversed path
	sub := p.Subpath(0.75, 0.25)

	x0, y0 := sub.PointOf(0)
	if !approxEqual(x0, 75, 0.001) || !approxEqual(y0, 0, 0.001) {
		t.Errorf("Reversed subpath start = (%g, %g), want (75, 0)", x0, y0)
	}

	x1, y1 := sub.PointOf(1)
	if !approxEqual(x1, 25, 0.001) || !approxEqual(y1, 0, 0.001) {
		t.Errorf("Reversed subpath end = (%g, %g), want (25, 0)", x1, y1)
	}
}

func TestSubpath_FullPath(t *testing.T) {
	p := makeLinePath() // (0,0)--(100,0)

	// subpath (0, 1) should give the full path
	sub := p.Subpath(0, 1)

	x0, y0 := sub.PointOf(0)
	if !approxEqual(x0, 0, 0.001) || !approxEqual(y0, 0, 0.001) {
		t.Errorf("Full subpath start = (%g, %g), want (0, 0)", x0, y0)
	}

	x1, y1 := sub.PointOf(1)
	if !approxEqual(x1, 100, 0.001) || !approxEqual(y1, 0, 0.001) {
		t.Errorf("Full subpath end = (%g, %g), want (100, 0)", x1, y1)
	}
}

func TestReversed(t *testing.T) {
	p := makeMultiSegmentPath() // (0,0)--(100,0)--(100,100)

	rev := p.Reversed()

	// Check that reversed path goes (100,100)--(100,0)--(0,0)
	x0, y0 := rev.PointOf(0)
	if !approxEqual(x0, 100, 0.001) || !approxEqual(y0, 100, 0.001) {
		t.Errorf("Reversed start = (%g, %g), want (100, 100)", x0, y0)
	}

	x1, y1 := rev.PointOf(1)
	if !approxEqual(x1, 100, 0.001) || !approxEqual(y1, 0, 0.001) {
		t.Errorf("Reversed middle = (%g, %g), want (100, 0)", x1, y1)
	}

	x2, y2 := rev.PointOf(2)
	if !approxEqual(x2, 0, 0.001) || !approxEqual(y2, 0, 0.001) {
		t.Errorf("Reversed end = (%g, %g), want (0, 0)", x2, y2)
	}
}

func TestEvalCubic(t *testing.T) {
	// Test with a simple line: (0,0) to (100,0)
	// Control points on the line: (33.33, 0) and (66.67, 0)
	p0x, p0y := 0.0, 0.0
	p1x, p1y := 33.33, 0.0
	p2x, p2y := 66.67, 0.0
	p3x, p3y := 100.0, 0.0

	tests := []struct {
		t  Number
		ex Number
		ey Number
	}{
		{0, 0, 0},
		{1, 100, 0},
		{0.5, 50, 0},
	}

	for _, tt := range tests {
		x, y := evalCubic(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, tt.t)
		if !approxEqual(x, tt.ex, 0.01) || !approxEqual(y, tt.ey, 0.01) {
			t.Errorf("evalCubic(t=%g) = (%g, %g), want (%g, %g)", tt.t, x, y, tt.ex, tt.ey)
		}
	}
}

func TestSplitCubicCoords(t *testing.T) {
	// Split a line at t=0.5
	p0x, p0y := 0.0, 0.0
	p1x, p1y := 0.0, 0.0
	p2x, p2y := 100.0, 0.0
	p3x, p3y := 100.0, 0.0

	a0x, a0y, _, _, _, _, a3x, a3y,
		b0x, b0y, _, _, _, _, b3x, b3y := splitCubicCoords(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, 0.5)

	// First half should go from (0,0) to (50,0)
	if !approxEqual(a0x, 0, 0.001) || !approxEqual(a0y, 0, 0.001) {
		t.Errorf("First half start = (%g, %g), want (0, 0)", a0x, a0y)
	}
	if !approxEqual(a3x, 50, 0.001) || !approxEqual(a3y, 0, 0.001) {
		t.Errorf("First half end = (%g, %g), want (50, 0)", a3x, a3y)
	}

	// Second half should go from (50,0) to (100,0)
	if !approxEqual(b0x, 50, 0.001) || !approxEqual(b0y, 0, 0.001) {
		t.Errorf("Second half start = (%g, %g), want (50, 0)", b0x, b0y)
	}
	if !approxEqual(b3x, 100, 0.001) || !approxEqual(b3y, 0, 0.001) {
		t.Errorf("Second half end = (%g, %g), want (100, 0)", b3x, b3y)
	}
}

// Helper to create a horizontal line from (x1,y) to (x2,y)
func makeHorizontalLine(x1, x2, y Number) *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = x1
	k1.YCoord = y
	k1.LeftX = x1
	k1.LeftY = y
	k1.RightX = x1 + (x2-x1)/3
	k1.RightY = y
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = x2
	k2.YCoord = y
	k2.LeftX = x1 + 2*(x2-x1)/3
	k2.LeftY = y
	k2.RightX = x2
	k2.RightY = y
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	p.Append(k2)

	return p
}

// Helper to create a vertical line from (x,y1) to (x,y2)
func makeVerticalLine(x, y1, y2 Number) *Path {
	p := NewPath()

	k1 := NewKnot()
	k1.XCoord = x
	k1.YCoord = y1
	k1.LeftX = x
	k1.LeftY = y1
	k1.RightX = x
	k1.RightY = y1 + (y2-y1)/3
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord = x
	k2.YCoord = y2
	k2.LeftX = x
	k2.LeftY = y1 + 2*(y2-y1)/3
	k2.RightX = x
	k2.RightY = y2
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	p.Append(k2)

	return p
}

func TestBuildCycle_TwoLines(t *testing.T) {
	// Two intersecting lines: horizontal (0,50)--(100,50) and vertical (50,0)--(50,100)
	// Should create a path with 2 segments forming a "corner"
	horiz := makeHorizontalLine(0, 100, 50)
	vert := makeVerticalLine(50, 0, 100)

	result := BuildCycle(horiz, vert)
	if result == nil {
		t.Fatal("BuildCycle returned nil")
	}

	// The cycle should contain points near (50,50) where the lines intersect
	// The resulting path goes from intersection on horiz to intersection on vert
	if result.Head == nil {
		t.Fatal("BuildCycle result has no head")
	}

	// Check that we have at least one knot
	x0, y0 := result.PointOf(0)
	t.Logf("BuildCycle result point 0: (%g, %g)", x0, y0)

	// The intersection point should be at (50, 50)
	if !approxEqual(x0, 50, 1) || !approxEqual(y0, 50, 1) {
		t.Errorf("Expected point near (50, 50), got (%g, %g)", x0, y0)
	}
}

func TestBuildCycle_FourLines_Square(t *testing.T) {
	// Four lines forming a square:
	// bottom: (-10,0)--(110,0)
	// right: (100,-10)--(100,110)
	// top: (110,100)--(-10,100)
	// left: (0,110)--(0,-10)
	// These extend beyond the square corners to ensure intersections

	bottom := makeHorizontalLine(-10, 110, 0)
	right := makeVerticalLine(100, -10, 110)
	top := makeHorizontalLine(110, -10, 100) // reversed direction
	left := makeVerticalLine(0, 110, -10)    // reversed direction

	result := BuildCycle(bottom, right, top, left)
	if result == nil {
		t.Fatal("BuildCycle returned nil for square")
	}

	// Check the four corners
	n := result.PathLength()
	t.Logf("BuildCycle square has %d segments", n)

	// Should have at least 4 segments for a square (may have 5 if cycle closes)
	if n < 4 {
		t.Errorf("Expected at least 4 segments, got %d", n)
	}

	// Check corner points - buildcycle returns corners in order:
	// (bottom∩left), (bottom∩right), (right∩top), (top∩left), cycle back
	expectedCorners := []struct{ x, y Number }{
		{0, 0},     // bottom-left (intersection of bottom with left)
		{100, 0},   // bottom-right (intersection of bottom with right)
		{100, 100}, // top-right (intersection of right with top)
		{0, 100},   // top-left (intersection of top with left)
	}

	for i, ec := range expectedCorners {
		x, y := result.PointOf(Number(i))
		t.Logf("Corner %d: (%g, %g)", i, x, y)
		if !approxEqual(x, ec.x, 1) || !approxEqual(y, ec.y, 1) {
			t.Errorf("Corner %d: got (%g, %g), want (%g, %g)", i, x, y, ec.x, ec.y)
		}
	}
}

func TestBuildCycle_NoIntersection(t *testing.T) {
	// Two parallel lines that don't intersect
	line1 := makeHorizontalLine(0, 100, 0)
	line2 := makeHorizontalLine(0, 100, 100)

	result := BuildCycle(line1, line2)
	if result != nil {
		t.Error("BuildCycle should return nil for non-intersecting paths")
	}
}

func TestBuildCycle_SinglePath(t *testing.T) {
	// Single path should return nil (need at least 2)
	line := makeHorizontalLine(0, 100, 50)

	result := BuildCycle(line)
	if result != nil {
		t.Error("BuildCycle should return nil for single path")
	}
}

// TestBuildCycle_CurveWithLines tests buildcycle with a curve and vertical lines.
// This is a regression test for a bug where control points of intermediate
// subpaths were incorrectly overwritten when joining subsequent subpaths.
func TestBuildCycle_CurveWithLines(t *testing.T) {
	// Create a solved curve: (0,0)..(50,80)..(100,0)
	curve := createSolvedCurve()

	// Vertical lines at x=20 and x=80
	leftLine := makeVerticalLine(20, -10, 90)
	rightLine := makeVerticalLine(80, 90, -10) // reversed direction

	// Reversed curve
	reverseCurve := curve.Reversed()

	// buildcycle(curve, rightLine, reverseCurve, leftLine)
	result := BuildCycle(curve, rightLine, reverseCurve, leftLine)
	if result == nil {
		t.Fatal("BuildCycle returned nil")
	}

	// The result should have knots that form a "lens" shape
	// going through the curve's peak at y≈80

	// Find the knot closest to x=50 (the peak)
	var peakKnot *Knot
	cur := result.Head
	for {
		if cur.XCoord > 45 && cur.XCoord < 55 && cur.YCoord > 75 {
			peakKnot = cur
			break
		}
		cur = cur.Next
		if cur == result.Head {
			break
		}
	}

	if peakKnot == nil {
		t.Fatal("Could not find peak knot near (50, 80)")
	}

	// The peak knot's control points should be curve-like, not vertical
	// RightX should be around 38-62 (moving towards next point on curve)
	// NOT at x=20 or x=80 (which would indicate the bug)
	if peakKnot.RightX < 30 || peakKnot.RightX > 70 {
		t.Errorf("Peak knot RightX=%.2f, expected between 30-70 (curve control point)", peakKnot.RightX)
	}
	if peakKnot.LeftX < 30 || peakKnot.LeftX > 70 {
		t.Errorf("Peak knot LeftX=%.2f, expected between 30-70 (curve control point)", peakKnot.LeftX)
	}

	t.Logf("Peak knot: (%.2f, %.2f) L=(%.2f, %.2f) R=(%.2f, %.2f)",
		peakKnot.XCoord, peakKnot.YCoord,
		peakKnot.LeftX, peakKnot.LeftY,
		peakKnot.RightX, peakKnot.RightY)
}

// createSolvedCurve creates a solved curve path: (0,0)..(50,80)..(100,0)
// with MetaPost-like control points.
func createSolvedCurve() *Path {
	p := NewPath()

	// Knot 0: (0,0)
	k0 := NewKnot()
	k0.XCoord, k0.YCoord = 0, 0
	k0.LeftX, k0.LeftY = 0, 0
	k0.RightX, k0.RightY = -18.01, 36.95 // control towards (50,80)
	k0.LType = KnotEndpoint
	k0.RType = KnotExplicit
	p.Append(k0)

	// Knot 1: (50,80) - the peak
	k1 := NewKnot()
	k1.XCoord, k1.YCoord = 50, 80
	k1.LeftX, k1.LeftY = 8.89, 80    // control from (0,0)
	k1.RightX, k1.RightY = 91.11, 80 // control towards (100,0)
	k1.LType = KnotExplicit
	k1.RType = KnotExplicit
	p.Append(k1)

	// Knot 2: (100,0)
	k2 := NewKnot()
	k2.XCoord, k2.YCoord = 100, 0
	k2.LeftX, k2.LeftY = 118.01, 36.95 // control from (50,80)
	k2.RightX, k2.RightY = 100, 0
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	p.Append(k2)

	return p
}

// TestDirectionTimeOf tests the directiontime operation.
func TestDirectionTimeOf(t *testing.T) {
	// Test 1: Horizontal line - direction is always (1,0)
	line := makeLinePath() // (0,0)--(100,0)
	tRight := line.DirectionTimeOf(1, 0)
	if tRight < 0 {
		t.Errorf("horizontal line should have direction (1,0), got t=%v", tRight)
	}

	// Direction (0,1) should not exist on horizontal line
	tUp := line.DirectionTimeOf(0, 1)
	if tUp >= 0 {
		t.Errorf("horizontal line should not have direction (0,1), got t=%v", tUp)
	}

	// Test 2: Arc path - should have various directions
	arc := createSolvedCurve() // (0,0)..(50,80)..(100,0)

	// Find where tangent is horizontal (pointing right)
	tHoriz := arc.DirectionTimeOf(1, 0)
	if tHoriz < 0 {
		t.Errorf("arc should have horizontal tangent somewhere, got t=%v", tHoriz)
	}
	// At horizontal tangent, the direction should be parallel to (1,0)
	if tHoriz >= 0 {
		dx, dy := arc.DirectionOf(tHoriz)
		// Check that direction is approximately horizontal
		if math.Abs(float64(dy)) > 0.01*math.Abs(float64(dx)) && dx != 0 {
			t.Errorf("at t=%v, direction should be horizontal, got (%v, %v)", tHoriz, dx, dy)
		}
	}

	// Test 3: Zero direction vector should return -1
	tZero := arc.DirectionTimeOf(0, 0)
	if tZero != -1 {
		t.Errorf("zero direction should return -1, got %v", tZero)
	}

	// Test 4: Nil path should return -1
	var nilPath *Path
	tNil := nilPath.DirectionTimeOf(1, 0)
	if tNil != -1 {
		t.Errorf("nil path should return -1, got %v", tNil)
	}
}

// TestDirectionPointOf tests the directionpoint operation.
func TestDirectionPointOf(t *testing.T) {
	arc := createSolvedCurve() // (0,0)..(50,80)..(100,0)

	// Find point where tangent is horizontal
	x, y, found := arc.DirectionPointOf(1, 0)
	if !found {
		t.Error("arc should have a point with horizontal tangent")
	}
	if found {
		// The point should be near the top of the arc
		// For this symmetric arc, x should be around 50
		if x < 30 || x > 70 {
			t.Errorf("horizontal tangent point x=%v should be near 50", x)
		}
		// y should be near the maximum (around 80)
		if y < 60 {
			t.Errorf("horizontal tangent point y=%v should be near the top", y)
		}
	}

	// Direction not on path
	_, _, found = arc.DirectionPointOf(-1, 0) // leftward
	// Note: The arc goes up then down, so leftward direction might exist on descent
	// This is just checking the function works, not the specific result
}

// TestDirectionTimeConsistency verifies that directiontime is inverse of direction.
func TestDirectionTimeConsistency(t *testing.T) {
	arc := createSolvedCurve()

	// For various t values, get direction and verify directiontime finds it
	testTs := []Number{0.1, 0.3, 0.5, 0.7, 0.9, 1.2, 1.5, 1.8}

	for _, testT := range testTs {
		dx, dy := arc.DirectionOf(testT)
		if dx == 0 && dy == 0 {
			continue // Skip degenerate directions
		}

		foundT := arc.DirectionTimeOf(dx, dy)
		if foundT < 0 {
			t.Errorf("direction at t=%v should be findable, got -1", testT)
			continue
		}

		// The found t should give approximately the same direction
		foundDx, foundDy := arc.DirectionOf(foundT)

		// Normalize both for comparison
		len1 := math.Sqrt(float64(dx*dx + dy*dy))
		len2 := math.Sqrt(float64(foundDx*foundDx + foundDy*foundDy))
		if len1 > 0 && len2 > 0 {
			ndx1, ndy1 := float64(dx)/len1, float64(dy)/len1
			ndx2, ndy2 := float64(foundDx)/len2, float64(foundDy)/len2

			// Directions should be parallel (dot product ≈ 1)
			dot := ndx1*ndx2 + ndy1*ndy2
			if dot < 0.99 {
				t.Errorf("at t=%v: direction (%v,%v) found at t=%v gives (%v,%v), dot=%v",
					testT, dx, dy, foundT, foundDx, foundDy, dot)
			}
		}
	}
}

// TestSolveQuadratic tests the quadratic equation solver.
func TestSolveQuadratic(t *testing.T) {
	tests := []struct {
		name     string
		a, b, c  Number
		expected []Number
	}{
		{"two roots", 1, -3, 2, []Number{1, 2}}, // x²-3x+2=0 → x=1,2
		{"one root", 1, -2, 1, []Number{1}},     // x²-2x+1=0 → x=1
		{"no real roots", 1, 0, 1, nil},         // x²+1=0 → no real
		{"linear", 0, 2, -4, []Number{2}},       // 2x-4=0 → x=2
		{"constant zero", 0, 0, 5, nil},         // 5=0 → no solution
		{"negative discriminant", 1, 1, 1, nil}, // x²+x+1=0 → no real
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roots := solveQuadratic(tt.a, tt.b, tt.c)
			if len(roots) != len(tt.expected) {
				t.Errorf("expected %d roots, got %d: %v", len(tt.expected), len(roots), roots)
				return
			}
			for i, exp := range tt.expected {
				if math.Abs(float64(roots[i]-exp)) > 0.001 {
					t.Errorf("root[%d]: expected %v, got %v", i, exp, roots[i])
				}
			}
		})
	}
}
