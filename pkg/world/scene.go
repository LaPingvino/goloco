package world

import (
	"os"

	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// The ported software renderer (pkg/scene: gfx8 + paint) as the world view —
// LOOP item 3j. Behind GOLOCO_SCENE=1 while it is incomplete: tracks,
// stations, signals, walls, bridges and industries are not drawn by it yet (the
// old per-tile painters still own those), and it covers zoom 0 only.

// sceneEnabled reports whether this frame is drawn by pkg/scene.
func (w *World) sceneEnabled() bool {
	return os.Getenv("GOLOCO_SCENE") == "1" && w.zoom == 0 &&
		w.renderer != nil && w.renderer.G1 != nil && w.renderer.ObjMgr != nil
}

// drawScene renders the visible map through pkg/scene and draws it on screen.
func (w *World) drawScene(screen *ebiten.Image) {
	if w.scene == nil {
		w.scene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	rt := w.scene.Render(&w.Map, scene.View{X: int(w.camX), Y: int(w.camY), W: sw, H: sh, Ticks: uint32(w.animTick)})
	if w.sceneImg == nil || w.sceneImg.Bounds().Dx() != sw || w.sceneImg.Bounds().Dy() != sh {
		w.sceneImg = ebiten.NewImage(sw, sh)
	}
	w.sceneImg.WritePixels(w.scene.ToRGBA(rt).Pix)
	screen.DrawImage(w.sceneImg, nil)
}
