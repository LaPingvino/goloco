package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// paintLevelCrossing: four gate images from image + (rot&1)·4 + frame·8;
// frame 15 animates: ((ticks >> speed) & (closing − 1)) + closed + 1.
func TestPaintLevelCrossing(t *testing.T) {
	var asked []int
	sprite := func(id int) *gfx8.Sprite {
		asked = append(asked, id)
		return &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{1}}
	}
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	el := RoadEl{Rotation: 1, BaseZ: 4, Crossing: &LevelCrossing{Image: 500, AnimationSpeed: 2, ClosingFrames: 4, ClosedFrames: 3, Frame: 0}}
	s.paintLevelCrossing(el, 1, sprite)
	if want := []int{504, 505, 506, 507}; len(asked) != 4 || asked[0] != want[0] || asked[3] != want[3] {
		t.Fatalf("open crossing images %v, want %v", asked, want)
	}
	// ticks 21 >> 2 = 5, & 3 = 1, + 3 + 1 = frame 5: 500 + 4 + 40
	asked = nil
	el.Crossing.Frame, el.Crossing.Ticks = 15, 21
	s.paintLevelCrossing(el, 1, sprite)
	if len(asked) != 4 || asked[0] != 544 {
		t.Fatalf("animating crossing images %v, want from 544", asked)
	}
	// Beyond zoom 1 PaintRoad draws no crossing.
	asked = nil
	s.Zoom = 2
	s.PaintRoad(RoadEl{RoadID: 0xFF, Crossing: el.Crossing}, 0, 0, sprite)
	if len(asked) != 0 {
		t.Errorf("zoom 2: %v drawn", asked)
	}
}
