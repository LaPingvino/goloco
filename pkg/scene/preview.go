package scene

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The construction window's piece preview (ConstructionTab.cpp drawTrack /
// drawRoad): a paint session of its own on the window's render target, the
// map's tiles painted one at a time while pkg/world has put the preview's
// temporary elements on them.

// BeginPreview is drawTrack's Paint::PaintSession: zoom 0, the view's
// rotation, default session options (no view flags), the w × h target's
// top-left at viewport point (x, y).
func (r *Renderer) BeginPreview(rotation uint8, x, y, w, h int) *paint.Session {
	if r.spriteFor == nil {
		r.spriteFor = r.Sprite
	}
	if r.colourMapFn == nil {
		r.colourMapFn = r.colourMap
	}
	s := paint.NewSession(rotation&3, x, y, w, h)
	s.Zoom, s.SpriteFor, s.ColourMap = 0, r.spriteFor, r.colourMapFn
	return s
}

// PaintPreviewTile is Paint::paintTileElements(session, pos) for the tile
// (tx, ty) of m as it stands now (with the arrow, ConstructionArrow).
func (r *Renderer) PaintPreviewTile(s *paint.Session, m *sim.Map, tx, ty int) {
	r.paintTile(s, m, tx, ty, false, 0)
}

// EndPreview is session.arrangeStructs and drawStructs on rt.
func (r *Renderer) EndPreview(s *paint.Session, rt *gfx8.RenderTarget) {
	s.Arrange()
	s.Draw(rt)
}
