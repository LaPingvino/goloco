package world

import (
	"image/color"
	"log"
	"math"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Track construction: an OpenLoco-style construction head that pieces connect
// to, instead of free placement under the cursor.
//
// The head is the world position, height and heading where the NEXT piece
// begins. Placing a piece appends its elements to every sequence tile and
// advances the head by the piece's end transform (kTrackCoordinates).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Construction.cpp
// and Map/Track/TrackData.cpp.

// ConstructionHead is the "next piece starts here" cursor.
type ConstructionHead struct {
	Active   bool
	X, Y     int   // tile coordinates of the piece origin
	BaseZ    uint8 // SmallZ height of the piece origin
	Rotation uint8 // heading 0-3 (diagonal headings 12-15 not yet buildable)
}

// placedPiece remembers one build action so it can be undone.
type placedPiece struct {
	trackID  int
	rotation uint8
	tiles    [][2]int // tile coords that received elements
	baseZ    []uint8  // element BaseZ per tile
}

// rotOffset rotates a rotation-0 world offset to the given rotation.
// Matches OpenLoco Math::Vector::rotate.
func rotOffset(x, y int16, rotation uint8) (int16, int16) {
	switch rotation & 3 {
	case 1:
		return y, -x
	case 2:
		return -x, -y
	case 3:
		return -y, x
	}
	return x, y
}

// SetConstructionHead places the head on a tile (click-to-start). The head
// takes the tile's surface height and keeps its current heading.
func (w *World) SetConstructionHead(tileX, tileY int) {
	t := w.getTile(tileX, tileY)
	if t == nil {
		return
	}
	w.consHead.Active = true
	w.consHead.X = tileX
	w.consHead.Y = tileY
	w.consHead.BaseZ = t.baseZ
}

// ConstructionHeadState returns the current head (value copy).
func (w *World) ConstructionHeadState() ConstructionHead { return w.consHead }

// RotateConstructionHead turns the head 90° clockwise.
func (w *World) RotateConstructionHead() {
	w.consHead.Rotation = (w.consHead.Rotation + 1) & 3
}

// ClearConstructionHead deactivates the head (leave build mode).
func (w *World) ClearConstructionHead() { w.consHead.Active = false }

// CanPlaceTrackAtHead reports whether trackID can start at the head's exact
// heading: the piece's begin rotation (kTrackCoordinates) must equal the head
// heading, which naturally restricts cardinal vs diagonal (12-15) pieces.
func (w *World) CanPlaceTrackAtHead(trackID int) bool {
	h := w.consHead
	if !h.Active || trackID < 0 || trackID<<3 >= len(kTrackCoordinates) {
		return false
	}
	if kTrackCoordinates[trackID<<3|int(h.Rotation&3)].RotBegin != h.Rotation {
		return false
	}
	_, _, ok := w.trackPieceTiles(trackID)
	return ok
}

// trackPieceTiles returns the sequence-tile placements for trackID at the
// head: tile coords and BaseZ per sequence tile. ok=false when the piece
// would leave the map or the data table has no entry.
func (w *World) trackPieceTiles(trackID int) (tiles [][2]int, baseZ []uint8, ok bool) {
	if trackID < 0 || trackID >= len(kTrackPieceTiles) || len(kTrackPieceTiles[trackID]) == 0 {
		return nil, nil, false
	}
	h := w.consHead
	for _, off := range kTrackPieceTiles[trackID] {
		dx, dy := rotOffset(off[0], off[1], h.Rotation)
		tx := h.X + int(dx)/32
		ty := h.Y + int(dy)/32
		if tx < 0 || ty < 0 || tx >= w.width || ty >= w.height {
			return nil, nil, false
		}
		z := int(h.BaseZ) + int(off[2])/4 // piece z is world units; BaseZ is SmallZ
		if z < 0 || z > 255 {
			return nil, nil, false
		}
		tiles = append(tiles, [2]int{tx, ty})
		baseZ = append(baseZ, uint8(z))
	}
	return tiles, baseZ, true
}

// PlaceTrackAtHead appends trackID at the construction head and advances the
// head along the piece's end transform. Returns false if the piece cannot be
// placed (inactive head, unsupported heading, off-map).
func (w *World) PlaceTrackAtHead(trackID int, trackObjID uint8) bool {
	h := w.consHead
	if !w.CanPlaceTrackAtHead(trackID) {
		return false
	}
	tiles, baseZ, ok := w.trackPieceTiles(trackID)
	if !ok {
		return false
	}
	_ = h

	placed := placedPiece{trackID: trackID, rotation: w.consHead.Rotation, tiles: tiles, baseZ: baseZ}
	for i, tc := range tiles {
		w.AddTrack(tc[0], tc[1], scenario.TrackElement{
			TrackObjectID: trackObjID,
			TrackID:       uint8(trackID),
			Rotation:      w.consHead.Rotation & 3,
			SequenceIndex: uint8(i),
			BaseZ:         baseZ[i],
			ClearZ:        baseZ[i] + 8,
		})
	}
	w.consStack = append(w.consStack, placed)
	log.Printf("[Cons] placed track id=%d objSlot=%d rot=%d tiles=%v z=%v", trackID, trackObjID, w.consHead.Rotation&3, tiles, baseZ)

	// Advance the head: trackAndDirection = trackID<<3 | rotation (forward).
	c := kTrackCoordinates[trackID<<3|int(w.consHead.Rotation&3)]
	w.consHead.X += int(c.DX) / 32
	w.consHead.Y += int(c.DY) / 32
	w.consHead.BaseZ = uint8(int(w.consHead.BaseZ) + int(c.DZ)/4)
	w.consHead.Rotation = c.RotEnd // may become diagonal (12-15); UI restricts pieces then
	return true
}

// UndoLastTrack removes the most recently placed piece and moves the head
// back to its origin.
func (w *World) UndoLastTrack() bool {
	if len(w.consStack) == 0 {
		return false
	}
	p := w.consStack[len(w.consStack)-1]
	w.consStack = w.consStack[:len(w.consStack)-1]
	for i, tc := range p.tiles {
		w.removeTrackElement(tc[0], tc[1], uint8(p.trackID), p.rotation, uint8(i), p.baseZ[i])
	}
	w.consHead.X = p.tiles[0][0]
	w.consHead.Y = p.tiles[0][1]
	w.consHead.BaseZ = p.baseZ[0]
	w.consHead.Rotation = p.rotation
	w.consHead.Active = true
	return true
}

// drawHeadMarker outlines the construction-head tile with a bright pulsing
// diamond so the head is visible even over busy terrain (OpenLoco shows a
// construction arrow here).
func (w *World) drawHeadMarker(screen *ebiten.Image, scale float64) {
	h := w.consHead
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
func (w *World) removeTrackElement(x, y int, trackID, rotation, seq, baseZ uint8) {
	t := w.getTile(x, y)
	if t == nil {
		return
	}
	for i, te := range t.tracks {
		if te.TrackID == trackID && te.Rotation == rotation &&
			te.SequenceIndex == seq && te.BaseZ == baseZ {
			t.tracks = append(t.tracks[:i], t.tracks[i+1:]...)
			return
		}
	}
}

// --- Road construction (mirrors the track functions above) ----------------
//
// Roads reuse the same construction head. The 80-entry kRoadCoordinates table
// and kRoadPieceTiles footprints are generated from the same upstream file.
// Roads never produce diagonal end rotations, so the head stays cardinal.

// CanPlaceRoadAtHead reports whether roadID can start at the head's exact
// heading: the piece's begin rotation (kRoadCoordinates) must equal the head
// heading.
func (w *World) CanPlaceRoadAtHead(roadID int) bool {
	h := w.consHead
	if !h.Active || roadID < 0 || roadID<<3 >= len(kRoadCoordinates) {
		return false
	}
	if kRoadCoordinates[roadID<<3|int(h.Rotation&3)].RotBegin != h.Rotation {
		return false
	}
	_, _, ok := w.roadPieceTiles(roadID)
	return ok
}

// roadPieceTiles returns the sequence-tile placements for roadID at the head:
// tile coords and BaseZ per sequence tile. ok=false when the piece would leave
// the map or the data table has no entry.
func (w *World) roadPieceTiles(roadID int) (tiles [][2]int, baseZ []uint8, ok bool) {
	if roadID < 0 || roadID >= len(kRoadPieceTiles) || len(kRoadPieceTiles[roadID]) == 0 {
		return nil, nil, false
	}
	h := w.consHead
	for _, off := range kRoadPieceTiles[roadID] {
		dx, dy := rotOffset(off[0], off[1], h.Rotation)
		tx := h.X + int(dx)/32
		ty := h.Y + int(dy)/32
		if tx < 0 || ty < 0 || tx >= w.width || ty >= w.height {
			return nil, nil, false
		}
		z := int(h.BaseZ) + int(off[2])/4 // piece z is world units; BaseZ is SmallZ
		if z < 0 || z > 255 {
			return nil, nil, false
		}
		tiles = append(tiles, [2]int{tx, ty})
		baseZ = append(baseZ, uint8(z))
	}
	return tiles, baseZ, true
}

// PlaceRoadAtHead appends roadID at the construction head and advances the head
// along the piece's end transform. Returns false if the piece cannot be placed.
func (w *World) PlaceRoadAtHead(roadID int, roadObjID uint8) bool {
	if !w.CanPlaceRoadAtHead(roadID) {
		return false
	}
	tiles, baseZ, ok := w.roadPieceTiles(roadID)
	if !ok {
		return false
	}

	placed := placedPiece{trackID: roadID, rotation: w.consHead.Rotation, tiles: tiles, baseZ: baseZ}
	for i, tc := range tiles {
		w.AddRoad(tc[0], tc[1], scenario.RoadElement{
			RoadObjectID:  roadObjID,
			RoadID:        uint8(roadID),
			Rotation:      w.consHead.Rotation & 3,
			SequenceIndex: uint8(i),
			BaseZ:         baseZ[i],
			ClearZ:        baseZ[i] + 8,
		})
	}
	w.consRoadStack = append(w.consRoadStack, placed)

	// Advance the head: roadAndDirection = roadID<<3 | rotation (forward).
	c := kRoadCoordinates[roadID<<3|int(w.consHead.Rotation&3)]
	w.consHead.X += int(c.DX) / 32
	w.consHead.Y += int(c.DY) / 32
	w.consHead.BaseZ = uint8(int(w.consHead.BaseZ) + int(c.DZ)/4)
	w.consHead.Rotation = c.RotEnd
	return true
}

// UndoLastRoad removes the most recently placed road piece and moves the head
// back to its origin.
func (w *World) UndoLastRoad() bool {
	if len(w.consRoadStack) == 0 {
		return false
	}
	p := w.consRoadStack[len(w.consRoadStack)-1]
	w.consRoadStack = w.consRoadStack[:len(w.consRoadStack)-1]
	for i, tc := range p.tiles {
		w.removeRoadElement(tc[0], tc[1], uint8(p.trackID), p.rotation, uint8(i), p.baseZ[i])
	}
	w.consHead.X = p.tiles[0][0]
	w.consHead.Y = p.tiles[0][1]
	w.consHead.BaseZ = p.baseZ[0]
	w.consHead.Rotation = p.rotation
	w.consHead.Active = true
	return true
}

// removeRoadElement deletes one exactly-matching road element from a tile.
func (w *World) removeRoadElement(x, y int, roadID, rotation, seq, baseZ uint8) {
	t := w.getTile(x, y)
	if t == nil {
		return
	}
	for i, re := range t.roads {
		if re.RoadID == roadID && re.Rotation == rotation &&
			re.SequenceIndex == seq && re.BaseZ == baseZ {
			t.roads = append(t.roads[:i], t.roads[i+1:]...)
			return
		}
	}
}

// DrawRoadConstructionGhost renders the selected road piece translucently at
// the head, covering every sequence tile. Uses the Style0 single-image-per-
// rotation offsets (roads 0-2 base merge sprites are valid direct images).
func (w *World) DrawRoadConstructionGhost(screen *ebiten.Image, roadID, objID int) {
	h := w.consHead
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
	tiles, baseZ, ok := w.roadPieceTiles(roadID)
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

// DrawConstructionGhost renders the selected piece translucently at the head,
// covering every sequence tile, plus a marker diamond on the head tile.
func (w *World) DrawConstructionGhost(screen *ebiten.Image, trackID, objID int) {
	h := w.consHead
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
		if tiles, baseZ, ok := w.trackPieceTiles(trackID); ok {
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
