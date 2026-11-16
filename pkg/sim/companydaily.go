package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Company::updateDaily (Company.cpp): the owner's face from the emotions in
// play, the emotions a day shorter, the day's observation (what the company
// window says the owner is doing), the observation and jail a day shorter,
// then the player's and the controlling player's daily parts.
//
// PARTIAL: the window invalidations (company, info panel, lists, news) are
// left to the windows, which draw from the state each frame.

// ObservationStatus values (Company.h).
const (
	observationCheckingServices   uint8 = 4
	observationSurveyingLandscape uint8 = 5
)

// aiThinkStateUnk3 is AiThinkState::unk3 (CompanyAi.h).
const aiThinkStateUnk3 = 3

// emotionWeightings is Company.cpp's: which active emotion shows.
var emotionWeightings = [9]uint8{0, 4, 6, 5, 3, 8, 1, 2, 7}

// Emotion values (Company.h).
const (
	emotionHappy     = 1
	emotionWorried   = 2
	emotionDejected  = 4 // also a bankrupt owner's face
	emotionSurprised = 5
	emotionScared    = 6
	emotionAngry     = 7
)

// kEmotionDurations (0x004F9462): the days an emotion lasts.
var kEmotionDurations = [9]uint8{0, 31, 10, 7, 31, 10, 31, 31, 11}

// companyEmotionEvent is Company.cpp's (0x00437F29): the emotion for its days.
func (m *Map) companyEmotionEvent(id int, emotion int) {
	m.Companies[id].ActiveEmotions[emotion] = kEmotionDurations[emotion]
}

// thoughtDestinationAIsIndustry is kThoughtTypeFlags' destinationAIsIndustry
// per AI thought type (CompanyAi.cpp 0x004FE720): destinationA is an
// industry, else a town.
var thoughtDestinationAIsIndustry = [20]bool{7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 14: true, 16: true, 17: true, 18: true, 19: true}

func (m *Map) companyUpdateDaily(id int, om *objects.ObjectManager) {
	c := &m.Companies[id]
	m.updateOwnerEmotion(id)
	for i := range c.ActiveEmotions {
		if c.ActiveEmotions[i] > 0 {
			c.ActiveEmotions[i]--
		}
	}
	m.companyUpdateDailyLogic(id)
	if c.ObservationTimeout > 0 {
		c.ObservationTimeout--
	}
	if c.JailStatus > 0 {
		c.JailStatus--
	}
	if m.IsPlayerCompany(id) {
		m.companyUpdateDailyPlayer(id, om)
	}
	if id == int(m.ControllingCompany) {
		m.updateLoanAutorepay(id) // updateDailyControllingPlayer
	}
}

// updateOwnerEmotion is Company::updateOwnerEmotion: the heaviest emotion in
// play (neutral without one), dejected when bankrupt.
func (m *Map) updateOwnerEmotion(id int) {
	if id >= len(m.CompanyFaces) {
		return
	}
	f := &m.CompanyFaces[id]
	newEmotion := uint8(emotionDejected)
	if f.ChallengeFlags&companyFlagBankrupt == 0 {
		newEmotion = 0
		weight := uint8(0)
		for e, d := range m.Companies[id].ActiveEmotions {
			if d != 0 && weight <= emotionWeightings[e] {
				weight, newEmotion = emotionWeightings[e], uint8(e)
			}
		}
	}
	f.Emotion = newEmotion
}

// companyUpdateDailyLogic is Company::updateDailyLogic: once the last
// observation has run out, a player's owner watches what its ownerStatus
// names (a vehicle, or a place it surveys); an AI's: setAiObservation.
func (m *Map) companyUpdateDailyLogic(id int) {
	c := &m.Companies[id]
	if !m.IsPlayerCompany(id) {
		m.setAiObservation(id)
		return
	}
	if c.ObservationTimeout != 0 || c.OwnerStatus[0] == -1 { // isEmpty
		return
	}
	if c.OwnerStatus[0] != -2 { // a position
		m.companySetObservation(id, observationSurveyingLandscape, c.OwnerStatus[0], c.OwnerStatus[1], entityIDNull, 0xFFFF)
		return
	}
	m.observeTrain(id, uint16(c.OwnerStatus[1])) // no train: upstream asserts and returns
}

// setAiObservation is CompanyAi.cpp's (0x0043821D): an AI building a thought
// surveys its first destination; otherwise, once the last observation has
// run out, it checks on one of its vehicles picked from gPrng1.
func (m *Map) setAiObservation(id int) {
	c := &m.Companies[id]
	if c.AiThinkState == aiThinkStateUnk3 {
		t := int(c.ActiveThoughtID)
		if t >= len(c.ThoughtType) {
			return
		}
		var x, y int16
		if int(c.ThoughtType[t]) < len(thoughtDestinationAIsIndustry) && thoughtDestinationAIsIndustry[c.ThoughtType[t]] {
			if ind := m.industryByID(c.ThoughtDestinationA[t]); ind != nil {
				x, y = ind.X, ind.Y
			}
		} else if town := m.TownByID(uint16(c.ThoughtDestinationA[t])); town != nil {
			x, y = town.X, town.Y
		}
		m.companySetObservation(id, observationSurveyingLandscape, x, y, entityIDNull, 0xFFFF)
		return
	}
	if c.ObservationTimeout != 0 {
		return
	}
	total := uint64(0)
	for _, n := range c.TransportTypeCount {
		total += uint64(n)
	}
	pick := total * uint64(m.Rng1.RandNext()) >> 32
	trains, err := m.Trains()
	if err != nil {
		return
	}
	i := uint64(0)
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if int(h.Owner) != id || h.Flags38&flags38IsGhost != 0 {
			continue
		}
		if i == pick {
			m.observeTrain(id, h.ID)
			return
		}
		i++
	}
}

// observeTrain is the checkingServices observation of the train with this
// head entity id, at its Vehicle2 (none while it isn't placed).
func (m *Map) observeTrain(id int, head uint16) {
	tr, err := m.TrainOf(head)
	if err != nil {
		return
	}
	v2 := &m.Entities[tr.Veh2]
	if v2.PosX == locationNull {
		return
	}
	m.companySetObservation(id, observationCheckingServices, int16(v2.PosX), int16(v2.PosY), head, 0xFFFF)
}

// companySetObservation is Company.cpp's (0x00438167): a new observation
// (shouldSetObservation) takes the place, entity, object and the closest
// town; either way it lasts 5 days.
func (m *Map) companySetObservation(id int, status uint8, x, y int16, entity, object uint16) {
	c := &m.Companies[id]
	if shouldSetObservation(c.ObservationTimeout, c.ObservationStatus, c.ObservationX, c.ObservationY, c.ObservationEntity, c.ObservationObject, status, x, y, entity, object) {
		c.ObservationX, c.ObservationY = x, y
		c.ObservationEntity, c.ObservationObject, c.ObservationStatus = entity, object, status
		if t, _, ok := m.ClosestTownAndDensity(int32(x), int32(y)); ok {
			c.ObservationTownID = uint16(t)
		}
	}
	c.ObservationTimeout = 5
}

// shouldSetObservation is Company.cpp's: always once the last one ran out;
// a survey never replaces another kind; else when anything differs.
func shouldSetObservation(timeout uint16, curStatus uint8, curX, curY int16, curEntity, curObject uint16, status uint8, x, y int16, entity, object uint16) bool {
	if timeout == 0 {
		return true
	}
	if status == observationSurveyingLandscape && curStatus != observationSurveyingLandscape {
		return false
	}
	return status != curStatus || x != curX || y != curY || object != curObject || entity != curEntity
}

// companiesUpdateQuarterly is CompanyManager::updateQuarterly → Company::
// updateQuarterly (0x00487FCC), at the start of January, April, July and
// October: each AI thought a quarter older, its income kept as the previous
// quarter's and counted afresh, its running cost its vehicles' now.
func (m *Map) companiesUpdateQuarterly(om *objects.ObjectManager) {
	for id := range m.Companies {
		if m.CompanyEmpty(id) {
			continue
		}
		c := &m.Companies[id]
		for k := range c.ThoughtType {
			if c.ThoughtType[k] == 0xFF { // AiThoughtType::null
				continue
			}
			c.ThoughtAge[k] = uint8(min(0xFF, int(c.ThoughtAge[k])+1))
			c.ThoughtPrevIncome[k], c.ThoughtIncome[k] = c.ThoughtIncome[k], 0
			total := int32(0)
			for v := range min(int(c.ThoughtNumVehicles[k]), 8) {
				if tr, err := m.TrainOf(c.ThoughtVehicles[k][v]); err == nil {
					total += m.calculateRunningCost(tr, om)
				}
			}
			c.ThoughtRunningCost[k] = total
		}
	}
}
