package sim

import (
	"strings"
	"testing"
)

// Every live industry on the title map formats: its town's name and then the
// object's name ("Springfield Coal Mines"), or a user string.
func TestIndustryNames(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	n := 0
	for i, ind := range m.Industries {
		if ind.Name == 0xFFFF {
			continue
		}
		name, ok := m.IndustryName(i, om)
		if !ok || name == "" {
			t.Errorf("industry %d (object %d, name %#x): %q ok %v", i, ind.ObjectID, ind.Name, name, ok)
			continue
		}
		if ind.Name < kUserStringsStart {
			town := m.TownName(m.Towns[ind.Town].Name, int(ind.Town))
			if !strings.HasPrefix(name, town+" ") {
				t.Errorf("industry %d %q doesn't start with its town %q", i, name, town)
			}
		}
		if n < 3 {
			t.Logf("industry %d: %q", i, name)
		}
		n++
	}
	if n == 0 {
		t.Skip("no industries")
	}
}

func TestRenameIndustry(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Industries) == 0 {
		t.Skip("no industries")
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	cur, _ := m.IndustryName(0, om)
	if _, ok := m.RenameIndustry(c, CmdApply, 0, cur, om); !ok || m.Industries[0].Name >= kUserStringsStart {
		t.Error("renaming to the same name allocated a string")
	}
	if _, ok := m.RenameIndustry(c, 0, 0, "Joop's Works", om); !ok {
		t.Fatal("the dry run refused")
	}
	if got, _ := m.IndustryName(0, om); got != cur {
		t.Fatal("the dry run renamed it")
	}
	m.RenameIndustry(c, CmdApply, 0, "Joop's Works", om)
	if got, _ := m.IndustryName(0, om); got != "Joop's Works" {
		t.Errorf("name %q", got)
	}
}
