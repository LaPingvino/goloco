package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

func diffCount(a, b []uint8) int {
	n := 0
	for i := range a {
		if a[i] != b[i] {
			n++
		}
	}
	return n
}

// TestRenderRect: Viewport::paint(rect) over the whole screen, and over the
// rectangles a fully invalidated InvalidationGrid hands out, paints exactly
// the full frame, at every zoom.
func TestRenderRect(t *testing.T) {
	r, m, v := titleView(t)
	for _, zoom := range []int{0, 1, 2, 3} {
		v := v
		v.Zoom = zoom
		v.X, v.Y = v.X&^31, v.Y&^31 // a camera on the 32 grid: the screen's 64-wide blocks fall on whole columns
		full := gfx8.NewRenderTarget(v.W, v.H)
		r.RenderColumns(m, v, full, 0, v.W)
		one := gfx8.NewRenderTarget(v.W, v.H)
		r.RenderRect(m, v, one, 0, 0, v.W, v.H)
		if d := diffCount(full.Bits, one.Bits); d != 0 {
			t.Errorf("zoom %d, one rect: %d pixels differ", zoom, d)
		}
		var g gfx8.InvalidationGrid
		g.Reset(v.W, v.H, gfx8.InvalidationBlockWidth, gfx8.InvalidationBlockHeight)
		g.Invalidate(0, 0, v.W, v.H)
		grid := gfx8.NewRenderTarget(v.W, v.H)
		g.TraverseDirtyCells(func(l, tp, rt, b int) { r.RenderRect(m, v, grid, l, tp, rt, b) })
		if d := diffCount(full.Bits, grid.Bits); d != 0 {
			t.Errorf("zoom %d, the grid's rects: %d pixels differ", zoom, d)
		}
		if zoom == 0 { // information: a camera off the 32 grid splits the rects' columns differently (upstream does too)
			u := v
			u.X += 13
			f2, g2 := gfx8.NewRenderTarget(v.W, v.H), gfx8.NewRenderTarget(v.W, v.H)
			r.RenderColumns(m, u, f2, 0, v.W)
			g.Invalidate(0, 0, v.W, v.H)
			g.TraverseDirtyCells(func(l, tp, rt, b int) { r.RenderRect(m, u, g2, l, tp, rt, b) })
			t.Logf("zoom 0, camera 13 off the column grid: the grid's rects differ from one full paint in %d pixels (the column split moves)", diffCount(f2.Bits, g2.Bits))
		}
		if zoom == 0 { // information: every 64x8 block painted on its own (as small invalidations are)
			blocks := gfx8.NewRenderTarget(v.W, v.H)
			for y := 0; y < v.H; y += 8 {
				for x := 0; x < v.W; x += 64 {
					r.RenderRect(m, v, blocks, x, y, min(x+64, v.W), min(y+8, v.H))
				}
			}
			t.Logf("zoom 0, each 64x8 block on its own: %d of %d pixels differ from the full frame", diffCount(full.Bits, blocks.Bits), len(full.Bits))
		}
	}
}
