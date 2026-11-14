package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A planted sapling grows (updateTreeElement): each 16th update (unk5h
// wrapping), in its season, a step of growth and its clear height to the
// object's curve, until the object's last growth stage.
func TestSaplingGrows(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	// a tile of plain, flat land with no trees: plant a sapling of a tree object
	var tx, ty int
	var obj uint8 = 0xFF
	for i, o := range om.TreeObjects {
		if o != nil && o.Growth > 2 {
			obj = uint8(i)
			break
		}
	}
	if obj == 0xFF {
		t.Skip("no growing tree object")
	}
	planted := false
	for ty = 50; ty < 300 && !planted; ty += 7 {
		for tx = 50; tx < 300 && !planted; tx += 7 {
			tl := m.TileAt(tx, ty)
			if tl == nil || tl.WaterLevel != 0 || tl.Slope&0x1F != 0 || len(tl.Order) != 1 || int(tl.BaseZ)-4 > int(m.CurrentSnowLine) {
				continue
			}
			a := TreePlacement{X: int32(tx) * 32, Y: int32(ty) * 32, Type: obj, RequiresFullClearance: true}
			if _, _, ok := m.CreateTreeCmd(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, a); ok {
				planted = true
			}
		}
	}
	if !planted {
		t.Fatal("found no tile to plant on")
	}
	tx, ty = tx-7, ty-7
	tl := m.TileAt(tx, ty)
	k := -1
	for i, r := range tl.Order {
		if r.Kind == scenario.KindTree {
			k = i
		}
	}
	if k < 0 {
		t.Fatal("the sapling isn't on the tile")
	}
	e := &tl.Trees[tl.Order[k].Index]
	e.Season, e.Unk7l = m.CurrentSeason, 7 // its season settled
	g0, c0 := e.Growth, e.ClearZ
	for i := 0; i < 16; i++ {
		if !m.updateTreeElement(tl, k, int32(tx)*32, int32(ty)*32, om) {
			t.Fatal("the tile changed")
		}
	}
	e = &tl.Trees[tl.Order[k].Index]
	if e.IsDying {
		t.Fatal("the sapling is dying")
	}
	if e.Growth != g0+1 || e.ClearZ <= c0 {
		t.Errorf("after 16 updates: growth %d (was %d), clearZ %d (was %d)", e.Growth, g0, e.ClearZ, c0)
	}
}
