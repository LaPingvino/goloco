package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/paint"
)

// The map selection (paintSurfaceFullTileSelection) draws its box on the
// selected tile's surface and nowhere else: the view with one snow-free tile
// selected differs from the plain one only around that tile. (Upstream
// attaches partial snow after the box, so on snow it can be hidden.)
func TestMapSelectionBox(t *testing.T) {
	r, m, v := saveView(t, "Data/title.dat", true)
	sx, sy := -1, -1
	for d := 0; d < 12 && sx < 0; d++ {
		for _, p := range [][2]int{{169 + d, 150}, {169, 150 + d}, {169 - d, 150}, {169, 150 - d}} {
			if tl := m.TileAt(p[0], p[1]); tl != nil && tl.Slope>>5 == 0 && !tl.IsIndustrial && tl.WaterLevel == 0 {
				sx, sy = p[0], p[1]
				break
			}
		}
	}
	if sx < 0 {
		t.Skip("no snow-free tile near the view's centre")
	}
	plain := append([]byte(nil), r.Render(m, v).Bits...)
	r.Selection = func(tx, ty int) bool { return tx == sx && ty == sy }
	sel := r.Render(m, v)
	tl := m.TileAt(sx, sy)
	cx, cy := paint.GameToScreen(paint.Pos3{X: sx*32 + 16, Y: sy*32 + 16, Z: int(tl.BaseZ) * 4}, 0)
	cx, cy = cx-v.X, cy-v.Y
	changed, far := 0, 0
	for i := range plain {
		if plain[i] == sel.Bits[i] {
			continue
		}
		changed++
		if dx, dy := i%sel.Width-cx, i/sel.Width-cy; dx*dx+dy*dy > 40*40 {
			far++
		}
	}
	if changed == 0 {
		t.Fatalf("the selection box on tile (%d, %d) drew nothing", sx, sy)
	}
	if far != 0 {
		t.Errorf("%d of %d changed pixels are far from tile (%d, %d)", far, changed, sx, sy)
	}
	t.Logf("tile (%d, %d): %d pixels of the selection box", sx, sy, changed)
}
