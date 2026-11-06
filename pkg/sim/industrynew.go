package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
)

// New industries (item CR part 2): IndustryManager.cpp's choice of which
// industry types may start up and how many of each the scenario wants —
// getMostCommonBuildingCargoType, canCargoTypeBeProducedInWorld,
// canIndustryObjBeCreated, capOfTypeOfIndustry / hasReachedCapOfTypeOfIndustry,
// isTooCloseToNearbyIndustries, isOutwithCluster, industryNearPosition.
// PARTIAL: findRandomNewIndustryLocation, createNewIndustry and the
// createIndustry game command, so tryCreateNewIndustriesMonthly is not run
// yet (its gPrng1 draws would come without the creation's).

const (
	industryManagerDisallowStartUp = 1 << 1 // IndustryManager::Flags::disallowIndustriesStartUp
	closeIndustryDistanceMax       = 480    // kCloseIndustryDistanceMax
	industryWithinClusterDistance  = 960    // kIndustryWithinClusterDistance
	numIndustryInCluster           = 3      // kNumIndustryInCluster
)

// mostCommonBuildingCargoType is getMostCommonBuildingCargoType: the cargo
// most (non-misc) town building types produce first — the first of the
// most counted —; with none, the first loaded cargo; else 0.
func mostCommonBuildingCargoType(om *objects.ObjectManager) uint8 {
	var counts [32]uint32
	for i := 0; i < 128; i++ {
		b := om.GetBuildingObjectByIndex(i)
		if b == nil || b.Flags&objects.BuildingFlagMiscBuilding != 0 || b.ProducedQuantity[0] == 0 {
			continue
		}
		if c := om.ProducedCargoIDs(b)[0]; c < 32 {
			counts[c]++
		}
	}
	best := 0
	for i := range counts {
		if counts[i] > counts[best] {
			best = i
		}
	}
	if counts[best] != 0 {
		return uint8(best)
	}
	for i, c := range om.CargoObjects {
		if c != nil && i < 32 {
			return uint8(i)
		}
	}
	return 0
}

// canCargoTypeBeProducedInWorld: the towns' most common cargo, or one an
// existing industry produces.
func (m *Map) canCargoTypeBeProducedInWorld(cargo uint8, om *objects.ObjectManager) bool {
	if cargo == mostCommonBuildingCargoType(om) {
		return true
	}
	for i := range m.Industries {
		obj := om.GetIndustryObjectByIndex(int(m.Industries[i].ObjectID))
		if obj == nil {
			continue
		}
		if p, _ := om.IndustryCargoIDs(obj); p[0] == cargo || p[1] == cargo {
			return true
		}
	}
	return false
}

// canIndustryObjBeCreated: designed and not obsolete this year; a primary
// industry always, else when every required cargo (requiresAllCargo) or at
// least one can be produced in the world.
func (m *Map) canIndustryObjBeCreated(obj *objects.IndustryObject, om *objects.ObjectManager) bool {
	if m.CurrentYear < obj.DesignedYear || m.CurrentYear >= obj.ObsoleteYear {
		return false
	}
	_, req := om.IndustryCargoIDs(obj)
	if req[0] == 0xFF {
		return true
	}
	all := obj.Flags&objects.IndustryFlagRequiresAllCargo != 0
	for _, c := range req {
		if c == 0xFF {
			continue
		}
		ok := m.canCargoTypeBeProducedInWorld(c, om)
		if all && !ok {
			return false
		}
		if !all && ok {
			return true
		}
	}
	return all
}

// capOfTypeOfIndustry: ((numberOfIndustries + 1) × totalOfTypeInScenario / 3)
// less a quarter of it, plus up to half that at random — upstream's
// (half × (randNext(0xFF) / 256)) is integer arithmetic, so 0, but it draws
// from gPrng1 all the same.
func (m *Map) capOfTypeOfIndustry(obj *objects.IndustryObject, factor uint8) int32 {
	i1 := (int32(factor) + 1) * int32(obj.TotalOfTypeInScenario) / 3
	static := i1 - i1/4
	random := (static / 2) * (m.Rng1.RandNextBound(0xFF) / 256)
	return static + random
}

// hasReachedCapOfTypeOfIndustry: the industries of the type already at the cap.
func (m *Map) hasReachedCapOfTypeOfIndustry(objID uint8, om *objects.ObjectManager) bool {
	obj := om.GetIndustryObjectByIndex(int(objID))
	if obj == nil {
		return true
	}
	cap := m.capOfTypeOfIndustry(obj, m.NumberOfIndustries)
	n := int32(0)
	for i := range m.Industries {
		if m.Industries[i].ObjectID == objID {
			n++
		}
	}
	return n >= cap
}

// isTooCloseToNearbyIndustries: an industry within 480 (Manhattan) of loc.
func (m *Map) isTooCloseToNearbyIndustries(x, y int32) bool {
	for i := range m.Industries {
		if abs32(x-int32(m.Industries[i].X))+abs32(y-int32(m.Industries[i].Y)) < closeIndustryDistanceMax {
			return true
		}
	}
	return false
}

// isOutwithCluster: loc is outside every cluster of the type (none of its
// industries within 960) and the type already has 3 or more.
func (m *Map) isOutwithCluster(x, y int32, objID uint8) bool {
	n := 0
	for i := range m.Industries {
		ind := &m.Industries[i]
		if ind.ObjectID != objID {
			continue
		}
		n++
		if abs32(x-int32(ind.X))+abs32(y-int32(ind.Y)) < industryWithinClusterDistance {
			return false
		}
	}
	return n >= numIndustryInCluster
}

// industryNearPosition: an industry with one of the object flags within 11
// tiles (Manhattan, ÷ 32).
func (m *Map) industryNearPosition(x, y int32, flags uint32, om *objects.ObjectManager) bool {
	for i := range m.Industries {
		ind := &m.Industries[i]
		obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if obj == nil || obj.Flags&flags == 0 {
			continue
		}
		if (abs32(int32(ind.X)-x)+abs32(int32(ind.Y)-y))/32 < 11 {
			return true
		}
	}
	return false
}
