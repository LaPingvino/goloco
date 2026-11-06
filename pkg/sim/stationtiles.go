package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A station's tiles (World/Station.cpp): addTileToStation (0x0048F321),
// removeTileFromStation (0x0048F43A) and …AndRecalcCargo (0x0048F482),
// recalculateStationCenter and recalculateStationModes (0x0048F529). Item ST.

const (
	trackObjFlagIsRoad = 1 << 2 // TrackObjectFlags::isRoad
	roadObjFlagIsRail  = 1 << 1 // RoadObjectFlags::isRail
	cargoStatAccepted  = 1 << 0 // StationCargoStatsFlags::accepted (isAccepted)
)

// setAcceptance is the cargo stats' update after calcAcceptedCargo: each
// cargo's industry and whether it is accepted.
func (m *Map) setAcceptance(st *scenario.StationName, accepted uint32, s *CargoSearchState, om *objects.ObjectManager) {
	for c := range st.Cargo {
		if om != nil && c < len(om.CargoObjects) && om.CargoObjects[c] == nil {
			continue // getMaxObjects(cargo) — goloco's stats go to 32 as upstream's
		}
		st.Cargo[c].IndustryID = s.Industry[c]
		if accepted&(1<<c) != 0 {
			st.Cargo[c].Flags |= cargoStatAccepted
		} else {
			st.Cargo[c].Flags &^= cargoStatAccepted
		}
	}
}

// AddTileToStation is addTileToStation: (x, y, z) with the rotation in z's
// low bits appended to the station's tiles, its accepted cargo recalculated;
// the first time a station element is found the station is fully created
// (flag_5 off) and its town counts one station more. PARTIAL: the rival
// companies' "surprised" emotion (company emotions aren't kept).
func (m *Map) AddTileToStation(id int, x, y, z int32, rotation uint8, om *objects.ObjectManager) {
	st := &m.StationNames[id]
	if int(st.TileSize) >= len(st.Tiles) {
		return
	}
	st.Tiles[st.TileSize] = [3]int16{int16(x), int16(y), int16(z&^3 | int32(rotation&3))}
	st.TileSize++
	accepted, s := m.CalcAcceptedCargo(id, om)
	if s.CargoFilterHasBeenRecalculated() && st.Flags&stationFlag5 != 0 {
		st.Flags &^= stationFlag5
		if t := m.townBySlot(st.Town); t != nil {
			t.NumStations++
		}
	}
	m.setAcceptance(st, accepted, &s, om)
}

// RemoveTileFromStation is removeTileFromStation: the tile (with its
// rotation) taken out, the rest moved up; not found, the count still drops
// (clamped), as upstream's bug mitigation.
func (m *Map) RemoveTileFromStation(id int, x, y, z int32, rotation uint8) {
	st := &m.StationNames[id]
	want := [3]int16{int16(x), int16(y), int16(z | int32(rotation))}
	for i := range st.Tiles {
		if st.Tiles[i] != want {
			continue
		}
		copy(st.Tiles[i:], st.Tiles[i+1:])
		st.Tiles[len(st.Tiles)-1] = [3]int16{}
		st.TileSize--
		return
	}
	if st.TileSize > 0 {
		st.TileSize--
	}
}

// RemoveTileFromStationAndRecalcCargo is …AndRecalcCargo: the tile removed,
// the accepted cargo and industries recalculated.
func (m *Map) RemoveTileFromStationAndRecalcCargo(id int, x, y, z int32, rotation uint8, om *objects.ObjectManager) {
	m.RemoveTileFromStation(id, x, y, z, rotation)
	accepted, s := m.CalcAcceptedCargo(id, om)
	m.setAcceptance(&m.StationNames[id], accepted, &s, om)
}

// RecalculateStationCenter is recalculateStationCenter: the station at the
// average of its tiles that still have a station element (+16, their
// centre) and the highest of their heights, its label updated.
func (m *Map) RecalculateStationCenter(id int, g1 G1Raw) {
	st := &m.StationNames[id]
	var sx, sy int32
	var maxZ int16
	n := int32(0)
	for i := 0; i < int(st.TileSize) && i < len(st.Tiles); i++ {
		p := st.Tiles[i]
		if m.getStationElement(int32(p[0]), int32(p[1]), int32(p[2])) == nil {
			continue
		}
		sx += int32(p[0])
		sy += int32(p[1])
		maxZ = max(maxZ, p[2]&^3)
		n++
	}
	if n == 0 {
		return
	}
	st.Z = maxZ
	st.X = int16(sx/n + 16)
	st.Y = int16(sy/n + 16)
	if g1 != nil {
		m.UpdateStationLabel(id, g1)
	}
}

// RecalculateStationModes is recalculateStationModes: the transport modes
// from its (not ghost) station elements — a train station by the track
// under it (a road-like track: road), a road station by the roads under it
// (a rail-like road: rail) and its object's cargo, an airport air, docks
// water — and each cargo's acceptedFromProducer from the cargo they take.
func (m *Map) RecalculateStationModes(id int, om *objects.ObjectManager) {
	st := &m.StationNames[id]
	var accepted uint32
	st.Flags &^= 0xF
	for i := 0; i < int(st.TileSize) && i < len(st.Tiles); i++ {
		p := st.Tiles[i]
		t := m.TileAt(int(p[0])>>5, int(p[1])>>5)
		if t == nil {
			continue
		}
		baseZ := uint8(p[2] / 4)
		k := -1
		var el *scenario.StationElement
		for j, r := range t.Order {
			if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) || t.Stations[r.Index].BaseZ != baseZ {
				continue
			}
			if !t.Stations[r.Index].AiAllocated {
				k, el = j, &t.Stations[r.Index]
			}
			break
		}
		if el == nil || el.Ghost {
			continue
		}
		switch el.StationType {
		case 0: // train station: the element before it is its track
			if k == 0 || t.Order[k-1].Kind != scenario.KindTrack || int(t.Order[k-1].Index) >= len(t.Tracks) {
				break
			}
			tr := &t.Tracks[t.Order[k-1].Index]
			if om != nil && int(tr.TrackObjectID) < len(om.TrackObjects) && om.TrackObjects[tr.TrackObjectID] != nil &&
				om.TrackObjects[tr.TrackObjectID].Flags&trackObjFlagIsRoad != 0 {
				st.Flags |= 1 << 1
			} else {
				st.Flags |= 1 << 0
			}
			accepted = ^uint32(0)
		case 1: // road station: the roads below it at its height
			for j := 0; j < k; j++ {
				r := t.Order[j]
				if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
					continue
				}
				rd := &t.Roads[r.Index]
				if rd.BaseZ != el.BaseZ || rd.AiAllocated || rd.Ghost {
					continue
				}
				if om != nil && int(rd.RoadObjectID) < len(om.RoadObjects) && om.RoadObjects[rd.RoadObjectID] != nil &&
					om.RoadObjects[rd.RoadObjectID].Flags&roadObjFlagIsRail != 0 {
					st.Flags |= 1 << 0
				} else {
					st.Flags |= 1 << 1
				}
			}
			ro := roadStationObject(om, el.ObjectID)
			slot := -1
			if ro != nil {
				slot = cargoSlotByName(om, ro.CargoName)
			}
			switch {
			case ro != nil && ro.Flags&roadStationFlagPassenger != 0 && slot >= 0:
				accepted |= 1 << uint(slot)
			case ro != nil && ro.Flags&roadStationFlagFreight != 0 && slot >= 0:
				accepted |= ^(uint32(1) << uint(slot))
			default:
				accepted = ^uint32(0)
			}
		case 2:
			st.Flags |= 1 << 2
			accepted = ^uint32(0)
		case 3:
			st.Flags |= 1 << 3
			accepted = ^uint32(0)
		}
	}
	for c := range st.Cargo {
		if om == nil || c >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
			continue
		}
		st.Cargo[c].Flags &^= stationCargoAcceptedFromProducer
		if accepted&(1<<c) != 0 {
			st.Cargo[c].Flags |= stationCargoAcceptedFromProducer
		}
	}
}
