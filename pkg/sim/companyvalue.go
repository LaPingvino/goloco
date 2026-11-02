package sim

// Company performance and value — World/Company.cpp calculatePerformanceIndex
// and calculateCompanyValue, Math::Vector::fastSquareRoot (table generated,
// checked against 2000 upstream vectors).
//
// Not checkable against a save: Company::updateMonthly1 computes both BEFORE
// VehicleManager::updateMonthly rotates the vehicles' profit history (dateTick
// order), so a save holds results of a profit history it no longer has. Ported
// line for line; wired into the monthly update next (item 5 monthly tick).

const companyFlagBankrupt = 1 << 9

// FastSquareRoot is Math::Vector::fastSquareRoot.
func FastSquareRoot(d uint32) uint16 {
	i := uint(10)
	for ; d >= 4096; i, d = i-1, d>>2 {
	}
	return fastSquareRootTable[(d&0xFFE)>>1] >> i
}

// totalRecentProfit is Vehicle2::totalRecentProfit.
func totalRecentProfit(p [4]int32) int32 { return p[0] + p[1] + p[2] + p[3] }

// CalculatePerformanceIndex is calculatePerformanceIndex.
func (m *Map) CalculatePerformanceIndex(id int) int16 {
	if m.CompanyFaces[id].ChallengeFlags&companyFlagBankrupt != 0 {
		return 0
	}
	var total int32
	trains, _ := m.Trains()
	for _, tr := range trains {
		if int(m.Entities[tr.Head].Owner) != id {
			continue
		}
		p := totalRecentProfit(m.Entities[tr.Veh2].Profit) >> 1
		if int64(total)+int64(p) < 0x7FFFFFFF {
			total += p
		}
	}
	total = max(0, total)
	partial := uint32(FastSquareRoot(uint32(total))) * 75
	eco := uint32(FastSquareRoot(m.CurrencyMultiplicationFactor[0]))
	profitFactor := uint16(500)
	if partial < eco*500 {
		profitFactor = uint16(partial / eco)
	}
	cargoFactor := min(500, FastSquareRoot(m.Companies[id].CargoUnitsDistanceHist0/4))
	return int16(profitFactor + cargoFactor)
}

// CalculateCompanyValue is calculateCompanyValue: (vehicle profit, value).
func (m *Map) CalculateCompanyValue(id int) (vehicleProfit, value int64) {
	if m.CompanyFaces[id].ChallengeFlags&companyFlagBankrupt != 0 {
		return 0, 0
	}
	c := &m.Companies[id]
	value = c.Cash - int64(c.CurrentLoan)
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if int(h.Owner) != id || h.Flags38&flags38IsGhost != 0 {
			continue
		}
		p := totalRecentProfit(m.Entities[tr.Veh2].Profit)
		vehicleProfit += int64(p >> 2)
		value += int64(p) * 8
		for _, car := range tr.Cars {
			value += int64(m.Entities[car[0].Front].RefundCost)
		}
	}
	return vehicleProfit, max(0, value)
}
