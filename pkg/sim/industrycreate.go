package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Creating an industry (item CR part 2c): GameCommands/Industries/
// CreateIndustry.cpp's getProduced/RequiredCargoBitSet, sub_454C91
// (0x00454C91), getRandomAvailableColour and placeIndustryBuilding, and
// IndustryManager::allocateNewIndustry. PARTIAL: the createIndustry command
// itself (its building loop, the claimed fields and walls, expandGrounds) —
// next; checkFreeElementsAndReorganise (goloco's tiles have no element limit); the name:
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
			if flags&CmdGhost != 0 {
				for i := range t.Order {
					if r := &t.Order[i]; r.Kind == scenario.KindIndustry && int(r.Index) == len(t.Industries)-1 {
						r.Flags |= 0x10
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
