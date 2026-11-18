package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// PaintPiecePreview is the Construction tab's preview (ConstructionTab.cpp
// draw → drawTrackCost 0x0049D106 / drawRoadCost 0x0049D325 → drawTrack /
// drawRoad 0x00478F1F) on rt, the construct widget's area (its left + 1, top
// + 1, width × height) holding the window's pixels: the piece (pieceID, a
// track's or — road — a road's, of object objID, in direction, with the
// window's mods less those whose extra doesn't fit the piece) painted with
// the map's own paint code at (256, 256) tiles, z 120 small, the
// construction arrow on its first tile. The target's top-left is the
// screen point of the piece's middle (its last tile's offset halved, +
// 0x2010, z + 0x1CC) less half the width and half the height + 16. Each
// piece tile goes on the map for its painting — the tile alone with the
// piece's element, its four neighbours a surface at 255 (slope 0xF) —
// and the map is put back before the next.
func (w *World) PaintPiecePreview(rt *gfx8.RenderTarget, road bool, pieceID, objID int, direction, mods uint8) {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || pieceID < 0 {
		return
	}
	om := w.renderer.ObjMgr
	var pieces []sim.PreviewTrack
	var compat uint16
	var modNames []string
	nMods := 4
	if road {
		o := om.GetRoadObjectByIndex(objID)
		if o == nil || pieceID >= len(sim.RoadPieces) {
			return
		}
		pieces, compat, modNames, nMods = sim.RoadPieces[pieceID], sim.RoadCompatibleFlags[pieceID], o.ModNames, 2
	} else {
		o := om.GetTrackObjectByIndex(objID)
		if o == nil || pieceID >= len(sim.TrackPieces) {
			return
		}
		pieces, compat, modNames = sim.TrackPieces[pieceID], sim.TrackCompatibleFlags[pieceID], o.ModNames
	}
	if len(pieces) == 0 || w.Map.Width < 258 || w.Map.Height < 258 {
		return
	}
	// Remove any none compatible track / road mods.
	for i := range nMods {
		if mods&(1<<i) == 0 {
			continue
		}
		if i >= len(modNames) {
			mods &^= 1 << i
			continue
		}
		if ex := om.Extra(modNames[i]); ex == nil || ex.Pieces&compat != compat {
			mods &^= 1 << i
		}
	}
	// draw: the piece's middle, centred in the widget.
	last := pieces[len(pieces)-1]
	lx, ly := int(last.X), int(last.Y)
	if last.Flags&(1<<6) != 0 { // PreviewTrackFlags::unused
		lx, ly = 0, 0
	}
	rx, ry := rotateXY(lx, ly, direction&3)
	mid := paint.Pos3{X: rx/2 + 0x2010, Y: ry/2 + 0x2010, Z: int(last.Z) + 0x1CC}
	rotation := w.Map.CurrentRotation & 3
	sx, sy := paint.GameToScreen(mid, rotation)
	width, height := rt.Width, rt.Height
	sx -= width >> 1
	sy -= height>>1 + 16

	if w.previewScene == nil {
		w.previewScene = scene.New(w.renderer.G1, om)
	}
	r := w.previewScene
	const px, py, pz = 256 * 32, 256 * 32, 120 * 4
	// The arrow: enableConstructionArrow at the preview's position
	// (TrackRoadPreviewFlags::displayConstructionArrow), the map's own
	// selection flags otherwise kept (no tool selection reaches tile 256).
	r.ConstructionArrow = scene.ConstructionArrow{Visible: true, X: px, Y: py, Z: pz, Direction: direction}
	s := r.BeginPreview(rotation, sx, sy, width, height)
	company, _ := w.ControllingCompany()
	dir := direction & 3
	for _, p := range pieces {
		ox, oy := rotateXY(int(p.X), int(p.Y), dir)
		x, y, z := px+ox, py+oy, pz+int(p.Z)
		baseZ := uint8(z / 4)
		clearZ := baseZ + uint8((int(p.ClearZ)+32)/4)
		quarter := sim.RotateQuarterTile(p.SubTileClearance, dir) & 0xF
		centre := sim.Tile{}
		if road {
			centre.Roads = []scenario.RoadElement{{BaseZ: baseZ, ClearZ: clearZ, Rotation: dir, SequenceIndex: p.Index,
				RoadObjectID: uint8(objID), RoadID: uint8(pieceID), Mods: mods & 3, Owner: uint8(company)}}
			centre.Order = []scenario.ElemRef{{Kind: scenario.KindRoad, Flags: quarter}}
		} else {
			centre.Tracks = []scenario.TrackElement{{BaseZ: baseZ, ClearZ: clearZ, Rotation: dir, SequenceIndex: p.Index,
				TrackObjectID: uint8(objID) & 0xF, TrackID: uint8(pieceID), Mods: mods & 0xF, Owner: uint8(company)}}
			centre.Order = []scenario.ElemRef{{Kind: scenario.KindTrack, Flags: quarter}}
		}
		side := sim.Tile{BaseZ: 255, ClearZ: 255, Slope: 0xF, Order: []scenario.ElemRef{{Kind: scenario.KindSurface}}}
		tx, ty := x/32, y/32
		at := [5][2]int{{tx, ty}, {tx, ty + 1}, {tx, ty - 1}, {tx + 1, ty}, {tx - 1, ty}} // centre, ± kOffsets[1] (east, west), ± kOffsets[3] (north, south)
		var backup [5]sim.Tile
		for i, c := range at {
			t := w.Map.TileAt(c[0], c[1])
			backup[i] = *t
			if i == 0 {
				*t = centre
			} else {
				*t = side
			}
		}
		r.PaintPreviewTile(s, &w.Map, tx, ty)
		for i, c := range at {
			*w.Map.TileAt(c[0], c[1]) = backup[i]
		}
	}
	r.EndPreview(s, rt)
}

// rotateXY is Math::Vector::rotate for a 2D offset.
func rotateXY(x, y int, rotation uint8) (int, int) {
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
