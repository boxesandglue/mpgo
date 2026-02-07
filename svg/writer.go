package svg

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/boxesandglue/mpgo/mp"
)

// formatLineCap returns the SVG stroke-linecap value for a given LineCap constant.
// Defaults to "round" (MetaPost default) if unset or unknown.
func formatLineCap(cap int) string {
	switch cap {
	case mp.LineCapButt:
		return "butt"
	case mp.LineCapSquared:
		return "square"
	default: // LineCapDefault, LineCapRounded, or unknown
		return "round"
	}
}

// formatLineJoin returns the SVG stroke-linejoin value for a given LineJoin constant.
// Defaults to "round" (MetaPost default) if unset or unknown.
func formatLineJoin(join int) string {
	switch join {
	case mp.LineJoinMiter:
		return "miter"
	case mp.LineJoinBevel:
		return "bevel"
	default: // LineJoinDefault, LineJoinRound, or unknown
		return "round"
	}
}

// FormatDashAttrs returns SVG stroke-dasharray and stroke-dashoffset attributes
// for a dash pattern. Returns empty string if dash is nil.
// Mirrors MetaPost's SVG output (svgout.w:1089ff).
func FormatDashAttrs(dash *mp.DashPattern) string {
	if dash == nil || len(dash.Array) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(` stroke-dasharray="`)
	for i, v := range dash.Array {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%.2f", v)
	}
	b.WriteString(`"`)
	if dash.Offset != 0 {
		fmt.Fprintf(&b, ` stroke-dashoffset="%.2f"`, dash.Offset)
	}
	return b.String()
}

// PathToSVG converts a solved mp.Path into an SVG path string.
// It expects explicit control points to be present (after Solve).
func PathToSVG(path *mp.Path) string {
	if path == nil || path.Head == nil {
		return ""
	}
	var b strings.Builder
	h := path.Head
	fmt.Fprintf(&b, "M %.3f %.3f", h.XCoord, h.YCoord)
	p := h
	isClosed := false
	for {
		q := p.Next
		// Check if this is a straight line:
		// 1. Controls equal endpoints (degenerate case)
		// 2. Controls are collinear with endpoints (common for curl/curl segments)
		isLine := (p.RightX == p.XCoord && p.RightY == p.YCoord &&
			q.LeftX == q.XCoord && q.LeftY == q.YCoord)
		if !isLine {
			// Check collinearity using cross product: (c1-p) x (q-p) and (c2-p) x (q-p)
			// If both are ~0, the control points lie on the line from p to q
			dx := q.XCoord - p.XCoord
			dy := q.YCoord - p.YCoord
			cross1 := (p.RightX-p.XCoord)*dy - (p.RightY-p.YCoord)*dx
			cross2 := (q.LeftX-p.XCoord)*dy - (q.LeftY-p.YCoord)*dx
			const eps = 1e-6
			if cross1 > -eps && cross1 < eps && cross2 > -eps && cross2 < eps {
				isLine = true
			}
		}
		if isLine {
			fmt.Fprintf(&b, " L %.3f %.3f", q.XCoord, q.YCoord)
		} else {
			fmt.Fprintf(&b, " C %.3f %.3f %.3f %.3f %.3f %.3f",
				p.RightX, p.RightY,
				q.LeftX, q.LeftY,
				q.XCoord, q.YCoord)
		}
		p = q
		if p.RType == mp.KnotEndpoint {
			break
		}
		if p == h {
			isClosed = true
			break
		}
	}
	if isClosed {
		b.WriteString("Z")
	}
	return b.String()
}

// PathToSVGFlipped converts a solved mp.Path into an SVG path string with Y-coordinates
// flipped around the given height. This produces MetaPost-compatible output where
// y_svg = height - y_math. Use height = maxY + minY for proper alignment.
func PathToSVGFlipped(path *mp.Path, height float64) string {
	return PathToSVGTransformed(path, 0, 0, height)
}

// PathToSVGTransformed converts a path to SVG with coordinate transformation.
// offsetX, offsetY: subtracted from coordinates (to shift origin)
// flipHeight: Y is flipped as (flipHeight - y)
func PathToSVGTransformed(path *mp.Path, offsetX, offsetY, flipHeight float64) string {
	if path == nil || path.Head == nil {
		return ""
	}
	transformX := func(x float64) float64 { return x - offsetX }
	transformY := func(y float64) float64 { return flipHeight - y + offsetY }
	var b strings.Builder
	h := path.Head
	fmt.Fprintf(&b, "M %.6f %.6f", transformX(h.XCoord), transformY(h.YCoord))
	p := h
	isClosed := false
	for {
		q := p.Next
		isLine := (p.RightX == p.XCoord && p.RightY == p.YCoord &&
			q.LeftX == q.XCoord && q.LeftY == q.YCoord)
		if !isLine {
			dx := q.XCoord - p.XCoord
			dy := q.YCoord - p.YCoord
			cross1 := (p.RightX-p.XCoord)*dy - (p.RightY-p.YCoord)*dx
			cross2 := (q.LeftX-p.XCoord)*dy - (q.LeftY-p.YCoord)*dx
			const eps = 1e-6
			if cross1 > -eps && cross1 < eps && cross2 > -eps && cross2 < eps {
				isLine = true
			}
		}
		if isLine {
			fmt.Fprintf(&b, "L %.6f %.6f", transformX(q.XCoord), transformY(q.YCoord))
		} else {
			fmt.Fprintf(&b, "C %.6f %.6f,%.6f %.6f,%.6f %.6f",
				transformX(p.RightX), transformY(p.RightY),
				transformX(q.LeftX), transformY(q.LeftY),
				transformX(q.XCoord), transformY(q.YCoord))
		}
		p = q
		if p.RType == mp.KnotEndpoint {
			break
		}
		if p == h {
			isClosed = true
			break
		}
	}
	if isClosed {
		b.WriteString("Z")
	}
	return b.String()
}

// bbox1D computes min/max for a cubic Bezier in one dimension and merges it
// into the provided running min/max. Mirrors mp_bound_cubic ideas (mp.c:9434ff).
func bbox1D(p0, p1, p2, p3 float64, curMin, curMax float64) (float64, float64) {
	expand := func(v float64) {
		if v < curMin {
			curMin = v
		}
		if v > curMax {
			curMax = v
		}
	}
	expand(p0)
	expand(p3)

	// Derivative coefficients: 3(-p0+3p1-3p2+p3) t^2 + 2(2p0-4p1+2p2) t + (p1-p0)
	a := -p0 + 3*p1 - 3*p2 + p3
	b := 2 * (p0 - 2*p1 + p2)
	c := p1 - p0

	if a == 0 {
		if b != 0 {
			t := -c / b
			if t > 0 && t < 1 {
				expand(cubicAt(p0, p1, p2, p3, t))
			}
		}
		return curMin, curMax
	}

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return curMin, curMax
	}
	sqrtDisc := math.Sqrt(discriminant)
	for _, t := range []float64{(-b + sqrtDisc) / (2 * a), (-b - sqrtDisc) / (2 * a)} {
		if t > 0 && t < 1 {
			expand(cubicAt(p0, p1, p2, p3, t))
		}
	}
	return curMin, curMax
}

func cubicAt(p0, p1, p2, p3, t float64) float64 {
	mt := 1 - t
	return mt*mt*mt*p0 + 3*mt*mt*t*p1 + 3*mt*t*t*p2 + t*t*t*p3
}

// PathBBox computes the bounding box (minX, minY, maxX, maxY) for a path,
// including cubic extrema.
func PathBBox(p *mp.Path) (minX, minY, maxX, maxY float64) {
	if p == nil || p.Head == nil {
		return 0, 0, 0, 0
	}
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)
	expand := func(x, y float64) {
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	k := p.Head
	for {
		q := k.Next
		expand(k.XCoord, k.YCoord)
		minX, maxX = bbox1D(k.XCoord, k.RightX, q.LeftX, q.XCoord, minX, maxX)
		minY, maxY = bbox1D(k.YCoord, k.RightY, q.LeftY, q.YCoord, minY, maxY)
		k = q
		if k == p.Head || k == nil || k.RType == mp.KnotEndpoint {
			expand(k.XCoord, k.YCoord)
			break
		}
	}
	return minX, minY, maxX, maxY
}

// Builder is a tiny SVG writer for demo purposes.
type Builder struct {
	width, height  float64
	paths          []string
	labels         []*mp.Label // Labels to render as SVG text elements
	bg             string
	viewBox        string
	viewBoxSet     bool // True if viewBox was explicitly set (FitViewBoxToPaths called)
	stroke         mp.Color
	fill           mp.Color
	strokeWidth    float64
	lineCap        int // Default line cap (uses mp.LineCap* constants)
	lineJoin       int // Default line join (uses mp.LineJoin* constants)
	flipY          bool
	autoSize       bool
	padding        float64
	metaPostCompat bool           // Output in MetaPost-compatible format (Y-down in path data, no transform)
	mpPaths        []*mp.Path     // Store paths for MetaPost-compatible rendering
	mpOrigPaths    []*mp.Path     // Store original paths (before envelope substitution) for auto viewBox
	mpMinY, mpMaxY float64        // Bounding box for Y-flip calculation
	mpOffsetX      float64        // X offset for coordinate transformation (minX - halfStroke)
	mpOffsetY      float64        // Y offset for coordinate transformation (minY - halfStroke)
	clipPaths      []*mp.Path     // Clip paths (each gets an ID)
	clippedGroups  []clippedGroup // Groups of paths with their clip path index
}

// clippedGroup represents a set of paths that share a clip path.
type clippedGroup struct {
	clipIndex int        // Index into clipPaths (-1 means no clip)
	paths     []*mp.Path // Paths in this group
}

// NewBuilder constructs a Builder. When called without dimensions, it enables
// auto sizing based on the computed viewBox (e.g., via FitViewBoxToPaths).
// MetaPost-compatible mode is enabled by default for accurate output matching.
func NewBuilder(dim ...float64) *Builder {
	var w, h float64
	if len(dim) >= 2 {
		w, h = dim[0], dim[1]
	}
	return &Builder{
		width:          w,
		height:         h,
		paths:          make([]string, 0),
		bg:             "",
		stroke:         mp.ColorCSS("black"),
		fill:           mp.ColorCSS("none"),
		strokeWidth:    0.5,   // MetaPost default: pencircle scaled 0.5pt
		metaPostCompat: true,  // Default to MetaPost-compatible output
		flipY:          false, // Not needed in MetaPost-compatible mode
		autoSize:       len(dim) == 0,
	}
}

// Padding sets a default padding (in the same units as the paths) to apply when
// computing a viewBox via FitViewBoxToPaths or AutoViewBox.
func (s *Builder) Padding(p float64) *Builder {
	s.padding = p
	return s
}

// AutoViewBox sets a viewBox with padding.
func (s *Builder) AutoViewBox(pad float64) *Builder {
	if s.padding != 0 {
		pad = s.padding
	} else {
		s.padding = pad
	}
	s.viewBox = fmt.Sprintf("%g %g %g %g", -pad, -pad, s.width+2*pad, s.height+2*pad)
	return s
}

// FitViewBoxToPaths computes a tight bounding box over the given paths (including
// cubic extrema) and applies padding plus half the max stroke width (builder
// default or per-path Style). Useful to ensure the full drawing is visible
// regardless of content size.
//
// Note: This is now called automatically by WriteTo if not called explicitly.
// You only need to call it manually if you want to set padding or use specific paths.
func (s *Builder) FitViewBoxToPaths(paths ...*mp.Path) *Builder {
	s.viewBoxSet = true
	pad := s.padding
	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)
	maxStroke := s.strokeWidth
	hasEnvelope := false // Track if any path has an envelope (for stroke padding)
	expand := func(x, y float64) {
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
	}
	// Helper to expand bounds from a path
	expandPath := func(p *mp.Path) {
		if p == nil || p.Head == nil {
			return
		}
		k := p.Head
		for {
			q := k.Next
			// Include endpoints
			expand(k.XCoord, k.YCoord)
			// Include cubic extrema
			minx, maxx = bbox1D(k.XCoord, k.RightX, q.LeftX, q.XCoord, minx, maxx)
			miny, maxy = bbox1D(k.YCoord, k.RightY, q.LeftY, q.YCoord, miny, maxy)
			k = q
			if k == p.Head || k == nil || k.RType == mp.KnotEndpoint {
				expand(k.XCoord, k.YCoord)
				break
			}
		}
	}
	for _, p := range paths {
		if p == nil || p.Head == nil {
			continue
		}
		if p.Style.StrokeWidth > 0 && p.Style.StrokeWidth > maxStroke {
			maxStroke = p.Style.StrokeWidth
		}
		// For elliptical pens, use GetPenScale to determine stroke width (mp.w:11529-11547)
		if pen := p.Style.Pen; pen != nil && pen.Elliptical {
			if scale := mp.GetPenScale(pen); scale > maxStroke {
				maxStroke = scale
			}
		}
		// If there's an envelope (from non-elliptical pen), use it for bounds;
		// otherwise use the main path
		if p.Envelope != nil {
			hasEnvelope = true
			expandPath(p.Envelope)
		} else {
			expandPath(p)
		}
		// Also include arrow heads in bounds calculation
		if p.Style.Arrow.End {
			ahLen := p.Style.Arrow.Length
			ahAng := p.Style.Arrow.Angle
			if ahLen <= 0 {
				ahLen = mp.DefaultAHLength
			}
			if ahAng <= 0 {
				ahAng = mp.DefaultAHAngle
			}
			if arrow := mp.ArrowHeadEnd(p, ahLen, ahAng); arrow != nil {
				expandPath(arrow)
			}
		}
		if p.Style.Arrow.Start {
			ahLen := p.Style.Arrow.Length
			ahAng := p.Style.Arrow.Angle
			if ahLen <= 0 {
				ahLen = mp.DefaultAHLength
			}
			if ahAng <= 0 {
				ahAng = mp.DefaultAHAngle
			}
			if arrow := mp.ArrowHeadStart(p, ahLen, ahAng); arrow != nil {
				expandPath(arrow)
			}
		}
	}
	if math.IsInf(minx, 1) {
		// fallback: keep existing viewBox if nothing found
		return s
	}
	w := maxx - minx
	h := maxy - miny
	// In MetaPost-compatible mode, set viewBox to start at (0,0) like MetaPost does.
	// For paths WITHOUT envelopes, add half stroke width as padding.
	// For paths WITH envelopes, the envelope already includes the pen, so no extra padding.
	// MetaPost viewBox: (0, 0, maxX - minX, maxY - minY)
	// Y coordinates are flipped: y_svg = maxY - y_orig, x_svg = x_orig - minX
	if s.metaPostCompat {
		halfStroke := float64(0)
		if !hasEnvelope {
			halfStroke = maxStroke / 2
		}
		s.mpOffsetX = 0
		s.mpOffsetY = 0
		// viewBox dimensions: account for negative coordinates
		viewBoxMinX := minx - halfStroke
		viewBoxMaxX := maxx + halfStroke
		viewBoxMinY := miny - halfStroke
		viewBoxMaxY := maxy + halfStroke
		viewBoxW := viewBoxMaxX - viewBoxMinX
		viewBoxH := viewBoxMaxY - viewBoxMinY
		s.mpMaxY = viewBoxMaxY // Y-flip reference point
		s.mpOffsetX = viewBoxMinX
		s.viewBox = fmt.Sprintf("0 0 %g %g", viewBoxW, viewBoxH)
		if s.autoSize {
			s.width = viewBoxW
			s.height = viewBoxH
		}
		return s
	}
	halfStroke := maxStroke / 2
	totalPad := pad + halfStroke
	s.viewBox = fmt.Sprintf("%g %g %g %g", minx-totalPad, miny-totalPad, w+2*totalPad, h+2*totalPad)
	if s.autoSize {
		s.width = w + 2*totalPad
		s.height = h + 2*totalPad
	}
	return s
}

// Picture interface for adding pictures to the SVG
type Picture interface {
	Paths() []*mp.Path
	Labels() []*mp.Label
	ClipPath() *mp.Path
}

// FitViewBoxToPictures computes a viewBox over all paths in the provided pictures.
// If a picture has a clip path, the clip path's bounding box is used instead of
// the content paths (matching MetaPost's behavior where viewBox reflects visible content).
func (s *Builder) FitViewBoxToPictures(pics ...Picture) *Builder {
	var paths []*mp.Path
	for _, pic := range pics {
		if pic == nil {
			continue
		}
		// If picture has a clip path, use clip path bounds instead of content
		if clip := pic.ClipPath(); clip != nil {
			paths = append(paths, clip)
		} else {
			paths = append(paths, pic.Paths()...)
		}
	}
	return s.FitViewBoxToPaths(paths...)
}

func (s *Builder) SetBackground(color string) *Builder {
	s.bg = color
	return s
}

func (s *Builder) SetStroke(color string, width float64) *Builder {
	s.stroke = mp.ColorCSS(color)
	s.strokeWidth = width
	return s
}

// WithColor sets the stroke color using a Color helper (e.g., ColorRGB/ColorCSS).
func (s *Builder) WithColor(c mp.Color) *Builder {
	s.stroke = c
	return s
}

// AddPathWithColor adds a path overriding the stroke color with a Color helper.
func (s *Builder) AddPathWithColor(pathData string, stroke mp.Color) *Builder {
	return s.AddPath(pathData, stroke.CSS())
}

// FlipY mirrors the Y axis (MetaPost coordinates grow upward; SVG grows downward).
// Implemented as translate(y0+height) then scale(1,-1) in viewBox units.
func (s *Builder) FlipY() *Builder {
	s.flipY = true
	return s
}

// DisableFlipY turns off the default Y flip if a raw SVG coordinate space is desired.
func (s *Builder) DisableFlipY() *Builder {
	s.flipY = false
	return s
}

// MetaPostCompatible enables MetaPost-compatible SVG output. This is now the default.
// Instead of using a transform to flip Y, this mode converts coordinates directly
// in the path data (y_svg = height - y_math). This makes the output directly
// comparable to MetaPost SVG.
func (s *Builder) MetaPostCompatible() *Builder {
	s.metaPostCompat = true
	s.flipY = false // Don't use transform when in MetaPost-compatible mode
	return s
}

// DisableMetaPostCompat disables MetaPost-compatible mode and uses SVG transforms
// instead. This may be useful for specific use cases where transform-based Y-flip
// is preferred.
func (s *Builder) DisableMetaPostCompat() *Builder {
	s.metaPostCompat = false
	s.flipY = true // Use transform for Y-flip
	return s
}

func (s *Builder) AddPath(pathData string, stroke string) *Builder {
	color := s.stroke
	if stroke != "" {
		color = mp.ColorCSS(stroke)
	}
	linecap := formatLineCap(s.lineCap)
	linejoin := formatLineJoin(s.lineJoin)
	attrs := fmt.Sprintf(`fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="%s" stroke-linejoin="%s"`,
		s.fill.CSS(), color.CSS(), s.strokeWidth, linecap, linejoin)
	if op, ok := color.Opacity(); ok {
		attrs += fmt.Sprintf(` stroke-opacity="%.3f"`, op)
	}
	if op, ok := s.fill.Opacity(); ok {
		attrs += fmt.Sprintf(` fill-opacity="%.3f"`, op)
	}
	s.paths = append(s.paths, fmt.Sprintf(
		`<path d="%s" %s/>`,
		pathData, attrs))
	return s
}

// AddPathFromPath renders an mp.Path using its Style (if present), falling
// back to the builder defaults.
func (s *Builder) AddPathFromPath(p *mp.Path) *Builder {
	if p == nil {
		return s
	}
	// If an envelope was precomputed, render that instead
	if p.Envelope != nil {
		// Store original path for auto viewBox calculation (includes envelope info)
		s.mpOrigPaths = append(s.mpOrigPaths, p)
		// Envelope is a filled shape representing the stroked path.
		// MetaPost renders envelopes with fill only, no stroke (stroke: none).
		envelope := p.Envelope
		envelope.Style.Arrow = p.Style.Arrow
		envelope.Style.Fill = p.Style.Stroke        // Fill with the stroke color
		envelope.Style.Stroke = mp.ColorCSS("none") // No SVG stroke on envelope
		return s.AddPathFromPath(envelope)
	}
	// For MetaPost-compatible mode, store paths and defer rendering to WriteTo
	if s.metaPostCompat {
		// Store original path for auto viewBox calculation (no envelope case)
		s.mpOrigPaths = append(s.mpOrigPaths, p)
		// Determine arrow lengths for shortening
		var shortenStart, shortenEnd mp.Number
		ahLenEnd := p.Style.Arrow.Length
		ahAngEnd := p.Style.Arrow.Angle
		if ahLenEnd <= 0 {
			ahLenEnd = mp.DefaultAHLength
		}
		if ahAngEnd <= 0 {
			ahAngEnd = mp.DefaultAHAngle
		}
		ahLenStart := ahLenEnd
		ahAngStart := ahAngEnd

		// The arrow base is at distance ahlength * cos(ahangle/2) from the apex
		// (not the full ahlength, which is the distance to the base corners)
		halfAngleRad := ahAngEnd * 3.14159265358979323846 / 360.0
		cosHalfAngle := math.Cos(halfAngleRad)

		if p.Style.Arrow.End {
			shortenEnd = ahLenEnd * mp.Number(cosHalfAngle)
		}
		if p.Style.Arrow.Start {
			shortenStart = ahLenStart * mp.Number(cosHalfAngle)
		}

		// If arrows are present, shorten the path (like MetaPost's cutafter)
		pathToAdd := p
		if shortenStart > 0 || shortenEnd > 0 {
			shortened := mp.ShortenPathForArrow(p, shortenStart, shortenEnd)
			if shortened != nil {
				pathToAdd = shortened
			}
		}

		s.mpPaths = append(s.mpPaths, pathToAdd)
		// Note: mpMinY/mpMaxY should already be set by FitViewBoxToPaths.
		// We don't update them here to avoid overwriting the correctly
		// calculated values that include stroke padding.
		// Also add arrowhead paths if arrows are enabled
		if p.Style.Arrow.End {
			if arrow := mp.ArrowHeadEnd(p, ahLenEnd, ahAngEnd); arrow != nil {
				arrow.Style.Fill = p.Style.Stroke
				arrow.Style.Stroke = mp.ColorCSS("none") // Filled, no stroke
				s.mpPaths = append(s.mpPaths, arrow)
			}
		}
		if p.Style.Arrow.Start {
			if arrow := mp.ArrowHeadStart(p, ahLenStart, ahAngStart); arrow != nil {
				arrow.Style.Fill = p.Style.Stroke
				arrow.Style.Stroke = mp.ColorCSS("none") // Filled, no stroke
				s.mpPaths = append(s.mpPaths, arrow)
			}
		}
		return s // Don't add to s.paths yet; will be rendered in WriteTo
	}
	pathData := PathToSVG(p)
	color := s.stroke
	fill := s.fill
	width := s.strokeWidth
	linecap := "round"
	linejoin := "round"
	pen := (*mp.Pen)(nil)
	dash := p.Style.Dash
	if p.Style.Stroke.CSS() != "" {
		color = p.Style.Stroke
	}
	if p.Style.Fill.CSS() != "" {
		fill = p.Style.Fill
	}
	if p.Style.StrokeWidth > 0 {
		width = p.Style.StrokeWidth
	}
	if p.Style.Pen != nil {
		pen = p.Style.Pen
	}
	if color.CSS() == "none" {
		width = 0
	}

	// If an elliptical pen is present (single explicit knot, mp.w:10439 pen_is_elliptical),
	// compute stroke-width using GetPenScale (mp.w:11529-11547 mp_get_pen_scale).
	if pen != nil && pen.Elliptical && pen.Head != nil && pen.Head.Next == pen.Head {
		// mp.w:11529-11547: GetPenScale returns sqrt(|det(transformation)|)
		// For an untransformed pencircle of diameter d, this equals d.
		scale := mp.GetPenScale(pen)
		if scale > 0 {
			width = scale
		}
		attrs := fmt.Sprintf(`fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="%s" stroke-linejoin="%s"`,
			fill.CSS(), color.CSS(), width, linecap, linejoin)
		if op, ok := color.Opacity(); ok {
			attrs += fmt.Sprintf(` stroke-opacity="%.3f"`, op)
		}
		if op, ok := fill.Opacity(); ok {
			attrs += fmt.Sprintf(` fill-opacity="%.3f"`, op)
		}
		attrs += FormatDashAttrs(dash)
		s.paths = append(s.paths, fmt.Sprintf(
			`<path d="%s" %s/>`,
			pathData, attrs))
		return s
	}

	// Non-elliptical pen: fall back to stroking with a width derived from the
	// pen's bounding box (mp_pen_bbox analogue, mp.c:10670ff) and square caps/
	// miter joins to better approximate a box pen (mp_offset_prep/mp_apply_offset).
	if pen != nil && !pen.Elliptical {
		if minx, miny, maxx, maxy, ok := mp.PenBBox(pen); ok {
			pw := maxx - minx
			ph := maxy - miny
			if pw < ph {
				pw = ph
			}
			if pw > 0 {
				width = pw
				linecap = "square"
				linejoin = "miter"
			}
		}
	}

	attrs := fmt.Sprintf(`fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="%s" stroke-linejoin="%s"`,
		fill.CSS(), color.CSS(), width, linecap, linejoin)
	if op, ok := color.Opacity(); ok {
		attrs += fmt.Sprintf(` stroke-opacity="%.3f"`, op)
	}
	if op, ok := fill.Opacity(); ok {
		attrs += fmt.Sprintf(` fill-opacity="%.3f"`, op)
	}
	attrs += FormatDashAttrs(dash)
	s.paths = append(s.paths, fmt.Sprintf(
		`<path d="%s" %s/>`,
		pathData, attrs))
	return s
}

// AddPicture renders every path stored in the picture using their Style (if set),
// mirroring how MetaPost pictures collect edges before backend output.
// If the picture has a clip path set, all paths will be clipped to that boundary.
func (s *Builder) AddPicture(pic Picture) *Builder {
	if pic == nil {
		return s
	}

	// Handle clipping
	if clip := pic.ClipPath(); clip != nil {
		// Add clip path to the list and create a clipped group
		clipIndex := len(s.clipPaths)
		s.clipPaths = append(s.clipPaths, clip)
		s.clippedGroups = append(s.clippedGroups, clippedGroup{
			clipIndex: clipIndex,
			paths:     pic.Paths(),
		})
		// Add labels (not clipped for now, matching MetaPost behavior)
		s.labels = append(s.labels, pic.Labels()...)
		return s
	}

	// No clipping - add paths directly
	for _, p := range pic.Paths() {
		s.AddPathFromPath(p)
	}
	// Add labels
	s.labels = append(s.labels, pic.Labels()...)
	return s
}

// fitViewBoxToContent computes the viewBox including both paths and labels.
func (s *Builder) fitViewBoxToContent() {
	s.viewBoxSet = true
	pad := s.padding
	minx, miny := math.Inf(1), math.Inf(1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)
	maxStroke := s.strokeWidth
	hasEnvelope := false

	expand := func(x, y float64) {
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
	}

	// Include paths (same logic as FitViewBoxToPaths)
	for _, p := range s.mpOrigPaths {
		if p == nil || p.Head == nil {
			continue
		}
		if p.Style.StrokeWidth > 0 && p.Style.StrokeWidth > maxStroke {
			maxStroke = p.Style.StrokeWidth
		}
		if pen := p.Style.Pen; pen != nil && pen.Elliptical {
			if scale := mp.GetPenScale(pen); scale > maxStroke {
				maxStroke = scale
			}
		}
		if p.Envelope != nil {
			hasEnvelope = true
			lminX, lminY, lmaxX, lmaxY := PathBBox(p.Envelope)
			expand(lminX, lminY)
			expand(lmaxX, lmaxY)
		} else {
			lminX, lminY, lmaxX, lmaxY := PathBBox(p)
			expand(lminX, lminY)
			expand(lmaxX, lmaxY)
		}
		// Include arrow heads
		if p.Style.Arrow.End {
			ahLen := p.Style.Arrow.Length
			if ahLen <= 0 {
				ahLen = mp.DefaultAHLength
			}
			ahAng := p.Style.Arrow.Angle
			if ahAng <= 0 {
				ahAng = mp.DefaultAHAngle
			}
			if arrow := mp.ArrowHeadEnd(p, ahLen, ahAng); arrow != nil {
				lminX, lminY, lmaxX, lmaxY := PathBBox(arrow)
				expand(lminX, lminY)
				expand(lmaxX, lmaxY)
			}
		}
		if p.Style.Arrow.Start {
			ahLen := p.Style.Arrow.Length
			if ahLen <= 0 {
				ahLen = mp.DefaultAHLength
			}
			ahAng := p.Style.Arrow.Angle
			if ahAng <= 0 {
				ahAng = mp.DefaultAHAngle
			}
			if arrow := mp.ArrowHeadStart(p, ahLen, ahAng); arrow != nil {
				lminX, lminY, lmaxX, lmaxY := PathBBox(arrow)
				expand(lminX, lminY)
				expand(lmaxX, lmaxY)
			}
		}
	}

	// Include clipped groups (use clip path bounds, like MetaPost)
	for _, cg := range s.clippedGroups {
		if cg.clipIndex >= 0 && cg.clipIndex < len(s.clipPaths) {
			clip := s.clipPaths[cg.clipIndex]
			if clip != nil && clip.Head != nil {
				lminX, lminY, lmaxX, lmaxY := PathBBox(clip)
				expand(lminX, lminY)
				expand(lmaxX, lmaxY)
			}
		}
	}

	// Include labels
	for _, label := range s.labels {
		if label == nil {
			continue
		}
		lminX, lminY, lmaxX, lmaxY := label.EstimateBounds()
		expand(lminX, lminY)
		expand(lmaxX, lmaxY)
	}

	if math.IsInf(minx, 1) {
		return // No content
	}

	// Calculate viewBox with stroke padding
	halfStroke := float64(0)
	if !hasEnvelope {
		halfStroke = maxStroke / 2
	}
	halfStroke += pad

	viewBoxMinX := minx - halfStroke
	viewBoxMaxX := maxx + halfStroke
	viewBoxMinY := miny - halfStroke
	viewBoxMaxY := maxy + halfStroke
	viewBoxW := viewBoxMaxX - viewBoxMinX
	viewBoxH := viewBoxMaxY - viewBoxMinY

	s.mpMaxY = viewBoxMaxY
	s.mpOffsetX = viewBoxMinX
	s.viewBox = fmt.Sprintf("0 0 %g %g", viewBoxW, viewBoxH)
	if s.autoSize {
		s.width = viewBoxW
		s.height = viewBoxH
	}
}

func (s *Builder) WriteTo(w io.Writer) error {
	// Auto-fit viewBox if not explicitly set and we have content
	if !s.viewBoxSet && (len(s.mpOrigPaths) > 0 || len(s.labels) > 0 || len(s.clippedGroups) > 0) {
		s.fitViewBoxToContent()
	}
	vb := s.viewBox
	if vb == "" {
		vb = fmt.Sprintf("0 0 %g %g", s.width, s.height)
	}
	if s.autoSize {
		if _, err := fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%s">`, vb); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="%g" height="%g" viewBox="%s">`, s.width, s.height, vb); err != nil {
		return err
	}

	// Write clip path definitions if any
	if len(s.clipPaths) > 0 {
		if _, err := io.WriteString(w, "<defs>"); err != nil {
			return err
		}
		for i, clipPath := range s.clipPaths {
			pathData := PathToSVG(clipPath)
			if s.metaPostCompat {
				pathData = PathToSVGTransformed(clipPath, s.mpOffsetX, s.mpOffsetY, s.mpMaxY)
			}
			if _, err := fmt.Fprintf(w, `<clipPath id="clip%d"><path d="%s"/></clipPath>`, i, pathData); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "</defs>"); err != nil {
			return err
		}
	}

	if s.bg != "" {
		if _, err := fmt.Fprintf(w, `<rect x="0" y="0" width="100%%" height="100%%" fill="%s"/>`, s.bg); err != nil {
			return err
		}
	}
	if s.flipY {
		var vx, vy, vw, vh float64
		_, _ = fmt.Sscanf(vb, "%f %f %f %f", &vx, &vy, &vw, &vh)
		// Flip around the vertical midpoint so [minY,maxY] stays in view.
		ty := (2 * vy) + vh // minY + maxY
		if _, err := fmt.Fprintf(w, `<g transform="translate(0 %g) scale(1 -1)">`, ty); err != nil {
			return err
		}
	}

	// Render clipped groups
	for _, group := range s.clippedGroups {
		if _, err := fmt.Fprintf(w, `<g clip-path="url(#clip%d)">`, group.clipIndex); err != nil {
			return err
		}
		for _, p := range group.paths {
			if err := s.writePathElement(w, p); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "</g>"); err != nil {
			return err
		}
	}

	// MetaPost-compatible mode: render paths with transformed coordinates
	if s.metaPostCompat && len(s.mpPaths) > 0 {
		// MetaPost uses y_svg = maxY - y_orig and shifts X so viewBox starts at (0,0).
		// We apply the same transformation using stored offsets.
		flipHeight := s.mpMaxY
		for _, p := range s.mpPaths {
			pathData := PathToSVGTransformed(p, s.mpOffsetX, s.mpOffsetY, flipHeight)
			fill := s.fill
			color := s.stroke
			if p.Style.Fill.CSS() != "" {
				fill = p.Style.Fill
			}
			if p.Style.Stroke.CSS() != "" {
				color = p.Style.Stroke
			}
			if color.CSS() == "none" {
				if _, err := fmt.Fprintf(w, `<path d="%s" fill="%s" stroke="none"/>`, pathData, fill.CSS()); err != nil {
					return err
				}
			} else {
				width := s.strokeWidth
				if p.Style.StrokeWidth > 0 {
					width = p.Style.StrokeWidth
				}
				// For elliptical pens, use the pen's scale as stroke width
				if pen := p.Style.Pen; pen != nil && pen.Elliptical {
					if scale := mp.GetPenScale(pen); scale > 0 {
						width = scale
					}
				}
				dashAttrs := FormatDashAttrs(p.Style.Dash)
				linecap := formatLineCap(p.Style.LineCap)
				linejoin := formatLineJoin(p.Style.LineJoin)
				if _, err := fmt.Fprintf(w, `<path d="%s" fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="%s" stroke-linejoin="%s"%s/>`,
					pathData, fill.CSS(), color.CSS(), width, linecap, linejoin, dashAttrs); err != nil {
					return err
				}
			}
		}
	}
	for _, p := range s.paths {
		if _, err := io.WriteString(w, p); err != nil {
			return err
		}
	}
	// Render labels
	for _, label := range s.labels {
		if err := s.writeLabelElement(w, label); err != nil {
			return err
		}
	}
	if s.flipY {
		if _, err := io.WriteString(w, "</g>"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</svg>\n")
	return err
}

// writePathElement writes a single path element to the SVG output.
func (s *Builder) writePathElement(w io.Writer, p *mp.Path) error {
	pathData := PathToSVG(p)
	if s.metaPostCompat {
		pathData = PathToSVGTransformed(p, s.mpOffsetX, s.mpOffsetY, s.mpMaxY)
	}
	fill := s.fill
	color := s.stroke
	if p.Style.Fill.CSS() != "" {
		fill = p.Style.Fill
	}
	if p.Style.Stroke.CSS() != "" {
		color = p.Style.Stroke
	}
	if color.CSS() == "none" {
		_, err := fmt.Fprintf(w, `<path d="%s" fill="%s" stroke="none"/>`, pathData, fill.CSS())
		return err
	}
	width := s.strokeWidth
	if p.Style.StrokeWidth > 0 {
		width = p.Style.StrokeWidth
	}
	// For elliptical pens, use the pen's scale as stroke width
	if pen := p.Style.Pen; pen != nil && pen.Elliptical {
		if scale := mp.GetPenScale(pen); scale > 0 {
			width = scale
		}
	}
	dashAttrs := FormatDashAttrs(p.Style.Dash)
	linecap := formatLineCap(p.Style.LineCap)
	linejoin := formatLineJoin(p.Style.LineJoin)
	_, err := fmt.Fprintf(w, `<path d="%s" fill="%s" stroke="%s" stroke-width="%.2f" stroke-linecap="%s" stroke-linejoin="%s"%s/>`,
		pathData, fill.CSS(), color.CSS(), width, linecap, linejoin, dashAttrs)
	return err
}

// AddLabel adds a label to the SVG output.
func (s *Builder) AddLabel(label *mp.Label) *Builder {
	if label != nil {
		s.labels = append(s.labels, label)
	}
	return s
}

// formatTextAnchor returns the SVG text-anchor value for a given Anchor.
// Maps MetaPost's labxf values to SVG text-anchor:
//   - labxf=0 (left edge at position) → text-anchor="start"
//   - labxf=0.5 (center at position) → text-anchor="middle"
//   - labxf=1 (right edge at position) → text-anchor="end"
func formatTextAnchor(anchor mp.Anchor) string {
	xf, _ := mp.LabelAnchorFactors(anchor)
	if xf < 0.25 {
		return "start"
	} else if xf > 0.75 {
		return "end"
	}
	return "middle"
}

// formatDominantBaseline returns the SVG dominant-baseline value for a given Anchor.
// Maps MetaPost's labyf values to SVG dominant-baseline:
//   - labyf=0 (bottom edge of text at position) → dominant-baseline="text-after-edge"
//   - labyf=0.5 (middle at position) → dominant-baseline="central"
//   - labyf=1 (top edge of text at position) → dominant-baseline="hanging"
func formatDominantBaseline(anchor mp.Anchor) string {
	_, yf := mp.LabelAnchorFactors(anchor)
	if yf < 0.25 {
		// yf=0: bottom edge of text at position (e.g., label.top places text above point)
		return "text-after-edge"
	} else if yf > 0.75 {
		// yf=1: top edge of text at position (e.g., label.bot places text below point)
		return "hanging"
	}
	return "central"
}

// writeLabelElement writes a single label as an SVG text element.
func (s *Builder) writeLabelElement(w io.Writer, label *mp.Label) error {
	if label == nil {
		return nil
	}

	// Calculate the position with offset
	dx, dy := mp.LabelOffsetVector(label.Anchor)
	offset := label.LabelOffset
	if offset == 0 {
		offset = mp.DefaultLabelOffset
	}
	x := label.Position.X + dx*offset
	y := label.Position.Y + dy*offset

	// Transform coordinates for MetaPost-compatible mode
	if s.metaPostCompat {
		x = x - s.mpOffsetX
		y = s.mpMaxY - y + s.mpOffsetY
	}

	// Get SVG text attributes
	textAnchor := formatTextAnchor(label.Anchor)
	dominantBaseline := formatDominantBaseline(label.Anchor)

	// Get font settings
	fontSize := label.FontSize
	if fontSize == 0 {
		fontSize = mp.DefaultFontSize
	}
	fontFamily := label.FontFamily
	if fontFamily == "" {
		fontFamily = "sans-serif"
	}

	// Get color
	color := label.Color
	if color.CSS() == "" {
		color = mp.ColorCSS("black")
	}

	// Write the text element
	_, err := fmt.Fprintf(w, `<text x="%.3f" y="%.3f" font-family="%s" font-size="%.2f" fill="%s" text-anchor="%s" dominant-baseline="%s">%s</text>`,
		x, y, fontFamily, fontSize, color.CSS(), textAnchor, dominantBaseline, escapeXML(label.Text))
	return err
}

// escapeXML escapes special characters for XML/SVG text content.
func escapeXML(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '&':
			b.WriteString("&amp;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
