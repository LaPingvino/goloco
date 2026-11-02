package objects

import "github.com/LaPingvino/goloco/pkg/olmath"

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
	// Colour is the car part's colour scheme (primary, secondary) in the car
	// variant; zero in the object variant, whose caller passes one.
	Colour [2]uint8
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
			items = append(items, DrawItem{v.bogieImage(c.FrontBogieSpriteIdx, yaw, roll, reversed), total + frontDist, false, [2]uint8{}})
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
			items = append(items, DrawItem{v.bogieImage(c.BackBogieSpriteIdx, yaw, roll, !reversed), backDist, false, [2]uint8{}})
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
			items = append(items, DrawItem{b.BodyImageIndex(PitchFlat, by, rollIndex, 0), bodyDist, true, [2]uint8{}})
		}
		total += length
	}
	return items, total
}

// screenDistance is toScreenDistDrawItems' getScreenDistance: a distance
// along the vehicle at yaw, as a screen x offset (gameToScreen at rotation 0
// of the int16 Pos3, then −x / 4).
func screenDistance(gameDist int32, yaw uint8) int32 {
	x, y := olmath.ComputeXYVector(gameDist, yaw)
	px := int16(int16(y) - int16(x))
	return -int32(px) / 4
}

// InlineDrawItems is getDrawItemsForVehicle then toScreenDistDrawItems: the
// object's items with screen distances, and its screen length.
func (v *VehicleObject) InlineDrawItems(yaw, roll uint8) ([]DrawItem, int32) {
	items, total := v.DrawItems(yaw, roll)
	for i := range items {
		items[i].Dist = screenDistance(items[i].Dist, yaw)
	}
	return items, screenDistance(total, yaw)
}

// VehicleYDisplayOffset is getVehicleObjectYDisplayOffset: a road vehicle
// sits at its road's displayOffset (0 with no road type), anything else at
// its track's (3 with no track type).
func (m *ObjectManager) VehicleYDisplayOffset(v *VehicleObject) int {
	if v.Mode == TransportModeRoad {
		if v.TrackType != 0xFF && int(v.TrackType) < len(m.RoadObjects) && m.RoadObjects[v.TrackType] != nil {
			return int(m.RoadObjects[v.TrackType].DisplayOffset)
		}
		return 0
	}
	if v.TrackType == 0xFF {
		return 3
	}
	if int(v.TrackType) < len(m.TrackObjects) && m.TrackObjects[v.TrackType] != nil {
		return int(m.TrackObjects[v.TrackType].DisplayOffset)
	}
	return 0
}

// CarPart is one bogie or body entity as getDrawItemsForVehicle reads it:
// its objectSpriteType, Flags38 isReversed, animation (bogie animationIndex /
// body animationFrame), cargo frame (body) and colour scheme.
type CarPart struct {
	SpriteType uint8
	Reversed   bool
	Animation  uint8
	CargoFrame uint8
	Colour     [2]uint8
}

// CarComponent is a car component's front bogie, back bogie and body.
type CarComponent struct{ Front, Back, Body CarPart }

// CarDrawItems is getDrawItemsForVehicle for a Car: its components walked in
// order (the object's components from the last one when the car's body is
// reversed), bogies and bodies from the entities' own sprite types, reversal,
// animation frames (only when animated) and colours; with brakeLights (the
// train's veh2 brakeLightTimeout, animated, and a body with braking lights)
// the braking frame is added over the body.
func (v *VehicleObject) CarDrawItems(yaw uint8, comps []CarComponent, carReversed, animated, brakeLights bool) ([]DrawItem, int32) {
	var items []DrawItem
	total := int32(0)
	bogies := v.Mode == TransportModeRail || v.Mode == TransportModeRoad
	ci := 0
	if carReversed {
		ci = int(v.NumCarComponents) - 1
	}
	anim := func(a uint8) uint8 {
		if animated {
			return a
		}
		return 0
	}
	bogie := func(p CarPart) uint32 {
		y := yaw
		if p.Reversed {
			y ^= 1 << 5
		}
		y /= 2
		b := &v.BogieSprites[p.SpriteType] // callers keep SpriteType < 2 (guarded below)
		if b.Flags&BogieFlagRotationalSymmetry != 0 {
			y &= 0xF
		}
		return uint32(b.NumFramesPerRotation)*uint32(y) + uint32(anim(p.Animation)) + b.FlatImageIDs
	}
	for _, cc := range comps {
		if ci < 0 || ci >= len(v.CarComponents) {
			break
		}
		c := &v.CarComponents[ci]
		front, back := int32(c.FrontBogiePosition), int32(c.BackBogiePosition)
		if carReversed {
			front, back = back, front
		}
		if cc.Front.SpriteType != SpriteIndexNull && bogies && int(cc.Front.SpriteType) < len(v.BogieSprites) {
			items = append(items, DrawItem{bogie(cc.Front), total + front, false, cc.Front.Colour})
		}
		length := int32(0)
		backDist := total
		if c.BodySpriteIdx != SpriteIndexNull {
			length = int32(v.BodySprites[c.BodySpriteIdx&^SpriteIndexIsReversed].HalfLength) * 2
			backDist += length
		}
		backDist -= back
		if cc.Back.SpriteType != SpriteIndexNull && bogies && int(cc.Back.SpriteType) < len(v.BogieSprites) {
			items = append(items, DrawItem{bogie(cc.Back), backDist, false, cc.Back.Colour})
		}
		bodyDist := total + (front+length-back)/2
		if cc.Body.SpriteType != SpriteIndexNull && int(cc.Body.SpriteType) < len(v.BodySprites) {
			b := &v.BodySprites[cc.Body.SpriteType]
			y := yaw
			if cc.Body.Reversed {
				y ^= 1 << 5
			}
			items = append(items, DrawItem{b.BodyImageIndex(PitchFlat, y, anim(cc.Body.Animation), cc.Body.CargoFrame), bodyDist, true, cc.Body.Colour})
			if animated && b.Flags&BodyFlagHasBrakingLights != 0 && brakeLights {
				items = append(items, DrawItem{b.BrakingImageIndex(PitchFlat, yaw), bodyDist, true, cc.Body.Colour})
			}
		}
		total += length
		if carReversed {
			ci--
		} else {
			ci++
		}
	}
	return items, total
}
