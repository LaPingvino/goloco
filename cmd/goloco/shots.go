package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// shots.go — captures of the finished frame: --shots, a script's "shot".

// maybeSaveShot ends a frame's Draw: the diagnostics that act on or draw over
// the finished frame (diag.go), then the captures — a script's "shot", and
// --shots mode: the full frame saved every shotsEvery, quitting when the
// requested count has been captured.
func (g *Game) maybeSaveShot(screen *ebiten.Image) {
	g.diagOpenWindows()
	g.diagDrawPanels(screen)
	if name := g.scriptShotName(); name != "" {
		name = debugPath(name)
		if err := savePNG(name, screen.SubImage(screen.Bounds())); err == nil {
			log.Printf("[Script] saved %s", name)
		}
	}
	if g.kb.screenshotNow {
		g.kb.screenshotNow = false
		g.saveScreenshot(screen)
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
	if err := savePNG(name, screen.SubImage(screen.Bounds())); err != nil {
		log.Printf("[Shots] %v", err)
	} else {
		log.Printf("[Shots] saved %s (%d remaining)", name, g.shotsLeft)
	}
	g.shotsLeft--
	if g.shotsLeft == 0 {
		g.quitAfterSave = true
	}
}

// savePNG writes img to the file name; the error says which step failed
// ("create name: …", "encode name: …").
func savePNG(name string, img image.Image) error {
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	return nil
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

// triggerScreenshotCountdown is Ui::triggerScreenshotCountdown (the
// screenshot shortcut: 2 ticks; the Load/Save menu's entry: 10 — time for
// the menu to go).
func (g *Game) triggerScreenshotCountdown(ticks int) { g.kb.screenshotCountdown = ticks }

// handleScreenshotCountdown is Ui::handleScreenshotCountdown: at the
// countdown's end the next frame drawn is saved (saveScreenshot).
func (g *Game) handleScreenshotCountdown() {
	if g.kb.screenshotCountdown != 0 {
		g.kb.screenshotCountdown--
		g.kb.screenshotNow = g.kb.screenshotCountdown == 0
	}
}

// screenshotsDir is Environment's PathId::screenshots for goloco: beside its
// saves.
func screenshotsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goloco", "screenshots")
}

// saveScreenshot is Screenshot.cpp's saveScreenshot / prepareSaveScreenshot:
// the frame as "<scenario name>.png" in the screenshots folder — "… (1).png"
// and on when that exists —, then the quiet error window "Screenshot saved
// to disk as ‘name’", or "Screenshot failed!".
// PARTIAL: the name is the scenario file's (gameState.scenarioName isn't
// kept), "Screenshot" without one; an RGBA PNG, not the 8-bit screen's;
// the giant screenshot.
func (g *Game) saveScreenshot(screen *ebiten.Image) {
	dir := screenshotsDir()
	name := strings.TrimSuffix(filepath.Base(g.currentScenarioPath), filepath.Ext(g.currentScenarioPath))
	if g.currentScenarioPath == "" || name == "" {
		name = i18n.Label(i18n.ScreenshotFilenameTemplate)
	}
	file := name + ".png"
	err := os.MkdirAll(dir, 0o755)
	for suffix := 1; err == nil; suffix++ {
		if _, statErr := os.Stat(filepath.Join(dir, file)); statErr != nil {
			break
		}
		if suffix == 1<<15-1 {
			err = fmt.Errorf("no free file name")
		}
		file = fmt.Sprintf("%s (%d).png", name, suffix)
	}
	if err == nil {
		err = savePNG(filepath.Join(dir, file), screen.SubImage(screen.Bounds()))
	}
	if err != nil {
		log.Printf("[Game] screenshot: %v", err)
		g.openError(i18n.Label(i18n.ScreenshotFailed), "")
		return
	}
	log.Printf("[Game] screenshot saved as %s", filepath.Join(dir, file))
	g.openError(ui.FormatErrorArgs(i18n.Label(i18n.ScreenshotSavedAs), file), "") // Error::openQuiet
}
