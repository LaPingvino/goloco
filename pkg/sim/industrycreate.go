package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Creating an industry (item CR part 2c): GameCommands/Industries/
// CreateIndustry.cpp's getProduced/RequiredCargoBitSet, sub_454C91
// (0x00454C91), getRandomAvailableColour and placeIndustryBuilding, and
// IndustryManager::allocateNewIndustry; the createIndustry command itself
// (CreateIndustry below: its building loop, the claimed fields and walls,
// expandGrounds). PARTIAL: checkFreeElementsAndReorganise (goloco's tiles have no element limit); the name:
// upstream takes the object's var_02 (a StringId given at load) and, when
// another industry of the town has it, a numbered user string — goloco has
// neither string ids nor user strings, so Name only marks the slot used.

const maxIndustries = 128 // Limits::kMaxIndustries

// industryCargoBitSets are getProducedCargoBitSet / getRequiredCargoBitSet:
// a bit per cargo type the object makes / needs.
func industryCargoBitSets(obj *objects.IndustryObject, om *objects.ObjectManager) (produced, required uint32) {
	p, r := om.IndustryCargoIDs(obj)
	for _, c := range p {
		if c != 0xFF {
			produced |= 1 << c
		}
	}
	for _, c := range r {
		if c != 0xFF {
			required |= 1 << c
		}
	}
	return
}

// industryAllocateChecked is sub_454C91: refused when an industry is closer
// than 9 tiles (24 when one feeds the other: a cargo one makes the other
// needs), or no town is near; else allocateNewIndustry with the closest
// town. ok false with the reason upstream sets as the error.
func (m *Map) industryAllocateChecked(objID uint8, x, y int32, prng [2]uint32, om *objects.ObjectManager) (id uint8, reason string, ok bool) {
	obj := om.GetIndustryObjectByIndex(int(objID))
	if obj == nil {
		return 0, "no such industry object", false
	}
	prod, req := industryCargoBitSets(obj, om)
	for i := range m.Industries {
		ind := &m.Industries[i]
		o2 := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if o2 == nil {
			continue
		}
		p2, r2 := industryCargoBitSets(o2, om)
		minDist := int32(32 * 9)
		if req&p2 != 0 || prod&r2 != 0 {
			minDist = 32 * 24
		}
		if abs32(int32(ind.X)-x)+abs32(int32(ind.Y)-y) < minDist {
			return 0, "too_close_to_another_industry", false
		}
	}
	ti, _, found := m.ClosestTownAndDensity(x, y)
	if !found {
		return 0, "town_must_be_built_nearby_first", false
	}
	id, ok = m.allocateNewIndustry(objID, x, y, prng, m.Towns[ti].ID)
	if !ok {
		return 0, "too_many_industries", false
	}
	return id, "", true
}

// allocateNewIndustry is IndustryManager::allocateNewIndustry: the first
// free of the 128 slots made a new, empty industry of the type at (x, y) —
// its prng, under construction (0), production rate 25, founded this year,
// no stations, the histories one entry of 0 —, for the town; kept in slot
// (id) order with the others. ok false: no free slot.
func (m *Map) allocateNewIndustry(objID uint8, x, y int32, prng [2]uint32, town uint16) (uint8, bool) {
	used := map[uint8]bool{}
	for i := range m.Industries {
		used[m.Industries[i].ID] = true
	}
	for id := 0; id < maxIndustries; id++ {
		if used[uint8(id)] {
			continue
		}
		ind := scenario.Industry{ID: uint8(id), Name: 0, ObjectID: objID, X: int16(x), Y: int16(y), Prng: prng,
			UnderConstruction: 0, ProductionRate: 25, FoundingYear: m.CurrentYear, Town: town,
			HistorySize: [2]uint8{1, 1}}
		for c := range ind.ProducedStations {
			for k := range ind.ProducedStations[c] {
				ind.ProducedStations[c][k] = stationIDNull
			}
		}
		at := len(m.Industries)
		for i := range m.Industries {
			if m.Industries[i].ID > uint8(id) {
				at = i
				break
			}
		}
		m.Industries = append(m.Industries, scenario.Industry{})
		copy(m.Industries[at+1:], m.Industries[at:])
		m.Industries[at] = ind
		return uint8(id), true
	}
	return 0, false
}

// randomAvailableColour is getRandomAvailableColour (0x0045442A): one of the
// object's colours by the industry's prng — drawn even when there is only
// one (as upstream: a draw fewer would diverge) —; black without any.
func randomAvailableColour(p *core.Prng, obj *objects.IndustryObject) uint8 {
	var colours []uint8
	for c := 0; c < 32; c++ {
		if obj.AvailableColours&(1<<c) != 0 {
			colours = append(colours, uint8(c))
		}
	}
	if len(colours) == 0 {
		return 0 // Colour::black
	}
	return colours[p.RandNextBound(int32(len(colours))-1)]
}

// placeIndustryBuilding is CreateIndustry.cpp's (0x004551CC): one building
// of the industry at pos — on its 1 or 4 tiles, at the footprint's highest
// corner (or water) height, as tall as its parts (with scaffolding while
// built: whole segments and the roof), rounded up to 4 —; each tile cleared
// of trees (anything else collides), on water only for an object built on
// water (not a ship channel), else off water, the ground levelled (paid by
// the land object, and refused where track, road, an airport, a building, an
// industry or a tree it would undercut stands above), the industry element
// placed (its animation made, a ghost when asked) — with CmdApply; then the
// tile added to the industry's list. ok false: refused.
func (m *Map) placeIndustryBuilding(c *Cmd, id uint8, x, y int32, direction, buildingType, colour uint8, buildImmediate bool, flags uint8) (int32, bool) {
	ind := m.industryByID(id)
	if ind == nil || c.Objects == nil {
		return 0, false
	}
	obj := c.Objects.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return 0, false
	}
	large := obj.BuildingSizeFlags&(1<<buildingType) != 0
	footprint := buildingTileOffsets(large)
	var highest uint8
	for _, off := range footprint {
		t := m.TileAt(int(x+off[0])>>5, int(y+off[1])>>5)
		if t == nil {
			continue
		}
		highest = max(highest, surfaceCornerHeight(t))
		if t.WaterLevel != 0 {
			highest = max(highest, t.WaterLevel*4) // kMicroToSmallZStep
		}
	}
	clearHeight := 0
	if int(buildingType) < len(obj.Variations) {
		for _, part := range obj.Variations[buildingType] {
			if int(part) < len(obj.PartHeights) {
				clearHeight += int(obj.PartHeights[part])
			}
		}
	}
	if !buildImmediate && obj.ScaffoldingSegmentType != 0xFF && c.Objects.Scaffolding != nil && obj.ScaffoldingSegmentType < 3 {
		seg := int(c.Objects.Scaffolding.SegmentHeights[obj.ScaffoldingSegmentType])
		if seg > 0 {
			n := clearHeight / seg
			if clearHeight%seg != 0 {
				n++
			}
			clearHeight = n*seg + int(c.Objects.Scaffolding.RoofHeights[obj.ScaffoldingSegmentType])
		}
	}
	clearHeight = (clearHeight + 3) &^ 3
	clearZ := uint8(clearHeight/4) + highest
	var cost int32
	for seq, off := range footprint {
		tx, ty := int(x+off[0])>>5, int(y+off[1])>>5
		if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height {
			return 0, false
		}
		t := m.TileAt(tx, ty)
		px, py := int32(tx)*32, int32(ty)*32
		if flags&CmdApply != 0 && flags&CmdGhost == 0 {
			t.removeAllWallsBelow(highest + uint8(clearHeight/4))
		}
		clear := func(t *Tile, k int) ClearResult { // tileClearFunction: trees cleared, anything else collides
			if t.Order[k].Kind != scenario.KindTree {
				return ClearCollision
			}
			return m.clearTreeCollision(c, t, k, px, py, flags, &cost)
		}
		if obj.Flags&industryBuiltOnWater != 0 {
			if t.WaterLevel*4 != highest || t.Type6Flag {
				return 0, false // can_only_be_built_on_water / water_channel_currently_needed_by_ships
			}
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, highest, clearZ, 0xF, 0xF, clear); !ok {
				return 0, false
			}
		} else {
			if t.WaterLevel != 0 {
				return 0, false // cant_build_this_underwater
			}
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, t.BaseZ, clearZ, 0xF, 0xF, clear); !ok {
				return 0, false
			}
			t = m.TileAt(tx, ty)
			if t.Slope&0x1F != 0 || t.BaseZ != highest {
				if int(t.TerrainIndex) < len(c.Objects.LandObjects) && c.Objects.LandObjects[t.TerrainIndex] != nil {
					lo := c.Objects.LandObjects[t.TerrainIndex]
					cost += m.InflationAdjustedCost(lo.CostFactor, lo.CostIndex, 10)
				}
			}
		}
		if flags&CmdGhost == 0 && obj.Flags&industryBuiltOnWater == 0 && (t.Slope&0x1F != 0 || highest != t.BaseZ) {
			if highest < t.BaseZ {
				passed := false
				for _, r := range t.Order {
					if r.Kind == scenario.KindSurface {
						passed = true
						continue
					}
					if !passed {
						continue
					}
					ghost := r.Flags&0x10 != 0
					switch r.Kind {
					case scenario.KindTrack:
						if !ghost && !t.Tracks[r.Index].HasBridge {
							return 0, false
						}
					case scenario.KindRoad:
						if !ghost && !t.Roads[r.Index].HasBridge {
							return 0, false
						}
					case scenario.KindStation:
						if t.Stations[r.Index].StationType == 2 {
							return 0, false
						}
					case scenario.KindBuilding, scenario.KindIndustry:
						return 0, false
					case scenario.KindTree:
						if clearZ <= t.Trees[r.Index].BaseZ {
							return 0, false
						}
					}
				}
			}
			if flags&CmdApply != 0 {
				t.BaseZ = highest // setBaseZ / setClearZ (the sim tile keeps one height)
				t.Slope = 0       // the slope and the snow coverage (its top bits)
				t.GrowthStage = 0
			}
		}
		if flags&CmdApply != 0 {
			if flags&CmdGhost == 0 {
				m.removeSurfaceIndustry(px, py, t, c)
				t.setTerrainStyleAsCleared()
			}
			t.Industries = append(t.Industries, scenario.IndustryElement{IndustryID: id, BuildingType: buildingType, Colour: colour,
				SequenceIndex: uint8(seq), Rotation: direction, IsConstructed: buildImmediate, BaseZ: highest, ClearZ: clearZ})
			t.insertRef(scenario.KindIndustry, len(t.Industries)-1, highest)
			m.invalidateTile(px&^31, py&^31, int32(highest)*4, int32(clearZ)*4, ZoomEighth, defaultRadius) // ViewportManager::invalidate
			for i := range t.Order {                                                                       // insertElement<IndustryElement>(…, 0xF): every quarter occupied
				if r := &t.Order[i]; r.Kind == scenario.KindIndustry && int(r.Index) == len(t.Industries)-1 {
					r.Flags |= 0xF
					if flags&CmdGhost != 0 {
						r.Flags |= elemFlagGhost
					}
				}
			}
			m.CreateAnimation(3, px, py, highest)
		}
	}
	z := int16(highest) * 4
	if large {
		z |= -1 << 15 // bit 15: a 2×2 building
	}
	if ind = m.industryByID(id); ind != nil {
		ind.Tiles = append(ind.Tiles, [3]int16{int16(x), int16(y), z})
	}
	return cost, true
}

// IndustryPlacement is IndustryPlacementArgs: the industry object, where,
// whether it is built at once, and the command's seed (srand1; upstream
// takes srand0 as the command's number — a vanilla bug it keeps).
type IndustryPlacement struct {
	Type             uint8
	X, Y             int32
	BuildImmediately bool
	Srand0, Srand1   uint32
}

const (
	gameCommandCreateIndustry = 47      // GameCommand::createIndustry
	industryFlag04            = 1 << 3  // IndustryFlags::flag_04: founded by a company
	industryNotRotatable      = 1 << 15 // IndustryObjectFlags::notRotatable
)

// placementAttempts is kPlacementAttemptDistanceRangeAndMin: as a building's
// tries go on, the tiles searched around the last building widen.
var placementAttempts = [25][2]int32{{1, 0}, {3, -1}, {3, -1}, {3, -1}, {3, -1}, {3, -1}, {3, -1}, {3, -1},
	{5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2},
	{5, -2}, {5, -2}, {5, -2}, {5, -2}, {5, -2}}

// CreateIndustry is the createIndustry game command (0x0045436B): an
// industry allocated (sub_454C91), built at once or under construction, a
// ghost when asked, a company's when one founds it; its colour; its two
// production targets from the object's initial ranges (the second from the
// first's low byte, as upstream); between min and max of its buildings, each
// tried up to 25 times at a random tile — widening around the last one
// placed — not overlapping a placed one, a random direction (unless not
// rotatable), a query then the placement; its centre the average of its
// tiles; its grounds (industryGrounds: the claimed fields, perimeter walls,
// expandGrounds); the cost. The industry is freed again when a building can't
// be placed, or on a query. A real one not built at once posts newIndustry.
func (m *Map) CreateIndustry(c *Cmd, flags uint8, a IndustryPlacement) (int32, uint8, bool) {
	prng := [2]uint32{gameCommandCreateIndustry, a.Srand1}
	id, _, ok := m.industryAllocateChecked(a.Type, a.X, a.Y, prng, c.Objects)
	if !ok {
		return 0, 0, false
	}
	free := func() { // the slot freed (its name null)
		for i := range m.Industries {
			if m.Industries[i].ID == id {
				m.Industries = append(m.Industries[:i], m.Industries[i+1:]...)
				return
			}
		}
	}
	ind := m.industryByID(id)
	obj := c.Objects.GetIndustryObjectByIndex(int(a.Type))
	if a.BuildImmediately {
		ind.UnderConstruction = 0xFF
	}
	if flags&CmdGhost != 0 {
		ind.Flags |= industryFlagIsGhost
	}
	if c.Company != CompanyNeutral {
		ind.Flags |= industryFlag04
		ind.Owner = c.Company
	}
	p := core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
	sync := func() { ind.Prng = [2]uint32{p.Srand0(), p.Srand1()} }
	colour := randomAvailableColour(p, obj)
	lastX, lastY := int32(ind.X), int32(ind.Y)
	randVal := p.RandNext() & 0xFF
	rate := randVal
	for i := 0; i < 2; i++ {
		ind.DailyProduction[i] = 0
		lo, hi := uint32(obj.InitialProductionRate[i][0]), uint32(obj.InitialProductionRate[i][1])
		ind.DailyProductionTarget[i] = uint16((hi-lo)*rate/256 + lo)
		rate = uint32(ind.DailyProductionTarget[i] & 0xFF)
	}
	var total int32
	n := (uint32(obj.MaxNumBuildings)-uint32(obj.MinNumBuildings)+1)*randVal/256 + uint32(obj.MinNumBuildings)
	for i := uint32(0); i < n && int(i) < len(obj.Buildings); i++ {
		b := obj.Buildings[i]
		large := obj.BuildingSizeFlags&(1<<b) != 0
		placed := false
		for j := 0; j < 25 && !placed; j++ {
			r := p.RandNext()
			sync()
			dir := uint8(r & 3)
			if obj.Flags&industryNotRotatable != 0 {
				dir = 0
			}
			rng, minD := placementAttempts[j][0], placementAttempts[j][1]
			x := lastX + (rng*int32((r>>18)&0xFF)/256+minD)*32
			y := lastY + (rng*int32((r>>2)&0xFF)/256+minD)*32
			maxX, maxY := x, y
			if large {
				maxX, maxY = x+32, y+32
			}
			overlap := false
			for _, tl := range ind.Tiles {
				tx, ty := int32(tl[0]), int32(tl[1])
				tmx, tmy := tx, ty
				if tl[2]&(-1<<15) != 0 {
					tmx, tmy = tx+32, ty+32
				}
				if x > tmx || y > tmy || maxX < tx || maxY < ty {
					continue
				}
				overlap = true
				break
			}
			if overlap {
				continue
			}
			if flags&CmdApply != 0 { // the query first
				if _, ok := m.placeIndustryBuilding(c, id, x, y, dir, b, colour, a.BuildImmediately, flags&^CmdApply); !ok {
					continue
				}
				ind = m.industryByID(id)
				ind.Tiles = ind.Tiles[:len(ind.Tiles)-1] // numTiles-- (the query added it)
			}
			cost, ok := m.placeIndustryBuilding(c, id, x, y, dir, b, colour, a.BuildImmediately, flags)
			ind = m.industryByID(id)
			if !ok {
				continue
			}
			total += cost
			lastX, lastY, placed = x, y, true
		}
		if !placed {
			free()
			return 0, 0, false
		}
	}
	sync()
	if len(ind.Tiles) != 0 {
		var sx, sy int32
		for _, tl := range ind.Tiles {
			sx, sy = sx+int32(tl[0]), sy+int32(tl[1])
			if tl[2]&(-1<<15) != 0 {
				sx, sy = sx+16, sy+16
			}
		}
		k := int32(len(ind.Tiles))
		ind.X, ind.Y = int16(int32(-32)&(sx/k+16)), int16(int32(-32)&(sy/k+16))
	}
	if flags&CmdApply != 0 && flags&CmdGhost == 0 && len(ind.Tiles) != 0 {
		m.industryGrounds(c, ind, obj, p)
	}
	if flags&CmdApply == 0 {
		free()
	}
	total += m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 3)
	if flags&CmdApply != 0 && flags&CmdGhost == 0 && !a.BuildImmediately {
		m.PostMessage(c.Objects, MsgNewIndustry, 0xFF, uint16(id), 0xFFFF, 0xFFFF)
	}
	return total, id, true
}

const (
	industryFarmTilesDesync          = 1 << 23 // IndustryObjectFlags::farmTilesGrowthStageDesynchronized
	industryFarmTilesPartialCoverage = 1 << 27 // farmTilesPartialCoverage
)

// industryWall places one fence piece (the createWall command, apply).
func (m *Map) industryWall(c *Cmd, tx, ty int, rotation, wallType uint8) {
	m.CreateWall(c, CmdApply, WallPlacement{X: int32(tx) * 32, Y: int32(ty) * 32, Rotation: rotation, Type: wallType})
}

// expandGrounds is Industry::expandGrounds (0x00454A43): when at least 20 of
// the 5×5 tiles around pos can be claimed, all of them become the industry's
// fields (the growth stage and timer per tile from a copy of its prng for a
// desynchronised farm; 7 in 8 skipped for partial coverage — upstream's
// copy too, its vanilla mistake kept) and the 5×5's edges get walls (an
// entrance where the mask says, from the prng's state without a draw).
func (m *Map) expandGrounds(c *Cmd, ind *scenario.Industry, x, y int32, wallType, wallEntrance, growth, timer uint8) {
	obj := c.Objects.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	cx, cy := int(x)>>5, int(y)>>5
	x0, y0, x1, y1 := max(cx-2, 0), max(cy-2, 0), min(cx+2, m.Width-1), min(cy+2, m.Height-1) // getClampedRange
	n := 0
	for ty := y0; ty <= y1; ty++ {
		for tx := x0; tx <= x1; tx++ {
			if m.isSurfaceClaimed(tx, ty) {
				n++
			}
		}
	}
	if n < 20 {
		return
	}
	var desync, partial *core.Prng
	if obj.Flags&industryFarmTilesDesync != 0 {
		desync = core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
	}
	if obj.Flags&industryFarmTilesPartialCoverage != 0 {
		partial = core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
	}
	var mask uint32
	if wallType != 0xFF && wallEntrance != 0xFF {
		mask = 1 << (ind.Prng[0] & 0xF)
		mask |= 1 << ((ind.Prng[0] >> 4) & 0x1F)
	}
	i := uint32(0)
	wall := func(tx, ty int, rot uint8) {
		t := wallType
		if mask&(1<<i) != 0 {
			t = wallEntrance
		}
		i++
		m.industryWall(c, tx, ty, rot, t)
	}
	for ty := y0; ty <= y1; ty++ {
		for tx := x0; tx <= x1; tx++ {
			if desync != nil {
				r := desync.RandNext()
				growth, timer = uint8((r&0xFF)*uint32(obj.FarmTileNumGrowthStages)/256), uint8((r>>8)&7)
			}
			if partial == nil || partial.RandNext()&7 == 0 {
				m.claimSurfaceForIndustry(tx, ty, ind.ID, growth, timer)
			}
			if wallType == 0xFF {
				continue
			}
			if tx == cx-2 {
				wall(tx, ty, 0)
			}
			if tx == cx+2 {
				wall(tx, ty, 2)
			}
			if ty == cy-2 {
				wall(tx, ty, 3)
			}
			if ty == cy+2 {
				wall(tx, ty, 1)
			}
		}
	}
}

// industryGrounds is createIndustry's 0x00454745 (apply, not a ghost): the
// tiles around each building claimed as fields with a perimeter fence
// (entrances where the prng's state says), then the farm's grounds expanded
// a random number of times.
func (m *Map) industryGrounds(c *Cmd, ind *scenario.Industry, obj *objects.IndustryObject, p *core.Prng) {
	if obj.FarmTileGrowthStageNoProduction != 0xFF {
		var mask uint32
		if obj.BuildingWallEntrance != 0xFF {
			r := p.Srand0()
			mask |= 1 << (r & 0xF)
			mask |= 1 << ((r >> 4) & 0xF)
		}
		wall := func(tx, ty int, rot uint8) {
			t := obj.BuildingWall
			if mask&1 != 0 {
				t = obj.BuildingWallEntrance
			}
			mask = mask>>1 | mask<<31 // std::rotr(mask, 1)
			m.industryWall(c, tx, ty, rot, t)
		}
		for _, tl := range ind.Tiles {
			x0, y0 := int(tl[0])>>5-1, int(tl[1])>>5-1
			x1, y1 := x0+2, y0+2
			if tl[2]&(-1<<15) != 0 {
				x1, y1 = x0+3, y0+3
			}
			for ty := y0; ty <= y1; ty++ { // TilePosRangeView: x inner
				for tx := x0; tx <= x1; tx++ {
					m.claimSurfaceForIndustry(tx, ty, ind.ID, obj.FarmTileGrowthStageNoProduction, 0)
					if obj.BuildingWall == 0xFF {
						continue
					}
					if tx == x0 {
						wall(tx, ty, 0)
					}
					if tx == x1 {
						wall(tx, ty, 2)
					}
					if ty == y0 {
						wall(tx, ty, 3)
					}
					if ty == y1 {
						wall(tx, ty, 1)
					}
				}
			}
		}
	}
	if obj.FarmTileNumGrowthStages != 0 {
		n := (int(obj.FarmNumFields)*int(p.RandNextBound(0xFF))/256 + 1) * 4
		for i := 0; i < n; i++ {
			r := p.RandNext()
			ind.Prng = [2]uint32{p.Srand0(), p.Srand1()}
			growth := uint8((r & 0xFF) * uint32(obj.FarmTileNumGrowthStages) / 256)
			timer := uint8((r >> 8) & 7)
			ox, oy := int32((r>>11)&0x1F)-15, int32((r>>16)&0x1F)-15
			second := (r>>21)&1 != 0 && obj.WallTypes[2] != 0xFF
			wt, we := obj.WallTypes[0], obj.WallTypes[1]
			if second {
				wt, we = obj.WallTypes[2], obj.WallTypes[3]
			}
			m.expandGrounds(c, ind, int32(ind.X)+ox*32, int32(ind.Y)+oy*32, wt, we, growth, timer)
		}
	}
	ind.Prng = [2]uint32{p.Srand0(), p.Srand1()}
}

// createNewIndustry is IndustryManager::createNewIndustry (0x0045979C): up
// to numAttempts random places (findRandomNewIndustryLocation), each tried
// with the createIndustry command, its seeds from gPrng1 as vanilla does
// (srand0's difference over one draw, then srand1), until one is built.
func (m *Map) createNewIndustry(objID uint8, buildImmediately bool, numAttempts int, om *objects.ObjectManager) bool {
	for a := 0; a < numAttempts; a++ {
		x, y, ok := m.findRandomNewIndustryLocation(objID, om)
		if !ok {
			continue
		}
		before := m.Rng1.Srand0()
		m.Rng1.RandNext()
		args := IndustryPlacement{Type: objID, X: x, Y: y, BuildImmediately: buildImmediately, Srand0: m.Rng1.Srand0() - before, Srand1: m.Rng1.Srand1()}
		if _, _, ok := m.CreateIndustry(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, args); ok {
			return true
		}
	}
	return false
}

// tryCreateNewIndustriesMonthly is IndustryManager's (0x00459659): unless the
// scenario disallows start-ups, each industry type that may start now and is
// under its cap gets 25 attempts at a new industry.
func (m *Map) tryCreateNewIndustriesMonthly(om *objects.ObjectManager) {
	if m.IndustryManagerFlags&industryManagerDisallowStartUp != 0 {
		return
	}
	for id := 0; id < 16; id++ {
		obj := om.GetIndustryObjectByIndex(id)
		if obj == nil || !m.canIndustryObjBeCreated(obj, om) || m.hasReachedCapOfTypeOfIndustry(uint8(id), om) {
			continue
		}
		m.createNewIndustry(uint8(id), false, 25, om)
	}
}

// DiagNewFarmIndustry (DIAGNOSTIC ONLY: the newindustry diags) builds an
// industry with fields and fences — at once, or under construction — at a
// searched place: its id and centre.
func (m *Map) DiagNewFarmIndustry(om *objects.ObjectManager, immediately bool) (uint8, int32, int32, bool) {
	for k := 0; k < 16; k++ {
		o := om.GetIndustryObjectByIndex(k)
		if o == nil || o.FarmTileGrowthStageNoProduction == 0xFF || o.BuildingWall == 0xFF {
			continue
		}
		for try := 0; try < 40; try++ {
			x, y, ok := m.findRandomNewIndustryLocation(uint8(k), om)
			if !ok {
				continue
			}
			if _, id, ok := m.CreateIndustry(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, IndustryPlacement{Type: uint8(k), X: x, Y: y, BuildImmediately: immediately, Srand1: uint32(try)}); ok {
				ind := m.industryByID(id)
				return id, int32(ind.X), int32(ind.Y), true
			}
		}
	}
	return 0, 0, 0, false
}
