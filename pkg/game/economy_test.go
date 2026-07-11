package game

import (
	"testing"
	"time"
)

// advanceDays ticks the game state forward by n game-days worth of ticks,
// exercising the real Update path (date advance + month rollover).
func advanceDays(gs *GameState, n int) {
	for i := 0; i < n*int(gs.TicksPerDay); i++ {
		gs.Update()
	}
}

func TestCreditDeliveryFares(t *testing.T) {
	gs := NewGameState(4, 4)
	start := gs.PlayerMoney

	gs.CreditDelivery(0, 10)  // passengers: 6 each = 60
	gs.CreditDelivery(11, 10) // passengers (base-game PASS slot): 6 each = 60
	gs.CreditDelivery(1, 5)   // mail: 8 each = 40
	gs.CreditDelivery(7, 3)   // other: 10 each = 30

	wantIncome := int64(60 + 60 + 40 + 30)
	if gs.MonthlyIncome != wantIncome {
		t.Errorf("MonthlyIncome = %d, want %d", gs.MonthlyIncome, wantIncome)
	}
	if got := gs.PlayerMoney - start; got != wantIncome {
		t.Errorf("PlayerMoney delta = %d, want %d", got, wantIncome)
	}
}

func TestMonthRolloverSnapshotsProfit(t *testing.T) {
	gs := NewGameState(4, 4)
	gs.GameDate = time.Date(1950, 1, 28, 0, 0, 0, 0, time.UTC)
	gs.VehicleCount = func() int { return 2 } // 2 * £120 = £240 running cost

	// Earn £600 (100 passengers) during January.
	gs.CreditDelivery(0, 100)
	if gs.MonthlyIncome != 600 {
		t.Fatalf("MonthlyIncome = %d, want 600", gs.MonthlyIncome)
	}

	start := gs.PlayerMoney

	// Advance into February to trigger exactly one rollover.
	advanceDays(gs, 5)

	if gs.GameDate.Month() != time.February {
		t.Fatalf("GameDate month = %v, want February", gs.GameDate.Month())
	}
	// Profit = income 600 - expenses (2 vehicles * 120) = 360.
	if gs.LastMonthProfit != 360 {
		t.Errorf("LastMonthProfit = %d, want 360", gs.LastMonthProfit)
	}
	// Accumulators reset for the new month.
	if gs.MonthlyIncome != 0 || gs.MonthlyExpenses != 0 {
		t.Errorf("accumulators not reset: income=%d expenses=%d",
			gs.MonthlyIncome, gs.MonthlyExpenses)
	}
	// Money should have dropped by the £240 running cost since the snapshot.
	if got := start - gs.PlayerMoney; got != 240 {
		t.Errorf("running cost charged = %d, want 240", got)
	}
}

func TestMonthRolloverNoVehicles(t *testing.T) {
	gs := NewGameState(4, 4)
	gs.GameDate = time.Date(1950, 3, 30, 0, 0, 0, 0, time.UTC)
	// VehicleCount nil: no running cost charged.

	gs.CreditDelivery(1, 10) // mail: 80
	start := gs.PlayerMoney

	advanceDays(gs, 3) // cross into April

	if gs.GameDate.Month() != time.April {
		t.Fatalf("GameDate month = %v, want April", gs.GameDate.Month())
	}
	if gs.LastMonthProfit != 80 {
		t.Errorf("LastMonthProfit = %d, want 80", gs.LastMonthProfit)
	}
	if gs.PlayerMoney != start {
		t.Errorf("PlayerMoney changed with no vehicles: %d != %d", gs.PlayerMoney, start)
	}
}

func TestProfitObjectiveWinnable(t *testing.T) {
	gs := NewGameState(4, 4)
	gs.GameDate = time.Date(1950, 6, 29, 0, 0, 0, 0, time.UTC)
	gs.VehicleCount = func() int { return 3 }

	// Deliver enough to beat a £5000/month objective after running costs
	// (3 * 120 = 360). Need income >= 5360; 900 passengers * 6 = 5400.
	gs.CreditDelivery(0, 900)
	advanceDays(gs, 3) // into July

	const objective = 5000
	if gs.LastMonthProfit < objective {
		t.Errorf("LastMonthProfit = %d, expected >= %d (objective unwinnable)",
			gs.LastMonthProfit, objective)
	}
	if gs.LastMonthProfit != 5400-360 {
		t.Errorf("LastMonthProfit = %d, want %d", gs.LastMonthProfit, 5400-360)
	}
}
