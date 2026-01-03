package mp

import (
	"fmt"
	"math"
)

// zeroOff mirrors the C constant zero_off (mp.c:623) that encodes
// pen-walking offsets in Knot.Info during offset/envelope prep.
const zeroOff = 16384

// penWalk steps k knots forward (positive) or backward (negative) in a cyclic
// pen path. Mirrors mp_pen_walk (mp.c:12997 / mp.w:13707).
func penWalk(w *Knot, k int) *Knot {
	if w == nil {
		return nil
	}
	for k > 0 {
		w = w.Next
		k--
	}
	for k < 0 {
		w = w.Prev
		k++
	}
	return w
}

// insertKnot inserts a new explicit knot after q at position (x,y), copying
// q's outgoing controls to the new knot and collapsing q's outgoing control
// to its anchor. Mirrors mp_insert_knot (mp.c:14905 / mp.w:14905).
func insertKnot(q *Knot, x, y Number) *Knot {
	if q == nil || q.Next == nil {
		return nil
	}
	r := NewKnot()
	r.Next = q.Next
	q.Next.Prev = r
	q.Next = r
	r.Prev = q

	r.RightX, r.RightY = q.RightX, q.RightY
	r.XCoord, r.YCoord = x, y
	q.RightX, q.RightY = q.XCoord, q.YCoord
	r.LeftX, r.LeftY = r.XCoord, r.YCoord
	r.LType, r.RType = KnotExplicit, KnotExplicit
	r.Origin = OriginProgram
	return r
}

// finOffsetPrep performs the directional splitting pass from mp_offset_prep,
// ensuring each cubic stays on one side of the current pen edge direction.
// Mirrors mp_fin_offset_prep (mp.c:12997ff / mp.w:13837ff).
//
// Parameters x0,x1,x2/y0,y1,y2 are the cubic derivative control deltas
// prepared in mp_offset_prep (p->right - p->point, q->left - p->right, etc.).
// rise determines the sign stored in Knot.Info (zeroOff ± 1), and turnAmt
// tracks remaining turn steps for the pen walk.
func finOffsetPrep(p, w *Knot, x0, x1, x2, y0, y1, y2 Number, rise, turnAmt int) {
	debug := false // Set true for debug output
	if p == nil || p.Next == nil || w == nil {
		return
	}
	q := p.Next // original successor; used to detect earlier splits

	if debug {
		fmt.Printf("  finOffsetPrep: p=(%.1f,%.1f) w=(%.1f,%.1f) rise=%d turnAmt=%d\n",
			p.XCoord, p.YCoord, w.XCoord, w.YCoord, rise, turnAmt)
	}

	for {
		var ww *Knot
		if rise > 0 {
			ww = w.Next
		} else {
			ww = w.Prev
		}

		// Compute test coefficients t0,t1,t2 for the projection of the cubic
		// derivative onto the current pen edge (mp.c:13927ff).
		du := ww.XCoord - w.XCoord
		dv := ww.YCoord - w.YCoord
		absDu := math.Abs(float64(du))
		absDv := math.Abs(float64(dv))
		var t0, t1, t2 Number
		if absDu >= absDv {
			s := makeFraction(dv, du)     // dv/du in fraction units
			t0 = takeFraction(x0, s) - y0 // x0*s - y0
			t1 = takeFraction(x1, s) - y1 // x1*s - y1
			t2 = takeFraction(x2, s) - y2 // x2*s - y2
			if du < 0 {
				t0 = numberNegate(t0)
				t1 = numberNegate(t1)
				t2 = numberNegate(t2)
			}
		} else {
			s := makeFraction(du, dv)     // du/dv
			t0 = x0 - takeFraction(y0, s) // x0 - y0*s
			t1 = x1 - takeFraction(y1, s)
			t2 = x2 - takeFraction(y2, s)
			if dv < 0 {
				t0 = numberNegate(t0)
				t1 = numberNegate(t1)
				t2 = numberNegate(t2)
			}
		}
		if t0 < 0 {
			t0 = 0
		}

		t := crossingPoint(t0, t1, t2)
		if t >= fractionOne {
			if turnAmt > 0 {
				t = fractionOne
			} else {
				break
			}
		}

		// Split the cubic at t (fraction units) and tag the new knot.
		// mp.c:12617 - mp_split_cubic(mp, p, t)
		splitCubic(p, t)
		p = p.Next
		p.Info = int32(zeroOff + rise)
		turnAmt--

		// Update derivative controls for the remaining tail (mp.c:13999ff).
		v := ofTheWay(x0, x1, t)
		x1 = ofTheWay(x1, x2, t)
		x0 = ofTheWay(v, x1, t)
		v = ofTheWay(y0, y1, t)
		y1 = ofTheWay(y1, y2, t)
		y0 = ofTheWay(v, y1, t)

		// mp.c:12640ff - If the derivative crosses back, split once more (turnAmt < 0 case)
		if turnAmt < 0 {
			t1 = ofTheWay(t1, t2, t)
			if t1 > 0 {
				t1 = 0
			}
			t = crossingPoint(0, -t1, -t2)
			if t > fractionOne {
				t = fractionOne
			}
			turnAmt++
			if t == fractionOne && p.Next != q {
				// mp.c:12662 - adjust info of existing knot
				p.Next.Info = int32(p.Next.Info - int32(rise))
			} else {
				// mp.c:12661 - mp_split_cubic(mp, r, t)
				splitCubic(p, t)
				p.Next.Info = int32(zeroOff - rise)
				v = ofTheWay(x1, x2, t)
				x1 = ofTheWay(x0, x1, t)
				x2 = ofTheWay(x1, v, t)
				v = ofTheWay(y1, y2, t)
				y1 = ofTheWay(y0, y1, t)
				y2 = ofTheWay(y1, v, t)
			}
		}

		w = ww
	}
}
