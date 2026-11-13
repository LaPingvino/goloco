package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// The scenario challenge — Company::getNewChallengeProgress and
// applyBeTopProgressModifiers (World/Company.cpp).
//
// PARTIAL: evaluateChallengeProgress (completed / failed flags, messages,
// the challenge window, saveNewScore) and the daily call that stores the
// result in challengeProgress wait for the day tick.

const (
	companyFlagUnk2               = 1 << 2
	companyFlagChallengeCompleted = 1 << 6
	companyFlagChallengeFailed    = 1 << 7
	companyFlagChallengeBeaten    = 1 << 8
	objectiveFlagBeTop            = 1 << 0
	objectiveFlagTopThree         = 1 << 1
	objectiveFlagWithinTimeLimit  = 1 << 2 // Scenario::ObjectiveFlags::withinTimeLimit
)

// CompanyEmpty is Company::empty (name == StringIds::empty).
func (m *Map) CompanyEmpty(id int) bool {
	return id >= len(m.CompanyNames) || m.CompanyNames[id] == 0
}

func (m *Map) applyBeTopProgressModifiers(progress int, id int) int {
	f := m.Objective.Flags
	if f&(objectiveFlagBeTop|objectiveFlagTopThree) == 0 {
		return progress
	}
	better := 0
	for o := range m.CompanyFaces {
		if o == id || m.CompanyEmpty(o) {
			continue
		}
		if m.CompanyFaces[id].Performance < m.CompanyFaces[o].Performance {
			better++
		}
	}
	if f&objectiveFlagBeTop != 0 {
		if better != 0 {
			return max(progress-10, 0)
		}
		return progress
	}
	if better >= 3 {
		return max(progress-10*min(better-2, 3), 0)
	}
	return progress
}

// NewChallengeProgress is Company::getNewChallengeProgress: percent done
// (100 complete, 255 failed).
func (m *Map) NewChallengeProgress(id int, om *objects.ObjectManager) uint8 {
	if id >= len(m.Companies) || id >= len(m.CompanyFaces) {
		return 0
	}
	c, flags := &m.Companies[id], m.CompanyFaces[id].ChallengeFlags
	if flags&companyFlagChallengeCompleted != 0 {
		return 100
	}
	if flags&(companyFlagChallengeFailed|companyFlagChallengeBeaten) != 0 {
		return 255
	}
	// 37d7c01: the objective's time limit (vanilla tested companyFlags unk2,
	// set only when AI-allocated items turn real), and it kicks in on the
	// 2nd (the day of month counts from 0), not after January.
	if m.Objective.Flags&objectiveFlagWithinTimeLimit != 0 && m.ObjectiveUntilYear < m.CurrentYear && m.CurrentDayOfMonth != 0 {
		return 255
	}
	o := m.Objective
	progress := 0
	switch o.Type {
	case 0: // companyValue
		if c.CompanyValueHistory[0] >= int64(int32(o.CompanyValue)) {
			return 100
		}
		return uint8(c.CompanyValueHistory[0] / int64(o.CompanyValue/100))
	case 1: // vehicleProfit
		if c.VehicleProfit <= 0 {
			return 0
		}
		if c.VehicleProfit >= int64(int32(o.MonthlyVehicleProfit)) {
			progress = 100
		} else {
			progress = int(c.VehicleProfit / int64(o.MonthlyVehicleProfit/100))
		}
	case 2: // performanceIndex
		pi := int(m.CompanyFaces[id].Performance)
		if int(o.PerformanceIndex)*10 <= pi {
			progress = 100
		} else {
			progress = pi * 10 / int(o.PerformanceIndex)
		}
	case 3: // cargoDelivery
		if o.DeliveredCargoType == 0xFF || om == nil || int(o.DeliveredCargoType) >= len(om.CargoObjects) || om.CargoObjects[o.DeliveredCargoType] == nil {
			return 0
		}
		q := c.CargoDelivered[o.DeliveredCargoType&31]
		if q >= o.DeliveredCargoAmount {
			progress = 100
		} else {
			progress = int(q / (o.DeliveredCargoAmount / 100))
		}
	}
	return uint8(m.applyBeTopProgressModifiers(uint8n(progress), id))
}

// uint8n mirrors the uint8_t progress local.
func uint8n(v int) int { return int(uint8(v)) }

// CompanyFlagChallengeCompleted and CompanyFlagChallengeFailed are
// CompanyFlags::challengeCompleted / challengeFailed (the Challenge tab).
const (
	CompanyFlagChallengeCompleted = companyFlagChallengeCompleted
	CompanyFlagChallengeFailed    = companyFlagChallengeFailed
)
