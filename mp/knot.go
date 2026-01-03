package mp

type KnotType uint16

const (
	KnotEndpoint KnotType = iota
	KnotExplicit
	KnotGiven
	KnotCurl
	KnotOpen
	KnotEndCycle
)

type KnotOrigin uint8

const (
	OriginProgram KnotOrigin = iota
	OriginUser
)

// mplib.h 304
type Knot struct {
	XCoord Number
	YCoord Number
	LeftX  Number
	LeftY  Number
	RightX Number
	RightY Number
	Next   *Knot
	Prev   *Knot
	Info   int32
	LType  KnotType
	RType  KnotType
	Origin KnotOrigin
}

func NewKnot() *Knot {
	return &Knot{}
}

func CopyKnot(p *Knot) *Knot {
	if p == nil {
		return nil
	}
	q := *p
	return &q
}
