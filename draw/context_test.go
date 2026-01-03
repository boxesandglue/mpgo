package draw

import (
	"math"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestContext_KnownPoint(t *testing.T) {
	ctx := NewContext()
	z := ctx.Known(10, 20)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := z.XY()
	if x != 10 || y != 20 {
		t.Errorf("expected (10, 20), got (%v, %v)", x, y)
	}
}

func TestContext_MidPoint(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(0, 0)
	b := ctx.Known(100, 100)
	m := ctx.MidPointOf(a, b)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := m.XY()
	if math.Abs(x-50) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (50, 50), got (%v, %v)", x, y)
	}
}

func TestContext_Between(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(0, 0)
	b := ctx.Known(100, 0)
	p := ctx.BetweenAt(a, b, 0.25)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := p.XY()
	if math.Abs(x-25) > 1e-10 || math.Abs(y) > 1e-10 {
		t.Errorf("expected (25, 0), got (%v, %v)", x, y)
	}
}

func TestContext_Collinear(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(0, 0)
	b := ctx.Known(100, 100)
	p := ctx.Unknown()

	ctx.Collinear(p, a, b)
	ctx.EqX(p, 50) // Fix x to determine position on line

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := p.XY()
	if math.Abs(x-50) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (50, 50), got (%v, %v)", x, y)
	}
}

func TestContext_CollinearDiagonal(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(0, 0)
	b := ctx.Known(100, 50)
	p := ctx.Unknown()

	ctx.Collinear(p, a, b)
	ctx.EqX(p, 60)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := p.XY()
	if math.Abs(x-60) > 1e-10 || math.Abs(y-30) > 1e-10 {
		t.Errorf("expected (60, 30), got (%v, %v)", x, y)
	}
}

func TestContext_Intersection(t *testing.T) {
	ctx := NewContext()
	// Line 1: (0,0) to (100,100)
	a1 := ctx.Known(0, 0)
	a2 := ctx.Known(100, 100)
	// Line 2: (0,100) to (100,0)
	b1 := ctx.Known(0, 100)
	b2 := ctx.Known(100, 0)

	p, err := ctx.IntersectionOf(a1, a2, b1, b2)
	if err != nil {
		t.Fatalf("IntersectionOf failed: %v", err)
	}

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := p.XY()
	if math.Abs(x-50) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (50, 50), got (%v, %v)", x, y)
	}
}

func TestContext_Sum(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(10, 20)
	b := ctx.Known(30, 40)
	c := ctx.Unknown()
	ctx.Sum(c, a, b)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := c.XY()
	if math.Abs(x-40) > 1e-10 || math.Abs(y-60) > 1e-10 {
		t.Errorf("expected (40, 60), got (%v, %v)", x, y)
	}
}

func TestContext_Diff(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(50, 60)
	b := ctx.Known(20, 10)
	c := ctx.Unknown()
	ctx.Diff(c, a, b)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := c.XY()
	if math.Abs(x-30) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (30, 50), got (%v, %v)", x, y)
	}
}

func TestContext_Scaled(t *testing.T) {
	ctx := NewContext()
	v := ctx.Known(10, 20)
	s := ctx.Unknown()
	ctx.Scaled(s, v, 2.5)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := s.XY()
	if math.Abs(x-25) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (25, 50), got (%v, %v)", x, y)
	}
}

func TestContext_EqVar(t *testing.T) {
	ctx := NewContext()
	a := ctx.Known(30, 40)
	b := ctx.Unknown()
	ctx.EqVar(b, a)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := b.XY()
	if math.Abs(x-30) > 1e-10 || math.Abs(y-40) > 1e-10 {
		t.Errorf("expected (30, 40), got (%v, %v)", x, y)
	}
}

func TestContext_ComplexChain(t *testing.T) {
	// z0 = (0, 0)
	// z1 = (100, 0)
	// z2 = midpoint of z0 and z1 = (50, 0)
	// z3 = z2 + (0, 50) = (50, 50)
	ctx := NewContext()
	z0 := ctx.Known(0, 0)
	z1 := ctx.Known(100, 0)
	z2 := ctx.MidPointOf(z0, z1)
	z3 := ctx.Unknown()
	offset := ctx.Known(0, 50)
	ctx.Sum(z3, z2, offset)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := z3.XY()
	if math.Abs(x-50) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (50, 50), got (%v, %v)", x, y)
	}
}

func TestContext_VarPoint(t *testing.T) {
	ctx := NewContext()
	z := ctx.Known(10, 20)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	pt := z.Point()
	if pt.X != 10 || pt.Y != 20 {
		t.Errorf("expected Point{10, 20}, got %v", pt)
	}
}

func TestContext_SetXY(t *testing.T) {
	ctx := NewContext()
	z := ctx.Unknown()
	z.SetX(10)
	z.SetY(20)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := z.XY()
	if x != 10 || y != 20 {
		t.Errorf("expected (10, 20), got (%v, %v)", x, y)
	}
}

func TestContext_Eq(t *testing.T) {
	ctx := NewContext()
	z := ctx.Unknown()
	ctx.Eq(z, mp.P(30, 40))

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := z.XY()
	if math.Abs(x-30) > 1e-10 || math.Abs(y-40) > 1e-10 {
		t.Errorf("expected (30, 40), got (%v, %v)", x, y)
	}
}

func TestContext_TwoCollinearConstraints(t *testing.T) {
	// Find intersection of two lines using collinear constraints:
	// Line 1: (0,0) -- (100,100)
	// Line 2: (100,0) -- (0,100)
	// Intersection: (50, 50)
	ctx := NewContext()
	a1 := ctx.Known(0, 0)
	a2 := ctx.Known(100, 100)
	b1 := ctx.Known(100, 0)
	b2 := ctx.Known(0, 100)
	p := ctx.Unknown()

	ctx.Collinear(p, a1, a2)
	ctx.Collinear(p, b1, b2)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	x, y := p.XY()
	if math.Abs(x-50) > 1e-10 || math.Abs(y-50) > 1e-10 {
		t.Errorf("expected (50, 50), got (%v, %v)", x, y)
	}
}

func TestContext_PathBuilderIntegration(t *testing.T) {
	// Create a triangle where one vertex is computed as the midpoint
	// z0 = (0, 0), z1 = (100, 0), z2 = midpoint of z0 and z1 shifted up
	ctx := NewContext()
	z0 := ctx.Known(0, 0)
	z1 := ctx.Known(100, 0)
	z2 := ctx.MidPointOf(z0, z1)
	z3 := ctx.Unknown()
	offset := ctx.Known(0, 50)
	ctx.Sum(z3, z2, offset) // z3 = (50, 50)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	// Build a path using resolved variables
	engine := mp.NewEngine()
	path, err := ctx.NewPath().
		MoveToVar(z0).
		LineToVar(z1).
		LineToVar(z3).
		Close().
		SolveWithEngine(engine)

	if err != nil {
		t.Fatalf("SolveWithEngine failed: %v", err)
	}

	// Verify the path has 3 knots
	count := 0
	k := path.Head
	for {
		count++
		k = k.Next
		if k == path.Head {
			break
		}
	}
	if count != 3 {
		t.Errorf("expected 3 knots, got %d", count)
	}

	// Verify coordinates
	k = path.Head
	if math.Abs(float64(k.XCoord)-0) > 1e-10 || math.Abs(float64(k.YCoord)-0) > 1e-10 {
		t.Errorf("knot 0: expected (0, 0), got (%v, %v)", k.XCoord, k.YCoord)
	}
	k = k.Next
	if math.Abs(float64(k.XCoord)-100) > 1e-10 || math.Abs(float64(k.YCoord)-0) > 1e-10 {
		t.Errorf("knot 1: expected (100, 0), got (%v, %v)", k.XCoord, k.YCoord)
	}
	k = k.Next
	if math.Abs(float64(k.XCoord)-50) > 1e-10 || math.Abs(float64(k.YCoord)-50) > 1e-10 {
		t.Errorf("knot 2: expected (50, 50), got (%v, %v)", k.XCoord, k.YCoord)
	}
}

func TestContext_PathBuilderCurves(t *testing.T) {
	// Test CurveToVar with computed points
	ctx := NewContext()
	z0 := ctx.Known(0, 0)
	z1 := ctx.Known(100, 100)
	mid := ctx.MidPointOf(z0, z1) // (50, 50)

	if err := ctx.Solve(); err != nil {
		t.Fatalf("Solve failed: %v", err)
	}

	engine := mp.NewEngine()
	path, err := ctx.NewPath().
		MoveToVar(z0).
		CurveToVar(mid).
		CurveToVar(z1).
		SolveWithEngine(engine)

	if err != nil {
		t.Fatalf("SolveWithEngine failed: %v", err)
	}

	// Verify the middle knot is at (50, 50)
	k := path.Head.Next
	if math.Abs(float64(k.XCoord)-50) > 1e-10 || math.Abs(float64(k.YCoord)-50) > 1e-10 {
		t.Errorf("middle knot: expected (50, 50), got (%v, %v)", k.XCoord, k.YCoord)
	}
}
