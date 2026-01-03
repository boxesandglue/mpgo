package mp

import (
	"math"
	"testing"
)

func TestArcLength_Line(t *testing.T) {
	// Straight line (0,0)--(100,0) should have arc length 100
	line := NewPath()
	k0 := NewKnot()
	k0.XCoord, k0.YCoord = 0, 0
	k0.RightX, k0.RightY = 0, 0 // For a line, controls = points
	k0.LeftX, k0.LeftY = 0, 0
	k0.LType = KnotEndpoint
	k0.RType = KnotExplicit
	line.Append(k0)

	k1 := NewKnot()
	k1.XCoord, k1.YCoord = 100, 0
	k1.LeftX, k1.LeftY = 100, 0
	k1.RightX, k1.RightY = 100, 0
	k1.LType = KnotExplicit
	k1.RType = KnotEndpoint
	line.Append(k1)

	arcLen := line.ArcLength()
	expected := Number(100)
	if math.Abs(float64(arcLen-expected)) > 0.001 {
		t.Errorf("Line arc length: got %v, want %v", arcLen, expected)
	}
}

func TestArcLength_Circle(t *testing.T) {
	// 4-point circle approximation with radius 50
	// Expected circumference: π * 100 ≈ 314.159
	kappa := Number(4.0 / 3.0 * (math.Sqrt(2) - 1))
	r := Number(50)

	circle := NewPath()
	points := [][2]Number{{r, 0}, {0, r}, {-r, 0}, {0, -r}}
	for i := 0; i < 4; i++ {
		k := NewKnot()
		k.XCoord = points[i][0]
		k.YCoord = points[i][1]

		// Tangent directions
		nextIdx := (i + 1) % 4
		prevIdx := (i + 3) % 4

		// Tangent going forward (perpendicular to radius, counterclockwise)
		// At (r,0): tangent is (0,1)
		// At (0,r): tangent is (-1,0)
		// At (-r,0): tangent is (0,-1)
		// At (0,-r): tangent is (1,0)
		tx := -points[i][1] / r // perpendicular to radius
		ty := points[i][0] / r

		k.RightX = k.XCoord + kappa*r*tx
		k.RightY = k.YCoord + kappa*r*ty
		k.LeftX = k.XCoord - kappa*r*tx
		k.LeftY = k.YCoord - kappa*r*ty

		k.LType = KnotExplicit
		k.RType = KnotExplicit
		circle.Append(k)
		_ = nextIdx
		_ = prevIdx
	}

	arcLen := circle.ArcLength()
	expected := Number(math.Pi * 100)
	// 4-segment approximation is about 0.027% too short
	if math.Abs(float64(arcLen-expected)) > 0.1 {
		t.Errorf("Circle arc length: got %v, want ~%v", arcLen, expected)
	}
	t.Logf("Circle (4 segments, r=50): %.5f (expected: %.5f)", arcLen, expected)
}

func TestArcLength_MetaPostCurve(t *testing.T) {
	// Test with the curve (0,0)..(50,80)..(100,80)..(150,0)
	// MetaPost reports arclength = 254.22923
	p := NewPath()

	// Knot 0: (0,0)
	k0 := NewKnot()
	k0.XCoord, k0.YCoord = 0, 0
	k0.LeftX, k0.LeftY = 0, 0
	k0.RightX, k0.RightY = -3.61343, 34.96872
	k0.LType = KnotEndpoint
	k0.RType = KnotExplicit
	p.Append(k0)

	// Knot 1: (50,80)
	k1 := NewKnot()
	k1.XCoord, k1.YCoord = 50, 80
	k1.LeftX, k1.LeftY = 16.98402, 67.92467
	k1.RightX, k1.RightY = 66.14372, 85.90445
	k1.LType = KnotExplicit
	k1.RType = KnotExplicit
	p.Append(k1)

	// Knot 2: (100,80)
	k2 := NewKnot()
	k2.XCoord, k2.YCoord = 100, 80
	k2.LeftX, k2.LeftY = 83.85628, 85.90445
	k2.RightX, k2.RightY = 133.01596, 67.92467
	k2.LType = KnotExplicit
	k2.RType = KnotExplicit
	p.Append(k2)

	// Knot 3: (150,0)
	k3 := NewKnot()
	k3.XCoord, k3.YCoord = 150, 0
	k3.LeftX, k3.LeftY = 153.61343, 34.96872
	k3.RightX, k3.RightY = 150, 0
	k3.LType = KnotExplicit
	k3.RType = KnotEndpoint
	p.Append(k3)

	arcLen := p.ArcLength()
	expected := Number(254.22923)
	if math.Abs(float64(arcLen-expected)) > 0.001 {
		t.Errorf("Curve arc length: got %.5f, want %.5f", arcLen, expected)
	}
	t.Logf("Curve (0,0)..(50,80)..(100,80)..(150,0): Go=%.5f, MetaPost=%.5f", arcLen, expected)
}
