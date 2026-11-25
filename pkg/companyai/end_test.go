package companyai

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/sim"
)

// quitter ends its company at its first think; it hears the news.
type quitter struct{ heard []sim.AIEvent }

func (q *quitter) Name() string { return "test-quitter" }
func (q *quitter) Think(v *sim.AIView) {
	if err := v.EndCompany(); err != nil {
		panic(err)
	}
}
func (q *quitter) Save() []byte                         { return nil }
func (q *quitter) Load([]byte) error                    { return nil }
func (q *quitter) OnEvent(v *sim.AIView, e sim.AIEvent) { q.heard = append(q.heard, e) }

var lastQuitter *quitter

func init() {
	sim.RegisterAI("test-quitter", func() sim.CompanyAI { lastQuitter = &quitter{}; return lastQuitter })
}

// G5 and G12 on the title save: an AI company hears the news addressed to it
// and to everyone (not another company's), before its think; ending the
// company (aiDestroy) empties the slot — name, owner name, AI, competitor
// object, the messages' references to it — and frees it for the next one.
func TestEndCompanyAndEvents(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	m := &s.Map
	id := m.CreateAICompany("test-quitter", sim.AIOptions{}, om)
	if id == 0xFF {
		t.Fatal("no AI company created")
	}
	q := lastQuitter
	other := uint8((int(id) + 1) % 15)
	m.PostMessage(nil, sim.MsgVehicleCrashed, id, 7, 0xFFFF, 0xFFFF)
	m.PostMessage(nil, sim.MsgVehicleCrashed, other, 8, 0xFFFF, 0xFFFF)
	m.PostMessage(nil, sim.MsgCompanyPromoted, 0xFF, uint16(id), 3, 0xFFFF)
	competitor := m.CompanyFaces[id].Competitor
	removed := -1
	m.CompanyRemoved = func(c uint8) { removed = int(c) }
	for range 16 {
		step(s, om)
	}
	if len(q.heard) != 2 || q.heard[0].Type != sim.MsgVehicleCrashed || q.heard[0].Subjects[0] != 7 ||
		q.heard[1].Type != sim.MsgCompanyPromoted {
		t.Errorf("events heard: %+v; want its own crash and the promotion", q.heard)
	}
	if !m.CompanyEmpty(int(id)) || m.CompanyAIName(int(id)) != "" || removed != int(id) {
		t.Fatalf("company %d not removed (empty %v, AI %q, windows closed for %d)", id, m.CompanyEmpty(int(id)),
			m.CompanyAIName(int(id)), removed)
	}
	if m.CompanyFaces[id].OwnerName != 0 || om.CompetitorObjects[competitor] != nil {
		t.Errorf("owner name %#x, competitor object %d still loaded", m.CompanyFaces[id].OwnerName, competitor)
	}
	for _, msg := range m.Messages {
		if msg.Type == sim.MsgCompanyPromoted && msg.Subjects[0] == uint16(id) {
			t.Errorf("a message still names company %d", id)
		}
	}
	if m.EncodeAIChunk() != nil {
		t.Error("the save keeps an AI chunk for the removed company")
	}
	if again := m.CreateAICompany(BusLinkName, sim.AIOptions{}, om); again == 0xFF {
		t.Error("no company could start after the removal")
	}
}

// The events waiting for a think are saved with the AI (chunk version 2).
func TestAIEventsSaved(t *testing.T) {
	s, om := loadDataMap(t, "title.dat")
	m := &s.Map
	id := m.CreateAICompany(BusLinkName, sim.AIOptions{}, om)
	if id == 0xFF {
		t.Fatal("no AI company created")
	}
	m.PostMessage(nil, sim.MsgVehicleCrashed, id, 7, 0xFFFF, 0xFFFF)
	a := m.EncodeAIChunk()
	r := reload(t, saveBytes(t, s))
	if b := r.Map.EncodeAIChunk(); string(a) != string(b) {
		t.Errorf("the AI chunk changed through a save and load:\n%x\n%x", a, b)
	}
}
