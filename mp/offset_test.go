package mp

import "testing"

func TestPenBBox(t *testing.T) {
	p := PenSquare(2)
	minx, miny, maxx, maxy, ok := PenBBox(p)
	if !ok {
		t.Fatalf("bbox not ok")
	}
	if minx != -1 || maxx != 1 || miny != -1 || maxy != 1 {
		t.Fatalf("unexpected bbox: %v %v %v %v", minx, miny, maxx, maxy)
	}
}

func TestPathNormals(t *testing.T) {
	path := NewPath()
	k1 := NewKnot()
	k1.XCoord, k1.YCoord = 0, 0
	k2 := NewKnot()
	k2.XCoord, k2.YCoord = 3, 4
	k1.Next = k2
	k2.Next = k1
	k1.RType = KnotEndpoint
	k2.RType = KnotEndpoint
	path.Head = k1

	normals := PathNormals(path)
	if len(normals) != 1 {
		t.Fatalf("expected 1 normal, got %d", len(normals))
	}
	if normals[0].Len != 5 {
		t.Fatalf("expected length 5, got %v", normals[0].Len)
	}
}
