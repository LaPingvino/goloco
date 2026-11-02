package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// New towns — GameCommands/Town/CreateTown.cpp's placement (item 5.9f).
// PARTIAL: TownManager::initialiseTown (the town record, its name), the
// growth (Town::grow eight rounds of size² with the year moved back 51 years),
// recalculateSize — so the generator does not place towns yet.

// countNearbyWaterTiles is TileManager::countNearbyWaterTiles: water
// surfaces on every other tile (a checkerboard keyed to the area's corner) of
// the 21×21 tiles (clamped) around a tile.
func (m *Map) countNearbyWaterTiles(tx, ty int) int {
	ix, iy := tx-10, ty-10
	n := 0
	for y := max(0, iy); y <= min(m.Height-1, iy+20); y++ {
		for x := max(0, ix); x <= min(m.Width-1, ix+20); x++ {
			if x&1 != ix&1 || y&1 != iy&1 {
				continue
			}
			if t := m.TileAt(x, y); t != nil && t.WaterLevel > 0 {
				n++
			}
		}
	}
	return n
}

// townSiteOK is CreateTown's checkSurroundings: flat, dry; on desert or
// no-trees land (when checking water) little water nearby fails unless a
// 1-in-256 draw lets it; at least 768 (manhattan) from every town.
func (m *Map) townSiteOK(px, py int32, checkWater bool, om *objects.ObjectManager) bool {
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	if t.Slope&0x1F != 0 || t.WaterLevel != 0 {
		return false
	}
	if checkWater && int(t.TerrainIndex) < len(om.LandObjects) && om.LandObjects[t.TerrainIndex] != nil {
		if om.LandObjects[t.TerrainIndex].Flags&(objects.LandFlagIsDesert|objects.LandFlagNoTrees) != 0 {
			if m.countNearbyWaterTiles(int(px/tileSize), int(py/tileSize)) < 10 && m.Rng1.RandNext()&0xFF != 0 {
				return false
			}
		}
	}
	for _, tw := range m.Towns {
		d := absInt(int(px)-int(tw.X)) + absInt(int(py)-int(tw.Y))
		if d < 768 {
			return false
		}
	}
	return true
}

// FindTownSite is createTown's position search for a town placed anywhere
// (args.pos = −1): up to 200 random tiles inside the map's 12-tile border
// (gPrng1), the first whose surroundings pass. ok false when none did.
func (m *Map) FindTownSite(om *objects.ObjectManager) (px, py int32, ok bool) {
	for attempts := 200; attempts > 0; attempts-- {
		r := m.Rng1.RandNext()
		x := int32(((r >> 16) * uint32(m.Width)) >> 16 * tileSize)
		y := int32(((r & 0xFFFF) * uint32(m.Height)) >> 16 * tileSize)
		if x < 384 || y < 384 || x > 11904 || y > 11904 {
			continue
		}
		if m.townSiteOK(x, y, true, om) {
			return x, y, true
		}
	}
	return 0, 0, false
}

// countSurroundingDesertTiles is TileManager::countSurroundingDesertTiles:
// dry surfaces on desert land in the 11×11 tiles (clamped) around a tile.
func (m *Map) countSurroundingDesertTiles(tx, ty int, om *objects.ObjectManager) int {
	n := 0
	for y := max(0, ty-5); y <= min(m.Height-1, ty+5); y++ {
		for x := max(0, tx-5); x <= min(m.Width-1, tx+5); x++ {
			t := m.TileAt(x, y)
			if t == nil || t.WaterLevel != 0 || int(t.TerrainIndex) >= len(om.LandObjects) || om.LandObjects[t.TerrainIndex] == nil {
				continue
			}
			if om.LandObjects[t.TerrainIndex].Flags&objects.LandFlagIsDesert != 0 {
				n++
			}
		}
	}
	return n
}

// CalcCargoInfluenceFlags is TownManager's calcCargoInfluenceFlags: a bit per
// region cargo influence object whose town filter the town passes (all towns;
// may snow — its tile at or above the summer snow line; in desert — ≥ 100
// desert tiles around).
func (m *Map) CalcCargoInfluenceFlags(townX, townY int16, om *objects.ObjectManager) uint32 {
	r := om.Region
	if r == nil {
		return 0
	}
	ids := om.CargoInfluenceObjectIDs(r)
	var flags uint32
	tx, ty := int(townX)/tileSize, int(townY)/tileSize
	for i := range ids {
		has := false
		switch r.CargoInfluenceTownFilter[i] {
		case 0:
			has = true
		case 1:
			if t := m.TileAt(tx, ty); t != nil && om.Climate != nil {
				has = t.BaseZ >= om.Climate.SummerSnowLine
			}
		case 2:
			has = m.countSurroundingDesertTiles(tx, ty, om) >= 100
		}
		if has {
			flags |= 1 << ids[i]
		}
	}
	return flags
}
