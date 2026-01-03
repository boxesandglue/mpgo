package draw

import (
	"math"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

/*
prologues:=0;
path p;
r:= 80;
for a=0 upto 9:
    p := (0,0){dir 45}..{dir -10a}(6cm,0);
    show p;
endfor
end
*/
// Controls taken from MetaPost 2.02 "show p" for fan.mp (dir 45 out, dir -10*a in, length=6cm).
func TestFanControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	const length = 6 * 28.3464567 // same as cmd/fan
	expected := []seg{
		{44.36261, 44.36261, 110.4153, 0, length, 0},
		{43.43579, 43.43579, 109.59146, 10.6654, length, 0},
		{43.14322, 43.14322, 110.57832, 21.65662, length, 0},
		{43.78325, 43.78325, 113.80388, 32.49019, length, 0},
		{45.58353, 45.58353, 119.43259, 42.49783, length, 0},
		{48.68285, 48.68285, 127.31259, 50.96584, length, 0},
		{53.1185, 53.1185, 136.99841, 57.2969, length, 0},
		{58.8096, 58.8096, 147.82527, 61.14009, length, 0},
		{65.54742, 65.54742, 159.0586, 62.49892, length, 0},
		{72.98096, 72.98096, 170.0787, 61.77214, length, 0},
	}

	approx := func(a, b float64) bool {
		const tol = 1e-3
		return math.Abs(a-b) <= tol
	}

	for a, exp := range expected {
		p := NewPath().
			MoveTo(P(0, 0)).
			WithDirection(45.0).
			WithIncomingDirection(float64(-10 * a)).
			CurveTo(P(length, 0))

		solved, err := p.SolveWithEngine(mp.NewEngine())
		if err != nil {
			t.Fatalf("solve failed for a=%d: %v", a, err)
		}
		k := solved.Head
		if k == nil {
			t.Fatalf("a=%d: no knots", a)
		}
		q := k.Next
		if q == nil {
			t.Fatalf("a=%d: missing second knot", a)
		}

		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("a=%d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", a, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("a=%d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", a, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("a=%d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", a, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
	}
}

/*
prologues:=0;
path p;
p := (0,0)--(50,0){dir 90}..tension 2..(100,50)..
     controls (120,70) and (140,70)..
     (160,50)..(210,0)--(260,0);
show p;
end
*/

// Controls taken from MetaPost 2.02 "show p" output for:
// draw (0,0)--(50,0){dir 90}..tension 2..(100,50)..controls (120,70) and (140,70)..(160,50)..(210,0)--(260,0);
func TestConnectionShowcaseControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{16.666666666666668, 0, 33.33333333333333, 0, 50, 0},
		{50, 13.0417860789382, 91.23002937648121, 41.230029376481205, 100, 50},
		{120, 70, 140, 70, 160, 50},
		{176.66666666666666, 33.333333333333336, 193.33333333333334, 16.666666666666668, 210, 0},
		{226.66666666666666, 0, 243.33333333333334, 0, 260, 0},
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		LineTo(P(50, 0)).
		WithDirection(90).WithTension(2).CurveTo(P(100, 50)).
		CurveToWithControls(P(160, 50), P(120, 70), P(140, 70)).
		CurveTo(P(210, 0)).
		LineTo(P(260, 0))

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	approx := func(a, b float64) bool {
		const tol = 1e-6
		return math.Abs(a-b) <= tol
	}

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.10f,%.10f) want (%.10f,%.10f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.10f,%.10f) want (%.10f,%.10f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.10f,%.10f) want (%.10f,%.10f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
	if k.RType != mp.KnotEndpoint {
		t.Fatalf("expected to stop at endpoint after %d segments, found RType=%v", len(expected), k.RType)
	}
}

/*
prologues:=0;
path p;
r:= 80;
p:= (r,0)..(0,r)..(-r,0)..(0,-r)..(r,0)..cycle;
show p;
end
*/

// Controls taken from MetaPost 2.02 "show p" for:
// r:=80; draw (r,0)..(0,r)..(-r,0)..(0,-r)..(r,0)..cycle;
func TestCircleControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{80, 44.18279, 44.18279, 80, 0, 80},
		{-44.18279, 80, -80, 44.18279, -80, 0},
		{-80, -44.18279, -44.18279, -80, 0, -80},
		{44.18279, -80, 80, -44.18279, 80, 0},
		{80, 0, 80, 0, 80, 0}, // closing segment
	}

	const r = 80.0
	path := NewPath().
		MoveTo(P(r, 0)).
		CurveTo(P(0, r)).
		CurveTo(P(-r, 0)).
		CurveTo(P(0, -r)).
		CurveTo(P(r, 0)).
		Close()

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	approx := func(a, b float64) bool {
		const tol = 1e-4 // MetaPost output rounded to 5 decimals
		return math.Abs(a-b) <= tol
	}

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if q == nil {
			t.Fatalf("seg %d: next knot is nil", i)
		}
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
	if k != solved.Head {
		t.Fatalf("expected to return to head after %d segments; at (%.5f,%.5f)", len(expected), k.XCoord, k.YCoord)
	}
}

/*
prologues:=0;
path p;
z0 = (0,0); z1 = (60,40); z2 = (40,90); z3 = (10,70); z4 = (30,50);
draw z0..z1..z2..z3..z4;
show p;
end
*/

// mpcurve: z0..z1..z2..z3..z4; controls from MetaPost 2.02 "show p".
func TestMpCurveControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{26.76463, -1.84543, 51.4094, 14.58441, 60, 40},
		{67.09875, 61.00188, 59.76253, 84.57518, 40, 90},
		{25.35715, 94.01947, 10.48064, 84.5022, 10, 70},
		{9.62895, 58.80421, 18.80421, 49.62895, 30, 50},
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		CurveTo(P(60, 40)).
		CurveTo(P(40, 90)).
		CurveTo(P(10, 70)).
		CurveTo(P(30, 50))

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	const tol = 1e-4
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
}

/*
prologues:=0;
path p;
z0 = (0,0); z1 = (60,40); z2 = (40,90); z3 = (10,70); z4 = (30,50);
p := z0..z1..z2..z3..z4..cycle;
show p;
end
*/

// mpcurve cyclic: z0..z1..z2..z3..z4..cycle with MetaPost 2.02 control points.
func TestMpCurveCycleControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{5.18756, -26.8353, 60.36073, -18.40036, 60, 40},
		{59.87714, 59.889, 57.33896, 81.64203, 40, 90},
		{22.39987, 98.48387, 4.72404, 84.46368, 10, 70},
		{13.38637, 60.7165, 26.35591, 59.1351, 30, 50},
		{39.19409, 26.95198, -4.10555, 21.23804, 0, 0}, // closing segment to start
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		CurveTo(P(60, 40)).
		CurveTo(P(40, 90)).
		CurveTo(P(10, 70)).
		CurveTo(P(30, 50)).
		Close()

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	const tol = 1e-4
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
		if k == solved.Head {
			break
		}
	}
}

/*
path p;
z0 = (0,0); z1 = (60,10); z2 = (120,0);
p := z0{up}..z1{right}..z2{down};
show p;
end
*/

// direction demo: (0,0){up}..(60,10){right}..(120,0){down}
// Controls from MetaPost 2.02 "show p".
func TestDirectionBoundaryControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{0, 25.83095, 34.33913, 10, 60, 10},
		{85.66087, 10, 120, 25.83095, 120, 0},
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		WithDirection(90). // z0{up} (outgoing)
		CurveTo(P(60, 10)).
		WithDirection(0).   // z1{right} (outgoing)
		CurveTo(P(120, 0)). // z2 endpoint, outgoing direction stored
		WithDirection(270)

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	const tol = 1e-4
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
	if k.RType != mp.KnotEndpoint {
		t.Fatalf("expected to end at endpoint, got RType=%v", k.RType)
	}
}

/*
path p;
z0 = (0,0); z1 = (60,10); z2 = (120,0);
p := z0{up}...z1{right}...z2{down};
show p;
end
*/

// direction demo with "..." (tension atleast 1):
// draw z0{up}...z1{right}...z2{down}
// MetaPost 2.02 "show p" control points.
func TestDirectionAtLeastControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{0, 9.99756, 34.33913, 10, 60, 10},
		{85.66087, 10, 120, 9.99756, 120, 0},
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		WithDirection(90).     // z0{up}
		WithTensionAtLeast(1). // "..."
		CurveTo(P(60, 10)).
		WithDirection(0).      // z1{right}
		WithTensionAtLeast(1). // "..."
		CurveTo(P(120, 0)).
		WithDirection(270) // z2{down}

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	const tol = 3e-3
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
	if k.RType != mp.KnotEndpoint {
		t.Fatalf("expected endpoint, got RType=%v", k.RType)
	}
}

// direction_tension demo:
// draw z0..z1..tension 1.5 and 1..z2..z3;
// Controls from MetaPost 2.02 "show p".
func TestDirectionMixedTensionControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	expected := []seg{
		{2.09846, 12.3886, 9.37148, 23.29811, 20, 30},
		{40.56642, 42.96829, 90.2319, 50.68138, 120, 30},
		{130.2267, 22.895, 137.3749, 12.17271, 140, 0},
	}

	path := NewPath().
		MoveTo(P(0, 0)).
		CurveTo(P(20, 30)).
		WithOutgoingTension(1.5). // between z1(out) and z2(in)
		WithIncomingTension(1).
		CurveTo(P(120, 30)).
		CurveTo(P(140, 0))

	solved, err := path.SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}
	if solved.Head == nil {
		t.Fatalf("no knots")
	}

	const tol = 1e-4
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	k := solved.Head
	for i, exp := range expected {
		q := k.Next
		if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
			t.Fatalf("seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, k.RightX, k.RightY, exp.c1x, exp.c1y)
		}
		if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
			t.Fatalf("seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
		}
		if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
			t.Fatalf("seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, q.XCoord, q.YCoord, exp.ex, exp.ey)
		}
		k = q
	}
	if k.RType != mp.KnotEndpoint {
		t.Fatalf("expected endpoint, got RType=%v", k.RType)
	}
}

// curl demo: draw z0{curl c}..z1..{curl c}z2 for c in {0,1,2,inf}.
func TestCurlControlsMatchMetaPost(t *testing.T) {
	type seg struct {
		c1x, c1y float64
		c2x, c2y float64
		ex, ey   float64
	}
	type sample struct {
		curl float64
		exp  []seg
	}
	z0 := P(10, 0)
	z1 := P(0, 60)
	z2 := P(10, 120)

	samples := []sample{
		{
			curl: 0,
			exp: []seg{
				{5.00978, 19.73059, 0, 39.62689, 0, 60},
				{0, 80.37311, 5.00978, 100.26941, 10, 120},
			},
		},
		{
			curl: 1,
			exp: []seg{
				{3.379, 19.31125, 0, 39.58524, 0, 60},
				{0, 80.41476, 3.379, 100.68875, 10, 120},
			},
		},
		{
			curl: 2,
			exp: []seg{
				{2.5711, 19.06372, 0, 39.552, 0, 60},
				{0, 80.448, 2.5711, 100.93628, 10, 120},
			},
		},
		{
			curl: mp.Inf(),
			exp: []seg{
				{0.18536, 18.16626, -0.00015, 39.39874, 0, 60},
				{0.00015, 80.6012, 0.18594, 101.83351, 10, 120},
			},
		},
	}

	const tol = 3e-3
	approx := func(a, b float64) bool { return math.Abs(a-b) <= tol }

	for i, s := range samples {
		curl := s.curl
		if math.IsInf(curl, 1) || curl == mp.Inf() {
			curl = 1e9 // avoid overflow; matches cmd/curl approximation for curl infinity
		}
		p := NewPath().
			MoveTo(z0).
			WithOutgoingCurl(curl).
			CurveTo(z1).
			WithIncomingCurl(curl).
			CurveTo(z2)

		solved, err := p.SolveWithEngine(mp.NewEngine())
		if err != nil {
			t.Fatalf("curl idx %d solve failed: %v", i, err)
		}
		if solved.Head == nil {
			t.Fatalf("curl idx %d: no knots", i)
		}

		k := solved.Head
		for segIdx, exp := range s.exp {
			q := k.Next
			if !approx(k.RightX, exp.c1x) || !approx(k.RightY, exp.c1y) {
				t.Fatalf("curl idx %d seg %d: c1 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, segIdx, k.RightX, k.RightY, exp.c1x, exp.c1y)
			}
			if !approx(q.LeftX, exp.c2x) || !approx(q.LeftY, exp.c2y) {
				t.Fatalf("curl idx %d seg %d: c2 mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, segIdx, q.LeftX, q.LeftY, exp.c2x, exp.c2y)
			}
			if !approx(q.XCoord, exp.ex) || !approx(q.YCoord, exp.ey) {
				t.Fatalf("curl idx %d seg %d: end mismatch got (%.5f,%.5f) want (%.5f,%.5f)", i, segIdx, q.XCoord, q.YCoord, exp.ex, exp.ey)
			}
			k = q
		}
		if k.RType != mp.KnotEndpoint {
			t.Fatalf("curl idx %d: expected endpoint, got RType=%v", i, k.RType)
		}
	}
}
