package render

import (
	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer holds rendering state and sprite data
type Renderer struct {
	Screen *ebiten.Image
	Atlas  *Atlas
	G1     *assets.G1File

	// Cache for decoded G1 sprites
	spriteCache map[int]*ebiten.Image
}

func NewRenderer() *Renderer {
	r := &Renderer{
		spriteCache: make(map[int]*ebiten.Image),
	}
	// attempt to load atlas from default extracted assets directory; ignore errors
	if at, err := LoadAtlasFromDir("assets/extracted"); err == nil {
		r.Atlas = at
	}
	return r
}

func (r *Renderer) SetScreen(s *ebiten.Image) {
	r.Screen = s
}

// GetSprite returns an ebiten image for the given G1 sprite index
// Results are cached for performance
func (r *Renderer) GetSprite(index int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}

	// Check cache first
	if img, ok := r.spriteCache[index]; ok {
		return img
	}

	// Decode sprite
	rgba, err := r.G1.DecodeSprite(index)
	if err != nil {
		return nil
	}

	// Convert to ebiten image and cache
	img := ebiten.NewImageFromImage(rgba)
	r.spriteCache[index] = img
	return img
}

// GetSpriteInfo returns the dimensions and offsets for a sprite
func (r *Renderer) GetSpriteInfo(index int) (width, height, xOff, yOff int16, ok bool) {
	if r.G1 == nil || index < 0 || index >= len(r.G1.Elements) {
		return 0, 0, 0, 0, false
	}
	elem := &r.G1.Elements[index]
	return elem.Width, elem.Height, elem.XOffset, elem.YOffset, true
}

func (r *Renderer) Clear() {
	// placeholder
}
