// Package scene renders a view of the map through the ported software renderer
// (pkg/gfx8 + pkg/paint): it walks the visible tiles of a sim.Map, calls the
// ported painters in upstream's per-tile order, sorts with the paint session and
// draws into an 8-bit target. No rendering-library dependencies — used by
// cmd/render8 (headless) and by the game (LOOP item 3j).
package scene

import (
	"image"
	"image/color"

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
	G1        *assets.G1File
	Objects   *objects.ObjectManager
	sprites   map[int]*gfx8.Sprite
	water     paint.WaterImages
	shadowMap []uint8

	// Reused across frames (Render/ToRGBA allocate nothing in steady state).
	sess *paint.Session
	rt   *gfx8.RenderTarget
	rgba *image.RGBA

	colourMaps map[int][]uint8  // GetPaletteMap(colour), which copies, cached
	anims      map[any]anim     // per object: part animation frames and speeds
	tunnels    map[string]int   // tunnel header name → tunnel image base
	nbBuf      [4]paint.Surface // neighbour surfaces of the tile being painted
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
	r := &Renderer{G1: g1, Objects: om, sprites: map[int]*gfx8.Sprite{}}
	if wo := om.WaterObj; wo != nil {
		r.water.Base = int(wo.ImageOffset)
		r.water.PalMap, _ = g1.GetPaletteMapAt(r.water.Base + 41)
	}
	r.shadowMap, _ = g1.GetPaletteMapAt(2208) // ImageIds::paletteMapUnk32 (tree shadows)
	return r
}

// Sprite decodes (and caches) a g1 sprite; nil if it can't.
func (r *Renderer) Sprite(i int) *gfx8.Sprite {
	if sp, ok := r.sprites[i]; ok {
		return sp
	}
	var sp *gfx8.Sprite
	if i >= 0 && i < len(r.G1.Elements) {
		sp, _ = gfx8.Decode(&r.G1.Elements[i])
		if sp != nil {
			sp.Index = i // for the zoom-sprite recursion
		}
	}
	r.sprites[i] = sp
	return sp
}

// SpriteCount is the number of sprites decoded so far.
func (r *Renderer) SpriteCount() int { return len(r.sprites) }

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
	lw, lh := v.W<<v.Zoom, v.H<<v.Zoom // the render target in viewport units (rt.width/height)
	if r.sess == nil {
		r.sess = paint.NewSession(0, v.X, v.Y, lw, lh)
	} else {
		r.sess.Reset(0, v.X, v.Y, lw, lh)
	}
	s := r.sess
	s.Zoom, s.SpriteFor = v.Zoom, r.Sprite
	x0, y0, x1, y1 := tileWindow(m, View{X: v.X, Y: v.Y, W: lw, H: lh})
	if !v.Tiles.Empty() {
		x0, y0, x1, y1 = max(v.Tiles.Min.X, 0), max(v.Tiles.Min.Y, 0), min(v.Tiles.Max.X, m.Width-1), min(v.Tiles.Max.Y, m.Height-1)
	}
	sprite := r.Sprite
	ents := r.entitiesByTile(m, v, lw, lh)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := m.TileAt(x, y)
			if t == nil {
				continue
			}
			s.SetSpritePosition(x*32, y*32)
			s.ResetTunnels() // paintTileElementsSetup
			s.SetGeneralSupportHeight(0xFFFF, 0)
			s.SetSegmentsSupportHeight(paint.SegmentsAll, 0xFFFF, 0)
			// paintTileElements: the tile's elements in stored order, each followed by
			// paintTileElementsEndLoop. Signals and stations read entry.prev().
			for i, ref := range t.Order {
				var prev *scenario.ElemRef
				if i > 0 {
					prev = &t.Order[i-1]
				}
				r.paintElement(s, m, t, ref, prev, x, y, v.Ticks)
				// paintTileElementsEndLoop: at the last element or a change of baseZ.
				// PARTIAL: paintSupports (track/road addition supports) is not ported yet.
				if i == len(t.Order)-1 || tileElemZ(t, t.Order[i+1]) != tileElemZ(t, ref) {
					s.FinalisePaintRoad(sprite)
					s.FinaliseTrackRoadOrdering()
					s.FinaliseTrackRoadAdditionsOrdering()
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
			// paintEntities(loc): the tile's vehicle entities after its elements.
			for _, e := range ents[[2]int{x, y}] {
				r.paintVehicleEntity(s, m, e)
			}
		}
	}
	s.Arrange()
	if r.rt == nil || r.rt.Width != v.W || r.rt.Height != v.H {
		r.rt = gfx8.NewRenderTarget(v.W, v.H)
	} else {
		clear(r.rt.Bits)
	}
	rt := r.rt
	s.Draw(rt)
	return rt
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
		}
	case scenario.KindRoad:
		rd := t.Roads[ref.Index]
		if ro := om.GetRoadObjectByIndex(int(rd.RoadObjectID)); ro != nil {
			s.PaintRoad(paint.RoadEl{RoadID: rd.RoadID, SequenceIndex: rd.SequenceIndex, Rotation: rd.Rotation, BaseZ: rd.BaseZ, Remap: r.companyRemap(m, rd.Owner),
				Tunnel: r.tunnelImage(ro.TunnelName), HasBridge: rd.HasBridge, Bridge: rd.BridgeID}, int(ro.Image), ro.PaintStyle, sprite)
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
		s.PaintBuilding(paint.BuildingEl{Rotation: b.Rotation, SequenceIndex: b.SequenceIndex, Variation: b.Variation, BaseZ: b.BaseZ, ClearZ: b.ClearZ},
			paint.BuildingObj{Image: int(bo.ImageOffset), NumParts: int(bo.NumParts), Parts: bo.VariationParts[b.Variation], PartHeights: bo.PartHeights,
				PartAnimFrames: frames, PartAnimSpeed: speeds, ElevatorSequences: bo.ElevatorSequences, LargeTile: bo.HasFlags(objects.BuildingFlagLargeTile),
				Shadow: r.shadow(s.Zoom, bo.HasFlags(objects.BuildingFlagHasShadows), (int(bo.NumParts)+int(bo.NumElevatorSequences)+int(b.Variation))*4+int(bo.ImageOffset), colourOrange, false)},
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
				Shadow:    r.shadow(s.Zoom, io.Flags&objects.IndustryFlagHasShadows != 0, int(ie.BuildingType)*4+int(io.ShadowImageIDs), int(ie.Colour), true)},
			cmap, x, y, v.Ticks, sprite)
	case scenario.KindWall:
		wl := t.Walls[ref.Index]
		wo := om.GetWallObjectByIndex(int(wl.WallObjectID))
		if wo == nil {
			return
		}
		var cmap []uint8
		if wo.Flags&objects.WallFlagHasPrimaryColour != 0 {
			cmap = r.colourMap(int(wl.PrimaryColour))
		}
		s.PaintWall(paint.WallEl{Rotation: wl.Rotation, EdgeSlope: wl.EdgeSlope, PrimaryColour: wl.PrimaryColour, BaseZ: wl.BaseZ},
			paint.WallObj{Sprite: int(wo.Sprite), Height: wo.Height, TwoSided: wo.Flags&objects.WallFlagTwoSided != 0,
				HasPrimary: wo.Flags&objects.WallFlagHasPrimaryColour != 0, HasGlass: wo.Flags&objects.WallFlagHasGlass != 0},
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
	// Variation) is then the industry id. PARTIAL: snow cover not drawn.
	if t.IsIndustrial && int(t.Variation) < len(m.IndustryObjectIDs) {
		if io := om.GetIndustryObjectByIndex(int(m.IndustryObjectIDs[t.Variation])); io != nil && io.FieldImageIDs != 0 {
			variation := int(io.NumImagesPerFieldGrowthStage)*int(t.GrowthStage) + ((int(io.FarmTileNumImageAngles)-1)&0)*21
			th := z + int(paint.SurfaceTreeHeight[ds])
			if tree := sprite(int(io.FieldImageIDs) + variation + 19); tree != nil {
				s.AddAsParent(paint.Image{Sprite: tree, Op: gfx8.OpFor(tree, false, false)}, paint.Pos3{Z: th}, paint.Pos3{X: 14, Y: 14, Z: th + 4}, paint.Pos3{X: 4, Y: 4, Z: 14})
			}
			idx = int(io.FieldImageIDs) + variation + ds
		}
	}
	sp := sprite(idx)
	if sp == nil {
		return
	}
	// Snow (PaintSurface.cpp snow branch, land tiles): full cover replaces the
	// surface image; partial cover attaches the snow image masked per slope.
	// PARTIAL: industry fields' snow (farmTilesDrawAboveSnow) not ported.
	var snowAttach *paint.Image
	if cov := int(t.Slope >> 5); cov != 0 && !t.IsIndustrial && om.SnowObj != nil {
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
		s.PaintRoadStation(st.Rotation, st.BaseZ, rd.RoadID, int(rso.SequenceImageOffsets[1]), remap, tl, sprite)
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
		Translucent: r.canopyTranslucent(m, tk.Owner, so.Flags&objects.TrainStationFlagRecolourable != 0)},
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
	out := map[[2]int][]*scenario.VehicleEntity{}
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
	pal := func(p, q uint8) []uint8 {
		pm, _ := r.G1.GetPaletteMap2(int(p), int(q))
		return pm
	}
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
