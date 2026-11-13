package world

import "github.com/hajimehoshi/ebiten/v2"

// DrawPiecePreview draws the selected construction piece's sprite(s) centred at
// (screenCX, screenCY) using the given rotation and scale, into an arbitrary
// destination image (used by the construction window's preview panel and tabs).
// It is drawn from screen-space coordinates only (no camera/world transform).
// Returns true if any sprite was drawn.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/ConstructionTab.cpp
// draw() — the construct Wt3 widget renders the piece into a clipped mini-viewport.
func (w *World) DrawPiecePreview(dst *ebiten.Image, screenCX, screenCY, scale float64, isRoad bool, pieceID, objID, rotation int) bool {
	if w.renderer == nil || w.renderer.ObjMgr == nil || pieceID < 0 {
		return false
	}
	rot := rotation & 3
	drawSprite := func(spriteID int) {
		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			return
		}
		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		// Track/road tile sprites are anchored at the tile's top-left in isometric
		// space (xOff≈-32 centres a 64px tile). Nudge down so the deck sits mid-panel.
		op.GeoM.Translate(screenCX+float64(xOff)*scale, screenCY+(float64(yOff)+8)*scale)
		dst.DrawImage(img, op)
	}

	if isRoad {
		if pieceID >= len(kRoadPartsStyle0) || len(kRoadPartsStyle0[pieceID]) == 0 {
			return false
		}
		roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(objID)
		if roadObj == nil || roadObj.Image == 0 {
			return false
		}
		drawn := false
		for _, p := range kRoadPartsStyle0[pieceID] {
			off := p.img[rot]
			if off == 0 {
				continue
			}
			drawSprite(int(roadObj.Image) + int(off))
			drawn = true
		}
		return drawn
	}

	if pieceID >= len(kTrackParts) || len(kTrackParts[pieceID]) == 0 {
		return false
	}
	trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(objID)
	if trackObj == nil || trackObj.Image == 0 {
		return false
	}
	drawn := false
	for _, p := range kTrackParts[pieceID] {
		for layer := 0; layer < 3; layer++ {
			off := p.img[rot][layer]
			if off == 0 {
				continue
			}
			drawSprite(int(trackObj.Image) + int(off))
			drawn = true
			if p.nonMergeable {
				break
			}
		}
	}
	return drawn
}
