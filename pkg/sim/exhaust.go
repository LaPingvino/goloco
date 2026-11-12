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
// PARTIAL: the chuff sounds (no sound queue in the sim); the bodies of ships
// and aircraft (shipWakeAnimationUpdate, diesel exhaust on ships) — goloco's
// air and water ticks don't run VehicleBody::update yet.

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
	if a.Type == objects.EmitterNone || a.Type == objects.EmitterShipWake {
		return
	}
	body, front, back := &m.Entities[cc.Body], &m.Entities[cc.Front], &m.Entities[cc.Back]
	if front.BreakdownFlags&breakdownBrokenDown != 0 {
		return
	}
	v2 := &m.Entities[tr.Veh2]
	reversed := body.Flags38&f38IsReversed != 0
	steamID := uint8(om.SteamObjectIndex(a.SteamName)) // objectId (0xFF: not loaded)
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
// var_44 step (an eighth of the distance, against the body's direction).
// PARTIAL: the rest of sub_4AAB0B (speed animation and tilt frames).
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
}
