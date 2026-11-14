package sim

import "testing"

// IndustryCargoMonths (the Industry window's Transported tab): an industry
// that needs cargo lists its received lines, then what awaits processing,
// before its produced lines and what awaits transport — as many awaiting
// lines as received / produced ones; every line has its cargo's unit name.
func TestIndustryCargoMonths(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	seenRecv := false
	for i := range m.Industries {
		if m.Industries[i].Name == 0xFFFF {
			continue
		}
		rec, prod, lines := m.IndustryCargoMonths(i, om)
		last, count := -1, [4]int{}
		for _, l := range lines {
			if l.Unit == "" {
				t.Errorf("industry %d: a line without a unit name: %+v", i, l)
			}
			order := [4]int{CargoReceived: 0, CargoAwaitingProcessing: 1, CargoProduced: 2, CargoAwaitingTransport: 3}[l.Kind]
			if order < last {
				t.Errorf("industry %d: kind %d after a later one", i, l.Kind)
			}
			last = order
			count[l.Kind]++
		}
		if count[CargoReceived] != count[CargoAwaitingProcessing] || count[CargoProduced] != count[CargoAwaitingTransport] {
			t.Errorf("industry %d: line counts %v", i, count)
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
