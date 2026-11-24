package sim

import "github.com/LaPingvino/goloco/pkg/i18n"

import "github.com/LaPingvino/goloco/pkg/objects"

import "github.com/LaPingvino/goloco/pkg/olmath"

// Company performance and value — World/Company.cpp calculatePerformanceIndex
// and calculateCompanyValue, olmath.FastSquareRoot (Math::Vector::fastSquareRoot).
//
// Not checkable against a save: Company::updateMonthly1 computes both BEFORE
// VehicleManager::updateMonthly rotates the vehicles' profit history (dateTick
// order), so a save holds results of a profit history it no longer has. Ported
// line for line; wired into the monthly update next (item 5 monthly tick).

const companyFlagBankrupt = 1 << 9

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
	partial := uint32(olmath.FastSquareRoot(uint32(total))) * 75
	eco := uint32(olmath.FastSquareRoot(m.CurrencyMultiplicationFactor[0]))
	profitFactor := uint16(500)
	if partial < eco*500 {
		profitFactor = uint16(partial / eco)
	}
	cargoFactor := min(500, olmath.FastSquareRoot(m.Companies[id].CargoUnitsDistanceHist[0]/4))
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

const (
	companyFlagIncreasedPerformance = 1 << 4
	companyFlagDecreasedPerformance = 1 << 5
	aiLoanStep                      = 1000
	expenditureLoanInterest         = 15
	companyIDNeutral                = 15
)

// IsPlayerCompany is CompanyManager::isPlayerCompany.
func (m *Map) IsPlayerCompany(id int) bool {
	return id == int(m.PlayerCompanies[0]) || id == int(m.PlayerCompanies[1])
}

// ensureCompanyFunding is CompanyManager::ensureCompanyFunding (outside the
// editor): a player needs the cash, and so does a company run by one of
// goloco's AIs without AIOptions.AutoBorrow (AI-INTERFACE.md G6); any other
// AI company takes a loan in 1000 steps up to the inflation-adjusted maximum. A refusal sets the command's error text
// (cmd may be nil: none kept): company_is_bankrupt, or
// not_enough_cash_requires_currency32 with the payment.
func (m *Map) ensureCompanyFunding(cmd *Cmd, id int, payment int32) bool {
	if payment <= 0 || id == companyIDNeutral {
		return true
	}
	refuse := func(text int) bool {
		if cmd != nil {
			cmd.ErrorText, cmd.ErrorCost = text, payment
		}
		return false
	}
	c := &m.Companies[id]
	if m.IsPlayerCompany(id) || m.aiPaysAsPlayer(id) { // goloco's AIs without AutoBorrow pay as players (companyai.go, G6)
		if m.CompanyFaces[id].ChallengeFlags&companyFlagBankrupt != 0 {
			return refuse(i18n.CompanyIsBankrupt)
		}
		if c.Cash < int64(payment) {
			return refuse(i18n.NotEnoughCashRequiresCurrency32)
		}
		return true
	}
	if c.Cash < int64(payment) {
		if need := -(c.Cash - int64(payment)); need > 0 {
			loan := (need + aiLoanStep - 1) / aiLoanStep * aiLoanStep
			if loan+int64(c.CurrentLoan) > int64(m.InflationAdjustedCost(int16(m.MaxLoanSize), 0, 8)) {
				return refuse(i18n.NotEnoughCashRequiresCurrency32)
			}
			c.CurrentLoan += int32(loan)
			c.Cash += loan
		}
	}
	return true
}

func rotate120[T any](h *[120]T, v T) {
	copy(h[1:], h[:119])
	h[0] = v
}

// companiesUpdateMonthly1 is CompanyManager::updateMonthly1 → each company's
// Company::updateMonthly1 (the controlling company's promotion posts
// companyPromoted), with the owner's emotions: angry at two months of falling
// deliveries, happy or scared as the performance rises or falls, worried
// without cash, scared at the bankruptcy warnings, dejected at bankruptcy —
// and, of more than one company, the lowest performer dejected; first, the
// competitor start delay counts down (competitorsUpdateMonthly1). PARTIAL:
// the Ui invalidations are not ported. The bankruptcy news are posted (warnings at 3 and 6 months in
// the red for the controlling company, declared at 9), and a bankruptcy
// stops the company's vehicles (stopAllCompanyVehicles).
func (m *Map) companiesUpdateMonthly1(om *objects.ObjectManager) {
	m.competitorsUpdateMonthly1()
	for id := range m.Companies {
		if id >= len(m.CompanyFaces) || m.CompanyEmpty(id) {
			continue
		}
		c, face := &m.Companies[id], &m.CompanyFaces[id]
		rotate120(&c.CargoUnitsDeliveredHist, c.CargoUnitsTotalDelivered)
		c.CargoUnitsTotalDelivered = 0
		rotate120(&c.CargoUnitsDistanceHist, c.CargoUnitsTotalDistance)
		c.CargoUnitsTotalDistance = 0
		if c.HistorySize >= 2 && c.CargoUnitsDeliveredHist[0] < c.CargoUnitsDeliveredHist[1] && c.CargoUnitsDeliveredHist[1] < c.CargoUnitsDeliveredHist[2] {
			m.companyEmotionEvent(id, emotionAngry)
		}
		perf := m.CalculatePerformanceIndex(id)
		face.ChallengeFlags &^= companyFlagIncreasedPerformance | companyFlagDecreasedPerformance
		if perf < face.Performance {
			face.ChallengeFlags |= companyFlagDecreasedPerformance
		} else if perf > face.Performance {
			face.ChallengeFlags |= companyFlagIncreasedPerformance
		}
		rotate120(&c.PerformanceIndexHist, perf)
		face.Performance = perf
		if rating := uint8(min(9, perf/100)); rating > c.CurrentRating { // performanceToRating
			c.CurrentRating = rating
			if uint8(id) == m.ControllingCompany {
				m.PostMessage(nil, MsgCompanyPromoted, 0xFF, uint16(id), uint16(rating), 0xFFFF)
			}
		}
		if face.ChallengeFlags&companyFlagIncreasedPerformance != 0 {
			m.companyEmotionEvent(id, emotionHappy)
		}
		if face.ChallengeFlags&companyFlagDecreasedPerformance != 0 {
			m.companyEmotionEvent(id, emotionScared)
		}
		vp, value := m.CalculateCompanyValue(id)
		rotate120(&c.CompanyValueHistory, value)
		c.VehicleProfit = vp
		c.HistorySize = min(c.HistorySize+1, 120)
		if !m.IsPlayerCompany(id) && !m.aiPaysAsPlayer(id) { // AI: auto pay the loan (not goloco's AIs that borrow as players: G6)
			for c.CurrentLoan >= 1000 && c.Cash >= 1000 {
				c.CurrentLoan -= 1000
				c.Cash -= 1000
			}
		}
		interest := c.CurrentLoan * int32(m.LoanInterestRate) / 1200
		m.ensureCompanyFunding(nil, id, interest+1)
		m.ApplyPaymentToCompany(uint8(id), interest, expenditureLoanInterest)
		if c.Cash <= 0 {
			m.companyEmotionEvent(id, emotionWorried)
		}
		if face.ChallengeFlags&companyFlagBankrupt == 0 {
			if c.Cash < 0 {
				c.NumMonthsInTheRed++
				if c.NumMonthsInTheRed == 9 {
					msg := MsgBankruptcyDeclaredCompetitor
					if uint8(id) == m.ControllingCompany {
						msg = MsgBankruptcyDeclared
					}
					m.PostMessage(nil, msg, uint8(id), uint16(id), 0xFFFF, 0xFFFF) // the text needs no objects
					face.ChallengeFlags |= companyFlagBankrupt
					if face.ChallengeFlags&(companyFlagChallengeBeaten|companyFlagChallengeCompleted) == 0 {
						face.ChallengeFlags |= companyFlagChallengeFailed
					}
					m.companyEmotionEvent(id, emotionDejected)
					m.stopAllCompanyVehicles(uint8(id), om)
				}
				if uint8(id) == m.ControllingCompany {
					switch c.NumMonthsInTheRed {
					case 3:
						m.PostMessage(nil, MsgBankruptcyWarning6MonthsRemaining, uint8(id), uint16(id), 0xFFFF, 0xFFFF)
						m.companyEmotionEvent(id, emotionScared)
					case 6:
						m.PostMessage(nil, MsgBankruptcyWarning3MonthsRemaining, uint8(id), uint16(id), 0xFFFF, 0xFFFF)
						m.companyEmotionEvent(id, emotionScared)
					}
				}
			} else {
				c.NumMonthsInTheRed = 0
			}
		}
	}
	// the lowest performance (the first of equals) is dejected, unless alone
	low, n := -1, 0
	for id := range m.Companies {
		if id >= len(m.CompanyFaces) || m.CompanyEmpty(id) {
			continue
		}
		n++
		if low < 0 || m.CompanyFaces[id].Performance < m.CompanyFaces[low].Performance {
			low = id
		}
	}
	if low >= 0 && n != 1 {
		m.companyEmotionEvent(low, emotionDejected)
	}
}

// stopAllCompanyVehicles (Company.cpp): as the company, every non-ghost
// vehicle of it stopped (vehicleChangeRunningMode stopVehicle, applied).
func (m *Map) stopAllCompanyVehicles(company uint8, om *objects.ObjectManager) {
	trains, _ := m.Trains()
	c := &Cmd{Company: company, Objects: om}
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Owner != company || h.Flags38&flags38IsGhost != 0 {
			continue
		}
		m.VehicleChangeRunningMode(c, CmdApply, h.ID, RunningModeStop)
	}
}
