package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// GameCommands/Terraform/ClearLand.cpp and RemoveWall.cpp.

// clearTile (0x00469D76): applying, the surface's terrain style cleared
// (outside the editor) and the walls above the land removed; then
// everything at all heights on the land's level is cleared without the
// default collision (trees and buildings go, the rest stays). The cost.
func (m *Map) clearTile(c *Cmd, px, py int32, removed RemovedBuildings, flags uint8) (int32, bool) {
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	if t == nil || px < 0 || py < 0 {
		return 0, false // off_edge_of_map (getClampedRange keeps it inside)
	}
	if flags&CmdApply != 0 {
		if !c.Editor {
			t.setTerrainStyleAsCleared()
		}
		land, _ := m.GetHeight(px, py)
		t.removeAllWallsAbove(uint8(land / 4))
	}
	var cost int32
	clear := func(tt *Tile, k int) ClearResult {
		return m.tileClearFunction(c, tt, k, px, py, removed, flags, &cost, false)
	}
	land, _ := m.GetHeight(px, py)
	if ok, _ := m.ApplyClearAtAllHeights(px, py, uint8(land/4), uint8(land/4), 0xF, 0, clear); !ok {
		return 0, false
	}
	return cost, true
}

// ClearLand is the clearLand game command (0x00469CCB) over the tiles a..b
// (tile coordinates, clamped to the map), each building's removal counted
// once; the position the centre (world units) at its land height. The
// caller pays under Construction.
func (m *Map) ClearLand(c *Cmd, flags uint8, cx, cy int32, ax, ay, bx, by int) (int32, bool) {
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	removed := RemovedBuildings{}
	var total int32
	for y := y0; y <= y1; y++ { // TilePosRangeView: y outer, x inner
		for x := x0; x <= x1; x++ {
			cost, ok := m.clearTile(c, int32(x*tileSize), int32(y*tileSize), removed, flags)
			if !ok {
				return 0, false
			}
			total += cost
		}
	}
	land, _ := m.GetHeight(cx, cy)
	c.Pos = &[3]int32{cx, cy, int32(land)}
	return total, true
}

// RemoveWall is the removeWall game command (0x004C466C): the first wall on
// the tile at this height (world z) and rotation — a ghost only by a ghost
// removal — goes when applying. Always succeeds, costs nothing.
// PARTIAL: madeAnyChanges; the invalidation (goloco redraws).
func (m *Map) RemoveWall(c *Cmd, flags uint8, x, y, z int32, rotation uint8) (int32, bool) {
	c.Pos = &[3]int32{x + tileSize/2, y + tileSize/2, z}
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return 0, true
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindWall || int(r.Index) >= len(t.Walls) {
			continue
		}
		w := t.Walls[r.Index]
		if int32(w.BaseZ) != z/4 || w.Rotation != rotation {
			continue
		}
		if flags&CmdGhost == 0 && r.Flags&elemFlagGhost != 0 {
			continue
		}
		if flags&CmdApply == 0 {
			return 0, true
		}
		i := int(r.Index)
		t.Walls = append(t.Walls[:i], t.Walls[i+1:]...)
		t.removeRef(scenario.KindWall, i)
		return 0, true
	}
	return 0, true
}
