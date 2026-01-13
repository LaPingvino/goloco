package render

import "github.com/hajimehoshi/ebiten/v2"

// Renderer is a small adapter that will be expanded to integrate with Ebiten.
type Renderer struct {
	// future fields for renderer state
	Screen *ebiten.Image
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) SetScreen(s *ebiten.Image) {
	r.Screen = s
}

func (r *Renderer) Clear() {
	// placeholder
}
