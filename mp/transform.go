package mp

import "math"

// Transform represents an affine transformation matrix.
// The transformation is applied as:
//
//	x' = Txx*x + Txy*y + Tx
//	y' = Tyx*x + Tyy*y + Ty
//
// This mirrors MetaPost's transform type (mp.w:6196ff).
// The matrix form is:
//
//	| Txx  Txy  Tx |
//	| Tyx  Tyy  Ty |
//	| 0    0    1  |
type Transform struct {
	Txx, Txy, Tx Number // first row: x' = Txx*x + Txy*y + Tx
	Tyx, Tyy, Ty Number // second row: y' = Tyx*x + Tyy*y + Ty
}

// Identity returns the identity transformation.
// (mp.w:6367ff mp_id_transform)
func Identity() Transform {
	return Transform{
		Txx: 1, Txy: 0, Tx: 0,
		Tyx: 0, Tyy: 1, Ty: 0,
	}
}

// Shifted returns a translation transformation.
// Mirrors MetaPost's "shifted (dx, dy)" (mp.w:28509ff).
func Shifted(dx, dy Number) Transform {
	return Transform{
		Txx: 1, Txy: 0, Tx: dx,
		Tyx: 0, Tyy: 1, Ty: dy,
	}
}

// Scaled returns a uniform scaling transformation around the origin.
// Mirrors MetaPost's "scaled s" (mp.w:28502ff).
func Scaled(s Number) Transform {
	return Transform{
		Txx: s, Txy: 0, Tx: 0,
		Tyx: 0, Tyy: s, Ty: 0,
	}
}

// XScaled returns a horizontal scaling transformation.
// Mirrors MetaPost's "xscaled s".
func XScaled(s Number) Transform {
	return Transform{
		Txx: s, Txy: 0, Tx: 0,
		Tyx: 0, Tyy: 1, Ty: 0,
	}
}

// YScaled returns a vertical scaling transformation.
// Mirrors MetaPost's "yscaled s".
func YScaled(s Number) Transform {
	return Transform{
		Txx: 1, Txy: 0, Tx: 0,
		Tyx: 0, Tyy: s, Ty: 0,
	}
}

// Rotated returns a rotation transformation around the origin.
// Angle is in degrees (positive = counter-clockwise).
// Mirrors MetaPost's "rotated angle" (mp.w:28537ff).
func Rotated(angleDeg Number) Transform {
	// Convert to radians
	rad := angleDeg * math.Pi / 180.0
	cos := Number(math.Cos(float64(rad)))
	sin := Number(math.Sin(float64(rad)))
	return Transform{
		Txx: cos, Txy: -sin, Tx: 0,
		Tyx: sin, Tyy: cos, Ty: 0,
	}
}

// Slanted returns a horizontal shear transformation.
// Mirrors MetaPost's "slanted s" (mp.w:28496ff).
// The transformation is: x' = x + s*y, y' = y
func Slanted(s Number) Transform {
	return Transform{
		Txx: 1, Txy: s, Tx: 0,
		Tyx: 0, Tyy: 1, Ty: 0,
	}
}

// ZScaled returns a scaling+rotation transformation using a complex number.
// Mirrors MetaPost's "zscaled (a, b)" which scales by sqrt(a²+b²) and
// rotates by atan2(b, a).
func ZScaled(a, b Number) Transform {
	return Transform{
		Txx: a, Txy: -b, Tx: 0,
		Tyx: b, Tyy: a, Ty: 0,
	}
}

// RotatedAround returns a rotation transformation around a given point.
// This is equivalent to: shifted(-cx,-cy) rotated(angle) shifted(cx,cy)
func RotatedAround(cx, cy, angleDeg Number) Transform {
	return Shifted(-cx, -cy).Then(Rotated(angleDeg)).Then(Shifted(cx, cy))
}

// ScaledAround returns a scaling transformation around a given point.
// This is equivalent to: shifted(-cx,-cy) scaled(s) shifted(cx,cy)
func ScaledAround(cx, cy, s Number) Transform {
	return Shifted(-cx, -cy).Then(Scaled(s)).Then(Shifted(cx, cy))
}

// ReflectedAbout returns a reflection transformation about the line
// passing through points (x1,y1) and (x2,y2).
// Mirrors MetaPost's "reflectedabout(p1, p2)" (plain.mp).
func ReflectedAbout(x1, y1, x2, y2 Number) Transform {
	// Reflection about a line through origin at angle θ:
	// | cos(2θ)   sin(2θ)  0 |
	// | sin(2θ)  -cos(2θ)  0 |
	// | 0         0        1 |
	//
	// For a line through (x1,y1) and (x2,y2):
	// 1. Translate so (x1,y1) is at origin
	// 2. Reflect about line through origin
	// 3. Translate back
	dx := x2 - x1
	dy := y2 - y1
	// θ = atan2(dy, dx), so 2θ rotation
	// cos(2θ) = cos²θ - sin²θ = (dx² - dy²) / (dx² + dy²)
	// sin(2θ) = 2·sinθ·cosθ = 2·dx·dy / (dx² + dy²)
	d2 := dx*dx + dy*dy
	if d2 == 0 {
		return Identity() // degenerate case: same point
	}
	cos2 := (dx*dx - dy*dy) / d2
	sin2 := 2 * dx * dy / d2
	// Reflection matrix about line through origin
	reflect := Transform{
		Txx: cos2, Txy: sin2, Tx: 0,
		Tyx: sin2, Tyy: -cos2, Ty: 0,
	}
	// Compose: shift to origin, reflect, shift back
	return Shifted(-x1, -y1).Then(reflect).Then(Shifted(x1, y1))
}

// Then composes this transformation with another.
// Returns a transformation equivalent to applying t first, then other.
// (i.e., other ∘ t in mathematical notation)
func (t Transform) Then(other Transform) Transform {
	// Matrix multiplication:
	// | other.Txx  other.Txy  other.Tx |   | t.Txx  t.Txy  t.Tx |
	// | other.Tyx  other.Tyy  other.Ty | × | t.Tyx  t.Tyy  t.Ty |
	// | 0          0          1        |   | 0      0      1    |
	return Transform{
		Txx: other.Txx*t.Txx + other.Txy*t.Tyx,
		Txy: other.Txx*t.Txy + other.Txy*t.Tyy,
		Tx:  other.Txx*t.Tx + other.Txy*t.Ty + other.Tx,
		Tyx: other.Tyx*t.Txx + other.Tyy*t.Tyx,
		Tyy: other.Tyx*t.Txy + other.Tyy*t.Tyy,
		Ty:  other.Tyx*t.Tx + other.Tyy*t.Ty + other.Ty,
	}
}

// ApplyToPoint applies the transformation to a point (x, y).
// Returns the transformed coordinates (x', y').
// Mirrors MetaPost's mp_number_trans (mp.w:28617ff).
func (t Transform) ApplyToPoint(x, y Number) (Number, Number) {
	xp := t.Txx*x + t.Txy*y + t.Tx
	yp := t.Tyx*x + t.Tyy*y + t.Ty
	return xp, yp
}

// ApplyToKnot applies the transformation to all coordinates of a knot.
func (t Transform) ApplyToKnot(k *Knot) {
	if k == nil {
		return
	}
	k.XCoord, k.YCoord = t.ApplyToPoint(k.XCoord, k.YCoord)
	k.LeftX, k.LeftY = t.ApplyToPoint(k.LeftX, k.LeftY)
	k.RightX, k.RightY = t.ApplyToPoint(k.RightX, k.RightY)
}

// ApplyToPath applies the transformation to all knots in a path.
// Mirrors MetaPost's mp_do_path_trans (mp.w:28647ff).
// Returns a new transformed path (does not modify the original).
func (t Transform) ApplyToPath(p *Path) *Path {
	if p == nil || p.Head == nil {
		return nil
	}
	// Copy the path first
	result := p.Copy()
	// Apply transform to each knot
	cur := result.Head
	for {
		t.ApplyToKnot(cur)
		cur = cur.Next
		if cur == nil || cur == result.Head {
			break
		}
	}
	return result
}

// Inverse returns the inverse transformation, if it exists.
// Returns Identity() if the transformation is singular (determinant = 0).
func (t Transform) Inverse() Transform {
	det := t.Txx*t.Tyy - t.Txy*t.Tyx
	if det == 0 {
		return Identity()
	}
	invDet := 1.0 / det
	return Transform{
		Txx: t.Tyy * invDet,
		Txy: -t.Txy * invDet,
		Tx:  (t.Txy*t.Ty - t.Tyy*t.Tx) * invDet,
		Tyx: -t.Tyx * invDet,
		Tyy: t.Txx * invDet,
		Ty:  (t.Tyx*t.Tx - t.Txx*t.Ty) * invDet,
	}
}

// Determinant returns the determinant of the transformation matrix.
// This represents the scaling factor for areas.
func (t Transform) Determinant() Number {
	return t.Txx*t.Tyy - t.Txy*t.Tyx
}

// Path transformation methods

// Shifted returns a new path shifted by (dx, dy).
func (p *Path) Shifted(dx, dy Number) *Path {
	return Shifted(dx, dy).ApplyToPath(p)
}

// Scaled returns a new path scaled uniformly around the origin.
func (p *Path) Scaled(s Number) *Path {
	return Scaled(s).ApplyToPath(p)
}

// Rotated returns a new path rotated around the origin.
// Angle is in degrees (positive = counter-clockwise).
func (p *Path) Rotated(angleDeg Number) *Path {
	return Rotated(angleDeg).ApplyToPath(p)
}

// Slanted returns a new path with horizontal shear applied.
func (p *Path) Slanted(s Number) *Path {
	return Slanted(s).ApplyToPath(p)
}

// XScaled returns a new path scaled horizontally.
func (p *Path) XScaled(s Number) *Path {
	return XScaled(s).ApplyToPath(p)
}

// YScaled returns a new path scaled vertically.
func (p *Path) YScaled(s Number) *Path {
	return YScaled(s).ApplyToPath(p)
}

// ZScaled returns a new path scaled and rotated using complex multiplication.
func (p *Path) ZScaled(a, b Number) *Path {
	return ZScaled(a, b).ApplyToPath(p)
}

// Transformed returns a new path with the given transformation applied.
func (p *Path) Transformed(t Transform) *Path {
	return t.ApplyToPath(p)
}

// RotatedAround returns a new path rotated around a given point.
func (p *Path) RotatedAround(cx, cy, angleDeg Number) *Path {
	return RotatedAround(cx, cy, angleDeg).ApplyToPath(p)
}

// ScaledAround returns a new path scaled around a given point.
func (p *Path) ScaledAround(cx, cy, s Number) *Path {
	return ScaledAround(cx, cy, s).ApplyToPath(p)
}

// ReflectedAbout returns a new path reflected about the line through (x1,y1) and (x2,y2).
func (p *Path) ReflectedAbout(x1, y1, x2, y2 Number) *Path {
	return ReflectedAbout(x1, y1, x2, y2).ApplyToPath(p)
}
