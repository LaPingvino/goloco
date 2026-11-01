package gfx8

import "testing"

func TestTwoColourMap(t *testing.T) {
	p, s := make([]uint8, 256), make([]uint8, 256)
	for i := range p {
		p[i], s[i] = uint8(255-i), uint8(i^0x55)
	}
	m := TwoColourMap(p, s)
	for i := range 256 {
		want := uint8(i)
		for k, idx := range PrimaryRemap {
			if i == int(idx) {
				want = p[idx]
			}
			if i == int(SecondaryRemap[k]) {
				want = s[idx]
			}
		}
		if m[i] != want {
			t.Fatalf("entry %d = %d, want %d", i, m[i], want)
		}
	}
}
