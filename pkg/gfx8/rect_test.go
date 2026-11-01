package gfx8

import "testing"

// With an identity colour map, Shade(m, s) is PrimaryRemap[2+s], so the bevel
// pixels name the shade upstream's fillRectInset picks for each edge.
func TestRectInsetShades(t *testing.T) {
	id := make([]uint8, 256)
	for i := range id {
		id[i] = uint8(i)
	}
	c := InsetColour{Map: id}
	sh := func(s int) uint8 { return PrimaryRemap[2+s] }
	rt := NewRenderTarget(10, 10)
	DrawRectInset(rt, 0, 0, 10, 10, c, 0) // raised: hilight 7 top/left, shadow 3 right/bottom, fill 5
	at := func(x, y int) uint8 { return rt.Bits[y*rt.Width+x] }
	for _, ck := range []struct {
		x, y int
		want uint8
	}{{0, 0, sh(7)}, {5, 0, sh(7)}, {0, 5, sh(7)}, {9, 5, sh(3)}, {5, 9, sh(3)}, {5, 5, sh(5)}, {0, 9, sh(3)}} {
		if got := at(ck.x, ck.y); got != ck.want {
			t.Errorf("raised (%d,%d) = %d, want %d", ck.x, ck.y, got, ck.want)
		}
	}
	DrawRectInset(rt, 0, 0, 10, 10, c, InsetBorderInset) // inset: shadow top/left, hilight right/bottom, fill2 6
	for _, ck := range []struct {
		x, y int
		want uint8
	}{{0, 0, sh(3)}, {5, 0, sh(3)}, {9, 5, sh(7)}, {5, 9, sh(7)}, {5, 5, sh(6)}} {
		if got := at(ck.x, ck.y); got != ck.want {
			t.Errorf("inset (%d,%d) = %d, want %d", ck.x, ck.y, got, ck.want)
		}
	}
}
