package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// TrainStationPlacement is TrainStationPlacementArgs: a track piece (its
// position, rotation, piece id, the tile's sequence index, track object)
// and the train station object to build on it.
type TrainStationPlacement struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	TrackObjectID uint8
	Type          uint8
}

// insertStationAfter is TileManager::insertElementAfterNoReorg for a station
// element: directly after the element at order position k, with the track's
// occupied quarters (and ghost / AI) in its ref's flags.
func (t *Tile) insertStationAfter(k int, el scenario.StationElement, quarters uint8) int {
	t.Stations = append(t.Stations, el)
	f := quarters & 0xF
	if el.Ghost {
		f |= elemFlagGhost
	}
	if el.AiAllocated {
		f |= elemFlagAiAllocated
	}
	t.Order = append(t.Order, scenario.ElemRef{})
	copy(t.Order[k+2:], t.Order[k+1:])
	t.Order[k+1] = scenario.ElemRef{Kind: scenario.KindStation, Index: uint16(len(t.Stations) - 1), Flags: f}
	return len(t.Stations) - 1
}

// CreateTrainStation is the createTrainStation game command
// (CreateTrainStation.cpp 0x0048BB20): a station of a.Type on every tile of
// the track piece, joining the station it touches or a new one
// (validateNearbyStation / allocateNewStation). Refused: a track or station
// that doesn't take stations on this piece (track_road_unsuitable_for_station),
// no such track, another company's, a signal or level crossing, a junction,
// no room above. The cost: per piece, the station's build cost (× the piece's
// cost factor / 256) — replacing another station type, its sell cost too.
// Applied (not ghost): the station element on each tile (or the old one
// changed), the track's clear height raised by the station's, then
// addTileToStation, the modes, centre, label and canopies.
// PARTIAL: the AI's placement without a track (index 0 query), the AI's
// nearby search, the viewport invalidations and the legacy return state's
// ghost adjoining-station fields.
func (m *Map) CreateTrainStation(c *Cmd, flags uint8, a TrainStationPlacement, g1 G1Raw) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z} // setPosition(args.pos + (16, 16, 0)): where the money effect shows
	om := c.Objects
	if om == nil || int(a.TrackObjectID) >= len(om.TrackObjects) || om.TrackObjects[a.TrackObjectID] == nil ||
		int(a.Type) >= len(om.TrainStationObjects) || om.TrainStationObjects[a.Type] == nil || int(a.TrackID) >= len(TrackPieces) {
		return 0, false
	}
	trackObj, stationObj := om.TrackObjects[a.TrackObjectID], om.TrainStationObjects[a.Type]
	compat := TrackCompatibleFlags[a.TrackID]
	if trackObj.StationTrackPieces&stationObj.TrackPieces&compat != compat {
		return 0, false // track_road_unsuitable_for_station
	}
	initial, _ := m.getElTrack(a.X, a.Y, a.Z, a.Rotation, a.TrackObjectID, a.TrackID, a.Index)
	if initial == nil || !checkCompanyCompatibility(c, initial.Owner) {
		return 0, false
	}
	pieces := TrackPieces[a.TrackID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	ax, ay := rotOffset(pieces[a.Index].X, pieces[a.Index].Y, a.Rotation)
	sx, sy, sz := a.X-int32(ax), a.Y-int32(ay), a.Z-int32(pieces[a.Index].Z)
	tad := uint16(a.TrackID)<<3 | uint16(a.Rotation)

	m.LastAdjoiningStation, m.LastAdjoiningStationX, m.LastAdjoiningStationY = -1, -1, -1
	if flags&CmdGhost != 0 && flags&CmdApply != 0 { // the ghost's would-be station (findNearbyStationOnTrack)
		m.LastAdjoiningStationX, m.LastAdjoiningStationY = sx, sy
		if id, ok := m.findNearbyStationOnTrack(sx, sy, sz, tad, a.TrackObjectID, c.Company); ok {
			m.LastAdjoiningStation = id
		}
	}
	stationID := -1
	registerTile := false
	if flags&CmdGhost == 0 {
		res, near := m.validateNearbyStation(sx, sy, sz, tad, a.TrackObjectID, c.Company)
		if flags&CmdApply != 0 {
			switch res {
			case nearbyRequiresNewStation:
				if id, ok := m.AllocateNewStation(sx, sy, sz, c.Company, 0, om); ok {
					stationID = id
					if g1 != nil {
						m.UpdateStationLabel(id, g1)
					}
				}
			case nearbyOkay:
				stationID = near
			}
			registerTile = true
		} else {
			switch res {
			case nearbyFailure:
				return 0, false
			case nearbyRequiresNewStation:
				id, ok := m.AllocateNewStation(sx, sy, sz, c.Company, 0, om)
				if !ok {
					return 0, false
				}
				m.DeallocateStation(id)
			}
		}
	}

	cost := func(factor int16) int32 {
		return m.InflationAdjustedCost(factor, stationObj.CostIndex, 8) * int32(TrackCostFactor[a.TrackID]) / 256
	}
	var total int32
	for _, piece := range pieces {
		ox, oy := rotOffset(piece.X, piece.Y, a.Rotation)
		px, py, pz := sx+int32(ox), sy+int32(oy), sz+int32(piece.Z)
		el, k := m.getElTrack(px, py, pz, a.Rotation, a.TrackObjectID, a.TrackID, piece.Index)
		if el == nil {
			return 0, false // (the AI's no-track query isn't ported)
		}
		if el.HasSignal || el.HasLevelCrossing {
			return 0, false // signal_in_the_way / level_crossing_in_the_way
		}
		t := m.TileAt(int(px)>>5, int(py)>>5)
		connect := piece.ConnectFlags[el.Rotation]
		for _, r := range t.Order {
			if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
				continue
			}
			o := &t.Tracks[r.Index]
			if o == el || int32(o.BaseZ)*4 != pz || o.Ghost {
				continue
			}
			if connect&TrackPieces[o.TrackID][o.SequenceIndex].ConnectFlags[o.Rotation] != 0 {
				return 0, false // station_cannot_be_built_on_a_junction
			}
		}
		var old *scenario.StationElement
		if el.HasStation {
			if k+1 >= len(t.Order) || t.Order[k+1].Kind != scenario.KindStation {
				return 0, false
			}
			old = &t.Stations[t.Order[k+1].Index]
		}
		if piece.Index == 0 {
			if old == nil || old.ObjectID != a.Type {
				if old != nil && int(old.ObjectID) < len(om.TrainStationObjects) && om.TrainStationObjects[old.ObjectID] != nil {
					total += m.InflationAdjustedCost(om.TrainStationObjects[old.ObjectID].SellCostFactor, om.TrainStationObjects[old.ObjectID].CostIndex, 8) *
						int32(TrackCostFactor[a.TrackID]) / 256
				}
				total += cost(stationObj.BuildCostFactor)
			}
		}
		baseZ := el.BaseZ
		clearZ := baseZ + 8 + stationObj.Height/4
		quarters := t.Order[k].Flags & 0xF
		self := el
		if flags&CmdAiAllocated == 0 {
			aiRes := func(tt *Tile, kk int) ClearResult { // clearFuncAiReservation
				r := tt.Order[kk]
				switch {
				case r.Kind == scenario.KindStation && tt.Stations[r.Index].StationType == 0:
					return ClearNoCollision
				case r.Kind == scenario.KindTrack && &tt.Tracks[r.Index] == self:
					return ClearNoCollision
				}
				return ClearCollision
			}
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ+8, clearZ, quarters, 0, aiRes); !ok {
				return 0, false
			}
		}
		surface := func(tt *Tile, kk int) ClearResult { // clearFuncCollideWithSurface
			if tt.Order[kk].Kind == scenario.KindSurface {
				return ClearCollision
			}
			return ClearNoCollision
		}
		if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ, clearZ, quarters, 0, surface); !ok {
			return 0, false
		}
		if el.HasStation && flags&CmdGhost != 0 {
			return 0, false
		}
		if flags&CmdApply == 0 {
			continue
		}
		var st *scenario.StationElement
		if old != nil {
			if int(old.ObjectID) < len(om.TrainStationObjects) && om.TrainStationObjects[old.ObjectID] != nil {
				el.ClearZ -= om.TrainStationObjects[old.ObjectID].Height / 4
			}
			old.SequenceIndex = 0
			registerTile = false
			st = old
		} else {
			ne := scenario.StationElement{Rotation: el.Rotation, Ghost: flags&CmdGhost != 0, AiAllocated: flags&CmdAiAllocated != 0,
				StationType: 0, BaseZ: el.BaseZ}
			if flags&CmdGhost == 0 && stationID >= 0 {
				ne.StationID = uint16(stationID)
			}
			i := t.insertStationAfter(k, ne, quarters)
			el, _ = m.getElTrack(px, py, pz, a.Rotation, a.TrackObjectID, a.TrackID, piece.Index)
			if el == nil {
				return 0, false
			}
			el.HasStation = true
			st = &t.Stations[i]
		}
		st.ObjectID = a.Type
		el.ClearZ += stationObj.Height / 4
		st.ClearZ = el.ClearZ
		st.Owner = c.Company
	}
	if flags&CmdGhost == 0 && flags&CmdApply != 0 && stationID >= 0 {
		if registerTile {
			m.AddTileToStation(stationID, sx, sy, sz, a.Rotation, om)
		}
		m.RecalculateStationModes(stationID, om)
		m.RecalculateStationCenter(stationID, g1)
		if g1 != nil {
			m.UpdateStationLabel(stationID, g1)
		}
		m.UpdateStationCanopies(stationID, om)
	}
	return total, true
}
