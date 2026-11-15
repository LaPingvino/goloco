package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// updateSurface is SurfaceElement.cpp's updateSurface (0x004691FA), the tile
// loop's surface step: a field's or the grass's growth (each 8th visit),
// the snow cover a step towards the snow line, the water's variation
// counter; then, unless a wave is running, a wave (wave.go createWave).
// False when a field's walls were placed or the field went: the tile's
// elements changed.
func (s *State) updateSurface(t *Tile, tx, ty int, visible func(tileX, tileY int) bool, om *objects.ObjectManager) bool {
	m := &s.Map
	valid := true
	if t.IsIndustrial {
		valid = !m.updateIndustrialSurface(t, tx, ty, om)
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
		return valid
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
	return valid
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

// industryFlagBuiltOnSnow is IndustryObjectFlags::builtOnSnow.
const industryFlagBuiltOnSnow = 1 << 3

// updateIndustrialSurface is SurfaceElement.cpp's: a builtOnSnow industry's
// field goes once the snow cover is under 4; a field not resting at its
// no-production stage grows a stage each 8th visit (round to 0 after the
// last); then its edges: a wall between two resting fields of the industry
// goes, and an edge to land that isn't the industry's (nor holds one of its
// buildings) gets the object's wall — its building wall on a resting field
// that has one. True when the tile's elements changed.
func (m *Map) updateIndustrialSurface(t *Tile, tx, ty int, om *objects.ObjectManager) bool {
	obj := m.IndustryObjectByID(t.Variation, om)
	if obj == nil {
		return false
	}
	x, y := int32(tx)*32, int32(ty)*32
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	if obj.Flags&industryFlagBuiltOnSnow != 0 && t.Slope>>5 < 4 {
		m.removeSurfaceIndustry(x, y, t, c)
		return true
	}
	if t.GrowthStage != obj.FarmTileGrowthStageNoProduction || t.GrowthStage == 0 {
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

	changed := false
	resting := obj.FarmTileGrowthStageNoProduction
	for i := uint8(0); i < 4; i++ {
		nx, ny := x+kRotationOffset[i][0], y+kRotationOffset[i][1]
		if w, ok := m.fieldWall(t, t.BaseZ, i); ok {
			// A wall here: it goes between two resting fields of the industry.
			if t.GrowthStage != resting || !m.validCoords(nx, ny) {
				continue
			}
			n := m.TileAt(int(nx)>>5, int(ny)>>5)
			if n == nil || !n.IsIndustrial || n.Variation != t.Variation || n.GrowthStage != resting {
				continue
			}
			m.RemoveWall(c, CmdApply, x, y, int32(w.BaseZ)*smallZStep, i)
			// upstream carries on with this tile (it notes the divergence a
			// stop would be)
			continue
		}
		if !m.validCoords(nx, ny) {
			continue
		}
		n := m.TileAt(int(nx)>>5, int(ny)>>5)
		if n == nil {
			continue
		}
		if _, ok := m.fieldWall(n, t.BaseZ, i^2); ok {
			continue
		}
		if n.IsIndustrial {
			if n.Variation == t.Variation {
				continue
			}
		} else if n.hasIndustryBuilding() {
			continue
		}
		wallType := obj.WallTypes[0]
		if wallType == 0xFF {
			continue
		}
		if t.GrowthStage != resting || obj.BuildingWall != 0xFF {
			if t.GrowthStage == resting {
				wallType = obj.BuildingWall
			}
			m.CreateWall(c, CmdApply, WallPlacement{X: x, Y: y, Rotation: i, Type: wallType})
			changed = true
		}
	}
	return changed
}

// fieldWall is SurfaceElement.cpp's getWallElement: a wall (not a ghost) on
// t's edge dir overlapping baseZ..baseZ+16.
func (m *Map) fieldWall(t *Tile, baseZ, dir uint8) (*scenario.WallElement, bool) {
	for _, r := range t.Order {
		if r.Kind != scenario.KindWall || r.Flags&elemFlagGhost != 0 || int(r.Index) >= len(t.Walls) {
			continue
		}
		w := &t.Walls[r.Index]
		if int(w.ClearZ) < int(baseZ) || int(w.BaseZ) > int(baseZ)+16 || w.Rotation != dir {
			continue
		}
		return w, true
	}
	return nil, false
}

// hasIndustryBuilding: an industry element (not a ghost) above the tile's
// surface.
func (t *Tile) hasIndustryBuilding() bool {
	passed := false
	for _, r := range t.Order {
		if r.Kind == scenario.KindSurface {
			passed = true
			continue
		}
		if passed && r.Kind == scenario.KindIndustry && r.Flags&elemFlagGhost == 0 {
			return true
		}
	}
	return false
}

// invalidateSurface is ViewportManager::invalidate(loc, baseHeight,
// baseHeight): the defaults, zoom eighth and radius 32.
func (m *Map) invalidateSurface(tx, ty int, t *Tile) {
	z := int32(t.BaseZ) * smallZStep
	m.invalidateTile(int32(tx)*32, int32(ty)*32, z, z, ZoomEighth, defaultRadius)
}
