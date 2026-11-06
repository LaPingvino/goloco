package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Creating an industry (item CR part 2c): GameCommands/Industries/
// CreateIndustry.cpp's getProduced/RequiredCargoBitSet and sub_454C91
// (0x00454C91), and IndustryManager::allocateNewIndustry. PARTIAL: the
// createIndustry command itself (placeIndustryBuilding,
// getRandomAvailableColour, the buildings and their tiles) — next; the name:
// upstream takes the object's var_02 (a StringId given at load) and, when
// another industry of the town has it, a numbered user string — goloco has
// neither string ids nor user strings, so Name only marks the slot used.

const maxIndustries = 128 // Limits::kMaxIndustries

// industryCargoBitSets are getProducedCargoBitSet / getRequiredCargoBitSet:
// a bit per cargo type the object makes / needs.
func industryCargoBitSets(obj *objects.IndustryObject, om *objects.ObjectManager) (produced, required uint32) {
	p, r := om.IndustryCargoIDs(obj)
	for _, c := range p {
		if c != 0xFF {
			produced |= 1 << c
		}
	}
	for _, c := range r {
		if c != 0xFF {
			required |= 1 << c
		}
	}
	return
}

// industryAllocateChecked is sub_454C91: refused when an industry is closer
// than 9 tiles (24 when one feeds the other: a cargo one makes the other
// needs), or no town is near; else allocateNewIndustry with the closest
// town. ok false with the reason upstream sets as the error.
func (m *Map) industryAllocateChecked(objID uint8, x, y int32, prng [2]uint32, om *objects.ObjectManager) (id uint8, reason string, ok bool) {
	obj := om.GetIndustryObjectByIndex(int(objID))
	if obj == nil {
		return 0, "no such industry object", false
	}
	prod, req := industryCargoBitSets(obj, om)
	for i := range m.Industries {
		ind := &m.Industries[i]
		o2 := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if o2 == nil {
			continue
		}
		p2, r2 := industryCargoBitSets(o2, om)
		minDist := int32(32 * 9)
		if req&p2 != 0 || prod&r2 != 0 {
			minDist = 32 * 24
		}
		if abs32(int32(ind.X)-x)+abs32(int32(ind.Y)-y) < minDist {
			return 0, "too_close_to_another_industry", false
		}
	}
	ti, _, found := m.ClosestTownAndDensity(x, y)
	if !found {
		return 0, "town_must_be_built_nearby_first", false
	}
	id, ok = m.allocateNewIndustry(objID, x, y, prng, m.Towns[ti].ID)
	if !ok {
		return 0, "too_many_industries", false
	}
	return id, "", true
}

// allocateNewIndustry is IndustryManager::allocateNewIndustry: the first
// free of the 128 slots made a new, empty industry of the type at (x, y) —
// its prng, under construction (0), production rate 25, founded this year,
// no stations, the histories one entry of 0 —, for the town; kept in slot
// (id) order with the others. ok false: no free slot.
func (m *Map) allocateNewIndustry(objID uint8, x, y int32, prng [2]uint32, town uint16) (uint8, bool) {
	used := map[uint8]bool{}
	for i := range m.Industries {
		used[m.Industries[i].ID] = true
	}
	for id := 0; id < maxIndustries; id++ {
		if used[uint8(id)] {
			continue
		}
		ind := scenario.Industry{ID: uint8(id), Name: 0, ObjectID: objID, X: int16(x), Y: int16(y), Prng: prng,
			UnderConstruction: 0, ProductionRate: 25, FoundingYear: m.CurrentYear, Town: town,
			HistorySize: [2]uint8{1, 1}}
		for c := range ind.ProducedStations {
			for k := range ind.ProducedStations[c] {
				ind.ProducedStations[c][k] = stationIDNull
			}
		}
		at := len(m.Industries)
		for i := range m.Industries {
			if m.Industries[i].ID > uint8(id) {
				at = i
				break
			}
		}
		m.Industries = append(m.Industries, scenario.Industry{})
		copy(m.Industries[at+1:], m.Industries[at:])
		m.Industries[at] = ind
		return uint8(id), true
	}
	return 0, false
}
