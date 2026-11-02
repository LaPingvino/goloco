package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// The year's end — CompanyManager::updateYearly (Company.cpp updateYearly),
// ObjectManager::updateDefaultLevelCrossingType and updateYearly2.
//
// PARTIAL: TileManager::updateYearly; in updateYearly2 the availableVehicles
// recalculation (determineAvailableVehicleTypes), the newVehicle message and
// the construction window's road / rail options.

const expenditureHistoryCapacity = 16

// companiesUpdateYearly: each company's expenditure table moves on a year
// (the oldest year drops off).
func (m *Map) companiesUpdateYearly() {
	for id := range m.Companies {
		if m.CompanyEmpty(id) {
			continue
		}
		c := &m.Companies[id]
		for t := range c.Expenditures[0] {
			var last int32
			for i := range expenditureHistoryCapacity {
				last, c.Expenditures[i][t] = c.Expenditures[i][t], last
			}
		}
		c.NumExpenditureYears = min(c.NumExpenditureYears+1, expenditureHistoryCapacity)
	}
}

// UpdateDefaultLevelCrossingType is updateDefaultLevelCrossingType: the
// newest level crossing designed by now (else the newest of all) for crossings
// built from now on.
func (m *Map) UpdateDefaultLevelCrossingType(om *objects.ObjectManager) {
	best, bestYear := 0, int32(-1<<31)
	pick := func(onlyDesigned bool) {
		for i, o := range om.LevelCrossingObjects {
			if o == nil || (onlyDesigned && int32(m.CurrentYear) < int32(o.DesignedYear)) || bestYear > int32(o.DesignedYear) {
				continue
			}
			bestYear, best = int32(o.DesignedYear), i
		}
	}
	pick(true)
	if bestYear == -1<<31 {
		pick(false)
	}
	m.DefaultLevelCrossingType = uint8(best)
}

// updateYearly2: vehicles designed this year are unlocked for every company
// allowed their type; vehicles going obsolete this year are locked for all.
func (m *Map) updateYearly2(om *objects.ObjectManager) {
	for vid, v := range om.Vehicles {
		if v == nil {
			continue
		}
		if m.CurrentYear == v.Designed {
			for id := range m.Companies {
				if m.CompanyEmpty(id) || id >= len(m.UnlockedVehicles) {
					continue
				}
				forbidden := m.ForbiddenVehiclesCompetitors
				if m.IsPlayerCompany(id) {
					forbidden = m.ForbiddenVehiclesPlayers
				}
				if forbidden&(1<<v.Type) != 0 {
					continue
				}
				m.UnlockedVehicles[id][vid/32] |= 1 << (vid % 32)
			}
		}
		if m.CurrentYear == v.Obsolete {
			for id := range m.UnlockedVehicles {
				if !m.CompanyEmpty(id) {
					m.UnlockedVehicles[id][vid/32] &^= 1 << (vid % 32)
				}
			}
		}
	}
}
