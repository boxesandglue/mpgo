package mp

import (
	"fmt"
	"strings"
)

// LineCap constants. Values are offset by 1 so that Go's zero value (0)
// means "unset" and defaults to LineCapRounded (matching MetaPost's default).
// Internal MetaPost values: butt=0, rounded=1, squared=2.
const (
	LineCapDefault = 0 // Unset - uses MetaPost default (rounded)
	LineCapButt    = 1 // MetaPost linecap 0
	LineCapRounded = 2 // MetaPost linecap 1 (MetaPost default)
	LineCapSquared = 3 // MetaPost linecap 2
)

// LineJoin constants.
// Note: MetaPost's default is linejoin=1 (rounded), not 0 (mitered).
// Go's zero value means "unset", which defaults to rounded (MetaPost behavior).
const (
	LineJoinDefault = 0 // Unset - uses MetaPost default (rounded)
	LineJoinMiter   = 1 // MetaPost linejoin 0
	LineJoinRound   = 2 // MetaPost linejoin 1 (MetaPost default)
	LineJoinBevel   = 3 // MetaPost linejoin 2
)

// Arrow constants (MetaPost defaults from plain.mp)
const (
	DefaultAHLength = 4.0  // default arrowhead length (4bp)
	DefaultAHAngle  = 45.0 // default arrowhead angle (45 degrees)
)

// ArrowStyle defines arrow head appearance.
type ArrowStyle struct {
	Start  bool   // arrow at start of path (for drawdblarrow)
	End    bool   // arrow at end of path (for drawarrow)
	Length Number // ahlength - arrow head length
	Angle  Number // ahangle - arrow head angle in degrees
}

// DashPattern represents a dash pattern for stroked paths.
// Mirrors MetaPost's picture-based dash pattern (plain.mp dashpattern macro).
//
// In MetaPost, a dash pattern is a picture containing horizontal line segments
// where the y-coordinate encodes the cumulative position (total pattern length).
// The pattern is built using "on" (visible) and "off" (gap) segments.
//
// Internal structure (mp.w:11778ff):
//   - dash_node: start_x, stop_x, dash_y (period)
//   - mp_export_dashes converts to offset + array[] for SVG output
//
// Example: dashpattern(on 3 off 3) creates evenly spaced dashes.
type DashPattern struct {
	// Array contains alternating on/off lengths: [on1, off1, on2, off2, ...]
	// This mirrors mp_dash_object.array from psout.w:5219ff
	Array []float64
	// Offset is the starting offset into the pattern (for phase shifting)
	// Mirrors mp_dash_object.offset
	Offset float64
}

// NewDashPattern creates a dash pattern from alternating on/off lengths.
// Example: NewDashPattern(3, 3) creates "on 3 off 3" (evenly spaced dashes)
func NewDashPattern(onOff ...float64) *DashPattern {
	if len(onOff) == 0 {
		return nil
	}
	return &DashPattern{Array: onOff}
}

// Evenly returns the standard "evenly" dash pattern (on 3 off 3).
// This is the MetaPost default: dashpattern(on 3 off 3)
func DashEvenly() *DashPattern {
	return &DashPattern{Array: []float64{3, 3}}
}

// WithDots returns the "withdots" dash pattern (off 2.5 on 0 off 2.5).
// In MetaPost: dashpattern(off 2.5 on 0 off 2.5)
// Note: "on 0" creates a dot when linecap is round.
func DashWithDots() *DashPattern {
	// MetaPost withdots is: off 2.5 on 0 off 2.5
	// This means: start with gap, then zero-length dash (dot), then gap
	// For SVG, we need to represent this carefully.
	// The pattern starts with "off", so we need offset to skip the first "on"
	return &DashPattern{
		Array:  []float64{0, 5}, // on 0, off 5 (2.5 + 2.5)
		Offset: 2.5,             // start 2.5 into the pattern
	}
}

// Scaled returns a new dash pattern with all values multiplied by factor.
// Mirrors MetaPost's "dashed evenly scaled 2" syntax.
func (d *DashPattern) Scaled(factor float64) *DashPattern {
	if d == nil {
		return nil
	}
	result := &DashPattern{
		Array:  make([]float64, len(d.Array)),
		Offset: d.Offset * factor,
	}
	for i, v := range d.Array {
		result.Array[i] = v * factor
	}
	return result
}

// Shifted returns a new dash pattern with the offset adjusted.
// Mirrors MetaPost's phase shifting.
func (d *DashPattern) Shifted(offset float64) *DashPattern {
	if d == nil {
		return nil
	}
	result := &DashPattern{
		Array:  make([]float64, len(d.Array)),
		Offset: d.Offset + offset,
	}
	copy(result.Array, d.Array)
	return result
}

// Style holds drawing attributes attached to a path.
type Style struct {
	Stroke      Color
	StrokeWidth float64
	Fill        Color
	Pen         *Pen // mirrors pen_p in mp.c (mp.c:564)
	// LineJoin/LineCap mirror MetaPost linejoin/linecap (mp.c:23894ff).
	// LineCap uses offset constants (0=default/unset → rounded).
	// LineJoin uses direct MetaPost values (0=miter, 1=round, 2=bevel).
	LineJoin int
	LineCap  int // Use LineCapButt, LineCapRounded, LineCapSquared constants
	Arrow    ArrowStyle
	Dash     *DashPattern // dash pattern for stroked paths (mp.w:11362ff)
}

type Path struct {
	Head     *Knot
	Style    Style
	Envelope *Path // optional precomputed offset/envelope (mp_apply_offset analogue)
}

func (p *Path) String() string {
	if p == nil || p.Head == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "(%.5g,%.5g)", p.Head.XCoord, p.Head.YCoord)
	cur := p.Head
	for {
		next := cur.Next
		if next == nil {
			break
		}
		fmt.Fprintf(&b, "..controls (%.5g,%.5g) and (%.5g,%.5g)\n..(%.5g,%.5g)",
			cur.RightX, cur.RightY, next.LeftX, next.LeftY, next.XCoord, next.YCoord)
		cur = next
		if cur == p.Head || cur.RType == KnotEndpoint {
			break
		}
	}
	if cur != nil && cur.RType != KnotEndpoint {
		b.WriteString("\n..cycle")
	}
	return b.String()
}

func NewPath() *Path {
	return &Path{}
}

func (p *Path) Append(k *Knot) {
	if p.Head == nil {
		p.Head = k
		k.Next = k
		k.Prev = k
		return
	}
	tail := p.Head.Prev
	tail.Next = k
	k.Prev = tail
	k.Next = p.Head
	p.Head.Prev = k
}

func (p *Path) Copy() *Path {
	if p == nil || p.Head == nil {
		return &Path{}
	}
	q := NewPath()
	cur := p.Head
	for {
		q.Append(CopyKnot(cur))
		cur = cur.Next
		if cur == nil || cur == p.Head {
			break
		}
	}
	q.Style = p.Style
	if p.Envelope != nil {
		q.Envelope = p.Envelope.Copy()
	}
	return q
}

// ShortenPathForArrow creates a copy of path p with endpoints moved inward
// to make room for arrowheads. This mimics MetaPost's "cutafter" behavior.
// shortenStart/shortenEnd specify how much to shorten at each end.
func ShortenPathForArrow(p *Path, shortenStart, shortenEnd Number) *Path {
	if p == nil || p.Head == nil {
		return nil
	}
	// Make a copy of the path
	q := p.Copy()
	if q.Head == nil {
		return q
	}

	// Shorten at end
	if shortenEnd > 0 {
		// Find the last knot
		last := q.Head
		for last.Next != nil && last.Next != q.Head {
			last = last.Next
		}
		// Get direction at the end
		dx := last.XCoord - last.LeftX
		dy := last.YCoord - last.LeftY
		if dx == 0 && dy == 0 && last.Prev != nil {
			dx = last.XCoord - last.Prev.XCoord
			dy = last.YCoord - last.Prev.YCoord
		}
		length := sqrtNumber(dx*dx + dy*dy)
		if length > 0.0001 {
			dx /= length
			dy /= length
			// Move endpoint back
			last.XCoord -= dx * shortenEnd
			last.YCoord -= dy * shortenEnd
			// Also adjust control point
			last.LeftX -= dx * shortenEnd
			last.LeftY -= dy * shortenEnd
		}
	}

	// Shorten at start
	if shortenStart > 0 {
		start := q.Head
		// Get direction at the start
		dx := start.RightX - start.XCoord
		dy := start.RightY - start.YCoord
		if dx == 0 && dy == 0 && start.Next != nil {
			dx = start.Next.XCoord - start.XCoord
			dy = start.Next.YCoord - start.YCoord
		}
		length := sqrtNumber(dx*dx + dy*dy)
		if length > 0.0001 {
			dx /= length
			dy /= length
			// Move startpoint forward
			start.XCoord += dx * shortenStart
			start.YCoord += dy * shortenStart
			// Also adjust control point
			start.RightX += dx * shortenStart
			start.RightY += dy * shortenStart
		}
	}

	return q
}

// ArrowHeadEnd creates an arrowhead path at the end of path p.
// The arrowhead is a filled triangle with apex at the endpoint.
// Uses ahLength for the arrow length and ahAngle for the head angle (degrees).
func ArrowHeadEnd(p *Path, ahLength, ahAngle Number) *Path {
	if p == nil || p.Head == nil {
		return nil
	}
	// Find the last knot
	last := p.Head
	for last.Next != nil && last.Next != p.Head {
		last = last.Next
	}
	// Get the direction at the end (from second-to-last control point to endpoint)
	// For a cubic bezier ending at 'last', the direction is from last.LeftX/Y to last
	dx := last.XCoord - last.LeftX
	dy := last.YCoord - last.LeftY
	// If control point equals endpoint (degenerate), use previous knot
	if dx == 0 && dy == 0 {
		prev := last.Prev
		if prev != nil {
			dx = last.XCoord - prev.XCoord
			dy = last.YCoord - prev.YCoord
		}
	}
	// Normalize direction
	length := sqrtNumber(dx*dx + dy*dy)
	if length < 0.0001 {
		return nil
	}
	dx /= length
	dy /= length

	return createArrowHead(last.XCoord, last.YCoord, dx, dy, ahLength, ahAngle)
}

// ArrowHeadStart creates an arrowhead path at the start of path p.
// The arrowhead is a filled triangle with apex at the start point.
func ArrowHeadStart(p *Path, ahLength, ahAngle Number) *Path {
	if p == nil || p.Head == nil {
		return nil
	}
	start := p.Head
	// Get the direction at the start (from start to first control point)
	dx := start.RightX - start.XCoord
	dy := start.RightY - start.YCoord
	// If control point equals endpoint (degenerate), use next knot
	if dx == 0 && dy == 0 && start.Next != nil {
		dx = start.Next.XCoord - start.XCoord
		dy = start.Next.YCoord - start.YCoord
	}
	// Normalize and reverse direction (arrow points toward start)
	length := sqrtNumber(dx*dx + dy*dy)
	if length < 0.0001 {
		return nil
	}
	dx /= length
	dy /= length
	// Reverse direction so arrow points inward
	dx = -dx
	dy = -dy

	return createArrowHead(start.XCoord, start.YCoord, dx, dy, ahLength, ahAngle)
}

// createArrowHead creates a triangular arrowhead path.
// (tipX, tipY) is the apex, (dx, dy) is the unit direction vector pointing
// toward the tip, ahLength is the arrow length, ahAngle is the full angle in degrees.
func createArrowHead(tipX, tipY, dx, dy, ahLength, ahAngle Number) *Path {
	// Convert angle to radians (half angle for each side)
	halfAngle := ahAngle * 3.14159265358979323846 / 360.0

	// Calculate the base points
	// Rotate (-dx, -dy) by ±halfAngle and scale by ahLength
	cos := cosNumber(halfAngle)
	sin := sinNumber(halfAngle)

	// Base direction (opposite to arrow direction)
	bx := -dx
	by := -dy

	// Left base point: rotate by +halfAngle
	lx := bx*cos - by*sin
	ly := bx*sin + by*cos
	leftX := tipX + lx*ahLength
	leftY := tipY + ly*ahLength

	// Right base point: rotate by -halfAngle
	rx := bx*cos + by*sin
	ry := -bx*sin + by*cos
	rightX := tipX + rx*ahLength
	rightY := tipY + ry*ahLength

	// Create triangular path: left -> tip -> right -> cycle
	arrow := NewPath()
	arrow.Append(&Knot{XCoord: leftX, YCoord: leftY, LType: KnotExplicit, RType: KnotExplicit})
	arrow.Append(&Knot{XCoord: tipX, YCoord: tipY, LType: KnotExplicit, RType: KnotExplicit})
	arrow.Append(&Knot{XCoord: rightX, YCoord: rightY, LType: KnotExplicit, RType: KnotExplicit})

	// Set control points to make straight lines
	k := arrow.Head
	for {
		next := k.Next
		if next == nil {
			break
		}
		k.RightX = k.XCoord
		k.RightY = k.YCoord
		next.LeftX = next.XCoord
		next.LeftY = next.YCoord
		k = next
		if k == arrow.Head {
			break
		}
	}

	return arrow
}
