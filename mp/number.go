package mp

import "math"

// Number aliases float64 to mirror the double backend only.
type Number = float64

// Constants aligned to mpmathdouble.c (mpmathdouble.w:22-47).
// epsilon = 2^-52, unity = 1.0, EL_GORDO = DBL_MAX/2-1, one_third_EL_GORDO = EL_GORDO/3.
const (
	unity    Number = 1.0
	zero     Number = 0.0
	epsilon  Number = 2.220446049250313e-16 // pow(2.0,-52.0)
	elGordo  Number = math.MaxFloat64/2.0 - 1.0
	oneThird Number = 1.0 / 3.0
	half     Number = 0.5
)

var (
	inf       Number = elGordo
	oneThirdI Number = elGordo / 3.0
	// ueps is not explicit in mpmathdouble; we mirror epsilon for now and can tune later if needed.
	ueps Number = epsilon

	// Thresholds and multipliers (mpmathdouble.w:22-47).
	fractionMultiplier Number = 4096.0
	angleMultiplier    Number = 16.0
	fractionThreshold  Number = 0.04096
	scaledThreshold    Number = 0.000122
	nearZeroAngle      Number = 0.0256 * angleMultiplier
	pOverVThreshold    int    = 0x80000
	equationThreshold  Number = 0.001
	tfmWarnThreshold   Number = 0.0625
	warningLimit       Number = math.Pow(2.0, 52.0)
	fractionHalf       Number = 0.5 * fractionMultiplier
	fractionOne        Number = 1.0 * fractionMultiplier
	fractionTwo        Number = 2.0 * fractionMultiplier
	fractionThree      Number = 3.0 * fractionMultiplier
	fractionFour       Number = 4.0 * fractionMultiplier
	arcTolK            Number = unity / 4096.0 // mpmathdouble.w:196-197
	twelveBits3        Number = 1365.0 / 65536.0
)

// AngleMultiplier exposes the angle scaling used for MetaPost angles (mpmathdouble angle_multiplier).
func AngleMultiplier() Number { return angleMultiplier }

// Fraction/Scaled helpers mirroring mpmathdouble conversion macros.
// convert_fraction_to_scaled: fraction (scaled by fractionMultiplier) -> scaled (unit = 1.0)
func fractionToScaled(x Number) Number {
	return x / fractionMultiplier
}

// convert_scaled_to_fraction: scaled -> fraction (scaled by fractionMultiplier)
func scaledToFraction(x Number) Number {
	return x * fractionMultiplier
}

// set_number_from_div equivalent: a = b / c
func divNumber(b, c Number) Number {
	return b / c
}

func twoToThe(a uint) int {
	return 1 << a
}

func numberAbs(a Number) Number {
	if a < 0 {
		return -a
	}
	return a
}

func numberToDouble(a Number) float64 {
	return float64(a)
}

// Inf exposes a large sentinel value used when mapping MetaPost "infinity".
func Inf() Number {
	return inf
}
