package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The four view rotations (PaintSession::generate's generateParameters<0–3>,
// paintTileElementsSetup's kUnkOffsets, getRotatedSlope, kNeighbourOffsets):
// the same place seen from each side paints the whole frame — next to no
// pixel in the viewport's clear colour (a few sprites use it too) — and the
// surface under a pixel (getMapCoordinatesFromPos) is a tile whose own screen
// position is at that pixel (a cliff face, also a surface item, hangs below
// its tile). Pixel equality with OpenLoco at each rotation is checked by
// hand with scripts/olref (setview + the rotate key in both games).
func TestViewRotations(t *testing.T) {
	r, m, _ := titleView(t)
	const cx, cy = 169*32 + 16, 150*32 + 16
	land, _ := m.GetHeight(cx, cy)
	surface := func(it paint.InteractionItem) bool { return it == paint.ItemSurface }
	for rot := uint8(0); rot < 4; rot++ {
		m.CurrentRotation = rot
		m.UpdateSpatialIndex()
		vx, vy := m.GameToScreen(cx, cy, int32(land))
		v := View{X: int(vx) - 400, Y: int(vy) - 300, W: 800, H: 600}
		full := gfx8.NewRenderTarget(v.W, v.H)
		r.RenderColumns(m, v, full, 0, v.W)
		empty := 0
		for _, p := range full.Bits {
			if p == brown2 {
				empty++
			}
		}
		if empty > len(full.Bits)/30 {
			t.Errorf("rotation %d: %d of %d pixels unpainted", rot, empty, len(full.Bits))
		}
		hits := 0
		for sy := 20; sy < v.H; sy += 37 {
			for sx := 20; sx < v.W; sx += 37 {
				ps := r.InteractionAt(m, v, sx, sy, surface)
				if ps == nil {
					continue
				}
				hits++
				tl := m.TileAt(ps.MapX/32, ps.MapY/32)
				if tl == nil {
					t.Fatalf("rotation %d (%d,%d): a surface hit off the map: %d,%d", rot, sx, sy, ps.MapX, ps.MapY)
				}
				// the tile's centre on screen: within a tile's sprite (64 × 32, its slope, its cliff faces) of the pixel
				tx, ty := sim.GameToScreen(int32(ps.MapX)+16, int32(ps.MapY)+16, int32(tl.BaseZ)*4, rot)
				dx, dy := int(tx)-(v.X+sx), int(ty)-(v.Y+sy)
				if dx < -36 || dx > 36 || dy < -80 || dy > 48 {
					t.Errorf("rotation %d (%d,%d): surface tile %d,%d is %d,%d px away", rot, sx, sy, ps.MapX/32, ps.MapY/32, dx, dy)
				}
			}
		}
		if hits < 200 {
			t.Errorf("rotation %d: only %d surface hits", rot, hits)
		}
	}
}

// getRotatedSlope: the corner bits turn with the view, the double-height bit
// stays, the snow coverage goloco keeps above them goes.
func TestRotatedSlope(t *testing.T) {
	for _, c := range []struct{ slope, rot, want uint8 }{
		{0b0001, 1, 0b0010}, {0b1000, 1, 0b0001}, {0b0011, 2, 0b1100}, {0b1_0111, 3, 0b1_1011}, {0b101_0_0001, 0, 0b0001},
	} {
		if got := rotatedSlope(c.slope, c.rot); got != c.want {
			t.Errorf("rotatedSlope(%#b, %d) = %#b, want %#b", c.slope, c.rot, got, c.want)
		}
	}
}
