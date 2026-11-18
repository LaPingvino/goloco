package world

import (
	"slices"

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
	w.syncConstructSelection(on)
}

// syncConstructSelection hands the construction tool's selected tiles
// (sim.State.ConsSelection: enableConstruct and the free-form tiles) to the
// renderer, the old and new tiles redrawn when they change
// (mapInvalidateMapSelectionFreeFormTiles); on false switches them off.
func (w *World) syncConstructSelection(on bool) {
	var tiles [][2]int32
	if on && w.ConsSelection.On {
		tiles = w.ConsSelection.Tiles
	}
	unk03 := w.ConsSelection.Unk03
	if slices.Equal(tiles, w.consSelection) && unk03 == w.consSelectionUnk03 {
		return
	}
	w.consSelectionUnk03 = unk03
	for _, t := range w.consSelection {
		w.invalidateTile(t[0], t[1], 0, 1120, 3, 32)
	}
	for _, t := range tiles {
		w.invalidateTile(t[0], t[1], 0, 1120, 3, 32)
	}
	w.consSelection = append(w.consSelection[:0], tiles...)
	if w.scene == nil {
		return
	}
	if len(tiles) == 0 {
		w.scene.ConstructSelection = nil
		return
	}
	w.scene.ConstructSelectionUnk03 = unk03
	sel := slices.Clone(tiles)
	w.scene.ConstructSelection = func(tx, ty int) bool {
		return slices.Contains(sel, [2]int32{int32(tx) * 32, int32(ty) * 32})
	}
}
