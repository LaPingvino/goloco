package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Walls placed (for the industries' fences, CR): the createWall game
// command (GameCommands/Terraform/CreateWall.cpp 0x004C436C) with
// canConstructWall (0x004C4A3B) and edgeWallMapping. PARTIAL: the tertiary
// colour (unused in Locomotion — only two remaps work), the viewport
// invalidation, checkFreeElementsAndReorganise.

// WallPlacement is WallPlacementArgs: the tile (z 0: on the ground, the
// edge's slope deciding), its edge (rotation), the wall object and colours.
type WallPlacement struct {
	X, Y, Z            int32
	Rotation, Type     uint8
	Primary, Secondary uint8
}

// Edge slopes of a wall (EdgeSlope): along a sloped surface edge.
const (
	edgeSlopeUpwards   uint8 = 1
	edgeSlopeDownwards uint8 = 2
)

// edgeWallMapping: per surface slope (32) and edge (4), the wall's edge slope
// and whether it stands 16 higher — upstream's table (none 0, up 1, down 2;
// +4: elevated).
var edgeWallMapping = [32][4]uint8{
	{0, 0, 0, 0}, {0, 1, 2, 0}, {0, 0, 1, 2}, {0, 1, 4, 2}, {2, 0, 0, 1}, {2, 1, 2, 1}, {2, 0, 1, 4}, {2, 1, 4, 4},
	{1, 2, 0, 0}, {1, 4, 2, 0}, {1, 2, 1, 2}, {1, 4, 4, 2}, {4, 2, 0, 1}, {4, 4, 2, 1}, {4, 2, 1, 4}, {4, 4, 4, 4},
	{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {2, 1, 5, 6},
	{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {1, 5, 6, 2}, {0, 0, 0, 0}, {5, 6, 2, 1}, {6, 2, 1, 5}, {0, 0, 0, 0},
}

// canConstructWall: on the drawable map, nothing (but the surface, trees and
// what occupies no quarter) overlapping baseZ–clearZ, and no wall on the same
// edge there.
func (m *Map) canConstructWall(x, y int32, baseZ, clearZ, edge uint8) bool {
	tx, ty := int(x)>>5, int(y)>>5
	if tx < 1 || ty < 1 || tx > m.Width-2 || ty > m.Height-2 { // drawableCoords
		return false
	}
	t := m.TileAt(tx, ty)
	for _, r := range t.Order {
		if r.Kind == scenario.KindSurface || baseZ >= t.clearZOf(r) || clearZ <= t.baseZOf(r) {
			continue
		}
		if r.Kind == scenario.KindWall {
			if t.Walls[r.Index].Rotation == edge {
				return false
			}
			continue
		}
		if r.Flags&0x0F == 0 || r.Kind == scenario.KindTree {
			continue
		}
		return false
	}
	return true
}

// CreateWall is the createWall game command: on the ground (z 0) its height
// and edge slope from the surface's slope on that edge (16 higher where
// elevated); refused under water, below ground, partly underground on a
// level wall over a sloped corner, a sloped wall of a level-only object, or
// where canConstructWall says no; with CmdApply the wall element placed (a
// ghost when asked).
func (m *Map) CreateWall(c *Cmd, flags uint8, a WallPlacement) (int32, bool) {
	tx, ty := int(a.X)>>5, int(a.Y)>>5
	if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height || c.Objects == nil || int(a.Type) >= len(c.Objects.WallObjects) || c.Objects.WallObjects[a.Type] == nil {
		return 0, false
	}
	t := m.TileAt(tx, ty)
	var slope uint8
	target := a.Z
	if target == 0 {
		target = int32(t.BaseZ) * 4
		e := edgeWallMapping[t.Slope&0x1F][a.Rotation&3]
		slope = e & 3
		if e&4 != 0 {
			target += 16
		}
	}
	baseZ := uint8(target / 4)
	if t.WaterLevel != 0 && baseZ < t.WaterLevel*4 {
		return 0, false // cant_build_this_underwater
	}
	if baseZ < t.BaseZ {
		return 0, false // error_can_only_build_above_ground
	}
	if slope&(edgeSlopeUpwards|edgeSlopeDownwards) == 0 {
		test := t.BaseZ + 4
		for i := uint8(2); i <= 3; i++ {
			e := (a.Rotation + i) & 3
			if t.Slope&(1<<e) == 0 {
				continue
			}
			if baseZ < test {
				return 0, false
			}
			if t.Slope&0x10 != 0 { // isSlopeDoubleHeight
				e = (e - 1) & 3
				if t.Slope&(1<<e) != 0 {
					e = (e + 2) & 3
					if t.Slope&(1<<e) != 0 && baseZ < test+4 {
						return 0, false
					}
				}
			}
		}
	}
	w := c.Objects.WallObjects[a.Type]
	clearZ := baseZ
	if slope&(edgeSlopeUpwards|edgeSlopeDownwards) != 0 {
		if w.Flags&objects.WallFlagOnlyOnLevelLand != 0 {
			return 0, false // can_only_build_this_on_level_land
		}
		clearZ += 4
	}
	clearZ += w.Height
	if !m.canConstructWall(a.X, a.Y, baseZ, clearZ, a.Rotation) {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	t.Walls = append(t.Walls, scenario.WallElement{WallObjectID: a.Type, Rotation: a.Rotation, EdgeSlope: slope,
		PrimaryColour: a.Primary, SecondaryColour: a.Secondary, BaseZ: baseZ, ClearZ: clearZ})
	t.insertRef(scenario.KindWall, len(t.Walls)-1, baseZ)
	if flags&CmdGhost != 0 {
		for i := range t.Order {
			if r := &t.Order[i]; r.Kind == scenario.KindWall && int(r.Index) == len(t.Walls)-1 {
				r.Flags |= elemFlagGhost
			}
		}
	}
	return 0, true
}

// isSurfaceClaimed is Industry.cpp's (0x0045510C, bl 0): a valid tile with a
// dry surface no ship needs, and above it nothing but ghosts, walls and trees.
func (m *Map) isSurfaceClaimed(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height {
		return false
	}
	t := m.TileAt(tx, ty)
	passed := false
	for _, r := range t.Order {
		if r.Kind == scenario.KindSurface {
			if t.WaterLevel != 0 || t.Type6Flag {
				return false
			}
			passed = true
			continue
		}
		if !passed || r.Flags&elemFlagGhost != 0 || r.Kind == scenario.KindWall || r.Kind == scenario.KindTree {
			continue
		}
		return false
	}
	return true
}

// claimSurfaceForIndustry is claimSurfaceForIndustry (0x0045510C, bl 1): a
// claimable surface becomes the industry's field — industrial, its id, update
// timer and growth stage —, the walls above it removed.
func (m *Map) claimSurfaceForIndustry(tx, ty int, id, growthStage, updateTimer uint8) bool {
	if !m.isSurfaceClaimed(tx, ty) {
		return false
	}
	t := m.TileAt(tx, ty)
	t.IsIndustrial, t.Variation, t.GrowthStage = true, id, growthStage // setIndustry: byte 7
	t.UpdateTimer = updateTimer
	t.removeAllWallsAbove(t.BaseZ)
	return true
}
