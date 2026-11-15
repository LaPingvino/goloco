package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/sim"
)

// SoundAt: calculateVolumeFromViewport's −1024 per zoom level, ×256 below the
// surface, and no sound off the view.
func TestSoundAt(t *testing.T) {
	const sz = 16
	w := &World{State: sim.State{Map: sim.Map{Width: sz, Height: sz, Tiles: make([][]tile, sz)}}, viewW: 640, viewH: 480}
	for y := range w.Map.Tiles {
		w.Map.Tiles[y] = make([]tile, sz)
		for x := range w.Map.Tiles[y] {
			w.Map.Tiles[y][x].BaseZ = 4 // 16 world units
		}
	}
	x, y := int32(8*32+16), int32(8*32+16)
	vx, vy := w.Map.GameToScreen(x, y, 16)
	w.camX, w.camY = float64(vx-320), float64(vy-240)
	if vol, sx, ok := w.SoundAt(x, y, 16); !ok || vol != 0 || sx != 320 {
		t.Errorf("zoom 0 on the surface: %d %d %v, want 0 320 true", vol, sx, ok)
	}
	if vol, _, _ := w.SoundAt(x, y, 0); vol != ((-1)<<8)+1 {
		t.Errorf("below the surface: %d, want %d", vol, ((-1)<<8)+1)
	}
	w.zoom = 1
	if vol, sx, ok := w.SoundAt(x, y, 16); !ok || vol != -1024 || sx != 160 {
		t.Errorf("zoom 1: %d %d %v, want -1024 160 true", vol, sx, ok)
	}
	w.zoom = 0
	if _, _, ok := w.SoundAt(0, 15*32, 16); ok {
		t.Error("a point off the view is heard")
	}
}
