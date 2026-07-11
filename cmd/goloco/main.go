package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/game"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/title"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	defaultWidth    = 800
	defaultHeight   = 600
	edgeScrollZone  = 20  // px from screen edge that triggers edge scroll
	edgeScrollSpeed = 8.0 // world-space pixels per frame at zoom 0
)

const (
	buildModeNone  = 0
	buildModeTrack = 1
	buildModeRoad  = 2
)

type Game struct {
	w                   *world.World
	titleWorld          *world.World // preserved title-screen world; restored on Quit to Menu
	r                   *render.Renderer
	toolbar             *ui.Toolbar
	windowMgr           *ui.SimpleWindowManager
	objMgr              *objects.ObjectManager
	audioMgr            *audio.Manager
	titleSeq            *title.Sequence
	gameState           *game.GameState
	dropdown            *ui.DropdownMenu
	mouseX              int
	mouseY              int
	dataDir             string
	inTitleScreen       bool
	soundEnabled        bool   // whether sound effects are on
	musicEnabled        bool   // whether music is on
	currentScenarioPath string // path of the currently loaded scenario/save
	saveMsg             string // transient "Game Saved!" status overlay
	saveMsgFrames       int    // frames remaining to show saveMsg

	// Current logical screen dimensions (updated by Layout each frame)
	sw, sh int

	// Right-click drag pan state
	rightDragging bool
	dragLastX     int
	dragLastY     int

	// Title screen globe animation counter (incremented each frame)
	titleFrame int

	// Speed controls
	isPaused  bool
	speedMult int // 1, 2, 4, or 8

	// Build mode (track/road placement)
	buildMode  int // buildModeNone / buildModeTrack / buildModeRoad
	consCurve  int // construction curve selection: -4 (left large) … 0 … +4
	consSlope  int // construction slope selection: -2 (steep down) … 0 … +2
	buildObjID int // index into TrackObjects / RoadObjects
	hoverTileX int // tile under cursor in build mode (-1 if invalid)
	hoverTileY int

	// Diagnostic crop: when set, save a 480×360 PNG centred on diagTileX/Y after the next Draw.
	// diagQuit=true means quit afterwards (CLI --diag mode); false keeps the game running (F12).
	diagPending bool
	diagQuit    bool
	diagTileX   int
	diagTileY   int

	// scripted input playback (GOLOCO_SCRIPT)
	script *inputScript

	// --shots mode: periodic full-frame captures (shot_NN.png), then quit.
	shotsEvery time.Duration // wall-clock interval between captures
	shotsLeft  int
	shotsLast  time.Time
	shotsN     int
	// quitAfterSave: exit cleanly once the map save has completed.
	quitAfterSave bool

	// Loading-screen state.  When loadingSpriteIDs != nil, the game is in the
	// sprite-preload phase and will draw the OpenLoco-style loading screen.
	loadingSpriteIDs []int  // full list collected from the world
	loadingStep      int    // next index to process
	loadingStepSize  int    // sprites per frame
	loadingCaption   string // text shown in the loading window caption
	loadingStyleFlip bool   // alternates between train style 0 and 1 each load
}

func findLocoDataDir() string {
	candidates := []string{
		"../locomotion/Data",
		"locomotion/Data",
		filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/common/Locomotion/Data"),
		filepath.Join(os.Getenv("HOME"), ".steam/steam/steamapps/common/Locomotion/Data"),
		"/usr/share/games/locomotion/Data",
	}

	for _, dir := range candidates {
		g1Path := filepath.Join(dir, "g1.DAT")
		if _, err := os.Stat(g1Path); err == nil {
			return dir
		}
	}
	return ""
}

func NewGame() *Game {
	log.Println("[Game] Creating new game...")

	// Initialize fonts
	log.Println("[Game] Loading fonts...")
	_, err := ui.InitFonts("../fonts")
	if err != nil {
		log.Printf("[Game] Failed to load fonts: %v (will use fallback)", err)
	}

	r := render.NewRenderer()

	// Initialize audio manager
	log.Println("[Game] Creating audio manager...")
	audioMgr := audio.NewManager()

	// Try to load real Locomotion G1.DAT sprites
	dataDir := findLocoDataDir()
	var objMgr *objects.ObjectManager

	if dataDir != "" {
		log.Printf("[Game] Found Locomotion data directory: %s", dataDir)
		g1Path := filepath.Join(dataDir, "g1.DAT")
		log.Printf("Loading G1 sprites from: %s", g1Path)

		g1, err := assets.LoadG1(g1Path)
		if err != nil {
			log.Printf("Failed to load G1: %v", err)
		} else {
			log.Printf("Loaded %d sprites from G1.DAT", g1.GetSpriteCount())
			r.G1 = g1

			// Authentic Locomotion bitmap text from the g1 glyph sprites;
			// overrides the TTF/embedded fallback set up by InitFonts.
			ui.InitG1Font(g1)

			// Populate the UI drawing palette from the loaded G1 so window
			// backgrounds, buttons and other widgets more closely match the
			// original Locomotion look-and-feel.
			//
			// Set the package-level global palette so DrawingContext and UI
			// widgets can use the game's palette everywhere. As a fallback,
			// also attempt to copy the palette into a temporary DrawingContext.
			if graphics.SetGlobalPalette(r.G1) {
				log.Printf("Applied G1 palette to UI drawing context")
			} else {
				// Fallback: still try to copy into a temporary DC for local use
				var tmpDC graphics.DrawingContext
				tmpDC.UseRendererPalette(r.G1)
			}
		}

		// Load objects from ObjData
		objDataDir := filepath.Join(filepath.Dir(dataDir), "ObjData")
		if _, err := os.Stat(objDataDir); err == nil {
			log.Printf("Loading objects from: %s", objDataDir)
			objMgr = objects.NewObjectManager(objDataDir)

			// Set base sprite index to after G1 sprites and attach G1 for dynamic loading
			if g1 != nil {
				objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
				objMgr.G1File = g1
				log.Printf("ObjectManager will use dynamic G1 sprite pool (starting at index %d)", g1.GetSpriteCount())
			}

			// Load InterfaceSkin first (INTERDEF.DAT)
			interfaceSkinPath := filepath.Join(objDataDir, "INTERDEF.DAT")
			if _, err := os.Stat(interfaceSkinPath); err == nil {
				log.Printf("Loading InterfaceSkin from: %s", interfaceSkinPath)
				if _, err := objMgr.LoadObject(interfaceSkinPath); err != nil {
					log.Printf("Failed to load InterfaceSkin: %v", err)
				} else if objMgr.InterfaceSkin != nil {
					log.Printf("Loaded InterfaceSkin: %s with %d sprites",
						objMgr.InterfaceSkin.DisplayName,
						objMgr.InterfaceSkin.ImageCount)
				}
			}

			if err := objMgr.LoadAllObjects(); err != nil {
				log.Printf("Failed to load objects: %v", err)
			} else {
				log.Printf("Loaded %d vehicles, %d land objects", len(objMgr.Vehicles), len(objMgr.LandObjects))
			}

			// Connect object manager to renderer for sprite access
			r.ObjMgr = objMgr
		}
	} else {
		log.Printf("Locomotion data directory not found")
	}

	// Fall back to placeholder tiles if G1 not loaded
	if r.G1 == nil {
		log.Printf("Using placeholder tiles...")
		if err := assets.GeneratePlaceholderAtlas("assets"); err != nil {
			log.Printf("Failed to generate placeholder: %v", err)
		} else {
			if at, err := render.LoadAtlasFromDir("assets/extracted"); err == nil {
				r.Atlas = at
				log.Printf("Loaded %d placeholder tiles", len(at.Images))
			}
		}
	}

	w := world.NewWorld(r)
	toolbar := ui.NewToolbar(defaultWidth)
	windowMgr := ui.NewSimpleWindowManager()

	// Try to load the title scenario
	if dataDir != "" {
		titlePath := filepath.Join(dataDir, "title.dat")
		log.Printf("[Game] Loading title scenario from: %s", titlePath)
		if sc, err := scenario.LoadScenarioData(titlePath); err == nil {
			// Load packed objects embedded in the scenario file first.
			// These take priority over ObjData files and provide scenario-
			// specific assets (buildings, vehicles, etc.) that may not exist
			// in the standard ObjData directory.
			if objMgr != nil && len(sc.PackedObjectsRaw) > 0 {
				packedOK := 0
				for _, po := range sc.PackedObjectsRaw {
					if _, err := objMgr.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data); err != nil {
						log.Printf("[Game] Failed to load packed object: %v", err)
					} else {
						packedOK++
					}
				}
				log.Printf("[Game] Loaded %d/%d packed objects from scenario", packedOK, len(sc.PackedObjectsRaw))
			}

			// Reorder objects to match the slot order declared in the
			// scenario's RequiredObjects chunk before loading tiles,
			// so that element indices resolve to the correct objects.
			if objMgr != nil {
				if len(sc.LandObjectOrder) > 0 {
					objMgr.ReorderLandObjects(sc.LandObjectOrder)
					log.Printf("[Game] Reordered land objects to match scenario slot order")
				}
				if len(sc.TreeObjectOrder) > 0 {
					objMgr.ReorderTreeObjects(sc.TreeObjectOrder)
					log.Printf("[Game] Reordered tree objects to match scenario slot order (%d slots)", len(sc.TreeObjectOrder))
				}
				if len(sc.BuildingObjectOrder) > 0 {
					objMgr.ReorderBuildingObjects(sc.BuildingObjectOrder)
					log.Printf("[Game] Reordered building objects to match scenario slot order (%d slots)", len(sc.BuildingObjectOrder))
				}
				if len(sc.WallObjectOrder) > 0 {
					objMgr.ReorderWallObjects(sc.WallObjectOrder)
					log.Printf("[Game] Reordered wall objects to match scenario slot order (%d slots)", len(sc.WallObjectOrder))
				}
				if len(sc.TrainSignalObjectOrder) > 0 {
					objMgr.ReorderTrainSignalObjects(sc.TrainSignalObjectOrder)
					log.Printf("[Game] Reordered train signal objects to match scenario slot order (%d slots)", len(sc.TrainSignalObjectOrder))
				}
				if len(sc.LevelCrossingObjectOrder) > 0 {
					objMgr.ReorderLevelCrossingObjects(sc.LevelCrossingObjectOrder)
					log.Printf("[Game] Reordered level crossing objects to match scenario slot order (%d slots)", len(sc.LevelCrossingObjectOrder))
				}
				if len(sc.BridgeObjectOrder) > 0 {
					objMgr.ReorderBridgeObjects(sc.BridgeObjectOrder)
					log.Printf("[Game] Reordered bridge objects to match scenario slot order (%d slots)", len(sc.BridgeObjectOrder))
				}
				if len(sc.TrainStationObjectOrder) > 0 {
					objMgr.ReorderTrainStationObjects(sc.TrainStationObjectOrder)
					log.Printf("[Game] Reordered train station objects to match scenario slot order (%d slots)", len(sc.TrainStationObjectOrder))
				}
				if len(sc.RoadStationObjectOrder) > 0 {
					objMgr.ReorderRoadStationObjects(sc.RoadStationObjectOrder)
					log.Printf("[Game] Reordered road station objects to match scenario slot order (%d slots)", len(sc.RoadStationObjectOrder))
				}
				if len(sc.TrackObjectOrder) > 0 {
					objMgr.ReorderTrackObjects(sc.TrackObjectOrder)
					log.Printf("[Game] Reordered track objects to match scenario slot order (%d slots)", len(sc.TrackObjectOrder))
				}
				if len(sc.RoadObjectOrder) > 0 {
					objMgr.ReorderRoadObjects(sc.RoadObjectOrder)
					log.Printf("[Game] Reordered road objects to match scenario slot order (%d slots)", len(sc.RoadObjectOrder))
				}
				if len(sc.VehicleObjectOrder) > 0 {
					objMgr.ReorderVehicleObjects(sc.VehicleObjectOrder)
					log.Printf("[Game] Reordered vehicle objects to match scenario slot order (%d slots)", len(sc.VehicleObjectOrder))
				}
			}
			w.LoadFromScenario(sc)
			log.Printf("[Game] Loaded title scenario: %dx%d map", sc.MapWidth, sc.MapHeight)
		} else {
			log.Printf("[Game] Could not load title scenario: %v", err)
		}
	}
	titleSpriteIDs := w.CollectSpriteIDs()
	log.Printf("[Game] Collected %d sprite IDs for preload", len(titleSpriteIDs))

	// Create title sequence
	log.Println("[Game] Creating title sequence...")
	titleSeq := title.NewSequence(audioMgr)

	// Set title screen camera to map center at full zoom (matching OpenLoco)
	// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Main.cpp:45-51
	//   Camera position = (kMapRows * kTileSize) / 2 - 1 = (384 * 32) / 2 - 1 = 6143
	//   Zoom = ZoomLevel::full (0)
	mapW, mapH := w.GetMapSize()
	if mapW > 0 && mapH > 0 {
		titleSeq.Start(mapW, mapH)
		w.SetZoom(0) // full zoom for title screen
	} else {
		titleSeq.Start(384, 384)
		w.SetZoom(0)
	}
	// Apply initial camera position from sequence
	{
		cx, cy := titleSeq.GetCameraPosition()
		w.SetCamera(cx, cy)
	}

	// Start playing title music
	if dataDir != "" {
		musicPath := filepath.Join(dataDir, "css5.dat")
		log.Printf("[Game] Loading music from: %s", musicPath)
		if err := audioMgr.LoadAndPlayMusic(musicPath, true); err != nil {
			log.Printf("[Game] WARNING: Could not load music: %v", err)
		} else {
			log.Println("[Game] Music loaded and playing!")
		}
	}

	log.Println("[Game] Game initialization complete")
	return &Game{
		w:             w,
		titleWorld:    w, // keep a reference so Quit to Menu can restore it
		r:             r,
		toolbar:       toolbar,
		windowMgr:     windowMgr,
		objMgr:        objMgr,
		audioMgr:      audioMgr,
		titleSeq:      titleSeq,
		dataDir:       dataDir,
		inTitleScreen: true,
		soundEnabled:  true,
		musicEnabled:  true,
		sw:            defaultWidth,
		sh:            defaultHeight,
		speedMult:     1,
		hoverTileX:    -1,
		hoverTileY:    -1,
		// Begin preloading title sprites on the first Draw() frame.
		loadingSpriteIDs: titleSpriteIDs,
		loadingStep:      0,
		loadingStepSize:  60,
		loadingCaption:   "title.dat",
	}
}

func (g *Game) Update() error {
	if g.quitAfterSave {
		return ebiten.Termination
	}
	if g.script != nil {
		g.script.step()
		if g.script.quit {
			return ebiten.Termination
		}
	}
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	if g.script != nil && g.script.hasMouse {
		g.mouseX, g.mouseY = g.script.mouseX, g.script.mouseY
	}

	if g.inTitleScreen {
		// Title screen: advance camera animation and apply to world.
		// w.Update() must be called here too — it increments animTick which
		// drives water wave animation and demo train movement.
		g.titleFrame++
		g.w.Update()
		if g.titleSeq != nil {
			g.titleSeq.Update()
			cx, cy := g.titleSeq.GetCameraPosition()
			g.w.SetCamera(cx, cy)
		}
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
		}
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			// Window manager takes priority (e.g. Load Game file browser)
			if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
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

	// Build mode: update hover tile every frame
	if g.buildMode != buildModeNone {
		g.hoverTileX, g.hoverTileY = g.w.ScreenToTile(g.mouseX, g.mouseY)
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

	// Update dropdown hover if visible
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	} else {
		g.toolbar.UpdateHover(g.mouseX, g.mouseY)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
	}
	if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
		// Check dropdown first (it's on top)
		if g.dropdown != nil && g.dropdown.Visible {
			if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
				// Click outside dropdown - close it
				g.dropdown = nil
			}
		} else if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
			if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
				g.toolbar.Buttons[btnIdx].Pressed = true
				g.handleToolbarButton(btnIdx)
			} else if g.buildMode != buildModeNone && g.hoverTileX >= 0 {
				// Track/road mode: clicking the map sets the construction head.
				g.w.SetConstructionHead(g.hoverTileX, g.hoverTileY)
			}
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
	if panX != 0 || panY != 0 {
		g.w.PanCamera(panX, panY)
	}

	// Zoom: keyboard Q/E, +/-, and mouse scroll wheel
	if g.inKeyJustPressed(ebiten.KeyQ) || g.inKeyJustPressed(ebiten.KeyEqual) || g.inKeyJustPressed(ebiten.KeyKPAdd) {
		g.w.ZoomInAt(g.sw/2, g.sh/2)
	}
	if g.inKeyJustPressed(ebiten.KeyE) || g.inKeyJustPressed(ebiten.KeyMinus) || g.inKeyJustPressed(ebiten.KeyKPSubtract) {
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

	// Edge scrolling: pan when cursor is within edgeScrollZone px of screen edge.
	// OpenLoco reference: src/OpenLoco/src/Input/Keyboard.cpp edgeScrolling()
	// Speed is scaled by zoom so scrolling feels consistent at all zoom levels.
	if ebiten.IsFocused() && !g.rightDragging {
		edgeSpeed := edgeScrollSpeed * float64(int(1)<<g.w.GetZoom())
		if g.mouseX <= edgeScrollZone {
			g.w.PanCamera(-edgeSpeed, 0)
		}
		if g.mouseX >= g.sw-edgeScrollZone {
			g.w.PanCamera(edgeSpeed, 0)
		}
		if g.mouseY <= edgeScrollZone {
			g.w.PanCamera(0, -edgeSpeed)
		}
		if g.mouseY >= g.sh-edgeScrollZone {
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
	if !g.isPaused {
		for i := 0; i < g.speedMult; i++ {
			g.w.Update()
			if g.gameState != nil {
				g.gameState.Update()
			}
		}
	}
	if g.saveMsgFrames > 0 {
		g.saveMsgFrames--
	}
	return nil
}

// Title screen layout constants — match OpenLoco's window positions.
const (
	// Main menu buttons (TitleMenu.cpp)
	titleBtnSize    = 74
	titleBtnCount   = 4
	titleMenuW      = titleBtnSize * titleBtnCount // 296
	titleMenuH      = titleBtnSize
	titleMenuMargin = 25 // pixels from bottom

	// Exit button (TitleExit.cpp) — bottom-right
	titleExitW = 60
	titleExitH = 28

	// Options button (TitleOptions.cpp) — top-right
	titleOptionsW = 60
	titleOptionsH = 15
)

func (g *Game) handleTitleMenuClick(mx, my int) {
	// --- Exit button: bottom-right corner ---
	exitX := g.sw - titleExitW
	exitY := g.sh - titleExitH
	if mx >= exitX && mx < g.sw && my >= exitY && my < g.sh {
		log.Println("[Game] Exit clicked")
		os.Exit(0)
	}

	// --- Options button: top-right corner ---
	optX := g.sw - titleOptionsW
	if mx >= optX && mx < g.sw && my >= 0 && my < titleOptionsH {
		g.openOptionsWindow()
		return
	}

	// --- Main menu buttons: bottom-centre ---
	menuX := (g.sw - titleMenuW) / 2
	menuY := g.sh - titleMenuH - titleMenuMargin
	relX := mx - menuX
	if relX < 0 || relX >= titleMenuW || my < menuY || my >= menuY+titleMenuH {
		return
	}
	btnIdx := relX / titleBtnSize

	switch btnIdx {
	case 0: // New Game — open scenario browser (SC5 only)
		g.openNewGameWindow()
	case 1:
		g.openLoadGameWindow()
	case 2:
		log.Println("[Game] Title menu: Tutorial (not yet implemented)")
	case 3:
		log.Println("[Game] Title menu: Scenario Editor (not yet implemented)")
	}
}

// handleToolbarButton handles clicks on toolbar buttons, matching the new
// OpenLoco-based button layout.
//
// Button order (indices):
//
//	0: Load/Save, 1: Audio, 2: (reserved/hidden)
//	3: Zoom, 4: Rotate, 5: View
//	6: Terraform, 7: Railroad, 8: Road, 9: Port/Airport, 10: Build Vehicles
//	11: Vehicles, 12: Stations, 13: Towns
func (g *Game) handleToolbarButton(idx int) {
	btn := &g.toolbar.Buttons[idx]
	tooltip := btn.Tooltip

	// Create appropriate dropdown or window based on button tooltip
	var win *ui.SimpleWindow
	switch tooltip {
	case "Load/Save":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp:117-154
		items := []ui.DropdownItem{
			{Text: "Load Game", Action: func() { g.openLoadGameWindow() }},
			{Text: "Save Game", Action: func() {
				if err := g.saveGame(); err != nil {
					log.Printf("[Game] Save error: %v", err)
				}
			}},
			{Separator: true},
			{Text: "Screenshot", Action: func() {
				g.diagPending = true
				g.diagQuit = false
				g.diagTileX = -1
				g.diagTileY = -1
			}},
			{Separator: true},
			{Text: "About", Action: func() { log.Println("[Game] About selected") }},
			{Text: "Options", Action: func() { g.openOptionsWindow() }},
			{Separator: true},
			{Text: "Quit to Menu", Action: func() { g.quitToMenu() }},
			{Text: "Quit to Desktop", Action: func() { os.Exit(0) }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Audio":
		items := []ui.DropdownItem{
			{Text: "Stop Music", Action: func() {
				if g.audioMgr != nil {
					g.audioMgr.StopMusic()
					log.Println("[Game] Music stopped")
				}
			}},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Zoom":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:95-132
		items := []ui.DropdownItem{
			{Text: "Zoom In", Action: func() { g.w.ZoomIn() }},
			{Text: "Zoom Out", Action: func() { g.w.ZoomOut() }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Rotate":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:135-143
		items := []ui.DropdownItem{
			{Text: "Rotate Clockwise", Action: func() { log.Println("[Game] Rotate CW") }},
			{Text: "Rotate Counter-CW", Action: func() { log.Println("[Game] Rotate CCW") }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "View":
		items := []ui.DropdownItem{
			{Text: "Underground View", Action: func() { log.Println("[Game] Underground view") }},
			{Text: "See Through Tracks", Action: func() { log.Println("[Game] See through tracks") }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return
	case "Terraform":
		log.Println("[Game] Terraform menu (not yet implemented)")
		return
	case "Railroad":
		g.buildMode = buildModeTrack
		g.buildObjID = 0
		g.consCurve = 0
		g.consSlope = 0
		win = g.newConstructionWindow()

	case "Road":
		g.buildMode = buildModeRoad
		g.buildObjID = 0
		g.consCurve = 0
		g.consSlope = 0
		win = g.newConstructionWindow()
	case "Port/Airport":
		log.Println("[Game] Port/Airport construction menu (not yet implemented)")
		return
	case "Build Vehicles":
		log.Println("[Game] Build vehicles menu (not yet implemented)")
		return
	case "Vehicles":
		win = ui.NewSimpleWindow("Available Vehicles", 50, 50, 400, 350)
		vehicles := g.objMgr
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			if vehicles == nil {
				ui.DrawText(screen, "No vehicles loaded", x+10, y+20, color.White)
				return
			}

			// Group by type
			trains := 0
			buses := 0
			trucks := 0
			aircraft := 0
			ships := 0
			for _, v := range vehicles.Vehicles {
				switch v.Type {
				case objects.VehicleTypeTrain:
					trains++
				case objects.VehicleTypeBus:
					buses++
				case objects.VehicleTypeTruck:
					trucks++
				case objects.VehicleTypeAircraft:
					aircraft++
				case objects.VehicleTypeShip:
					ships++
				}
			}

			ui.DrawText(screen, fmt.Sprintf("Total Vehicles: %d", len(vehicles.Vehicles)), x+10, y+20, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Trains:   %d", trains), x+10, y+40, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Buses:    %d", buses), x+10, y+55, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Trucks:   %d", trucks), x+10, y+70, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Aircraft: %d", aircraft), x+10, y+85, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Ships:    %d", ships), x+10, y+100, color.White)

			ui.DrawText(screen, "--- Sample Vehicles ---", x+10, y+125, color.White)

			// Show first 10 vehicles
			yOff := y + 140
			count := 0
			for _, v := range vehicles.Vehicles {
				if count >= 10 {
					break
				}
				name := v.DisplayName
				if name == "" {
					name = v.Header.GetName()
				}
				if len(name) > 28 {
					name = name[:28] + "..."
				}
				line := fmt.Sprintf("%-30s %3d km/h", name, int(v.GetSpeedKmh()))
				ui.DrawText(screen, line, x+10, yOff, color.White)
				yOff += 15
				count++
			}
		}
	case "Stations":
		win = ui.NewSimpleWindow("Stations", 150, 120, 350, 250)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Station list will go here", x+10, y+20, color.White)
		}
	case "Towns":
		win = ui.NewSimpleWindow("Towns", 150, 120, 280, 180)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Town list will go here", x+10, y+20, color.White)
		}
	default:
		// Generic window for any unhandled buttons
		win = ui.NewSimpleWindow(tooltip, 100+idx*20, 100+idx*10, 250, 150)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Coming soon...", x+10, y+20, color.White)
		}
	}

	if win != nil {
		g.windowMgr.OpenWindow(win)
	}
}

// drawLoadingScreen draws the OpenLoco-style progress window.
// The G1 progress bar sprites are:
//
//	2326-2329 style-0 train (green locomotive, 4 frames)
//	2330      track background (345×28)
//	2331-2334 style-1 train (darker locomotive, 4 frames)
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ProgressBar.cpp
func (g *Game) drawLoadingScreen(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Dark background
	screen.Fill(color.RGBA{0, 0, 0, 255})

	total := len(g.loadingSpriteIDs)
	if total == 0 {
		return
	}
	barValue := uint8(g.loadingStep * 255 / total) // 0-255 like OpenLoco

	// Window geometry (matches OpenLoco ProgressBar 350×47)
	const winW, winH = 350, 47
	winX := (sw - winW) / 2
	winY := (sh - winH) / 2

	// Window chrome: dark panel with single-pixel bevel
	fillRect(screen, float64(winX), float64(winY), float64(winW), float64(winH), color.RGBA{40, 40, 40, 255})
	fillRect(screen, float64(winX), float64(winY), float64(winW), 1, color.RGBA{120, 120, 120, 255})
	fillRect(screen, float64(winX), float64(winY), 1, float64(winH), color.RGBA{120, 120, 120, 255})
	fillRect(screen, float64(winX), float64(winY+winH-1), float64(winW), 1, color.RGBA{20, 20, 20, 255})
	fillRect(screen, float64(winX+winW-1), float64(winY), 1, float64(winH), color.RGBA{20, 20, 20, 255})
	// Caption bar background
	fillRect(screen, float64(winX+1), float64(winY+1), float64(winW-2), 13, color.RGBA{0, 0, 128, 255})

	// Caption bar at top (14px)
	captionText := "Loading: " + g.loadingCaption
	ui.DrawText(screen, captionText, winX+4, winY+3, color.RGBA{255, 255, 255, 255})

	// Clip region for progress bar: (winX+2, winY+15, 345, 28).
	// OpenLoco reference: Ui/Windows/ProgressBar.cpp draw() — the track is a
	// single full-width image at the clip origin and the train slides in from
	// the left (xPos = barValue - 255), both clipped to the bar region.
	clipX, clipY := winX+2, winY+15
	clip, _ := screen.SubImage(image.Rect(clipX, clipY, clipX+345, clipY+28)).(*ebiten.Image)
	if clip == nil {
		return
	}

	if track := g.r.GetSprite(2330); track != nil {
		_, _, xo, yo, _ := g.r.GetSpriteInfo(2330)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(clipX+int(xo)), float64(clipY+int(yo)))
		clip.DrawImage(track, op)
	}

	// Style 0: sea-green/dark-red steam train; style 1: black/grass-green.
	trainBase, c1, c2 := 2326, 11, 25 // mutedSeaGreen, mutedDarkRed
	if g.loadingStyleFlip {
		trainBase, c1, c2 = 2331, 0, 12 // black, mutedGrassGreen
	}
	trainFrame := int(barValue/4) % 4
	if train := g.r.GetSpriteColoured2(trainBase+trainFrame, c1, c2); train != nil {
		_, _, xo, yo, _ := g.r.GetSpriteInfo(trainBase + trainFrame)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(clipX+int(barValue)-255+int(xo)), float64(clipY+int(yo)))
		clip.DrawImage(train, op)
	}

	// Numeric progress below the window
	pct := g.loadingStep * 100 / total
	ui.DrawText(screen, fmt.Sprintf("%d%%", pct), winX+winW/2-10, winY+winH+4, color.RGBA{200, 200, 200, 200})
}

// doLoadingStep preloads up to loadingStepSize sprites per frame into the atlas.
// Returns true when all sprites have been packed and loading is complete.
func (g *Game) doLoadingStep() bool {
	size := g.loadingStepSize
	if size <= 0 {
		size = 50
	}
	end := g.loadingStep + size
	if end > len(g.loadingSpriteIDs) {
		end = len(g.loadingSpriteIDs)
	}
	g.r.PrewarmSpriteList(g.loadingSpriteIDs[g.loadingStep:end])
	g.loadingStep = end
	if g.loadingStep >= len(g.loadingSpriteIDs) {
		g.loadingSpriteIDs = nil // loading complete
		log.Printf("[Game] Sprite preload complete — atlas has %d sprites", g.r.AtlasPackedCount())
		return true
	}
	return false
}

func (g *Game) Draw(screen *ebiten.Image) {
	// ── Loading screen ──────────────────────────────────────────────────────
	// When loadingSpriteIDs is set, show OpenLoco-style progress window and
	// pack sprites into the atlas N per frame. Normal draw is skipped.
	if g.loadingSpriteIDs != nil {
		g.drawLoadingScreen(screen)
		g.doLoadingStep()
		g.maybeSaveShot(screen)
		return
	}

	if g.diagPending {
		diagBackground(screen)
	} else {
		screen.Fill(color.RGBA{0, 0, 0, 255})
	}

	// Full-map PNG save: skip normal game draw; show progress and render one chunk.
	if done, total := g.w.MapSaveProgress(); total > 0 {
		pct := float64(done) / float64(total)
		sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
		fillRect(screen, 0, 0, float64(sw), float64(sh), color.RGBA{20, 20, 30, 220})
		barX, barY, barW, barH := float64(sw/2-200), float64(sh/2-12), 400.0, 24.0
		fillRect(screen, barX, barY, barW, barH, color.RGBA{60, 60, 80, 255})
		fillRect(screen, barX, barY, barW*pct, barH, color.RGBA{80, 180, 80, 255})
		label := fmt.Sprintf("Saving full map... %d/%d (%.0f%%)", done, total, pct*100)
		ui.DrawText(screen, label, sw/2-100, sh/2-8, color.RGBA{255, 255, 255, 255})
		if g.w.StepMapSave() {
			if err := g.w.FinishMapSave(); err != nil {
				log.Printf("FinishMapSave: %v", err)
			}
			g.quitAfterSave = true
		}
		return
	}

	// If a diagnostic crop is pending for a specific tile, move the camera
	// here (after Update has run) so the title-sequence camera override
	// doesn't undo the positioning we set earlier.
	if g.diagPending && g.diagTileX >= 0 {
		g.w.SetCamera(float64(g.diagTileX), float64(g.diagTileY))
	}

	g.r.SetScreen(screen)
	g.w.Draw(screen)

	// Diagnostic crop: save a small PNG centred on the requested tile.
	if g.diagPending {
		g.diagPending = false
		saveDiagCrop(screen, g, g.diagTileX, g.diagTileY)
		world.LogBuildingStats()
		if g.diagQuit {
			g.quitAfterSave = true
		}
	}

	// Ghost preview for build mode (both chain from the construction head).
	if g.buildMode == buildModeTrack {
		g.w.DrawConstructionGhost(screen, g.consSelectedTrackID(), g.buildObjID)
	} else if g.buildMode == buildModeRoad {
		g.w.DrawRoadConstructionGhost(screen, g.consSelectedRoadID(), g.buildObjID)
	}

	if g.inTitleScreen {
		g.drawTitleMenu(screen)
		g.windowMgr.Draw(screen, g.r) // allow windows (e.g. Load Game) on title screen
		ui.DrawText(screen, fmt.Sprintf("%.0ffps", ebiten.ActualFPS()), 4, g.sh-8, color.RGBA{200, 200, 200, 200})
		g.maybeSaveShot(screen)
		return
	}

	// --- Gameplay HUD ---
	g.toolbar.Draw(screen, g.r)
	g.windowMgr.Draw(screen, g.r)

	// Draw dropdown menu on top of everything else
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.Draw(screen)
	}

	statusY := g.sh - 20
	fillRect(screen, 0, float64(statusY), float64(g.sw), 20, color.RGBA{50, 50, 50, 240})
	var statusText string
	speedStr := fmt.Sprintf("×%d", g.speedMult)
	if g.isPaused {
		speedStr = "PAUSED"
	}
	fps := ebiten.ActualFPS()
	fpsStr := fmt.Sprintf("%.0ffps", fps)
	if g.gameState != nil {
		date := g.gameState.GameDate
		statusText = fmt.Sprintf("%s %d  |  £%d  |  %s  |  Space=pause F3-F6=speed  |  %s",
			date.Month().String()[:3], date.Year(), g.gameState.PlayerMoney, speedStr, fpsStr)
	} else {
		statusText = fmt.Sprintf("GoLoco | %s | Scroll to zoom | %s", speedStr, fpsStr)
	}
	ui.DrawText(screen, statusText, 4, statusY+14, color.White)

	// Transient save confirmation overlay
	if g.saveMsgFrames > 0 {
		alpha := uint8(255)
		if g.saveMsgFrames < 30 {
			alpha = uint8(g.saveMsgFrames * 8)
		}
		msgW, _ := ui.MeasureText(g.saveMsg)
		mx := (g.sw - msgW) / 2
		fillRect(screen, float64(mx-8), float64(statusY-30), float64(msgW+16), 24,
			color.RGBA{0, 0, 0, alpha / 2})
		ui.DrawText(screen, g.saveMsg, mx, statusY-12, color.RGBA{100, 255, 100, alpha})
	}

	g.maybeSaveShot(screen)
}

// maybeSaveShot implements --shots mode: save the full frame every shotsEvery,
// quit when the requested count has been captured. Script `shot` commands are
// honoured here too.
func (g *Game) maybeSaveShot(screen *ebiten.Image) {
	if name := g.scriptShotName(); name != "" {
		if f, err := os.Create(name); err == nil {
			png.Encode(f, screen.SubImage(screen.Bounds()))
			f.Close()
			log.Printf("[Script] saved %s", name)
		}
	}
	if g.shotsEvery <= 0 || g.shotsLeft <= 0 || time.Since(g.shotsLast) < g.shotsEvery {
		return
	}
	g.shotsLast = time.Now()
	g.shotsN++
	name := fmt.Sprintf("shot_%02d.png", g.shotsN)
	if f, err := os.Create(name); err == nil {
		if err := png.Encode(f, screen.SubImage(screen.Bounds())); err != nil {
			log.Printf("[Shots] encode %s: %v", name, err)
		}
		f.Close()
		log.Printf("[Shots] saved %s (%d remaining)", name, g.shotsLeft)
	} else {
		log.Printf("[Shots] create %s: %v", name, err)
	}
	g.shotsLeft--
	if g.shotsLeft == 0 {
		g.quitAfterSave = true
	}
}

// iclamp clamps v to [lo, hi].
func iclamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// fillRect is a thin wrapper around vector.FillRect using float64 args (no anti-aliasing).
// Replaces the deprecated ebitenutil.DrawRect.
func fillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

// drawButton renders a Windows 95-style 3D beveled button with a centered text label.
//
//   - pressed=true inverts the bevel so the button looks inset/active (e.g. toggle ON state).
//   - enabled=false renders the button greyed out.
//
// Colors match the overall Locomotion beige palette; when a global palette is
// loaded the nearest indices are selected automatically via FillRect chains.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/SoftwareDrawingContext.cpp drawRect3d
func drawButton(screen *ebiten.Image, x, y, w, h int, label string, pressed, enabled bool) {
	var bgC, lightC, darkC, textC color.RGBA
	if !enabled {
		bgC = color.RGBA{168, 163, 148, 255}
		lightC = color.RGBA{194, 190, 175, 255}
		darkC = color.RGBA{118, 114, 102, 255}
		textC = color.RGBA{108, 104, 94, 255}
	} else if pressed {
		// Invert bevel: dark on top/left, light on bottom/right
		bgC = color.RGBA{190, 185, 168, 255}
		lightC = color.RGBA{110, 107, 96, 255}
		darkC = color.RGBA{232, 229, 214, 255}
		textC = color.RGBA{20, 15, 10, 255}
	} else {
		bgC = color.RGBA{205, 200, 183, 255}
		lightC = color.RGBA{238, 235, 220, 255}
		darkC = color.RGBA{112, 108, 97, 255}
		textC = color.RGBA{20, 15, 10, 255}
	}
	fillRect(screen, float64(x), float64(y), float64(w), float64(h), bgC)
	// Outer bevel
	fillRect(screen, float64(x), float64(y), float64(w), 1, lightC)
	fillRect(screen, float64(x), float64(y), 1, float64(h), lightC)
	fillRect(screen, float64(x), float64(y+h-1), float64(w), 1, darkC)
	fillRect(screen, float64(x+w-1), float64(y), 1, float64(h), darkC)
	// Inner bevel
	if w > 3 && h > 3 {
		fillRect(screen, float64(x+1), float64(y+1), float64(w-2), 1, lightC)
		fillRect(screen, float64(x+1), float64(y+1), 1, float64(h-2), lightC)
		fillRect(screen, float64(x+1), float64(y+h-2), float64(w-2), 1, darkC)
		fillRect(screen, float64(x+w-2), float64(y+1), 1, float64(h-2), darkC)
	}
	lw, _ := ui.MeasureText(label)
	lx := x + (w-lw)/2
	ly := y + (h-10)/2 // top-anchored 10px bitmap glyphs, vertically centred
	if pressed {
		lx++
		ly++
	}
	ui.DrawText(screen, label, lx, ly, textC)
}

// ColourNone means no palette remap — draw with the base palette as-is.
const ColourNone = -1

// OpenLoco Colour enum indices (Colour.h).
const (
	ColourBlack           = 0
	ColourGrey            = 1
	ColourWhite           = 2
	ColourMutedDarkPurple = 3
	ColourMutedPurple     = 4
	ColourPurple          = 5
	ColourDarkBlue        = 6
	ColourBlue            = 7
)

// drawUISprite draws a G1 sprite at (x,y) in screen coordinates.
// colour is an OpenLoco Colour index (0=black, 1=grey, …) or ColourNone for no remap.
// The sprite's G1 xOffset/yOffset are applied, matching OpenLoco's drawImage().
//
// OpenLoco reference: src/OpenLoco/src/Graphics/SoftwareDrawingContext.cpp drawImage()
//
//	→ PaletteMap::getForColour() → G1[2170+colour] 256-byte remap table
func (g *Game) drawUISprite(screen *ebiten.Image, spriteID, x, y, colour int) {
	var img *ebiten.Image
	if colour == ColourNone {
		img = g.r.GetSprite(spriteID)
	} else {
		img = g.r.GetSpriteColoured(spriteID, colour)
	}
	if img == nil {
		return
	}
	_, _, xOff, yOff, ok := g.r.GetSpriteInfo(spriteID)
	if !ok {
		xOff, yOff = 0, 0
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x+int(xOff)), float64(y+int(yOff)))
	screen.DrawImage(img, opts)
}

func (g *Game) drawTitleMenu(screen *ebiten.Image) {
	// Sprite ID constants.
	// OpenLoco reference: src/OpenLoco/src/Graphics/ImageIds.h
	const (
		spriteGlobeSpinBase  = 3552 // title_menu_globe_spin_0..31 (sequential)
		spriteGlobeConstBase = 3584 // title_menu_globe_construct_0..31 (sequential)
		spriteGlobeConstIdle = 3608 // globe_construct_24 — default editor frame
		spriteSparkle        = 3547 // title_menu_sparkle (New Game overlay)
		spriteSave           = 3548 // title_menu_save (Load Game overlay)
		spriteLessonL        = 3549 // title_menu_lesson_l (Tutorial overlay)
	)

	// --- Logo: the original Locomotion logo sprite, like OpenLoco's TitleLogo
	// window (298×170 at the top-left). Requires the disc-aligned G1 layout.
	// OpenLoco reference: src/OpenLoco/src/Ui/Windows/TitleLogo.cpp,
	// ImageIds::locomotion_logo = 3624.
	const spriteLocomotionLogo = 3624
	var subY int
	if logo := g.r.GetSprite(spriteLocomotionLogo); logo != nil {
		_, lh, xOff, yOff, _ := g.r.GetSpriteInfo(spriteLocomotionLogo)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(xOff), float64(yOff))
		screen.DrawImage(logo, op)
		subY = int(yOff) + int(lh) + 6
	} else {
		// Fallback branding when the logo sprite is unavailable.
		const logoScale = 4.0
		logoText := "GoLoco"
		logoW, _ := ui.MeasureTextBold(logoText)
		logoX := (298 - int(float64(logoW)*logoScale)) / 2
		logoY := 72
		shadow := color.RGBA{0, 0, 0, 220}
		for _, d := range [][2]int{{-4, -4}, {0, -4}, {4, -4}, {-4, 0}, {4, 0}, {-4, 4}, {0, 4}, {4, 4}} {
			ui.DrawTextBoldScaled(screen, logoText, logoX+d[0], logoY+d[1], logoScale, shadow)
		}
		ui.DrawTextBoldScaled(screen, logoText, logoX, logoY, logoScale, color.RGBA{220, 220, 255, 255})
		subY = logoY + int(9.0*logoScale) + 8
	}

	subtitle := "GoLoco - A Locomotion Reimplementation"
	subW, _ := ui.MeasureText(subtitle)
	subX := (298 - subW) / 2
	for _, d := range [][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		ui.DrawText(screen, subtitle, subX+d[0], subY+d[1], color.RGBA{0, 0, 0, 160})
	}
	ui.DrawText(screen, subtitle, subX, subY, color.RGBA{200, 200, 230, 255})

	// --- Options button: top-right (sw-60, 0), 60×15 ---
	// OpenLoco: TitleOptions.cpp at (Ui::width()-60, 0)
	optX := g.sw - titleOptionsW
	optHovered := g.mouseX >= optX && g.mouseX < g.sw &&
		g.mouseY >= 0 && g.mouseY < titleOptionsH
	// OpenLoco's TitleOptions window is transparent — plain white text with a
	// subtle darkening only on hover (the solid coloured boxes read as
	// misrendered rectangles).
	if optHovered {
		fillRect(screen, float64(optX), 0, float64(titleOptionsW), float64(titleOptionsH), color.RGBA{0, 0, 0, 90})
	}
	optLabel := "Options"
	optLabelW, _ := ui.MeasureText(optLabel)
	drawOutlinedText(screen, optLabel, optX+(titleOptionsW-optLabelW)/2, (titleOptionsH-10)/2, color.White)

	// --- Main menu buttons: horizontally centred, 25px from bottom ---
	// OpenLoco: TitleMenu.cpp — kBtnMainSize=74, kWW=296,
	//   window at ((Ui::width()-296)/2, Ui::height()-92-25)
	// Globe animation: (var_846/2) % 32, var_846 incremented each frame.
	animFrame := (g.titleFrame / 2) % 32

	type btnDef struct {
		globeBase  int // base sprite ID for animation frames
		idleGlobe  int // idle (non-hover) sprite ID
		overlay    int // overlay sprite ID (-1 = none)
		label      string
		frameCount int // number of animation frames in the sequence
	}
	btns := [titleBtnCount]btnDef{
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSparkle, "New Game", 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSave, "Load Game", 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteLessonL, "Tutorial", 32},
		// globe_construct has 25 frames (0–24 = sprites 3584–3608); clamping to 25
		// prevents overshooting into unrelated sprites (e.g. the Sawyer logo).
		{spriteGlobeConstBase, spriteGlobeConstIdle, -1, "Scenario Editor", 25},
	}

	menuX := (g.sw - titleMenuW) / 2
	menuY := g.sh - titleMenuH - titleMenuMargin

	for i, btn := range btns {
		bx := menuX + i*titleBtnSize
		by := menuY

		hovered := g.mouseX >= bx && g.mouseX < bx+titleBtnSize &&
			g.mouseY >= by && g.mouseY < by+titleBtnSize

		globeID := btn.idleGlobe
		if hovered {
			globeID = btn.globeBase + animFrame%btn.frameCount
		}

		if g.r.GetSprite(globeID) != nil {
			// Draw globe base then overlay (OpenLoco: drawImage x2 per button)
			g.drawUISprite(screen, globeID, bx, by, ColourNone)
			if btn.overlay >= 0 {
				g.drawUISprite(screen, btn.overlay, bx, by, ColourNone)
			}
		} else {
			// Fallback: coloured rectangle + text label
			var bgColor color.RGBA
			if hovered {
				bgColor = color.RGBA{80, 120, 80, 220}
			} else {
				bgColor = color.RGBA{50, 90, 50, 200}
			}
			fillRect(screen, float64(bx), float64(by), float64(titleBtnSize), float64(titleBtnSize), bgColor)
			light := color.RGBA{100, 160, 100, 255}
			dark := color.RGBA{30, 60, 30, 255}
			fillRect(screen, float64(bx), float64(by), float64(titleBtnSize), 2, light)
			fillRect(screen, float64(bx), float64(by), 2, float64(titleBtnSize), light)
			fillRect(screen, float64(bx), float64(by+titleBtnSize-2), float64(titleBtnSize), 2, dark)
			fillRect(screen, float64(bx+titleBtnSize-2), float64(by), 2, float64(titleBtnSize), dark)
			labelW, _ := ui.MeasureText(btn.label)
			labelX := bx + (titleBtnSize-labelW)/2
			labelY := by + titleBtnSize/2 + 4
			ui.DrawText(screen, btn.label, labelX, labelY, color.White)
		}
	}

	// --- Exit button: bottom-right (sw-40, sh-28), 40×28 ---
	// OpenLoco: TitleExit.cpp at (Ui::width()-40, Ui::height()-28)
	exitX := g.sw - titleExitW
	exitY := g.sh - titleExitH
	exitHovered := g.mouseX >= exitX && g.mouseX < g.sw &&
		g.mouseY >= exitY && g.mouseY < g.sh
	// Transparent like OpenLoco's TitleExit window; hover darkens.
	if exitHovered {
		fillRect(screen, float64(exitX), float64(exitY), float64(titleExitW), float64(titleExitH), color.RGBA{0, 0, 0, 90})
	}
	exitLabel := "Exit Game"
	exitLabelW, _ := ui.MeasureText(exitLabel)
	drawOutlinedText(screen, exitLabel, exitX+(titleExitW-exitLabelW)/2, exitY+(titleExitH-10)/2, color.White)

	// --- Version text: bottom-left ---
	ui.DrawText(screen, "GoLoco v0.1", 8, g.sh-20, color.RGBA{200, 200, 200, 255})
}

// drawOutlinedText draws white-ish text with a 1px dark outline so it stays
// readable over the moving title-screen world (like OpenLoco's title labels).
func drawOutlinedText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	shadow := color.RGBA{0, 0, 0, 200}
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		ui.DrawText(screen, str, x+d[0], y+d[1], shadow)
	}
	ui.DrawText(screen, str, x, y, clr)
}

// --- Load / Save -------------------------------------------------------

// GoLocoSave is the JSON format for GoLoco native save files (.goloco).
type GoLocoSave struct {
	Version      int    `json:"version"`
	ScenarioPath string `json:"scenarioPath"`
}

// goLocoSavesDir returns (and creates) the directory for .goloco save files.
func goLocoSavesDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".goloco", "saves")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// quitToMenu restores the title screen world and restarts the title sequence.
func (g *Game) quitToMenu() {
	if g.titleWorld != nil {
		g.w = g.titleWorld
	}
	g.inTitleScreen = true
	g.currentScenarioPath = ""
	if g.titleSeq != nil {
		g.titleSeq.Restart()
	}
}

// loadScenario loads a scenario or GoLoco save file into the world.
// Supports .SC5, .SV5 (Locomotion native) and .goloco (GoLoco JSON saves).
func (g *Game) loadScenario(filePath string) error {
	// If it is a GoLoco JSON save, extract the referenced scenario path.
	if strings.EqualFold(filepath.Ext(filePath), ".goloco") {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading save file: %w", err)
		}
		var save GoLocoSave
		if err := json.Unmarshal(data, &save); err != nil {
			return fmt.Errorf("parsing save file: %w", err)
		}
		filePath = save.ScenarioPath
	}

	sc, err := scenario.LoadScenarioData(filePath)
	if err != nil {
		return fmt.Errorf("loading scenario: %w", err)
	}

	if g.objMgr != nil {
		if len(sc.LandObjectOrder) > 0 {
			g.objMgr.ReorderLandObjects(sc.LandObjectOrder)
		}
		if len(sc.TreeObjectOrder) > 0 {
			g.objMgr.ReorderTreeObjects(sc.TreeObjectOrder)
		}
		if len(sc.BuildingObjectOrder) > 0 {
			g.objMgr.ReorderBuildingObjects(sc.BuildingObjectOrder)
		}
		if len(sc.WallObjectOrder) > 0 {
			g.objMgr.ReorderWallObjects(sc.WallObjectOrder)
		}
		if len(sc.TrainSignalObjectOrder) > 0 {
			g.objMgr.ReorderTrainSignalObjects(sc.TrainSignalObjectOrder)
		}
		if len(sc.LevelCrossingObjectOrder) > 0 {
			g.objMgr.ReorderLevelCrossingObjects(sc.LevelCrossingObjectOrder)
		}
		if len(sc.BridgeObjectOrder) > 0 {
			g.objMgr.ReorderBridgeObjects(sc.BridgeObjectOrder)
		}
		if len(sc.TrainStationObjectOrder) > 0 {
			g.objMgr.ReorderTrainStationObjects(sc.TrainStationObjectOrder)
		}
		if len(sc.RoadStationObjectOrder) > 0 {
			g.objMgr.ReorderRoadStationObjects(sc.RoadStationObjectOrder)
		}
		if len(sc.TrackObjectOrder) > 0 {
			g.objMgr.ReorderTrackObjects(sc.TrackObjectOrder)
		}
		if len(sc.RoadObjectOrder) > 0 {
			g.objMgr.ReorderRoadObjects(sc.RoadObjectOrder)
		}
		if len(sc.VehicleObjectOrder) > 0 {
			g.objMgr.ReorderVehicleObjects(sc.VehicleObjectOrder)
		}
	}

	// Create a fresh world for the new scenario so the title world is preserved.
	gameW := world.NewWorld(g.r)
	gameW.LoadFromScenario(sc)
	g.w = gameW
	g.currentScenarioPath = filePath
	g.inTitleScreen = false

	// Kick off sprite preload for the new scenario's sprites.
	g.loadingSpriteIDs = gameW.CollectSpriteIDs()
	g.loadingStep = 0
	g.loadingStepSize = 60
	g.loadingCaption = filepath.Base(filePath)
	g.loadingStyleFlip = !g.loadingStyleFlip
	log.Printf("[Game] Queued %d sprites for preload from %s", len(g.loadingSpriteIDs), filePath)

	mapW, mapH := g.w.GetMapSize()
	g.gameState = game.NewGameState(mapW, mapH)

	if g.titleSeq != nil {
		g.titleSeq.Stop()
	}
	g.windowMgr.CloseWindow("Load Game")
	g.windowMgr.CloseWindow("New Game")
	log.Printf("[Game] Loaded scenario: %s (%dx%d)", filePath, sc.MapWidth, sc.MapHeight)
	return nil
}

// saveGame writes the current scenario path to a .goloco JSON save file.
func (g *Game) saveGame() error {
	if g.currentScenarioPath == "" {
		g.saveMsg = "Nothing to save"
		g.saveMsgFrames = 120
		return nil
	}
	savesDir := goLocoSavesDir()
	filename := fmt.Sprintf("save_%d.goloco", time.Now().Unix())
	path := filepath.Join(savesDir, filename)

	data, err := json.Marshal(GoLocoSave{Version: 1, ScenarioPath: g.currentScenarioPath})
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	g.saveMsg = "Game saved!"
	g.saveMsgFrames = 120
	log.Printf("[Game] Saved to: %s", path)
	return nil
}

// routeWheelToWindows sends mouse-wheel events to any visible window under the
// cursor that has an OnScrollWheel handler.  Returns true if consumed.
func (g *Game) routeWheelToWindows() bool {
	_, dy := g.inWheel()
	if dy == 0 {
		return false
	}
	for i := len(g.windowMgr.Windows) - 1; i >= 0; i-- {
		w := g.windowMgr.Windows[i]
		if w.Visible && w.OnScrollWheel != nil &&
			g.mouseX >= w.X && g.mouseX < w.X+w.Width &&
			g.mouseY >= w.Y && g.mouseY < w.Y+w.Height {
			w.OnScrollWheel(dy)
			return true
		}
	}
	return false
}

// openNewGameWindow opens the OpenLoco-style scenario selection screen.
func (g *Game) openNewGameWindow() {
	const winTitle = "New Game"
	const winW, winH = 620, 440

	// Scan for scenarios once when the window opens.
	metas, _ := scenario.ScanScenarios(g.dataDir + "/../Scenarios")
	if len(metas) == 0 {
		// Fallback: scan the data dir itself for .SC5 files.
		metas, _ = scenario.ScanScenarios(g.dataDir)
	}

	// Group by category.
	type catGroup struct {
		label string
		items []*scenario.ScenarioMeta
	}
	cats := []catGroup{
		{label: "Beginner"},
		{label: "Easy"},
		{label: "Medium"},
		{label: "Challenging"},
		{label: "Expert"},
	}
	for _, m := range metas {
		idx := int(m.Category)
		if idx < 0 || idx >= len(cats) {
			idx = 0
		}
		cats[idx].items = append(cats[idx].items, m)
	}

	// Find first non-empty category.
	activeCat := 0
	for i, c := range cats {
		if len(c.items) > 0 {
			activeCat = i
			break
		}
	}

	selectedIdx := -1
	listScroll := 0

	// Cached preview image for the selected scenario.
	var previewImg *ebiten.Image
	var previewFor string // file path the image was built for

	getPreview := func(m *scenario.ScenarioMeta) *ebiten.Image {
		if m == nil || m.IsGenerated {
			return nil
		}
		if previewFor == m.FilePath {
			return previewImg
		}
		// Convert 128×128 palette-indexed preview → RGBA using the global palette.
		img := ebiten.NewImage(scenario.PreviewSize, scenario.PreviewSize)
		if graphics.IsGlobalPaletteLoaded() {
			for py := 0; py < scenario.PreviewSize; py++ {
				for px := 0; px < scenario.PreviewSize; px++ {
					idx := m.Preview[py*scenario.PreviewSize+px]
					c := graphics.GetGlobalColor(idx)
					img.Set(px, py, c)
				}
			}
		}
		previewImg = img
		previewFor = m.FilePath
		return previewImg
	}

	win := ui.NewSimpleWindow(winTitle, 60, 30, winW, winH)

	// Layout constants (content area starts after title bar + border).
	const (
		tabH      = 24 // category tab height
		listW     = 220
		divW      = 4
		infoX     = listW + divW
		itemH     = 20
		btnAreaH  = 32
		previewSz = scenario.PreviewSize // 128
	)

	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, r *render.Renderer) {
		// ── Category tabs ────────────────────────────────────────────────
		tabW := cw / len(cats)
		for i, cat := range cats {
			tx := cx + i*tabW
			active := i == activeCat
			drawButton(screen, tx, cy, tabW-2, tabH-2, cat.label, active, len(cat.items) > 0)
		}

		contentY := cy + tabH
		contentH := ch - tabH - btnAreaH
		items := cats[activeCat].items
		visible := contentH / itemH

		// ── Scenario list (left panel) ───────────────────────────────────
		fillRect(screen, float64(cx), float64(contentY), float64(listW), float64(contentH), color.RGBA{210, 205, 185, 255})
		// Clamp scroll
		maxScroll := len(items) - visible
		if maxScroll < 0 {
			maxScroll = 0
		}
		if listScroll > maxScroll {
			listScroll = maxScroll
		}
		if listScroll < 0 {
			listScroll = 0
		}

		for i := 0; i < visible; i++ {
			idx := listScroll + i
			if idx >= len(items) {
				break
			}
			m := items[idx]
			iy := contentY + i*itemH
			rowBg := color.RGBA{210, 205, 185, 255}
			textCol := color.RGBA{20, 20, 20, 255}
			if idx == selectedIdx {
				rowBg = color.RGBA{60, 100, 180, 255}
				textCol = color.RGBA{255, 255, 255, 255}
			}
			fillRect(screen, float64(cx), float64(iy), float64(listW), float64(itemH), rowBg)
			label := m.Name
			if m.IsGenerated {
				label = "~ " + label
			}
			ui.DrawText(screen, label, cx+4, iy+(itemH-10)/2, textCol)
		}

		// Divider
		fillRect(screen, float64(cx+listW), float64(contentY), float64(divW), float64(contentH), color.RGBA{120, 115, 100, 255})

		// ── Info panel (right panel) ─────────────────────────────────────
		infoW := cw - infoX
		fillRect(screen, float64(cx+infoX), float64(contentY), float64(infoW), float64(contentH), color.RGBA{225, 220, 200, 255})

		if selectedIdx >= 0 && selectedIdx < len(items) {
			m := items[selectedIdx]
			ix := cx + infoX + 6
			iy := contentY + 6

			// Preview image or "Randomly Generated" text
			if !m.IsGenerated {
				if img := getPreview(m); img != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(float64(ix), float64(iy))
					screen.DrawImage(img, op)
					iy += previewSz + 6
				}
			} else {
				fillRect(screen, float64(ix), float64(iy), float64(previewSz), float64(previewSz), color.RGBA{100, 130, 100, 255})
				rw, _ := ui.MeasureText("Randomly")
				gw, _ := ui.MeasureText("Generated")
				ui.DrawText(screen, "Randomly", ix+(previewSz-rw)/2, iy+previewSz/2-12, color.White)
				ui.DrawText(screen, "Generated", ix+(previewSz-gw)/2, iy+previewSz/2+2, color.White)
				iy += previewSz + 6
			}

			// Scenario name in the bold bitmap font
			ui.DrawTextBold(screen, m.Name, ix, iy, color.RGBA{20, 20, 20, 255})
			iy += 14

			// Start year
			if m.StartYear > 0 {
				ui.DrawText(screen, fmt.Sprintf("Start: %d", m.StartYear), ix, iy, color.RGBA{60, 60, 60, 255})
				iy += 12
			}
			// Competitors
			if m.MaxCompetitors > 0 {
				ui.DrawText(screen, fmt.Sprintf("Competitors: %d", m.MaxCompetitors), ix, iy, color.RGBA{60, 60, 60, 255})
				iy += 12
			}

			// Description (word-wrapped by character count)
			if m.Description != "" {
				descMaxW := infoW - 12
				charWidth := 6
				charsPerLine := descMaxW / charWidth
				desc := m.Description
				for len(desc) > 0 && iy < contentY+contentH-10 {
					line := desc
					if len(line) > charsPerLine {
						cut := charsPerLine
						for cut > 0 && desc[cut] != ' ' {
							cut--
						}
						if cut == 0 {
							cut = charsPerLine
						}
						line = desc[:cut]
						desc = strings.TrimLeft(desc[cut:], " ")
					} else {
						desc = ""
					}
					ui.DrawText(screen, line, ix, iy, color.RGBA{50, 50, 50, 255})
					iy += 12
				}
			}
		} else if len(items) == 0 {
			ui.DrawText(screen, "No scenarios found", cx+infoX+10, contentY+30, color.RGBA{100, 100, 100, 255})
		}

		// ── Bottom buttons ───────────────────────────────────────────────
		btnY := cy + ch - btnAreaH + 4
		canPlay := selectedIdx >= 0 && selectedIdx < len(items)
		drawButton(screen, cx+4, btnY, 80, 22, "Cancel", false, true)
		drawButton(screen, cx+cw-84, btnY, 80, 22, "Play", false, canPlay)
	}

	win.OnContentClick = func(relX, relY int) {
		// Tab clicks
		tabW := win.Width / len(cats)
		if relY < tabH {
			clicked := relX / tabW
			if clicked >= 0 && clicked < len(cats) && clicked != activeCat {
				activeCat = clicked
				selectedIdx = -1
				listScroll = 0
				previewFor = ""
			}
			return
		}

		// Button row
		if relY >= win.Height-22-3-btnAreaH {
			if relX >= 4 && relX < 84 {
				g.windowMgr.CloseWindow(winTitle)
			} else if relX >= win.Width-6-84 {
				items := cats[activeCat].items
				if selectedIdx >= 0 && selectedIdx < len(items) {
					if err := g.loadScenario(items[selectedIdx].FilePath); err != nil {
						log.Printf("[Game] Load error: %v", err)
					}
				}
			}
			return
		}

		contentY := tabH
		row := (relY - contentY) / itemH
		if row < 0 {
			return
		}
		if relX < listW {
			items := cats[activeCat].items
			idx := listScroll + row
			if idx >= 0 && idx < len(items) {
				selectedIdx = idx
				previewFor = ""
			}
		}
	}

	win.OnScrollWheel = func(dy float64) {
		items := cats[activeCat].items
		contentH := win.Height - 22 - 3 - tabH - btnAreaH
		visible := contentH / itemH
		maxScroll := len(items) - visible
		if maxScroll < 0 {
			maxScroll = 0
		}
		listScroll = iclamp(listScroll-int(dy), 0, maxScroll)
	}

	g.windowMgr.OpenWindow(win)
}

// openOptionsWindow opens an Options window with proper 3D toggle buttons for
// sound and music, styled after OpenLoco's option panels.
func (g *Game) openOptionsWindow() {
	const winTitle = "Options"
	const (
		winW   = 270
		winH   = 160
		rowH   = 32
		pad    = 10
		btnW   = 44
		btnH   = 22
		valueX = 140 // x offset (relative to content area) where ON button starts
	)
	// Content area dimensions (must match simplewindow.go constants).
	const (
		titleBarH = 22
		borderW   = 3
	)
	contentW := winW - borderW*2
	contentH := winH - titleBarH - borderW

	darkText := color.RGBA{20, 15, 10, 255}

	win := &ui.SimpleWindow{
		Title:   winTitle,
		X:       100,
		Y:       80,
		Width:   winW,
		Height:  winH,
		Visible: true,
	}
	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, _ *render.Renderer) {
		// Row 1: Sound Effects
		ui.DrawText(screen, "Sound Effects", cx+pad, cy+pad+btnH/2+4, darkText)
		drawButton(screen, cx+valueX, cy+pad, btnW, btnH, "On", g.soundEnabled, true)
		drawButton(screen, cx+valueX+btnW+4, cy+pad, btnW, btnH, "Off", !g.soundEnabled, true)

		// Row 2: Music
		ui.DrawText(screen, "Music", cx+pad, cy+pad+rowH+btnH/2+4, darkText)
		drawButton(screen, cx+valueX, cy+pad+rowH, btnW, btnH, "On", g.musicEnabled, true)
		drawButton(screen, cx+valueX+btnW+4, cy+pad+rowH, btnW, btnH, "Off", !g.musicEnabled, true)

		// Separator line
		sepY := cy + pad + rowH*2 + 6
		fillRect(screen, float64(cx+pad), float64(sepY), float64(cw-pad*2), 1, color.RGBA{140, 135, 120, 255})

		// Close button centred at bottom
		closeW := 60
		closeX := cx + (cw-closeW)/2
		closeY := cy + ch - btnH - 6
		drawButton(screen, closeX, closeY, closeW, btnH, "Close", false, true)
	}
	win.OnContentClick = func(relX, relY int) {
		// Sound row
		if relY >= pad && relY < pad+rowH {
			if relX >= valueX && relX < valueX+btnW {
				g.soundEnabled = true
			} else if relX >= valueX+btnW+4 && relX < valueX+btnW*2+4 {
				g.soundEnabled = false
			}
			return
		}
		// Music row
		if relY >= pad+rowH && relY < pad+rowH*2 {
			if relX >= valueX && relX < valueX+btnW {
				g.musicEnabled = true
			} else if relX >= valueX+btnW+4 && relX < valueX+btnW*2+4 {
				g.musicEnabled = false
				if g.audioMgr != nil {
					g.audioMgr.StopMusic()
				}
			}
			return
		}
		// Close button
		closeW := 60
		closeX := (contentW - closeW) / 2
		closeY := contentH - btnH - 6
		if relY >= closeY && relY < closeY+btnH && relX >= closeX && relX < closeX+closeW {
			g.windowMgr.CloseWindow(winTitle)
		}
	}
	g.windowMgr.OpenWindow(win)
}

// openLoadGameWindow opens the file browser for all supported file types.
func (g *Game) openLoadGameWindow() { g.openFileWindow("Load Game", false) }

// openFileWindow is the shared implementation for the two-panel file browser.
// title is the window title (also used as the CloseWindow key).
// scenariosOnly restricts the file list to .SC5 files when true.
func (g *Game) openFileWindow(title string, scenariosOnly bool) {
	const (
		winW      = 540
		winH      = 380
		itemH     = 18
		dirPanelW = 165 // width of the left directory panel
		sepW      = 2   // vertical separator between panels
		bottomH   = 46  // path display + button row at the bottom
	)

	// Colour scheme — all text is dark on the beige window background so it
	// stays readable regardless of the global palette.
	var (
		colText    = color.RGBA{20, 12, 4, 255}     // dark on beige — main text
		colLight   = color.RGBA{240, 238, 232, 255} // light — text on dark bg
		colDir     = color.RGBA{25, 60, 140, 255}   // blue — directory labels
		colSep     = color.RGBA{140, 132, 115, 255} // panel separator line
		colSelFile = color.RGBA{50, 100, 185, 220}  // selected file highlight
		colSelDir  = color.RGBA{35, 100, 45, 220}   // navigating into a dir
		colPathBg  = color.RGBA{178, 172, 155, 255} // path display bar
		colPanelHd = color.RGBA{155, 148, 130, 255} // panel header strip
	)

	// ── State captured in closure ──────────────────────────────────────────
	startDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		startDir = home
	}
	if g.dataDir != "" {
		startDir = filepath.Dir(g.dataDir) // one level up from "Data/"
	}

	var currentDir string
	var subDirs []string  // directory names within currentDir (basenames only)
	var fileList []string // full paths of .SC5/.SV5/.goloco files in currentDir
	selectedFile := -1
	dirScroll := 0
	fileScroll := 0
	lastFileIdx := -1
	var lastClickTime time.Time

	rescanDir := func(dir string) {
		currentDir = dir
		subDirs = nil
		fileList = nil
		selectedFile = -1
		dirScroll = 0
		fileScroll = 0
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				subDirs = append(subDirs, e.Name())
			} else {
				ext := strings.ToLower(filepath.Ext(e.Name()))
				isScenario := ext == ".sc5"
				isSave := ext == ".sv5" || ext == ".goloco"
				if isScenario || (!scenariosOnly && isSave) {
					fileList = append(fileList, filepath.Join(dir, e.Name()))
				}
			}
		}
	}
	rescanDir(startDir)

	win := ui.NewSimpleWindow(title, 60, 60, winW, winH)

	// ── DrawContent ────────────────────────────────────────────────────────
	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, r *render.Renderer) {
		listH := ch - bottomH
		visible := listH / itemH
		if visible < 1 {
			visible = 1
		}
		filePanelX := cx + dirPanelW + sepW
		filePanelW := cw - dirPanelW - sepW

		// ── Left panel header ──────────────────────────────────────────────
		fillRect(screen, float64(cx), float64(cy), float64(dirPanelW), float64(itemH), colPanelHd)
		ui.DrawText(screen, "Folders", cx+4, cy+itemH-4, colText)

		// ── Parent dir entry ──────────────────────────────────────────────
		parentY := cy + itemH
		parentLbl := "[↑  ..]"
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// at filesystem root — dim the entry
			ui.DrawText(screen, parentLbl, cx+4, parentY+itemH-4, colSep)
		} else {
			ui.DrawText(screen, parentLbl, cx+4, parentY+itemH-4, colDir)
		}

		// ── Subdirectory list ──────────────────────────────────────────────
		// The parent entry occupies row 0; subdirs start at row 1.
		// We reserve 1 slot for the parent, so only (visible-2) subdir rows
		// fit below it (minus 1 for header).
		subVisible := visible - 2
		if subVisible < 0 {
			subVisible = 0
		}
		for i := 0; i < subVisible; i++ {
			idx := dirScroll + i
			if idx >= len(subDirs) {
				break
			}
			ry := cy + (i+2)*itemH // +2: skip header and parent rows
			label := "[" + subDirs[idx] + "]"
			ui.DrawText(screen, label, cx+4, ry+itemH-4, colDir)
		}

		// Dir scroll indicator
		if len(subDirs) > subVisible && subVisible > 0 {
			total := len(subDirs)
			barH := listH * subVisible / total
			if barH < 6 {
				barH = 6
			}
			barY := listH * dirScroll / total
			fillRect(screen, float64(cx+dirPanelW-5), float64(cy), 5, float64(listH), colPanelHd)
			fillRect(screen, float64(cx+dirPanelW-5), float64(cy+barY), 5, float64(barH), colSep)
		}

		// ── Vertical separator ─────────────────────────────────────────────
		fillRect(screen, float64(cx+dirPanelW), float64(cy), float64(sepW), float64(listH), colSep)

		// ── Right panel header ─────────────────────────────────────────────
		fillRect(screen, float64(filePanelX), float64(cy), float64(filePanelW), float64(itemH), colPanelHd)
		ui.DrawText(screen, "Scenarios / Saves", filePanelX+4, cy+itemH-4, colText)

		// ── File list ──────────────────────────────────────────────────────
		// Files start below the header row.
		fileVisible := visible - 1
		if fileVisible < 0 {
			fileVisible = 0
		}
		for i := 0; i < fileVisible; i++ {
			idx := fileScroll + i
			if idx >= len(fileList) {
				break
			}
			ry := cy + (i+1)*itemH
			label := filepath.Base(fileList[idx])
			if idx == selectedFile {
				fillRect(screen, float64(filePanelX), float64(ry), float64(filePanelW), float64(itemH), colSelFile)
				ui.DrawText(screen, label, filePanelX+4, ry+itemH-4, colLight)
			} else {
				ui.DrawText(screen, label, filePanelX+4, ry+itemH-4, colText)
			}
		}
		if len(fileList) == 0 {
			ui.DrawText(screen, "No files here.", filePanelX+6, cy+itemH+14, colSep)
		}

		// File scroll indicator
		if len(fileList) > fileVisible && fileVisible > 0 {
			total := len(fileList)
			barH := listH * fileVisible / total
			if barH < 6 {
				barH = 6
			}
			barY := listH * fileScroll / total
			fillRect(screen, float64(filePanelX+filePanelW-5), float64(cy), 5, float64(listH), colPanelHd)
			fillRect(screen, float64(filePanelX+filePanelW-5), float64(cy+barY), 5, float64(barH), colSep)
		}

		// ── Bottom strip: current path + buttons ───────────────────────────
		stripY := cy + listH
		fillRect(screen, float64(cx), float64(stripY), float64(cw), float64(bottomH), colPathBg)

		// Path text (truncated on the left if too long)
		pathLabel := currentDir
		maxPathW := cw - 8
		for {
			w, _ := ui.MeasureText(pathLabel)
			if w <= maxPathW || len(pathLabel) < 4 {
				break
			}
			pathLabel = "…" + pathLabel[len(pathLabel)/4:]
		}
		ui.DrawText(screen, pathLabel, cx+4, stripY+14, colText)

		// Buttons
		btnY := stripY + bottomH - 26
		canLoad := selectedFile >= 0 && selectedFile < len(fileList)
		drawButton(screen, cx+4, btnY, 80, 22, "Cancel", false, true)
		drawButton(screen, cx+cw-84, btnY, 80, 22, "Load", false, canLoad)
		_ = colSelDir
	}

	// ── OnContentClick ─────────────────────────────────────────────────────
	win.OnContentClick = func(relX, relY int) {
		ch := win.Height - 22 - 3 // content height
		listH := ch - bottomH
		visible := listH / itemH

		if relY >= listH {
			// Button row
			btnRelY := relY - listH
			if btnRelY >= bottomH-26 {
				if relX >= 4 && relX < 84 {
					// Cancel
					g.windowMgr.CloseWindow(title)
				} else if relX >= win.Width-6-84 {
					// Load
					if selectedFile >= 0 && selectedFile < len(fileList) {
						if err := g.loadScenario(fileList[selectedFile]); err != nil {
							log.Printf("[Game] Load error: %v", err)
						}
					}
				}
			}
			return
		}

		row := relY / itemH

		if relX < dirPanelW {
			// ── Dir panel click ────────────────────────────────────────────
			switch row {
			case 0:
				// header — ignore
			case 1:
				// parent dir
				parent := filepath.Dir(currentDir)
				if parent != currentDir {
					rescanDir(parent)
				}
			default:
				// subdir rows start at row 2
				subVisible := visible - 2
				if subVisible < 0 {
					subVisible = 0
				}
				idx := dirScroll + (row - 2)
				if idx >= 0 && idx < len(subDirs) {
					rescanDir(filepath.Join(currentDir, subDirs[idx]))
				}
			}
		} else if relX >= dirPanelW+sepW {
			// ── File panel click ───────────────────────────────────────────
			if row == 0 {
				return // header
			}
			fileRow := row - 1
			idx := fileScroll + fileRow
			if idx < 0 || idx >= len(fileList) {
				return
			}
			now := time.Now()
			if idx == lastFileIdx && now.Sub(lastClickTime) < 400*time.Millisecond {
				// Double-click: load immediately
				if err := g.loadScenario(fileList[idx]); err != nil {
					log.Printf("[Game] Load error: %v", err)
				}
				return
			}
			selectedFile = idx
			lastFileIdx = idx
			lastClickTime = now
		}
	}

	// ── OnScrollWheel ──────────────────────────────────────────────────────
	win.OnScrollWheel = func(dy float64) {
		ch := win.Height - 22 - 3
		listH := ch - bottomH
		visible := listH / itemH
		subVisible := visible - 2
		fileVisible := visible - 1

		mx, _ := ebiten.CursorPosition()
		relX := mx - (win.X + 3)
		if relX < dirPanelW {
			// scroll dir panel
			maxD := iclamp(len(subDirs)-subVisible, 0, len(subDirs))
			dirScroll = iclamp(dirScroll-int(dy), 0, maxD)
		} else {
			// scroll file panel
			maxF := iclamp(len(fileList)-fileVisible, 0, len(fileList))
			fileScroll = iclamp(fileScroll-int(dy), 0, maxF)
		}
	}

	g.windowMgr.OpenWindow(win)
}

// -----------------------------------------------------------------------

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth < 1 {
		outsideWidth = defaultWidth
	}
	if outsideHeight < 1 {
		outsideHeight = defaultHeight
	}
	g.sw = outsideWidth
	g.sh = outsideHeight
	// Update toolbar to span the full window width on resize.
	if g.toolbar != nil && g.toolbar.Width != outsideWidth {
		g.toolbar.Width = outsideWidth
		g.toolbar.RightAlignButtons(outsideWidth)
	}
	return outsideWidth, outsideHeight
}

const diagCropW = 480
const diagCropH = 360

// saveDiagCrop saves a 480×360 PNG cropped from the centre of the screen
// (or centred on tile diagTileX/diagTileY when the world is loaded).
// The file is always named "diag.png" so Claude can read it without guessing the name.
// diagBackground draws a black+red cross-hatch pattern onto img so that
// transparent/unrendered areas are visually distinct from intentional black pixels.
func diagBackground(img *ebiten.Image) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	pixels := make([]byte, w*h*4)
	const stride = 16 // grid cell size
	red := [4]byte{180, 0, 0, 255}
	dark := [4]byte{20, 20, 20, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			onLine := (x%stride == 0) || (y%stride == 0)
			var c [4]byte
			if onLine {
				c = red
			} else {
				c = dark
			}
			i := (y*w + x) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = c[0], c[1], c[2], c[3]
		}
	}
	img.WritePixels(pixels)
}

func saveDiagCrop(screen *ebiten.Image, g *Game, tileX, tileY int) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Centre pixel: use tile position when valid, otherwise screen centre.
	cx, cy := sw/2, sh/2
	if g.w != nil && tileX >= 0 && tileY >= 0 {
		cx, cy = g.w.TileToScreenPx(tileX, tileY)
	}

	x0 := cx - diagCropW/2
	y0 := cy - diagCropH/2
	x0 = max(0, min(x0, sw-diagCropW))
	y0 = max(0, min(y0, sh-diagCropH))

	crop := screen.SubImage(image.Rect(x0, y0, x0+diagCropW, y0+diagCropH))

	f, err := os.Create("diag.png")
	if err != nil {
		log.Printf("[Diag] Cannot create diag.png: %v", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, crop); err != nil {
		log.Printf("[Diag] Encode failed: %v", err)
		return
	}
	log.Printf("[Diag] Saved diag.png (%dx%d crop centred on tile %d,%d, screen offset %d,%d)",
		diagCropW, diagCropH, tileX, tileY, x0, y0)
}

func main() {
	game := NewGame()

	// Parse CLI args:
	//   ./goloco [file]                    — load scenario normally
	//   ./goloco [file] --diag TX,TY       — save diag.png centred on tile TX,TY, quit
	//   ./goloco [file] --diag worst       — jump to tile with most render issues (dynamic)
	//   ./goloco [file] --diag water       — scan for nearest water tile, quit
	//   ./goloco [file] --diag building    — same for building / track / road / tree / station /
	//                                        industry / wall / train
	//   ./goloco [file] --diag             — save diag.png centred on map middle, quit
	//   ./goloco [file] --shots S,N       — save shot_01..NN.png every S seconds, quit
	args := os.Args[1:]
	var scenarioPath string
	diagTileX, diagTileY := -1, -1
	diagKind := ""
	diagMode := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--shots" && i+1 < len(args) {
			var secs, count int
			if n, _ := fmt.Sscanf(args[i+1], "%d,%d", &secs, &count); n >= 1 && secs > 0 {
				if count <= 0 {
					count = 5
				}
				game.shotsEvery = time.Duration(secs) * time.Second
				game.shotsLeft = count
				game.shotsLast = time.Now()
			}
			i++
			continue
		}
		if args[i] != "--diag" {
			scenarioPath = args[i]
			continue
		}
		diagMode = true
		if i+1 >= len(args) {
			break
		}
		next := args[i+1]
		n, _ := fmt.Sscanf(next, "%d,%d", &diagTileX, &diagTileY)
		if n == 2 {
			i++ // consumed as TX,TY
		} else if len(next) > 0 && next[0] != '-' {
			diagKind = next // consumed as kind keyword (water, worst, building, etc.)
			i++
		}
	}

	if scenarioPath != "" {
		if err := game.loadScenario(scenarioPath); err != nil {
			log.Printf("Failed to load %s: %v", scenarioPath, err)
		}
	}

	if diagMode && game.w != nil {
		if diagTileX < 0 {
			if diagKind == "" {
				diagKind = "worst"
			}
			game.w.CenterOn(diagKind) // handles "worst", "water", "building", etc.
			// Store the tile CenterOn snapped to so Draw() re-applies camera each
			// frame, overriding the title sequence tour camera.
			game.diagTileX = game.w.LastCenterTileX
			game.diagTileY = game.w.LastCenterTileY
			game.diagPending = true
			game.diagQuit = true
		} else {
			// Move camera to centre on the requested tile so it is
			// fully rendered when saveDiagCrop runs after Draw().
			game.w.SetCamera(float64(diagTileX), float64(diagTileY))
			game.diagPending = true
			game.diagQuit = true
			game.diagTileX = diagTileX
			game.diagTileY = diagTileY
		}
	}

	if p := os.Getenv("GOLOCO_SCRIPT"); p != "" {
		game.script = loadInputScript(p)
	}
	if os.Getenv("GOLOCO_OPEN") == "newgame" {
		game.openNewGameWindow()
	}
	if os.Getenv("GOLOCO_OPEN") == "track" && game.w != nil {
		// Headless construction test: open the window and chain a piece
		// sequence from the map centre so captures verify connections.
		game.buildMode = buildModeTrack
		game.windowMgr.OpenWindow(game.newConstructionWindow())
		mw, mh := game.w.GetMapSize()
		game.w.SetConstructionHead(mw/2, mh/2)
		for _, id := range []int{0, 0, 4, 0, 14, 0, 0, 6, 0} {
			if !game.w.PlaceTrackAtHead(id, 0) {
				log.Printf("[ConsTest] piece %d not placeable at head %+v", id, game.w.ConstructionHeadState())
			}
		}
		log.Printf("[ConsTest] head after sequence: %+v", game.w.ConstructionHeadState())
	}
	if os.Getenv("GOLOCO_OPEN") == "road" && game.w != nil {
		// Headless road construction test: chain straight + very-small curves.
		game.buildMode = buildModeRoad
		game.windowMgr.OpenWindow(game.newConstructionWindow())
		mw, mh := game.w.GetMapSize()
		game.w.SetConstructionHead(mw/2, mh/2)
		for _, id := range []int{0, 2, 0, 1, 0} {
			if !game.w.PlaceRoadAtHead(id, 0) {
				log.Printf("[RoadConsTest] piece %d not placeable at head %+v", id, game.w.ConstructionHeadState())
			}
		}
		log.Printf("[RoadConsTest] head after sequence: %+v", game.w.ConstructionHeadState())
	}

	ebiten.SetWindowSize(defaultWidth, defaultHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("GoLoco - Locomotion Reimplementation")
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

// --- RCT-style track construction window --------------------------------
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Construction.cpp
// Curve row + slope row of sprite buttons, Rotate/Undo, and a Build action.
// Click the map to place the construction head; pieces then chain from it.

// consTrackCandidates maps the selected curve (-4 left large … 0 straight …
// +4 right large) and slope (-2 steep down … 0 level … +2 steep up) to track
// piece candidates; the first placeable at the head's heading wins (this is
// how straight-vs-diagonal and cardinal-vs-diagonal larges resolve).
func consTrackCandidates(curve, slope int) []int {
	if curve == 0 {
		switch slope {
		case -2:
			return []int{17}
		case -1:
			return []int{15}
		case 1:
			return []int{14}
		case 2:
			return []int{16}
		default:
			return []int{0, 1} // straight, or diagonal when heading is diagonal
		}
	}
	if slope != 0 {
		return nil // curved slopes not yet supported
	}
	switch curve {
	case -1:
		return []int{2} // left very small
	case 1:
		return []int{3}
	case -2:
		return []int{4} // left small
	case 2:
		return []int{5}
	case -3:
		return []int{6} // left medium
	case 3:
		return []int{7}
	case -4:
		return []int{8, 10} // left large (to/from diagonal)
	case 4:
		return []int{9, 11}
	}
	return nil
}

// consSelectedTrackID resolves the current curve/slope selection to the first
// placeable track id at the head, or -1.
func (g *Game) consSelectedTrackID() int {
	for _, id := range consTrackCandidates(g.consCurve, g.consSlope) {
		if g.w.CanPlaceTrackAtHead(id) {
			return id
		}
	}
	return -1
}

// consRoadCandidates maps the selected curve/slope to road piece ids. Roads
// support only straight, very-small curves and (steep) slopes — VERIFIED
// against upstream kRoadCoordinates comments (straight=0, leftCurveVerySmall=1,
// rightCurveVerySmall=2, straightSlopeUp=5, straightSlopeDown=6,
// straightSteepSlopeUp=7, straightSteepSlopeDown=8).
func consRoadCandidates(curve, slope int) []int {
	if curve == 0 {
		switch slope {
		case -2:
			return []int{8} // straightSteepSlopeDown
		case -1:
			return []int{6} // straightSlopeDown
		case 1:
			return []int{5} // straightSlopeUp
		case 2:
			return []int{7} // straightSteepSlopeUp
		default:
			return []int{0} // straight
		}
	}
	if slope != 0 {
		return nil // curved slopes not supported for roads
	}
	switch curve {
	case -1:
		return []int{1} // leftCurveVerySmall
	case 1:
		return []int{2} // rightCurveVerySmall
	}
	return nil
}

// consSelectedRoadID resolves the current curve/slope selection to the first
// placeable road id at the head, or -1.
func (g *Game) consSelectedRoadID() int {
	for _, id := range consRoadCandidates(g.consCurve, g.consSlope) {
		if g.w.CanPlaceRoadAtHead(id) {
			return id
		}
	}
	return -1
}

// --- Mode-aware dispatch (track vs road) for the shared construction window --

func (g *Game) consIsRoad() bool { return g.buildMode == buildModeRoad }

func (g *Game) consCandidates(curve, slope int) []int {
	if g.consIsRoad() {
		return consRoadCandidates(curve, slope)
	}
	return consTrackCandidates(curve, slope)
}

func (g *Game) consCanPlace(id int) bool {
	if g.consIsRoad() {
		return g.w.CanPlaceRoadAtHead(id)
	}
	return g.w.CanPlaceTrackAtHead(id)
}

func (g *Game) consSelectedID() int {
	if g.consIsRoad() {
		return g.consSelectedRoadID()
	}
	return g.consSelectedTrackID()
}

func (g *Game) consPlace(id int) {
	if g.consIsRoad() {
		g.w.PlaceRoadAtHead(id, uint8(g.buildObjID))
	} else {
		g.w.PlaceTrackAtHead(id, uint8(g.buildObjID))
	}
}

func (g *Game) consUndo() {
	if g.consIsRoad() {
		g.w.UndoLastRoad()
	} else {
		g.w.UndoLastTrack()
	}
}

// consObjectNames returns the object display names for the current mode,
// indexed by object slot (empty string for unused slots).
func (g *Game) consObjectNames() []string {
	var names []string
	if g.r == nil || g.r.ObjMgr == nil {
		return names
	}
	if g.consIsRoad() {
		for _, o := range g.r.ObjMgr.RoadObjects {
			if o != nil {
				names = append(names, o.Name)
			} else {
				names = append(names, "")
			}
		}
	} else {
		for _, o := range g.r.ObjMgr.TrackObjects {
			if o != nil {
				names = append(names, o.Name)
			} else {
				names = append(names, "")
			}
		}
	}
	return names
}

// consObjName is the current object slot's display name, or "(none)".
func (g *Game) consObjName() string {
	names := g.consObjectNames()
	if g.buildObjID >= 0 && g.buildObjID < len(names) && names[g.buildObjID] != "" {
		return names[g.buildObjID]
	}
	return "(none)"
}

// consCycleType advances g.buildObjID to the next populated object slot.
func (g *Game) consCycleType(dir int) {
	names := g.consObjectNames()
	n := len(names)
	if n == 0 {
		return
	}
	for k := 0; k < n; k++ {
		g.buildObjID = ((g.buildObjID+dir)%n + n) % n
		if names[g.buildObjID] != "" {
			return
		}
	}
}

func (g *Game) newConstructionWindow() *ui.SimpleWindow {
	const winW, winH = 220, 182
	title := "Track Construction"
	if g.consIsRoad() {
		title = "Road Construction"
	}
	win := ui.NewSimpleWindow(title, 10, 40, winW, winH)

	// Sprite ids for the two button rows (ImageIds construction_*).
	curveSprites := []int{2346, 2344, 2342, 2340, 2335, 2341, 2343, 2345, 2347}
	curveValues := []int{-4, -3, -2, -1, 0, 1, 2, 3, 4}
	slopeSprites := []int{2348, 2349, 2350, 2351, 2352}
	slopeValues := []int{-2, -1, 0, 1, 2}

	const btnS = 22 // button size
	curveY, slopeY, actionY, typeY := 6, 34, 66, 122

	win.DrawContent = func(screen *ebiten.Image, cx, cy, cw, ch int, r *render.Renderer) {
		drawSpriteButton := func(x, y, spriteID int, selected, enabled bool) {
			drawButton(screen, x, y, btnS, btnS, "", selected, enabled)
			if img := r.GetSprite(spriteID); img != nil {
				b := img.Bounds()
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x+(btnS-b.Dx())/2), float64(y+(btnS-b.Dy())/2))
				if !enabled {
					op.ColorScale.ScaleAlpha(0.4)
				}
				screen.DrawImage(img, op)
			}
		}

		// Curve row
		rowX := cx + (cw-len(curveSprites)*btnS)/2
		for i, sp := range curveSprites {
			sel := g.consCurve == curveValues[i]
			en := false
			for _, id := range g.consCandidates(curveValues[i], 0) {
				if g.consCanPlace(id) {
					en = true
					break
				}
			}
			drawSpriteButton(rowX+i*btnS, cy+curveY, sp, sel, en)
		}
		// Slope row (centred, only meaningful for straight)
		rowX2 := cx + (cw-len(slopeSprites)*btnS)/2
		for i, sp := range slopeSprites {
			sel := g.consSlope == slopeValues[i]
			en := false
			for _, id := range g.consCandidates(0, slopeValues[i]) {
				if g.consCanPlace(id) {
					en = true
					break
				}
			}
			drawSpriteButton(rowX2+i*btnS, cy+slopeY, sp, sel, en)
		}

		// Action row
		head := g.w.ConstructionHeadState()
		canBuild := head.Active && g.consSelectedID() >= 0
		drawButton(screen, cx+6, cy+actionY, 62, 20, "Build", false, canBuild)
		drawButton(screen, cx+72, cy+actionY, 62, 20, "Undo", false, true)
		drawButton(screen, cx+138, cy+actionY, 62, 20, "Rotate", false, head.Active)

		// Status line
		status := "Click the map to start"
		if head.Active {
			dir := [16]string{"SW", "SE", "NE", "NW", "", "", "", "", "", "", "", "", "S", "E", "N", "W"}[head.Rotation]
			status = fmt.Sprintf("Head: %d,%d h%d %s", head.X, head.Y, head.BaseZ, dir)
		}
		ui.DrawText(screen, status, cx+6, cy+actionY+28, color.RGBA{220, 220, 180, 255})
		ui.DrawText(screen, "[R]=rotate  [X]=undo  [Esc]=done", cx+6, cy+actionY+42, color.RGBA{160, 160, 160, 255})

		// Type picker: prev/next cycling button showing the current object name.
		drawButton(screen, cx+6, cy+typeY, 20, 20, "<", false, true)
		drawButton(screen, cx+cw-26, cy+typeY, 20, 20, ">", false, true)
		drawButton(screen, cx+30, cy+typeY, cw-62, 20, g.consObjName(), false, true)
	}

	win.OnContentClick = func(relX, relY int) {
		rowX := (winW - len(curveSprites)*btnS) / 2
		if relY >= curveY && relY < curveY+btnS {
			if i := (relX - rowX) / btnS; i >= 0 && i < len(curveValues) && relX >= rowX {
				g.consCurve = curveValues[i]
				if g.consCurve != 0 {
					g.consSlope = 0
				}
			}
			return
		}
		rowX2 := (winW - len(slopeSprites)*btnS) / 2
		if relY >= slopeY && relY < slopeY+btnS {
			if i := (relX - rowX2) / btnS; i >= 0 && i < len(slopeValues) && relX >= rowX2 {
				g.consSlope = slopeValues[i]
				if g.consSlope != 0 {
					g.consCurve = 0
				}
			}
			return
		}
		if relY >= actionY && relY < actionY+20 {
			switch {
			case relX >= 6 && relX < 68:
				if id := g.consSelectedID(); id >= 0 {
					g.consPlace(id)
				}
			case relX >= 72 && relX < 134:
				g.consUndo()
			case relX >= 138 && relX < 200:
				g.w.RotateConstructionHead()
			}
			return
		}
		if relY >= typeY && relY < typeY+20 {
			switch {
			case relX >= 6 && relX < 26:
				g.consCycleType(-1)
			case relX >= winW-26 && relX < winW-6:
				g.consCycleType(1)
			}
		}
	}
	return win
}
