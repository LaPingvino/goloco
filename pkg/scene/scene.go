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
	}
	r.sprites[i] = sp
	return sp
}

// SpriteCount is the number of sprites decoded so far.
func (r *Renderer) SpriteCount() int { return len(r.sprites) }

// View is the part of the viewport to draw (viewport coordinates, rotation 0).
type View struct {
	X, Y, W, H int
	Ticks      uint32
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
	s := paint.NewSession(0, v.X, v.Y, v.W, v.H)
	x0, y0, x1, y1 := tileWindow(m, v)
	if !v.Tiles.Empty() {
		x0, y0, x1, y1 = max(v.Tiles.Min.X, 0), max(v.Tiles.Min.Y, 0), min(v.Tiles.Max.X, m.Width-1), min(v.Tiles.Max.Y, m.Height-1)
	}
	om := r.Objects
	sprite := r.Sprite
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := m.TileAt(x, y)
			if t == nil {
				continue
			}
			ti := int(t.TerrainIndex)
			if t.IsIndustrial {
				ti = 0
			}
			land := om.GetLandObjectByIndex(ti)
			if land == nil {
				continue
			}
			ds := slopeToDisplay[t.Slope&0x1F]
			growth := int(t.GrowthStage)
			if t.IsIndustrial || growth >= int(land.NumGrowthStages) {
				growth = 0
			}
			idx := int(land.ImageOffset) + land.GetFlatTerrainSpriteIndex() + growth*int(land.NumImagesPerGrowthStage) + ds
			if !t.IsIndustrial && t.Variation != 0 && ds == 0 && t.WaterLevel == 0 && land.NumGrowthStages > 0 && t.GrowthStage == land.NumGrowthStages-1 {
				idx = int(land.MapPixelImage) + 3 + int(t.Variation)
			}
			z := int(t.BaseZ) * 4
			s.SetSpritePosition(x*32, y*32)
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
				continue
			}
			// PaintSurface.cpp paintMainSurface: offset (0,0,baseHeight), bbox (32,32,-1).
			s.AddAsParent(paint.Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, paint.Pos3{Z: z}, paint.Pos3{Z: z}, paint.Pos3{X: 32, Y: 32, Z: -1})

			var nbs [4]*paint.Surface
			for d, o := range paint.NeighbourOffsets {
				if nt := m.TileAt(x+o[0]/32, y+o[1]/32); nt != nil {
					nbs[d] = &paint.Surface{BaseZ: nt.BaseZ, Slope: nt.Slope, Water: nt.WaterLevel}
				}
			}
			self := paint.Surface{BaseZ: t.BaseZ, Slope: t.Slope, Water: t.WaterLevel}
			if land.CliffEdgeImage > 0 {
				s.PaintCliffEdges(paint.Surface{BaseZ: t.BaseZ, Slope: t.Slope}, nbs, int(land.CliffEdgeImage), sprite)
			}
			if t.WaterLevel != 0 && r.water.PalMap != nil {
				s.PaintWater(self, nbs, r.water, objects.KSlopeToWaterShape[:], int(land.CliffEdgeImage), sprite)
			}

			// Tracks and roads, then the per-group finalise passes (PaintTile.cpp
			// paintTileElementsEndLoop: at the end of each run of elements with the
			// same baseZ). Tile elements are stored in ascending baseZ; the order of
			// tracks vs roads within one baseZ is not kept by the tile (separate
			// slices), so tracks go first — APPROXIMATION until the tile keeps the
			// element order.
			ti, ri := 0, 0
			for ti < len(t.Tracks) || ri < len(t.Roads) {
				z := 256
				if ti < len(t.Tracks) {
					z = int(t.Tracks[ti].BaseZ)
				}
				if ri < len(t.Roads) && int(t.Roads[ri].BaseZ) < z {
					z = int(t.Roads[ri].BaseZ)
				}
				for ; ti < len(t.Tracks) && int(t.Tracks[ti].BaseZ) == z; ti++ {
					tk := t.Tracks[ti]
					if to := om.GetTrackObjectByIndex(int(tk.TrackObjectID)); to != nil {
						s.PaintTrack(paint.TrackEl{TrackID: tk.TrackID, SequenceIndex: tk.SequenceIndex, Rotation: tk.Rotation, BaseZ: tk.BaseZ}, int(to.Image), sprite)
					}
				}
				for ; ri < len(t.Roads) && int(t.Roads[ri].BaseZ) == z; ri++ {
					rd := t.Roads[ri]
					if ro := om.GetRoadObjectByIndex(int(rd.RoadObjectID)); ro != nil {
						s.PaintRoad(paint.RoadEl{RoadID: rd.RoadID, SequenceIndex: rd.SequenceIndex, Rotation: rd.Rotation, BaseZ: rd.BaseZ}, int(ro.Image), ro.PaintStyle, sprite)
					}
				}
				s.FinalisePaintRoad(sprite)
				s.FinaliseTrackRoadOrdering()
				s.FinaliseTrackRoadAdditionsOrdering()
			}
			// Signals: upstream paints a signal against the track element just before it
			// in the tile; the tile keeps signals in their own slice, so the track with
			// the same baseZ stands in for it (APPROXIMATION until element order is kept).
			for _, sg := range t.Signals {
				for _, tk := range t.Tracks {
					if tk.BaseZ != sg.BaseZ {
						continue
					}
					s.PaintSignal(sg.Rotation, sg.BaseZ, int(tk.TrackID), int(tk.SequenceIndex), tk.IsLast,
						paint.SignalSide{Has: sg.HasLeftSignal, Object: int(sg.SignalObjectID), Frame: sg.LeftFrame},
						paint.SignalSide{Has: sg.HasRightSignal, Object: int(sg.RightSignalObjectID), Frame: sg.RightFrame},
						r.signalObj, sprite)
					break
				}
			}
			// Train stations (style 0), against the track of the same baseZ (upstream:
			// the element before — same APPROXIMATION as signals).
			for _, st := range t.Stations {
				if st.StationType != 0 {
					continue // road stations, airports, docks: not ported yet
				}
				so := om.GetTrainStationObjectByIndex(int(st.ObjectID))
				if so == nil || so.PaintStyle != 0 || int(st.SequenceIndex) >= len(so.SequenceImageOffsets) {
					continue
				}
				for _, tk := range t.Tracks {
					if tk.BaseZ == st.BaseZ {
						s.PaintTrainStation(paint.StationEl{Rotation: st.Rotation, BaseZ: st.BaseZ, TrackID: tk.TrackID, TrackSeq: tk.SequenceIndex},
							int(so.SequenceImageOffsets[st.SequenceIndex]), sprite)
						break
					}
				}
			}
			s.FinaliseTrackRoadAdditionsOrdering()
			for _, b := range t.Buildings {
				bo := om.GetBuildingObjectByIndex(int(b.ObjectID))
				if bo == nil || int(b.Variation) >= len(bo.VariationParts) {
					continue
				}
				frames := make([]uint8, len(bo.PartAnimations))
				speeds := make([]uint8, len(bo.PartAnimations))
				for i, pa := range bo.PartAnimations {
					frames[i], speeds[i] = pa.NumFrames, pa.AnimationSpeed
				}
				cmap, _ := r.G1.GetPaletteMap(int(b.Colour))
				s.PaintBuilding(paint.BuildingEl{Rotation: b.Rotation, SequenceIndex: b.SequenceIndex, Variation: b.Variation, BaseZ: b.BaseZ, ClearZ: b.ClearZ},
					paint.BuildingObj{Image: int(bo.ImageOffset), NumParts: int(bo.NumParts), Parts: bo.VariationParts[b.Variation], PartHeights: bo.PartHeights,
						PartAnimFrames: frames, PartAnimSpeed: speeds, ElevatorSequences: bo.ElevatorSequences, LargeTile: bo.HasFlags(objects.BuildingFlagLargeTile)},
					cmap, x, y, v.Ticks, sprite)
			}
			for _, ie := range t.Industries {
				if int(ie.IndustryID) >= len(m.IndustryObjectIDs) {
					continue
				}
				io := om.GetIndustryObjectByIndex(int(m.IndustryObjectIDs[ie.IndustryID]))
				if io == nil || int(ie.BuildingType) >= len(io.Variations) {
					continue
				}
				frames := make([]uint8, len(io.PartAnimations))
				speeds := make([]uint8, len(io.PartAnimations))
				for i, pa := range io.PartAnimations {
					frames[i], speeds[i] = pa.NumFrames, pa.AnimationSpeed
				}
				cmap, _ := r.G1.GetPaletteMap(int(ie.Colour))
				s.PaintIndustry(paint.IndustryEl{Rotation: ie.Rotation, SequenceIndex: ie.SequenceIndex, BuildingType: ie.BuildingType,
					Var6Low: ie.Var6Low, SectionProgress: ie.SectionProgress, IsConstructed: ie.IsConstructed, BaseZ: ie.BaseZ, ClearZ: ie.ClearZ},
					paint.IndustryObj{BuildingImageIDs: int(io.BuildingImageIDs), Parts: io.Variations[ie.BuildingType], PartHeights: io.PartHeights,
						PartAnimFrames: frames, PartAnimSpeed: speeds, AnimationSequences: io.AnimationSequences,
						LargeTile: io.BuildingSizeFlags&(1<<ie.BuildingType) != 0},
					cmap, x, y, v.Ticks, sprite)
			}
			for _, wl := range t.Walls {
				wo := om.GetWallObjectByIndex(int(wl.WallObjectID))
				if wo == nil {
					continue
				}
				var cmap []uint8
				if wo.Flags&objects.WallFlagHasPrimaryColour != 0 {
					cmap, _ = r.G1.GetPaletteMap(int(wl.PrimaryColour))
				}
				s.PaintWall(paint.WallEl{Rotation: wl.Rotation, EdgeSlope: wl.EdgeSlope, PrimaryColour: wl.PrimaryColour, BaseZ: wl.BaseZ},
					paint.WallObj{Sprite: int(wo.Sprite), Height: wo.Height, TwoSided: wo.Flags&objects.WallFlagTwoSided != 0,
						HasPrimary: wo.Flags&objects.WallFlagHasPrimaryColour != 0, HasGlass: wo.Flags&objects.WallFlagHasGlass != 0},
					cmap, sprite)
			}
			for _, tr := range t.Trees {
				to := om.GetTreeObjectByIndex(int(tr.TreeObjectID))
				if to == nil {
					continue
				}
				s.PaintTree(paint.TreeEl{Rotation: tr.Rotation, Quadrant: tr.Quadrant, Growth: tr.Growth, Season: tr.Season,
					Unk7l: tr.Unk7l, HasSnow: tr.HasSnow, BaseZ: tr.BaseZ, ClearZ: tr.ClearZ},
					paint.TreeObj{NumRotations: to.NumRotations, Sprites: to.Sprites, SnowSprites: to.SnowSprites,
						ShadowImageOffset: uint32(to.ShadowImageOffset), HasShadow: to.Flags&objects.TreeFlagHasShadow != 0},
					sprite, r.shadowMap)
			}
		}
	}
	s.Arrange()
	rt := gfx8.NewRenderTarget(v.W, v.H)
	s.Draw(rt)
	return rt
}

// ToRGBA converts an 8-bit target through the palette (index 0 → opaque black).
func (r *Renderer) ToRGBA(rt *gfx8.RenderTarget) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, rt.Width, rt.Height))
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
