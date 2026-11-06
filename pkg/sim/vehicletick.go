package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The vehicle tick — VehicleHead::updateVehicle's order over the saved
// entities: head (sets the update distances from Vehicle2's speed), Vehicle1,
// Vehicle2 (speed, then sub_4A9F20 which trims the distances the rest use),
// each car's front bogie, back bogie and body, then the tail.
//
// PARTIAL: of the head's own logic (updateLand) the normal-movement path runs
// (headupdate.go: orders, sub_4ACEF1 / sub_47DA8D reservation, approaching a
// station, waiting at a signal); stations, stopping, reversing are not
// (5.6/5.7): a vehicle that reaches a station stays approaching, one at a
// signal waits without reversing. Air and water vehicles are not updated.
// On noRouteFound upstream destroys the train or calls sub_4AD93A; here that
// train's update stops for the tick.

func (m *Map) component(i int) MotionComponent {
	e := &m.Entities[i]
	c := MotionComponent{
		Road: e.Mode == 1, RightHand: m.TrafficHandedness != 0,
		TrackAndDirection: e.TrackAndDirection, SubPosition: e.SubPosition,
		TileX: int32(e.TileX), TileY: int32(e.TileY), TileBaseZ: e.TileBaseZ,
		RemainingDistance: e.RemainingDistance, X: int32(e.PosX), Y: int32(e.PosY), Z: int32(e.PosZ),
		SpriteYaw: e.SpriteYaw, SpritePitch: e.SpritePitch,
		RoutingHandle: RoutingHandle(e.RoutingHandle), Owner: e.Owner, TrackType: e.TrackType,
		IsVehicle2: e.EntityType == scenario.VehicleTypeVeh2,
	}
	if !c.Road {
		c.TrackAndDirection &= TaDBasicMask
	}
	return c
}

func (m *Map) store(i int, c *MotionComponent) {
	e := &m.Entities[i]
	if c.Road {
		e.TrackAndDirection = c.TrackAndDirection
	} else {
		e.TrackAndDirection = e.TrackAndDirection&^TaDBasicMask | c.TrackAndDirection&TaDBasicMask
	}
	e.SubPosition = c.SubPosition
	e.TileX, e.TileY, e.TileBaseZ = coords.WorldUnits(c.TileX), coords.WorldUnits(c.TileY), c.TileBaseZ
	e.RemainingDistance = c.RemainingDistance
	if int32(e.PosX) != c.X || int32(e.PosY) != c.Y || int32(e.PosZ) != c.Z { // updateTrackMotion: moveTo when it moved
		m.moveTo(e, c.X, c.Y, c.Z)
	}
	e.SpriteYaw, e.SpritePitch = c.SpriteYaw, c.SpritePitch
	e.RoutingHandle = uint16(c.RoutingHandle)
	e.TrackType = c.TrackType
	if c.IsVehicle2 {
		e.Var4F = int8(c.Var4F)
	}
}

// UpdateVehicles runs one vehicle tick for every land vehicle.
func (m *Map) UpdateVehicles(om *objects.ObjectManager) {
	m.ScenarioTicks++            // ScenarioManager: one scenario tick per game tick
	m.dateTick(om)               // tickLogic: dateTick before the managers' updates
	m.UpdateTowns(om)            // TownManager::update (towngrowroad.go)
	defer m.UpdateAnimations(om) // AnimationManager::update: after the vehicles (tickLogic's order)
	trains, err := m.Trains()
	if err != nil {
		return
	}
	for _, tr := range trains {
		h := &m.Entities[tr.Head]
		if h.TileX == -1 {
			continue
		}
		switch h.Mode {
		case 0, 1:
			m.updateTrain(tr, om)
		case 2: // air (airmotion.go)
			m.updateAircraft(tr, om)
		case 3: // water (watermotion.go)
			m.updateBoat(tr, om)
		}
	}
}

func (m *Map) updateTrain(tr Train, om *objects.ObjectManager) {
	head, v1, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh1], &m.Entities[tr.Veh2]
	road := head.Mode == 1
	// VehicleHead::update: the distances (head logic: see PARTIAL).
	d1, d2 := UpdateDistances(v2.CurrentSpeed)
	d := VehicleUpdateDistances{D1: d1, D2: d2}
	if head.Var5C != 0 { // VehicleHead::update
		head.Var5C--
	}
	m.updateBreakdown(tr, om)
	// VehicleHead::updateLand (headupdate.go; its PARTIAL states do nothing).
	m.updateHead(tr, d.D1, om)

	// Vehicle1::update.
	if road {
		v1.TargetSpeed = m.Vehicle1RoadTargetSpeed(head, v1, v2, om)
		v := &Veh1{MotionComponent: m.component(tr.Veh1), Var3C: v1.Var3C, Veh2MaxSpeed: v2.MaxSpeed,
			HeadVar3C: &head.Var3C, HeadTaD: &head.TrackAndDirection,
			HeadVar52: head.HeadVar52, HeadTrackType: head.TrackType, HeadVar53: head.HeadVar53}
		f := m.Vehicle1RoadMove(v, v2.CurrentSpeed, d.D1, om)
		m.store(tr.Veh1, &v.MotionComponent)
		v1.Var3C = v.Var3C
		if f&MotionNoRouteFound != 0 { // Vehicle1::update: sub_4AD93A, stop approaching
			m.Sub4AD93A(tr)
			if head.Status == StatusApproaching {
				head.Status = StatusTravelling
			}
		}
	} else {
		v1.TargetSpeed = m.Vehicle1RailTargetSpeed(head, v1, v2, om)
		c := m.component(tr.Veh1)
		f := m.Vehicle1RailMove(&c, v2.CurrentSpeed, d.D1, &v1.Var3C, head.HeadVar53)
		m.store(tr.Veh1, &c)
		if f&MotionNoRouteFound != 0 { // Vehicle1::update: sub_4AD93A, stop approaching
			m.Sub4AD93A(tr)
			if head.Status == StatusApproaching {
				head.Status = StatusTravelling
			}
		}
	}

	// Vehicle2::update: speed, then sub_4A9F20.
	cars := m.CarPhysicsOf(tr, om)
	sp := Vehicle2Speed{CurrentSpeed: v2.CurrentSpeed, MotorState: v2.MotorState}
	sp.UpdateSpeed(v1.TargetSpeed, head.VehicleFlags, head.ManualPower, v2.TotalPower, v2.TotalWeight, v2.Var73&1 != 0, cars)
	for ci, car := range tr.Cars {
		for _, cc := range car {
			for _, i := range []int{cc.Front, cc.Back, cc.Body} {
				m.Entities[i].WheelSlipping = cars[ci].WheelSlipping
			}
		}
	}
	c2 := m.component(tr.Veh2)
	c2.Var4F = uint8(v2.Var4F)
	brakeWas := v2.BrakeLightTimeout
	ok := m.Vehicle2Move(&c2, &sp, &d, head.VehicleFlags&VehicleFlagManualControl != 0, RoutingHandle(v1.RoutingHandle),
		head.HeadVar53, &head.Var3C, &v1.Var3C, &v2.BrakeLightTimeout)
	m.store(tr.Veh2, &c2)
	v2.CurrentSpeed, v2.MotorState = sp.CurrentSpeed, sp.MotorState
	if (brakeWas == 0) != (v2.BrakeLightTimeout == 0) {
		// Vehicle2::update: the brake lights going on or off change how
		// every part is drawn — the whole train is invalidated.
		m.invalidateTrain(tr)
	}
	if !ok {
		return
	}

	// Cars: front bogie, back bogie, body.
	for _, car := range tr.Cars {
		for _, cc := range car {
			moved := false
			for _, i := range []int{cc.Front, cc.Back} {
				e := &m.Entities[i]
				c := m.component(i)
				x0, y0, z0 := c.X, c.Y, c.Z
				roll := BogieRoll{Var44: e.Var44, WheelSlipping: e.WheelSlipping, Reversed: e.Flags38&2 != 0, AnimationIndex: e.Animation}
				if veh := vehicleObject(om, e.ObjectID); veh != nil && e.ObjectSpriteType != 0xFF {
					roll.NumAnimationFrames = veh.BogieSprites[e.ObjectSpriteType&1].NumAnimationFrames
				}
				ok := m.BogieUpdate(&c, &roll, d.D1, d.D2, head.HeadVar53)
				m.store(i, &c)
				animWas := e.Animation
				e.Var44, e.Animation = roll.Var44, roll.AnimationIndex
				if e.Animation != animWas { // VehicleBogie::updateRoll: a new wheel frame
					m.invalidateSprite(e)
				}
				if !ok {
					return
				}
				moved = moved || c.X != x0 || c.Y != y0 || c.Z != z0
			}
			if moved { // VehicleBody::update: sub_4AC255 when a bogie moved
				f, b, body := &m.Entities[cc.Front], &m.Entities[cc.Back], &m.Entities[cc.Body]
				m.invalidateSprite(body) // VehicleBody::update: before and after sub_4AC255 (a turn alone moves nothing)
				p := PlaceBody(int32(f.PosX), int32(f.PosY), int32(f.PosZ), int32(b.PosX), int32(b.PosY), int32(b.PosZ), body.ObjectSpriteType, vehicleObject(om, body.ObjectID))
				m.moveTo(body, p.X, p.Y, p.Z) // sub_4AC255: body->moveTo
				if p.Turned {
					body.SpritePitch, body.SpriteYaw = p.Pitch, p.Yaw
				}
				m.invalidateSprite(body)
			}
		}
	}

	// VehicleTail::update.
	ct := m.component(tr.Tail)
	m.TailUpdate(&ct, d.D1, head.HeadVar53)
	m.store(tr.Tail, &ct)
}
