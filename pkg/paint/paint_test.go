package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

func dot(idx uint8) Image {
	return Image{Sprite: &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{idx}, HasTransparency: true}, Op: gfx8.OpTransparent}
}

func TestGameToScreen(t *testing.T) {
	// Tile.cpp: (y - x, ((y + x) >> 1) - z) at rotation 0.
	if x, y := GameToScreen(Pos3{32, 0, 0}, 0); x != -32 || y != 16 {
		t.Errorf("(32,0,0) → (%d,%d)", x, y)
	}
	if x, y := GameToScreen(Pos3{0, 32, 8}, 0); x != 32 || y != 8 {
		t.Errorf("(0,32,8) → (%d,%d)", x, y)
	}
}

// Tiles further forward (larger x+y at rotation 0) draw later.
func TestQuadrantOrderBackToFront(t *testing.T) {
	s := NewSession(0, -1000, -1000, 2000, 2000)
	var want []*Struct
	for _, p := range [][2]int{{64, 64}, {0, 0}, {32, 0}} {
		s.SetSpritePosition(p[0], p[1])
		ps := s.AddAsParent(dot(1), Pos3{}, Pos3{}, Pos3{32, 32, 1})
		want = append(want, ps)
	}
	s.Arrange()
	got := s.Order()
	if len(got) != 3 || got[0] != want[1] || got[1] != want[2] || got[2] != want[0] {
		t.Fatalf("order wrong: tiles (0,0),(32,0),(64,64) expected back to front")
	}
}

// On one tile: an object on top of the ground draws after the ground, even
// when it was added first.
func TestObjectOnGroundDrawsAfterGround(t *testing.T) {
	s := NewSession(0, -1000, -1000, 2000, 2000)
	s.SetSpritePosition(0, 0)
	obj := s.AddAsParent(dot(2), Pos3{16, 16, 0}, Pos3{8, 8, 1}, Pos3{16, 16, 20})
	ground := s.AddAsParent(dot(1), Pos3{}, Pos3{0, 0, 0}, Pos3{32, 32, -1})
	s.Arrange()
	got := s.Order()
	if len(got) != 2 || got[0] != ground || got[1] != obj {
		t.Fatalf("ground must draw before the object on it")
	}
}

func TestDrawChildrenAndAttached(t *testing.T) {
	s := NewSession(0, 0, 0, 4, 4)
	s.SetSpritePosition(0, 0)
	s.AddAsParent(dot(1), Pos3{}, Pos3{}, Pos3{32, 32, 1})
	s.AttachToPrevious(dot(3), 1, 0)
	s.AddAsChild(dot(2), Pos3{}, Pos3{}, Pos3{32, 32, 1})
	s.Arrange()
	rt := gfx8.NewRenderTarget(4, 4)
	s.Draw(rt)
	if rt.Bits[0] != 2 || rt.Bits[1] != 3 {
		t.Fatalf("child over parent at (0,0) and attachment at (1,0): %v", rt.Bits[:4])
	}
}
