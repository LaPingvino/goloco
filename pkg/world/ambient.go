package world

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// AmbientScan is updateAmbientNoise's look around the main view
// (Audio/AmbientAudio.cpp): over the 12 × 12 tiles from 5 before the tile at
// the view's centre (getCentreMapPosition; the range clamped to the map),
// each tile counts once as water (its surface has water) or wilderness (the
// surface the tile's last element and snow-covered, or at baseZ 64 or
// higher); every tree after the surface whose object is not drought
// resistant counts as a tree. Zoom is the main view's.
func (w *World) AmbientScan(om *objects.ObjectManager) (water, wilderness, trees, zoom int) {
	cx, cy := w.centreMapPosition()
	water, wilderness, trees = w.ambientCounts(om, cx, cy)
	return water, wilderness, trees, w.zoom
}

// ambientCounts is AmbientScan's count around the map position (cx, cy).
func (w *World) ambientCounts(om *objects.ObjectManager, cx, cy int32) (water, wilderness, trees int) {
	tx, ty := int(cx/32)-5, int(cy/32)-5 // toTileSpace − {5, 5}
	clamp := func(v, n int) int { return min(max(v, 0), n-1) }
	x0, y0 := clamp(tx, int(w.Map.Width)), clamp(ty, int(w.Map.Height))
	x1, y1 := clamp(tx+11, int(w.Map.Width)), clamp(ty+11, int(w.Map.Height))
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := w.Map.TileAt(x, y)
			if t == nil {
				continue
			}
			passed := false
		elements:
			for i, r := range t.Order {
				switch r.Kind {
				case scenario.KindSurface:
					passed = true
					last := i == len(t.Order)-1
					switch {
					case t.WaterLevel != 0:
						water++
						break elements
					case t.Slope>>5 != 0 && last: // snowCoverage
						wilderness++
						break elements
					case t.BaseZ >= 64 && last:
						wilderness++
						break elements
					}
				case scenario.KindTree:
					if !passed {
						continue
					}
					if obj := treeObject(om, int(t.Trees[r.Index].TreeObjectID)); obj == nil || obj.Flags&objects.TreeFlagDroughtResistant == 0 {
						trees++
					}
				}
			}
		}
	}
	return water, wilderness, trees
}

func treeObject(om *objects.ObjectManager, i int) *objects.TreeObject {
	if om == nil {
		return nil
	}
	return om.GetTreeObjectByIndex(i)
}
