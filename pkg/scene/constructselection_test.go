package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// changedNear renders m with and without draw's change and counts the
// pixels that differ, and those farther than 40 px from screen point (cx, cy).
func changedNear(r *Renderer, m *sim.Map, v View, before []byte, cx, cy int) (changed, far int, after []byte) {
	got := r.Render(m, v)
	for i := range before {
		if before[i] == got.Bits[i] {
			continue
		}
		changed++
		if dx, dy := i%got.Width-cx, i/got.Width-cy; dx*dx+dy*dy > 40*40 {
			far++
		}
	}
	return changed, far, append([]byte(nil), got.Bits...)
}

// enableConstruct's free-form tiles (PaintSurface.cpp 1655): the box on the
// tile's surface in unk25, in another colour (unk2B) with unk_03; on a
// flooded tile it stands at the water's surface, not the land's.
func TestConstructSelectionBox(t *testing.T) {
	r, m, v := saveView(t, "Data/title.dat", true)
	land, water := [2]int{-1, -1}, [2]int{-1, -1}
	for y := 8; y < m.Height-8 && (land[0] < 0 || water[0] < 0); y++ {
		for x := 8; x < m.Width-8; x++ {
			tl := m.TileAt(x, y)
			if tl == nil || tl.IsIndustrial || len(tl.Order) != 1 {
				continue
			}
			if land[0] < 0 && tl.Slope == 0 && tl.WaterLevel == 0 {
				land = [2]int{x, y}
			}
			if water[0] < 0 && tl.Slope == 0 && int(tl.WaterLevel)*16 > int(tl.BaseZ)*4 {
				water = [2]int{x, y}
			}
		}
	}
	if land[0] < 0 || water[0] < 0 {
		t.Skip("no plain land and flooded tile on the map")
	}
	at := func(p [2]int, z int) (View, int, int) {
		cx, cy := paint.GameToScreen(paint.Pos3{X: p[0]*32 + 16, Y: p[1]*32 + 16, Z: z}, 0)
		w := v
		w.X, w.Y = cx-w.W/2, cy-w.H/2
		return w, w.W / 2, w.H / 2
	}

	tl := m.TileAt(land[0], land[1])
	lv, cx, cy := at(land, int(tl.BaseZ)*4)
	plain := append([]byte(nil), r.Render(m, lv).Bits...)
	r.ConstructSelection = func(tx, ty int) bool { return tx == land[0] && ty == land[1] }
	n, far, unk25 := changedNear(r, m, lv, plain, cx, cy)
	if n == 0 || far != 0 {
		t.Fatalf("land tile %v: %d pixels changed, %d far from it", land, n, far)
	}
	r.ConstructSelectionUnk03 = true
	if n, _, _ := changedNear(r, m, lv, unk25, cx, cy); n == 0 {
		t.Errorf("unk_03: the box in the same colour as without")
	}
	r.ConstructSelectionUnk03 = false

	wt := m.TileAt(water[0], water[1])
	wz := int(wt.WaterLevel) * 16
	wv, wx, wy := at(water, wz)
	r.ConstructSelection = nil
	plain = append([]byte(nil), r.Render(m, wv).Bits...)
	r.ConstructSelection = func(tx, ty int) bool { return tx == water[0] && ty == water[1] }
	n, far, boxed := changedNear(r, m, wv, plain, wx, wy)
	if n == 0 || far != 0 {
		t.Fatalf("flooded tile %v: %d pixels changed, %d farther than 40 px from it", water, n, far)
	}
	// The box's middle row: at the water surface (wy), not 16·(water - land)
	// / 16 px lower at the land's height.
	sum := 0
	for i := range plain {
		if plain[i] != boxed[i] {
			sum += i / wv.W
		}
	}
	mid, landY := sum/n, wy+(wz-int(wt.BaseZ)*4)
	if d, dl := abs(mid-wy), abs(mid-landY); d >= dl {
		t.Errorf("flooded tile %v: the box's middle at y %d, nearer the land's height (%d) than the water's (%d)", water, mid, landY, wy)
	}
	r.ConstructSelection = nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
