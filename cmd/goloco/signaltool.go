package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// signaltool.go — the construction window's Signal tab tool
// (Construction/SignalTab.cpp onToolUpdate / onToolUp, 04e7900): under the
// cursor the track piece gets a ghost signal (or, with repeated placement, a
// ghost every step pieces both ways to the next junction) and the tab shows
// its cost; the release builds it. One side — the side of the piece nearer the
// cursor's end of it (getSide) — or both, as the tab's buttons choose.
// PARTIAL: the construct sound; the signal's "wrong type" check is against
// the window's track type only.

// signalToolOn is the Signal tab's tool on the map (a track window — roads
// have no signal tab — its tool not cancelled).
func (g *Game) signalToolOn() bool {
	return g.buildMode != buildModeNone && !g.consIsRoad() && g.consTab == 2 && !g.inTitleScreen && !g.consToolCancelled()
}

// signalPlacementAt is getSignalPlacementArgsFromCursor: the track element
// under (sx, sy) with the tab's signal, sides and (repeated) step.
func (g *Game) signalPlacementAt(sx, sy int) (sim.SignalsPlacementAuto, bool) {
	el, x, y, ok := g.w.TrackUnder(sx, sy)
	if !ok || g.r.ObjMgr == nil {
		return sim.SignalsPlacementAuto{}, false
	}
	list := g.r.ObjMgr.AvailableCompatibleSignals(g.buildObjID, int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.SignalsPlacementAuto{}, false
	}
	a := sim.SignalsPlacementAuto{SignalPlacement: sim.SignalPlacement{X: x, Y: y, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation,
		TrackID: el.TrackID, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID, Type: uint8(pickSlot(list, g.consSignal))}}
	if g.consSigRepeat {
		a.Step = uint8(g.consSigStep)
	}
	if !g.consSigOne {
		a.Left, a.Right = true, true // 0xC000
		return a, true
	}
	// getSide: closer to the piece's next tile → right (0x4000), else left.
	next, prev, ok := sim.NextAndPreviousTile(false, el.TrackID, el.SequenceIndex, el.Rotation, x, y, a.Z)
	if !ok {
		return sim.SignalsPlacementAuto{}, false
	}
	dist := func(e sim.ConstructionEnd) int {
		px, py := g.w.WorldToScreen(e.X+16, e.Y+16, e.Z)
		return abs(px-sx) + abs(py-sy)
	}
	if dist(next) < dist(prev) {
		a.Right = true
	} else {
		a.Left = true
	}
	return a, true
}

// signalToolUpdate is onToolUpdate: no piece (or another track type) — the
// ghosts go and the cost with them; the same placement as the ghost shown —
// nothing; else the ghosts replaced and the tab's cost the new one's.
func (g *Game) signalToolUpdate(sx, sy int) {
	a, ok := g.signalPlacementAt(sx, sy)
	if g.overWindow(sx, sy) || !ok || int(a.TrackObjectID) != g.buildObjID {
		g.removeConstructionGhosts()
		g.consSigCost = -1
		return
	}
	if g.w.SignalGhostIs(a) {
		return
	}
	g.removeConstructionGhosts()
	cost, placed := g.w.PlaceSignalGhost(g.playerCmd(), a)
	if !placed {
		cost = -1
	}
	g.consSigCost = cost
}

// signalToolUp is onToolUp: the ghosts off, then the placement for real under
// "Can't build signal(s) here" — another track type refused at once
// (wrong_type_of_track_road).
func (g *Game) signalToolUp(sx, sy int) {
	g.removeConstructionGhosts()
	a, ok := g.signalPlacementAt(sx, sy)
	if !ok {
		return
	}
	title := i18n.CantBuildSignalHere
	if !g.consSigOne || g.consSigRepeat {
		title = i18n.CantBuildSignalsHere
	}
	if int(a.TrackObjectID) != g.buildObjID {
		g.openError(i18n.Label(title), i18n.Label(i18n.WrongTypeOfTrackRoad))
		return
	}
	c := g.playerCmd()
	g.unpauseForCommand(sim.CmdApply)
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) {
		return g.w.CreateSignalsAuto(c, f, a)
	}); !ok {
		g.openError(i18n.Label(title), cmdErrorText(c))
	}
}
