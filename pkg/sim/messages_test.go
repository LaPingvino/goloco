package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

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

// A controlling-company station whose acceptance comes back posts "now
// accepts", formatted with the station and the cargo.
func TestCargoAcceptedMessage(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	id, cargo := -1, -1
	for i, st := range m.StationNames {
		if st.Name == stationNameNull {
			continue
		}
		for c := range st.Cargo {
			if st.Cargo[c].Flags&cargoStatAccepted != 0 {
				id, cargo = i, c
				break
			}
		}
		if id >= 0 {
			break
		}
	}
	if id < 0 {
		t.Skip("no station accepting cargo")
	}
	m.ControllingCompany = m.StationNames[id].Owner
	m.StationNames[id].Cargo[cargo].Flags &^= cargoStatAccepted
	n := len(m.Messages)
	m.ScenarioTicks = uint32(id) // stationsUpdate runs station scenarioTicks & 0x3FF
	m.stationsUpdate(om)
	if len(m.Messages) != n+1 {
		t.Fatalf("%d messages posted, want 1", len(m.Messages)-n)
	}
	got := m.Messages[len(m.Messages)-1]
	t.Logf("%q", got.Text)
	if got.Type != MsgCargoNowAccepted || int(got.Subjects[0]) != id || int(got.Subjects[1]) != cargo || !contains(got.Text, " now accepts ") {
		t.Errorf("message %+v", got)
	}
}

// The celebration and new-industry texts, with the pack's {NEWLINE} as the
// newline control code (5) the save's messages carry.
func TestCelebrationAndNewIndustryText(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Towns) == 0 || len(m.Industries) == 0 {
		t.Skip("no towns / industries")
	}
	tw, ind := m.Towns[0], m.Industries[0]
	m.PostMessage(om, MsgCitizensCelebrate, 0, 1, tw.ID, tw.ID|0<<8)
	m.PostMessage(om, MsgWorkersCelebrate, 0, 1, uint16(ind.ID), uint16(ind.ID)|0<<8)
	m.PostMessage(om, MsgNewIndustry, 0xFF, uint16(ind.ID), 0xFFFF, 0xFFFF)
	n := len(m.Messages)
	for _, e := range m.Messages[n-3:] {
		t.Logf("%d: %q", e.Type, e.Text)
		if e.Text == "" {
			t.Errorf("type %d has no text", e.Type)
		}
	}
	if !contains(m.Messages[n-3].Text, "Citizens Celebrate!\x05First ") {
		t.Errorf("citizens: %q", m.Messages[n-3].Text)
	}
	name, _ := m.IndustryName(0, om)
	if !contains(m.Messages[n-2].Text, name) {
		t.Errorf("workers: %q lacks %q", m.Messages[n-2].Text, name)
	}
}

// The vehicle news texts name the vehicle ("Train 3") and, for the full load,
// the station.
func TestVehicleMessageText(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	if len(trains) == 0 {
		t.Skip("no trains")
	}
	h := m.Entities[trains[0].Head]
	st := -1
	for i, s := range m.StationNames {
		if s.Name != stationNameNull {
			st = i
			break
		}
	}
	m.PostMessage(om, MsgVehicleSlipped, h.Owner, h.ID, 0xFFFF, 0xFFFF)
	m.PostMessage(om, MsgCantWaitForFullLoad, h.Owner, h.ID, uint16(st), 0xFFFF)
	n := len(m.Messages)
	veh := m.VehicleName(h.Name, int(h.OrdinalNumber))
	for _, e := range m.Messages[n-2:] {
		t.Logf("%q", e.Text)
		if !contains(e.Text, veh) {
			t.Errorf("%q doesn't name %q", e.Text, veh)
		}
	}
	if !contains(m.Messages[n-1].Text, m.StationName(st)) {
		t.Error("the full-load message doesn't name the station")
	}
}

// The challenge's end: the controlling company completing it posts the
// congratulations (with its owner's name); another company completing it
// posts "have been beaten" and marks the controlling company beaten.
func TestChallengeMessages(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	pc := int(m.ControllingCompany)
	other := -1
	for i := range m.Companies {
		if i != pc && !m.CompanyEmpty(i) {
			other = i
			break
		}
	}
	m.CompanyFaces[pc].ChallengeFlags = 0
	m.Companies[pc].ChallengeProgress = 100
	m.evaluateChallengeProgress(pc, om)
	got := m.Messages[len(m.Messages)-1]
	t.Logf("%q", got.Text)
	if got.Type != MsgCongratulationsCompleted || got.Text == "" {
		t.Errorf("controlling company: %+v", got)
	}
	if other < 0 {
		return
	}
	m.CompanyFaces[pc].ChallengeFlags = 0
	m.Companies[other].ChallengeProgress = 100
	m.evaluateChallengeProgress(other, om)
	got = m.Messages[len(m.Messages)-1]
	t.Logf("%q", got.Text)
	if got.Type != MsgHaveBeenBeaten || m.CompanyFaces[pc].ChallengeFlags&companyFlagChallengeBeaten == 0 {
		t.Errorf("another company: %+v, beaten flag %v", got, m.CompanyFaces[pc].ChallengeFlags&companyFlagChallengeBeaten != 0)
	}
}

// A vehicle designed this year, not yet unlocked for the controlling company,
// is unlocked and announced ("New Vehicle Invented - “…”!").
func TestNewVehicleMessage(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	pc := int(m.ControllingCompany)
	vid := -1
	for i, v := range om.Vehicles {
		if v != nil && v.Flags&vehicleFlagQuietInvention == 0 && (m.ForbiddenVehiclesPlayers|m.ForbiddenVehiclesCompetitors)&(1<<v.Type) == 0 {
			vid = i
			break
		}
	}
	if vid < 0 || pc >= len(m.UnlockedVehicles) {
		t.Skip("no announced vehicle")
	}
	m.CurrentYear = om.Vehicles[vid].Designed
	m.UnlockedVehicles[pc][vid/32] &^= 1 << (vid % 32)
	n := len(m.Messages)
	m.updateYearly2(om)
	if m.UnlockedVehicles[pc][vid/32]&(1<<(vid%32)) == 0 {
		t.Fatal("not unlocked")
	}
	found := false
	for _, e := range m.Messages[n:] {
		if e.Type == MsgNewVehicle && int(e.Subjects[0]) == vid {
			found = true
			t.Logf("%q", e.Text)
			if !contains(e.Text, om.Vehicles[vid].DisplayName) || om.Vehicles[vid].DisplayName == "" {
				t.Errorf("text %q, name %q", e.Text, om.Vehicles[vid].DisplayName)
			}
		}
	}
	if !found {
		t.Error("no newVehicle message")
	}
}

// Months in the red: the controlling company is warned at 3 and 6 and
// declared bankrupt at 9, each with its company's name.
func TestBankruptcyMessages(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	pc := int(m.ControllingCompany)
	for i := range m.Companies {
		if i != pc {
			m.Companies[i].Cash = 1 << 30 // keep the others out of the red
		}
	}
	m.CompanyFaces[pc].ChallengeFlags &^= companyFlagBankrupt
	m.Companies[pc].NumMonthsInTheRed = 0
	n := len(m.Messages)
	var types []MessageType
	for range 9 {
		m.Companies[pc].Cash = -1000000
		m.companiesUpdateMonthly1()
	}
	for _, e := range m.Messages[n:] {
		if e.CompanyID == uint8(pc) {
			types = append(types, e.Type)
			t.Logf("%q", e.Text)
		}
	}
	want := []MessageType{MsgBankruptcyWarning6MonthsRemaining, MsgBankruptcyWarning3MonthsRemaining, MsgBankruptcyDeclared}
	if len(types) != 3 || types[0] != want[0] || types[1] != want[1] || types[2] != want[2] {
		t.Errorf("posted %v, want %v", types, want)
	}
	if m.CompanyFaces[pc].ChallengeFlags&companyFlagBankrupt == 0 {
		t.Error("not bankrupt after 9 months")
	}
}

// A promotion: the controlling company's rating going up posts
// companyPromoted naming the old and new rating.
func TestPromotionMessage(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	pc := int(m.ControllingCompany)
	m.PostMessage(om, MsgCompanyPromoted, 0xFF, uint16(pc), 3, 0xFFFF)
	got := m.Messages[len(m.Messages)-1].Text
	t.Logf("%q", got)
	if !contains(got, " is promoted from ") || !contains(got, locoText(i18n.Str(i18n.CorporateRatingPlatelayer+2))) || !contains(got, locoText(i18n.Str(i18n.CorporateRatingPlatelayer+3))) {
		t.Errorf("text %q", got)
	}
}
