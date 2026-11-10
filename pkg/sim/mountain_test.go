package sim

import "testing"

// Raising a one-tile mountain three times on flat land: the centre is three
// steps up and the land round it slopes (more tiles than the centre changed,
// each side neighbour's highest corner above the old ground); the dry run
// changes nothing; lowered three times the centre is back.
func TestLowerRaiseLandMountain(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om, Editor: true}
	// A flat, empty, dry 9×9 area.
	cx, cy := -1, -1
	for y := 12; y < m.Height-12 && cx < 0; y++ {
		for x := 12; x < m.Width-12; x++ {
			ok, z := true, m.TileAt(x, y).BaseZ
			for dy := -4; dy <= 4 && ok; dy++ {
				for dx := -4; dx <= 4; dx++ {
					tl := m.TileAt(x+dx, y+dy)
					if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 || tl.BaseZ != z {
						ok = false
						break
					}
				}
			}
			if ok {
				cx, cy = x, y
				break
			}
		}
	}
	if cx < 0 {
		t.Skip("no flat 9×9")
	}
	z0 := m.TileAt(cx, cy).BaseZ
	a := MountainArgs{CX: int32(cx)*32 + 16, CY: int32(cy)*32 + 16, AX: int32(cx) * 32, AY: int32(cy) * 32, BX: int32(cx) * 32, BY: int32(cy) * 32, Adjustment: 1}
	if _, ok := m.LowerRaiseLandMountain(c, 0, a); !ok || m.TileAt(cx, cy).BaseZ != z0 {
		t.Fatal("the dry run changed the land")
	}
	var cost int32
	for range 3 {
		got, ok := m.LowerRaiseLandMountain(c, CmdApply, a)
		if !ok {
			t.Fatal("refused")
		}
		cost += got
	}
	if got := m.TileAt(cx, cy).BaseZ; got != z0+12 {
		t.Errorf("centre at %d, want %d", got, z0+12)
	}
	changed := 0
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			tl := m.TileAt(cx+dx, cy+dy)
			if tl.BaseZ != z0 || tl.Slope != 0 {
				changed++
			}
		}
	}
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		if tl := m.TileAt(cx+d[0], cy+d[1]); surfaceCornerHeight(tl) <= z0 {
			t.Errorf("neighbour %v left flat at %d", d, tl.BaseZ)
		}
	}
	t.Logf("cost %d, %d of the 81 tiles changed", cost, changed)
	if changed <= 1 || cost <= 0 {
		t.Errorf("only %d tiles changed, cost %d", changed, cost)
	}
	// Lowered three times, the centre is back.
	a.Adjustment = 0
	for range 3 {
		m.LowerRaiseLandMountain(c, CmdApply, a)
	}
	if got := m.TileAt(cx, cy).BaseZ; got != z0 {
		t.Errorf("after lowering the centre is at %d, want %d", got, z0)
	}
}
