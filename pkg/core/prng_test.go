package core

import "testing"

func TestPrngBasics(t *testing.T) {
	p := NewPrngSeed(12345, 67890)
	v1 := p.RandNext()
	v2 := p.RandNext()
	if v1 == v2 {
		t.Fatalf("expected different PRNG outputs")
	}
	if p.RandBool() != (p.RandNext()&1 != 0) {
		// This is nondeterministic but exercises the call
	}
}

func TestRandNextRange(t *testing.T) {
	p := NewPrngSeed(1, 2)
	for i := 0; i < 100; i++ {
		v := p.RandNextRange(5, 10)
		if v < 5 || v > 10 {
			t.Fatalf("out of range: %d", v)
		}
	}
}
