package mp

import (
	"fmt"
	"math"
	"sort"
)

// absInt returns the absolute value of an integer.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// PathNormal holds a normalized direction and its length for a path edge.
type PathNormal struct {
	DX, DY Number // original edge delta (mp.delta_x/delta_y analogue)
	Len    Number // edge length
	NX, NY Number // unit normal (rotated left)
}

// PathNormals computes edge deltas and unit normals for a path, mirroring the
// delta/psi preparation leading into offset computations (mp.c:7398ff before
// mp_offset_prep). This is a building block for mp_offset_prep/mp_apply_offset.
func PathNormals(path *Path) []PathNormal {
	if path == nil || path.Head == nil {
		return nil
	}
	var normals []PathNormal
	cur := path.Head
	for {
		next := cur.Next
		if next == nil {
			break
		}
		dx := next.XCoord - cur.XCoord
		dy := next.YCoord - cur.YCoord
		len := math.Hypot(float64(dx), float64(dy))
		n := PathNormal{DX: dx, DY: dy, Len: Number(len)}
		if len != 0 {
			// Rotate left for outward normal; sign adjustments will follow join logic later.
			n.NX = -dy / Number(len)
			n.NY = dx / Number(len)
		}
		normals = append(normals, n)
		cur = next
		if cur == path.Head || cur.RType == KnotEndpoint {
			break
		}
	}
	return normals
}

// htapYpoc creates a reversed copy of a path, mirroring mp_htap_ypoc (mp.c:7062ff).
// Used for open paths to create a doubled outline (forward + backward trace).
// Returns (q, pathTail) where:
//   - q = copy of original head (first node in reversed list, which is at the same position as original head)
//   - pathTail = the last original knot (whose Next == p)
//
// The reversed list is circular: q -> ... -> (copy of original tail) -> q
// mp.c:7062-7087
func htapYpoc(p *Knot) (q *Knot, pathTail *Knot) {
	if p == nil {
		return nil, nil
	}
	// mp.c:7064-7066: q = new knot, qq = q, pp = p
	q = NewKnot()
	qq := q
	pp := p
	for {
		// Copy pp to qq with left/right swapped (mp.c:7068-7076)
		qq.RType = pp.LType
		qq.LType = pp.RType
		qq.XCoord = pp.XCoord
		qq.YCoord = pp.YCoord
		qq.RightX = pp.LeftX
		qq.RightY = pp.LeftY
		qq.LeftX = pp.RightX
		qq.LeftY = pp.RightY
		qq.Origin = pp.Origin
		qq.Info = pp.Info

		// Check if we've gone around the full cycle (mp.c:7077-7080)
		if pp.Next == p {
			// Close reversed list: q.Next = qq (mp.c:7078)
			q.Next = qq
			qq.Prev = q
			// Return q (copy of original head) and pathTail = pp (original tail)
			return q, pp
		}
		// Create next reversed knot and prepend (mp.c:7082-7085)
		rr := NewKnot()
		rr.Next = qq
		qq.Prev = rr
		qq = rr
		pp = pp.Next
	}
}

// PenBBox returns the axis-aligned bounding box of a pen outline, equivalent to
// mp_pen_bbox (mp.c:10670ff) but without transforming; used in offset prep.
func PenBBox(pen *Pen) (minx, miny, maxx, maxy Number, ok bool) {
	pts := penPoints(pen)
	if len(pts) == 0 {
		return 0, 0, 0, 0, false
	}
	minx, miny = math.Inf(1), math.Inf(1)
	maxx, maxy = math.Inf(-1), math.Inf(-1)
	for _, pt := range pts {
		if pt[0] < minx {
			minx = pt[0]
		}
		if pt[0] > maxx {
			maxx = pt[0]
		}
		if pt[1] < miny {
			miny = pt[1]
		}
		if pt[1] > maxy {
			maxy = pt[1]
		}
	}
	return minx, miny, maxx, maxy, true
}

// penOffsetPoints returns the pen points translated by (dx,dy).
func penOffsetPoints(pen *Pen, dx, dy Number) [][2]Number {
	base := penPoints(pen)
	out := make([][2]Number, 0, len(base))
	for _, pt := range base {
		out = append(out, [2]Number{pt[0] + dx, pt[1] + dy})
	}
	return out
}

// OffsetOutline builds a swept outline for a non-elliptical pen by using
// makeEnvelope (mp.c:13445ff). Falls back to polygon approximation if needed.
func OffsetOutline(path *Path, pen *Pen) *Path {
	if pen == nil || pen.Elliptical || path == nil || path.Head == nil {
		return nil
	}
	// Try proper envelope generation (mp.c:13445ff)
	if env := MakeEnvelope(path, pen); env != nil {
		return env
	}
	return nil
}

// flattenPath samples each cubic segment into a polyline (uniform param steps).
// Returns points and whether the path is closed (cycle).
func flattenPath(path *Path, samples int) (pts [][2]Number, closed bool) {
	if path == nil || path.Head == nil {
		return nil, false
	}
	if samples < 1 {
		samples = 1
	}
	// helper: quadratic roots for derivative = 0
	quadRoots := func(a, b, c Number) []Number {
		if a == 0 {
			if b == 0 {
				return nil
			}
			t := -c / b
			return []Number{t}
		}
		d := b*b - 4*a*c
		if d < 0 {
			return nil
		}
		sd := math.Sqrt(float64(d))
		return []Number{
			(-b + Number(sd)) / (2 * a),
			(-b - Number(sd)) / (2 * a),
		}
	}
	cubicPoint := func(x0, y0, x1, y1, x2, y2, x3, y3, t Number) (Number, Number) {
		mt := 1 - t
		mt2 := mt * mt
		t2 := t * t
		x := mt2*mt*x0 + 3*mt2*t*x1 + 3*mt*t2*x2 + t2*t*x3
		y := mt2*mt*y0 + 3*mt2*t*y1 + 3*mt*t2*y2 + t2*t*y3
		return x, y
	}
	addUnique := func(arr []Number, v Number) []Number {
		const eps = 1e-9
		for _, x := range arr {
			if math.Abs(float64(x-v)) < eps {
				return arr
			}
		}
		return append(arr, v)
	}
	cur := path.Head
	for {
		next := cur.Next
		if next == nil {
			break
		}
		x0, y0 := cur.XCoord, cur.YCoord
		x1, y1 := cur.RightX, cur.RightY
		x2, y2 := next.LeftX, next.LeftY
		x3, y3 := next.XCoord, next.YCoord
		// Collect split parameters: 0,1 and extrema of dx/dt, dy/dt in (0,1).
		ts := []Number{0, 1}
		// derivative coefficients for x: 3(-x0+3x1-3x2+x3) t^2 + 2(2x0-4x1+2x2)t + (x1-x0)
		ax := 3 * (-x0 + 3*x1 - 3*x2 + x3)
		bx := 2 * (2*x0 - 4*x1 + 2*x2)
		cx := x1 - x0
		for _, r := range quadRoots(ax, bx, cx) {
			if r > 0 && r < 1 {
				ts = addUnique(ts, r)
			}
		}
		ay := 3 * (-y0 + 3*y1 - 3*y2 + y3)
		by := 2 * (2*y0 - 4*y1 + 2*y2)
		cy := y1 - y0
		for _, r := range quadRoots(ay, by, cy) {
			if r > 0 && r < 1 {
				ts = addUnique(ts, r)
			}
		}
		sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })

		// Sample each sub-interval with `samples` steps to honor extrema splits.
		for seg := 0; seg < len(ts)-1; seg++ {
			t0 := ts[seg]
			t1 := ts[seg+1]
			for i := 0; i < samples; i++ {
				t := t0 + (t1-t0)*(Number(i)/Number(samples))
				x, y := cubicPoint(x0, y0, x1, y1, x2, y2, x3, y3, t)
				pts = append(pts, [2]Number{x, y})
			}
		}
		pts = append(pts, [2]Number{x3, y3})
		if next.RType == KnotEndpoint {
			break
		}
		cur = next
		if cur == path.Head {
			closed = true
			break
		}
	}
	// ensure closed list ends with start if cyclic
	if closed && (pts[0][0] != pts[len(pts)-1][0] || pts[0][1] != pts[len(pts)-1][1]) {
		pts = append(pts, pts[0])
	}
	return pts, closed
}

// splitCubicAt splits the cubic between p (knot) and p.Next at parameter t (0..1),
// inserting a new knot r at the split point with explicit controls. Mirrors
// mp_split_cubic (mp.c:12939ff).
func splitCubicAt(p *Knot, t Number) *Knot {
	if p == nil || p.Next == nil {
		return nil
	}
	q := p.Next
	x0, y0 := p.XCoord, p.YCoord
	x1, y1 := p.RightX, p.RightY
	x2, y2 := q.LeftX, q.LeftY
	x3, y3 := q.XCoord, q.YCoord

	lerp := func(a, b Number) Number { return a + t*(b-a) }
	x01, y01 := lerp(x0, x1), lerp(y0, y1)
	x12, y12 := lerp(x1, x2), lerp(y1, y2)
	x23, y23 := lerp(x2, x3), lerp(y2, y3)
	x012, y012 := lerp(x01, x12), lerp(y01, y12)
	x123, y123 := lerp(x12, x23), lerp(y12, y23)
	x0123, y0123 := lerp(x012, x123), lerp(y012, y123)

	// Left segment controls
	p.RightX, p.RightY = x01, y01
	// Right segment controls for q
	q.LeftX, q.LeftY = x23, y23

	// New knot r at split point.
	r := NewKnot()
	r.XCoord, r.YCoord = x0123, y0123
	r.LeftX, r.LeftY = x012, y012   // incoming control
	r.RightX, r.RightY = x123, y123 // outgoing control
	r.LType, r.RType = KnotExplicit, KnotExplicit

	// Insert r between p and q.
	r.Prev = p
	r.Next = q
	p.Next = r
	q.Prev = r
	return r
}

// PathPoints collects knot coordinates (ignores controls) for polygon export.
func PathPoints(path *Path) [][2]Number {
	if path == nil || path.Head == nil {
		return nil
	}
	var pts [][2]Number
	cur := path.Head
	for {
		pts = append(pts, [2]Number{cur.XCoord, cur.YCoord})
		cur = cur.Next
		if cur == path.Head || cur == nil {
			break
		}
	}
	return pts
}

// splitCubic splits the cubic between p and q=p.Next at parameter t (in fraction units),
// inserting a new knot r at the split point. Mirrors mp_split_cubic (mp.c:12870ff / mp.w:13630ff).
// The parameter t is in fraction units (i.e., t/fractionOne gives the 0..1 parameter).
func splitCubic(p *Knot, t Number) *Knot {
	if p == nil || p.Next == nil {
		return nil
	}
	q := p.Next

	// De Casteljau algorithm (mp.c:12881-12893)
	// set_number_from_of_the_way computes: a + t*(b-a) where t is in fraction units
	v := ofTheWay(p.RightX, q.LeftX, t)
	p.RightX = ofTheWay(p.XCoord, p.RightX, t)
	q.LeftX = ofTheWay(q.LeftX, q.XCoord, t)

	r := NewKnot()
	r.LeftX = ofTheWay(p.RightX, v, t)
	r.RightX = ofTheWay(v, q.LeftX, t)
	r.XCoord = ofTheWay(r.LeftX, r.RightX, t)

	vy := ofTheWay(p.RightY, q.LeftY, t)
	p.RightY = ofTheWay(p.YCoord, p.RightY, t)
	q.LeftY = ofTheWay(q.LeftY, q.YCoord, t)
	r.LeftY = ofTheWay(p.RightY, vy, t)
	r.RightY = ofTheWay(vy, q.LeftY, t)
	r.YCoord = ofTheWay(r.LeftY, r.RightY, t)

	r.LType = KnotExplicit
	r.RType = KnotExplicit
	r.Origin = OriginProgram

	// Insert r between p and q
	r.Next = q
	r.Prev = p
	p.Next = r
	q.Prev = r

	return r
}

// removeCubicNode removes the cubic segment after p by merging it with the next.
// Mirrors mp_remove_cubic (mp.c:12900ff / mp.w:13663ff).
func removeCubicNode(p *Knot) {
	if p == nil || p.Next == nil {
		return
	}
	q := p.Next
	p.Next = q.Next
	if q.Next != nil {
		q.Next.Prev = p
	}
	p.RightX = q.RightX
	p.RightY = q.RightY
}

// offsetPrep ports mp_offset_prep (mp.c:11891ff / mp.w:13372ff).
// It computes turn_amt for each segment and stores offset info in Knot.Info.
// Also splits cubics at direction crossings and returns spec_offset.
func offsetPrep(path *Path, pen *Pen) int {
	if path == nil || path.Head == nil || pen == nil || pen.Head == nil {
		return 0
	}
	if pen.Head.Next == nil || pen.Head.Prev == nil {
		return 0
	}

	// Count pen vertices n (mp.c:13534 / mp.w:13534: section 550)
	n := 0
	pw := pen.Head
	for {
		n++
		pw = pw.Next
		if pw == pen.Head {
			break
		}
	}

	// Compute initial dxin, dyin from pen (mp.c:13546ff / mp.w:13546: section 551)
	h := pen.Head
	hn := h.Next
	hp := h.Prev
	dxin := hn.XCoord - hp.XCoord
	dyin := hn.YCoord - hp.YCoord
	if dxin == 0 && dyin == 0 {
		dxin = hp.YCoord - h.YCoord
		dyin = h.XCoord - hp.XCoord
	}

	w0 := pen.Head
	c := path.Head
	c0 := c
	p := c
	kNeeded := 0
	var dx0, dy0 Number // Save first direction for spec_offset computation
	var ww *Knot        // Declare outside loop to avoid goto jump issue

	debugMain := false

	// Main loop: process each cubic segment (mp.c:13433 do loop)
	segIdx := 0
	for {
		q := p.Next
		if q == nil {
			break
		}

		// Section 558: Set info(p) = zero_off + k_needed (mp.c:12026 / mp.w:13695)
		if debugMain && segIdx == 0 {
			fmt.Printf("offsetPrep: BEFORE first iteration, p==c=%v, c.Info=%d, kNeeded=%d\n",
				p == c, c.Info, kNeeded)
		}
		p.Info = int32(zeroOff + kNeeded)
		kNeeded = 0
		if debugMain {
			fmt.Printf("offsetPrep seg[%d]: p=(%.1f,%.1f) q=(%.1f,%.1f) w0=(%.1f,%.1f)\n",
				segIdx, p.XCoord, p.YCoord, q.XCoord, q.YCoord, w0.XCoord, w0.YCoord)
			fmt.Printf("  p.RType=%d p.LType=%d q.RType=%d q.LType=%d\n",
				p.RType, p.LType, q.RType, q.LType)
			fmt.Printf("  p.Right=(%.1f,%.1f) q.Left=(%.1f,%.1f)\n",
				p.RightX, p.RightY, q.LeftX, q.LeftY)
		}

		// Section 562: Prepare derivative coefficients (mp.c:12035ff / mp.w:13757ff)
		x0 := p.RightX - p.XCoord
		x2 := q.XCoord - q.LeftX
		x1 := q.LeftX - p.RightX
		y0 := p.RightY - p.YCoord
		y2 := q.YCoord - q.LeftY
		y1 := q.LeftY - p.RightY

		// Scale up coefficients if needed (mp.c:12082ff)
		maxCoef := maxAbs6(x0, x1, x2, y0, y1, y2)
		skipProcessing := maxCoef == 0
		if !skipProcessing {
			for maxCoef < fractionHalf {
				maxCoef *= 2
				x0 *= 2
				x1 *= 2
				x2 *= 2
				y0 *= 2
				y1 *= 2
				y2 *= 2
			}

			// Section 567: Find initial direction dx,dy (mp.c:12107ff / mp.w:14063ff)
			dx, dy := x0, y0
			if dx == 0 && dy == 0 {
				dx, dy = x1, y1
				if dx == 0 && dy == 0 {
					dx, dy = x2, y2
				}
			}
			if p == c {
				dx0, dy0 = dx, dy
			}

			// Section 569: Compute turn_amt and update info(p) (mp.c:12173ff / mp.w:14162ff)
			abVsCD := abVsCd(dy, dxin, dx, dyin)
			ccw := numberNonnegative(abVsCD)
			turnAmt := getTurnAmt(w0, dx, dy, ccw)
			if debugMain {
				fmt.Printf("  FIRST turnAmt=%d (dx=%.1f dy=%.1f ccw=%v)\n", turnAmt, dx, dy, ccw)
			}
			w := penWalk(w0, turnAmt)
			w0 = w
			p.Info = int32(int(p.Info) + turnAmt)

			// Section 568: Update dxin, dyin for end direction (mp.c:12224ff / mp.w:14125ff)
			dxin, dyin = x2, y2
			if dxin == 0 && dyin == 0 {
				dxin, dyin = x1, y1
				if dxin == 0 && dyin == 0 {
					dxin, dyin = x0, y0
				}
			}

			// Section 577: Decide on net change in pen offsets (mp.c:12260ff / mp.w:14486ff)
			abVsCD = abVsCd(dx, dyin, dxin, dy)
			var dSign int
			if numberNegative(abVsCD) {
				dSign = -1
			} else if abVsCD == 0 {
				dSign = 0
			} else {
				dSign = 1
			}

			// Section 578: Refine d_sign if zero (mp.c:12294ff / mp.w:14561ff)
			if dSign == 0 {
				u0 := q.XCoord - p.XCoord
				u1 := q.YCoord - p.YCoord
				abVsCD1 := abVsCd(dx, u1, u0, dy)
				abVsCD2 := abVsCd(u0, dyin, dxin, u1)
				tSum := (abVsCD1 + abVsCD2) / 2
				if numberNegative(tSum) {
					dSign = -1
				} else if tSum == 0 {
					dSign = 0
				} else {
					dSign = 1
				}
			}
			if dSign == 0 {
				if dx == 0 {
					if dy > 0 {
						dSign = 1
					} else {
						dSign = -1
					}
				} else {
					if dx > 0 {
						dSign = 1
					} else {
						dSign = -1
					}
				}
			}

			// Section 579: Compute ss for turn direction (mp.c:12335ff / mp.w:14590ff)
			r1 := takeFraction(x0, y2)
			r2 := takeFraction(x2, y0)
			t0 := (r1 - r2) / 2
			r1 = takeFraction(x1, y0+y2)
			r2 = takeFraction(y1, x0+x2)
			t1 := (r1 - r2) / 2
			if t0 == 0 {
				t0 = Number(dSign)
			}

			var u0v, u1v, v0v, v1v, tCross Number
			if t0 > 0 {
				tCross = crossingPoint(t0, t1, -t0)
				u0v = ofTheWay(x0, x1, tCross)
				u1v = ofTheWay(x1, x2, tCross)
				v0v = ofTheWay(y0, y1, tCross)
				v1v = ofTheWay(y1, y2, tCross)
			} else {
				tCross = crossingPoint(-t0, t1, t0)
				u0v = ofTheWay(x2, x1, tCross)
				u1v = ofTheWay(x1, x0, tCross)
				v0v = ofTheWay(y2, y1, tCross)
				v1v = ofTheWay(y1, y0, tCross)
			}
			tmp1 := ofTheWay(u0v, u1v, tCross)
			tmp2 := ofTheWay(v0v, v1v, tCross)
			r1 = takeFraction(x0+x2, tmp1)
			r2 = takeFraction(y0+y2, tmp2)
			ss := r1 + r2

			// Second turn_amt computation (mp.c:12437ff)
			turnAmt = getTurnAmt(w, dxin, dyin, dSign > 0)
			if numberNegative(ss) {
				turnAmt = turnAmt - dSign*n
			}

			// Section 573: Complete the offset splitting process (mp.c:12461ff / mp.w:14329ff)
			ww = w.Prev

			// Section 565: Compute test coefficients t0,t1,t2 (mp.c:12471ff / mp.w:13935ff)
			du := ww.XCoord - w.XCoord
			dv := ww.YCoord - w.YCoord
			absDu := absNumber(du)
			absDv := absNumber(dv)
			var t0c, t1c, t2c Number
			if absDu >= absDv {
				s := makeFraction(dv, du)
				t0c = takeFraction(x0, s) - y0
				t1c = takeFraction(x1, s) - y1
				t2c = takeFraction(x2, s) - y2
				if du < 0 {
					t0c = -t0c
					t1c = -t1c
					t2c = -t2c
				}
			} else {
				s := makeFraction(du, dv)
				t0c = x0 - takeFraction(y0, s)
				t1c = x1 - takeFraction(y1, s)
				t2c = x2 - takeFraction(y2, s)
				if dv < 0 {
					t0c = -t0c
					t1c = -t1c
					t2c = -t2c
				}
			}
			if t0c < 0 {
				t0c = 0
			}

			// Section 575: Find the first t where d(t) crosses (mp.c:12554ff / mp.w:14433ff)
			tCross2 := crossingPoint(t0c, t1c, t2c)
			if turnAmt >= 0 {
				if t2c < 0 {
					tCross2 = fractionOne + 1
				} else {
					// Check if crossing is valid (mp.c:12566ff)
					u0c := ofTheWay(x0, x1, tCross2)
					u1c := ofTheWay(x1, x2, tCross2)
					tmpC := ofTheWay(u0c, u1c, tCross2)
					ssC := takeFraction(-du, tmpC)
					v0c := ofTheWay(y0, y1, tCross2)
					v1c := ofTheWay(y1, y2, tCross2)
					tmpC = ofTheWay(v0c, v1c, tCross2)
					ssC += takeFraction(-dv, tmpC)
					if ssC < 0 {
						tCross2 = fractionOne + 1
					}
				}
			} else if tCross2 > fractionOne {
				tCross2 = fractionOne
			}

			// mp.c:12605ff - Branch based on crossing point
			if tCross2 > fractionOne {
				// Simple case: no splitting needed
				finOffsetPrep(p, w, x0, x1, x2, y0, y1, y2, 1, turnAmt)
			} else {
				// Complex case: split cubic and process both parts (mp.c:12617ff)
				splitCubic(p, tCross2)
				r := p.Next

				// Update coefficients for first part (mp.c:12619-12624)
				x1a := ofTheWay(x0, x1, tCross2)
				x1 = ofTheWay(x1, x2, tCross2)
				x2a := ofTheWay(x1a, x1, tCross2)
				y1a := ofTheWay(y0, y1, tCross2)
				y1 = ofTheWay(y1, y2, tCross2)
				y2a := ofTheWay(y1a, y1, tCross2)

				// Process first part with rise=1, turn_amt=0 (mp.c:12636)
				finOffsetPrep(p, w, x0, x1a, x2a, y0, y1a, y2a, 1, 0)
				x0 = x2a
				y0 = y2a
				r.Info = int32(zeroOff - 1)

				if turnAmt >= 0 {
					// mp.c:12640ff - Additional splitting for turn_amt >= 0
					t1c = ofTheWay(t1c, t2c, tCross2)
					if t1c > 0 {
						t1c = 0
					}
					tCross3 := crossingPoint(0, -t1c, -t2c)
					if tCross3 > fractionOne {
						tCross3 = fractionOne
					}

					// Section 574: Split at r (mp.c:12661ff / mp.w:14411ff)
					splitCubic(r, tCross3)
					r.Next.Info = int32(zeroOff + 1)
					x1a = ofTheWay(x1, x2, tCross3)
					x1 = ofTheWay(x0, x1, tCross3)
					x0a := ofTheWay(x1, x1a, tCross3)
					y1a = ofTheWay(y1, y2, tCross3)
					y1 = ofTheWay(y0, y1, tCross3)
					y0a := ofTheWay(y1, y1a, tCross3)

					finOffsetPrep(r.Next, w, x0a, x1a, x2, y0a, y1a, y2, 1, turnAmt)
					x2 = x0a
					y2 = y0a
					finOffsetPrep(r, ww, x0, x1, x2, y0, y1, y2, -1, 0)
				} else {
					// mp.c:12678 - Process with rise=-1
					finOffsetPrep(r, ww, x0, x1, x2, y0, y1, y2, -1, -1-turnAmt)
				}
			}

			if debugMain {
				fmt.Printf("  turnAmt=%d, w0 before=(%.1f,%.1f)", turnAmt, w0.XCoord, w0.YCoord)
			}
			w0 = penWalk(w0, turnAmt)
			if debugMain {
				fmt.Printf(" after=(%.1f,%.1f) p.Info=%d\n", w0.XCoord, w0.YCoord, p.Info)
			}
		}
		segIdx++

		// Section 552: Advance p to node q (mp.c:12707ff / mp.w:13569ff)
		q0 := q
		for {
			r := p.Next
			if r == nil {
				break
			}
			// Check for degenerate cubic to remove
			if p.XCoord == p.RightX && p.YCoord == p.RightY &&
				p.XCoord == r.LeftX && p.YCoord == r.LeftY &&
				p.XCoord == r.XCoord && p.YCoord == r.YCoord &&
				r != p && r != q {
				// Section 553: Remove the cubic following p
				kNeeded = int(p.Info) - zeroOff
				if r == q {
					q = p
				} else {
					p.Info = int32(kNeeded + int(r.Info))
					kNeeded = 0
				}
				if r == c {
					p.Info = c.Info
					c = p
				}
				removeCubicNode(p)
				r = p
			}
			p = r
			if p == q {
				break
			}
		}
		if q != q0 && (q != c || c == c0) {
			q = q.Next
		}
		if q == c {
			break
		}
	}

	// Section 572: Fix the offset change and compute spec_offset (mp.c:12787ff / mp.w:14302ff)
	debug572 := false
	specOffset := int(c.Info) - zeroOff
	if debug572 {
		fmt.Printf("offsetPrep section572: c=(%.1f,%.1f) c.Info=%d (offset=%d), kNeeded=%d\n",
			c.XCoord, c.YCoord, c.Info, specOffset, kNeeded)
		fmt.Printf("  w0=(%.1f,%.1f) h=(%.1f,%.1f) n=%d\n",
			w0.XCoord, w0.YCoord, h.XCoord, h.YCoord, n)
		fmt.Printf("  dx0=%.1f dy0=%.1f dxin=%.1f dyin=%.1f\n", dx0, dy0, dxin, dyin)
	}
	if c.Next == c {
		c.Info = int32(zeroOff + n)
	} else {
		// Fix by k_needed
		c.Info = int32(int(c.Info) + kNeeded)
		if debug572 {
			fmt.Printf("  after kNeeded: c.Info=%d\n", c.Info)
		}
		// Walk w0 back to h
		walkCount := 0
		for w0 != h {
			c.Info = int32(int(c.Info) + 1)
			w0 = w0.Next
			walkCount++
		}
		if debug572 {
			fmt.Printf("  after walk (%d steps): c.Info=%d\n", walkCount, c.Info)
		}
		// Normalize to range (-n, 0]
		for int(c.Info) <= zeroOff-n {
			c.Info = int32(int(c.Info) + n)
		}
		for int(c.Info) > zeroOff {
			c.Info = int32(int(c.Info) - n)
		}
		if debug572 {
			fmt.Printf("  after normalize: c.Info=%d\n", c.Info)
		}
		// Adjust based on initial direction
		if int(c.Info) != zeroOff && numberNonnegative(abVsCd(dy0, dxin, dx0, dyin)) {
			c.Info = int32(int(c.Info) + n)
			if debug572 {
				fmt.Printf("  after direction adjust (+%d): c.Info=%d\n", n, c.Info)
			}
		}
	}
	// spec_offset is computed BEFORE the fix_by operations (mp.w line 14303).
	// The fix_by operations adjust c.Info for the envelope construction,
	// but spec_offset keeps the original value for pen walking.
	if debug572 {
		fmt.Printf("  final c.Info=%d, specOffset=%d (unchanged from before fix_by)\n", c.Info, specOffset)
	}

	return specOffset
}

// maxAbs6 returns the maximum absolute value of six numbers.
func maxAbs6(a, b, c, d, e, f Number) Number {
	max := absNumber(a)
	if v := absNumber(b); v > max {
		max = v
	}
	if v := absNumber(c); v > max {
		max = v
	}
	if v := absNumber(d); v > max {
		max = v
	}
	if v := absNumber(e); v > max {
		max = v
	}
	if v := absNumber(f); v > max {
		max = v
	}
	return max
}

// absNumber returns the absolute value of a Number.
func absNumber(x Number) Number {
	if x < 0 {
		return -x
	}
	return x
}

// MakeEnvelope creates an envelope outline by walking the pen around the path.
// Mirrors mp_make_envelope (mp.c:13304ff / mp.w:14748ff).
func MakeEnvelope(path *Path, pen *Pen) *Path {
	if path == nil || path.Head == nil || pen == nil || pen.Head == nil {
		return nil
	}
	if pen.Head.Next == nil || pen.Head.Prev == nil {
		return nil
	}

	debug := false // Set to true for debugging

	// Get join/cap settings from path style (mp.c:14827ff)
	// Convert LineJoin from offset constants to MetaPost internal values.
	// LineJoin constants are offset by 1 so 0 = "unset/default" → rounded (1).
	// LineJoinMiter=1 → 0, LineJoinRound=2 → 1, LineJoinBevel=3 → 2.
	ljoin := path.Style.LineJoin
	if ljoin == 0 {
		ljoin = 1 // Default to rounded (MetaPost default is linejoin=1)
	} else {
		ljoin-- // Convert from offset constant to MetaPost value
	}

	// Convert LineCap from offset constants to MetaPost internal values.
	// LineCap constants are offset by 1 so 0 = "unset/default" → rounded (1).
	// LineCapButt=1 → 0, LineCapRounded=2 → 1, LineCapSquared=3 → 2.
	lcap := path.Style.LineCap
	if lcap == 0 {
		lcap = 1 // Default to rounded (MetaPost default)
	} else {
		lcap-- // Convert from offset constant to MetaPost value
	}
	miterlim := Number(4.0) // default miter limit

	// mp.c:14769-14770 - Copy path
	c := path.Copy()

	if debug {
		fmt.Printf("=== MakeEnvelope ===\n")
		fmt.Printf("Pen vertices:\n")
		pw := pen.Head
		for i := 0; i < 10; i++ {
			fmt.Printf("  pen[%d]: (%.3f, %.3f)\n", i, pw.XCoord, pw.YCoord)
			pw = pw.Next
			if pw == pen.Head {
				break
			}
		}
	}

	// mp.w:35192-35196 - For stroked cycles, cut the cycle into an open path
	// by inserting a duplicate knot at the start with endpoint types.
	// This causes htapYpoc to create both outer and inner contours.
	if c.Head.LType != KnotEndpoint {
		if debug {
			fmt.Printf("Stroked cycle detected, cutting at start point\n")
		}
		// Insert duplicate knot at the same position as head
		// mp_left_type(mp_insert_knot(mp, pc, pc->x_coord, pc->y_coord)) = mp_endpoint
		newKnot := mpInsertKnot(c.Head, c.Head.XCoord, c.Head.YCoord)
		newKnot.LType = KnotEndpoint
		// mp_right_type(pc) = mp_endpoint
		c.Head.RType = KnotEndpoint
		// pc = mp_next_knot(pc) - advance head to the old head (now second knot)
		c.Head = c.Head.Next
		// t = 1 (round cap)
		lcap = 1
		if debug {
			fmt.Printf("After cutting: c.Head=(%.1f,%.1f) LType=%d RType=%d\n",
				c.Head.XCoord, c.Head.YCoord, c.Head.LType, c.Head.RType)
		}
	}

	// For open paths, create doubled path using htapYpoc (mp.c:13340-13367 / mp.w:15121ff)
	var specP1, specP2 *Knot
	if debug {
		fmt.Printf("c.Head.LType=%d (KnotEndpoint=%d)\n", c.Head.LType, KnotEndpoint)
	}
	if c.Head.LType == KnotEndpoint {
		if debug {
			fmt.Printf("Open path detected, creating doubled path with htapYpoc\n")
		}
		// mp.c:13341-13347: Create reverse copy and rewire
		// spec_p1 = htap_ypoc(c) returns copy of original head
		// spec_p2 = path_tail = original tail
		specP1, specP2 = htapYpoc(c.Head)
		if specP1 != nil {
			specP1.Origin = OriginProgram

			if debug {
				fmt.Printf("htapYpoc: specP1=(%.1f,%.1f) specP2=(%.1f,%.1f)\n",
					specP1.XCoord, specP1.YCoord, specP2.XCoord, specP2.YCoord)
				// Show reversed list
				fmt.Printf("Reversed list:\n")
				rk := specP1
				for i := 0; i < 10; i++ {
					fmt.Printf("  rev[%d] (%.1f,%.1f)\n", i, rk.XCoord, rk.YCoord)
					rk = rk.Next
					if rk == specP1 || rk == nil {
						break
					}
				}
			}

			if debug {
				fmt.Printf("Before linking:\n")
				fmt.Printf("  specP1=%p (%.1f,%.1f) specP1.Next=%p\n", specP1, specP1.XCoord, specP1.YCoord, specP1.Next)
				fmt.Printf("  specP2=%p (%.1f,%.1f) specP2.Next=%p\n", specP2, specP2.XCoord, specP2.YCoord, specP2.Next)
				fmt.Printf("  c.Head=%p (%.1f,%.1f) c.Head.Next=%p\n", c.Head, c.Head.XCoord, c.Head.YCoord, c.Head.Next)
			}

			// mp.c:13344: spec_p2.Next = spec_p1.Next (orig_tail.Next = copy of 2nd orig node)
			specP2.Next = specP1.Next
			if specP1.Next != nil {
				specP1.Next.Prev = specP2
			}
			// mp.c:13345: spec_p1.Next = c (copy of orig head -> orig head)
			specP1.Next = c.Head
			c.Head.Prev = specP1

			// mp.c:13346: remove cubic from spec_p1 to c (collapse to straight)
			removeCubic(specP1)

			// mp.c:13347: c = spec_p1
			c.Head = specP1

			if debug {
				fmt.Printf("After linking:\n")
				fmt.Printf("  c.Head=%p specP1.Next=%p specP2.Next=%p\n", c.Head, specP1.Next, specP2.Next)
			}

			// mp.c:13348-13350: if not single point, also remove cubic at spec_p2
			if c.Head != c.Head.Next {
				specP2.Origin = OriginProgram
				removeCubic(specP2)
			} else {
				// mp.c:13355-13361: single point case
				c.Head.LType = KnotExplicit
				c.Head.RType = KnotExplicit
				c.Head.LeftX = c.Head.XCoord
				c.Head.LeftY = c.Head.YCoord
				c.Head.RightX = c.Head.XCoord
				c.Head.RightY = c.Head.YCoord
			}

			if debug {
				fmt.Printf("Doubled path knots:\n")
				pk := c.Head
				for i := 0; i < 10; i++ {
					fmt.Printf("  [%d] (%.1f,%.1f) LType=%d RType=%d Left=(%.1f,%.1f) Right=(%.1f,%.1f)\n",
						i, pk.XCoord, pk.YCoord, pk.LType, pk.RType,
						pk.LeftX, pk.LeftY, pk.RightX, pk.RightY)
					pk = pk.Next
					if pk == c.Head || pk == nil {
						break
					}
				}
			}
		}
	}

	// Run offset_prep on the (possibly doubled) path (mp.c:14814-14815 / mp.w:14814ff)
	specOffset := offsetPrep(c, pen)

	// Get initial pen position: h = pen_walk(h, spec_offset) (mp.c:14818 / mp.w:14818)
	h := penWalk(pen.Head, specOffset)
	w := h

	if debug {
		fmt.Printf("specOffset=%d, initial pen h=(%.3f, %.3f)\n", specOffset, h.XCoord, h.YCoord)
		fmt.Printf("Path knots after offsetPrep:\n")
		pk := c.Head
		for i := 0; i < 20; i++ {
			fmt.Printf("  knot[%d]: (%.3f, %.3f) info=%d\n", i, pk.XCoord, pk.YCoord, pk.Info)
			pk = pk.Next
			if pk == c.Head || pk == nil {
				break
			}
		}
	}

	p := c.Head

	if debug {
		fmt.Printf("Before loop: c.Head=(%.1f,%.1f) h=(%.1f,%.1f)\n", c.Head.XCoord, c.Head.YCoord, h.XCoord, h.YCoord)
	}

	// Main envelope loop (mp.c:14770-14798 / mp.w:14770ff)
	// do { ... } while (q0 != c)
	maxIter := 1000
	passedSpecP2 := false // Track when we've passed specP2 (entering inner contour)
	for iter := 0; iter < maxIter; iter++ {
		q := p.Next
		if q == nil {
			break
		}
		q0 := q

		// mp.c:14777-14779 - Save original q coordinates before modification
		qx := q.XCoord
		qy := q.YCoord

		k := int(q.Info)
		k0 := k
		w0 := w

		// mp.c:14782 - Determine join_type (mp.c:14827-14847 / mp.w:14827ff)
		joinType := 0
		var dirs joinDirections
		if k != zeroOff {
			if k < zeroOff {
				// Inner turn - use bevel (mp.c:14838)
				joinType = 2
			} else {
				// Outer turn
				if q != specP1 && q != specP2 {
					joinType = ljoin
				} else if lcap == 2 {
					joinType = 3 // squared cap
				} else {
					joinType = 2 - lcap
				}

				// For the inner contour (after specP2), use bevel join instead of miter.
				// This prevents spikes on the inside of stroked closed paths.
				// MetaPost handles this similarly - both contours have symmetric corner treatment.
				if joinType == 0 && passedSpecP2 {
					joinType = 2 // bevel for inner contour
					if debug {
						fmt.Printf("  Using bevel for inner contour at q=(%.1f,%.1f)\n", q.XCoord, q.YCoord)
					}
				}

				if debug {
					fmt.Printf("  Before computeJoinType2: joinType=%d ljoin=%d\n", joinType, ljoin)
				}

				// For miter/squared joins, compute direction vectors (mp.c:14847ff)
				if joinType == 0 || joinType == 3 {
					joinType, dirs = computeJoinType2(p, q, c.Head, w, h, joinType, miterlim)
				}
			}
		}

		if debug {
			fmt.Printf("Loop: p=(%.1f,%.1f) q=(%.1f,%.1f) k=%d joinType=%d w=(%.1f,%.1f)\n",
				p.XCoord, p.YCoord, q.XCoord, q.YCoord, k-zeroOff, joinType, w.XCoord, w.YCoord)
		}

		// mp.c:14878-14887 / mp.w:14878ff - Translate control points by pen offset w
		// All use the SAME w position (current pen before walk)
		// mp.w:14883-14884: "The coordinates of |p| have already been shifted unless
		// |p| is the first knot in which case they get shifted at the very end."
		// This means: p's coordinates were shifted when p was q in the previous iteration.
		// The first knot gets shifted when it appears as q in the last iteration.
		p.RightX += w.XCoord
		p.RightY += w.YCoord
		q.LeftX += w.XCoord
		q.LeftY += w.YCoord
		q.XCoord += w.XCoord
		q.YCoord += w.YCoord
		q.LType = KnotExplicit
		q.RType = KnotExplicit

		// mp.c:14785-14792 / mp.w:14785ff - Walk pen and insert join knots
		for k != zeroOff {
			// mp.c:14888-14897 / mp.w:14888ff - Step pen forward/backward
			if k > zeroOff {
				w = w.Next
				k--
			} else {
				w = w.Prev
				k++
			}
			// mp.c:14794 - Insert join knots for round joins or at final position
			if joinType == 1 || k == zeroOff {
				xtot := qx + w.XCoord
				ytot := qy + w.YCoord
				q = mpInsertKnot(q, xtot, ytot)
			}
		}

		// mp.c:14929-14945 / mp.w:14929ff - Handle miter/squared joins after inserting pen walk knots
		if q != p.Next {
			// There were join knots inserted, may need miter/squared handling
			insertJoinKnots2(p, q, w, w0, k0, joinType, miterlim, dirs)
		}

		// mp.c:14877 - Advance to next segment
		p = q

		// Track when we pass specP2 (boundary between outer and inner contour)
		if q0 == specP2 && specP2 != nil {
			passedSpecP2 = true
			if debug {
				fmt.Printf("  Passed specP2, entering inner contour\n")
			}
		}

		// mp.c:14878 - Exit condition: when we've processed the segment ending at c (head)
		if q0 == c.Head {
			if debug {
				fmt.Printf("  Exit: q0 == c.Head, c.Head now at (%.1f,%.1f), w was (%.1f,%.1f)\n",
					c.Head.XCoord, c.Head.YCoord, w.XCoord, w.YCoord)
			}
			break
		}
	}

	if debug {
		fmt.Printf("Final: c.Head=(%.1f,%.1f)\n", c.Head.XCoord, c.Head.YCoord)
	}

	// Style: envelope is filled with stroke color, no stroke
	c.Style = path.Style
	c.Style.Fill = path.Style.Stroke
	c.Style.Stroke = ColorCSS("none")
	c.Style.StrokeWidth = 0
	c.Style.Pen = nil
	return c
}

// joinDirections holds the computed direction vectors for miter/squared joins.
type joinDirections struct {
	dxin, dyin, dxout, dyout Number
}

// computeJoinType2 determines if miter/squared join should be used or fall back to bevel.
// Returns the joinType and the computed direction vectors for use in miter/squared insertion.
// Mirrors mp.c:15143-15206 / mp.w:15143ff for direction computation and
// mp.c:14848-14867 / mp.w:14848ff for miter limit check.
func computeJoinType2(p, q, cHead *Knot, w, h *Knot, joinType int, miterlim Number) (int, joinDirections) {
	debug := false

	// Compute incoming direction dxin, dyin (mp.c:15143-15165 / mp.w:15143ff)
	// Note: At this point, q's coordinates have NOT been translated yet (translation happens after)
	dxin := q.XCoord - q.LeftX
	dyin := q.YCoord - q.LeftY
	if debug {
		fmt.Printf("  computeJoinType2: q=(%.1f,%.1f) q.Left=(%.1f,%.1f) dxin=(%.1f,%.1f)\n",
			q.XCoord, q.YCoord, q.LeftX, q.LeftY, dxin, dyin)
	}
	if dxin == 0 && dyin == 0 {
		dxin = q.XCoord - p.RightX
		dyin = q.YCoord - p.RightY
		if debug {
			fmt.Printf("  fallback1: p.Right=(%.1f,%.1f) dxin=(%.1f,%.1f)\n", p.RightX, p.RightY, dxin, dyin)
		}
		if dxin == 0 && dyin == 0 {
			dxin = q.XCoord - p.XCoord
			dyin = q.YCoord - p.YCoord
			if p != cHead {
				dxin += w.XCoord
				dyin += w.YCoord
			}
			if debug {
				fmt.Printf("  fallback2: p=(%.1f,%.1f) dxin=(%.1f,%.1f)\n", p.XCoord, p.YCoord, dxin, dyin)
			}
		}
	}
	tmp := pythAdd(dxin, dyin)
	if tmp == 0 {
		if debug {
			fmt.Printf("  -> bevel (zero direction)\n")
		}
		return 2, joinDirections{} // bevel
	}
	dxin = makeFraction(dxin, tmp)
	dyin = makeFraction(dyin, tmp)

	// Compute outgoing direction dxout, dyout (mp.c:15178-15206 / mp.w:15178ff)
	r := q.Next
	dxout := q.RightX - q.XCoord
	dyout := q.RightY - q.YCoord
	if dxout == 0 && dyout == 0 && r != nil {
		dxout = r.LeftX - q.XCoord
		dyout = r.LeftY - q.YCoord
		if dxout == 0 && dyout == 0 {
			dxout = r.XCoord - q.XCoord
			dyout = r.YCoord - q.YCoord
		}
	}
	if q == cHead {
		dxout -= h.XCoord
		dyout -= h.YCoord
	}
	tmp = pythAdd(dxout, dyout)
	if tmp != 0 {
		dxout = makeFraction(dxout, tmp)
		dyout = makeFraction(dyout, tmp)
	}

	dirs := joinDirections{dxin: dxin, dyin: dyin, dxout: dxout, dyout: dyout}

	// Miter limit check (mp.c:14848-14867 / mp.w:14848ff)
	if joinType == 0 {
		r1 := takeFraction(dxin, dxout)
		r2 := takeFraction(dyin, dyout)
		cosAngle := (r1 + r2) / 2 // half of (1 + cos(angle))
		cosAngle += fractionHalf
		miterTest := takeFraction(miterlim, cosAngle)
		if debug {
			fmt.Printf("  miterLimitCheck: r1=%.3f r2=%.3f cosAngle=%.3f miterTest=%.3f unity=%.3f\n",
				r1, r2, cosAngle, miterTest, unity)
		}
		if miterTest < unity {
			ret := takeScaled(miterlim, miterTest)
			if debug {
				fmt.Printf("  miterLimitCheck: ret=%.3f -> bevel=%v\n", ret, ret < unity)
			}
			if ret < unity {
				return 2, joinDirections{} // bevel
			}
		}
	}
	return joinType, dirs
}

// insertJoinKnots2 inserts miter/squared join knots after pen walk.
// Uses the pre-computed directions from computeJoinType2.
// Mirrors mp.c:14929-15062 / mp.w:14929ff.
func insertJoinKnots2(p, q *Knot, w, w0 *Knot, k0 int, joinType int, miterlim Number, dirs joinDirections) {
	debug := false
	pNext := p.Next
	if pNext == nil {
		return
	}

	if debug {
		fmt.Printf("  insertJoinKnots2: joinType=%d p=(%.1f,%.1f) pNext=(%.1f,%.1f) q=(%.1f,%.1f)\n",
			joinType, p.XCoord, p.YCoord, pNext.XCoord, pNext.YCoord, q.XCoord, q.YCoord)
		fmt.Printf("    dirs: dxin=(%.1f,%.1f) dxout=(%.1f,%.1f)\n",
			dirs.dxin, dirs.dyin, dirs.dxout, dirs.dyout)
	}

	if joinType != 0 && joinType != 3 {
		if debug {
			fmt.Printf("  -> skipping (joinType != 0 and != 3)\n")
		}
		return // Only miter (0) and squared (3) need extra processing
	}

	// Use pre-computed directions from computeJoinType2
	dxin := dirs.dxin
	dyin := dirs.dyin
	dxout := dirs.dxout
	dyout := dirs.dyout

	if joinType == 0 {
		// Miter join (mp.c:14951-14993 / mp.w:14951ff)
		// MetaPost advances p = p.Next first, so use pNext (not original p)
		r := insertMiterJoin2(pNext, q, dxin, dyin, dxout, dyout)
		if r != nil {
			// mp.c:14940-14941
			r.RightX = r.XCoord
			r.RightY = r.YCoord
		}
	} else {
		// Squared join (mp.c:14997-15062 / mp.w:14997ff)
		insertSquaredJoin2(pNext, q, w, w0, k0, dxin, dyin, dxout, dyout)
	}
}

// insertMiterJoin2 inserts a miter join knot. Mirrors mp.c:14951-14993 / mp.w:14951ff.
// pNext is the first inserted knot (after MetaPost's p = p.Next), q is the last inserted knot.
// The miter point is computed using the distance from pNext to q, and inserted after pNext.
func insertMiterJoin2(pNext, q *Knot, dxin, dyin, dxout, dyout Number) *Knot {
	debug := false
	if debug {
		fmt.Printf("    insertMiterJoin2: pNext=(%.1f,%.1f) q=(%.1f,%.1f)\n",
			pNext.XCoord, pNext.YCoord, q.XCoord, q.YCoord)
		fmt.Printf("    dxin=(%.3f,%.3f) dxout=(%.3f,%.3f)\n", dxin, dyin, dxout, dyout)
	}

	// Compute determinant (mp.c:14965-14971)
	r1 := takeFraction(dyout, dxin)
	r2 := takeFraction(dxout, dyin)
	det := r1 - r2
	absDet := det
	if absDet < 0 {
		absDet = -absDet
	}
	if debug {
		fmt.Printf("    det=%.3f nearZeroAngle=%.3f\n", det, nearZeroAngle)
	}
	if absDet < nearZeroAngle {
		if debug {
			fmt.Printf("    -> skipping (det too small)\n")
		}
		return nil
	}

	// Compute intersection point (mp.c:14972-14992)
	// MetaPost advances p = p.Next first, then uses (q - p). So we use (q - pNext).
	tmp := q.XCoord - pNext.XCoord
	r1 = takeFraction(tmp, dyout)
	tmp = q.YCoord - pNext.YCoord
	r2 = takeFraction(tmp, dxout)
	tmp = r1 - r2
	r1 = makeFraction(tmp, det)
	xsub := takeFraction(r1, dxin)
	ysub := takeFraction(r1, dyin)
	// Result is relative to pNext (mp.c:14988-14989)
	xtot := pNext.XCoord + xsub
	ytot := pNext.YCoord + ysub
	if debug {
		fmt.Printf("    -> miter point: (%.3f, %.3f) xsub=%.3f ysub=%.3f\n", xtot, ytot, xsub, ysub)
	}
	// Insert after pNext (mp.c:14990)
	inserted := mpInsertKnot(pNext, xtot, ytot)
	if debug {
		fmt.Printf("    -> inserted knot: %p\n", inserted)
	}
	return inserted
}

// insertSquaredJoin2 inserts squared cap/join knots. Mirrors mp.c:14997-15062 / mp.w:14997ff.
func insertSquaredJoin2(p, q *Knot, w, w0 *Knot, k0 int, dxin, dyin, dxout, dyout Number) {
	// Compute ht vector perpendicular to pen edge (mp.c:15008-15027)
	htx := w.YCoord - w0.YCoord
	hty := w0.XCoord - w.XCoord
	absHtx := htx
	if absHtx < 0 {
		absHtx = -absHtx
	}
	absHty := hty
	if absHty < 0 {
		absHty = -absHty
	}
	// mp.c:15019-15027 - scale up if too small
	for absHtx < fractionHalf && absHty < fractionHalf {
		htx *= 2
		hty *= 2
		absHtx = htx
		if absHtx < 0 {
			absHtx = -absHtx
		}
		absHty = hty
		if absHty < 0 {
			absHty = -absHty
		}
	}

	// Find max height of pen vertices in ht direction (mp.c:15078-15100 / mp.w:15078ff)
	maxHt := Number(0)
	kk := zeroOff
	ww := w
	for {
		// mp.c:15103-15108 - step towards k0
		if kk > k0 {
			ww = ww.Next
			kk--
		} else {
			ww = ww.Prev
			kk++
		}
		if kk == k0 {
			break
		}
		// mp.c:15090-15095
		tmp := ww.XCoord - w0.XCoord
		r1 := takeFraction(tmp, htx)
		tmp = ww.YCoord - w0.YCoord
		r2 := takeFraction(tmp, hty)
		tmp = r1 + r2
		if tmp > maxHt {
			maxHt = tmp
		}
	}

	// Insert first point along dxin direction (mp.c:15032-15037)
	r1 := takeFraction(dxin, htx)
	r2 := takeFraction(dyin, hty)
	tmp := r1 + r2
	if tmp != 0 {
		tmp = makeFraction(maxHt, tmp)
	}
	xsub := takeFraction(tmp, dxin)
	ysub := takeFraction(tmp, dyin)
	xtot := p.XCoord + xsub
	ytot := p.YCoord + ysub
	r := mpInsertKnot(p, xtot, ytot)
	if r == nil {
		return
	}

	// Insert second point along dxout direction (mp.c:15049-15054)
	r1 = takeFraction(dxout, htx)
	r2 = takeFraction(dyout, hty)
	tmp = r1 + r2
	if tmp != 0 {
		tmp = makeFraction(maxHt, tmp)
	}
	xsub = takeFraction(tmp, dxout)
	ysub = takeFraction(tmp, dyout)
	xtot = q.XCoord + xsub
	ytot = q.YCoord + ysub
	mpInsertKnot(r, xtot, ytot)
}

// mpInsertKnot inserts a new knot after q at position (x,y).
// Mirrors mp_insert_knot (mp.c:13893-13909 / mp.w:14905ff).
// Key behavior: copies q's right control to new knot r, then collapses q's right to its anchor.
func mpInsertKnot(q *Knot, x, y Number) *Knot {
	if q == nil || q.Next == nil {
		return nil
	}
	// mp.c:13893-13909
	r := NewKnot()
	r.Next = q.Next
	if q.Next != nil {
		q.Next.Prev = r
	}
	q.Next = r
	r.Prev = q

	// mp.c:13898-13899 - copy q's right control to r
	r.RightX = q.RightX
	r.RightY = q.RightY
	// mp.c:13900-13901 - set r's position
	r.XCoord = x
	r.YCoord = y
	// mp.c:13902-13903 - collapse q's right control to its anchor (creates straight edge)
	q.RightX = q.XCoord
	q.RightY = q.YCoord
	// mp.c:13904-13905 - set r's left control to its anchor (creates straight edge)
	r.LeftX = r.XCoord
	r.LeftY = r.YCoord
	r.LType = KnotExplicit
	r.RType = KnotExplicit
	r.Origin = OriginProgram
	return r
}

// removeCubic removes the node after p and takes its right control point.
// Mirrors mp_remove_cubic (mp.c:12900ff).
// This effectively removes a segment by skipping over the next node.
func removeCubic(p *Knot) {
	if p == nil || p.Next == nil {
		return
	}
	q := p.Next
	// mp.c:12904: mp_next_knot(p) = mp_next_knot(q)
	p.Next = q.Next
	if q.Next != nil {
		q.Next.Prev = p
	}
	// mp.c:12905-12906: take q's right control
	p.RightX = q.RightX
	p.RightY = q.RightY
	// Node q is now orphaned and will be garbage collected
}
