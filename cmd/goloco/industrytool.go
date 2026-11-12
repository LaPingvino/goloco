package main

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// The Fund New Industries tab's tool (IndustryList.cpp NewIndustries): set
// while the tab is open (switchTab's toolSet); the chosen industry's ghost
// follows the cursor (onToolUpdate), a click funds it there (onToolDown).
// PARTIAL: the selection's tile highlight (setMapSelectionArea), the
// construct sound, the error window's text ("Can't build this here…").

// industryTool is the tool's state: on while the tab is open; the ghost
// placed (_industryGhostPlaced/Pos/Type/Id); the last ghost's cost
// (_industryPlacementCost, ok false: kFailure); the placement seed
// (_placementPrng, a copy of gPrng2 after a draw).
type industryTool struct {
	on                 bool
	ghost              bool
	ghostID, ghostType uint8
	ghostX, ghostY     int32
	cost               int32
	costOK             bool
	prng               core.Prng
	win                *ui.TownListWindow // the tab's window: the tool ends with it
}

// industryToolOn is ToolManager::isToolActive for the tab: on while its
// window is the open list (another list replacing it ends the tool).
func (g *Game) industryToolOn() bool {
	if g.indTool.on && g.townList != g.indTool.win {
		g.industryToolAbort()
	}
	return g.indTool.on
}

// industryToolStart is switchTab's toolSet and the tab's prng step
// (gPrng2().randNext(); _placementPrng = gPrng2()).
func (g *Game) industryToolStart(win *ui.TownListWindow) {
	g.industryToolAbort()
	g.w.Rng2.RandNext()
	g.indTool = industryTool{on: true, prng: g.w.Rng2, win: win}
}

// industryToolAbort is onToolAbort (removeIndustryGhost) and the tool's end.
func (g *Game) industryToolAbort() {
	g.removeIndustryGhost()
	g.indTool.on = false
}

// industryPlacementAt is getIndustryPlacementArgsFromCursor: the surface or
// water under the cursor, the chosen industry (none chosen: none), the
// placement seed.
func (g *Game) industryPlacementAt(x, y int) (sim.IndustryPlacement, bool) {
	if g.lastIndustryOption < 0 {
		return sim.IndustryPlacement{}, false
	}
	px, py, ok := g.w.SurfaceOrWaterAt(x, y)
	if !ok {
		return sim.IndustryPlacement{}, false
	}
	return sim.IndustryPlacement{Type: uint8(g.lastIndustryOption), X: px, Y: py,
		Srand0: g.indTool.prng.Srand0(), Srand1: g.indTool.prng.Srand1()}, true
}

// removeIndustryGhost is 0x00458C09: the ghost removed (ghost, no payment).
func (g *Game) removeIndustryGhost() {
	t := &g.indTool
	if !t.ghost {
		return
	}
	t.ghost = false
	c := g.playerCmd()
	g.w.Map.DoCommandFor(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, sim.ExpenditureMiscellaneous,
		func(f uint8) (int32, bool) { return g.w.Map.RemoveIndustry(c, f, t.ghostID) })
}

// industryToolUpdate is onToolUpdate: no placement, no ghost; else a ghost
// always built at once, replaced only when the place or industry changed,
// its cost kept for the window.
func (g *Game) industryToolUpdate(x, y int) {
	a, ok := g.industryPlacementAt(x, y)
	if !ok {
		g.removeIndustryGhost()
		return
	}
	a.BuildImmediately = true // ghosts show the buildings, not scaffolding
	t := &g.indTool
	if t.ghost && t.ghostX == a.X && t.ghostY == a.Y && t.ghostType == a.Type {
		return
	}
	g.removeIndustryGhost()
	c := g.playerCmd()
	var id uint8
	cost, ok := g.w.Map.DoCommandFor(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, sim.ExpenditureMiscellaneous,
		func(f uint8) (int32, bool) {
			cost, gid, ok := g.w.Map.CreateIndustry(c, f, a)
			id = gid
			return cost, ok
		})
	t.cost, t.costOK = cost, ok
	if ok { // placeIndustryGhost: lastPlacedIndustryId
		t.ghost, t.ghostID, t.ghostX, t.ghostY, t.ghostType = true, id, a.X, a.Y, a.Type
	}
}

// industryToolDown is onToolDown: the ghost removed, the industry funded
// where clicked (createIndustry, apply; its cost at the place's land height),
// then a fresh placement seed.
func (g *Game) industryToolDown(x, y int) {
	g.removeIndustryGhost()
	if a, ok := g.industryPlacementAt(x, y); ok {
		c := g.playerCmd()
		land, _ := g.w.Map.GetHeight(a.X, a.Y)
		c.Pos = &[3]int32{a.X, a.Y, int32(land)} // setPosition
		g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) {
			cost, _, ok := g.w.Map.CreateIndustry(c, f, a)
			return cost, ok
		})
	}
	g.w.Rng2.RandNext()
	g.indTool.prng = g.w.Rng2
}
