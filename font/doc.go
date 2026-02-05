// Package font provides optional TrueType/OpenType font support for mpgo.
//
// This package allows converting text labels to glyph outline paths, similar to
// MetaPost's "text infont fontname" mechanism. Import this package only when you
// need font rendering - it adds approximately 2.3 MB to the binary size due to
// the text shaping dependencies.
//
// # Loading Fonts
//
// Load a font from a file or reader:
//
//	f, err := os.Open("/path/to/font.ttf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer f.Close()
//
//	face, err := font.Load(f)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Or load from byte data:
//
//	face, err := font.LoadFromBytes(fontData)
//
// # Converting Labels to Paths
//
// Convert a single label to glyph paths:
//
//	label := mp.NewLabel("Hello", mp.P(0, 0), mp.AnchorCenter)
//	paths, err := label.ToPaths(face)
//
// Convert all labels in a picture:
//
//	pic := draw.NewPicture()
//	pic.Label("A", mp.P(0, 0), mp.AnchorLowerLeft)
//	pic.Label("B", mp.P(100, 0), mp.AnchorLowerRight)
//
//	pic.ConvertLabelsToPathsWithFont(face)
//	// Labels are now glyph outline paths in pic.Paths()
//
// # Text to Paths Directly
//
// Convert text to paths without using labels:
//
//	paths, err := face.TextToPaths("Hello", mp.TextToPathsOptions{
//	    FontSize: 12,
//	    X:        0,
//	    Y:        0,
//	    Color:    mp.ColorCSS("black"),
//	})
//
// # Text Bounds
//
// Get the dimensions of shaped text:
//
//	width, height := face.TextBounds("Hello", 12)
//
// # Without Font Support
//
// If you don't import this package, labels are rendered as SVG <text> elements
// instead of glyph paths. This keeps the binary smaller but relies on the
// viewing application having the specified font installed.
//
// # Implementation Details
//
// This package uses the textshape library for:
//   - TrueType/OpenType font parsing
//   - OpenType text shaping (proper glyph selection and positioning)
//   - Glyph outline extraction
//
// Quadratic Bézier curves (common in TrueType) are automatically converted
// to cubic Bézier curves to match mpgo's path representation.
package font
