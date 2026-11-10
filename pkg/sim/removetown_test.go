package sim

import "testing"

func TestRemoveTown(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(G1Raw)
	if !ok {
		t.Skip("no g1")
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	// A town with a station is refused.
	for _, st := range m.StationNames {
		if st.Name != stationNameNull {
			if _, ok := m.RemoveTown(c, CmdApply, st.Town, g1, om); ok {
				t.Error("a town with a station was removed")
			}
			break
		}
	}
	i, ok := m.CreateTown(c, CmdApply, -1, -1, 2, g1, om)
	if !ok {
		t.Skip("no site for a new town")
	}
	tw := m.Towns[i]
	id, cx, cy := tw.ID, int(tw.X)/32, int(tw.Y)/32
	near := func() (b, r int) {
		for y := cy - 12; y <= cy+12; y++ {
			for x := cx - 12; x <= cx+12; x++ {
				if tl := m.TileAt(x, y); tl != nil {
					b += len(tl.Buildings)
					r += len(tl.Roads)
				}
			}
		}
		return
	}
	b0, r0 := near()
	n := len(m.Towns)
	if _, ok := m.RemoveTown(c, 0, id, g1, om); !ok || len(m.Towns) != n {
		t.Fatal("the dry run removed the town")
	}
	if _, ok := m.RemoveTown(c, CmdApply, id, g1, om); !ok {
		t.Fatal("removal refused")
	}
	b1, r1 := near()
	t.Logf("near the town: buildings %d → %d, roads %d → %d", b0, b1, r0, r1)
	if len(m.Towns) != n-1 || m.TownByID(id) != nil {
		t.Errorf("towns %d → %d, still found %v", n, len(m.Towns), m.TownByID(id) != nil)
	}
	if b1 >= b0 || r1 >= r0 {
		t.Error("the town's buildings / roads stayed")
	}
	// The other towns' names still format (ids, not indices).
	for j, sn := range m.StationNames {
		if sn.Name != stationNameNull && m.TownByID(sn.Town) == nil {
			t.Errorf("station %d's town %d is gone", j, sn.Town)
		}
	}
}
