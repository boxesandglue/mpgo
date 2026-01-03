# mpgo

A Go port of MetaPost's curve-solving engine. Implements the Hobby-Knuth algorithm for computing smooth Bézier curves from high-level path specifications.

## Features

- **Curve Solving**: The complete Hobby-Knuth algorithm for smooth curves
- **Path Construction**: Curves (`..`), lines (`--`), directions, tension, curl
- **Pens**: pencircle, pensquare, penrazor with full envelope computation
- **Transformations**: shifted, scaled, rotated, slanted, reflected, etc.
- **Path Operations**: point/direction of, subpath, arclength, intersections, buildcycle
- **Geometry Helpers**: midpoint, line intersection, rotation, reflection
- **Equation Solver**: Constraint-based coordinate computation
- **SVG Output**: With automatic viewBox and MetaPost-compatible coordinates

## Installation

```bash
go get github.com/boxesandglue/mpgo
```

## Quick Start

```go
package main

import (
    "os"

    "github.com/boxesandglue/mpgo/draw"
    "github.com/boxesandglue/mpgo/mp"
    "github.com/boxesandglue/mpgo/svg"
)

func main() {
    // Create a curved path
    path, _ := draw.NewPath().
        MoveTo(draw.P(0, 0)).
        CurveTo(draw.P(100, 100)).
        CurveTo(draw.P(200, 0)).
        WithStrokeColor(mp.ColorRGB(0, 0, 0)).
        Solve()

    // Output SVG
    f, _ := os.Create("output.svg")
    defer f.Close()

    svg.NewBuilder().
        AddPathFromPath(path).
        WriteTo(f)
}
```

## Examples

### Smooth Curves with Direction Hints

```go
// MetaPost: z0{right}..z1{up}..z2
path, _ := draw.NewPath().
    MoveTo(draw.P(0, 0)).
    WithDirection(0).           // {right}
    CurveTo(draw.P(50, 50)).
    WithDirection(90).          // {up}
    CurveTo(draw.P(100, 0)).
    Solve()
```

### Closed Paths (Cycles)

```go
// MetaPost: z0..z1..z2..cycle
path, _ := draw.NewPath().
    MoveTo(draw.P(0, 0)).
    CurveTo(draw.P(100, 50)).
    CurveTo(draw.P(50, 100)).
    Close().
    Solve()
```

### Tension Control

```go
// MetaPost: z0..tension 2..z1
path, _ := draw.NewPath().
    MoveTo(draw.P(0, 0)).
    WithTension(2).
    CurveTo(draw.P(100, 0)).
    Solve()
```

### Pen Strokes

```go
// MetaPost: draw p withpen pensquare scaled 4
path, _ := draw.NewPath().
    MoveTo(draw.P(0, 0)).
    LineTo(draw.P(100, 0)).
    WithPen(mp.PenSquare().Scaled(4)).
    Solve()
```

### Transformations

```go
// MetaPost: fullcircle scaled 50 shifted (100, 100)
circle := mp.FullCircle()
circle = mp.Scaled(50).ApplyToPath(circle)
circle = mp.Shifted(100, 100).ApplyToPath(circle)
```

### Path Operations

```go
// Get point at parameter t
pt := path.PointOf(0.5)

// Get subpath
sub := path.Subpath(0.25, 0.75)

// Find intersection
t1, t2, found := path1.IntersectionTimes(path2)

// Build closed region from multiple paths
region := mp.BuildCycle(path1, path2, path3, path4)
```

### Equation Solver

```go
// Solve for unknown coordinates
ctx := draw.NewContext()
z0 := ctx.Known(0, 0)
z1 := ctx.Known(100, 0)
mid := ctx.MidPointOf(z0, z1)    // computed: (50, 0)
top := ctx.Unknown()
ctx.Sum(top, mid, ctx.Known(0, 50))  // top = mid + (0, 50)

ctx.Solve()

// Use solved points in path
path, _ := ctx.NewPath().
    MoveToVar(z0).
    LineToVar(z1).
    LineToVar(top).
    Close().
    Solve()
```

## Package Structure

```
mpgo/
├── mp/       # Core types and algorithms
├── svg/      # SVG rendering
├── draw/   # Builder API (optional)
└── cmd/      # Example programs
```

## What This Is Not

This is a **library**, not a MetaPost interpreter. It does not support:
- MetaPost macros (`def`, `vardef`)
- Loops (`for`, `forever`)
- Conditionals (`if`, `else`)
- Text/labels (`btex...etex`)
- File I/O

Use this library when you want to generate MetaPost-quality curves programmatically in Go.

## Credits

This library is a port of [MetaPost](https://www.tug.org/metapost.html), originally created by John Hobby based on Donald Knuth's METAFONT. The MetaPost source code is in the public domain.

- Original author of MetaPost: John Hobby
- CWEB version: Taco Hoekwater
- Current maintainer: Luigi Scarso

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.
