package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Breakdowns — VehicleHead::updateBreakdown (every tick) and
// applyBreakdownToTrain.
//
// The breakdown sound is queued (sounds.go). PARTIAL: the smoke puffs are
// not drawn;
// updateDaily's countdowns (a breakdown ending after breakdownTimeout days, a
// new one falling due after timeoutToBreakdown days, sub_4BA873) wait for a
// day tick; with Config breakdownsDisabled a pending breakdown never begins.

const (
	BreakdownPending  uint8 = 1 << 1
	BreakdownBrokenDn uint8 = 1 << 2
	flags73BrokenDown uint8 = 1 << 0
	flags73Powered    uint8 = 1 << 1
)

// updateBreakdown: a pending breakdown on a car's front bogie breaks it down
// (not on the title screen).
func (m *Map) updateBreakdown(tr Train, om *objects.ObjectManager) {
	switch m.Entities[tr.Head].Status {
	case StatusTravelling, StatusApproaching, StatusBrokenDown:
	default: // unk_0, stopped, waitingAtSignal, unloading, loading, crashed, stuck
		if m.Entities[tr.Head].Status <= StatusStuck {
			return
		}
	}
	for _, car := range tr.Cars {
		front := &m.Entities[car[0].Front]
		if front.BreakdownFlags&BreakdownBrokenDn != 0 && m.ScenarioTicks&3 == 0 { // smoke from the body, every 4th tick
			body := &m.Entities[car[0].Body]
			m.createSmoke(int32(body.PosX), int32(body.PosY), int32(body.PosZ)+4)
		}
		if front.BreakdownFlags&BreakdownPending != 0 && !m.TitleMode && !Config.BreakdownsDisabled {
			front.BreakdownFlags &^= BreakdownPending
			front.BreakdownFlags |= BreakdownBrokenDn
			front.BreakdownTimeout = 5
			m.ApplyBreakdownToTrain(tr, om)
			snd := m.Rng1.RandNextRange(26, 26+5) // a breakdown sound, at the body + 22 up
			body := &m.Entities[car[0].Body]
			m.playSound(uint16(snd), ChannelVehicles, int32(body.PosX), int32(body.PosY), int32(body.PosZ)+22)
		}
	}
}

// ApplyBreakdownToTrain is applyBreakdownToTrain: Vehicle2's var_73 from the
// powered cars' front bogies.
func (m *Map) ApplyBreakdownToTrain(tr Train, om *objects.ObjectManager) {
	broken, powered := false, false
	for _, car := range tr.Cars {
		front := &m.Entities[car[0].Front]
		obj := vehicleObject(om, front.ObjectID)
		if obj == nil || obj.Power == 0 {
			continue
		}
		if front.BreakdownFlags&BreakdownBrokenDn != 0 {
			broken = true
		} else {
			powered = true
		}
	}
	v2 := &m.Entities[tr.Veh2]
	switch {
	case broken && powered:
		v2.Var73 |= flags73BrokenDown | flags73Powered
	case broken:
		v2.Var73 = v2.Var73&^flags73Powered | flags73BrokenDown
	default:
		v2.Var73 &^= flags73BrokenDown | flags73Powered
	}
}
