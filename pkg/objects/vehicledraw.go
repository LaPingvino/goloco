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

// SpriteIndex values of a car component (VehicleObject.h SpriteIndex).
const (
	SpriteIndexNull       = 0xFF
	SpriteIndexIsReversed = 1 << 7
)

// VehicleObjectFlags used for drawing (VehicleObject.h).
const (
	VehicleFlagJacobsBogieFront = 1 << 2
	VehicleFlagJacobsBogieRear  = 1 << 3
)

// kUnk500264: the roll frame offset per roll (& 7).
var rollFrameOffset = [8]uint8{0, 1, 1, 1, 0, 2, 2, 2}

// DrawItem is VehicleDraw.cpp's DrawItem: an image (relative to the
// object's ImageOffset; the caller adds the colour scheme), its distance
// along the vehicle, and whether it is a body (else a bogie).
type DrawItem struct {
	Image  uint32
	Dist   int32
	IsBody bool
}

// bogieImage is the bogie frame: yaw (turned round for the given end) / 2,
// masked to 16 for a symmetric bogie, plus the roll frame.
func (v *VehicleObject) bogieImage(idx uint8, yaw, roll uint8, flip bool) uint32 {
	if flip {
		yaw ^= 1 << 5
	}
	yaw /= 2
	b := &v.BogieSprites[idx]
	if b.Flags&BogieFlagRotationalSymmetry != 0 {
		yaw &= 0xF
	}
	rollIndex := (b.NumFramesPerRotation - 1) & roll
	return uint32(b.NumFramesPerRotation)*uint32(yaw) + uint32(rollIndex) + b.FlatImageIDs
}

// DrawItems is getDrawItemsForVehicle for an object (no car: the build
// window and drawVehicleOverview): for each car component, skipping the
// Jacobs-bogie ends, its front bogie, back bogie and body (rail and road
// draw bogies), with their distances; and the total length.
func (v *VehicleObject) DrawItems(yaw, roll uint8) ([]DrawItem, int32) {
	var items []DrawItem
	total := int32(0)
	rollOff := rollFrameOffset[roll&7]
	bogies := v.Mode == TransportModeRail || v.Mode == TransportModeRoad
	for i := 0; i < int(v.NumCarComponents) && i < len(v.CarComponents); i++ {
		if v.Flags&VehicleFlagJacobsBogieFront != 0 && i == 0 {
			continue
		}
		if v.Flags&VehicleFlagJacobsBogieRear != 0 && i+1 == int(v.NumCarComponents) {
			continue
		}
		c := &v.CarComponents[i]
		reversed := c.BodySpriteIdx&SpriteIndexIsReversed != 0
		frontDist := int32(c.FrontBogiePosition)
		if c.FrontBogieSpriteIdx != SpriteIndexNull && bogies {
			items = append(items, DrawItem{v.bogieImage(c.FrontBogieSpriteIdx, yaw, roll, reversed), total + frontDist, false})
		}
		length := int32(0)
		backDist := total
		if c.BodySpriteIdx != SpriteIndexNull {
			length = int32(v.BodySprites[c.BodySpriteIdx&^SpriteIndexIsReversed].HalfLength) * 2
			backDist += length
		}
		backPos := int32(c.BackBogiePosition)
		backDist -= backPos
		if c.BackBogieSpriteIdx != SpriteIndexNull && bogies {
			items = append(items, DrawItem{v.bogieImage(c.BackBogieSpriteIdx, yaw, roll, !reversed), backDist, false})
		}
		bodyDist := total + (frontDist+length-backPos)/2
		if c.BodySpriteIdx != SpriteIndexNull {
			b := &v.BodySprites[c.BodySpriteIdx&^SpriteIndexIsReversed]
			by := yaw
			if reversed {
				by ^= 1 << 5
			}
			rollIndex := uint8(0)
			if b.Flags&BodyFlagHasSpeedAnimation == 0 {
				rollIndex = (b.NumAnimationFrames - 1) & roll
			}
			if b.NumRollFrames != 1 {
				rollIndex += rollOff
			}
			items = append(items, DrawItem{b.BodyImageIndex(PitchFlat, by, rollIndex, 0), bodyDist, true})
		}
		total += length
	}
	return items, total
}
