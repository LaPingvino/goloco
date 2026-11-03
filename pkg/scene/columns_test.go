package scene

import (
	"bytes"
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// TestColumnsIndependent: Viewport::paint's columns each paint the same
// whichever other columns are painted — rendering the frame's columns in
// three groups (as a per-frame budget would) gives exactly the full frame, at
// every zoom.
func TestColumnsIndependent(t *testing.T) {
	r, m, v := titleView(t)
	for _, zoom := range []int{0, 1, 2, 3} {
		v := v
		v.Zoom = zoom
		full := gfx8.NewRenderTarget(v.W, v.H)
		r.RenderColumns(m, v, full, 0, v.W)
		want := full.Bits
		got := gfx8.NewRenderTarget(v.W, v.H)
		third := v.W / 3
		for _, g := range [][2]int{{2 * third, v.W}, {0, third}, {third, 2 * third}} {
			r.RenderColumns(m, v, got, g[0], g[1])
		}
		if !bytes.Equal(want, got.Bits) {
			diff := 0
			for i := range want {
				if want[i] != got.Bits[i] {
					diff++
				}
			}
			t.Errorf("zoom %d: %d pixels differ", zoom, diff)
		}
	}
}
