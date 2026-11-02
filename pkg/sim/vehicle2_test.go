package sim

import "testing"

// Vehicle2::update's speed step, by hand from the upstream code: braking
// sheds speed/64 + 0.18311 mph down to max(target, 5 mph); accelerating on
// the level adds min(power·2048/weight − speed², 0.5 mph), capped at the
// target; within 1.5 mph of it the motor coasts; a stopped train at target 0
// stays stopped.
func TestVehicle2UpdateSpeed(t *testing.T) {
	v := Vehicle2Speed{CurrentSpeed: 60 * mph}
	v.UpdateSpeed(40, 0, 0, 1000, 100, false, nil)
	if want := 60*mph - (60*mph/64 + 0x2EE0); v.MotorState != MotorBraking || v.CurrentSpeed != want {
		t.Errorf("braking: state %d speed %#x, want %#x", v.MotorState, v.CurrentSpeed, want)
	}
	v = Vehicle2Speed{CurrentSpeed: 10 * mph}
	cars := []CarPhysics{{TotalCarWeight: 100, Powered: true, OnRackRail: true}}
	v.UpdateSpeed(60, 0, 0, 1000, 100, false, cars)
	// ebp = 1000·2048/100 − 10² = 20480 − 100 = 20380 < 0.5 mph (32768)
	if want := 10*mph + 20380; v.MotorState != MotorAccelerating || v.CurrentSpeed != want {
		t.Errorf("accelerating: state %d speed %#x, want %#x", v.MotorState, v.CurrentSpeed, want)
	}
	v = Vehicle2Speed{CurrentSpeed: 59 * mph}
	v.UpdateSpeed(60, 0, 0, 1000, 100, false, cars)
	if v.MotorState != MotorCoasting || v.CurrentSpeed > 60*mph {
		t.Errorf("near target: state %d speed %#x", v.MotorState, v.CurrentSpeed)
	}
	v = Vehicle2Speed{}
	v.UpdateSpeed(0, 0, 0, 1000, 100, false, cars)
	if v.MotorState != MotorStopped || v.CurrentSpeed != 0 {
		t.Errorf("stopped: state %d speed %#x", v.MotorState, v.CurrentSpeed)
	}
}
