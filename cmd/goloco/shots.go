package main

import (
	"fmt"
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

// debugPath prefixes capture filenames with the gitignored debug/ directory
// (created on demand) unless the caller already gave a path.
func debugPath(name string) string {
	if strings.ContainsRune(name, '/') {
		return name
	}
	_ = os.MkdirAll("debug", 0o755)
	return "debug/" + name
}
