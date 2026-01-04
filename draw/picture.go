package draw

import "github.com/boxesandglue/mpgo/mp"

// Picture mirrors MetaPost's picture container: it collects solved paths that can
// be drawn together. Tracks are stored as-is (no copying) similar to how MetaPost
// chains edge objects into a picture (mp.c around mp_make_dashes/export_dashes).
type Picture struct {
	paths    []*mp.Path
	labels   []*mp.Label
	clipPath *mp.Path // Optional clipping path
}

// NewPicture constructs an empty picture.
func NewPicture() *Picture {
	return &Picture{paths: make([]*mp.Path, 0)}
}

// AddPath appends a solved path to the picture.
func (p *Picture) AddPath(path *mp.Path) *Picture {
	if path != nil {
		p.paths = append(p.paths, path)
	}
	return p
}

// AddPicture appends all paths from another picture (no copies; mirrors MetaPost's
// picture addition semantics where edges are shared until output).
func (p *Picture) AddPicture(other *Picture) *Picture {
	if other == nil {
		return p
	}
	p.paths = append(p.paths, other.paths...)
	return p
}

// Paths exposes the collected paths.
func (p *Picture) Paths() []*mp.Path {
	return p.paths
}

// Clip sets the clipping path for this picture.
// Mirrors MetaPost's "clip p to q" where q is the clipping boundary.
// All paths in the picture will be clipped to this boundary when rendered.
func (p *Picture) Clip(clipPath *mp.Path) *Picture {
	p.clipPath = clipPath
	return p
}

// ClipPath returns the current clipping path, or nil if none is set.
func (p *Picture) ClipPath() *mp.Path {
	return p.clipPath
}

// Label adds a text label to the picture at the given position.
// Mirrors MetaPost's label@#(s, z) command.
//
// Example:
//
//	pic.Label("A", mp.P(0, 0), mp.AnchorTop)      // label.top("A", origin)
//	pic.Label("B", mp.P(100, 0), mp.AnchorRight)  // label.rt("B", z1)
func (p *Picture) Label(text string, pos mp.Point, anchor mp.Anchor) *Picture {
	label := mp.NewLabel(text, pos, anchor)
	p.labels = append(p.labels, label)
	return p
}

// LabelWithStyle adds a styled text label to the picture.
// Returns the created label for further customization.
func (p *Picture) LabelWithStyle(text string, pos mp.Point, anchor mp.Anchor) *mp.Label {
	label := mp.NewLabel(text, pos, anchor)
	p.labels = append(p.labels, label)
	return label
}

// DotLabel adds a text label with a dot at the reference point.
// Mirrors MetaPost's dotlabel@#(s, z) command.
//
// Example:
//
//	pic.DotLabel("$z_0$", z0, mp.AnchorLowerRight)  // dotlabel.lrt("$z_0$", z0)
func (p *Picture) DotLabel(text string, pos mp.Point, anchor mp.Anchor, color mp.Color) *Picture {
	// Add the label
	label := mp.NewLabel(text, pos, anchor)
	p.labels = append(p.labels, label)

	// Add the dot (a filled circle at the position)
	dot := mp.FullCircle()
	dot = mp.Scaled(mp.DefaultDotLabelDiam).ApplyToPath(dot)
	dot = mp.Shifted(pos.X, pos.Y).ApplyToPath(dot)
	dot.Style.Fill = color
	dot.Style.Stroke = mp.ColorCSS("none")
	p.paths = append(p.paths, dot)

	return p
}

// AddLabel adds a pre-configured label to the picture.
func (p *Picture) AddLabel(label *mp.Label) *Picture {
	if label != nil {
		p.labels = append(p.labels, label)
	}
	return p
}

// Labels returns all labels in the picture.
func (p *Picture) Labels() []*mp.Label {
	return p.labels
}

// ConvertLabelsToPathsWithFont converts all labels to glyph outline paths
// using the provided font. The converted paths are added to the picture's
// path list, and the labels are cleared.
// This mirrors MetaPost's behavior where text becomes a picture with glyph paths.
//
// The font parameter must implement mp.FontRenderer. Use the font package:
//
//	import "github.com/boxesandglue/mpgo/font"
//	face, _ := font.Load(fontReader)
//	pic.ConvertLabelsToPathsWithFont(face)
func (p *Picture) ConvertLabelsToPathsWithFont(f mp.FontRenderer) error {
	if f == nil {
		return nil
	}

	for _, label := range p.labels {
		paths, err := label.ToPaths(f)
		if err != nil {
			return err
		}
		p.paths = append(p.paths, paths...)
	}

	// Clear labels after conversion
	p.labels = nil
	return nil
}
