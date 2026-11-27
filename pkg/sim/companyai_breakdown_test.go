package sim

import "testing"

// listenerAI hears events and does nothing else.
type listenerAI struct{ heard []AIEvent }

func (a *listenerAI) Name() string                 { return "test-listener" }
func (a *listenerAI) Save() []byte                 { return nil }
func (a *listenerAI) Load([]byte) error            { return nil }
func (a *listenerAI) Think(*AIView)                {}
func (a *listenerAI) OnEvent(_ *AIView, e AIEvent) { a.heard = append(a.heard, e) }

func init() { RegisterAI("test-listener", func() CompanyAI { return &listenerAI{} }) }

// A car breaking down and its breakdown ending reach the owner's AI as
// EventVehicleBrokenDown / EventVehicleRepaired (head id, car's front bogie
// id) — not on the title screen, and not another company's AI.
func TestBreakdownEvents(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	var tr Train
	found := false
	for _, c := range trains {
		h := m.Entities[c.Head]
		if h.Mode == 0 && h.Status == StatusTravelling && m.Entities[c.Veh2].Var73 == 0 && !m.IsPlayerCompany(int(h.Owner)) {
			tr, found = c, true
			break
		}
	}
	if !found {
		t.Skip("no travelling AI-company train")
	}
	owner := int(m.Entities[tr.Head].Owner)
	if taken := m.TakeOverOpenLocoAIs("test-listener", AIOptions{}, om); len(taken) == 0 {
		t.Fatal("no company taken over")
	}
	if m.CompanyAIName(owner) != "test-listener" {
		t.Fatalf("the train's owner %d runs %q", owner, m.CompanyAIName(owner))
	}
	events := func(id int) []AIEvent { return m.companyAIs[id].events }
	others := func() (n int) {
		for id := range m.companyAIs {
			if id != owner && m.companyAIs[id] != nil {
				for _, e := range events(id) {
					if e.Subjects[0] == m.Entities[tr.Head].ID {
						n++
					}
				}
			}
		}
		return
	}
	head, front := &m.Entities[tr.Head], &m.Entities[tr.Cars[0][0].Front]
	front.BreakdownFlags |= BreakdownPending
	m.TitleMode = true
	m.UpdateVehicles(om)
	if len(events(owner)) != 0 {
		t.Fatalf("events on the title screen: %+v", events(owner))
	}
	m.TitleMode = false
	m.UpdateVehicles(om)
	want := AIEvent{Type: EventVehicleBrokenDown, Subjects: [3]uint16{head.ID, front.ID, 0xFFFF}, Day: m.CurrentDay}
	if ev := events(owner); len(ev) != 1 || ev[0] != want {
		t.Fatalf("after the breakdown: %+v, want [%+v]", ev, want)
	}
	front.BreakdownTimeout = 1
	m.vehiclesUpdateDaily(om)
	want.Type = EventVehicleRepaired
	if ev := events(owner); len(ev) != 2 || ev[1] != want || front.BreakdownFlags&BreakdownBrokenDn != 0 {
		t.Fatalf("after the breakdown's last day: %+v (flags %#x), want [.. %+v]", ev, front.BreakdownFlags, want)
	}
	if n := others(); n != 0 {
		t.Errorf("other companies' AIs heard %d events about this train", n)
	}
	// Handed over at the next think, oldest first, then cleared.
	ai := m.companyAIs[owner].ai.(*listenerAI)
	m.ScenarioTicks = m.ScenarioTicks&^0x0F | uint32(owner)
	m.companiesTick(om)
	if len(ai.heard) != 2 || len(events(owner)) != 0 || ai.heard[0].Type != EventVehicleBrokenDown || ai.heard[1].Type != EventVehicleRepaired {
		t.Errorf("heard %+v", ai.heard)
	}
}
