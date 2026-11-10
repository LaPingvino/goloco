package sim

import "testing"

func TestCheats(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: m.ControllingCompany, Objects: om}
	pc := int(m.PlayerCompanies[0])
	cash := m.Companies[pc].Cash
	if _, ok := m.Cheat(c, 0, CheatAddCash, 5000, 0, 0); !ok || m.Companies[pc].Cash != cash {
		t.Fatal("a dry run changed the cash")
	}
	m.Cheat(c, CmdApply, CheatAddCash, 5000, 0, 0)
	if m.Companies[pc].Cash != cash+5000 {
		t.Errorf("cash %d, want %d", m.Companies[pc].Cash, cash+5000)
	}
	m.Companies[pc].CurrentLoan = 1234
	m.Cheat(c, CmdApply, CheatClearLoan, 0, 0, 0)
	if m.Companies[pc].CurrentLoan != 0 {
		t.Error("the loan stayed")
	}
	f := m.CompanyFaces[pc].ChallengeFlags
	m.Cheat(c, CmdApply, CheatToggleBankruptcy, int32(pc), 0, 0)
	if m.CompanyFaces[pc].ChallengeFlags^f != companyFlagBankrupt {
		t.Error("bankruptcy didn't toggle")
	}
	m.Cheat(c, CmdApply, CheatToggleJail, int32(pc), 0, 0)
	j := m.Companies[pc].JailStatus
	m.Cheat(c, CmdApply, CheatToggleJail, int32(pc), 0, 0)
	if (j != 30 && j != 0) || m.Companies[pc].JailStatus == j {
		t.Errorf("jail toggled %d → %d", j, m.Companies[pc].JailStatus)
	}
	m.Cheat(c, CmdApply, CheatModifyDate, 1955, 5, 17)
	if m.CurrentYear != 1955 || m.CurrentMonth != 5 || m.CurrentDayOfMonth != 16 || m.CurrentDay != CalcDays(1955, 5, 17) {
		t.Errorf("date %d-%d-%d (day %d)", m.CurrentYear, m.CurrentMonth, m.CurrentDayOfMonth, m.CurrentDay)
	}
	// Acquire another company's assets: none of its tracks, roads, stations or vehicles are left.
	other := -1
	trains, _ := m.Trains()
	for _, tr := range trains {
		if o := int(m.Entities[tr.Head].Owner); o != int(c.Company) && o != CompanyNeutral {
			other = o
			break
		}
	}
	if other < 0 {
		t.Skip("no other company's train")
	}
	m.Cheat(c, CmdApply, CheatAcquireAssets, int32(other), 0, 0)
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			for _, tr := range m.Tiles[y][x].Tracks {
				if int(tr.Owner) == other {
					t.Fatal("a track of the other company is left")
				}
			}
		}
	}
	for i := range m.Entities {
		if int(m.Entities[i].Owner) == other {
			t.Fatalf("vehicle entity %d still belongs to company %d", m.Entities[i].ID, other)
		}
	}
	// Ratings: absolute 1 → 1000 where rated.
	m.Cheat(c, CmdApply, CheatCompanyRatings, 1, 1, 0)
	for _, tw := range m.Towns {
		if tw.CompaniesWithRating != 0 && tw.CompanyRatings[m.ControllingCompany] != 1000 {
			t.Errorf("town %d rating %d, want 1000", tw.ID, tw.CompanyRatings[m.ControllingCompany])
		}
	}
}

func TestShuntAndFreeCashCheats(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	if len(trains) == 0 {
		t.Skip("no trains")
	}
	h := m.Entities[trains[0].Head]
	c := &Cmd{Company: h.Owner, Objects: om}
	if _, ok := m.VehicleShuntCheat(c, CmdApply, h.ID); !ok || m.Entities[trains[0].Head].VehicleFlags&vehicleFlagShuntCheat == 0 {
		t.Error("shunt cheat not set")
	}
	if _, ok := m.VehicleShuntCheat(c, CmdApply, 0xFFFE); ok {
		t.Error("shunt cheat on no vehicle")
	}
	n := len(m.Messages)
	m.FreeCashCheat(c, CmdApply)
	if m.Companies[c.Company].JailStatus != 30 || len(m.Messages) != n+1 || m.Messages[n].Type != MsgCompanyCheated {
		t.Fatalf("jail %d, messages %d → %d", m.Companies[c.Company].JailStatus, n, len(m.Messages))
	}
	t.Logf("%q", m.Messages[n].Text)
	if m.Messages[n].Text == "" {
		t.Error("no scandal text")
	}
}
