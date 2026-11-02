package world

import (
	"log"
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
	// responsiveness"): while a window or menu is open the world is re-rendered
	// only every SceneEvery-th frame, the last picture reused in between, so the
	// UI runs at the full frame rate.
	w.sceneSkip++
	if w.SceneEvery > 1 && w.sceneImg != nil && w.sceneImg.Bounds().Dx() == sw && w.sceneImg.Bounds().Dy() == sh &&
		w.sceneSkip%w.SceneEvery != 0 {
		screen.DrawImage(w.sceneImg, nil)
		return
	}
	rt := w.scene.Render(&w.Map, scene.View{X: int(w.camX), Y: int(w.camY), W: sw, H: sh, Zoom: w.zoom, Ticks: uint32(w.animTick)})
	if w.sceneImg == nil || w.sceneImg.Bounds().Dx() != sw || w.sceneImg.Bounds().Dy() != sh {
		w.sceneImg = ebiten.NewImage(sw, sh)
	}
	w.sceneImg.WritePixels(w.scene.ToRGBA(rt).Pix)
	screen.DrawImage(w.sceneImg, nil)
}
