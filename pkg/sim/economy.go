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
