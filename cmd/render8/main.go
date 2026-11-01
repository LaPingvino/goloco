// Command render8 renders a map region through the ported software renderer
// (pkg/gfx8 + pkg/paint) to a PNG, with no graphics library — the CPU
// reference the Ebiten renderer is compared against (LOOP item 3i, step 3).
// So far: land surfaces, cliff edges, water (no waves), trees, buildings and roads.
//
//	go run ./cmd/render8 -scenario ../locomotion/Data/title.dat -tile 196,190 -out surf.png
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

func main() {
	data := flag.String("data", "../locomotion", "Locomotion data directory")
	scen := flag.String("scenario", "../locomotion/Data/title.dat", "scenario or save to render")
	tile := flag.String("tile", "196,190", "centre tile x,y")
	w := flag.Int("w", 480, "width")
	h := flag.Int("h", 360, "height")
	out := flag.String("out", "render8.png", "output PNG")
	flag.Parse()

	var tx, ty int
	if _, err := fmt.Sscanf(*tile, "%d,%d", &tx, &ty); err != nil {
		log.Fatalf("render8: -tile: %v", err)
	}
	g1, err := assets.LoadG1(filepath.Join(*data, "Data", "g1.DAT"))
	must(err)
	om := objects.NewObjectManager(filepath.Join(*data, "ObjData"))
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	must(om.LoadAllObjects())
	sc, err := scenario.LoadScenarioData(*scen)
	must(err)
	for _, po := range sc.PackedObjectsRaw {
		_, _ = om.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data)
	}
	if len(sc.LandObjectOrder) > 0 {
		om.ReorderLandObjects(sc.LandObjectOrder)
	}
	if len(sc.TreeObjectOrder) > 0 {
		om.ReorderTreeObjects(sc.TreeObjectOrder)
	}
	if len(sc.BuildingObjectOrder) > 0 {
		om.ReorderBuildingObjects(sc.BuildingObjectOrder)
	}
	if len(sc.RoadObjectOrder) > 0 {
		om.ReorderRoadObjects(sc.RoadObjectOrder)
	}
	om.ResolveCliffEdges()
	shadowMap, _ := g1.GetPaletteMapAt(2208) // ImageIds::paletteMapUnk32 (tree shadows)
	var water paint.WaterImages
	if wo := om.WaterObj; wo != nil {
		water.Base = int(wo.ImageOffset)
		water.PalMap, _ = g1.GetPaletteMapAt(water.Base + 41)
	}

	ct := sc.GetTile(tx, ty)
	cz := 0
	if ct != nil {
		cz = int(ct.Height) * 4
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx*32 + 16, Y: ty*32 + 16, Z: cz}, 0)
	s := paint.NewSession(0, cx-*w/2, cy-*h/2, *w, *h)

	sprites := map[int]*gfx8.Sprite{}
	sprite := func(i int) *gfx8.Sprite {
		if sp, ok := sprites[i]; ok {
			return sp
		}
		var sp *gfx8.Sprite
		if i >= 0 && i < len(g1.Elements) {
			sp, _ = gfx8.Decode(&g1.Elements[i])
		}
		sprites[i] = sp
		return sp
	}
	slopeToDisplay := [32]int{0, 2, 1, 3, 8, 10, 9, 11, 4, 6, 5, 7, 12, 14, 13, 15, 0, 0, 0, 0, 0, 0, 0, 17, 0, 0, 0, 16, 0, 18, 15, 0}

	const r = 24
	for y := ty - r; y <= ty+r; y++ {
		for x := tx - r; x <= tx+r; x++ {
			t := sc.GetTile(x, y)
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
			// Same choice as goloco's surface painter (world.go), for an equal comparison.
			idx := int(land.ImageOffset) + land.GetFlatTerrainSpriteIndex() + growth*int(land.NumImagesPerGrowthStage) + ds
			if !t.IsIndustrial && t.Variation != 0 && ds == 0 && t.Water == 0 && land.NumGrowthStages > 0 && t.GrowthStage == land.NumGrowthStages-1 {
				idx = int(land.MapPixelImage) + 3 + int(t.Variation)
			}
			sp := sprite(idx)
			if sp == nil {
				continue
			}
			z := int(t.Height) * 4
			s.SetSpritePosition(x*32, y*32)
			// PaintSurface.cpp paintMainSurface: offset (0,0,baseHeight), bbox (32,32,-1).
			s.AddAsParent(paint.Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, paint.Pos3{Z: z}, paint.Pos3{Z: z}, paint.Pos3{X: 32, Y: 32, Z: -1})

			// Cliff faces against the four neighbours (PaintSurface.cpp, edges 2,3,0,1).
			if land.CliffEdgeImage > 0 {
				var nbs [4]*paint.Surface
				for d, o := range paint.NeighbourOffsets {
					if nt := sc.GetTile(x+o[0]/32, y+o[1]/32); nt != nil {
						nbs[d] = &paint.Surface{BaseZ: nt.Height, Slope: nt.Slope}
					}
				}
				s.PaintCliffEdges(paint.Surface{BaseZ: t.Height, Slope: t.Slope}, nbs, int(land.CliffEdgeImage), sprite)
			}
			// Roads, then the per-group finalise passes (PaintTile.cpp paintTileElementsEndLoop:
			// at the end of each run of elements with the same baseZ).
			for i, rd := range t.Roads {
				if ro := om.GetRoadObjectByIndex(int(rd.RoadObjectID)); ro != nil {
					s.PaintRoad(paint.RoadEl{RoadID: rd.RoadID, SequenceIndex: rd.SequenceIndex, Rotation: rd.Rotation, BaseZ: rd.BaseZ}, int(ro.Image), ro.PaintStyle, sprite)
				}
				if i == len(t.Roads)-1 || t.Roads[i+1].BaseZ != rd.BaseZ {
					s.FinalisePaintRoad(sprite)
					s.FinaliseTrackRoadOrdering()
				}
			}
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
				cmap, _ := g1.GetPaletteMap(int(b.Colour))
				s.PaintBuilding(paint.BuildingEl{Rotation: b.Rotation, SequenceIndex: b.SequenceIndex, Variation: b.Variation, BaseZ: b.BaseZ, ClearZ: b.ClearZ},
					paint.BuildingObj{Image: int(bo.ImageOffset), NumParts: int(bo.NumParts), Parts: bo.VariationParts[b.Variation], PartHeights: bo.PartHeights,
						PartAnimFrames: frames, PartAnimSpeed: speeds, ElevatorSequences: bo.ElevatorSequences, LargeTile: bo.HasFlags(objects.BuildingFlagLargeTile)},
					cmap, x, y, 0, sprite)
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
					sprite, shadowMap)
			}
			if t.Water != 0 && water.PalMap != nil {
				var nbs [4]*paint.Surface
				for d, o := range paint.NeighbourOffsets {
					if nt := sc.GetTile(x+o[0]/32, y+o[1]/32); nt != nil {
						nbs[d] = &paint.Surface{BaseZ: nt.Height, Slope: nt.Slope, Water: nt.Water}
					}
				}
				s.PaintWater(paint.Surface{BaseZ: t.Height, Slope: t.Slope, Water: t.Water}, nbs, water, objects.KSlopeToWaterShape[:], int(land.CliffEdgeImage), sprite)
			}
		}
	}
	s.Arrange()
	rt := gfx8.NewRenderTarget(*w, *h)
	s.Draw(rt)

	img := image.NewRGBA(image.Rect(0, 0, *w, *h))
	for i, p := range rt.Bits {
		c := g1.Palette[p]
		if p == 0 {
			c = color.RGBA{0, 0, 0, 255}
		}
		c.A = 255
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = c.R, c.G, c.B, 255
	}
	f, err := os.Create(*out)
	must(err)
	must(png.Encode(f, img))
	must(f.Close())
	log.Printf("render8: %d sprites decoded → %s", len(sprites), *out)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
