package mp

import (
	"errors"
)

type Engine struct {
	paths []*Path
	// Path-working buffers (mp.c: delta_x/delta_y/delta/psi/theta/uu/vv/ww).
	pathSize int
	deltaX   []Number
	deltaY   []Number
	delta    []Number
	psi      []Number
	theta    []Number
	uu       []Number
	vv       []Number
	ww       []Number
	// Temp trig storage used by setControls (mp->st/ct/sf/cf in C).
	st, ct Number
	sf, cf Number
	// epsilon-like small value; align with mpmathdouble's epsilon_t use.
	epsilon Number
}

func NewEngine() *Engine {
	e := &Engine{
		paths:   make([]*Path, 0),
		epsilon: epsilon,
	}
	// mp.c mp_initt initializes path_size via mp_reallocate_paths(mp,1000).
	e.ensurePathCapacity(1000)
	return e
}

// AddPath appends a path to the engine queue.
func (e *Engine) AddPath(p *Path) {
	e.paths = append(e.paths, p)
}

// Solve runs the curve-solving and envelope pipeline on all paths.
func (e *Engine) Solve() error {
	if len(e.paths) == 0 {
		return errors.New("no paths loaded")
	}
	for _, p := range e.paths {
		if p == nil || p.Head == nil {
			continue
		}
		if err := e.solvePath(p); err != nil {
			return err
		}
		// After solving controls, compute a pen envelope for non-elliptical pens,
		// mirroring the offset/envelope phase (mp_apply_offset/mp_offset_prep, mp.c:13364ff, 15800ff).
		e.applyOffset(p)
	}
	return nil
}
