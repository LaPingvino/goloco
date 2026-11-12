package sim

import "github.com/LaPingvino/goloco/pkg/objects"

const roadFlagAllowUseByAllCompanies = 1 << 7 // RoadObjectFlags::allowUseByAllCompanies

// UpdateRoadObjectIdFlags is ObjectManager::updateRoadObjectIdFlags: the
// GameState bit sets of the loaded road objects that are any-road-type
// compatible and usable by all companies (Scenario::sub_4748D4 runs it after
// object selection, at a scenario's start and at the end of generation).
func (m *Map) UpdateRoadObjectIdFlags(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	var any, all uint32
	for i := range om.RoadObjects {
		r := om.GetRoadObjectByIndex(i)
		if r == nil || i >= 32 {
			continue
		}
		if r.Flags&roadFlagAnyRoadTypeCompatible != 0 {
			any |= 1 << i
		}
		if r.Flags&roadFlagAllowUseByAllCompanies != 0 {
			all |= 1 << i
		}
	}
	m.RoadAnyTypeCompatible, m.RoadUsableByAll = any, all
}

// UpdateLastTrackTypeOption is ObjectManager::updateLastTrackTypeOption
// (0x0047AC05): the two-way any-road-type road for the largest town size
// (the last of equals), 0xFF none.
func (m *Map) UpdateLastTrackTypeOption(om *objects.ObjectManager) {
	largest, last := uint8(0), uint8(0xFF)
	for i := range om.RoadObjects {
		r := om.GetRoadObjectByIndex(i)
		if r == nil || r.Flags&roadFlagAnyRoadTypeCompatible == 0 || r.Flags&roadFlagIsOneWay != 0 {
			continue
		}
		if largest <= r.TargetTownSize {
			largest, last = r.TargetTownSize, uint8(i)
		}
	}
	m.LastTrackTypeOption = last
}

// ResetDefaultLandObject is ObjectManager's resetDefaultLandObject: the
// first land object with distribution pattern 0 (everywhere), else the
// first land object, as lastLandOption.
func (m *Map) ResetDefaultLandObject(om *objects.ObjectManager) {
	for _, pass := range []bool{true, false} {
		for i := range om.LandObjects {
			if l := om.GetLandObjectByIndex(i); l != nil && (!pass || l.DistributionPattern == 0) {
				m.LastLandOption = uint8(i)
				return
			}
		}
	}
}

// UpdateScenarioObjectOptions is the rest of Scenario::sub_4748D4 after the
// road flags and level crossing: the last track type option, the terraform
// objects' default land (resetDefaultLandObject) and the traffic handedness
// (the region's rightHandTraffic). PARTIAL: updatePlayerInfrastructureOptions
// (the toolbar's last road / railroad), the land flag cache and water
// palette (goloco reads them from the objects directly), the currency.
func (m *Map) UpdateScenarioObjectOptions(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	m.UpdateLastTrackTypeOption(om)
	m.TrafficHandedness = 0
	if om.Region != nil && om.Region.Flags&1 != 0 {
		m.TrafficHandedness = 1
	}
	m.ResetDefaultLandObject(om)
}

// staticSet is sfl::static_unordered_set<uint8_t, 8>'s iteration as the
// compiled container gives it: unique values, the most recently first
// inserted first (a value inserted again keeps its place). Checked with a
// probe built against OpenLoco's sfl (2026-11-10: 83 81 80 85 82 84 → 84 82
// 85 80 81 83).
type staticSet []uint8

func (s *staticSet) insert(v uint8) {
	for _, e := range *s {
		if e == v {
			return
		}
	}
	*s = append(staticSet{v}, *s...)
}

// companyAvailableRoads is companyGetAvailableRoads (0x00478265): the road
// types the company's unlocked road vehicles run on and the any-road-type
// roads (bit 7 set), without the rail (tram) ones; then the track types of
// its unlocked rail vehicles that are roads (TrackObjectFlags::isRoad).
func (m *Map) companyAvailableRoads(company uint8, om *objects.ObjectManager) []uint8 {
	var roads staticSet
	for i, v := range om.Vehicles {
		if v != nil && m.isVehicleIndexUnlocked(company, uint16(i)) && v.Mode == objects.TransportModeRoad {
			if tt := om.TrackTypeFor(v); tt != 0xFF {
				roads.insert(tt | 1<<7)
			}
		}
	}
	for i := range om.RoadObjects {
		if r := om.GetRoadObjectByIndex(i); r != nil && r.Flags&roadFlagAnyRoadTypeCompatible != 0 {
			roads.insert(uint8(i) | 1<<7)
		}
	}
	var out []uint8
	for _, id := range roads {
		if r := om.GetRoadObjectByIndex(int(id &^ (1 << 7))); r != nil && r.Flags&roadObjFlagIsRail == 0 {
			out = append(out, id)
		}
	}
	var tracks staticSet // then the rail vehicles' track types that are roads (no bit 7)
	for i, v := range om.Vehicles {
		if v != nil && m.isVehicleIndexUnlocked(company, uint16(i)) && v.Mode == objects.TransportModeRail {
			tracks.insert(om.TrackTypeFor(v))
		}
	}
	for _, id := range tracks {
		if t := om.GetTrackObjectByIndex(int(id)); t != nil && t.Flags&trackObjFlagIsRoad != 0 {
			out = append(out, id)
		}
	}
	return out
}

// companyAvailableRailTracks is companyGetAvailableRailTracks (0x004A6841):
// the track types the company's unlocked rail vehicles run on (not road
// tracks), then the rail (tram) roads (bit 7 set).
func (m *Map) companyAvailableRailTracks(company uint8, om *objects.ObjectManager) []uint8 {
	var tracks staticSet
	for i, v := range om.Vehicles {
		if v != nil && m.isVehicleIndexUnlocked(company, uint16(i)) && v.Mode == objects.TransportModeRail {
			tracks.insert(om.TrackTypeFor(v))
		}
	}
	var out []uint8
	for _, id := range tracks {
		if t := om.GetTrackObjectByIndex(int(id)); t != nil && t.Flags&trackObjFlagIsRoad == 0 {
			out = append(out, id)
		}
	}
	var roads staticSet
	for i, v := range om.Vehicles {
		if v != nil && m.isVehicleIndexUnlocked(company, uint16(i)) && v.Mode == objects.TransportModeRoad {
			if tt := om.TrackTypeFor(v); tt != 0xFF {
				roads.insert(tt | 1<<7)
			}
		}
	}
	for i := range om.RoadObjects {
		if r := om.GetRoadObjectByIndex(i); r != nil && r.Flags&roadFlagAnyRoadTypeCompatible != 0 {
			roads.insert(uint8(i) | 1<<7)
		}
	}
	for _, id := range roads {
		if r := om.GetRoadObjectByIndex(int(id &^ (1 << 7))); r != nil && r.Flags&roadObjFlagIsRail != 0 {
			out = append(out, id)
		}
	}
	return out
}

// UpdatePlayerInfrastructureOptions is CompanyManager's
// updatePlayerInfrastructureOptions: the toolbar's last road (the last track
// type option as a road, else the player's first available road) and last
// railroad (the first available rail track), 0xFF none.
func (m *Map) UpdatePlayerInfrastructureOptions(om *objects.ObjectManager) {
	pc := m.PlayerCompanies[0]
	road := m.LastTrackTypeOption | 1<<7
	if road == 0xFF {
		road = 0xFF
		if r := m.companyAvailableRoads(pc, om); len(r) > 0 {
			road = r[0]
		}
	}
	m.LastRoadOption = road
	m.LastRailroadOption = 0xFF
	if t := m.companyAvailableRailTracks(pc, om); len(t) > 0 {
		m.LastRailroadOption = t[0]
	}
}

// CompanyAvailableRoads is companyGetAvailableRoads, for the toolbar's road
// menu (ToolbarTopCommon.cpp roadMenuMouseDown).
func (m *Map) CompanyAvailableRoads(company uint8, om *objects.ObjectManager) []uint8 {
	return m.companyAvailableRoads(company, om)
}

// CompanyAvailableRailTracks is companyGetAvailableRailTracks, for the
// toolbar's railroad menu (ToolbarTop.cpp railroadMenuMouseDown).
func (m *Map) CompanyAvailableRailTracks(company uint8, om *objects.ObjectManager) []uint8 {
	return m.companyAvailableRailTracks(company, om)
}

// SetTrackOptions is Construction::Common::setTrackOptions (0x004A3A06): the
// track type a construction window opens with becomes the toolbar's last
// road option (a road, or a track that is a road) or last railroad option (a
// track, or a road that is rail — a tram line).
func (m *Map) SetTrackOptions(trackType uint8, om *objects.ObjectManager) {
	if trackType&(1<<7) != 0 {
		if r := om.GetRoadObjectByIndex(int(trackType &^ (1 << 7))); r != nil && r.Flags&roadObjFlagIsRail == 0 {
			m.LastRoadOption = trackType
		} else {
			m.LastRailroadOption = trackType
		}
		return
	}
	if t := om.GetTrackObjectByIndex(int(trackType)); t != nil && t.Flags&trackObjFlagIsRoad == 0 {
		m.LastRailroadOption = trackType
	} else {
		m.LastRoadOption = trackType
	}
}
