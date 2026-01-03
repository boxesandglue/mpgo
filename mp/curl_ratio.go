package mp

// curlRatio ports mp_curl_ratio (mp.c:8147ff / mp.w:8697ff).
// gamma = curl, a_tension = right tension, b_tension = left tension.
// Returns a fraction in TeX scaled units (we stay in Number with fractionMultiplier semantics).
func curlRatio(gamma, aTension, bTension Number) Number {
	alpha := makeFraction(unity, aTension) // alpha = 1 / a_tension
	beta := makeFraction(unity, bTension)  // beta = 1 / b_tension
	g := gamma
	var ff Number
	var denom Number
	// convert_fraction_to_scaled is a no-op here (we store as Number).
	if alpha <= beta {
		ff = makeFraction(alpha, beta)
		ff = takeFraction(ff, ff)
		g = takeFraction(g, ff)
		// convert_fraction_to_scaled(beta)
		beta = fractionToScaled(beta)
		denom = takeFraction(g, alpha)
		denom = numberAdd(denom, 3) // three_t
	} else {
		ff = makeFraction(beta, alpha)
		ff = takeFraction(ff, ff)
		tmp := takeFraction(beta, ff)
		beta = fractionToScaled(tmp)
		denom = takeFraction(g, alpha)
		denom = denom + divNumber(ff, twelveBits3) // set_number_from_div(ff, twelvebits_3)
	}
	denom = denom - beta
	num := takeFraction(g, (3 - alpha))
	num = numberAdd(num, beta)
	if num >= 4*denom {
		return fractionFour
	}
	return makeFraction(num, denom)
}
