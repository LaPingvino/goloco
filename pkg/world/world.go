package world

import "github.com/LaPingvino/goloco/pkg/render"

// World is a minimal stub for now, will hold map, entities and simulation state.
type World struct {
	renderer *render.Renderer
}

func NewWorld(r *render.Renderer) *World {
	return &World{renderer: r}
}

func (w *World) Update() {
	// placeholder for world tick
}

func (w *World) Draw() {
	// placeholder for drawing world using w.renderer
}
