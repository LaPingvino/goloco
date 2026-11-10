package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// New towns — GameCommands/Town/CreateTown.cpp's placement (item 5.9f).
// initialiseTown (the record, its name) and recalculateSize are ported below
// and in towngrow.go; the command itself with its growth (eight rounds of
// size² Town::grow from 51 years back) is CreateTown (createtown.go), which
// the generator's GenerateTowns calls.

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

const (
	g1FontBase        = 1116 // the first glyph image (medium normal, then medium bold, small, large; 224 each)
	kUserStringSize   = 32
	kMaxUserStrings   = 2048
	kUserStringsStart = 0x8000
)

// MediumBoldStringWidth is Gfx::TextRenderer::getStringWidth for
// Font::medium_bold on plain text: the characters' widths (initialiseCharacterWidths:
// the glyph's g1 width − 1, characters 123–150 zero).
func MediumBoldStringWidth(g1 G1Raw, s string) int {
	w := 0
	for i := 0; i < len(s); i++ {
		chr := int(s[i])
		if chr < 32 || chr >= 123 && chr <= 150 {
			continue
		}
		if gw, _, _, ok := g1.ElementRaw(g1FontBase + 224 + chr - 32); ok {
			w += int(uint8(int8(gw) - 1))
		}
	}
	return w
}

// userStringAllocate is StringManager::userStringAllocate: the text into the
// LAST empty user string slot (upstream's loop keeps the last one it sees),
// refused when unique and already there; 0 (StringIds::empty) on failure.
func (m *Map) userStringAllocate(s string, unique bool) uint16 {
	for len(m.UserStrings) < kMaxUserStrings {
		m.UserStrings = append(m.UserStrings, "")
	}
	best := -1
	for i, u := range m.UserStrings[:kMaxUserStrings] {
		if u == "" {
			best = i
		} else if unique && u == s {
			return 0
		}
	}
	if best < 0 {
		return 0
	}
	if len(s) > kUserStringSize-1 {
		s = s[:kUserStringSize-1]
	}
	m.UserStrings[best] = s
	return uint16(best + kUserStringsStart)
}

// GenerateTownName is TownManager's generateTownName for town i: up to 400
// names drawn from the town's own prng, the first that is not empty, fits 32
// characters and 200 pixels of medium bold, suits the place (large / small
// water body nearby, not mountainous), and no other town has.
func (m *Map) GenerateTownName(i int, g1 G1Raw, om *objects.ObjectManager) bool {
	tw := &m.Towns[i]
	tn := om.TownNames
	if tn == nil {
		return false
	}
	p := core.NewPrngSeed(tw.Prng[0], tw.Prng[1])
	defer func() { tw.Prng = [2]uint32{p.Srand0(), p.Srand1()} }()
	waterAbove := func(threshold int) bool {
		tx, ty := int(tw.X)/tileSize, int(tw.Y)/tileSize
		return m.countSurroundingWaterTiles(tx+6, ty) > threshold || m.countSurroundingWaterTiles(tx, ty+6) > threshold ||
			m.countSurroundingWaterTiles(tx-6, ty) > threshold || m.countSurroundingWaterTiles(tx, ty-6) > threshold
	}
	for attempts := 400; attempts > 0; attempts-- {
		name, flags := tn.NameFromRand(p.RandNext())
		if name == "" || len(name) > kUserStringSize || MediumBoldStringWidth(g1, name) > 200 {
			continue
		}
		if flags&objects.TownNameAdjacentToLargeWater != 0 && !waterAbove(65) {
			continue
		}
		if flags&objects.TownNameNotMountainous != 0 {
			if land, _ := m.GetHeight(int32(tw.X)+tileSize/2, int32(tw.Y)+tileSize/2); land < 192 {
				continue
			}
		}
		if flags&objects.TownNameAdjacentToSmallWater != 0 && !waterAbove(15) {
			continue
		}
		inUse := false
		for j := range m.Towns {
			if m.TownName(m.Towns[j].Name, j) == name {
				inUse = true
				break
			}
		}
		if inUse {
			continue
		}
		id := m.userStringAllocate(name, true)
		if id == 0 {
			continue
		}
		tw.Name = id
		m.UpdateTownLabel(i, g1) // town->updateLabel()
		return true
	}
	return false
}

const kMaxTowns = 80

// ClosestTownAndDensity is TownManager::getClosestTownAndDensity: the town
// nearest loc by Manhattan distance (the first in id order on a tie), and how
// built-up loc is around it (0–3) by the real distance and the town's
// building count. ok false without towns.
func (m *Map) ClosestTownAndDensity(x, y int32) (idx int, density uint8, ok bool) {
	closest, idx := int32(0xFFFF), -1
	for i := range m.Towns {
		d := olmath.ManhattanDistance2D(olmath.Vector2i{X: int32(m.Towns[i].X), Y: int32(m.Towns[i].Y)}, olmath.Vector2i{X: x, Y: y})
		if d < closest {
			closest, idx = d, i
		}
	}
	if closest == 0xFFFF || idx < 0 {
		return -1, 0, false
	}
	tw := &m.Towns[idx]
	real := int32(olmath.Distance2D(olmath.Vector2i{X: int32(tw.X), Y: int32(tw.Y)}, olmath.Vector2i{X: x, Y: y}))
	unk := min(max((real-int32(tw.NumBuildings)*4+512)/128, 0), 4)
	return idx, uint8(min(4-unk, 3)), true
}

// updateTownInfo is TownManager::updateTownInfo: the closest town gains
// population (uint32: a negative amount wraps, as upstream subtracts),
// capacity and buildings, and — for a company other than the neutral one,
// outside the editor — the rating change (Town::adjustCompanyRating).
func (m *Map) updateTownInfo(x, y int32, population, capacity uint32, rating, numBuildings int16, c *Cmd) int {
	i, _, ok := m.ClosestTownAndDensity(x, y)
	if !ok {
		return -1
	}
	tw := &m.Towns[i]
	tw.PopulationCapacity += capacity
	tw.Population += population
	if rating != 0 && c != nil && c.Company != CompanyNeutral && !c.Editor {
		m.adjustCompanyRating(i, c.Company, int(rating))
	}
	if int32(tw.NumBuildings)+int32(numBuildings) <= 32767 {
		tw.NumBuildings += numBuildings
	}
	return i
}

// adjustCompanyRating is Town::adjustCompanyRating: the company's rating
// moved and clamped to −1000…1000, its bit set in companiesWithRating.
func (m *Map) adjustCompanyRating(i int, company uint8, amount int) {
	tw := &m.Towns[i]
	tw.CompaniesWithRating |= 1 << company
	tw.CompanyRatings[company] = int16(min(max(int(tw.CompanyRatings[company])+amount, -1000), 1000))
}

// ResetBuildingsInfluence is TownManager::resetBuildingsInfluence: every
// town's buildings, population, capacity and amenities recounted from the
// map's buildings (first tile of each; no ghosts, no misc buildings), each
// counted for its closest town.
func (m *Map) ResetBuildingsInfluence(om *objects.ObjectManager) {
	for i := range m.Towns {
		tw := &m.Towns[i]
		tw.NumBuildings, tw.Population, tw.PopulationCapacity = 0, 0, 0
		tw.AmenityCounts = [8]uint8{}
	}
	for ty := 0; ty < m.Height; ty++ {
		for tx := 0; tx < m.Width; tx++ {
			t := m.TileAt(tx, ty)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindBuilding || r.Flags&0x10 != 0 {
					continue
				}
				b := t.Buildings[r.Index]
				if b.IsMisc || b.SequenceIndex != 0 {
					continue
				}
				obj := om.GetBuildingObjectByIndex(int(b.ObjectID))
				if obj == nil {
					continue
				}
				produced := uint32(obj.ProducedQuantity[0])
				pop := uint32(0)
				if b.IsConstructed {
					pop = produced
				}
				if i := m.updateTownInfo(int32(tx*tileSize), int32(ty*tileSize), pop, produced, 0, 1, nil); i >= 0 && obj.TownAmenityCategory != 0xFF {
					m.Towns[i].AmenityCounts[obj.TownAmenityCategory&7]++
				}
			}
		}
	}
}

// InitialiseTown is TownManager::initialiseTown: a new town record in the
// first empty slot at (x, y) — a hamlet with no people, ratings of 500, the
// game rng copied as its own prng, its cargo influence flags and a generated
// name — and, if any other town already has buildings or people, the
// building influence recounted. Returns its index in m.Towns, or −1 (no slot,
// or no name found: the slot stays empty).
func (m *Map) InitialiseTown(x, y int16, g1 G1Raw, om *objects.ObjectManager) int {
	used := map[uint16]bool{}
	for _, tw := range m.Towns {
		used[tw.ID] = true
	}
	id := -1
	for i := range kMaxTowns {
		if !used[uint16(i)] {
			id = i
			break
		}
	}
	if id < 0 {
		return -1
	}
	tw := scenario.Town{ID: uint16(id), X: x, Y: y, HistorySize: 1, BuildSpeed: 1, Prng: [2]uint32{m.Rng1.Srand0(), m.Rng1.Srand1()}}
	for c := range tw.CompanyRatings {
		tw.CompanyRatings[c] = 500
	}
	tw.CargoInfluenceFlags = m.CalcCargoInfluenceFlags(x, y, om)
	// towns() runs in id order: keep m.Towns sorted by ID.
	at := len(m.Towns)
	for i := range m.Towns {
		if m.Towns[i].ID > tw.ID {
			at = i
			break
		}
	}
	m.Towns = append(m.Towns[:at], append([]scenario.Town{tw}, m.Towns[at:]...)...)
	if !m.GenerateTownName(at, g1, om) {
		m.Towns = append(m.Towns[:at], m.Towns[at+1:]...)
		return -1
	}
	for i := range m.Towns {
		o := &m.Towns[i]
		if o.NumBuildings == 0 && o.Population == 0 && o.PopulationCapacity == 0 {
			continue
		}
		m.ResetBuildingsInfluence(om)
		break
	}
	return at
}
