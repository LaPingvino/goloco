package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Tile clearance — Map/TileClearance.cpp canConstructAtWithClear (the
// standard and any-height collision types, a clear function, the position
// flags) and canConstructAt (no clear function: any collision fails).
//
// PARTIAL (aiCompanyAboutToBuildCheck): exact for a neutral or non-player
// updating company (the map generator, towns, AIs: always "another company is
// about to build here"); for a player company building over an AI's
// allocation upstream checks the tile owner and the AI's challenge flags —
// taken as a collision here until companies' tile ownership is ported. The
// error texts are not kept.

const (
	elemFlagGhost       = 1 << 4
	elemFlagAiAllocated = 1 << 5
)

// clearZOf is an element's clearZ (the surface's for KindSurface).
func (t *Tile) clearZOf(r scenario.ElemRef) uint8 {
	switch r.Kind {
	case scenario.KindTrack:
		return t.Tracks[r.Index].ClearZ
	case scenario.KindStation:
		return t.Stations[r.Index].ClearZ
	case scenario.KindSignal:
		return t.Signals[r.Index].ClearZ
	case scenario.KindBuilding:
		return t.Buildings[r.Index].ClearZ
	case scenario.KindTree:
		return t.Trees[r.Index].ClearZ
	case scenario.KindWall:
		return t.Walls[r.Index].ClearZ
	case scenario.KindRoad:
		return t.Roads[r.Index].ClearZ
	case scenario.KindIndustry:
		return t.Industries[r.Index].ClearZ
	}
	return t.SurfaceClearZ()
}

// ClearResult is TileClearance's ClearFuncResult.
type ClearResult uint8

const (
	ClearAllCollisionsRemoved ClearResult = iota
	ClearCollision
	ClearCollisionErrorSet
	ClearNoCollision
	ClearCollisionRemoved
)

func isCollisionResult(r ClearResult) bool {
	return r == ClearCollision || r == ClearCollisionErrorSet
}

// ClearFunc is the clear function a game command passes: called on a
// colliding element (t.Order[k]); it may remove it (ClearCollisionRemoved —
// the tile is then scanned again from the start).
type ClearFunc func(t *Tile, k int) ClearResult

// ElementPositionFlags (TileManager.h).
const (
	PosAboveGround         = 1 << 0
	PosUnderground         = 1 << 1
	PosPartiallyUnderwater = 1 << 2
	PosUnderwater          = 1 << 3
)

// aiCompanyAboutToBuildCheck: see PARTIAL above.
func aiCompanyAboutToBuildCheck() ClearResult {
	return ClearCollisionErrorSet
}

func callClear(t *Tile, k int, clear ClearFunc) ClearResult {
	if clear == nil {
		return ClearCollision
	}
	return clear(t, k)
}

// canConstructAtWithClear is TileClearance's canConstructAtWithClear: ok,
// and the position flags it leaves (getPositionFlags).
func (m *Map) canConstructAtWithClear(px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, anyHeight bool, clear ClearFunc) (bool, uint8) {
	var pos uint8 = PosAboveGround
	drawable := func(c int32) bool { return c >= tileSize && c < int32(m.Width*tileSize)-tileSize-1 }
	if !drawable(px) || !drawable(py) {
		return false, pos
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	for {
		removed := false
		for k := 0; k < len(t.Order); k++ {
			var res ClearResult
			if t.Order[k].Kind == scenario.KindSurface {
				res = m.checkSurface(t, k, baseZ, clearZ, baseQuarter, zQuarter, clear, &pos)
			} else {
				res = checkNonSurface(t, k, baseZ, clearZ, baseQuarter, anyHeight, clear)
			}
			switch res {
			case ClearAllCollisionsRemoved:
				return true, pos
			case ClearCollision, ClearCollisionErrorSet:
				return false, pos
			case ClearCollisionRemoved:
				removed = true
			}
			if removed {
				break
			}
		}
		if !removed {
			return true, pos
		}
	}
}

// ApplyClearAtStandardHeight / ApplyClearAtAllHeights are TileClearance's.
func (m *Map) ApplyClearAtStandardHeight(px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc) (bool, uint8) {
	return m.canConstructAtWithClear(px, py, baseZ, clearZ, baseQuarter, zQuarter, false, clear)
}

func (m *Map) ApplyClearAtAllHeights(px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc) (bool, uint8) {
	return m.canConstructAtWithClear(px, py, baseZ, clearZ, baseQuarter, zQuarter, true, clear)
}

// CanConstructAt is TileClearance::canConstructAt: whether something from
// baseZ to clearZ over the quarters baseQuarter (z quarters zQuarter) fits on
// the tile at world (px, py).
func (m *Map) CanConstructAt(px, py int32, baseZ, clearZ uint8, baseQuarter, zQuarter uint8) bool {
	ok, _ := m.canConstructAtWithClear(px, py, baseZ, clearZ, baseQuarter, zQuarter, false, nil)
	return ok
}

// checkNonSurface is canConstructAtCheckNonSurfaceElement.
func checkNonSurface(t *Tile, k int, baseZ, clearZ, baseQuarter uint8, anyHeight bool, clear ClearFunc) ClearResult {
	r := t.Order[k]
	if anyHeight {
		if r.Kind == scenario.KindTree || r.Kind == scenario.KindBuilding || r.Kind == scenario.KindIndustry {
			return callClear(t, k, clear)
		}
		if r.Kind == scenario.KindStation {
			if st := t.Stations[r.Index].StationType; st == 2 || st == 3 {
				return callClear(t, k, clear)
			}
		}
	}
	if baseZ >= t.clearZOf(r) || clearZ <= t.baseZOf(r) || r.Flags&elemFlagGhost != 0 || r.Flags&0x0F&baseQuarter == 0 {
		return ClearNoCollision
	}
	if r.Flags&elemFlagAiAllocated == 0 {
		return callClear(t, k, clear)
	}
	if clear != nil {
		if res := clear(t, k); !isCollisionResult(res) {
			return res
		}
	}
	return aiCompanyAboutToBuildCheck()
}

// checkSurface is canConstructAtCheckSurfaceElement.
func (m *Map) checkSurface(t *Tile, k int, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc, pos *uint8) ClearResult {
	r := t.Order[k]
	if r.Flags&elemFlagAiAllocated != 0 {
		if res := aiCompanyAboutToBuildCheck(); res != ClearNoCollision {
			return res
		}
	}
	if waterZ := int(t.WaterLevel&0x1F) * microToSmallZStep; waterZ != 0 {
		if clearZ > t.SurfaceClearZ() && int(baseZ) < waterZ+4 {
			*pos |= PosPartiallyUnderwater
			if int(baseZ) < waterZ {
				*pos |= PosUnderwater
				if int(clearZ) > waterZ {
					if clear == nil {
						return ClearCollisionErrorSet
					}
					if res := clear(t, k); isCollisionResult(res) {
						return ClearCollisionErrorSet
					}
				}
			}
		}
	}
	if zQuarter == 0x0F {
		return ClearNoCollision
	}
	if clearZ <= t.BaseZ {
		*pos |= PosUnderground
		*pos &^= PosAboveGround
		return ClearNoCollision
	}
	if surfaceCornersFit(t, baseZ, baseQuarter, zQuarter) {
		return ClearNoCollision
	}
	return callClear(t, k, clear)
}

// surfaceCornersFit is the slope-corner part of canConstructAtCheckSurfaceElement.
func surfaceCornersFit(t *Tile, baseZ, baseQuarter, zQuarter uint8) bool {
	slope := t.Slope & 0x1F
	n, e, s, w := t.BaseZ, t.BaseZ, t.BaseZ, t.BaseZ
	if slope&slopeCornerNorth != 0 {
		n += smallZStep
		if slope == slopeDownSouth|slopeDoubleHeight {
			n += smallZStep
		}
	}
	if slope&slopeCornerEast != 0 {
		e += smallZStep
		if slope == slopeDownWest|slopeDoubleHeight {
			e += smallZStep
		}
	}
	if slope&slopeCornerSouth != 0 {
		s += smallZStep
		if slope == slopeDownNorth|slopeDoubleHeight {
			s += smallZStep
		}
	}
	if slope&slopeCornerWest != 0 {
		w += smallZStep
		if slope == slopeDownEast|slopeDoubleHeight {
			w += smallZStep
		}
	}
	dbl := int(baseZ) + 8
	ok := func(bit uint8, cz uint8) bool {
		return baseQuarter&bit == 0 || ((zQuarter&bit != 0 || baseZ >= cz) && dbl >= int(cz))
	}
	return ok(1, n) && ok(2, e) && ok(4, s) && ok(8, w)
}

// MaxHeight is PaintSession::setMaxHeight's _maxHeight for this tile: the
// highest clearZ of its elements (water, and an industrial surface + 24,
// count too), × kSmallZStep, + 32 — how far the tile can reach up the
// screen (the paint's vertical cull).
func (t *Tile) MaxHeight() int {
	var maxClearZ uint8
	for _, r := range t.Order {
		maxClearZ = max(maxClearZ, t.clearZOf(r))
		if r.Kind != scenario.KindSurface {
			continue
		}
		if t.WaterLevel != 0 {
			maxClearZ = max(maxClearZ, uint8(int(t.WaterLevel)*microToSmallZStep))
		}
		if t.IsIndustrial {
			maxClearZ = max(maxClearZ, t.SurfaceClearZ()+24)
		}
	}
	return int(maxClearZ)*smallZStep + 32
}
