package sim

import "testing"

// Trees over a year of the title map (the date and the tile loop): every season comes
// round (Scenario::updateSnowLine, daily) and the forests age, die and spread,
// the total drifting — not collapsing or overrunning the map. (Without the
// daily season the title map sat in winter: 13483 trees fell to 2908 in
// three years.)
func TestTreesOverYears(t *testing.T) {
	if testing.Short() {
		t.Skip("long")
	}
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	count := func() (n, dying, snow int) {
		for y := range s.Tiles {
			for x := range s.Tiles[y] {
				for _, e := range s.Tiles[y][x].Trees {
					n++
					if e.IsDying {
						dying++
					}
					if e.HasSnow {
						snow++
					}
				}
			}
		}
		return
	}
	n0, _, _ := count()
	d0 := s.CurrentDay
	seasons := map[uint8]bool{}
	for i := 0; s.CurrentDay < d0+366 && i < 3_000_000; i++ {
		s.ScenarioTicks++ // tickLogic's date and tile loop: the managers left out
		s.dateTick(om)
		s.UpdateTiles(nil, om)
		seasons[s.CurrentSeason] = true
		if i%(256*16*8) == 0 {
			n, d, sn := count()
			t.Logf("%d-%02d season %d: %d trees, %d dying, %d snow", s.CurrentYear, s.CurrentMonth, s.CurrentSeason, n, d, sn)
		}
	}
	n, _, _ := count()
	if len(seasons) != 4 {
		t.Errorf("seasons seen in a year: %v", seasons)
	}
	if n < n0*9/10 || n > n0*5/4 {
		t.Errorf("trees: %d at the start, %d a year on", n0, n)
	}
}
