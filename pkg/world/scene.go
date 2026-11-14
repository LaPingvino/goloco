package world

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// The ported software renderer (pkg/scene: gfx8 + paint) as the world view —
// LOOP item 3j. The DEFAULT since 2026-11-01 (Joop: "use that as the
// baseline"); it draws every zoom (3m) and the saved vehicles (paintEntities).
// GOLOCO_SCENE=0 still selects the old Ebitengine renderer, for comparison
// only — being retired (LOOP item RR). Still drawn outside the scene: the demo
// trains (maps without vehicles), the construction ghost and the label overlay.

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
		w.scene.WaveFrame = func(tx, ty int) uint8 { return uint8(w.Waves[sim.WaveIndex(tx, ty)].Frame) }
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
	view := scene.View{X: int(w.camX), Y: int(w.camY), W: sw, H: sh, Zoom: w.zoom, Ticks: w.Map.ScenarioTicks, Flags: w.viewFlags} // ScenarioManager::getScenarioTicks(): what upstream's paint animates by (not the frame count)
	if diagScreen != nil {                                                                                                         // GOLOCO_DIAG_PAINT_SCREEN: the paint diagnostic's pixel, as a screen pixel at this view's zoom
		bm := ^0 << view.Zoom
		paint.DiagAt = &[2]int{view.X&bm + diagScreen[0]<<view.Zoom, view.Y&bm + diagScreen[1]<<view.Zoom}
	}
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
	if gpuPaletteOn && rt.Width == sw && rt.Height == sh && gpuPaletteFits(sw) && w.scene.G1 != nil {
		if g := w.gpuPalFor(sw, sh); g != nil {
			if !g.palSet {
				g.setPalette(&w.scene.G1.Palette)
			}
			for _, r := range ranges {
				g.upload(rt, r[0], r[1])
			}
			g.draw(screen)
		}
	}
	if !gpuPaletteOn || w.gpal == nil || rt.Width != sw || rt.Height != sh || !gpuPaletteFits(sw) {
		for _, r := range ranges {
			if r[0] == 0 && r[1] == sw {
				w.sceneImg.WritePixels(w.scene.ToRGBA(rt).Pix)
			} else { // only the repainted columns converted and uploaded
				w.sceneImg.SubImage(image.Rect(r[0], 0, r[1], sh)).(*ebiten.Image).WritePixels(w.scene.ToRGBAColumns(rt, r[0], r[1]))
			}
		}
		screen.DrawImage(w.sceneImg, nil)
	}
	if rt.Width == sw && rt.Height == sh { // the palette indices under the UI, for draw8 (no GPU readback)
		w.lastBits = rt.Bits
		ui.BeginFrame8(rt.Bits, sw, sh)
	}
}

// dirtyRegions (on by default since 2026-11-05; GOLOCO_DIRTY=0 turns it off):
// the world drawn as upstream draws it, into a screen that persists,
// repainted by renderDirtyRegions: the invalidation grid's rects, each
// through Viewport::paint(rect) (scene.RenderRect, upstream's column split).
// GOLOCO_DIRTY_CHECK=N compares with a full repaint every Nth frame.
var dirtyRegions = os.Getenv("GOLOCO_DIRTY") != "0"

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
	if w.invalTile == nil {
		w.invalTile = w.invalidateTile
	}
	w.Map.InvalidateTile = w.invalTile // ViewportManager::invalidate(pos, …)
	moved := view.X != w.dirtyView.X || view.Y != w.dirtyView.Y || view.W != w.dirtyView.W || view.H != w.dirtyView.H || view.Zoom != w.dirtyView.Zoom
	if fresh || moved { // (Ticks — the animation clock — changes every frame: not a new view)
		// PARTIAL (D4): upstream's Viewport::move shifts the pixels already drawn
		// and invalidates only the uncovered strip; a new view repaints it all.
		w.dirtyGrid.Invalidate(0, 0, sw, sh)
	}
	// The invalidations (D3) are ported wherever goloco has the code that
	// changes the map: entities moving or changing their picture (moveTo,
	// invalidateSprite: body turns, wheel frames, brake lights), signals, level
	// crossings, roads, trees, buildings removed, terraforming, station cargo,
	// animations. PARTIAL: code goloco doesn't have yet brings its own
	// invalidations with it when ported (BuildingElement::update's construction
	// stages, a vehicle's body animation and cargo frames). Differences the
	// GOLOCO_DIRTY_CHECK still reports come from the column-split sort, as in
	// upstream (LOOP 2026-11-05).
	w.dirtyView = view
	rt := w.dirtyRT
	// The colours on the GPU (FPS (3)) when the screen has room for the
	// palette row; else the CPU conversion of each dirty rect.
	var gp *gpuPal
	if gpuPaletteOn && gpuPaletteFits(sw) && w.scene.G1 != nil {
		gp = w.gpuPalFor(sw, sh)
	}
	if gp != nil {
		if !gp.palSet {
			gp.setPalette(&w.scene.G1.Palette)
		}
		if gp.fresh { // a new packed image: every pixel uploaded once
			w.dirtyGrid.Invalidate(0, 0, sw, sh)
		}
	}
	w.dirtyGrid.TraverseDirtyCells(func(l, t, r, b int) {
		w.scene.RenderRect(&w.Map, view, rt, l, t, r, b)
		if gp != nil {
			gp.uploadRect(rt, l, t, r, b)
		} else {
			w.sceneImg.SubImage(image.Rect(l, t, r, b)).(*ebiten.Image).WritePixels(w.scene.ToRGBARect(rt, l, t, r, b))
		}
	})
	w.lastView = view
	if dirtyCheck > 0 {
		w.checkDirty(view)
	}
	if gp != nil {
		gp.draw(screen)
	} else {
		screen.DrawImage(w.sceneImg, nil)
	}
	w.lastBits = rt.Bits
	ui.BeginFrame8(rt.Bits, sw, sh)
}

// invalidateEntity is ViewportManager::invalidate(entity, ZoomLevel::eighth)
// for the main viewport: skipped when zoomed out past min(vehiclesMinScale,
// eighth) (Config default 2); the sprite rect intersected with the view
// (getIntersection), made relative to the view, shifted by the zoom, and
// invalidated on the screen (the main viewport sits at 0, 0).
func (w *World) invalidateEntity(left, top, right, bottom int16) {
	w.invalidateRect(int(left), int(top), int(right), int(bottom), min(2, sim.ZoomEighth)) // min(vehiclesMinScale (default 2), eighth)
}

// invalidateTile is ViewportManager::invalidate(pos, zMin, zMax, zoom,
// radius): from gameToScreen(pos + 16, zMax) less radius to gameToScreen(pos +
// 16, zMin) plus radius, at the current rotation.
func (w *World) invalidateTile(x, y, zMin, zMax int32, zoom uint8, radius int32) {
	ax, ay := w.Map.GameToScreen(x+16, y+16, zMax)
	dx, dy := w.Map.GameToScreen(x+16, y+16, zMin)
	w.invalidateRect(int(ax-radius), int(ay-radius), int(dx+radius), int(dy+radius), zoom)
}

// invalidateRect is ViewportManager's static invalidate(rect, zoom) for the
// main viewport: skipped when it is zoomed out further than zoom; the rect ∩
// the view (getIntersection), made relative to the view, shifted by the zoom,
// invalidated on the screen (the main viewport sits at 0, 0).
func (w *World) invalidateRect(l, t, r, b int, zoom uint8) {
	v := w.dirtyView
	if v.W == 0 || v.Zoom > int(zoom) {
		return
	}
	vw, vh := v.W<<v.Zoom, v.H<<v.Zoom
	if r <= v.X || b <= v.Y || l >= v.X+vw || t >= v.Y+vh { // !viewport.intersects(rect)
		return
	}
	l, r = max(l, v.X), min(r, v.X+vw)
	t, b = max(t, v.Y), min(b, v.Y+vh)
	w.dirtyGrid.Invalidate((l-v.X)>>v.Zoom, (t-v.Y)>>v.Zoom, (r-v.X)>>v.Zoom, (b-v.Y)>>v.Zoom)
}

// dirtyCheck (GOLOCO_DIRTY_CHECK=N, diagnostic): every Nth frame the whole
// view is repainted into a second screen and compared with the dirty-region
// one; differing pixels are logged and the first frames with any are saved
// (debug/dirtycheck_NN.png: the difference in red over the full repaint) —
// what an invalidation upstream does that goloco misses.
var dirtyCheck = func() int {
	n, _ := strconv.Atoi(os.Getenv("GOLOCO_DIRTY_CHECK"))
	return n
}()

var dirtyCheckMin = func() int {
	n, err := strconv.Atoi(os.Getenv("GOLOCO_DIRTY_CHECK_MIN"))
	if err != nil || n < 1 {
		return 1
	}
	return n
}()

func (w *World) checkDirty(view scene.View) {
	w.dirtyChecks++
	if w.dirtyChecks%dirtyCheck != 0 {
		return
	}
	// the reference is upstream's own full repaint: the whole screen invalidated,
	// then renderDirtyRegions — the grid's 64-wide rects, not one rect (the
	// column split decides some pixels: one rect would differ from upstream too)
	full := gfx8.NewRenderTarget(view.W, view.H)
	var g gfx8.InvalidationGrid
	g.Reset(view.W, view.H, gfx8.InvalidationBlockWidth, gfx8.InvalidationBlockHeight)
	g.Invalidate(0, 0, view.W, view.H)
	g.TraverseDirtyCells(func(l, t, r, b int) { w.scene.RenderRect(&w.Map, view, full, l, t, r, b) })
	diff := 0
	for i := range full.Bits {
		if full.Bits[i] != w.dirtyRT.Bits[i] {
			diff++
		}
	}
	log.Printf("[DirtyCheck] frame %d: %d of %d pixels differ from a full repaint (view %d,%d zoom %d — GOLOCO_DIAG_PAINT_AT takes view + screen pixel)", w.dirtyChecks, diff, len(full.Bits), view.X, view.Y, view.Zoom)
	if diff < dirtyCheckMin || w.dirtySaved >= 6 { // GOLOCO_DIRTY_CHECK_MIN: the smallest difference saved (default 1)
		return
	}
	w.dirtySaved++
	img := w.scene.ToRGBA(full)
	out := image.NewRGBA(img.Rect)
	copy(out.Pix, img.Pix)
	for i := range full.Bits {
		if full.Bits[i] != w.dirtyRT.Bits[i] {
			out.Pix[i*4], out.Pix[i*4+1], out.Pix[i*4+2] = 255, 0, 0
		}
	}
	if f, err := os.Create(fmt.Sprintf("debug/dirtycheck_%02d.png", w.dirtySaved)); err == nil {
		png.Encode(f, out)
		f.Close()
	}
}

// diagScreen (GOLOCO_DIAG_PAINT_SCREEN=x,y): GOLOCO_DIAG_PAINT_AT for a screen
// pixel, whatever the zoom.
var diagScreen = func() *[2]int {
	var p [2]int
	if _, err := fmt.Sscanf(os.Getenv("GOLOCO_DIAG_PAINT_SCREEN"), "%d,%d", &p[0], &p[1]); err != nil {
		return nil
	}
	return &p
}()

// SetCatchmentDisplay sets (nil: clears) the map selection's catchment area —
// a byte per tile, bit 1 inside — invalidating its tiles in full
// (sub_491BC6: mapInvalidateTileFull = invalidate(pos, 0, 1120, eighth)).
func (w *World) SetCatchmentDisplay(cm []uint8) {
	inval := func(c []uint8) {
		for i, v := range c {
			if v&(1<<1) != 0 {
				w.invalidateTile(int32(i%w.Map.Width)*32, int32(i/w.Map.Width)*32, 0, 1120, 3, 32)
			}
		}
	}
	if w.catchment == nil && cm == nil {
		return
	}
	inval(w.catchment)
	w.catchment = cm
	if w.scene == nil {
		return
	}
	inval(cm)
	if cm == nil {
		w.scene.Catchment = nil
		return
	}
	mw := w.Map.Width
	w.scene.Catchment = func(tx, ty int) bool {
		i := ty*mw + tx
		return tx >= 0 && tx < mw && i >= 0 && i < len(cm) && cm[i]&(1<<1) != 0
	}
}

// SetMapSelection sets the map selection (setMapSelectionFlags(enable) with
// setMapSelectionArea, MapSelectionType::full) to the tiles a..b inclusive,
// or clears it (on false: resetMapSelectionFlag(enable)); the old and new
// areas are invalidated (mapInvalidateSelectionRect).
func (w *World) SetMapSelection(on bool, ax, ay, bx, by int) {
	w.SetMapSelectionType(on, scene.SelFull, ax, ay, bx, by)
}

// SetMapSelectionType is SetMapSelection with the selection's type
// (setMapSelectionCorner: a corner, the whole tile, the water's surface, a
// quarter or an edge — scene.SelCorner0 …). It returns how many of the
// flag, the type and the area's four coordinates changed (the count
// setMapSelectionTiles / setMapSelectionSingleTile return: 0, nothing did).
func (w *World) SetMapSelectionType(on bool, typ uint8, ax, ay, bx, by int) int {
	area := [4]int{ax, ay, bx, by}
	if !on {
		area, typ = w.selArea, w.selType // resetMapSelectionFlag keeps the area and type
	}
	count := 0
	if on != w.selOn {
		count++
	}
	if typ != w.selType {
		count++
	}
	for i := range area {
		if area[i] != w.selArea[i] {
			count++
		}
	}
	if count == 0 {
		return 0
	}
	inval := func(a [4]int) {
		for ty := a[1]; ty <= a[3]; ty++ {
			for tx := a[0]; tx <= a[2]; tx++ {
				w.invalidateTile(int32(tx)*32, int32(ty)*32, 0, 1120, 3, 32)
			}
		}
	}
	if w.selOn {
		inval(w.selArea)
	}
	w.selOn, w.selArea, w.selType = on, area, typ
	if w.scene == nil {
		return count
	}
	if !on {
		w.scene.Selection = nil
		return count
	}
	inval(area)
	w.scene.SelectionType = typ
	w.scene.Selection = func(tx, ty int) bool {
		a := w.selArea
		return tx >= a[0] && tx <= a[2] && ty >= a[1] && ty <= a[3]
	}
	return count
}

// MapSelection is hasMapSelectionFlag(enable), getMapSelectionCorner and
// getMapSelectionArea (tiles: min x, min y, max x, max y).
func (w *World) MapSelection() (on bool, typ uint8, area [4]int) {
	return w.selOn, w.selType, w.selArea
}

// InvalidateView is the main window's invalidate(): the whole view redrawn
// (the dirty grid's every cell; without dirty regions, the next frame whole).
func (w *World) InvalidateView() {
	v := w.dirtyView
	w.dirtyGrid.Invalidate(0, 0, v.W, v.H)
	w.lastView = scene.View{}
}

// SurfaceAt is ViewportInteraction::getSurfaceLocFromUi (main view): the map
// position (world units, not snapped to the tile) on the surface under the
// screen point; false when no surface is there.
func (w *World) SurfaceAt(sx, sy int) (int32, int32, bool) {
	return w.surfaceLocAt(sx, sy, false)
}
