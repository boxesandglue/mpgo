package mp

import (
	"math"
	"testing"
)

func TestArrowHeadEnd_HorizontalLine(t *testing.T) {
	// Create a simple horizontal line from (0,0) to (100,0)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 0, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	// Set control points for straight line
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 0

	arrow := ArrowHeadEnd(p, DefaultAHLength, DefaultAHAngle)
	if arrow == nil {
		t.Fatal("ArrowHeadEnd returned nil")
	}
	if arrow.Head == nil {
		t.Fatal("arrow has no head knot")
	}

	// Arrow tip should be at (100, 0)
	// Find the tip (middle point of the triangle)
	tip := arrow.Head.Next
	if tip == nil {
		t.Fatal("arrow has no tip knot")
	}

	if math.Abs(tip.XCoord-100) > 0.001 || math.Abs(tip.YCoord) > 0.001 {
		t.Errorf("arrow tip at wrong position: got (%.4f, %.4f), want (100, 0)", tip.XCoord, tip.YCoord)
	}

	// Check that left and right base points are symmetric around the line
	left := arrow.Head
	right := tip.Next
	if right == nil {
		t.Fatal("arrow has no right base knot")
	}

	// Both base points should have same X coordinate
	if math.Abs(left.XCoord-right.XCoord) > 0.001 {
		t.Errorf("base points X not equal: left=%.4f, right=%.4f", left.XCoord, right.XCoord)
	}

	// Y coordinates should be opposite
	if math.Abs(left.YCoord+right.YCoord) > 0.001 {
		t.Errorf("base points Y not symmetric: left=%.4f, right=%.4f", left.YCoord, right.YCoord)
	}

	// Base should be at ahlength * cos(ahangle/2) back from tip
	// cos(22.5°) ≈ 0.9239
	expectedBaseX := 100 - DefaultAHLength*math.Cos(DefaultAHAngle*math.Pi/360)
	if math.Abs(left.XCoord-expectedBaseX) > 0.01 {
		t.Errorf("base X coordinate wrong: got %.4f, want %.4f", left.XCoord, expectedBaseX)
	}
}

func TestArrowHeadEnd_VerticalLine(t *testing.T) {
	// Create a vertical line from (0,0) to (0,100)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 0, YCoord: 100, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 0, 100

	arrow := ArrowHeadEnd(p, DefaultAHLength, DefaultAHAngle)
	if arrow == nil {
		t.Fatal("ArrowHeadEnd returned nil")
	}

	tip := arrow.Head.Next
	if tip == nil {
		t.Fatal("arrow has no tip knot")
	}

	if math.Abs(tip.XCoord) > 0.001 || math.Abs(tip.YCoord-100) > 0.001 {
		t.Errorf("arrow tip at wrong position: got (%.4f, %.4f), want (0, 100)", tip.XCoord, tip.YCoord)
	}
}

func TestArrowHeadEnd_DiagonalLine(t *testing.T) {
	// Create a diagonal line from (0,0) to (100,100)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 100, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 100

	arrow := ArrowHeadEnd(p, DefaultAHLength, DefaultAHAngle)
	if arrow == nil {
		t.Fatal("ArrowHeadEnd returned nil")
	}

	tip := arrow.Head.Next
	if tip == nil {
		t.Fatal("arrow has no tip knot")
	}

	if math.Abs(tip.XCoord-100) > 0.001 || math.Abs(tip.YCoord-100) > 0.001 {
		t.Errorf("arrow tip at wrong position: got (%.4f, %.4f), want (100, 100)", tip.XCoord, tip.YCoord)
	}
}

func TestArrowHeadStart_HorizontalLine(t *testing.T) {
	// Create a simple horizontal line from (0,0) to (100,0)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 0, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 0

	arrow := ArrowHeadStart(p, DefaultAHLength, DefaultAHAngle)
	if arrow == nil {
		t.Fatal("ArrowHeadStart returned nil")
	}

	tip := arrow.Head.Next
	if tip == nil {
		t.Fatal("arrow has no tip knot")
	}

	// Arrow tip should be at (0, 0) pointing backward
	if math.Abs(tip.XCoord) > 0.001 || math.Abs(tip.YCoord) > 0.001 {
		t.Errorf("arrow tip at wrong position: got (%.4f, %.4f), want (0, 0)", tip.XCoord, tip.YCoord)
	}

	// Base should be to the right of the tip (in positive X direction)
	left := arrow.Head
	if left.XCoord < tip.XCoord {
		t.Errorf("left base should be to right of tip for start arrow: left.X=%.4f, tip.X=%.4f", left.XCoord, tip.XCoord)
	}
}

func TestShortenPathForArrow_HorizontalLine(t *testing.T) {
	// Create a horizontal line from (0,0) to (100,0)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 0, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 0

	// Shorten by 10 at end
	shortened := ShortenPathForArrow(p, 0, 10)
	if shortened == nil {
		t.Fatal("ShortenPathForArrow returned nil")
	}

	// Find last knot
	last := shortened.Head
	for last.Next != nil && last.Next != shortened.Head {
		last = last.Next
	}

	// End should now be at (90, 0)
	if math.Abs(last.XCoord-90) > 0.001 || math.Abs(last.YCoord) > 0.001 {
		t.Errorf("shortened end at wrong position: got (%.4f, %.4f), want (90, 0)", last.XCoord, last.YCoord)
	}

	// Original should be unchanged
	origLast := p.Head
	for origLast.Next != nil && origLast.Next != p.Head {
		origLast = origLast.Next
	}
	if math.Abs(origLast.XCoord-100) > 0.001 {
		t.Errorf("original path was modified: end.X=%.4f, want 100", origLast.XCoord)
	}
}

func TestShortenPathForArrow_BothEnds(t *testing.T) {
	// Create a horizontal line from (0,0) to (100,0)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 0, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 0

	// Shorten by 10 at both ends
	shortened := ShortenPathForArrow(p, 10, 10)
	if shortened == nil {
		t.Fatal("ShortenPathForArrow returned nil")
	}

	// Start should now be at (10, 0)
	if math.Abs(shortened.Head.XCoord-10) > 0.001 || math.Abs(shortened.Head.YCoord) > 0.001 {
		t.Errorf("shortened start at wrong position: got (%.4f, %.4f), want (10, 0)", shortened.Head.XCoord, shortened.Head.YCoord)
	}

	// Find last knot
	last := shortened.Head
	for last.Next != nil && last.Next != shortened.Head {
		last = last.Next
	}

	// End should now be at (90, 0)
	if math.Abs(last.XCoord-90) > 0.001 || math.Abs(last.YCoord) > 0.001 {
		t.Errorf("shortened end at wrong position: got (%.4f, %.4f), want (90, 0)", last.XCoord, last.YCoord)
	}
}

func TestShortenPathForArrow_DiagonalLine(t *testing.T) {
	// Create a diagonal line from (0,0) to (100,100)
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 100, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 100

	// Shorten by 10*sqrt(2) at end (so the endpoint moves by 10 in each direction)
	shortenAmount := 10 * math.Sqrt(2)
	shortened := ShortenPathForArrow(p, 0, Number(shortenAmount))
	if shortened == nil {
		t.Fatal("ShortenPathForArrow returned nil")
	}

	// Find last knot
	last := shortened.Head
	for last.Next != nil && last.Next != shortened.Head {
		last = last.Next
	}

	// End should now be at approximately (90, 90)
	if math.Abs(last.XCoord-90) > 0.1 || math.Abs(last.YCoord-90) > 0.1 {
		t.Errorf("shortened end at wrong position: got (%.4f, %.4f), want (90, 90)", last.XCoord, last.YCoord)
	}
}

func TestShortenPathForArrow_NilPath(t *testing.T) {
	result := ShortenPathForArrow(nil, 10, 10)
	if result != nil {
		t.Error("ShortenPathForArrow should return nil for nil input")
	}
}

func TestArrowHeadEnd_NilPath(t *testing.T) {
	result := ArrowHeadEnd(nil, DefaultAHLength, DefaultAHAngle)
	if result != nil {
		t.Error("ArrowHeadEnd should return nil for nil input")
	}
}

func TestArrowHeadStart_NilPath(t *testing.T) {
	result := ArrowHeadStart(nil, DefaultAHLength, DefaultAHAngle)
	if result != nil {
		t.Error("ArrowHeadStart should return nil for nil input")
	}
}

func TestArrowHeadGeometry(t *testing.T) {
	// Test that arrow head has correct geometry for MetaPost defaults
	// ahlength = 4bp, ahangle = 45°
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotEndpoint, RType: KnotExplicit}
	k2 := &Knot{XCoord: 100, YCoord: 0, LType: KnotExplicit, RType: KnotEndpoint}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 100, 0

	arrow := ArrowHeadEnd(p, DefaultAHLength, DefaultAHAngle)

	// Get the three points
	left := arrow.Head
	tip := left.Next
	right := tip.Next

	// Calculate distances
	leftToTip := math.Sqrt(math.Pow(tip.XCoord-left.XCoord, 2) + math.Pow(tip.YCoord-left.YCoord, 2))
	rightToTip := math.Sqrt(math.Pow(tip.XCoord-right.XCoord, 2) + math.Pow(tip.YCoord-right.YCoord, 2))

	// Both sides should be equal to ahlength
	if math.Abs(leftToTip-DefaultAHLength) > 0.01 {
		t.Errorf("left side length wrong: got %.4f, want %.4f", leftToTip, DefaultAHLength)
	}
	if math.Abs(rightToTip-DefaultAHLength) > 0.01 {
		t.Errorf("right side length wrong: got %.4f, want %.4f", rightToTip, DefaultAHLength)
	}

	// The half-width at the base should be ahlength * sin(ahangle/2)
	// sin(22.5°) ≈ 0.3827
	expectedHalfWidth := DefaultAHLength * math.Sin(DefaultAHAngle*math.Pi/360)
	actualHalfWidth := math.Abs(left.YCoord)
	if math.Abs(actualHalfWidth-expectedHalfWidth) > 0.01 {
		t.Errorf("half-width wrong: got %.4f, want %.4f", actualHalfWidth, expectedHalfWidth)
	}
}

// DashPattern tests

func TestDashEvenly(t *testing.T) {
	d := DashEvenly()
	if d == nil {
		t.Fatal("DashEvenly() returned nil")
	}
	if len(d.Array) != 2 {
		t.Errorf("expected 2 elements, got %d", len(d.Array))
	}
	if d.Array[0] != 3 || d.Array[1] != 3 {
		t.Errorf("expected [3, 3], got %v", d.Array)
	}
	if d.Offset != 0 {
		t.Errorf("expected offset 0, got %f", d.Offset)
	}
}

func TestDashWithDots(t *testing.T) {
	d := DashWithDots()
	if d == nil {
		t.Fatal("DashWithDots() returned nil")
	}
	// withdots: on 0, off 5 (2.5 + 2.5), offset 2.5
	if d.Array[0] != 0 {
		t.Errorf("expected on=0 for dots, got %f", d.Array[0])
	}
	if d.Array[1] != 5 {
		t.Errorf("expected off=5, got %f", d.Array[1])
	}
	if d.Offset != 2.5 {
		t.Errorf("expected offset=2.5, got %f", d.Offset)
	}
}

func TestNewDashPattern(t *testing.T) {
	d := NewDashPattern(6, 3, 2, 3)
	if d == nil {
		t.Fatal("NewDashPattern returned nil")
	}
	expected := []float64{6, 3, 2, 3}
	if len(d.Array) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(d.Array))
	}
	for i, v := range expected {
		if d.Array[i] != v {
			t.Errorf("element %d: expected %f, got %f", i, v, d.Array[i])
		}
	}
}

func TestNewDashPatternEmpty(t *testing.T) {
	d := NewDashPattern()
	if d != nil {
		t.Error("NewDashPattern with no args should return nil")
	}
}

func TestDashPatternScaled(t *testing.T) {
	original := DashEvenly()
	scaled := original.Scaled(2)

	// Check scaled values
	if scaled.Array[0] != 6 || scaled.Array[1] != 6 {
		t.Errorf("scaled should be [6, 6], got %v", scaled.Array)
	}

	// Original should be unchanged
	if original.Array[0] != 3 || original.Array[1] != 3 {
		t.Errorf("original should be unchanged, got %v", original.Array)
	}
}

func TestDashPatternScaledWithOffset(t *testing.T) {
	d := &DashPattern{Array: []float64{3, 3}, Offset: 1}
	scaled := d.Scaled(2)

	if scaled.Offset != 2 {
		t.Errorf("scaled offset should be 2, got %f", scaled.Offset)
	}
}

func TestDashPatternShifted(t *testing.T) {
	original := DashEvenly()
	shifted := original.Shifted(1.5)

	// Check shifted offset
	if shifted.Offset != 1.5 {
		t.Errorf("shifted offset should be 1.5, got %f", shifted.Offset)
	}

	// Array should be copied
	if shifted.Array[0] != 3 || shifted.Array[1] != 3 {
		t.Errorf("shifted array should be [3, 3], got %v", shifted.Array)
	}

	// Original should be unchanged
	if original.Offset != 0 {
		t.Errorf("original offset should be 0, got %f", original.Offset)
	}
}

func TestDashPatternNilScaled(t *testing.T) {
	var d *DashPattern = nil
	result := d.Scaled(2)
	if result != nil {
		t.Error("scaling nil should return nil")
	}
}

func TestDashPatternNilShifted(t *testing.T) {
	var d *DashPattern = nil
	result := d.Shifted(1)
	if result != nil {
		t.Error("shifting nil should return nil")
	}
}
