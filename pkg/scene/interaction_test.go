package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// TestInteractionAt (getMapCoordinatesFromPos): over a grid of screen pixels
// on the title view, every pixel the frame drew something at has an item
// under it (anything accepted); with only track accepted, the items found
// are track elements painted from a tile whose screen position is near the
// pixel; at zoom 1 too.
func TestInteractionAt(t *testing.T) {
	r, m, v := titleView(t)
	for _, zoom := range []int{0, 1} {
		v := v
		v.Zoom = zoom
		full := gfx8.NewRenderTarget(v.W, v.H)
		r.RenderColumns(m, v, full, 0, v.W)
		any := func(paint.InteractionItem) bool { return true }
		track := func(it paint.InteractionItem) bool { return it == paint.ItemTrack }
		drawn, found, tracks := 0, 0, 0
		for sy := 4; sy < v.H; sy += 23 {
			for sx := 4; sx < v.W; sx += 23 {
				if full.Bits[sy*v.W+sx] == brown2 {
					continue
				}
				drawn++
				if r.InteractionAt(m, v, sx, sy, any) != nil {
					found++
				}
				ps := r.InteractionAt(m, v, sx, sy, track)
				if ps == nil {
					continue
				}
				tracks++
				te, ok := ps.Item.(*scenario.TrackElement)
				if !ok {
					t.Fatalf("zoom %d (%d,%d): a track hit with item %T", zoom, sx, sy, ps.Item)
				}
				tx, ty := paint.GameToScreen(paint.Pos3{X: ps.MapX + 16, Y: ps.MapY + 16, Z: int(te.BaseZ) * 4}, 0)
				dx, dy := (tx-(v.X+sx<<zoom))>>zoom, (ty-(v.Y+sy<<zoom))>>zoom
				if dx*dx+dy*dy > 80*80 {
					t.Errorf("zoom %d (%d,%d): track tile %d,%d is %d,%d px away", zoom, sx, sy, ps.MapX/32, ps.MapY/32, dx, dy)
				}
			}
		}
		t.Logf("zoom %d: %d drawn pixels sampled, %d with an item, %d on track", zoom, drawn, found, tracks)
		if drawn == 0 || found < drawn*95/100 || tracks == 0 {
			t.Errorf("zoom %d: %d drawn, %d with an item, %d on track", zoom, drawn, found, tracks)
		}
	}
}
