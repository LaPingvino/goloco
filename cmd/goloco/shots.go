package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/paint"
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
