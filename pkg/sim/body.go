package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// VehicleBody placement — a port of VehicleBody.cpp sub_4AC255 (the body at
// its bogies' midpoint, pitched and turned by the vector between them),
// updateSpritePitch / updateSpritePitchSteepSlopes and the
// calculateYaw*FromVector family (yaw_gen.go).
//
// PARTIAL: VehicleBody::update's animationUpdate (smoke, steam, sparks,
// wakes) and sub_4AAB0B (body animation / tilt frame) are not ported.

// kVehicleBodyIndexToPitch.
var bodyIndexToPitch = [10]uint8{0, 1, 2, 3, 4, 0, 5, 6, 7, 8}

func (f *yawFn) calc(x, y int16) uint8 {
	i := 0
	if x < 0 {
		i += f.XNeg
		x = -x
	}
	if y < 0 {
		i += f.YNeg
		y = -y
	}
	xy := uint32(0xFFFFFFFF)
	if y != 0 {
		xy = uint32((uint64(uint16(x)) << 16) / uint64(uint16(y)))
	}
	for _, t := range f.Thresholds {
		if xy >= t {
			i++
		}
	}
	return f.Table[i]
}

// Yaw0..Yaw4 and Yaw1Plane are calculateYaw*FromVector.
func Yaw0(x, y int16) uint8      { return yaw0.calc(x, y) }
func Yaw1(x, y int16) uint8      { return yaw1.calc(x, y) }
func Yaw1Plane(x, y int16) uint8 { return yaw1Plane.calc(x, y) }
func Yaw2(x, y int16) uint8      { return yaw2.calc(x, y) }
func Yaw3(x, y int16) uint8      { return yaw3.calc(x, y) }
func Yaw4(x, y int16) uint8      { return yaw4.calc(x, y) }

func pitchRatio(xyOffset uint16, zOffset int16) (int, uint32) {
	i := 0
	if zOffset < 0 {
		i = 5
		zOffset = -zOffset
	}
	xyz := uint32(0xFFFFFFFF)
	if xyOffset != 0 {
		xyz = uint32((uint64(uint16(zOffset)) << 16) / uint64(xyOffset))
	}
	return i, xyz
}

// UpdateSpritePitch is updateSpritePitch (0x004BF49D).
func UpdateSpritePitch(xyOffset uint16, zOffset int16) uint8 {
	i, xyz := pitchRatio(xyOffset, zOffset)
	if xyz >= 3331 {
		i++
		if xyz >= 9000 {
			i++
		}
	}
	return bodyIndexToPitch[i]
}

// UpdateSpritePitchSteepSlopes is updateSpritePitchSteepSlopes (0x004BF4DA).
func UpdateSpritePitchSteepSlopes(xyOffset uint16, zOffset int16) uint8 {
	i, xyz := pitchRatio(xyOffset, zOffset)
	if xyz > 10064 {
		i += 2
		if xyz >= 20500 {
			i++
			if xyz >= 22000 {
				i++
			}
		}
	} else if xyz >= 3331 {
		i++
	}
	return bodyIndexToPitch[i]
}

// BodyPlacement is what sub_4AC255 sets on a body.
type BodyPlacement struct {
	X, Y, Z    int32
	Pitch, Yaw uint8
	Turned     bool // false when objectSpriteType is 0xFF (only moved)
}

// PlaceBody is sub_4AC255: the body at the midpoint of its back and front
// bogies; pitch from their height difference over distance2D; yaw at the
// accuracy its body sprite has (4 bits on a transition pitch).
func PlaceBody(fx, fy, fz, bx, by, bz int32, objectSpriteType uint8, veh *objects.VehicleObject) BodyPlacement {
	p := BodyPlacement{X: (fx + bx) / 2, Y: (fy + by) / 2, Z: (fz + bz) / 2}
	if objectSpriteType == 0xFF || veh == nil || int(objectSpriteType) >= len(veh.BodySprites) {
		return p
	}
	p.Turned = true
	dx, dy, dz := int16(fx-bx), int16(fy-by), int16(fz-bz)
	dist := olmath.Distance2D(olmath.Vector2i{X: fx, Y: fy}, olmath.Vector2i{X: bx, Y: by})
	sp := veh.BodySprites[objectSpriteType]
	if sp.Flags&objects.BodyFlagHasSteepSprites != 0 {
		p.Pitch = UpdateSpritePitchSteepSlopes(dist, dz)
	} else {
		p.Pitch = UpdateSpritePitch(dist, dz)
	}
	if p.Pitch&1 != 0 {
		p.Yaw = Yaw1(dx, dy)
		return p
	}
	acc := sp.SlopedYawAccuracy
	if p.Pitch == 0 {
		acc = sp.FlatYawAccuracy
	}
	switch acc {
	case 0:
		p.Yaw = Yaw0(dx, dy)
	case 1:
		p.Yaw = Yaw1(dx, dy)
	case 2:
		p.Yaw = Yaw2(dx, dy)
	case 3:
		p.Yaw = Yaw3(dx, dy)
	case 4:
		p.Yaw = Yaw4(dx, dy)
	}
	return p
}
