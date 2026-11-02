package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A new game's companies — CompanyManager::reset, createPlayerCompany,
// createCompany (the player branch), selectNewCompetitor,
// competingColourMask, getInflationAdjustedStartingLoan,
// determineAvailableVehicles' unlocked set.
//
// PARTIAL: createCompany's competitor branch (AI names, playstyles, town),
// selectNewCompetitor's fallback that loads an installed competitor object
// when every loaded one is taken, the preferred owner face / name options,
// the network player slots, the owner name, startedDate, secondary and
// vehicle colours, availableVehicles, AI state, transport type counts,
// observation / owner status, town company ratings, and
// updatePlayerInfrastructureOptions.

const (
	companyMaxCount     = 15
	stringIDNewCompany  = 2 // StringIds::new_company
	colourMax           = 31
	colourMutedSeaGreen = 11
	companyIDNull       = 0xFF
)

// ensureCompanySlots gives a map without company records (a .SC5) the
// fifteen empty slots a game has.
func (m *Map) ensureCompanySlots() {
	for len(m.Companies) < companyMaxCount {
		m.Companies = append(m.Companies, scenario.CompanyEconomy{})
	}
	for len(m.CompanyFaces) < companyMaxCount {
		m.CompanyFaces = append(m.CompanyFaces, scenario.CompanyFace{Competitor: 0xFF})
	}
	for len(m.CompanyNames) < companyMaxCount {
		m.CompanyNames = append(m.CompanyNames, 0)
	}
	for len(m.CompanyColours) < companyMaxCount {
		m.CompanyColours = append(m.CompanyColours, 0)
	}
	for len(m.UnlockedVehicles) < companyMaxCount {
		m.UnlockedVehicles = append(m.UnlockedVehicles, [7]uint32{})
	}
}

// ResetCompanies is CompanyManager::reset (single player): every company
// emptied, the player slots 0 / none, company 0's colour muted sea green.
func (m *Map) ResetCompanies() {
	m.ensureCompanySlots()
	for id := range m.CompanyNames {
		m.CompanyNames[id] = 0
	}
	m.PlayerCompanies = [2]uint8{0, companyIDNull}
	m.ControllingCompany = 0
	m.CompanyColours[0] = colourMutedSeaGreen
}

// competingColourMask: the colours too similar to another company's.
func (m *Map) competingColourMask(id int) uint32 {
	var mask uint32
	for o := range m.CompanyNames {
		if o != id && !m.CompanyEmpty(o) && int(m.CompanyColours[o]) < len(similarColourMask) {
			mask |= similarColourMask[m.CompanyColours[o]]
		}
	}
	return mask
}

// selectNewCompetitor: a loaded competitor object no company uses, chosen
// by gPrng1; 0xFF when none (the installed-object fallback is PARTIAL).
func (m *Map) selectNewCompetitor(om *objects.ObjectManager) uint8 {
	var free []int
	for i, c := range om.CompetitorObjects {
		if c == nil {
			continue
		}
		used := false
		for o := range m.CompanyNames {
			if !m.CompanyEmpty(o) && int(m.CompanyFaces[o].Competitor) == i {
				used = true
			}
		}
		if !used {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		return 0xFF
	}
	return uint8(free[int(m.Rng1.RandNext()&0xFFFF)*len(free)/65536])
}

// startingLoan is getInflationAdjustedStartingLoan.
func (m *Map) startingLoan() int32 {
	return m.InflationAdjustedCost(int16(m.StartingLoanSize), 0, 8) / 100 * 100
}

// createPlayerCompany is CompanyManager::createPlayerCompany: a company for
// the player with a free competitor's face. Returns its id (0xFF if none).
func (m *Map) CreatePlayerCompany(om *objects.ObjectManager) uint8 {
	id := m.createCompany(m.selectNewCompetitor(om), true, om)
	m.PlayerCompanies = [2]uint8{id, companyIDNull}
	m.ControllingCompany = id
	return id
}

// createCompany (the player branch; see PARTIAL).
func (m *Map) createCompany(competitor uint8, isPlayer bool, om *objects.ObjectManager) uint8 {
	m.ensureCompanySlots()
	id := -1
	for o := range m.CompanyNames {
		if m.CompanyEmpty(o) {
			id = o
			break
		}
	}
	if id < 0 || !isPlayer {
		return companyIDNull // the competitor branch: PARTIAL
	}
	face, c := &m.CompanyFaces[id], &m.Companies[id]
	face.Competitor, face.ChallengeFlags, face.Emotion = competitor, 0, 0
	c.Var49C, c.Var4A0 = 0, 0
	m.CompanyNames[id] = stringIDNewCompany
	for {
		primary := uint8((m.Rng1.RandNext() & 0xFF) * colourMax / 256)
		if m.competingColourMask(id)&(1<<primary) == 0 {
			m.CompanyColours[id] = primary
			break
		}
	}
	c.NumExpenditureYears = 1
	c.Expenditures[0] = [17]int32{}
	c.CurrentLoan = m.startingLoan()
	c.Cash = int64(c.CurrentLoan)
	m.UnlockedVehicles[id] = m.DetermineUnlockedVehicles(id, om)
	c.CargoUnitsTotalDelivered, c.CargoUnitsTotalDistance = 0, 0
	c.HistorySize = 1
	face.Performance = 0
	c.PerformanceIndexHist[0] = 0
	c.CargoUnitsDeliveredHist[0] = 0
	c.VehicleProfit, c.CompanyValueHistory[0] = m.CalculateCompanyValue(id)
	c.CurrentRating, c.ChallengeProgress, c.NumMonthsInTheRed = 0, 0, 0
	c.CargoDelivered = [32]uint32{}
	return uint8(id)
}
