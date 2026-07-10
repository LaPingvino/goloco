package core

import "math/bits"

// Prng is Locomotion's pseudo-random generator, ported from OpenLoco
// (src/Core/src/Prng.cpp). The exact sequence matters for parity with
// scenario/save state.
type Prng struct {
	s0 uint32
	s1 uint32
}

func NewPrng() *Prng                  { return &Prng{s0: 0, s1: 0} }
func NewPrngSeed(s0, s1 uint32) *Prng { return &Prng{s0: s0, s1: s1} }

func (p *Prng) Srand0() uint32 { return p.s0 }
func (p *Prng) Srand1() uint32 { return p.s1 }

func (p *Prng) RandNext() uint32 {
	s0 := p.s0
	p.s0 += bits.RotateLeft32(p.s1^0x1234567F, -7)
	p.s1 = bits.RotateLeft32(s0, -3)
	return p.s1
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
