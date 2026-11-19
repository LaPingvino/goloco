package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// A window's viewport holds the points of its rectangle only
// (containsWindowPos: left and top in, right and bottom out), the first of a
// window's viewports that holds the point is the one read, and a world
// position projects into the viewport's place on the screen
// (viewportToWindow + the window's position) at the viewport's zoom.
func TestWindowViewport(t *testing.T) {
	vps := []Viewport{
		{X: 100, Y: 50, View: scene.View{X: 1000, Y: 400, W: 20, H: 10, Zoom: 2}},
		{X: 110, Y: 55, View: scene.View{W: 50, H: 50}},
	}
	for _, c := range []struct{ x, y, want int }{
		{100, 50, 0}, {119, 59, 0}, {120, 59, 1}, {119, 60, 1}, {99, 50, -1}, {200, 200, -1},
	} {
		if got := containingViewport(vps, c.x, c.y); got != c.want {
			t.Errorf("containingViewport(%d, %d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}
	w := &World{}
	w.Map = sim.Map{Width: 64, Height: 64}
	gx, gy := w.Map.GameToScreen(320, 320, 16)
	vp := Viewport{X: 100, Y: 50, View: scene.View{X: int(gx) - 40, Y: int(gy) - 80, W: 40, H: 40, Zoom: 2}}
	if x, y := vp.ToScreen(w, 320, 320, 16); x != 110 || y != 70 {
		t.Errorf("ToScreen = %d, %d, want 110, 70", x, y)
	}
}

// The hovered station (getStationArguments(id)) lasts until processMouseOver
// lets it go.
func TestHoveredStationReset(t *testing.T) {
	w := &World{hoveredStation: 3, hoveringOverStation: true}
	w.ResetHoveredStation()
	if w.hoveringOverStation {
		t.Error("the hovered station stayed after ResetHoveredStation")
	}
}
