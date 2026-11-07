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
// the preferred owner face / name options,
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
	for len(m.CompanyColourSchemes) < companyMaxCount {
		m.CompanyColourSchemes = append(m.CompanyColourSchemes, scenario.CompanyColourScheme{})
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
// by gPrng1; else an installed one by the preferred-AI metric.
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
	if len(free) > 0 {
		return uint8(free[int(m.Rng1.RandNext()&0xFFFF)*len(free)/65536])
	}
	// No loaded competitor is free: the installed one closest to the
	// preferred AI character (ties all kept), chosen by gPrng1, is loaded.
	var best []*objects.CompetitorObject
	bestValue := 0xFF
	for _, c := range om.InstalledCompetitors() {
		inUse := false
		for _, l := range om.CompetitorObjects {
			if l == c {
				inUse = true
			}
		}
		if inUse {
			continue
		}
		metric := 0
		for k, v := range [3]uint8{c.Intelligence, c.Aggressiveness, c.Competitiveness} {
			if p := int(m.PreferredAI[k]); p != 0 && int(v) < len(aiToMetric) {
				metric += absInt(int(aiToMetric[v]) - p)
			}
		}
		if metric < bestValue {
			best, bestValue = []*objects.CompetitorObject{c}, metric
		} else if metric == bestValue {
			best = append(best, c)
		}
	}
	if len(best) == 0 {
		return 0xFF
	}
	return m.selectNewCompetitorFromHeader(best[int(m.Rng1.RandNext()&0xFFFF)*len(best)/65536], om)
}

var aiToMetric = [10]uint8{1, 1, 1, 1, 2, 2, 2, 3, 3, 3} // kAiToMetric

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// selectNewCompetitorFromHeader: the object's slot if loaded; else it is
// loaded into the first free competitor slot (ObjectManager::load's
// findFreeObjectId), unloading the first competitor no company uses when all
// 32 are taken.
func (m *Map) selectNewCompetitorFromHeader(c *objects.CompetitorObject, om *objects.ObjectManager) uint8 {
	for len(om.CompetitorObjects) < 32 {
		om.CompetitorObjects = append(om.CompetitorObjects, nil)
	}
	for i, l := range om.CompetitorObjects {
		if l == c {
			return uint8(i)
		}
	}
	slot := -1
	for i, l := range om.CompetitorObjects {
		if l == nil {
			slot = i
			break
		}
	}
	if slot < 0 {
		for i := range om.CompetitorObjects {
			used := false
			for o := range m.CompanyNames {
				if !m.CompanyEmpty(o) && int(m.CompanyFaces[o].Competitor) == i {
					used = true
				}
			}
			if !used {
				slot = i
				break
			}
		}
		if slot < 0 {
			return 0xFF
		}
	}
	om.CompetitorObjects[slot] = c
	return uint8(slot)
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
	face.OwnerName = stringIDNewOwner // company->ownerName = StringIds::new_owner
	c.Var49C, c.Var4A0 = 0, 0
	m.CompanyNames[id] = stringIDNewCompany
	for {
		primary := uint8((m.Rng1.RandNext() & 0xFF) * colourMax / 256)
		if m.competingColourMask(id)&(1<<primary) == 0 {
			m.CompanyColours[id] = primary
			scheme := [2]uint8{primary, primaryToSecondary[primary]}
			cs := &m.CompanyColourSchemes[id]
			cs.Main, cs.CustomSet = scheme, 0
			for k := range cs.Vehicle {
				cs.Vehicle[k] = scheme
			}
			break
		}
	}
	c.NumExpenditureYears = 1
	c.Expenditures[0] = [17]int32{}
	c.CurrentLoan = m.startingLoan()
	c.Cash = int64(c.CurrentLoan)
	m.DetermineAvailableVehicles(id, om)
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

// StartScenario is Scenario::start's game-state part for a new game from a
// .SC5, in upstream's order: companies reset and the player company created
// (before the date, so its loan uses the file's inflation), initialiseDate
// for the scenario's start year (every company's unlocked vehicles
// redetermined), sub_4748D4's default level crossing, the objective's time
// limit and month count.
//
// PARTIAL: the snow line and season, sub_4748D4's road object flags, last
// track type, infrastructure options and terraform objects, setRecords(zero),
// initialiseDefaultTrackRoadMods, the scenario name / details copy, labels
// and the windows / viewport / music resets (cmd/goloco's side).
func (m *Map) StartScenario(startYear uint16, om *objects.ObjectManager) {
	m.StartScenarioGenerating(startYear, nil, nil, om)
}

// StartScenarioGenerating is StartScenario for a .SC5 whose landscape is not
// in the file (tileManagerLoaded clear): generateLandscape runs first, on
// the freshly seeded rng, with the scenario's terrain options (gen non-nil).
func (m *Map) StartScenarioGenerating(startYear uint16, gen *TerrainOptions, g1 G1Raw, om *objects.ObjectManager) {
	if gen != nil {
		m.GenerateLandscape(startYear, *gen, g1, om)
	}
	m.ResetCompanies()
	m.CreatePlayerCompany(om)
	m.InitialiseDate(startYear)
	for id := range m.Companies {
		if !m.CompanyEmpty(id) {
			m.DetermineAvailableVehicles(id, om) // CompanyManager::determineAvailableVehicles
		}
	}
	m.UpdateDefaultLevelCrossingType(om)
	m.ObjectiveUntilYear = uint16(m.Objective.TimeLimitYears) - 1 + m.CurrentYear
	m.ObjectiveMonthsInChallenge = 0
}

// vehicleColourScheme is createCar's colour choice: the company's main
// colours, or its scheme for the vehicle's colour type when that one is set.
func (m *Map) vehicleColourScheme(company, colourType uint8) [2]uint8 {
	if int(company) >= len(m.CompanyColourSchemes) {
		return [2]uint8{}
	}
	cs := &m.CompanyColourSchemes[company]
	if cs.CustomSet&(1<<colourType) != 0 && colourType >= 1 && int(colourType) <= len(cs.Vehicle) {
		return cs.Vehicle[colourType-1]
	}
	return cs.Main
}
