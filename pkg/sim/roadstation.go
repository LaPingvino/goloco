package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// CreateRoadStation.cpp's helpers (item ST, road stations): the station a
// road piece joins, its validation, the road element a placement names, and
// the AI's junction test. The command itself follows.

// findNearbyStationOnRoad is CreateRoadStation.cpp's (0x004900B8): the
// company's own station the road continues into at the piece's end, or at its
// start (the piece reversed), else FindNearbyStation. Attached: the first two.
func (m *Map) findNearbyStationOnRoad(x, y, z int32, tad uint16, roadObjectID, company uint8) (int, bool) {
	own := func(id int) bool {
		return id >= 0 && id < len(m.StationNames) && m.StationNames[id].Owner == company
	}
	if e, ok := RoadConnectionEnd(x, y, z, tad); ok {
		if rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, company, roadObjectID, 0, 0); own(rc.StationID) {
			return rc.StationID, true
		}
	}
	if int(tad) < len(RoadCoordinates) {
		c := RoadCoordinates[tad] // getUnkRoad
		tx, ty, tz := x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		if c.RotEnd < 12 {
			tx -= int32(RotationOffset[c.RotEnd][0])
			ty -= int32(RotationOffset[c.RotEnd][1])
		}
		if e, ok := RoadConnectionEnd(tx, ty, tz, tad^(1<<2)); ok {
			if rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, company, roadObjectID, 0, 0); own(rc.StationID) {
				return rc.StationID, true
			}
		}
	}
	return m.FindNearbyStation(x, y, z, company)
}

// validateNearbyRoadStation is CreateRoadStation.cpp's validateNearbyStation
// (0x0048CA3E / 0x0048C9C8): none nearby, a new station; full or too spread
// out, a failure when attached (station_too_large), else a new one.
func (m *Map) validateNearbyRoadStation(x, y, z int32, tad uint16, roadObjectID, company uint8) (int, int) {
	id, attached := m.findNearbyStationOnRoad(x, y, z, tad, roadObjectID, company)
	if id < 0 {
		return nearbyRequiresNewStation, -1
	}
	st := &m.StationNames[id]
	if int(st.TileSize) >= len(st.Tiles) || m.exceedsStationSize(st, x, y) {
		if attached {
			return nearbyFailure, -1
		}
		return nearbyRequiresNewStation, -1
	}
	return nearbyOkay, id
}

// getElRoad is CreateRoadStation.cpp's: the road element on the tile at that
// height of that rotation, piece, road object and sequence index (its index in
// Order, -1 for none).
func (m *Map) getElRoad(x, y, z int32, rotation, roadObjectID, roadID, index uint8) (*scenario.RoadElement, int) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return nil, -1
	}
	for k, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		el := &t.Roads[r.Index]
		if int32(el.BaseZ)*4 == z && el.Rotation == rotation && el.RoadID == roadID && el.RoadObjectID == roadObjectID && el.SequenceIndex == index {
			return el, k
		}
	}
	return nil, -1
}

// elRoadIsJunctionAi is CreateRoadStation.cpp's: a non-ghost road at that
// height that is not a straight (roadId 0) or runs across the rotation.
func (m *Map) elRoadIsJunctionAi(x, y, z int32, rotation uint8) bool {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return false
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		el := &t.Roads[r.Index]
		if int32(el.BaseZ)*4 != z || el.Ghost {
			continue
		}
		if el.RoadID != 0 || (el.Rotation-rotation)&1 != 0 {
			return true
		}
	}
	return false
}
