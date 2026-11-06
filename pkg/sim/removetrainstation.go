package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// TrainStationRemoval is TrainStationRemovalArgs: the track piece (position,
// rotation, piece id, the tile's sequence index, track object) whose
// station is removed.
type TrainStationRemoval struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	TrackObjectID uint8
}

// removeStationElement is TileManager::removeElement for the station element
// at order position k.
func (t *Tile) removeStationElement(k int) {
	idx := int(t.Order[k].Index)
	t.Stations = append(t.Stations[:idx], t.Stations[idx+1:]...)
	t.removeRef(scenario.KindStation, idx)
}

// RemoveTrainStation is the removeTrainStation game command
// (RemoveTrainStation.cpp 0x0048C402): the station elements on every tile of
// the track piece taken off — refused when the track or a station element
// isn't there, or it is another company's. The cost: the station's sell cost
// (× the piece's cost factor / 256; negative — a refund) once per piece.
// Applied: each track's clear height lowered by the station's height (as
// upstream writes it: the full height, though creation adds height / 4),
// then — unless it was a ghost — the tile out of the station and its cargo,
// modes, centre, label and canopies recalculated. The emptied station is
// deallocated later by Station::update, as upstream.
// PARTIAL: the viewport invalidations.
func (m *Map) RemoveTrainStation(c *Cmd, flags uint8, a TrainStationRemoval, g1 G1Raw) (int32, bool) {
	om := c.Objects
	initial, _ := m.getElTrack(a.X, a.Y, a.Z, a.Rotation, a.TrackObjectID, a.TrackID, a.Index)
	if initial == nil || !checkCompanyCompatibility(c, initial.Owner) || int(a.TrackID) >= len(TrackPieces) {
		return 0, false
	}
	pieces := TrackPieces[a.TrackID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	ax, ay := rotOffset(pieces[a.Index].X, pieces[a.Index].Y, a.Rotation)
	sx, sy, sz := a.X-int32(ax), a.Y-int32(ay), a.Z-int32(pieces[a.Index].Z)
	register := true
	found := -1
	var total int32
	for _, piece := range pieces {
		ox, oy := rotOffset(piece.X, piece.Y, a.Rotation)
		px, py, pz := sx+int32(ox), sy+int32(oy), sz+int32(piece.Z)
		el, k := m.getElTrack(px, py, pz, a.Rotation, a.TrackObjectID, a.TrackID, piece.Index)
		if el == nil {
			return 0, false
		}
		t := m.TileAt(int(px)>>5, int(py)>>5)
		if k+1 >= len(t.Order) || t.Order[k+1].Kind != scenario.KindStation {
			return 0, false
		}
		st := &t.Stations[t.Order[k+1].Index]
		if st.Ghost {
			register = false
		}
		found = int(st.StationID)
		if om == nil || int(st.ObjectID) >= len(om.TrainStationObjects) || om.TrainStationObjects[st.ObjectID] == nil {
			return 0, false
		}
		obj := om.TrainStationObjects[st.ObjectID]
		if piece.Index == 0 {
			total += m.InflationAdjustedCost(obj.SellCostFactor, obj.CostIndex, 8) * int32(TrackCostFactor[a.TrackID]) / 256
		}
		if flags&CmdApply != 0 {
			el.ClearZ -= obj.Height
			el.HasStation = false
			t.removeStationElement(k + 1)
		}
	}
	if register && flags&CmdApply != 0 && found >= 0 && found < len(m.StationNames) {
		m.RemoveTileFromStationAndRecalcCargo(found, sx, sy, sz, a.Rotation, om)
		m.RecalculateStationModes(found, om)
		m.RecalculateStationCenter(found, g1)
		if g1 != nil {
			m.UpdateStationLabel(found, g1)
		}
		m.UpdateStationCanopies(found, om)
	}
	return total, true
}
