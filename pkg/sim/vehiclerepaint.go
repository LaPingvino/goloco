package sim

// GameCommands/Vehicles/VehicleRepaint.cpp: a train, a car or one of its
// parts in a colour scheme.

// VehicleRepaintFlags (VehicleRepaint.h).
const (
	RepaintBody             uint8 = 1 << 0
	RepaintFrontBogie       uint8 = 1 << 1
	RepaintBackBogie        uint8 = 1 << 2
	RepaintEntireCar        uint8 = 1 << 3
	RepaintEntireTrain      uint8 = 1 << 4
	RepaintFromVehicleUi          = RepaintBody | RepaintFrontBogie | RepaintBackBogie | RepaintEntireCar
	kBodyColour                   = 0
	kFrontBogieColour             = 1
	kBackBogieColour              = 2
)

// ColourScheme is a vehicle's primary and secondary colour.
type ColourScheme struct{ Primary, Secondary uint8 }

// VehicleRepaint is VehicleRepaintArgs: the vehicle (any part of a train),
// the colours per part (body, front bogie, back bogie) and the flags.
type VehicleRepaint struct {
	Vehicle uint16
	Colours [4]ColourScheme
	Flags   uint8
}

func (m *Map) paintComponent(comp CarComponent, a VehicleRepaint) {
	set := func(i, k int) {
		if i >= 0 && i < len(m.Entities) {
			m.Entities[i].ColourPrimary, m.Entities[i].ColourSecondary = a.Colours[k].Primary, a.Colours[k].Secondary
		}
	}
	if a.Flags&RepaintBody != 0 {
		set(comp.Body, kBodyColour)
	}
	if a.Flags&RepaintFrontBogie != 0 {
		set(comp.Front, kFrontBogieColour)
	}
	if a.Flags&RepaintBackBogie != 0 {
		set(comp.Back, kBackBogieColour)
	}
}

func (m *Map) paintEntireCar(car Car, a VehicleRepaint) {
	for _, comp := range car {
		m.paintComponent(comp, a)
	}
}

// VehicleRepaint is the vehicleRepaint game command: refused when the
// vehicle isn't one or its owner isn't compatible; applying, every car of
// the train (entire train), or the car holding the vehicle (entire car), or
// just its component is painted. Costs nothing.
// PARTIAL: invalidateSprite (goloco redraws).
func (m *Map) VehicleRepaint(c *Cmd, flags uint8, a VehicleRepaint) (int32, bool) {
	v := m.entity(a.Vehicle)
	if v == nil || !c.companyCompatible(v.Owner) {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	tr, err := m.TrainOf(v.HeadID)
	if err != nil {
		return 0, false
	}
	is := func(i int) bool { return i >= 0 && i < len(m.Entities) && m.Entities[i].ID == a.Vehicle }
	for _, car := range tr.Cars {
		if a.Flags&RepaintEntireTrain != 0 {
			m.paintEntireCar(car, a)
			continue
		}
		for _, comp := range car {
			if is(comp.Front) || is(comp.Back) || is(comp.Body) {
				if a.Flags&RepaintEntireCar != 0 {
					m.paintEntireCar(car, a)
				} else {
					m.paintComponent(comp, a)
				}
				return 0, true
			}
		}
	}
	return 0, true
}
