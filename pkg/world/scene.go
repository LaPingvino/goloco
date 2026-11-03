package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/ui"
	"image"
	"log"
	"math"
	"os"
	"time"

	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// The ported software renderer (pkg/scene: gfx8 + paint) as the world view —
// LOOP item 3j. The DEFAULT since 2026-11-01 (Joop: "use that as the
// baseline"); GOLOCO_SCENE=0 selects the old Ebitengine renderer. PARTIAL: it
// covers zoom 0 only — at zoom 1-3 the old renderer still draws the frame (item
// 3m) — and vehicles are drawn over it by the old post-pass (item 5).

// sceneEnabled reports whether this frame is drawn by pkg/scene.
func (w *World) sceneEnabled() bool {
	// Every zoom (3m): GOLOCO_SCENE=0 selects the old renderer.
	return os.Getenv("GOLOCO_SCENE") != "0" &&
		w.renderer != nil && w.renderer.G1 != nil && w.renderer.ObjMgr != nil
}

// drawScene renders the visible map through pkg/scene and draws it on screen.
// sceneBudgetMS is the world's share of a frame while a window is open.
const sceneBudgetMS = 8.0

func (w *World) drawScene(screen *ebiten.Image) {
	if w.scene == nil {
		w.scene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	t0 := time.Now()
	defer func() {
		w.sceneTime += time.Since(t0)
		if w.sceneFrames++; w.sceneFrames == 120 {
			log.Printf("[Scene] CPU renderer: %v per frame (avg of 120, %dx%d)", (w.sceneTime / 120).Round(100*time.Microsecond), sw, sh)
			w.sceneFrames, w.sceneTime = 0, 0
		}
	}()
	// Responsiveness first (LOOP item 0, Joop: "give up internal speed for
	// responsiveness"): while a window or menu is open the world's columns
	// are repainted a part a frame (SceneEvery parts), so no frame waits for
	// a whole world render; a moved camera, zoom or size repaints all of it.
	view := scene.View{X: int(w.camX), Y: int(w.camY), W: sw, H: sh, Zoom: w.zoom, Ticks: uint32(w.animTick)}
	var rt *gfx8.RenderTarget
	var x0, x1 int
	same := view.X == w.lastView.X && view.Y == w.lastView.Y && view.W == w.lastView.W && view.H == w.lastView.H && view.Zoom == w.lastView.Zoom
	if w.sceneImg == nil || w.sceneImg.Bounds().Dx() != sw || w.sceneImg.Bounds().Dy() != sh {
		w.sceneImg = ebiten.NewImage(sw, sh)
		same = false // a new image: all of it
	}
	parts := 1
	if w.SceneEvery > 1 && same {
		// Adaptive (tick 309): as many parts as keep the world within the
		// budget, judged by what a whole frame costs (an average); at least
		// SceneEvery, at most a column a part.
		parts = max(w.SceneEvery, int(math.Ceil(w.frameCost/sceneBudgetMS)))
		parts = min(parts, max(sw/32, 1))
		w.sceneSkip = (w.sceneSkip + 1) % parts
	} else {
		w.sceneSkip = 0
	}
	t1 := time.Now()
	rt, x0, x1 = w.scene.RenderPart(&w.Map, view, w.sceneSkip, parts)
	if x1 > x0 { // the whole frame's cost, from this part's
		whole := float64(time.Since(t1).Microseconds()) / 1000 * float64(sw) / float64(x1-x0)
		if w.frameCost == 0 {
			w.frameCost = whole
		} else {
			w.frameCost += (whole - w.frameCost) / 8
		}
	}
	w.SceneParts = parts
	w.lastView = view
	if x0 == 0 && x1 == sw {
		w.sceneImg.WritePixels(w.scene.ToRGBA(rt).Pix)
	} else { // only the repainted columns converted and uploaded
		w.sceneImg.SubImage(image.Rect(x0, 0, x1, sh)).(*ebiten.Image).WritePixels(w.scene.ToRGBAColumns(rt, x0, x1))
	}
	screen.DrawImage(w.sceneImg, nil)
	if rt.Width == sw && rt.Height == sh { // the palette indices under the UI, for draw8 (no GPU readback)
		w.lastBits = rt.Bits
		ui.BeginFrame8(rt.Bits, sw, sh)
	}
}
