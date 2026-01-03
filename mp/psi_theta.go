package mp

// computePsiTheta ports the psi/theta setup in mp_make_choices (mp.c:7398-7453 / mp.w ~8014ff).
// It fills deltaX/deltaY/delta and psi for a run from start up to (but not including) stop,
// stopping early if end_cycle is hit. Returns the count n (number of edges processed).
func (e *Engine) computePsiTheta(start, stop *Knot) int {
	// Ensure we have room for at least n+1 slots (n edges plus closing slot).
	e.ensurePathCapacity(e.pathSize + 1)
	k := 0
	s := start
	n := e.pathSize
	for {
		t := s.Next
		e.deltaX[k] = t.XCoord - s.XCoord
		e.deltaY[k] = t.YCoord - s.YCoord
		e.delta[k] = pythAdd(e.deltaX[k], e.deltaY[k])
		if k > 0 {
			// psi[k] = angle between segment k-1 and k
			r1 := makeFraction(e.deltaY[k-1], e.delta[k-1])
			sine := r1
			r2 := makeFraction(e.deltaX[k-1], e.delta[k-1])
			cosine := r2
			r1 = takeFraction(e.deltaX[k], cosine)
			r2 = takeFraction(e.deltaY[k], sine)
			arg1 := numberAdd(r1, r2)
			r1 = takeFraction(e.deltaY[k], cosine)
			r2 = takeFraction(e.deltaX[k], sine)
			arg2 := numberSub(r1, r2)
			e.psi[k] = nArg(arg1, arg2)
		}
		k++
		s = t
		if s == stop {
			n = k
		}
		// stop when k>=n && left_type(s)!=end_cycle
		if k >= n && s.LType != KnotEndCycle {
			break
		}
		if k == len(e.deltaX) {
			e.ensurePathCapacity(k + k/4 + 1)
		}
	}
	// Ensure psi has n+1 entries; set psi[n] per mp.c logic.
	if k >= len(e.psi) {
		e.ensurePathCapacity(k + 1)
	}
	if k == n {
		e.psi[k] = 0
	} else {
		e.psi[k] = e.psi[1]
	}
	return n
}
