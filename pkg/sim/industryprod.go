package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Industries produce (item CP): Industry::updateDaily (World/Industry.cpp
// 0x004534BD), from dateTick's IndustryManager::updateDaily — the cargo it
// received turned into output (all its required cargoes at once with
// requiresAllCargo, else their sum), a sixteenth of what's left lost, the
// daily production moving one step towards its target × productionRate /
// 256, and at 8 or more in the output buffer up to 255 sent to the stations
// its produced cargo goes to (producedCargoStatsStation).
// PARTIAL: Industry::update's farm tiles, updateMonthly's production changes.

const industryFlagIsGhost = 1 << 0 // IndustryFlags::isGhost

func boundAddU16(a, b uint16) uint16 { return uint16(min(uint32(a)+uint32(b), 0xFFFF)) }

// industriesUpdateDaily is IndustryManager::updateDaily.
func (m *Map) industriesUpdateDaily(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	for i := range m.Industries {
		m.industryUpdateDaily(&m.Industries[i], om)
	}
}

// industryUpdateDaily is Industry::updateDaily.
func (m *Map) industryUpdateDaily(ind *scenario.Industry, om *objects.ObjectManager) {
	if ind.Flags&industryFlagIsGhost != 0 {
		return
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	produced, required := om.IndustryCargoIDs(obj)
	var amount uint16
	if obj.Flags&objects.IndustryFlagRequiresAllCargo != 0 {
		amount = 0xFFFF
		for i := 0; i < 3; i++ {
			if required[i] != 0xFF {
				amount = min(amount, ind.ReceivedDailyTotal[i])
			}
		}
		if amount != 0 {
			for i := 0; i < 3; i++ {
				if required[i] != 0xFF {
					ind.ReceivedDailyTotal[i] -= amount
				}
			}
		}
	} else {
		for i := 0; i < 3; i++ {
			if required[i] != 0xFF {
				amount = boundAddU16(amount, ind.ReceivedDailyTotal[i])
				ind.ReceivedDailyTotal[i] = 0
			}
		}
	}
	if amount != 0 {
		for i := 0; i < 2; i++ {
			if produced[i] != 0xFF {
				ind.OutputBuffer[i] = boundAddU16(ind.OutputBuffer[i], amount)
			}
		}
	}
	for i := 0; i < 3; i++ { // a sixteenth of the required cargo left lost
		ind.ReceivedDailyTotal[i] -= boundAddU16(ind.ReceivedDailyTotal[i], 15) / 16
	}
	for i := 0; i < 2; i++ {
		if produced[i] == 0xFF {
			continue
		}
		target := uint16(uint32(ind.DailyProductionTarget[i]) * uint32(ind.ProductionRate) / 256)
		switch {
		case target < ind.DailyProduction[i]:
			ind.DailyProduction[i]--
		case target > ind.DailyProduction[i]:
			ind.DailyProduction[i]++
		}
		ind.OutputBuffer[i] = boundAddU16(ind.DailyProduction[i], ind.OutputBuffer[i])
		if ind.OutputBuffer[i] < 8 {
			continue
		}
		send := min(ind.OutputBuffer[i], 255)
		ind.OutputBuffer[i] -= send
		ind.ProducedMonthlyTotal[i] = boundAddU16(ind.ProducedMonthlyTotal[i], send)
		var found []stationShare // deliverCargoToStations(stations): each with its rating for the cargo
		for _, sid := range ind.ProducedStations[i] {
			if sid != stationIDNull && int(sid) < len(m.StationNames) {
				found = append(found, stationShare{sid, m.StationNames[sid].Cargo[produced[i]].Rating})
			}
		}
		var delivered uint16
		if len(found) > 0 {
			delivered = m.deliverCargoToStations(found, produced[i], uint8(send), om)
		}
		ind.DeliveredMonthlyTotal[i] = boundAddU16(delivered, ind.DeliveredMonthlyTotal[i])
	}
}
