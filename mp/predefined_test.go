package mp

import (
	"math"
	"testing"
)

func TestFullCircle(t *testing.T) {
	circle := FullCircle()

	if circle == nil || circle.Head == nil {
		t.Fatal("FullCircle returned nil")
	}

	// Count knots - should be 8
	count := 0
	k := circle.Head
	for {
		count++
		k = k.Next
		if k == circle.Head {
			break
		}
	}
	if count != 8 {
		t.Errorf("expected 8 knots, got %d", count)
	}

	// Check first point is at (0.5, 0)
	if math.Abs(float64(circle.Head.XCoord)-0.5) > 1e-10 {
		t.Errorf("first point X: expected 0.5, got %v", circle.Head.XCoord)
	}
	if math.Abs(float64(circle.Head.YCoord)) > 1e-10 {
		t.Errorf("first point Y: expected 0, got %v", circle.Head.YCoord)
	}

	// Check that all points are at radius 0.5
	k = circle.Head
	for {
		x := float64(k.XCoord)
		y := float64(k.YCoord)
		r := math.Sqrt(x*x + y*y)
		if math.Abs(r-0.5) > 1e-10 {
			t.Errorf("point (%v, %v) has radius %v, expected 0.5", x, y, r)
		}
		k = k.Next
		if k == circle.Head {
			break
		}
	}
}

func TestHalfCircle(t *testing.T) {
	half := HalfCircle()

	if half == nil || half.Head == nil {
		t.Fatal("HalfCircle returned nil")
	}

	// Count knots - should be 5
	count := 0
	k := half.Head
	for {
		count++
		if k.Next == nil || k.RType == KnotEndpoint {
			break
		}
		k = k.Next
	}
	if count != 5 {
		t.Errorf("expected 5 knots, got %d", count)
	}

	// Check first point is at (0.5, 0)
	if math.Abs(float64(half.Head.XCoord)-0.5) > 1e-10 {
		t.Errorf("first point X: expected 0.5, got %v", half.Head.XCoord)
	}

	// Check last point is at (-0.5, 0)
	last := half.Head
	for last.Next != nil && last.RType != KnotEndpoint {
		last = last.Next
	}
	if math.Abs(float64(last.XCoord)+0.5) > 1e-10 {
		t.Errorf("last point X: expected -0.5, got %v", last.XCoord)
	}
}

func TestQuarterCircle(t *testing.T) {
	quarter := QuarterCircle()

	if quarter == nil || quarter.Head == nil {
		t.Fatal("QuarterCircle returned nil")
	}

	// Count knots - should be 3
	count := 0
	k := quarter.Head
	for {
		count++
		if k.Next == nil || k.RType == KnotEndpoint {
			break
		}
		k = k.Next
	}
	if count != 3 {
		t.Errorf("expected 3 knots, got %d", count)
	}

	// Check first point is at (0.5, 0)
	if math.Abs(float64(quarter.Head.XCoord)-0.5) > 1e-10 {
		t.Errorf("first point X: expected 0.5, got %v", quarter.Head.XCoord)
	}

	// Check last point is at (0, 0.5)
	last := quarter.Head
	for last.Next != nil && last.RType != KnotEndpoint {
		last = last.Next
	}
	if math.Abs(float64(last.XCoord)) > 1e-10 {
		t.Errorf("last point X: expected 0, got %v", last.XCoord)
	}
	if math.Abs(float64(last.YCoord)-0.5) > 1e-10 {
		t.Errorf("last point Y: expected 0.5, got %v", last.YCoord)
	}
}

func TestUnitSquare(t *testing.T) {
	square := UnitSquare()

	if square == nil || square.Head == nil {
		t.Fatal("UnitSquare returned nil")
	}

	// Count knots - should be 4
	count := 0
	k := square.Head
	for {
		count++
		k = k.Next
		if k == square.Head {
			break
		}
	}
	if count != 4 {
		t.Errorf("expected 4 knots, got %d", count)
	}

	// Check corners
	expected := [][2]float64{
		{0, 0},
		{1, 0},
		{1, 1},
		{0, 1},
	}
	k = square.Head
	for i, exp := range expected {
		if math.Abs(float64(k.XCoord)-exp[0]) > 1e-10 {
			t.Errorf("corner %d X: expected %v, got %v", i, exp[0], k.XCoord)
		}
		if math.Abs(float64(k.YCoord)-exp[1]) > 1e-10 {
			t.Errorf("corner %d Y: expected %v, got %v", i, exp[1], k.YCoord)
		}
		k = k.Next
	}

	// Check that control points equal knot coordinates (straight lines)
	k = square.Head
	for {
		if k.LeftX != k.XCoord || k.LeftY != k.YCoord {
			t.Errorf("left control point should equal knot for straight lines")
		}
		if k.RightX != k.XCoord || k.RightY != k.YCoord {
			t.Errorf("right control point should equal knot for straight lines")
		}
		k = k.Next
		if k == square.Head {
			break
		}
	}
}
