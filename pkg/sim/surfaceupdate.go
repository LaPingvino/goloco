package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// updateSurface is SurfaceElement.cpp's updateSurface (0x004691FA), the tile
// loop's surface step: a field's or the grass's growth (each 8th visit),
// the snow cover a step towards the snow line, the water's variation
// counter; then, unless a wave is running, a wave (wave.go createWave).
//
// PARTIAL: updateIndustrialSurface's removal of a builtOnSnow industry's
// field once it thaws, and the field walls it places
// (removeSurfaceIndustry, the wall placement game command).
func (s *State) updateSurface(t *Tile, tx, ty int, visible func(tileX, tileY int) bool, om *objects.ObjectManager) {
	m := &s.Map
	if t.IsIndustrial {
		m.updateIndustrialSurface(t, tx, ty, om)
	} else {
		m.updateNonIndustrialSurface(t, tx, ty, om)
	}

	snow := int(m.CurrentSnowLine) / 4
	target := uint8(min(max(int(t.BaseZ)/4+1-snow, 0), 5))
	if cover := t.Slope >> 5; cover != target {
		if cover > target {
			cover--
		} else {
			cover++
		}
		t.Slope = t.Slope&0x1F | cover<<5
		m.invalidateSurface(tx, ty, t)
	}

	if t.WaterLevel == 0 {
		return
	}
	if t.Type6Flag {
		t.Variation++
		if t.Variation >= 64 {
			t.Variation, t.Type6Flag = 0, false
		}
	}
	if !t.WaveFlag {
		s.createWave(t, tx, ty, visible)
	}
}

// updateNonIndustrialSurface: land with growth stages grows a stage each
// 8th visit (the 3-bit timer wrapping), unless under water — which takes it
// back to stage 0.
func (m *Map) updateNonIndustrialSurface(t *Tile, tx, ty int, om *objects.ObjectManager) {
	lo := om.GetLandObjectByIndex(int(t.TerrainIndex))
	if lo == nil || lo.Flags&objects.LandFlagHasGrowthStages == 0 {
		return
	}
	t.UpdateTimer = (t.UpdateTimer + 1) & 7
	if t.UpdateTimer != 0 {
		return
	}
	if t.WaterLevel != 0 {
		waterZ := int(t.WaterLevel) * 4
		if t.Slope&0x1F != 0 {
			waterZ -= 4
		}
		if waterZ > int(t.BaseZ) {
			if t.GrowthStage != 0 {
				t.GrowthStage = 0
				m.invalidateSurface(tx, ty, t)
			}
			return
		}
	}
	if int(t.GrowthStage)+1 < int(lo.NumGrowthStages) {
		t.GrowthStage++
		m.invalidateSurface(tx, ty, t)
	}
}

// updateIndustrialSurface's growth: a field not resting at its
// no-production stage grows a stage each 8th visit, round to 0 after the
// last.
func (m *Map) updateIndustrialSurface(t *Tile, tx, ty int, om *objects.ObjectManager) {
	obj := m.IndustryObjectByID(t.Variation, om)
	if obj == nil {
		return
	}
	if t.GrowthStage == obj.FarmTileGrowthStageNoProduction && t.GrowthStage != 0 {
		return
	}
	t.UpdateTimer = (t.UpdateTimer + 1) & 7
	if t.UpdateTimer == 0 {
		next := t.GrowthStage + 1
		if next >= obj.FarmTileNumGrowthStages {
			next = 0
		}
		t.GrowthStage = next
		m.invalidateSurface(tx, ty, t)
	}
}

// invalidateSurface is ViewportManager::invalidate(loc, baseHeight,
// baseHeight): the defaults, zoom eighth and radius 32.
func (m *Map) invalidateSurface(tx, ty int, t *Tile) {
	z := int32(t.BaseZ) * smallZStep
	m.invalidateTile(int32(tx)*32, int32(ty)*32, z, z, ZoomEighth, defaultRadius)
}
