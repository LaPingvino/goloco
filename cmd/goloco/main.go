package main

import (
	"encoding/json"
	"fmt"
	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/title"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
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
	buildModeNone    = 0
	buildModeTrack   = 1
	buildModeRoad    = 2
	buildModeDock    = 3 // Construction::openWithFlags(1 << 30): the dock window (Station tab only)
	buildModeAirport = 4 // openWithFlags(1 << 31): the airport window (Station tab only, rotate button)
)

type Game struct {
	diagConsDone                   bool
	noEdgeScrolling                bool // Options > Controls: edge scrolling off (Config::edgeScrolling)
	w                              *world.World
	titleWorld                     *world.World // preserved title-screen world; restored on Quit to Menu
	r                              *render.Renderer
	toolbar                        *ui.Toolbar
	windowMgr                      *ui.SimpleWindowManager
	objMgr                         *objects.ObjectManager
	audioMgr                       *audio.Manager
	scenarioSelect                 *ui.ScenarioSelectWindow // the New Game window when open
	diagNewGameDone                bool
	browse                         *ui.PromptBrowseWindow    // the Load window when open
	options                        *ui.OptionsWindow         // the Options window when open
	townList                       *ui.TownListWindow        // the town list when open
	buildVehicle                   *ui.BuildVehicleWindow    // the Build Vehicle window when open (5P.f)
	vehicleWin                     *ui.VehicleWindow         // the Vehicle window when open (5P.g)
	diagMouse                      *[2]int                   // DIAGNOSTIC ONLY: the mouse held at a point (the roadghost diag's hover)
	ghostWater                     sim.VehicleWaterPlacement // _ghostVehiclePos for a boat ghost (ghostProgress 0)
	ghostAirStation                uint16                    // _ghostAirportStationId (0xFFFF none)
	ghostAirNode                   uint8                     // _ghostAirportNode
	diagShip                       uint16                    // DIAGNOSTIC ONLY: the shipsail diag's ship, logged as it goes
	diagFollow                     bool                      // DIAGNOSTIC ONLY: keep the view on diagShip (planefly)
	diagShipFrames                 int
	buildVehDiagDone               bool
	diagTownsDone                  bool
	diagOptionsDone                bool
	diagBrowseDone                 bool
	titleSeq                       *title.Sequence
	inGame                         bool // a scenario / save is loaded into gameplay
	dropdown                       *ui.DropdownMenu
	mouseX                         int
	lastTickMouseX, lastTickMouseY int       // the cursor at the last GameTicks (mouseMoved)
	showFPS                        bool      // Config::showFPS (drawFPS)
	scale                          float64   // Config::scaleFactor: the game drawn at window ÷ scale, scaled up
	fpsRef                         time.Time // FPSCounter's _referenceTime
	fpsFrames                      int
	fpsCurrent                     float64
	mouseY                         int
	dataDir                        string
	inTitleScreen                  bool
	soundEnabled                   bool   // whether sound effects are on
	musicEnabled                   bool   // whether music is on
	currentScenarioPath            string // path of the currently loaded scenario/save
	saveMsg                        string // transient "Game Saved!" status overlay
	saveMsgFrames                  int    // frames remaining to show saveMsg

	// Current logical screen dimensions (updated by Layout each frame)
	sw, sh int

	// Right-click drag pan state
	rightDragging bool
	dragLastX     int
	dragLastY     int

	// Title screen globe animation counter (incremented each frame)
	titleFrame  int
	windowFrame int // counts Updates: the windows' frameNo (onUpdate's frameNo += 1)

	// Speed controls
	isPaused  bool
	speedMult int // 1, 2, 4, or 8

	// Build mode (track/road placement)
	buildMode  int    // buildModeNone / buildModeTrack / buildModeRoad
	pickupOn   bool   // the Vehicle window's pickup (placing) tool is active
	pickupHead uint16 // the vehicle head it places
	routeSel   int    // the Vehicle window's Route selection (orderTableIndex; -1: none)
	routeTool  bool   // the Route tab's tool (adding stops by clicking the map) is active
	drawn      int    // frames drawn so far (diagnostics that need a painted view wait for some)
	routeHead  uint16 // whose Route that selection is
	// The ghost's placement (_ghostLandTrackAndDirection, _ghostVehiclePos,
	// _ghostTrackProgress; ghostProgress -1: none).
	ghostPlace    sim.VehiclePlacement
	ghostProgress int
	diagLogged    bool // GOLOCO_DIAG_BUILDVEHICLE=place: the first placement logged
	objective     scenario.Objective
	hasObjective  bool
	gameWon       bool
	progress      *progressFile

	consCurve         int    // construction curve selection: -4 (left large) … 0 … +4
	consTab           int    // construction window tab: 0 construction, 2 signal (1 station places directly)
	consSigOne        bool   // signal tab: single direction chosen (isSignalBothDirections = 0)
	preferredCurrency string // Config::preferredCurrency (DAT name; PARTIAL: not saved)
	// preferredCurrencyFlags are Config usePreferredCurrencyForNewGames and
	// usePreferredCurrencyAlways (default off).
	vehDiagDone            bool   // GOLOCO_DIAG_VEHLIST opened
	fsResolution           [2]int // Config display.fullscreenResolution
	preferredCurrencyFlags [2]bool
	frameLimit             int                        // Options' frame rate limit (0 internal, 1 vsync, 2 unrestricted)
	prompt                 *ui.PromptOkCancel         // an open confirmation (modal: it takes all input)
	textInput              *ui.TextInputWindow        // an open text input (renaming); PARTIAL: takes all input, upstream's is not modal
	consBridge             int                        // lastSelectedBridge (slot; -1 = the first available)
	consOneShot            int                        // track: an s-bend/turnaround piece chosen (construct.TrackPiece ≥ 9), -1 none
	consSignal             int                        // signal tab: lastSelectedSignal (slot; -1 = the first available)
	consStation            int                        // station tab: lastSelectedStationType (slot; -1 = the first available)
	stationGhost           *sim.TrainStationPlacement // the station tab's ghost (cState.stationGhost*), nil none
	stationGhostCost       int32                      // cState.stationCost
	roadStationGhost       *sim.RoadStationPlacement  // the road stop ghost (stationGhostType bit 7)
	dockGhost              *sim.PortPlacement         // the dock ghost (stationGhostType bit 14)
	diagMoney              bool                       // DIAGNOSTIC ONLY: keep a money popup alive (moneyeffect)
	airportGhost           *sim.AirportPlacement      // the airport ghost (stationGhostType bit 15)
	airportRotation        uint8                      // cState.constructionRotation in the airport window
	lastPortOption         int                        // ToolbarTop's _lastPortOption: 0 airport, 1 dock
	errorWin               *ui.ErrorWindow            // WindowType::error (one at a time)
	stationGhostStation    int                        // cState.constructingStationId (-1 new)
	stationGhostAccepted   uint32                     // constructingStationAcceptedCargoTypes
	stationGhostProduced   uint32                     // constructingStationProducedCargoTypes
	diagHoldGhost          bool                       // DIAGNOSTIC ONLY: the stationghost diag keeps its ghost
	consMods               uint8                      // overhead tab: lastSelectedMods (PARTIAL: not from the scenario's trackMods/roadMods)
	consSlope              int                        // construction slope selection: -2 (steep down) … 0 … +2
	buildObjID             int                        // index into TrackObjects / RoadObjects
	hoverTileX             int                        // tile under cursor in build mode (-1 if invalid)
	hoverTileY             int

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
	titleTip   string        // the hovered title button's tooltip (drawn last)
	shotsLeft  int
	shotsLast  time.Time
	shotsN     int
	// quitAfterSave: exit cleanly once the map save has completed.
	quitAfterSave bool

	// Loading-screen state.  When loadingSpriteIDs != nil, the game is in the
	// sprite-preload phase and will draw the OpenLoco-style loading screen.
	loadingSpriteIDs []int // full list collected from the world
	loadingStep      int   // next index to process

	// Background sprite prewarm: decode+pack a few sprites per frame within a
	// small time budget, so the game shows at once instead of after a blocking
	// preload of every sprite (11,057 for the title).
	prewarmIDs       []int
	prewarmPos       int
	firstFrameLogged bool   // the start-to-first-world-frame time has been logged
	loadingStepSize  int    // sprites per frame
	loadingCaption   string // text shown in the loading window caption
	loadingStyleFlip bool   // alternates between train style 0 and 1 each load
}

func findLocoDataDir() string {
	home := os.Getenv("HOME")
	var candidates []string
	if d := os.Getenv("GOLOCO_DATA"); d != "" { // the Locomotion folder (or its Data) named explicitly
		candidates = append(candidates, filepath.Join(d, "Data"), d)
	}
	candidates = append(candidates,
		"../locomotion/Data",
		"locomotion/Data",
		filepath.Join(home, "Locomotion/Data"), // an unpacked copy in the home folder
		filepath.Join(home, "locomotion/Data"),
		filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/common/Locomotion/Data"),
		filepath.Join(os.Getenv("HOME"), ".steam/steam/steamapps/common/Locomotion/Data"),
		"/usr/share/games/locomotion/Data",
	)

	for _, dir := range candidates {
		g1Path := filepath.Join(dir, "g1.DAT")
		if _, err := os.Stat(g1Path); err == nil {
			return dir
		}
	}
	return ""
}

func NewGame() *Game {
	if v := os.Getenv("GOLOCO_DIAG_PAINT_AT"); v != "" { // diagnostic: the painters of one viewport pixel
		var at [2]int
		if _, err := fmt.Sscanf(v, "%d,%d", &at[0], &at[1]); err == nil {
			paint.DiagAt = &at
		}
	}
	log.Println("[Game] Creating new game...")

	// Load the UI language pack (GOLOCO_LANG, default en-GB) before any object
	// string tables are parsed — Load also sets objects.SelectedLocoLanguage
	// from the pack header so DAT strings match. A missing pack is non-fatal.
	lang := os.Getenv("GOLOCO_LANG")
	if lang == "" {
		lang = configValue("language") // the language chosen in Options last time (Config.cpp)
	}
	if err := i18n.Load(lang); err != nil {
		log.Printf("[Game] i18n: %v (using English)", err)
	} else {
		log.Printf("[Game] Language: %s (%s)", i18n.Locale(), i18n.NativeName())
	}

	// Initialize fonts
	log.Println("[Game] Loading fonts...")
	_, err := ui.InitFonts("../fonts")
	if err != nil {
		log.Printf("[Game] Failed to load fonts: %v (will use fallback)", err)
	}

	r := render.NewRenderer()

	// Initialize audio manager (GOLOCO_MUTE=1 skips it entirely — parallel
	// headless instances would otherwise fight over the ALSA device). With no
	// sound device the game plays on silently, as OpenLoco does (item AU).
	var audioMgr *audio.Manager
	switch {
	case os.Getenv("GOLOCO_MUTE") == "1":
	case !audio.DeviceAvailable():
		log.Println("[Game] No sound device: playing without sound")
	default:
		log.Println("[Game] Creating audio manager...")
		audioMgr = audio.NewManager()
	}

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
	} else if og := openGraphicsDir(); og != "" {
		// The open build (item O): no Locomotion folder — OpenGraphics (MIT):
		// its palette and an empty base pool (assets.NewOpenG1), its objects.
		// The bitmap fonts are g1.DAT's, so the TTF fallback stays.
		log.Printf("[Game] No Locomotion data; the open build from %s", og)
		if g1, err := assets.NewOpenG1(filepath.Join(og, "locopal.jasc.pal")); err != nil {
			log.Printf("[Game] open build: %v", err)
		} else {
			r.G1 = g1
			graphics.SetGlobalPalette(r.G1)
			objMgr = objects.NewObjectManager(filepath.Join(og, "ObjData"))
			objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
			objMgr.G1File = g1
			if err := objMgr.LoadAllObjects(); err != nil {
				log.Printf("[Game] open build: objects: %v", err)
			}
			log.Printf("[Game] open build: %d vehicles, %d land objects, %d images", len(objMgr.Vehicles), len(objMgr.LandObjects), g1.GetTotalSprites())
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
				reorderScenarioObjects(objMgr, sc)
			}
			w.LoadFromScenario(sc)
			log.Printf("[Game] Loaded title scenario: %dx%d map", sc.MapWidth, sc.MapHeight)
		} else {
			log.Printf("[Game] Could not load title scenario: %v", err)
		}
	} else if r.G1 != nil && objMgr != nil { // the open build: no title.dat — a landscape with the editor's defaults
		m := sim.NewGeneratedMap(0x4C4F434F, 0x474F4C4F, 1900, sim.EditorDefaultTerrainOptions(objMgr), r.G1, objMgr)
		w.LoadGenerated(m)
		log.Printf("[Game] open build: generated title landscape %dx%d", m.Width, m.Height)
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
	if z, err := strconv.Atoi(os.Getenv("GOLOCO_ZOOM")); err == nil { // diagnostic: start at another zoom (0-3)
		w.SetZoom(z)
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
		showFPS:       configValue("showFPS") == "true" || os.Getenv("GOLOCO_SHOWFPS") != "", // Config::showFPS (Config.cpp: default false)
		scale:         loadScale(),
		w:             w,
		consSignal:    -1,
		consOneShot:   -1,
		consBridge:    -1,
		consStation:   -1,
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
		// Warm the title's sprites in the background (prewarmTick); the title
		// draws at once and decodes anything it needs first on use, as upstream
		// draws g1 sprites directly with no preload phase.
		prewarmIDs:     titleSpriteIDs,
		loadingCaption: "title.dat",
	}
}

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
	if g.textInput != nil { // TextInputWindow: keyUp (typed characters, editing keys, Enter / Escape), onMouseUp, onUpdate
		t := g.textInput
		t.Update()
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			t.HandleClick(mx, my)
		}
		var keys []int
		for k, code := range map[ebiten.Key]int{ebiten.KeyBackspace: ui.KeyBackspace, ebiten.KeyDelete: ui.KeyDelete, ebiten.KeyHome: ui.KeyHome,
			ebiten.KeyEnd: ui.KeyEnd, ebiten.KeyArrowLeft: ui.KeyLeft, ebiten.KeyArrowRight: ui.KeyRight} {
			if inpututil.IsKeyJustPressed(k) || inpututil.KeyPressDuration(k) > 30 && inpututil.KeyPressDuration(k)%3 == 0 { // key repeat
				keys = append(keys, code)
			}
		}
		t.Keys(ebiten.AppendInputChars(nil), keys, inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),
			inpututil.IsKeyJustPressed(ebiten.KeyEscape))
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
	if g.script != nil {
		g.script.step()
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
			g.browse.Drag(g.mouseY, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
		}
		if g.scenarioSelect != nil { // and the scenario list's
			g.scenarioSelect.Drag(g.mouseY, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
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

	// Update dropdown hover if visible
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	} else {
		g.toolbar.UpdateHover(g.mouseX, g.mouseY)
	}

	if g.townList != nil { // the list's scrollbar thumb follows the mouse while held
		g.townList.Drag(g.mouseY, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
	}
	if g.browse != nil { // the load/save list's thumb too
		g.browse.Drag(g.mouseY, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
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
		} else if g.vehicleWin != nil && g.vehicleWin.Contains(g.mouseX, g.mouseY) {
			g.vehicleWin.HandleClick(g.mouseX, g.mouseY)
		} else if g.buildVehicle != nil && g.buildVehicle.Contains(g.mouseX, g.mouseY) {
			g.buildVehicle.HandleClick(g.mouseX, g.mouseY)
		} else if g.townList != nil && g.townList.Contains(g.mouseX, g.mouseY) {
			g.townList.HandleClick(g.mouseX, g.mouseY)
		} else if g.options != nil && g.options.Contains(g.mouseX, g.mouseY) {
			g.options.HandleClick(g.mouseX, g.mouseY)
		} else if g.browse != nil && g.browse.Contains(g.mouseX, g.mouseY) {
			g.browse.HandleClick(g.mouseX, g.mouseY)
		} else if g.scenarioSelect != nil && g.scenarioSelect.Contains(g.mouseX, g.mouseY) {
			g.scenarioSelect.HandleClick(g.mouseX, g.mouseY)
		} else if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
			if sp := uilayout.TimePanelSpeedAt(g.mouseX, g.mouseY, g.sw, g.sh); sp >= 0 && !g.inTitleScreen {
				// TimePanel speed buttons: pause toggles; the others set the speed.
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
				// the bottom-bar windows swallow the click
			} else if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
				g.toolbar.Buttons[btnIdx].Pressed = true
				g.handleToolbarButton(btnIdx)
			} else if g.pickupOn && !g.inTitleScreen {
				g.pickupToolDown(g.mouseX, g.mouseY) // the Vehicle window's placing tool
			} else if g.routeTool && g.vehicleWin != nil && !g.inTitleScreen {
				g.routeToolDown(g.mouseX, g.mouseY) // the Route tab's tool: a stop at the station clicked
			} else if (g.buildMode == buildModeTrack || g.buildMode == buildModeRoad || g.buildMode == buildModeDock || g.buildMode == buildModeAirport) && g.consTab == 1 && !g.inTitleScreen {
				g.stationToolUp(g.mouseX, g.mouseY) // the Station tab's tool: a station on the track or road clicked, or a dock
			} else if (g.buildMode == buildModeTrack || g.buildMode == buildModeRoad) && g.hoverTileX >= 0 {
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

	// Edge scrolling: pan when cursor is within edgeScrollZone px of screen edge.
	// OpenLoco reference: src/OpenLoco/src/Input/Keyboard.cpp edgeScrolling()
	// Speed is scaled by zoom so scrolling feels consistent at all zoom levels.
	if ebiten.IsFocused() && !g.rightDragging && !g.noEdgeScrolling {
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

// drawPlayerPanel draws the ported PlayerInfoPanel (bottom-left).
func (g *Game) drawPlayerPanel(screen *ebiten.Image) {
	if !g.inGame {
		return
	}
	var st ui.PlayerPanelState
	// The controlling company's own cash (the save, kept by the ported sim).
	if cash, ok := g.w.ControllingCompanyCash(); ok {
		st.Cash = cash
	}
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		st.SkinColour = g.r.ObjMgr.InterfaceSkin.PlayerInfoToolbarColour
	}
	g.playerPanelCompany(&st)
	wx, wy := uilayout.PlayerPanelPos(g.sw, g.sh)
	cv, pf := uilayout.PlayerPanelCompanyVal, uilayout.PlayerPanelPerformance
	st.HoverCash = g.mouseX >= wx+cv[0] && g.mouseX < wx+cv[0]+cv[2] && g.mouseY >= wy+cv[1] && g.mouseY < wy+cv[1]+cv[3]
	st.HoverPerf = g.mouseX >= wx+pf[0] && g.mouseX < wx+pf[0]+pf[2] && g.mouseY >= wy+pf[1] && g.mouseY < wy+pf[1]+pf[3]
	ui.DrawPlayerPanel(screen, st)
}

// drawTimePanel draws the ported TimePanel (bottom-right) and, when its date
// is hovered, the date tooltip with the scenario objective.
func (g *Game) drawTimePanel(screen *ebiten.Image) {
	if !g.inGame {
		return
	}
	mp := &g.w.Map // getCurrentDay's date (pkg/sim dateTick)
	st := ui.TimePanelState{Date: uilayout.DateDMY(int(mp.CurrentDayOfMonth)+1, i18n.Month(int(mp.CurrentMonth)+1), int(mp.CurrentYear))}
	switch {
	case g.isPaused:
		st.Speed = 0
	case g.speedMult <= 1:
		st.Speed = 1
	case g.speedMult == 2:
		st.Speed = 2
	default:
		st.Speed = 3
	}
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		st.SkinColour = g.r.ObjMgr.InterfaceSkin.TimeToolbarColour
		st.SkinImage = int(g.r.ObjMgr.InterfaceSkin.ImageOffset)
	}
	wx, wy := uilayout.TimePanelPos(g.sw, g.sh)
	d := uilayout.TimePanelDate
	st.HoverDate = g.mouseX >= wx+d[0] && g.mouseX < wx+d[0]+d[2] && g.mouseY >= wy+d[1] && g.mouseY < wy+d[1]+d[3]
	ui.DrawTimePanel(screen, st)
	if st.HoverDate && g.hasObjective {
		// formatChallenge (string 647: date, then 695 / 696 / 699 + 700).
		tip := st.Date
		completed, failed, beaten, progress := mp.ChallengeStatus()
		switch {
		case completed:
			tip += "\nChallenge completed!"
		case failed:
			tip += "\nChallenge failed!"
		case beaten:
		default:
			tip += fmt.Sprintf("\nChallenge progress: %d%%", progress)
			if mp.Objective.Flags&(1<<2) != 0 { // withinTimeLimit
				left := uint16(mp.Objective.TimeLimitYears)*12 - mp.ObjectiveMonthsInChallenge
				tip += fmt.Sprintf("\nTime left: %d years %d months", left/12, left%12)
			}
		}
		var c uint8
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			c = g.r.ObjMgr.InterfaceSkin.TooltipColour
		}
		ui.DrawTooltip(screen, tip, g.mouseX, g.mouseY, c)
	}
}

// toolbarDropdown is Dropdown::showBelow for a toolbar button: below the
// button, in the button's window colour (the interface skin's top-toolbar
// colour for its group), translucent.
func (g *Game) toolbarDropdown(btn *ui.ToolbarButton, items []ui.DropdownItem) *ui.DropdownMenu {
	d := ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
	d.Place(btn.X, btn.Y, btn.Width, btn.Height, g.sw, g.sh)
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		sk := g.r.ObjMgr.InterfaceSkin
		cols := [4]uint8{sk.TopToolbarPrimaryColour, sk.TopToolbarSecondaryColour, sk.TopToolbarTertiaryColour, sk.TopToolbarQuaternaryColour}
		if btn.ColorGroup >= 0 && btn.ColorGroup < 4 {
			d.SetToolbarColour(cols[btn.ColorGroup])
		}
	}
	return d
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
//
// clickable(ToolbarTop.cpp): build_vehicles_menu
// clickable(ToolbarTopCommon.cpp): zoom_menu
// clickable-partial(ToolbarTop.cpp): loadsave_menu audio_menu vehicles_menu stations_menu towns_menu
// clickable-partial(ToolbarTopCommon.cpp): railroad_menu road_menu
// (not yet: rotate_menu and view_menu only log; terraform_menu, port_menu, cheats_menu nothing)
func (g *Game) handleToolbarButton(idx int) {
	btn := &g.toolbar.Buttons[idx]
	tooltip := btn.Tooltip

	// Create appropriate dropdown or window based on button tooltip
	var win *ui.SimpleWindow
	switch tooltip {
	case "Load/Save":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp:117-154
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuLoadGame), Action: func() { g.openLoadGameWindow() }},
			{Text: i18n.Label(i18n.MenuSaveGame), Action: func() { g.openSaveWindow() }},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuScreenshot), Action: func() {
				g.diagPending = true
				g.diagQuit = false
				g.diagTileX = -1
				g.diagTileY = -1
			}},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuAbout), Action: func() { log.Println("[Game] About selected") }},
			{Text: i18n.Label(i18n.Options), Action: func() { g.openOptionsWindow() }},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuQuitToMenu), Action: func() { g.quitToMenu() }},
			{Text: i18n.Label(i18n.MenuExitOpenloco), Action: func() { os.Exit(0) }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)
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
		g.dropdown = g.toolbarDropdown(btn, items)
		return

	case "Zoom":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:95-132
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuZoomIn), Action: func() { g.w.ZoomIn() }},
			{Text: i18n.Label(i18n.MenuZoomOut), Action: func() { g.w.ZoomOut() }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)
		return

	case "Rotate":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:135-143
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuRotateClockwise), Action: func() { log.Println("[Game] Rotate CW") }},
			{Text: i18n.Label(i18n.MenuRotateAntiClockwise), Action: func() { log.Println("[Game] Rotate CCW") }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)
		return

	case "View":
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuUndergroundView), Action: func() { log.Println("[Game] Underground view") }},
			{Text: "See Through Tracks", Action: func() { log.Println("[Game] See through tracks") }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)
		return
	case "Terraform":
		log.Println("[Game] Terraform menu (not yet implemented)")
		return
	case "Railroad":
		g.buildMode = buildModeTrack
		g.buildObjID = 0
		if names := g.consObjectNames(); len(names) > 0 && names[0] == "" {
			g.consCycleType(1) // slot 0 empty in this scenario — first populated slot
		}
		g.consCurve = 0
		g.consSlope = 0
		win = g.newPortedConstructionWindow()

	case "Road":
		g.buildMode = buildModeRoad
		g.buildObjID = 0
		if names := g.consObjectNames(); len(names) > 0 && names[0] == "" {
			g.consCycleType(1) // slot 0 empty in this scenario — first populated slot
		}
		g.consCurve = 0
		g.consSlope = 0
		win = g.newPortedConstructionWindow()
	case "Port/Airport":
		g.portMenu(btn)
		return
	case "Build Vehicles":
		g.buildVehiclesMenu(btn)
		return
	case "Vehicles":
		g.openVehiclesWindow()
		return
	case "Stations":
		g.openStationsWindow()
		return
	case "Towns":
		g.openTownsWindow()
		return
	default:
		// Not ported yet: no window (an invented placeholder window would not
		// be the game's own).
		log.Printf("[Game] %s: window not ported yet", tooltip)
		return
	}

	if win != nil {
		g.windowMgr.OpenWindow(win)
	}
}

// drawLoadingScreen draws the OpenLoco-style progress window.
//
// NOT REACHED since the blocking sprite preload was dropped (loop tick 6):
// upstream shows this window while an S5 file loads (S5/S5.cpp:728,
// Ui::ProgressBar::begin(StringIds::loading)), not for sprites. goloco loads
// the file synchronously before the next frame; wire this to real S5 load
// progress (Ui/Windows/ProgressBar.cpp) when loading moves off the frame.
func (g *Game) drawLoadingScreen(screen *ebiten.Image) {
	total := len(g.loadingSpriteIDs)
	if total == 0 {
		return
	}
	style := 0
	if g.loadingStyleFlip {
		style = 1
	}
	// Ui::ProgressBar (pkg/ui/progressbar.go); the caption is StringIds::loading.
	ui.DrawProgressBar(screen, "Loading...", uint8(g.loadingStep*255/total), style)
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

// prewarmBudget is the per-frame time spent warming sprites in the background.
const prewarmBudget = 4 * time.Millisecond

// prewarmTick decodes and packs queued sprites until the frame budget is used.
func (g *Game) prewarmTick() {
	if g.prewarmPos >= len(g.prewarmIDs) {
		return
	}
	deadline := time.Now().Add(prewarmBudget)
	for g.prewarmPos < len(g.prewarmIDs) && time.Now().Before(deadline) {
		g.r.GetSprite(g.prewarmIDs[g.prewarmPos])
		g.prewarmPos++
	}
	if g.prewarmPos >= len(g.prewarmIDs) {
		log.Printf("[Game] Background prewarm complete after %v — atlas has %d sprites", time.Since(processStart).Round(time.Millisecond), g.r.AtlasPackedCount())
		g.prewarmIDs, g.prewarmPos = nil, 0
	}
}

// processStart is when the program started; used for startup timings.
var processStart = time.Now()

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawn++
	if g.shotsEvery > 0 && g.shotsLeft > 0 && time.Since(g.shotsLast) >= g.shotsEvery && os.Getenv("GOLOCO_SHOT_PAINTDUMP") != "" {
		paint.Dump = &[]string{} // diagnostic: this frame's paint structs, saved with the shot
	}
	defer g.maybeSaveShot(screen) // captures (--shots) of the finished frame: after the FPS counter
	defer g.drawFPS(screen)       // SoftwareDrawingEngine: the FPS counter over everything, last
	// ── Loading screen ──────────────────────────────────────────────────────
	// When loadingSpriteIDs is set, show OpenLoco-style progress window and
	// pack sprites into the atlas N per frame. Normal draw is skipped.
	if g.loadingSpriteIDs != nil {
		g.drawLoadingScreen(screen)
		g.doLoadingStep()
		return
	}
	if !g.firstFrameLogged {
		g.firstFrameLogged = true
		log.Printf("[Game] First world frame %v after start", time.Since(processStart).Round(time.Millisecond))
	}
	if fstats.on {
		t0 := time.Now()
		defer func() { fstats.addDraw(t0, time.Since(t0)) }() // runs last: the prewarm is included
	}
	defer g.prewarmTick()

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
	ui.InvalidateFrame8() // the world (drawScene) starts draw8's screen copy
	g.w.SceneEvery = 1
	if g.uiOpen() {
		g.w.SceneEvery = 3 // the world at a third of the rate, the menus at full (LOOP item 0)
	}
	tw := time.Now()
	g.w.Draw(screen)
	fstats.addWorld(time.Since(tw))
	fstats.parts = g.w.SceneParts

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
		ui.MarkForeign(screen.Bounds())
	} else if g.buildMode == buildModeRoad {
		g.w.DrawRoadConstructionGhost(screen, g.consSelectedRoadID(), g.buildObjID)
		ui.MarkForeign(screen.Bounds())
	}

	if g.inTitleScreen {
		g.drawTitleMenu(screen)
		g.windowMgr.Draw(screen, g.r) // allow windows (e.g. Load Game) on title screen
		g.windowMgr.MarkForeign()
		if g.scenarioSelect != nil {
			g.scenarioSelect.Draw(screen)
		}
		if g.browse != nil {
			g.browse.Draw(screen)
		}
		if g.options != nil {
			g.options.Draw(screen)
		}
		if g.townList != nil {
			g.townList.Draw(screen)
		}
		if g.dropdown != nil && g.dropdown.Visible { // a window's dropdown, on top
			g.dropdown.Draw(screen)
		}
		if g.prompt != nil { // a confirmation, on top of everything
			g.prompt.Draw(screen)
		}
		// A title button's tooltip (ToolTip.cpp), when no window is over it.
		// PARTIAL: shown at once, as goloco's other tooltips (upstream waits).
		if g.titleTip != "" && g.options == nil && g.browse == nil && g.scenarioSelect == nil && g.prompt == nil &&
			(g.dropdown == nil || !g.dropdown.Visible) {
			var c uint8
			if g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
				c = g.r.ObjMgr.InterfaceSkin.TooltipColour
			}
			ui.DrawTooltip(screen, g.titleTip, g.mouseX, g.mouseY, c)
		}
		g.drawInlineDiag(screen)
		return
	}

	// --- Gameplay HUD ---
	g.toolbar.Draw(screen, g.r)
	ui.MarkForeign(image.Rect(g.toolbar.X, g.toolbar.Y, g.toolbar.X+g.toolbar.Width, g.toolbar.Y+g.toolbar.Height))
	g.windowMgr.Draw(screen, g.r)
	g.windowMgr.MarkForeign()
	if g.scenarioSelect != nil {
		g.scenarioSelect.Draw(screen)
	}
	if g.browse != nil {
		g.browse.Draw(screen)
	}
	if g.options != nil {
		g.options.Draw(screen)
	}
	if g.buildVehicle != nil {
		g.buildVehicle.Draw(screen)
	}
	if g.vehicleWin != nil { // opened after the Build window: on top (PARTIAL: the window stack's real order)
		g.vehicleWin.FrameNo = g.windowFrame
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			g.vehicleWin.SkinImage = int(g.r.ObjMgr.InterfaceSkin.ImageOffset)
		}
		g.vehicleWin.Draw(screen)
		g.vehicleWin.DrawDrag(screen) // DragVehiclePart: stickToFront
	}
	if g.townList != nil {
		g.townList.Draw(screen)
	}

	// Draw dropdown menu on top of everything else
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.Draw(screen)
	}
	if g.textInput != nil { // stickToFront
		g.textInput.Draw(screen)
	}
	if g.prompt != nil { // a confirmation, on top of everything
		g.prompt.Draw(screen)
	}
	if g.errorWin != nil { // Error.cpp: stickToFront
		if g.errorWin.Expired() {
			g.errorWin = nil
		} else {
			g.errorWin.Draw(screen)
		}
	}

	statusY := g.sh - 20
	// Bottom bar: the ported PlayerInfoPanel (left) and TimePanel (right).
	g.drawPlayerPanel(screen)
	// The objective is not an overlay upstream: it is the time panel's date
	// tooltip (StringIds::tooltip_daymonthyear_challenge), shown on hover.
	g.drawTimePanel(screen)
	if g.gameWon {
		msg := "OBJECTIVE ACHIEVED!"
		mw, _ := ui.MeasureTextBold(msg)
		fillRect(screen, float64(g.sw/2-mw), float64(g.sh/2-30), float64(mw*2), 60, color.RGBA{0, 80, 0, 200})
		ui.DrawTextBoldScaled(screen, msg, g.sw/2-mw, g.sh/2-10, 2.0, color.RGBA{255, 255, 160, 255})
	}

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

	g.drawInlineDiag(screen)
}

// drawFPS is Gfx::drawFPS (Graphics/FPSCounter.cpp), drawn when
// Config::showFPS is on (Options → Display, or the test page's "Show FPS"):
// the frames per second measured over a second, "%.0f" (below 10: "%.1f"),
// medium bold, white outlined, centred at the top.
func (g *Game) drawFPS(screen *ebiten.Image) {
	if !g.showFPS {
		return
	}
	g.fpsFrames++
	now := time.Now()
	if g.fpsRef.IsZero() {
		g.fpsRef = now
	}
	if el := now.Sub(g.fpsRef).Seconds(); el > 1.0 {
		g.fpsCurrent = float64(g.fpsFrames) / el
		g.fpsFrames, g.fpsRef = 0, now
	}
	format := "%.0f"
	if g.fpsCurrent < 10 {
		format = "%.1f"
	}
	str := fmt.Sprintf(format, g.fpsCurrent)
	w, _ := ui.MeasureTextBold(str)
	wc, oc := ui.TextColours(2)
	ui.DrawTextOutlined(screen, str, g.sw/2-w/2, 2, wc, oc)
}

// debugPath prefixes capture filenames with the gitignored debug/ directory
// (created on demand) unless the caller already gave a path.
func debugPath(name string) string {
	if strings.ContainsRune(name, '/') {
		return name
	}
	_ = os.MkdirAll("debug", 0o755)
	return "debug/" + name
}

// maybeSaveShot implements --shots mode: save the full frame every shotsEvery,
// quit when the requested count has been captured. Script `shot` commands are
// honoured here too.
func (g *Game) maybeSaveShot(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TOWNS") != "" && g.townList == nil && !g.diagTownsDone && g.w != nil {
		g.diagTownsDone = true
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "stations" {
			g.openStationsWindow()
		} else if d == "vehicles" {
			g.openVehiclesWindow()
		} else {
			g.openTownsWindow()
		}
		g.townList.Hover = 2
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "scrolled" { // a click on the track below the thumb
			l := g.townList.List
			g.townList.HandleClick(g.townList.X+l[0]+l[2]-6, g.townList.Y+l[1]+l[3]-20)
		}
	}
	if d := os.Getenv("GOLOCO_DIAG_CONSTRUCTION"); (d == "window" || d == "road" || d == "extend" || d == "signal" || d == "overhead" || d == "station" || d == "sigdrop" || d == "bridgedrop") && !g.diagConsDone {
		g.diagConsDone = true
		if d == "signal" {
			g.consTab = 2
		}
		if d == "station" {
			g.consTab = 1
		}
		if d == "bridgedrop" { // the bridge list open
			defer func() {
				f, _ := g.constructionFrame(10, 40)
				g.openBridgeDropdown(10, 40, f.SkinColour, f.CompanyColour)
			}()
		}
		if d == "sigdrop" { // the signal tab with its dropdown open
			g.consTab = 2
			defer g.openConsDropdown(10, 40, g.r.ObjMgr.InterfaceSkin.WindowConstructionColour, false)
		}
		if d == "overhead" { // the tram road (its overhead wires), one ticked
			g.buildMode = buildModeRoad
			for range 8 {
				if st := g.overheadTabState(); st.Names[0] != "" {
					break
				}
				g.consCycleType(1)
			}
			g.consTab, g.consMods = 3, 1
		}
		if d == "extend" {
			g.buildMode = buildModeTrack
			mw, mh := g.w.GetMapSize()
			g.w.SetConstructionHead(mw/2, mh/2)
		}
		if d == "road" {
			g.buildMode = buildModeRoad
			if names := g.consObjectNames(); len(names) > 0 && names[0] == "" {
				g.consCycleType(1)
			}
		}
		g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
	}
	if v := os.Getenv("GOLOCO_DIAG_BUILDVEHICLE"); v != "" && !g.buildVehDiagDone && g.w != nil && ((v != "routeadd" && v != "routewaypoint" && v != "express" && v != "roadghost" && v != "shipghost" && v != "shipsail" && v != "planefly" && v != "planeghost" && v != "newindustry" && v != "newindustrybuild" && v != "waves" && v != "playable" && v != "towns" && v != "buildstation" && v != "stationtool" && v != "stationghost") || g.w.SceneReady()) { // the Build Vehicle window on tab N
		g.buildVehDiagDone = true
		if v == "buildstation" { // DIAGNOSTIC ONLY: CreateTrainStation on the first free straight track piece that takes one, the view on it
			m := &g.w.Map
			om := g.r.ObjMgr
			g1, _ := om.G1File.(sim.G1Raw)
			typ := uint8(0)
			for i, s := range om.TrainStationObjects {
				if s != nil {
					typ = uint8(i)
					break
				}
			}
			done := false
			for y := 4; y < m.Height-4 && !done; y++ {
				for x := 4; x < m.Width-4 && !done; x++ {
					tl := m.TileAt(x, y)
					for _, r := range tl.Order {
						if r.Kind != scenario.KindTrack {
							continue
						}
						tr := tl.Tracks[r.Index]
						if tr.TrackID != 0 || tr.HasStation || tr.HasSignal || tr.HasLevelCrossing || tr.Ghost || tr.AiAllocated || tr.HasBridge {
							continue
						}
						c := &sim.Cmd{Company: tr.Owner, Objects: om}
						a := sim.TrainStationPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tr.BaseZ) * 4, Rotation: tr.Rotation, TrackObjectID: tr.TrackObjectID, Type: typ}
						if _, ok := m.CreateTrainStation(c, 0, a, g1); !ok {
							continue
						}
						cost, ok := m.CreateTrainStation(c, sim.CmdApply, a, g1)
						g.inTitleScreen = false
						g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
						log.Printf("[Diag] buildstation: at %d,%d company %d: ok %v cost %d", x, y, tr.Owner, ok, cost)
						done = true
						break
					}
				}
			}
		} else if v == "roadstationtool" || v == "roadstationghost" { // DIAGNOSTIC ONLY: the Station tab's road tool over the first free straight road near a town, through picking (ghost: hover, leave, back, held for the shot; tool: clicked)
			m := &g.w.Map
			done := false
			for y := 4; y < m.Height-4 && !done; y++ {
				for x := 4; x < m.Width-4 && !done; x++ {
					near := false
					for _, t := range m.Towns {
						if dx, dy := int(t.X)/32-x, int(t.Y)/32-y; dx*dx+dy*dy <= 64 {
							near = true
						}
					}
					if !near {
						continue
					}
					tl := m.TileAt(x, y)
					for _, r := range tl.Order {
						if r.Kind != scenario.KindRoad {
							continue
						}
						rd := tl.Roads[r.Index]
						if rd.RoadID != 0 || rd.HasStationElement || rd.HasSignalElement || rd.HasLevelCrossing || rd.Ghost || rd.HasBridge {
							continue
						}
						g.inTitleScreen = false
						g.buildMode, g.consTab, g.buildObjID = buildModeRoad, 1, int(rd.RoadObjectID)
						g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
						sx, sy := g.w.TileToScreenPx(x, y)
						ghost := func() bool {
							for _, st := range m.TileAt(x, y).Stations {
								if st.Ghost && st.StationType == 1 {
									return true
								}
							}
							return false
						}
						if v == "roadstationghost" {
							g.stationToolUpdate(sx, sy+16)
							on := ghost()
							g.stationToolUpdate(5, 5)
							off := ghost()
							g.stationToolUpdate(sx, sy+16)
							log.Printf("[Diag] roadstationghost: tile %d,%d: ghost on hover %v, after leaving %v, back %v, cost %d, accepts %#x produces %#x", x, y, on, off, ghost(), g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
							g.diagHoldGhost = true
							g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
						} else {
							g.stationToolUp(sx, sy+16)
							has := false
							for _, st := range m.TileAt(x, y).Stations {
								has = has || (!st.Ghost && st.StationType == 1)
							}
							log.Printf("[Diag] roadstationtool: clicked tile %d,%d: stop now %v, error window %v", x, y, has, g.errorWin != nil)
						}
						done = true
						break
					}
				}
			}
		} else if v == "stationtool" || v == "stationghost" { // DIAGNOSTIC ONLY: the Station tab's tool clicked on the screen over the first free straight track piece of a company — through picking, as the player would (stationghost: near a town first, for the catchment lists)
			m := &g.w.Map
			done := false
			nearTown := func(x, y int) bool {
				for _, t := range m.Towns {
					if dx, dy := int(t.X)/32-x, int(t.Y)/32-y; dx*dx+dy*dy <= 64 {
						return true
					}
				}
				return false
			}
			for pass := 0; pass < 2 && !done; pass++ {
				for y := 4; y < m.Height-4 && !done; y++ {
					for x := 4; x < m.Width-4 && !done; x++ {
						if pass == 0 && (v != "stationghost" || !nearTown(x, y)) {
							continue
						}
						tl := m.TileAt(x, y)
						for _, r := range tl.Order {
							if r.Kind != scenario.KindTrack {
								continue
							}
							tr := tl.Tracks[r.Index]
							if tr.TrackID != 0 || tr.HasStation || tr.HasSignal || tr.HasLevelCrossing || tr.Ghost || tr.AiAllocated || tr.HasBridge {
								continue
							}
							g.inTitleScreen = false
							m.ControllingCompany = tr.Owner
							g.buildMode, g.consTab, g.buildObjID = buildModeTrack, 1, int(tr.TrackObjectID)
							g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
							sx, sy := g.w.TileToScreenPx(x, y)
							if v == "stationghost" { // hover: a ghost appears; off the track it goes; back on, it stays for the shot
								g.stationToolUpdate(sx, sy+16)
								ghost := func() bool {
									for _, st := range m.TileAt(x, y).Stations {
										if st.Ghost {
											return true
										}
									}
									return false
								}
								on := ghost()
								g.stationToolUpdate(5, 5)
								off := ghost()
								g.stationToolUpdate(sx, sy+16)
								log.Printf("[Diag] stationghost: tile %d,%d: ghost on hover %v, after leaving %v, back %v, cost %d", x, y, on, off, ghost(), g.stationGhostCost)
								g.diagHoldGhost = true // the ghost and the tab stay for the shot, its cost in the window
								g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
							} else {
								g.stationToolUp(sx, sy+16)
								log.Printf("[Diag] stationtool: clicked %d,%d (tile %d,%d company %d): station now %v", sx, sy+16, x, y, tr.Owner, m.TileAt(x, y).Tracks[r.Index].HasStation)
								g.openError(ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), "City Station"), i18n.Label(i18n.WrongTypeOfTrackRoad)) // the window, held for the shot
								g.errorWin.Hold = true
								log.Printf("[Diag] stationtool: error window %q", g.errorWin.Lines)
							}
							done = true
							break
						}
					}
				}
			}
		} else if v == "towns" { // DIAGNOSTIC ONLY: the view on the biggest town (GOLOCO_DIAG_ZOOM: at that zoom), for the town labels
			g.inTitleScreen = false
			m := &g.w.Map
			best := 0
			for i, t := range m.Towns {
				if t.Population > m.Towns[best].Population {
					best = i
				}
			}
			if z, err := strconv.Atoi(os.Getenv("GOLOCO_DIAG_ZOOM")); err == nil {
				g.w.SetZoom(z)
			}
			t := m.Towns[best]
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", int(t.X)/32, int(t.Y)/32))
			log.Printf("[Diag] towns: %s, frame %+v", m.TownName(t.Name, best), t.LabelFrame)
		} else if v == "playable" { // DIAGNOSTIC ONLY: 5P end to end through the game's commands, as the player — a stopped rail train with two stops lifted, a copy bought (same cars), placed where it stood, its orders given, started; the view follows it, its window on Main
			m := &g.w.Map
			om := g.r.ObjMgr
			trains, _ := m.Trains()
			for _, tr := range trains {
				oh := m.Entities[tr.Head]
				if oh.Mode != 0 || oh.TileX == -1 || len(tr.Cars) == 0 {
					continue
				}
				var orders [][]byte
				stops := 0
				for _, o := range m.OrderRing(oh.OrderTableOffset, 0) {
					if o.Type() == sim.OrderStopAt {
						stops++
					}
					orders = append(orders, append([]byte(nil), m.Orders[o.Offset:o.Offset+sim.OrderSize(int(o.Type()))]...))
				}
				c := &sim.Cmd{Company: oh.Owner, Objects: om}
				if stops < 2 {
					continue
				}
				v2 := m.Entities[tr.Veh2] // where it stands, before it is lifted
				place := sim.VehiclePlacement{X: int32(v2.TileX), Y: int32(v2.TileY), Z: int32(v2.TileBaseZ) * 4, TrackAndDirection: v2.TrackAndDirection, TrackProgress: v2.SubPosition}
				if _, ok := m.VehiclePickup(c, sim.CmdApply, oh.ID); !ok {
					continue // moving: can't be modified
				}
				m.ControllingCompany, m.PlayerCompanies[0] = oh.Owner, oh.Owner
				m.Companies[oh.Owner].Cash += 5_000_000 // funded: the title's companies are nearly broke
				var cars []uint16
				for _, car := range tr.Cars {
					cars = append(cars, m.Entities[car[0].Body].ObjectID)
				}
				var h uint16
				_, okBuy := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehiclePurchases, func(f uint8) (int32, bool) { return m.CreateVehicle(c, f, cars[0], true, &h) })
				for _, car := range cars[1:] {
					m.AddCarToVehicle(c, sim.CmdApply, car, h, true)
				}
				place.Head = h
				_, okPlace := m.VehiclePlace(c, sim.CmdApply, place)
				for _, raw := range orders {
					m.VehicleOrderInsert(c, sim.CmdApply, h, m.Entities[func() int { t, _ := m.TrainOf(h); return t.Head }()].SizeOfOrderTable-1, raw)
				}
				_, okStart := m.VehicleChangeRunningMode(c, sim.CmdApply, h, sim.RunningModeStart)
				g.inTitleScreen = false
				g.diagShip, g.diagFollow = h, true
				g.w.CentreOnVehicle(h)
				log.Printf("[Diag] playable: replaced train %d by bought %d (%d cars, %d orders): bought %v placed %v started %v", oh.ID, h, len(cars), len(orders), okBuy, okPlace, okStart)
				g.openVehicleWindow(h, 0)
				if g.vehicleWin != nil {
					g.vehicleWin.X, g.vehicleWin.Y = 8, 40
				}
				break
			}
		} else if v == "drag" { // DIAGNOSTIC ONLY: a stopped saved train's last car dragged over its second row (GOLOCO_DIAG_DRAG_DROP: dropped there)
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				h := m.Entities[tr.Head]
				if h.VehicleType != 0 || len(tr.Cars) < 3 {
					continue
				}
				m.ControllingCompany = h.Owner
				m.Entities[tr.Veh2].CurrentSpeed = 0
				m.Entities[tr.Head].VehicleFlags |= sim.VehicleFlagCommandStop
				g.inTitleScreen = false
				g.openVehicleWindow(h.ID, 1)
				w := g.vehicleWin
				order := func() []uint16 {
					t, _ := m.TrainOf(h.ID)
					var ids []uint16
					for _, car := range t.Cars {
						ids = append(ids, m.Entities[car[0].Front].ID)
					}
					return ids
				}
				before := order()
				listY := w.Y + 44 + 1 // the car list's inside (Details carList top 44)
				w.HandleClick(w.X+60, listY+(len(tr.Cars)-1)*22+11)
				w.DragTo(w.X+80, listY+22+11)
				w.MouseOver(w.X+80, listY+22+11) // scrollDrag: the drop line
				log.Printf("[Diag] drag: vehicle %d, dragging %v car %d", h.ID, w.Dragging, w.DragCar)
				if os.Getenv("GOLOCO_DIAG_DRAG_DROP") != "" {
					w.DragEnd(w.X+80, listY+22+11)
					if g.prompt != nil && g.prompt.OnOK != nil { // the cargo question: OK
						g.prompt.OnOK()
						g.prompt = nil
					}
					log.Printf("[Diag] drag: %v → %v", before, order())
				}
				break
			}
		} else if v == "sell" { // DIAGNOSTIC ONLY: as a saved train's company, a vehicle bought, its Details tab; GOLOCO_DIAG_SELL_OK: the bin pressed
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				h := m.Entities[tr.Head]
				if h.VehicleType != 0 || len(tr.Cars) == 0 {
					continue
				}
				m.ControllingCompany = h.Owner
				c := &sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}
				var nh uint16
				if _, ok := m.CreateVehicle(c, sim.CmdApply, m.Entities[tr.Cars[0][0].Body].ObjectID, true, &nh); !ok {
					continue
				}
				g.inTitleScreen = false
				g.openVehicleWindow(nh, 1)
				cash := m.Companies[h.Owner].Cash
				if os.Getenv("GOLOCO_DIAG_SELL_OK") != "" {
					g.vehicleWin.HandleClick(g.vehicleWin.X+g.vehicleWin.W-14, g.vehicleWin.Y+108) // the bin (right bar, 96–119)
					log.Printf("[Diag] sell: vehicle %d gone %v, window closed %v, cash %d → %d", nh, m.Entities != nil && func() bool { _, err := m.TrainOf(nh); return err != nil }(), g.vehicleWin == nil, cash, m.Companies[h.Owner].Cash)
				} else {
					log.Printf("[Diag] sell: vehicle %d, Details open", nh)
				}
				break
			}
		} else if v == "rename" { // DIAGNOSTIC ONLY: as a saved train's company, its window's caption clicked, " Express" typed
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				if h := m.Entities[tr.Head]; h.VehicleType == 0 {
					m.ControllingCompany = h.Owner
					g.inTitleScreen = false
					g.openVehicleWindow(h.ID, 3)
					g.vehicleWin.HandleClick(g.vehicleWin.X+40, g.vehicleWin.Y+6)
					if g.textInput != nil {
						g.textInput.Keys([]rune(" Express"), nil, false, false)
					}
					log.Printf("[Diag] rename: vehicle %d, text input open %v", h.ID, g.textInput != nil)
					if os.Getenv("GOLOCO_DIAG_RENAME_OK") != "" && g.textInput != nil { // Enter: OK
						g.textInput.Keys(nil, nil, true, false)
						g.textInput = nil
						nh := m.Entities[tr.Head]
						log.Printf("[Diag] rename: now %q, window %q", m.VehicleName(nh.Name, int(nh.OrdinalNumber)), g.vehicleWin.Title)
					}
					break
				}
			}
		} else if v == "refit" { // DIAGNOSTIC ONLY: as a saved train's company, a refittable vehicle bought, its Cargo tab, the refit dropdown
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				h := m.Entities[tr.Head]
				if h.VehicleType != 0 {
					continue
				}
				m.ControllingCompany, m.PlayerCompanies[0] = h.Owner, h.Owner // the diag plays as it (canRefit: a player's)
				c := &sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}
				for id, ob := range g.r.ObjMgr.Vehicles {
					var nh uint16
					if ob == nil || ob.Flags&(1<<14) == 0 {
						continue
					}
					if _, ok := m.CreateVehicle(c, sim.CmdApply, uint16(id), true, &nh); !ok {
						continue
					}
					g.inTitleScreen = false
					g.openVehicleWindow(nh, 2)
					x, y := g.vehicleWin.X+g.vehicleWin.W-14, g.vehicleWin.Y+56 // the refit button (right bar, 44–67)
					g.vehicleWin.HandleClick(x, y)
					log.Printf("[Diag] refit: vehicle %d (%s), dropdown %v", nh, ob.DisplayName, g.dropdown != nil)
					break
				}
				break
			}
		} else if v == "finances" || v == "cargo" { // DIAGNOSTIC ONLY: the first saved train with a last income (cargo: holding cargo), that tab
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				h := m.Entities[tr.Head]
				loaded := false
				for _, car := range tr.Cars {
					loaded = loaded || m.Entities[car[0].Body].Cargo.Qty != 0
				}
				if h.VehicleType == 0 && m.Entities[tr.Veh1].LastIncome.Day != -1 && (v == "finances" || loaded && len(tr.Cars) > 2) {
					g.inTitleScreen = false
					g.openVehicleWindow(h.ID, map[string]int{"finances": 3, "cargo": 2}[v])
					g.vehicleWin.MouseOver(g.vehicleWin.X+60, g.vehicleWin.Y+44+1+22+11) // the second row hovered
					log.Printf("[Diag] %s: vehicle %d", v, h.ID)
					break
				}
			}
		} else if v == "moneyeffect" { // DIAGNOSTIC ONLY: money popups at the screen's centre, one made whenever none is alive (they live 160 ticks)
			g.inTitleScreen = false
			g.w.Map.TitleMode = false
			g.diagMoney = true
		} else if v == "airportghost" { // DIAGNOSTIC ONLY: the airport window's tool over the first spot an airport takes (dry run), as the player; the ghost held, rotated once
			m := &g.w.Map
			m.ControllingCompany = m.PlayerCompanies[0]
			g.inTitleScreen = false
			g.openPortConstruction(0)
			g.airportRotation = 1
			c, g1 := g.stationCmd()
			done := false
			for y := 10; y < m.Height-10 && !done; y += 3 {
				for x := 10; x < m.Width-10 && !done; x += 3 {
					a, ok := m.AirportPlacementAt(int32(x)*32, int32(y)*32, uint8(g.consStation), g.airportRotation, g.r.ObjMgr)
					if !ok || m.TileAt(x, y).WaterLevel != 0 {
						continue
					}
					if _, ok := m.CreateAirport(c, 0, a, g1); !ok {
						continue
					}
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
					sx, sy := g.w.TileToScreenPx(x, y)
					g.stationToolUpdate(sx, sy+16)
					log.Printf("[Diag] airportghost: tile %d,%d rotation %d: ghost %v cost %d accepts %#x produces %#x", x, y, a.Rotation, g.airportGhost != nil, g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
					g.diagHoldGhost = true
					done = true
				}
			}
		} else if v == "dockghost" || v == "docktool" { // DIAGNOSTIC ONLY: the dock window's tool over the first shore spot a dock takes (dry run), as a station owner — hover (ghost held) or click
			m := &g.w.Map
			for i := range m.StationNames {
				if m.StationNames[i].Name != 0xFFFF {
					m.ControllingCompany = m.StationNames[i].Owner
					break
				}
			}
			g.inTitleScreen = false
			g.openPortConstruction(1)
			c, g1 := g.stationCmd()
			done := false
			for y := 3; y < m.Height-4 && !done; y++ {
				for x := 3; x < m.Width-4 && !done; x++ {
					a, ok := m.DockPlacementAt(int32(x)*32, int32(y)*32, uint8(g.consStation), g.r.ObjMgr)
					if !ok || m.TileAt(x, y).WaterLevel != 0 {
						continue
					}
					if _, ok := m.CreatePort(c, 0, a, g1); !ok {
						continue
					}
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
					sx, sy := g.w.TileToScreenPx(x, y)
					if v == "dockghost" {
						g.stationToolUpdate(sx, sy+16)
						log.Printf("[Diag] dockghost: tile %d,%d rotation %d: ghost %v cost %d accepts %#x produces %#x", x, y, a.Rotation, g.dockGhost != nil, g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
						g.diagHoldGhost = true
					} else {
						g.stationToolUp(sx, sy+16)
						docks := 0
						for _, st := range m.TileAt(x, y).Stations {
							if st.StationType == 3 && !st.Ghost {
								docks++
							}
						}
						log.Printf("[Diag] docktool: clicked tile %d,%d: dock elements %d, error window %v", x, y, docks, g.errorWin != nil)
					}
					done = true
				}
			}
		} else if v == "portmenu" { // DIAGNOSTIC ONLY: the toolbar's Port/Airport menu
			g.inTitleScreen = false
			for i := range g.toolbar.Buttons {
				if g.toolbar.Buttons[i].Tooltip == "Port/Airport" {
					g.handleToolbarButton(i)
				}
			}
			log.Printf("[Diag] portmenu: lastAirport %d lastShipPort %d, menu open %v", g.w.Map.LastAirport, g.w.Map.LastShipPort, g.dropdown != nil)
			if os.Getenv("GOLOCO_DIAG_PORT_DOCK") != "" { // choose Ship Port: the dock window (as a station owner: the title's company 0 is black)
				g.dropdown = nil
				for i := range g.w.Map.StationNames {
					if g.w.Map.StationNames[i].Name != 0xFFFF {
						g.w.Map.ControllingCompany = g.w.Map.StationNames[i].Owner
						break
					}
				}
				g.openPortConstruction(1)
				f, _ := g.constructionFrame(0, 0)
				c, _ := g.w.ControllingCompany()
				log.Printf("[Diag] portmenu: dock window, mode %d tab %d station %d, title %q raw %q, company %d colour %d", g.buildMode, g.consTab, g.consStation, f.Title, i18n.Str(i18n.TitleShipPort), c, f.CompanyColour)
			}
		} else if v == "menu" { // the toolbar's Build Vehicles menu instead
			for i := range g.toolbar.Buttons {
				if g.toolbar.Buttons[i].Tooltip == "Build Vehicles" {
					g.handleToolbarButton(i)
				}
			}
		} else if v == "routeadd" || v == "routewaypoint" { // the Route tool: click the first station (or plain track: a waypoint) found on screen
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				if h := m.Entities[tr.Head]; h.TileX != -1 && h.SizeOfOrderTable > 1 {
					m.ControllingCompany = h.Owner // the diag plays as the train's company (the title map's are AI)
					g.openVehicleWindow(h.ID, 4)
					log.Printf("[Diag] route tool on: %v (company %d)", g.routeTool, h.Owner)
					g.w.CentreOnVehicle(h.ID) // its stations in view
					done := false
					for sy := 40; sy < g.sh-40 && !done; sy += 8 {
						for sx := 0; sx < g.sw && !done; sx += 8 {
							hit, ok := g.w.RouteAt(sx, sy, h.Mode)
							if ok && (hit.Waypoint != nil) == (v == "routewaypoint") {
								log.Printf("[Diag] %+v under %d,%d — clicking", hit, sx, sy)
								g.routeToolDown(sx, sy)
								done = true
							}
						}
					}
					for _, l := range g.vehicleWin.Route {
						log.Printf("[Diag] route after the click: %q", l.Text)
					}
					break
				}
			}
		} else if v == "waves" { // DIAGNOSTIC ONLY: the view on the map's most open water (the tile with the most water within 6 tiles), for waves
			g.inTitleScreen = false
			m := &g.w.Map
			best, bx, by := -1, 0, 0
			for y := 6; y < m.Height-6; y += 2 {
				for x := 6; x < m.Width-6; x += 2 {
					n := 0
					for dy := -6; dy <= 6; dy += 2 {
						for dx := -6; dx <= 6; dx += 2 {
							if m.TileAt(x+dx, y+dy).WaterLevel != 0 {
								n++
							}
						}
					}
					if n > best {
						best, bx, by = n, x, y
					}
				}
			}
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", bx, by))
			log.Printf("[Diag] waves: view on tile (%d, %d), %d of 49 sampled tiles water", bx, by, best)
		} else if v == "newindustry" || v == "newindustrybuild" { // DIAGNOSTIC ONLY: a farm-type industry by the createIndustry command at a searched place (built at once, or — newindustrybuild — under construction), the view on it
			g.inTitleScreen = false
			m := &g.w.Map
			om := g.r.ObjMgr
			if id, x, y, ok := m.DiagNewFarmIndustry(om, v == "newindustry"); ok {
				g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x>>5, y>>5))
				log.Printf("[Diag] newindustry: industry %d built at tile (%d, %d)", id, x>>5, y>>5)
			} else {
				log.Printf("[Diag] newindustry: none could be built")
			}
		} else if v == "planeghost" { // DIAGNOSTIC ONLY: buy an aircraft for an airport's owner, its pickup tool on, hover over the airport until a ghost stands at a terminal
			m := &g.w.Map
			om := g.r.ObjMgr
			var sid uint16 = 0xFFFF
			for i, st := range m.StationNames {
				if st.Name != 0 && st.Flags&(1<<6) != 0 && st.Owner < 16 {
					sid = uint16(i)
					break
				}
			}
			var planeObj uint16 = 0xFFFF
			for i, v := range om.Vehicles {
				if v != nil && v.Mode == objects.TransportModeAir {
					planeObj = uint16(i)
					break
				}
			}
			if sid == 0xFFFF || planeObj == 0xFFFF {
				log.Printf("[Diag] planeghost: no airport or aircraft object")
			} else {
				st := m.StationNames[sid]
				m.ControllingCompany = st.Owner
				g.inTitleScreen = false
				g.w.CenterOn(fmt.Sprintf("tile:%d,%d", int(st.AirportStartPos[0])/32+2, int(st.AirportStartPos[1])/32+2))
				var head uint16
				if _, ok := m.CreateVehicle(&sim.Cmd{Company: st.Owner, Objects: om}, sim.CmdApply, planeObj, true, &head); !ok {
					log.Printf("[Diag] planeghost: buying aircraft %d refused", planeObj)
				} else {
					g.openVehicleWindow(head, 1)
					if x, y, ok := g.vehicleWin.PickupButton(); ok {
						g.vehicleWin.HandleClick(x, y)
					}
					if g.vehicleWin != nil {
						g.vehicleWin.X, g.vehicleWin.Y = 4, 40
					}
					for r := 0; r < 200 && !m.VehicleIsGhost(head); r += 4 {
						for _, o := range [][2]int{{r, r / 2}, {-r, r / 2}, {r, -r / 2}, {-r, -r / 2}, {0, r}, {0, -r}} {
							hx, hy := g.sw/2+o[0], g.sh/2+o[1]
							if g.overWindow(hx, hy) {
								continue
							}
							g.pickupToolUpdate(hx, hy)
							if m.VehicleIsGhost(head) {
								g.diagMouse = &[2]int{hx, hy}
								break
							}
						}
					}
					trh, _ := m.TrainOf(head)
					h := m.Entities[trh.Head]
					log.Printf("[Diag] planeghost: station %d, aircraft %d (object %d); pickup tool %v; a ghost now %v at %d,%d,%d edge %d; mouse %v",
						sid, head, planeObj, g.pickupOn, m.VehicleIsGhost(head), h.PosX, h.PosY, h.PosZ, h.AirportMovementEdge, g.diagMouse)
				}
			}
		} else if v == "planefly" { // DIAGNOSTIC ONLY: follow the title map's aircraft at an airport (its status logged)
			g.inTitleScreen = false
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				if h := m.Entities[tr.Head]; h.Mode == 2 && h.TileX != -1 && h.StationID != 0xFFFF {
					g.diagShip, g.diagFollow = h.ID, true
					g.w.CentreOnVehicle(h.ID)
					log.Printf("[Diag] planefly: aircraft %d at station %d", h.ID, h.StationID)
					break
				}
			}
		} else if v == "shipsail" { // DIAGNOSTIC ONLY: two docks (at the water's height, in their stations' tiles) on one lake, a ship placed at the first with "stop at" the second, started — watch it sail
			m := &g.w.Map
			om := g.r.ObjMgr
			var shipObj uint16 = 0xFFFF
			for i, v := range om.Vehicles {
				if v != nil && v.Mode == objects.TransportModeWater {
					shipObj = uint16(i)
					break
				}
			}
			var company uint8 = 0xFF
			var stations []uint16
			for i, st := range m.StationNames {
				if st.Name != 0 && st.Owner < 16 && (company == 0xFF || st.Owner == company) && int(st.TileSize) < len(st.Tiles) {
					company = st.Owner
					stations = append(stations, uint16(i))
					if len(stations) == 2 {
						break
					}
				}
			}
			if shipObj == 0xFFFF || len(stations) < 2 || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
				log.Printf("[Diag] shipsail: no ship object, two stations of a company, or dock object")
			} else {
				d := om.DockObjects[0]
				ox, oy := sim.RotateOffset(d.BoatPosition[0], d.BoatPosition[1], 0)
				var docks [][3]int32
				var level uint8
				for y := 4; y < m.Height-4 && len(docks) < 2; y++ {
					for x := 4; x < m.Width-4 && len(docks) < 2; x++ {
						t := m.TileAt(x, y)
						bx, by := int32(x)*32+int32(ox)+32, int32(y)*32+int32(oy)+32
						bt := m.TileAt(int(bx>>5), int(by>>5))
						if t == nil || bt == nil || t.WaterLevel != 0 || bt.WaterLevel == 0 || len(t.Stations) > 0 || (level != 0 && bt.WaterLevel != level) {
							continue
						}
						if len(docks) == 1 && (abs(int(int32(x)*32-docks[0][0]))+abs(int(int32(y)*32-docks[0][1])) < 6*32 || abs(int(int32(x)*32-docks[0][0]))+abs(int(int32(y)*32-docks[0][1])) > 14*32) {
							continue
						}
						level = bt.WaterLevel
						docks = append(docks, [3]int32{int32(x) * 32, int32(y) * 32, int32(bt.WaterLevel) * 16})
					}
				}
				if len(docks) < 2 {
					log.Printf("[Diag] shipsail: no two shore tiles on one water")
				} else {
					for i, dk := range docks {
						tl := m.TileAt(int(dk[0]>>5), int(dk[1]>>5))
						tl.Stations = append(tl.Stations, scenario.StationElement{StationType: 3, StationID: stations[i], Owner: company, BaseZ: uint8(dk[2] / 4)})
						st := &m.StationNames[stations[i]]
						st.Tiles[st.TileSize] = [3]int16{int16(dk[0]), int16(dk[1]), int16(dk[2])}
						st.TileSize++
					}
					m.ControllingCompany = company
					g.inTitleScreen = false
					own := &sim.Cmd{Company: company, Objects: om}
					var head uint16
					_, okBuy := m.CreateVehicle(own, sim.CmdApply, shipObj, true, &head)
					_, okPlace := m.VehiclePlaceWater(own, sim.CmdApply, sim.VehicleWaterPlacement{X: docks[0][0], Y: docks[0][1], Z: docks[0][2], Head: head})
					_, okOrder := m.VehicleOrderInsert(own, sim.CmdApply, head, 0, sim.OrderStopAtRaw(int(stations[1])))
					_, okStart := m.VehicleChangeRunningMode(own, sim.CmdApply, head, sim.RunningModeStart)
					mx, my := (docks[0][0]+docks[1][0])/2, (docks[0][1]+docks[1][1])/2
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", mx>>5, my>>5))
					g.diagShip = head
					log.Printf("[Diag] shipsail: docks at tiles %d,%d and %d,%d (water level %d, stations %d → %d, company %d); ship %d bought %v, placed %v, ordered %v, started %v",
						docks[0][0]>>5, docks[0][1]>>5, docks[1][0]>>5, docks[1][1]>>5, level, stations[0], stations[1], company, head, okBuy, okPlace, okOrder, okStart)
				}
			}
		} else if v == "shipghost" { // DIAGNOSTIC ONLY: the title map has no docks or boats — add a dock on water near the view's centre, buy a ship, hover over it
			m := &g.w.Map
			om := g.r.ObjMgr
			var shipObj uint16 = 0xFFFF
			for i, v := range om.Vehicles {
				if v != nil && v.Mode == objects.TransportModeWater {
					shipObj = uint16(i)
					break
				}
			}
			var company uint8 = 0xFF
			var stationID uint16
			for i, st := range m.StationNames {
				if st.Name != 0 && st.Owner < 16 {
					company, stationID = st.Owner, uint16(i)
					break
				}
			}
			if shipObj == 0xFFFF || company == 0xFF || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
				log.Printf("[Diag] shipghost: no ship object, company station or dock object")
			} else {
				m.ControllingCompany = company
				g.inTitleScreen = false
				d := om.DockObjects[0]
				ox, oy := sim.RotateOffset(d.BoatPosition[0], d.BoatPosition[1], 0)
				var tx, ty int = -1, -1 // a tile whose dock boat position lies on water
				for y := 4; y < m.Height-4 && tx < 0; y++ {
					for x := 4; x < m.Width-4; x++ {
						t := m.TileAt(x, y)
						if _, water := m.GetHeight(int32(x)*32+int32(ox)+32, int32(y)*32+int32(oy)+32); water != 0 && t != nil && len(t.Stations) == 0 {
							if _, here := m.GetHeight(int32(x)*32, int32(y)*32); here == 0 { // the dock itself on the shore
								tx, ty = x, y
								break
							}
						}
					}
				}
				if tx < 0 {
					log.Printf("[Diag] shipghost: no shore tile with water at the boat position")
				} else {
					t := m.TileAt(tx, ty)
					t.Stations = append(t.Stations, scenario.StationElement{StationType: 3, StationID: stationID, Owner: company, BaseZ: t.BaseZ})
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", tx, ty))
					var head uint16
					if _, ok := m.CreateVehicle(&sim.Cmd{Company: company, Objects: om}, sim.CmdApply, shipObj, true, &head); !ok {
						log.Printf("[Diag] shipghost: buying ship %d refused", shipObj)
					} else {
						g.openVehicleWindow(head, 1)
						if x, y, ok := g.vehicleWin.PickupButton(); ok {
							g.vehicleWin.HandleClick(x, y)
						}
						if g.vehicleWin != nil { // out of the way of the dock at the centre (the click re-opened it centred)
							g.vehicleWin.X, g.vehicleWin.Y = 4, 40
						}
						for r := 0; r < 120 && !m.VehicleIsGhost(head); r += 4 {
							for _, o := range [][2]int{{r, r / 2}, {-r, r / 2}, {r, -r / 2}, {-r, -r / 2}, {0, r}, {0, -r}} {
								hx, hy := g.sw/2+o[0], g.sh/2+o[1]
								if g.overWindow(hx, hy) {
									continue
								}
								g.pickupToolUpdate(hx, hy)
								if m.VehicleIsGhost(head) {
									g.diagMouse = &[2]int{hx, hy}
									break
								}
							}
						}
						tr, _ := m.TrainOf(head)
						v2 := m.Entities[tr.Veh2]
						log.Printf("[Diag] shipghost: dock added at tile %d,%d (station %d, company %d), ship %d bought; pickup tool %v; a ghost now %v at %d,%d,%d; mouse %v",
							tx, ty, stationID, company, head, g.pickupOn, m.VehicleIsGhost(head), v2.PosX, v2.PosY, v2.PosZ, g.diagMouse)
					}
				}
			}
		} else if v == "roadghost" { // DIAGNOSTIC ONLY: lift a road vehicle, its pickup tool on, hover until a placement — the ghost left standing for the shot
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				h := m.Entities[tr.Head]
				v2 := m.Entities[tr.Veh2]
				// a vehicle of the ordinary streets (road object 1, or any road type): trams need their own tracks
				if h.TileX == -1 || h.Mode != 1 || int(h.Owner) >= len(m.Companies) || (h.TrackType != 1 && h.TrackType != 0xFF) {
					continue
				}
				if _, ok := m.VehiclePickup(&sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}, sim.CmdApply, h.ID); !ok {
					continue
				}
				m.ControllingCompany = h.Owner
				g.inTitleScreen = false // windows draw over the title world
				g.w.CenterOn(fmt.Sprintf("tile:%d,%d", v2.TileX/32, v2.TileY/32))
				g.openVehicleWindow(h.ID, 1)
				g.vehicleWin.X, g.vehicleWin.Y = 8, 40 // out of the way of the road at the centre
				if x, y, ok := g.vehicleWin.PickupButton(); ok {
					g.vehicleWin.HandleClick(x, y)
				}
				log.Printf("[Diag] road vehicle %d of company %d lifted at tile %d,%d; pickup tool on: %v", h.ID, h.Owner, v2.TileX/32, v2.TileY/32, g.pickupOn)
				for d := 0; d < 200 && g.pickupOn && !m.VehicleIsGhost(h.ID); d += 4 { // outward from the centre
					for _, o := range [][2]int{{d, 0}, {-d, 0}, {0, d}, {0, -d}, {d, d}, {-d, -d}} {
						sx, sy := g.sw/2+o[0], g.sh/2+o[1]
						if g.overWindow(sx, sy) {
							continue
						}
						if p, ok := g.w.PlacementAt(h.ID, sx, sy); ok {
							g.pickupToolUpdate(sx, sy)
							g.diagMouse = &[2]int{sx, sy} // the mouse stays there: the frame's own pickupToolUpdate keeps the ghost
							log.Printf("[Diag] hover at %d,%d: placement %+v; a ghost now: %v", sx, sy, p, m.VehicleIsGhost(h.ID))
							break
						}
					}
				}
				break
			}
		} else if v == "express" { // the Route tab of the first placed train, as its owner: a click on Express (then left open)
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				if h := m.Entities[tr.Head]; h.TileX != -1 {
					m.ControllingCompany = h.Owner // DIAGNOSTIC ONLY: play as its owner
					g.inTitleScreen = false        // DIAGNOSTIC ONLY: draw windows over the title world (it stops cycling maps)
					g.openVehicleWindow(h.ID, 4)
					log.Printf("[Diag] train %d express before: %v", h.ID, g.vehicleWin.Express)
					lw := ui.VehicleRouteLayout(g.vehicleWin.W, g.vehicleWin.H, g.vehicleWin.State.HiddenTabs, false)
					e := lw[ui.VrExpressMode]
					g.vehicleWin.HandleClick(g.vehicleWin.X+(e.L+e.R)/2, g.vehicleWin.Y+(e.T+e.B)/2)
					log.Printf("[Diag] train %d express after the click: %v (veh1 var_48 %#x)", h.ID, g.vehicleWin.Express, m.Entities[tr.Veh1].Var48)
					break
				}
			}
		} else if v == "route" { // the Route tab of the first train with orders
			m := &g.w.Map
			trains, _ := m.Trains()
			for _, tr := range trains {
				if h := m.Entities[tr.Head]; h.TileX != -1 && h.SizeOfOrderTable > 1 {
					g.openVehicleWindow(h.ID, 4)
					for _, l := range g.vehicleWin.Route {
						log.Printf("[Diag] route of %d: %q current %v", h.ID, l.Text, l.Current)
					}
					if g.vehicleWin.OnRouteSelect != nil { // select the first order, move it down
						g.vehicleWin.OnRouteSelect(0)
						g.vehicleWin.OnRoute("down")
						for _, l := range g.vehicleWin.Route {
							log.Printf("[Diag] after down: %q", l.Text)
						}
						log.Printf("[Diag] selection now row %d", g.routeSel)
					}
					break
				}
			}
		} else if v == "main" { // buy a train, then the Vehicle window's Main tab (clicked)
			g.openBuildVehicle(0)
			if w := g.buildVehicle; w != nil {
				w.HandleClick(w.X+3+1+5, w.Y+102+1+5)
			}
			g.buildVehicle = nil
			if vw := g.vehicleWin; vw != nil {
				vw.HandleClick(vw.X+3+15, vw.Y+15+13) // the Main tab
				log.Printf("[Diag] vehicle window tab %d, flag image %d", g.vehicleWin.Tab, g.vehicleWin.StopStartImage)
			}
		} else if v == "buy" || v == "place" { // trains, then a click on the first row
			if v == "place" { // DIAGNOSTIC ONLY: lift a rail train that may be lifted, play as its owner, look at its freed track
				m := &g.w.Map
				trains, _ := m.Trains()
				for _, tr := range trains {
					h := m.Entities[tr.Head]
					v2 := m.Entities[tr.Veh2]
					if h.TileX == -1 || h.Mode != 0 || int(h.Owner) >= len(m.Companies) {
						continue
					}
					if _, ok := m.VehiclePickup(&sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}, sim.CmdApply, h.ID); !ok {
						continue
					}
					log.Printf("[Diag] lifted train %d of company %d at tile %d,%d; controlling company %d → %d", h.ID, h.Owner, v2.TileX/32, v2.TileY/32, m.ControllingCompany, h.Owner)
					m.ControllingCompany = h.Owner
					m.Companies[h.Owner].Cash += 10_000_000 // diagnostic: the title map's companies are nearly broke
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", v2.TileX/32, v2.TileY/32))
					break
				}
			}
			g.openBuildVehicle(0)
			if w := g.buildVehicle; w != nil {
				w.HandleClick(w.X+3+1+5, w.Y+102+1+5)
			}
			if v == "place" && g.vehicleWin != nil { // then the pickup button, and the first track under a 16-px grid
				g.buildVehicle = nil
				if x, y, ok := g.vehicleWin.PickupButton(); ok {
					g.vehicleWin.HandleClick(x, y)
				}
				log.Printf("[Diag] pickup tool on: %v", g.pickupOn)
				placed := false
				for sy := 40; sy < g.sh-40 && !placed && g.pickupOn; sy += 16 {
					for sx := 16; sx < g.sw-16 && !placed; sx += 16 {
						if g.vehicleWin.Contains(sx, sy) {
							continue
						}
						if p, ok := g.w.RailPlacementAt(g.pickupHead, sx, sy); ok && !g.diagLogged {
							g.diagLogged = true
							m := &g.w.Map
							tr, _ := m.TrainOf(g.pickupHead)
							h := m.Entities[tr.Head]
							log.Printf("[Diag] first placement %+v; the train: trackType %d, needs mods %#x", p, h.TrackType, h.HeadVar53)
							if t := m.TileAt(int(p.X)>>5, int(p.Y)>>5); t != nil {
								for _, el := range t.Tracks {
									log.Printf("[Diag]   tile track: id %d rot %d z %d obj %d owner %d mods %#x ai %v ghost %v", el.TrackID, el.Rotation, int(el.BaseZ)*4, el.TrackObjectID, el.Owner, el.Mods, el.AiAllocated, el.Ghost)
								}
							}
						}
						g.pickupToolUpdate(sx, sy) // the hover first: a ghost where the click will place
						ghost := g.w.Map.VehicleIsGhost(g.pickupHead)
						if placed = g.pickupToolDown(sx, sy); placed {
							log.Printf("[Diag] placed at screen %d,%d (a ghost there before the click: %v; still a ghost: %v); tool on: %v, window placed: %v",
								sx, sy, ghost, g.w.Map.VehicleIsGhost(g.pickupHead), g.pickupOn, g.vehicleWin.State.Placed)
							if x, y, ok := g.vehicleWin.StopStartButton(); ok { // then Main's stop / start → Start
								flag := g.vehicleWin.StopStartImage
								g.vehicleWin.HandleClick(x, y)
								if g.dropdown != nil && g.dropdown.OnSelect != nil {
									g.dropdown.OnSelect(1)
									g.dropdown = nil
								}
								tr, _ := g.w.Map.TrainOf(g.pickupHead)
								log.Printf("[Diag] Main tab %v; stop / start: flag %d → %d, commandStop now %v", g.vehicleWin.Tab == 0, flag, g.vehicleWin.StopStartImage,
									g.w.Map.Entities[tr.Head].VehicleFlags&sim.VehicleFlagCommandStop != 0)
								if x, y, ok := g.vehicleWin.CentreViewButton(); ok { // then centre view → focus
									bx, by := g.tileUnderCursor(g.sw/2, g.sh/2)
									g.vehicleWin.HandleClick(x+12, y-12)
									if g.dropdown != nil && g.dropdown.OnSelect != nil {
										g.dropdown.OnSelect(1)
										g.dropdown = nil
									}
									ax, ay := g.tileUnderCursor(g.sw/2, g.sh/2)
									v2 := g.w.Map.Entities[tr.Veh2]
									log.Printf("[Diag] focus on the train: screen centre tile %d,%d → %d,%d (the train's veh2 at tile %d,%d)", bx, by, ax, ay, v2.TileX/32, v2.TileY/32)
								}
							}
						}
					}
				}
			}
		} else {
			tab, _ := strconv.Atoi(v)
			g.openBuildVehicle(tab)
		}
	}
	if os.Getenv("GOLOCO_DIAG_VEHLIST") != "" && !g.vehDiagDone && g.w != nil { // the vehicle list (trains of their first owner)
		g.vehDiagDone = true
		g.openVehicleList(0)
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "cargo" && g.townList != nil { // transporting cargo, the cargo dropdown open
			w := g.townList
			w.HandleClick(w.X+270, w.Y+205)
			if g.dropdown != nil && g.dropdown.OnSelect != nil {
				g.dropdown.OnSelect(1)
				g.dropdown = nil
			}
			w.HandleClick(w.X+400, w.Y+205)
		}
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "company" && g.townList != nil { // the company select dropdown open
			w := g.townList
			w.HandleClick(w.X+w.W-20, w.Y+20)
		}
	}
	if os.Getenv("GOLOCO_DIAG_OPTIONS") != "" && g.options == nil && !g.diagOptionsDone {
		g.diagOptionsDone = true
		g.openOptionsWindow()
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "controls" && g.options != nil {
			g.options.Tab = ui.OptTabControls
			g.options.Controls = [3]bool{true, false, true}
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regional" && g.options != nil {
			g.options.Tab = ui.OptTabRegional
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "framesolid" && g.options != nil {
			ui.FrameStyle = 1 // the Rendering tab with solid frames
			g.options.Tab = ui.OptTabRendering
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "prompt" { // the replace-file confirmation
			g.options = nil
			g.prompt = ui.NewPromptOkCancel("Save Game", "Replace existing file: “Bottleneck Blues”?", "Replace", g.sw, g.sh, nil)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "renderscale" && g.options != nil {
			g.options.Tab = ui.OptTabRendering // and a click on Vehicles min. display scale
			g.options.HandleClick(g.options.X+240, g.options.Y+66)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "framelimit" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // and a click on Frame rate limit
			g.options.HandleClick(g.options.X+240, g.options.Y+114)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "displayres" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // in fullscreen mode, a click on Display Resolution
			g.options.ScreenMode = 1
			g.options.HandleClick(g.options.X+240, g.options.Y+82)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "displaymode" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // and a click on Screen mode
			g.options.HandleClick(g.options.X+240, g.options.Y+66)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalcurrency" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Preferred currency
			g.options.HandleClick(g.options.X+190, g.options.Y+122)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalcurrent" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Current currency (as in a game)
			g.options.TitleMode = false
			g.options.HandleClick(g.options.X+190, g.options.Y+107)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalalways" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // in a game, a click on 'always' (the current currency greys)
			g.options.TitleMode = false
			g.options.HandleClick(g.options.X+100, g.options.Y+153)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionallang" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Language
			g.options.HandleClick(g.options.X+190, g.options.Y+52)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalheights" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Heights
			g.options.HandleClick(g.options.X+190, g.options.Y+87)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionaldrop" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Distance and Speed
			g.options.HandleClick(g.options.X+190, g.options.Y+72)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "display" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "company" && g.options != nil {
			g.options.Tab = ui.OptTabCompany
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "rendering" && g.options != nil {
			g.options.Tab = ui.OptTabRendering
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "misc" && g.options != nil {
			g.options.Tab = ui.OptTabMisc
		}
	}
	if os.Getenv("GOLOCO_DIAG_BROWSE") != "" && g.browse == nil && !g.diagBrowseDone {
		g.diagBrowseDone = true
		if os.Getenv("GOLOCO_DIAG_BROWSE") == "preview" { // a synthetic saved-game preview on the first file
			g.openLoadGameWindow()
			if g.browse != nil {
				img := make([]byte, 250*200)
				for i := range img {
					img[i] = byte(10 + (i%250)/25*10 + (i/250)/50) // bands of palette colours
				}
				g.browse.Preview = &scenario.SaveDetails{Company: "Joop's Railways", Owner: "Joop", Date: 55000, ChallengeProgress: 42, Image: img}
				g.browse.ChangeDir(filepath.Join(g.browse.Dir, "Scenarios"))
				g.browse.SelectFirstFile()
			}
		} else if os.Getenv("GOLOCO_DIAG_BROWSE") == "save" {
			g.openSaveWindow()
		} else {
			g.openFileWindow("Load Game", false)
		}
	}
	if os.Getenv("GOLOCO_DIAG_NEWGAME") != "" && g.scenarioSelect == nil && !g.diagNewGameDone {
		// Diagnostic: open the ScenarioSelect window with its first scenario selected.
		g.diagNewGameDone = true
		g.openNewGameWindow()
		g.scenarioSelect.SelectFirst()
		if os.Getenv("GOLOCO_DIAG_NEWGAME") == "many" {
			g.scenarioSelect.DiagRepeat(3)
			g.scenarioSelect.Tab = 3
			g.scenarioSelect.HandleWheel(-4)
		}
	}
	if os.Getenv("GOLOCO_DIAG_TIMEPANEL") != "" {
		st := ui.TimePanelState{Date: uilayout.DateDMY(1, "January", 1950), Speed: 1}
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			st.SkinColour = g.r.ObjMgr.InterfaceSkin.TimeToolbarColour
			st.SkinImage = int(g.r.ObjMgr.InterfaceSkin.ImageOffset)
		}
		ui.DrawTimePanel(screen, st)
		ps := ui.PlayerPanelState{Cash: 2500000}
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			ps.SkinColour = g.r.ObjMgr.InterfaceSkin.PlayerInfoToolbarColour
		}
		g.playerPanelCompany(&ps)
		ui.DrawPlayerPanel(screen, ps)
	}
	if os.Getenv("GOLOCO_DIAG_WINDOW") != "" {
		// Diagnostic: an upstream-style window frame + caption (FrameWidget,
		// CaptionWidget ports) for headless shots; the colour is the skin's window colour.
		var c uint8 = 1
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			c = g.r.ObjMgr.InterfaceSkin.WindowColour
		}
		ui.DrawWindowFrame(screen, 200, 60, 360, 220, c, false)
		ui.DrawWindowCaption(screen, 201, 61, 358, "Select Scenario")
	}
	if os.Getenv("GOLOCO_DIAG_CONSTRUCTION") == "frame" && g.r != nil && g.r.ObjMgr != nil {
		// Diagnostic: the ported construction frame for the first track object.
		f := &ui.ConstructionFrame{X: 10, Y: 40, Title: "Railway", FrameNo: g.titleFrame}
		if sk := g.r.ObjMgr.InterfaceSkin; sk != nil {
			f.SkinColour = sk.WindowConstructionColour
		}
		if len(g.w.CompanyColours) > 0 {
			f.CompanyColour = g.w.CompanyColours[0]
		}
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			f.TrackImage = int(g.r.ObjMgr.TrackObjects[0].Image)
			f.Title = g.r.ObjMgr.TrackObjects[0].Name
		}
		f.Draw(screen)
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			ui.DrawConstructionTab(screen, f, ui.ConstructionTabState{Traits: g.r.ObjMgr.TrackObjects[0].TrackPieces,
				Piece: "straight", Gradient: "level"})
		}
		f.DrawCaption(screen)
	}
	if os.Getenv("GOLOCO_DIAG_PROGRESS") != "" {
		// Diagnostic: the ProgressBar window at 3/4 in both styles, for headless shots.
		ui.DrawProgressBar(screen, "Loading...", 190, 0)
	}
	if os.Getenv("GOLOCO_DIAG_TOOLTIP") != "" {
		// Diagnostic: a tooltip over the frame (ToolTip.cpp port), for headless shots.
		var c uint8
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			c = g.r.ObjMgr.InterfaceSkin.TooltipColour
		}
		ui.DrawTooltip(screen, "Change viewport settings", 380, 120, c)
		ui.DrawTooltip(screen, "Build tracks, stations, signals and train depots, and place them on the map", 560, 240, c)
	}
	if os.Getenv("GOLOCO_DIAG_DROPDOWN") != "" {
		// Diagnostic: a toolbar-style dropdown on every frame (drawn last, just
		// before the capture), for headless screenshots of the Dropdown.cpp port.
		d := ui.NewDropdownMenu(40, 28, []ui.DropdownItem{{Text: "Load Game"}, {Text: "Save Game"}, {Separator: true},
			{Text: "Screenshot"}, {Separator: true}, {Text: "About", Disabled: true}, {Text: "Options"}, {Separator: true},
			{Text: "Quit to Menu"}, {Text: "Quit to Desktop"}})
		d.Highlighted = 1
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			d.SetToolbarColour(g.r.ObjMgr.InterfaceSkin.TopToolbarPrimaryColour)
		}
		d.Draw(screen)
	}
	if name := g.scriptShotName(); name != "" {
		name = debugPath(name)
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
	name := debugPath(fmt.Sprintf("shot_%02d.png", g.shotsN))
	if paint.Dump != nil { // GOLOCO_SHOT_PAINTDUMP: rtX order type image x y w h minX minY minZ maxX maxY maxZ quadrant
		os.WriteFile(strings.TrimSuffix(name, ".png")+".paint.txt", []byte(strings.Join(*paint.Dump, "\n")+"\n"), 0o644)
		paint.Dump = nil
	}
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
	var logoRect image.Rectangle
	defer func() { // what the title menu draws, for draw8's screen copy (pkg/ui translucent.go)
		ui.MarkForeign(logoRect)                                                     // the logo
		ui.MarkForeign(image.Rect(g.sw-titleOptionsW, 0, g.sw, 30))                  // Options
		ui.MarkForeign(image.Rect(0, g.sh-titleMenuH-titleMenuMargin-4, g.sw, g.sh)) // the globes, the corner texts
	}()
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
		logoRect = logo.Bounds().Add(image.Pt(int(xOff), int(yOff)))
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
		logoRect = image.Rect(0, 0, 298, subY)
	}

	_ = subY // upstream draws no subtitle under the logo (TitleLogo.cpp)

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
	optLabel := i18n.Label(i18n.Options) // TitleOptions.cpp: StringIds::options
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
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSparkle, i18n.Label(i18n.TitleMenuNewGame), 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSave, i18n.Label(i18n.TitleMenuLoadGame), 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteLessonL, i18n.Label(i18n.TitleMenuShowTutorial), 32},
		// globe_construct has 25 frames (0–24 = sprites 3584–3608); clamping to 25
		// prevents overshooting into unrelated sprites (e.g. the Sawyer logo).
		{spriteGlobeConstBase, spriteGlobeConstIdle, -1, i18n.Label(i18n.TitleMenuScenarioEditor), 25},
	}

	menuX := (g.sw - titleMenuW) / 2
	menuY := g.sh - titleMenuH - titleMenuMargin

	tip := ""
	for i, btn := range btns {
		bx := menuX + i*titleBtnSize
		by := menuY

		hovered := g.mouseX >= bx && g.mouseX < bx+titleBtnSize &&
			g.mouseY >= by && g.mouseY < by+titleBtnSize
		if hovered {
			tip = btn.label // the button's tooltip (TitleMenu.cpp's title_menu_* StringIds)
		}

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
	g.titleTip = tip

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
	// TitleExit.cpp: title_exit_game, centred and wrapped — its {NEWLINE}
	// splits it ("Exit / Game", "Spiel / beenden").
	exitLines := strings.Split(i18n.Label(i18n.TitleExitGame), "{NEWLINE}")
	ey := exitY + (titleExitH-10*len(exitLines))/2
	for _, l := range exitLines {
		lw, _ := ui.MeasureText(l)
		drawOutlinedText(screen, l, exitX+(titleExitW-lw)/2, ey, color.White)
		ey += 10
	}

	// --- Version text: bottom-left ---
	// TitleVersion.cpp: the version and the platform, white with the outline
	// style, at (8, height - 30) and 12 px below.
	{
		wc, oc := ui.TextColours(2) // Colour::white
		ui.DrawTextOutlined(screen, "GoLoco v0.1", 8, g.sh-30, wc, oc)
		ui.DrawTextOutlined(screen, runtime.GOOS+" ("+runtime.GOARCH+")", 8, g.sh-30+12, wc, oc)
	}
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
	g.objective = sc.Objective
	g.hasObjective = sc.HasObjective
	g.gameWon = false

	if g.objMgr != nil {
		reorderScenarioObjects(g.objMgr, sc)
		// Scenario::start (a .SC5: a new game) runs loadPreferredCurrencyNewGame;
		// loading a saved game (S5 import) runs loadPreferredCurrencyAlways.
		if strings.EqualFold(filepath.Ext(filePath), ".sc5") {
			g.loadPreferredCurrencyNewGame()
		} else {
			g.loadPreferredCurrencyAlways()
		}
	}

	// Create a fresh world for the new scenario so the title world is preserved.
	oldRng := g.w.Map.Rng1 // loadAndStart: the rng before the load
	gameW := world.NewWorld(g.r)
	gameW.LoadFromScenario(sc)
	if strings.EqualFold(filepath.Ext(filePath), ".sc5") && g.objMgr != nil {
		// Scenario::loadAndStart: a new rng seed from the time and the old
		// rng, then Scenario::start (pkg/sim StartScenario).
		gameW.Map.Rng1 = *core.NewPrngSeed(uint32(time.Now().UnixMilli())^oldRng.Srand0(), oldRng.Srand1())
		var gen *sim.TerrainOptions
		if sc.LandscapeToGenerate {
			gen = &sim.TerrainOptions{MinLandHeight: sc.Options.MinLandHeight, TopographyStyle: sc.Options.TopographyStyle,
				HillDensity: sc.Options.HillDensity, HillsEdgeOfMap: sc.Options.Flags&2 != 0, SeaLevel: sc.SeaLevel,
				NumRiverbeds: sc.Options.NumRiverbeds, MinRiverWidth: sc.Options.MinRiverWidth, MaxRiverWidth: sc.Options.MaxRiverWidth,
				RiverbankWidth: sc.Options.RiverbankWidth, RiverMeanderRate: sc.Options.RiverMeanderRate,
				LandDistributionPatterns: sc.Options.LandDistributionPatterns,
				NumberOfForests:          sc.Options.NumberOfForests, NumberRandomTrees: sc.Options.NumberRandomTrees,
				MinForestRadius: sc.Options.MinForestRadius, MaxForestRadius: sc.Options.MaxForestRadius,
				MinForestDensity: sc.Options.MinForestDensity, MaxForestDensity: sc.Options.MaxForestDensity,
				MinAltitudeForTrees: sc.Options.MinAltitudeForTrees, MaxAltitudeForTrees: sc.Options.MaxAltitudeForTrees}
		}
		gameW.Map.StartScenarioGenerating(sc.Options.StartYear, gen, g.r.G1, g.objMgr)
	}
	g.w = gameW
	g.currentScenarioPath = filePath
	g.inTitleScreen = false

	// Warm the new scenario's sprites in the background (see prewarmTick).
	g.prewarmIDs, g.prewarmPos = gameW.CollectSpriteIDs(), 0
	g.loadingCaption = filepath.Base(filePath)
	g.loadingStyleFlip = !g.loadingStyleFlip
	log.Printf("[Game] Queued %d sprites for background prewarm from %s", len(g.prewarmIDs), filePath)

	g.inGame = true

	if g.titleSeq != nil {
		g.titleSeq.Stop()
	}
	g.windowMgr.CloseWindow("Load Game")
	g.windowMgr.CloseWindow("New Game")
	log.Printf("[Game] Loaded scenario: %s (%dx%d)", filePath, sc.MapWidth, sc.MapHeight)
	if z, err := strconv.Atoi(os.Getenv("GOLOCO_ZOOM")); err == nil && g.w != nil { // diagnostic: start at another zoom
		g.w.SetZoom(z)
	}
	return nil
}

// openSaveWindow is the Save Game prompt (PromptBrowse, save mode) in the
// saves folder, named after the current game.
// DEVIATION (existing): goloco writes its .goloco JSON save (the scenario path),
// not an S5 saved game — the S5 writer isn't ported.
func (g *Game) openSaveWindow() {
	name := ""
	if g.currentScenarioPath != "" {
		name = strings.TrimSuffix(filepath.Base(g.currentScenarioPath), filepath.Ext(g.currentScenarioPath))
	}
	w := ui.NewPromptBrowseSave("Save Game", goLocoSavesDir(), name, ".goloco", g.sw, g.sh)
	w.OnClose = func() { g.browse = nil }
	w.Confirm = func(caption, text, ok string, then func()) {
		g.prompt = ui.NewPromptOkCancel(caption, text, ok, g.sw, g.sh, then)
	}
	w.OnSave = func(path string) {
		g.browse = nil
		if err := g.saveGameTo(path); err != nil {
			log.Printf("[Game] Save error: %v", err)
		}
	}
	g.browse = w
}

// saveGame writes the current scenario path to a new .goloco JSON save file.
func (g *Game) saveGame() error {
	return g.saveGameTo(filepath.Join(goLocoSavesDir(), fmt.Sprintf("save_%d.goloco", time.Now().Unix())))
}

// saveGameTo writes the current scenario path to a .goloco JSON save file.
func (g *Game) saveGameTo(path string) error {
	if g.currentScenarioPath == "" {
		g.saveMsg = "Nothing to save"
		g.saveMsgFrames = 120
		return nil
	}
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
	if g.scenarioSelect != nil && g.scenarioSelect.Contains(g.mouseX, g.mouseY) {
		g.scenarioSelect.HandleWheel(dy)
		return true
	}
	if g.browse != nil && g.browse.Contains(g.mouseX, g.mouseY) {
		g.browse.HandleWheel(dy)
		return true
	}
	if g.townList != nil && g.townList.Contains(g.mouseX, g.mouseY) {
		g.townList.HandleWheel(dy)
		return true
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
	// The ported ScenarioSelect window (pkg/ui/scenarioselect.go).
	metas, _ := scenario.ScanScenarios(g.dataDir + "/../Scenarios")
	if len(metas) == 0 {
		metas, _ = scenario.ScanScenarios(g.dataDir)
	}
	w := ui.NewScenarioSelect(metas, g.sw, g.sh)
	w.OnClose = func() { g.scenarioSelect = nil }
	w.OnStart = func(m *scenario.ScenarioMeta) {
		g.scenarioSelect = nil
		if err := g.loadScenario(m.FilePath); err != nil {
			log.Printf("[Game] Load error: %v", err)
		}
	}
	g.scenarioSelect = w
}

// openOptionsWindow opens an Options window with proper 3D toggle buttons for
// sound and music, styled after OpenLoco's option panels.
func (g *Game) openOptionsWindow() {
	// The ported Options window, Audio tab (pkg/ui/options.go).
	var tc, oc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, oc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowOptionsColour
	}
	w := ui.NewOptions(g.sw, g.sh, tc, oc)
	w.TitleMusic = g.musicEnabled
	w.MusicVolume = 1
	if g.audioMgr != nil {
		w.MusicVolume = g.audioMgr.MusicVolume()
	}
	w.OnClose = func() { g.options = nil }
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	// Screen mode: Ebitengine's fullscreen stands for both fullscreen modes
	// (it has no exclusive mode); windowed leaves it.
	if ebiten.IsFullscreen() {
		w.ScreenMode = 2
	}
	w.OnScreenMode = func(m int) { ebiten.SetFullscreen(m != 0) }
	// getFullscreenResolutions: SDL lists the display's modes; Ebitengine's
	// fullscreen always uses the monitor's own mode, so that is the one
	// resolution it can offer (PLATFORM: no mode switching).
	if g.fsResolution == [2]int{} {
		g.fsResolution = [2]int{1920, 1080} // Config default
	}
	w.Resolution = g.fsResolution
	if mw, mh := ebiten.Monitor().Size(); mw > 0 && mh > 0 {
		w.Resolutions = [][2]int{{mw, mh}}
	}
	w.OnResolution = func(i int) {
		g.fsResolution = w.Resolutions[i]
		ebiten.SetFullscreen(true)
	}
	// Frame rate limit: vertical sync = Ebitengine's vsync; unrestricted =
	// vsync off; internal speed draws with vsync too (upstream draws only at
	// its update rate — goloco draws every frame; noted in LOOP).
	w.FrameLimit = g.frameLimit
	w.OnFrameLimit = func(m int) {
		g.frameLimit = m
		ebiten.SetVsyncEnabled(m != 2)
	}
	// Regional tab: the scenario's currency object, the preferred one
	// (Config default CURRDOLL), and title mode.
	w.TitleMode = g.inTitleScreen
	if g.objMgr != nil {
		if c := g.objMgr.GetCurrencyObject(g.objMgr.CurrencySlot); c != nil {
			w.CurrencyName = c.Name
		}
		if g.preferredCurrency == "" {
			g.preferredCurrency = "CURRDOLL" // Config default
		}
		if c := g.objMgr.GetCurrencyObject(g.preferredCurrency); c != nil {
			w.PreferredCurrencyName = c.Name
		}
		w.Currencies = g.objMgr.AvailableCurrencies()
		langs := i18n.Available()
		for i, l := range langs {
			w.Languages = append(w.Languages, l.NativeName)
			if l.Locale == i18n.Locale() {
				w.Language = i
			}
		}
		w.OnLanguage = func(i int) { // languageDropdown: the pack, then Config::write
			if err := i18n.Load(langs[i].Locale); err != nil {
				log.Printf("[Options] language: %v", err)
				return
			}
			if err := setConfigValue("language", langs[i].Locale); err != nil {
				log.Printf("[Options] saving the language: %v", err)
			}
		}
		w.OnPreferredCurrency = func(dat string) { // preferredCurrencyDropdown
			g.preferredCurrency = dat
			g.loadPreferredCurrencyAlways()
			if c := g.objMgr.GetCurrencyObject(g.objMgr.CurrencySlot); c != nil {
				w.CurrencyName = c.Name
			}
		}
		w.PreferredCurrency = g.preferredCurrencyFlags
		w.OnPreferredToggle = func(i int, v bool) { // PARTIAL: Config::write
			g.preferredCurrencyFlags[i] = v
			if i == 1 {
				g.loadPreferredCurrencyAlways()
				if c := g.objMgr.GetCurrencyObject(g.objMgr.CurrencySlot); c != nil {
					w.CurrencyName = c.Name
				}
			}
		}
		w.OnCurrency = func(dat string) { useCurrency(g.objMgr, dat) } // currencyDropdown
	}
	// Controls tab: Config::edgeScrolling (default on). zoomToCursor and
	// invertRightMouseViewPan are kept in the window only: goloco has neither yet.
	w.Controls[0] = !g.noEdgeScrolling
	w.OnControl = func(i int, v bool) {
		if i == 0 {
			g.noEdgeScrolling = !v
		}
	}
	w.Scale = g.scale
	w.OnScale = func(s float64) { // setWindowScaling: Config::write, then the resize (Layout)
		g.scale = s
		setConfigValue("scaleFactor", strconv.FormatFloat(s, 'f', 2, 64))
	}
	w.ShowFPS = g.showFPS
	w.OnShowFPS = func(on bool) {
		g.showFPS = on
		v := "false"
		if on {
			v = "true"
		}
		setConfigValue("showFPS", v)
	}
	w.OnTitleMusic = func(on bool) {
		g.musicEnabled = on
		if !on && g.audioMgr != nil {
			g.audioMgr.StopMusic()
		}
	}
	w.OnMusicVol = func(v float64) {
		if g.audioMgr != nil {
			g.audioMgr.SetMusicVolume(v)
		}
	}
	g.options = w
}

// openLoadGameWindow opens the file browser for all supported file types.
func (g *Game) openLoadGameWindow() { g.openFileWindow("Load Game", false) }

// openFileWindow is the shared implementation for the two-panel file browser.
// title is the window title (also used as the CloseWindow key).
// scenariosOnly restricts the file list to .SC5 files when true.
func (g *Game) openFileWindow(title string, scenariosOnly bool) {
	// The ported PromptBrowse window (pkg/ui/promptbrowse.go), load mode.
	startDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		startDir = home
	}
	if g.dataDir != "" {
		startDir = filepath.Dir(g.dataDir) // one level up from "Data/"
	}
	exts := []string{".sv5", ".sc5", ".goloco"}
	if scenariosOnly {
		exts = []string{".sc5"}
	}
	w := ui.NewPromptBrowse(title, startDir, exts, g.sw, g.sh)
	w.OnClose = func() { g.browse = nil }
	w.Confirm = func(caption, text, ok string, then func()) { // right-click delete asks first
		g.prompt = ui.NewPromptOkCancel(caption, text, ok, g.sw, g.sh, then)
	}
	w.OnPick = func(path string) {
		g.browse = nil
		if err := g.loadScenario(path); err != nil {
			log.Printf("[Game] Load error: %v", err)
		}
	}
	g.browse = w
}

// -----------------------------------------------------------------------

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth < 1 {
		outsideWidth = defaultWidth
	}
	if outsideHeight < 1 {
		outsideHeight = defaultHeight
	}
	// Config::scaleFactor (SoftwareDrawingEngine::resize): the screen is the
	// window divided by the scale; Ebitengine scales it up to the window (sharp
	// at whole scales, smoothed between — upstream's nearest-then-linear) and
	// reports the mouse in the game's pixels (getCursorPosScaled).
	if s := g.scale; s > 1 {
		outsideWidth, outsideHeight = int(float64(outsideWidth)/s), int(float64(outsideHeight)/s)
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

	f, err := os.Create(debugPath("diag.png"))
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
	audio.ProbeChild() // the sound-device probe (item AU): exits when this process is it
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

	game.progress = loadProgress()
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
		game.windowMgr.OpenWindow(game.newPortedConstructionWindow())
		mw, mh := game.w.GetMapSize()
		game.w.SetConstructionHead(mw/2, mh/2)
		// Straight line so the piece origins fall on a predictable tile row and
		// the first/last pieces can carry stations for the board→deliver loop.
		seq := []int{0, 0, 0, 0, 0}
		var firstX, firstY, lastX, lastY int
		for i, id := range seq {
			h := game.w.ConstructionHeadState()
			if i == 0 {
				firstX, firstY = h.X, h.Y
			}
			if i == len(seq)-1 {
				lastX, lastY = h.X, h.Y
			}
			if !game.w.PlaceTrackAtHead(id, 0) {
				log.Printf("[ConsTest] piece %d not placeable at head %+v", id, game.w.ConstructionHeadState())
			}
		}
		log.Printf("[ConsTest] head after sequence: %+v", game.w.ConstructionHeadState())
		// Stations on the first and last pieces of the line.
		stationSlot := game.firstTrainStationSlot()
		if game.w.PlaceStationOnTile(firstX, firstY, stationSlot) {
			log.Printf("[ConsTest] station A placed at (%d,%d)", firstX, firstY)
		}
		if game.w.PlaceStationOnTile(lastX, lastY, stationSlot) {
			log.Printf("[ConsTest] station B placed at (%d,%d)", lastX, lastY)
		}
	}
	if os.Getenv("GOLOCO_OPEN") == "road" && game.w != nil {
		// Headless road construction test: chain straight + very-small curves.
		game.buildMode = buildModeRoad
		game.windowMgr.OpenWindow(game.newPortedConstructionWindow())
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
	defer startCPUProfile()() // GOLOCO_CPUPROFILE=file (LOOP item 0)
	if err := ebiten.RunGame(game); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
}

// checkObjective reports the scenario won once the sim has marked the player
// company's challenge completed (Company::evaluateChallengeProgress).
// OpenLoco reference: src/OpenLoco/src/Scenario/ScenarioObjective.cpp
func (g *Game) checkObjective() {
	if !g.hasObjective || g.gameWon || !g.inGame {
		return
	}
	// The player company's challengeCompleted flag (set by the sim's daily
	// evaluateChallengeProgress).
	if completed, _, _, _ := g.w.Map.ChallengeStatus(); completed {
		g.gameWon = true
	}
	if g.gameWon {
		log.Printf("[Game] OBJECTIVE ACHIEVED: %s", g.objective.Describe())
		if g.progress != nil && g.currentScenarioPath != "" {
			score := uint32(0)
			if id := int(g.w.Map.ControllingCompany); g.objective.Type == 3 && id < len(g.w.Map.Companies) {
				score = g.w.Map.Companies[id].CargoDelivered[g.objective.DeliveredCargoType&31]
			}
			g.progress.markCompleted(g.currentScenarioPath, g.objective.Describe(), score)
		}
	}
}

// --- RCT-style track construction window --------------------------------
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Construction.cpp
// Curve row + slope row of sprite buttons, Rotate/Undo, and a Build action.
// Click the map to place the construction head; pieces then chain from it.

// consSelectedTrackID resolves the current curve/slope selection to the first
// placeable track id at the head, or -1.
func (g *Game) consSelectedTrackID() int {
	// getTrackPieceId (pkg/construct) for the chosen piece and gradient at
	// the head's constructionRotation.
	rot := g.w.ConstructionHeadState().Rotation
	p, ok := construct.GetTrackPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), rot)
	if !ok || !g.w.CanPlaceTrackAtHead(int(p.ID)) {
		return -1
	}
	return int(p.ID)
}

// consTrackPiece is lastSelectedTrackPiece: a one-shot piece if chosen, else
// the curve buttons' piece.
func (g *Game) consTrackPiece() uint8 {
	if g.consOneShot >= 0 {
		return uint8(g.consOneShot)
	}
	switch g.consCurve {
	case -1:
		return construct.LeftCurveVerySmall
	case 1:
		return construct.RightCurveVerySmall
	case -2:
		return construct.LeftCurveSmall
	case 2:
		return construct.RightCurveSmall
	case -3:
		return construct.LeftCurve
	case 3:
		return construct.RightCurve
	case -4:
		return construct.LeftCurveLarge
	case 4:
		return construct.RightCurveLarge
	}
	return construct.Straight
}

// consTrackGradient is lastSelectedTrackGradient for goloco's slope choice.
func consTrackGradient(slope int) uint8 {
	switch slope {
	case -2:
		return construct.SteepSlopeDown
	case -1:
		return construct.SlopeDown
	case 1:
		return construct.SlopeUp
	case 2:
		return construct.SteepSlopeUp
	}
	return construct.Level
}

// consDualTrackPiece is the piece the left or right dual-track button
// chooses at this heading (construct.DualTrackButtons).
func consDualTrackPiece(right bool, rot uint8) int {
	_, _, l, r := construct.DualTrackButtons(rot)
	if right {
		return int(r)
	}
	return int(l)
}

// consSelectedRoadID is getRoadPieceId (pkg/construct) for the chosen piece
// and gradient at the head, if it can be placed there; else -1.
func (g *Game) consSelectedRoadID() int {
	rot := g.w.ConstructionHeadState().Rotation
	p, ok := construct.GetRoadPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), rot)
	if !ok || !g.w.CanPlaceRoadAtHead(int(p.ID)) {
		return -1
	}
	return int(p.ID)
}

// --- Mode-aware dispatch (track vs road) for the shared construction window --

func (g *Game) consIsRoad() bool { return g.buildMode == buildModeRoad }

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

// consDisplayID returns a piece id to show in the preview panel for the current
// curve/slope selection, regardless of whether it is placeable at the head (so
// the panel always shows the piece the buttons describe). Falls back to the
// first candidate, or straight.
func (g *Game) consDisplayID() int {
	if id := g.consSelectedID(); id >= 0 {
		return id
	}
	// The piece getTrackPieceId/getRoadPieceId gives; none (−1) when it
	// can't be built at this heading.
	rot := g.w.ConstructionHeadState().Rotation
	get := construct.GetTrackPieceID
	if g.consIsRoad() {
		get = construct.GetRoadPieceID
	}
	if p, ok := get(g.consTrackPiece(), consTrackGradient(g.consSlope), rot); ok {
		return int(p.ID)
	}
	return -1
}

func (g *Game) consPlace(id int) {
	var built bool
	if g.consIsRoad() {
		built = g.w.PlaceRoadAtHead(id, uint8(g.buildObjID))
	} else {
		built = g.w.PlaceTrackAtHead(id, uint8(g.buildObjID))
	}
	if built && g.consOneShot >= 0 {
		// constructTrack/constructRoad: one-shot pieces (≥ 9) go back to straight.
		g.consOneShot, g.consCurve = -1, 0
	}
}

func (g *Game) consUndo() {
	if g.consIsRoad() {
		g.w.UndoLastRoad()
	} else {
		g.w.UndoLastTrack()
	}
}

// firstTrainStationSlot returns the first populated train-station object slot,
// or 0 when none are loaded. Used to pick a station type for placement.
func (g *Game) firstTrainStationSlot() uint8 {
	if g.objMgr == nil {
		return 0
	}
	for i := 0; i < 16; i++ {
		if g.objMgr.GetTrainStationObjectByIndex(i) != nil {
			return uint8(i)
		}
	}
	return 0
}

// consPlaceStation attaches a station to the track piece under the head, using
// the first populated train-station object slot. Track mode only.
func (g *Game) consPlaceStation() {
	if g.consIsRoad() {
		return
	}
	g.w.PlaceStationAtHead(g.firstTrainStationSlot())
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

// Construction window layout — content-relative offsets, styled after the
// original Locomotion "Railway Track" window (reference/gallery-screenshot-0.png).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/ConstructionTab.cpp
const (
	consWinW = 138
	consWinH = 282

	consBtnS   = 22 // small (curve/slope) button size
	consBigBtn = 24 // large-curve / straight button size

	consTabY = 2
	consTabH = 26
	consTabW = 42

	consCurveY = 32 // primary curve row (6 buttons, no straight)
	consRow2Y  = 56 // large-left / straight / large-right
	consSlopeY = 82 // slope row (5 buttons)
	consTypeY  = 108
	consTypeH  = 20
	consPrevY  = 132 // piece preview panel
	consPrevH  = 58
	consCostY  = 196
	consIconY  = 214
	consIconS  = 24
	consStatY  = 244
)

// consCurveVals / consCurveSprites: the six primary-row curve buttons, sharpest
// (very-small) on the outside, medium nearest the centre — mirrors the original.
var consCurveVals = []int{-1, -2, -3, 3, 2, 1}
var consCurveSprites = []int{2340, 2342, 2344, 2345, 2343, 2341}

// consRow2Vals: large-left curve, straight (up-arrow), large-right curve.
var consRow2Vals = []int{-4, 0, 4}
var consRow2Sprites = []int{2346, 2335, 2347}

var consSlopeVals = []int{-2, -1, 0, 1, 2}
var consSlopeSprites = []int{2348, 2349, 2350, 2351, 2352}

const (
	spriteRotateObject = 2365 // ImageIds::rotate_object (red curved arrow)
	spriteRemoveObject = 2361 // ImageIds::construction_remove (undo/remove)
)

// drawInlineDiag (GOLOCO_DIAG_INLINE) is drawVehicleInline for the first ten
// rail and road vehicle types, in a panel at (20, 60).
func (g *Game) drawInlineDiag(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TTF") != "" { // the G1 font and the TTF fallback on one line
		fillRect(screen, 10, 10, 300, 30, color.RGBA{200, 200, 200, 255})
		fillRect(screen, 10, 20, 300, 1, color.RGBA{255, 0, 0, 255}) // y = 20: the text tops
		ui.DrawTextBold(screen, "Bitmap text", 20, 20, color.Black)
		ui.DrawTextBold(screen, "TTF 日本語 text", 120, 20, color.Black)
	}
	if os.Getenv("GOLOCO_DIAG_INLINE") == "" || g.objMgr == nil {
		return
	}
	var vs []*objects.VehicleObject
	var offs []int
	for _, v := range g.objMgr.Vehicles {
		if v != nil && v.ImageOffset != 0 && (v.Mode == objects.TransportModeRail || v.Mode == objects.TransportModeRoad) && len(vs) < 10 {
			vs = append(vs, v)
			offs = append(offs, g.objMgr.VehicleYDisplayOffset(v))
		}
	}
	log.Printf("[diag] inline: %d vehicle types", len(vs))
	ui.DrawVehiclesInlineDiag(screen, 20, 60, vs, offs)
}

// playerPanelCompany fills the player panel's company part from the
// controlling company: the owner's image (companyFace), performanceIndex and
// the increased/decreasedPerformance flags.
func (g *Game) playerPanelCompany(st *ui.PlayerPanelState) {
	if g.w == nil {
		return
	}
	id, _ := g.w.ControllingCompany()
	if id >= len(g.w.CompanyFaces) {
		return
	}
	st.Face, st.FaceColour = g.companyFace(id)
	f := g.w.CompanyFaces[id]
	st.Performance = f.Performance
	switch {
	case f.ChallengeFlags&(1<<4) != 0:
		st.Trend = 1
	case f.ChallengeFlags&(1<<5) != 0:
		st.Trend = -1
	}
}

// uiOpen reports a window, menu or prompt open: the world may then be drawn
// less often so the UI stays responsive (LOOP item 0).
func (g *Game) uiOpen() bool {
	return g.prompt != nil || g.options != nil || g.townList != nil || g.browse != nil || g.scenarioSelect != nil ||
		(g.dropdown != nil && g.dropdown.Visible) || (g.windowMgr != nil && len(g.windowMgr.Windows) > 0)
}

// mouseMoved is Input::hasPendingMouseInputUpdate for GameTicks: the cursor
// moved since the last frame.
func (g *Game) mouseMoved() bool {
	moved := g.mouseX != g.lastTickMouseX || g.mouseY != g.lastTickMouseY
	g.lastTickMouseX, g.lastTickMouseY = g.mouseX, g.mouseY
	return moved
}

// loadScale is Config::scaleFactor (default 1.0; GOLOCO_SCALE overrides it,
// for the test page), clamped to ScaleFactor's range.
func loadScale() float64 {
	v := configValue("scaleFactor")
	if e := os.Getenv("GOLOCO_SCALE"); e != "" {
		v = e
	}
	s, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 1
	}
	return min(max(s, ui.ScaleFactorMin), ui.ScaleFactorMax)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// openGraphicsDir is the open build's OpenGraphics folder — GOLOCO_OPENGRAPHICS,
// else the web pack's /w/opengraphics — when it holds the palette and ObjData;
// "" otherwise.
func openGraphicsDir() string {
	for _, d := range []string{os.Getenv("GOLOCO_OPENGRAPHICS"), "/w/opengraphics"} {
		if d == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(d, "locopal.jasc.pal")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(d, "ObjData")); err == nil {
			return d
		}
	}
	return ""
}

// tileUnderCursor is ConstructionTab's getSurfaceOrWaterLocFromUi (the
// surface or water the viewport draws at the screen point) as a tile, -1, -1
// for none.
func (g *Game) tileUnderCursor(sx, sy int) (int, int) {
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return -1, -1
	}
	return int(px) >> 5, int(py) >> 5
}
