package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

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

// quadrantCentreOffsets is CreateTree's _quadrantCentreOffsets.
var quadrantCentreOffsets = [4][2]int32{{7, 7}, {7, 23}, {23, 23}, {23, 7}}

// quadrantFromPos is getQuadrantFromPos.
func quadrantFromPos(px, py int32) uint8 {
	xn, yn := px&0x1F, py&0x1F
	if xn > 16 {
		if yn >= 16 {
			return 0
		}
		return 1
	}
	if yn >= 16 {
		return 3
	}
	return 2
}

// CreateTree is the CreateTree game command (apply, built immediately, no
// ghost) as the map generator uses it: on the tile at (px, py) — already
// tile-aligned — in the quadrant, unless underwater, on no-trees land or
// without clearance (full clearance when requiresFullClearance). PARTIAL:
// TownManager::updateTownInfo (the tree's rating effect on a town), the
// command's cost, error texts and the free-elements check.
func (m *Map) CreateTree(px, py int32, quadrant, rotation, treeType, colour uint8, requiresFullClearance bool, om *objects.ObjectManager) bool {
	if px < 0 || py < 0 || px >= int32(m.Width*tileSize) || py >= int32(m.Height*tileSize) {
		return false
	}
	if int(treeType) >= len(om.TreeObjects) || om.TreeObjects[treeType] == nil {
		return false
	}
	to := om.TreeObjects[treeType]
	qLand, _ := m.GetHeight(px+quadrantCentreOffsets[quadrant][0]-1, py+quadrantCentreOffsets[quadrant][1]-1)
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	if t.WaterLevel&0x1F != 0 && int16(t.WaterLevel&0x1F)*microZStep-1 > qLand {
		return false // cant_build_this_underwater
	}
	if int(t.TerrainIndex) < len(om.LandObjects) && om.LandObjects[t.TerrainIndex] != nil && om.LandObjects[t.TerrainIndex].Flags&objects.LandFlagNoTrees != 0 {
		return false
	}
	baseZ := uint8(qLand / smallZStep)
	clearZ := baseZ + to.InitialHeight/smallZStep
	if requiresFullClearance {
		clearZ = 0xFF
	}
	baseQuarter := uint8(1) << (quadrant ^ 2)
	if !m.CanConstructAt(px, py, baseZ, clearZ, baseQuarter, 0x0F) {
		return false
	}
	el := scenario.TreeElement{TreeObjectID: treeType, Rotation: rotation, Quadrant: quadrant, Colour: colour,
		Season: to.CurrentSeason, Unk7l: 7, BaseZ: baseZ}
	el.Growth = to.Growth - 1 // buildImmediately
	el.ClearZ = to.Height/smallZStep + baseZ
	if int(baseZ)-4 > int(m.CurrentSnowLine) && uint16(to.Flags)&treeHasSnowVariation != 0 {
		el.HasSnow = true
	}
	if kMaxElements-m.EntriesEnd <= kMaxElementsOnOneTile { // checkFreeElementsAndReorganise
		m.EntriesEnd = m.countElements() // its full reorganise (the periodic defrags first: PARTIAL)
	}
	m.EntriesEnd += len(t.Order) + 1 // insertElement copies the tile to the end of the entries
	t.Trees = append(t.Trees, el)
	t.insertRef(scenario.KindTree, len(t.Trees)-1, baseZ)
	for i := range t.Order { // the new ref carries the occupied quarter
		if t.Order[i].Kind == scenario.KindTree && int(t.Order[i].Index) == len(t.Trees)-1 {
			t.Order[i].Flags = baseQuarter
		}
	}
	return true
}

// PlaceRandomTree is placeRandomTree: a tree at a world point — its quadrant
// from the point, a random rotation, a type for the surface unless given —
// placed if CreateTree allows (with full clearance).
func (m *Map) PlaceRandomTree(px, py int32, treeType int, om *objects.ObjectManager) bool {
	quadrant := quadrantFromPos(px, py)
	ax, ay := px&0xFFE0, py&0xFFE0
	rotation := uint8(m.Rng1.RandNextBound(3))
	tt := uint8(treeType)
	if treeType < 0 {
		var ok bool
		if tt, ok = m.RandomTreeTypeFromSurface(int(ax/tileSize), int(ay/tileSize), false, om); !ok {
			return false
		}
	}
	return m.CreateTree(ax, ay, quadrant, rotation, tt, 0, true, om)
}

// PlaceTreeCluster is placeTreeCluster: range² × density / 8192 trees at
// random points within range of the tile's corner (gPrng1, the high-precision
// sine / cosine); true when any was placed.
func (m *Map) PlaceTreeCluster(tx, ty int, rng, density uint16, treeType int, om *objects.ObjectManager) bool {
	n := int(rng) * int(rng) * int(density) / 8192
	errors := 0
	for range n {
		mag := m.Rng1.RandNextBound(0xFFFF) * int32(rng) / 65536
		dir := uint16(m.Rng1.RandNextBound(0x4000 - 1))
		ox, oy := IntegerSinePrecisionHigh(dir, mag), IntegerCosinePrecisionHigh(dir, mag)
		if !m.PlaceRandomTree(ox+int32(tx*tileSize), oy+int32(ty*tileSize), treeType, om) {
			errors++
		}
	}
	return errors < n
}

const (
	kMaxElements          = 442368 // TileManager::kMaxElements
	kMaxElementsOnOneTile = 1024
)

func (m *Map) countElements() int {
	n := 0
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			n += len(m.Tiles[y][x].Order)
		}
	}
	return n
}

// GenerateTrees is generateTrees: forests (random radius, place, density;
// stopping once fewer than 0x1B000 elements are free), random single trees,
// then the altitude cull — on each tile the first tree at or beyond the
// options' altitudes goes (at exactly the limit only when the rotating bit of
// one random mask is set) — and updateTreeSeasons.
func (m *Map) GenerateTrees(o TerrainOptions, om *objects.ObjectManager) {
	for range int(o.NumberOfForests) {
		r := (m.Rng1.RandNextBound(255)*int32(max(int(o.MaxForestRadius)-int(o.MinForestRadius), 0))/255 + int32(o.MinForestRadius)) * tileSize
		// TilePos2(randNext(kMapRows), randNext(kMapColumns)): C++ leaves the
		// order of argument evaluation open; GCC and MSVC evaluate right to
		// left, so the second (y) draws first — as the upstream harness
		// (GCC) does. A Clang build of OpenLoco would swap them.
		ly := int(m.Rng1.RandNextBound(int32(m.Width)))
		lx := int(m.Rng1.RandNextBound(int32(m.Height)))
		density := m.Rng1.RandNextBound(15)*int32(max(int(o.MaxForestDensity)-int(o.MinForestDensity), 0))/15 + int32(o.MinForestDensity)
		m.PlaceTreeCluster(lx, ly, uint16(r), uint16(density), -1, om)
		if kMaxElements-m.EntriesEnd < 0x1B000 {
			break
		}
	}
	for range int(o.NumberRandomTrees) {
		py := m.Rng1.RandNextBound(int32(m.Height * tileSize)) // Pos2(randNext(kMapWidth), randNext(kMapHeight)): right to left, as above
		px := m.Rng1.RandNextBound(int32(m.Width * tileSize))
		m.PlaceRandomTree(px, py, -1, om)
	}
	mask := m.Rng1.RandNext()
	bit := uint32(0)
	for y := range m.Height {
		for x := range m.Width {
			t := m.TileAt(x, y)
			for _, r := range t.Order {
				if r.Kind != scenario.KindTree {
					continue
				}
				h := t.Trees[r.Index].BaseZ // baseHeight / kMicroToSmallZStep
				cull := (h <= o.MinAltitudeForTrees && (h != o.MinAltitudeForTrees || mask&(1<<bit) != 0)) ||
					(h >= o.MaxAltitudeForTrees && (h != o.MaxAltitudeForTrees || mask&(1<<bit) != 0))
				if cull {
					idx := int(r.Index)
					t.Trees = append(t.Trees[:idx], t.Trees[idx+1:]...)
					t.removeRef(scenario.KindTree, idx) // drops the ref, renumbers the later trees
					bit = (bit + 1) % 32
					break
				}
			}
		}
	}
	// updateTreeSeasons: each drawable tile's first tree takes the season.
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			t := m.TileAt(x, y)
			for _, r := range t.Order {
				if r.Kind == scenario.KindTree {
					if tr := &t.Trees[r.Index]; tr.Season < 4 { // isDying is never set on generated trees
						tr.Season, tr.Unk7l = m.CurrentSeason, 7
					}
					break
				}
			}
		}
	}
}
