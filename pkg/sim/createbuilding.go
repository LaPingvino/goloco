package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// BuildingPlacement is BuildingPlacementArgs: where (world units, z the
// base height), rotation, the building object, its variation, colour, and
// whether it stands built at once.
type BuildingPlacement struct {
	X, Y, Z          int32
	Rotation         uint8
	Type             uint8
	Variation        uint8
	Colour           uint8
	BuildImmediately bool
}

// buildingClearHeight: the variation's parts' heights, or — not built at
// once, with scaffolding — whole scaffolding segments plus its roof; rounded
// up to 4.
func buildingClearHeight(obj *objects.BuildingObject, variation uint8, buildImmediately bool, om *objects.ObjectManager) int {
	h := 0
	for _, part := range obj.GetBuildingParts(variation) {
		if int(part) < len(obj.PartHeights) {
			h += int(obj.PartHeights[part])
		}
	}
	if !buildImmediately && obj.ScaffoldingSegmentType != 0xFF && om.Scaffolding != nil && obj.ScaffoldingSegmentType < 3 {
		if seg := int(om.Scaffolding.SegmentHeights[obj.ScaffoldingSegmentType]); seg > 0 {
			n := h / seg
			if h%seg != 0 {
				n++
			}
			h = n*seg + int(om.Scaffolding.RoofHeights[obj.ScaffoldingSegmentType])
		}
	}
	return (h + 3) &^ 3
}

// CreateBuilding is the createBuilding game command (CreateBuilding.cpp
// 0x0042D133): a town building needs a town (town_must_be_built_first); on
// each tile of the footprint (inside the map) — applying a real one, the walls
// below its top go —, no water (cant_build_this_underwater), trees and (unless
// preventBuildingClearing) buildings cleared at standard height from the lower
// of surface and target, anything else colliding; a sloped or other-height
// surface costs its land and, not a ghost, is flattened (refused when a
// non-bridged track / road, an airport, a building, an industry or a tree
// above the top stands over a surface that would rise); applying, the surface
// industry and terrain style go (not a ghost), the building element goes in
// (all quarters; its elevator animation 5 on the first tile when built at
// once, the part animation 6 when a part has frames). A real town building
// then adds its population (when built) and capacity, one building and its
// amenity to the closest town. The cost.
// PARTIAL: checkFreeElementsAndReorganise; madeAnyChanges; invalidation.
func (m *Map) CreateBuilding(c *Cmd, flags uint8, a BuildingPlacement) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	if c.Objects == nil {
		return 0, false
	}
	obj := c.Objects.GetBuildingObjectByIndex(int(a.Type))
	if obj == nil {
		return 0, false
	}
	misc := obj.Flags&objects.BuildingFlagMiscBuilding != 0
	if !misc {
		if _, _, ok := m.ClosestTownAndDensity(a.X, a.Y); !ok {
			c.ErrorText = i18n.TownMustBeBuiltFirst
			return 0, false
		}
	}
	clearHeight := buildingClearHeight(obj, a.Variation, a.BuildImmediately, c.Objects)
	removed := RemovedBuildings{}
	var total int32
	ghost := flags&CmdGhost != 0
	for seq, off := range buildingTileOffsets(obj.Flags&objects.BuildingFlagLargeTile != 0) {
		px, py := a.X+off[0], a.Y+off[1]
		tx, ty := int(px)>>5, int(py)>>5
		if px < 0 || py < 0 || tx >= m.Width || ty >= m.Height {
			return 0, false
		}
		px, py = int32(tx)*tileSize, int32(ty)*tileSize
		t := m.TileAt(tx, ty)
		if flags&CmdApply != 0 && !ghost {
			t.removeAllWallsBelow(uint8((a.Z + int32(clearHeight)) / smallZStep))
		}
		if t.WaterLevel != 0 {
			c.ErrorText = i18n.CantBuildThisUnderwater
			return 0, false
		}
		baseZ := min(t.BaseZ, uint8(a.Z/smallZStep))
		clearZ := uint8((a.Z + int32(clearHeight)) / smallZStep)
		clear := func(tt *Tile, k int) ClearResult {
			switch tt.Order[k].Kind {
			case scenario.KindTree:
				return m.clearTreeCollision(c, tt, k, px, py, flags, &total)
			case scenario.KindBuilding:
				if flags&CmdPreventBuildingClearing != 0 {
					return ClearCollision
				}
				return m.clearBuildingCollision(c, tt, k, px, py, removed, flags, &total)
			}
			return ClearCollision
		}
		if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ, clearZ, 0xF, 0xF, clear); !ok {
			return 0, false
		}
		t = m.TileAt(tx, ty)
		uneven := t.Slope&0x1F != 0 || int32(t.BaseZ)*smallZStep != a.Z
		if uneven && int(t.TerrainIndex) < len(c.Objects.LandObjects) && c.Objects.LandObjects[t.TerrainIndex] != nil {
			lo := c.Objects.LandObjects[t.TerrainIndex]
			total += m.InflationAdjustedCost(lo.CostFactor, lo.CostIndex, 10)
		}
		if !ghost && uneven {
			if a.Z < int32(t.BaseZ)*smallZStep {
				passed := false
				for _, r := range t.Order {
					if r.Kind == scenario.KindSurface {
						passed = true
						continue
					}
					if !passed {
						continue
					}
					g := r.Flags&elemFlagGhost != 0
					switch r.Kind {
					case scenario.KindTrack:
						if !g && !t.Tracks[r.Index].HasBridge {
							return 0, false
						}
					case scenario.KindRoad:
						if !g && !t.Roads[r.Index].HasBridge {
							return 0, false
						}
					case scenario.KindStation:
						if t.Stations[r.Index].StationType == 2 { // airport
							return 0, false
						}
					case scenario.KindBuilding, scenario.KindIndustry:
						return 0, false
					case scenario.KindTree:
						if a.Z+int32(clearHeight) <= int32(t.Trees[r.Index].BaseZ)*smallZStep {
							return 0, false
						}
					}
				}
			}
			if flags&CmdApply != 0 {
				m.mapInvalidateTileFull(px, py)
				t.BaseZ = uint8(a.Z / smallZStep) // setBaseZ / setClearZ (the sim tile keeps one height)
				t.Slope = 0                       // the slope and the snow coverage (its top bits)
				t.GrowthStage = 0
			}
		}
		if flags&CmdApply == 0 {
			continue
		}
		if !ghost {
			m.removeSurfaceIndustry(px, py, t, c)
			t.setTerrainStyleAsCleared()
		}
		bz := uint8(a.Z / smallZStep)
		t.Buildings = append(t.Buildings, scenario.BuildingElement{
			ObjectID: a.Type, Rotation: a.Rotation & 3, SequenceIndex: uint8(seq), Variation: a.Variation, Colour: a.Colour,
			IsConstructed: a.BuildImmediately, IsMisc: misc, BaseZ: bz, ClearZ: uint8(clearHeight/smallZStep) + bz,
		})
		t.insertRef(scenario.KindBuilding, len(t.Buildings)-1, bz)
		for i := range t.Order {
			if r := &t.Order[i]; r.Kind == scenario.KindBuilding && int(r.Index) == len(t.Buildings)-1 {
				r.Flags |= 0xF // insertElement(…, 0xF): every quarter occupied
				if ghost {
					r.Flags |= elemFlagGhost
				}
			}
		}
		if a.BuildImmediately && seq == 0 && obj.NumElevatorSequences != 0 {
			m.CreateAnimation(5, px, py, bz)
		}
		for _, part := range obj.GetBuildingParts(a.Variation) {
			if int(part) < len(obj.PartAnimations) && obj.PartAnimations[part].NumFrames > 1 {
				m.CreateAnimation(6, px, py, bz)
				break
			}
		}
	}
	if flags&CmdApply != 0 && !ghost && !misc {
		capacity := uint32(obj.ProducedQuantity[0])
		pop := uint32(0)
		if a.BuildImmediately {
			pop = capacity
		}
		if town := m.updateTownInfo(a.X, a.Y, pop, capacity, 0, 1, c); town >= 0 && town < len(m.Towns) && obj.TownAmenityCategory != 0xFF && int(obj.TownAmenityCategory) < len(m.Towns[town].AmenityCounts) {
			m.Towns[town].AmenityCounts[obj.TownAmenityCategory]++
		}
	}
	return total, true
}
