package world

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (w *World) drawHeadMarker(screen *ebiten.Image, scale float64) {
	h := w.ConsHead
	if !h.Active {
		return
	}
	vpX, vpY := w.tileToScreen(h.X, h.Y)
	vpY -= float64(h.BaseZ) * 4.0
	cx := float32(math.Round((vpX - w.camX) * scale))
	cy := float32(math.Round((vpY - w.camY) * scale))
	s := float32(scale)
	// Tile diamond corners relative to the anchor (top corner).
	pts := [4][2]float32{{0, 0}, {32 * s, 16 * s}, {0, 32 * s}, {-32 * s, 16 * s}}
	pulse := uint8(180 + 75*math.Sin(float64(w.frameCounter)*0.2))
	col := color.RGBA{255, 255, pulse, 255}
	for i := range pts {
		a, b := pts[i], pts[(i+1)%4]
		vector.StrokeLine(screen, cx+a[0], cy+a[1], cx+b[0], cy+b[1], 2, col, false)
	}
}

// removeTrackElement deletes one exactly-matching track element from a tile.

func (w *World) DrawRoadConstructionGhost(screen *ebiten.Image, roadID, objID int) {
	h := w.ConsHead
	if !h.Active || w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	if !w.CanPlaceRoadAtHead(roadID) || roadID < 0 || roadID >= len(kRoadPartsStyle0) {
		return
	}
	roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(objID)
	if roadObj == nil || roadObj.Image == 0 {
		return
	}
	pieces := kRoadPartsStyle0[roadID]
	tiles, baseZ, ok := w.RoadFootprint(roadID)
	if !ok {
		return
	}
	scale := 1.0 / float64(int(1)<<w.zoom)
	for seq, tc := range tiles {
		if seq >= len(pieces) {
			break
		}
		offset := pieces[seq].img[h.Rotation&3]
		if offset == 0 {
			continue
		}
		spriteID := int(roadObj.Image) + int(offset)
		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}
		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}
		vpX, vpY := w.tileToScreen(tc[0], tc[1])
		vpY -= float64(baseZ[seq]) * 4.0
		drawX := math.Round((vpX - w.camX) * scale)
		drawY := math.Round((vpY - w.camY) * scale)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX+float64(xOff)*scale, drawY+float64(yOff)*scale)
		op.ColorScale.ScaleAlpha(0.6)
		screen.DrawImage(img, op)
	}
}

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

// DrawConstructionGhost renders the selected piece translucently at the head,
// covering every sequence tile, plus a marker diamond on the head tile.
func (w *World) DrawConstructionGhost(screen *ebiten.Image, trackID, objID int) {
	h := w.ConsHead
	if !h.Active || w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(objID)
	scale := 1.0 / float64(int(1)<<w.zoom)

	drawAt := func(tileX, tileY int, z uint8, spriteID int) {
		vpX, vpY := w.tileToScreen(tileX, tileY)
		vpY -= float64(z) * 4.0
		drawX := math.Round((vpX - w.camX) * scale)
		drawY := math.Round((vpY - w.camY) * scale)
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
		op.GeoM.Translate(drawX+float64(xOff)*scale, drawY+float64(yOff)*scale)
		op.ColorScale.ScaleAlpha(0.6)
		screen.DrawImage(img, op)
	}

	w.drawHeadMarker(screen, scale)

	if trackObj != nil && trackObj.Image != 0 && w.CanPlaceTrackAtHead(trackID) &&
		trackID < len(kTrackParts) && kTrackParts[trackID] != nil {
		if tiles, baseZ, ok := w.TrackFootprint(trackID); ok {
			parts := kTrackParts[trackID]
			for seq, tc := range tiles {
				if seq >= len(parts) {
					break
				}
				for layer := 0; layer < 3; layer++ {
					offset := parts[seq].img[h.Rotation&3][layer]
					if offset == 0 {
						continue
					}
					drawAt(tc[0], tc[1], baseZ[seq], int(trackObj.Image)+int(offset))
				}
			}
		}
	}
}
