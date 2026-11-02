package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Tile clearance — Map/TileClearance.cpp canConstructAt (standard height,
// no clear function: any collision fails).
//
// PARTIAL: aiCompanyAboutToBuildCheck is taken as a collision (what it is
// for the neutral company the map generator runs as; a player building over
// an AI's allocation is decided by the updating company there), the clear
// functions (applyClearAt…), the anyHeight collision type, the
// position flags (underwater / underground …) and the error texts.

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

// CanConstructAt is TileClearance::canConstructAt: whether something from
// baseZ to clearZ over the quarters baseQuarter (z quarters zQuarter) fits on
// the tile at world (px, py).
func (m *Map) CanConstructAt(px, py int32, baseZ, clearZ uint8, baseQuarter, zQuarter uint8) bool {
	drawable := func(c int32) bool { return c >= tileSize && c < int32(m.Width*tileSize)-tileSize-1 }
	if !drawable(px) || !drawable(py) {
		return false
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	for _, r := range t.Order {
		if r.Kind == scenario.KindSurface {
			if !m.surfaceAllows(t, r, baseZ, clearZ, baseQuarter, zQuarter) {
				return false
			}
			continue
		}
		bz, cz := t.baseZOf(r), t.clearZOf(r)
		if baseZ >= cz || clearZ <= bz || r.Flags&elemFlagGhost != 0 || r.Flags&0x0F&baseQuarter == 0 {
			continue
		}
		return false // a collision (or an AI's allocation — see PARTIAL)
	}
	return true
}

// surfaceAllows is canConstructAtCheckSurfaceElement without a clear function.
func (m *Map) surfaceAllows(t *Tile, r scenario.ElemRef, baseZ, clearZ, baseQuarter, zQuarter uint8) bool {
	if r.Flags&elemFlagAiAllocated != 0 {
		return false // aiCompanyAboutToBuildCheck: PARTIAL
	}
	if waterZ := int(t.WaterLevel&0x1F) * microToSmallZStep; waterZ != 0 {
		if clearZ > t.SurfaceClearZ() && int(baseZ) < waterZ+4 && int(baseZ) < waterZ && int(clearZ) > waterZ {
			return false // partly above, partly below water
		}
	}
	if zQuarter == 0x0F {
		return true
	}
	if clearZ <= t.BaseZ {
		return true // underground
	}
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
