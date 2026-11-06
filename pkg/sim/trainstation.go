package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Building a train station (GameCommands/Track/CreateTrainStation.cpp): the
// station a new piece joins (findNearbyStationOnTrack, validateNearbyStation)
// and StationManager's findNearbyStation / exceedsStationSize. Item ST.

const (
	kStationDistanceLimit     = 8 * TileSize // StationManager.cpp
	kMaxStationNearbyDistance = 64           // StationManager.h
)

func chebyshev(ax, ay, bx, by int32) int32 { return max(abs32(ax-bx), abs32(ay-by)) }

// exceedsStationSize is StationManager::exceedsStationSize: the tile's
// centre more than 8 tiles from the station's (disableStationSizeLimit is
// off by default).
func (m *Map) exceedsStationSize(st *scenario.StationName, x, y int32) bool {
	return chebyshev(x+TileSize/2, y+TileSize/2, int32(st.X), int32(st.Y)) > kStationDistanceLimit
}

// findNearbyEmptyStation is StationManager's: a station of this company with
// no tiles left (removed, kept a while with its name) within 64 of (x, y, z)
// and nearer than minDistance — the nearest; -1 for none.
func (m *Map) findNearbyEmptyStation(x, y, z int32, company uint8, minDistance int32) int {
	best := -1
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.TileSize != 0 || st.Owner != company {
			continue
		}
		d := chebyshev(int32(st.X), int32(st.Y), x, y)
		if abs32(int32(st.Z)-z) > kMaxStationNearbyDistance || d > kMaxStationNearbyDistance {
			continue
		}
		if d < minDistance {
			minDistance, best = d, id
		}
	}
	return best
}

// FindNearbyStation is StationManager::findNearbyStation (0x004901B0): the
// company's nearest station element (not ghost) within 2 tiles — distance
// plus half the height difference, none more than 64 apart in height;
// attached when within 64 — or an empty station of the company nearer still.
// -1 for none.
func (m *Map) FindNearbyStation(x, y, z int32, company uint8) (int, bool) {
	tx, ty := int(x)>>5, int(y)>>5
	best, minDistance, attached := -1, int32(0x7FFF), false
	for py := max(ty-2, 0); py <= min(ty+2, m.Height-1); py++ {
		for px := max(tx-2, 0); px <= min(tx+2, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el := &t.Stations[r.Index]
				if el.Ghost || int(el.StationID) >= len(m.StationNames) || m.StationNames[el.StationID].Owner != company {
					continue
				}
				d := chebyshev(int32(px)*TileSize, int32(py)*TileSize, x, y)
				if d < minDistance {
					dz := abs32(int32(el.BaseZ)*4 - z)
					if dz > kMaxStationNearbyDistance {
						continue
					}
					minDistance = d + dz/2
					if minDistance <= kMaxStationNearbyDistance {
						attached = true
					}
					best = int(el.StationID)
				}
			}
		}
	}
	if e := m.findNearbyEmptyStation(x, y, z, company, minDistance); e >= 0 {
		return e, attached
	}
	return best, attached
}

// findNearbyStationOnTrack is CreateTrainStation.cpp's (0x0048FF36): the
// station the track continues into at the piece's end, or at its start
// (the piece reversed), else FindNearbyStation. Attached: the first two.
func (m *Map) findNearbyStationOnTrack(x, y, z int32, tad uint16, trackObjectID, company uint8) (int, bool) {
	if e, ok := TrackConnectionEnd(x, y, z, tad); ok {
		if tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, company, trackObjectID, 0, 0); tc.StationID >= 0 {
			return tc.StationID, true
		}
	}
	if int(tad) < len(TrackCoordinates) {
		c := TrackCoordinates[tad]
		tx, ty, tz := x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		if c.RotEnd < 12 {
			tx -= int32(RotationOffset[c.RotEnd][0])
			ty -= int32(RotationOffset[c.RotEnd][1])
		}
		if e, ok := TrackConnectionEnd(tx, ty, tz, tad^(1<<2)); ok {
			if tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, company, trackObjectID, 0, 0); tc.StationID >= 0 {
				return tc.StationID, true
			}
		}
	}
	return m.FindNearbyStation(x, y, z, company)
}

// Nearby station validation (NearbyStationValidation).
const (
	nearbyOkay = iota
	nearbyRequiresNewStation
	nearbyFailure
)

// validateNearbyStation is CreateTrainStation.cpp's (0x0048BDCE): the station
// a piece at (x, y, z) joins — none: a new one; full (80 tiles) or too
// spread out: a failure when attached (station_too_large /
// station_too_spread_out), else a new one. PARTIAL: the AI's variant
// (getTrackConnectionsAi; flags aiAllocated) isn't ported.
func (m *Map) validateNearbyStation(x, y, z int32, tad uint16, trackObjectID, company uint8) (int, int) {
	id, attached := m.findNearbyStationOnTrack(x, y, z, tad, trackObjectID, company)
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

// getElTrack is CreateTrainStation.cpp's: the track element on the tile at
// that height of that rotation, piece, track object and sequence index (its
// index in Order, -1 for none).
func (m *Map) getElTrack(x, y, z int32, rotation, trackObjectID, trackID, index uint8) (*scenario.TrackElement, int) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return nil, -1
	}
	for k, r := range t.Order {
		if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
			continue
		}
		el := &t.Tracks[r.Index]
		if int32(el.BaseZ)*4 == z && el.Rotation == rotation && el.TrackID == trackID && el.TrackObjectID == trackObjectID && el.SequenceIndex == index {
			return el, k
		}
	}
	return nil, -1
}
