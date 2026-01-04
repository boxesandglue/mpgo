package mp

import "fmt"

// Anchor specifies the positioning of a label relative to its reference point.
// These mirror MetaPost's label suffixes (.lft, .rt, .top, .bot, etc.).
type Anchor int

const (
	AnchorCenter     Anchor = iota // label(s, z) - centered at z
	AnchorLeft                     // label.lft(s, z) - label to the left of z
	AnchorRight                    // label.rt(s, z) - label to the right of z
	AnchorTop                      // label.top(s, z) - label above z
	AnchorBottom                   // label.bot(s, z) - label below z
	AnchorUpperLeft                // label.ulft(s, z) - label upper-left of z
	AnchorUpperRight               // label.urt(s, z) - label upper-right of z
	AnchorLowerLeft                // label.llft(s, z) - label lower-left of z
	AnchorLowerRight               // label.lrt(s, z) - label lower-right of z
)

// DefaultLabelOffset is the default distance between the reference point and
// the label text. Mirrors MetaPost's labeloffset (3bp in plain.mp).
const DefaultLabelOffset = 3.0

// DefaultFontSize is the default font size for labels.
// Corresponds to MetaPost's defaultscale with cmr10 (10pt).
const DefaultFontSize = 10.0

// DefaultDotLabelDiam is the default diameter for dots in DotLabel.
// Mirrors MetaPost's dotlabeldiam (3bp in plain.mp).
const DefaultDotLabelDiam = 3.0

// Label represents a text label positioned near a point.
// This is a simplified version of MetaPost's label that works with plain text
// instead of btex...etex typeset content.
type Label struct {
	Text        string  // The label text
	Position    Point   // Reference point (z in MetaPost's label(s, z))
	Anchor      Anchor  // Positioning relative to the reference point
	Color       Color   // Text color (default: black)
	FontSize    float64 // Font size in points (default: 10)
	FontFamily  string  // Font family (default: sans-serif for SVG)
	LabelOffset float64 // Distance from reference point (default: 3bp)
}

// NewLabel creates a new label with default settings.
func NewLabel(text string, pos Point, anchor Anchor) *Label {
	return &Label{
		Text:        text,
		Position:    pos,
		Anchor:      anchor,
		Color:       ColorCSS("black"),
		FontSize:    DefaultFontSize,
		FontFamily:  "sans-serif",
		LabelOffset: DefaultLabelOffset,
	}
}

// WithColor sets the label color.
func (l *Label) WithColor(c Color) *Label {
	l.Color = c
	return l
}

// WithFontSize sets the font size.
func (l *Label) WithFontSize(size float64) *Label {
	l.FontSize = size
	return l
}

// WithFontFamily sets the font family.
func (l *Label) WithFontFamily(family string) *Label {
	l.FontFamily = family
	return l
}

// WithOffset sets the label offset distance.
func (l *Label) WithOffset(offset float64) *Label {
	l.LabelOffset = offset
	return l
}

// LabelOffsetVector returns the offset direction vector for an anchor.
// These values mirror MetaPost's laboff pairs from plain.mp.
func LabelOffsetVector(anchor Anchor) (dx, dy float64) {
	switch anchor {
	case AnchorCenter:
		return 0, 0
	case AnchorLeft:
		return -1, 0
	case AnchorRight:
		return 1, 0
	case AnchorTop:
		return 0, 1
	case AnchorBottom:
		return 0, -1
	case AnchorUpperLeft:
		return -0.7, 0.7
	case AnchorUpperRight:
		return 0.7, 0.7
	case AnchorLowerLeft:
		return -0.7, -0.7
	case AnchorLowerRight:
		return 0.7, -0.7
	default:
		return 0, 0
	}
}

// LabelAnchorFactors returns the anchor factors (labxf, labyf) for positioning.
// These determine which point of the label's bounding box is placed at the
// offset position. Values mirror MetaPost's labxf/labyf from plain.mp.
//
// Returns (xf, yf) where:
//   - xf=0 means left edge of text, xf=1 means right edge, xf=0.5 means center
//   - yf=0 means top edge of text, yf=1 means bottom edge, yf=0.5 means middle
func LabelAnchorFactors(anchor Anchor) (xf, yf float64) {
	// Values from plain.mp:
	// labxf=.5;  labyf=.5;     (center)
	// labxf.lft=1;   labyf.lft=.5;
	// labxf.rt=0;    labyf.rt=.5;
	// labxf.bot=.5;  labyf.bot=1;
	// labxf.top=.5;  labyf.top=0;
	// labxf.ulft=1;  labyf.ulft=0;
	// labxf.urt=0;   labyf.urt=0;
	// labxf.llft=1;  labyf.llft=1;
	// labxf.lrt=0;   labyf.lrt=1;
	switch anchor {
	case AnchorCenter:
		return 0.5, 0.5
	case AnchorLeft:
		return 1, 0.5 // right edge of text at position (text is to the left)
	case AnchorRight:
		return 0, 0.5 // left edge of text at position (text is to the right)
	case AnchorTop:
		return 0.5, 0 // top edge of text at offset position (text hangs down from above)
	case AnchorBottom:
		return 0.5, 1 // bottom edge of text at offset position (text extends up from below)
	case AnchorUpperLeft:
		return 1, 0 // top-right corner at position
	case AnchorUpperRight:
		return 0, 0 // top-left corner at position
	case AnchorLowerLeft:
		return 1, 1 // bottom-right corner at position
	case AnchorLowerRight:
		return 0, 1 // bottom-left corner at position
	default:
		return 0.5, 0.5
	}
}

// ToPaths converts the label text to glyph outline paths using the provided font.
// The paths are positioned according to the label's Position, Anchor, and LabelOffset.
// Returns a slice of filled paths representing each glyph.
//
// The font parameter must implement the FontRenderer interface. Use the font package
// to load fonts:
//
//	import "github.com/boxesandglue/mpgo/font"
//	face, _ := font.Load(fontReader)
//	paths, _ := label.ToPaths(face)
func (l *Label) ToPaths(f FontRenderer) ([]*Path, error) {
	if f == nil {
		return nil, fmt.Errorf("font is required for ToPaths")
	}

	fontSize := l.FontSize
	if fontSize == 0 {
		fontSize = DefaultFontSize
	}
	offset := l.LabelOffset
	if offset == 0 {
		offset = DefaultLabelOffset
	}

	// Get text dimensions for anchor calculation
	textWidth, textHeight := f.TextBounds(l.Text, fontSize)

	// Calculate offset direction
	dx, dy := LabelOffsetVector(l.Anchor)

	// Calculate anchor point (where text reference point goes)
	anchorX := l.Position.X + dx*offset
	anchorY := l.Position.Y + dy*offset

	// Calculate text origin based on anchor factors
	xf, yf := LabelAnchorFactors(l.Anchor)
	// xf=0: left edge at anchor, xf=1: right edge at anchor
	// yf=0: bottom edge at anchor, yf=1: top edge at anchor
	textX := anchorX - xf*textWidth
	textY := anchorY - yf*textHeight // yf=0: baseline at anchor, yf=1: top at anchor

	// Get color
	color := l.Color
	if color.CSS() == "" {
		color = ColorCSS("black")
	}

	// Convert text to paths
	return f.TextToPaths(l.Text, TextToPathsOptions{
		FontSize: fontSize,
		X:        textX,
		Y:        textY,
		Color:    color,
	})
}

// EstimateBounds returns an estimated bounding box for the label.
// Since we don't have actual font metrics, this uses approximations:
//   - Character width ≈ fontSize * 0.6 (average for sans-serif)
//   - Character height ≈ fontSize
//
// Returns (minX, minY, maxX, maxY) in the same coordinate system as Position.
func (l *Label) EstimateBounds() (minX, minY, maxX, maxY float64) {
	fontSize := l.FontSize
	if fontSize == 0 {
		fontSize = DefaultFontSize
	}
	offset := l.LabelOffset
	if offset == 0 {
		offset = DefaultLabelOffset
	}

	// Estimate text dimensions
	textWidth := fontSize * 0.6 * float64(len(l.Text))
	textHeight := fontSize

	// Get offset direction and anchor factors
	dx, dy := LabelOffsetVector(l.Anchor)
	xf, yf := LabelAnchorFactors(l.Anchor)

	// Calculate the anchor point (where the text reference point is)
	anchorX := l.Position.X + dx*offset
	anchorY := l.Position.Y + dy*offset

	// Calculate text bounding box based on anchor factors
	// xf=0: left edge at anchor, xf=1: right edge at anchor
	// yf=0: top edge at anchor, yf=1: bottom edge at anchor
	minX = anchorX - xf*textWidth
	maxX = anchorX + (1-xf)*textWidth
	minY = anchorY - (1-yf)*textHeight // In MetaPost coords, y increases upward
	maxY = anchorY + yf*textHeight

	return minX, minY, maxX, maxY
}
