// Package mp implements MetaPost's Hobby-Knuth curve-solving algorithm in Go.
//
// This package is a port of the core MetaPost engine, providing the same
// smooth curve generation that MetaPost is famous for. Given a sequence of
// points with optional direction and tension constraints, the solver computes
// optimal cubic Bézier control points.
//
// # Architecture
//
// The package is organized around these core concepts:
//
//   - [Knot]: A point on a path with coordinates, control points, and type information
//   - [Path]: A linked list of knots forming an open or closed curve
//   - [Engine]: The solver that computes Bézier control points
//   - [Transform]: Affine transformations (scale, rotate, shift, etc.)
//
// # Quick Start
//
// The simplest way to create paths is using the higher-level draw package:
//
//	import "github.com/boxesandglue/mpgo/draw"
//
//	path, _ := draw.NewPath().
//	    MoveTo(mp.P(0, 0)).
//	    CurveTo(mp.P(100, 50)).
//	    CurveTo(mp.P(200, 0)).
//	    Solve()
//
// For direct use of the mp package:
//
//	engine := mp.NewEngine()
//	path := mp.NewPath()
//	// ... add knots ...
//	engine.AddPath(path)
//	engine.Solve()
//
// # Predefined Paths
//
// The package provides MetaPost's standard path primitives:
//
//	mp.FullCircle()    // Unit circle (diameter 1) centered at origin
//	mp.HalfCircle()    // Upper half of unit circle
//	mp.QuarterCircle() // First quadrant arc
//	mp.UnitSquare()    // Unit square from (0,0) to (1,1)
//
// These can be transformed using [Transform]:
//
//	circle := mp.FullCircle()
//	circle = mp.Scaled(50).ApplyToPath(circle)           // Scale to diameter 50
//	circle = mp.Shifted(100, 100).ApplyToPath(circle)    // Move center to (100,100)
//
// # Transformations
//
// Affine transformations mirror MetaPost's transform operations:
//
//	mp.Scaled(s)           // Uniform scaling
//	mp.XScaled(s)          // Horizontal scaling
//	mp.YScaled(s)          // Vertical scaling
//	mp.Shifted(dx, dy)     // Translation
//	mp.Rotated(degrees)    // Rotation around origin
//	mp.RotatedAround(p, d) // Rotation around point p
//	mp.Slanted(s)          // Slant (shear) transformation
//
// Transformations can be combined:
//
//	t := mp.Scaled(2).Concat(mp.Rotated(45)).Concat(mp.Shifted(10, 20))
//	path = t.ApplyToPath(path)
//
// # Points and Colors
//
// Helper functions for creating points and colors:
//
//	mp.P(x, y)              // Create a point
//	mp.ColorRGB(r, g, b)    // RGB color (0-1 range)
//	mp.ColorCSS("red")      // CSS color name or hex
//	mp.ColorCMYK(c, m, y, k) // CMYK color
//
// # Labels
//
// Text labels can be attached to points with anchor positioning:
//
//	label := mp.NewLabel("A", mp.P(0, 0), mp.AnchorLowerLeft)
//
// Available anchors match MetaPost's label suffixes:
//
//	mp.AnchorCenter      // label(s, z)
//	mp.AnchorLeft        // label.lft(s, z)
//	mp.AnchorRight       // label.rt(s, z)
//	mp.AnchorTop         // label.top(s, z)
//	mp.AnchorBottom      // label.bot(s, z)
//	mp.AnchorUpperLeft   // label.ulft(s, z)
//	mp.AnchorUpperRight  // label.urt(s, z)
//	mp.AnchorLowerLeft   // label.llft(s, z)
//	mp.AnchorLowerRight  // label.lrt(s, z)
//
// Labels can be converted to glyph outline paths using the font package:
//
//	import "github.com/boxesandglue/mpgo/font"
//
//	face, _ := font.Load(fontFile)
//	paths, _ := label.ToPaths(face)
//
// # Path Styling
//
// Paths have a Style field for stroke, fill, and other attributes:
//
//	path.Style.Stroke = mp.ColorCSS("blue")
//	path.Style.Fill = mp.ColorCSS("yellow")
//	path.Style.StrokeWidth = 2.0
//	path.Style.Dash = &mp.DashPattern{Array: []float64{5, 3}}
//	path.Style.Arrow.End = true  // Arrow at end of path
//
// # Path Operations
//
// The package provides path manipulation functions:
//
//	mp.Reverse(path)           // Reverse path direction
//	mp.Subpath(path, t1, t2)   // Extract portion of path
//	mp.ArcLength(path)         // Total arc length
//	mp.PointAt(path, t)        // Point at parameter t
//	mp.DirectionAt(path, t)    // Tangent direction at t
//
// # Pens and Envelopes
//
// Non-circular pens create envelope paths (like MetaPost's pencircle transformations):
//
//	pen := mp.PenCircle()
//	pen = mp.XScaled(3).ApplyToPen(pen)  // Elliptical pen
//	path.Style.Pen = pen
//
// # References
//
// This implementation follows the algorithms described in:
//   - John D. Hobby, "Smooth, Easy to Compute Interpolating Splines" (1986)
//   - Donald E. Knuth, "The METAFONTbook" (1986)
//   - The MetaPost source code (mp.w, mp.c)
package mp
