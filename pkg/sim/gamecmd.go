package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// GameCommands::Flags.
const (
	CmdApply                   = 1 << 0
	CmdPreventBuildingClearing = 1 << 1
	CmdAllowNegativeCashFlow   = 1 << 2
	CmdNoErrorWindow           = 1 << 3
	CmdAiAllocated             = 1 << 4
	CmdNoPayment               = 1 << 5
	CmdGhost                   = 1 << 6
	CmdFlag7                   = 1 << 7
)

// CompanyNeutral is CompanyId::neutral.
const CompanyNeutral = 15

// Cmd is what a game command runs with besides its arguments: the updating
// company (GameCommands::getUpdatingCompanyId — neutral for towns, the
// generator and the editor) and gPrng2 (sounds draw from it).
type Cmd struct {
	Company uint8
	Rng2    *core.Prng
	Objects *objects.ObjectManager
}

// TileClearance::clearTreeCollision: the tree's clear cost (inflation
// adjusted, divisor 12); applied (not ghost / AI allocated) the tree is
// removed (TileManager::removeTree).
func (m *Map) clearTreeCollision(c *Cmd, t *Tile, k int, px, py int32, flags uint8, cost *int32) ClearResult {
	tr := t.Trees[t.Order[k].Index]
	var obj *objects.TreeObject
	if int(tr.TreeObjectID) < len(c.Objects.TreeObjects) {
		obj = c.Objects.TreeObjects[tr.TreeObjectID]
	}
	if obj != nil {
		*cost += m.InflationAdjustedCost(obj.ClearCostFactor, obj.CostIndex, 12)
	}
	if flags&(CmdGhost|CmdAiAllocated) != 0 || flags&CmdApply == 0 {
		return ClearNoCollision
	}
	last := k == len(t.Order)-1
	m.removeTree(c, t, k, px, py, CmdApply, obj)
	if last {
		return ClearAllCollisionsRemoved
	}
	return ClearCollisionRemoved
}

// removeTree is TileManager::removeTree: the demolish sound (a gPrng2 draw —
// randNext(20003, 24098) — when a company, not the neutral one, removes a
// real tree), the town rating reduction (updateTownInfo; its rating branch
// applies to companies only and is not ported yet), and the element removed.
func (m *Map) removeTree(c *Cmd, t *Tile, k int, px, py int32, flags uint8, obj *objects.TreeObject) {
	r := t.Order[k]
	if r.Flags&(elemFlagGhost|elemFlagAiAllocated) == 0 && c.Company != CompanyNeutral && c.Rng2 != nil {
		c.Rng2.RandNextRange(20003, 24098) // playDemolishTreeSound's frequency
	}
	if flags&CmdGhost == 0 {
		m.updateTownInfo(px, py, 0, 0, 0)
	}
	idx := int(r.Index)
	t.Trees = append(t.Trees[:idx], t.Trees[idx+1:]...)
	t.removeRef(scenario.KindTree, idx)
}

// tileClearFunction is TileClearance's tileClearFunction (clearWithDefaultCollision
// / clearWithoutDefaultCollision): surfaces pass, trees are cleared, buildings
// removed, anything else a collision or not as asked.
//
// PARTIAL: the building branch (clearBuildingCollision → the removeBuilding
// game command) is a collision until removeBuilding is ported (next in 5.9f3).
func (m *Map) tileClearFunction(c *Cmd, t *Tile, k int, px, py int32, flags uint8, cost *int32, defaultCollision bool) ClearResult {
	switch t.Order[k].Kind {
	case scenario.KindSurface:
		return ClearNoCollision
	case scenario.KindTree:
		return m.clearTreeCollision(c, t, k, px, py, flags, cost)
	case scenario.KindBuilding:
		return ClearCollision // PARTIAL: see above
	}
	if defaultCollision {
		return ClearCollision
	}
	return ClearNoCollision
}

// isBridgeRequired is CreateRoad.cpp's isBridgeRequired: above the surface
// always; level with it unless the piece is flat-capable (no unk4 preview
// flag) and its corner flags are none or the surface's slope corners.
func isBridgeRequired(baseZ uint8, t *Tile, pieceUnk4 bool, unk uint8) bool {
	if baseZ > t.BaseZ {
		return true
	}
	if baseZ == t.BaseZ {
		if !pieceUnk4 && (unk == 0 || unk == t.Slope&0x0F) {
			return false
		}
		return true
	}
	return false
}
