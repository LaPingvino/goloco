package sim

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/uilayout"
)

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
// post time (messageText), Loco-encoded bytes.
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
// else the same over all; the text formatted at posting (messageText).
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

// clearActiveMessage (0x00428E0F): the shown message no longer active, the
// news window closed (CloseNewsWindow: the game's, nil without one).
func (m *Map) clearActiveMessage() {
	if i := int(m.ActiveMessageIndex); i < len(m.Messages) {
		m.Messages[i].TimeActive = 0xFFFF
	}
	if m.CloseNewsWindow != nil {
		m.CloseNewsWindow()
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

// messageText is post's formatString per message type, as Loco-encoded bytes
// (messageString: the save's messages are stored so), cut at 197 bytes as
// upstream (locoStrlen, then the two-byte 0xFF check). Formatted so far: the
// cargo accepted / no longer accepted, citizens / workers celebrate, new
// industry messages, and the industry messages (closing down, production up / down — the industry
// object's nameClosingDown / nameUpProduction / nameDownProduction with the
// industry's name for its stringidArgs). unk3–unk8 have no text upstream
// either (0x004287F4: a nop).
func (m *Map) messageText(msg Message, om *objects.ObjectManager) string {
	var s string
	switch msg.Type {
	case MsgIndustryClosingDown, MsgIndustryProductionUp, MsgIndustryProductionDown:
		k := map[MessageType]int{MsgIndustryClosingDown: 3, MsgIndustryProductionUp: 4, MsgIndustryProductionDown: 5}[msg.Type]
		s = m.industryMessage(om, msg.Subjects[0], k)
	case MsgCargoNowAccepted, MsgCargoNoLongerAccepted: // the station (with its town), the cargo
		id := i18n.MessageNowAccepts
		if msg.Type == MsgCargoNoLongerAccepted {
			id = i18n.MessageNoLongerAccepts
		}
		cargo := ""
		if om != nil && int(msg.Subjects[1]) < len(om.CargoObjects) && om.CargoObjects[msg.Subjects[1]] != nil {
			cargo = om.CargoObjects[msg.Subjects[1]].Name // an object string: Loco bytes
		}
		s = fillStringIDs(locoText(i18n.Str(id)), m.StationName(int(msg.Subjects[0])), cargo)
	case MsgCitizensCelebrate, MsgWorkersCelebrate: // subject 2: the cargo << 8 | the town / industry
		cargo := ""
		if co := m.cargoObject(om, msg.Subjects[2]>>8); co != nil {
			cargo = co.Name // an object string: Loco bytes
		}
		if msg.Type == MsgCitizensCelebrate {
			id := i18n.MessageFirstStringArriveAt
			if co := m.cargoObject(om, msg.Subjects[2]>>8); co != nil && co.Flags&cargoFlagDelivering != 0 {
				id = i18n.MessageFirstStringDeliveryArivesAt
			}
			town := ""
			if tw := m.TownByID(msg.Subjects[2] & 0xFF); tw != nil {
				town = m.TownName(tw.Name, int(tw.ID))
			}
			s = fillStringIDs(locoText(i18n.Str(id)), cargo, town)
		} else {
			name := ""
			if i := m.industryIndex(msg.Subjects[2] & 0xFF); i >= 0 {
				name, _ = m.IndustryName(i, om)
			}
			s = fillStringIDs(locoText(i18n.Str(i18n.MessageFirstStringDeliveredTo)), cargo, name)
		}
	case MsgUnableToLandAtAirport: // the vehicle (name + ordinal), the airport's station
		if h := m.entity(msg.Subjects[0]); h != nil {
			s = fillStringIDs(locoText(i18n.Str(i18n.MessageNotAbleToLandAt)), m.VehicleName(h.Name, int(h.OrdinalNumber)), m.StationName(int(msg.Subjects[1])))
		}
	case MsgNewCompany: // the company's name, the town's
		town := ""
		if tw := m.TownByID(msg.Subjects[1]); tw != nil {
			town = m.TownName(tw.Name, int(tw.ID))
		}
		s = fillStringIDs(locoText(i18n.Str(i18n.MessageNewTransportCompany)), m.companyNameLoco(int(msg.Subjects[0])), town)
	case MsgCantWaitForFullLoad, MsgVehicleSlipped, MsgVehicleCrashed: // the vehicle (name + ordinal); the station
		if h := m.entity(msg.Subjects[0]); h != nil {
			veh := m.VehicleName(h.Name, int(h.OrdinalNumber))
			if msg.Type == MsgVehicleCrashed {
				s = fillStringIDs(locoText(i18n.Str(i18n.MessageHasCrashed)), veh)
			} else if msg.Type == MsgVehicleSlipped {
				s = fillStringIDs(locoText(i18n.Str(i18n.MessageHasSlippedToAHaltOnIncline)), veh)
			} else {
				s = fillStringIDs(locoText(i18n.Str(i18n.MessageNotAllowedToWaitForFullLoadAt)), veh, m.StationName(int(msg.Subjects[1])))
			}
		}
	case MsgCongratulationsCompleted, MsgFailedObjectives, MsgHaveBeenBeaten: // the company's owner name
		id := map[MessageType]int{MsgCongratulationsCompleted: i18n.MessageCongratulationsYouHaveCompleted,
			MsgFailedObjectives: i18n.MessageFailureYouHaveFailed, MsgHaveBeenBeaten: i18n.MessageBeatenHasCompleted}[msg.Type]
		s = fillStringIDs(locoText(i18n.Str(id)), m.ownerName(int(msg.Subjects[0]), om))
	case MsgNewVehicle: // the vehicle object's name
		if om != nil && int(msg.Subjects[0]) < len(om.Vehicles) && om.Vehicles[msg.Subjects[0]] != nil {
			s = fillStringIDs(locoText(i18n.Str(i18n.MessageNewVehicleInvented)), om.Vehicles[msg.Subjects[0]].DisplayName)
		}
	case MsgBankruptcyWarning6MonthsRemaining, MsgBankruptcyWarning3MonthsRemaining, MsgBankruptcyDeclared, MsgBankruptcyDeclaredCompetitor: // the company's name
		id := map[MessageType]int{MsgBankruptcyWarning6MonthsRemaining: i18n.MessageBankruptcyWarning6MonthsRemaining,
			MsgBankruptcyWarning3MonthsRemaining: i18n.MessageBankruptcyWarning3MonthsRemaining,
			MsgBankruptcyDeclared:                i18n.MessageBankrupt, MsgBankruptcyDeclaredCompetitor: i18n.MessageBankrupt2}[msg.Type]
		s = fillStringIDs(locoText(i18n.Str(id)), m.companyNameLoco(int(msg.Subjects[0])))
	case MsgCompanyPromoted: // the owner's name, the old and the new rating (subject 1)
		rating := func(r int) string {
			if r < 0 || r > 9 { // getCorporateRatingAsStringId: unknown → platelayer
				r = 0
			}
			return locoText(i18n.Str(i18n.CorporateRatingPlatelayer + r))
		}
		s = fillStringIDs(locoText(i18n.Str(i18n.MessageIsPromotedTo)), m.ownerName(int(msg.Subjects[0]), om),
			rating(int(msg.Subjects[1])-1), rating(int(msg.Subjects[1])))
	case MsgNewSpeedRecord, MsgNewSpeedRecord2: // the record kind, the vehicle's type, the record speed
		if h := m.entity(msg.Subjects[0]); h != nil && msg.Subjects[2] < 3 && h.VehicleType < 6 {
			s = fillStringIDs(locoText(i18n.Str(i18n.MessageNewSpeedRecord)),
				locoText(i18n.Str(i18n.Land+int(msg.Subjects[2]))), locoText(i18n.Str(i18n.Train2+int(h.VehicleType))))
			s = strings.Replace(s, "{VELOCITY}", locoText(uilayout.Velocity(int(m.CompanyRecords.Speed[msg.Subjects[2]]))), 1)
		}
	case MsgCompanyCheated: // the owner's name
		s = fillStringIDs(locoText(i18n.Str(i18n.MessageCorporateScandal)), m.ownerName(int(msg.Subjects[0]), om))
	case MsgNewIndustry: // the industry object's name, the industry's town
		if i := m.industryIndex(msg.Subjects[0]); i >= 0 && om != nil {
			ind := m.Industries[i]
			obj, town := "", ""
			if o := om.GetIndustryObjectByIndex(int(ind.ObjectID)); o != nil {
				obj = o.Name // an object string (cleaned to ASCII)
			}
			if tw := m.TownByID(ind.Town); tw != nil {
				town = m.TownName(tw.Name, int(tw.ID))
			}
			s = fillStringIDs(locoText(i18n.Str(i18n.MessageNewStringUnderConstructionNear)), obj, town)
		}
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

// locoText is a UTF-8 string (a language pack's — object strings are Loco bytes already) as Loco bytes
// (convertUnicodeToLoco per character), {NEWLINE} as its control code (5).
func locoText(s string) string {
	s = strings.ReplaceAll(s, "{NEWLINE}", "\x05")
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 0x20 { // control codes (the newline) pass as they are
			b = append(b, byte(r))
			continue
		}
		b = append(b, byte(i18n.LocoChar(r)))
	}
	return string(b)
}

const cargoFlagDelivering = 1 << 2 // CargoObjectFlags::delivering

func (m *Map) cargoObject(om *objects.ObjectManager, id uint16) *objects.CargoObject {
	if om == nil || int(id) >= len(om.CargoObjects) {
		return nil
	}
	return om.CargoObjects[id]
}

// industryIndex is the index in m.Industries of the industry with this id (−1 none).
func (m *Map) industryIndex(id uint16) int {
	for i := range m.Industries {
		if uint16(m.Industries[i].ID) == id {
			return i
		}
	}
	return -1
}

// fillStringIDs puts the (already formatted, Loco-encoded) arguments in for
// the template's {STRINGID}s in order; other codes are left as written.
func fillStringIDs(tmpl string, args ...string) string {
	var b strings.Builder
	for {
		i := strings.Index(tmpl, "{STRINGID}")
		if i < 0 || len(args) == 0 {
			b.WriteString(tmpl)
			return b.String()
		}
		b.WriteString(tmpl[:i])
		b.WriteString(args[0])
		args, tmpl = args[1:], tmpl[i+len("{STRINGID}"):]
	}
}

// ownerName is OwnerName as Loco bytes: user strings and object strings are
// stored so; only the pack's new_owner is converted.
func (m *Map) ownerName(company int, om *objects.ObjectManager) string {
	if company >= 0 && company < len(m.CompanyFaces) && m.CompanyFaces[company].OwnerName == stringIDNewOwner {
		return locoText(m.OwnerName(company, om)) // the pack's (UTF-8) string
	}
	return m.OwnerName(company, om) // a user string or an object's: Loco bytes already
}

// companyNameLoco is CompanyName as Loco bytes (user strings are stored so;
// the pack's strings converted).
func (m *Map) companyNameLoco(company int) string {
	if company >= 0 && company < len(m.CompanyNames) && m.CompanyNames[company] >= kUserStringsStart {
		return m.CompanyName(company)
	}
	return locoText(m.CompanyName(company))
}
