// Package draw provides a high-level fluent API for constructing MetaPost-style paths.
//
// This package wraps the low-level [mp] package with a builder pattern that makes
// path construction more intuitive and Go-idiomatic. It also provides equation solving
// for geometric constraints.
//
// # Path Building
//
// The [PathBuilder] provides a fluent interface for constructing paths:
//
//	path, err := draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    CurveTo(mp.P(50, 30)).
//	    CurveTo(mp.P(100, 0)).
//	    WithStrokeColor(mp.ColorCSS("blue")).
//	    Solve()
//
// # Curve Types
//
// Different curve segment types are supported:
//
//	CurveTo(pt)           // Smooth curve (Hobby algorithm)
//	LineTo(pt)            // Straight line segment
//	Controls(c1, c2, pt)  // Explicit Bézier control points
//
// # Direction and Tension
//
// Control the curve shape with directions and tensions:
//
//	draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    WithOutDir(45).          // Leave at 45°
//	    CurveTo(mp.P(100, 0)).
//	    WithInDir(-45).          // Arrive at -45°
//	    Solve()
//
//	draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    WithTension(2).          // Tighter curve
//	    CurveTo(mp.P(100, 0)).
//	    Solve()
//
// # Closed Paths
//
// Create closed paths with [PathBuilder.Close]:
//
//	triangle, _ := draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    LineTo(mp.P(100, 0)).
//	    LineTo(mp.P(50, 86)).
//	    Close().
//	    Solve()
//
// # Styling
//
// Apply stroke, fill, and other styling:
//
//	draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    CurveTo(mp.P(100, 0)).
//	    WithStrokeColor(mp.ColorCSS("red")).
//	    WithFillColor(mp.ColorCSS("yellow")).
//	    WithStrokeWidth(2.0).
//	    WithDash([]float64{5, 3}, 0).
//	    WithArrowEnd().
//	    Solve()
//
// # Pictures
//
// The [Picture] type collects multiple paths and labels, similar to MetaPost's picture:
//
//	pic := draw.NewPicture()
//	pic.AddPath(path1)
//	pic.AddPath(path2)
//	pic.Label("A", mp.P(0, 0), mp.AnchorLowerLeft)
//	pic.DotLabel("B", mp.P(100, 0), mp.AnchorRight, mp.ColorCSS("blue"))
//
// # Label Conversion
//
// Labels can be converted to glyph paths using the font package:
//
//	import "github.com/boxesandglue/mpgo/font"
//
//	face, _ := font.Load(fontFile)
//	pic.ConvertLabelsToPathsWithFont(face)
//
// # Equation Solving
//
// The [Context] type provides linear equation solving for geometric constraints:
//
//	ctx := draw.NewContext()
//	z0 := ctx.Known(0, 0)      // Fixed point
//	z1 := ctx.Unknown()        // Unknown point
//	z2 := ctx.Known(100, 100)  // Fixed point
//
//	ctx.Collinear(z1, z0, z2)  // z1 lies on line z0--z2
//	ctx.EqX(z1, 50)            // z1.x = 50
//	ctx.Solve()
//
//	fmt.Println(z1.XY())       // (50, 50)
//
// Variables can be used in path building:
//
//	path, _ := draw.NewPath().
//	    WithContext(ctx).
//	    MoveToVar(z0).
//	    CurveToVar(z1).
//	    CurveToVar(z2).
//	    Solve()
//
// # Transformations
//
// Apply transformations to paths:
//
//	draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    LineTo(mp.P(10, 0)).
//	    Scaled(5).
//	    Rotated(45).
//	    Shifted(100, 100).
//	    Solve()
package draw
