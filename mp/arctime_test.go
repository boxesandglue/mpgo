package mp

import (
	"math"
	"testing"
)

func TestArcTime_MetaPostCurve(t *testing.T) {
	// Test with the curve (0,0)..(50,80)..(100,80)..(150,0)
	// MetaPost control points
	p := NewPath()

	k0 := NewKnot()
	k0.XCoord, k0.YCoord = 0, 0
	k0.LeftX, k0.LeftY = 0, 0
	k0.RightX, k0.RightY = -3.61343, 34.96872
	k0.LType = KnotEndpoint
	k0.RType = KnotExplicit
	p.Append(k0)

	k1 := NewKnot()
	k1.XCoord, k1.YCoord = 50, 80
	k1.LeftX, k1.LeftY = 16.98402, 67.92467
	k1.RightX, k1.RightY = 66.14372, 85.90445
	k1.LType = KnotExplicit
	k1.RType = KnotExplicit
	p.Append(k1)

	k2 := NewKnot()
	k2.XCoord, k2.YCoord = 100, 80
	k2.LeftX, k2.LeftY = 83.85628, 85.90445
	k2.RightX, k2.RightY = 133.01596, 67.92467
	k2.LType = KnotExplicit
	k2.RType = KnotExplicit
	p.Append(k2)

	k3 := NewKnot()
	k3.XCoord, k3.YCoord = 150, 0
	k3.LeftX, k3.LeftY = 153.61343, 34.96872
	k3.RightX, k3.RightY = 150, 0
	k3.LType = KnotExplicit
	k3.RType = KnotEndpoint
	p.Append(k3)

	tests := []struct {
		arcLen   Number
		expected Number // MetaPost values
	}{
		{0, 0},
		{50, 0.49202},
		{100, 0.98485},
		{127.11, 1.49991},
		{200, 2.46565},
		{254.22923, 3},
	}

	for _, tc := range tests {
		got := p.ArcTime(tc.arcLen)
		diff := math.Abs(float64(got - tc.expected))
		if diff > 0.01 { // Allow 1% tolerance
			t.Errorf("ArcTime(%.2f): got %.5f, want %.5f (diff: %.5f)",
				tc.arcLen, got, tc.expected, diff)
		} else {
			t.Logf("ArcTime(%.2f): Go=%.5f, MetaPost=%.5f ✓", tc.arcLen, got, tc.expected)
		}
	}
}

func TestArcTime_Line(t *testing.T) {
	// Straight line (0,0)--(100,0)
	// arctime should be linear: arctime x = x/100 * 1 = x/100
	line := NewPath()
	k0 := NewKnot()
	k0.XCoord, k0.YCoord = 0, 0
	k0.RightX, k0.RightY = 0, 0
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

	tests := []struct {
		arcLen   Number
		expected Number
	}{
		{0, 0},
		{25, 0.25},
		{50, 0.5},
		{75, 0.75},
		{100, 1},
	}

	for _, tc := range tests {
		got := line.ArcTime(tc.arcLen)
		diff := math.Abs(float64(got - tc.expected))
		if diff > 0.01 {
			t.Errorf("Line ArcTime(%.2f): got %.5f, want %.5f", tc.arcLen, got, tc.expected)
		} else {
			t.Logf("Line ArcTime(%.2f): %.5f ✓", tc.arcLen, got)
		}
	}
}
