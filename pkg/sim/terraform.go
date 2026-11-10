package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Raising and lowering land: GameCommands/Terraform/RaiseLand.cpp and
// LowerLand.cpp on TileManager::adjustSurfaceHeight (TileManager.cpp 1499),
// item 5.9f3t. Town road growth uses them on one whole tile.
// PARTIAL: the error texts, the construction sound, madeAnyChanges, the
// expenditure type / position, viewport invalidation.

// mapSelectionFull is MapSelectionType::full (the whole tile).
const mapSelectionFull = 4

const slopeRequiresHeightAdjustment = 1 << 5 // SurfaceSlope::requiresHeightAdjustment

// surfaceCornerHeight is TileManager::getSurfaceCornerHeight: the surface's
// highest corner (baseZ, + a step if sloped, + another if double height).
func surfaceCornerHeight(t *Tile) uint8 {
	z := t.BaseZ
	if t.Slope&0x1F != 0 {
		z += smallZStep
		if t.Slope&slopeDoubleHeight != 0 {
			z += smallZStep
		}
	}
	return z
}

// removeAllWallsAbove is TileManager::removeAllWallsOnTileAbove: the walls
// reaching above baseZ that start no more than 12 above it.
func (t *Tile) removeAllWallsAbove(baseZ uint8) {
	for i := len(t.Walls) - 1; i >= 0; i-- {
		w := t.Walls[i]
		if baseZ >= w.ClearZ || int(baseZ)+12 < int(w.BaseZ) {
			continue
		}
		t.Walls = append(t.Walls[:i], t.Walls[i+1:]...)
		t.removeRef(scenario.KindWall, i)
	}
}

// adjustSurfaceHeight is TileManager::adjustSurfaceHeight: the tile at world
// (px, py) to baseZ targetBaseZ with the slope slopeFlags — within the map's
// heights (4–160), its land's cost, everything in the way cleared (trees,
// buildings; anything else collides), a track or road above it still within
// its bridge's reach (a road or track on the ground: it would need a bridge),
// a ship route's water kept. Applied: the industry, growth / snow and walls
// above removed first (as upstream, before the checks), then the surface
// set, its land swapped for its replacement outside the editor, water under
// the new ground dropped. The cost, false where upstream fails.
func (m *Map) adjustSurfaceHeight(c *Cmd, px, py int32, targetBaseZ int, slopeFlags uint8, removed RemovedBuildings, flags uint8) (int32, bool) {
	if !m.validCoords(px, py) || targetBaseZ < 4 { // off the map / too low
		return 0, false
	}
	if targetBaseZ > 160 || targetBaseZ == 160 && slopeFlags&0x1F != 0 || targetBaseZ == 156 && slopeFlags&0x10 != 0 { // too high
		return 0, false
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	var cost int32
	if flags&CmdApply != 0 {
		m.removeSurfaceIndustry(px, py, t, c)
		if !c.Editor {
			t.setTerrainStyleAsCleared()
		}
		land, _ := m.GetHeight(px, py)
		t.removeAllWallsAbove(uint8(land / 4))
	}
	if c.Objects != nil && int(t.TerrainIndex) < len(c.Objects.LandObjects) {
		if lo := c.Objects.LandObjects[t.TerrainIndex]; lo != nil {
			cost += m.InflationAdjustedCost(lo.CostFactor, lo.CostIndex, 10)
		}
	}
	targetClearZ := targetBaseZ
	if slopeFlags&0x1F != 0 {
		targetClearZ += smallZStep
	}
	if slopeFlags&slopeDoubleHeight != 0 {
		targetClearZ += smallZStep
	}
	if t.Type6Flag && t.WaterLevel > 0 && (int(t.WaterLevel)-1)*microToSmallZStep < targetClearZ { // water a ship route needs
		return 0, false
	}
	clear := func(tt *Tile, k int) ClearResult {
		return m.tileClearFunction(c, tt, k, px, py, removed, flags, &cost, true)
	}
	if ok, _ := m.ApplyClearAtAllHeights(px, py, uint8(targetBaseZ), uint8(targetClearZ), 0xF, 0, clear); !ok {
		return 0, false
	}
	for _, r := range t.Order { // a track or road above: within its bridge's reach
		var baseZ uint8
		var hasBridge bool
		var bridge uint8
		switch r.Kind {
		case scenario.KindTrack:
			e := t.Tracks[r.Index]
			if e.Ghost {
				continue
			}
			baseZ, hasBridge, bridge = e.BaseZ, e.HasBridge, e.BridgeID
		case scenario.KindRoad:
			e := t.Roads[r.Index]
			if e.Ghost {
				continue
			}
			baseZ, hasBridge, bridge = e.BaseZ, e.HasBridge, e.BridgeID
		default:
			continue
		}
		height := int(baseZ) - targetBaseZ
		if height < 0 {
			continue
		}
		if !hasBridge { // '… requires a bridge'
			return 0, false
		}
		if b := c.bridge(bridge); b != nil && height > int(b.MaxHeight)*microToSmallZStep { // bridge already at maximum height
			return 0, false
		}
	}
	if flags&CmdApply == 0 {
		return cost, true
	}
	if !c.Editor {
		t.GrowthStage = 0
	}
	t.BaseZ = uint8(targetBaseZ)
	t.Slope = slopeFlags // setSlope: the whole byte, the snow bits with it
	if !c.Editor && c.Objects != nil && int(t.TerrainIndex) < len(c.Objects.LandObjects) {
		if lo := c.Objects.LandObjects[t.TerrainIndex]; lo != nil && lo.Flags&objects.LandFlagHasReplacementLandHeader != 0 {
			t.TerrainIndex = lo.ReplacementLandHeaderIdx
		}
	}
	if int(t.WaterLevel)*microToSmallZStep <= targetBaseZ {
		t.WaterLevel = 0
	}
	m.mapInvalidateTileFull(px, py)
	return cost, true
}

// tileRange is getClampedRange: the tiles from a to b (either order), inside the map.
func (m *Map) tileRange(ax, ay, bx, by int) (x0, y0, x1, y1 int) {
	x0, x1 = min(ax, bx), max(ax, bx)
	y0, y1 = min(ay, by), max(ay, by)
	return max(x0, 0), max(y0, 0), min(x1, m.Width-1), min(y1, m.Height-1)
}

// raiseLand is the RaiseLand game command over tiles a..b (tile
// coordinates): only the lowest tiles of the range move, each by
// raiseSurfaceCornerFlags (a step up where the table says so). The cost.
func (m *Map) raiseLand(c *Cmd, ax, ay, bx, by int, corner uint8, flags uint8) (int32, bool) {
	return m.raiseLandWith(c, ax, ay, bx, by, corner, flags, RemovedBuildings{})
}

// raiseLandWith is raiseLand with the caller's removed-buildings set (upstream
// passes it in: the mountain tool shares one over all its steps).
func (m *Map) raiseLandWith(c *Cmd, ax, ay, bx, by int, corner uint8, flags uint8, removed RemovedBuildings) (int32, bool) {
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	lowest := 255
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			lowest = min(lowest, int(m.Tiles[y][x].BaseZ))
		}
	}
	var total int32
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &m.Tiles[y][x]
			if int(t.BaseZ) > lowest {
				continue
			}
			target := lowest
			slope := raiseSurfaceCornerMapping[corner][t.Slope&0x1F]
			if slope&slopeRequiresHeightAdjustment != 0 {
				target += smallZStep
				slope &^= slopeRequiresHeightAdjustment
			}
			cost, ok := m.adjustSurfaceHeight(c, int32(x*tileSize), int32(y*tileSize), target, slope, removed, flags)
			if !ok {
				return 0, false
			}
			total += cost
		}
	}
	return total, true
}

// lowerLand is the LowerLand game command over tiles a..b: only the tiles
// whose highest corner is the range's highest move, each by
// lowerSurfaceCornerFlags (a step down where the table says so). The cost.
func (m *Map) lowerLand(c *Cmd, ax, ay, bx, by int, corner uint8, flags uint8) (int32, bool) {
	return m.lowerLandWith(c, ax, ay, bx, by, corner, flags, RemovedBuildings{})
}

// lowerLandWith is lowerLand with the caller's removed-buildings set (upstream
// passes it in: the mountain tool shares one over all its steps).
func (m *Map) lowerLandWith(c *Cmd, ax, ay, bx, by int, corner uint8, flags uint8, removed RemovedBuildings) (int32, bool) {
	x0, y0, x1, y1 := m.tileRange(ax, ay, bx, by)
	highest := 0
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			highest = max(highest, int(surfaceCornerHeight(&m.Tiles[y][x])))
		}
	}
	var total int32
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &m.Tiles[y][x]
			if int(surfaceCornerHeight(t)) < highest {
				continue
			}
			target := int(t.BaseZ)
			slope := lowerSurfaceCornerMapping[corner][t.Slope&0x1F]
			if slope&slopeRequiresHeightAdjustment != 0 {
				target -= smallZStep
				slope &^= slopeRequiresHeightAdjustment
			}
			cost, ok := m.adjustSurfaceHeight(c, int32(x*tileSize), int32(y*tileSize), target, slope, removed, flags)
			if !ok {
				return 0, false
			}
			total += cost
		}
	}
	return total, true
}
