package mp

import (
	"fmt"
	"math"
)

// Solver pipeline skeleton, aligned to mp.c (mp_make_choices, etc.).

// solvePath coordinates the steps for a single path.
func (e *Engine) solvePath(p *Path) error {
	// mp_make_choices is the core solver stage; offset/envelope steps remain TODO.
	if err := e.makeChoices(p); err != nil {
		return err
	}
	return nil
}

// applyOffset mirrors the offset/envelope step for non-elliptical pens by
// computing a swept outline (mp_offset_prep/mp_apply_offset, mp.c:13364ff, 15800ff).
// For now this uses OffsetOutline (convex hull over translated pen) as an
// approximation and stores it on the path for later backends.
func (e *Engine) applyOffset(p *Path) {
	if p == nil || p.Style.Pen == nil {
		return
	}
	pen := p.Style.Pen
	// Defaults for linejoin/linecap as in mp_set_up_envelope (mp.c:29290ff).
	// LineJoin: 0=miter (default), 1=round, 2=bevel
	// LineCap: 0=butt (default), 1=round, 2=square
	// Note: 0 is a valid value (miter/butt), so don't override it.
	if pen.Elliptical {
		// Elliptical pens are handled via stroke width in the backend.
		return
	}
	if env := OffsetOutline(p, pen); env != nil && env.Head != nil {
		// Fill the envelope with the stroke color; stroke is unused for the envelope.
		env.Style = p.Style
		env.Style.Fill = p.Style.Stroke
		env.Style.Stroke = ColorCSS("none")
		env.Style.StrokeWidth = 0
		env.Style.Pen = nil
		// Avoid recursion if reused.
		env.Envelope = nil
		p.Envelope = env
	}
}

// makeChoices mirrors mp_make_choices (mp.c ~7321ff / mp.w ~7788ff).
func (e *Engine) makeChoices(p *Path) error {
	if p == nil || p.Head == nil {
		return fmt.Errorf("makeChoices: empty path")
	}
	knots := p.Head

	// Block: Coincident knots -> force explicit and align control points.
	// mp.c:7340-7362 (mp.w ~7831ff)
	cur := knots
	for {
		q := cur.Next
		if numberEqual(cur.XCoord, q.XCoord) &&
			numberEqual(cur.YCoord, q.YCoord) &&
			cur.RType > KnotExplicit {
			cur.RType = KnotExplicit
			if cur.LType == KnotOpen {
				cur.LType = KnotCurl
				cur.LeftX = unity // curl is stored in left_x in C
			}
			q.LType = KnotExplicit
			if q.RType == KnotOpen {
				q.RType = KnotCurl
				q.RightX = unity // curl is stored in right_x in C
			}
			q.LeftX = cur.XCoord
			cur.RightX = cur.XCoord
			q.LeftY = cur.YCoord
			cur.RightY = cur.YCoord
		}
		cur = q
		if cur.RType == KnotEndpoint {
			break
		}
		if cur == knots {
			break
		}
	}

	// Block: detect pure open cycle -> mark end_cycle.
	// mp.c:7370-7379 (mp.w ~7860ff)
	h := knots
	for {
		if h.LType != KnotOpen {
			break
		}
		if h.RType != KnotOpen {
			break
		}
		h = h.Next
		if h.RType == KnotEndpoint {
			break
		}
		if h == knots {
			h.LType = KnotEndCycle
			break
		}
	}

	cur = h
	for {
		q := cur.Next
		if cur.RType >= KnotGiven {
			// mp.c:7392-7396 — skip over open/open runs
			for q.LType == KnotOpen && q.RType == KnotOpen {
				q = q.Next
			}
			// mp.c:7466-7486 — open -> curl/given based on control deltas
			delx := q.RightX - q.XCoord
			dely := q.RightY - q.YCoord
			if q.LType == KnotOpen {
				if q.RType == KnotCurl {
					// Keep curl information instead of deriving a direction from bogus deltas.
					q.LType = KnotCurl
					q.LeftX = q.RightX // curl stored in right_x
				} else if numberZero(delx) && numberZero(dely) {
					q.LType = KnotCurl
					q.LeftX = unity // curl stored in left_x
				} else {
					q.LType = KnotGiven
					q.LeftX = nArg(delx, dely) // angle stored in left_x
				}
			}
			if cur.RType == KnotOpen && cur.LType == KnotExplicit {
				delx = cur.XCoord - cur.LeftX
				dely = cur.YCoord - cur.LeftY
				if numberZero(delx) && numberZero(dely) {
					cur.RType = KnotCurl
					cur.RightX = unity // curl stored in right_x
				} else {
					cur.RType = KnotGiven
					cur.RightX = nArg(delx, dely) // angle stored in right_x
				}
			}
			n := e.computePsiTheta(cur, q)
			e.solveChoices(cur, q, n) // mp.c:7495
		} else if cur.RType == KnotEndpoint {
			// mp.c:7505-7509 — endpoint: control points equal to knot coords.
			cur.RightX = cur.XCoord
			cur.RightY = cur.YCoord
			q.LeftX = q.XCoord
			q.LeftY = q.YCoord
		}
		cur = q
		if cur.RType == KnotEndpoint {
			break
		}
		if cur == h {
			break
		}
	}

	// Tracing/mp->arith_error handling from mp.c is omitted for now.
	return nil
}

// getTurnAmt mirrors mp_get_turn_amt (mp.c:14208ff / mp.w ~14208ff) and
// returns the signed turn count as an integer.
func getTurnAmt(w *Knot, dx, dy Number, ccw bool) int {
	// mp_get_turn_amt (mp.c:14208ff / mp.w ~14208ff).
	if w == nil {
		return 0
	}
	s := 0
	if ccw {
		ww := w.Next
		for {
			arg1 := ww.XCoord - w.XCoord
			arg2 := ww.YCoord - w.YCoord
			t := abVsCd(dy, arg1, dx, arg2)
			if numberNegative(t) {
				break
			}
			s++
			w = ww
			ww = ww.Next
			if !numberPositive(t) {
				break
			}
		}
	} else {
		ww := w.Prev
		arg1 := w.XCoord - ww.XCoord
		arg2 := w.YCoord - ww.YCoord
		t := abVsCd(dy, arg1, dx, arg2)
		for numberNegative(t) {
			s--
			w = ww
			ww = ww.Prev
			arg1 = w.XCoord - ww.XCoord
			arg2 = w.YCoord - ww.YCoord
			t = abVsCd(dy, arg1, dx, arg2)
		}
	}
	return s
}

// ensurePathCapacity grows working buffers similar to mp_reallocate_paths (mp.c:7548-7569).
func (e *Engine) ensurePathCapacity(size int) {
	if size <= e.pathSize && size <= len(e.deltaX) {
		return
	}
	newSize := size
	if newSize < e.pathSize+(e.pathSize/4) {
		newSize = e.pathSize + (e.pathSize / 4)
	}
	extend := func(s []Number, n int) []Number {
		if len(s) >= n {
			return s
		}
		oldLen := len(s)
		s = append(s, make([]Number, n-oldLen)...)
		return s
	}
	e.deltaX = extend(e.deltaX, newSize)
	e.deltaY = extend(e.deltaY, newSize)
	e.delta = extend(e.delta, newSize)
	e.psi = extend(e.psi, newSize)
	e.theta = extend(e.theta, newSize)
	e.uu = extend(e.uu, newSize)
	e.vv = extend(e.vv, newSize)
	e.ww = extend(e.ww, newSize)
	e.pathSize = newSize
}

// countSegments counts edges from p up to (but not including) stop, respecting cycles.
func (e *Engine) countSegments(p, stop *Knot) int {
	if p == nil {
		return 0
	}
	k := 0
	cur := p
	for {
		k++
		cur = cur.Next
		if cur == nil || cur == stop {
			break
		}
		if cur == p {
			break
		}
	}
	return k
}

// solveChoices mirrors mp_solve_choices (mp.c:7575ff) — full control solver
// for a run of n segments starting at p and ending before q.
func (e *Engine) solveChoices(p, q *Knot, n int) {
	if p == nil || q == nil || n <= 0 {
		return
	}
	t := p.Next
	e.deltaX[0] = t.XCoord - p.XCoord
	e.deltaY[0] = t.YCoord - p.YCoord

	// Main loop mirrors mp_solve_choices (mp.c:7575ff / mp.w:8238ff).
	s := p
	var r *Knot
	k := 0
	found := false
	for !found {
		t = s.Next
		if k == 0 {
			switch s.RType {
			case KnotGiven:
				if t.LType == KnotGiven {
					// mp.c:7591-7612 — both directions given.
					narg := nArg(e.deltaX[0], e.deltaY[0])
					e.ct, e.st = numberSinCos(s.RightX - narg) // right_given stored in RightX
					e.cf, e.sf = numberSinCos(t.LeftX - narg)  // left_given stored in LeftX
					e.sf = numberNegate(e.sf)
					e.setControls(s, t, 0)
					return
				}
				// mp.c:7620-7632 — right given, left not given.
				narg := nArg(e.deltaX[0], e.deltaY[0])
				e.vv[0] = reduceAngle(s.RightX - narg)
				e.uu[0] = 0
				e.ww[0] = 0
			case KnotCurl:
				if t.LType == KnotCurl {
					// mp.c:8942-9014 — curl/curl explicit controls (k==0).
					rt := numberAbsVal(s.RightY) // right_tension in RightY
					lt := numberAbsVal(t.LeftY)  // left_tension in LeftY
					s.RType = KnotExplicit
					t.LType = KnotExplicit
					if numberEqual(rt, unity) {
						p.RightX = s.XCoord + numberDivideInt(e.deltaX[0], 3)
						p.RightY = s.YCoord + numberDivideInt(e.deltaY[0], 3)
					} else {
						ff := makeFraction(unity, numberMultiplyInt(rt, 3))
						p.RightX = s.XCoord + takeFraction(e.deltaX[0], ff)
						p.RightY = s.YCoord + takeFraction(e.deltaY[0], ff)
					}
					if numberEqual(lt, unity) {
						t.LeftX = t.XCoord - numberDivideInt(e.deltaX[0], 3)
						t.LeftY = t.YCoord - numberDivideInt(e.deltaY[0], 3)
					} else {
						ff := makeFraction(unity, numberMultiplyInt(lt, 3))
						t.LeftX = t.XCoord - takeFraction(e.deltaX[0], ff)
						t.LeftY = t.YCoord - takeFraction(e.deltaY[0], ff)
					}
					return
				}
				// mp.c:8602-8625 — curl boundary (non-curl neighbor).
				rt := numberAbsVal(s.RightY)
				lt := numberAbsVal(t.LeftY)
				cc := s.RightX // curl stored in RightX
				if numberEqual(rt, unity) && numberEqual(lt, unity) {
					if math.IsInf(float64(cc), 1) || cc > math.MaxFloat64/4 {
						// limit cc->inf of (2cc+1)/(cc+2) = 2.
						e.uu[0] = 2
					} else {
						num := numberAdd(numberDouble(cc), unity)
						den := numberAdd(cc, 2)
						e.uu[0] = makeFraction(num, den)
					}
				} else {
					e.uu[0] = curlRatio(cc, rt, lt)
				}
				e.vv[0] = numberNegate(takeFraction(e.psi[1], e.uu[0]))
				e.ww[0] = 0
			default:
				// mp.c:7692ff — open fallback.
				e.uu[0] = 0
				e.vv[0] = 0
				e.ww[0] = fractionOne
			}
		} else {
			switch s.LType {
			case KnotEndCycle, KnotOpen:
				// mp.c:8325ff (blocks 358-361) — solve interior opens.
				e.deltaX[k] = t.XCoord - s.XCoord
				e.deltaY[k] = t.YCoord - s.YCoord
				e.delta[k] = pythAdd(e.deltaX[k], e.deltaY[k])

				aa := fractionHalf
				bb := fractionHalf
				cc := fractionOne
				dd := numberDouble(e.delta[k])
				ee := numberDouble(e.delta[k-1])

				// rtPrev from r.right_tension, ltNext from t.left_tension.
				rtPrev := numberAbsVal(r.RightY)
				ltNext := numberAbsVal(t.LeftY)
				if !numberEqual(rtPrev, unity) {
					arg2 := numberMultiplyInt(rtPrev, 3)
					arg2 = numberSub(arg2, unity)
					aa = makeFraction(unity, arg2)
					ret := makeFraction(unity, rtPrev)
					arg1 := numberSub(fractionThree, ret)
					dd = takeFraction(e.delta[k], arg1)
				}
				if !numberEqual(ltNext, unity) {
					arg2 := numberMultiplyInt(ltNext, 3)
					arg2 = numberSub(arg2, unity)
					bb = makeFraction(unity, arg2)
					ret := makeFraction(unity, ltNext)
					arg1 := numberSub(fractionThree, ret)
					ee = takeFraction(e.delta[k-1], arg1)
				}
				// cc = 1 - uu[k-1]*aa
				r1 := takeFraction(e.uu[k-1], aa)
				cc = numberSub(fractionOne, r1)
				// dd <- dd*cc (mp.c:8426 uses take_fraction(dd, dd, cc))
				dd = takeFraction(dd, cc)

				// adjust dd/ee if |left_tension| != |right_tension|
				ltS := numberAbsVal(s.LeftY)
				rtS := numberAbsVal(s.RightY)
				if !numberEqual(ltS, rtS) {
					if ltS < rtS {
						r1 = makeFraction(ltS, rtS)
						ff := takeFraction(r1, r1)
						dd = takeFraction(dd, ff)
					} else {
						r1 = makeFraction(rtS, ltS)
						ff := takeFraction(r1, r1)
						ee = takeFraction(ee, ff)
					}
				}

				arg2 := numberAdd(dd, ee)
				ff := makeFraction(ee, arg2) // ff = ee/(dd+ee)
				e.uu[k] = takeFraction(ff, bb)

				acc := takeFraction(e.psi[k+1], e.uu[k])
				acc = numberNegate(acc)
				if r.RType == KnotCurl {
					arg2 = numberSub(fractionOne, ff)
					r1 = takeFraction(e.psi[1], arg2)
					e.ww[k] = 0
					e.vv[k] = numberSub(acc, r1)
				} else {
					arg1 := numberSub(fractionOne, ff)
					ff = makeFraction(arg1, cc)
					r1 = takeFraction(e.psi[k], ff)
					acc = numberSub(acc, r1)
					r1copy := ff
					ff = takeFraction(r1copy, aa)
					r1 = takeFraction(e.vv[k-1], ff)
					e.vv[k] = numberSub(acc, r1)
					if numberZero(e.ww[k-1]) {
						e.ww[k] = 0
					} else {
						e.ww[k] = takeFraction(e.ww[k-1], ff)
						e.ww[k] = numberNegate(e.ww[k])
					}
				}

				if s.LType == KnotEndCycle {
					// mp.c:8519ff — close cycle and seed theta[n].
					aa = 0
					bb = fractionOne
					for {
						k--
						if k == 0 {
							k = n
						}
						r1 = takeFraction(aa, e.uu[k])
						aa = numberSub(e.vv[k], r1)
						r1 = takeFraction(bb, e.uu[k])
						bb = numberSub(e.ww[k], r1)
						if k == n {
							break
						}
					}
					arg2 = numberSub(fractionOne, bb)
					r1 = makeFraction(aa, arg2)
					e.theta[n] = r1
					e.vv[0] = r1
					for k = 1; k < n; k++ {
						adj := takeFraction(r1, e.ww[k])
						e.vv[k] = numberAdd(e.vv[k], adj)
					}
					found = true
				}
			case KnotCurl:
				// mp.c:8637ff — left curl hits FOUND.
				cc := s.LeftX               // curl in left_x
				lt := numberAbsVal(s.LeftY) // left tension
				rt := numberAbsVal(r.RightY)
				var ff Number
				if numberEqual(rt, unity) && numberEqual(lt, unity) {
					ff = makeFraction(numberAdd(numberDouble(cc), unity), numberAdd(cc, 2))
				} else {
					ff = curlRatio(cc, lt, rt)
				}
				arg1 := takeFraction(e.vv[n-1], ff)
				r1 := takeFraction(ff, e.uu[n-1])
				arg2 := numberSub(fractionOne, r1)
				e.theta[n] = makeFraction(arg1, arg2)
				e.theta[n] = numberNegate(e.theta[n])
				found = true
			case KnotGiven:
				// mp.c:8577ff — left given sets theta[n] then FOUND.
				narg := nArg(e.deltaX[n-1], e.deltaY[n-1])
				e.theta[n] = reduceAngle(s.LeftX - narg)
				found = true
			}
		}
		if found {
			break
		}
		r = s
		s = t
		k++
	}

	// Backward propagation of theta (mp.c:8755ff / mp.w ~8755ff).
	for k := n - 1; k >= 0; k-- {
		r1 := takeFraction(e.theta[k+1], e.uu[k])
		e.theta[k] = numberSub(e.vv[k], r1)
	}

	// Final control computation (mp.c:8766ff).
	s = p
	for k = 0; k < n; k++ {
		t = s.Next
		e.ct, e.st = numberSinCos(e.theta[k])
		arg := numberNegate(numberAdd(e.psi[k+1], e.theta[k+1]))
		e.cf, e.sf = numberSinCos(arg)
		e.setControls(s, t, k)
		s = t
	}
}

// setControls mirrors mp_set_controls (mp.c:8205ff) — simplified.
// Uses precomputed trig (e.st/ct/sf/cf) and tensions in left_y/right_y.
func (e *Engine) setControls(p, q *Knot, k int) {
	lt := numberAbsVal(q.LeftY)  // left tension in left_y
	rt := numberAbsVal(p.RightY) // right tension in right_y
	rr := velocity(e.st, e.ct, e.sf, e.cf, rt)
	ss := velocity(e.sf, e.cf, e.st, e.ct, lt)

	// Negative tension correction (mp.c:8066-8114 / mp.w ~8874ff).
	if p.RightY < 0 || q.LeftY < 0 {
		if (numberNonnegative(e.st) && numberNonnegative(e.sf)) ||
			(numberNonpositive(e.st) && numberNonpositive(e.sf)) {
			sine := numberAdd(
				takeFraction(numberAbsVal(e.st), e.cf),
				takeFraction(numberAbsVal(e.sf), e.ct),
			)
			if numberPositive(sine) {
				if p.RightY < 0 {
					ab := abVsCd(numberAbsVal(e.sf), fractionOne, rr, sine)
					if ab < 0 {
						rr = makeFraction(numberAbsVal(e.sf), sine)
					}
				}
				if q.LeftY < 0 {
					ab := abVsCd(numberAbsVal(e.st), fractionOne, ss, sine)
					if ab < 0 {
						ss = makeFraction(numberAbsVal(e.st), sine)
					}
				}
			}
		}
	}

	// Compute control points (mp.c:8119-8138 / mp.w ~8835ff).
	r1 := takeFraction(e.deltaX[k], e.ct)
	r2 := takeFraction(e.deltaY[k], e.st)
	tmp := numberSub(r1, r2)
	tmp = takeFraction(tmp, rr)
	p.RightX = numberAdd(p.XCoord, tmp)

	r1 = takeFraction(e.deltaY[k], e.ct)
	r2 = takeFraction(e.deltaX[k], e.st)
	tmp = numberAdd(r1, r2)
	tmp = takeFraction(tmp, rr)
	p.RightY = numberAdd(p.YCoord, tmp)

	r1 = takeFraction(e.deltaX[k], e.cf)
	r2 = takeFraction(e.deltaY[k], e.sf)
	tmp = numberAdd(r1, r2)
	tmp = takeFraction(tmp, ss)
	q.LeftX = numberSub(q.XCoord, tmp)

	r1 = takeFraction(e.deltaY[k], e.cf)
	r2 = takeFraction(e.deltaX[k], e.sf)
	tmp = numberSub(r1, r2)
	tmp = takeFraction(tmp, ss)
	q.LeftY = numberSub(q.YCoord, tmp)

	p.RType = KnotExplicit
	q.LType = KnotExplicit
}
