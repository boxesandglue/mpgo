package mp

import "math"

// Path operations mirroring MetaPost path queries (mp.c / mp.w).
// These implement "point t of p", "direction t of p", and "subpath (t1,t2) of p".

// PathLength returns the number of segments in the path.
// For a path with n knots, there are n-1 segments (open) or n segments (cycle).
// This corresponds to the maximum integer value of the path parameter t.
func (p *Path) PathLength() int {
	if p == nil || p.Head == nil {
		return 0
	}
	// Count knots
	nKnots := 0
	cur := p.Head
	for {
		nKnots++
		cur = cur.Next
		if cur == nil || cur == p.Head {
			break
		}
	}
	// Check if cycle: neither endpoint has KnotEndpoint type
	isCycle := p.Head.LType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint
	if isCycle {
		return nKnots // cycle has same number of segments as knots
	}
	// Open path has one less segment than knots
	if nKnots > 1 {
		return nKnots - 1
	}
	return 0
}

// getSegment returns the knot at the start of segment index i (0-based),
// along with whether the path is a cycle.
// Returns nil if index is out of bounds.
func (p *Path) getSegment(i int) (*Knot, bool) {
	if p == nil || p.Head == nil || i < 0 {
		return nil, false
	}
	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint
	n := p.PathLength()
	if !isCycle && i >= n {
		return nil, false
	}
	if isCycle {
		// For cycles, wrap around
		i = i % n
		if i < 0 {
			i += n
		}
	}
	cur := p.Head
	for j := 0; j < i; j++ {
		cur = cur.Next
		if cur == nil || cur == p.Head {
			return nil, isCycle
		}
	}
	return cur, isCycle
}

// PointOf returns the point at parameter t on the path.
// Mirrors MetaPost's "point t of p" (mp.c:8750ff / mp.w:9401ff).
//
// The parameter t has integer part selecting the segment (0-based)
// and fractional part [0,1) selecting position within that segment.
// For a path z0..z1..z2:
//   - t=0 gives z0
//   - t=0.5 gives midpoint of first curve
//   - t=1 gives z1
//   - t=1.5 gives midpoint of second curve
//   - t=2 gives z2
//
// For values outside [0, length], the path is linearly extrapolated
// along the tangent at the endpoint.
func (p *Path) PointOf(t Number) (x, y Number) {
	if p == nil || p.Head == nil {
		return 0, 0
	}

	n := p.PathLength()
	if n == 0 {
		return p.Head.XCoord, p.Head.YCoord
	}

	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint

	// Handle extrapolation for open paths
	if !isCycle {
		if t <= 0 {
			if t == 0 {
				return p.Head.XCoord, p.Head.YCoord
			}
			// Extrapolate before start
			dx, dy := p.DirectionOf(0)
			return p.Head.XCoord + t*dx, p.Head.YCoord + t*dy
		}
		if t >= Number(n) {
			// Find last knot
			last := p.Head
			for last.Next != nil && last.Next != p.Head {
				last = last.Next
			}
			if t == Number(n) {
				return last.XCoord, last.YCoord
			}
			// Extrapolate after end
			dx, dy := p.DirectionOf(Number(n))
			excess := t - Number(n)
			return last.XCoord + excess*dx, last.YCoord + excess*dy
		}
	}

	// Get segment index and fractional part
	seg := int(math.Floor(float64(t)))
	frac := t - Number(seg)

	// For cycles, handle wrapping
	if isCycle {
		seg = seg % n
		if seg < 0 {
			seg += n
		}
	}

	// Get the knot at the start of this segment
	knot, _ := p.getSegment(seg)
	if knot == nil || knot.Next == nil {
		// Shouldn't happen, but return last known point
		return p.Head.XCoord, p.Head.YCoord
	}

	// De Casteljau algorithm for cubic Bézier
	// P(t) = (1-t)³P₀ + 3(1-t)²tP₁ + 3(1-t)t²P₂ + t³P₃
	return evalCubic(
		knot.XCoord, knot.YCoord,
		knot.RightX, knot.RightY,
		knot.Next.LeftX, knot.Next.LeftY,
		knot.Next.XCoord, knot.Next.YCoord,
		frac,
	)
}

// PrecontrolOf returns the control point "coming into" parameter t on the path.
// Mirrors MetaPost's "precontrol t of p" (mp.w:9523ff).
//
// At integer t values, this is the left control point of that knot.
// At fractional t values, we split the cubic and return the precontrol of the split point.
func (p *Path) PrecontrolOf(t Number) (x, y Number) {
	if p == nil || p.Head == nil {
		return 0, 0
	}

	n := p.PathLength()
	if n == 0 {
		return p.Head.XCoord, p.Head.YCoord
	}

	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint

	// Clamp t for open paths
	if !isCycle {
		if t < 0 {
			t = 0
		} else if t > Number(n) {
			t = Number(n)
		}
	}

	// Get segment index and fractional part
	seg := int(math.Floor(float64(t)))
	frac := t - Number(seg)

	// Handle endpoint exactly at path length
	if !isCycle && seg >= n {
		seg = n - 1
		frac = 1.0
	}

	// For cycles, handle wrapping
	if isCycle {
		seg = seg % n
		if seg < 0 {
			seg += n
		}
	}

	// Get the knot at the start of this segment
	knot, _ := p.getSegment(seg)
	if knot == nil {
		return 0, 0
	}

	// At integer t (frac == 0), return the left control of this knot
	// But for endpoints, return the point itself
	if frac == 0 {
		if knot.LType == KnotEndpoint {
			return knot.XCoord, knot.YCoord
		}
		return knot.LeftX, knot.LeftY
	}

	// At fractional t, split the cubic and return precontrol of split point
	if knot.Next == nil {
		return knot.XCoord, knot.YCoord
	}

	// De Casteljau to find the control point coming into the split point
	// After splitting at t, the precontrol is r0 (second level, first point)
	u := 1 - frac
	q0x := u*knot.XCoord + frac*knot.RightX
	q0y := u*knot.YCoord + frac*knot.RightY
	q1x := u*knot.RightX + frac*knot.Next.LeftX
	q1y := u*knot.RightY + frac*knot.Next.LeftY

	r0x := u*q0x + frac*q1x
	r0y := u*q0y + frac*q1y

	return r0x, r0y
}

// PostcontrolOf returns the control point "going out of" parameter t on the path.
// Mirrors MetaPost's "postcontrol t of p" (mp.w:9533ff).
//
// At integer t values, this is the right control point of that knot.
// At fractional t values, we split the cubic and return the postcontrol of the split point.
func (p *Path) PostcontrolOf(t Number) (x, y Number) {
	if p == nil || p.Head == nil {
		return 0, 0
	}

	n := p.PathLength()
	if n == 0 {
		return p.Head.XCoord, p.Head.YCoord
	}

	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint

	// Clamp t for open paths
	if !isCycle {
		if t < 0 {
			t = 0
		} else if t > Number(n) {
			t = Number(n)
		}
	}

	// Get segment index and fractional part
	seg := int(math.Floor(float64(t)))
	frac := t - Number(seg)

	// Handle endpoint exactly at path length
	if !isCycle && seg >= n {
		seg = n - 1
		frac = 1.0
	}

	// For cycles, handle wrapping
	if isCycle {
		seg = seg % n
		if seg < 0 {
			seg += n
		}
	}

	// Get the knot at the start of this segment
	knot, _ := p.getSegment(seg)
	if knot == nil {
		return 0, 0
	}

	// At integer t (frac == 0), return the right control of this knot
	// But for endpoints, return the point itself
	if frac == 0 {
		if knot.RType == KnotEndpoint {
			return knot.XCoord, knot.YCoord
		}
		return knot.RightX, knot.RightY
	}

	// At fractional t, split the cubic and return postcontrol of split point
	if knot.Next == nil {
		return knot.XCoord, knot.YCoord
	}

	// De Casteljau to find the control point going out of the split point
	// After splitting at t, the postcontrol is r1 (second level, second point)
	u := 1 - frac
	q1x := u*knot.RightX + frac*knot.Next.LeftX
	q1y := u*knot.RightY + frac*knot.Next.LeftY
	q2x := u*knot.Next.LeftX + frac*knot.Next.XCoord
	q2y := u*knot.Next.LeftY + frac*knot.Next.YCoord

	r1x := u*q1x + frac*q2x
	r1y := u*q1y + frac*q2y

	return r1x, r1y
}

// DirectionOf returns the tangent direction at parameter t on the path.
// Mirrors MetaPost's "direction t of p" defined in plain.mp as:
//
//	postcontrol t of p - precontrol t of p
//
// Returns (dx, dy) representing the tangent vector (not normalized).
func (p *Path) DirectionOf(t Number) (dx, dy Number) {
	postX, postY := p.PostcontrolOf(t)
	preX, preY := p.PrecontrolOf(t)
	return postX - preX, postY - preY
}

// Subpath returns a new path representing the portion from t1 to t2.
// Mirrors MetaPost's "subpath (t1,t2) of p" (mp.c:8869ff / mp.w:9543ff).
//
// If t1 > t2, the subpath runs backwards.
// The returned path is always open (non-cyclic).
func (p *Path) Subpath(t1, t2 Number) *Path {
	if p == nil || p.Head == nil {
		return NewPath()
	}

	n := p.PathLength()
	if n == 0 {
		// Single point path
		result := NewPath()
		k := NewKnot()
		k.XCoord = p.Head.XCoord
		k.YCoord = p.Head.YCoord
		k.RightX = k.XCoord
		k.RightY = k.YCoord
		k.LeftX = k.XCoord
		k.LeftY = k.YCoord
		k.LType = KnotEndpoint
		k.RType = KnotEndpoint
		result.Append(k)
		return result
	}

	// Handle reversed subpath
	if t1 > t2 {
		// Get forward subpath and reverse it
		fwd := p.Subpath(t2, t1)
		return fwd.Reversed()
	}

	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint

	// Clamp for open paths
	if !isCycle {
		if t1 < 0 {
			t1 = 0
		}
		if t2 > Number(n) {
			t2 = Number(n)
		}
	}

	result := NewPath()

	// Get start and end segment indices
	seg1 := int(math.Floor(float64(t1)))
	frac1 := t1 - Number(seg1)
	seg2 := int(math.Floor(float64(t2)))
	frac2 := t2 - Number(seg2)

	// Handle cycles
	if isCycle {
		seg1 = seg1 % n
		if seg1 < 0 {
			seg1 += n
		}
		seg2 = seg2 % n
		if seg2 < 0 {
			seg2 += n
		}
	} else {
		// Clamp segment indices
		if seg1 >= n {
			seg1 = n - 1
			frac1 = 1.0
		}
		if seg2 >= n {
			seg2 = n - 1
			frac2 = 1.0
		}
	}

	if seg1 == seg2 && frac1 <= frac2 {
		// Subpath within a single segment
		return p.subpathSingleSegment(seg1, frac1, frac2)
	}

	// Multi-segment subpath
	// First partial segment: from frac1 to 1
	knot1, _ := p.getSegment(seg1)
	if knot1 == nil || knot1.Next == nil {
		return result
	}

	// Split at frac1 to get the second half
	p0x, p0y := knot1.XCoord, knot1.YCoord
	p1x, p1y := knot1.RightX, knot1.RightY
	p2x, p2y := knot1.Next.LeftX, knot1.Next.LeftY
	p3x, p3y := knot1.Next.XCoord, knot1.Next.YCoord

	if frac1 > 0 {
		// Split and take second half
		_, _, _, _, _, _, _, _, q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y := splitCubicCoords(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, frac1)
		p0x, p0y = q0x, q0y
		p1x, p1y = q1x, q1y
		p2x, p2y = q2x, q2y
		p3x, p3y = q3x, q3y
	}

	// Add first knot
	k := NewKnot()
	k.XCoord = p0x
	k.YCoord = p0y
	k.LeftX = p0x
	k.LeftY = p0y
	k.RightX = p1x
	k.RightY = p1y
	k.LType = KnotExplicit
	k.RType = KnotExplicit
	result.Append(k)

	// Add endpoint of first partial segment
	k = NewKnot()
	k.XCoord = p3x
	k.YCoord = p3y
	k.LeftX = p2x
	k.LeftY = p2y
	k.RightX = p3x
	k.RightY = p3y
	k.LType = KnotExplicit
	k.RType = KnotExplicit
	result.Append(k)

	// Add complete middle segments
	for s := seg1 + 1; s < seg2; s++ {
		actualSeg := s
		if isCycle {
			actualSeg = s % n
			if actualSeg < 0 {
				actualSeg += n
			}
		}
		knot, _ := p.getSegment(actualSeg)
		if knot == nil || knot.Next == nil {
			continue
		}

		// Update control points of previous knot
		lastKnot := result.Head.Prev
		lastKnot.RightX = knot.RightX
		lastKnot.RightY = knot.RightY

		// Add new endpoint
		k = NewKnot()
		k.XCoord = knot.Next.XCoord
		k.YCoord = knot.Next.YCoord
		k.LeftX = knot.Next.LeftX
		k.LeftY = knot.Next.LeftY
		k.RightX = k.XCoord
		k.RightY = k.YCoord
		k.LType = KnotExplicit
		k.RType = KnotExplicit
		result.Append(k)
	}

	// Final partial segment: from 0 to frac2
	if seg2 != seg1 || frac2 < frac1 {
		knot2, _ := p.getSegment(seg2)
		if knot2 != nil && knot2.Next != nil {
			fp0x, fp0y := knot2.XCoord, knot2.YCoord
			fp1x, fp1y := knot2.RightX, knot2.RightY
			fp2x, fp2y := knot2.Next.LeftX, knot2.Next.LeftY
			fp3x, fp3y := knot2.Next.XCoord, knot2.Next.YCoord

			if frac2 < 1 {
				// Split and take first half
				r0x, r0y, r1x, r1y, r2x, r2y, r3x, r3y, _, _, _, _, _, _, _, _ := splitCubicCoords(
					fp0x, fp0y, fp1x, fp1y, fp2x, fp2y, fp3x, fp3y, frac2)
				fp0x, fp0y = r0x, r0y
				fp1x, fp1y = r1x, r1y
				fp2x, fp2y = r2x, r2y
				fp3x, fp3y = r3x, r3y
			}

			// Update control points of previous knot
			lastKnot := result.Head.Prev
			lastKnot.RightX = fp1x
			lastKnot.RightY = fp1y

			// Add final endpoint
			k = NewKnot()
			k.XCoord = fp3x
			k.YCoord = fp3y
			k.LeftX = fp2x
			k.LeftY = fp2y
			k.RightX = k.XCoord
			k.RightY = k.YCoord
			k.LType = KnotExplicit
			k.RType = KnotExplicit
			result.Append(k)
		}
	}

	// Set endpoint types for open path
	if result.Head != nil {
		result.Head.LType = KnotEndpoint
		result.Head.Prev.RType = KnotEndpoint
	}

	return result
}

// subpathSingleSegment extracts a subpath within a single segment.
func (p *Path) subpathSingleSegment(seg int, frac1, frac2 Number) *Path {
	result := NewPath()

	knot, _ := p.getSegment(seg)
	if knot == nil || knot.Next == nil {
		return result
	}

	p0x, p0y := knot.XCoord, knot.YCoord
	p1x, p1y := knot.RightX, knot.RightY
	p2x, p2y := knot.Next.LeftX, knot.Next.LeftY
	p3x, p3y := knot.Next.XCoord, knot.Next.YCoord

	// If frac1 == frac2, return a single point
	if frac1 == frac2 {
		x, y := evalCubic(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, frac1)
		k := NewKnot()
		k.XCoord = x
		k.YCoord = y
		k.LeftX = x
		k.LeftY = y
		k.RightX = x
		k.RightY = y
		k.LType = KnotEndpoint
		k.RType = KnotEndpoint
		result.Append(k)
		return result
	}

	// First split at frac1
	if frac1 > 0 {
		_, _, _, _, _, _, _, _, q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y := splitCubicCoords(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, frac1)
		p0x, p0y = q0x, q0y
		p1x, p1y = q1x, q1y
		p2x, p2y = q2x, q2y
		p3x, p3y = q3x, q3y
		// Adjust frac2 relative to remaining curve
		frac2 = (frac2 - frac1) / (1 - frac1)
	}

	// Then split at (adjusted) frac2 and take first half
	if frac2 < 1 {
		r0x, r0y, r1x, r1y, r2x, r2y, r3x, r3y, _, _, _, _, _, _, _, _ := splitCubicCoords(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, frac2)
		p0x, p0y = r0x, r0y
		p1x, p1y = r1x, r1y
		p2x, p2y = r2x, r2y
		p3x, p3y = r3x, r3y
	}

	// Create the two-knot path
	k1 := NewKnot()
	k1.XCoord = p0x
	k1.YCoord = p0y
	k1.LeftX = p0x
	k1.LeftY = p0y
	k1.RightX = p1x
	k1.RightY = p1y
	k1.LType = KnotEndpoint
	k1.RType = KnotExplicit
	result.Append(k1)

	k2 := NewKnot()
	k2.XCoord = p3x
	k2.YCoord = p3y
	k2.LeftX = p2x
	k2.LeftY = p2y
	k2.RightX = p3x
	k2.RightY = p3y
	k2.LType = KnotExplicit
	k2.RType = KnotEndpoint
	result.Append(k2)

	return result
}

// Reversed returns a copy of the path with direction reversed.
// Used for subpath when t1 > t2.
func (p *Path) Reversed() *Path {
	if p == nil || p.Head == nil {
		return NewPath()
	}

	result := NewPath()

	// Collect all knots in order
	var knots []*Knot
	cur := p.Head
	for {
		knots = append(knots, cur)
		cur = cur.Next
		if cur == nil || cur == p.Head {
			break
		}
	}

	// Add knots in reverse order, swapping left/right control points
	for i := len(knots) - 1; i >= 0; i-- {
		old := knots[i]
		k := NewKnot()
		k.XCoord = old.XCoord
		k.YCoord = old.YCoord
		// Swap left and right control points
		k.LeftX = old.RightX
		k.LeftY = old.RightY
		k.RightX = old.LeftX
		k.RightY = old.LeftY
		// Swap left and right types
		k.LType = old.RType
		k.RType = old.LType
		result.Append(k)
	}

	result.Style = p.Style
	return result
}

// CutBefore returns the portion of path p after its first intersection with path q.
// Mirrors MetaPost's "p cutbefore q" (plain.mp).
//
// If there is no intersection, returns a copy of p.
//
// Example:
//
//	result := p.CutBefore(q)  // p from intersection to end
func (p *Path) CutBefore(q *Path) *Path {
	if p == nil || p.Head == nil {
		return NewPath()
	}
	if q == nil || q.Head == nil {
		return p.Copy()
	}

	t1, _ := p.IntersectionTimes(q)
	if t1 < 0 {
		// No intersection, return copy of original
		return p.Copy()
	}

	// Return subpath from intersection to end
	n := Number(p.PathLength())
	return p.Subpath(t1, n)
}

// CutAfter returns the portion of path p before its first intersection with path q.
// Mirrors MetaPost's "p cutafter q" (plain.mp).
//
// If there is no intersection, returns a copy of p.
//
// Example:
//
//	result := p.CutAfter(q)  // p from start to intersection
func (p *Path) CutAfter(q *Path) *Path {
	if p == nil || p.Head == nil {
		return NewPath()
	}
	if q == nil || q.Head == nil {
		return p.Copy()
	}

	t1, _ := p.IntersectionTimes(q)
	if t1 < 0 {
		// No intersection, return copy of original
		return p.Copy()
	}

	// Return subpath from start to intersection
	return p.Subpath(0, t1)
}

// ArcLength returns the total arc length of the path.
// Mirrors MetaPost's "arclength p" (mp.w:10197ff).
//
// Uses adaptive Simpson's rule to integrate |B'(t)| along each segment.
func (p *Path) ArcLength() Number {
	if p == nil || p.Head == nil {
		return 0
	}

	n := p.PathLength()
	if n == 0 {
		return 0
	}

	var total Number = 0

	// Iterate over each segment
	cur := p.Head
	for i := 0; i < n; i++ {
		if cur.Next == nil {
			break
		}

		// Compute the velocity vectors (control point differences)
		// dx0, dy0 = P1 - P0 (right control - current point)
		// dx1, dy1 = P2 - P1 (left control of next - right control)
		// dx2, dy2 = P3 - P2 (next point - left control of next)
		dx0 := cur.RightX - cur.XCoord
		dy0 := cur.RightY - cur.YCoord
		dx1 := cur.Next.LeftX - cur.RightX
		dy1 := cur.Next.LeftY - cur.RightY
		dx2 := cur.Next.XCoord - cur.Next.LeftX
		dy2 := cur.Next.YCoord - cur.Next.LeftY

		// Add arc length of this segment
		total += doArcTest(dx0, dy0, dx1, dy1, dx2, dy2)

		cur = cur.Next
		if cur == p.Head {
			break
		}
	}

	return total
}

// ArcLengthSegment returns the arc length of a single segment starting at parameter t.
// This is useful for computing arc length of subpaths.
func (p *Path) ArcLengthSegment(segIdx int) Number {
	knot, _ := p.getSegment(segIdx)
	if knot == nil || knot.Next == nil {
		return 0
	}

	dx0 := knot.RightX - knot.XCoord
	dy0 := knot.RightY - knot.YCoord
	dx1 := knot.Next.LeftX - knot.RightX
	dy1 := knot.Next.LeftY - knot.RightY
	dx2 := knot.Next.XCoord - knot.Next.LeftX
	dy2 := knot.Next.YCoord - knot.Next.LeftY

	return doArcTest(dx0, dy0, dx1, dy1, dx2, dy2)
}

// arcTolerance is the tolerance for arc length computation (unity/4096 in MetaPost)
const arcTolerance = 1.0 / 4096.0

// ArcTime returns the time parameter t where the arc length from the start
// of the path reaches the given value arcLen.
// Mirrors MetaPost's "arctime x of p" (mp.w:10255ff).
//
// For non-cyclic paths:
//   - If arcLen < 0, returns 0
//   - If arcLen > total arc length, returns path length
//
// For cyclic paths:
//   - Negative arcLen traverses backwards
//   - arcLen > total wraps around multiple times
func (p *Path) ArcTime(arcLen Number) Number {
	if p == nil || p.Head == nil {
		return 0
	}

	n := p.PathLength()
	if n == 0 {
		return 0
	}

	isCycle := p.Head.RType != KnotEndpoint && p.Head.Prev != nil && p.Head.Prev.RType != KnotEndpoint

	// Handle negative arc length
	if arcLen < 0 {
		if !isCycle {
			return 0
		}
		// For cycles, reverse the path and negate
		rev := p.Reversed()
		t := rev.ArcTime(-arcLen)
		return -t
	}

	// Handle zero arc length
	if arcLen == 0 {
		return 0
	}

	var tTotal Number = 0
	remainingArc := arcLen

	// Iterate over segments
	cur := p.Head
	for i := 0; i < n && remainingArc > 0; i++ {
		if cur.Next == nil {
			break
		}

		// Compute velocity vectors for this segment
		dx0 := cur.RightX - cur.XCoord
		dy0 := cur.RightY - cur.YCoord
		dx1 := cur.Next.LeftX - cur.RightX
		dy1 := cur.Next.LeftY - cur.RightY
		dx2 := cur.Next.XCoord - cur.Next.LeftX
		dy2 := cur.Next.YCoord - cur.Next.LeftY

		// Call arc test with remaining arc as goal
		t := doArcTestWithGoal(dx0, dy0, dx1, dy1, dx2, dy2, remainingArc)

		if t < 0 {
			// Goal was reached in this segment
			// Actual time within segment is t + 2
			tTotal += t + 2
			remainingArc = 0
		} else {
			// Goal not reached, t is the arc length of this segment
			tTotal += 1
			remainingArc -= t
		}

		cur = cur.Next
		if cur == p.Head {
			// Completed one cycle
			if isCycle && remainingArc > 0 {
				// For cycles with remaining arc, compute how many full cycles
				totalArcLen := p.ArcLength()
				if totalArcLen > 0 && remainingArc > totalArcLen {
					fullCycles := int(remainingArc / totalArcLen)
					tTotal += Number(fullCycles * n)
					remainingArc -= Number(fullCycles) * totalArcLen
				}
			}
			break
		}
	}

	return tTotal
}

// doArcTestWithGoal computes arc length or finds time when goal is reached.
// Returns:
//   - Positive value: arc length of segment (goal not reached)
//   - Negative value: -(2 - time) where time is when goal was reached
func doArcTestWithGoal(dx0, dy0, dx1, dy1, dx2, dy2, goal Number) Number {
	v0 := math.Hypot(float64(dx0), float64(dy0))
	v2 := math.Hypot(float64(dx2), float64(dy2))

	vx02 := (dx0+dx2)/2 + dx1
	vy02 := (dy0+dy2)/2 + dy1
	v02 := math.Hypot(float64(vx02), float64(vy02))

	result := arcTestWithGoal(dx0, dy0, dx1, dy1, dx2, dy2, v0, v02, v2, float64(goal), arcTolerance)
	return Number(result)
}

// arcTestWithGoal is like arcTest but also finds time when goal is reached.
// Returns:
//   - Positive: arc length (goal not reached)
//   - Negative: -(2 - time) where 0 <= time <= 1 is when goal was reached
func arcTestWithGoal(dx0, dy0, dx1, dy1, dx2, dy2 Number, v0, v02, v2, goal, tol float64) float64 {
	// Bisect the Bézier quadratic
	dx01 := (dx0 + dx1) / 2
	dy01 := (dy0 + dy1) / 2
	dx12 := (dx1 + dx2) / 2
	dy12 := (dy1 + dy2) / 2
	dx02 := (dx01 + dx12) / 2
	dy02 := (dy01 + dy12) / 2

	// Velocity magnitudes at t=1/4 and t=3/4
	vx002 := (dx0+dx02)/2 + dx01
	vy002 := (dy0+dy02)/2 + dy01
	v002 := math.Hypot(float64(vx002), float64(vy002))

	vx022 := (dx02+dx2)/2 + dx12
	vy022 := (dy02+dy2)/2 + dy12
	v022 := math.Hypot(float64(vx022), float64(vy022))

	// Arc length estimate
	halfV02 := v02 / 2
	arc1 := (v0 + halfV02) / 2
	arc1 = v002 + (arc1-v002)/2
	arc2 := (v2 + halfV02) / 2
	arc2 = v022 + (arc2-v022)/2
	arc := arc1 + arc2

	// Check if goal is reached
	if goal < arc {
		// Goal will be reached in this segment
		simple := isSimple(dx0, dy0, dx1, dy1, dx2, dy2)
		simplyTest := math.Abs(arc - (v0+v2)/2 - v02)

		if simple && simplyTest <= tol {
			// Use parabolic approximation to find time
			// mp.w:9987-10036: solve rising cubic
			return solveForTime(v0, v02, v2, arc1, arc, goal)
		}
	}

	// Check convergence for arc length
	simple := isSimple(dx0, dy0, dx1, dy1, dx2, dy2)
	simplyTest := math.Abs(arc - (v0+v2)/2 - v02)
	if simple && simplyTest <= tol {
		if goal >= arc {
			return arc // Goal not reached, return arc length
		}
		return solveForTime(v0, v02, v2, arc1, arc, goal)
	}

	// Recursive subdivision
	// mp.w:9754-9812
	newTol := tol * 1.5

	// Double the goal for recursive call since control points aren't halved
	// mp.w:9815-9824: a_new = 2 * a_goal
	doubledGoal := goal * 2

	// First half
	a := arcTestWithGoal(dx0, dy0, dx01, dy01, dx02, dy02, v0, v002, halfV02, doubledGoal, newTol)
	if a < 0 {
		// Goal reached in first half, scale time to [0, 0.5]
		// mp.w:9777-9782
		t := a + 2        // time in [0,1] within first half
		return -(2 - t/2) // scale to [0, 0.5]
	}

	// a is arc length of first half (doubled because control points weren't halved)
	// Reduce goal by half of a for second half
	// mp.w:9831-9835
	remainingGoal := doubledGoal - a
	if remainingGoal <= 0 {
		// Goal was exactly at boundary, return time = 0.5
		return -(2 - 0.5)
	}

	b := arcTestWithGoal(dx02, dy02, dx12, dy12, dx2, dy2, halfV02, v022, v2, remainingGoal, newTol)
	if b < 0 {
		// Goal reached in second half, scale time to [0.5, 1]
		// mp.w:9788-9799
		t := b + 2                // time in [0,1] within second half
		return -(2 - (0.5 + t/2)) // scale to [0.5, 1]
	}

	// Goal not reached, return total arc length (halved to account for doubling)
	// mp.w:9801-9803
	return a + (b-a)/2
}

// solveForTime finds t where arc length reaches goal using linear interpolation.
// For more accuracy, we could implement the full solve_rising_cubic (mp.w:10051ff).
func solveForTime(v0, v02, v2, arc1, arc, goal float64) float64 {
	// Simple linear interpolation: t = goal / arc
	// This is accurate for curves with nearly uniform speed
	if arc <= 0 {
		return -(2 - 0)
	}
	t := goal / arc
	if t > 1 {
		t = 1
	}
	return -(2 - t)
}

// doArcTest computes the arc length of a cubic Bézier segment.
// Arguments are the control point differences:
//
//	dx0, dy0 = P1 - P0
//	dx1, dy1 = P2 - P1
//	dx2, dy2 = P3 - P2
//
// Mirrors MetaPost's mp_do_arc_test (mp.w:10150ff).
func doArcTest(dx0, dy0, dx1, dy1, dx2, dy2 Number) Number {
	// Compute velocity magnitudes at t=0, t=0.5, t=1
	v0 := math.Hypot(float64(dx0), float64(dy0))
	v2 := math.Hypot(float64(dx2), float64(dy2))

	// v02 = velocity magnitude at t=0.5 (times 2)
	// At t=0.5: velocity direction is (dx0+dx2)/2 + dx1, (dy0+dy2)/2 + dy1
	vx02 := (dx0+dx2)/2 + dx1
	vy02 := (dy0+dy2)/2 + dy1
	v02 := math.Hypot(float64(vx02), float64(vy02))

	return Number(arcTest(dx0, dy0, dx1, dy1, dx2, dy2, v0, v02, v2, arcTolerance))
}

// arcTest is the recursive arc length computation.
// Mirrors MetaPost's mp_arc_test (mp.w:9688ff).
//
// Uses adaptive Simpson's rule with recursive subdivision.
// The tolerance is increased by factor 1.5 on each recursion.
func arcTest(dx0, dy0, dx1, dy1, dx2, dy2 Number, v0, v02, v2, tol float64) float64 {
	// Bisect the Bézier quadratic (control point differences form a quadratic)
	// mp.w:9842-9854
	dx01 := (dx0 + dx1) / 2
	dy01 := (dy0 + dy1) / 2
	dx12 := (dx1 + dx2) / 2
	dy12 := (dy1 + dy2) / 2
	dx02 := (dx01 + dx12) / 2
	dy02 := (dy01 + dy12) / 2

	// Compute velocity magnitudes at t=1/4 and t=3/4 (times 2)
	// mp.w:9859-9916
	// v002 at t=1/4: (dx0+dx02)/2 + dx01
	vx002 := (dx0+dx02)/2 + dx01
	vy002 := (dy0+dy02)/2 + dy01
	v002 := math.Hypot(float64(vx002), float64(vy002))

	// v022 at t=3/4: (dx02+dx2)/2 + dx12
	vx022 := (dx02+dx2)/2 + dx12
	vy022 := (dy02+dy2)/2 + dy12
	v022 := math.Hypot(float64(vx022), float64(vy022))

	// Compute arc length estimate using Simpson's rule
	// For first half: (v0 + 4*v002/2 + v02/2) / 6 = (v0 + 2*v002 + v02/2) / 6
	// For second half: (v02/2 + 4*v022/2 + v2) / 6 = (v02/2 + 2*v022 + v2) / 6
	// mp.w:9887-9897
	halfV02 := v02 / 2
	arc1 := (v0 + halfV02) / 2
	arc1 = v002 + (arc1-v002)/2 // weighted average
	arc2 := (v2 + halfV02) / 2
	arc2 = v022 + (arc2-v022)/2
	arc := arc1 + arc2

	// Check if control points are in same quadrant (simple case)
	// mp.w:9919-9947
	simple := isSimple(dx0, dy0, dx1, dy1, dx2, dy2)

	// Check if estimate is accurate enough
	// mp.w:9718-9732
	simplyTest := math.Abs(arc - (v0+v2)/2 - v02)
	if simple && simplyTest <= tol {
		return arc
	}

	// Recursive subdivision with increased tolerance
	// mp.w:9754-9812
	newTol := tol * 1.5

	// First half
	a := arcTest(dx0, dy0, dx01, dy01, dx02, dy02, v0, v002, halfV02, newTol)

	// Second half
	b := arcTest(dx02, dy02, dx12, dy12, dx2, dy2, halfV02, v022, v2, newTol)

	// mp.w:9801-9803: return a + half(b - a) = (a + b) / 2
	// Note: control points aren't halved, so results are halved instead
	return a + (b-a)/2
}

// isSimple checks if control points are confined to one quadrant
// or if rotating 45° would put them in one quadrant.
// mp.w:9919-9947
func isSimple(dx0, dy0, dx1, dy1, dx2, dy2 Number) bool {
	// Check if all dx have same sign
	allDxPos := dx0 >= 0 && dx1 >= 0 && dx2 >= 0
	allDxNeg := dx0 <= 0 && dx1 <= 0 && dx2 <= 0
	if allDxPos || allDxNeg {
		// Check if all dy have same sign
		allDyPos := dy0 >= 0 && dy1 >= 0 && dy2 >= 0
		allDyNeg := dy0 <= 0 && dy1 <= 0 && dy2 <= 0
		if allDyPos || allDyNeg {
			return true
		}
	}

	// Check rotated 45° case
	// dx >= dy for all, or dx <= dy for all
	allGE := dx0 >= dy0 && dx1 >= dy1 && dx2 >= dy2
	allLE := dx0 <= dy0 && dx1 <= dy1 && dx2 <= dy2
	if allGE || allLE {
		// Also check -dx vs dy
		allNegGE := -dx0 >= dy0 && -dx1 >= dy1 && -dx2 >= dy2
		allNegLE := -dx0 <= dy0 && -dx1 <= dy1 && -dx2 <= dy2
		if allNegGE || allNegLE {
			return true
		}
	}

	return false
}

// evalCubic evaluates a cubic Bézier curve at parameter t ∈ [0,1].
// P(t) = (1-t)³P₀ + 3(1-t)²tP₁ + 3(1-t)t²P₂ + t³P₃
func evalCubic(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, t Number) (x, y Number) {
	// Use De Casteljau for numerical stability
	u := 1 - t

	// First level
	q0x := u*p0x + t*p1x
	q0y := u*p0y + t*p1y
	q1x := u*p1x + t*p2x
	q1y := u*p1y + t*p2y
	q2x := u*p2x + t*p3x
	q2y := u*p2y + t*p3y

	// Second level
	r0x := u*q0x + t*q1x
	r0y := u*q0y + t*q1y
	r1x := u*q1x + t*q2x
	r1y := u*q1y + t*q2y

	// Third level (the point)
	x = u*r0x + t*r1x
	y = u*r0y + t*r1y
	return
}

// evalCubicDerivative evaluates the derivative of a cubic Bézier at parameter t.
// P'(t) = 3(1-t)²(P₁-P₀) + 6(1-t)t(P₂-P₁) + 3t²(P₃-P₂)
func evalCubicDerivative(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, t Number) (dx, dy Number) {
	u := 1 - t
	u2 := u * u
	t2 := t * t
	ut := u * t

	// Derivative components
	dx = 3*u2*(p1x-p0x) + 6*ut*(p2x-p1x) + 3*t2*(p3x-p2x)
	dy = 3*u2*(p1y-p0y) + 6*ut*(p2y-p1y) + 3*t2*(p3y-p2y)
	return
}

// splitCubicCoords splits a cubic Bézier at parameter t, returning both halves.
// Returns: first half (a0,a1,a2,a3), second half (b0,b1,b2,b3)
func splitCubicCoords(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, t Number) (
	a0x, a0y, a1x, a1y, a2x, a2y, a3x, a3y,
	b0x, b0y, b1x, b1y, b2x, b2y, b3x, b3y Number) {

	u := 1 - t

	// De Casteljau algorithm
	// First level
	q0x := u*p0x + t*p1x
	q0y := u*p0y + t*p1y
	q1x := u*p1x + t*p2x
	q1y := u*p1y + t*p2y
	q2x := u*p2x + t*p3x
	q2y := u*p2y + t*p3y

	// Second level
	r0x := u*q0x + t*q1x
	r0y := u*q0y + t*q1y
	r1x := u*q1x + t*q2x
	r1y := u*q1y + t*q2y

	// Third level (split point)
	sx := u*r0x + t*r1x
	sy := u*r0y + t*r1y

	// First half: p0 -> q0 -> r0 -> s
	a0x, a0y = p0x, p0y
	a1x, a1y = q0x, q0y
	a2x, a2y = r0x, r0y
	a3x, a3y = sx, sy

	// Second half: s -> r1 -> q2 -> p3
	b0x, b0y = sx, sy
	b1x, b1y = r1x, r1y
	b2x, b2y = q2x, q2y
	b3x, b3y = p3x, p3y

	return
}

// IntersectionTimes returns the time parameters (t1, t2) where paths p and q intersect.
// Mirrors MetaPost's "intersectiontimes (p, q)" (mp.w:16130ff).
//
// Returns:
//   - (t1, t2) where p.PointOf(t1) == q.PointOf(t2) (within tolerance)
//   - (-1, -1) if no intersection exists
//
// The algorithm iterates over all pairs of segments and uses recursive bisection
// to find the intersection point.
func (p *Path) IntersectionTimes(q *Path) (t1, t2 Number) {
	if p == nil || p.Head == nil || q == nil || q.Head == nil {
		return -1, -1
	}

	np := p.PathLength()
	nq := q.PathLength()
	if np == 0 || nq == 0 {
		return -1, -1
	}

	// Iterate over all segment pairs (mp.w:16137-16161)
	for tolStep := 0; tolStep <= 3; tolStep += 3 {
		curP := p.Head
		for i := 0; i < np; i++ {
			if curP.Next == nil || curP.RType == KnotEndpoint {
				curP = curP.Next
				if curP == p.Head {
					break
				}
				continue
			}

			curQ := q.Head
			for j := 0; j < nq; j++ {
				if curQ.Next == nil || curQ.RType == KnotEndpoint {
					curQ = curQ.Next
					if curQ == q.Head {
						break
					}
					continue
				}

				// Try to find intersection between segment i of p and segment j of q
				t1Local, t2Local, found := cubicIntersection(
					curP.XCoord, curP.YCoord,
					curP.RightX, curP.RightY,
					curP.Next.LeftX, curP.Next.LeftY,
					curP.Next.XCoord, curP.Next.YCoord,
					curQ.XCoord, curQ.YCoord,
					curQ.RightX, curQ.RightY,
					curQ.Next.LeftX, curQ.Next.LeftY,
					curQ.Next.XCoord, curQ.Next.YCoord,
					tolStep,
				)

				if found {
					// Add segment offsets
					return Number(i) + t1Local, Number(j) + t2Local
				}

				curQ = curQ.Next
				if curQ == q.Head {
					break
				}
			}

			curP = curP.Next
			if curP == p.Head {
				break
			}
		}
	}

	return -1, -1
}

// IntersectionPoint returns the point where paths p and q intersect.
// Mirrors MetaPost's "intersectionpoint (p, q)".
//
// Returns:
//   - (x, y, true) if an intersection exists
//   - (0, 0, false) if no intersection exists
func (p *Path) IntersectionPoint(q *Path) (x, y Number, found bool) {
	t1, _ := p.IntersectionTimes(q)
	if t1 < 0 {
		return 0, 0, false
	}
	x, y = p.PointOf(t1)
	return x, y, true
}

// maxIntersectionPatience limits backtracking to prevent infinite loops (mp.w:15868)
const maxIntersectionPatience = 5000

// cubicIntersection finds the intersection of two cubic Bézier curves.
// Uses recursive bisection to find where two curves overlap.
//
// Returns (t1, t2, found) where t1, t2 ∈ [0,1] are the parameters on each curve.
func cubicIntersection(
	p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y Number, // First curve
	q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y Number, // Second curve
	tolStep int,
) (t1, t2 Number, found bool) {
	// Use recursive subdivision approach
	const maxDepth = 20
	const tolerance = 0.0001

	// Bézier bounding box check
	pMinX := minOf4(p0x, p1x, p2x, p3x)
	pMaxX := maxOf4(p0x, p1x, p2x, p3x)
	pMinY := minOf4(p0y, p1y, p2y, p3y)
	pMaxY := maxOf4(p0y, p1y, p2y, p3y)

	qMinX := minOf4(q0x, q1x, q2x, q3x)
	qMaxX := maxOf4(q0x, q1x, q2x, q3x)
	qMinY := minOf4(q0y, q1y, q2y, q3y)
	qMaxY := maxOf4(q0y, q1y, q2y, q3y)

	// Check for bounding box overlap
	tol := Number(tolStep)
	if pMaxX+tol < qMinX || qMaxX+tol < pMinX ||
		pMaxY+tol < qMinY || qMaxY+tol < pMinY {
		return -1, -1, false
	}

	return cubicIntersectionRecursive(
		p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, 0, 1,
		q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y, 0, 1,
		maxDepth, tolerance,
	)
}

// cubicIntersectionRecursive is the recursive bisection algorithm.
func cubicIntersectionRecursive(
	p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y Number, pT0, pT1 Number,
	q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y Number, qT0, qT1 Number,
	depth int, tolerance Number,
) (t1, t2 Number, found bool) {

	// Check bounding box overlap
	pMinX := minOf4(p0x, p1x, p2x, p3x)
	pMaxX := maxOf4(p0x, p1x, p2x, p3x)
	pMinY := minOf4(p0y, p1y, p2y, p3y)
	pMaxY := maxOf4(p0y, p1y, p2y, p3y)

	qMinX := minOf4(q0x, q1x, q2x, q3x)
	qMaxX := maxOf4(q0x, q1x, q2x, q3x)
	qMinY := minOf4(q0y, q1y, q2y, q3y)
	qMaxY := maxOf4(q0y, q1y, q2y, q3y)

	// No overlap
	if pMaxX < qMinX || qMaxX < pMinX ||
		pMaxY < qMinY || qMaxY < pMinY {
		return -1, -1, false
	}

	// Check if curves are small enough
	pSize := max(pMaxX-pMinX, pMaxY-pMinY)
	qSize := max(qMaxX-qMinX, qMaxY-qMinY)

	if pSize <= tolerance && qSize <= tolerance {
		// Curves overlap in a small region - found intersection
		return (pT0 + pT1) / 2, (qT0 + qT1) / 2, true
	}

	// Depth limit reached
	if depth <= 0 {
		return (pT0 + pT1) / 2, (qT0 + qT1) / 2, true
	}

	// Split the larger curve
	if pSize >= qSize {
		// Split first curve at t=0.5
		pMid := (pT0 + pT1) / 2
		pa0x, pa0y, pa1x, pa1y, pa2x, pa2y, pa3x, pa3y,
			pb0x, pb0y, pb1x, pb1y, pb2x, pb2y, pb3x, pb3y := splitCubicCoords(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, 0.5)

		// Try first half
		t1, t2, found = cubicIntersectionRecursive(
			pa0x, pa0y, pa1x, pa1y, pa2x, pa2y, pa3x, pa3y, pT0, pMid,
			q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y, qT0, qT1,
			depth-1, tolerance,
		)
		if found {
			return t1, t2, true
		}

		// Try second half
		return cubicIntersectionRecursive(
			pb0x, pb0y, pb1x, pb1y, pb2x, pb2y, pb3x, pb3y, pMid, pT1,
			q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y, qT0, qT1,
			depth-1, tolerance,
		)
	} else {
		// Split second curve at t=0.5
		qMid := (qT0 + qT1) / 2
		qa0x, qa0y, qa1x, qa1y, qa2x, qa2y, qa3x, qa3y,
			qb0x, qb0y, qb1x, qb1y, qb2x, qb2y, qb3x, qb3y := splitCubicCoords(
			q0x, q0y, q1x, q1y, q2x, q2y, q3x, q3y, 0.5)

		// Try first half
		t1, t2, found = cubicIntersectionRecursive(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, pT0, pT1,
			qa0x, qa0y, qa1x, qa1y, qa2x, qa2y, qa3x, qa3y, qT0, qMid,
			depth-1, tolerance,
		)
		if found {
			return t1, t2, true
		}

		// Try second half
		return cubicIntersectionRecursive(
			p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, pT0, pT1,
			qb0x, qb0y, qb1x, qb1y, qb2x, qb2y, qb3x, qb3y, qMid, qT1,
			depth-1, tolerance,
		)
	}
}

// minOf4 returns the minimum of four numbers
func minOf4(a, b, c, d Number) Number {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	if d < m {
		m = d
	}
	return m
}

// maxOf4 returns the maximum of four numbers
func maxOf4(a, b, c, d Number) Number {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	return m
}

// max returns the maximum of two numbers
func max(a, b Number) Number {
	if a > b {
		return a
	}
	return b
}

// BuildCycle constructs a cyclic path from multiple paths by finding their
// intersection points and connecting them. This mirrors MetaPost's buildcycle
// macro from plain.mp.
//
// The algorithm:
//  1. For each consecutive pair of paths (wrapping around), find intersection
//  2. Extract the subpath of each path between its two intersection points
//  3. Join them into a closed cycle
//
// Returns nil if any consecutive pair of paths doesn't intersect.
func BuildCycle(paths ...*Path) *Path {
	n := len(paths)
	if n < 2 {
		return nil
	}

	// Arrays to store intersection times for each path:
	// ta[i] = start time on path[i] (intersection with previous path)
	// tb[i] = end time on path[i] (intersection with next path)
	ta := make([]Number, n)
	tb := make([]Number, n)

	// Find intersections between consecutive paths
	// MetaPost uses: pp[i] intersectiontimes reverse pp[i_]
	// where i_ is the previous path index
	prevIdx := n - 1 // Start with last path as "previous"
	for i := 0; i < n; i++ {
		// Intersect path[i] with reversed path[prevIdx]
		reversedPrev := paths[prevIdx].Reversed()
		if reversedPrev == nil {
			return nil
		}

		t1, t2 := paths[i].IntersectionTimes(reversedPrev)
		if t1 < 0 {
			// No intersection found
			return nil
		}

		// ta[i] is where path[i] starts (intersection with previous)
		ta[i] = t1

		// tb[prevIdx] is where previous path ends
		// Since we intersected with reversed path, we need to convert the time:
		// tb = length - t2
		prevLen := Number(paths[prevIdx].PathLength())
		tb[prevIdx] = prevLen - t2

		prevIdx = i
	}

	// Build the result by extracting and joining subpaths
	result := NewPath()
	for i := 0; i < n; i++ {
		sub := paths[i].Subpath(ta[i], tb[i])
		if sub == nil || sub.Head == nil {
			continue
		}

		if result.Head == nil {
			// First subpath - copy all knots
			cur := sub.Head
			for {
				k := NewKnot()
				k.XCoord = cur.XCoord
				k.YCoord = cur.YCoord
				k.LeftX = cur.LeftX
				k.LeftY = cur.LeftY
				k.RightX = cur.RightX
				k.RightY = cur.RightY
				k.LType = KnotExplicit
				k.RType = KnotExplicit
				result.Append(k)

				cur = cur.Next
				if cur == nil || cur == sub.Head || cur.RType == KnotEndpoint {
					break
				}
			}
			// Include last knot if it's an endpoint
			if cur != nil && cur != sub.Head && cur.RType == KnotEndpoint {
				k := NewKnot()
				k.XCoord = cur.XCoord
				k.YCoord = cur.YCoord
				k.LeftX = cur.LeftX
				k.LeftY = cur.LeftY
				k.RightX = cur.RightX
				k.RightY = cur.RightY
				k.LType = KnotExplicit
				k.RType = KnotExplicit
				result.Append(k)
			}
		} else {
			// Subsequent subpaths - skip first knot (it overlaps with previous)
			// but update the connection controls
			cur := sub.Head
			if cur != nil {
				// Update the last knot's right control to connect smoothly
				lastKnot := result.Head.Prev
				if lastKnot != nil {
					lastKnot.RightX = cur.RightX
					lastKnot.RightY = cur.RightY
				}
				cur = cur.Next
			}

			// Add remaining knots (including the last one for intermediate subpaths)
			isLastSubpath := (i == n-1)
			for cur != nil && cur != sub.Head {
				// Only skip final endpoint for the LAST subpath (will be handled by cycle close)
				if isLastSubpath && cur.RType == KnotEndpoint && cur.Next == nil {
					break
				}
				k := NewKnot()
				k.XCoord = cur.XCoord
				k.YCoord = cur.YCoord
				k.LeftX = cur.LeftX
				k.LeftY = cur.LeftY
				k.RightX = cur.RightX
				k.RightY = cur.RightY
				k.LType = KnotExplicit
				k.RType = KnotExplicit
				result.Append(k)

				cur = cur.Next
				if isLastSubpath && cur != nil && cur.RType == KnotEndpoint {
					break
				}
			}
		}
	}

	// Close the cycle by connecting last knot to first
	if result.Head != nil && result.Head.Prev != nil {
		first := result.Head
		last := result.Head.Prev

		// Make it a proper cycle
		last.RType = KnotExplicit
		first.LType = KnotExplicit

		// The last subpath's endpoint should connect back to first
		// Update controls for smooth connection
		if n > 0 {
			sub := paths[n-1].Subpath(ta[n-1], tb[n-1])
			if sub != nil && sub.Head != nil {
				// Find endpoint of last subpath
				endpoint := sub.Head
				for endpoint.Next != nil && endpoint.Next != sub.Head {
					if endpoint.Next.RType == KnotEndpoint {
						endpoint = endpoint.Next
						break
					}
					endpoint = endpoint.Next
				}
				last.RightX = endpoint.RightX
				last.RightY = endpoint.RightY
			}
		}

		// Connect first knot's left control from first subpath start
		if n > 0 {
			sub := paths[0].Subpath(ta[0], tb[0])
			if sub != nil && sub.Head != nil {
				first.LeftX = sub.Head.LeftX
				first.LeftY = sub.Head.LeftY
			}
		}
	}

	return result
}

// DirectionTimeOf returns the first time t when the path has the given direction.
// Mirrors MetaPost's "directiontime (dx,dy) of p" (mp.w:9593ff).
//
// Returns -1 if the direction is never achieved on the path.
//
// The direction vector (dx, dy) does not need to be normalized.
// For example, directiontime (1, 1) finds where the tangent is at 45°.
//
// Example:
//
//	t := path.DirectionTimeOf(1, 0)  // Find where tangent is horizontal (rightward)
//	t := path.DirectionTimeOf(0, 1)  // Find where tangent is vertical (upward)
func (p *Path) DirectionTimeOf(dx, dy Number) Number {
	if p == nil || p.Head == nil {
		return -1
	}
	if dx == 0 && dy == 0 {
		return -1
	}

	n := p.PathLength()
	if n == 0 {
		return -1
	}

	// For each segment, solve for t where direction equals (dx, dy)
	cur := p.Head
	for seg := 0; seg < n; seg++ {
		if cur.Next == nil {
			break
		}

		// Get segment control points
		p0x, p0y := cur.XCoord, cur.YCoord
		p1x, p1y := cur.RightX, cur.RightY
		p2x, p2y := cur.Next.LeftX, cur.Next.LeftY
		p3x, p3y := cur.Next.XCoord, cur.Next.YCoord

		// Find t where derivative is parallel to (dx, dy)
		t, found := directionTimeInSegment(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, dx, dy)
		if found {
			return Number(seg) + t
		}

		cur = cur.Next
		if cur == p.Head {
			break
		}
	}

	return -1
}

// directionTimeInSegment finds the first t in [0,1] where the cubic Bézier
// segment has direction parallel to (dx, dy).
//
// The derivative of a cubic Bézier is:
//
//	B'(t) = 3[(1-t)²(P1-P0) + 2(1-t)t(P2-P1) + t²(P3-P2)]
//
// We want B'(t) × (dx,dy) = 0 (cross product = 0 means parallel)
// This gives a quadratic equation in t.
func directionTimeInSegment(
	p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y Number,
	dx, dy Number,
) (Number, bool) {
	// Compute the three difference vectors
	ax, ay := p1x-p0x, p1y-p0y // A = P1 - P0
	bx, by := p2x-p1x, p2y-p1y // B = P2 - P1
	cx, cy := p3x-p2x, p3y-p2y // C = P3 - P2

	// Cross products with target direction
	// a = A × D, b = B × D, c = C × D
	a := ax*dy - ay*dx
	b := bx*dy - by*dx
	c := cx*dy - cy*dx

	// Quadratic coefficients: αt² + βt + γ = 0
	// Derived from expanding B'(t) × D = 0
	alpha := a - 2*b + c
	beta := 2 * (b - a)
	gamma := a

	// Handle degenerate case: direction is constant and parallel to target
	// This happens for straight lines where a = b = c = 0
	const epsDegenerate = 1e-12
	if math.Abs(float64(alpha)) < epsDegenerate && math.Abs(float64(beta)) < epsDegenerate && math.Abs(float64(gamma)) < epsDegenerate {
		// Check if initial direction is parallel to target
		dirX, dirY := cubicDerivative(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, 0)
		dot := dirX*dx + dirY*dy
		if dot >= 0 { // Same direction
			return 0, true // Any t works, return 0
		}
		return -1, false // Opposite direction
	}

	// Solve quadratic equation
	roots := solveQuadratic(alpha, beta, gamma)

	// Find smallest valid root in [0, 1]
	const eps = 1e-9
	bestT := Number(-1)
	for _, t := range roots {
		if t >= -eps && t <= 1+eps {
			// Clamp to [0, 1]
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			// Verify the direction is actually parallel (same orientation, not opposite)
			// This is important because cross product = 0 includes antiparallel vectors
			dirX, dirY := cubicDerivative(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, t)
			dot := dirX*dx + dirY*dy
			if dot >= 0 { // Same direction (not opposite)
				if bestT < 0 || t < bestT {
					bestT = t
				}
			}
		}
	}

	return bestT, bestT >= 0
}

// cubicDerivative computes the derivative of a cubic Bézier at parameter t.
// Returns the tangent vector (not normalized).
func cubicDerivative(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y, t Number) (Number, Number) {
	// B'(t) = 3[(1-t)²(P1-P0) + 2(1-t)t(P2-P1) + t²(P3-P2)]
	u := 1 - t
	u2 := u * u
	t2 := t * t
	ut2 := 2 * u * t

	ax, ay := p1x-p0x, p1y-p0y
	bx, by := p2x-p1x, p2y-p1y
	cx, cy := p3x-p2x, p3y-p2y

	dx := 3 * (u2*ax + ut2*bx + t2*cx)
	dy := 3 * (u2*ay + ut2*by + t2*cy)
	return dx, dy
}

// solveQuadratic solves αx² + βx + γ = 0 and returns real roots.
func solveQuadratic(alpha, beta, gamma Number) []Number {
	const eps = 1e-12

	// Handle linear case (α ≈ 0)
	if math.Abs(float64(alpha)) < eps {
		if math.Abs(float64(beta)) < eps {
			return nil // No solution or infinite solutions
		}
		return []Number{-gamma / beta}
	}

	// Quadratic formula: x = (-β ± √(β²-4αγ)) / 2α
	discriminant := beta*beta - 4*alpha*gamma

	if discriminant < -eps {
		return nil // No real roots
	}

	if discriminant < eps {
		// One repeated root
		return []Number{-beta / (2 * alpha)}
	}

	// Two distinct roots
	sqrtD := Number(math.Sqrt(float64(discriminant)))
	r1 := (-beta + sqrtD) / (2 * alpha)
	r2 := (-beta - sqrtD) / (2 * alpha)

	// Return in ascending order
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	return []Number{r1, r2}
}

// DirectionPointOf returns the first point on the path where it has the given direction.
// Mirrors MetaPost's "directionpoint (dx,dy) of p" macro.
//
// Returns (0, 0) and false if the direction is never achieved.
//
// Example:
//
//	x, y, ok := path.DirectionPointOf(1, 0)  // Point where tangent is horizontal
func (p *Path) DirectionPointOf(dx, dy Number) (x, y Number, found bool) {
	t := p.DirectionTimeOf(dx, dy)
	if t < 0 {
		return 0, 0, false
	}
	x, y = p.PointOf(t)
	return x, y, true
}
