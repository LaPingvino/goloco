package main

import (
	"github.com/LaPingvino/goloco/pkg/audio"
	"log"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
)

// rightclick.go — the right button: MouseInput.cpp's stateNormalRight
// (0x004C834A: the tooltip closed, the window under the mouse brought to the
// front, then by its widget) with State::viewportRight (the main view:
// stateViewportRight 0x004C74BB — a drag pans the view, a click, released
// within 500 ms without the mouse having moved, acts on what is under it) and
// State::scrollRight (a window's scroll view: scrollDragBegin,
// stateScrollRight 0x004C76A7 — a drag scrolls it), and
// ViewportInteraction::handleRightReleased with its rightReleased… actions.
// OpenLoco reference: src/OpenLoco/src/Input/MouseInput.cpp,
// src/OpenLoco/src/Input.cpp (startCursorDrag, getNextDragOffset),
// src/OpenLoco/src/Ui/ScrollView.cpp, src/OpenLoco/src/Ui/ViewportInteraction.cpp
// The view pans the way the mouse moves, twice as far (zoom + 1), the cursor
// hidden meanwhile (Ui::hideCursor); Options → Controls' "invert right mouse
// dragging" turns it round — the map then follows the mouse, as goloco's pan
// did before this port (at half the speed). A scroll view moves the way its
// thumb would (verticalDragFollow / horizontalDragFollow), the same switch
// turning it round.
// The Town, Industry, Station and Vehicle windows' viewports are
// viewportNoScrolling: a right press there holds the cursor (hidden, not
// pinned) and pans nothing; released within 500 ms — however far the mouse
// went — it acts on what is under the press point in that viewport.
// PARTIAL: a view following a vehicle is not let go (viewportUnfocusFromEntity).

// cursorDrag is startCursorDrag's hold on the cursor: hidden, and put back
// at the press point after every frame (getNextDragOffset) — the frame's move
// is the drag, and the rest of the frame (hover, tooltips) sees the mouse
// where it was pressed.
type cursorDrag struct {
	pressX, pressY int // _dragLast / _cursorDragStart: where the button went down
	lastX, lastY   int // the mouse as last read
	dx, dy         int // this frame's getNextDragOffset
}

func (c *cursorDrag) begin(x, y int) { *c = cursorDrag{pressX: x, pressY: y, lastX: x, lastY: y} }

// pin takes the frame's move and puts the mouse back at the press point.
func (c *cursorDrag) pin(x, y *int) {
	c.dx, c.dy = *x-c.lastX, *y-c.lastY
	c.lastX, c.lastY = *x, *y
	*x, *y = c.pressX, c.pressY
}

// viewportRightState is Input's state for State::viewportRight.
type viewportRightState struct {
	on    bool      // Input::state() == State::viewportRight
	ticks int       // _ticksSinceDragStart (ms; 1000 once the mouse moved)
	last  time.Time // getTimeSinceLastTick's clock
	// the window dragged in (_dragWindowType, _dragWindowNumber; main: the
	// main view) and its viewportNoScrolling
	main     bool
	typ      ui.WindowType
	number   int
	noScroll bool
	cursorDrag
}

// scrollRightState is Input's state for State::scrollRight: the window
// (_dragWindowType, _dragWindowNumber) and its scroll area
// (_dragScrollIndex).
type scrollRightState struct {
	on     bool
	typ    ui.WindowType
	number int
	index  int
	cursorDrag
}

// rightDragging reports a right drag holding the cursor (viewportRight or
// scrollRight).
func (g *Game) rightDragging() bool { return g.vpRight.on || g.scRight.on }

// pinRightDrag is the frame's getNextDragOffset for a right drag: called as
// soon as the frame's mouse is read.
func (g *Game) pinRightDrag() {
	var c *cursorDrag
	switch {
	case g.vpRight.on && g.vpRight.noScroll:
		return // stateViewportRight returns before getNextDragOffset: the cursor is not put back
	case g.vpRight.on:
		c = &g.vpRight.cursorDrag
	case g.scRight.on:
		c = &g.scRight.cursorDrag
	default:
		return
	}
	c.pin(&g.mouseX, &g.mouseY)
	if g.script != nil && g.script.hasMouse {
		// setCursorPos: a script's pointer goes back to the press point too,
		// so its next move counts from there (as xt.py's do in OpenLoco)
		g.script.mouseX, g.script.mouseY = g.mouseX, g.mouseY
		c.lastX, c.lastY = g.mouseX, g.mouseY
	}
}

// startCursorDrag hides the cursor and holds it (captured: the mouse still
// moves the drag, unbounded by the screen's edge, and the cursor comes back
// at the press point).
func startCursorDrag() { ebiten.SetCursorMode(ebiten.CursorModeCaptured) }

// updateViewportRight is the right button's part of handleMouse.
func (g *Game) updateViewportRight() {
	if g.inMouseJustPressed(ebiten.MouseButtonRight) && !g.rightDragging() {
		g.rightPressed(g.mouseX, g.mouseY)
	}
	g.updateScrollRight()
	r := &g.vpRight
	if !r.on {
		return
	}
	if !r.main && g.wins.Find(r.typ, r.number) == nil { // the window gone: State::reset
		r.on = false
		g.stopCursorDrag(&r.cursorDrag)
		return
	}
	if g.inMouseJustReleased(ebiten.MouseButtonRight) || !g.inMousePressed(ebiten.MouseButtonRight) && !g.inMouseJustPressed(ebiten.MouseButtonRight) {
		r.on = false // Input::state(State::reset)
		g.stopCursorDrag(&r.cursorDrag)
		if r.ticks <= 500 {
			g.handleRightReleased(r.pressX, r.pressY)
		}
		return
	}
	now := gameNow()
	r.ticks = min(r.ticks+int(now.Sub(r.last)/time.Millisecond), 2000)
	r.last = now
	if r.noScroll { // viewportNoScrolling: only the time counts
		return
	}
	if dx, dy := r.dx, r.dy; dx != 0 || dy != 0 {
		r.ticks = 1000
		// the view moves the mouse's way, twice as far (panZoom = zoom + 1);
		// Config::invertRightMouseViewPan turns it round
		shift := g.w.GetZoom() + 1
		if g.invertRightPan {
			dx, dy = -dx, -dy
		}
		g.w.PanCamera(float64(dx<<shift), float64(dy<<shift))
	}
}

// rightPressed is stateNormalRight: an open dropdown takes the press
// instead; on a window the tooltip closes and the window comes to the front —
// with a modal window open nothing more (its scrollModalRight: the file
// browser's, update.go), nor in the title — then a scroll view begins
// scrollRight, a viewport widget viewportRight (viewportDragBegin); on no
// window the main view begins viewportRight.
func (g *Game) rightPressed(x, y int) {
	if (g.dropdown != nil && g.dropdown.Visible) || (g.colourDropdown != nil && g.colourDropdown.Visible) {
		return
	}
	if w := g.wins.FindAt(x, y); w != nil {
		g.tip.close(gameNow()) // WindowManager::close(WindowType::tooltip)
		g.wins.BringToFront(w)
		if g.wins.Modal != 0 || g.prompt != nil || g.inTitleScreen {
			return
		}
		b := w.Base()
		if i := ui.ScrollAreaAt(w, x, y); i >= 0 { // scrollDragBegin
			g.scRight = scrollRightState{on: true, typ: b.WinType, number: b.Number, index: i}
			g.scRight.begin(x, y)
			startCursorDrag()
		} else if vw, ok := w.(ui.ViewportWidgets); ok && vw.ViewportWidgetAt(x, y) {
			// viewportDragBegin: these windows' viewports are viewportNoScrolling
			// (initViewport sets it with the viewport)
			g.vpRight = viewportRightState{on: true, last: gameNow(), typ: b.WinType, number: b.Number, noScroll: true}
			g.vpRight.begin(x, y)
			startCursorDrag()
		}
		return
	}
	if g.inTitleScreen {
		return
	}
	g.vpRight = viewportRightState{on: true, last: gameNow(), main: true} // viewportDragBegin
	g.w.StopGlide()                                                       // viewportDragBegin: scrollingToLocation off
	g.vpRight.begin(x, y)
	startCursorDrag()
}

// updateScrollRight is stateScrollRight: the window gone, the state resets;
// released, it resets (no click); held, the frame's move scrolls the area
// (horizontalDragFollow, then verticalDragFollow), inverted by
// Config::invertRightMouseViewPan.
func (g *Game) updateScrollRight() {
	r := &g.scRight
	if !r.on {
		return
	}
	w := g.wins.Find(r.typ, r.number)
	if w == nil || !g.inMousePressed(ebiten.MouseButtonRight) {
		r.on = false // Input::state(State::reset)
		g.stopCursorDrag(&r.cursorDrag)
		return
	}
	dx, dy := r.dx, r.dy
	if dx == 0 && dy == 0 {
		return
	}
	s, ok := w.(ui.ScrollAreas)
	if !ok {
		return
	}
	areas := s.ScrollAreas()
	if r.index >= len(areas) {
		return
	}
	if g.invertRightPan {
		dx, dy = -dx, -dy
	}
	areas[r.index].DragFollow(dx, dy)
}

// stopCursorDrag is Input::stopCursorDrag: the cursor shown again, back at
// the press point (the captured mode gives it back there; a script's pointer
// is put there too, as setCursorPos puts X's).
func (g *Game) stopCursorDrag(c *cursorDrag) {
	g.applyCursor()
	if g.script != nil && g.script.hasMouse {
		g.script.mouseX, g.script.mouseY = c.pressX, c.pressY
	}
}

// rightOverState is what ViewportInteraction::rightOver reads from the
// windows: the construction window and its tab; a building tool on the map
// (Input's toolActive with flag6: the Terraform window's tools, the industry
// tool, the Construction tab's — the last always with the window).
func (g *Game) rightOverState() world.RightOverState {
	open := g.buildMode != buildModeNone && g.consWin() != nil
	return world.RightOverState{
		ConstructionOpen: open,
		StationTab:       open && g.consTab == 1,
		SignalTab:        open && g.consTab == 2,
		RepeatedSignals:  g.consSigRepeat,
		OverheadTab:      open && g.consTab == 3,
		ConstructTool:    g.tf.on || g.industryToolOn(),
		TipTimeout:       g.mapTip.timeout,
	}
}

// mapCommand runs a removal from the map as the player (doCommand with
// Flags::apply): refused, the error window under title; done, upstream plays
// the demolish sound at the command's position for the track-side removals
// (demolish: signals, stations, airports, docks, track and road extras; not
// trees, buildings, walls or headquarters). Every one of them unpauses the
// game and is booked as construction.
func (g *Game) mapCommand(title string, demolish bool, run func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool)) bool {
	c, g1 := g.stationCmd()
	g.unpauseForCommand(sim.CmdApply)
	_, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return run(c, g1, f) })
	if !ok {
		g.openCmdError(title, c)
	} else if demolish {
		g.playCommandSound(audio.SoundDemolish, c)
	}
	return ok
}

// handleRightReleased is ViewportInteraction::handleRightReleased: rightOver
// at the point, then the item's action — the player's track or road opens
// the construction window on it (another company's its Company window); on
// the construction window's tabs a signal, station or extra is removed; an
// airport or dock opens its window, or is removed on the Station tab; in a
// building mode a tree, wall, building or headquarters is removed. Nothing
// to act on: getItemLeft's item opens its list (the vehicle's owner's
// vehicle list of its type, the town list, the station's owner's station
// list, the industry list). (x, y) is the press point; the viewports are
// those of the window there (the main view's, or a window's).
func (g *Game) handleRightReleased(x, y int) {
	if g.r.ObjMgr == nil {
		return
	}
	m := &g.w.Map
	ctrl, _ := g.w.ControllingCompany()
	vps, _ := g.viewportsAt(x, y)
	item, _, _, ok := g.w.RightOver(vps, x, y, g.rightOverState())
	if !ok {
		item = world.RightItem{}
	}
	label := func(id int) string { return i18n.Label(id) }
	log.Printf("[Game] right click at %d,%d: %v, tile %d,%d", x, y, item.Type, item.X>>5, item.Y>>5)
	switch item.Type {
	case paint.ItemTrack:
		if t := item.Track; t != nil {
			if int(t.Owner) == ctrl {
				g.openConstructionAt(vps, false, t.TrackObjectID, t.TrackID, t.SequenceIndex, t.Rotation, item.X, item.Y, int32(t.BaseZ)*4,
					t.Mods, t.HasBridge, t.BridgeID, x, y)
			} else {
				g.openCompanyWindow(int(t.Owner))
			}
		}
	case paint.ItemRoad:
		if rd := item.Road; rd != nil {
			ro := g.r.ObjMgr.GetRoadObjectByIndex(int(rd.RoadObjectID))
			anyRoad := ro != nil && ro.Flags&(1<<3) != 0 // RoadObjectFlags::anyRoadTypeCompatible
			if int(rd.Owner) == ctrl || rd.Owner == sim.CompanyNeutral || anyRoad {
				mods := rd.Mods
				if anyRoad { // openAtRoad: such a road's mods are not taken
					mods = 0
				}
				g.openConstructionAt(vps, true, rd.RoadObjectID, rd.RoadID, rd.SequenceIndex, rd.Rotation, item.X, item.Y, int32(rd.BaseZ)*4,
					mods, rd.HasBridge, rd.BridgeID, x, y)
			} else {
				g.openCompanyWindow(int(rd.Owner))
			}
		}
	case paint.ItemTrackExtra: // rightReleasedTrackExtra (0x004A1303)
		if t := item.Track; t != nil {
			g.removeConstructionGhosts()
			name := ""
			if o := g.r.ObjMgr.GetTrackObjectByIndex(int(t.TrackObjectID)); o != nil && int(item.ModID) < len(o.ModNames) {
				name = g.r.ObjMgr.ExtraName(o.ModNames[item.ModID])
			}
			a := sim.TrackModsPlacement{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Rotation: t.Rotation, TrackID: t.TrackID,
				Index: t.SequenceIndex, Type: 1 << item.ModID, TrackObjectID: t.TrackObjectID, Section: g.consModSection}
			g.mapCommand(ui.FormatErrorArgs(label(i18n.CantRemovePop3String), name), true, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveTrackMod(c, f, a)
			})
		}
	case paint.ItemRoadExtra: // rightReleaseRoadExtra (0x004A13C1)
		if rd := item.Road; rd != nil {
			g.removeConstructionGhosts()
			name := ""
			if o := g.r.ObjMgr.GetRoadObjectByIndex(int(rd.RoadObjectID)); o != nil && int(item.ModID) < len(o.ModNames) {
				name = g.r.ObjMgr.ExtraName(o.ModNames[item.ModID])
			}
			a := sim.RoadModsPlacement{X: item.X, Y: item.Y, Z: int32(rd.BaseZ) * 4, Rotation: rd.Rotation, RoadID: rd.RoadID,
				Index: rd.SequenceIndex, Type: 1 << item.ModID, RoadObjectID: rd.RoadObjectID, Section: g.consModSection}
			g.mapCommand(ui.FormatErrorArgs(label(i18n.CantRemovePop3String), name), true, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveRoadMod(c, f, a)
			})
		}
	case paint.ItemSignal:
		// rightReleasedSignal (0x004A5AA1): the side under the mouse (the
		// paint struct's mod id: 0 the left signal, 1 the right) — both when
		// the element has both and the single-direction tool is not chosen.
		// removeSignalsAuto: with repeated placement step 1 (every signal
		// along the block), else step 0 (removeSignal).
		if s, t := item.Signal, item.Track; s != nil && t != nil {
			left, right := item.ModID == 0, item.ModID != 0
			if !g.consSigOne && s.Sides[0].HasSignal() && s.Sides[1].HasSignal() {
				left, right = true, true
			}
			g.removeConstructionGhosts()
			a := sim.SignalRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Rotation: t.Rotation, TrackID: t.TrackID,
				Index: t.SequenceIndex, TrackObjectID: t.TrackObjectID, Left: left, Right: right}
			step := uint8(0)
			if g.consSigRepeat {
				step = 1
			}
			g.mapCommand(label(i18n.CantRemoveSignal), true, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveSignalsAuto(c, f, sim.SignalsRemovalAuto{SignalRemoval: a, Step: step})
			})
		}
	case paint.ItemTrainStation: // rightReleasedTrainStation (0x004A5B66)
		if t := item.Track; item.Station != nil && t != nil {
			a := sim.TrainStationRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Rotation: t.Rotation, TrackID: t.TrackID,
				Index: t.SequenceIndex, TrackObjectID: t.TrackObjectID}
			g.mapCommand(label(i18n.CantRemoveStation), true, func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveTrainStation(c, f, a, g1)
			})
		}
	case paint.ItemRoadStation: // rightReleasedRoadStation (0x004A5BDF)
		if rd := item.Road; item.Station != nil && rd != nil {
			a := sim.RoadStationRemoval{X: item.X, Y: item.Y, Z: int32(rd.BaseZ) * 4, Rotation: rd.Rotation, RoadID: rd.RoadID,
				Index: rd.SequenceIndex, RoadObjectID: rd.RoadObjectID}
			g.mapCommand(label(i18n.CantRemoveStation), true, func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveRoadStation(c, f, a, g1)
			})
		}
	case paint.ItemAirport: // rightReleasedAirport (0x004A5C58)
		if s := item.Station; s != nil {
			if !g.rightOverState().StationTab {
				g.openPortConstruction(0) // openWithFlags(1 << 31)
				return
			}
			a := sim.AirportRemoval{X: item.X, Y: item.Y, Z: int32(s.BaseZ) * 4}
			g.mapCommand(label(i18n.CantRemoveAirport), true, func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) { return m.RemoveAirport(c, f, a, g1) })
		}
	case paint.ItemDock: // rightReleasedDock (0x004A5CC5): from the dock's first tile
		if s := item.Station; s != nil {
			if !g.rightOverState().StationTab {
				g.openPortConstruction(1) // openWithFlags(1 << 30)
				return
			}
			fx, fy := firstTile(item.X, item.Y, s.SequenceIndex)
			a := sim.PortRemoval{X: fx, Y: fy, Z: int32(s.BaseZ) * 4}
			g.mapCommand(label(i18n.CantRemoveShipPort), true, func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) { return m.RemovePort(c, f, a, g1) })
		}
	case paint.ItemTree: // rightReleasedTree (0x004BB116)
		if t := item.Tree; t != nil {
			a := sim.TreeRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Type: t.TreeObjectID, Quadrant: t.Quadrant, Rotation: t.Rotation}
			g.mapCommand(label(i18n.ErrorCantRemoveThis), false, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) { return m.RemoveTree(c, f, a) })
		}
	case paint.ItemBuilding: // rightReleasedBuilding (0x0042D9BF): from its first tile
		if b := item.Building; b != nil {
			fx, fy := firstTile(item.X, item.Y, b.SequenceIndex)
			g.mapCommand(label(i18n.ErrorCantRemoveThis), false, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveBuilding(c, fx, fy, int32(b.BaseZ)*4, f)
			})
		}
	case paint.ItemWall: // rightReleasedWall (0x004C4809)
		if wl := item.Wall; wl != nil {
			g.mapCommand(label(i18n.ErrorCantRemoveThis), false, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveWall(c, f, item.X, item.Y, int32(wl.BaseZ)*4, wl.Rotation)
			})
		}
	case paint.ItemHeadquarterBuilding: // rightReleasedHeadquarter (0x0042F007)
		if b := item.Building; b != nil {
			fx, fy := firstTile(item.X, item.Y, b.SequenceIndex)
			g.mapCommand(label(i18n.ErrorCantRemoveThis), false, func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveCompanyHeadquarters(c, f, fx, fy, int32(b.BaseZ)*4)
			})
		}
	default: // noInteraction and the rest: getItemLeft's item opens its list
		it, ok := g.w.ItemLeftAt(vps, x, y)
		if !ok {
			return
		}
		switch it.Kind {
		case paint.ItemEntity:
			if v, ok := m.LiveVehicle(uint16(it.ID)); ok {
				g.openVehicleListFor(int(v.Owner), int(v.VehicleType))
			}
		case paint.ItemTownLabel:
			g.openTownsWindow()
		case paint.ItemStationLabel:
			if owner, ok := m.StationOwner(it.ID); ok {
				g.openStationList(int(owner), 0)
			}
		case paint.ItemIndustry:
			g.openIndustryList()
		}
	}
}

// firstTile is pos − World::kOffsets[sequenceIndex]: a multi-tile building's
// first tile.
func firstTile(x, y int32, seq uint8) (int32, int32) {
	off := [4][2]int32{{0, 0}, {0, 32}, {32, 32}, {32, 0}}[seq&3]
	return x - off[0], y - off[1]
}

// openConstructionAt is Construction::openAtTrack (0x004A0EAD) / openAtRoad
// (0x004A147F): the construction window (opened, or reset to its
// Construction tab) for the element's track or road type, building on from
// the piece's end nearer the click (isPointCloserToNextOrPreviousTile: by
// Manhattan distance on the screen to each end tile's centre, as the clicked
// window's first viewport shows them — openAtTrack(*window, …)'s
// viewports[0]), straight and level, with the element's mods and — if it
// stands on one — its bridge.
// The signal, station and bridge are the ones the scenario remembers for
// the type (Scenario::getConstruction, openConstructionWithFlags' consRecall),
// the element's own bridge over the remembered one.
func (g *Game) openConstructionAt(vps []world.Viewport, road bool, obj, id, seq, rotation uint8, x, y, z int32, mods uint8, hasBridge bool, bridge uint8, sx, sy int) {
	next, prev, ok := sim.NextAndPreviousTile(road, id, seq, rotation, x, y, z)
	if !ok {
		return
	}
	g.removeConstructionGhosts()
	flags := obj
	if road {
		flags |= 1 << 7
	}
	g.consTab, g.consSigOne, g.consToolOff = 0, false, false
	g.consModSection = sim.ModSectionSingle // lastSelectedTrackModSection = single
	g.openConstructionWithFlags(flags)      // createConstructionWindow / resetWindow, setTrackOptions; the remembered choices
	if hasBridge {
		g.consBridge = int(bridge) // copyElement->hasBridge(): its bridge
	}
	vp := g.w.MainViewport()
	if len(vps) > 0 {
		vp = vps[0]
	}
	dist := func(e sim.ConstructionEnd) int {
		px, py := vp.ToScreen(g.w, e.X+16, e.Y+16, e.Z)
		return abs(px-sx) + abs(py-sy)
	}
	end := prev
	if dist(next) < dist(prev) {
		end = next
	}
	g.consMods = mods
	g.w.ConstructionOpenAt(end, mods)
}
