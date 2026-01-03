package draw

import (
	"bytes"
	"github.com/boxesandglue/mpgo/svg"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestDashedEvenly(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		DashedEvenly().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if solved.Style.Dash == nil {
		t.Fatal("Dash should not be nil")
	}
	if len(solved.Style.Dash.Array) != 2 {
		t.Errorf("expected 2 elements in dash array, got %d", len(solved.Style.Dash.Array))
	}
	if solved.Style.Dash.Array[0] != 3 || solved.Style.Dash.Array[1] != 3 {
		t.Errorf("expected [3, 3], got %v", solved.Style.Dash.Array)
	}
}

func TestDashedWithDots(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		DashedWithDots().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if solved.Style.Dash == nil {
		t.Fatal("Dash should not be nil")
	}
	// withdots: on 0, off 5, with offset 2.5
	if solved.Style.Dash.Array[0] != 0 {
		t.Errorf("expected on=0 for dots, got %f", solved.Style.Dash.Array[0])
	}
	if solved.Style.Dash.Offset != 2.5 {
		t.Errorf("expected offset=2.5 for withdots, got %f", solved.Style.Dash.Offset)
	}
}

func TestCustomDashPattern(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		Dashed(6, 3, 2, 3). // on 6 off 3 on 2 off 3
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if solved.Style.Dash == nil {
		t.Fatal("Dash should not be nil")
	}
	expected := []float64{6, 3, 2, 3}
	if len(solved.Style.Dash.Array) != len(expected) {
		t.Errorf("expected %d elements, got %d", len(expected), len(solved.Style.Dash.Array))
	}
	for i, v := range expected {
		if solved.Style.Dash.Array[i] != v {
			t.Errorf("element %d: expected %f, got %f", i, v, solved.Style.Dash.Array[i])
		}
	}
}

func TestDashSVGOutput(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		DashedEvenly().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Should have stroke-dasharray attribute
	if !strings.Contains(output, `stroke-dasharray="3.00 3.00"`) {
		t.Errorf("SVG should contain stroke-dasharray, got: %s", output)
	}
}

func TestDashWithDotsSVGOutput(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		DashedWithDots().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Should have stroke-dasharray and stroke-dashoffset attributes
	if !strings.Contains(output, `stroke-dasharray="0.00 5.00"`) {
		t.Errorf("SVG should contain stroke-dasharray for dots, got: %s", output)
	}
	if !strings.Contains(output, `stroke-dashoffset="2.50"`) {
		t.Errorf("SVG should contain stroke-dashoffset for dots, got: %s", output)
	}
}

func TestDashPatternScaled(t *testing.T) {
	evenly := mp.DashEvenly()
	scaled := evenly.Scaled(2)

	if scaled.Array[0] != 6 || scaled.Array[1] != 6 {
		t.Errorf("scaled pattern should be [6, 6], got %v", scaled.Array)
	}

	// Original should be unchanged
	if evenly.Array[0] != 3 || evenly.Array[1] != 3 {
		t.Errorf("original pattern should be unchanged, got %v", evenly.Array)
	}
}

func TestDashPatternShifted(t *testing.T) {
	evenly := mp.DashEvenly()
	shifted := evenly.Shifted(1.5)

	if shifted.Offset != 1.5 {
		t.Errorf("shifted pattern offset should be 1.5, got %f", shifted.Offset)
	}

	// Original should be unchanged
	if evenly.Offset != 0 {
		t.Errorf("original pattern offset should be 0, got %f", evenly.Offset)
	}
}

func TestWithDashPattern(t *testing.T) {
	customPattern := mp.DashEvenly().Scaled(2)

	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		WithDashPattern(customPattern).
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if solved.Style.Dash == nil {
		t.Fatal("Dash should not be nil")
	}
	if solved.Style.Dash.Array[0] != 6 {
		t.Errorf("expected scaled dash of 6, got %f", solved.Style.Dash.Array[0])
	}
}

func TestNoDashPattern(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Should NOT have stroke-dasharray attribute
	if strings.Contains(output, "stroke-dasharray") {
		t.Error("SVG should not contain stroke-dasharray for solid lines")
	}
}

func TestDashPatternNil(t *testing.T) {
	// Test that nil patterns don't cause panic
	var d *mp.DashPattern = nil
	scaled := d.Scaled(2)
	if scaled != nil {
		t.Error("scaling nil dash should return nil")
	}
	shifted := d.Shifted(1)
	if shifted != nil {
		t.Error("shifting nil dash should return nil")
	}
}

func TestNewDashPatternEmpty(t *testing.T) {
	d := mp.NewDashPattern()
	if d != nil {
		t.Error("empty dash pattern should return nil")
	}
}

func TestFormatDashAttrsNil(t *testing.T) {
	result := svg.FormatDashAttrs(nil)
	if result != "" {
		t.Errorf("svg.FormatDashAttrs(nil) should return empty string, got %q", result)
	}
}

func TestFormatDashAttrsEmpty(t *testing.T) {
	d := &mp.DashPattern{Array: []float64{}}
	result := svg.FormatDashAttrs(d)
	if result != "" {
		t.Errorf("svg.FormatDashAttrs with empty array should return empty string, got %q", result)
	}
}

func TestFormatDashAttrsWithOffset(t *testing.T) {
	d := &mp.DashPattern{Array: []float64{3, 3}, Offset: 1.5}
	result := svg.FormatDashAttrs(d)
	if !strings.Contains(result, `stroke-dasharray="3.00 3.00"`) {
		t.Errorf("should contain dasharray, got %q", result)
	}
	if !strings.Contains(result, `stroke-dashoffset="1.50"`) {
		t.Errorf("should contain dashoffset, got %q", result)
	}
}
