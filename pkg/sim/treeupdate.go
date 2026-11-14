package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Trees in the tile loop — Map/TreeElement.cpp updateTreeElement
// (0x004BD52B): the seasons turning, growing, snow, dying near industries
// that kill trees or too near buildings, and spreading.

// kSeasonDeathSequence: the season a dying tree shows next (0xFF: gone).
var kSeasonDeathSequence = [6]uint8{1, 4, 3, 0, 5, 0xFF}

// kTreeQuadrantOffset: a quadrant's tree position within its tile.
var kTreeQuadrantOffset = [4][2]int32{{7, 7}, {7, 23}, {23, 23}, {23, 7}}

// kTileOffsets is World::kOffsets: the four tiles of a 2×2 footprint.
var kTileOffsets = [4][2]int32{{0, 0}, {0, 32}, {32, 32}, {32, 0}}

// invalidateTreeElement is invalidateTree: the tree's column, radius 56.
func (m *Map) invalidateTreeElement(e *scenario.TreeElement, x, y int32) {
	m.invalidateTile(x, y, int32(e.BaseZ)*4, int32(e.ClearZ)*4, ZoomEighth, 56)
}

// updateTreeElement is updateTreeElement for the tree at t.Order[k] on the
// tile at world (x, y): false when the tile's elements changed (the tree
// removed, or a new one planted). PARTIAL: editor mode (it neither ages nor
// spreads trees there).
func (m *Map) updateTreeElement(t *Tile, k int, x, y int32, om *objects.ObjectManager) bool {
	if om == nil || int(t.Order[k].Index) >= len(t.Trees) {
		return true
	}
	e := &t.Trees[t.Order[k].Index]
	if e.Unk7l != 7 { // the season's change runs over 7 steps
		e.Unk7l++
		m.invalidateTreeElement(e, x, y)
		return true
	}
	obj := om.GetTreeObjectByIndex(int(e.TreeObjectID))
	if obj == nil {
		return true
	}
	if e.IsDying {
		next := kSeasonDeathSequence[e.Season%6]
		if next == 0xFF {
			m.invalidateTreeElement(e, x, y)
			idx := int(t.Order[k].Index)
			t.Trees = append(t.Trees[:idx], t.Trees[idx+1:]...)
			t.removeRef(scenario.KindTree, idx)
			return false
		}
		e.Season, e.Unk7l = next, 0
		m.invalidateTreeElement(e, x, y)
		return true
	}
	e.Unk5h = (e.Unk5h + 1) & 0xF
	if e.Unk5h != 0 {
		return true
	}

	if int(e.BaseZ)-4 <= int(m.CurrentSnowLine) { // below the snow line
		if e.HasSnow {
			e.HasSnow = false
			m.invalidateTreeElement(e, x, y)
		}
	} else if obj.Flags&objects.TreeFlagHasSnowVariation != 0 {
		if !e.HasSnow {
			e.HasSnow = true
			m.invalidateTreeElement(e, x, y)
		}
	} else {
		e.IsDying = true // killTree
		return true
	}

	if t.IsIndustrial { // the surface's industry (its byte 7) kills trees
		if io := m.IndustryObjectByID(t.Variation, om); io != nil && io.Flags&objects.IndustryFlagKillsTrees != 0 {
			e.IsDying = true
			return true
		}
	}

	if e.Season != m.CurrentSeason {
		e.Season, e.Unk7l = (e.Season+1)&0x3, 0
		m.invalidateTreeElement(e, x, y)
		return true
	}

	if newGrowth := e.Growth + 1; newGrowth < obj.Growth {
		oldClearZ := e.ClearZ
		e.ClearZ = e.BaseZ
		length := (int(obj.Height) - int(obj.InitialHeight)) * (int(newGrowth) + 1)
		if d := int(obj.Growth) - 1; d != 0 {
			length /= d
		}
		newClearZ := int(e.BaseZ) + (int(obj.InitialHeight)+length)/4
		occupied := uint8(1) << ((e.Quadrant + 2) & 0x3)
		if newClearZ > 255 || !m.CanConstructAt(x, y, e.BaseZ, uint8(newClearZ), occupied, 0b1111) {
			e.ClearZ = oldClearZ
			e.IsDying = true
			return true
		}
		e.ClearZ, e.Growth, e.Unk5h = uint8(newClearZ), newGrowth, 0
		m.invalidateTreeElement(e, x, y)
		return true
	}

	if obj.Var05 > 34 && m.treeObstructionsTooNear(x, y, e.Quadrant, e.BaseZ, e.ClearZ) {
		e.IsDying = true
		return true
	}

	rand := m.Rng1.RandNext()
	switch decision := rand & 0x3F; {
	case decision < 52:
		return true
	case decision < 58: // spread: a new tree nearby (7 tiles, 31 on 57)
		rand2 := m.Rng1.RandNext()
		dist := int32(7)
		if decision == 57 {
			dist = 31
		}
		ox := int32((rand2>>16)&uint32(dist*32)) - dist/2*32
		oy := int32((rand2&0xFFFF)&uint32(dist*32)) - dist/2*32
		if m.CurrentSeason == seasonWinter {
			return true
		}
		nx, ny := x+ox, y+oy
		if !m.validCoords(nx, ny) {
			return true
		}
		newID := e.TreeObjectID
		if rand&0x0F000000 == 0 {
			id, ok := m.RandomTreeTypeFromSurface(int(nx)>>5, int(ny)>>5, true, om)
			if !ok {
				return true
			}
			newID = id
		}
		newObj := om.GetTreeObjectByIndex(int(newID))
		if newObj == nil {
			return true
		}
		quadrant := uint8(rand>>5) & 0x3
		if newObj.Var05 > 34 {
			qx, qy := nx+kTreeQuadrantOffset[quadrant][0]-1, ny+kTreeQuadrantOffset[quadrant][1]-1
			land, _ := m.GetHeight(qx, qy)
			baseZ := int(land) / 4
			clearZ := int(newObj.Height)/4 + baseZ
			if clearZ > 255 || m.treeObstructionsTooNear(nx, ny, quadrant, uint8(baseZ), uint8(clearZ)) {
				return true
			}
		}
		numRot := (rand >> 19) & 0x1F
		colours := newObj.Colours>>numRot | newObj.Colours<<(32-numRot) // std::rotr
		colour := uint8(0)                                              // Colour::black
		if colours != 0 {
			low := uint32(0)
			for colours&(1<<low) == 0 {
				low++
			}
			colour = uint8((low + numRot) & 0x1F)
		}
		a := TreePlacement{X: nx, Y: ny, Quadrant: quadrant, Rotation: uint8(rand>>16) & 0x3, Type: newID, Colour: colour,
			RequiresFullClearance: true}
		m.CreateTreeCmd(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, a)
		return false
	default:
		if m.CurrentSeason == seasonAutumn || m.CurrentSeason == seasonWinter {
			e.IsDying = true
		}
		return true
	}
}

// treeObstructionsTooNear is hasObstructionsTooNear (0x004BD64A): around the
// quadrant's point a 3 × 3 grid of quarter tiles 16 apart, each cleared at
// baseZ–clearZ with clearFunction — a signal, station, surface or wall, or a
// track or road with a bridge, there (not AI-allocated or a ghost) means no
// trees.
func (m *Map) treeObstructionsTooNear(x, y int32, quadrant, baseZ, clearZ uint8) bool {
	left := x + kTileOffsets[quadrant][0]/2 - 16
	top := y + kTileOffsets[quadrant][1]/2 - 16
	noTrees := false
	clear := func(tt *Tile, kk int) ClearResult {
		r := tt.Order[kk]
		if r.Flags&(elemFlagGhost|elemFlagAiAllocated) != 0 {
			return ClearNoCollision
		}
		switch r.Kind {
		case scenario.KindRoad:
			if int(r.Index) < len(tt.Roads) && tt.Roads[r.Index].HasBridge {
				noTrees = true
			}
		case scenario.KindTrack:
			if int(r.Index) < len(tt.Tracks) && tt.Tracks[r.Index].HasBridge {
				noTrees = true
			}
		case scenario.KindSignal, scenario.KindStation, scenario.KindSurface, scenario.KindWall:
			noTrees = true
		}
		return ClearNoCollision
	}
	py := top
	for i := 0; i < 3; i++ {
		px := left
		for j := 0; j < 3; j++ {
			cx, cy := px, py
			var quarter uint8
			if px&0x1F != 0 {
				cx &^= 0x1F
				quarter = 1 << 1
				if py&0x1F != 0 {
					cy &^= 0x1F
					quarter = 1 << 0
				}
			} else {
				quarter = 1 << 2
				if py&0x1F != 0 {
					cy &^= 0x1F
					quarter = 1 << 3
				}
			}
			m.ApplyClearAtStandardHeight(cx, cy, baseZ, clearZ, quarter, 0b1111, clear)
			px += 16
		}
		py += 16
	}
	return noTrees
}
