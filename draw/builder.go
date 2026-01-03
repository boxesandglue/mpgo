package draw

import (
	"github.com/boxesandglue/mpgo/mp"
)

// Point is an alias for mp.Point for convenience in the fluent API.
type Point = mp.Point

// P creates a Point from x, y coordinates.
// This is a convenience re-export of mp.P for use in fluent path building.
func P(x, y float64) Point {
	return mp.P(x, y)
}

// PathBuilder constructs a path using outgoing/incoming directions and delegates solving to mp.Engine.
type PathBuilder struct {
	ctx             *Context // optional equation context for variable resolution
	startSet        bool
	start           mp.Point
	startVar        *Var // alternative: start from a Var
	outDir          float64
	inDir           float64
	outTension      float64
	inTension       float64
	outCurl         float64
	inCurl          float64
	outSet          bool
	inSet           bool
	outTSet         bool
	inTSet          bool
	outCurlSet      bool
	inCurlSet       bool
	segments        []segment
	closed          bool
	closeIn         float64
	closeOut        float64
	closeInSet      bool
	closeOutSet     bool
	closeCurlIn     float64
	closeCurlOut    float64
	closeCurlInSet  bool
	closeCurlOutSet bool
	stroke          mp.Color
	fill            mp.Color
	strokeWidth     float64
	pen             *mp.Pen
	lineJoin        int
	lineCap         int
	arrowEnd        bool
	arrowStart      bool
	arrowLength     float64
	arrowAngle      float64
	dash            *mp.DashPattern
	transforms      []mp.Transform // transformations to apply after solving
	styleSet        bool
}

type segment struct {
	to     mp.Point
	toVar  *Var // alternative: destination from a Var
	outDir float64
	inDir  float64
	outSet bool
	inSet  bool
	line   bool
	// tension per end; if unset default is 1.
	outTension float64
	inTension  float64
	outTSet    bool
	inTSet     bool
	outCurl    float64
	inCurl     float64
	outCurlSet bool
	inCurlSet  bool
	// explicit controls skip solving.
	explicit bool
	ctrl1    mp.Point
	ctrl2    mp.Point
}

func NewPath() *PathBuilder {
	return &PathBuilder{
		segments:    make([]segment, 0),
		outTension:  1,
		inTension:   1,
		strokeWidth: 0.5, // MetaPost default: pencircle scaled 0.5pt
	}
}

// WithContext links this path builder to an equation context.
// Variables used in MoveToVar/CurveToVar/LineToVar will be resolved
// when the context is solved.
func (p *PathBuilder) WithContext(ctx *Context) *PathBuilder {
	p.ctx = ctx
	return p
}

func (p *PathBuilder) MoveTo(pt mp.Point) *PathBuilder {
	p.startSet = true
	p.start = pt
	p.startVar = nil
	return p
}

// MoveToVar sets the start point from a context variable.
// The variable must be resolved (via ctx.Solve()) before BuildPath is called.
func (p *PathBuilder) MoveToVar(v *Var) *PathBuilder {
	p.startSet = true
	p.startVar = v
	return p
}

// WithDirection sets the outgoing direction in degrees for the next segment.
func (p *PathBuilder) WithDirection(deg float64) *PathBuilder {
	p.outDir = deg
	p.outSet = true
	return p
}

// WithIncomingDirection sets the incoming direction in degrees for the next segment.
func (p *PathBuilder) WithIncomingDirection(deg float64) *PathBuilder {
	p.inDir = deg
	p.inSet = true
	return p
}

// WithCurl sets both outgoing and incoming curl for the next segment.
func (p *PathBuilder) WithCurl(c float64) *PathBuilder {
	p.outCurl = c
	p.inCurl = c
	p.outCurlSet = true
	p.inCurlSet = true
	return p
}

// WithOutgoingCurl sets outgoing curl for the next segment.
func (p *PathBuilder) WithOutgoingCurl(c float64) *PathBuilder {
	p.outCurl = c
	p.outCurlSet = true
	return p
}

// WithIncomingCurl sets incoming curl for the next segment.
func (p *PathBuilder) WithIncomingCurl(c float64) *PathBuilder {
	p.inCurl = c
	p.inCurlSet = true
	return p
}

// WithTension sets both outgoing and incoming tension for the next segment.
func (p *PathBuilder) WithTension(t float64) *PathBuilder {
	p.outTension = t
	p.inTension = t
	p.outTSet = true
	p.inTSet = true
	return p
}

// WithTensionAtLeast mirrors MetaPost's "tension atleast t"; currently treated
// by storing a negative tension to signal "atleast" to the solver (mp.c uses
// negative values for the flag).
func (p *PathBuilder) WithTensionAtLeast(t float64) *PathBuilder {
	p.outTension = -t
	p.inTension = -t
	p.outTSet = true
	p.inTSet = true
	return p
}

// WithTensionInfinity mirrors MetaPost's "tension infinity"; mapped to a large
// tension value.
func (p *PathBuilder) WithTensionInfinity() *PathBuilder {
	p.outTension = mp.Inf()
	p.inTension = mp.Inf()
	p.outTSet = true
	p.inTSet = true
	return p
}

// WithOutgoingTension sets outgoing tension for the next segment.
func (p *PathBuilder) WithOutgoingTension(t float64) *PathBuilder {
	p.outTension = t
	p.outTSet = true
	return p
}

// WithIncomingTension sets incoming tension for the next segment.
func (p *PathBuilder) WithIncomingTension(t float64) *PathBuilder {
	p.inTension = t
	p.inTSet = true
	return p
}

// WithStrokeColor stores a stroke color for this path (MetaPost: withcolor).
func (p *PathBuilder) WithStrokeColor(c mp.Color) *PathBuilder {
	p.stroke = c
	p.styleSet = true
	return p
}

// WithStrokeWidth sets the stroke width for this path.
func (p *PathBuilder) WithStrokeWidth(w float64) *PathBuilder {
	p.strokeWidth = w
	p.styleSet = true
	return p
}

// WithFill sets a fill color for this path.
func (p *PathBuilder) WithFill(c mp.Color) *PathBuilder {
	p.fill = c
	p.styleSet = true
	return p
}

// WithPen attaches a pen to this path style (mirrors pen_p in mp.c:564).
func (p *PathBuilder) WithPen(pen *mp.Pen) *PathBuilder {
	p.pen = pen
	p.styleSet = true
	return p
}

// WithLineJoin sets the line join style for corners.
// Use mp.LineJoinMiter (0), mp.LineJoinRound (1), or mp.LineJoinBevel (2).
func (p *PathBuilder) WithLineJoin(join int) *PathBuilder {
	p.lineJoin = join
	p.styleSet = true
	return p
}

// WithLineCap sets the line cap style for endpoints.
// Use mp.LineCapButt (1), mp.LineCapRounded (2), or mp.LineCapSquared (3).
// 0 means default (rounded).
func (p *PathBuilder) WithLineCap(cap int) *PathBuilder {
	p.lineCap = cap
	p.styleSet = true
	return p
}

// WithArrow adds an arrowhead at the end of the path (like drawarrow).
func (p *PathBuilder) WithArrow() *PathBuilder {
	p.arrowEnd = true
	p.styleSet = true
	return p
}

// WithDoubleArrow adds arrowheads at both ends of the path (like drawdblarrow).
func (p *PathBuilder) WithDoubleArrow() *PathBuilder {
	p.arrowStart = true
	p.arrowEnd = true
	p.styleSet = true
	return p
}

// WithArrowStyle sets custom arrow head dimensions.
// length is the arrow head length (default 4), angle is the head angle in degrees (default 45).
func (p *PathBuilder) WithArrowStyle(length, angle float64) *PathBuilder {
	p.arrowLength = length
	p.arrowAngle = angle
	p.styleSet = true
	return p
}

// Dashed sets a custom dash pattern.
// The pattern is given as alternating on/off lengths: on1, off1, on2, off2, ...
// Example: Dashed(6, 3) creates "on 6 off 3" (long dashes with short gaps)
func (p *PathBuilder) Dashed(onOff ...float64) *PathBuilder {
	p.dash = mp.NewDashPattern(onOff...)
	p.styleSet = true
	return p
}

// DashedEvenly sets the standard "evenly" dash pattern (on 3 off 3).
// This mirrors MetaPost's "dashed evenly" from plain.mp.
func (p *PathBuilder) DashedEvenly() *PathBuilder {
	p.dash = mp.DashEvenly()
	p.styleSet = true
	return p
}

// DashedWithDots sets the "withdots" dash pattern.
// This creates dots when used with round linecap.
// Mirrors MetaPost's "dashed withdots" from plain.mp.
func (p *PathBuilder) DashedWithDots() *PathBuilder {
	p.dash = mp.DashWithDots()
	p.styleSet = true
	return p
}

// WithDashPattern sets a pre-created dash pattern.
// Use this for patterns created with mp.NewDashPattern() or mp.DashEvenly().Scaled(2), etc.
func (p *PathBuilder) WithDashPattern(d *mp.DashPattern) *PathBuilder {
	p.dash = d
	p.styleSet = true
	return p
}

// Shifted adds a translation transformation to be applied after solving.
// Mirrors MetaPost's "path shifted (dx, dy)".
func (p *PathBuilder) Shifted(dx, dy float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.Shifted(mp.Number(dx), mp.Number(dy)))
	return p
}

// Scaled adds a uniform scaling transformation around the origin.
// Mirrors MetaPost's "path scaled s".
func (p *PathBuilder) Scaled(s float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.Scaled(mp.Number(s)))
	return p
}

// Rotated adds a rotation transformation around the origin.
// Angle is in degrees (positive = counter-clockwise).
// Mirrors MetaPost's "path rotated angle".
func (p *PathBuilder) Rotated(angleDeg float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.Rotated(mp.Number(angleDeg)))
	return p
}

// Slanted adds a horizontal shear transformation.
// Mirrors MetaPost's "path slanted s".
func (p *PathBuilder) Slanted(s float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.Slanted(mp.Number(s)))
	return p
}

// XScaled adds a horizontal scaling transformation.
// Mirrors MetaPost's "path xscaled s".
func (p *PathBuilder) XScaled(s float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.XScaled(mp.Number(s)))
	return p
}

// YScaled adds a vertical scaling transformation.
// Mirrors MetaPost's "path yscaled s".
func (p *PathBuilder) YScaled(s float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.YScaled(mp.Number(s)))
	return p
}

// ZScaled adds a scaling+rotation using complex multiplication.
// Mirrors MetaPost's "path zscaled (a, b)".
func (p *PathBuilder) ZScaled(a, b float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.ZScaled(mp.Number(a), mp.Number(b)))
	return p
}

// RotatedAround adds a rotation around a given point.
// Equivalent to: shifted(-cx,-cy) rotated(angle) shifted(cx,cy)
func (p *PathBuilder) RotatedAround(cx, cy, angleDeg float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.RotatedAround(mp.Number(cx), mp.Number(cy), mp.Number(angleDeg)))
	return p
}

// ScaledAround adds a scaling around a given point.
// Equivalent to: shifted(-cx,-cy) scaled(s) shifted(cx,cy)
func (p *PathBuilder) ScaledAround(cx, cy, s float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.ScaledAround(mp.Number(cx), mp.Number(cy), mp.Number(s)))
	return p
}

// ReflectedAbout adds a reflection about the line through (x1,y1) and (x2,y2).
// Mirrors MetaPost's "reflectedabout(z1, z2)".
func (p *PathBuilder) ReflectedAbout(x1, y1, x2, y2 float64) *PathBuilder {
	p.transforms = append(p.transforms, mp.ReflectedAbout(mp.Number(x1), mp.Number(y1), mp.Number(x2), mp.Number(y2)))
	return p
}

// Transformed adds a custom transformation.
func (p *PathBuilder) Transformed(t mp.Transform) *PathBuilder {
	p.transforms = append(p.transforms, t)
	return p
}

// CurveTo adds a segment to pt with the stored directions.
func (p *PathBuilder) CurveTo(pt mp.Point) *PathBuilder {
	p.segments = append(p.segments, segment{
		to:         pt,
		toVar:      nil,
		outDir:     p.outDir,
		inDir:      p.inDir,
		outSet:     p.outSet,
		inSet:      p.inSet,
		outTension: p.outTension,
		inTension:  p.inTension,
		outTSet:    p.outTSet,
		inTSet:     p.inTSet,
		outCurl:    p.outCurl,
		inCurl:     p.inCurl,
		outCurlSet: p.outCurlSet,
		inCurlSet:  p.inCurlSet,
	})
	p.resetAfterSegment()
	return p
}

// CurveToVar adds a curve segment to a context variable.
func (p *PathBuilder) CurveToVar(v *Var) *PathBuilder {
	p.segments = append(p.segments, segment{
		toVar:      v,
		outDir:     p.outDir,
		inDir:      p.inDir,
		outSet:     p.outSet,
		inSet:      p.inSet,
		outTension: p.outTension,
		inTension:  p.inTension,
		outTSet:    p.outTSet,
		inTSet:     p.inTSet,
		outCurl:    p.outCurl,
		inCurl:     p.inCurl,
		outCurlSet: p.outCurlSet,
		inCurlSet:  p.inCurlSet,
	})
	p.resetAfterSegment()
	return p
}

// CurveToDir adds a segment and sets outgoing/incoming directions just for this segment.
func (p *PathBuilder) CurveToDir(pt mp.Point, outDeg, inDeg float64) *PathBuilder {
	p.outDir, p.inDir = outDeg, inDeg
	p.outSet, p.inSet = true, true
	return p.CurveTo(pt)
}

// CurveToWithControls adds a segment with explicit control points (skips solving).
func (p *PathBuilder) CurveToWithControls(pt mp.Point, c1, c2 mp.Point) *PathBuilder {
	p.segments = append(p.segments, segment{
		to:         pt,
		line:       false,
		explicit:   true,
		ctrl1:      c1,
		ctrl2:      c2,
		outTension: p.outTension,
		inTension:  p.inTension,
		outTSet:    p.outTSet,
		inTSet:     p.inTSet,
	})
	p.resetAfterSegment()
	return p
}

// LineTo adds a straight segment (MetaPost "--", i.e., {curl 1}..{curl 1}) to (x,y).
func (p *PathBuilder) LineTo(pt mp.Point) *PathBuilder {
	p.segments = append(p.segments, segment{
		to:         pt,
		toVar:      nil,
		line:       true,
		outTension: p.outTension,
		inTension:  p.inTension,
		outTSet:    p.outTSet,
		inTSet:     p.inTSet,
		outCurl:    p.outCurl,
		inCurl:     p.inCurl,
		outCurlSet: p.outCurlSet,
		inCurlSet:  p.inCurlSet,
	})
	p.resetAfterSegment()
	return p
}

// LineToVar adds a straight segment to a context variable.
func (p *PathBuilder) LineToVar(v *Var) *PathBuilder {
	p.segments = append(p.segments, segment{
		toVar:      v,
		line:       true,
		outTension: p.outTension,
		inTension:  p.inTension,
		outTSet:    p.outTSet,
		inTSet:     p.inTSet,
		outCurl:    p.outCurl,
		inCurl:     p.inCurl,
		outCurlSet: p.outCurlSet,
		inCurlSet:  p.inCurlSet,
	})
	p.resetAfterSegment()
	return p
}

func (p *PathBuilder) resetAfterSegment() {
	p.outSet = false
	p.inSet = false
	p.outTSet = false
	p.inTSet = false
	p.outTension = 1
	p.inTension = 1
	p.outCurlSet = false
	p.inCurlSet = false
	p.outCurl = 0
	p.inCurl = 0
}

// Close marks the path as cyclic; use the currently set outDir/inDir as the
// outgoing/incoming directions for the closing segment back to the start.
func (p *PathBuilder) Close() *PathBuilder {
	p.closed = true
	p.closeOut = p.outDir
	p.closeIn = p.inDir
	p.closeOutSet = p.outSet
	p.closeInSet = p.inSet
	return p
}

// resolveStart returns the start point, resolving from Var if needed.
func (p *PathBuilder) resolveStart() mp.Point {
	if p.startVar != nil {
		return p.startVar.Point()
	}
	return p.start
}

// resolveSegmentTarget returns the target point of a segment, resolving from Var if needed.
func (seg *segment) resolveTarget() mp.Point {
	if seg.toVar != nil {
		return seg.toVar.Point()
	}
	return seg.to
}

// BuildPath constructs an mp.Path with knots configured for given directions.
// Directions are converted to MetaPost's scaled degrees.
// If the path uses context variables, the context must be solved first.
func (p *PathBuilder) BuildPath() *mp.Path {
	if !p.startSet || len(p.segments) == 0 {
		return &mp.Path{}
	}
	path := mp.NewPath()
	if p.styleSet {
		path.Style.Stroke = p.stroke
		path.Style.StrokeWidth = p.strokeWidth
		path.Style.Fill = p.fill
		path.Style.Pen = p.pen
		path.Style.LineJoin = p.lineJoin
		path.Style.LineCap = p.lineCap
		path.Style.Arrow.Start = p.arrowStart
		path.Style.Arrow.End = p.arrowEnd
		if p.arrowLength > 0 {
			path.Style.Arrow.Length = p.arrowLength
		} else {
			path.Style.Arrow.Length = mp.DefaultAHLength
		}
		if p.arrowAngle > 0 {
			path.Style.Arrow.Angle = p.arrowAngle
		} else {
			path.Style.Arrow.Angle = mp.DefaultAHAngle
		}
		path.Style.Dash = p.dash
	}

	// Resolve start point (from Var if set)
	startPt := p.resolveStart()

	// start knot
	start := mp.NewKnot()
	start.XCoord = startPt.X
	start.YCoord = startPt.Y
	// Default tensions to 1.0 even when directions are open; MetaPost assumes tension=1
	// unless explicitly overridden.
	start.LeftY = 1
	start.RightY = 1
	if p.closed && len(p.segments) > 0 {
		last := p.segments[len(p.segments)-1]
		if last.inTSet {
			start.LeftY = last.inTension
		}
	}
	if len(p.segments) > 0 && p.segments[0].outTSet {
		start.RightY = p.segments[0].outTension
	}
	if p.closed {
		if p.closeInSet {
			start.LType = mp.KnotGiven
			start.LeftX = degToAngle(p.closeIn)
		} else if p.closeCurlInSet {
			start.LType = mp.KnotCurl
			start.LeftX = p.closeCurlIn
		} else if p.segments[len(p.segments)-1].line {
			start.LType = mp.KnotCurl
			start.LeftX = 1 // curl 1
		} else {
			start.LType = mp.KnotOpen
		}
	} else {
		start.LType = mp.KnotEndpoint
	}
	if p.segments[0].explicit {
		start.RType = mp.KnotExplicit
		start.RightX = p.segments[0].ctrl1.X
		start.RightY = p.segments[0].ctrl1.Y
	} else if p.segments[0].outCurlSet {
		start.RType = mp.KnotCurl
		start.RightX = p.segments[0].outCurl
	} else if p.segments[0].line {
		start.RType = mp.KnotCurl
		start.RightX = 1 // curl 1
	} else if p.segments[0].outSet {
		start.RType = mp.KnotGiven
		start.RightX = degToAngle(p.segments[0].outDir) // store given angle (scaled degrees) in right_x
	} else if !p.closed {
		// Default boundary condition for open paths: curl 1 (mp.w ~7890).
		start.RType = mp.KnotCurl
		start.RightX = 1
	} else if p.segments[len(p.segments)-1].line {
		// Closed path ending with a line (e.g., z0..z1--cycle): the first knot
		// inherits curl boundary from the line closure, matching MetaPost semantics.
		start.RType = mp.KnotCurl
		start.RightX = 1
	} else {
		start.RType = mp.KnotOpen
	}
	path.Append(start)

	// each segment adds an end knot; if multiple segments, intermediate knots chain.
	for i, seg := range p.segments {
		target := seg.resolveTarget()
		end := mp.NewKnot()
		end.XCoord = target.X
		end.YCoord = target.Y
		// Default tensions to 1.0; solver expects tension fields to be initialized.
		end.LeftY = 1
		end.RightY = 1
		if seg.inTSet {
			end.LeftY = seg.inTension
		}
		if seg.explicit {
			end.LType = mp.KnotExplicit
			end.LeftX = seg.ctrl2.X
			end.LeftY = seg.ctrl2.Y
		} else if seg.inCurlSet {
			end.LType = mp.KnotCurl
			end.LeftX = seg.inCurl
		} else if seg.line {
			end.LType = mp.KnotCurl
			end.LeftX = 1 // curl 1
		} else if seg.inSet {
			end.LType = mp.KnotGiven
			end.LeftX = degToAngle(seg.inDir) // incoming angle (scaled degrees) stored in left_x
		} else if i < len(p.segments)-1 {
			// If the next segment specifies an outgoing direction/curl, treat it as
			// the incoming boundary here to mirror MetaPost's point-level direction.
			next := p.segments[i+1]
			if next.line {
				// If next segment is a line, this knot acts as endpoint: curl on both sides
				// (MetaPost rule: -- on either side means curl on BOTH sides)
				end.LType = mp.KnotCurl
				end.LeftX = 1
			} else if next.outCurlSet {
				end.LType = mp.KnotCurl
				end.LeftX = next.outCurl
			} else if next.outSet {
				end.LType = mp.KnotGiven
				end.LeftX = degToAngle(next.outDir)
			} else {
				end.LType = mp.KnotOpen
			}
		} else if !p.closed {
			// Allow a final direction/curl set after the last segment to act as the
			// incoming boundary condition for open paths, mirroring MetaPost's
			// ability to specify a terminal direction.
			switch {
			case p.inSet:
				end.LType = mp.KnotGiven
				end.LeftX = degToAngle(p.inDir)
			case p.inCurlSet:
				end.LType = mp.KnotCurl
				end.LeftX = p.inCurl
			case p.outSet:
				end.LType = mp.KnotGiven
				end.LeftX = degToAngle(p.outDir)
			case p.outCurlSet:
				end.LType = mp.KnotCurl
				end.LeftX = p.outCurl
			default:
				// Default boundary condition for open paths: curl 1 on the final knot.
				end.LType = mp.KnotCurl
				end.LeftX = 1
			}
		} else {
			end.LType = mp.KnotOpen
		}
		if i == len(p.segments)-1 {
			if p.closed {
				if p.closeOutSet {
					end.RType = mp.KnotGiven
					end.RightX = degToAngle(p.closeOut)
				} else if seg.line {
					end.RType = mp.KnotCurl
					end.RightX = 1 // curl 1
				} else {
					end.RType = mp.KnotOpen
				}
			} else {
				end.RType = mp.KnotEndpoint
			}
		} else {
			next := p.segments[i+1]
			if next.line {
				end.RType = mp.KnotCurl
				end.RightX = 1 // curl 1
			} else if next.outCurlSet {
				end.RType = mp.KnotCurl
				end.RightX = next.outCurl
			} else if next.explicit {
				end.RType = mp.KnotExplicit
				end.RightX = next.ctrl1.X
				end.RightY = next.ctrl1.Y
			} else if next.outSet {
				end.RType = mp.KnotGiven
				// carry outgoing direction for next segment
				end.RightX = degToAngle(next.outDir)
			} else {
				end.RType = mp.KnotOpen
			}
			if next.outTSet {
				end.RightY = next.outTension
			}
		}
		path.Append(end)
	}

	return path
}

// Solve builds the path, solves it with a new engine, and applies
// any pending transformations. For better performance when solving
// many paths, use SolveWithEngine to reuse an engine.
func (p *PathBuilder) Solve() (*mp.Path, error) {
	return p.SolveWithEngine(mp.NewEngine())
}

// SolveWithEngine appends the built path to the engine, runs Solve,
// and applies any pending transformations.
func (p *PathBuilder) SolveWithEngine(e *mp.Engine) (*mp.Path, error) {
	path := p.BuildPath()
	e.AddPath(path)
	if err := e.Solve(); err != nil {
		return nil, err
	}
	// Apply any pending transformations
	result := path
	for _, t := range p.transforms {
		result = t.ApplyToPath(result)
	}
	// Copy the style to the transformed result
	if result != path {
		result.Style = path.Style
	}
	return result, nil
}

// MetaPost angles are stored as degrees scaled by angleMultiplier.
func degToAngle(d float64) float64 {
	return d * mp.AngleMultiplier()
}
