package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
)

// save.go — saving a game: the Save Game prompt (ToolbarTop.cpp saveGame,
// Game::saveSaveGameOpen), S5::exportGameStateToFile (pkg/scenario's writer
// over the map's sim.State.ExportScenario), and the monthly autosave
// (GameScene.cpp autosaveCheck / autosave / autosaveClean).
//
// The folders are OpenLoco's, under goloco's own config folder (Platform::
// getUserDirectory: $XDG_CONFIG_HOME/goloco): save/ and save/autosave/.

// userDir is Platform::getUserDirectory: goloco's config folder.
func userDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goloco")
}

// savesDir is Environment::getPath(PathId::save): last_save_path when that
// folder exists (resolvePaths' tryPathOrDefault), else the user folder's save/.
func savesDir() string {
	if p := configValue("last_save_path"); p != "" {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			return p
		}
	}
	return filepath.Join(userDir(), "save")
}

// autosaveDir is PathId::autosave.
func autosaveDir() string { return filepath.Join(userDir(), "save", "autosave") }

// newGameSavePath is Scenario::start's setActiveSavePath: the saves folder,
// the scenario's name.
func newGameSavePath(scenarioName string) string {
	return filepath.Join(savesDir(), scenarioName+".SV5")
}

// openSaveWindow is ToolbarTop's saveGame: the Save Game prompt on the active
// save path (its folder, its name), then the save as .SV5; a failure opens
// the error window (error_game_save_failed).
func (g *Game) openSaveWindow() {
	if !g.inGame || g.inTitleScreen {
		return
	}
	dir, name := savesDir(), ""
	if g.activeSavePath != "" {
		dir = filepath.Dir(g.activeSavePath)
		name = strings.TrimSuffix(filepath.Base(g.activeSavePath), filepath.Ext(g.activeSavePath))
	}
	_ = os.MkdirAll(dir, 0o755)
	w := ui.NewPromptBrowseSave(i18n.Label(i18n.TitlePromptSaveGame), dir, name, ".SV5", g.sw, g.sh)
	w.Confirm = func(caption, text, ok string, then func()) {
		g.prompt = ui.NewPromptOkCancel(caption, text, ok, g.sw, g.sh, then)
	}
	w.OnSave = func(path string) {
		w.Close()
		rememberSavePath(w.Dir)
		path = strings.TrimSuffix(path, filepath.Ext(path)) + ".SV5" // replace_extension(extensionSV5)
		g.activeSavePath = path
		if err := g.saveGameTo(path, false); err != nil {
			log.Printf("[Game] Save error: %v", err)
			g.openError(i18n.Label(i18n.ErrorGameSaveFailed), "")
		}
	}
	g.showCentred(w)
}

// saveGameTo is S5::exportGameStateToFile for a saved game: the construction
// windows closed (unless an autosave: noWindowClose), the live state's
// pre-save clean-up (StationManager::zeroUnused …), then prepareGameState's
// file — the header, the save details, the packed objects, the object list,
// the game state with the main view saved in it, the tile elements.
func (g *Game) saveGameTo(path string, autosave bool) error {
	if !autosave {
		g.wins.CloseType(ui.WindowTypeConstruction, ui.AnyNumber) // closeConstructionWindows
		g.toolCancel()
	}
	m := &g.w.Map
	if m.Raw == nil {
		return fmt.Errorf("no game to save")
	}
	m.PrepareSave()
	sc := g.w.ExportScenario()
	sc.SavedViewX, sc.SavedViewY, sc.SavedViewZoom, sc.SavedViewRotation = g.w.SavedView()
	var packed []scenario.PackedObjectRaw
	if configValue("exportObjectsWithSaves") != "false" { // shouldPackObjects: Config exportObjectsWithSaves (default true)
		packed = g.packedObjects
	}
	f := &scenario.S5File{
		Header:          scenario.SavedGameHeader(len(packed)),
		SaveDetails:     g.saveDetails(),
		PackedObjects:   packed,
		RequiredObjects: g.requiredObjects(),
		GameState:       sc.EncodeGameState(),
		TileElements:    scenario.EncodeTileElements(sc.Tiles),
	}
	if g.script != nil {
		g.script.lastSave = append(append([]byte(nil), f.GameState...), f.TileElements...)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	log.Printf("[Game] Saved to: %s", path)
	return nil
}

// saveStateBytes is the game state and tile elements a save of w now would
// hold (the script's "expect saved").
func saveStateBytes(w *world.World) []byte {
	sc := w.ExportScenario()
	sc.SavedViewX, sc.SavedViewY, sc.SavedViewZoom, sc.SavedViewRotation = w.SavedView()
	return append(sc.EncodeGameState(), scenario.EncodeTileElements(sc.Tiles)...)
}

// saveDetails is prepareSaveDetails: the player company's name and owner,
// the date, its performance index and challenge, the scenario's name, the
// preview of the main view (drawSavePreviewImage, 250 × 200).
func (g *Game) saveDetails() []byte {
	m := &g.w.Map
	id := int(m.PlayerCompanies[0])
	d := scenario.SaveDetails{Date: m.CurrentDay, Image: g.w.SavePreview(250, 200)}
	var perf uint16
	if id < len(m.CompanyNames) {
		d.Company, d.Owner = m.CompanyName(id), m.OwnerName(id, g.r.ObjMgr)
	}
	if id < len(m.CompanyFaces) {
		perf, d.ChallengeFlags = uint16(m.CompanyFaces[id].Performance), m.CompanyFaces[id].ChallengeFlags
	}
	if id < len(m.Companies) {
		d.ChallengeProgress = m.Companies[id].ChallengeProgress
	}
	return scenario.EncodeSaveDetails(d, perf, m.ScenarioName)
}

// requiredObjects is ObjectManager::getHeaders: the object list the game was
// loaded with, the currency and competitor slots as they are now
// (sim.RequiredObjects).
func (g *Game) requiredObjects() []byte {
	return sim.RequiredObjects(g.w.Map.Raw.RequiredObjects, g.r.ObjMgr)
}

// autosaveCheck is GameScene's autosaveCheck, at every month's end: after
// autosave_frequency months (openloco.yml; 0 never), an autosave, the old
// ones cleaned away, the count reset. None on the title.
func (g *Game) autosaveCheck() {
	g.monthsSinceAutosave++
	if g.inTitleScreen || !g.inGame {
		return
	}
	freq := 1
	if v, err := strconv.Atoi(configValue("autosave_frequency")); err == nil {
		freq = v
	}
	if freq > 0 && g.monthsSinceAutosave >= freq {
		g.autosave()
		g.autosaveClean()
		g.monthsSinceAutosave = 0 // autosaveReset
	}
}

// autosave writes autosave_YYYY-MM-DD_HH-MM-SS.SV5 (the local time) into the
// autosave folder (isAutosave | noWindowClose).
func (g *Game) autosave() {
	name := "autosave_" + time.Now().Format("2006-01-02_15-04-05") + ".SV5"
	path := filepath.Join(autosaveDir(), name)
	log.Printf("[Game] Autosaving game to %s", path)
	if err := g.saveGameTo(path, true); err != nil {
		log.Printf("[Game] Unable to autosave game: %v", err)
	}
}

// autosaveClean deletes the oldest autosaves beyond autosave_amount (default
// 12, at least 1), by name — the date order.
func (g *Game) autosaveClean() {
	keep := 12
	if v, err := strconv.Atoi(configValue("autosave_amount")); err == nil {
		keep = max(1, v)
	}
	ents, err := os.ReadDir(autosaveDir())
	if err != nil {
		return
	}
	var files []string
	for _, e := range ents {
		n := e.Name()
		if e.Type().IsRegular() && strings.HasPrefix(n, "autosave_") && strings.EqualFold(filepath.Ext(n), ".SV5") {
			files = append(files, n)
		}
	}
	if len(files) <= keep {
		return
	}
	sort.Strings(files)
	for _, n := range files[:len(files)-keep] {
		log.Printf("[Game] Deleting old autosave: %s", n)
		_ = os.Remove(filepath.Join(autosaveDir(), n))
	}
}

// rememberSavePath is processFileForLoadSave for a saved game: the browser's
// folder becomes Config::lastSavePath, written to openloco.yml (so the next
// browser opens there: savesDir).
func rememberSavePath(dir string) {
	if dir == "" {
		return
	}
	if err := setConfigValue("last_save_path", dir); err != nil {
		log.Printf("[Game] saving last_save_path: %v", err)
	}
}
