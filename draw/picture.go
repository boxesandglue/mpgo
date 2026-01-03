package draw

import "github.com/boxesandglue/mpgo/mp"

// Picture mirrors MetaPost's picture container: it collects solved paths that can
// be drawn together. Tracks are stored as-is (no copying) similar to how MetaPost
// chains edge objects into a picture (mp.c around mp_make_dashes/export_dashes).
type Picture struct {
	paths    []*mp.Path
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
