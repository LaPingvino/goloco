package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// The last income run (Vehicle1 lastIncome): VehicleHead.cpp's
// IncomeStats::beginNewIncome (0x004BA7FC), addToStats (0x4BA817) and
// VehicleHead::updateLastIncomeStats (0x004BA7C7). What the vehicle window's
// Finances tab shows as "Last income on …".

// beginNewIncome: today's day, every cargo line empty.
func (m *Map) beginNewIncome(s *scenario.IncomeStats) {
	s.Day = int32(m.CurrentDay)
	s.CargoTypes = [4]uint8{0xFF, 0xFF, 0xFF, 0xFF}
}

// addToStats adds to the line of that cargo and distance, else takes an empty
// line; false when all four are taken by others.
func addToStats(s *scenario.IncomeStats, cargoType uint8, qty, dist uint16, age uint8, profit int32) bool {
	for i := 0; i < 4; i++ {
		if s.CargoTypes[i] != cargoType || s.CargoDistances[i] != dist {
			continue
		}
		s.CargoQtys[i] += qty
		s.CargoProfits[i] += profit
		s.CargoAges[i] = max(age, s.CargoAges[i])
		return true
	}
	for i := 0; i < 4; i++ {
		if s.CargoTypes[i] != 0xFF {
			continue
		}
		s.CargoTypes[i], s.CargoDistances[i], s.CargoQtys[i], s.CargoAges[i], s.CargoProfits[i] = cargoType, dist, qty, age, profit
		return true
	}
	return false
}

// updateLastIncomeStats: a delivery of qty (none: nothing) — the first since
// the vehicle was placed or last paid (Flags48::flag2) begins a new run.
// PARTIAL: the vehicle window's invalidation (goloco's windows redraw anyway).
func (m *Map) updateLastIncomeStats(tr Train, cargoType uint8, qty, dist uint16, age uint8, profit int32) {
	if qty == 0 || cargoType == 0xFF {
		return
	}
	v1 := &m.Entities[tr.Veh1]
	if v1.Var48&flags48Flag2 != 0 {
		v1.Var48 &^= flags48Flag2
		m.beginNewIncome(&v1.LastIncome)
	}
	addToStats(&v1.LastIncome, cargoType, qty, dist, age, profit)
}
