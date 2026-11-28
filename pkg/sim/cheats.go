package sim

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// GameCommands/Cheats/Cheat.cpp: the cheat game command's subcommands, and
// the vehicle shunt cheat.

// CheatCommand (Cheat.h).
type CheatCommand uint8

const (
	CheatAcquireAssets CheatCommand = iota
	CheatAddCash
	CheatClearLoan
	CheatCompanyRatings
	CheatSwitchCompany
	CheatToggleBankruptcy
	CheatToggleJail
	CheatVehicleReliability
	CheatModifyDate
	CheatCompleteChallenge
)

const (
	maxCompanyRating = 1000  // kMaxCompanyRating
	minCompanyRating = -1000 // kMinCompanyRating
)

// Cheat is the cheat game command (GenericCheatArgs: the subcommand and up
// to three parameters): nothing unless applied. Its cost (always 0).
// The vehicle shunt and free-cash cheats are their own commands below.
func (m *Map) Cheat(c *Cmd, flags uint8, sub CheatCommand, p1, p2, p3 int32) (int32, bool) {
	if flags&CmdApply == 0 {
		return 0, true
	}
	switch sub {
	case CheatAcquireAssets:
		m.cheatAcquireAssets(c.Company, uint8(p1), c.Objects)
	case CheatAddCash:
		if pc := int(m.PlayerCompanies[0]); pc < len(m.Companies) {
			m.Companies[pc].Cash += int64(p1)
		}
	case CheatClearLoan:
		if pc := int(m.PlayerCompanies[0]); pc < len(m.Companies) {
			m.Companies[pc].CurrentLoan = 0
		}
	case CheatCompanyRatings:
		m.cheatCompanyRatings(p1 != 0, p2)
	case CheatSwitchCompany:
		m.cheatSwitchCompany(uint8(p1))
	case CheatToggleBankruptcy:
		if int(p1) < len(m.CompanyFaces) {
			m.CompanyFaces[p1].ChallengeFlags ^= companyFlagBankrupt
		}
	case CheatVehicleReliability:
		m.cheatVehicleReliability(c.Company, p1)
	case CheatToggleJail:
		if int(p1) < len(m.Companies) {
			if m.Companies[p1].JailStatus == 0 {
				m.Companies[p1].JailStatus = 30
			} else {
				m.Companies[p1].JailStatus = 0
			}
		}
	case CheatModifyDate:
		m.InitialiseDateYMD(uint16(p1), uint8(p2), uint8(p3))
		log.Printf("Date set to: Day=%d Month=%d Year=%d", p3, p2, p1) // Logging::info
	case CheatCompleteChallenge:
		m.cheatCompleteChallenge(int(p1), c.Objects)
	}
	return 0, true
}

// cheatAcquireAssets is Cheats::acquireAssets (Cheat.cpp, as upstream
// 6d7f5b1 / #4041 has it): every road, track and station element of the
// target company becomes ours (the updating company's) unless it is
// AI-allocated — the AI's plans aren't inherited —; then its stations, and
// its vehicles, each bogie and body repainted in our colours (its own
// vehicle object's colour type: resetVehicleColour); then the target AI is
// shut down — aiThinkState endCompany, sub-state 0, aiPathfindTargetPos 0,0
// — so its next think removes what is left (aiThinkEndCompany). A company
// a goloco AI runs loses it: the company ends the same way. PARTIAL: the
// station / vehicle windows' owner and Gfx::invalidateScreen (no Cheats
// window calls this yet; a Station window takes its owner's colour when it
// opens).
func (m *Map) cheatAcquireAssets(ours, target uint8, om *objects.ObjectManager) {
	for y := range m.Tiles { // acquireTiles<RoadElement>, <TrackElement>, <StationElement>
		for x := range m.Tiles[y] {
			t := &m.Tiles[y][x]
			for _, r := range t.Order {
				ai := r.Flags&elemFlagAiAllocated != 0
				switch r.Kind {
				case scenario.KindRoad:
					if el := &t.Roads[r.Index]; el.Owner == target && !ai && !el.AiAllocated {
						el.Owner = ours
					}
				case scenario.KindTrack:
					if el := &t.Tracks[r.Index]; el.Owner == target && !ai && !el.AiAllocated {
						el.Owner = ours
					}
				case scenario.KindStation:
					if el := &t.Stations[r.Index]; el.Owner == target && !ai && !el.AiAllocated {
						el.Owner = ours
					}
				}
			}
		}
	}
	for i := range m.StationNames {
		if m.StationNames[i].Name != stationNameNull && m.StationNames[i].Owner == target {
			m.StationNames[i].Owner = ours
		}
	}
	trains, _ := m.Trains()
	for _, tr := range trains {
		if m.Entities[tr.Head].Owner != target {
			continue
		}
		for _, i := range trainComponents(tr) { // applyToComponents
			m.Entities[i].Owner = ours
		}
		for _, car := range tr.Cars { // the bogies and bodies: resetVehicleColour
			for _, comp := range car {
				for _, i := range []int{comp.Front, comp.Back, comp.Body} {
					m.resetVehicleColour(i, ours, om)
				}
			}
		}
	}
	if int(target) < len(m.Companies) { // fourth phase: shut down the AI company
		c := &m.Companies[target]
		c.AiThinkState, c.AiThinkSubState = aiThinkStateEndCompany, 0
		c.AiPathfindTargetX, c.AiPathfindTargetY = 0, 0
	}
	if int(target) < companyMaxCount {
		m.companyAIs[target] = nil
	}
}

// resetVehicleColour is Cheat.cpp's resetVehicleColour: the component takes
// the company's main colours, or its scheme for the component's vehicle
// object's colour type when that one is set.
func (m *Map) resetVehicleColour(i int, company uint8, om *objects.ObjectManager) {
	e := &m.Entities[i]
	if om == nil || int(e.ObjectID) >= len(om.Vehicles) || om.Vehicles[e.ObjectID] == nil {
		return
	}
	scheme := m.vehicleColourScheme(company, om.Vehicles[e.ObjectID].ColourType)
	e.ColourPrimary, e.ColourSecondary = scheme[0], scheme[1]
}

// cheatCompanyRatings: in every town that rates the controlling company —
// upstream tests with "companiesWithRating &= bit", which also clears the
// other companies' bits; kept — the rating set to value × 1000 (absolute) or
// scaled by 1 + 1 / value around −1000, clamped to ±1000.
func (m *Map) cheatCompanyRatings(absolute bool, value int32) {
	id := m.ControllingCompany
	for i := range m.Towns {
		tw := &m.Towns[i]
		tw.CompaniesWithRating &= 1 << id
		if tw.CompaniesWithRating == 0 || int(id) >= len(tw.CompanyRatings) {
			continue
		}
		var r int16
		if absolute {
			r = int16(value * maxCompanyRating)
		} else {
			r = tw.CompanyRatings[id] + maxCompanyRating
			r = int16(float32(r) * (1 + 1/float32(value))) // newRanking *= 1.0f + 1.0f / value (int16 ← float)
			r -= maxCompanyRating
		}
		tw.CompanyRatings[id] = min(max(r, minCompanyRating), maxCompanyRating)
	}
}

// cheatSwitchCompany: control passes to the target; when the other player
// has it, the two swap.
func (m *Map) cheatSwitchCompany(target uint8) {
	ours, other := m.PlayerCompanies[0], m.PlayerCompanies[1]
	if target == ours {
		return
	}
	if target == other {
		m.PlayerCompanies[0], m.PlayerCompanies[1] = other, ours
	} else {
		m.PlayerCompanies[0] = target
	}
	m.ControllingCompany = m.PlayerCompanies[0]
}

// cheatVehicleReliability: each of our trains' vehicle 2 takes the
// reliability, and each car's front bogie that × 256.
func (m *Map) cheatVehicleReliability(ours uint8, rel int32) {
	trains, _ := m.Trains()
	for _, tr := range trains {
		if m.Entities[tr.Head].Owner != ours {
			continue
		}
		m.Entities[tr.Veh2].Veh2Reliability = uint8(rel)
		for _, car := range tr.Cars {
			if len(car) > 0 {
				m.Entities[car[0].Front].Reliability = uint16(rel * 256)
			}
		}
	}
}

// cheatCompleteChallenge: unless already completed, the challenge flags
// cleared, progress 100, and the challenge evaluated (completing it).
func (m *Map) cheatCompleteChallenge(id int, om *objects.ObjectManager) {
	if id >= len(m.Companies) || id >= len(m.CompanyFaces) || m.CompanyFaces[id].ChallengeFlags&companyFlagChallengeCompleted != 0 {
		return
	}
	m.CompanyFaces[id].ChallengeFlags &^= companyFlagChallengeBeaten | companyFlagChallengeCompleted | companyFlagChallengeFailed
	m.Companies[id].ChallengeProgress = 100
	m.evaluateChallengeProgress(id, om)
}

// VehicleShuntCheat is vehicleShuntCheat (0x004BAC53): refused for no such
// head; applied, the head gets shuntCheat (its cars keep the player's order).
func (m *Map) VehicleShuntCheat(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	if flags&CmdApply != 0 {
		m.Entities[tr.Head].VehicleFlags |= vehicleFlagShuntCheat
	}
	return 0, true
}

// FreeCashCheat is freeCashCheat (0x00438A08): applied, the updating company
// goes to jail for 30 days and companyCheated is posted. PARTIAL: the
// company's dejected emotion, the window invalidations.
func (m *Map) FreeCashCheat(c *Cmd, flags uint8) (int32, bool) {
	if flags&CmdApply != 0 && int(c.Company) < len(m.Companies) {
		m.Companies[c.Company].JailStatus = 30
		m.PostMessage(c.Objects, MsgCompanyCheated, 0xFF, uint16(c.Company), 0xFFFF, 0xFFFF)
	}
	return 0, true
}
