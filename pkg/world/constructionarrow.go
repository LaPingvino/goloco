package world

import "github.com/LaPingvino/goloco/pkg/scene"

// The blinking yellow arrow at the construction head (ConstructionTab
// updateConstructionArrow 0x0049FD66's tail; PaintTile.cpp paintConstructionArrow
// 0x00461EA7): once the head is placed, every 6th window update toggles
// GhostVisibilityFlags::constructArrow, re-aims the arrow at the head
// (cState.x, cState.y, constructionZ, constructionRotation), sets
// MapSelectionFlags::enableConstructionArrow from the toggle and invalidates
// the head's tile in full.

// constructionArrow is the arrow's state: World::getConstructionArrow plus
// enableConstructionArrow, and cState.constructionArrowFrameNum.
type constructionArrow struct {
	frame   uint8 // constructionArrowFrameNum
	visible bool  // constructArrow ghost flag = enableConstructionArrow
	x, y    int32 // world units (tile × 32)
	z       int32 // world z
	dir     uint8 // constructionRotation
}

// UpdateConstructionArrow is updateConstructionArrow, called on each update
// of the track window's Construction tab with the head placed (head true; the
// window's updateConstruction runs only while not hovering). With head false
// the arrow is switched off (resetMapSelectionFlag(enableConstructionArrow)
// when the tool / tab / window goes, Common.cpp) and the frame counter reset
// (constructionArrowFrameNum = 0, ConstructionTab when the head is set).
func (w *World) UpdateConstructionArrow(head bool) {
	a := &w.consArrow
	if !head {
		if a.visible {
			w.invalidateTile(a.x, a.y, 0, 1120, 3, 32)
		}
		*a = constructionArrow{}
		w.passArrow()
		return
	}
	a.frame--
	if a.frame == 0xFF {
		a.frame = 5
		a.visible = !a.visible
		h := w.ConstructionHeadState()
		if a.x != int32(h.X)*32 || a.y != int32(h.Y)*32 {
			w.invalidateTile(a.x, a.y, 0, 1120, 3, 32) // the old head's tile
		}
		a.x, a.y, a.z, a.dir = int32(h.X)*32, int32(h.Y)*32, int32(h.BaseZ)*4, h.Rotation
		w.invalidateTile(a.x, a.y, 0, 1120, 3, 32) // mapInvalidateTileFull
	}
	w.passArrow()
}

// passArrow hands the arrow to the renderer (a value: no allocation per update).
func (w *World) passArrow() {
	if w.scene == nil {
		return
	}
	a := w.consArrow
	w.scene.ConstructionArrow = scene.ConstructionArrow{Visible: a.visible, X: a.x, Y: a.y, Z: a.z, Direction: a.dir}
}
