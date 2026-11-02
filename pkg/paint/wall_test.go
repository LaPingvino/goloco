package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// A glass wall gets its glass image (kImageOffsetsGlass) as a blended child;
// a plain wall only its own image.
func TestWallGlass(t *testing.T) {
	var asked []int
	sprite := func(id int) *gfx8.Sprite {
		asked = append(asked, id)
		return &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{1}}
	}
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	pal := make([]uint8, 256)
	before := len(s.structs)
	s.PaintWall(WallEl{Rotation: 1}, WallObj{Sprite: 100, Height: 4, HasGlass: true, GlassPal: pal}, nil, sprite)
	if n := len(s.structs) - before; n != 2 {
		t.Fatalf("glass wall: %d structs, want 2", n)
	}
	if want := 100 + wallGlassOffsets[1][2]; asked[len(asked)-1] != want {
		t.Errorf("glass image %d, want %d", asked[len(asked)-1], want)
	}
	before = len(s.structs)
	s.PaintWall(WallEl{Rotation: 1}, WallObj{Sprite: 100, Height: 4}, nil, sprite)
	if n := len(s.structs) - before; n != 1 {
		t.Errorf("plain wall: %d structs, want 1", n)
	}
}
