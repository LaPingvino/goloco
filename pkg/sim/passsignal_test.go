package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

// vehiclePassSignal on a saved train: another company's is refused; its own
// query changes nothing; applied, var_5C = 2 and Vehicle1's passSignal set.
func TestVehiclePassSignal(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	tr := firstRailTrain(t, m, trains, 1)
	head := m.Entities[tr.Head]
	if _, ok := m.VehiclePassSignal(&Cmd{Company: head.Owner ^ 1, Objects: om}, CmdApply, head.ID); ok {
		t.Fatal("another company's train passed")
	}
	c := &Cmd{Company: head.Owner, Objects: om}
	m.Entities[tr.Veh1].Var48 &^= flags48PassSignal
	m.Entities[tr.Head].Var5C = 0
	m.Entities[tr.Head].Status = StatusTravelling
	m.Entities[tr.Veh2].CurrentSpeed, m.Entities[tr.Veh1].Var3C = 10, 0x4000 // moving: canBeModified is false
	if _, ok := m.VehiclePassSignal(c, CmdApply, head.ID); ok {
		t.Fatal("a moving train passed")
	}
	m.Entities[tr.Veh2].CurrentSpeed = 0 // stopped at the signal
	if _, ok := m.VehiclePassSignal(c, 0, head.ID); !ok || m.Entities[tr.Head].Var5C != 0 {
		t.Fatalf("query: ok %v var5C %d", ok, m.Entities[tr.Head].Var5C)
	}
	_, ok := m.VehiclePassSignal(c, CmdApply, head.ID)
	if !ok || m.Entities[tr.Head].Var5C != 2 || m.Entities[tr.Veh1].Var48&flags48PassSignal == 0 || c.Pos == nil {
		t.Fatalf("apply: ok %v var5C %d var48 %#x pos %v", ok, m.Entities[tr.Head].Var5C, m.Entities[tr.Veh1].Var48, c.Pos)
	}
}

// canBeModified's reasons (VehicleHead::canBeModified's setErrorText) reach
// the command: a moving train, a crashed, stuck or broken-down one.
func TestCanBeModifiedErrorText(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	tr := firstRailTrain(t, m, trains, 1)
	m.Entities[tr.Head].Status = StatusTravelling
	m.Entities[tr.Veh2].CurrentSpeed, m.Entities[tr.Veh1].Var3C = 10, 0x4000
	for _, k := range []struct {
		status uint8
		want   int
	}{{StatusTravelling, i18n.VehicleMustBeStopped}, {StatusCrashed, i18n.VehicleHasCrashed}, {StatusStuck, i18n.VehicleIsStuck}, {StatusBrokenDown, i18n.VehicleHasBrokenDown}} {
		m.Entities[tr.Head].Status = k.status
		c := &Cmd{}
		if m.canBeModified(c, tr) || c.ErrorText != k.want {
			t.Errorf("status %d: error text %d, want %d", k.status, c.ErrorText, k.want)
		}
	}
	m.Entities[tr.Head].Status = StatusTravelling
	m.Entities[tr.Veh2].CurrentSpeed = 0
	if c := (&Cmd{}); !m.canBeModified(c, tr) || c.ErrorText != 0 {
		t.Errorf("stopped: refused (%d)", c.ErrorText)
	}
}
