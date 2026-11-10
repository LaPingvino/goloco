package sim

import "testing"

func TestMessages(t *testing.T) {
	m := &Map{ActiveMessageIndex: messageIDNull, CurrentDay: 1000}
	// Dedup: cantWaitForFullLoad (shouldDedup) posted twice with the same subjects is kept once.
	m.PostMessage(MsgCantWaitForFullLoad, 0, 5, 7, 0xFFFF)
	m.PostMessage(MsgCantWaitForFullLoad, 1, 5, 7, 0xFFFF)
	if len(m.Messages) != 1 {
		t.Fatalf("dedup: %d messages", len(m.Messages))
	}
	// newCompany doesn't dedup.
	m.PostMessage(MsgNewCompany, 0, 2, 3, 0xFFFF)
	m.PostMessage(MsgNewCompany, 0, 2, 3, 0xFFFF)
	if len(m.Messages) != 3 {
		t.Fatalf("no dedup: %d messages", len(m.Messages))
	}
	// A vehicle subject removed: its references blank (not in the first message — upstream's loop).
	m.PostMessage(MsgVehicleSlipped, 0, 5, 0xFFFF, 0xFFFF)
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
		m.PostMessage(MsgNewCompany, 0, uint16(i), 0, 0)
		m.Messages[len(m.Messages)-1].TimeActive = 0
	}
	m.Messages[10].TimeActive = 0xFFFF // shown and closed
	m.PostMessage(MsgNewCompany, 0, 999, 0, 0)
	if len(m.Messages) != kMaxMessages || m.Messages[10].Subjects[0] == 10 {
		t.Errorf("eviction: %d messages, slot 10 subject %d", len(m.Messages), m.Messages[10].Subjects[0])
	}
}
