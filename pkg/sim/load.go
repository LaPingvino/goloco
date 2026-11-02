package sim

import (
	"math/bits"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Loading — VehicleHead.cpp's beginLoading, updateLoadCargo (0x004B9B5F-side),
// updateLoadCargoComponent, beginNewJourney (0x004BAC74) and landLoadingUpdate
// (0x004A9011).
//
// PARTIAL: the cantWaitForFullLoad message, the cargo sprite update (drawn
// from the cargo state anyway) and the Ui invalidations are not ported.
// After loading, pathingShouldReverse / tryReverse run as upstream.

const stationFlag7 uint16 = 1 << 7 // StationFlags::flag_7
const cargoStatFlag2 uint8 = 1 << 2

// BeginLoading is beginLoading.
func (m *Map) BeginLoading(tr Train) {
	head := &m.Entities[tr.Head]
	head.Status = StatusLoading
	head.CargoTransferTimeout = 10
	for _, car := range tr.Cars {
		for _, cc := range car {
			for _, i := range []int{cc.Front, cc.Back, cc.Body} {
				m.Entities[i].BreakdownFlags |= BreakdownAwaitingCargoTransfer
			}
		}
	}
}

// updateLoadCargoComponent loads one cargo slot from the head's station:
// an empty slot picks its cargo (wait-for orders first, else the most
// waiting); an UnloadAll order for it or a full slot loads nothing.
func (m *Map) updateLoadCargoComponent(tr Train, cargo *scenario.VehicleCargo, bogieIdx int, om *objects.ObjectManager) bool {
	head := &m.Entities[tr.Head]
	if cargo.MaxQty == 0 || head.StationID == stationIDNull || int(head.StationID) >= len(m.StationNames) {
		return false
	}
	mod := m.loadingModifier(head, &m.Entities[bogieIdx], om)
	station := &m.StationNames[head.StationID]
	orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder)
	if cargo.Qty == 0 {
		var waitFor uint32
		for _, o := range orders {
			if o.Type() == OrderWaitFor {
				waitFor |= 1 << o.Cargo()
			}
			if o.HasFlags(OrderIsRoutable) {
				break
			}
		}
		var all uint32
		for _, o := range orders {
			if o.Type() == OrderWaitFor {
				all |= 1 << o.Cargo()
			}
		}
		if all == 0 {
			all = 0xFFFFFFFF
		}
		chosen, highest := uint8(0xFF), uint16(0)
		for acc := cargo.AcceptedTypes; acc != 0; {
			c := uint8(bits.TrailingZeros32(acc))
			acc &^= 1 << c
			if all&(1<<c) == 0 {
				continue
			}
			if waitFor&(1<<c) != 0 {
				chosen, highest = c, 0xFFFF
			} else if highest < station.Cargo[c].Quantity {
				chosen, highest = c, station.Cargo[c].Quantity
			}
		}
		if highest != 0 {
			cargo.Type = chosen
		}
	}
	for _, o := range orders {
		if !o.HasFlags(OrderHasCargo) {
			break
		}
		if o.Type() == OrderUnloadAll && o.Cargo() == cargo.Type {
			return false
		}
	}
	if cargo.Qty == cargo.MaxQty || cargo.Type >= 32 {
		return false
	}
	sc := &station.Cargo[cargo.Type]
	qty := min(uint16(cargo.MaxQty-cargo.Qty), sc.Quantity)
	if int(cargo.Type) < len(om.CargoObjects) && om.CargoObjects[cargo.Type] != nil {
		head.CargoTransferTimeout = uint16(min(uint32(om.CargoObjects[cargo.Type].CargoTransferTime)*uint32(qty)*uint32(mod)/256, 0xFFFF))
	}
	if sc.Quantity != 0 {
		if cargo.Qty == 0 {
			cargo.TownFrom, cargo.NumDays = sc.Origin, sc.EnrouteAge
		} else {
			cargo.NumDays = max(cargo.NumDays, sc.EnrouteAge)
			if int(sc.Origin) < len(m.StationNames) && int(cargo.TownFrom) < len(m.StationNames) {
				here := olmath.Vector2i{X: int32(station.X), Y: int32(station.Y)}
				o, s := m.StationNames[sc.Origin], m.StationNames[cargo.TownFrom]
				if olmath.Distance2D(here, olmath.Vector2i{X: int32(s.X), Y: int32(s.Y)}) >= olmath.Distance2D(here, olmath.Vector2i{X: int32(o.X), Y: int32(o.Y)}) {
					cargo.TownFrom = sc.Origin
				}
			}
		}
	}
	sc.Quantity -= qty
	m.UpdateCargoDistribution(int(head.StationID), om)
	cargo.Qty += uint8(qty)
	typeAge := [6]uint8{0, 5, 3, 2, 0, 0}
	sc.Age = min(sc.Age, typeAge[head.VehicleType%6])
	station.Var3B0 = 0
	station.Flags |= stationFlag7
	v2max := m.Entities[tr.Veh2].MaxSpeed
	carAge := uint8(min((m.CurrentDay-m.Entities[bogieIdx].CreationDay)/256, 0xFF))
	if sc.Flags&cargoStatFlag2 != 0 {
		sc.VehicleSpeed = (sc.VehicleSpeed + v2max) / 2
		sc.VehicleAge = uint8((uint16(sc.VehicleAge) + uint16(carAge)) / 2)
	} else {
		sc.Flags |= cargoStatFlag2
		sc.VehicleSpeed, sc.VehicleAge = v2max, carAge
	}
	if int(head.Owner) < len(m.Companies) {
		m.Companies[head.Owner].Var49C |= 1 << cargo.Type
	}
	m.UpdateTrainProperties(tr, om)
	return true
}

// UpdateLoadCargo is updateLoadCargo: true while loading goes on.
func (m *Map) UpdateLoadCargo(tr Train, om *objects.ObjectManager) bool {
	head := &m.Entities[tr.Head]
	if head.CargoTransferTimeout != 0 {
		head.CargoTransferTimeout--
		return true
	}
	for _, car := range tr.Cars {
		for _, cc := range car {
			front, back, body := &m.Entities[cc.Front], &m.Entities[cc.Back], &m.Entities[cc.Body]
			switch {
			case front.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				front.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				if front.Cargo.Type != 0xFF {
					m.updateLoadCargoComponent(tr, &front.Cargo, cc.Front, om)
				}
				return true
			case back.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				back.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				return true
			case body.BreakdownFlags&BreakdownAwaitingCargoTransfer != 0:
				body.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
				if body.Cargo.Type != 0xFF {
					m.updateLoadCargoComponent(tr, &body.Cargo, cc.Back, om)
				}
				return true
			}
		}
	}
	for _, o := range m.OrderRing(head.OrderTableOffset, head.CurrentOrder) {
		if !o.HasFlags(OrderHasCargo) {
			head.CurrentOrder = uint16(o.Offset - int(head.OrderTableOffset))
			break
		}
		if o.Type() != OrderWaitFor {
			continue
		}
		cantWait := false
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, c := range []*scenario.VehicleCargo{&m.Entities[cc.Front].Cargo, &m.Entities[cc.Body].Cargo} {
					if c.Type == o.Cargo() && c.MaxQty != c.Qty {
						if head.BreakdownFlags&BreakdownAwaitingCargoTransfer == 0 {
							m.BeginLoading(tr)
							return true
						}
						cantWait = true // (cantWaitForFullLoad message: PARTIAL)
						break
					}
				}
				if cantWait {
					break
				}
			}
			if cantWait {
				break
			}
		}
	}
	return false
}

// LandLoadingUpdate is landLoadingUpdate: when loading is done the journey
// begins (beginNewJourney), the head stops and moves to its next routable
// order.
func (m *Map) LandLoadingUpdate(tr Train, om *objects.ObjectManager) {
	if m.UpdateLoadCargo(tr, om) {
		return
	}
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	head.JourneyStartTicks = m.ScenarioTicks
	head.JourneyStartX, head.JourneyStartY = int16(v2.PosX), int16(v2.PosY)
	head.BreakdownFlags |= BreakdownJourneyStarted
	head.Status = StatusStopped
	m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
	if head.VehicleFlags&VehicleFlagManualControl != 0 {
		return
	}
	if m.PathingShouldReverse(tr, om) {
		m.TryReverse(tr, om)
	}
}
