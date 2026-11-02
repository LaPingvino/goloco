package objects

// Body image selection, ported from Vehicles/VehicleDraw.cpp
// (getBodyImageIndex / getBrakingImageIndex and the per-pitch helpers). Image
// ids are relative to the object's ImageOffset, as BodySprites are stored.
// yaw is the entity's 0–63 sprite yaw held in a uint8, so the +64 steps wrap
// as upstream's do.

// Pitch is Entity.h Pitch (flat, up6deg … down25deg, then the curve pitches,
// which draw as flat here as upstream's default case does).
type Pitch uint8

const (
	PitchFlat Pitch = iota
	PitchUp6
	PitchUp12
	PitchUp18
	PitchUp25
	PitchDown6
	PitchDown12
	PitchDown18
	PitchDown25
)

// yawShift is _503F20: the yaw shift per yaw accuracy.
var yawShift = [8]uint8{4, 3, 2, 1, 0, 0, 0, 0}

func (s *VehicleBodySprite) has(f uint8) bool { return s.Flags&f != 0 }

func (s *VehicleBodySprite) pitchDefault(yaw uint8) uint32 {
	if s.has(BodyFlagRotationalSymmetry) {
		yaw &= 0x1F
	}
	return uint32(yaw>>yawShift[s.FlatYawAccuracy&7])*uint32(s.NumFramesPerRotation) + s.FlatImageID
}

func (s *VehicleBodySprite) sloped(yaw uint8, base uint32) uint32 {
	off := uint32(8)
	if s.has(BodyFlagRotationalSymmetry) {
		off = 4
	}
	return (uint32(yaw>>yawShift[s.SlopedYawAccuracy&7])+off)*uint32(s.NumFramesPerRotation) + base
}

func (s *VehicleBodySprite) up12(yaw uint8) uint32 {
	if !s.has(BodyFlagHasGentleSprites) {
		return s.pitchDefault(yaw)
	}
	return s.sloped(yaw, s.GentleImageID)
}

// reverse is the down-pitch yaw: the opposite half for a symmetric sprite,
// else the next 64 (the down frames follow the up ones).
func (s *VehicleBodySprite) reverse(yaw uint8) uint8 {
	if s.has(BodyFlagRotationalSymmetry) {
		return yaw ^ (1 << 5)
	}
	return yaw + (1 << 6)
}

func (s *VehicleBodySprite) down12(yaw uint8) uint32 {
	if !s.has(BodyFlagHasGentleSprites) {
		return s.pitchDefault(yaw)
	}
	return s.up12(s.reverse(yaw))
}

// transition is the 6° / 18° frames: four directions (yaw + 7) >> 4.
func (s *VehicleBodySprite) transition(yaw uint8, down bool, base uint32) uint32 {
	if down && s.has(BodyFlagRotationalSymmetry) {
		yaw ^= 1 << 5
		down = false
	}
	yaw += 7
	yaw >>= 4
	yaw &= 0x3
	if down {
		yaw += 4
	}
	return uint32(yaw)*uint32(s.NumFramesPerRotation) + base
}

func (s *VehicleBodySprite) pitchImage(p Pitch, yaw uint8) uint32 {
	switch p {
	case PitchUp12:
		return s.up12(yaw)
	case PitchDown12:
		return s.down12(yaw)
	case PitchUp6, PitchDown6:
		if !s.has(BodyFlagHasGentleSprites) {
			return s.pitchDefault(yaw)
		}
		return s.transition(yaw, p == PitchDown6, s.GentleImageID)
	case PitchUp25:
		if !s.has(BodyFlagHasSteepSprites) {
			return s.up12(yaw)
		}
		return s.sloped(yaw, s.SteepImageID)
	case PitchDown25:
		if !s.has(BodyFlagHasSteepSprites) {
			return s.down12(yaw)
		}
		return s.sloped(s.reverse(yaw), s.SteepImageID)
	case PitchUp18:
		if !s.has(BodyFlagHasSteepSprites) {
			return s.up12(yaw)
		}
		return s.transition(yaw, false, s.SteepImageID)
	case PitchDown18:
		if !s.has(BodyFlagHasSteepSprites) {
			return s.down12(yaw)
		}
		return s.transition(yaw, true, s.SteepImageID)
	}
	return s.pitchDefault(yaw)
}

// BodyImageIndex is getBodyImageIndex: the pitch's frame for yaw, plus the
// roll and cargo frame.
func (s *VehicleBodySprite) BodyImageIndex(p Pitch, yaw, roll, cargoIndex uint8) uint32 {
	return s.pitchImage(p, yaw) + uint32(roll) + uint32(cargoIndex)
}

// BrakingImageIndex is getBrakingImageIndex: the rotation's last frame.
func (s *VehicleBodySprite) BrakingImageIndex(p Pitch, yaw uint8) uint32 {
	return s.pitchImage(p, yaw) + uint32(s.NumFramesPerRotation) - 1
}
