package main

import (
	"log"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
)

// rightclick.go — the main view's right button: MouseInput.cpp's
// State::viewportRight (stateNormalRight 0x004C8048's viewport case,
// stateViewportRight 0x004C74BB) — a drag pans the view, a click (released
// within 500 ms without the mouse having moved) acts on what is under it —
// and ViewportInteraction::handleRightReleased with its rightReleased…
// actions.
// OpenLoco reference: src/OpenLoco/src/Input/MouseInput.cpp,
// src/OpenLoco/src/Ui/ViewportInteraction.cpp
// The view pans the way the mouse moves, twice as far (zoom + 1), the cursor
// hidden meanwhile (Ui::hideCursor); Options → Controls' "invert right mouse
// dragging" turns it round — the map then follows the mouse, as goloco's pan
// did before this port (at half the speed).
// PARTIAL: upstream also holds the hidden cursor where it was pressed
// (startCursorDrag) — goloco's moves on unseen and comes back where the
// mouse went, and stops panning at the screen's edge; a view following a
// vehicle is not let go (viewportUnfocusFromEntity); viewports of other
// windows; the right drag in scroll views (State::scrollRight).

// viewportRightState is Input's state for State::viewportRight.
type viewportRightState struct {
	on             bool      // Input::state() == State::viewportRight
	ticks          int       // _ticksSinceDragStart (ms; 1000 once the mouse moved)
	last           time.Time // getTimeSinceLastTick's clock
	pressX, pressY int       // _dragLast: where the button went down (the click's point)
	lastX, lastY   int       // the mouse at the last frame (the pan's step)
}

// updateViewportRight is the right button's part of handleMouse.
func (g *Game) updateViewportRight() {
	r := &g.vpRight
	if g.inMouseJustPressed(ebiten.MouseButtonRight) {
		// stateNormalRight: not in the title; only on the viewport widget (no
		// window under the mouse); an open dropdown takes the press instead.
		dropdown := (g.dropdown != nil && g.dropdown.Visible) || (g.colourDropdown != nil && g.colourDropdown.Visible)
		if !g.inTitleScreen && !dropdown && !g.overWindow(g.mouseX, g.mouseY) {
			*r = viewportRightState{on: true, last: gameNow(), pressX: g.mouseX, pressY: g.mouseY, lastX: g.mouseX, lastY: g.mouseY}
			ebiten.SetCursorMode(ebiten.CursorModeHidden) // Ui::hideCursor
		}
	}
	if !r.on {
		return
	}
	if g.inMouseJustReleased(ebiten.MouseButtonRight) || !g.inMousePressed(ebiten.MouseButtonRight) && !g.inMouseJustPressed(ebiten.MouseButtonRight) {
		r.on = false // Input::state(State::reset)
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
		if r.ticks <= 500 {
			g.handleRightReleased(r.pressX, r.pressY)
		}
		return
	}
	now := gameNow()
	r.ticks = min(r.ticks+int(now.Sub(r.last)/time.Millisecond), 2000)
	r.last = now
	if dx, dy := g.mouseX-r.lastX, g.mouseY-r.lastY; dx != 0 || dy != 0 {
		r.ticks = 1000
		// the view moves the mouse's way, twice as far (panZoom = zoom + 1);
		// Config::invertRightMouseViewPan turns it round
		shift := g.w.GetZoom() + 1
		if g.invertRightPan {
			dx, dy = -dx, -dy
		}
		g.w.PanCamera(float64(dx<<shift), float64(dy<<shift))
		r.lastX, r.lastY = g.mouseX, g.mouseY
	}
}

// rightOverState is what ViewportInteraction::rightOver reads from the
// windows: the construction window and its tab; a building tool on the map
// (Input's toolActive with flag6: the Terraform window's tools, the industry
// tool, the Construction tab's — the last always with the window).
func (g *Game) rightOverState() world.RightOverState {
	open := g.buildMode != buildModeNone && g.consWin() != nil
	vf := g.w.ViewFlags() // the viewport's see-through flags: what they hide is not found
	return world.RightOverState{
		SeeThroughBuildings: vf.Has(paint.ViewSeeThroughBuildings), SeeThroughTrees: vf.Has(paint.ViewSeeThroughTrees),
		SeeThroughScenery: vf.Has(paint.ViewSeeThroughScenery), SeeThroughTracks: vf.Has(paint.ViewSeeThroughTracks),
		SeeThroughRoads:  vf.Has(paint.ViewSeeThroughRoads),
		ConstructionOpen: open,
		StationTab:       open && g.consTab == 1,
		SignalTab:        open && g.consTab == 2,
		OverheadTab:      open && g.consTab == 3,
		ConstructTool:    g.tf.on || g.industryToolOn(),
		TipTimeout:       g.mapTip.timeout,
	}
}

// mapCommand runs a removal from the map as the player (doCommand with
// Flags::apply): refused, the error window under title; done, upstream plays
// the demolish sound for the track-side removals. Every one of them unpauses
// the game and is booked as construction.
// PARTIAL: Audio::playSound(SoundId::demolish) (no effect sounds yet).
func (g *Game) mapCommand(title string, run func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool)) bool {
	c, g1 := g.stationCmd()
	g.unpauseForCommand(sim.CmdApply)
	_, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return run(c, g1, f) })
	if !ok {
		g.openError(title, cmdErrorText(c))
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
// list, the industry list).
func (g *Game) handleRightReleased(x, y int) {
	if g.r.ObjMgr == nil {
		return
	}
	m := &g.w.Map
	ctrl, _ := g.w.ControllingCompany()
	item, _, _, ok := g.w.RightOver(x, y, g.rightOverState())
	if !ok {
		item = world.RightItem{}
	}
	label := func(id int) string { return i18n.Label(id) }
	log.Printf("[Game] right click at %d,%d: %v, tile %d,%d", x, y, item.Type, item.X>>5, item.Y>>5)
	switch item.Type {
	case paint.ItemTrack:
		if t := item.Track; t != nil {
			if int(t.Owner) == ctrl {
				g.openConstructionAt(false, t.TrackObjectID, t.TrackID, t.SequenceIndex, t.Rotation, item.X, item.Y, int32(t.BaseZ)*4,
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
				g.openConstructionAt(true, rd.RoadObjectID, rd.RoadID, rd.SequenceIndex, rd.Rotation, item.X, item.Y, int32(rd.BaseZ)*4,
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
			g.mapCommand(ui.FormatErrorArgs(label(i18n.CantRemovePop3String), name), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
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
			g.mapCommand(ui.FormatErrorArgs(label(i18n.CantRemovePop3String), name), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveRoadMod(c, f, a)
			})
		}
	case paint.ItemSignal:
		// rightReleasedSignal (0x004A5AA1): the side under the mouse (the
		// paint struct's mod id: 0 the left signal, 1 the right) — both when
		// the element has both and the single-direction tool is not chosen.
		// removeSignalsAuto with step 0 is removeSignal (AutoSignals.h).
		// PARTIAL: repeatedSignalMode's step (signals along the block).
		if s, t := item.Signal, item.Track; s != nil && t != nil {
			left, right := item.ModID == 0, item.ModID != 0
			if !g.consSigOne && s.Sides[0].HasSignal() && s.Sides[1].HasSignal() {
				left, right = true, true
			}
			g.removeConstructionGhosts()
			a := sim.SignalRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Rotation: t.Rotation, TrackID: t.TrackID,
				Index: t.SequenceIndex, TrackObjectID: t.TrackObjectID, Left: left, Right: right}
			g.mapCommand(label(i18n.CantRemoveSignal), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) { return m.RemoveSignal(c, f, a) })
		}
	case paint.ItemTrainStation: // rightReleasedTrainStation (0x004A5B66)
		if t := item.Track; item.Station != nil && t != nil {
			a := sim.TrainStationRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Rotation: t.Rotation, TrackID: t.TrackID,
				Index: t.SequenceIndex, TrackObjectID: t.TrackObjectID}
			g.mapCommand(label(i18n.CantRemoveStation), func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveTrainStation(c, f, a, g1)
			})
		}
	case paint.ItemRoadStation: // rightReleasedRoadStation (0x004A5BDF)
		if rd := item.Road; item.Station != nil && rd != nil {
			a := sim.RoadStationRemoval{X: item.X, Y: item.Y, Z: int32(rd.BaseZ) * 4, Rotation: rd.Rotation, RoadID: rd.RoadID,
				Index: rd.SequenceIndex, RoadObjectID: rd.RoadObjectID}
			g.mapCommand(label(i18n.CantRemoveStation), func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) {
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
			g.mapCommand(label(i18n.CantRemoveAirport), func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) { return m.RemoveAirport(c, f, a, g1) })
		}
	case paint.ItemDock: // rightReleasedDock (0x004A5CC5): from the dock's first tile
		if s := item.Station; s != nil {
			if !g.rightOverState().StationTab {
				g.openPortConstruction(1) // openWithFlags(1 << 30)
				return
			}
			fx, fy := firstTile(item.X, item.Y, s.SequenceIndex)
			a := sim.PortRemoval{X: fx, Y: fy, Z: int32(s.BaseZ) * 4}
			g.mapCommand(label(i18n.CantRemoveShipPort), func(c *sim.Cmd, g1 sim.G1Raw, f uint8) (int32, bool) { return m.RemovePort(c, f, a, g1) })
		}
	case paint.ItemTree: // rightReleasedTree (0x004BB116)
		if t := item.Tree; t != nil {
			a := sim.TreeRemoval{X: item.X, Y: item.Y, Z: int32(t.BaseZ) * 4, Type: t.TreeObjectID, Quadrant: t.Quadrant, Rotation: t.Rotation}
			g.mapCommand(label(i18n.ErrorCantRemoveThis), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) { return m.RemoveTree(c, f, a) })
		}
	case paint.ItemBuilding: // rightReleasedBuilding (0x0042D9BF): from its first tile
		if b := item.Building; b != nil {
			fx, fy := firstTile(item.X, item.Y, b.SequenceIndex)
			g.mapCommand(label(i18n.ErrorCantRemoveThis), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveBuilding(c, fx, fy, int32(b.BaseZ)*4, f)
			})
		}
	case paint.ItemWall: // rightReleasedWall (0x004C4809)
		if wl := item.Wall; wl != nil {
			g.mapCommand(label(i18n.ErrorCantRemoveThis), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveWall(c, f, item.X, item.Y, int32(wl.BaseZ)*4, wl.Rotation)
			})
		}
	case paint.ItemHeadquarterBuilding: // rightReleasedHeadquarter (0x0042F007)
		if b := item.Building; b != nil {
			fx, fy := firstTile(item.X, item.Y, b.SequenceIndex)
			g.mapCommand(label(i18n.ErrorCantRemoveThis), func(c *sim.Cmd, _ sim.G1Raw, f uint8) (int32, bool) {
				return m.RemoveCompanyHeadquarters(c, f, fx, fy, int32(b.BaseZ)*4)
			})
		}
	default: // noInteraction and the rest: getItemLeft's item opens its list
		it, ok := g.w.ItemLeftAt(x, y)
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
// Manhattan distance on the screen to each end tile's centre), straight and
// level, with the element's mods and — if it stands on one — its bridge.
// PARTIAL: the signal, station and bridge the scenario remembers per type
// (Scenario::getConstruction) — the window's first available, as elsewhere.
func (g *Game) openConstructionAt(road bool, obj, id, seq, rotation uint8, x, y, z int32, mods uint8, hasBridge bool, bridge uint8, sx, sy int) {
	next, prev, ok := sim.NextAndPreviousTile(road, id, seq, rotation, x, y, z)
	if !ok {
		return
	}
	g.removeConstructionGhosts()
	flags := obj
	if road {
		flags |= 1 << 7
	}
	g.consTab, g.consSigOne = 0, false
	g.consModSection = sim.ModSectionSingle // lastSelectedTrackModSection = single
	if hasBridge {
		g.consBridge = int(bridge)
	}
	g.openConstructionWithFlags(flags) // createConstructionWindow / resetWindow, setTrackOptions
	dist := func(e sim.ConstructionEnd) int {
		px, py := g.w.WorldToScreen(e.X+16, e.Y+16, e.Z)
		return abs(px-sx) + abs(py-sy)
	}
	end := prev
	if dist(next) < dist(prev) {
		end = next
	}
	g.consMods = mods
	g.w.ConstructionOpenAt(end, mods)
}
