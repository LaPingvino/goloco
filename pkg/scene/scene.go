// Package scene renders a view of the map through the ported software renderer
// (pkg/gfx8 + pkg/paint): it walks the visible tiles of a sim.Map, calls the
// ported painters in upstream's per-tile order, sorts with the paint session and
// draws into an 8-bit target. No rendering-library dependencies — used by
// cmd/render8 (headless) and by the game (LOOP item 3j).
package scene

import (
	"image"
	"image/color"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// Renderer holds the graphics and objects a scene is drawn with, and a cache of
// decoded sprites.
type Renderer struct {
	G1      *assets.G1File
	Objects *objects.ObjectManager
	sprites *spriteCache // shared by the column workers
	water   paint.WaterImages
	// workers paint columns in parallel (Viewport::paint's
	// std::execution::par): clones with their own session and scratch.
	workers   []*Renderer
	cols      [][2]int // RenderColumns' column list, kept
	need      []bool   // RenderPlay's columns with vehicles, kept
	ranges    [][2]int // RenderPlay's repainted ranges, kept
	edges     []int    // RenderPart's column edges, kept
	partPix   []byte   // ToRGBAColumns' buffer, kept
	shadowMap []uint8

	// Reused across frames (Render/ToRGBA allocate nothing in steady state).
	sess  *paint.Session
	rt    *gfx8.RenderTarget
	colRT *gfx8.RenderTarget // one column's target (RenderColumns)
	// spriteFor is r.Sprite as a func value, kept for the paint sessions.
	spriteFor func(int) *gfx8.Sprite
	rgba      *image.RGBA

	colourMaps map[int][]uint8 // GetPaletteMap(colour), which copies, cached
	// Per-frame scratch kept for a frame without allocations: vehicle
	// entities by tile (slices truncated, keys kept), two-colour palette maps
	// (GetPaletteMap2 allocates), and the palette lookup handed to the painters.
	entsByTile map[[2]int][]*scenario.VehicleEntity
	// station cargo inputs, kept (stationCargo)
	cargoPlatforms map[int]*paint.CargoPlatform
	cargoSlots     map[string]int               // cargoSlot cache (GetName allocates)
	scaffolds      map[[2]uint8]*paint.Scaffold // scaffold inputs, kept
	stationCargos  map[stationCargoKey]*paint.StationCargo
	pal2           map[[2]uint8][]uint8
	pal2Fn         func(p, q uint8) []uint8
	anims          map[any]anim     // per object: part animation frames and speeds
	tunnels        map[string]int   // tunnel header name → tunnel image base
	nbBuf          [4]paint.Surface // neighbour surfaces of the tile being painted
}

type anim struct{ frames, speeds []uint8 }

// colourMap returns the remap for a colour (cached; G1File.GetPaletteMap copies).
func (r *Renderer) colourMap(c int) []uint8 {
	if m, ok := r.colourMaps[c]; ok {
		return m
	}
	m, _ := r.G1.GetPaletteMap(c)
	if r.colourMaps == nil {
		r.colourMaps = map[int][]uint8{}
	}
	r.colourMaps[c] = m
	return m
}

// companyRemap is CompanyManager::getCompanyColour(owner) as a palette remap:
// the scenario's company colours (neutral = grey). nil when the map carries
// none (a scenario without saved companies), which draws the images uncoloured.
func (r *Renderer) companyRemap(m *sim.Map, owner uint8) []uint8 {
	if int(owner) >= len(m.CompanyColours) {
		return nil
	}
	return r.colourMap(int(m.CompanyColours[owner]))
}

// canopyTranslucent is the station canopy's translucent palette map:
// Colours::getTranslucent(companyColour) (shade 1), or ExtColour::unk2E when the
// station object is not recolourable. nil without company colours.
func (r *Renderer) canopyTranslucent(m *sim.Map, owner uint8, recolourable bool) []uint8 {
	if int(owner) >= len(m.CompanyColours) {
		return nil
	}
	ext := paint.ExtColourUnk2E
	if recolourable {
		ext = int(paint.TranslucentColour[m.CompanyColours[owner]][1])
	}
	return r.colourMap(ext)
}

// colourOrange is Colour::orange, the shadow colour of buildings, airports, docks.
const colourOrange = 18

// shadow builds a paint.Shadow: image (before + rotation) through
// Colours::getShadow(colour). zoom is always 0 here (upstream: zoomLevel <= 1).
// zoom: the shadows are drawn only up to zoom 1 (zoomLevel <= 1 in each painter).
func (r *Renderer) shadow(zoom int, has bool, image, colour int, primary bool) paint.Shadow {
	if !has || zoom > 1 || colour < 0 || colour >= len(paint.ShadowColour) {
		return paint.Shadow{}
	}
	return paint.Shadow{Image: image, Pal: r.colourMap(int(paint.ShadowColour[colour])), Primary: primary}
}

// paintBridge hands the session's bridge entry to paint.PaintBridge with its
// bridge object (scenario bridge slot).
// extColourUnk32 is ExtColour::unk32, the bridge shadow's translucency.
const extColourUnk32 = 0x32

func (r *Renderer) paintBridge(s *paint.Session, b paint.BridgeEntry, tx, ty int) bool {
	bo := r.Objects.GetBridgeObjectByIndex(int(b.ObjectID))
	if bo == nil {
		return false
	}
	return s.PaintBridge(b, paint.BridgeObj{Image: int(bo.ImageOffset), SpanLength: bo.SpanLength, PillarSpacing: bo.PillarSpacing,
		HasRoof: bo.Flags&objects.BridgeFlagHasRoof != 0, DeckDepth: int(bo.DeckDepth), TileX: tx, TileY: ty,
		ShadowPal: r.colourMap(extColourUnk32)}, r.Sprite)
}

// tunnelImage resolves a track/road object's tunnel header name to the tunnel
// object's image base once (ObjectManager.TunnelImage), then from a cache.
func (r *Renderer) tunnelImage(name string) int {
	if v, ok := r.tunnels[name]; ok {
		return v
	}
	if r.tunnels == nil {
		r.tunnels = map[string]int{}
	}
	v := r.Objects.TunnelImage(name)
	r.tunnels[name] = v
	return v
}

// animOf splits an object's part animations into frames and speeds (cached).
func (r *Renderer) animOf(key any, pa []objects.BuildingPartAnimation) anim {
	if a, ok := r.anims[key]; ok {
		return a
	}
	a := anim{make([]uint8, len(pa)), make([]uint8, len(pa))}
	for i, p := range pa {
		a.frames[i], a.speeds[i] = p.NumFrames, p.AnimationSpeed
	}
	if r.anims == nil {
		r.anims = map[any]anim{}
	}
	r.anims[key] = a
	return a
}

// New prepares a renderer. The object manager must already be reordered to the
// scenario's slots and have its cliff edges resolved.
func New(g1 *assets.G1File, om *objects.ObjectManager) *Renderer {
	r := &Renderer{G1: g1, Objects: om, sprites: &spriteCache{m: map[int]*gfx8.Sprite{}}}
	if wo := om.WaterObj; wo != nil {
		r.water.Base = int(wo.ImageOffset)
		r.water.PalMap, _ = g1.GetPaletteMapAt(r.water.Base + 41)
	}
	r.shadowMap, _ = g1.GetPaletteMapAt(2208) // ImageIds::paletteMapUnk32 (tree shadows)
	return r
}

// spriteCache is the decoded sprites, shared by the column workers.
type spriteCache struct {
	mu sync.RWMutex
	m  map[int]*gfx8.Sprite
}

// Sprite decodes (and caches) a g1 sprite; nil if it can't.
func (r *Renderer) Sprite(i int) *gfx8.Sprite {
	c := r.sprites
	c.mu.RLock()
	sp, ok := c.m[i]
	c.mu.RUnlock()
	if ok {
		return sp
	}
	if i >= 0 && i < len(r.G1.Elements) {
		sp, _ = gfx8.Decode(&r.G1.Elements[i])
		if sp != nil {
			sp.Index = i // for the zoom-sprite recursion
		}
	}
	c.mu.Lock()
	c.m[i] = sp
	c.mu.Unlock()
	return sp
}

// SpriteCount is the number of sprites decoded so far.
func (r *Renderer) SpriteCount() int {
	r.sprites.mu.RLock()
	defer r.sprites.mu.RUnlock()
	return len(r.sprites.m)
}

// worker is the i-th column worker: a clone sharing the sprites and the
// read-only data, with its own session and scratch.
func (r *Renderer) worker(i int) *Renderer {
	for len(r.workers) <= i {
		r.workers = append(r.workers, &Renderer{G1: r.G1, Objects: r.Objects, sprites: r.sprites, water: r.water, shadowMap: r.shadowMap})
	}
	return r.workers[i]
}

// View is the part of the viewport to draw (viewport coordinates, rotation 0).
type View struct {
	// X, Y are the viewport position (zoom-0 pixels); W, H the screen size
	// in pixels; Zoom 0-3 (each level halves: the view covers W<<Zoom ×
	// H<<Zoom viewport pixels).
	X, Y, W, H int
	Zoom       int
	// VehiclesMinScale is Config::vehiclesMinScale (0 = its default 2):
	// vehicles are not drawn at a greater zoom.
	VehiclesMinScale int
	Ticks            uint32
	// Tiles optionally restricts the tile walk to a window (minX, minY, maxX,
	// maxY inclusive); empty = every tile that can reach the view.
	Tiles image.Rectangle
}

var slopeToDisplay = [32]int{0, 2, 1, 3, 8, 10, 9, 11, 4, 6, 5, 7, 12, 14, 13, 15, 0, 0, 0, 0, 0, 0, 0, 17, 0, 0, 0, 16, 0, 18, 15, 0}

// tileWindow is the box of tiles whose sprites can reach the view at rotation 0:
// a tile's screen x is (y-x)*32 and its y (x+y)*16 minus its height (up to
// 255*4 px) plus tall objects above it.
func tileWindow(m *sim.Map, v View) (x0, y0, x1, y1 int) {
	const margin = 4
	dMin, dMax := v.X/32-margin, (v.X+v.W)/32+margin           // y - x
	sMin, sMax := v.Y/16-margin, (v.Y+v.H+255*4+256)/16+margin // x + y
	x0, x1 = (sMin-dMax)/2, (sMax-dMin)/2
	y0, y1 = (sMin+dMin)/2, (sMax+dMax)/2
	return max(x0, 0), max(y0, 0), min(x1, m.Width-1), min(y1, m.Height-1)
}

// Render draws the view of m into a new 8-bit target.
func (r *Renderer) Render(m *sim.Map, v View) *gfx8.RenderTarget {
	// Viewport::paint: in columns, each its own session with upstream's tile
	// walk (generate).
	if r.rt == nil || r.rt.Width != v.W || r.rt.Height != v.H {
		r.rt = gfx8.NewRenderTarget(v.W, v.H)
	}
	r.RenderColumns(m, v, r.rt, 0, v.W)
	return r.rt
}

// RenderPart repaints part k of n of the view's columns (screen x from k·W/n
// to (k+1)·W/n) into the frame kept from earlier calls with the same view
// size, and returns that frame: n calls with k = 0…n−1 make a whole frame,
// a frame's work spread over n frames (LOOP item 0).
// It also returns the screen x range it repainted.
func (r *Renderer) RenderPart(m *sim.Map, v View, k, n int) (*gfx8.RenderTarget, int, int) {
	if r.rt == nil || r.rt.Width != v.W || r.rt.Height != v.H {
		r.rt = gfx8.NewRenderTarget(v.W, v.H)
		k, n = 0, 1 // no earlier frame: all of it
	}
	// Parts are whole columns (a part split mid-column would paint both
	// columns it touches): the column edges in screen x, shared out.
	edges := r.edges[:0]
	edges = append(edges, 0)
	lw := v.W << v.Zoom
	for colX := (v.X &^ 31) + 32; colX < v.X+lw; colX += 32 {
		edges = append(edges, (colX-v.X)>>v.Zoom)
	}
	edges = append(edges, v.W)
	r.edges = edges
	cols := len(edges) - 1
	n = min(n, cols)
	k = min(k, n-1)
	x0, x1 := edges[k*cols/n], edges[(k+1)*cols/n]
	r.RenderColumns(m, v, r.rt, x0, x1)
	return r.rt, x0, x1
}

// RenderPlay is RenderPart for play (LOOP item 0): part k of n, plus every
// column a vehicle overlaps, so moving trains are never torn between columns
// painted at different times — as upstream invalidates what moves each tick.
// It returns the frame and the screen x ranges repainted.
func (r *Renderer) RenderPlay(m *sim.Map, v View, k, n int) (*gfx8.RenderTarget, [][2]int) {
	rt, p0, p1 := r.RenderPart(m, v, k, n)
	if p0 == 0 && p1 == v.W {
		r.ranges = append(r.ranges[:0], [2]int{0, v.W})
		return rt, r.ranges
	}
	edges := r.edges
	need := r.need[:0]
	for range len(edges) - 1 {
		need = append(need, false)
	}
	left, right := v.X, v.X+v.W<<v.Zoom
	for i := range m.Entities {
		e := &m.Entities[i]
		if int(e.SpriteRight) <= left || int(e.SpriteLeft) > right || e.SpriteLeft >= e.SpriteRight {
			continue
		}
		a, b := (max(int(e.SpriteLeft), left)-v.X)>>v.Zoom, (min(int(e.SpriteRight), right)-v.X)>>v.Zoom
		for c := 0; c < len(need); c++ {
			if edges[c+1] > a && edges[c] <= b && !(edges[c] >= p0 && edges[c+1] <= p1) {
				need[c] = true
			}
		}
	}
	r.need = need
	ranges := append(r.ranges[:0], [2]int{p0, p1})
	for c := 0; c < len(need); c++ {
		if !need[c] {
			continue
		}
		end := c
		for end+1 < len(need) && need[end+1] {
			end++
		}
		r.RenderColumns(m, v, rt, edges[c], edges[end+1])
		ranges = append(ranges, [2]int{edges[c], edges[end+1]})
		c = end
	}
	r.ranges = ranges
	return rt, ranges
}

// RenderColumns paints the screen columns [sx0, sx1) of the view into dst
// (v.W×v.H) the way Viewport::paint does: in columns 32 viewport pixels wide
// (one tile; 32>>zoom screen pixels), aligned to 32 in viewport coordinates,
// each its own paint session (generate, arrange, draw) on a target cleared
// first. A column paints the same whichever other columns are painted, so
// the game can spread a frame's columns over several frames (LOOP item 0).
// The columns are painted in parallel, as upstream does, one worker a core.
func (r *Renderer) RenderColumns(m *sim.Map, v View, dst *gfx8.RenderTarget, sx0, sx1 int) {
	lw := v.W << v.Zoom
	cols := r.cols[:0]
	for colX := v.X &^ 31; colX < v.X+lw; colX += 32 {
		a, b := (max(colX, v.X)-v.X)>>v.Zoom, (min(colX+32, v.X+lw)-v.X)>>v.Zoom
		if b <= a || b <= sx0 || a >= sx1 {
			continue
		}
		cols = append(cols, [2]int{a, b})
	}
	r.cols = cols
	paintCol := func(w *Renderer, c [2]int) {
		cv := v
		cv.X, cv.W = v.X+c[0]<<v.Zoom, c[1]-c[0]
		crt := w.renderRect(m, cv)
		for y := 0; y < v.H; y++ {
			copy(dst.Bits[y*dst.Width+c[0]:y*dst.Width+c[1]], crt.Bits[y*crt.Width:(y+1)*crt.Width])
		}
	}
	n := min(runtime.GOMAXPROCS(0), len(cols))
	if n <= 1 { // one core (the browser): no goroutines
		for _, c := range cols {
			paintCol(r.worker(0), c)
		}
		return
	}
	var next atomic.Int32
	var wg sync.WaitGroup
	for i := range n {
		w := r.worker(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				k := int(next.Add(1)) - 1
				if k >= len(cols) {
					return
				}
				paintCol(w, cols[k])
			}
		}()
	}
	wg.Wait()
}

// renderRect paints one rectangle of the viewport in one paint session (one
// of Viewport::paint's columns).
func (r *Renderer) renderRect(m *sim.Map, v View) *gfx8.RenderTarget {
	lw, lh := v.W<<v.Zoom, v.H<<v.Zoom // the render target in viewport units (rt.width/height)
	if r.sess == nil {
		r.sess = paint.NewSession(0, v.X, v.Y, lw, lh)
	} else {
		r.sess.Reset(0, v.X, v.Y, lw, lh)
	}
	s := r.sess
	if r.spriteFor == nil {
		r.spriteFor = r.Sprite // a method value allocates: taken once, not per column
	}
	s.Zoom, s.SpriteFor = v.Zoom, r.spriteFor
	var x0, y0, x1, y1 int
	if !v.Tiles.Empty() {
		x0, y0, x1, y1 = max(v.Tiles.Min.X, 0), max(v.Tiles.Min.Y, 0), min(v.Tiles.Max.X, m.Width-1), min(v.Tiles.Max.Y, m.Height-1)
	}
	ents := r.entitiesByTile(m, v, lw, lh)
	if v.Tiles.Empty() {
		r.generate(s, m, v, lw, lh, ents)
	} else { // a tile window asked for (diagnostics): every tile in it, row by row
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				r.paintTile(s, m, x, y, false, v.Ticks)
				r.paintTileEntities(s, m, x, y, ents, false)
			}
		}
	}
	s.Arrange()
	if r.colRT == nil || cap(r.colRT.Bits) < v.W*v.H { // columns differ in width (the first is trimmed): one buffer, resliced
		r.colRT = gfx8.NewRenderTarget(v.W, v.H)
	} else {
		r.colRT.Width, r.colRT.Height, r.colRT.Bits = v.W, v.H, r.colRT.Bits[:v.W*v.H]
	}
	// Viewport::paint clears each column to PaletteIndex::brown2 (black0 in
	// the underground view, which the port doesn't have).
	for i := range r.colRT.Bits {
		r.colRT.Bits[i] = brown2
	}
	rt := r.colRT
	s.Draw(rt)
	return rt
}

// ToRGBAColumns converts only screen columns [x0, x1) of rt into a packed
// RGBA buffer (x1−x0 wide, for WritePixels on that part of the image).
func (r *Renderer) ToRGBAColumns(rt *gfx8.RenderTarget, x0, x1 int) []byte {
	w := x1 - x0
	if n := w * rt.Height * 4; cap(r.partPix) < n {
		r.partPix = make([]byte, n)
	} else {
		r.partPix = r.partPix[:n]
	}
	pix := r.partPix
	for y := 0; y < rt.Height; y++ {
		row := rt.Bits[y*rt.Width+x0 : y*rt.Width+x1]
		o := y * w * 4
		for i, p := range row {
			c := r.G1.Palette[p]
			if p == 0 {
				c = color.RGBA{}
			}
			pix[o+i*4], pix[o+i*4+1], pix[o+i*4+2], pix[o+i*4+3] = c.R, c.G, c.B, 255
		}
	}
	return pix
}

// ToRGBA converts an 8-bit target through the palette (index 0 → opaque black).
func (r *Renderer) ToRGBA(rt *gfx8.RenderTarget) *image.RGBA {
	if r.rgba == nil || r.rgba.Rect.Dx() != rt.Width || r.rgba.Rect.Dy() != rt.Height {
		r.rgba = image.NewRGBA(image.Rect(0, 0, rt.Width, rt.Height))
	}
	img := r.rgba
	for i, p := range rt.Bits {
		c := r.G1.Palette[p]
		if p == 0 {
			c = color.RGBA{}
		}
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = c.R, c.G, c.B, 255
	}
	return img
}

func (r *Renderer) signalObj(i int) (paint.SignalObj, bool) {
	o := r.Objects.GetTrainSignalObjectByIndex(i)
	if o == nil {
		return paint.SignalObj{}, false
	}
	return paint.SignalObj{Image: int(o.ImageOffset), IsLeft: o.IsLeft}, true
}

// tileElemZ is an element's baseZ (the surface's for KindSurface).
func tileElemZ(t *sim.Tile, ref scenario.ElemRef) uint8 {
	switch ref.Kind {
	case scenario.KindTrack:
		return t.Tracks[ref.Index].BaseZ
	case scenario.KindStation:
		return t.Stations[ref.Index].BaseZ
	case scenario.KindSignal:
		return t.Signals[ref.Index].BaseZ
	case scenario.KindBuilding:
		return t.Buildings[ref.Index].BaseZ
	case scenario.KindTree:
		return t.Trees[ref.Index].BaseZ
	case scenario.KindWall:
		return t.Walls[ref.Index].BaseZ
	case scenario.KindRoad:
		return t.Roads[ref.Index].BaseZ
	case scenario.KindIndustry:
		return t.Industries[ref.Index].BaseZ
	}
	return t.BaseZ
}

// paintElement is the switch in paintTileElements.
func (r *Renderer) paintElement(s *paint.Session, m *sim.Map, t *sim.Tile, ref scenario.ElemRef, prev *scenario.ElemRef, x, y int, ticks uint32) {
	om, sprite, v := r.Objects, r.Sprite, struct{ Ticks uint32 }{ticks}
	_ = v
	switch ref.Kind {
	case scenario.KindSurface:
		r.paintSurface(s, m, t, x, y)
	case scenario.KindTrack:
		tk := t.Tracks[ref.Index]
		if to := om.GetTrackObjectByIndex(int(tk.TrackObjectID)); to != nil {
			s.PaintTrack(paint.TrackEl{TrackID: tk.TrackID, SequenceIndex: tk.SequenceIndex, Rotation: tk.Rotation, BaseZ: tk.BaseZ, Remap: r.companyRemap(m, tk.Owner),
				Tunnel: r.tunnelImage(to.TunnelName), HasBridge: tk.HasBridge, Bridge: tk.BridgeID}, int(to.Image), sprite)
			var mods [4]paint.TrackMod // trackObj->mods: the track extras, by the element's hasMod bits
			for i, name := range to.ModNames {
				if ex := om.Extra(name); i < 4 && ex != nil && ex.Image != 0 {
					mods[i] = paint.TrackMod{Has: tk.Mods&(1<<i) != 0, Image: int(ex.Image), PaintStyle: ex.PaintStyle}
				}
			}
			s.PaintTrackMods(paint.TrackEl{TrackID: tk.TrackID, SequenceIndex: tk.SequenceIndex, Rotation: tk.Rotation, BaseZ: tk.BaseZ,
				Remap: r.companyRemap(m, tk.Owner)}, mods, sprite)
		}
	case scenario.KindRoad:
		rd := t.Roads[ref.Index]
		if ro := om.GetRoadObjectByIndex(int(rd.RoadObjectID)); ro != nil {
			var light uint8 // streetLightStyle shares the level crossing id's bits; only without crossing, signal or station
			if !rd.HasLevelCrossing && !rd.HasSignalElement && !rd.HasStationElement {
				light = rd.LevelCrossingObjectID
			}
			s.PaintRoad(paint.RoadEl{RoadID: rd.RoadID, SequenceIndex: rd.SequenceIndex, Rotation: rd.Rotation, BaseZ: rd.BaseZ, Remap: r.companyRemap(m, rd.Owner),
				Tunnel: r.tunnelImage(ro.TunnelName), HasBridge: rd.HasBridge, Bridge: rd.BridgeID,
				Streetlight: light, StreetlightImage: int(om.StreetLightImage), Crossing: r.levelCrossing(rd, v.Ticks),
				Mods: r.roadMods(ro, rd), AnyRoadTypeCompatible: ro.Flags&(1<<3) != 0}, int(ro.Image), ro.PaintStyle, sprite)
		}
	case scenario.KindSignal:
		// paintSignal: against entry.prev(), which must be a track element.
		sg := t.Signals[ref.Index]
		if prev == nil || prev.Kind != scenario.KindTrack {
			return
		}
		tk := t.Tracks[prev.Index]
		s.PaintSignal(sg.Rotation, sg.BaseZ, int(tk.TrackID), int(tk.SequenceIndex), tk.IsLast,
			paint.SignalSide{Has: sg.HasLeftSignal, Object: int(sg.SignalObjectID), Frame: sg.LeftFrame},
			paint.SignalSide{Has: sg.HasRightSignal, Object: int(sg.RightSignalObjectID), Frame: sg.RightFrame},
			r.signalObj, sprite)
	case scenario.KindStation:
		r.paintStation(s, m, t, t.Stations[ref.Index], prev, x, y, ticks)
	case scenario.KindBuilding:
		b := t.Buildings[ref.Index]
		bo := om.GetBuildingObjectByIndex(int(b.ObjectID))
		if bo == nil || int(b.Variation) >= len(bo.VariationParts) {
			return
		}
		an := r.animOf(bo, bo.PartAnimations)
		frames, speeds := an.frames, an.speeds
		cmap := r.colourMap(int(b.Colour))
		s.PaintBuilding(paint.BuildingEl{Rotation: b.Rotation, SequenceIndex: b.SequenceIndex, Variation: b.Variation, BaseZ: b.BaseZ, ClearZ: b.ClearZ,
			IsConstructed: b.IsConstructed, Age: b.Age, SectionProgress: b.SectionProgress},
			paint.BuildingObj{Image: int(bo.ImageOffset), NumParts: int(bo.NumParts), Parts: bo.VariationParts[b.Variation], PartHeights: bo.PartHeights,
				PartAnimFrames: frames, PartAnimSpeed: speeds, ElevatorSequences: bo.ElevatorSequences, LargeTile: bo.HasFlags(objects.BuildingFlagLargeTile),
				Shadow:   r.shadow(s.Zoom, bo.HasFlags(objects.BuildingFlagHasShadows), (int(bo.NumParts)+int(bo.NumElevatorSequences)+int(b.Variation))*4+int(bo.ImageOffset), colourOrange, false),
				Scaffold: r.scaffold(bo.ScaffoldingSegmentType, bo.ScaffoldingColour)},
			cmap, x, y, v.Ticks, sprite)
	case scenario.KindIndustry:
		ie := t.Industries[ref.Index]
		if int(ie.IndustryID) >= len(m.IndustryObjectIDs) {
			return
		}
		io := om.GetIndustryObjectByIndex(int(m.IndustryObjectIDs[ie.IndustryID]))
		if io == nil || int(ie.BuildingType) >= len(io.Variations) {
			return
		}
		an := r.animOf(io, io.PartAnimations)
		frames, speeds := an.frames, an.speeds
		cmap := r.colourMap(int(ie.Colour))
		s.PaintIndustry(paint.IndustryEl{Rotation: ie.Rotation, SequenceIndex: ie.SequenceIndex, BuildingType: ie.BuildingType,
			Var6Low: ie.Var6Low, SectionProgress: ie.SectionProgress, IsConstructed: ie.IsConstructed, BaseZ: ie.BaseZ, ClearZ: ie.ClearZ},
			paint.IndustryObj{BuildingImageIDs: int(io.BuildingImageIDs), Parts: io.Variations[ie.BuildingType], PartHeights: io.PartHeights,
				PartAnimFrames: frames, PartAnimSpeed: speeds, AnimationSequences: io.AnimationSequences,
				LargeTile: io.BuildingSizeFlags&(1<<ie.BuildingType) != 0,
				Shadow:    r.shadow(s.Zoom, io.Flags&objects.IndustryFlagHasShadows != 0, int(ie.BuildingType)*4+int(io.ShadowImageIDs), int(ie.Colour), true),
				Scaffold:  r.scaffold(io.ScaffoldingSegmentType, io.ScaffoldingColour)},
			cmap, x, y, v.Ticks, sprite)
	case scenario.KindWall:
		wl := t.Walls[ref.Index]
		wo := om.GetWallObjectByIndex(int(wl.WallObjectID))
		if wo == nil {
			return
		}
		// getWallImageId: withPrimary / withSecondary as the object has them
		// (an image with a secondary but no primary colour uses primary 0).
		var cmap []uint8
		primary := uint8(0)
		if wo.Flags&objects.WallFlagHasPrimaryColour != 0 {
			primary = wl.PrimaryColour
			cmap = r.colourMap(int(primary))
		}
		if wo.Flags&objects.WallFlagHasSecondaryColour != 0 {
			cmap = r.palette2()(primary, wl.SecondaryColour)
		}
		s.PaintWall(paint.WallEl{Rotation: wl.Rotation, EdgeSlope: wl.EdgeSlope, PrimaryColour: wl.PrimaryColour, BaseZ: wl.BaseZ},
			paint.WallObj{Sprite: int(wo.Sprite), Height: wo.Height, TwoSided: wo.Flags&objects.WallFlagTwoSided != 0,
				HasPrimary: wo.Flags&objects.WallFlagHasPrimaryColour != 0, HasGlass: wo.Flags&objects.WallFlagHasGlass != 0,
				GlassPal: r.colourMap(int(primary) + 0x74)},
			cmap, sprite)
	case scenario.KindTree:
		tr := t.Trees[ref.Index]
		to := om.GetTreeObjectByIndex(int(tr.TreeObjectID))
		if to == nil {
			return
		}
		s.PaintTree(paint.TreeEl{Rotation: tr.Rotation, Quadrant: tr.Quadrant, Growth: tr.Growth, Season: tr.Season,
			Unk7l: tr.Unk7l, HasSnow: tr.HasSnow, BaseZ: tr.BaseZ, ClearZ: tr.ClearZ},
			paint.TreeObj{NumRotations: to.NumRotations, Sprites: to.Sprites, SnowSprites: to.SnowSprites,
				ShadowImageOffset: uint32(to.ShadowImageOffset), HasShadow: to.Flags&objects.TreeFlagHasShadow != 0},
			sprite, r.shadowMap)
	}
}

// paintSurface is PaintSurface.cpp paintSurface for one tile (surface, industry
// field, snow, cliff edges, water).
func (r *Renderer) paintSurface(s *paint.Session, m *sim.Map, t *sim.Tile, x, y int) {
	om, sprite := r.Objects, r.Sprite
	// Session state paintSurface sets (bridge shadows and supports): surface
	// height/slope, setWaterHeight2, and at the end setGeneralSupportHeight.
	// Rotation is 0 here, so rotatedSlope is the slope.
	z := int(t.BaseZ) * 4
	s.SetSurface(z, t.Slope&0x1F, int(t.WaterLevel)*16)
	if t.WaterLevel != 0 {
		s.SetWaterHeight(int(t.WaterLevel) * 16)
	}
	defer s.SetSurfaceSegments(t.Slope & 0x1F)
	defer s.SetGeneralSupportHeight(z, t.Slope&0x1F)
	ti := int(t.TerrainIndex)
	if t.IsIndustrial {
		ti = 0
	}
	land := om.GetLandObjectByIndex(ti)
	if land == nil {
		return
	}
	ds := slopeToDisplay[t.Slope&0x1F]
	growth := int(t.GrowthStage)
	if t.IsIndustrial || growth >= int(land.NumGrowthStages) {
		growth = 0
	}
	idx := int(land.ImageOffset) + land.GetFlatTerrainSpriteIndex() + growth*int(land.NumImagesPerGrowthStage) + ds
	if !t.IsIndustrial && t.Variation != 0 && s.Zoom == 0 && ds == 0 && t.WaterLevel == 0 && land.NumGrowthStages > 0 && t.GrowthStage == land.NumGrowthStages-1 {
		idx = int(land.MapPixelImage) + 3 + int(t.Variation)
	}
	// Industry fields (PaintSurface.cpp industrial branch): the field surface and
	// its tree image come from the industry object; the tile's byte 7 (goloco's
	// Variation) is then the industry id. Snow: below (fieldSnow) unless the
	// object draws its farm tiles above snow (at zoom 0).
	fieldSnow := false
	if t.IsIndustrial && int(t.Variation) < len(m.IndustryObjectIDs) {
		if io := om.GetIndustryObjectByIndex(int(m.IndustryObjectIDs[t.Variation])); io != nil && io.FieldImageIDs != 0 {
			variation := int(io.NumImagesPerFieldGrowthStage)*int(t.GrowthStage) + ((int(io.FarmTileNumImageAngles)-1)&0)*21
			th := z + int(paint.SurfaceTreeHeight[ds])
			if tree := sprite(int(io.FieldImageIDs) + variation + 19); tree != nil {
				s.AddAsParent(paint.Image{Sprite: tree, Op: gfx8.OpFor(tree, false, false)}, paint.Pos3{Z: th}, paint.Pos3{X: 14, Y: 14, Z: th + 4}, paint.Pos3{X: 4, Y: 4, Z: 14})
			}
			idx = int(io.FieldImageIDs) + variation + ds
			fieldSnow = !(s.Zoom == 0 && io.Flags&objects.IndustryFlagFarmTilesDrawAboveSnow != 0) && t.Slope>>5 != 0
		}
	}
	sp := sprite(idx)
	if sp == nil {
		return
	}
	// Snow (PaintSurface.cpp snow branch): full cover replaces the surface
	// image; partial cover attaches the snow image masked per slope — on land
	// tiles, and on industry fields unless farmTilesDrawAboveSnow (zoom 0).
	var snowAttach *paint.Image
	if cov := int(t.Slope >> 5); cov != 0 && (!t.IsIndustrial || fieldSnow) && om.SnowObj != nil {
		snowImg := int(om.SnowObj.ImageOffset) + paint.SnowSurfaceFullZoom + ds
		if cov == 5 {
			if full := sprite(snowImg); full != nil {
				sp = full
			}
		} else if cov <= 4 {
			if si, mk := sprite(snowImg), sprite(int(paint.SnowCoverageSlopeToMask[cov-1][ds])); si != nil && mk != nil {
				snowAttach = &paint.Image{Sprite: si, Mask: mk}
			}
		}
	}
	// PaintSurface.cpp paintMainSurface: offset (0,0,baseHeight), bbox (32,32,-1).
	s.AddAsParent(paint.Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, paint.Pos3{Z: z}, paint.Pos3{Z: z}, paint.Pos3{X: 32, Y: 32, Z: -1})
	if snowAttach != nil && s.Zoom <= 2 { // the partial snow image only up to zoom 2
		s.AttachToPrevious(*snowAttach, 0, 0)
	}
	// Landscape smoothing (zoom 0; Config::landscapeSmoothing, default on).
	if s.Zoom == 0 {
		desc := func(tt *sim.Tile) paint.SmoothTile {
			d := paint.SmoothTile{Surface: paint.Surface{BaseZ: tt.BaseZ, Slope: tt.Slope & 0x1F, Water: tt.WaterLevel}, Present: true,
				Land: tt.TerrainIndex, DisplaySlope: slopeToDisplay[tt.Slope&0x1F], Snow: tt.Slope >> 5, Growth: tt.GrowthStage}
			if tt.IsIndustrial {
				d.Land = 0xFF
			}
			return d
		}
		var nbd [4]paint.SmoothTile
		for d, o := range paint.NeighbourOffsets {
			if nt := m.TileAt(x+o[0]/32, y+o[1]/32); nt != nil {
				nbd[d] = desc(nt)
			}
		}
		snowImage := 0
		if om.SnowObj != nil {
			snowImage = int(om.SnowObj.ImageOffset) + paint.SnowSurfaceFullZoom // SnowObject::image
		}
		s.PaintSurfaceSmoothing(desc(t), nbd, r.smoothLand, snowImage, sprite)
	}

	var nbs [4]*paint.Surface
	for d, o := range paint.NeighbourOffsets {
		if nt := m.TileAt(x+o[0]/32, y+o[1]/32); nt != nil {
			// Scratch kept on the renderer: &Surface{} here escaped, 4 allocations per tile.
			r.nbBuf[d] = paint.Surface{BaseZ: nt.BaseZ, Slope: nt.Slope, Water: nt.WaterLevel}
			nbs[d] = &r.nbBuf[d]
		}
	}
	self := paint.Surface{BaseZ: t.BaseZ, Slope: t.Slope, Water: t.WaterLevel}
	if land.CliffEdgeImage > 0 {
		s.PaintCliffEdges(paint.Surface{BaseZ: t.BaseZ, Slope: t.Slope}, nbs, int(land.CliffEdgeImage), sprite)
	}
	if t.WaterLevel != 0 && r.water.PalMap != nil {
		s.PaintWater(self, nbs, r.water, objects.KSlopeToWaterShape[:], int(land.CliffEdgeImage), sprite)
	}

}

// paintStation is PaintStation.cpp paintStation: by station type; train and road
// stations paint against entry.prev() (a track / road element).
func (r *Renderer) paintStation(s *paint.Session, m *sim.Map, t *sim.Tile, st scenario.StationElement, prev *scenario.ElemRef, x, y int, ticks uint32) {
	om, sprite, v := r.Objects, r.Sprite, struct{ Ticks uint32 }{ticks}
	if st.StationType == 1 { // road station: paintRoadStation, against entry.prev() (a road element)
		rso := om.GetRoadStationObjectByIndex(int(st.ObjectID))
		if rso == nil || rso.PaintStyle != 0 || prev == nil || prev.Kind != scenario.KindRoad {
			return
		}
		rd := t.Roads[prev.Index]
		var remap, tl []uint8 // getCompanyColour(station->owner) and its translucent canopy colour
		if int(st.StationID) < len(m.StationOwners) {
			remap = r.companyRemap(m, m.StationOwners[st.StationID])
			tl = r.canopyTranslucent(m, m.StationOwners[st.StationID], rso.Flags&objects.RoadStationFlagRecolourable != 0)
		}
		// cargoTypeFlags: all; a passenger/freight station drops its cargo type,
		// a passenger-only one then keeps just that type.
		types := uint32(0xFFFFFFFF)
		if rso.Flags&(1<<1|1<<2) != 0 {
			types &^= 1 << (r.cargoSlot(rso.CargoName) & 31)
			if rso.Flags&(1<<2) == 0 {
				types = ^types
			}
		}
		cargo := r.stationCargo(m, rso, rso.CargoOffsets, types, st.StationID, int((s.Rotation+st.Rotation)&3), x+y)
		s.PaintRoadStation(st.Rotation, st.BaseZ, rd.RoadID, int(rso.SequenceImageOffsets[1]), remap, tl, cargo, sprite)
		return
	}
	if st.StationType == 2 { // airport
		ao := om.GetAirportObjectByIndex(int(st.ObjectID))
		if ao == nil || int(st.BuildingType) >= len(ao.Variations) {
			return
		}
		an := r.animOf(ao, ao.PartAnimations)
		s.PaintAirport(st.Rotation, st.SequenceIndex, st.BaseZ, paint.AirportObj{BuildingImage: int(ao.BuildingImage),
			Parts: ao.Variations[st.BuildingType], PartHeights: ao.PartHeights, PartAnimFrames: an.frames, PartAnimSpeed: an.speeds,
			LargeTile: ao.LargeTiles&(1<<st.BuildingType) != 0, Remap: r.companyRemap(m, st.Owner),
			Shadow: r.shadow(s.Zoom, ao.HasShadows(), int(st.BuildingType)*4+int(ao.Image)+1, colourOrange, false)}, x, y, v.Ticks, sprite)
		return
	}
	if st.StationType == 3 { // docks
		do := om.GetDockObjectByIndex(int(st.ObjectID))
		if do == nil || int(st.BuildingType) >= len(do.Variations) {
			return
		}
		an := r.animOf(do, do.PartAnimations)
		s.PaintDocks(st.Rotation, st.SequenceIndex, st.BaseZ, st.ClearZ, paint.AirportObj{BuildingImage: int(do.BuildingImage),
			Parts: do.Variations[st.BuildingType], PartHeights: do.PartHeights, PartAnimFrames: an.frames, PartAnimSpeed: an.speeds, Remap: r.companyRemap(m, st.Owner),
			Shadow: r.shadow(s.Zoom, do.Flags&objects.DockFlagHasShadows != 0, int(st.BuildingType)*4+int(do.Image)+1, colourOrange, false)}, x, y, v.Ticks, sprite)
		return
	}
	if st.StationType != 0 {
		return
	}
	so := om.GetTrainStationObjectByIndex(int(st.ObjectID))
	if so == nil || so.PaintStyle != 0 || int(st.SequenceIndex) >= len(so.SequenceImageOffsets) {
		return
	}
	// paintTrainStation: against entry.prev(), which must be a track element.
	if prev == nil || prev.Kind != scenario.KindTrack {
		return
	}
	tk := t.Tracks[prev.Index]
	s.PaintTrainStation(paint.StationEl{Rotation: st.Rotation, BaseZ: st.BaseZ, TrackID: tk.TrackID, TrackSeq: tk.SequenceIndex, Remap: r.companyRemap(m, tk.Owner),
		Translucent: r.canopyTranslucent(m, tk.Owner, so.Flags&objects.TrainStationFlagRecolourable != 0),
		Cargo:       r.stationCargo(m, so, so.CargoOffsets, 0xFFFFFFFF, st.StationID, int((s.Rotation+st.Rotation)&3), x+y)},
		int(so.SequenceImageOffsets[st.SequenceIndex]), sprite)
}

// entitiesByTile is EntityTileList per tile for paintEntities: the bogie and
// body entities (the only vehicle parts paintVehicleEntity draws) whose
// screen rectangle meets the render target (lw × lh viewport pixels at
// v.X, v.Y); none beyond Config::vehiclesMinScale. ASSUMPTION: within a
// tile, entity slot order (upstream walks its spatial list).
func (r *Renderer) entitiesByTile(m *sim.Map, v View, lw, lh int) map[[2]int][]*scenario.VehicleEntity {
	minScale := v.VehiclesMinScale
	if minScale == 0 {
		minScale = 2 // Config default
	}
	if minScale < v.Zoom || len(m.Entities) == 0 {
		return nil
	}
	if r.entsByTile == nil {
		r.entsByTile = map[[2]int][]*scenario.VehicleEntity{}
	}
	out := r.entsByTile
	for k, l := range out {
		out[k] = l[:0]
	}
	left, top := v.X, v.Y
	right, bottom := left+lw, top+lh
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType != scenario.VehicleTypeBogie && !scenario.IsVehicleBody(e.EntityType) {
			continue
		}
		if int(e.PosX) >= 0x4000 || int(e.PosY) >= 0x4000 || int(e.PosX) < 0 || int(e.PosY) < 0 {
			continue
		}
		if int(e.SpriteTop) > bottom || int(e.SpriteBottom) <= top || int(e.SpriteLeft) > right || int(e.SpriteRight) <= left {
			continue
		}
		k := [2]int{int(e.PosX) / 32, int(e.PosY) / 32}
		out[k] = append(out[k], e)
	}
	return out
}

// paintVehicleEntity is PaintVehicle.cpp paintVehicleEntity: a ghost of
// another company is not drawn; bogies and bodies by their painters at the
// entity's position (setEntityPosition), in their colour schemes.
func (r *Renderer) paintVehicleEntity(s *paint.Session, m *sim.Map, e *scenario.VehicleEntity) {
	const isReversed, isGhost = 1 << 1, 1 << 4
	if e.Flags38&isGhost != 0 && int(e.Owner) != int(m.ControllingCompany) {
		return
	}
	om := r.Objects
	if int(e.ObjectID) >= len(om.Vehicles) || om.Vehicles[e.ObjectID] == nil {
		return
	}
	v := om.Vehicles[e.ObjectID]
	s.SetSpritePosition(int(e.PosX), int(e.PosY))
	pal := r.palette2()
	col := [2]uint8{e.ColourPrimary, e.ColourSecondary}
	if e.EntityType == scenario.VehicleTypeBogie {
		s.PaintBogie(paint.BogieEl{ObjectSpriteType: e.ObjectSpriteType, SpriteYaw: e.SpriteYaw, SpritePitch: e.SpritePitch,
			Reversed: e.Flags38&isReversed != 0, AnimationIndex: e.Animation, Z: int(e.PosZ), Air: v.Mode == objects.TransportModeAir,
			Colour: col, Ghost: e.Flags38&isGhost != 0}, v, r.Sprite, pal, r.colourMap(0x32))
		return
	}
	s.PaintBody(paint.BodyEl{ObjectSpriteType: e.ObjectSpriteType, SpriteYaw: e.SpriteYaw, SpritePitch: e.SpritePitch,
		Reversed: e.Flags38&isReversed != 0, AnimationFrame: e.Animation, CargoFrame: e.CargoFrame, BodyIndex: e.BodyIndex,
		Z: int(e.PosZ), AirOrWater: v.Mode == objects.TransportModeAir || v.Mode == objects.TransportModeWater,
		Colour: col, BrakeLights: r.brakeLights(m, e.HeadID), Ghost: e.Flags38&isGhost != 0}, v, r.Sprite, pal)
}

// brakeLights is the train's veh2 brakeLightTimeout != 0 (head → veh1 → veh2).
func (r *Renderer) brakeLights(m *sim.Map, head uint16) bool {
	for i := range m.Entities {
		if e := &m.Entities[i]; e.EntityType == scenario.VehicleTypeVeh2 && e.HeadID == head {
			return e.BrakeLightTimeout != 0
		}
	}
	return false
}

// palette2 returns the cached GetPaletteMap2 lookup (one closure, maps kept).
func (r *Renderer) palette2() func(p, q uint8) []uint8 {
	if r.pal2Fn == nil {
		r.pal2 = map[[2]uint8][]uint8{}
		r.pal2Fn = func(p, q uint8) []uint8 {
			k := [2]uint8{p, q}
			if pm, ok := r.pal2[k]; ok {
				return pm
			}
			pm, _ := r.G1.GetPaletteMap2(int(p), int(q))
			r.pal2[k] = pm
			return pm
		}
	}
	return r.pal2Fn
}

// smoothLand is what landscape smoothing reads of a land object (by terrain).
func (r *Renderer) smoothLand(id uint8) (paint.SmoothLand, bool) {
	l := r.Objects.GetLandObjectByIndex(int(id))
	if l == nil {
		return paint.SmoothLand{}, false
	}
	return paint.SmoothLand{Image: int(l.ImageOffset) + l.GetFlatTerrainSpriteIndex(), ImagesPerGrowthStage: int(l.NumImagesPerGrowthStage), // LandObject::image
		SharpSlopeTransition:    l.Flags&objects.LandFlagHasSharpSlopeTransition != 0,
		DisableSmoothTransition: l.Flags&objects.LandFlagDisableSmoothTileTransition != 0}, true
}

// levelCrossing is the road's level crossing for paintLevelCrossing (nil:
// none or its object not loaded).
func (r *Renderer) levelCrossing(rd scenario.RoadElement, ticks uint32) *paint.LevelCrossing {
	if !rd.HasLevelCrossing {
		return nil
	}
	lc := r.Objects.GetLevelCrossingObjectByIndex(int(rd.LevelCrossingObjectID))
	if lc == nil {
		return nil
	}
	return &paint.LevelCrossing{Image: int(lc.ImageOffset), AnimationSpeed: lc.AnimSpeed, ClosingFrames: lc.ClosingFrames,
		ClosedFrames: lc.ClosedFrames, Frame: rd.AnimFrame, Ticks: ticks}
}

// roadMods is the road object's mods as paintRoad reads them: the extras (by
// DAT name) with the element's hasMod bits (byte 7 bits 6, 7).
func (r *Renderer) roadMods(ro *objects.RoadObject, rd scenario.RoadElement) [2]paint.TrackMod {
	var mods [2]paint.TrackMod
	for i, name := range ro.ModNames {
		if ex := r.Objects.Extra(name); i < 2 && ex != nil && ex.Image != 0 {
			mods[i] = paint.TrackMod{Has: rd.Mods&(1<<i) != 0, Image: int(ex.Image), PaintStyle: ex.PaintStyle}
		}
	}
	return mods
}

// bridgeHasRoof is BridgeObjectFlags::hasRoof of a bridge slot (paintSupports).
func (r *Renderer) bridgeHasRoof(id uint8) bool {
	bo := r.Objects.GetBridgeObjectByIndex(int(id))
	return bo != nil && bo.Flags&1 != 0
}

// stationCargo is paintStationCargo's input for a station element: the
// station's densities, the cargo objects' platform images, and the station
// object's cargo offsets for (rotation, (x + y) & 3) — kept per key so a
// frame allocates nothing; nil without station data.
func (r *Renderer) stationCargo(m *sim.Map, so any, offsets [4][4][][2][3]int, types uint32, station uint16, rot, hash int) *paint.StationCargo {
	if int(station) >= len(m.StationNames) {
		return nil
	}
	if r.cargoPlatforms == nil {
		r.cargoPlatforms = map[int]*paint.CargoPlatform{}
		r.stationCargos = map[stationCargoKey]*paint.StationCargo{}
	}
	k := stationCargoKey{so, station, rot, hash}
	c, ok := r.stationCargos[k]
	if !ok {
		c = &paint.StationCargo{Offsets: offsets[rot][hash&3], Hash: hash, Types: types}
		for id := range 32 {
			if id >= len(r.Objects.CargoObjects) || r.Objects.CargoObjects[id] == nil {
				continue
			}
			p, ok := r.cargoPlatforms[id]
			if !ok {
				co := r.Objects.CargoObjects[id]
				p = &paint.CargoPlatform{Sprite: int(co.ImageOffset), Variations: co.NumPlatformVariations, Unk0: co.Flags&1 != 0}
				r.cargoPlatforms[id] = p
			}
			c.Cargos[id] = p
		}
		r.stationCargos[k] = c
	}
	c.Density = m.StationCargoDensities(int(station))
	return c
}

type stationCargoKey struct {
	so        any // the train or road station object
	station   uint16
	rot, hash int
}

// cargoSlot is the loaded cargo object's id (slot) for a DAT name (load's
// findObjectHandle); 0 when not found, as cargoType stays 0.
func (r *Renderer) cargoSlot(name string) int {
	if v, ok := r.cargoSlots[name]; ok {
		return v
	}
	if r.cargoSlots == nil {
		r.cargoSlots = map[string]int{}
	}
	v := r.findCargoSlot(name)
	r.cargoSlots[name] = v
	return v
}

func (r *Renderer) findCargoSlot(name string) int {
	for i, c := range r.Objects.CargoObjects {
		if c != nil && c.Header != nil && strings.EqualFold(strings.TrimSpace(c.Header.GetName()), strings.TrimSpace(name)) {
			return i
		}
	}
	return 0
}

// scaffold is the Scaffolding object's input for a building or industry of
// this scaffolding type (0xFF: none) and colour; kept per (type, colour).
func (r *Renderer) scaffold(segType, colour uint8) *paint.Scaffold {
	so := r.Objects.Scaffolding
	if so == nil || segType >= 3 || so.Image == 0 {
		return nil
	}
	k := [2]uint8{segType, colour}
	if sc, ok := r.scaffolds[k]; ok {
		return sc
	}
	if r.scaffolds == nil {
		r.scaffolds = map[[2]uint8]*paint.Scaffold{}
	}
	sc := &paint.Scaffold{Image: int(so.Image), Type: int(segType), SegmentHeight: int(so.SegmentHeights[segType]), Colour: r.colourMap(int(colour))}
	r.scaffolds[k] = sc
	return sc
}

// generate is PaintSession::generate → generateTilesAndEntities with
// generateParameters<0> (rotation 0): from the tile at the target's top-left,
// down one "vertical quadrant" (+32, +32) at a time, (height + 1040) >> 5
// times, painting at each step the tile, then four neighbours in upstream's
// order — (−32, +32) only its tall things (paintTileElements2), (0, +32)
// fully, (+32, 0) tall things, (+32, −32) and (−32, +64) only flying or
// floating vehicles (paintEntities2) — each tile's entities after it.
func (r *Renderer) generate(s *paint.Session, m *sim.Map, v View, lw, lh int, ents map[[2]int][]*scenario.VehicleEntity) {
	n := (lh + 1040) >> 5
	vx, vy := floor2(v.X, 32), floor2(v.Y-16, 32)
	mx, my := floor2(vy-(vx>>1), 32), floor2(vy+(vx>>1), 32) // viewportCoordToMapCoord, rotation 0
	for ; n > 0; n-- {
		tx, ty := mx>>5, my>>5
		r.paintTile(s, m, tx, ty, false, v.Ticks)
		r.paintTileEntities(s, m, tx, ty, ents, false)
		r.paintTile(s, m, tx-1, ty+1, true, v.Ticks)
		r.paintTileEntities(s, m, tx-1, ty+1, ents, false)
		r.paintTile(s, m, tx, ty+1, false, v.Ticks)
		r.paintTileEntities(s, m, tx, ty+1, ents, false)
		r.paintTile(s, m, tx+1, ty, true, v.Ticks)
		r.paintTileEntities(s, m, tx+1, ty, ents, false)
		r.paintTileEntities(s, m, tx+1, ty-1, ents, true)
		r.paintTileEntities(s, m, tx-1, ty+2, ents, true)
		mx, my = mx+32, my+32
	}
}

// brown2 is PaletteIndex::brown2 (Colour.h), the viewport's clear colour.
const brown2 = 0xD8

// floor2 is Numerics::floor2: x rounded down to a multiple of n (a power of 2).
func floor2(x, n int) int { return x &^ (n - 1) }

// paintTile is paintTileElements (tall = false) or paintTileElements2 (tall:
// only buildings, trees, industries, and airport / dock stations — the things
// that reach into the neighbouring column).
// PARTIAL: paintTileElementsSetup's vertical cull (the tile's screen y
// against the target, with the tile's max height) and paintVoid for tiles
// outside the drawable coordinates are not ported: such tiles paint as before.
func (r *Renderer) paintTile(s *paint.Session, m *sim.Map, x, y int, tall bool, ticks uint32) {
	// World::drawableCoords: outside it, paintTileElements paints the void
	// and paintTileElements2 nothing.
	drawable := func(c, n int) bool { return c*32 >= 32 && c*32 < n*32-32-1 }
	if !drawable(x, m.Width) || !drawable(y, m.Height) {
		if !tall {
			r.paintVoid(s, x, y)
		}
		return
	}
	t := m.TileAt(x, y)
	if t == nil {
		return
	}
	sprite := r.spriteFor
	s.SetSpritePosition(x*32, y*32)
	s.ResetTunnels() // paintTileElementsSetup
	s.SetGeneralSupportHeight(0xFFFF, 0)
	s.SetSegmentsSupportHeight(paint.SegmentsAll, 0xFFFF, 0)
	// paintTileElements: the tile's elements in stored order, each followed by
	// paintTileElementsEndLoop. Signals and stations read entry.prev().
	for i, ref := range t.Order {
		if tall && !tallElement(t, ref) {
			continue // paintTileElements2: 'continue' also skips the end loop
		}
		var prev *scenario.ElemRef
		if i > 0 {
			prev = &t.Order[i-1]
		}
		r.paintElement(s, m, t, ref, prev, x, y, ticks)
		// paintTileElementsEndLoop: at the last element or a change of baseZ.
		if i == len(t.Order)-1 || tileElemZ(t, t.Order[i+1]) != tileElemZ(t, ref) {
			s.FinalisePaintRoad(sprite)
			s.PaintSupports(r.bridgeHasRoof, sprite)
			s.FinaliseTrackRoadOrdering()
			s.FinaliseTrackRoadAdditionsOrdering()
			s.SetOccupiedAdditionSupportSegments(0)
			if b, ok := s.TakeBridge(); ok {
				if r.paintBridge(s, b, x, y) {
					s.SetSegmentsSupportHeight(paint.SegmentsAll, 0xFFFF, 0)
				}
				if s.GeneralSupportHeight().Height >= b.Height {
					s.SetGeneralSupportHeight(b.Height, 0x20)
				}
			}
			s.FlushPendingSegments()
		}
	}
}

// paintVoid is PaintTile.cpp's paintVoid (rotation 0): ImageIds::blank_tile
// at z 16 for a location outside the drawable map, when its screen position
// can reach the target.
func (r *Renderer) paintVoid(s *paint.Session, x, y int) {
	const blankTile = 448
	px, py := x*32, y*32
	vy := (px+py)>>1 - 16 // gameToScreen(loc, 16), rotation 0
	_, rtY, _, rtH := s.RTRect()
	if vy+32 <= rtY || vy-20 >= rtH+rtY {
		return
	}
	sp := r.spriteFor(blankTile)
	if sp == nil {
		return
	}
	s.SetSpritePosition(px, py)
	s.AddAsParent(paint.Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, paint.Pos3{Z: 16}, paint.Pos3{Z: 16}, paint.Pos3{X: 32, Y: 32, Z: -1})
}

// tallElement: what paintTileElements2 paints.
func tallElement(t *sim.Tile, ref scenario.ElemRef) bool {
	switch ref.Kind {
	case scenario.KindBuilding, scenario.KindTree, scenario.KindIndustry:
		return true
	case scenario.KindStation:
		st := t.Stations[ref.Index].StationType
		return st == 2 || st == 3 // airport, docks
	}
	return false
}

// paintTileEntities is paintEntities (all) or paintEntities2 (only flying or
// floating vehicles: air and water) for a tile.
func (r *Renderer) paintTileEntities(s *paint.Session, m *sim.Map, x, y int, ents map[[2]int][]*scenario.VehicleEntity, flyingOnly bool) {
	for _, e := range ents[[2]int{x, y}] {
		if flyingOnly && !r.flyingOrFloating(m, e) {
			continue
		}
		r.paintVehicleEntity(s, m, e)
	}
}

// flyingOrFloating is isEntityFlyingOrFloating: a vehicle whose transport
// mode (vhCommonMode) is air or water.
func (r *Renderer) flyingOrFloating(m *sim.Map, e *scenario.VehicleEntity) bool {
	return e.Mode == 2 || e.Mode == 3
}
