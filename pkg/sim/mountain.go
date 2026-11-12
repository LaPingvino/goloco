package sim

// GameCommands/Terraform/LowerRaiseLandMountain.cpp: the mountain tool —
// the selection raised / lowered as a whole, then the land round it smoothed
// in rings so the slope never steps more than allowed.

// MountainArgs is LowerRaiseLandMountainArgs (world units): the centre, the
// selection's corners, and +1 to raise (anything else lowers).
type MountainArgs struct {
	CX, CY, AX, AY, BX, BY int32
	Adjustment             int16
}

// mountainTool holds the command's statics (_mtnToolCost, _mtnToolGCFlags,
// _mtnToolHeightDiff, _mtnToolOuterLoopIndex), as upstream's 8-bit types.
type mountainTool struct {
	m          *Map
	c          *Cmd
	cost       uint32
	flags      uint8
	heightDiff uint8
	removed    RemovedBuildings
}

func (m *Map) validWorld(x, y int32) bool {
	return x >= 0 && y >= 0 && x < int32(m.Width)*tileSize && y < int32(m.Height)*tileSize
}

// surfaceCornerDownHeight is getSurfaceCornerDownHeight: the base, a step
// up when the corner is raised, and another for a double-height slope whose
// corners are all but the opposite one.
func surfaceCornerDownHeight(t *Tile, cornerUp uint8) uint8 {
	z := t.BaseZ
	if t.Slope&0x1F&cornerUp != 0 {
		z += smallZStep
		var down uint8
		switch cornerUp {
		case slopeCornerNorth:
			down = slopeCornersAll &^ slopeCornerSouth
		case slopeCornerSouth:
			down = slopeCornersAll &^ slopeCornerNorth
		case slopeCornerWest:
			down = slopeCornersAll &^ slopeCornerEast
		case slopeCornerEast:
			down = slopeCornersAll &^ slopeCornerWest
		}
		if t.Slope&slopeDoubleHeight != 0 && t.Slope&0x0F == down {
			z += smallZStep
		}
	}
	return z
}

func (mt *mountainTool) cornerDown(x, y int32, cornerUp uint8) uint8 {
	return surfaceCornerDownHeight(mt.m.TileAt(int(x)>>5, int(y)>>5), cornerUp)
}

// adjustSurfaceSlope: a tile whose reference corner differs from the target
// by more than the allowed height difference gets that corner moved a step
// towards it (lower / raiseSurfaceCornerFlags), the cost kept when it worked.
func (mt *mountainTool) adjustSurfaceSlope(x, y int32, targetBaseZ int8, targetCorner, refCorner uint8) {
	if !mt.m.validWorld(x, y) {
		return
	}
	t := mt.m.TileAt(int(x)>>5, int(y)>>5)
	diff := int8(surfaceCornerDownHeight(t, refCorner)) - targetBaseZ
	var slope uint8
	switch {
	case diff > 0:
		if uint8(diff) <= mt.heightDiff {
			return
		}
		slope = lowerSurfaceCornerMapping[targetCorner][t.Slope&0x1F]
		targetBaseZ = int8(t.BaseZ)
		if slope&slopeRequiresHeightAdjustment != 0 {
			targetBaseZ -= smallZStep
			slope &^= slopeRequiresHeightAdjustment
		}
	case diff < 0:
		if uint8(-diff) <= mt.heightDiff {
			return
		}
		slope = raiseSurfaceCornerMapping[targetCorner][t.Slope&0x1F]
		targetBaseZ = int8(t.BaseZ)
		if slope&slopeRequiresHeightAdjustment != 0 {
			targetBaseZ += smallZStep
			slope &^= slopeRequiresHeightAdjustment
		}
	default:
		return
	}
	if cost, ok := mt.m.adjustSurfaceHeight(mt.c, x&^31, y&^31, int(uint8(targetBaseZ)), slope, mt.removed, mt.flags); ok {
		mt.cost += uint32(cost)
	}
}

func (mt *mountainTool) south(x, y int32, z uint8) {
	mt.adjustSurfaceSlope(x, y, int8(z), 2, slopeCornerSouth)
}
func (mt *mountainTool) west(x, y int32, z uint8) {
	mt.adjustSurfaceSlope(x, y, int8(z), 3, slopeCornerWest)
}
func (mt *mountainTool) north(x, y int32, z uint8) {
	mt.adjustSurfaceSlope(x, y, int8(z), 0, slopeCornerNorth)
}
func (mt *mountainTool) east(x, y int32, z uint8) {
	mt.adjustSurfaceSlope(x, y, int8(z), 1, slopeCornerEast)
}

// adjustMountainCentre: raiseLand / lowerLand over the selection, every corner.
func (mt *mountainTool) adjustMountainCentre(a MountainArgs) (int32, bool) {
	ax, ay, bx, by := int(a.AX)>>5, int(a.AY)>>5, int(a.BX)>>5, int(a.BY)>>5
	if a.Adjustment == 1 {
		return mt.m.raiseLandWith(mt.c, ax, ay, bx, by, mapSelectionFull, mt.flags, mt.removed)
	}
	return mt.m.lowerLandWith(mt.c, ax, ay, bx, by, mapSelectionFull, mt.flags, mt.removed)
}

// The four smoothen steps: two corners of the target matched to the
// reference tile's facing corners, the allowed difference widened as the
// target passes the selection's edge.
func (mt *mountainTool) smoothenNorth(rx, ry, tx, ty int32, a MountainArgs) {
	mt.east(tx, ty, mt.cornerDown(rx, ry, slopeCornerSouth))
	mt.heightDiff -= smallZStep
	if ty >= a.AY {
		mt.heightDiff += smallZStep
		if ty > a.BY {
			mt.heightDiff += smallZStep
		}
	}
	mt.north(tx, ty, mt.cornerDown(rx, ry, slopeCornerWest))
}

func (mt *mountainTool) smoothenEast(rx, ry, tx, ty int32, a MountainArgs) {
	mt.south(tx, ty, mt.cornerDown(rx, ry, slopeCornerWest))
	mt.heightDiff -= smallZStep
	if tx >= a.AX {
		mt.heightDiff += smallZStep
		if tx > a.BX {
			mt.heightDiff += smallZStep
		}
	}
	mt.east(tx, ty, mt.cornerDown(rx, ry, slopeCornerNorth))
}

func (mt *mountainTool) smoothenSouth(rx, ry, tx, ty int32, a MountainArgs) {
	mt.west(tx, ty, mt.cornerDown(rx, ry, slopeCornerNorth))
	mt.heightDiff -= smallZStep
	if ty <= a.BY {
		mt.heightDiff += smallZStep
		if ty < a.AY {
			mt.heightDiff += smallZStep
		}
	}
	mt.south(tx, ty, mt.cornerDown(rx, ry, slopeCornerEast))
}

func (mt *mountainTool) smoothenWest(rx, ry, tx, ty int32, a MountainArgs) {
	mt.north(tx, ty, mt.cornerDown(rx, ry, slopeCornerEast))
	mt.heightDiff -= smallZStep
	if tx <= a.BX {
		mt.heightDiff += smallZStep
		if tx < a.AX {
			mt.heightDiff += smallZStep
		}
	}
	mt.west(tx, ty, mt.cornerDown(rx, ry, slopeCornerSouth))
}

func clamp32(v, lo, hi int32) int32 { return min(max(v, lo), hi) }

// LowerRaiseLandMountain is the lowerRaiseLandMountain game command
// (0x00462DCE): refused outside the map; the selection raised / lowered (a
// second time when its first tile is still sloped — stopping there when it
// still is), then rings round it from the selection's width out to 64 more,
// each ring's four sides smoothed towards the selection's edge corners with
// an allowed difference growing by 8 a ring. The cost (failed steps cost
// nothing and don't stop it). The caller pays under Construction.
// PARTIAL: the construct sound; madeAnyChanges.
func (m *Map) LowerRaiseLandMountain(c *Cmd, flags uint8, a MountainArgs) (int32, bool) {
	mt := &mountainTool{m: m, c: c, flags: flags, removed: RemovedBuildings{}}
	done := func() (int32, bool) {
		land, _ := m.GetHeight(a.CX, a.CY)
		c.Pos = &[3]int32{a.CX, a.CY, int32(land)}
		return int32(mt.cost), true
	}
	if !m.validWorld(a.AX, a.AY) || !m.validWorld(a.BX, a.BY) {
		return 0, false
	}
	if cost, ok := mt.adjustMountainCentre(a); ok {
		mt.cost += uint32(cost)
	}
	if m.TileAt(int(a.AX)>>5, int(a.AY)>>5).Slope&0x1F != 0 {
		if cost, ok := mt.adjustMountainCentre(a); ok {
			mt.cost += uint32(cost)
		}
		if m.TileAt(int(a.AX)>>5, int(a.AY)>>5).Slope&0x1F != 0 {
			return done()
		}
	}
	minR := (a.BX-a.AX)/tileSize - 1
	maxR := minR + 64
	r := minR
	bx, by := a.AX, a.AY
	outer := int8(-4)
	for {
		outer += 4
		r += 2
		bx -= tileSize
		by -= tileSize
		if r >= maxR {
			break
		}
		mt.heightDiff = uint8(outer * 2)

		mt.north(bx, by, mt.cornerDown(a.AX, a.AY, slopeCornerSouth))
		by += tileSize
		for i := r; i > 0; i-- {
			mt.smoothenNorth(a.AX, clamp32(by, a.AY, a.BY), bx, by, a)
			by += tileSize
		}

		mt.east(bx, by, mt.cornerDown(a.AX, a.BY, slopeCornerWest))
		bx += tileSize
		for i := r; i > 0; i-- {
			mt.smoothenEast(clamp32(bx, a.AX, a.BX), a.BY, bx, by, a)
			bx += tileSize
		}

		mt.south(bx, by, mt.cornerDown(a.BX, a.BY, slopeCornerNorth))
		by -= tileSize
		for i := r; i > 0; i-- {
			mt.smoothenSouth(a.BX, clamp32(by, a.AY, a.BY), bx, by, a)
			by -= tileSize
		}

		mt.west(bx, by, mt.cornerDown(a.BX, a.AY, slopeCornerEast))
		bx -= tileSize
		for i := r; i > 0; i-- {
			mt.smoothenWest(clamp32(bx, a.AX, a.BX), a.AY, bx, by, a)
			bx -= tileSize
		}
	}
	return done()
}
