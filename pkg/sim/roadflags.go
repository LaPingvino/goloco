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
