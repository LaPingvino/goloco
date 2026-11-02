package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Trees — Map/Tree.cpp (getRandomTreeTypeFromSurface; placeRandomTree and
// placeTreeCluster need the CreateTree game command, not ported yet),
// TileManager::countSurroundingWaterTiles and Math::Trigonometry's
// high-precision integer sine / cosine (kQuarterSine generated).

// IntegerSinePrecisionHigh is integerSinePrecisionHigh (directions 0–0x3FFF).
func IntegerSinePrecisionHigh(direction uint16, magnitude int32) int32 {
	d := int32(direction)
	idx := d
	if direction&(1<<12) != 0 {
		idx = ^d
	}
	v := int32(quarterSine[idx&0xFFF])
	if direction&(1<<13) != 0 {
		v = ^v
	}
	return v * magnitude / 0x8000
}

// IntegerCosinePrecisionHigh is integerCosinePrecisionHigh: sine plus π/2.
func IntegerCosinePrecisionHigh(direction uint16, magnitude int32) int32 {
	return IntegerSinePrecisionHigh(direction+0x4000/4, magnitude)
}

// countSurroundingWaterTiles: surfaces with water in the 11×11 tiles
// (clamped) around a tile.
func (m *Map) countSurroundingWaterTiles(tx, ty int) int {
	n := 0
	for y := max(0, ty-5); y <= min(m.Height-1, ty+5); y++ {
		for x := max(0, tx-5); x <= min(m.Width-1, tx+5); x++ {
			if t := m.TileAt(x, y); t != nil && t.WaterLevel > 0 {
				n++
			}
		}
	}
	return n
}

// Tree object flags (TreeObjectFlags).
const (
	treeHasSnowVariation = 1 << 0
	treeUnk1             = 1 << 1
	treeHighAltitude     = 1 << 2
	treeLowAltitude      = 1 << 3
	treeRequiresWater    = 1 << 4
	treeDroughtResistant = 1 << 6
)

// RandomTreeTypeFromSurface is getRandomTreeTypeFromSurface: a tree object
// suited to the tile's height, snow, land (desert / no trees) and nearby
// water, chosen by gPrng1; false when none fits.
func (m *Map) RandomTreeTypeFromSurface(tx, ty int, unk bool, om *objects.ObjectManager) (uint8, bool) {
	if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height {
		return 0, false
	}
	t := m.TileAt(tx, ty)
	var mustNot, must uint16
	if !unk {
		mustNot |= treeUnk1
	}
	if int(t.BaseZ)-4 > int(m.CurrentSnowLine) {
		must |= treeHasSnowVariation
	}
	if t.BaseZ >= 68 {
		must |= treeHighAltitude
	}
	if t.BaseZ <= 48 {
		must |= treeLowAltitude
	}
	if int(t.TerrainIndex) >= len(om.LandObjects) || om.LandObjects[t.TerrainIndex] == nil {
		return 0, false
	}
	lo := om.LandObjects[t.TerrainIndex]
	mustNot |= treeDroughtResistant
	if lo.Flags&objects.LandFlagIsDesert != 0 {
		must |= treeDroughtResistant
		mustNot &^= treeDroughtResistant
	}
	if lo.Flags&objects.LandFlagNoTrees != 0 {
		return 0, false
	}
	mustNot |= treeRequiresWater
	if m.countSurroundingWaterTiles(tx, ty) >= 8 {
		mustNot &^= treeRequiresWater
	}
	var ok []uint8
	for i, tr := range om.TreeObjects {
		if tr == nil || i >= 64 {
			continue
		}
		f := uint16(tr.Flags)
		if f&mustNot != 0 || f&must != must {
			continue
		}
		ok = append(ok, uint8(i))
	}
	if len(ok) == 0 {
		return 0, false
	}
	r := m.Rng1.RandNext()
	return ok[int(r&0xFFFF)*len(ok)/65536], true
}
