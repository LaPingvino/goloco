package sim

import "testing"

func TestMessages(t *testing.T) {
	m := &Map{ActiveMessageIndex: messageIDNull, CurrentDay: 1000}
	// Dedup: cantWaitForFullLoad (shouldDedup) posted twice with the same subjects is kept once.
	m.PostMessage(nil, MsgCantWaitForFullLoad, 0, 5, 7, 0xFFFF)
	m.PostMessage(nil, MsgCantWaitForFullLoad, 1, 5, 7, 0xFFFF)
	if len(m.Messages) != 1 {
		t.Fatalf("dedup: %d messages", len(m.Messages))
	}
	// newCompany doesn't dedup.
	m.PostMessage(nil, MsgNewCompany, 0, 2, 3, 0xFFFF)
	m.PostMessage(nil, MsgNewCompany, 0, 2, 3, 0xFFFF)
	if len(m.Messages) != 3 {
		t.Fatalf("no dedup: %d messages", len(m.Messages))
	}
	// A vehicle subject removed: its references blank (not in the first message — upstream's loop).
	m.PostMessage(nil, MsgVehicleSlipped, 0, 5, 0xFFFF, 0xFFFF)
	m.RemoveAllSubjectRefs(5, MsgArgVehicle)
	if m.Messages[3].Subjects[0] != 0xFFFF {
		t.Error("the slipped vehicle's reference stayed")
	}
	if m.Messages[0].Subjects[0] != 5 {
		t.Error("the first message was touched (upstream skips index 0)")
	}
	// Expiry: vehicleSlipped lasts 30 days, newCompany 180, cantWaitForFullLoad 60.
	m.CurrentDay += 61
	m.UpdateMessagesDaily()
	if len(m.Messages) != 2 || m.Messages[0].Type != MsgNewCompany {
		t.Fatalf("after 61 days: %+v", m.Messages)
	}
	// The limit: among the messages already shown (timeActive ≠ 0) the one
	// with the most life left (duration + date − today; upstream's "oldest")
	// goes; here only slot 10 was shown.
	m.Messages = nil
	for i := range kMaxMessages {
		m.PostMessage(nil, MsgNewCompany, 0, uint16(i), 0, 0)
		m.Messages[len(m.Messages)-1].TimeActive = 0
	}
	m.Messages[10].TimeActive = 0xFFFF // shown and closed
	m.PostMessage(nil, MsgNewCompany, 0, 999, 0, 0)
	if len(m.Messages) != kMaxMessages || m.Messages[10].Subjects[0] == 10 {
		t.Errorf("eviction: %d messages, slot 10 subject %d", len(m.Messages), m.Messages[10].Subjects[0])
	}
}

// The title save's messages load: a count within the limit, known types,
// dates not after today, and the stored (formatted) text.
func TestLoadedMessages(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	t.Logf("%d messages, active %#x", len(m.Messages), m.ActiveMessageIndex)
	if len(m.Messages) > kMaxMessages {
		t.Fatalf("%d messages", len(m.Messages))
	}
	for i, e := range m.Messages {
		if int(e.Type) >= len(MessageTypeDescriptors) || e.Date > m.CurrentDay {
			t.Errorf("message %d: type %d date %d (today %d)", i, e.Type, e.Date, m.CurrentDay)
		}
		if i < 3 {
			t.Logf("  %d: type %d company %d subjects %v date %d %q", i, e.Type, e.CompanyID, e.Subjects, e.Date, e.Text)
		}
	}
	// None of them is past its duration (the game expires them daily).
	n := len(m.Messages)
	m.UpdateMessagesDaily()
	if len(m.Messages) != n {
		t.Errorf("a day's update dropped %d loaded messages already past their time", n-len(m.Messages))
	}
}

// An industry's production-up message reads as the game's: the object's
// nameUpProduction with the industry's name.
func TestIndustryMessageText(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Industries) == 0 {
		t.Skip("no industries")
	}
	ind := m.Industries[0]
	m.PostMessage(om, MsgIndustryProductionUp, 0xFF, uint16(ind.ID), 0xFFFF, 0xFFFF)
	last := m.Messages[len(m.Messages)-1]
	name, _ := m.IndustryName(0, om)
	t.Logf("%q", last.Text)
	if last.Text == "" || !contains(last.Text, name) {
		t.Errorf("text %q doesn't name %q", last.Text, name)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
