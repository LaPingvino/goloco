package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Economy — Economy.cpp getInflationAdjustedCost and CompanyManager.cpp
// calculateDeliveredCargoPayment.

// InflationAdjustedCost is getInflationAdjustedCost(costFactor, costIndex,
// divisor): costFactor × currencyMultiplicationFactor[costIndex] >> divisor
// (a 64-bit product divided, as upstream).
func (m *Map) InflationAdjustedCost(costFactor int16, costIndex, divisor uint8) int32 {
	val := int64(costFactor) * int64(m.CurrencyMultiplicationFactor[costIndex&31])
	return int32(val / (int64(1) << divisor))
}

// DeliveredCargoPayment is calculateDeliveredCargoPayment: the cargo's
// payment factor, reduced for days past its premium days (nonPremiumRate,
// up to maxNonPremiumDays) and beyond those (penaltyRate, at most 255
// days), inflation adjusted, × units × distance (tiles) / 4096. The 32-bit
// int arithmetic and the int16 cost factor are upstream's.
func (m *Map) DeliveredCargoPayment(c *objects.CargoObject, numUnits, distance int32, numDays uint16) int32 {
	pf := int32(c.PaymentFactor) << 16
	update := func(percentage int32) {
		pf = max(0, pf-int32(c.PaymentFactor)*percentage)
	}
	nonPremium := int32(numDays) - int32(c.PremiumDays)
	if nonPremium > 0 {
		update(int32(c.NonPremiumRate) * min(nonPremium, int32(c.MaxNonPremiumDays)))
		if penalty := min(255, nonPremium-int32(c.MaxNonPremiumDays)); penalty > 0 {
			update(int32(c.PenaltyRate) * penalty)
		}
	}
	pf >>= 16
	unit := int64(m.InflationAdjustedCost(int16(pf), c.PaymentIndex, 8))
	return int32(unit * int64(numUnits) * int64(distance) / 4096)
}

// ApplyPaymentToCompany is CompanyManager::applyPaymentToCompany: a payment
// (negative: income) comes off the company's cash and is booked under this
// month's expenditure type (Expenditures[0][type]).
func (m *Map) ApplyPaymentToCompany(id uint8, payment int32, expenditureType uint8) {
	if int(id) >= len(m.Companies) || expenditureType >= 17 {
		return
	}
	c := &m.Companies[id]
	c.Cash -= int64(payment)
	c.Expenditures[0][expenditureType] -= payment
}

// inflationFactors is Economy.cpp's kInflationFactors (monthly, /4096).
var inflationFactors = [32]uint32{20, 20, 20, 20, 23, 20, 23, 23, 20, 17, 17, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20}

// EconomyUpdateMonthly is Economy::updateMonthly (dateTick runs it while the
// year is ≤ 2029). PARTIAL: buildDeliveredCargoPaymentsTable (the company
// list's cargo payment graph) and the window invalidations.
func (m *Map) EconomyUpdateMonthly() {
	for i := range m.CurrencyMultiplicationFactor {
		f := &m.CurrencyMultiplicationFactor[i]
		*f += uint32(uint64(inflationFactors[i]) * uint64(*f) >> 12)
	}
}

// SetInflationForYear is Economy::setInflationForYear: the factors from 1024
// through one monthly update per month since 1900 (years clamped to
// 1900–2030). The unused factors are not kept.
func (m *Map) SetInflationForYear(year uint16) {
	for i := range m.CurrencyMultiplicationFactor {
		m.CurrencyMultiplicationFactor[i] = 1024
	}
	for n := (uint32(min(max(year, 1900), 2030)) - 1900) * 12; n > 0; n-- {
		m.EconomyUpdateMonthly()
	}
}
