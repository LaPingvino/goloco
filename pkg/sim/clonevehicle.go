package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// GameCommands/Vehicles/CloneVehicle.cpp: a new vehicle made of the same
// cars as an existing one, with its colours, orders, express mode and (ships
// and aircraft) refit.

// nested is GameCommands::doCommand below the top level (loc_4313C6 with
// _gameCommandNestLevel > 1): the query pass, then, applying, the applying
// pass and the lower of the two costs — no funding check and no payment (the
// top-level command pays the total).
func nested(flags uint8, run func(flags uint8) (int32, bool)) (int32, bool) {
	cost, ok := run(flags &^ CmdApply)
	if !ok || flags&CmdApply == 0 {
		return cost, ok
	}
	cost2, ok := run(flags)
	if !ok {
		return 0, false
	}
	return min(cost, cost2), true
}

// CloneVehicle is the cloneVehicle game command for the vehicle headed by
// headID. Not applying, its cost is the sum of what each car's VehicleCreate
// would cost. Applying, the first car creates the new vehicle and the others
// are added to it (each a nested VehicleCreate); then the colours are copied
// car by car, every order is inserted at the end of the new table, express
// mode is toggled on if the original has it, and a ship or aircraft is
// refitted to the first car's cargo. The new head's id goes to *newHead
// (lastCreatedVehicleId). The caller pays under ExpenditureVehiclePurchases
// (setExpenditureType, set after the sub-commands). A failing car leaves the
// cars made so far, as upstream does.
func (m *Map) CloneVehicle(c *Cmd, flags uint8, headID uint16, newHead *uint16) (int32, bool) {
	if newHead != nil {
		*newHead = entityIDNull
	}
	src, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	objIDs := make([]uint16, len(src.Cars))
	for i, car := range src.Cars {
		objIDs[i] = m.Entities[car[0].Front].ObjectID
	}
	if flags&CmdApply == 0 {
		var total int32
		for _, id := range objIDs {
			cost, ok := nested(0, func(f uint8) (int32, bool) { return m.CreateVehicle(c, f, id, false, nil) })
			if !ok {
				return 0, false
			}
			total += cost
		}
		return total, true
	}
	var cargoType uint8
	if len(src.Cars) > 0 {
		cargoType = m.Entities[src.Cars[0][0].Body].Cargo.Type
	}
	nh := uint16(entityIDNull)
	var total int32
	for i, id := range objIDs {
		var cost int32
		var ok bool
		if i == 0 {
			cost, ok = nested(CmdApply, func(f uint8) (int32, bool) {
				var h uint16
				cost, ok := m.CreateVehicle(c, f, id, false, &h)
				if f&CmdApply != 0 {
					nh = h
				}
				return cost, ok
			})
			if ok && nh == entityIDNull {
				return 0, false
			}
		} else {
			cost, ok = nested(CmdApply, func(f uint8) (int32, bool) { return m.AddCarToVehicle(c, f, id, nh, false) })
		}
		if !ok {
			return 0, false
		}
		total += cost
	}
	if nh == entityIDNull {
		return 0, false
	}
	if newHead != nil {
		*newHead = nh
	}
	src, err = m.TrainOf(headID) // the entities moved as the new ones were made
	if err != nil {
		return 0, false
	}
	dst, err := m.TrainOf(nh)
	if err != nil {
		return 0, false
	}
	// copyVehicleColours: each component's body, front and back colour scheme.
	for i := 0; i < len(src.Cars) && i < len(dst.Cars); i++ {
		for j := 0; j < len(src.Cars[i]) && j < len(dst.Cars[i]); j++ {
			s, d := src.Cars[i][j], dst.Cars[i][j]
			for _, p := range [][2]int{{s.Body, d.Body}, {s.Front, d.Front}, {s.Back, d.Back}} {
				m.Entities[p[1]].ColourPrimary = m.Entities[p[0]].ColourPrimary
				m.Entities[p[1]].ColourSecondary = m.Entities[p[0]].ColourSecondary
			}
		}
	}
	// The orders, each inserted at the new table's end (sizeOfOrderTable − 1,
	// read again every time).
	var raws [][]byte
	for _, o := range m.OrderTable(m.Entities[src.Head].OrderTableOffset) {
		size := orderSizes[o.Type()]
		raws = append(raws, append([]byte(nil), m.Orders[o.Offset:o.Offset+size]...))
	}
	for _, raw := range raws {
		h := m.entity(nh)
		if h == nil {
			break
		}
		at := h.SizeOfOrderTable - 1
		nested(CmdApply, func(f uint8) (int32, bool) { return m.VehicleOrderInsert(c, f, nh, at, raw) })
	}
	if m.Entities[src.Veh1].Var48&Flags48ExpressMode != 0 {
		nested(CmdApply, func(f uint8) (int32, bool) {
			return m.VehicleChangeRunningMode(c, f, nh, RunningModeToggleLocalExpress)
		})
	}
	if h := m.entity(nh); h != nil && (h.VehicleType == uint8(objects.VehicleTypeShip) || h.VehicleType == uint8(objects.VehicleTypeAircraft)) {
		nested(CmdApply, func(f uint8) (int32, bool) { return m.VehicleRefit(c, f, nh, cargoType) })
	}
	return total, true
}
