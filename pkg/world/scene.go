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
	if dirtyRegions {
		w.drawSceneDirty(screen, view)
		return
	}
	var rt *gfx8.RenderTarget
	same := view.X == w.lastView.X && view.Y == w.lastView.Y && view.W == w.lastView.W && view.H == w.lastView.H && view.Zoom == w.lastView.Zoom
	if w.sceneImg == nil || w.sceneImg.Bounds().Dx() != sw || w.sceneImg.Bounds().Dy() != sh {
		w.sceneImg = ebiten.NewImage(sw, sh)
		same = false // a new image: all of it
	}
	parts := 1
	if same && (w.SceneEvery > 1 || w.frameCost > sceneBudgetMS) {
		// Adaptive (ticks 309, 311): as many parts as keep the world within
		// the budget, judged by what a whole frame costs (an average); at
		// least SceneEvery while a window is open, at most a column a part.
		parts = max(w.SceneEvery, int(math.Ceil(w.frameCost/sceneBudgetMS)))
		parts = min(parts, max(sw/32, 1))
		w.sceneSkip = (w.sceneSkip + 1) % parts
	} else {
		w.sceneSkip = 0
	}
	t1 := time.Now()
	var ranges [][2]int
	if w.SceneEvery > 1 { // a window open: the part alone
		var x0, x1 int
		rt, x0, x1 = w.scene.RenderPart(&w.Map, view, w.sceneSkip, parts)
		ranges = [][2]int{{x0, x1}}
	} else { // in play: the part and every column with a vehicle (no torn trains)
		rt, ranges = w.scene.RenderPlay(&w.Map, view, w.sceneSkip, parts)
	}
	if p := ranges[0]; p[1] > p[0] { // the whole frame's cost, from this frame's part
		painted := 0
		for _, r := range ranges {
			painted += r[1] - r[0]
		}
		whole := float64(time.Since(t1).Microseconds()) / 1000 * float64(sw) / float64(painted)
		if w.frameCost == 0 {
			w.frameCost = whole
		} else {
			w.frameCost += (whole - w.frameCost) / 8
		}
	}
	w.SceneParts = parts
	w.lastView = view
	for _, r := range ranges {
		if r[0] == 0 && r[1] == sw {
			w.sceneImg.WritePixels(w.scene.ToRGBA(rt).Pix)
		} else { // only the repainted columns converted and uploaded
			w.sceneImg.SubImage(image.Rect(r[0], 0, r[1], sh)).(*ebiten.Image).WritePixels(w.scene.ToRGBAColumns(rt, r[0], r[1]))
		}
	}
	screen.DrawImage(w.sceneImg, nil)
	if rt.Width == sw && rt.Height == sh { // the palette indices under the UI, for draw8 (no GPU readback)
		w.lastBits = rt.Bits
		ui.BeginFrame8(rt.Bits, sw, sh)
	}
}

// dirtyRegions (GOLOCO_DIRTY=1, while the port is under way — LOOP's DIRTY
// REGIONS D2): the world drawn as upstream draws it, into a screen that
// persists, repainted by renderDirtyRegions: the invalidation grid's rects,
// each through Viewport::paint(rect) (scene.RenderRect, upstream's column
// split). Until the invalidations are ported (D3) the whole screen is
// invalidated every frame.
var dirtyRegions = os.Getenv("GOLOCO_DIRTY") != ""

func (w *World) drawSceneDirty(screen *ebiten.Image, view scene.View) {
	sw, sh := view.W, view.H
	fresh := false
	if w.dirtyRT == nil || w.dirtyRT.Width != sw || w.dirtyRT.Height != sh {
		w.dirtyRT = gfx8.NewRenderTarget(sw, sh)
		w.dirtyGrid.Reset(sw, sh, gfx8.InvalidationBlockWidth, gfx8.InvalidationBlockHeight)
		fresh = true
	}
	if w.sceneImg == nil || w.sceneImg.Bounds().Dx() != sw || w.sceneImg.Bounds().Dy() != sh {
		w.sceneImg = ebiten.NewImage(sw, sh)
	}
	if w.invalEntity == nil {
		w.invalEntity = w.invalidateEntity // a method value allocates: made once
	}
	w.Map.InvalidateEntity = w.invalEntity // EntityBase::invalidateSprite → this screen (the Map is replaced on load)
	moved := view.X != w.dirtyView.X || view.Y != w.dirtyView.Y || view.W != w.dirtyView.W || view.H != w.dirtyView.H || view.Zoom != w.dirtyView.Zoom
	if fresh || moved { // (Ticks — the animation clock — changes every frame: not a new view)
		// PARTIAL (D4): upstream's Viewport::move shifts the pixels already drawn
		// and invalidates only the uncovered strip; a new view repaints it all.
		w.dirtyGrid.Invalidate(0, 0, sw, sh)
	}
	// PARTIAL (D3): the tile invalidations (town growth, construction, signals,
	// level crossings, animations — ViewportManager::invalidate(pos, …)) are not
	// ported yet, so GOLOCO_DIRTY stays off by default.
	w.dirtyView = view
	rt := w.dirtyRT
	w.dirtyGrid.TraverseDirtyCells(func(l, t, r, b int) {
		w.scene.RenderRect(&w.Map, view, rt, l, t, r, b)
		w.sceneImg.SubImage(image.Rect(l, t, r, b)).(*ebiten.Image).WritePixels(w.scene.ToRGBARect(rt, l, t, r, b))
	})
	w.lastView = view
	screen.DrawImage(w.sceneImg, nil)
	w.lastBits = rt.Bits
	ui.BeginFrame8(rt.Bits, sw, sh)
}

// invalidateEntity is ViewportManager::invalidate(entity, ZoomLevel::eighth)
// for the main viewport: skipped when zoomed out past min(vehiclesMinScale,
// eighth) (Config default 2); the sprite rect intersected with the view
// (getIntersection), made relative to the view, shifted by the zoom, and
// invalidated on the screen (the main viewport sits at 0, 0).
func (w *World) invalidateEntity(left, top, right, bottom int16) {
	v := w.dirtyView
	if v.W == 0 || w.zoom > min(2, 3) { // vehiclesMinScale (Config default 2), ZoomLevel::eighth
		return
	}
	vw, vh := v.W<<v.Zoom, v.H<<v.Zoom
	l, t, r, b := int(left), int(top), int(right), int(bottom)
	if r <= v.X || b <= v.Y || l >= v.X+vw || t >= v.Y+vh { // !viewport.intersects(rect)
		return
	}
	l, r = max(l, v.X), min(r, v.X+vw)
	t, b = max(t, v.Y), min(b, v.Y+vh)
	w.dirtyGrid.Invalidate((l-v.X)>>v.Zoom, (t-v.Y)>>v.Zoom, (r-v.X)>>v.Zoom, (b-v.Y)>>v.Zoom)
}
