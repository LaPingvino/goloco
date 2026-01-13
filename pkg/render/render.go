package render

import "github.com/hajimehoshi/ebiten/v2"

// Renderer is a small adapter that will be expanded to integrate with Ebiten.
type Renderer struct {
	// future fields for renderer state
	Screen *ebiten.Image
	Atlas  *Atlas
}

func NewRenderer() *Renderer {
	r := &Renderer{}
	// attempt to load atlas from default extracted assets directory; ignore errors
	if at, err := LoadAtlasFromDir("assets/extracted"); err == nil {
		r.Atlas = at
	}
	return r
}

func (r *Renderer) SetScreen(s *ebiten.Image) {
	r.Screen = s
}

func (r *Renderer) Clear() {
	// placeholder
}
