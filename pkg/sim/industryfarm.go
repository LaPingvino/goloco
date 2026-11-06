package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The farm tiles (item CR part 3): Industry::update (Industry.cpp), run for
// every industry each tick by IndustryManager::update — a tile loop over the
// whole map, 100 tiles a tick, counting the industry's fields that produce
// (isFarmTileProducing, 0x0045329B) and, each time round, the production
// rate they give (calculateFarmProduction, 0x00453366), the farm growing a
// field when it is small and its cargo is taken. As upstream, a field under
// snow is idle unless the object has farmProductionIgnoresSnow — which the
// original game didn't honour: its saves count a Managed Forest's snowy
// fields idle (title.dat's industry 23: 12 snowy + 9 treed = its 21).

const industryFarmIgnoresSnow = 1 << 28 // IndustryObjectFlags::farmProductionIgnoresSnow

// tileLoopNext is TileLoop::next: a tile on, a row down at the end, back to
// the start at the bottom.
func tileLoopNext(p [2]int16) [2]int16 {
	const mapSize = 384 * 32 // kMapWidth / kMapHeight
	p[0] += 32
	if int32(p[0]) >= mapSize-1 {
		p[0] = 0
		p[1] += 32
		if int32(p[1]) >= mapSize-1 {
			p[1] = 0
		}
	}
	return p
}

// industriesUpdate is IndustryManager::update.
func (m *Map) industriesUpdate(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	for i := range m.Industries {
		m.industryUpdate(&m.Industries[i], om)
	}
}

// industryUpdate is Industry::update: a built, real industry's tile loop
// takes 100 steps a tick; at the map's end its farm production is worked out.
func (m *Map) industryUpdate(ind *scenario.Industry, om *objects.ObjectManager) {
	if ind.Flags&industryFlagIsGhost != 0 || ind.UnderConstruction != 0xFF {
		return
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	for i := 0; i < 100; i++ {
		m.isFarmTileProducing(ind, obj, ind.TileLoop)
		ind.TileLoop = tileLoopNext(ind.TileLoop)
		if ind.TileLoop == [2]int16{} {
			m.calculateFarmProduction(ind, obj, om)
			break
		}
	}
}

// isFarmTileProducing counts the tile when it is one of the industry's fields
// past the stage that makes nothing (or at stage 0), and as idle when snow
// covers it (unless the object ignores snow) or a tree stands on it.
func (m *Map) isFarmTileProducing(ind *scenario.Industry, obj *objects.IndustryObject, p [2]int16) {
	t := m.TileAt(int(p[0])>>5, int(p[1])>>5)
	if t == nil || !t.IsIndustrial || t.Variation != ind.ID {
		return
	}
	if t.GrowthStage != 0 && t.GrowthStage == obj.FarmTileGrowthStageNoProduction {
		return
	}
	ind.NumFarmTiles++
	// findTreeAboveSurface: a tree above the surface — every tree is (they
	// stand on the ground).
	tree := len(t.Trees) > 0
	if (obj.Flags&industryFarmIgnoresSnow == 0 && t.Slope>>5 != 0) || tree {
		ind.NumIdleFarmTiles++
	}
}

// calculateFarmProduction: the productive fields (the idle less a sixteenth of
// all, so winter never stops it) per 25 make the production rate out of 256
// against the object's ideal number of fields (255 at or over it); the
// counts reset; a farm under 224 whose cargo was at least an eighth taken
// last month grows, half the time, by expandGrounds at a random spot within
// 15 tiles (a random wall pair, update timer).
func (m *Map) calculateFarmProduction(ind *scenario.Industry, obj *objects.IndustryObject, om *objects.ObjectManager) {
	idle := max(0, ind.NumIdleFarmTiles-ind.NumFarmTiles/16)
	productive := ind.NumFarmTiles - idle
	size := min(productive/25, 255)
	if size < int16(obj.FarmNumFields) {
		ind.ProductionRate = uint8(int32(size) * 256 / int32(obj.FarmNumFields))
	} else {
		ind.ProductionRate = 255
	}
	ind.NumFarmTiles, ind.NumIdleFarmTiles = 0, 0
	if ind.ProductionRate >= 224 {
		return
	}
	if ind.ProducedPrevMonth[0]/8 > ind.DeliveredPrevMonth[0] && ind.ProducedPrevMonth[1]/8 > ind.DeliveredPrevMonth[1] {
		return
	}
	p := core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
	defer func() { ind.Prng = [2]uint32{p.Srand0(), p.Srand1()} }()
	if !p.RandBool() {
		return
	}
	x := int32(ind.X) + p.RandNextRange(-15, 16)*32
	y := int32(ind.Y) + p.RandNextRange(-15, 16)*32
	wt, we := obj.WallTypes[0], obj.WallTypes[1]
	if obj.WallTypes[2] != 0xFF && p.RandBool() {
		wt, we = obj.WallTypes[2], obj.WallTypes[3]
	}
	timer := uint8(p.RandNextBound(7))
	ind.Prng = [2]uint32{p.Srand0(), p.Srand1()}
	m.expandGrounds(&Cmd{Company: CompanyNeutral, Objects: om}, ind, x, y, wt, we, 0, timer)
}
