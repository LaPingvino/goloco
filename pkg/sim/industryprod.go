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

// industryElementStations is the station-range part of IndustryElement::update
// (Map/IndustryElement.cpp): from a finished industry's building (its first
// tile, not a ghost), every station element within 4 tiles (5 on the far
// sides of a 2×2 building), not ghost or AI-allocated, marks its station in
// the industry's stationsInRange. PARTIAL: the element's construction and
// animation parts.
func (m *Map) industryElementStations(el *scenario.IndustryElement, x, y int32, om *objects.ObjectManager) {
	if el.SequenceIndex != 0 || om == nil {
		return
	}
	ind := m.industryByID(el.IndustryID)
	if ind == nil || ind.UnderConstruction != 0xFF {
		return
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	upper := 4
	if obj.BuildingSizeFlags&(1<<el.BuildingType) != 0 {
		upper = 5
	}
	tx, ty := int(x)>>5, int(y)>>5
	for py := max(ty-4, 0); py <= min(ty+upper, m.Height-1); py++ { // getClampedRange
		for px := max(tx-4, 0); px <= min(tx+upper, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, s := range t.Stations {
				if s.Ghost || s.AiAllocated || int(s.StationID) >= 1024 {
					continue
				}
				ind.StationsInRange[s.StationID/32] |= 1 << (s.StationID % 32)
			}
		}
	}
}

// industriesUpdateProducedCargoStats is IndustryManager::updateProducedCargoStats
// (TileManager::update, once a tile-loop round): each industry's best four
// stations by rating per produced cargo, from its stationsInRange (stations
// not empty, taking the cargo from producers; an insertion sort, the later
// one first on equal ratings), each rating then scaled by what the better
// ones left; the range cleared.
func (m *Map) industriesUpdateProducedCargoStats(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	for i := range m.Industries {
		ind := &m.Industries[i]
		obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if obj == nil {
			continue
		}
		produced, _ := om.IndustryCargoIDs(obj)
		for c := 0; c < 2; c++ {
			st, rt := &ind.ProducedStations[c], &ind.ProducedRatings[c]
			for k := range st {
				st[k] = stationIDNull
			}
			cargo := produced[c]
			if cargo == 0xFF {
				continue
			}
			for id := 0; id < 1024 && id < len(m.StationNames); id++ {
				if ind.StationsInRange[id/32]&(1<<(id%32)) == 0 {
					continue
				}
				sn := &m.StationNames[id]
				if sn.Name == 0 { // station->empty()
					continue
				}
				cs := &sn.Cargo[cargo]
				if cs.Flags&stationCargoAcceptedFromProducer == 0 {
					continue
				}
				for k := 0; k < 4; k++ {
					if st[k] == stationIDNull || rt[k] <= cs.Rating {
						copy(st[k+1:], st[k:3]) // rotate: the last falls off
						copy(rt[k+1:], rt[k:3])
						st[k], rt[k] = uint16(id), cs.Rating
						break
					}
				}
			}
			frac := uint8(0xFF)
			for k := range rt {
				rt[k] = uint8(uint16(rt[k]) * uint16(frac) / 256)
				frac = -rt[k]
			}
		}
		ind.StationsInRange = [32]uint32{}
	}
}
