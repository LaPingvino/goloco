package core

// Prng is a small pseudo-random generator ported from OpenLoco C++.
type Prng struct {
	s0 uint32
	s1 uint32
}

func NewPrng() *Prng                  { return &Prng{s0: 0, s1: 0} }
func NewPrngSeed(s0, s1 uint32) *Prng { return &Prng{s0: s0, s1: s1} }

func (p *Prng) Srand0() uint32 { return p.s0 }
func (p *Prng) Srand1() uint32 { return p.s1 }

// randNext implements an XORSHIFT-like PRNG adapted from OpenLoco's implementation.
func (p *Prng) RandNext() uint32 {
	// Simple xorshift algorithm based on two 32-bit state values.
	a := p.s0
	b := p.s1
	p.s0 = b
	a ^= a << 23
	a ^= a >> 17
	a ^= b ^ (b >> 26)
	p.s1 = a
	return p.s0 + p.s1
}

// RandNextBound returns random int32 in [0, high]
func (p *Prng) RandNextBound(high int32) int32 {
	if high <= 0 {
		return 0
	}
	positive := int32(p.RandNext() & 0x7FFFFFFF)
	return positive % (high + 1)
}

// RandNextRange returns random int32 in [low, high]
func (p *Prng) RandNextRange(low, high int32) int32 {
	if low > high {
		low, high = high, low
	}
	positive := int32(p.RandNext() & 0x7FFFFFFF)
	return low + (positive % ((high + 1) - low))
}

func (p *Prng) RandBool() bool {
	return (p.RandNext() & 1) != 0
}
