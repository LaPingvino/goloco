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

// shouldStartWheelSlipping is asked inside Vehicle2::update's car loop
// (Vehicle2.cpp 0x004A9B0B): after the motor state is set from this tick's
// target and the slip counters are incremented, and never while braking.
// goloco asked it before the speed step, with last tick's motor state and
// counter: at the end of a slip (counter 63 → 0) OpenLoco draws for a new one
// and goloco did not — Boulder Breakers' rng parted there (1925-04-05).
func TestVehicle2AsksWheelSlipInTheLoop(t *testing.T) {
	type ask struct{ motor, slip uint8 }
	var asked []ask
	start := func(motor, slip uint8) bool { asked = append(asked, ask{motor, slip}); return false }

	v := Vehicle2Speed{CurrentSpeed: 60 * mph, MotorState: MotorAccelerating}
	v.UpdateSpeed(40, 0, 0, 1000, 100, false, []CarPhysics{{TotalCarWeight: 100, Powered: true, OnRackRail: true, StartSlipping: start}})
	if len(asked) != 0 {
		t.Errorf("braking: asked %v, want no question (and no rng draw)", asked)
	}

	v = Vehicle2Speed{CurrentSpeed: 2 * mph, MotorState: MotorCoasting}
	cars := []CarPhysics{{TotalCarWeight: 100, Powered: true, OnRackRail: true, StartSlipping: start, WheelSlipping: WheelSlippingDuration - 1}}
	v.UpdateSpeed(60, 0, 0, 1000, 100, false, cars)
	if len(asked) != 1 || asked[0] != (ask{MotorAccelerating, 0}) {
		t.Errorf("asked %v, want once with this tick's motor state (accelerating) and the counter run out (0)", asked)
	}
}
