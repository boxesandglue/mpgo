package mp

import (
	"math"
	"testing"
)

func TestIdentity(t *testing.T) {
	id := Identity()
	x, y := id.ApplyToPoint(10, 20)
	if x != 10 || y != 20 {
		t.Errorf("Identity should not change point: got (%f, %f), want (10, 20)", x, y)
	}
}

func TestShifted(t *testing.T) {
	tr := Shifted(5, -3)
	x, y := tr.ApplyToPoint(10, 20)
	if x != 15 || y != 17 {
		t.Errorf("Shifted(5,-3) on (10,20): got (%f, %f), want (15, 17)", x, y)
	}
}

func TestScaled(t *testing.T) {
	tr := Scaled(2)
	x, y := tr.ApplyToPoint(10, 20)
	if x != 20 || y != 40 {
		t.Errorf("Scaled(2) on (10,20): got (%f, %f), want (20, 40)", x, y)
	}
}

func TestXScaled(t *testing.T) {
	tr := XScaled(3)
	x, y := tr.ApplyToPoint(10, 20)
	if x != 30 || y != 20 {
		t.Errorf("XScaled(3) on (10,20): got (%f, %f), want (30, 20)", x, y)
	}
}

func TestYScaled(t *testing.T) {
	tr := YScaled(3)
	x, y := tr.ApplyToPoint(10, 20)
	if x != 10 || y != 60 {
		t.Errorf("YScaled(3) on (10,20): got (%f, %f), want (10, 60)", x, y)
	}
}

func TestRotated90(t *testing.T) {
	tr := Rotated(90)
	x, y := tr.ApplyToPoint(10, 0)
	// 90 degree rotation: (10, 0) -> (0, 10)
	if math.Abs(x) > 0.0001 || math.Abs(y-10) > 0.0001 {
		t.Errorf("Rotated(90) on (10,0): got (%f, %f), want (0, 10)", x, y)
	}
}

func TestRotated45(t *testing.T) {
	tr := Rotated(45)
	x, y := tr.ApplyToPoint(10, 0)
	// 45 degree rotation: (10, 0) -> (10*cos45, 10*sin45) ≈ (7.07, 7.07)
	expected := 10 * math.Cos(math.Pi/4)
	if math.Abs(x-expected) > 0.0001 || math.Abs(y-expected) > 0.0001 {
		t.Errorf("Rotated(45) on (10,0): got (%f, %f), want (%f, %f)", x, y, expected, expected)
	}
}

func TestRotated180(t *testing.T) {
	tr := Rotated(180)
	x, y := tr.ApplyToPoint(10, 5)
	// 180 degree rotation: (10, 5) -> (-10, -5)
	if math.Abs(x+10) > 0.0001 || math.Abs(y+5) > 0.0001 {
		t.Errorf("Rotated(180) on (10,5): got (%f, %f), want (-10, -5)", x, y)
	}
}

func TestSlanted(t *testing.T) {
	tr := Slanted(0.5)
	x, y := tr.ApplyToPoint(10, 20)
	// Slant: x' = x + s*y = 10 + 0.5*20 = 20, y' = y = 20
	if x != 20 || y != 20 {
		t.Errorf("Slanted(0.5) on (10,20): got (%f, %f), want (20, 20)", x, y)
	}
}

func TestZScaled(t *testing.T) {
	// ZScaled(0, 1) should rotate by 90 degrees
	tr := ZScaled(0, 1)
	x, y := tr.ApplyToPoint(10, 0)
	if math.Abs(x) > 0.0001 || math.Abs(y-10) > 0.0001 {
		t.Errorf("ZScaled(0,1) on (10,0): got (%f, %f), want (0, 10)", x, y)
	}

	// ZScaled(2, 0) should scale by 2
	tr2 := ZScaled(2, 0)
	x2, y2 := tr2.ApplyToPoint(10, 5)
	if x2 != 20 || y2 != 10 {
		t.Errorf("ZScaled(2,0) on (10,5): got (%f, %f), want (20, 10)", x2, y2)
	}
}

func TestThen(t *testing.T) {
	// Shift then scale
	tr := Shifted(10, 0).Then(Scaled(2))
	x, y := tr.ApplyToPoint(5, 5)
	// First shift: (5, 5) -> (15, 5)
	// Then scale: (15, 5) -> (30, 10)
	if x != 30 || y != 10 {
		t.Errorf("Shifted(10,0).Then(Scaled(2)) on (5,5): got (%f, %f), want (30, 10)", x, y)
	}
}

func TestThenOrder(t *testing.T) {
	// Scale then shift (different order)
	tr := Scaled(2).Then(Shifted(10, 0))
	x, y := tr.ApplyToPoint(5, 5)
	// First scale: (5, 5) -> (10, 10)
	// Then shift: (10, 10) -> (20, 10)
	if x != 20 || y != 10 {
		t.Errorf("Scaled(2).Then(Shifted(10,0)) on (5,5): got (%f, %f), want (20, 10)", x, y)
	}
}

func TestRotatedAround(t *testing.T) {
	// Rotate 90 degrees around (10, 10)
	tr := RotatedAround(10, 10, 90)
	x, y := tr.ApplyToPoint(20, 10)
	// Point is 10 units to the right of center
	// After 90 degree rotation: should be 10 units above center
	if math.Abs(x-10) > 0.0001 || math.Abs(y-20) > 0.0001 {
		t.Errorf("RotatedAround(10,10,90) on (20,10): got (%f, %f), want (10, 20)", x, y)
	}
}

func TestScaledAround(t *testing.T) {
	// Scale by 2 around (10, 10)
	tr := ScaledAround(10, 10, 2)
	x, y := tr.ApplyToPoint(15, 15)
	// Point is 5 units from center in each direction
	// After scaling: should be 10 units from center in each direction
	if x != 20 || y != 20 {
		t.Errorf("ScaledAround(10,10,2) on (15,15): got (%f, %f), want (20, 20)", x, y)
	}
}

func TestInverse(t *testing.T) {
	tr := Shifted(10, 5).Then(Scaled(2)).Then(Rotated(45))
	inv := tr.Inverse()

	x, y := tr.ApplyToPoint(7, 3)
	xBack, yBack := inv.ApplyToPoint(x, y)

	if math.Abs(xBack-7) > 0.0001 || math.Abs(yBack-3) > 0.0001 {
		t.Errorf("Inverse failed: got (%f, %f), want (7, 3)", xBack, yBack)
	}
}

func TestDeterminant(t *testing.T) {
	// Identity has determinant 1
	id := Identity()
	if id.Determinant() != 1 {
		t.Errorf("Identity determinant: got %f, want 1", id.Determinant())
	}

	// Scaled(2) has determinant 4 (scales area by 4)
	s := Scaled(2)
	if s.Determinant() != 4 {
		t.Errorf("Scaled(2) determinant: got %f, want 4", s.Determinant())
	}

	// Rotation preserves area, determinant = 1
	r := Rotated(45)
	if math.Abs(r.Determinant()-1) > 0.0001 {
		t.Errorf("Rotated(45) determinant: got %f, want 1", r.Determinant())
	}
}

func TestApplyToPath(t *testing.T) {
	// Create a simple path
	p := NewPath()
	k1 := &Knot{XCoord: 0, YCoord: 0, LType: KnotExplicit, RType: KnotExplicit}
	k2 := &Knot{XCoord: 10, YCoord: 0, LType: KnotExplicit, RType: KnotExplicit}
	p.Append(k1)
	p.Append(k2)
	k1.RightX, k1.RightY = 0, 0
	k2.LeftX, k2.LeftY = 10, 0

	// Shift by (5, 10)
	shifted := p.Shifted(5, 10)
	if shifted.Head.XCoord != 5 || shifted.Head.YCoord != 10 {
		t.Errorf("Shifted path start: got (%f, %f), want (5, 10)", shifted.Head.XCoord, shifted.Head.YCoord)
	}

	// Original should be unchanged
	if p.Head.XCoord != 0 || p.Head.YCoord != 0 {
		t.Errorf("Original path was modified")
	}
}

func TestPathRotated(t *testing.T) {
	p := NewPath()
	k1 := &Knot{XCoord: 10, YCoord: 0, LType: KnotExplicit, RType: KnotExplicit}
	p.Append(k1)
	k1.RightX, k1.RightY = 10, 0
	k1.LeftX, k1.LeftY = 10, 0

	rotated := p.Rotated(90)
	if math.Abs(rotated.Head.XCoord) > 0.0001 || math.Abs(rotated.Head.YCoord-10) > 0.0001 {
		t.Errorf("Rotated path: got (%f, %f), want (0, 10)", rotated.Head.XCoord, rotated.Head.YCoord)
	}
}

func TestPathScaled(t *testing.T) {
	p := NewPath()
	k1 := &Knot{XCoord: 10, YCoord: 5, LType: KnotExplicit, RType: KnotExplicit}
	p.Append(k1)

	scaled := p.Scaled(3)
	if scaled.Head.XCoord != 30 || scaled.Head.YCoord != 15 {
		t.Errorf("Scaled path: got (%f, %f), want (30, 15)", scaled.Head.XCoord, scaled.Head.YCoord)
	}
}

func TestPathSlanted(t *testing.T) {
	p := NewPath()
	k1 := &Knot{XCoord: 10, YCoord: 20, LType: KnotExplicit, RType: KnotExplicit}
	p.Append(k1)

	slanted := p.Slanted(0.5)
	// x' = x + 0.5*y = 10 + 10 = 20
	if slanted.Head.XCoord != 20 || slanted.Head.YCoord != 20 {
		t.Errorf("Slanted path: got (%f, %f), want (20, 20)", slanted.Head.XCoord, slanted.Head.YCoord)
	}
}

func TestReflectedAboutHorizontal(t *testing.T) {
	// Reflect about horizontal line y=10: (5, 15) -> (5, 5)
	tr := ReflectedAbout(0, 10, 10, 10)
	x, y := tr.ApplyToPoint(5, 15)
	if math.Abs(x-5) > 0.0001 || math.Abs(y-5) > 0.0001 {
		t.Errorf("ReflectedAbout horizontal: got (%f, %f), want (5, 5)", x, y)
	}
}

func TestReflectedAboutVertical(t *testing.T) {
	// Reflect about vertical line x=10: (15, 5) -> (5, 5)
	tr := ReflectedAbout(10, 0, 10, 10)
	x, y := tr.ApplyToPoint(15, 5)
	if math.Abs(x-5) > 0.0001 || math.Abs(y-5) > 0.0001 {
		t.Errorf("ReflectedAbout vertical: got (%f, %f), want (5, 5)", x, y)
	}
}

func TestReflectedAboutDiagonal(t *testing.T) {
	// Reflect about y=x line (through origin): (3, 1) -> (1, 3)
	tr := ReflectedAbout(0, 0, 1, 1)
	x, y := tr.ApplyToPoint(3, 1)
	if math.Abs(x-1) > 0.0001 || math.Abs(y-3) > 0.0001 {
		t.Errorf("ReflectedAbout diagonal: got (%f, %f), want (1, 3)", x, y)
	}
}

func TestPathReflectedAbout(t *testing.T) {
	p := NewPath()
	k1 := &Knot{XCoord: 5, YCoord: 15, LType: KnotExplicit, RType: KnotExplicit}
	p.Append(k1)

	// Reflect about horizontal line y=10
	reflected := p.ReflectedAbout(0, 10, 10, 10)
	if math.Abs(reflected.Head.XCoord-5) > 0.0001 || math.Abs(reflected.Head.YCoord-5) > 0.0001 {
		t.Errorf("ReflectedAbout path: got (%f, %f), want (5, 5)", reflected.Head.XCoord, reflected.Head.YCoord)
	}
}
