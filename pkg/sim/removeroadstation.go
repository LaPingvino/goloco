package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// RoadStationRemoval is RoadStationRemovalArgs: the road piece (position,
// rotation, piece id, the tile's sequence index, road object) whose stop is
// removed.
type RoadStationRemoval struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	Index        uint8
	RoadObjectID uint8
}

// roadElementsRange is RemoveRoadStation.cpp's getRoadElementsRange: the
// order positions of the first road at that height and of the element after
// the last consecutive one (-1, -1: none).
func (m *Map) roadElementsRange(x, y, z int32) (int, int) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	begin, end := -1, -1
	if t == nil {
		return begin, end
	}
	for k, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) || int32(t.Roads[r.Index].BaseZ)*4 != z {
			if end >= 0 {
				break
			}
			continue
		}
		end = k + 1
		if begin < 0 {
			begin = k
		}
	}
	return begin, end
}

// RemoveRoadStation is the removeRoadStation game command
// (RemoveRoadStation.cpp 0x0048D2AC): the stop on the road piece taken off —
// refused without the road, without a stop on it, or when the stop (the
// element right after the tile's roads) is another company's. The cost: the
// stop's sell cost (× the piece's cost factor / 256; negative, a refund) once.
// Applied: the roads lose the station flag and their clear height drops by
// the stop's height (upstream's: the full height, though creation adds
// height / 4), the element goes; unless it was a ghost, the tile leaves the
// station and its modes, centre and label are recalculated.
// Mirrors upstream: each piece's range is looked up at args.pos, not at the
// piece's own tile (the invalidation is the piece's tile's).
func (m *Map) RemoveRoadStation(c *Cmd, flags uint8, a RoadStationRemoval, g1 G1Raw) (int32, bool) {
	om := c.Objects
	initial, _ := m.getElRoad(a.X, a.Y, a.Z, a.Rotation, a.RoadObjectID, a.RoadID, a.Index)
	if initial == nil || !initial.HasStationElement || int(a.RoadID) >= len(RoadPieces) {
		return 0, false
	}
	t := m.TileAt(int(a.X)>>5, int(a.Y)>>5)
	stationAt := func() (*scenario.StationElement, int, int) {
		b, e := m.roadElementsRange(a.X, a.Y, a.Z)
		if b < 0 || e >= len(t.Order) || t.Order[e].Kind != scenario.KindStation {
			return nil, b, e
		}
		return &t.Stations[t.Order[e].Index], b, e
	}
	if st, _, _ := stationAt(); st == nil || !checkCompanyCompatibility(c, st.Owner) {
		return 0, false // vanilla asked the station struct, not the element
	}
	pieces := RoadPieces[a.RoadID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	ax, ay := rotOffset(pieces[a.Index].X, pieces[a.Index].Y, a.Rotation)
	sx, sy, sz := a.X-int32(ax), a.Y-int32(ay), a.Z-int32(pieces[a.Index].Z)
	register := true
	found := -1
	var total int32
	for _, piece := range pieces {
		st, b, e := stationAt() // at args.pos, as upstream
		if st == nil {
			return 0, false
		}
		if st.Ghost {
			register = false
		}
		found = int(st.StationID)
		if om == nil || int(st.ObjectID) >= len(om.RoadStationObjects) || om.RoadStationObjects[st.ObjectID] == nil {
			return 0, false
		}
		obj := om.RoadStationObjects[st.ObjectID]
		if piece.Index == 0 {
			total += m.InflationAdjustedCost(obj.SellCostFactor, obj.CostIndex, 8) * int32(RoadCostFactor[a.RoadID]) / 256
		}
		if flags&CmdApply != 0 {
			for k := b; k < e; k++ {
				if r := t.Order[k]; r.Kind == scenario.KindRoad && int(r.Index) < len(t.Roads) {
					t.Roads[r.Index].HasStationElement = false
					t.Roads[r.Index].ClearZ -= obj.Height
				}
			}
			ox, oy := rotOffset(piece.X, piece.Y, a.Rotation) // roadLoc: the piece's own tile
			m.invalidateElement(sx+int32(ox), sy+int32(oy), st.BaseZ, st.ClearZ)
			t.removeStationElement(e)
		}
	}
	if register && flags&CmdApply != 0 && found >= 0 && found < len(m.StationNames) {
		m.RemoveTileFromStation(found, sx, sy, sz, a.Rotation)
		m.RecalculateStationModes(found, om)
		m.RecalculateStationCenter(found, g1)
		if g1 != nil {
			m.UpdateStationLabel(found, g1)
		}
	}
	return total, true
}
