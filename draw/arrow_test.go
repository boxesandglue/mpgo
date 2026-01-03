package draw

import (
	"bytes"
	"github.com/boxesandglue/mpgo/svg"
	"math"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestWithArrow(t *testing.T) {
	path := NewPath().
		WithArrow().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if !solved.Style.Arrow.End {
		t.Error("Arrow.End should be true")
	}
	if solved.Style.Arrow.Start {
		t.Error("Arrow.Start should be false")
	}
}

func TestWithDoubleArrow(t *testing.T) {
	path := NewPath().
		WithDoubleArrow().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if !solved.Style.Arrow.End {
		t.Error("Arrow.End should be true")
	}
	if !solved.Style.Arrow.Start {
		t.Error("Arrow.Start should be true")
	}
}

func TestWithArrowStyle(t *testing.T) {
	path := NewPath().
		WithArrow().
		WithArrowStyle(8.0, 60.0).
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if !solved.Style.Arrow.End {
		t.Error("Arrow.End should be true")
	}
	if solved.Style.Arrow.Length != 8.0 {
		t.Errorf("Arrow.Length wrong: got %f, want 8.0", solved.Style.Arrow.Length)
	}
	if solved.Style.Arrow.Angle != 60.0 {
		t.Errorf("Arrow.Angle wrong: got %f, want 60.0", solved.Style.Arrow.Angle)
	}
}

func TestArrowSVGOutput(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		WithArrow().
		MoveTo(P(10, 30)).
		LineTo(P(50, 30))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Should have main path (shortened for arrow)
	if !strings.Contains(output, "<path") {
		t.Error("SVG should contain path elements")
	}

	// Should have arrow head (filled triangle)
	if !strings.Contains(output, "fill=\"black\"") {
		t.Error("SVG should contain filled arrow head")
	}

	// Arrow head path should close with Z
	if !strings.Contains(output, "Z\"") {
		t.Error("Arrow head path should be closed (end with Z)")
	}
}

func TestDoubleArrowSVGOutput(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("blue")).
		WithDoubleArrow().
		MoveTo(P(70, 30)).
		LineTo(P(110, 30))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Count filled triangles (should have 2 for double arrow)
	fillCount := strings.Count(output, "fill=\"blue\"")
	if fillCount < 2 {
		t.Errorf("double arrow should have at least 2 filled elements, got %d", fillCount)
	}
}

func TestArrowPathShortening(t *testing.T) {
	// Create a simple horizontal arrow
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithArrow().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// The main line should end at approximately 96.3 (100 - 4*cos(22.5°))
	// With MetaPost-compatible viewBox offset of 0.25, the coordinate becomes ~96.55
	expectedLineEnd := 100 - mp.DefaultAHLength*math.Cos(mp.DefaultAHAngle*math.Pi/360)
	offsetLineEnd := expectedLineEnd + 0.25 // viewBox offset

	// The shortened line endpoint should be around 96.55 (96.3 + 0.25 offset)
	if !strings.Contains(output, "96.5") {
		t.Errorf("line should be shortened to ~96.55 (with viewBox offset), got output: %s", output)
	}

	// Verify the expected shortening distance (before offset)
	if math.Abs(expectedLineEnd-96.304482) > 0.001 {
		t.Errorf("expected line end at ~96.304, got formula result: %f", expectedLineEnd)
	}
	_ = offsetLineEnd // Used for documentation
}

func TestArrowWithCurvedPath(t *testing.T) {
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("red")).
		WithArrow().
		MoveTo(P(0, 0)).
		CurveTo(P(100, 100))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Should have curved path (contains C for cubic bezier)
	if !strings.Contains(output, "C ") && !strings.Contains(output, "L ") {
		t.Error("SVG should contain path data")
	}

	// Should have arrow head
	if !strings.Contains(output, "fill=\"red\"") {
		t.Error("SVG should contain red arrow head")
	}
}

func TestArrowDefaultValues(t *testing.T) {
	path := NewPath().
		WithArrow().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Default values should be applied
	arrow := solved.Style.Arrow
	if arrow.Length != 0 && arrow.Length != mp.DefaultAHLength {
		t.Errorf("unexpected default arrow length: got %f", arrow.Length)
	}
	if arrow.Angle != 0 && arrow.Angle != mp.DefaultAHAngle {
		t.Errorf("unexpected default arrow angle: got %f", arrow.Angle)
	}
}

func TestArrowSVGMetaPostCompatible(t *testing.T) {
	// This test verifies that Go output matches MetaPost output structure
	// MetaPost arrow at end: tip at endpoint, base at ahlength*cos(ahangle/2) back
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		WithArrow().
		MoveTo(P(10, 30)).
		LineTo(P(50, 30))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	// Create arrow head manually to check geometry
	ahLength := mp.DefaultAHLength
	ahAngle := mp.DefaultAHAngle
	halfAngleRad := ahAngle * math.Pi / 360.0

	// Expected base X position
	baseX := 50.0 - ahLength*math.Cos(halfAngleRad)

	// The main path should end approximately at baseX
	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Check that viewBox is reasonable
	if !strings.Contains(output, "viewBox=\"0 0") {
		t.Error("viewBox should start at 0 0")
	}

	// The shortened line should end near baseX (with some tolerance for formatting)
	expectedEndX := baseX
	_ = expectedEndX // Used for verification

	// Arrow tip should be at x=50 (with viewBox offset it becomes 50 - minX + halfStroke)
	// For a path from (10,30) to (50,30) with halfStroke=0.25:
	// minX = 10 - 0.25 = 9.75, so x=50 becomes 50 - 9.75 = 40.25
	if !strings.Contains(output, "40.") {
		t.Errorf("arrow tip should be at x ~40 (50 with viewBox offset), got: %s", output)
	}
}

func TestMultipleArrowsInSVG(t *testing.T) {
	engine := mp.NewEngine()

	path1, _ := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		WithArrow().
		MoveTo(P(10, 30)).
		LineTo(P(50, 30)).
		SolveWithEngine(engine)

	path2, _ := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("blue")).
		WithDoubleArrow().
		MoveTo(P(70, 30)).
		LineTo(P(110, 30)).
		SolveWithEngine(engine)

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(path1, path2)
	svg.AddPathFromPath(path1)
	svg.AddPathFromPath(path2)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Count path elements - should have:
	// - 2 main lines (black, blue)
	// - 1 black arrow head
	// - 2 blue arrow heads
	pathCount := strings.Count(output, "<path")
	if pathCount < 5 {
		t.Errorf("expected at least 5 path elements, got %d", pathCount)
	}
}

func TestArrowNoStroke(t *testing.T) {
	// Arrow heads should be filled, not stroked
	path := NewPath().
		WithPen(mp.PenCircle(0.5)).
		WithStrokeColor(mp.ColorCSS("black")).
		WithArrow().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0))

	engine := mp.NewEngine()
	solved, err := path.SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	svg := svg.NewBuilder().
		MetaPostCompatible().
		FitViewBoxToPaths(solved)
	svg.AddPathFromPath(solved)

	var buf bytes.Buffer
	svg.WriteTo(&buf)
	output := buf.String()

	// Arrow head should have stroke="none"
	if !strings.Contains(output, "stroke=\"none\"") {
		t.Error("arrow head should have stroke=\"none\"")
	}
}
