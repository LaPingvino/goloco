package sim

// Vehicle2::update's speed step (0x004A9B0B) — the train's motor state and
// currentSpeed from Vehicle1's targetSpeed, the cars' weight on slopes,
// power over weight and air drag.
//
// The per-car inputs come from Map.CarPhysicsOf (isOnRackRail,
// shouldStartWheelSlipping: wheelslip.go). PARTIAL: the "vehicle slipped"
// news message is not posted. sub_4A9F20 (the move that
// follows) is Map.Vehicle2Move.

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

// BrakeLightTimeout is kBrakeLightTimeout (ticks the brake lights stay on).
const BrakeLightTimeout = 7

// VehicleUpdateDistances is VehicleUpdateDistances (0x0113612C/30).
type VehicleUpdateDistances struct{ D1, D2 int32 }

// Vehicle2Move is sub_4A9F20 (0x004A9F20): Vehicle2 moves by D1 with
// isVeh2UnkM15 (it stops where Vehicle1's routing is); what it could not
// travel comes off both distances; blocked, it stops (unless manual); the
// distance travelled comes off the head's and Vehicle1's var_3C; the brake
// lights hold for kBrakeLightTimeout ticks after braking. false: the train is
// destroyed (noRouteFound). (Sprite invalidation is upstream's redraw only.)
func (m *Map) Vehicle2Move(c *MotionComponent, v *Vehicle2Speed, d *VehicleUpdateDistances, manual bool, veh1Handle RoutingHandle, requiredMods uint8, headVar3C, veh1Var3C *int32, brakeLight *uint8) bool {
	next := m.TrackNewPiece
	if c.Road {
		next = m.RoadNewPiece
	}
	res := c.UpdateMotion(d.D1, func(c *MotionComponent, f *uint32) bool {
		return next(c, f, requiredMods, true, veh1Handle)
	})
	d.D1 -= res.RemainingDistance
	d.D2 -= res.RemainingDistance
	if res.Flags&MotionNoRouteFound != 0 {
		return false
	}
	if res.Blocked() && !manual {
		v.CurrentSpeed = 0
		v.MotorState = MotorStopped
	}
	if v.MotorState == MotorStoppedOnIncline {
		d.D2 = d.D1 + 0x1388
	}
	*headVar3C -= d.D1
	*veh1Var3C -= d.D1
	if v.MotorState == MotorBraking {
		*brakeLight = BrakeLightTimeout
	} else if *brakeLight != 0 {
		*brakeLight--
	}
	return true
}
