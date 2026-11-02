package sim

// Vehicle2::update's speed step (0x004A9B0B) — the train's motor state and
// currentSpeed from Vehicle1's targetSpeed, the cars' weight on slopes,
// power over weight and air drag.
//
// The per-car inputs come from Map.CarPhysicsOf (isOnRackRail,
// shouldStartWheelSlipping: wheelslip.go). PARTIAL: the "vehicle slipped"
// news message is not posted. sub_4A9F20 (the move that
// follows) is Vehicle2Motion.

// MotorState values.
const (
	MotorStopped          uint8 = 0
	MotorAccelerating     uint8 = 1
	MotorCoasting         uint8 = 2
	MotorBraking          uint8 = 3
	MotorStoppedOnIncline uint8 = 4
)

// VehicleFlags bits used here.
const (
	VehicleFlagUnk0          uint16 = 1 << 0
	VehicleFlagManualControl uint16 = 1 << 6
)

// PitchLateralGravity is kPitchLateralGravity (×256, by Pitch).
var PitchLateralGravity = [13]int32{0, -6645, -13223, -19182, -24963, 6645, 13223, 19182, 24963, -11290, 11290, -21628, 21628}

// Speed32 constants (mph << 16; upstream's _mph literals).
const (
	mph            int32 = 1 << 16
	mph0_18311     int32 = 0x2EE0
	mph0_5         int32 = 0x8000
	mph1_5         int32 = 0x18000
	mph5           int32 = 5 * mph
	brakeFloorMph  int32 = 5 * mph
	stoppedSlipMph int32 = 3 * mph
)

// CarPhysics is what the speed step reads of one car's front bogie.
type CarPhysics struct {
	Pitch          uint8  // front bogie spritePitch
	TotalCarWeight uint16 // front bogie totalCarWeight
	Powered        bool   // its vehicle object has power != 0
	OnRackRail     bool   // isOnRackRail (see PARTIAL)
	StartSlipping  bool   // shouldStartWheelSlipping (see PARTIAL)
	// WheelSlipping is the car's components' wheelSlipping (all equal: the
	// car starts and stops slipping together); updated in place.
	WheelSlipping uint8
}

// Vehicle2Speed is the state the speed step changes.
type Vehicle2Speed struct {
	CurrentSpeed int32 // Speed32
	MotorState   uint8
}

// UpdateSpeed is Vehicle2::update up to sub_4A9F20: it sets MotorState and
// CurrentSpeed. targetSpeed is Vehicle1's (Speed16, mph); headFlags and
// manualPower the head's; brokenDown is var_73 bit 0.
func (v *Vehicle2Speed) UpdateSpeed(targetSpeed int16, headFlags uint16, manualPower int8, totalPower, totalWeight uint16, brokenDown bool, cars []CarPhysics) {
	target := int32(targetSpeed) << 16
	manual := headFlags&VehicleFlagManualControl != 0
	v.MotorState = MotorAccelerating
	diff := v.CurrentSpeed - target
	if diff > 0 {
		v.MotorState = MotorBraking
		ns := v.CurrentSpeed - (v.CurrentSpeed/64 + mph0_18311)
		v.CurrentSpeed = max(ns, max(target, brakeFloorMph))
		return
	}
	if !manual {
		if diff >= -mph1_5 {
			v.MotorState = MotorCoasting
		}
		if v.CurrentSpeed == 0 {
			v.MotorState = MotorStopped
		}
	}
	for i := range cars {
		if c := &cars[i]; c.WheelSlipping != 0 {
			c.WheelSlipping++
			if c.WheelSlipping >= WheelSlippingDuration {
				c.WheelSlipping = 0
			}
		}
	}
	onRack, slipping := true, false
	var ebp int32
	for i := range cars {
		c := &cars[i]
		if c.StartSlipping {
			c.WheelSlipping = 1
		}
		if c.WheelSlipping != 0 {
			slipping = true
		}
		g := PitchLateralGravity[c.Pitch%13]
		if g <= -19182 && c.Powered {
			onRack = onRack && c.OnRackRail
		}
		ebp += (int32(c.TotalCarWeight) * g) >> 8
	}
	if !onRack {
		ebp /= 2
		if headFlags&VehicleFlagUnk0 == 0 {
			v.MotorState = MotorStoppedOnIncline
			// (a "vehicle slipped" message is posted at ≤ 3 mph: not ported)
		}
	}
	if onRack && !slipping && totalWeight != 0 {
		power := int32(totalPower)
		if brokenDown {
			power = int32(totalPower) / 4
		}
		if manual {
			switch {
			case manualPower < 0 && manualPower <= -10:
				v.MotorState = MotorBraking
			case manualPower < 0:
				v.MotorState = MotorCoasting
			case manualPower >= 10:
				v.MotorState = MotorAccelerating
			default:
				v.MotorState = MotorCoasting
			}
			ebp += (power * 2048 * int32(manualPower)) / (int32(totalWeight) * 40)
		} else {
			ebp += (power * 2048) / int32(totalWeight)
		}
	}
	s16 := v.CurrentSpeed / 65536 // toSpeed16
	ebp -= s16 * s16
	adj := min(ebp, mph0_5)
	if v.CurrentSpeed < 0 {
		adj = mph0_5
	}
	ns := adj + v.CurrentSpeed
	if adj < 0 {
		minSpeed := mph5
		if manual || !onRack {
			minSpeed = 0
		}
		if v.CurrentSpeed >= minSpeed {
			ns = max(ns, minSpeed)
		}
	}
	if !manual {
		if ns < 0 {
			ns = target
		} else {
			ns = min(ns, target)
		}
	}
	v.CurrentSpeed = ns
}
