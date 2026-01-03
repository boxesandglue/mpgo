package draw

import (
	"github.com/boxesandglue/mpgo/svg"
	"strings"
	"testing"

	"github.com/boxesandglue/mpgo/mp"
)

func TestPictureCollectsPaths(t *testing.T) {
	engine := mp.NewEngine()
	p1, err := NewPath().MoveTo(P(0, 0)).CurveTo(P(10, 0)).SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve p1: %v", err)
	}
	p2, err := NewPath().MoveTo(P(20, 0)).CurveTo(P(30, 10)).SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve p2: %v", err)
	}

	picA := NewPicture().AddPath(p1)
	picB := NewPicture().AddPath(p2)
	pic := NewPicture().AddPicture(picA).AddPicture(picB)

	if got := len(pic.Paths()); got != 2 {
		t.Fatalf("expected 2 paths in picture, got %d", got)
	}
	if pic.Paths()[0] != p1 || pic.Paths()[1] != p2 {
		t.Fatalf("unexpected path ordering in picture")
	}
}

func TestPictureSVGIntegration(t *testing.T) {
	pic := NewPicture()
	path, err := NewPath().MoveTo(P(0, 0)).CurveTo(P(10, 0)).SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve: %v", err)
	}
	other, err := NewPath().MoveTo(P(20, 0)).CurveTo(P(30, 10)).SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve other: %v", err)
	}
	pic.AddPath(path).AddPath(other)

	svg := svg.NewBuilder().FitViewBoxToPictures(pic).AddPicture(pic)
	var b strings.Builder
	if err := svg.WriteTo(&b); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	out := b.String()
	if cnt := strings.Count(out, "<path"); cnt != 2 {
		t.Fatalf("expected 2 path elements in svg, got %d", cnt)
	}
	if !strings.Contains(out, "viewBox") {
		t.Fatalf("viewBox not present in svg output")
	}
}

// TestPictureClip tests clipping a picture to a boundary path.
// Mirrors MetaPost's "clip p to q" syntax.
func TestPictureClip(t *testing.T) {
	engine := mp.NewEngine()

	// Create a path to clip
	content, err := NewPath().MoveTo(P(0, 0)).LineTo(P(100, 100)).SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve content: %v", err)
	}

	// Create a clipping boundary (a square)
	clipBoundary, err := NewPath().
		MoveTo(P(25, 25)).
		LineTo(P(75, 25)).
		LineTo(P(75, 75)).
		LineTo(P(25, 75)).
		Close().
		SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve clip boundary: %v", err)
	}

	// Create picture with clipping
	pic := NewPicture().AddPath(content).Clip(clipBoundary)

	// Verify clip path is set
	if pic.ClipPath() != clipBoundary {
		t.Fatalf("clip path not set correctly")
	}

	// Generate SVG and verify clip path elements
	svg := svg.NewBuilder().FitViewBoxToPaths(content, clipBoundary).AddPicture(pic)
	var b strings.Builder
	if err := svg.WriteTo(&b); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	out := b.String()

	// Check for clipPath definition
	if !strings.Contains(out, "<defs>") {
		t.Errorf("expected <defs> element for clip path")
	}
	if !strings.Contains(out, `<clipPath id="clip0"`) {
		t.Errorf("expected clipPath with id='clip0'")
	}

	// Check for clipped group
	if !strings.Contains(out, `clip-path="url(#clip0)"`) {
		t.Errorf("expected clip-path reference")
	}
}

// TestPictureClipSVGOutput verifies the complete SVG structure with clipping.
func TestPictureClipSVGOutput(t *testing.T) {
	// Create a circle-like path to clip
	engine := mp.NewEngine()
	circle, err := NewPath().
		MoveTo(P(50, 0)).
		CurveTo(P(100, 50)).
		CurveTo(P(50, 100)).
		CurveTo(P(0, 50)).
		Close().
		SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve circle: %v", err)
	}

	// Clip to a smaller square
	clipSquare, err := NewPath().
		MoveTo(P(20, 20)).
		LineTo(P(80, 20)).
		LineTo(P(80, 80)).
		LineTo(P(20, 80)).
		Close().
		SolveWithEngine(mp.NewEngine())
	if err != nil {
		t.Fatalf("solve clip square: %v", err)
	}

	pic := NewPicture().AddPath(circle).Clip(clipSquare)

	svg := svg.NewBuilder().FitViewBoxToPaths(circle).AddPicture(pic)
	var b strings.Builder
	if err := svg.WriteTo(&b); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	out := b.String()

	// The output should have the structure:
	// <svg ...>
	//   <defs><clipPath id="clip0"><path .../></clipPath></defs>
	//   <g clip-path="url(#clip0)"><path .../></g>
	// </svg>

	if !strings.Contains(out, "<defs>") {
		t.Errorf("missing <defs> section")
	}
	if !strings.Contains(out, "</defs>") {
		t.Errorf("missing </defs> closing tag")
	}
	if !strings.Contains(out, "<g clip-path=") {
		t.Errorf("missing clipped group")
	}
	if !strings.Contains(out, "</g>") {
		t.Errorf("missing </g> closing tag")
	}
}

// TestPictureNoClip verifies that pictures without clipping work normally.
func TestPictureNoClip(t *testing.T) {
	engine := mp.NewEngine()
	path, err := NewPath().MoveTo(P(0, 0)).CurveTo(P(50, 50)).SolveWithEngine(engine)
	if err != nil {
		t.Fatalf("solve: %v", err)
	}

	pic := NewPicture().AddPath(path)

	// ClipPath should be nil
	if pic.ClipPath() != nil {
		t.Errorf("expected nil clip path for unclipped picture")
	}

	svg := svg.NewBuilder().FitViewBoxToPaths(path).AddPicture(pic)
	var b strings.Builder
	if err := svg.WriteTo(&b); err != nil {
		t.Fatalf("write svg: %v", err)
	}
	out := b.String()

	// Should not have clipPath elements
	if strings.Contains(out, "<clipPath") {
		t.Errorf("unexpected clipPath element in unclipped picture")
	}
	if strings.Contains(out, "clip-path=") {
		t.Errorf("unexpected clip-path attribute in unclipped picture")
	}
}
