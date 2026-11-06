package sim

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// RoadStationPlacement is RoadStationPlacementArgs: a road piece (position,
// rotation, piece id, the tile's sequence index, road object) and the road
// station object to build on it.
type RoadStationPlacement struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	Index        uint8
	RoadObjectID uint8
	Type         uint8
}

const (
	roadStationFlagRoadEnd          = 1 << 3 // RoadStationFlags::roadEnd
	roadObjFlagAnyRoadTypeCompat    = 1 << 3 // RoadObjectFlags::anyRoadTypeCompatible
	roadObjFlagIsRoad               = 1 << 6 // RoadObjectFlags::isRoad
	roadObjFlagAllowUseByAllCompany = 1 << 7 // RoadObjectFlags::allowUseByAllCompanies
)

// CreateRoadStation is the createRoadStation game command
// (CreateRoadStation.cpp 0x0048C708): a road stop of a.Type on the road
// piece's tiles, joining the station it touches or a new one. Refused: a stop
// that doesn't fit the piece (track_road_unsuitable_for_station), a road-end
// stop with road on both sides (station_type_can_only_be_built_at_road_ends,
// queries only), no such road, another company's stop, a signal or level
// crossing, a junction, a tram track for a road-end stop
// (station_not_compatible_with_string_id), a road the stop doesn't fit
// (wrong_type_of_station_for_string_id), a road nobody may use, no room.
// Cost on the piece's first tile: the stop's build cost (× the piece's cost
// factor / 256), and replacing another stop type its sell cost too. Applied:
// the station element after the tile's roads at that height (or the old one
// changed), the roads' clear height raised by the stop's; then
// addTileToStation, the modes, centre and label.
// Mirrors upstream's quirks: the clearance's base height is the piece start's
// (roadStart.z, "vanilla bug"), and the first road at the height isn't spared
// by clearFunc0 (elRoad > elReferenceRoad0).
// PARTIAL: the AI's placement without a road (aiAllocated, index 0), the
// error texts, the company's thinking emotion, the viewport invalidations.
func (m *Map) CreateRoadStation(c *Cmd, flags uint8, a RoadStationPlacement, g1 G1Raw) (int32, bool) {
	om := c.Objects
	if om == nil || int(a.RoadObjectID) >= len(om.RoadObjects) || om.RoadObjects[a.RoadObjectID] == nil ||
		int(a.Type) >= len(om.RoadStationObjects) || om.RoadStationObjects[a.Type] == nil || int(a.RoadID) >= len(RoadPieces) {
		return 0, false
	}
	stationObj := om.RoadStationObjects[a.Type]
	compat := RoadCompatibleFlags[a.RoadID]
	if stationObj.RoadPieces&compat != compat {
		return 0, false // track_road_unsuitable_for_station
	}
	if flags&CmdApply == 0 && stationObj.Flags&roadStationFlagRoadEnd != 0 {
		connected := func(tad uint16) bool {
			e, ok := RoadConnectionEnd(a.X, a.Y, a.Z, tad)
			return ok && len(m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, c.Company, a.RoadObjectID, 0, 0).Connections) > 0
		}
		if connected(uint16(a.Rotation)) && connected(1<<2|uint16(a.Rotation)) {
			return 0, false // station_type_can_only_be_built_at_road_ends
		}
	}
	initial, _ := m.getElRoad(a.X, a.Y, a.Z, a.Rotation, a.RoadObjectID, a.RoadID, a.Index)
	if initial == nil {
		return 0, false // (the AI's no-road query isn't ported)
	}
	if initial.HasStationElement {
		st := m.getStationElement(a.X, a.Y, a.Z)
		if st == nil || !checkCompanyCompatibility(c, st.Owner) {
			return 0, false
		}
	}
	pieces := RoadPieces[a.RoadID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	ax, ay := rotOffset(pieces[a.Index].X, pieces[a.Index].Y, a.Rotation)
	sx, sy, sz := a.X-int32(ax), a.Y-int32(ay), a.Z-int32(pieces[a.Index].Z)
	tad := uint16(a.RoadID)<<3 | uint16(a.Rotation)

	m.LastAdjoiningStation, m.LastAdjoiningStationX, m.LastAdjoiningStationY = -1, -1, -1
	if flags&CmdGhost != 0 && flags&CmdApply != 0 {
		m.LastAdjoiningStationX, m.LastAdjoiningStationY = sx, sy
		if id, ok := m.findNearbyStationOnRoad(sx, sy, sz, tad, a.RoadObjectID, c.Company); ok || id >= 0 {
			m.LastAdjoiningStation = id
		}
	}
	stationID := -1
	registerTile := false
	if flags&CmdGhost == 0 {
		res, near := m.validateNearbyRoadStation(sx, sy, sz, tad, a.RoadObjectID, c.Company)
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

	stationName := ""
	if stationObj.Header != nil {
		stationName = stationObj.Header.GetName()
	}
	cost := func(factor int16, costIndex uint8) int32 {
		return m.InflationAdjustedCost(factor, costIndex, 8) * int32(RoadCostFactor[a.RoadID]) / 256
	}
	var total int32
	for _, piece := range pieces {
		ox, oy := rotOffset(piece.X, piece.Y, a.Rotation)
		px, py, pz := sx+int32(ox), sy+int32(oy), sz+int32(piece.Z)
		el, _ := m.getElRoad(px, py, pz, a.Rotation, a.RoadObjectID, a.RoadID, piece.Index)
		if el == nil {
			return 0, false
		}
		if el.HasSignalElement || el.HasLevelCrossing {
			return 0, false // signal_in_the_way / level_crossing_in_the_way
		}
		t := m.TileAt(int(px)>>5, int(py)>>5)
		if el.RoadID == 0 {
			if m.elRoadIsJunctionAi(px, py, pz, a.Rotation) {
				return 0, false // station_cannot_be_built_on_a_junction
			}
		} else {
			connect := piece.ConnectFlags[el.Rotation]
			for _, r := range t.Order {
				if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
					continue
				}
				o := &t.Roads[r.Index]
				if o == el || int32(o.BaseZ)*4 != pz || o.Ghost || int(o.RoadID) >= len(RoadPieces) || int(o.SequenceIndex) >= len(RoadPieces[o.RoadID]) {
					continue
				}
				if connect&RoadPieces[o.RoadID][o.SequenceIndex].ConnectFlags[o.Rotation] != 0 {
					return 0, false // station_cannot_be_built_on_a_junction
				}
			}
		}

		// the tile's roads at this height: the first (k0) and last (k1), each
		// one the stop fits — by the stop's compatible roads or the road's
		// stations — and one at least the company may use
		k0, k1 := -1, -1
		usable := false
		for k, r := range t.Order {
			if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
				continue
			}
			o := &t.Roads[r.Index]
			if int32(o.BaseZ)*4 != pz {
				continue
			}
			if int(o.RoadObjectID) >= len(om.RoadObjects) || om.RoadObjects[o.RoadObjectID] == nil {
				return 0, false
			}
			ro := om.RoadObjects[o.RoadObjectID]
			if stationObj.Flags&roadStationFlagRoadEnd != 0 && ro.Flags&roadObjFlagIsRoad == 0 {
				return 0, false // station_not_compatible_with_string_id
			}
			if ro.Flags&(roadObjFlagAnyRoadTypeCompat|roadObjFlagAllowUseByAllCompany) != 0 || checkCompanyCompatibility(c, o.Owner) {
				usable = true
			}
			fits := false
			if ro.Header != nil {
				for _, n := range stationObj.CompatibleNames { // stationObj->mods[]
					fits = fits || strings.EqualFold(n, ro.Header.GetName())
				}
			}
			for _, n := range ro.StationNames { // roadObj->stations[]
				fits = fits || strings.EqualFold(n, stationName)
			}
			if !fits {
				return 0, false // wrong_type_of_station_for_string_id
			}
			k1 = k
			if k0 < 0 {
				k0 = k
			}
		}
		if !usable || k1 < 0 {
			return 0, false
		}
		last := &t.Roads[t.Order[k1].Index]
		var old *scenario.StationElement
		if last.HasStationElement {
			if k1+1 >= len(t.Order) || t.Order[k1+1].Kind != scenario.KindStation {
				return 0, false
			}
			old = &t.Stations[t.Order[k1+1].Index]
		}
		if piece.Index == 0 && (old == nil || old.ObjectID != a.Type) {
			if old != nil && int(old.ObjectID) < len(om.RoadStationObjects) && om.RoadStationObjects[old.ObjectID] != nil {
				o := om.RoadStationObjects[old.ObjectID]
				total += cost(o.SellCostFactor, o.CostIndex)
			}
			total += cost(stationObj.BuildCostFactor, stationObj.CostIndex)
		}

		baseZ := uint8(sz / 4) // roadStart.z: "Vanilla bug? should really be roadLoc.z"
		clearZ := baseZ + 8 + stationObj.Height/4
		roads := func(tt *Tile, kk int) ClearResult { // clearFunc0
			r := tt.Order[kk]
			switch {
			case r.Kind == scenario.KindStation && tt.Stations[r.Index].StationType == 1:
				return ClearNoCollision
			case r.Kind == scenario.KindRoad && tt == t && kk > k0 && kk <= k1:
				return ClearNoCollision
			}
			return ClearCollision
		}
		if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ+8, clearZ, 0xF, 0, roads); !ok {
			return 0, false
		}
		if flags&CmdAiAllocated == 0 {
			surface := func(tt *Tile, kk int) ClearResult { // clearFuncCollideWithSurface
				if tt.Order[kk].Kind == scenario.KindSurface {
					return ClearCollision
				}
				return ClearNoCollision
			}
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ, clearZ, 0xF, 0, surface); !ok {
				return 0, false
			}
		}
		if old != nil && flags&CmdGhost != 0 {
			return 0, false
		}
		if flags&CmdApply == 0 {
			continue
		}
		eachRoad := func(f func(*scenario.RoadElement)) { // elRoadEntries[0] … [1]
			for k := k0; k <= k1; k++ {
				if r := t.Order[k]; r.Kind == scenario.KindRoad && int(r.Index) < len(t.Roads) {
					f(&t.Roads[r.Index])
				}
			}
		}
		var st *scenario.StationElement
		if old != nil {
			if int(old.ObjectID) < len(om.RoadStationObjects) && om.RoadStationObjects[old.ObjectID] != nil {
				h := om.RoadStationObjects[old.ObjectID].Height / 4
				eachRoad(func(r *scenario.RoadElement) { r.ClearZ -= h })
			}
			registerTile = false // just replacing: not a new tile
			st = old
		} else {
			ne := scenario.StationElement{Rotation: last.Rotation, Ghost: flags&CmdGhost != 0, AiAllocated: flags&CmdAiAllocated != 0,
				StationType: 1, BaseZ: last.BaseZ}
			if flags&CmdGhost == 0 && stationID >= 0 {
				ne.StationID = uint16(stationID)
			}
			i := t.insertStationAfter(k1, ne, t.Order[k1].Flags&0xF)
			eachRoad(func(r *scenario.RoadElement) { r.HasStationElement = true })
			st = &t.Stations[i]
		}
		st.ObjectID = a.Type
		eachRoad(func(r *scenario.RoadElement) { r.ClearZ += stationObj.Height / 4 })
		st.ClearZ = t.Roads[t.Order[k0].Index].ClearZ
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
	}
	return total, true
}

// RoadStationCargoFilter is onToolUpdateRoadStation's filter: a passenger stop
// counts only its cargo, a freight bay every cargo but that one, else all.
func RoadStationCargoFilter(om *objects.ObjectManager, typ int) uint32 {
	if om == nil || typ < 0 || typ >= len(om.RoadStationObjects) || om.RoadStationObjects[typ] == nil {
		return ^uint32(0)
	}
	ro := om.RoadStationObjects[typ]
	cargo := uint(cargoSlotByName(om, ro.CargoName)) & 31
	filter := ^uint32(0)
	if ro.Flags&roadStationFlagPassenger != 0 {
		filter = 1 << cargo
	}
	if ro.Flags&roadStationFlagFreight != 0 {
		filter = ^uint32(0) &^ (1 << cargo)
	}
	return filter
}
