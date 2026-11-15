package sim

import (
	"math/bits"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// townsUpdateMonthly is TownManager::updateMonthly: each town's
// Town::updateMonthly.
func (m *Map) townsUpdateMonthly() {
	for i := range m.Towns {
		if m.Towns[i].Name == 0xFFFF { // an empty town slot
			continue
		}
		townUpdateMonthly(&m.Towns[i])
	}
}

// townUpdateMonthly is Town::updateMonthly (Town.cpp): the population
// history scrolled and a new point written (populations above 255 steps of
// 50 folded into historyMinPopulation, and back while the history is low),
// then buildSpeed from the least delivered of the town's influence cargoes
// this month (1 slow … 4 fast; no influence cargo: as fast as 1 + 65535/100,
// clamped to 4), and the month's deliveries cleared.
func townUpdateMonthly(t *scenario.Town) {
	n := len(t.History)
	if int(t.HistorySize) == n {
		copy(t.History[:n-1], t.History[1:])
	} else {
		t.HistorySize++
	}
	popSteps := uint32(max(int32(t.Population)-t.HistoryMinPopulation, 0)) / 50
	popGrowth := uint32(0)
	for popSteps > 255 {
		popSteps -= 20
		popGrowth += 1000
	}
	if popGrowth != 0 {
		t.HistoryMinPopulation += int32(popGrowth)
		offset := uint8(popGrowth / 50)
		for i := 0; i < int(t.HistorySize); i++ {
			v := int16(t.History[i]) - int16(offset)
			t.History[i] = uint8(max(v, 0))
		}
	}
	histIndex := min(max(int(t.HistorySize)-1, 0), n-1) // std::clamp(historySize − 1, 0, size) — an index of 240 can't be written
	t.History[histIndex] = uint8(popSteps)
	maxPop := uint8(0)
	for i := 0; i < int(t.HistorySize); i++ {
		maxPop = max(maxPop, t.History[i])
	}
	popOffset := t.HistoryMinPopulation
	for maxPop <= 235 && popOffset > 0 {
		maxPop += 20
		popOffset -= 1000
	}
	popOffset -= t.HistoryMinPopulation
	if popOffset != 0 {
		popOffset = -popOffset
		t.HistoryMinPopulation -= popOffset
		popOffset /= 50
		for i := 0; i < int(t.HistorySize); i++ {
			t.History[i] += uint8(popOffset)
		}
	}
	minDelivered := uint16(0xFFFF)
	for flags := t.CargoInfluenceFlags; flags != 0; flags &= flags - 1 {
		minDelivered = min(minDelivered, t.MonthlyCargoDelivered[bits.TrailingZeros32(flags)])
	}
	t.BuildSpeed = uint8(min(max(int(minDelivered)/100+1, 1), 4))
	t.MonthlyCargoDelivered = [32]uint16{}
}
