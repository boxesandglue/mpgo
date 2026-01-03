package draw

import (
	"errors"
	"fmt"
	"math"

	"github.com/boxesandglue/mpgo/mp"
)

// Context provides an equation solver for geometric constraints.
// It tracks variables (points) that can be known or unknown, and
// solves linear equations to determine unknown values.
//
// Usage:
//
//	ctx := draw.NewContext()
//	z0 := ctx.Known(0, 0)
//	z1 := ctx.Unknown()
//	z2 := ctx.Known(100, 100)
//	ctx.Collinear(z1, z0, z2)  // z1 on line z0--z2
//	ctx.EqX(z1, 50)            // z1.x = 50
//	ctx.Solve()
//	fmt.Println(z1.XY())       // (50, 50)
type Context struct {
	vars   []*Var
	eqns   []equation
	solved bool
}

// Var represents a point variable with x and y components.
// Components can be known (fixed value) or unknown (to be solved).
type Var struct {
	ctx    *Context
	index  int
	x, y   float64
	xKnown bool
	yKnown bool
}

// equation represents a linear equation: sum of (coeff * variable) = constant
// For point equations, we have separate equations for x and y components.
type equation struct {
	// coeffs maps variable index -> coefficient
	// For x-equations: index*2, for y-equations: index*2+1
	coeffs   map[int]float64
	constant float64
}

// NewContext creates a new equation-solving context.
func NewContext() *Context {
	return &Context{
		vars: make([]*Var, 0),
		eqns: make([]equation, 0),
	}
}

// Unknown creates a new unknown point variable.
func (c *Context) Unknown() *Var {
	v := &Var{
		ctx:   c,
		index: len(c.vars),
	}
	c.vars = append(c.vars, v)
	return v
}

// Known creates a new point variable with known coordinates.
func (c *Context) Known(x, y float64) *Var {
	v := &Var{
		ctx:    c,
		index:  len(c.vars),
		x:      x,
		y:      y,
		xKnown: true,
		yKnown: true,
	}
	c.vars = append(c.vars, v)
	return v
}

// Point creates a new unknown point variable (alias for Unknown).
func (c *Context) Point() *Var {
	return c.Unknown()
}

// Points creates n unknown point variables.
func (c *Context) Points(n int) []*Var {
	vars := make([]*Var, n)
	for i := 0; i < n; i++ {
		vars[i] = c.Unknown()
	}
	return vars
}

// XY returns the point's coordinates. Only valid after Solve().
func (v *Var) XY() (float64, float64) {
	return v.x, v.y
}

// X returns the x-coordinate. Only valid after Solve().
func (v *Var) X() float64 {
	return v.x
}

// Y returns the y-coordinate. Only valid after Solve().
func (v *Var) Y() float64 {
	return v.y
}

// Point returns the variable as an mp.Point. Only valid after Solve().
func (v *Var) Point() mp.Point {
	return mp.P(v.x, v.y)
}

// SetX sets the x-coordinate to a known value.
func (v *Var) SetX(x float64) *Var {
	v.x = x
	v.xKnown = true
	return v
}

// SetY sets the y-coordinate to a known value.
func (v *Var) SetY(y float64) *Var {
	v.y = y
	v.yKnown = true
	return v
}

// SetXY sets both coordinates to known values.
func (v *Var) SetXY(x, y float64) *Var {
	v.x = x
	v.y = y
	v.xKnown = true
	v.yKnown = true
	return v
}

// --- Equations ---

// Eq constrains a variable to equal a known point.
func (c *Context) Eq(v *Var, p mp.Point) {
	c.EqX(v, p.X)
	c.EqY(v, p.Y)
}

// EqX constrains a variable's x-coordinate to a value.
func (c *Context) EqX(v *Var, x float64) {
	eq := equation{
		coeffs:   map[int]float64{v.index * 2: 1},
		constant: x,
	}
	c.eqns = append(c.eqns, eq)
}

// EqY constrains a variable's y-coordinate to a value.
func (c *Context) EqY(v *Var, y float64) {
	eq := equation{
		coeffs:   map[int]float64{v.index*2 + 1: 1},
		constant: y,
	}
	c.eqns = append(c.eqns, eq)
}

// EqVar constrains two variables to be equal.
func (c *Context) EqVar(a, b *Var) {
	c.EqVarX(a, b)
	c.EqVarY(a, b)
}

// EqVarX constrains two variables to have equal x-coordinates.
func (c *Context) EqVarX(a, b *Var) {
	// a.x - b.x = 0
	eq := equation{
		coeffs:   map[int]float64{a.index * 2: 1, b.index * 2: -1},
		constant: 0,
	}
	c.eqns = append(c.eqns, eq)
}

// EqVarY constrains two variables to have equal y-coordinates.
func (c *Context) EqVarY(a, b *Var) {
	// a.y - b.y = 0
	eq := equation{
		coeffs:   map[int]float64{a.index*2 + 1: 1, b.index*2 + 1: -1},
		constant: 0,
	}
	c.eqns = append(c.eqns, eq)
}

// MidPoint constrains m to be the midpoint of a and b.
// Equivalent to: m = 0.5[a, b]
func (c *Context) MidPoint(m, a, b *Var) {
	c.Between(m, a, b, 0.5)
}

// Between constrains p to lie at parameter t on the line from a to b.
// Equivalent to: p = t[a, b] = (1-t)*a + t*b
func (c *Context) Between(p, a, b *Var, t float64) {
	// p.x = (1-t)*a.x + t*b.x
	// p.x - (1-t)*a.x - t*b.x = 0
	eqX := equation{
		coeffs: map[int]float64{
			p.index * 2: 1,
			a.index * 2: -(1 - t),
			b.index * 2: -t,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqX)

	// p.y = (1-t)*a.y + t*b.y
	eqY := equation{
		coeffs: map[int]float64{
			p.index*2 + 1: 1,
			a.index*2 + 1: -(1 - t),
			b.index*2 + 1: -t,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqY)
}

// Collinear constrains p to lie on the line through a and b.
// This adds the constraint that p, a, b are collinear, but doesn't
// determine WHERE on the line p is - you need another constraint for that.
func (c *Context) Collinear(p, a, b *Var) {
	// Collinearity: (p - a) × (b - a) = 0
	// (p.x - a.x)(b.y - a.y) - (p.y - a.y)(b.x - a.x) = 0
	//
	// This is NOT a linear equation in general. However, if a and b are known,
	// it becomes linear in p.
	//
	// For the general case, we use a parametric approach:
	// p = a + t*(b - a) for some t
	// This requires introducing an auxiliary variable t.
	//
	// For now, we handle the case where a and b are known:
	if a.xKnown && a.yKnown && b.xKnown && b.yKnown {
		// (p.x - a.x)(b.y - a.y) = (p.y - a.y)(b.x - a.x)
		// p.x*(b.y - a.y) - p.y*(b.x - a.x) = a.x*(b.y - a.y) - a.y*(b.x - a.x)
		dx := b.x - a.x
		dy := b.y - a.y
		eq := equation{
			coeffs: map[int]float64{
				p.index * 2:   dy,  // coeff for p.x
				p.index*2 + 1: -dx, // coeff for p.y
			},
			constant: a.x*dy - a.y*dx,
		}
		c.eqns = append(c.eqns, eq)
	}
	// TODO: Handle case where a or b are unknown (requires nonlinear solving)
}

// Intersection constrains p to be the intersection of line a1-a2 and line b1-b2.
// All of a1, a2, b1, b2 must be known.
func (c *Context) Intersection(p, a1, a2, b1, b2 *Var) error {
	if !a1.xKnown || !a1.yKnown || !a2.xKnown || !a2.yKnown ||
		!b1.xKnown || !b1.yKnown || !b2.xKnown || !b2.yKnown {
		return errors.New("Intersection requires all line endpoints to be known")
	}

	// Use the geometry helper to compute intersection
	pt, ok := mp.LineIntersection(
		mp.P(a1.x, a1.y), mp.P(a2.x, a2.y),
		mp.P(b1.x, b1.y), mp.P(b2.x, b2.y),
	)
	if !ok {
		return errors.New("lines are parallel, no intersection")
	}

	c.Eq(p, pt)
	return nil
}

// Sum constrains: result = a + b (vector addition)
func (c *Context) Sum(result, a, b *Var) {
	// result.x = a.x + b.x
	eqX := equation{
		coeffs: map[int]float64{
			result.index * 2: 1,
			a.index * 2:      -1,
			b.index * 2:      -1,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqX)

	// result.y = a.y + b.y
	eqY := equation{
		coeffs: map[int]float64{
			result.index*2 + 1: 1,
			a.index*2 + 1:      -1,
			b.index*2 + 1:      -1,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqY)
}

// Diff constrains: result = a - b (vector subtraction)
func (c *Context) Diff(result, a, b *Var) {
	// result.x = a.x - b.x
	eqX := equation{
		coeffs: map[int]float64{
			result.index * 2: 1,
			a.index * 2:      -1,
			b.index * 2:      1,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqX)

	// result.y = a.y - b.y
	eqY := equation{
		coeffs: map[int]float64{
			result.index*2 + 1: 1,
			a.index*2 + 1:      -1,
			b.index*2 + 1:      1,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqY)
}

// Scaled constrains: result = t * v (scalar multiplication)
func (c *Context) Scaled(result, v *Var, t float64) {
	// result.x = t * v.x
	eqX := equation{
		coeffs: map[int]float64{
			result.index * 2: 1,
			v.index * 2:      -t,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqX)

	// result.y = t * v.y
	eqY := equation{
		coeffs: map[int]float64{
			result.index*2 + 1: 1,
			v.index*2 + 1:      -t,
		},
		constant: 0,
	}
	c.eqns = append(c.eqns, eqY)
}

// --- Solver ---

// Solve solves the system of equations and updates all variables.
// Returns an error if the system is unsolvable or underdetermined.
func (c *Context) Solve() error {
	if c.solved {
		return nil
	}

	// First, apply known values as equations
	for _, v := range c.vars {
		if v.xKnown {
			c.eqns = append(c.eqns, equation{
				coeffs:   map[int]float64{v.index * 2: 1},
				constant: v.x,
			})
		}
		if v.yKnown {
			c.eqns = append(c.eqns, equation{
				coeffs:   map[int]float64{v.index*2 + 1: 1},
				constant: v.y,
			})
		}
	}

	// Number of variables (each point has x and y)
	numVars := len(c.vars) * 2
	if numVars == 0 {
		c.solved = true
		return nil
	}

	// Build augmented matrix for Gaussian elimination
	// Each row is an equation, columns are variable coefficients + constant
	numEqs := len(c.eqns)
	if numEqs < numVars {
		return fmt.Errorf("underdetermined system: %d equations for %d variables", numEqs, numVars)
	}

	matrix := make([][]float64, numEqs)
	for i, eq := range c.eqns {
		row := make([]float64, numVars+1)
		for varIdx, coeff := range eq.coeffs {
			if varIdx < numVars {
				row[varIdx] = coeff
			}
		}
		row[numVars] = eq.constant
		matrix[i] = row
	}

	// Gaussian elimination with partial pivoting
	solution, err := gaussianElimination(matrix, numVars)
	if err != nil {
		return err
	}

	// Apply solution to variables
	for i, v := range c.vars {
		v.x = solution[i*2]
		v.y = solution[i*2+1]
		v.xKnown = true
		v.yKnown = true
	}

	c.solved = true
	return nil
}

// gaussianElimination solves Ax = b using Gaussian elimination with partial pivoting.
func gaussianElimination(augmented [][]float64, numVars int) ([]float64, error) {
	numRows := len(augmented)
	if numRows == 0 {
		return make([]float64, numVars), nil
	}

	const eps = 1e-10

	// Forward elimination
	for col := 0; col < numVars; col++ {
		// Find pivot
		maxRow := -1
		maxVal := eps
		for row := col; row < numRows; row++ {
			if math.Abs(augmented[row][col]) > maxVal {
				maxVal = math.Abs(augmented[row][col])
				maxRow = row
			}
		}

		if maxRow == -1 {
			// No pivot found, variable might be free or determined by other equations
			continue
		}

		// Swap rows
		if maxRow != col && col < numRows {
			augmented[col], augmented[maxRow] = augmented[maxRow], augmented[col]
		}

		pivot := augmented[col][col]
		if math.Abs(pivot) < eps {
			continue
		}

		// Eliminate column
		for row := 0; row < numRows; row++ {
			if row == col {
				continue
			}
			factor := augmented[row][col] / pivot
			for j := col; j <= numVars; j++ {
				augmented[row][j] -= factor * augmented[col][j]
			}
		}
	}

	// Back substitution
	solution := make([]float64, numVars)
	for col := 0; col < numVars && col < numRows; col++ {
		pivot := augmented[col][col]
		if math.Abs(pivot) < eps {
			// Check if this is inconsistent
			if math.Abs(augmented[col][numVars]) > eps {
				return nil, fmt.Errorf("inconsistent system at variable %d", col)
			}
			// Variable is free, assume 0
			solution[col] = 0
		} else {
			solution[col] = augmented[col][numVars] / pivot
		}
	}

	return solution, nil
}

// --- PathBuilder integration ---

// NewPath creates a PathBuilder linked to this context.
// Use MoveToVar/CurveToVar/LineToVar to reference context variables.
// Call Solve() before building the path to resolve all variables.
func (c *Context) NewPath() *PathBuilder {
	return NewPath().WithContext(c)
}

// --- Geometry convenience methods ---

// MidPointOf returns a new variable constrained to be the midpoint of a and b.
func (c *Context) MidPointOf(a, b *Var) *Var {
	m := c.Unknown()
	c.MidPoint(m, a, b)
	return m
}

// IntersectionOf returns a new variable at the intersection of lines a1-a2 and b1-b2.
func (c *Context) IntersectionOf(a1, a2, b1, b2 *Var) (*Var, error) {
	p := c.Unknown()
	err := c.Intersection(p, a1, a2, b1, b2)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// BetweenAt returns a new variable at parameter t on line from a to b.
func (c *Context) BetweenAt(a, b *Var, t float64) *Var {
	p := c.Unknown()
	c.Between(p, a, b, t)
	return p
}
