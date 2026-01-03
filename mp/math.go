package mp

import "math"

// Math helpers mirroring mpmathdouble.c (double backend).

func makeScaled(p, q Number) Number {
	return p / q
}

func takeScaled(p, q Number) Number {
	return p * q
}

// Fractions follow mpmathdouble semantics: scale by fractionMultiplier (4096.0).
func makeFraction(p, q Number) Number {
	return (p / q) * fractionMultiplier
}

func takeFraction(p, q Number) Number {
	return (p * q) / fractionMultiplier
}

func roundUnscaled(x Number) int {
	return int(math.Floor(x + 0.5))
}

func numberFloor(x Number) Number {
	return math.Floor(x)
}

func numberFloorInPlace(x *Number) {
	*x = math.Floor(*x)
}

func numberSqrt(x Number) Number {
	return math.Sqrt(x)
}

// numberSinCos mirrors mp_double_sin_cos (mpmathdouble.c:1066ff):
// input angle z is in degrees scaled by angleMultiplier; outputs are
// cos/sin scaled by fractionMultiplier.
func numberSinCos(z Number) (cos Number, sin Number) {
	rad := (z / angleMultiplier) * math.Pi / 180.0
	switch z / angleMultiplier {
	case 90, -270:
		return 0, fractionMultiplier
	case -90, 270:
		return 0, -fractionMultiplier
	case 180, -180:
		return -fractionMultiplier, 0
	default:
		return math.Cos(rad) * fractionMultiplier, math.Sin(rad) * fractionMultiplier
	}
}

func numberNegate(x Number) Number {
	if x == 0 {
		return 0 // avoid -0.0
	}
	return -x
}

func numberAdd(a, b Number) Number {
	return a + b
}

func numberSub(a, b Number) Number {
	return a - b
}

func numberHalf(a Number) Number {
	return a / 2.0
}

func numberHalfp(a Number) Number {
	return a / 2.0
}

func numberDouble(a Number) Number {
	return a * 2.0
}

func numberAddScaled(a Number, b int) Number {
	return a + float64(b)
}

func numberMultiplyInt(a Number, b int) Number {
	return a * float64(b)
}

func numberDivideInt(a Number, b int) Number {
	return a / float64(b)
}

func numberAbsVal(a Number) Number {
	if a < 0 {
		return -a
	}
	return a
}

func numberPositive(a Number) bool {
	return a > 0
}

func numberNonnegative(a Number) bool {
	return a >= 0
}

func numberNonpositive(a Number) bool {
	return a <= 0
}

func numberNegative(a Number) bool {
	return a < 0
}

func numberZero(a Number) bool {
	return a == 0
}

func numberEqual(a, b Number) bool {
	return a == b
}

func numberLess(a, b Number) bool {
	return a < b
}

func numberGreater(a, b Number) bool {
	return a > b
}

// reduceAngle mirrors mp_reduce_angle (mp.c:7969ff / mp.w:8556ff):
// clamp to [-180,180] degrees scaled by angleMultiplier.
func reduceAngle(a Number) Number {
	oneEighty := 180 * angleMultiplier
	threeSixty := 360 * angleMultiplier
	if a > oneEighty {
		a -= threeSixty
	} else if a < -oneEighty {
		a += threeSixty
	}
	return a
}

// ofTheWay mirrors mp_set_double_from_of_the_way (mpmathdouble.w:397):
// A = B - take_fraction(B-C, t).
func ofTheWay(b, c, t Number) Number {
	return b - takeFraction(b-c, t)
}

func pythAdd(a, b Number) Number {
	return math.Hypot(a, b)
}

func pythSub(a, b Number) Number {
	v := a*a - b*b
	if v <= 0 {
		return 0
	}
	return math.Sqrt(v)
}

func nArg(x, y Number) Number {
	// mp_double_n_arg (mpmathdouble.c:1041ff) — returns degrees*angleMultiplier.
	return math.Atan2(y, x) * (180.0 / math.Pi) * angleMultiplier
}

// velocity mirrors mp_double_velocity (mpmathdouble.w:783-807).
func velocity(st, ct, sf, cf, t Number) Number {
	acc := takeFraction(st-(sf/16.0), sf-(st/16.0))
	acc = takeFraction(acc, ct-cf)
	num := fractionTwo + takeFraction(acc, math.Sqrt2*fractionOne)
	denom := fractionThree +
		takeFraction(ct, 3*fractionHalf*(math.Sqrt(5.0)-1.0)) +
		takeFraction(cf, 3*fractionHalf*(3.0-math.Sqrt(5.0)))
	if t != unity {
		num = makeScaled(num, t)
	}
	if num/4 >= denom {
		return fractionFour
	}
	return makeFraction(num, denom)
}

// abVsCd returns the sign of a*b - c*d (mpmathdouble.w:1479-1502).
func abVsCd(a, b, c, d Number) Number {
	ab := a * b
	cd := c * d
	switch {
	case ab > cd:
		return 1
	case ab < cd:
		return -1
	default:
		return 0
	}
}

// crossingPoint mirrors mp_double_crossing_point (mpmathdouble.w:931-993).
func crossingPoint(a, b, c Number) Number {
	if a < 0 {
		return 0 // zero_crossing
	}
	if c >= 0 {
		if b >= 0 {
			if c > 0 {
				return fractionOne + 1 // no_crossing
			}
			if a == 0 && b == 0 {
				return fractionOne + 1 // no_crossing
			}
			return fractionOne // one_crossing
		}
		if a == 0 {
			return 0 // zero_crossing
		}
	} else if a == 0 {
		if b <= 0 {
			return 0 // zero_crossing
		}
	}

	d := epsilon
	x0 := a
	x1 := a - b
	x2 := b - c
	for {
		x := (x1+x2)/2 + 1e-12
		if x1-x0 > x0 {
			x2 = x
			x0 += x0
			d += d
		} else {
			xx := x1 + x - x0
			if xx > x0 {
				x2 = x
				x0 += x0
				d += d
			} else {
				x0 = x0 - xx
				if x <= x0 {
					if x+x2 <= x0 {
						return fractionOne + 1 // no_crossing
					}
				}
				x1 = x
				d = d + d + epsilon
			}
		}
		if d >= fractionOne {
			break
		}
	}
	return d - fractionOne
}

// slowAdd mirrors mp_double_slow_add: here we simply add (overflow not handled).
func slowAdd(x, y Number) Number {
	return x + y
}

// squareRt with negative check similar to mp_double_square_rt.
func squareRt(x Number) Number {
	if x <= 0 {
		return 0
	}
	return math.Sqrt(x)
}

// pythSub with negative check similar to mp_double_pyth_sub.
func pythSubChecked(a, b Number) Number {
	a = math.Abs(a)
	b = math.Abs(b)
	if a <= b {
		if a < b {
			return 0
		}
		return 0
	}
	return math.Sqrt(a*a - b*b)
}
