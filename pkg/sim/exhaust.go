package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Vehicle smoke, steam, sparks and wakes: Effects/ExhaustEffect.cpp (an
// Exhaust entity per puff, a steam object's frames rising) made by
// VehicleBody::animationUpdate (VehicleBody.cpp 0x004AAC4E) and
// secondaryAnimationUpdate, one emitter each (animation[0], animation[1]).
// Ships and aircraft run animationUpdate from their ticks (VehicleBody::update
// for air and water: animationUpdate only): the ship's wakes and diesel
// exhaust. PARTIAL: the chuff sounds (no sound queue in the sim).

const effectTypeExhaust = 0 // EffectType::exhaust

// EffectExhaust is EffectType::exhaust for the scene.
const EffectExhaust = effectTypeExhaust

const breakdownBrokenDown = 1 << 2 // BreakdownFlags::brokenDown

// createExhaust is Exhaust::create (0x0044080C): off the map, or at or below
// the ground, none. objectID is the steam object's slot, 0x80 its second
// frame run (isSubObjType1).
func (m *Map) createExhaust(om *objects.ObjectManager, x, y, z int32, objectID uint8) {
	if x < 0 || y < 0 || x >= int32(m.Width)*32 || y >= int32(m.Height)*32 {
		return
	}
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil || z <= int32(t.BaseZ)*smallZStep {
		return
	}
	obj := steamObject(om, objectID)
	if obj == nil {
		return
	}
	m.newEffect(Effect{X: x, Y: y, Z: z, SubType: effectTypeExhaust, ObjectID: objectID,
		Width: obj.SpriteWidth, HeightNeg: obj.SpriteHeightNeg, HeightPos: obj.SpriteHeightPos})
}

// steamObject is Exhaust::getObject: the steam object in slot objectId & 0x7F.
func steamObject(om *objects.ObjectManager, objectID uint8) *objects.SteamObject {
	if om == nil || int(objectID&0x7F) >= len(om.SteamObjects) {
		return nil
	}
	return om.SteamObjects[objectID&0x7F]
}

// SteamObjectOf is the scene's getObject for an exhaust.
func SteamObjectOf(om *objects.ObjectManager, e *Effect) *objects.SteamObject {
	return steamObject(om, e.ObjectID)
}

// updateExhaust is Exhaust::update (0x004408C2): a pixel of wind along x
// about every tenth tick; each numStationaryTicks a frame on, rising by its
// height, until the run ends; a dispersing puff ends where it meets an
// element (or a station canopy).
func (m *Map) updateExhaust(om *objects.ObjectManager, id uint16, e *Effect) {
	obj := steamObject(om, e.ObjectID)
	if obj == nil {
		m.freeEffect(id)
		return
	}
	if obj.Flags&objects.SteamFlagApplyWind != 0 {
		res := uint32(e.Wind) + 7000
		e.Wind = uint16(res)
		if dx := int32(res >> 16); dx != 0 {
			m.moveEffect(e, e.X+dx, e.Y, e.Z)
		}
	}
	e.Stationary++
	if e.Stationary < obj.NumStationaryTicks {
		return
	}
	e.Stationary = 0
	e.Frame++
	run := obj.Frames[(e.ObjectID>>7)&1]
	if int(e.Frame) >= len(run) {
		m.freeEffect(id)
		return
	}
	m.moveEffect(e, e.X, e.Y, e.Z+int32(run[e.Frame].Height))
	if obj.Flags&objects.SteamFlagDisperseOnCollision != 0 && m.exhaustMeetsElement(om, e) {
		m.freeEffect(id)
	}
}

// exhaustMeetsElement is update's disperseOnCollision walk: an element whose
// base lies within 3 small steps of the puff, or the canopy (track base + 7)
// of a station that has one.
func (m *Map) exhaustMeetsElement(om *objects.ObjectManager, e *Effect) bool {
	t := m.TileAt(int(e.X>>5), int(e.Y>>5))
	if t == nil {
		return false
	}
	lowZ := int(e.Z/smallZStep) - 3
	highZ := lowZ + 6
	const ghostOrAI = elemFlagGhost | elemFlagAiAllocated
	for k, r := range t.Order {
		if r.Flags&ghostOrAI != 0 {
			continue
		}
		if bz := int(t.baseZOf(r)); lowZ < bz && highZ > bz {
			return true
		}
		if r.Kind != scenario.KindTrack || !t.Tracks[r.Index].HasStation || k+1 >= len(t.Order) {
			continue
		}
		sr := t.Order[k+1]
		if sr.Kind != scenario.KindStation || sr.Flags&ghostOrAI != 0 {
			continue
		}
		st := t.Stations[sr.Index]
		so := om.GetTrainStationObjectByIndex(int(st.ObjectID))
		if so == nil || so.Flags&objects.TrainStationFlagNoCanopy != 0 || st.SequenceIndex == 0 {
			continue
		}
		if cmp := int(t.Tracks[r.Index].BaseZ) + 7; lowZ < cmp && highZ > cmp {
			return true
		}
	}
	return false
}

// bodyAnimationUpdate is VehicleBody::animationUpdate with
// secondaryAnimationUpdate for body i of car component cc: none for a ghost,
// a crashed or stuck train, or a car without an emitter position.
func (m *Map) bodyAnimationUpdate(tr Train, cc CarComponent, om *objects.ObjectManager, unkDistance int32) {
	body := &m.Entities[cc.Body]
	if body.Flags38&flags38IsGhost != 0 {
		return
	}
	if h := &m.Entities[tr.Head]; h.Status == StatusCrashed || h.Status == StatusStuck {
		return
	}
	v := vehicleObject(om, body.ObjectID)
	if v == nil || int(body.BodyIndex) >= len(v.CarComponents) {
		return
	}
	pos := v.CarComponents[body.BodyIndex].EmitterHorizontalPos
	if pos == 0 {
		return
	}
	// animationUpdate's position is signed (int32 − 0x80); secondaryAnimationUpdate's
	// stays a uint8 (−0x80 wraps), as upstream.
	m.emitterUpdate(tr, cc, om, v, 0, int32(pos)-0x80, unkDistance)
	m.emitterUpdate(tr, cc, om, v, 1, int32(pos-0x80), unkDistance)
}

// emitterUpdate runs animation[num]'s kind: steamPuffsAnimationUpdate,
// dieselExhaust1/2AnimationUpdate, electricSpark1/2AnimationUpdate.
func (m *Map) emitterUpdate(tr Train, cc CarComponent, om *objects.ObjectManager, v *objects.VehicleObject, num int, hpos, unkDistance int32) {
	a := v.Animations[num]
	if a.Type == objects.EmitterNone {
		return
	}
	body, front, back := &m.Entities[cc.Body], &m.Entities[cc.Front], &m.Entities[cc.Back]
	v2 := &m.Entities[tr.Veh2]
	steamID := uint8(om.SteamObjectIndex(a.SteamName)) // objectId (0xFF: not loaded)
	if a.Type == objects.EmitterShipWake {
		m.shipWake(om, body, v, v2, steamID)
		return
	}
	if front.BreakdownFlags&breakdownBrokenDown != 0 {
		return
	}
	reversed := body.Flags38&f38IsReversed != 0
	var44 := int32(int16(body.Var44))
	if reversed {
		hpos, var44 = -hpos, -var44
	}
	ticks := m.ScenarioTicks
	// smokeLoc: along the bogies by the horizontal position, up by the vertical
	// one (and back by the body's pitch).
	loc := func() (int32, int32, int32) {
		p := hpos + 64
		dx, dy, dz := int32(back.PosX)-int32(front.PosX), int32(back.PosY)-int32(front.PosY), int32(back.PosZ)-int32(front.PosZ)
		xyMag := int32(olmath.PitchHorizontalFactor[body.SpritePitch]) * int32(a.VerticalPos) / 256 // computeXYMagnitude
		fx, fy := olmath.ComputeXYVector(xyMag, body.SpriteYaw)
		return dx*p/128 + int32(front.PosX) + fx, dy*p/128 + int32(front.PosY) + fy, dz*p/128 + int32(front.PosZ) + int32(a.VerticalPos)
	}
	switch a.Type {
	case objects.EmitterSteamPuff1, objects.EmitterSteamPuff2, objects.EmitterSteamPuff3:
		working := v2.MotorState == MotorAccelerating || v2.MotorState == MotorStoppedOnIncline // soundCode
		idle := v2.MotorState == MotorStopped || v2.CurrentSpeed < 1<<16                        // tickCalc
		if idle && !working {
			if ticks&7 != 0 {
				return
			}
		} else if unkDistance+int32(uint16(var44*8)) < 0xFFFF {
			return
		}
		x, y, z := loc()
		id := steamID
		if !working {
			id |= 0x80
		}
		m.createExhaust(om, x, y, z, id)
		if working { // chuffSoundIndex (PARTIAL: the sound itself)
			body.ChuffSoundIndex++
			if body.ChuffSoundIndex >= a.Type+1 {
				body.ChuffSoundIndex = 0
			}
		}
	case objects.EmitterDieselExhaust1:
		if m.Entities[tr.Head].VehicleType == vehicleTypeShip { // a ship: back along the hull from its position
			if v2.CurrentSpeed == 0 || ticks&3 != 0 {
				return
			}
			f := int32(0)
			if len(v.BodySprites) > 0 {
				f = int32(v.BodySprites[0].HalfLength) * hpos / 256
			}
			dx, dy := olmath.ComputeXYVector(f, body.SpriteYaw^(1<<5))
			m.createExhaust(om, int32(body.PosX)+dx/2, int32(body.PosY)+dy/2, int32(body.PosZ)+int32(a.VerticalPos), steamID)
			return
		}
		if v2.MotorState != MotorAccelerating || ticks&3 != 0 || body.WheelSlipping != 0 {
			return
		}
		x, y, z := loc()
		m.createExhaust(om, x, y, z, steamID)
	case objects.EmitterDieselExhaust2:
		if v2.MotorState != MotorAccelerating || v2.CurrentSpeed > 14<<16 || ticks&7 != 0 {
			return
		}
		x, y, z := loc()
		side := int32(5)
		if m.TrafficHandedness != 0 {
			side = -5
		}
		sx, sy := olmath.ComputeXYVector(side, (body.SpriteYaw+16)&0x3F) // 90° clockwise
		m.createExhaust(om, x+sx, y+sy, z, steamID)
	case objects.EmitterElectricSpark1, objects.EmitterElectricSpark2:
		if v2.MotorState != MotorCoasting && v2.MotorState != MotorAccelerating {
			return
		}
		if int32(uint16(unkDistance))+int32(uint16(var44))*8 < 0xFFFF {
			return
		}
		limit := uint32(819)
		if a.Type == objects.EmitterElectricSpark2 {
			limit = 936
		}
		if m.Rng1.RandNextBound(0xFFFF) > int32(limit) { // randNext(0xFFFF)
			return
		}
		x, y, z := loc()
		if a.Type == objects.EmitterElectricSpark2 {
			first := front
			if reversed {
				first = back
			}
			side := int32(5)
			if id := int(first.TrackAndDirection >> 3); id >= len(TrackSparkDirection) || !TrackSparkDirection[id] {
				side = -5
			}
			if first.TrackAndDirection&(1<<2) != 0 {
				side = -side
			}
			sx, sy := olmath.ComputeXYVector(side, (body.SpriteYaw+16)&0x3F)
			x, y = x+sx, y+sy
		}
		m.createExhaust(om, x, y, z, steamID)
	}
}

// bodyUpdateAnimation is VehicleBody::update after sub_4AC255: a slipping
// wheel adds to the distance (unkDistance2 + slip·320 + 500, the slip
// counted back down past half), then animationUpdate, then sub_4AAB0B's
// var_44 step (an eighth of the distance, against the body's direction) and
// the rest of sub_4AAB0B: the body's animation frame (bodyAnimationFrame).
func (m *Map) bodyUpdateAnimation(tr Train, cc CarComponent, om *objects.ObjectManager, unkDistance int32) {
	body := &m.Entities[cc.Body]
	if ws := int32(body.WheelSlipping); ws != 0 {
		if ws > WheelSlippingDuration/2 {
			ws = WheelSlippingDuration - ws
		}
		unkDistance += ws*320 + 500
	}
	m.bodyAnimationUpdate(tr, cc, om, unkDistance)
	step := unkDistance >> 3
	if body.Flags38&f38IsReversed != 0 {
		step = -step
	}
	body.Var44 += uint16(step)
	m.bodyAnimationFrame(tr, cc, om)
}

// bodyAnimationFrame is sub_4AAB0B's frame: a body with a speed animation
// shows the frame its speed reaches of the vehicle's top speed (speed /
// (top / frames), the last at most); a rail body with roll frames tilts into
// the curve above 35 mph (kCurvatureDegrees by the front bogie's track, an
// S-bend's second half the other way; frame 1 or 2 by side and direction,
// through 0 when changing sides); any other the wheel frame (var_44 >> 12).
func (m *Map) bodyAnimationFrame(tr Train, cc CarComponent, om *objects.ObjectManager) {
	body := &m.Entities[cc.Body]
	v := vehicleObject(om, body.ObjectID)
	if body.ObjectSpriteType == 0xFF || v == nil || int(body.ObjectSpriteType) >= len(v.BodySprites) {
		return
	}
	sp := v.BodySprites[body.ObjectSpriteType]
	v2 := &m.Entities[tr.Veh2]
	var target uint8
	switch {
	case sp.Flags&objects.BodyFlagHasSpeedAnimation != 0:
		if n := sp.NumAnimationFrames; n > 0 {
			if per := int64(v.Speed / uint16(n)); per > 0 { // Speed32 / Speed16: both in 16.16
				target = min(uint8(int64(v2.CurrentSpeed)/(per<<16)), n-1)
			}
		}
	case sp.NumRollFrames != 1 && body.Mode == 0: // TransportMode::rail
		target = body.Animation
		var tilt int8
		if v2.CurrentSpeed >= 35<<16 {
			front := &m.Entities[cc.Front]
			if i := int(front.TrackAndDirection >> 2); i < len(TrackCurvatureDegree) {
				tilt = TrackCurvatureDegree[i]
			}
			if id := (front.TrackAndDirection >> 3) & 0x3F; (id == 12 || id == 13) && front.SubPosition >= 48 {
				tilt = -tilt // the S-bend's second half
			}
			rev := body.Flags38&f38IsReversed != 0
			switch {
			case tilt < 0 && rev, tilt > 0 && !rev:
				tilt = 2
			case tilt != 0:
				tilt = 1
			}
			if tilt != 0 && target != 0 && target != uint8(tilt) { // changing sides: through 0
				tilt = 0
			}
		}
		target = uint8(tilt)
	default:
		if n := sp.NumAnimationFrames; n > 0 {
			target = uint8(int16(body.Var44)>>12) & (n - 1)
		}
	}
	if body.Animation != target {
		body.Animation = target
		m.invalidateSprite(body)
	}
}

const vehicleTypeShip = 5 // VehicleType::ship

// shipWake is shipWakeAnimationUpdate (0x004ABC8A): moving at 6 mph or more,
// every 32, 16 (9 mph), 8 (13 mph) or 4 (25 mph) ticks a wake a quarter of
// the hull back from the ship, half the wake spacing to its right; with a
// spacing, a second half of it back to the left (from there, as upstream).
func (m *Map) shipWake(om *objects.ObjectManager, body *scenario.VehicleEntity, v *objects.VehicleObject, v2 *scenario.VehicleEntity, steamID uint8) {
	const mph = 1 << 16
	speed := v2.CurrentSpeed
	if v2.MotorState == MotorStopped || speed < 6*mph {
		return
	}
	freq := uint32(32)
	switch {
	case speed >= 25*mph:
		freq = 4
	case speed >= 13*mph:
		freq = 8
	case speed >= 9*mph:
		freq = 16
	}
	if m.ScenarioTicks%freq != 0 {
		return
	}
	half := int32(0)
	if len(v.BodySprites) > 0 {
		half = int32(v.BodySprites[0].HalfLength)
	}
	dx, dy := olmath.ComputeXYVector(half, body.SpriteYaw^(1<<5))
	x, y, z := int32(body.PosX)+dx/4, int32(body.PosY)+dy/4, int32(body.PosZ)
	dx, dy = olmath.ComputeXYVector(int32(v.ShipWakeSpacing), (body.SpriteYaw+16)&0x3F) // 90° clockwise
	x, y = x+dx/2, y+dy/2
	m.createExhaust(om, x, y, z, steamID)
	if v.ShipWakeSpacing == 0 {
		return
	}
	dx, dy = olmath.ComputeXYVector(int32(v.ShipWakeSpacing), (body.SpriteYaw-16)&0x3F) // 90° anticlockwise
	m.createExhaust(om, x+dx/2, y+dy/2, z, steamID)
}

// bodiesAnimate is VehicleBody::update for an air or water vehicle: each
// body's animationUpdate with unkDistance2, nothing else.
func (m *Map) bodiesAnimate(tr Train, om *objects.ObjectManager) {
	_, d2 := UpdateDistances(m.Entities[tr.Veh2].CurrentSpeed)
	for _, car := range tr.Cars {
		for _, cc := range car {
			m.bodyAnimationUpdate(tr, cc, om, d2)
		}
	}
}
