package core

import "testing"

// Golden vectors from OpenLoco src/Core/tests/PrngTests.cpp — the sequence
// must match Locomotion exactly for save/scenario parity.
func TestPrngGoldenVectors(t *testing.T) {
	p := NewPrng()
	if got := p.RandNext(); got != 0 {
		t.Errorf("zero-state RandNext() = %#x, want 0", got)
	}
	if p.Srand0() != 0xFE2468AC || p.Srand1() != 0 {
		t.Errorf("zero-state advance: s0=%#x s1=%#x, want s0=0xFE2468AC s1=0", p.Srand0(), p.Srand1())
	}

	p = NewPrngSeed(0x1234, 0x4321)
	if got := p.RandNext(); got != 0x80000246 {
		t.Errorf("seeded RandNext() = %#x, want 0x80000246", got)
	}
	if p.Srand0() != 0xBC247A5E || p.Srand1() != 0x80000246 {
		t.Errorf("seeded advance: s0=%#x s1=%#x, want s0=0xBC247A5E s1=0x80000246", p.Srand0(), p.Srand1())
	}

	for i := 0; i < 1000; i++ {
		p.RandNext()
	}
	if p.Srand0() != 0x0A597A43 || p.Srand1() != 0x12FC0827 {
		t.Errorf("after 1000 iterations: s0=%#x s1=%#x, want s0=0x0A597A43 s1=0x12FC0827", p.Srand0(), p.Srand1())
	}

	if got := p.RandNextBound(30); got != 11 {
		t.Errorf("RandNextBound(30) = %d, want 11", got)
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
