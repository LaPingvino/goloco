package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// blueprint.go — the Construction tab's Copy and Paste tools (the blueprint
// itself is pkg/sim blueprint.go). Copy: a press marks a tile, a drag shows
// the area, the release copies the company's track, signals and stations
// there and turns the paste tool on. Paste: the blueprint follows the cursor
// as a ghost (turned by rotate_90 while hovering) and a click builds it.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/ConstructionTab.cpp
// (onMouseUp kCopy / kPaste / kRotate90, onToolUpdate, onToolDown,
// onToolDrag, onToolUp, onToolAbort, onUpdate).

// The Construction tab's map tools (ToolManager::toolSet's widget).
const (
	consToolConstruct = iota // widx::construct
	consToolCopy             // widx::copy
	consToolPaste            // widx::paste
)

// consCopy is the Copy button: the ghosts gone, the blueprint forgotten, the
// overground view, the copy tool, hovering.
func (g *Game) consCopy() {
	c := g.playerCmd()
	g.w.RemoveConstructionGhosts(c)
	g.w.SetBlueprint(nil)
	g.consOverground()
	g.consTool = consToolCopy
	g.w.Cons.Hover = true
	g.w.ActivateSelection()
}

// consPaste is the Paste button: the ghosts gone; with the paste tool on it
// goes off (toolCancel → onToolAbort), else the overground view, the paste
// tool, hovering.
func (g *Game) consPaste() {
	g.w.RemoveConstructionGhosts(g.playerCmd())
	if g.consTool == consToolPaste {
		g.pasteToolAbort()
		return
	}
	g.consOverground()
	g.consTool = consToolPaste
	g.w.Cons.Hover = true
	g.w.ActivateSelection()
}

// pasteToolAbort is onToolAbort for the paste tool: paste no longer
// activated, the window no longer hovering (it builds on from its position).
func (g *Game) pasteToolAbort() {
	g.consTool = consToolConstruct
	g.bpDrag = stationDrag{from: g.bpDrag.from, to: g.bpDrag.to}
	g.w.Cons.Hover = false
}

// bpToolDown is onToolDown for the copy and paste tools: the tile under the
// cursor (the surface or water) is _toolPosInitial; not dragging.
func (g *Game) bpToolDown(sx, sy int) {
	g.bpDrag.pressed, g.bpDrag.dragging = true, false
	if px, py, ok := g.w.SurfaceOrWaterAt(sx, sy); ok {
		g.bpDrag.from = [2]int{int(px) >> 5, int(py) >> 5}
	}
}

// bpToolDrag is onToolDrag: nothing for paste; for copy the ghosts go and,
// once the cursor is over another tile (_toolPosDrag), the area between is
// the map selection.
func (g *Game) bpToolDrag(sx, sy int) {
	if g.consTool == consToolPaste {
		return
	}
	g.w.RemoveConstructionGhosts(g.playerCmd())
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return
	}
	g.bpDrag.to = [2]int{int(px) >> 5, int(py) >> 5}
	g.bpDrag.dragging = g.bpDrag.to != g.bpDrag.from
	if !g.bpDrag.dragging {
		return
	}
	a, b := g.bpDrag.from, g.bpDrag.to
	g.w.SetMapSelection(true, min(a[0], b[0]), min(a[1], b[1]), max(a[0], b[0]), max(a[1], b[1]))
}

// bpToolUpdate is onToolUpdate's paste part: the blueprint's ghost at
// getConstructionPos under the cursor (the copy tool has no update).
func (g *Game) bpToolUpdate(sx, sy int) {
	if g.consTool != consToolPaste {
		return
	}
	g.syncBlueprintG1()
	tx, ty := g.tileUnderCursor(sx, sy)
	if tx < 0 {
		return
	}
	if pos, ok := g.w.ConstructionPos(tx, ty); ok {
		g.w.PasteToolUpdate(g.playerCmd(), pos)
	}
}

// bpToolUp is onToolUp. Copy: the selection gone, the area from the press to
// the last drag tile copied (both kept from earlier presses as upstream's
// statics are), the overground view and the paste tool. Paste: the
// blueprint built at getConstructionPos under the cursor; refused,
// error_cant_build_this_here with the reason; built, the error windows the
// placements opened close (placeBlueprint), else the last stays.
func (g *Game) bpToolUp(sx, sy int) {
	g.bpDrag.pressed, g.bpDrag.dragging = false, false
	c := g.playerCmd()
	switch g.consTool {
	case consToolCopy:
		g.w.SetMapSelection(false, 0, 0, 0, 0) // resetMapSelectionFlags: every selection flag
		g.w.ConsSelection.On, g.w.ConsArrow.On = false, false
		g.w.RemoveConstructionGhosts(c)
		bp := g.w.CopyTrackToBlueprint(g.bpDrag.from, g.bpDrag.to)
		g.w.SetBlueprint(&bp)
		g.consOverground()
		g.consTool = consToolPaste
	case consToolPaste:
		tx, ty := g.tileUnderCursor(sx, sy)
		if tx < 0 || g.w.Blueprint() == nil {
			return
		}
		pos, ok := g.w.ConstructionPos(tx, ty)
		if !ok {
			return
		}
		g.syncBlueprintG1()
		_, res, ok := g.w.PasteToolUp(c, pos)
		switch {
		case !ok:
			g.openError(i18n.Label(i18n.ErrorCantBuildThisHere), cmdErrorText(c))
		case res.Built:
			g.errorWin = nil // WindowManager::close(WindowType::error)
		case res.Failed != nil:
			g.openErrorQuiet(i18n.Label(i18n.ErrorCantBuildThisHere), cmdErrorText(res.Failed))
		}
	}
}

// syncBlueprintG1 hands the station commands the blueprint places the G1.
func (g *Game) syncBlueprintG1() {
	if g.r.ObjMgr != nil {
		g1, _ := g.r.ObjMgr.G1File.(sim.G1Raw)
		g.w.SetBlueprintG1(g1)
	}
}

// bpToolOn: the Construction tab's tool is copy or paste.
func (g *Game) bpToolOn() bool {
	return g.consHasHead() && g.consTab == 0 && g.w.Cons.Hover && g.consTool != consToolConstruct && !g.inTitleScreen
}
