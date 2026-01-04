package mp

// TextToPathsOptions configures text-to-path conversion.
type TextToPathsOptions struct {
	FontSize float64 // Font size in points (default: 10)
	X, Y     float64 // Starting position
	Color    Color   // Fill color for the glyphs
}

// FontRenderer is the interface for converting text to glyph paths.
// This allows the font support to be in a separate package (github.com/boxesandglue/mpgo/font)
// which users can optionally import.
type FontRenderer interface {
	// TextToPaths converts a text string to a slice of filled paths (one per glyph).
	// Each path represents a glyph outline positioned correctly for the text layout.
	TextToPaths(text string, opts TextToPathsOptions) ([]*Path, error)

	// TextBounds returns the bounding box of shaped text.
	// Returns (width, height) in output units.
	TextBounds(text string, fontSize float64) (width, height float64)
}
