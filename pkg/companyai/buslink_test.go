package companyai

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// loadDataMap loads a save from the local game data with its objects in the
// save's slot order (as pkg/sim's tests); skips without the data.
func loadDataMap(t *testing.T, file string) (*sim.State, *objects.ObjectManager) {
	t.Helper()
	sc, err := scenario.LoadScenarioData("../../../locomotion/Data/" + file)
	if err != nil {
		t.Skip("game data not present")
	}
	om := objects.NewObjectManager("../../../locomotion/ObjData")
	if g1, err := assets.LoadG1("../../../locomotion/Data/g1.DAT"); err == nil {
		om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
		om.G1File = g1
	}
	if err := om.LoadAllObjects(); err != nil {
		t.Skip("objects not present")
	}
	for _, po := range sc.PackedObjectsRaw {
		_, _ = om.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data)
	}
	sim.ReorderObjectsForScenario(om, sc)
	s := &sim.State{}
	s.Map.LoadScenario(sc)
	return s, om
}

// The reference AI end to end on the title save: a new competitor company
// running buslink plans a bus service between two towns joined by road,
// builds both stops, buys, places, orders and starts a bus — every command
// paid through DoCommandFor at a player's price, the loan raised with
// ChangeLoan where the cash ran short (no automatic borrowing) — and the
// bus calls at both stops and earns.
func TestBusLinkEndToEnd(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	m := &s.Map
	id := m.CreateAICompany(BusLinkName, sim.AIOptions{}, om)
	if id == 0xFF {
		t.Fatal("no AI company created")
	}
	t.Logf("company %d %q (owner %q) runs %q; cash %d, loan %d", id, m.CompanyName(int(id)), m.OwnerName(int(id), om),
		m.CompanyAIName(int(id)), m.Companies[id].Cash, m.Companies[id].CurrentLoan)
	if m.CompanyName(int(id)) == "" || m.CompanyName(int(id)) == "New Company" {
		t.Errorf("the competitor company has no name of its own: %q", m.CompanyName(int(id)))
	}
	cash0, loan0 := m.Companies[id].Cash, m.Companies[id].CurrentLoan
	var head uint16 = sim.EntityIDNull
	called := map[uint16]bool{}
	var earned int64
	running := -1
	for tick := 0; tick < 60000; tick++ {
		m.UpdateVehicles(om)
		s.UpdateTiles(func(int, int) bool { return false }, om)
		if head == sim.EntityIDNull {
			for _, d := range m.AIDecisions(int(id)) {
				if d.Command && d.OK && len(d.Text) > 5 && d.Text[:5] == "start" {
					running = tick
				}
			}
			if running >= 0 {
				for _, e := range m.Entities {
					if e.EntityType == scenario.VehicleTypeHead && e.Owner == id {
						head = e.ID
					}
				}
			}
			continue
		}
		e := m.Entities
		for i := range e {
			if e[i].ID == head && (e[i].Status == sim.StatusLoading || e[i].Status == sim.StatusUnloading) && e[i].StationID != 0xFFFF {
				called[e[i].StationID] = true
			}
		}
		tr, err := m.TrainOf(head)
		if err != nil {
			t.Fatal(err)
		}
		earned = max(earned, int64(m.Entities[tr.Veh2].CurMonthRevenue))
		if len(called) >= 2 && earned > 0 {
			break
		}
	}
	for _, d := range m.AIDecisions(int(id)) {
		t.Logf("tick %d: %s (cmd %v ok %v cost %d err %d)", d.Tick, d.Text, d.Command, d.OK, d.Cost, d.Err)
	}
	if running < 0 {
		t.Fatal("the AI never started a bus")
	}
	c := m.Companies[id]
	t.Logf("started at tick %d; called at %v; earned %d; cash %d → %d, loan %d → %d", running, called, earned, cash0, c.Cash, loan0, c.CurrentLoan)
	if len(called) < 2 {
		t.Errorf("the bus did not call at both stops: %v", called)
	}
	if earned <= 0 {
		t.Error("the bus earned nothing")
	}
	var spent int64
	for _, d := range m.AIDecisions(int(id)) {
		if d.Command && d.OK {
			spent += int64(d.Cost)
		}
	}
	if spent <= 0 {
		t.Errorf("the AI paid nothing for its commands (%d)", spent)
	}
}
