package mp

import "math"

// Point represents a 2D coordinate pair.
// This is the basic type for geometric helper functions.
type Point struct {
	X, Y float64
}

// P creates a Point from x, y coordinates.
func P(x, y float64) Point {
	return Point{X: x, Y: y}
}

// MidPoint returns the midpoint between two points.
// Equivalent to MetaPost's "0.5[a,b]".
func MidPoint(a, b Point) Point {
	return Point{
		X: (a.X + b.X) / 2,
		Y: (a.Y + b.Y) / 2,
	}
}

// PointBetween returns the point at parameter t along the line from a to b.
// Equivalent to MetaPost's "t[a,b]".
//   - t=0 returns a
//   - t=1 returns b
//   - t=0.5 returns midpoint
//   - t<0 or t>1 extrapolates beyond the segment
func PointBetween(a, b Point, t float64) Point {
	return Point{
		X: a.X + t*(b.X-a.X),
		Y: a.Y + t*(b.Y-a.Y),
	}
}

// LineIntersection returns the intersection point of two lines.
// Line 1 passes through p1 and p2.
// Line 2 passes through p3 and p4.
// Returns the intersection point and true if lines intersect.
// Returns zero point and false if lines are parallel.
func LineIntersection(p1, p2, p3, p4 Point) (Point, bool) {
	// Line 1: p1 + t*(p2-p1)
	// Line 2: p3 + s*(p4-p3)
	// Solve: p1 + t*(p2-p1) = p3 + s*(p4-p3)

	d1x := p2.X - p1.X
	d1y := p2.Y - p1.Y
	d2x := p4.X - p3.X
	d2y := p4.Y - p3.Y

	// Cross product of directions (determinant)
	denom := d1x*d2y - d1y*d2x

	if math.Abs(denom) < 1e-10 {
		// Lines are parallel
		return Point{}, false
	}

	// Vector from p1 to p3
	dx := p3.X - p1.X
	dy := p3.Y - p1.Y

	// Parameter t for line 1
	t := (dx*d2y - dy*d2x) / denom

	return Point{
		X: p1.X + t*d1x,
		Y: p1.Y + t*d1y,
	}, true
}

// PointOnLineAtX returns the point on the line through p1 and p2 at the given x coordinate.
// Returns the point and true if the line is not vertical.
// Returns zero point and false if the line is vertical (infinite or no solutions).
func PointOnLineAtX(p1, p2 Point, x float64) (Point, bool) {
	dx := p2.X - p1.X
	if math.Abs(dx) < 1e-10 {
		// Vertical line - no unique solution
		return Point{}, false
	}
	t := (x - p1.X) / dx
	return Point{
		X: x,
		Y: p1.Y + t*(p2.Y-p1.Y),
	}, true
}

// PointOnLineAtY returns the point on the line through p1 and p2 at the given y coordinate.
// Returns the point and true if the line is not horizontal.
// Returns zero point and false if the line is horizontal (infinite or no solutions).
func PointOnLineAtY(p1, p2 Point, y float64) (Point, bool) {
	dy := p2.Y - p1.Y
	if math.Abs(dy) < 1e-10 {
		// Horizontal line - no unique solution
		return Point{}, false
	}
	t := (y - p1.Y) / dy
	return Point{
		X: p1.X + t*(p2.X-p1.X),
		Y: y,
	}, true
}

// PerpendicularFoot returns the point on the line through p1 and p2
// that is closest to point p (the foot of the perpendicular).
func PerpendicularFoot(p, p1, p2 Point) Point {
	// Direction vector of line
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	// Length squared
	lenSq := dx*dx + dy*dy
	if lenSq < 1e-20 {
		// p1 and p2 are the same point
		return p1
	}

	// Project (p - p1) onto direction vector
	t := ((p.X-p1.X)*dx + (p.Y-p1.Y)*dy) / lenSq

	return Point{
		X: p1.X + t*dx,
		Y: p1.Y + t*dy,
	}
}

// Distance returns the Euclidean distance between two points.
func Distance(a, b Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Reflection returns the reflection of point p about the line through p1 and p2.
func Reflection(p, p1, p2 Point) Point {
	foot := PerpendicularFoot(p, p1, p2)
	// Reflected point is on the opposite side of the line, same distance from foot
	return Point{
		X: 2*foot.X - p.X,
		Y: 2*foot.Y - p.Y,
	}
}

// Rotate returns point p rotated by angle degrees around the origin.
func Rotate(p Point, angle float64) Point {
	rad := angle * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return Point{
		X: p.X*cos - p.Y*sin,
		Y: p.X*sin + p.Y*cos,
	}
}

// RotateAround returns point p rotated by angle degrees around center point c.
func RotateAround(p, c Point, angle float64) Point {
	// Translate to origin, rotate, translate back
	translated := Point{X: p.X - c.X, Y: p.Y - c.Y}
	rotated := Rotate(translated, angle)
	return Point{
		X: rotated.X + c.X,
		Y: rotated.Y + c.Y,
	}
}

// Scale returns point p scaled by factor s from the origin.
func Scale(p Point, s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// ScaleAround returns point p scaled by factor s from center point c.
func ScaleAround(p, c Point, s float64) Point {
	return Point{
		X: c.X + s*(p.X-c.X),
		Y: c.Y + s*(p.Y-c.Y),
	}
}

// Add returns the vector sum of two points.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Sub returns the vector difference p - q.
func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

// Mul returns the point scaled by scalar s.
func (p Point) Mul(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Length returns the distance from the origin (vector magnitude).
func (p Point) Length() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

// Normalized returns a unit vector in the same direction.
// Returns zero vector if p is zero.
func (p Point) Normalized() Point {
	len := p.Length()
	if len < 1e-10 {
		return Point{}
	}
	return Point{X: p.X / len, Y: p.Y / len}
}

// Dot returns the dot product of two points (as vectors).
func (p Point) Dot(q Point) float64 {
	return p.X*q.X + p.Y*q.Y
}

// Cross returns the 2D cross product (z-component of 3D cross product).
// Positive if q is counter-clockwise from p.
func (p Point) Cross(q Point) float64 {
	return p.X*q.Y - p.Y*q.X
}

// Angle returns the angle of the vector in degrees (0-360).
func (p Point) Angle() float64 {
	return math.Atan2(p.Y, p.X) * 180 / math.Pi
}

// Dir returns a unit vector at the given angle in degrees.
// Equivalent to MetaPost's "dir(angle)".
func Dir(angle float64) Point {
	rad := angle * math.Pi / 180
	return Point{X: math.Cos(rad), Y: math.Sin(rad)}
}
