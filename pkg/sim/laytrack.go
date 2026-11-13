package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Track and road laid straight into the map, without the game commands (no
// checks, no owner's money, the default flags): the fixture tests and
// GOLOCO_OPEN=track|road build their lines with. The construction window
// builds through ConstructTrack / ConstructRoad.

// LayTrack lays the track piece trackID of track object objID from (x, y, z)
// (world units) in rotation rot (0–3) and returns where it ends
// (getTrackConnectionEnd). ok false when a tile is off the map or the height
// out of range (nothing is laid then).
func (m *Map) LayTrack(x, y, z int32, rot uint8, trackID int, objID uint8) (ConnectionEnd, bool) {
	if trackID < 0 || trackID >= len(TrackPieceTiles) || len(TrackPieceTiles[trackID]) == 0 {
		return ConnectionEnd{}, false
	}
	tiles, ok := m.layFootprint(TrackPieceTiles[trackID], x, y, z, rot)
	if !ok {
		return ConnectionEnd{}, false
	}
	for i, t := range tiles {
		m.AddTrack(t[0], t[1], scenario.TrackElement{
			TrackObjectID: objID,
			TrackID:       uint8(trackID),
			Rotation:      rot & 3,
			SequenceIndex: uint8(i),
			BaseZ:         uint8(t[2]),
			ClearZ:        uint8(t[2]) + 8,
			IsLast:        i == len(tiles)-1, // CreateTrack: setFlag6 on the piece's last tile
		})
	}
	return TrackConnectionEnd(x, y, z, uint16(trackID)<<3|uint16(rot&3))
}

// LayRoad is LayTrack for the road piece roadID of road object objID.
func (m *Map) LayRoad(x, y, z int32, rot uint8, roadID int, objID uint8) (ConnectionEnd, bool) {
	if roadID < 0 || roadID >= len(RoadPieceTiles) || len(RoadPieceTiles[roadID]) == 0 {
		return ConnectionEnd{}, false
	}
	tiles, ok := m.layFootprint(RoadPieceTiles[roadID], x, y, z, rot)
	if !ok {
		return ConnectionEnd{}, false
	}
	for i, t := range tiles {
		m.AddRoad(t[0], t[1], scenario.RoadElement{
			RoadObjectID:  objID,
			RoadID:        uint8(roadID),
			Rotation:      rot & 3,
			SequenceIndex: uint8(i),
			BaseZ:         uint8(t[2]),
			ClearZ:        uint8(t[2]) + 8,
			IsLast:        i == len(tiles)-1, // CreateRoad: setFlag6 on the piece's last tile
		})
	}
	return RoadConnectionEnd(x, y, z, uint16(roadID)<<3|uint16(rot&3))
}

// layFootprint is a piece's tiles from (x, y, z) in rotation rot: tile x,
// tile y and base height (small z) each.
func (m *Map) layFootprint(piece [][3]int16, x, y, z int32, rot uint8) ([][3]int, bool) {
	var tiles [][3]int
	for _, off := range piece {
		dx, dy := rotOffset(off[0], off[1], rot)
		tx, ty := int(x+int32(dx))/32, int(y+int32(dy))/32
		bz := (int(z) + int(off[2])) / 4
		if x+int32(dx) < 0 || y+int32(dy) < 0 || tx >= m.Width || ty >= m.Height || bz < 0 || bz > 255 {
			return nil, false
		}
		tiles = append(tiles, [3]int{tx, ty, bz})
	}
	return tiles, true
}
