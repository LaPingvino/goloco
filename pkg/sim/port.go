package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// CreatePort.cpp's helpers (item ST, docks): the dock a new one joins, its
// validation, and the water tests. The command itself follows.

// portBorderOffsets is kPortBorderOffsets (Station.h): the eight tiles round a
// 2×2 dock's edges.
var portBorderOffsets = [8][2]int{{-1, 0}, {-1, 1}, {0, 2}, {1, 2}, {2, 1}, {2, 0}, {1, -1}, {0, -1}}

// findNearbyStationDocks is CreatePort.cpp's (0x0049060C; ai: the AI variant
// 0x004906AC): in the 4×4 round the dock's tile (-1 … +2), the company's own
// dock — not a ghost; an AI-allocated one only for ai, else never — attached;
// else FindNearbyStation.
func (m *Map) findNearbyStationDocks(x, y, z int32, company uint8, ai bool) (int, bool) {
	tx, ty := int(x)>>5, int(y)>>5
	for py := max(ty-1, 0); py <= min(ty+2, m.Height-1); py++ {
		for px := max(tx-1, 0); px <= min(tx+2, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el := &t.Stations[r.Index]
				if el.StationType != 3 || el.Ghost || el.AiAllocated != ai || el.Owner != company {
					continue
				}
				return int(el.StationID), true
			}
		}
	}
	return m.FindNearbyStation(x, y, z, company)
}

// validateNearbyDock is CreatePort.cpp's validateNearbyStation (0x0048BDCE /
// 0x0048BD40): none nearby, a new station; full, a failure when attached
// (station_too_large), else new; too spread out (not for the AI), a failure
// when attached (station_too_spread_out), else new.
func (m *Map) validateNearbyDock(x, y, z int32, company uint8, ai bool) (int, int) {
	id, attached := m.findNearbyStationDocks(x, y, z, company, ai)
	if id < 0 || id >= len(m.StationNames) {
		return nearbyRequiresNewStation, -1
	}
	st := &m.StationNames[id]
	if int(st.TileSize) >= len(st.Tiles) || (!ai && m.exceedsStationSize(st, x, y)) {
		if attached {
			return nearbyFailure, -1
		}
		return nearbyRequiresNewStation, -1
	}
	return nearbyOkay, id
}

// isValidWaterIndustryPort is CreatePort.cpp's: a non-ghost industry built on
// water (IndustryObjectFlags::builtOnWater) on one of the dock's border tiles.
func (m *Map) isValidWaterIndustryPort(x, y int32, om *objects.ObjectManager) bool {
	tx, ty := int(x)>>5, int(y)>>5
	for _, o := range portBorderOffsets {
		px, py := tx+o[0], ty+o[1]
		if px < 0 || py < 0 || px >= m.Width || py >= m.Height { // validCoords
			continue
		}
		t := m.TileAt(px, py)
		for _, r := range t.Order {
			if r.Kind != scenario.KindIndustry || int(r.Index) >= len(t.Industries) || r.Flags&elemFlagGhost != 0 {
				continue
			}
			ind := m.industryByID(t.Industries[r.Index].IndustryID)
			if ind == nil || om == nil {
				continue
			}
			if obj := om.GetIndustryObjectByIndex(int(ind.ObjectID)); obj != nil && obj.Flags&industryBuiltOnWater != 0 {
				return true
			}
		}
	}
	return false
}

// tileHasWater is CreatePort.cpp's: the tile is on the map and its surface
// has water.
func (m *Map) tileHasWater(x, y int32) bool {
	tx, ty := int(x)>>5, int(y)>>5
	if x < 0 || y < 0 || tx >= m.Width || ty >= m.Height {
		return false
	}
	t := m.TileAt(tx, ty)
	return t != nil && t.WaterLevel != 0
}
