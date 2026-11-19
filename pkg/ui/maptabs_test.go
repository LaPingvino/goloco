package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// setHoverItemTab's search: rows 10 px apart from the key's top; a point
// above the key hovers the first row, one past a fixed-length key's rows or
// off its 100 px width none; an industry key with empty slots hovers the
// object's bit, not the row's; the routes key counts its own rows.
func TestMapHoverItem(t *testing.T) {
	overall := MapOverallKey()
	for _, c := range []struct {
		tab, x, y int
		rows      []MapKeyRow
		want      uint32
	}{
		{0, 5, 0, overall, 1 << 0},
		{0, 5, 9, overall, 1 << 0},
		{0, 5, 10, overall, 1 << 1},
		{0, 5, 59, overall, 1 << 5},
		{0, 5, 60, overall, 0},
		{0, 5, -30, overall, 1 << 0},
		{0, -1, 5, overall, 0},
		{0, 101, 5, overall, 0},
		{0, 100, 5, overall, 1 << 0},
		{2, 5, 15, []MapKeyRow{{Bit: 3}, {Bit: 7}}, 1 << 7},
		{2, 5, 25, []MapKeyRow{{Bit: 3}, {Bit: 7}}, 0},
		{2, 5, 160, []MapKeyRow{{Bit: 3}, {Bit: 7}}, 0},
		{3, 5, 25, []MapKeyRow{{Bit: 0}, {Bit: 1}, {Bit: 2}}, 1 << 2},
		{3, 5, 30, []MapKeyRow{{Bit: 0}, {Bit: 1}, {Bit: 2}}, 0},
		{4, 5, 5, []MapKeyRow{{Bit: 2}}, 1 << 2},
	} {
		if got := MapHoverItem(c.tab, c.x, c.y, c.rows); got != c.want {
			t.Errorf("tab %d at %d,%d: %#x, want %#x", c.tab, c.x, c.y, got, c.want)
		}
	}
}

// drawLine on the scroll's target: the endpoints drawn, the target's
// offset taken off, a line off the target (by upstream's unordered
// rectangle test) not drawn.
func TestMapDrawLine(t *testing.T) {
	rt := gfx8.NewRenderTarget(40, 40)
	tg := mapTarget{rt: rt, OrgX: 5, OrgY: 5, W: 20, H: 20, X: 100, Y: 200}
	at := func(mx, my int) uint8 { return rt.Bits[(my-200+5)*40+(mx-100+5)] }
	tg.drawMapLine(102, 203, 110, 203, 7) // horizontal
	for x := 102; x < 110; x++ {
		if at(x, 203) != 7 {
			t.Errorf("horizontal: %d not drawn", x)
		}
	}
	tg.drawMapLine(104, 201, 104, 210, 9) // vertical: steep, a pixel a step
	for y := 201; y < 210; y++ {
		if at(104, y) != 9 {
			t.Errorf("vertical: %d not drawn", y)
		}
	}
	for i := range rt.Bits {
		rt.Bits[i] = 0
	}
	tg.drawMapLine(50, 210, 90, 215, 3) // all left of the target
	tg.drawMapLine(130, 210, 125, 215, 3)
	for _, b := range rt.Bits {
		if b != 0 {
			t.Fatal("a line off the target was drawn")
		}
	}
}
