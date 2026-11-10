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
