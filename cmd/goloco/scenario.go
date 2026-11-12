package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
)

// scenario.go — starting, loading and saving games: the scenario / saved game
// into a fresh world, the New Game, Load and Save windows, the objective, and
// back to the title.

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

	if g.r.ObjMgr != nil {
		reorderScenarioObjects(g.r.ObjMgr, sc)
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
	if strings.EqualFold(filepath.Ext(filePath), ".sc5") && g.r.ObjMgr != nil {
		// Scenario::loadAndStart: a new rng seed from the time and the old
		// rng, then Scenario::start (pkg/sim StartScenario).
		gameW.Map.Rng1 = *core.NewPrngSeed(uint32(time.Now().UnixMilli())^oldRng.Srand0(), oldRng.Srand1())
		if g.script != nil && g.script.seed != nil { // a playthrough's "seed N": the whole state fixed, the same game every run
			gameW.Map.Rng1 = *core.NewPrngSeed(*g.script.seed, *g.script.seed^0x5EED)
		}
		g.scenarioDetails = sc.Options.Description // GameState scenarioDetails (Scenario::start copies the options')
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
				MinAltitudeForTrees: sc.Options.MinAltitudeForTrees, MaxAltitudeForTrees: sc.Options.MaxAltitudeForTrees,
				NumberOfTowns: sc.Options.NumberOfTowns, MaxTownSize: sc.Options.MaxTownSize,
				NumberOfIndustries: sc.Options.OptNumberOfIndustries}
		}
		if g.script != nil && g.script.seed != nil {
			log.Printf("[Seed] before generating: rng %08x %08x, generate %v", gameW.Map.Rng1.Srand0(), gameW.Map.Rng1.Srand1(), sc.LandscapeToGenerate)
		}
		gameW.Map.StartScenarioGenerating(sc.Options.StartYear, gen, g.r.G1, g.r.ObjMgr)
		if g.script != nil && g.script.seed != nil {
			log.Printf("[Seed] after starting: rng %08x %08x, company colour %v", gameW.Map.Rng1.Srand0(), gameW.Map.Rng1.Srand1(), gameW.Map.CompanyColours)
		}
	}
	g.w = gameW
	g.currentScenarioPath = filePath
	g.inTitleScreen = false

	// Warm the new scenario's sprites in the background (see prewarmTick).
	// PARTIAL: upstream shows the progress window while an S5 file loads
	// (S5/S5.cpp:728, Ui::ProgressBar::begin(StringIds::loading)), not for
	// sprites. goloco loads the file synchronously before the next frame; wire
	// ui.DrawProgressBar (pkg/ui/progressbar.go, Ui/Windows/ProgressBar.cpp) to
	// real S5 load progress when loading moves off the frame.
	g.prewarmIDs, g.prewarmPos = gameW.CollectSpriteIDs(), 0
	log.Printf("[Game] Queued %d sprites for background prewarm from %s", len(g.prewarmIDs), filePath)

	g.inGame = true

	if g.titleSeq != nil {
		g.titleSeq.Stop()
	}
	g.windowMgr.CloseWindow("Load Game")
	g.windowMgr.CloseWindow("New Game")
	log.Printf("[Game] Loaded scenario: %s (%dx%d)", filePath, sc.MapWidth, sc.MapHeight)
	if z, err := strconv.Atoi(os.Getenv("GOLOCO_ZOOM")); err == nil { // diagnostic: start at another zoom
		g.w.SetZoom(z)
	}
	return nil
}

// reorderScenarioObjects puts every object type the map refers to by index in
// the scenario's slot order (its RequiredObjects chunk), as the object
// manager's slots are upstream. One list for the title map and loaded games,
// so no type can be missed on one path (industry, airport and dock were, which
// gave title-map industries and their fields another industry's images).
func reorderScenarioObjects(om *objects.ObjectManager, sc *scenario.Scenario) {
	sim.ReorderObjectsForScenario(om, sc)
	useCurrency(om, sc.CurrencyObject)
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

// scenarioObjectiveText is loadScenarioProgress's objective string for a
// scenario in the list: formatChallengeArguments with the time limit's year
// (timeLimitYears + the start year − 1) and the delivered cargo's object
// (looked up by its DAT name, as loadTemporaryObject loads it). PARTIAL:
// the scenario's own currency for the amounts.
func (g *Game) scenarioObjectiveText(m *scenario.ScenarioMeta) string {
	o := m.Objective
	co := ui.ChallengeObjective{Type: o.Type, Flags: o.Flags, CompanyValue: o.CompanyValue, MonthlyProfit: o.MonthlyVehicleProfit,
		PerformanceIndex: o.PerformanceIndex, CargoAmount: o.DeliveredCargoAmount, TimeLimitYears: o.TimeLimitYears,
		UntilYear: int(o.TimeLimitYears) + int(m.StartYear) - 1}
	if g.r.ObjMgr != nil && m.ObjectiveCargoName != "" {
		if lo := g.r.ObjMgr.GetObject(m.ObjectiveCargoName); lo != nil {
			if c, ok := lo.Object.(*objects.CargoObject); ok {
				co.CargoUnitPlural = ui.ObjectString(c.UnitNamePlural)
			}
		}
	}
	return stripCodesKeepText(ui.ChallengeObjectiveText(co))
}

// stripCodesKeepText drops the {…} codes left in a formatted string (the
// list wraps plain text).
func stripCodesKeepText(s string) string {
	var b strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		j := strings.IndexByte(s[i:], '}')
		if j < 0 {
			break
		}
		s = s[i+j+1:]
	}
	return b.String()
}

// openNewGameWindow opens the OpenLoco-style scenario selection screen.
func (g *Game) openNewGameWindow() {
	// The ported ScenarioSelect window (pkg/ui/scenarioselect.go).
	metas, _ := scenario.ScanScenarios(g.dataDir + "/../Scenarios")
	if len(metas) == 0 {
		metas, _ = scenario.ScanScenarios(g.dataDir)
	}
	w := ui.NewScenarioSelect(metas, g.sw, g.sh)
	w.ObjectiveText = g.scenarioObjectiveText
	w.OnClose = func() { g.scenarioSelect = nil }
	w.OnStart = func(m *scenario.ScenarioMeta) {
		g.scenarioSelect = nil
		if err := g.loadScenario(m.FilePath); err != nil {
			log.Printf("[Game] Load error: %v", err)
		}
	}
	g.scenarioSelect = w
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
