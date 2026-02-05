// Package font provides optional font support for converting text to glyph paths.
// Import this package only when you need to convert labels to outline paths.
//
// Example:
//
//	import "github.com/boxesandglue/mpgo/font"
//
//	f, err := os.Open("arial.ttf")
//	face, err := font.Load(f)
//	paths, err := label.ToPaths(face)
package font

import (
	"fmt"
	"io"

	"github.com/boxesandglue/mpgo/mp"
	"github.com/boxesandglue/textshape/ot"
)

// Face wraps a loaded font for text-to-path conversion.
// It implements the mp.FontRenderer interface.
type Face struct {
	font   *ot.Font
	face   *ot.Face
	shaper *ot.Shaper
	upem   float64
}

// Load loads a TrueType or OpenType font from a reader.
// The reader's contents are read into memory.
func Load(r io.Reader) (*Face, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read font data: %w", err)
	}

	return LoadFromBytes(data)
}

// LoadFromBytes loads a TrueType or OpenType font from byte data.
func LoadFromBytes(data []byte) (*Face, error) {
	font, err := ot.ParseFont(data, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	face, err := ot.NewFace(font)
	if err != nil {
		return nil, fmt.Errorf("failed to create face: %w", err)
	}

	shaper, err := ot.NewShaperFromFace(face)
	if err != nil {
		return nil, fmt.Errorf("failed to create shaper: %w", err)
	}

	return &Face{
		font:   font,
		face:   face,
		shaper: shaper,
		upem:   float64(face.Upem()),
	}, nil
}

// TextToPaths converts a text string to a slice of filled paths (one per glyph).
// Each path represents a glyph outline positioned correctly for the text layout.
func (f *Face) TextToPaths(text string, opts mp.TextToPathsOptions) ([]*mp.Path, error) {
	if opts.FontSize == 0 {
		opts.FontSize = mp.DefaultFontSize
	}

	// Create shaping buffer
	buf := ot.NewBuffer()
	buf.Direction = ot.DirectionLTR
	buf.Script = ot.ScriptToTag(ot.ScriptLatin)
	if tags := ot.LanguageToTag("en"); len(tags) > 0 {
		buf.Language = tags[0]
	}

	buf.AddString(text)

	// Shape the text
	f.shaper.Shape(buf, nil)

	// Scale factor from font units to output units
	scale := opts.FontSize / f.upem

	// Current position
	curX := opts.X
	curY := opts.Y

	var paths []*mp.Path

	// Process each shaped glyph
	for i := range buf.Info {
		glyphInfo := buf.Info[i]
		glyphPos := buf.Pos[i]

		// Get glyph outline
		outline, ok := f.face.GlyphOutline(glyphInfo.GlyphID)
		if !ok {
			// Skip non-outline glyphs (e.g., spaces)
			curX += float64(glyphPos.XAdvance) * scale
			curY += float64(glyphPos.YAdvance) * scale
			continue
		}

		// Calculate glyph position with offset
		glyphX := curX + float64(glyphPos.XOffset)*scale
		glyphY := curY + float64(glyphPos.YOffset)*scale

		// Convert outline to path
		path := outlineToPath(outline, scale, glyphX, glyphY)
		if path != nil && path.Head != nil {
			path.Style.Fill = opts.Color
			path.Style.Stroke = mp.ColorCSS("none")
			paths = append(paths, path)
		}

		// Advance position
		curX += float64(glyphPos.XAdvance) * scale
		curY += float64(glyphPos.YAdvance) * scale
	}

	return paths, nil
}

// TextBounds returns the bounding box of shaped text.
// Returns (width, height) in output units.
func (f *Face) TextBounds(text string, fontSize float64) (width, height float64) {
	if fontSize == 0 {
		fontSize = mp.DefaultFontSize
	}

	buf := ot.NewBuffer()
	buf.Direction = ot.DirectionLTR
	buf.Script = ot.ScriptToTag(ot.ScriptLatin)
	if tags := ot.LanguageToTag("en"); len(tags) > 0 {
		buf.Language = tags[0]
	}

	buf.AddString(text)
	f.shaper.Shape(buf, nil)

	scale := fontSize / f.upem

	// Calculate total advance
	var totalAdvance float64
	for i := range buf.Pos {
		totalAdvance += float64(buf.Pos[i].XAdvance)
	}

	// Use real font metrics for height
	ext := f.face.GetHExtents()
	ascender := float64(ext.Ascender)
	descender := float64(-ext.Descender) // Descender is negative

	return totalAdvance * scale, (ascender + descender) * scale
}

// outlineToPath converts a font glyph outline to an mp.Path.
func outlineToPath(outline ot.GlyphOutline, scale, offsetX, offsetY float64) *mp.Path {
	if len(outline.Segments) == 0 {
		return nil
	}

	path := mp.NewPath()
	var firstKnot *mp.Knot
	var lastKnot *mp.Knot
	var startX, startY float64

	for _, seg := range outline.Segments {
		switch seg.Op {
		case ot.SegmentMoveTo:
			// Start a new contour
			if lastKnot != nil && firstKnot != nil && lastKnot != firstKnot {
				// Close previous contour
				closeContour(firstKnot, lastKnot)
			}

			x := float64(seg.Args[0].X)*scale + offsetX
			y := float64(seg.Args[0].Y)*scale + offsetY
			startX, startY = x, y

			knot := &mp.Knot{
				XCoord: x,
				YCoord: y,
				LType:  mp.KnotExplicit,
				RType:  mp.KnotExplicit,
				LeftX:  x,
				LeftY:  y,
				RightX: x,
				RightY: y,
			}
			path.Append(knot)
			firstKnot = knot
			lastKnot = knot

		case ot.SegmentLineTo:
			x := float64(seg.Args[0].X)*scale + offsetX
			y := float64(seg.Args[0].Y)*scale + offsetY

			// Set the right control point of the previous knot
			if lastKnot != nil {
				lastKnot.RightX = lastKnot.XCoord
				lastKnot.RightY = lastKnot.YCoord
			}

			knot := &mp.Knot{
				XCoord: x,
				YCoord: y,
				LType:  mp.KnotExplicit,
				RType:  mp.KnotExplicit,
				LeftX:  x,
				LeftY:  y,
				RightX: x,
				RightY: y,
			}
			path.Append(knot)
			lastKnot = knot

		case ot.SegmentQuadTo:
			// Quadratic Bezier - convert to cubic
			ctrlX := float64(seg.Args[0].X)*scale + offsetX
			ctrlY := float64(seg.Args[0].Y)*scale + offsetY
			endX := float64(seg.Args[1].X)*scale + offsetX
			endY := float64(seg.Args[1].Y)*scale + offsetY

			if lastKnot != nil {
				// Convert quadratic to cubic control points
				// Cubic ctrl1 = start + 2/3 * (qctrl - start)
				// Cubic ctrl2 = end + 2/3 * (qctrl - end)
				lastKnot.RightX = lastKnot.XCoord + 2.0/3.0*(ctrlX-lastKnot.XCoord)
				lastKnot.RightY = lastKnot.YCoord + 2.0/3.0*(ctrlY-lastKnot.YCoord)

				knot := &mp.Knot{
					XCoord: endX,
					YCoord: endY,
					LType:  mp.KnotExplicit,
					RType:  mp.KnotExplicit,
					LeftX:  endX + 2.0/3.0*(ctrlX-endX),
					LeftY:  endY + 2.0/3.0*(ctrlY-endY),
					RightX: endX,
					RightY: endY,
				}
				path.Append(knot)
				lastKnot = knot
			}

		case ot.SegmentCubeTo:
			// Cubic Bezier
			ctrl1X := float64(seg.Args[0].X)*scale + offsetX
			ctrl1Y := float64(seg.Args[0].Y)*scale + offsetY
			ctrl2X := float64(seg.Args[1].X)*scale + offsetX
			ctrl2Y := float64(seg.Args[1].Y)*scale + offsetY
			endX := float64(seg.Args[2].X)*scale + offsetX
			endY := float64(seg.Args[2].Y)*scale + offsetY

			if lastKnot != nil {
				lastKnot.RightX = ctrl1X
				lastKnot.RightY = ctrl1Y
			}

			knot := &mp.Knot{
				XCoord: endX,
				YCoord: endY,
				LType:  mp.KnotExplicit,
				RType:  mp.KnotExplicit,
				LeftX:  ctrl2X,
				LeftY:  ctrl2Y,
				RightX: endX,
				RightY: endY,
			}
			path.Append(knot)
			lastKnot = knot
		}
	}

	// Close the last contour
	if lastKnot != nil && firstKnot != nil && lastKnot != firstKnot {
		// Check if we need to close back to start
		if lastKnot.XCoord != startX || lastKnot.YCoord != startY {
			closeContour(firstKnot, lastKnot)
		} else {
			closeContour(firstKnot, lastKnot)
		}
	}

	return path
}

// closeContour closes a contour by connecting the last knot back to the first.
func closeContour(first, last *mp.Knot) {
	if first == nil || last == nil {
		return
	}
	// The path is already a circular linked list from Append,
	// but we need to ensure control points are set correctly for the closing segment
	last.RightX = last.XCoord
	last.RightY = last.YCoord
	first.LeftX = first.XCoord
	first.LeftY = first.YCoord
}
