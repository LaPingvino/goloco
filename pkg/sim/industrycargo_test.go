package sim

import "testing"

// IndustryCargoMonths (the Industry window's Transported tab): an industry
// that needs cargo lists one received line per required cargo before its
// produced lines; every line has its cargo's unit name.
func TestIndustryCargoMonths(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	seenRecv := false
	for i := range m.Industries {
		if m.Industries[i].Name == 0xFFFF {
			continue
		}
		rec, prod, lines := m.IndustryCargoMonths(i, om)
		producedSeen := false
		for _, l := range lines {
			if l.Unit == "" {
				t.Errorf("industry %d: a line without a unit name: %+v", i, l)
			}
			if l.Produced {
				producedSeen = true
			} else if producedSeen {
				t.Errorf("industry %d: a received line after the produced ones", i)
			}
		}
		if rec {
			seenRecv = true
		}
		if !rec && !prod && len(lines) != 0 {
			t.Errorf("industry %d: lines but neither receives nor produces", i)
		}
	}
	if !seenRecv {
		t.Error("no industry on the title map receives cargo")
	}
}
