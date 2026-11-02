package sim

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// calculateDeliveredCargoPayment against upstream's own function (extracted
// from CompanyManager.cpp with getInflationAdjustedCost, compiled, run over
// cargo parameters × currency factors × units × distances × days:
// testdata/payment.txt).
func TestDeliveredCargoPaymentMatchesUpstream(t *testing.T) {
	cargos := []objects.CargoObject{
		{PremiumDays: 2, MaxNonPremiumDays: 8, NonPremiumRate: 5, PenaltyRate: 0x1000, PaymentFactor: 3000, PaymentIndex: 3},
		{PremiumDays: 6, MaxNonPremiumDays: 30, NonPremiumRate: 0x200, PenaltyRate: 0x1800, PaymentFactor: 4500, PaymentIndex: 6},
		{PremiumDays: 1, MaxNonPremiumDays: 2, NonPremiumRate: 0x7FF, PenaltyRate: 0x7FF, PaymentFactor: 900, PaymentIndex: 0},
		{PremiumDays: 20, MaxNonPremiumDays: 255, NonPremiumRate: 0x80, PenaltyRate: 0x100, PaymentFactor: 60000, PaymentIndex: 31},
	}
	f, err := os.Open("testdata/payment.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		var c int
		var factor uint32
		var units, dist, days, want int32
		if _, err := fmt.Sscan(sc.Text(), &c, &factor, &units, &dist, &days, &want); err != nil {
			t.Fatal(err)
		}
		var m Map
		for i := range m.CurrencyMultiplicationFactor {
			m.CurrencyMultiplicationFactor[i] = factor
		}
		if got := m.DeliveredCargoPayment(&cargos[c], units, dist, uint16(days)); got != want {
			t.Errorf("cargo %d factor %d units %d dist %d days %d: %d, upstream %d", c, factor, units, dist, days, got, want)
		}
		n++
	}
	if n != 864 {
		t.Errorf("%d cases", n)
	}
}
