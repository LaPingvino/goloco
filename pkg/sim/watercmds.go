package sim

import "github.com/LaPingvino/goloco/pkg/i18n"

// GameCommands/Terraform/ChangeLandMaterial.cpp, RaiseWater.cpp,
// LowerWater.cpp and TileManager::adjustWaterHeight.

// ChangeLandMaterial is the changeLandMaterial game command: applying, every
// surface in tiles a..b (clamped; y outer) gets the land type — the growth
// stage 0, a random variation and then the land's last growth stage on
// non-industrial ones —, invalidated. Costs nothing; the position the
// range's middle (world units: (a+b)/2 + 16) at its land height.
// PARTIAL: madeAnyChanges.
func (m *Map) ChangeLandMaterial(c *Cmd, flags uint8, ax, ay, bx, by int, landType uint8) (int32, bool) {
	mx, my := int32(ax+bx)*tileSize/2+16, int32(ay+by)*tileSize/2+16
	land, _ := m.GetHeight(mx, my)
	c.Pos = &[3]int32{mx, my, int32(land)}
	if flags&CmdApply == 0 {
		return 0, true
	}
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &m.Tiles[y][x]
			t.TerrainIndex = landType
			if !t.IsIndustrial {
				t.GrowthStage = 0
			}
			if c.Objects != nil {
				if v, ok := m.randomTerrainVariation(t, c.Objects); ok && !t.IsIndustrial {
					t.Variation = v
				}
				if int(landType) < len(c.Objects.LandObjects) && c.Objects.LandObjects[landType] != nil && !t.IsIndustrial {
					t.GrowthStage = c.Objects.LandObjects[landType].NumGrowthStages - 1
				}
			}
			m.mapInvalidateTileFull(int32(x*tileSize), int32(y*tileSize))
		}
	}
	return 0, true
}

// adjustWaterHeight is TileManager::adjustWaterHeight (target in SmallZ):
// refused below 4 (error_too_low) or from 116 (error_too_high); applying, the
// surface's industry, terrain style (outside the editor) and walls above the
// land go first (as upstream, before the checks); everything between the
// target and the present water / land is cleared with the default
// collision; a ship route's water refused; the water object's cost; then the
// water set (none at or below the land) with the route flag and variation
// cleared.
func (m *Map) adjustWaterHeight(c *Cmd, px, py int32, target int, removed RemovedBuildings, flags uint8) (int32, bool) {
	c.Pos = &[3]int32{px + tileSize/2, py + tileSize/2, int32(target) * microToSmallZStep}
	if target < 4 {
		c.ErrorText = i18n.ErrorTooLow
		return 0, false
	}
	if target >= 116 {
		c.ErrorText = i18n.ErrorTooHigh
		return 0, false
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	if t == nil {
		return 0, false
	}
	var total int32
	if flags&CmdApply != 0 {
		m.removeSurfaceIndustry(px, py, t, c)
		if !c.Editor {
			t.setTerrainStyleAsCleared()
		}
		land, _ := m.GetHeight(px, py)
		t.removeAllWallsAbove(uint8(land / smallZStep))
	}
	ref := int(t.BaseZ)
	if t.WaterLevel > 0 {
		ref = int(t.WaterLevel) * microToSmallZStep
	}
	clear := func(tt *Tile, k int) ClearResult {
		return m.tileClearFunction(c, tt, k, px, py, removed, flags, &total, true)
	}
	if ok, _ := m.ApplyClearAtAllHeights(px, py, uint8(min(target, ref)), uint8(max(target, ref)), 0xF, 0, clear); !ok {
		return 0, false
	}
	if t.Type6Flag {
		c.ErrorText = i18n.WaterChannelCurrentlyNeededByShips
		return 0, false
	}
	if c.Objects != nil && c.Objects.WaterObj != nil {
		total += m.InflationAdjustedCost(c.Objects.WaterObj.CostFactor, c.Objects.WaterObj.CostIndex, 10)
	}
	if flags&CmdApply != 0 {
		if target <= int(t.BaseZ) {
			t.WaterLevel = 0
		} else {
			t.WaterLevel = uint8(target/microToSmallZStep) & 0x1F
		}
		t.Type6Flag = false
		t.Variation = 0
		m.mapInvalidateTileFull(px, py)
	}
	return total, true
}

// RaiseWater is the raiseWater game command (0x004C4F19) over tiles a..b:
// the range's lowest level is found, and only tiles whose land is at or
// below it get water a step higher (dry ones a step above their land).
// Upstream's first pass takes a wet tile's waterHeight() — world units
// (water × kMicroZStep) — against the dry tiles' SmallZ baseZ; ported as is.
// The cost. PARTIAL: madeAnyChanges, the water sound.
func (m *Map) RaiseWater(c *Cmd, flags uint8, ax, ay, bx, by int) (int32, bool) {
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	lowest := 255
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &m.Tiles[y][x]
			z := int(t.BaseZ)
			if t.WaterLevel != 0 {
				z = int(t.WaterLevel) * microZStep // waterHeight(): world units, as upstream
			}
			lowest = min(lowest, int(uint8(z))) // std::min<SmallZ>: an 8-bit SmallZ
		}
	}
	removed := RemovedBuildings{}
	var total int32
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &m.Tiles[y][x]
			if int(t.BaseZ) > lowest {
				continue
			}
			h := int(t.WaterLevel) * microToSmallZStep
			if h == 0 {
				h = int(t.BaseZ) + smallZStep
			} else if h > lowest {
				continue
			} else {
				h += smallZStep
			}
			cost, ok := m.adjustWaterHeight(c, int32(x*tileSize), int32(y*tileSize), h, removed, flags)
			if !ok {
				return 0, false
			}
			total += cost
		}
	}
	return total, true
}

// LowerWater is the lowerWater game command (0x004C5126) over tiles a..b:
// the range's highest water is found and the tiles at it lowered a step.
// The cost. PARTIAL: madeAnyChanges, the water sound (and its position).
func (m *Map) LowerWater(c *Cmd, flags uint8, ax, ay, bx, by int) (int32, bool) {
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	highest := 0
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if w := m.Tiles[y][x].WaterLevel; w != 0 {
				highest = max(highest, int(w)*microToSmallZStep)
			}
		}
	}
	removed := RemovedBuildings{}
	var total int32
	if highest > 0 {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				h := int(m.Tiles[y][x].WaterLevel) * microToSmallZStep
				if h < highest {
					continue
				}
				cost, ok := m.adjustWaterHeight(c, int32(x*tileSize), int32(y*tileSize), h-smallZStep, removed, flags)
				if !ok {
					return 0, false
				}
				total += cost
			}
		}
	}
	return total, true
}
