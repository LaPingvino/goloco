package sim

import "testing"

func TestVehicleRename(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	a := firstRailTrain(t, m, trains, 1)
	var b Train
	for _, tr := range trains {
		if tr.Head != a.Head {
			b = tr
			break
		}
	}
	ha, hb := m.Entities[a.Head].ID, m.Entities[b.Head].ID
	c := &Cmd{Company: m.Entities[a.Head].Owner, Objects: om}
	free := func() int {
		n := 0
		for _, s := range m.UserStrings[:kMaxUserStrings] {
			if s == "" {
				n++
			}
		}
		return n
	}
	if _, ok := m.VehicleRename(c, 0, ha, "Night Express"); !ok { // a dry run allocates nothing
		t.Fatal("dry run refused")
	}
	free0 := free()
	if _, ok := m.VehicleRename(c, CmdApply, ha, "Night Express"); !ok {
		t.Fatal("rename refused")
	}
	h := m.Entities[a.Head]
	if got := m.VehicleName(h.Name, int(h.OrdinalNumber)); got != "Night Express" || free() != free0-1 {
		t.Fatalf("named %q, free user strings %d → %d", got, free0, free())
	}
	if _, ok := m.VehicleRename(c, CmdApply, hb, "Night Express"); ok {
		t.Error("a second vehicle took the same name")
	}
	if _, ok := m.VehicleRename(c, CmdApply, ha, ""); !ok {
		t.Fatal("clearing the name refused")
	}
	h = m.Entities[a.Head]
	if h.Name != defaultVehicleNames[h.VehicleType] || free() != free0 {
		t.Errorf("cleared: name id %d, free user strings %d (want %d)", h.Name, free(), free0)
	}
	t.Logf("renamed and back: %q", m.VehicleName(h.Name, int(h.OrdinalNumber)))
}
