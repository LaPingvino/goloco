package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
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
// PARTIAL: Industry::update's farm tiles; see industryUpdateMonthly.

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
// the industry's stationsInRange (the last part of industryElementUpdate).
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

const (
	industryFlagClosingDown          = 1 << 2 // IndustryFlags::closingDown
	industryManagerDisallowCloseDown = 1 << 0 // IndustryManager::Flags::disallowIndustriesCloseDown
)

// industriesUpdateMonthly is IndustryManager::updateMonthly: new industries
// first (tryCreateNewIndustriesMonthly), then each industry's month.
func (m *Map) industriesUpdateMonthly(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	m.tryCreateNewIndustriesMonthly(om)
	ids := make([]uint8, len(m.Industries)) // upstream's fixed slots: an industry removed this month leaves the others in place
	for i := range m.Industries {
		ids[i] = m.Industries[i].ID
	}
	for _, id := range ids {
		if ind := m.industryByID(id); ind != nil {
			m.industryUpdateMonthly(ind, om)
		}
	}
}

// industryUpdateMonthly is Industry::updateMonthly (0x00453868): a primary
// industry's production may double (isMonthlyProductionUp: > 70% taken,
// 1 in 32) or halve (isMonthlyProductionDown), else it may start closing
// (obsolete, or the object's monthly closure chance — the industry's own
// prng); the month's produced share kept in the 20-year histories (÷ 50,
// at most 255); the monthly totals moved to the previous month's and the
// percentage transported worked out; an industry closing down with no
// production left removed (RemoveIndustry). PARTIAL: the news messages
// (MessageManager::post: production up / down, closing down) — goloco has
// no messages yet.
func (m *Map) industryUpdateMonthly(ind *scenario.Industry, om *objects.ObjectManager) {
	if ind.Flags&industryFlagIsGhost != 0 {
		return
	}
	if ind.Flags&industryFlagClosingDown != 0 && ind.DailyProduction[0] == 0 && ind.DailyProduction[1] == 0 {
		m.RemoveIndustry(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, ind.ID) // IndustryRemovalArgs (industryremove.go)
		return
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	_, required := om.IndustryCargoIDs(obj)
	event := false
	if ind.UnderConstruction == 0xFF && ind.Flags&industryFlagClosingDown == 0 && required[0] == 0xFF {
		if m.isMonthlyProductionUp(ind, obj) {
			ind.DailyProductionTarget[0] = min(100, ind.DailyProductionTarget[0]*2)
			ind.DailyProductionTarget[1] = min(100, ind.DailyProductionTarget[1]*2)
			event = true
		} else if m.isMonthlyProductionDown(ind, obj) {
			ind.DailyProductionTarget[0] /= 2
			ind.DailyProductionTarget[1] /= 2
			event = true
		}
	}
	if !event && m.IndustryManagerFlags&industryManagerDisallowCloseDown == 0 && ind.UnderConstruction == 0xFF && ind.Flags&industryFlagClosingDown == 0 {
		if m.isMonthlyProductionClosing(ind, obj) {
			ind.Flags |= industryFlagClosingDown
			ind.DailyProductionTarget = [2]uint16{}
		}
	}
	for k := 0; k < 2; k++ {
		h := &ind.History[k]
		if int(ind.HistorySize[k]) == len(h) {
			copy(h[:], h[1:]) // rotate left by one; the new value goes last
		} else {
			ind.HistorySize[k]++
		}
		h[ind.HistorySize[k]-1] = uint8(min(uint32(ind.ProducedMonthlyTotal[k]), 12750) / 50)
	}
	for k := 0; k < 2; k++ {
		ind.ProducedPrevMonth[k], ind.ProducedMonthlyTotal[k] = ind.ProducedMonthlyTotal[k], 0
		ind.DeliveredPrevMonth[k], ind.DeliveredMonthlyTotal[k] = ind.DeliveredMonthlyTotal[k], 0
		transported := uint32(min(ind.ProducedPrevMonth[k], ind.DeliveredPrevMonth[k]))
		if ind.ProducedPrevMonth[k] != 0 {
			transported = transported * 100 / uint32(ind.ProducedPrevMonth[k])
		}
		ind.PercentTransported[k] = uint8(transported)
	}
	ind.ReceivedPrevMonth, ind.ReceivedMonthlyTotal = ind.ReceivedMonthlyTotal, [3]uint16{}
}

func (m *Map) isMonthlyProductionUp(ind *scenario.Industry, obj *objects.IndustryObject) bool {
	if obj.Flags&objects.IndustryFlagCanIncreaseProduction == 0 {
		return false
	}
	return ind.PercentTransported[0] > 70 && m.Rng1.RandNextBound(31) == 0 &&
		ind.DailyProductionTarget[0] < 100 && ind.DailyProductionTarget[1] < 100
}

func (m *Map) isMonthlyProductionDown(ind *scenario.Industry, obj *objects.IndustryObject) bool {
	if obj.Flags&objects.IndustryFlagCanDecreaseProduction == 0 {
		return false
	}
	return (ind.PercentTransported[0] > 50 && ind.DailyProductionTarget[0] > 20 && m.Rng1.RandNextBound(31) == 0) ||
		(ind.PercentTransported[0] <= 50 && ind.DailyProductionTarget[0] > 10 && m.Rng1.RandNextBound(15) == 0)
}

// isMonthlyProductionClosing draws from the industry's own prng.
func (m *Map) isMonthlyProductionClosing(ind *scenario.Industry, obj *objects.IndustryObject) bool {
	p := core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
	defer func() { ind.Prng = [2]uint32{p.Srand0(), p.Srand1()} }()
	return (m.CurrentYear > obj.ObsoleteYear && p.RandNextBound(0xFFFF) < 102) ||
		(obj.MonthlyClosureChance != 0 && int32(obj.MonthlyClosureChance) > p.RandNextBound(0xFFFF))
}

// industryElementUpdate is IndustryElement::update (Map/IndustryElement.cpp)
// from the tile loop, on a building's first tile (not a ghost): under
// construction, a section at a time — 8 steps of progress a section; the last
// section done counts the building in its industry (under_construction to
// 0xFF when all are) and sets its clear height from its parts —; built, a
// building with a part that has no frames may, 1 in 8 by the industry's
// prng, start one of its type's animation sequences (unk38) as an
// animation; then the stations in range (industryElementStations).
func (m *Map) industryElementUpdate(el *scenario.IndustryElement, x, y int32, om *objects.ObjectManager) {
	if el.SequenceIndex != 0 || om == nil {
		return
	}
	ind := m.industryByID(el.IndustryID)
	if ind == nil {
		return
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	if obj == nil {
		return
	}
	large := obj.BuildingSizeFlags&(1<<el.BuildingType) != 0
	var parts []uint8
	if int(el.BuildingType) < len(obj.Variations) {
		parts = obj.Variations[el.BuildingType]
	}
	if !el.IsConstructed {
		built, progress, sections := false, el.SectionProgress, el.Var6Low
		if el.SectionProgress == 7 {
			if len(parts) <= int(el.Var6Low)+1 {
				ind.UnderConstruction++
				if int(ind.UnderConstruction) >= len(ind.Tiles) {
					ind.UnderConstruction = 0xFF
				}
				h := 0
				for _, p := range parts {
					if int(p) < len(obj.PartHeights) {
						h += int(obj.PartHeights[p])
					}
				}
				clearZ := uint8((h+3)/4) + el.BaseZ
				m.industryMultiTile(el, x, y, large, func(e *scenario.IndustryElement, _, _ int32) { e.ClearZ = clearZ })
				built, progress, sections = true, 0, 0
			} else {
				progress, sections = 0, sections+1
			}
		} else {
			progress++
		}
		m.industryMultiTile(el, x, y, large, func(e *scenario.IndustryElement, _, _ int32) {
			e.IsConstructed, e.SectionProgress, e.Var6Low = built, progress, sections
		})
	}
	if el.IsConstructed {
		zeroFrame := false
		for _, p := range parts {
			if int(p) < len(obj.PartAnimations) && obj.PartAnimations[p].NumFrames == 0 {
				zeroFrame = true
				break
			}
		}
		if zeroFrame && el.Var6Low&(1<<5) == 0 {
			var anims []uint8
			for _, u := range obj.Unk38 {
				if u[0] == el.BuildingType && len(anims) < 8 {
					anims = append(anims, u[1])
				}
			}
			if len(anims) != 0 {
				p := core.NewPrngSeed(ind.Prng[0], ind.Prng[1])
				r := p.RandNext()
				ind.Prng = [2]uint32{p.Srand0(), p.Srand1()}
				if r&0x700 == 0 {
					v := anims[uint32(len(anims))*(r&0xFF)/256] | 1<<5 | el.Var6Low&0xC
					m.industryMultiTile(el, x, y, large, func(e *scenario.IndustryElement, _, _ int32) { e.Var6Low = v })
					m.CreateAnimation(4, x, y, el.BaseZ)
				}
			}
		}
	}
	m.industryElementStations(el, x, y, om)
}
