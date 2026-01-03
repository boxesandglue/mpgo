package mp

import "testing"

func TestAbVsCd(t *testing.T) {
	if v := abVsCd(2, 3, 1, 5); v != 1 {
		t.Fatalf("expected 1, got %v", v)
	}
	if v := abVsCd(1, 2, 2, 1); v != 0 {
		t.Fatalf("expected 0, got %v", v)
	}
	if v := abVsCd(1, 1, 2, 2); v != -1 {
		t.Fatalf("expected -1, got %v", v)
	}
}

func TestCrossingPoint(t *testing.T) {
	// a<0 -> zero_crossing
	if v := crossingPoint(-1, 0, 0); v != 0 {
		t.Fatalf("expected 0 for a<0, got %v", v)
	}
	// c>0, b>=0 -> no_crossing => fractionOne+1
	if v := crossingPoint(1, 1, 1); v != fractionOne+1 {
		t.Fatalf("expected no_crossing (%v), got %v", fractionOne+1, v)
	}
	// c==0, b>=0, a>0 -> one_crossing => fractionOne
	if v := crossingPoint(1, 1, 0); v != fractionOne {
		t.Fatalf("expected one_crossing (%v), got %v", fractionOne, v)
	}
}

func TestNumberNegateZero(t *testing.T) {
	if v := numberNegate(0); v != 0 {
		t.Fatalf("expected -0 handled as 0, got %v", v)
	}
}

func TestNumberNonpositive(t *testing.T) {
	if !numberNonpositive(0) {
		t.Fatalf("expected 0 to be nonpositive")
	}
	if numberNonpositive(1) {
		t.Fatalf("expected 1 to be positive")
	}
}
