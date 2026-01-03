package draw

import (
	"math"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestFluentShifted(t *testing.T) {
	path := NewPath().
		MoveTo(P(0, 0)).
		LineTo(P(10, 0)).
		Shifted(5, 10)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at (5, 10)
	if solved.Head.XCoord != 5 || solved.Head.YCoord != 10 {
		t.Errorf("Shifted start: got (%f, %f), want (5, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}

	// End should be at (15, 10)
	end := solved.Head.Next
	if end.XCoord != 15 || end.YCoord != 10 {
		t.Errorf("Shifted end: got (%f, %f), want (15, 10)", end.XCoord, end.YCoord)
	}
}

func TestFluentScaled(t *testing.T) {
	path := NewPath().
		MoveTo(P(10, 5)).
		LineTo(P(20, 10)).
		Scaled(2)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at (20, 10)
	if solved.Head.XCoord != 20 || solved.Head.YCoord != 10 {
		t.Errorf("Scaled start: got (%f, %f), want (20, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}

	// End should be at (40, 20)
	end := solved.Head.Next
	if end.XCoord != 40 || end.YCoord != 20 {
		t.Errorf("Scaled end: got (%f, %f), want (40, 20)", end.XCoord, end.YCoord)
	}
}

func TestFluentRotated(t *testing.T) {
	path := NewPath().
		MoveTo(P(10, 0)).
		LineTo(P(20, 0)).
		Rotated(90)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at (0, 10)
	if math.Abs(solved.Head.XCoord) > 0.0001 || math.Abs(solved.Head.YCoord-10) > 0.0001 {
		t.Errorf("Rotated start: got (%f, %f), want (0, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}

	// End should be at (0, 20)
	end := solved.Head.Next
	if math.Abs(end.XCoord) > 0.0001 || math.Abs(end.YCoord-20) > 0.0001 {
		t.Errorf("Rotated end: got (%f, %f), want (0, 20)", end.XCoord, end.YCoord)
	}
}

func TestFluentSlanted(t *testing.T) {
	path := NewPath().
		MoveTo(P(0, 10)).
		LineTo(P(0, 20)).
		Slanted(0.5)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start (0, 10) slanted by 0.5: x' = 0 + 0.5*10 = 5, y' = 10
	if solved.Head.XCoord != 5 || solved.Head.YCoord != 10 {
		t.Errorf("Slanted start: got (%f, %f), want (5, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}

	// End (0, 20) slanted by 0.5: x' = 0 + 0.5*20 = 10, y' = 20
	end := solved.Head.Next
	if end.XCoord != 10 || end.YCoord != 20 {
		t.Errorf("Slanted end: got (%f, %f), want (10, 20)", end.XCoord, end.YCoord)
	}
}

func TestFluentXScaled(t *testing.T) {
	path := NewPath().
		MoveTo(P(10, 10)).
		LineTo(P(20, 10)).
		XScaled(3)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at (30, 10) - only X scaled
	if solved.Head.XCoord != 30 || solved.Head.YCoord != 10 {
		t.Errorf("XScaled start: got (%f, %f), want (30, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentYScaled(t *testing.T) {
	path := NewPath().
		MoveTo(P(10, 10)).
		LineTo(P(10, 20)).
		YScaled(3)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at (10, 30) - only Y scaled
	if solved.Head.XCoord != 10 || solved.Head.YCoord != 30 {
		t.Errorf("YScaled start: got (%f, %f), want (10, 30)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentChainedTransforms(t *testing.T) {
	// Scale first, then shift
	path := NewPath().
		MoveTo(P(10, 10)).
		LineTo(P(20, 10)).
		Scaled(2).
		Shifted(5, 5)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// (10, 10) scaled by 2 = (20, 20), then shifted by (5, 5) = (25, 25)
	if solved.Head.XCoord != 25 || solved.Head.YCoord != 25 {
		t.Errorf("Chained transform start: got (%f, %f), want (25, 25)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentRotatedAround(t *testing.T) {
	// Create a point at (20, 10) and rotate 90 degrees around (10, 10)
	path := NewPath().
		MoveTo(P(20, 10)).
		LineTo(P(20, 10)).
		RotatedAround(10, 10, 90)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// (20, 10) rotated 90 around (10, 10) should go to (10, 20)
	if math.Abs(solved.Head.XCoord-10) > 0.0001 || math.Abs(solved.Head.YCoord-20) > 0.0001 {
		t.Errorf("RotatedAround start: got (%f, %f), want (10, 20)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentScaledAround(t *testing.T) {
	// Create a point at (15, 15) and scale by 2 around (10, 10)
	path := NewPath().
		MoveTo(P(15, 15)).
		LineTo(P(15, 15)).
		ScaledAround(10, 10, 2)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// (15, 15) is 5 units from (10, 10), after scaling by 2 it should be at (20, 20)
	if solved.Head.XCoord != 20 || solved.Head.YCoord != 20 {
		t.Errorf("ScaledAround start: got (%f, %f), want (20, 20)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentZScaled(t *testing.T) {
	// ZScaled(0, 1) is equivalent to Rotated(90)
	path := NewPath().
		MoveTo(P(10, 0)).
		LineTo(P(10, 0)).
		ZScaled(0, 1)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// (10, 0) rotated 90 -> (0, 10)
	if math.Abs(solved.Head.XCoord) > 0.0001 || math.Abs(solved.Head.YCoord-10) > 0.0001 {
		t.Errorf("ZScaled start: got (%f, %f), want (0, 10)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentTransformedPreservesStyle(t *testing.T) {
	path := NewPath().
		WithStrokeColor(mp.ColorCSS("red")).
		WithPen(mp.PenCircle(2)).
		MoveTo(P(0, 0)).
		LineTo(P(10, 0)).
		Scaled(2)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Style should be preserved
	if solved.Style.Stroke.CSS() != "red" {
		t.Errorf("Style not preserved: got stroke %s, want red", solved.Style.Stroke.CSS())
	}
	if solved.Style.Pen == nil {
		t.Error("Pen not preserved")
	}
}

func TestFluentReflectedAbout(t *testing.T) {
	// Reflect point (5, 15) about horizontal line y=10 -> (5, 5)
	path := NewPath().
		MoveTo(P(5, 15)).
		LineTo(P(5, 15)).
		ReflectedAbout(0, 10, 10, 10)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if math.Abs(solved.Head.XCoord-5) > 0.0001 || math.Abs(solved.Head.YCoord-5) > 0.0001 {
		t.Errorf("ReflectedAbout: got (%f, %f), want (5, 5)", solved.Head.XCoord, solved.Head.YCoord)
	}
}

func TestFluentTransformWithCurve(t *testing.T) {
	// A curved path should also transform correctly
	path := NewPath().
		MoveTo(P(0, 0)).
		CurveTo(P(100, 0)).
		Rotated(90)

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Start should be at origin
	if math.Abs(solved.Head.XCoord) > 0.0001 || math.Abs(solved.Head.YCoord) > 0.0001 {
		t.Errorf("Rotated curve start: got (%f, %f), want (0, 0)", solved.Head.XCoord, solved.Head.YCoord)
	}

	// End (100, 0) rotated 90 -> (0, 100)
	end := solved.Head.Next
	if math.Abs(end.XCoord) > 0.0001 || math.Abs(end.YCoord-100) > 0.0001 {
		t.Errorf("Rotated curve end: got (%f, %f), want (0, 100)", end.XCoord, end.YCoord)
	}

	// Control points should also be rotated
	// The right control of start should have been rotated
	if math.Abs(solved.Head.RightX) > 1 {
		t.Errorf("Control point X should be near 0 after 90 degree rotation")
	}
}
