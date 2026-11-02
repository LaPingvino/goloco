package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Vehicle1::updateRail (0x004A97A6) — the train's target speed from the
// curves, bridges and rack rail on its reserved route, then Vehicle1's own
// move by the distance Vehicle2's speed gives, against var_3C.
//
// PARTIAL: updateRoad (0x004A9969: the road target speed and Vehicle1's own
// overtaking road motion) is not ported; on noRouteFound upstream calls the
// head's sub_4AD93A (not ported: Vehicle1RailMove reports it); the level
// crossing whistle is sound only.

const (
	speed16Max  int16 = 0x7FFF
	speed16Null int16 = -1

	VehicleFlagCommandStop uint16 = 1 << 1
	flags38FasterCurves    uint8  = 1 << 5
)

// Vehicle1RailTargetSpeed is updateRail's targetSpeed (Speed16, mph).
func (m *Map) Vehicle1RailTargetSpeed(head, veh1, veh2 *scenario.VehicleEntity, om *objects.ObjectManager) int16 {
	manual := head.VehicleFlags&VehicleFlagManualControl != 0
	curve := uint16(0xFFFF)
	target := speed16Max
	onRack := false
	if !manual {
		for _, h := range m.RoutingRing(RoutingHandle(veh1.RoutingHandle), true) {
			res := m.Routing(h)
			onRack = onRack || res&(1<<13) != 0
			curve = min(curve, TrackCurveSpeedFraction[(res>>3)&0x3F%44])
			if res&(1<<12) != 0 {
				if b := om.GetBridgeObjectByIndex(int(res&0xE00) >> 9); b != nil && int16(b.MaxSpeed) != speed16Null {
					target = min(int16(b.MaxSpeed), target)
				}
			}
		}
	}
	var curveSpeed int16
	if tr := om.GetTrackObjectByIndex(int(veh1.TrackType)); tr != nil {
		curveSpeed = tr.CurveSpeed
	}
	frac := int32(uint32(curve) * uint32(uint16(curveSpeed))) // Speed32(curveSpeedFraction × curveSpeed raw)
	target = min(int16((frac+mph)/65536), target)
	if head.Flags38&flags38FasterCurves != 0 {
		target += target / 4
		target = min(curveSpeed, target)
	}
	vmax := veh2.MaxSpeed
	if veh2.Var73&1 != 0 {
		vmax /= 4
	}
	target = min(target, vmax)
	if onRack {
		target = min(target, veh2.RackRailMaxSpeed)
	}
	if !manual {
		// upstream keeps the original's bug: var_3C >> 15 as a Speed16
		target = min(target, int16(uint32(veh1.Var3C)>>15)+5)
	}
	if (manual && head.ManualPower <= -20) || head.VehicleFlags&VehicleFlagCommandStop != 0 {
		if veh2.CurrentSpeed == 0 {
			target = 0
		}
	}
	return target
}

// Vehicle1RailMove is the move that ends updateRail: by
// min(currentSpeed/2 − var_3C, max(unkDistance1·4, 0xCC48)); var_3C grows by
// what was travelled. It returns the motion flags (noRouteFound: upstream
// then calls the head's sub_4AD93A).
func (m *Map) Vehicle1RailMove(c *MotionComponent, veh2CurrentSpeed int32, unkDistance1 int32, var3C *int32, requiredMods uint8) uint32 {
	d := veh2CurrentSpeed/2 - *var3C
	d = min(d, max(unkDistance1*4, 0xCC48))
	res := c.UpdateMotion(d, func(c *MotionComponent, f *uint32) bool {
		return m.TrackNewPiece(c, f, requiredMods, false, 0)
	})
	*var3C += d - res.RemainingDistance
	return res.Flags
}
