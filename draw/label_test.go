package draw

import (
	"bytes"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
	"github.com/boxesandglue/mpgo/svg"
)

func TestLabelBasic(t *testing.T) {
	pic := NewPicture()
	pic.Label("A", mp.P(0, 0), mp.AnchorCenter)
	pic.Label("B", mp.P(100, 0), mp.AnchorRight)
	pic.Label("C", mp.P(50, 50), mp.AnchorTop)

	labels := pic.Labels()
	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}

	// Check first label
	if labels[0].Text != "A" {
		t.Errorf("expected text 'A', got '%s'", labels[0].Text)
	}
	if labels[0].Anchor != mp.AnchorCenter {
		t.Errorf("expected AnchorCenter, got %d", labels[0].Anchor)
	}
}

func TestLabelWithStyle(t *testing.T) {
	pic := NewPicture()
	label := pic.LabelWithStyle("styled", mp.P(10, 20), mp.AnchorTop)
	label.WithColor(mp.ColorRGB(1, 0, 0)).WithFontSize(14)

	labels := pic.Labels()
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}

	if labels[0].FontSize != 14 {
		t.Errorf("expected font size 14, got %f", labels[0].FontSize)
	}
	if labels[0].Color.CSS() != "rgb(255,0,0)" {
		t.Errorf("expected red color, got %s", labels[0].Color.CSS())
	}
}

func TestDotLabel(t *testing.T) {
	pic := NewPicture()
	pic.DotLabel("z0", mp.P(0, 0), mp.AnchorLowerRight, mp.ColorCSS("black"))

	labels := pic.Labels()
	paths := pic.Paths()

	if len(labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(labels))
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path (dot), got %d", len(paths))
	}

	// Check the dot is a filled circle
	dot := paths[0]
	if dot.Style.Fill.CSS() != "black" {
		t.Errorf("expected black fill, got %s", dot.Style.Fill.CSS())
	}
	if dot.Style.Stroke.CSS() != "none" {
		t.Errorf("expected no stroke, got %s", dot.Style.Stroke.CSS())
	}
}

func TestLabelOffsetVector(t *testing.T) {
	tests := []struct {
		anchor mp.Anchor
		dx, dy float64
	}{
		{mp.AnchorCenter, 0, 0},
		{mp.AnchorLeft, -1, 0},
		{mp.AnchorRight, 1, 0},
		{mp.AnchorTop, 0, 1},
		{mp.AnchorBottom, 0, -1},
		{mp.AnchorUpperLeft, -0.7, 0.7},
		{mp.AnchorUpperRight, 0.7, 0.7},
		{mp.AnchorLowerLeft, -0.7, -0.7},
		{mp.AnchorLowerRight, 0.7, -0.7},
	}

	for _, tc := range tests {
		dx, dy := mp.LabelOffsetVector(tc.anchor)
		if dx != tc.dx || dy != tc.dy {
			t.Errorf("LabelOffsetVector(%d): expected (%.1f, %.1f), got (%.1f, %.1f)",
				tc.anchor, tc.dx, tc.dy, dx, dy)
		}
	}
}

func TestLabelAnchorFactors(t *testing.T) {
	// Test that anchor factors match MetaPost's labxf/labyf from plain.mp
	tests := []struct {
		anchor mp.Anchor
		xf, yf float64
	}{
		{mp.AnchorCenter, 0.5, 0.5},
		{mp.AnchorLeft, 1, 0.5},     // labxf.lft=1, labyf.lft=.5
		{mp.AnchorRight, 0, 0.5},    // labxf.rt=0, labyf.rt=.5
		{mp.AnchorTop, 0.5, 0},      // labxf.top=.5, labyf.top=0
		{mp.AnchorBottom, 0.5, 1},   // labxf.bot=.5, labyf.bot=1
		{mp.AnchorUpperLeft, 1, 0},  // labxf.ulft=1, labyf.ulft=0
		{mp.AnchorUpperRight, 0, 0}, // labxf.urt=0, labyf.urt=0
		{mp.AnchorLowerLeft, 1, 1},  // labxf.llft=1, labyf.llft=1
		{mp.AnchorLowerRight, 0, 1}, // labxf.lrt=0, labyf.lrt=1
	}

	for _, tc := range tests {
		xf, yf := mp.LabelAnchorFactors(tc.anchor)
		if xf != tc.xf || yf != tc.yf {
			t.Errorf("LabelAnchorFactors(%d): expected (%.1f, %.1f), got (%.1f, %.1f)",
				tc.anchor, tc.xf, tc.yf, xf, yf)
		}
	}
}

func TestLabelInSVG(t *testing.T) {
	// Create a simple picture with a path and label
	pic := NewPicture()

	// Add a simple path
	path, _ := NewPath().
		MoveTo(P(0, 0)).
		LineTo(P(100, 0)).
		WithStrokeColor(mp.ColorCSS("black")).
		Solve()
	pic.AddPath(path)

	// Add a label
	pic.Label("Test Label", mp.P(50, 10), mp.AnchorBottom)

	// Render to SVG
	var buf bytes.Buffer
	builder := svg.NewBuilder()
	builder.AddPicture(pic)
	if err := builder.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	svgOutput := buf.String()

	// Check that the SVG contains a text element
	if !strings.Contains(svgOutput, "<text") {
		t.Error("SVG output should contain a <text> element")
	}
	if !strings.Contains(svgOutput, "Test Label") {
		t.Error("SVG output should contain the label text")
	}
	if !strings.Contains(svgOutput, "text-anchor=") {
		t.Error("SVG output should contain text-anchor attribute")
	}
}

func TestLabelXMLEscape(t *testing.T) {
	pic := NewPicture()
	pic.Label("<test> & \"quote\"", mp.P(0, 0), mp.AnchorCenter)

	path, _ := NewPath().
		MoveTo(P(0, 0)).
		LineTo(P(10, 0)).
		Solve()
	pic.AddPath(path)

	var buf bytes.Buffer
	builder := svg.NewBuilder()
	builder.AddPicture(pic)
	if err := builder.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	svgOutput := buf.String()

	// Check that special characters are escaped
	if strings.Contains(svgOutput, "<test>") {
		t.Error("SVG should escape < and > in label text")
	}
	if !strings.Contains(svgOutput, "&lt;test&gt;") {
		t.Error("SVG should contain escaped <test>")
	}
	if !strings.Contains(svgOutput, "&amp;") {
		t.Error("SVG should contain escaped &")
	}
}
