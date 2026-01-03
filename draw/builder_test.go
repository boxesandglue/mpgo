package draw

import (
	"github.com/boxesandglue/mpgo/svg"
	"math"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestSolvePathProducesFiniteControls(t *testing.T) {
	r := 80.0
	path := NewPath().
		MoveTo(P(r, 0)).
		CurveTo(P(0, r)).
		CurveTo(P(-r, 0)).
		CurveTo(P(0, -r)).
		CurveTo(P(r, 0)).
		Close()

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("path has no knots")
	}

	isFinite := func(v float64) bool {
		return !math.IsNaN(v) && !math.IsInf(v, 0)
	}

	k := solved.Head
	i := 0
	for {
		if !isFinite(k.RightX) || !isFinite(k.RightY) || !isFinite(k.LeftX) || !isFinite(k.LeftY) {
			t.Fatalf("knot %d has non-finite control points: R(%.4f, %.4f) L(%.4f, %.4f)", i, k.RightX, k.RightY, k.LeftX, k.LeftY)
		}
		k = k.Next
		i++
		if k == solved.Head {
			break
		}
	}

	svg := svg.PathToSVG(solved)
	if svg == "" || strings.Contains(svg, "NaN") {
		t.Fatalf("SVG output invalid: %q", svg)
	}
}
