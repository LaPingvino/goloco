package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/title"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
)

// game.go — the Game: its state, start-up (data folders, G1, objects, the
// title map), the screen layout and the background sprite prewarm.

const (
	defaultWidth       = 800
	defaultHeight      = 600
	edgeScrollingSpeed = 12 // Config::edgeScrollingSpeed's default
)

const (
	buildModeNone    = 0
	buildModeTrack   = 1
	buildModeRoad    = 2
	buildModeDock    = 3 // Construction::openWithFlags(1 << 30): the dock window (Station tab only)
	buildModeAirport = 4 // openWithFlags(1 << 31): the airport window (Station tab only, rotate button)
)

// Game is the running program: the world shown (w — the title map or a game's),
// the renderer and its object manager (r, r.ObjMgr: nil only without game
// data), the open windows, and the tools' state. w and r are never nil: NewGame
// sets both, and w is only ever replaced by another world (loadScenario,
// quitToMenu).
type Game struct {
	// The world, its renderer, sound.
	w          *world.World
	titleWorld *world.World // preserved title-screen world; restored on Quit to Menu
	r          *render.Renderer
	audioMgr   *audio.Manager
	dataDir    string

	// The screen and the mouse.
	sw, sh                         int     // current logical screen dimensions (updated by Layout each frame)
	scale                          float64 // Config::scaleFactor: the game drawn at window ÷ scale, scaled up
	mouseX, mouseY                 int
	lastTickMouseX, lastTickMouseY int  // the cursor at the last GameTicks (mouseMoved)
	rightDragging                  bool // right-click drag pan state
	dragLastX, dragLastY           int
	windowFrame                    int // counts Updates: the windows' frameNo (onUpdate's frameNo += 1)

	// The title screen.
	inTitleScreen bool
	titleSeq      *title.Sequence
	titleFrame    int    // title screen globe animation counter (incremented each frame)
	titleTip      string // the hovered title button's tooltip (drawn last)

	// The running game.
	inGame              bool   // a scenario / save is loaded into gameplay
	currentScenarioPath string // path of the currently loaded scenario/save
	scenarioDetails     string // the started scenario's description (the Challenge tab; PARTIAL: a saved game's)
	objective           scenario.Objective
	hasObjective        bool
	gameWon             bool
	progress            *progressFile
	isPaused            bool   // speed controls
	speedMult           int    // 1, 2, 4, or 8
	saveMsg             string // transient "Game Saved!" status overlay
	saveMsgFrames       int    // frames remaining to show saveMsg

	// The toolbar and the windows, each nil unless open.
	toolbar           *ui.Toolbar
	lastPortOption    int // ToolbarTop's _lastPortOption: 0 airport, 1 dock
	windowMgr         *ui.SimpleWindowManager
	dropdown          *ui.DropdownMenu
	colourDropdown    *ui.ColourDropdown        // Dropdown::showColour when open
	scenarioSelect    *ui.ScenarioSelectWindow  // the New Game window when open
	browse            *ui.PromptBrowseWindow    // the Load window when open
	options           *ui.OptionsWindow         // the Options window when open
	townList          *ui.TownListWindow        // the town list when open
	townWin           *ui.TownWindow            // the Town window (TownWindow.cpp)
	stationWin        *ui.StationWindow         // the Station window (StationWindow.cpp)
	catchmentStation  int                       // _lastSelectedStation: whose catchment shows (−1 none)
	buildVehicle      *ui.BuildVehicleWindow    // the Build Vehicle window when open (5P.f)
	vehicleWin        *ui.VehicleWindow         // the Vehicle window when open (5P.g)
	companyFinances   *ui.CompanyFinancesWindow // the Company window's Finances tab when open (PL.1)
	companyFinancesID int
	companyList       *ui.CompanyListWindow // CompanyList when open
	prompt            *ui.PromptOkCancel    // an open confirmation (modal: it takes all input)
	textInput         *ui.TextInputWindow   // an open text input (renaming); PARTIAL: takes all input, upstream's is not modal
	errorWin          *ui.ErrorWindow       // WindowType::error (one at a time)

	// The construction window (construction.go).
	buildMode   int   // buildModeNone / Track / Road / Dock / Airport
	buildObjID  int   // index into TrackObjects / RoadObjects
	consTab     int   // construction window tab: 0 construction, 1 station, 2 signal, 3 overhead
	consCurve   int   // construction curve selection: -4 (left large) … 0 … +4
	consSlope   int   // construction slope selection: -2 (steep down) … 0 … +2
	consOneShot int   // track: an s-bend/turnaround piece chosen (construct.TrackPiece ≥ 9), -1 none
	consBridge  int   // lastSelectedBridge (slot; -1 = the first available)
	consSignal  int   // signal tab: lastSelectedSignal (slot; -1 = the first available)
	consSigOne  bool  // signal tab: single direction chosen (isSignalBothDirections = 0)
	consStation int   // station tab: lastSelectedStationType (slot; -1 = the first available)
	consMods    uint8 // overhead tab: lastSelectedMods (PARTIAL: not from the scenario's trackMods/roadMods)
	hoverTileX  int   // tile under cursor in build mode (-1 if invalid)
	hoverTileY  int

	// The Station tab's ghosts (stationtool.go).
	stationGhost         *sim.TrainStationPlacement // the station tab's ghost (cState.stationGhost*), nil none
	stationGhostCost     int32                      // cState.stationCost
	stationGhostStation  int                        // cState.constructingStationId (-1 new)
	stationGhostAccepted uint32                     // constructingStationAcceptedCargoTypes
	stationGhostProduced uint32                     // constructingStationProducedCargoTypes
	roadStationGhost     *sim.RoadStationPlacement  // the road stop ghost (stationGhostType bit 7)
	dockGhost            *sim.PortPlacement         // the dock ghost (stationGhostType bit 14)
	airportGhost         *sim.AirportPlacement      // the airport ghost (stationGhostType bit 15)
	airportRotation      uint8                      // cState.constructionRotation in the airport window

	// The Vehicle window's tools (vehicle_pickup.go, vehicle_route.go).
	pickupOn       bool   // the Vehicle window's pickup (placing) tool is active
	pickupHead     uint16 // the vehicle head it places
	driverCheat    bool   // SceneManager driverCheatEnabled (the Cheats window's; off)
	routeSel       int    // the Vehicle window's Route selection (orderTableIndex; -1: none)
	lastTownOption int    // ToolbarTopCommon.cpp _lastTownOption: the Towns menu's last item (0 towns, 1 industries)
	routeTool      bool   // the Route tab's tool (adding stops by clicking the map) is active
	routeHead      uint16 // whose Route that selection is
	// The ghost's placement (_ghostLandTrackAndDirection, _ghostVehiclePos,
	// _ghostTrackProgress; ghostProgress -1: none).
	ghostPlace      sim.VehiclePlacement
	ghostProgress   int
	ghostWater      sim.VehicleWaterPlacement // _ghostVehiclePos for a boat ghost (ghostProgress 0)
	ghostAirStation uint16                    // _ghostAirportStationId (0xFFFF none)
	ghostAirNode    uint8                     // _ghostAirportNode

	// Options (options.go; Config).
	noEdgeScrolling   bool   // Options > Controls: edge scrolling off (Config::edgeScrolling)
	musicEnabled      bool   // whether music is on
	showFPS           bool   // Config::showFPS (drawFPS)
	frameLimit        int    // Options' frame rate limit (0 internal, 1 vsync, 2 unrestricted)
	fsResolution      [2]int // Config display.fullscreenResolution
	preferredCurrency string // Config::preferredCurrency (DAT name; PARTIAL: not saved)
	// preferredCurrencyFlags are Config usePreferredCurrencyForNewGames and
	// usePreferredCurrencyAlways (default off).
	preferredCurrencyFlags [2]bool

	// The FPS counter (drawFPS).
	fpsRef     time.Time // FPSCounter's _referenceTime
	fpsFrames  int
	fpsCurrent float64

	// Background sprite prewarm: decode+pack a few sprites per frame within a
	// small time budget, so the game shows at once instead of after a blocking
	// preload of every sprite (11,057 for the title).
	prewarmIDs       []int
	prewarmPos       int
	firstFrameLogged bool // the start-to-first-world-frame time has been logged

	// scripted input playback (GOLOCO_SCRIPT)
	script *inputScript

	// --shots mode: periodic full-frame captures (shot_NN.png), then quit.
	shotsEvery time.Duration // wall-clock interval between captures
	shotsLeft  int
	shotsLast  time.Time
	shotsN     int
	// quitAfterSave: exit cleanly once the map save has completed.
	quitAfterSave bool

	// Diagnostic crop: when set, save a 480×360 PNG centred on diagTileX/Y after the next Draw.
	// diagQuit=true means quit afterwards (CLI --diag mode); false keeps the game running (F12).
	diagPending bool
	diagQuit    bool
	diagTileX   int
	diagTileY   int

	// The GOLOCO_DIAG_* switches' state (diag*.go): each "Done" is a switch
	// that has run.
	diagConsDone     bool
	diagNewGameDone  bool
	diagTownsDone    bool
	diagOptionsDone  bool
	diagBrowseDone   bool
	buildVehDiagDone bool
	vehDiagDone      bool    // GOLOCO_DIAG_VEHLIST opened
	diagLogged       bool    // GOLOCO_DIAG_BUILDVEHICLE=place: the first placement logged
	diagMouse        *[2]int // DIAGNOSTIC ONLY: the mouse held at a point (the roadghost diag's hover)
	diagShip         uint16  // DIAGNOSTIC ONLY: the shipsail diag's ship, logged as it goes
	diagFollow       bool    // DIAGNOSTIC ONLY: keep the view on diagShip (planefly)
	diagShipFrames   int
	diagMoney        bool // DIAGNOSTIC ONLY: keep a money popup alive (moneyeffect)
	diagHoldGhost    bool // DIAGNOSTIC ONLY: the stationghost diag keeps its ghost
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
	if dataDir != "" {
		loadLocomotionData(r, dataDir)
	} else if og := openGraphicsDir(); og != "" {
		loadOpenGraphics(r, og)
	} else {
		log.Printf("Locomotion data directory not found")
	}

	// Fall back to placeholder tiles if G1 not loaded
	if r.G1 == nil {
		loadPlaceholderAtlas(r)
	}

	w := world.NewWorld(r)
	toolbar := ui.NewToolbar(defaultWidth)
	windowMgr := ui.NewSimpleWindowManager()

	// Try to load the title scenario
	if dataDir != "" {
		loadTitleMap(w, r, dataDir)
	} else if r.G1 != nil && r.ObjMgr != nil { // the open build: no title.dat — a landscape with the editor's defaults
		m := sim.NewGeneratedMap(0x4C4F434F, 0x474F4C4F, 1900, sim.EditorDefaultTerrainOptions(r.ObjMgr), r.G1, r.ObjMgr)
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
		catchmentStation: -1,
		routeSel:         -1,                                                                    // none selected
		ghostProgress:    -1,                                                                    // no ghost placed
		showFPS:          configValue("showFPS") == "true" || os.Getenv("GOLOCO_SHOWFPS") != "", // Config::showFPS (Config.cpp: default false)
		scale:            loadScale(),
		w:                w,
		consSignal:       -1,
		consOneShot:      -1,
		consBridge:       -1,
		consStation:      -1,
		titleWorld:       w, // keep a reference so Quit to Menu can restore it
		r:                r,
		toolbar:          toolbar,
		windowMgr:        windowMgr,
		audioMgr:         audioMgr,
		titleSeq:         titleSeq,
		dataDir:          dataDir,
		inTitleScreen:    true,
		musicEnabled:     true,
		sw:               defaultWidth,
		sh:               defaultHeight,
		speedMult:        1,
		hoverTileX:       -1,
		hoverTileY:       -1,
		// Warm the title's sprites in the background (prewarmTick); the title
		// draws at once and decodes anything it needs first on use, as upstream
		// draws g1 sprites directly with no preload phase.
		prewarmIDs: titleSpriteIDs,
	}
}

// loadLocomotionData loads the Locomotion folder's G1 (sprites, palette, the
// bitmap fonts) and its ObjData objects into the renderer.
func loadLocomotionData(r *render.Renderer, dataDir string) {
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

		// Set the package-level global palette from the loaded G1 so
		// DrawingContext and UI widgets (window backgrounds, buttons …)
		// use the game's palette everywhere.
		if graphics.SetGlobalPalette(r.G1) {
			log.Printf("Applied G1 palette to UI drawing context")
		}
	}

	// Load objects from ObjData
	objDataDir := filepath.Join(filepath.Dir(dataDir), "ObjData")
	if _, err := os.Stat(objDataDir); err == nil {
		log.Printf("Loading objects from: %s", objDataDir)
		objMgr := objects.NewObjectManager(objDataDir)

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
}

// loadOpenGraphics is the open build (item O): no Locomotion folder —
// OpenGraphics (MIT): its palette and an empty base pool (assets.NewOpenG1),
// its objects. The bitmap fonts are g1.DAT's, so the TTF fallback stays.
func loadOpenGraphics(r *render.Renderer, og string) {
	log.Printf("[Game] No Locomotion data; the open build from %s", og)
	if g1, err := assets.NewOpenG1(filepath.Join(og, "locopal.jasc.pal")); err != nil {
		log.Printf("[Game] open build: %v", err)
	} else {
		r.G1 = g1
		graphics.SetGlobalPalette(r.G1)
		objMgr := objects.NewObjectManager(filepath.Join(og, "ObjData"))
		objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
		objMgr.G1File = g1
		if err := objMgr.LoadAllObjects(); err != nil {
			log.Printf("[Game] open build: objects: %v", err)
		}
		log.Printf("[Game] open build: %d vehicles, %d land objects, %d images", len(objMgr.Vehicles), len(objMgr.LandObjects), g1.GetTotalSprites())
		r.ObjMgr = objMgr
	}
}

// loadPlaceholderAtlas is the fallback with no G1 at all: generated tiles.
func loadPlaceholderAtlas(r *render.Renderer) {
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

// loadTitleMap loads title.dat, the title screen's scenario, into the world.
func loadTitleMap(w *world.World, r *render.Renderer, dataDir string) {
	titlePath := filepath.Join(dataDir, "title.dat")
	log.Printf("[Game] Loading title scenario from: %s", titlePath)
	if sc, err := scenario.LoadScenarioData(titlePath); err == nil {
		// Load packed objects embedded in the scenario file first.
		// These take priority over ObjData files and provide scenario-
		// specific assets (buildings, vehicles, etc.) that may not exist
		// in the standard ObjData directory.
		if r.ObjMgr != nil && len(sc.PackedObjectsRaw) > 0 {
			packedOK := 0
			for _, po := range sc.PackedObjectsRaw {
				if _, err := r.ObjMgr.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data); err != nil {
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
		if r.ObjMgr != nil {
			reorderScenarioObjects(r.ObjMgr, sc)
		}
		w.LoadFromScenario(sc)
		log.Printf("[Game] Loaded title scenario: %dx%d map", sc.MapWidth, sc.MapHeight)
	} else {
		log.Printf("[Game] Could not load title scenario: %v", err)
	}
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

// processStart is when the program started; used for startup timings.
var processStart = time.Now()

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

// skin is the loaded interface skin object (INTERDEF.DAT): the windows'
// colours and the toolbar images; nil without game data.
func (g *Game) skin() *objects.InterfaceSkinObject {
	if g.r.ObjMgr == nil {
		return nil
	}
	return g.r.ObjMgr.InterfaceSkin
}

// playerCmd is a game command's context as the controlling company (what
// GameCommands::getUpdatingCompanyId is for a command from the UI).
func (g *Game) playerCmd() *sim.Cmd {
	company, _ := g.w.ControllingCompany()
	return &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
}

// mouseMoved is Input::hasPendingMouseInputUpdate for GameTicks: the cursor
// moved since the last frame.
func (g *Game) mouseMoved() bool {
	moved := g.mouseX != g.lastTickMouseX || g.mouseY != g.lastTickMouseY
	g.lastTickMouseX, g.lastTickMouseY = g.mouseX, g.mouseY
	return moved
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
