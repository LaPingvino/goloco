package main

import (
	"log"
	"os"
	"time"

	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// update.go — Game.Update: one frame of input and simulation. The modal
// windows first, then the title screen or the game (tools, windows, the map
// click, the camera, the game ticks).

func (g *Game) Update() error {
	g.windowFrame++
	if g.diagShip != 0 && g.w != nil {
		if tr, err := g.w.Map.TrainOf(g.diagShip); err == nil {
			g.diagShipFrames++
			if g.diagFollow && g.diagShipFrames%15 == 0 {
				g.w.CentreOnVehicle(g.diagShip)
			}
			if g.diagShipFrames%120 == 0 {
				h, v2 := g.w.Map.Entities[tr.Head], g.w.Map.Entities[tr.Veh2]
				log.Printf("[Diag] shipsail frame %d: ship at %d,%d yaw %d, %.0f mph, status %d, station %d", g.diagShipFrames, v2.PosX, v2.PosY, v2.SpriteYaw, float64(v2.CurrentSpeed)/65536, h.Status, h.StationID)
			}
		}
	}
	if g.w != nil {
		g.w.TitleMode = g.inTitleScreen // SceneManager::isTitleMode() for the viewport's names
	}
	if fstats.on {
		t0 := time.Now()
		defer func() { fstats.addUpdate(time.Since(t0)) }()
	}
	// The script steps before the modal windows, so a playthrough can type
	// into a text input or answer a prompt (it used to stall behind one).
	if g.script != nil {
		g.script.step()
		g.scriptScenario()
		g.scriptCamera()
		if g.script.quit {
			return ebiten.Termination
		}
	}
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	if g.diagMouse != nil {
		g.mouseX, g.mouseY = g.diagMouse[0], g.diagMouse[1]
	}
	if g.script != nil && g.script.hasMouse {
		g.mouseX, g.mouseY = g.script.mouseX, g.script.mouseY
	}
	if g.textInput != nil { // TextInputWindow: keyUp (typed characters, editing keys, Enter / Escape), onMouseUp, onUpdate
		t := g.textInput
		t.Update()
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			t.HandleClick(g.mouseX, g.mouseY)
		}
		var keys []int
		for k, code := range map[ebiten.Key]int{ebiten.KeyBackspace: ui.KeyBackspace, ebiten.KeyDelete: ui.KeyDelete, ebiten.KeyHome: ui.KeyHome,
			ebiten.KeyEnd: ui.KeyEnd, ebiten.KeyArrowLeft: ui.KeyLeft, ebiten.KeyArrowRight: ui.KeyRight} {
			if inpututil.IsKeyJustPressed(k) || inpututil.KeyPressDuration(k) > 30 && inpututil.KeyPressDuration(k)%3 == 0 { // key repeat
				keys = append(keys, code)
			}
		}
		t.Keys(g.inChars(), keys, g.inKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),
			g.inKeyJustPressed(ebiten.KeyEscape))
		if t.Closed && g.textInput == t {
			g.textInput = nil
		}
		return nil
	}
	if g.prompt != nil { // the modal confirmation takes all input (setCurrentModalType)
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			g.prompt.HandleClick(mx, my)
		}
		g.prompt.Key(inpututil.IsKeyJustPressed(ebiten.KeyEscape),
			inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter))
		if g.prompt.Closed {
			g.prompt = nil
		}
		return nil
	}
	// A right click on the load/save list deletes a file (after asking).
	if g.browse != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if mx, my := ebiten.CursorPosition(); g.browse.Contains(mx, my) {
			g.browse.HandleRightClick(mx, my)
		}
	}
	if g.quitAfterSave {
		return ebiten.Termination
	}
	if g.browse != nil { // PromptBrowse: mouse-over row, typed filename, Enter/Escape
		g.browse.HandleHover(g.mouseX, g.mouseY)
		g.browse.HandleText(ebiten.AppendInputChars(nil), inpututil.IsKeyJustPressed(ebiten.KeyBackspace),
			inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyEscape))
	}

	if g.inTitleScreen {
		// Title screen: advance camera animation and apply to world.
		// w.Update() must be called here too — it increments animTick which
		// drives water wave animation and demo train movement.
		g.titleFrame++
		g.w.Map.TitleMode = true
		for range g.w.GameTicks(g.mouseMoved()) { // tickLogic(numUpdates): the title is a running game upstream
			g.w.Update()
			g.w.TickSavedVehicles()
			if g.titleSeq != nil {
				g.titleSeq.Update()
			}
		}
		if g.titleSeq != nil {
			cx, cy := g.titleSeq.GetCameraPosition()
			g.w.SetCamera(cx, cy)
		}
		if g.browse != nil { // the load/save list's thumb follows the mouse while held
			g.browse.Drag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
		}
		if g.scenarioSelect != nil { // and the scenario list's
			g.scenarioSelect.Drag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
			g.scenarioSelect.MouseOver(g.mouseX, g.mouseY)
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
		}
		if g.dropdown != nil && g.dropdown.Visible { // a window's dropdown (Options' language, heights …)
			g.dropdown.UpdateHover(g.mouseX, g.mouseY)
		}
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			// An open dropdown is on top: it takes the click (a click outside
			// closes it); then the ScenarioSelect window, then the window manager.
			if g.dropdown != nil && g.dropdown.Visible {
				if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
					g.dropdown = nil
				}
			} else if g.townList != nil && g.townList.Contains(g.mouseX, g.mouseY) {
				g.townList.HandleClick(g.mouseX, g.mouseY)
			} else if g.options != nil && g.options.Contains(g.mouseX, g.mouseY) {
				g.options.HandleClick(g.mouseX, g.mouseY)
			} else if g.browse != nil && g.browse.Contains(g.mouseX, g.mouseY) {
				g.browse.HandleClick(g.mouseX, g.mouseY)
			} else if g.scenarioSelect != nil && g.scenarioSelect.Contains(g.mouseX, g.mouseY) {
				g.scenarioSelect.HandleClick(g.mouseX, g.mouseY)
			} else if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
				g.handleTitleMenuClick(g.mouseX, g.mouseY)
			}
		}
		if g.inMouseJustReleased(ebiten.MouseButtonLeft) {
			g.windowMgr.StopDrag()
		}
		g.routeWheelToWindows()
		return nil
	}

	// --- Gameplay mode ONLY below this point ---

	// Escape cancels the Route tab's tool (toolCancel).
	if g.routeTool && !g.pickupOn && g.inKeyJustPressed(ebiten.KeyEscape) {
		g.routeTool = false
	}
	// Escape cancels the pickup tool (pickupToolAbort: no ghost to remove yet).
	if g.pickupOn && g.inKeyJustPressed(ebiten.KeyEscape) {
		head := g.pickupHead
		g.removeLandGhost() // pickupToolAbort
		g.pickupOn = false
		if g.vehicleWin != nil {
			g.openVehicleDetails(head)
		}
	}

	if g.diagMoney && len(g.w.Map.MoneyEffects) == 0 { // DIAGNOSTIC ONLY (moneyeffect)
		g.w.Map.TitleMode = false
		if px, py, ok := g.w.SurfaceOrWaterAt(g.sw/2, g.sh/2); ok {
			t := g.w.Map.TileAt(int(px)>>5, int(py)>>5)
			g.w.Map.SpendMoneyEffect(px, py, int32(t.BaseZ)*4+28, g.w.Map.PlayerCompanies[0], -26447)
			if g.titleFrame%300 == 0 {
				log.Printf("[Diag] moneyeffect: at %d,%d: %d effects, lists %v", px, py, len(g.w.Map.MoneyEffects), g.w.Map.EntityLists != nil)
			}
		}
	}

	// The station tab's ghost follows the cursor (onToolUpdateTrainStation),
	// and goes once the tab or the mode does.
	if g.diagHoldGhost {
		// DIAGNOSTIC ONLY: the ghost stays for the screenshot
	} else if (g.buildMode == buildModeTrack || g.buildMode == buildModeRoad || g.buildMode == buildModeDock || g.buildMode == buildModeAirport) && g.consTab == 1 && !g.inTitleScreen {
		g.stationToolUpdate(g.mouseX, g.mouseY)
	} else if g.stationGhost != nil || g.roadStationGhost != nil || g.dockGhost != nil || g.airportGhost != nil {
		g.removeStationGhost()
	}

	// Build mode: update hover tile every frame (track and road; the dock
	// window has no construction head)
	if g.buildMode == buildModeTrack || g.buildMode == buildModeRoad {
		g.hoverTileX, g.hoverTileY = g.tileUnderCursor(g.mouseX, g.mouseY)
		// Keyboard shortcuts for build mode
		if g.inKeyJustPressed(ebiten.KeyR) {
			// Both track and road drive the shared construction head.
			g.w.RotateConstructionHead()
		}
		if g.inKeyJustPressed(ebiten.KeyX) {
			if g.buildMode == buildModeRoad {
				g.w.UndoLastRoad()
			} else {
				g.w.UndoLastTrack()
			}
		}
		if g.inKeyJustPressed(ebiten.KeyEscape) {
			g.buildMode = buildModeNone
			g.hoverTileX = -1
			g.hoverTileY = -1
			g.w.ClearConstructionHead()
		}
	}
	g.updateTrackGhost()
	// updateConstruction's updateConstructionArrow: the track / road window on
	// its Construction tab with the head placed (not while hovering).
	if g.w != nil {
		g.w.UpdateConstructionArrow((g.buildMode == buildModeTrack || g.buildMode == buildModeRoad) && g.consTab == 0 && !g.inTitleScreen && g.w.ConstructionHeadState().Active)
	}

	// Update dropdown hover if visible
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	} else if g.colourDropdown != nil && g.colourDropdown.Visible {
		g.colourDropdown.UpdateHover(g.mouseX, g.mouseY)
	} else {
		g.toolbar.UpdateHover(g.mouseX, g.mouseY)
	}

	if g.townList != nil { // the list's scrollbar thumb follows the mouse while held
		g.townList.Drag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.browse != nil { // the load/save list's thumb too
		g.browse.Drag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.pickupOn && !g.inTitleScreen { // pickupToolUpdate: the ghost under the cursor
		if g.overWindow(g.mouseX, g.mouseY) {
			g.removeLandGhost()
		} else {
			g.pickupToolUpdate(g.mouseX, g.mouseY)
		}
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
	}
	if g.companyList != nil {
		g.companyList.ScrollDrag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.vehicleWin != nil {
		g.vehicleWin.ScrollDrag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
		g.vehicleWin.SpeedHeld(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.buildVehicle != nil {
		g.buildVehicle.Drag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.companyFinances != nil {
		g.companyFinances.UpdateHeld(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
		g.companyFinances.UpdateDrag(g.mouseX, g.mouseY, g.inMousePressed(ebiten.MouseButtonLeft))
	}
	if g.companyList != nil {
		g.companyList.MouseOver(g.mouseX, g.mouseY)
		cl := g.companyList // findAt: in the window and no modal over it
		in := g.mouseX >= cl.X && g.mouseX < cl.X+cl.W && g.mouseY >= cl.Y && g.mouseY < cl.Y+cl.H
		cl.Update(g.mouseX, g.mouseY, in && g.companyFinances == nil, ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight))
	}
	if g.companyFinances != nil && g.companyFinances.Positioning { // statePositioningWindow
		if g.inMousePressed(ebiten.MouseButtonLeft) {
			g.companyFinances.PositionMove(g.mouseX, g.mouseY, g.sh)
		} else {
			g.companyFinances.PositionEnd(g.mouseX, g.mouseY, g.sh)
		}
	}
	if g.vehicleWin != nil && g.vehicleWin.Positioning { // statePositioningWindow
		if g.inMousePressed(ebiten.MouseButtonLeft) {
			g.vehicleWin.PositionMove(g.mouseX, g.mouseY, g.sh)
		} else {
			g.vehicleWin.PositionEnd(g.mouseX, g.mouseY, g.sh)
		}
	}
	if g.vehicleWin != nil && os.Getenv("GOLOCO_DIAG_DRAG_HOLD") == "" {
		g.vehicleWin.MouseOver(g.mouseX, g.mouseY)
	}
	if g.vehicleWin != nil && g.vehicleWin.Dragging && os.Getenv("GOLOCO_DIAG_DRAG_HOLD") == "" { // DragVehiclePart: follows the mouse, dropped on release (the diag's hold: neither)
		g.vehicleWin.DragTo(g.mouseX, g.mouseY)
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.vehicleWin.DragEnd(g.mouseX, g.mouseY)
		}
	}
	if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
		// Check dropdown first (it's on top: a window's dropdown lies over the
		// window that opened it)
		if g.dropdown != nil && g.dropdown.Visible {
			if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
				// Click outside dropdown - close it
				g.dropdown = nil
			}
		} else if g.colourDropdown != nil && g.colourDropdown.Visible {
			g.colourDropdown.HandleClick(g.mouseX, g.mouseY) // a colour picked, or closed
			g.colourDropdown = nil
		} else if g.companyFinances != nil && g.companyFinances.Contains(g.mouseX, g.mouseY) {
			g.companyFinances.HandleClick(g.mouseX, g.mouseY) // drawn on top: clicked first
		} else if g.companyList != nil && g.companyList.Contains(g.mouseX, g.mouseY) {
			g.companyList.HandleClick(g.mouseX, g.mouseY)
		} else if g.townWin != nil && g.townWin.Contains(g.mouseX, g.mouseY) { // the click goes to the window drawn on top (the draw order, reversed)
			g.townWin.HandleClick(g.mouseX, g.mouseY)
		} else if g.stationWin != nil && g.stationWin.Contains(g.mouseX, g.mouseY) {
			g.stationWin.HandleClick(g.mouseX, g.mouseY)
		} else if g.townList != nil && g.townList.Contains(g.mouseX, g.mouseY) {
			g.townList.HandleClick(g.mouseX, g.mouseY)
		} else if g.vehicleWin != nil && g.vehicleWin.Contains(g.mouseX, g.mouseY) {
			g.vehicleWin.HandleClick(g.mouseX, g.mouseY)
		} else if g.buildVehicle != nil && g.buildVehicle.Contains(g.mouseX, g.mouseY) {
			g.buildVehicle.HandleClick(g.mouseX, g.mouseY)
		} else if g.options != nil && g.options.Contains(g.mouseX, g.mouseY) {
			g.options.HandleClick(g.mouseX, g.mouseY)
		} else if g.browse != nil && g.browse.Contains(g.mouseX, g.mouseY) {
			g.browse.HandleClick(g.mouseX, g.mouseY)
		} else if g.scenarioSelect != nil && g.scenarioSelect.Contains(g.mouseX, g.mouseY) {
			g.scenarioSelect.HandleClick(g.mouseX, g.mouseY)
		} else if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
			if sp := uilayout.TimePanelSpeedAt(g.mouseX, g.mouseY, g.sw, g.sh); sp >= 0 && !g.inTitleScreen {
				// TimePanel speed buttons: pause toggles; the others set the speed.
				// clickable(TimePanel.cpp): pause_btn normal_speed_btn fast_forward_btn extra_fast_forward_btn
				switch sp {
				case 0:
					g.isPaused = !g.isPaused
				case 1:
					g.speedMult, g.isPaused = 1, false
				case 2:
					g.speedMult, g.isPaused = 2, false
				case 3:
					g.speedMult, g.isPaused = 4, false
				}
			} else if (uilayout.InTimePanel(g.mouseX, g.mouseY, g.sw, g.sh) || uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh)) && !g.inTitleScreen {
				// the bottom-bar windows swallow the click; the player panel's
				// company value is companyValueMouseUp: the Company window's Finances
				px, py := uilayout.PlayerPanelPos(g.sw, g.sh)
				if r := uilayout.PlayerPanelFace; uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh) &&
					g.mouseX >= px+r[0] && g.mouseX < px+r[0]+r[2] && g.mouseY >= py+r[1] && g.mouseY < py+r[1]+r[3] {
					g.playerCompanyDropdown(px+r[0], py+r[1], r[2], r[3]) // playerMouseDown
				}
				if r := uilayout.PlayerPanelCompanyVal; uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh) &&
					g.mouseX >= px+r[0] && g.mouseX < px+r[0]+r[2] && g.mouseY >= py+r[1] && g.mouseY < py+r[1]+r[3] {
					company, _ := g.w.ControllingCompany()
					g.openCompanyFinances(company)
				}
			} else if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
				g.toolbar.Buttons[btnIdx].Pressed = true
				g.handleToolbarButton(btnIdx)
			} else if g.pickupOn && !g.inTitleScreen {
				g.pickupToolDown(g.mouseX, g.mouseY) // the Vehicle window's placing tool
			} else if g.routeTool && g.vehicleWin != nil && !g.inTitleScreen {
				g.routeToolDown(g.mouseX, g.mouseY) // the Route tab's tool: a stop at the station clicked
			} else if (g.buildMode == buildModeTrack || g.buildMode == buildModeRoad || g.buildMode == buildModeDock || g.buildMode == buildModeAirport) && g.consTab == 1 && !g.inTitleScreen {
				g.stationToolUp(g.mouseX, g.mouseY) // the Station tab's tool: a station on the track or road clicked, or a dock
			} else if g.buildMode == buildModeTrack && g.hoverTileX >= 0 && g.r != nil && g.r.ObjMgr != nil {
				// onToolUpSingle + constructionLoop: the first click builds the
				// selected piece at the tile (the head set there), as the Construct
				// button would; refused, the error window shows and the rotation turns.
				g.removeTrackGhost()
				if p, ok := construct.GetTrackPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), g.w.ConstructionHeadState().Rotation); ok {
					c, _ := g.stationCmd()
					if _, built := g.w.BuildTrackAtTile(c, g.hoverTileX, g.hoverTileY, int(p.ID), uint8(g.buildObjID)); built {
						g.errorWin = nil // the loop's WindowManager::close(WindowType::error)
						if g.consOneShot >= 0 {
							g.consOneShot, g.consCurve = -1, 0
						}
					} else {
						g.openTrackBuildError(c)
					}
				}
			} else if g.buildMode == buildModeRoad && g.hoverTileX >= 0 {
				// Road mode: clicking the map sets the construction head. PARTIAL:
				// the first click building the road piece (as track's).
				g.w.SetConstructionHead(g.hoverTileX, g.hoverTileY)
			}
		}
	}
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) && g.dropdown != nil && g.dropdown.Visible {
		// the release of the press that opened a toolbar menu — after the press:
		// a quick click brings both in one frame (ui.DropdownMenu
		// HandleOpeningRelease): on its button, the highlighted item
		if g.dropdown.HandleOpeningRelease(g.mouseX, g.mouseY) {
			g.dropdown = nil
		}
	}
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) {
		for i := range g.toolbar.Buttons {
			g.toolbar.Buttons[i].Pressed = false
		}
		g.windowMgr.StopDrag()
	}

	// Right-click drag to pan the map.
	// OpenLoco reference: src/OpenLoco/src/Input/MouseInput.cpp stateViewportRight()
	// dragOffset is scaled by (1 << zoom) to convert screen pixels → world pixels.
	if g.inMouseJustPressed(ebiten.MouseButtonRight) {
		g.rightDragging = true
		g.dragLastX, g.dragLastY = g.mouseX, g.mouseY
	}
	if g.inMouseJustReleased(ebiten.MouseButtonRight) {
		g.rightDragging = false
	}
	if g.rightDragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		worldScale := float64(int(1) << g.w.GetZoom())
		dx := float64(g.mouseX-g.dragLastX) * worldScale
		dy := float64(g.mouseY-g.dragLastY) * worldScale
		g.w.PanCamera(-dx, -dy) // negative: drag right moves camera left
		g.dragLastX, g.dragLastY = g.mouseX, g.mouseY
	}

	// Camera panning: WASD or arrow keys (smooth, pixels per frame)
	const panSpeed = 4.0
	var panX, panY float64
	typing := g.browse != nil && g.browse.Save // the focused text box takes the keys
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		panY -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		panY += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		panX -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		panX += panSpeed
	}
	if typing {
		panX, panY = 0, 0
	}
	if panX != 0 || panY != 0 {
		g.w.PanCamera(panX, panY)
	}

	// Zoom: keyboard Q/E, +/-, and mouse scroll wheel
	if !typing && (g.inKeyJustPressed(ebiten.KeyQ) || g.inKeyJustPressed(ebiten.KeyEqual) || g.inKeyJustPressed(ebiten.KeyKPAdd)) {
		g.w.ZoomInAt(g.sw/2, g.sh/2)
	}
	if !typing && (g.inKeyJustPressed(ebiten.KeyE) || g.inKeyJustPressed(ebiten.KeyMinus) || g.inKeyJustPressed(ebiten.KeyKPSubtract)) {
		g.w.ZoomOutAt(g.sw/2, g.sh/2)
	}
	if _, dy := g.inWheel(); dy != 0 {
		if !g.routeWheelToWindows() {
			if dy > 0 {
				g.w.ZoomInAt(g.mouseX, g.mouseY)
			} else {
				g.w.ZoomOutAt(g.mouseX, g.mouseY)
			}
		}
	}

	// Keyboard.cpp edgeScroll: the cursor on the screen's very edge (x == 0,
	// x >= width − 1, likewise y) moves the main view by edgeScrollingSpeed
	// (Config default 12) << zoom — not on the title screen, not with Shift
	// or Ctrl held, only in the normal input state (not while dragging the
	// view or moving a window). It was a 20-px zone at 8 px: the view crept
	// while a toolbar button near the top was clicked.
	shiftCtrl := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyControl)
	positioning := g.vehicleWin != nil && g.vehicleWin.Positioning || g.companyFinances != nil && g.companyFinances.Positioning
	if ebiten.IsFocused() && !g.rightDragging && !g.noEdgeScrolling && !g.inTitleScreen && !shiftCtrl && !positioning {
		edgeSpeed := float64(int(edgeScrollingSpeed) << g.w.GetZoom())
		if g.mouseX == 0 {
			g.w.PanCamera(-edgeSpeed, 0)
		}
		if g.mouseX >= g.sw-1 {
			g.w.PanCamera(edgeSpeed, 0)
		}
		if g.mouseY == 0 {
			g.w.PanCamera(0, -edgeSpeed)
		}
		if g.mouseY >= g.sh-1 {
			g.w.PanCamera(0, edgeSpeed)
		}
	}

	// Speed controls: Space=pause, F3=×1, F4=×2, F5=×4, F6=×8
	if g.inKeyJustPressed(ebiten.KeySpace) {
		g.isPaused = !g.isPaused
	}
	if g.inKeyJustPressed(ebiten.KeyF3) {
		g.speedMult = 1
		g.isPaused = false
	}
	if g.inKeyJustPressed(ebiten.KeyF4) {
		g.speedMult = 2
		g.isPaused = false
	}
	if g.inKeyJustPressed(ebiten.KeyF5) {
		g.speedMult = 4
		g.isPaused = false
	}
	if g.inKeyJustPressed(ebiten.KeyF6) {
		g.speedMult = 8
		g.isPaused = false
	}
	if g.inKeyJustPressed(ebiten.KeyF12) {
		g.diagPending = true
		g.diagTileX = -1
		g.diagTileY = -1
	}
	// Update world (skipped when paused; repeated per speed multiplier)
	if n := g.w.GameTicks(g.mouseMoved()); !g.isPaused { // tickLogic(numUpdates × the speed)
		for i := 0; i < n*g.speedMult; i++ {
			g.w.Update()
			g.w.TickVehicles() // gameplay: the ported vehicle tick
			if g.inGame {
				g.checkObjective()
			}
		}
	}
	if g.saveMsgFrames > 0 {
		g.saveMsgFrames--
	}
	return nil
}
