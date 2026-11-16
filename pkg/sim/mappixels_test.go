package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The overall tab's colours on the title map: a bare land tile takes its land
// object's map pixels by height, water the water object's, a building muted
// red (flashing black0 when its key row is hovered), a tree green in the
// second pixel; the map's edge is left alone.
func TestMapPixelsOverall(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(*assets.G1File) // the sprite pool: g1.DAT and the objects' images
	if !ok {
		t.Skip("g1 not present")
	}
	pixel := func(img uint32, i int) uint8 { return g1.Elements[img].Data[i] }
	if _, _, _, _, ok := m.MapPixelsOverall(0, 5, om, pixel, 0); ok {
		t.Error("the map's edge was filled")
	}
	var land, water, building, tree bool
	for ty := 1; ty < m.Height-1; ty++ {
		for tx := 1; tx < m.Width-1; tx++ {
			tl := m.TileAt(tx, ty)
			c0, c1, f0, _, _ := m.MapPixelsOverall(tx, ty, om, pixel, MapFlashBuildings)
			only := len(tl.Order) == 1
			switch {
			case !land && only && tl.WaterLevel == 0:
				lo := om.GetLandObjectByIndex(int(tl.TerrainIndex))
				off := int(tl.BaseZ) / 4 * 2
				if want0, want1 := pixel(lo.MapPixelImage, off), pixel(lo.MapPixelImage, off+1); c0 != want0 || c1 != want1 {
					t.Errorf("land %d,%d: %#x %#x, want %#x %#x", tx, ty, c0, c1, want0, want1)
				}
				land = true
			case !water && only && tl.WaterLevel != 0:
				if c0 == 0 && c1 == 0 {
					t.Errorf("water %d,%d: no colour", tx, ty)
				}
				water = true
			case !building && len(tl.Buildings) == 1 && len(tl.Order) == 2:
				if c0 != palMutedDarkRed7 || c1 != palMutedDarkRed7 || f0 != palBlack0 {
					t.Errorf("building %d,%d: %#x %#x flash %#x", tx, ty, c0, c1, f0)
				}
				building = true
			case !tree && len(tl.Trees) > 0 && len(tl.Order) == 1+len(tl.Trees) && tl.Order[0].Kind == scenario.KindSurface && tl.WaterLevel == 0:
				if c1 != palGreen6 {
					t.Errorf("tree %d,%d: second pixel %#x", tx, ty, c1)
				}
				tree = true
			}
		}
	}
	if !land || !water || !building || !tree {
		t.Errorf("not every kind met: land %v water %v building %v tree %v", land, water, building, tree)
	}
}

// Every row filled, a tile's pair lands where upstream's stepping puts it:
// rotation 0, row tx, ty steps down; rotation 1, row ty, 383−tx steps.
func TestMapBufferLayout(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(*assets.G1File)
	if !ok {
		t.Skip("g1 not present")
	}
	pixel := func(img uint32, i int) uint8 { return g1.Elements[img].Data[i] }
	for rot, at := range []func(tx, ty int) int{
		func(tx, ty int) int { return tx*(MapRenderedWidth-1) + 383 + ty*(MapRenderedWidth+1) },
		func(tx, ty int) int { return ty*(MapRenderedWidth-1) + 383 + (383-tx)*(MapRenderedWidth+1) },
	} {
		b := NewMapBuffer()
		for range mapColumns + 1 {
			m.FillMapRowOverall(b, rot, om, pixel, 0)
		}
		if b.Row != 0 {
			t.Errorf("rotation %d: row %d after a round", rot, b.Row)
		}
		for _, p := range [][2]int{{100, 120}, {200, 50}, {57, 300}} {
			c0, c1, _, _, _ := m.MapPixelsOverall(p[0], p[1], om, pixel, 0)
			if o := at(p[0], p[1]); b.Pixels[o] != c0 || b.Pixels[o+1] != c1 {
				t.Errorf("rotation %d tile %v: %#x %#x, want %#x %#x", rot, p, b.Pixels[o], b.Pixels[o+1], c0, c1)
			}
		}
	}
}
