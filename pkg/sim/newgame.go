package sim

import (
	"math/bits"
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A new game's companies — CompanyManager::reset, createPlayerCompany,
// createCompany (the player branch), selectNewCompetitor,
// competingColourMask, getInflationAdjustedStartingLoan,
// determineAvailableVehicles' unlocked set.
//
// PARTIAL: the preferred owner face / name options,
// the network player slots, the owner name, startedDate, secondary and
// vehicle colours, availableVehicles, AI state, transport type counts,
// observation / owner status, town company ratings, and
// updatePlayerInfrastructureOptions.

const (
	companyMaxCount     = 15
	stringIDNewCompany  = 2 // StringIds::new_company
	colourMax           = 31
	colourMutedSeaGreen = 11
	colourGrey          = 1 // Colour::grey: the neutral company's
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
	for len(m.CompanyColours) < companyMaxCount+1 { // updateColours: every company's, then the neutral owner's (15) grey
		m.CompanyColours = append(m.CompanyColours, 0)
	}
	m.CompanyColours[companyMaxCount] = colourGrey
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

// createCompany is CompanyManager::createCompany (0x0042FE06): the first
// empty slot gets the competitor's face; a player's company a random free
// colour; a competitor's (isPlayer false) its owner's first name, a name
// prefix and playstyle drawn from the object's sets (250 tries for a colour
// no other company has: a prefix's own colour, or a random one), the
// playstyle's flags, a home town for the town-name prefix, and its name
// formatted from the playstyle's string (refused if another company has it).
// Then the common part: the expenditures, the AI state reset, the starting
// loan as cash, the unlocked vehicles, the history, the observation, and
// every town's rating of it at 500. Returns its id (0xFF: no slot, or the
// competitor branch refused — the slot is left empty).
// PARTIAL: a competitor's ownerName is 0, not the object's string id
// (object string ids are not modelled; OwnerName reads the competitor
// object's name, which is the same text, but a save written then shows no
// owner name in OpenLoco); updateCounter is not modelled.
func (m *Map) createCompany(competitor uint8, isPlayer bool, om *objects.ObjectManager) uint8 {
	m.ensureCompanySlots()
	id := -1
	for o := range m.CompanyNames {
		if m.CompanyEmpty(o) {
			id = o
			break
		}
	}
	if id < 0 {
		return companyIDNull
	}
	var compObj *objects.CompetitorObject
	if om != nil && int(competitor) < len(om.CompetitorObjects) {
		compObj = om.CompetitorObjects[competitor]
	}
	face, c := &m.CompanyFaces[id], &m.Companies[id]
	face.Competitor, face.ChallengeFlags, face.Emotion = competitor, 0, 0
	c.Var49C, c.Var4A0 = 0, 0 // cargoTypesDelivered, cargoTypesDelivered2
	m.CompanyNames[id] = stringIDNewCompany
	face.OwnerName = stringIDNewOwner // company->ownerName = StringIds::new_owner
	c.StartedDate = m.CurrentDay      // company->startedDate = getCurrentDay()
	setScheme := func(primary uint8) {
		m.CompanyColours[id] = primary
		scheme := [2]uint8{primary, primaryToSecondary[primary]}
		cs := &m.CompanyColourSchemes[id]
		cs.Main, cs.CustomSet = scheme, 0
		for k := range cs.Vehicle {
			cs.Vehicle[k] = scheme
		}
	}
	if isPlayer {
		for {
			primary := uint8((m.Rng1.RandNext() & 0xFF) * colourMax / 256)
			if m.competingColourMask(id)&(1<<primary) == 0 {
				setScheme(primary)
				break
			}
		}
	} else {
		face.OwnerName = 0 // competitorObj->firstName: see PARTIAL
		if !m.createCompetitorCompany(id, compObj, setScheme) {
			m.CompanyNames[id] = 0 // company->name = StringIds::empty
			return companyIDNull
		}
	}
	c.NumExpenditureYears = 1
	c.Expenditures[0] = [17]int32{}
	c.AiThinkState = 0 // AiThinkState::unk0; aiPlaceVehicleState, aiThinkTimer: the raw record (s5save initCreatedCompany)
	for k := range c.ThoughtType {
		c.ThoughtType[k] = 0xFF // AiThoughtType::null
	}
	c.HeadquartersX = -1
	m.UnlockedVehicles[id] = [7]uint32{} // unlockedVehicles.reset()
	c.AvailableVehicles = 0
	c.CurrentLoan = m.startingLoan()
	c.Cash = int64(c.CurrentLoan)
	m.DetermineAvailableVehicles(id, om)
	c.CargoUnitsTotalDelivered, c.CargoUnitsTotalDistance = 0, 0
	c.HistorySize = 1
	face.Performance = 0
	c.PerformanceIndexHist[0] = 0
	c.CargoUnitsDeliveredHist[0] = 0
	c.TransportTypeCount = [6]uint16{}
	c.ActiveEmotions = [9]uint8{}
	c.VehicleProfit, c.CompanyValueHistory[0] = m.CalculateCompanyValue(id)
	c.ObservationStatus, c.ObservationEntity = 0, entityIDNull // ObservationStatus::empty, EntityId::null
	c.ObservationX, c.ObservationY = -1, -1
	c.ObservationObject, c.ObservationTimeout = 0xFFFF, 0
	c.OwnerStatus = [2]int16{}
	c.CurrentRating, c.ChallengeProgress, c.NumMonthsInTheRed = 0, 0, 0
	c.JailStatus = 0
	c.CargoDelivered = [32]uint32{}
	for i := range m.Towns {
		if m.Towns[i].Name == 0xFFFF {
			continue
		}
		m.Towns[i].CompanyRatings[id] = 500
		m.Towns[i].CompaniesWithRating &^= 1 << id
	}
	return uint8(id)
}

// aiPlaystyleTownIDSet is AiPlaystyleFlags::townIdSet.
const aiPlaystyleTownIDSet = 1 << 8

// createCompetitorCompany is createCompany's competitor branch
// (CompanyManager.cpp 0x0042FE06, the else of isPlayer), as upstream draws
// gPrng1; false where upstream empties the name and returns
// CompanyId::null.
func (m *Map) createCompetitorCompany(id int, obj *objects.CompetitorObject, setScheme func(uint8)) bool {
	if obj == nil {
		return false
	}
	c := &m.Companies[id]
	pick := func(bits uint32, r uint32) uint8 {
		var avail []uint8
		for j := range uint8(32) {
			if bits&(1<<j) != 0 {
				avail = append(avail, j)
			}
		}
		if len(avail) == 0 {
			return 0
		}
		return avail[uint32(len(avail))*(r&0xFF)/256]
	}
	var randVal uint32
	var prefix, playstyle, primary uint8
	colourOK := false
	for range 250 {
		randVal = m.Rng1.RandNext()
		prefix = pick(obj.AvailableNamePrefixes, randVal)
		randVal = bits.RotateLeft32(randVal, -8)
		playstyle = pick(obj.AvailablePlayStyles, randVal)
		randVal = bits.RotateLeft32(randVal, -8)
		if int(prefix) >= len(aiPrimaryColours) || int(playstyle) >= len(companyAiPlaystyleFlags) {
			return false
		}
		primary = aiPrimaryColours[prefix]
		if primary == colourMax {
			primary = uint8((randVal & 0xFF) * 31 / 256)
			randVal = bits.RotateLeft32(randVal, -8)
		}
		if m.competingColourMask(id)&(1<<primary) != 0 {
			continue
		}
		colourOK = true
		break
	}
	if !colourOK {
		return false
	}
	setScheme(primary)
	c.AiPlaystyleFlags = companyAiPlaystyleFlags[playstyle]
	c.AiPlaystyleTownID = 0xFF
	if prefix == 12 { // company_ai_name_pop_string: named after a town
		var towns []int
		for i := range m.Towns {
			if m.Towns[i].Name != 0xFFFF {
				towns = append(towns, i)
			}
		}
		if len(towns) == 0 {
			return false
		}
		town := towns[(randVal&0xFF)*uint32(len(towns))/256]
		for o := range m.Companies {
			if o == id || m.CompanyEmpty(o) || m.Companies[o].AiPlaystyleFlags&aiPlaystyleTownIDSet != 0 { // as upstream: skips those WITH a town
				continue
			}
			if int(m.Companies[o].AiPlaystyleTownID) == town {
				return false
			}
		}
		c.AiPlaystyleTownID = uint8(town)
		c.AiPlaystyleFlags |= aiPlaystyleTownIDSet
	}
	args := []any{companyAiNamePrefixes[prefix], obj.LastName}
	if c.AiPlaystyleTownID != 0xFF {
		args = append(args, m.TownName(m.Towns[c.AiPlaystyleTownID].Name, int(c.AiPlaystyleTownID)))
	}
	name := formatNameString(companyAiPlaystyleString[playstyle], args)
	for o := range m.CompanyNames {
		if o != id && !m.CompanyEmpty(o) && m.CompanyName(o) == name {
			return false
		}
	}
	n := m.userStringAllocate(name, false)
	if n == 0 {
		return false
	}
	m.CompanyNames[id] = n
	return true
}

// formatNameString is StringManager::formatString for the company-name
// strings: {STRINGID} takes the next argument (a string id, formatted in
// turn with the arguments after it, or an object's string), {STRING} a
// string, {POP16} (and an object string's 0x88, its encoded form) skips one.
func formatNameString(id int, args []any) string {
	var b strings.Builder
	var format func(s string)
	next := func() any {
		if len(args) == 0 {
			return nil
		}
		a := args[0]
		args = args[1:]
		return a
	}
	format = func(s string) {
		for len(s) > 0 {
			switch {
			case strings.HasPrefix(s, "{STRINGID}"), strings.HasPrefix(s, "{STRING}"):
				s = s[strings.Index(s, "}")+1:]
				switch a := next().(type) {
				case int:
					format(i18n.Str(a))
				case string:
					format(a)
				}
			case strings.HasPrefix(s, "{POP16}"):
				s = s[len("{POP16}"):]
				next()
			case s[0] == 0x88:
				s = s[1:]
				next()
			default:
				b.WriteByte(s[0])
				s = s[1:]
			}
		}
	}
	format(i18n.Str(id))
	return b.String()
}

// StartScenario is Scenario::start's game-state part for a new game from a
// .SC5, in upstream's order: companies reset and the player company created
// (before the date, so its loan uses the file's inflation), initialiseDate
// for the scenario's start year (every company's unlocked vehicles
// redetermined), sub_4748D4's default level crossing, the objective's time
// limit and month count.
//
// PARTIAL: the season,
// initialiseDefaultTrackRoadMods. The
// scenario name / details / file name copy, the labels and the windows /
// viewport / music resets are cmd/goloco's side.
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
	m.InitialiseSnowLine(om)
	for id := range m.Companies {
		if !m.CompanyEmpty(id) {
			m.DetermineAvailableVehicles(id, om) // CompanyManager::determineAvailableVehicles
		}
	}
	// sub_4748D4 (Scenario::start, after the player company):
	m.UpdateRoadObjectIdFlags(om)
	m.UpdateDefaultLevelCrossingType(om)
	m.UpdateScenarioObjectOptions(om)
	m.UpdatePlayerInfrastructureOptions(om)
	// setRecords(kZeroRecords): no speed records, by no company, on no date.
	m.CompanyRecords = scenario.CompanyRecords{Company: [3]uint8{0xFF, 0xFF, 0xFF}}
	m.ObjectiveUntilYear = uint16(m.Objective.TimeLimitYears) - 1 + m.CurrentYear
	m.ObjectiveMonthsInChallenge = 0
	m.LastMapWindow.Flags = 0 // Scenario.cpp l. 366: lastMapWindowAttributes.flags = WindowFlags::none
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
