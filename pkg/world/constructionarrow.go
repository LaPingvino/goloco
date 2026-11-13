package world

import (
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The yellow construction arrow (PaintTile.cpp paintConstructionArrow
// 0x00461EA7): sim.State.ConsArrow is World::setConstructionArrow's arrow
// with MapSelectionFlags::enableConstructionArrow — steady under the cursor
// while the construction tool hovers (onToolUpdateSingle), blinking where
// the next piece starts while building (updateConstructionArrow). Here it
// goes to the renderer, and its tile is redrawn when it changes
// (mapInvalidateTileFull).

// SyncConstructionArrow hands the arrow to the renderer; on (false: the
// construction tool isn't up — another tab, no window) switches it off.
func (w *World) SyncConstructionArrow(on bool) {
	a := w.ConsArrow
	if !on {
		a = sim.ConstructionArrow{}
	}
	if a != w.consArrow {
		if w.consArrow.On {
			w.invalidateTile(w.consArrow.X, w.consArrow.Y, 0, 1120, 3, 32)
		}
		if a.On {
			w.invalidateTile(a.X, a.Y, 0, 1120, 3, 32)
		}
		w.consArrow = a
	}
	if w.scene != nil {
		w.scene.ConstructionArrow = scene.ConstructionArrow{Visible: a.On, X: a.X, Y: a.Y, Z: a.Z, Direction: a.Direction}
	}
}
