package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The driving sounds' state: VehicleHead::updateDrivingSounds (0x004A8882)
// and its four kinds (OpenLoco reference: src/OpenLoco/src/Vehicles/
// VehicleHead.cpp). Each tick the head sets Vehicle2's and the tail's
// VehicleSound — the sound object, a volume (0–255) and a frequency — from
// the vehicle object that sounds there (VehicleSound::objectId, set by
// updateTrainProperties) and the motor; the game plays them
// (Audio/VehicleAudio.cpp, cmd/goloco vehicleaudio.go). The state is saved.

const soundObjectIDNull = 0xFF // SoundObjectId::null

// initVehicleSound is CreateVehicle.cpp's createVehicle2 / createVehicleTail
// sound: no driving sound, no object, no handle.
func initVehicleSound(e *scenario.VehicleEntity) {
	e.DrivingSoundID = soundObjectIDNull
	e.SoundObjectID, e.SoundFlags = 0xFFFF, 0
	e.SoundHandle, e.ActiveSoundID = 0, soundObjectIDNull
}

// updateDrivingSounds is VehicleHead::updateDrivingSounds: Vehicle2's sound,
// then the tail's.
func (m *Map) updateDrivingSounds(tr Train, om *objects.ObjectManager) {
	m.updateDrivingSound(tr, &m.Entities[tr.Veh2], true, om)
	m.updateDrivingSound(tr, &m.Entities[tr.Tail], false, om)
}

// updateDrivingSound is VehicleHead::updateDrivingSound (0x004A88A6).
func (m *Map) updateDrivingSound(tr Train, s *scenario.VehicleEntity, isVeh2 bool, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if head.TileX == -1 || head.Status == StatusCrashed || head.Status == StatusStuck || head.Flags38&flags38IsGhost != 0 || s.SoundObjectID == 0xFFFF {
		s.DrivingSoundID = soundObjectIDNull // updateDrivingSoundNone
		return
	}
	obj := vehicleObject(om, s.SoundObjectID)
	if obj == nil {
		return
	}
	snd := &obj.DrivingSound
	slot := uint8(om.SoundObjectIndex(obj.DrivingSoundName)) // load's findObjectHandle: soundObjectId
	switch obj.DrivingSoundType {
	case objects.DrivingSoundNone:
		s.DrivingSoundID = soundObjectIDNull
	case objects.DrivingSoundFriction:
		m.updateDrivingSoundFriction(tr, s, &snd.Friction, slot)
	case objects.DrivingSoundSimpleMotor:
		m.updateSimpleMotorSound(tr, s, isVeh2, &snd.SimpleMotor, slot)
	case objects.DrivingSoundGearboxMotor:
		m.updateGearboxMotorSound(tr, s, isVeh2, &snd.Gearbox, slot)
	}
}

// updateDrivingSoundFriction is 0x004A88F7: none below minSpeed; above it the
// frequency and volume rise with the speed (the volume truncated to a byte,
// then capped at maxVolume, as std::min<uint8_t>).
func (m *Map) updateDrivingSoundFriction(tr Train, s *scenario.VehicleEntity, snd *objects.VehicleFrictionSound, slot uint8) {
	v2 := &m.Entities[tr.Veh2]
	if v2.CurrentSpeed < snd.MinSpeed {
		s.DrivingSoundID = soundObjectIDNull
		return
	}
	diff := v2.CurrentSpeed - snd.MinSpeed
	s.DrivingSoundFrequency = uint16((diff >> snd.SpeedFreqFactor) + int32(snd.BaseFrequency))
	volume := (diff >> snd.SpeedVolumeFactor) + int32(snd.BaseVolume)
	s.DrivingSoundVolume = min(uint8(volume), snd.MaxVolume)
	s.DrivingSoundID = slot
}

// motorBrokenDown is the simple and gearbox motors' first check: Vehicle2's
// sound of a train (not a ship or aircraft) whose first car is broken down
// is none.
func (m *Map) motorBrokenDown(tr Train, isVeh2 bool) bool {
	if !isVeh2 {
		return false
	}
	vt := m.Entities[tr.Head].VehicleType
	if vt == uint8(objects.VehicleTypeShip) || vt == uint8(objects.VehicleTypeAircraft) {
		return false
	}
	return len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 && m.Entities[tr.Cars[0][0].Front].BreakdownFlags&breakdownBrokenDown != 0
}

// firstCarSlipping is train.cars.firstCar.front->wheelSlipping != 0.
func (m *Map) firstCarSlipping(tr Train) bool {
	return len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 && m.Entities[tr.Cars[0][0].Front].WheelSlipping != 0
}

// stepDrivingSound is the motors' shared ending: a sound just started takes
// half the idle volume and a quarter of the idle frequency; else the
// frequency and volume step towards their targets (uint16 / uint8 arithmetic
// as upstream's std::max<uint16_t>(target, freq − step): a step past zero
// wraps). decreaseVolume is the gearbox's: a falling frequency aims the
// volume at the coasting volume.
func stepDrivingSound(s *scenario.VehicleEntity, slot uint8, idleFreq uint16, idleVol uint8, targetFreq uint16, targetVol uint8,
	freqInc, freqDec uint16, volInc, volDec uint8, fallingVol *uint8) {
	if s.DrivingSoundID == soundObjectIDNull {
		s.DrivingSoundVolume = idleVol >> 1
		s.DrivingSoundFrequency = idleFreq >> 2
		s.DrivingSoundID = slot
		return
	}
	if s.DrivingSoundFrequency != targetFreq {
		if s.DrivingSoundFrequency > targetFreq {
			if fallingVol != nil {
				targetVol = *fallingVol
			}
			s.DrivingSoundFrequency = max(targetFreq, uint16(int32(s.DrivingSoundFrequency)-int32(freqDec)))
		} else {
			s.DrivingSoundFrequency = min(targetFreq, uint16(int32(s.DrivingSoundFrequency)+int32(freqInc)))
		}
	}
	if s.DrivingSoundVolume != targetVol {
		if s.DrivingSoundVolume > targetVol {
			s.DrivingSoundVolume = max(targetVol, uint8(int32(s.DrivingSoundVolume)-int32(volDec)))
		} else {
			s.DrivingSoundVolume = min(targetVol, uint8(int32(s.DrivingSoundVolume)+int32(volInc)))
		}
	}
	s.DrivingSoundID = slot
}

// updateSimpleMotorSound is 0x004A8937: idle; accelerating (and, for
// Vehicle2, not slipping) the acceleration volume at a frequency rising with
// the speed; coasting at 12 mph or more the coasting sound.
func (m *Map) updateSimpleMotorSound(tr Train, s *scenario.VehicleEntity, isVeh2 bool, snd *objects.VehicleSimpleMotorSound, slot uint8) {
	if m.motorBrokenDown(tr, isVeh2) {
		s.DrivingSoundID = soundObjectIDNull
		return
	}
	v2 := &m.Entities[tr.Veh2]
	targetFreq, targetVol := snd.IdleFrequency, snd.IdleVolume
	if v2.MotorState == MotorAccelerating && (!isVeh2 || !m.firstCarSlipping(tr)) {
		targetFreq = snd.AccelerationBaseFreq + uint16(v2.CurrentSpeed>>snd.SpeedFreqFactor)
		targetVol = snd.AccelerationVolume
	} else if v2.MotorState == MotorCoasting && v2.CurrentSpeed >= 12<<16 {
		targetFreq, targetVol = snd.CoastingFrequency, snd.CoastingVolume
	}
	stepDrivingSound(s, slot, snd.IdleFrequency, snd.IdleVolume, targetFreq, targetVol,
		snd.FreqIncreaseStep, snd.FreqDecreaseStep, snd.VolumeIncreaseStep, snd.VolumeDecreaseStep, nil)
}

// updateGearboxMotorSound is 0x004A8A39: coasting or braking the coasting
// volume (in gear from 12 mph); accelerating in gear; else idle. In gear the
// frequency is the gear's (the first gear's less each higher gear's offset,
// by the speed, at least 7 mph) plus the speed's share — Vehicle2 slipping
// idles instead.
func (m *Map) updateGearboxMotorSound(tr Train, s *scenario.VehicleEntity, isVeh2 bool, snd *objects.VehicleGearboxMotorSound, slot uint8) {
	if m.motorBrokenDown(tr, isVeh2) {
		s.DrivingSoundID = soundObjectIDNull
		return
	}
	v2 := &m.Entities[tr.Veh2]
	var targetFreq uint16
	var targetVol uint8
	inGear := v2.MotorState == MotorAccelerating
	switch {
	case v2.MotorState == MotorCoasting || v2.MotorState == MotorBraking:
		targetVol = snd.CoastingVolume
		if v2.CurrentSpeed < 12<<16 {
			targetFreq = snd.IdleFrequency
		} else {
			inGear = true
		}
	case v2.MotorState == MotorAccelerating:
		targetVol = snd.AcceleratingVolume
	default:
		targetFreq, targetVol = snd.IdleFrequency, snd.IdleVolume
	}
	if inGear {
		if !isVeh2 || !m.firstCarSlipping(tr) {
			speed := max(v2.CurrentSpeed, 7<<16)
			freq := snd.FirstGearFrequency
			if speed >= int32(snd.FirstGearSpeed)<<16 {
				freq -= snd.SecondGearFreqOffset
				if speed >= int32(snd.SecondGearSpeed)<<16 {
					freq -= snd.ThirdGearFreqOffset
					if speed >= int32(snd.ThirdGearSpeed)<<16 {
						freq -= snd.FourthGearFreqOffset
					}
				}
			}
			targetFreq = uint16(speed>>snd.SpeedFreqFactor) + freq
		} else {
			targetFreq, targetVol = snd.IdleFrequency, snd.IdleVolume
		}
	}
	coasting := snd.CoastingVolume
	stepDrivingSound(s, slot, snd.IdleFrequency, snd.IdleVolume, targetFreq, targetVol,
		snd.FreqIncreaseStep, snd.FreqDecreaseStep, snd.VolumeIncreaseStep, snd.VolumeDecreaseStep, &coasting)
}

// updateSoundingObjects is updateTrainProperties' last part: the first car
// with a driving sound sounds at Vehicle2, the last other one at the tail;
// a train turned round (the front's object at the tail) swaps the two
// sounds, their handles with them.
func (m *Map) updateSoundingObjects(tr Train, om *objects.ObjectManager) {
	front, back := uint16(0xFFFF), uint16(0xFFFF)
	for _, car := range tr.Cars {
		if len(car) == 0 {
			continue
		}
		id := m.Entities[car[0].Body].ObjectID
		obj := vehicleObject(om, id)
		if obj == nil || obj.DrivingSoundType == objects.DrivingSoundNone {
			continue
		}
		if front == 0xFFFF {
			front = id
		}
		if front != id {
			back = id
		}
	}
	v2, tail := &m.Entities[tr.Veh2], &m.Entities[tr.Tail]
	if front != v2.SoundObjectID && front == tail.SoundObjectID {
		v2.DrivingSoundID, tail.DrivingSoundID = tail.DrivingSoundID, v2.DrivingSoundID
		v2.DrivingSoundVolume, tail.DrivingSoundVolume = tail.DrivingSoundVolume, v2.DrivingSoundVolume
		v2.DrivingSoundFrequency, tail.DrivingSoundFrequency = tail.DrivingSoundFrequency, v2.DrivingSoundFrequency
		v2.SoundHandle, tail.SoundHandle = tail.SoundHandle, v2.SoundHandle
		v2.SoundFlags, tail.SoundFlags = tail.SoundFlags, v2.SoundFlags
		v2.ActiveSoundID, tail.ActiveSoundID = tail.ActiveSoundID, v2.ActiveSoundID
	}
	v2.SoundObjectID, tail.SoundObjectID = front, back
}
