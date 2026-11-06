package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
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

// findRandomNewIndustryLocation's limits (IndustryManager.cpp).
const (
	findRandomNewIndustryAttempts    = 250
	industryHighGroundMin            = 48 // SmallZ
	industryLowGroundMax             = 56
	industryFlatGroundMountainMax    = 32
	industryTilesToBeInDesertMin     = 100
	industryTilesToBeNearDesertMax   = 70
	industryTilesToBeNearWaterMin    = 10
	industryTilesToBeAwayWaterMax    = 0
	industryDistToBeNearTownMax      = 576
	industryDistToBeAwayTownMin      = 768
	industryNumTreesToBeNearTreesMin = 25
	industryNumTreesToBeOpenSpaceMax = 3
)

// IndustryObjectFlags for where an industry may be built.
const (
	industryBuiltInClusters        = 1 << 0
	industryBuiltOnHighGround      = 1 << 1
	industryBuiltOnLowGround       = 1 << 2
	industryBuiltOnSnow            = 1 << 3
	industryBuiltBelowSnowLine     = 1 << 4
	industryBuiltOnFlatGround      = 1 << 5
	industryBuiltNearWater         = 1 << 6
	industryBuiltAwayFromWater     = 1 << 7
	industryBuiltOnWater           = 1 << 8
	industryBuiltNearTown          = 1 << 9
	industryBuiltAwayFromTown      = 1 << 10
	industryBuiltNearTrees         = 1 << 11
	industryBuiltRequiresOpenSpace = 1 << 12
	industryBuiltInDesert          = 1 << 24
	industryBuiltNearDesert        = 1 << 25
)

// mountainHeight is TileManager::mountainHeight (0x00469A81): over the
// 11×11 tiles (clamped) around a point, the highest surface (a sloped one
// 16 higher, a steep one 32) less the lowest.
func (m *Map) mountainHeight(x, y int32) int32 {
	tx, ty := int(x)>>5, int(y)>>5
	lowest, highest := int32(1<<15-1), int32(0)
	for py := max(ty-5, 0); py <= min(ty+5, m.Height-1); py++ {
		for px := max(tx-5, 0); px <= min(tx+5, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			h := int32(t.BaseZ) * 4
			lowest = min(lowest, h)
			if t.Slope&0x1F != 0 {
				h += 16
				if t.Slope&0x10 != 0 { // isSlopeDoubleHeight
					h += 16
				}
			}
			highest = max(highest, h)
		}
	}
	return highest - lowest
}

// countSurroundingTrees is TileManager::countSurroundingTrees (0x004BE048):
// the trees (not ghosts) on the 11×11 tiles (clamped) around a point.
func (m *Map) countSurroundingTrees(x, y int32) int {
	tx, ty := int(x)>>5-5, int(y)>>5-5
	n := 0
	for py := max(ty, 0); py <= min(ty+10, m.Height-1); py++ {
		for px := max(tx, 0); px <= min(tx+10, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind == scenario.KindTree && r.Flags&0x10 == 0 {
					n++
				}
			}
		}
	}
	return n
}

// findRandomNewIndustryLocation is IndustryManager's (0x004599B3): up to 250
// random tiles (one gPrng1 draw each: its high half the row, its low half the
// column) until one is far enough from other industries and meets the
// object's every "built …" rule (clusters, high / low ground, snow lines,
// flat ground, desert, water, towns, trees); ok false when none did.
func (m *Map) findRandomNewIndustryLocation(objID uint8, om *objects.ObjectManager) (x, y int32, ok bool) {
	obj := om.GetIndustryObjectByIndex(int(objID))
	if obj == nil {
		return 0, 0, false
	}
	f := obj.Flags
	for i := 0; i < findRandomNewIndustryAttempts; i++ {
		r := m.Rng1.RandNext()
		x, y = int32(((r>>16)*384)>>16)*32, int32(((r&0xFFFF)*384)>>16)*32 // toWorldSpace(TilePos2(… × kMapRows, … × kMapColumns))
		if m.isTooCloseToNearbyIndustries(x, y) {
			continue
		}
		if f&industryBuiltInClusters != 0 && m.isOutwithCluster(x, y, objID) {
			continue
		}
		t := m.TileAt(int(x)>>5, int(y)>>5)
		if f&industryBuiltOnHighGround != 0 && (t == nil || t.BaseZ < industryHighGroundMin) {
			continue
		}
		if f&industryBuiltOnLowGround != 0 && (t == nil || t.BaseZ > industryLowGroundMax) {
			continue
		}
		if f&industryBuiltOnSnow != 0 && (t == nil || om.Climate == nil || t.BaseZ < om.Climate.SummerSnowLine) {
			continue
		}
		if f&industryBuiltBelowSnowLine != 0 && (t == nil || om.Climate == nil || t.BaseZ > om.Climate.WinterSnowLine) {
			continue
		}
		if f&industryBuiltOnFlatGround != 0 && m.mountainHeight(x, y) > industryFlatGroundMountainMax {
			continue
		}
		if f&industryBuiltInDesert != 0 && m.countSurroundingDesertTiles(int(x)>>5, int(y)>>5, om) < industryTilesToBeInDesertMin {
			continue
		}
		if f&industryBuiltNearDesert != 0 && m.countSurroundingDesertTiles(int(x)>>5, int(y)>>5, om) >= industryTilesToBeNearDesertMax {
			continue
		}
		if f&industryBuiltNearWater != 0 && m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5) < industryTilesToBeNearWaterMin {
			continue
		}
		if f&industryBuiltAwayFromWater != 0 && m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5) > industryTilesToBeAwayWaterMax {
			continue
		}
		if f&industryBuiltOnWater != 0 && t != nil && t.WaterLevel == 0 {
			continue
		}
		if f&industryBuiltNearTown != 0 {
			ti, density, found := m.ClosestTownAndDensity(x, y)
			if !found {
				continue
			}
			if density == 0 && abs32(x-int32(m.Towns[ti].X))+abs32(y-int32(m.Towns[ti].Y)) > industryDistToBeNearTownMax {
				continue
			}
		}
		if f&industryBuiltAwayFromTown != 0 {
			ti, _, found := m.ClosestTownAndDensity(x, y)
			if !found || abs32(x-int32(m.Towns[ti].X))+abs32(y-int32(m.Towns[ti].Y)) < industryDistToBeAwayTownMin {
				continue
			}
		}
		if f&industryBuiltNearTrees != 0 && m.countSurroundingTrees(x, y) < industryNumTreesToBeNearTreesMin {
			continue
		}
		if f&industryBuiltRequiresOpenSpace != 0 && m.countSurroundingTrees(x, y) > industryNumTreesToBeOpenSpaceMax {
			continue
		}
		return x, y, true
	}
	return 0, 0, false
}
