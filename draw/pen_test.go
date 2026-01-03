package draw

import (
	"github.com/boxesandglue/mpgo/svg"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

// Basic sanity: attaching a pen stores it in the resulting path style.
func TestWithPenSetsStyle(t *testing.T) {
	pen := &mp.Pen{Head: mp.NewKnot()}
	path := NewPath().
		MoveTo(P(0, 0)).
		WithPen(pen).
		CurveTo(P(10, 0))
	if _, err := path.SolveWithEngine(mp.NewEngine()); err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if pathBuilt := path.BuildPath(); pathBuilt.Style.Pen != pen {
		t.Fatalf("expected pen to be attached to style")
	}
}

func TestPenCircleSetsStrokeWidthInSVG(t *testing.T) {
	pen := mp.PenCircle(10)
	path := NewPath().
		MoveTo(P(0, 0)).
		WithPen(pen).
		CurveTo(P(10, 0))
	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	var b strings.Builder
	svg := svg.NewBuilder().AddPathFromPath(solved)
	if err := svg.WriteTo(&b); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `stroke-width="10.00"`) {
		t.Fatalf("expected stroke-width from pen circle, got: %s", out)
	}
}

// TestPenCircleMatchesMetaPostFormat verifies that PenCircle stores
// the transformation matrix in MetaPost's format (mp.w:10440-10452).
func TestPenCircleMatchesMetaPostFormat(t *testing.T) {
	pen := mp.PenCircle(5)

	// Verify it's elliptical (single knot pointing to itself)
	if !pen.Elliptical {
		t.Fatal("expected elliptical pen")
	}
	if pen.Head.Next != pen.Head {
		t.Fatal("expected single knot (Next == self)")
	}
	if pen.Head.Prev != pen.Head {
		t.Fatal("expected single knot (Prev == self)")
	}

	// mp.w:10448-10453: For pencircle with diameter d:
	// x_coord = 0, y_coord = 0 (center)
	// left_x = d, left_y = 0 (where (1,0) maps to)
	// right_x = 0, right_y = d (where (0,1) maps to)
	h := pen.Head
	if h.XCoord != 0 || h.YCoord != 0 {
		t.Errorf("expected center (0,0), got (%v,%v)", h.XCoord, h.YCoord)
	}
	if h.LeftX != 5 || h.LeftY != 0 {
		t.Errorf("expected left=(5,0), got (%v,%v)", h.LeftX, h.LeftY)
	}
	if h.RightX != 0 || h.RightY != 5 {
		t.Errorf("expected right=(0,5), got (%v,%v)", h.RightX, h.RightY)
	}

	// Verify GetPenScale returns the diameter
	scale := mp.GetPenScale(pen)
	if scale != 5 {
		t.Errorf("expected GetPenScale=5, got %v", scale)
	}
}

// TestPenSquareEnvelopeRegression is a regression test for pensquare envelope calculation.
// Reference: mp/line.mp draws (0,0)--(100,0) with pensquare scaled 4.
//
// With rounded linecap (Go and MetaPost default):
// - MetaPost produces: (-2,2)->(-2,-2)->(2,-2)->(102,-2)->(102,2)->(98,2)->cycle (6 vertices)
// - Go produces: same 6 vertices ✓ MATCH
//
// Note: Go now defaults to rounded linecap (mp.LineCapDefault) to match MetaPost.
// Use mp.LineCapButt for butt caps (4 vertices).
func TestPenSquareEnvelopeRegression(t *testing.T) {
	pen := mp.PenSquare(4)
	path := NewPath().
		WithPen(pen).
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Check that an envelope was computed
	if solved.Envelope == nil {
		t.Fatal("expected envelope to be computed for pensquare")
	}

	// Extract envelope vertices
	env := solved.Envelope
	var vertices [][2]float64
	k := env.Head
	for {
		vertices = append(vertices, [2]float64{k.XCoord, k.YCoord})
		k = k.Next
		if k == env.Head || k == nil {
			break
		}
	}

	// Expected output with rounded linecap (Go and MetaPost default):
	// MetaPost produces these 6 vertices.
	expected := [][2]float64{
		{-2, 2}, {-2, -2}, {2, -2}, {102, -2}, {102, 2}, {98, 2},
	}

	if len(vertices) != len(expected) {
		t.Errorf("expected %d vertices, got %d: %v", len(expected), len(vertices), vertices)
	}

	// Check that all expected vertices are present (order may vary)
	for _, exp := range expected {
		found := false
		for _, v := range vertices {
			if abs(v[0]-exp[0]) < 0.01 && abs(v[1]-exp[1]) < 0.01 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected vertex (%.1f, %.1f) not found in envelope: %v", exp[0], exp[1], vertices)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestMakePenConvexHull(t *testing.T) {
	// points defining a triangle plus an interior point; hull should be triangle, non-elliptical.
	p := mp.NewPath()
	p.Append(&mp.Knot{XCoord: 0, YCoord: 0})
	p.Append(&mp.Knot{XCoord: 10, YCoord: 0})
	p.Append(&mp.Knot{XCoord: 5, YCoord: 10})
	p.Append(&mp.Knot{XCoord: 5, YCoord: 5})
	pen := mp.MakePen(p)
	if pen.Elliptical {
		t.Fatalf("expected non-elliptical pen")
	}
	// count knots
	k := pen.Head
	count := 0
	for {
		count++
		k = k.Next
		if k == pen.Head {
			break
		}
		if count > 4 {
			break
		}
	}
	if count != 3 {
		t.Fatalf("expected hull with 3 points, got %d", count)
	}
}

func TestPenSpeck(t *testing.T) {
	pen := mp.PenSpeck()
	if pen.Elliptical {
		t.Fatal("expected non-elliptical pen")
	}
	if pen.Head == nil {
		t.Fatal("expected pen head")
	}

	// PenSpeck should be a 4-vertex square with half-width = eps/2
	halfEps := mp.Eps / 2
	expected := [][2]float64{
		{-halfEps, -halfEps},
		{halfEps, -halfEps},
		{halfEps, halfEps},
		{-halfEps, halfEps},
	}

	k := pen.Head
	for i := 0; i < 4; i++ {
		if abs(k.XCoord-expected[i][0]) > 1e-10 || abs(k.YCoord-expected[i][1]) > 1e-10 {
			t.Errorf("vertex %d: expected (%.6f, %.6f), got (%.6f, %.6f)",
				i, expected[i][0], expected[i][1], k.XCoord, k.YCoord)
		}
		k = k.Next
	}
}
