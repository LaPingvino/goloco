package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Unloading — VehicleHead.cpp's updateUnloadCargo (0x004B9A3A-side),
// updateUnloadCargoComponent and getLoadingModifier.
//
// updateLastIncomeStats (Vehicle1 lastIncome) is here too (incomestats.go).
// The workersCelebrate / citizensCelebrate news are posted. PARTIAL: the AI
// thought revenue (aiThoughts[].var_80), the
// money effect and income sound are not ported.

const (
	stationIDNull         = 0xFFFF
	stationFlag8   uint16 = 1 << 8 // StationFlags::flag_8
	cargoStatFlag3 uint8  = 1 << 3 // StationCargoStatsFlags::flag3
	cargoAccepted  uint8  = 1 << 0 // acceptedForConsumer (isAccepted)

	minPastStationPenalty  = 1  // kMinVehiclePastStationPenalty
	railPastStationPenalty = 12 // kRailVehiclePastStationPenalty
	roadPastStationPenalty = 2  // kRoadVehiclePastStationPenalty
)

// loadingModifier is getLoadingModifier: a car standing past the station
// loads 12× (rail) / 2× (road) slower; a road stop that is not a road end
// sets awaitingCargoTransfer on the head. (Upstream's switch `break`s land
// on the final return: 1.) disableVehicleLoadPenaltyCheat is off.
func (m *Map) loadingModifier(head *scenario.VehicleEntity, bogie *scenario.VehicleEntity, om *objects.ObjectManager) int32 {
	t := m.TileAt(int(bogie.TileX)>>5, int(bogie.TileY)>>5)
	tad := bogie.TrackAndDirection
	switch head.Mode {
	case 0:
		mod := int32(railPastStationPenalty)
		if t == nil {
			return mod
		}
		if st := t.trainStation(uint8(tad>>3&0x3F), uint8(tad&3), bogie.TileBaseZ); st != nil {
			if st.AiAllocated || st.Ghost || st.StationID != head.StationID {
				return minPastStationPenalty
			}
			mod = minPastStationPenalty
		}
		return mod
	case 1:
		mod := int32(roadPastStationPenalty)
		if t == nil {
			return mod
		}
		if st := t.roadStation(uint8(tad>>3&0xF), uint8(tad&3), bogie.TileBaseZ); st != nil {
			if st.AiAllocated || st.Ghost || st.StationID != head.StationID {
				return minPastStationPenalty
			}
			if rs := om.GetRoadStationObjectByIndex(int(st.ObjectID)); rs == nil || rs.Flags&(1<<3) == 0 { // roadEnd
				head.BreakdownFlags |= BreakdownAwaitingCargoTransfer
			}
			mod = minPastStationPenalty
		}
		return mod
	}
	return minPastStationPenalty
}

func boundAdd32(a, b uint32) uint32 {
	if a+b < a {
		return 0xFFFFFFFF
	}
	return a + b
}

func (m *Map) industryByID(id uint8) *scenario.Industry {
	for i := range m.Industries {
		if m.Industries[i].ID == id {
			return &m.Industries[i]
		}
	}
	return nil
}

// updateUnloadCargoComponent unloads one cargo slot at the head's station:
// delivered and paid where the station accepts it (company, town, industry
// records), otherwise left at the station as a transfer; then the transfer
// time and the train's properties.
func (m *Map) updateUnloadCargoComponent(tr Train, cargo *scenario.VehicleCargo, bogieIdx int, om *objects.ObjectManager) bool {
	head := &m.Entities[tr.Head]
	if cargo.Qty == 0 || head.StationID == stationIDNull || int(head.StationID) >= len(m.StationNames) || cargo.Type >= 32 {
		return false
	}
	station := &m.StationNames[head.StationID]
	cs := &station.Cargo[cargo.Type]
	var cargoObj *objects.CargoObject
	if int(cargo.Type) < len(om.CargoObjects) {
		cargoObj = om.CargoObjects[cargo.Type]
	}
	if cs.Flags&cargoAccepted != 0 {
		m.DeliverCargoToTown(int(head.StationID), cargo.Type, uint16(cargo.Qty))
		var tiles int32
		if int(cargo.TownFrom) < len(m.StationNames) {
			src := m.StationNames[cargo.TownFrom]
			tiles = int32(olmath.Distance2D(olmath.Vector2i{X: int32(station.X), Y: int32(station.Y)}, olmath.Vector2i{X: int32(src.X), Y: int32(src.Y)})) / 32
		}
		if int(head.Owner) < len(m.Companies) && cargoObj != nil {
			c := &m.Companies[head.Owner]
			c.CargoUnitsTotalDelivered += uint32(cargo.Qty)
			c.CargoUnitsTotalDistance += uint32(min(uint64(cargo.Qty)*uint64(tiles), 0xFFFFFFFF))
			pay := m.DeliveredCargoPayment(cargoObj, int32(cargo.Qty), tiles, uint16(cargo.NumDays))
			c.CargoDelivered[cargo.Type] = boundAdd32(c.CargoDelivered[cargo.Type], uint32(cargo.Qty))
			m.updateLastIncomeStats(tr, cargo.Type, uint16(cargo.Qty), uint16(tiles), cargo.NumDays, pay)
			head.Var58 += uint32(pay)
			c.Var4A0 |= 1 << cargo.Type
		}
		station.Var3B1 = 0
		station.Flags |= stationFlag8
		if cs.IndustryID != 0xFF {
			if ind := m.industryByID(cs.IndustryID); ind != nil {
				if io := om.GetIndustryObjectByIndex(int(ind.ObjectID)); io != nil {
					for i, n := range io.RequiredCargoNames {
						if n == "" || cargoSlotByName(om, n) != int(cargo.Type) {
							continue
						}
						ind.ReceivedDailyTotal[i] = boundAdd16(ind.ReceivedDailyTotal[i], uint32(cargo.Qty))
						ind.ReceivedMonthlyTotal[i] = boundAdd16(ind.ReceivedMonthlyTotal[i], uint32(cargo.Qty))
					}
				}
				if ind.HistoryMinProduction&(1<<cargo.Type) == 0 {
					ind.HistoryMinProduction |= 1 << cargo.Type
					m.PostMessage(om, MsgWorkersCelebrate, head.Owner, head.ID, uint16(cs.IndustryID), uint16(cs.IndustryID)|uint16(cargo.Type)<<8)
				}
				if t := m.TownByID(ind.Town); t != nil {
					t.Var1A8 |= 1 << cargo.Type
				}
				if t := m.TownByID(station.Town); t != nil {
					t.Var1A8 |= 1 << cargo.Type
				}
			}
		}
		if t := m.TownByID(station.Town); t != nil && t.Var1A8&(1<<cargo.Type) == 0 {
			t.Var1A8 |= 1 << cargo.Type
			m.PostMessage(om, MsgCitizensCelebrate, head.Owner, head.ID, station.Town, station.Town|uint16(cargo.Type)<<8)
		}
		cs.Flags |= cargoStatFlag3
	} else {
		for _, o := range m.OrderRing(head.OrderTableOffset, head.CurrentOrder) {
			if !o.HasFlags(OrderHasCargo) {
				return false
			}
			if o.Type() == OrderUnloadAll && o.Cargo() == cargo.Type {
				break
			}
		}
		had := cs.Quantity != 0
		cs.Quantity = boundAdd16(cs.Quantity, uint32(cargo.Qty))
		m.UpdateCargoDistribution(int(head.StationID), om)
		cs.EnrouteAge = uint8(min(uint32(cs.EnrouteAge)+uint32(cargo.NumDays), 0xFF))
		if had {
			mult := int32(1<<16) - (int32(cargo.Qty)<<16)/int32(cs.Quantity)
			cs.EnrouteAge = uint8((int32(cs.EnrouteAge) * mult) >> 16)
		}
		setOrigin := true
		if cs.Origin != stationIDNull && int(cs.Origin) < len(m.StationNames) && int(cargo.TownFrom) < len(m.StationNames) {
			here := olmath.Vector2i{X: int32(station.X), Y: int32(station.Y)}
			o := m.StationNames[cs.Origin]
			s := m.StationNames[cargo.TownFrom]
			if olmath.Distance2D(here, olmath.Vector2i{X: int32(s.X), Y: int32(s.Y)}) > olmath.Distance2D(here, olmath.Vector2i{X: int32(o.X), Y: int32(o.Y)}) {
				setOrigin = false
			}
		}
		if setOrigin {
			cs.Origin = cargo.TownFrom
		}
	}
	mod := m.loadingModifier(head, &m.Entities[bogieIdx], om)
	if cargoObj != nil {
		head.CargoTransferTimeout = uint16(min(uint32(cargoObj.CargoTransferTime)*uint32(cargo.Qty)*uint32(mod)/256, 0xFFFF))
	}
	cargo.Qty = 0
	m.UpdateTrainProperties(tr, om)
	return true
}

// UpdateUnloadCargo is updateUnloadCargo: count down the transfer time; then
// unload the next component still awaiting transfer (front bogie's secondary
// cargo, back bogie, body's primary cargo), one per call; when none is left,
// book the trip's takings (var_58) to the company and begin loading.
func (m *Map) UpdateUnloadCargo(tr Train, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if head.CargoTransferTimeout != 0 {
		head.CargoTransferTimeout--
		return
	}
	for _, car := range tr.Cars {
		for _, cc := range car {
			front, back, body := &m.Entities[cc.Front], &m.Entities[cc.Back], &m.Entities[cc.Body]
			switch {
			case front.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				front.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				if front.Cargo.Type != 0xFF {
					m.updateUnloadCargoComponent(tr, &front.Cargo, cc.Front, om)
				}
				return
			case back.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				back.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				return
			case body.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				body.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				if body.Cargo.Type != 0xFF {
					m.updateUnloadCargoComponent(tr, &body.Cargo, cc.Back, om)
				}
				return
			}
		}
	}
	profit := int32(head.Var58)
	head.Var58 = 0
	if profit != 0 {
		m.Entities[tr.Veh2].CurMonthRevenue += profit
		m.Entities[tr.Veh1].Var48 |= 1 << 2 // Flags48::flag2
		m.ApplyPaymentToCompany(head.Owner, -profit, head.VehicleType*2)
		if len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 { // the first car's body + 28 up (spendMoneyEffect; PARTIAL: the income sound)
			b := &m.Entities[tr.Cars[0][0].Body]
			m.SpendMoneyEffect(int32(b.PosX), int32(b.PosY), int32(b.PosZ)+28, head.Owner, -profit)
		}
	}
	m.BeginLoading(tr)
}
