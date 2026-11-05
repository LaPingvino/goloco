package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// RailPlacementAtCursor is Vehicle.cpp's getTrackAtCursor (with
// getTrackProgressAtCursor) and getVehicleRailPlacementArgsFromCursor for a
// track element found under the cursor (InteractionAt) on the tile whose
// world x / y is (tx, ty): the piece's first tile (the element's
// sequence offset rotated back out), the sub-position whose screen point
// (toScreen, the viewport's gameToScreen + viewportToScreen) is nearest the
// cursor by Manhattan distance, then — when that sub-position faces the
// wrong way for the view (its yaw plus the view rotation, less 0x37, less
// 0x20 with pickupDirection, within 0 … 0x20) — the piece taken from its
// other end, reversed. ok is false for an element of no known piece.
func (m *Map) RailPlacementAtCursor(head uint16, tx, ty int32, te *scenario.TrackElement, viewRotation uint8, cx, cy int, toScreen func(x, y, z int32) (int, int)) (VehiclePlacement, bool) {
	id := int(te.TrackID)
	if id >= len(TrackPieces) || int(te.SequenceIndex) >= len(TrackPieces[id]) {
		return VehiclePlacement{}, false
	}
	td := TrackPieces[id][te.SequenceIndex]
	ox, oy := rotOffset(td.X, td.Y, te.Rotation) // Math::Vector::rotate
	fx, fy, fz := tx-int32(ox), ty-int32(oy), int32(te.BaseZ)*4-int32(td.Z)
	tad := uint16(te.Rotation) | uint16(te.TrackID)<<3
	moves := TrackSubPosition(tad)
	if len(moves) == 0 {
		return VehiclePlacement{}, false
	}
	best, progress := int(^uint(0)>>1), 0
	for i, mi := range moves { // getTrackProgressAtCursor
		sx, sy := toScreen(fx+int32(mi.X), fy+int32(mi.Y), fz+int32(mi.Z))
		d := absInt(sx-cx) + absInt(sy-cy)
		if d < best {
			best, progress = d, i
		}
	}
	p := VehiclePlacement{X: fx, Y: fy, Z: fz, TrackAndDirection: tad, TrackProgress: uint16(progress), Head: head}
	yaw := moves[progress].Yaw + viewRotation<<4 - 0x37
	if m.PickupDirection != 0 {
		yaw -= 0x20
	}
	if yaw&0x3F <= 0x20 {
		c := TrackCoordinates[tad&0x1FF] // getUnkTrack
		p.X, p.Y, p.Z = p.X+int32(c.DX), p.Y+int32(c.DY), p.Z+int32(c.DZ)
		if c.RotEnd < 12 {
			p.X -= int32(RotationOffset[c.RotEnd][0])
			p.Y -= int32(RotationOffset[c.RotEnd][1])
		}
		p.TrackProgress = uint16(max(len(moves)-progress, 0))
		if int(p.TrackProgress) >= len(moves) {
			p.TrackProgress = uint16(len(moves) - 1)
		}
		p.TrackAndDirection ^= 1 << 2
	}
	return p, true
}
