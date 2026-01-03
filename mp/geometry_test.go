package mp

import (
	"math"
	"testing"
)

func TestMidPoint(t *testing.T) {
	tests := []struct {
		a, b Point
		want Point
	}{
		{P(0, 0), P(10, 10), P(5, 5)},
		{P(-5, 0), P(5, 0), P(0, 0)},
		{P(0, 0), P(0, 0), P(0, 0)},
		{P(1, 2), P(3, 4), P(2, 3)},
	}
	for _, tt := range tests {
		got := MidPoint(tt.a, tt.b)
		if math.Abs(got.X-tt.want.X) > 1e-10 || math.Abs(got.Y-tt.want.Y) > 1e-10 {
			t.Errorf("MidPoint(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestPointBetween(t *testing.T) {
	a := P(0, 0)
	b := P(10, 20)

	tests := []struct {
		t    float64
		want Point
	}{
		{0, P(0, 0)},
		{1, P(10, 20)},
		{0.5, P(5, 10)},
		{0.25, P(2.5, 5)},
		{-0.5, P(-5, -10)}, // extrapolate before
		{1.5, P(15, 30)},   // extrapolate after
	}
	for _, tt := range tests {
		got := PointBetween(a, b, tt.t)
		if math.Abs(got.X-tt.want.X) > 1e-10 || math.Abs(got.Y-tt.want.Y) > 1e-10 {
			t.Errorf("PointBetween(%v, %v, %v) = %v, want %v", a, b, tt.t, got, tt.want)
		}
	}
}

func TestLineIntersection(t *testing.T) {
	// Two perpendicular lines through origin
	p, ok := LineIntersection(P(-10, 0), P(10, 0), P(0, -10), P(0, 10))
	if !ok {
		t.Error("LineIntersection: expected intersection")
	}
	if math.Abs(p.X) > 1e-10 || math.Abs(p.Y) > 1e-10 {
		t.Errorf("LineIntersection: got %v, want (0,0)", p)
	}

	// Diagonal lines
	p, ok = LineIntersection(P(0, 0), P(10, 10), P(0, 10), P(10, 0))
	if !ok {
		t.Error("LineIntersection: expected intersection")
	}
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y-5) > 1e-10 {
		t.Errorf("LineIntersection: got %v, want (5,5)", p)
	}

	// Parallel lines - no intersection
	_, ok = LineIntersection(P(0, 0), P(10, 0), P(0, 5), P(10, 5))
	if ok {
		t.Error("LineIntersection: expected no intersection for parallel lines")
	}
}

func TestPointOnLineAtX(t *testing.T) {
	// Line from (0,0) to (10,20): y = 2x
	p, ok := PointOnLineAtX(P(0, 0), P(10, 20), 5)
	if !ok {
		t.Error("PointOnLineAtX: expected success")
	}
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y-10) > 1e-10 {
		t.Errorf("PointOnLineAtX: got %v, want (5,10)", p)
	}

	// Vertical line - should fail
	_, ok = PointOnLineAtX(P(5, 0), P(5, 10), 5)
	if ok {
		t.Error("PointOnLineAtX: expected failure for vertical line")
	}
}

func TestPointOnLineAtY(t *testing.T) {
	// Line from (0,0) to (10,20): x = y/2
	p, ok := PointOnLineAtY(P(0, 0), P(10, 20), 10)
	if !ok {
		t.Error("PointOnLineAtY: expected success")
	}
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y-10) > 1e-10 {
		t.Errorf("PointOnLineAtY: got %v, want (5,10)", p)
	}

	// Horizontal line - should fail
	_, ok = PointOnLineAtY(P(0, 5), P(10, 5), 5)
	if ok {
		t.Error("PointOnLineAtY: expected failure for horizontal line")
	}
}

func TestPerpendicularFoot(t *testing.T) {
	// Point (5,5) perpendicular to x-axis
	p := PerpendicularFoot(P(5, 5), P(0, 0), P(10, 0))
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y) > 1e-10 {
		t.Errorf("PerpendicularFoot: got %v, want (5,0)", p)
	}

	// Point (0,0) perpendicular to diagonal line y=x+10
	p = PerpendicularFoot(P(0, 0), P(0, 10), P(10, 20))
	// Foot should be at (-5, 5)
	if math.Abs(p.X-(-5)) > 1e-10 || math.Abs(p.Y-5) > 1e-10 {
		t.Errorf("PerpendicularFoot: got %v, want (-5,5)", p)
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		a, b Point
		want float64
	}{
		{P(0, 0), P(3, 4), 5},
		{P(0, 0), P(0, 0), 0},
		{P(-1, -1), P(2, 3), 5},
	}
	for _, tt := range tests {
		got := Distance(tt.a, tt.b)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("Distance(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestReflection(t *testing.T) {
	// Reflect (5,5) about x-axis
	p := Reflection(P(5, 5), P(0, 0), P(10, 0))
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y-(-5)) > 1e-10 {
		t.Errorf("Reflection about x-axis: got %v, want (5,-5)", p)
	}

	// Reflect (5,0) about y=x line
	p = Reflection(P(5, 0), P(0, 0), P(10, 10))
	if math.Abs(p.X) > 1e-10 || math.Abs(p.Y-5) > 1e-10 {
		t.Errorf("Reflection about y=x: got %v, want (0,5)", p)
	}
}

func TestRotate(t *testing.T) {
	// Rotate (1,0) by 90 degrees
	p := Rotate(P(1, 0), 90)
	if math.Abs(p.X) > 1e-10 || math.Abs(p.Y-1) > 1e-10 {
		t.Errorf("Rotate 90: got %v, want (0,1)", p)
	}

	// Rotate (1,0) by 180 degrees
	p = Rotate(P(1, 0), 180)
	if math.Abs(p.X-(-1)) > 1e-10 || math.Abs(p.Y) > 1e-10 {
		t.Errorf("Rotate 180: got %v, want (-1,0)", p)
	}
}

func TestRotateAround(t *testing.T) {
	// Rotate (10,0) by 90 degrees around (5,0)
	p := RotateAround(P(10, 0), P(5, 0), 90)
	if math.Abs(p.X-5) > 1e-10 || math.Abs(p.Y-5) > 1e-10 {
		t.Errorf("RotateAround: got %v, want (5,5)", p)
	}
}

func TestScaleAround(t *testing.T) {
	// Scale (10,0) by 2 around (5,0)
	p := ScaleAround(P(10, 0), P(5, 0), 2)
	if math.Abs(p.X-15) > 1e-10 || math.Abs(p.Y) > 1e-10 {
		t.Errorf("ScaleAround: got %v, want (15,0)", p)
	}
}

func TestPointMethods(t *testing.T) {
	a := P(3, 4)
	b := P(1, 2)

	// Add
	sum := a.Add(b)
	if sum.X != 4 || sum.Y != 6 {
		t.Errorf("Add: got %v, want (4,6)", sum)
	}

	// Sub
	diff := a.Sub(b)
	if diff.X != 2 || diff.Y != 2 {
		t.Errorf("Sub: got %v, want (2,2)", diff)
	}

	// Mul
	scaled := a.Mul(2)
	if scaled.X != 6 || scaled.Y != 8 {
		t.Errorf("Mul: got %v, want (6,8)", scaled)
	}

	// Length
	if math.Abs(a.Length()-5) > 1e-10 {
		t.Errorf("Length: got %v, want 5", a.Length())
	}

	// Normalized
	n := a.Normalized()
	if math.Abs(n.X-0.6) > 1e-10 || math.Abs(n.Y-0.8) > 1e-10 {
		t.Errorf("Normalized: got %v, want (0.6,0.8)", n)
	}

	// Dot
	dot := a.Dot(b)
	if math.Abs(dot-11) > 1e-10 {
		t.Errorf("Dot: got %v, want 11", dot)
	}

	// Cross
	cross := a.Cross(b)
	if math.Abs(cross-2) > 1e-10 {
		t.Errorf("Cross: got %v, want 2", cross)
	}
}

func TestDir(t *testing.T) {
	tests := []struct {
		angle float64
		want  Point
	}{
		{0, P(1, 0)},
		{90, P(0, 1)},
		{180, P(-1, 0)},
		{270, P(0, -1)},
		{45, P(math.Sqrt(2)/2, math.Sqrt(2)/2)},
	}
	for _, tt := range tests {
		got := Dir(tt.angle)
		if math.Abs(got.X-tt.want.X) > 1e-10 || math.Abs(got.Y-tt.want.Y) > 1e-10 {
			t.Errorf("Dir(%v) = %v, want %v", tt.angle, got, tt.want)
		}
	}
}

func TestAngle(t *testing.T) {
	tests := []struct {
		p    Point
		want float64
	}{
		{P(1, 0), 0},
		{P(0, 1), 90},
		{P(-1, 0), 180},
		{P(0, -1), -90},
	}
	for _, tt := range tests {
		got := tt.p.Angle()
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("(%v).Angle() = %v, want %v", tt.p, got, tt.want)
		}
	}
}
