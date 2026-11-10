package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// MessageManager (MessageManager.cpp): the news / advice message store in the
// GameState — posting with dedup and eviction, removal, subject clean-up and
// the daily expiry. The descriptors are generated (messagetypes_gen.go).

// MessageType (Message.h).
type MessageType uint8

const (
	MsgCantWaitForFullLoad MessageType = iota
	MsgIndustryClosingDown
	MsgVehicleSlipped
	MsgUnk3
	MsgUnk4
	MsgUnk5
	MsgUnk6
	MsgUnk7
	MsgUnk8
	MsgCargoNowAccepted
	MsgCargoNoLongerAccepted
	MsgNewCompany
	MsgUnableToLandAtAirport
	MsgCitizensCelebrate
	MsgWorkersCelebrate
	MsgNewVehicle
	MsgCompanyPromoted
	MsgNewIndustry
	MsgIndustryProductionUp
	MsgIndustryProductionDown
	MsgCongratulationsCompleted
	MsgFailedObjectives
	MsgHaveBeenBeaten
	MsgBankruptcyWarning6MonthsRemaining
	MsgBankruptcyWarning3MonthsRemaining
	MsgBankruptcyDeclared
	MsgBankruptcyDeclaredCompetitor
	MsgVehicleCrashed
	MsgCompanyCheated
	MsgNewSpeedRecord
	MsgNewSpeedRecord2
)

// MessageItemArgumentType (Message.h).
const (
	MsgArgIndustry   uint8 = 0
	MsgArgStation    uint8 = 1
	MsgArgTown       uint8 = 2
	MsgArgVehicle    uint8 = 3
	MsgArgCompany    uint8 = 4
	MsgArgLocation   uint8 = 5
	MsgArgVehicleTab uint8 = 7
	MsgArgNull       uint8 = 0xFF
)

const (
	msgFlagShouldDedup = 1 << 0 // MessageTypeFlags::shouldDedup
	kMaxMessages       = 200    // Limits::kMaxMessages
	messageIDNull      = 0xFFFF // MessageId::null
)

// Message is Message.h's Message. Text is messageString as formatted at
// post time (PARTIAL: not formatted yet — the news window formats from the
// subjects).
type Message struct {
	Type       MessageType
	Text       string
	CompanyID  uint8
	TimeActive uint16 // 0xFFFF: no longer active; bit 15: opened by the user
	Subjects   [3]uint16
	Date       uint32
}

func (a Message) same(b Message) bool { // operator==: the company isn't compared
	return a.Type == b.Type && a.Subjects == b.Subjects
}

func (m *Map) messageDuration(t MessageType) int64 {
	if int(t) >= len(MessageTypeDescriptors) {
		return 0
	}
	return int64(MessageTypeDescriptors[t].Duration)
}

// PostMessage is MessageManager::post: a message of the type for the company
// and subjects, dated today — none when the type dedups and one with the same
// type and subjects is there; at the limit, among the messages already
// shown (timeActive ≠ 0) the one with the most life left (duration + date −
// today, which upstream names "oldest"; ages below 0 never qualify) goes,
// else the same over all; the text formatted at posting (messageText). PARTIAL: the window
// invalidation.
func (m *Map) PostMessage(om *objects.ObjectManager, t MessageType, company uint8, a, b, c uint16) {
	msg := Message{Type: t, CompanyID: company, Subjects: [3]uint16{a, b, c}, Date: m.CurrentDay}
	if int(t) < len(MessageTypeDescriptors) && MessageTypeDescriptors[t].Flags&msgFlagShouldDedup != 0 {
		for _, e := range m.Messages {
			if e.same(msg) {
				return
			}
		}
	}
	if len(m.Messages) >= kMaxMessages {
		oldest, at := int64(-1), -1
		age := func(e Message) int64 { return m.messageDuration(e.Type) + int64(e.Date) - int64(m.CurrentDay) }
		for i, e := range m.Messages {
			if e.TimeActive != 0 && age(e) > oldest {
				oldest, at = age(e), i
			}
		}
		if oldest == -1 {
			for i, e := range m.Messages {
				if age(e) > oldest {
					oldest, at = age(e), i
				}
			}
		}
		if at >= 0 {
			m.removeMessage(at)
		}
	}
	msg.Text = m.messageText(msg, om)
	m.Messages = append(m.Messages, msg)
}

// clearActiveMessage (0x00428E0F): the shown message no longer active.
// PARTIAL: closing the news window.
func (m *Map) clearActiveMessage() {
	if i := int(m.ActiveMessageIndex); i < len(m.Messages) {
		m.Messages[i].TimeActive = 0xFFFF
	}
	m.ActiveMessageIndex = messageIDNull
}

// removeMessage is remove (0x00428DAB): the shown one cleared or its index
// moved down, the rest moving up one.
func (m *Map) removeMessage(i int) {
	if m.ActiveMessageIndex != messageIDNull {
		if i == int(m.ActiveMessageIndex) {
			m.clearActiveMessage()
		} else if i < int(m.ActiveMessageIndex) {
			m.ActiveMessageIndex--
		}
	}
	m.Messages = append(m.Messages[:i], m.Messages[i+1:]...)
}

// RemoveAllSubjectRefs is removeAllSubjectRefs (0x0042851C): every message
// subject of that argument type and id becomes 0xFFFF. Upstream's loop runs
// i = numMessages … 1 — the first message skipped, the slot past the last
// looked at (empty here) —, kept.
func (m *Map) RemoveAllSubjectRefs(subject uint16, argType uint8) {
	for i := len(m.Messages); i > 0; i-- {
		if i >= len(m.Messages) {
			continue // the slot past the last: an empty message upstream
		}
		e := &m.Messages[i]
		if int(e.Type) >= len(MessageTypeDescriptors) {
			continue
		}
		d := MessageTypeDescriptors[e.Type]
		for j := range 3 {
			if d.ArgumentTypes[j] == argType && e.Subjects[j] == subject {
				e.Subjects[j] = 0xFFFF
			}
		}
	}
}

// UpdateMessagesDaily is updateDaily (0x004284DB): from the last back, the
// messages whose duration has run out go.
func (m *Map) UpdateMessagesDaily() {
	for i := len(m.Messages) - 1; i >= 0; i-- {
		e := m.Messages[i]
		if int64(m.CurrentDay) >= int64(e.Date)+m.messageDuration(e.Type) {
			m.removeMessage(i)
		}
	}
}

// messageText is post's formatString per message type, cut at 197 bytes as
// upstream (locoStrlen, then the two-byte 0xFF check). Formatted so far: the
// industry messages (closing down, production up / down — the industry
// object's nameClosingDown / nameUpProduction / nameDownProduction with the
// industry's name for its stringidArgs). PARTIAL: the other types (empty
// text; the news window will need them).
func (m *Map) messageText(msg Message, om *objects.ObjectManager) string {
	var s string
	switch msg.Type {
	case MsgIndustryClosingDown, MsgIndustryProductionUp, MsgIndustryProductionDown:
		k := map[MessageType]int{MsgIndustryClosingDown: 3, MsgIndustryProductionUp: 4, MsgIndustryProductionDown: 5}[msg.Type]
		s = m.industryMessage(om, msg.Subjects[0], k)
	}
	if len(s) > 197 {
		s = s[:197]
	}
	return s
}

// industryMessage formats the industry object's string table k with the
// industry (by id) as stringidArgs' argument (its name, which takes its town).
func (m *Map) industryMessage(om *objects.ObjectManager, id uint16, k int) string {
	if om == nil {
		return ""
	}
	for i := range m.Industries {
		if uint16(m.Industries[i].ID) != id {
			continue
		}
		obj := om.GetIndustryObjectByIndex(int(m.Industries[i].ObjectID))
		if obj == nil {
			return ""
		}
		name, _ := m.IndustryName(i, om)
		var out []byte
		for _, c := range obj.RawStrings[k] {
			switch {
			case c == 0x83: // stringidArgs: the industry's name
				out = append(out, name...)
			case c >= 0x20 && c < 0x7F:
				out = append(out, c)
			}
		}
		return string(out)
	}
	return ""
}
