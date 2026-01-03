package mp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Color carries color information in an output-agnostic way. CSS() returns a
// CSS-compatible color string for backends like SVG; Opacity() exposes a
// separate opacity channel when set.
type Color struct {
	css     string
	opacity float64 // NaN if unspecified
}

func (c Color) CSS() string { return c.css }
func (c Color) Opacity() (float64, bool) {
	if math.IsNaN(c.opacity) {
		return 0, false
	}
	return c.opacity, true
}

// ColorCSS uses the provided CSS color string verbatim.
func ColorCSS(css string) Color {
	if strings.HasPrefix(css, "#") {
		if c, op, ok := parseHexColor(css); ok {
			return Color{css: c, opacity: op}
		}
	}
	return Color{css: css, opacity: math.NaN()}
}

// ColorRGB expects components in [0,1] and formats an rgb() string.
func ColorRGB(r, g, b float64) Color {
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}
	r, g, b = clamp(r), clamp(g), clamp(b)
	return Color{css: fmt.Sprintf("rgb(%d,%d,%d)", int(r*255+0.5), int(g*255+0.5), int(b*255+0.5)), opacity: math.NaN()}
}

// ColorGray builds a grayscale rgb() string from gray in [0,1].
func ColorGray(gray float64) Color {
	return ColorRGB(gray, gray, gray)
}

// ColorRGBA expects components in [0,1] and formats an rgba() string.
func ColorRGBA(r, g, b, a float64) Color {
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}
	r, g, b, a = clamp(r), clamp(g), clamp(b), clamp(a)
	return Color{
		css:     fmt.Sprintf("rgb(%d,%d,%d)", int(r*255+0.5), int(g*255+0.5), int(b*255+0.5)),
		opacity: a,
	}
}

// ColorCMYK converts CMYK [0,1] to RGB and formats an rgb() string.
func ColorCMYK(c, m, y, k float64) Color {
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}
	c, m, y, k = clamp(c), clamp(m), clamp(y), clamp(k)
	r := (1 - c) * (1 - k)
	g := (1 - m) * (1 - k)
	b := (1 - y) * (1 - k)
	return ColorRGB(r, g, b)
}

// parseHexColor handles #RRGGBBAA and #RGBA forms; returns rgb() and opacity.
func parseHexColor(css string) (string, float64, bool) {
	switch len(css) {
	case 9: // #RRGGBBAA
		r, err1 := strconv.ParseUint(css[1:3], 16, 8)
		g, err2 := strconv.ParseUint(css[3:5], 16, 8)
		b, err3 := strconv.ParseUint(css[5:7], 16, 8)
		a, err4 := strconv.ParseUint(css[7:9], 16, 8)
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b), float64(a) / 255.0, true
		}
	case 5: // #RGBA (4-bit each)
		r, err1 := strconv.ParseUint(strings.Repeat(string(css[1]), 2), 16, 8)
		g, err2 := strconv.ParseUint(strings.Repeat(string(css[2]), 2), 16, 8)
		b, err3 := strconv.ParseUint(strings.Repeat(string(css[3]), 2), 16, 8)
		a, err4 := strconv.ParseUint(strings.Repeat(string(css[4]), 2), 16, 8)
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b), float64(a) / 255.0, true
		}
	}
	return "", 0, false
}
