package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Breakdowns — VehicleHead::updateBreakdown (every tick) and
// applyBreakdownToTrain.
//
// PARTIAL: the smoke puffs and the breakdown sound are not drawn / played
// (the sound's gPrng1 draw is made, so the random sequence stays upstream's);
// updateDaily's countdowns (a breakdown ending after breakdownTimeout days, a
// new one falling due after timeoutToBreakdown days, sub_4BA873) wait for a
// day tick; Config breakdownsDisabled is taken as off (upstream's default).

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
		// brokenDown: Smoke::create every 4th scenario tick — PARTIAL
		if front.BreakdownFlags&BreakdownPending != 0 && !m.TitleMode {
			front.BreakdownFlags &^= BreakdownPending
			front.BreakdownFlags |= BreakdownBrokenDn
			front.BreakdownTimeout = 5
			m.ApplyBreakdownToTrain(tr, om)
			m.Rng1.RandNextRange(26, 26+5) // the breakdown sound — PARTIAL (not played)
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
