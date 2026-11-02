package sim

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Rack rail and wheel slip — VehicleBogie.cpp isOnRackRail (rail 0x004AA9xx /
// road 0x004AA50) and Vehicle2.cpp shouldStartWheelSlipping (0x004A9BA0).
// Upstream compares object ids that load resolves from dependency headers
// (trackObj->mods[i], vehObj->rackRailType); goloco compares those same
// headers by name.

// Object flags used here.
const (
	vehicleFlagRackRail     uint16 = 1 << 6  // VehicleObjectFlags::rackRail
	vehicleFlagCanWheelslip uint16 = 1 << 13 // VehicleObjectFlags::canWheelslip
	trackFlagHasRackRail    uint16 = 1 << 0  // TrackObjectFlags
	trackFlagNoSlipSurface  uint16 = 1 << 1
	roadFlagNoSlipSurface   uint16 = 1 << 4 // RoadObjectFlags
	roadFlagHasRackRail     uint16 = 1 << 5
)

func sameObject(a, b string) bool {
	return a != "" && strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// IsOnRackRail is VehicleBogie::isOnRackRail.
func (m *Map) IsOnRackRail(b *scenario.VehicleEntity, om *objects.ObjectManager) bool {
	veh := vehicleObject(om, b.ObjectID)
	t := m.TileAt(int(b.TileX)>>5, int(b.TileY)>>5)
	if b.Mode == 0 {
		tr := om.GetTrackObjectByIndex(int(b.TrackType))
		if tr == nil || tr.Flags&trackFlagHasRackRail == 0 {
			return true
		}
		if veh == nil || veh.Flags&vehicleFlagRackRail == 0 || t == nil {
			return false
		}
		for _, el := range t.Tracks {
			if el.BaseZ != b.TileBaseZ || uint16(el.Rotation) != b.TrackAndDirection&3 || uint16(el.TrackID) != b.TrackAndDirection>>3&0x3F {
				continue
			}
			for i := range 4 {
				if el.Mods&(1<<i) != 0 && i < len(tr.ModNames) && sameObject(tr.ModNames[i], veh.RackRailName) {
					return true
				}
			}
		}
		return false
	}
	if b.TrackType == 0xFF {
		return true
	}
	rd := om.GetRoadObjectByIndex(int(b.TrackType))
	if rd == nil || rd.Flags&roadFlagHasRackRail == 0 {
		return true
	}
	if veh == nil || veh.Flags&vehicleFlagRackRail == 0 || t == nil {
		return false
	}
	for _, el := range t.Roads {
		if el.BaseZ != b.TileBaseZ || uint16(el.Rotation) != b.TrackAndDirection&3 || uint16(el.RoadID) != b.TrackAndDirection>>3&0xF {
			continue
		}
		for i := range 2 {
			if el.Mods&(1<<i) != 0 && i < len(rd.ModNames) && sameObject(rd.ModNames[i], veh.RackRailName) {
				return true
			}
		}
	}
	return false
}

func vehicleObject(om *objects.ObjectManager, id uint16) *objects.VehicleObject {
	if om == nil || int(id) >= len(om.Vehicles) {
		return nil
	}
	return om.Vehicles[id]
}

// Head status values used here (VehicleHead.h Status).
const StatusTravelling uint8 = 2

// ShouldStartWheelSlipping is shouldStartWheelSlipping for a car's front
// bogie. It draws from gPrng1 (m.Rng1) at the same point upstream does.
// The tutorial check is upstream's too: goloco has no tutorial (always none).
func (m *Map) ShouldStartWheelSlipping(head, veh2, front *scenario.VehicleEntity, om *objects.ObjectManager) bool {
	veh := vehicleObject(om, front.ObjectID)
	if veh == nil || veh.Flags&vehicleFlagCanWheelslip == 0 || veh.Power == 0 {
		return false
	}
	if head.Status != StatusTravelling {
		return false
	}
	if veh2.MotorState == MotorCoasting || veh2.CurrentSpeed > 10*mph {
		return false
	}
	if front.WheelSlipping != 0 {
		return false
	}
	if head.VehicleFlags&VehicleFlagManualControl != 0 && head.ManualPower <= 10 {
		return false
	}
	tot1 := uint64(128) * uint64(veh.Power) * uint64(veh2.TotalWeight)
	tot2 := uint64(front.TotalWeight) * uint64(veh2.TotalPower) // front: totalCarWeight
	fraction := tot1
	if tot2 != 0 {
		fraction = tot1 / tot2
	}
	fraction = min(fraction, 2000)
	if fraction < uint64(uint16(m.Rng1.RandNextBound(0xFFFF))) {
		return false
	}
	if front.Mode == 0 {
		tr := om.GetTrackObjectByIndex(int(front.TrackType))
		if tr != nil && tr.Flags&trackFlagNoSlipSurface != 0 {
			return false
		}
		if tr != nil && tr.Flags&trackFlagHasRackRail != 0 && m.IsOnRackRail(front, om) {
			return false
		}
		return true
	}
	if front.TrackType == 0xFF {
		return false
	}
	rd := om.GetRoadObjectByIndex(int(front.TrackType))
	if rd != nil && rd.Flags&roadFlagNoSlipSurface != 0 {
		return false
	}
	if rd != nil && rd.Flags&roadFlagHasRackRail != 0 && m.IsOnRackRail(front, om) {
		return false
	}
	return true
}

// CarPhysicsOf builds the speed step's per-car inputs for a train from its
// cars' front bogies (Vehicle2::update's loop over train.cars), deciding
// StartSlipping with shouldStartWheelSlipping as upstream does in that loop.
func (m *Map) CarPhysicsOf(tr Train, om *objects.ObjectManager) []CarPhysics {
	head, veh2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	out := make([]CarPhysics, 0, len(tr.Cars))
	for _, car := range tr.Cars {
		f := &m.Entities[car[0].Front]
		veh := vehicleObject(om, f.ObjectID)
		out = append(out, CarPhysics{
			Pitch:          f.SpritePitch,
			TotalCarWeight: f.TotalWeight,
			Powered:        veh != nil && veh.Power != 0,
			OnRackRail:     m.IsOnRackRail(f, om),
			StartSlipping:  m.ShouldStartWheelSlipping(head, veh2, f, om),
			WheelSlipping:  f.WheelSlipping,
		})
	}
	return out
}
