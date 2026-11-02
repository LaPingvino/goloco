package paint

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// Vehicle painting, ported from Paint/PaintVehicle.cpp.

// reversePitch is kReversePitch (0x00500160): flat, the down pitches for the
// up ones and back, the curve pitches swapped.
var reversePitch = [13]objects.Pitch{0, 5, 6, 7, 8, 1, 2, 3, 4, 10, 9, 12, 11}

// boundingBoxFactorsByYaw is kBoundingBoxFactorsByYaw (0x005001B4): x/y
// offset and size factors (× bodyLength / 256) per yaw 0-31 (copied from
// upstream's source).
var boundingBoxFactorsByYaw = [32][4]int{
	{-30, -3, 60, 6},
	{-29, -4, 58, 8},
	{-28, -6, 56, 12},
	{-26, -6, 56, 12},
	{-24, -6, 56, 12},
	{-22, -11, 48, 24},
	{-20, -20, 40, 40},
	{-20, -20, 45, 45},
	{-16, -16, 46, 46},
	{-20, -20, 45, 45},
	{-20, -20, 40, 40},
	{-11, -22, 24, 48},
	{-6, -24, 12, 56},
	{-6, -26, 12, 56},
	{-6, -28, 12, 56},
	{-4, -29, 8, 58},
	{-3, -30, 6, 60},
	{-4, -29, 8, 58},
	{-5, -28, 10, 57},
	{-2, -28, 6, 54},
	{-2, -20, 8, 50},
	{-2, -20, 8, 45},
	{-4, -15, 12, 40},
	{-5, -12, 17, 32},
	{-5, -5, 25, 25},
	{-12, -5, 32, 17},
	{-15, -4, 40, 12},
	{-20, -2, 45, 8},
	{-20, -2, 50, 8},
	{-28, -2, 54, 6},
	{-28, -5, 57, 10},
	{-29, -4, 58, 8},
}

// BodyEl is what paintBody reads of a VehicleBody entity.
type BodyEl struct {
	ObjectSpriteType   uint8 // 0xFF: no body sprite
	SpriteYaw          uint8
	SpritePitch        uint8
	Reversed           bool // Flags38::isReversed
	AnimationFrame     uint8
	CargoFrame         uint8
	BodyIndex          uint8 // car component of the object
	Z                  int   // position.z
	AirOrWater         bool  // getTransportMode() air or water
	Colour             [2]uint8
	BrakeLights        bool // the train's veh2 brakeLightTimeout != 0
	Ghost, VehicleUnk5 bool
}

// bodyBounds is paintBody's bounding box: air and water vehicles a fixed
// 48×48×15 at (−8, −8, z + 11); others from the car component's overhang
// (bogie positions, flipped for a reversed body or component) and the body
// length (halfLength·2 − 4) scaled by the yaw's factors.
func bodyBounds(b BodyEl, v *objects.VehicleObject, sp *objects.VehicleBodySprite, originalYaw uint8) (bbOff, bbSize Pos3) {
	if b.AirOrWater {
		return Pos3{-8, -8, b.Z + 11}, Pos3{48, 48, 15}
	}
	c := &v.CarComponents[b.BodyIndex&3]
	overhang := int(c.FrontBogiePosition) - int(c.BackBogiePosition)
	if b.Reversed {
		overhang = -overhang
	}
	if c.BodySpriteIdx&objects.SpriteIndexIsReversed != 0 {
		overhang = -overhang
	}
	ox, oy := olmath.ComputeXYVector(int32(overhang), originalYaw)
	bbOff.X, bbOff.Y = int(ox/8), int(oy/8) // TVector / 8
	length := int(sp.HalfLength)*2 - 4
	f := boundingBoxFactorsByYaw[originalYaw&0x1F]
	bbOff.X += (f[0] * length) >> 8
	bbOff.Y += (f[1] * length) >> 8
	bbOff.Z = b.Z + 11
	return bbOff, Pos3{(f[2] * length) >> 8, (f[3] * length) >> 8, 15}
}

// PaintBody is paintBody (0x004B0CFC): the body image for its yaw (turned by
// the view rotation; reversed bodies flip yaw and pitch) and pitch with the
// animation and cargo frames, at (0, 0, z) through addToPlotList4FD200, in
// the body's colour scheme (ghost and VehicleFlags::unk_5 PARTIAL: drawn
// plainly); the braking frame attached while the brake lights are on.
// palMap returns the colour scheme's palette map.
func (s *Session) PaintBody(b BodyEl, v *objects.VehicleObject, sprite func(int) *gfx8.Sprite, palMap func(primary, secondary uint8) []uint8) {
	if v == nil || b.ObjectSpriteType == objects.SpriteIndexNull || int(b.ObjectSpriteType) >= len(v.BodySprites) {
		return
	}
	sp := &v.BodySprites[b.ObjectSpriteType]
	yaw := (b.SpriteYaw + s.Rotation<<4) & 0x3F
	originalYaw := yaw
	pitch := objects.Pitch(b.SpritePitch)
	if b.Reversed {
		yaw ^= 1 << 5
		if int(b.SpritePitch) < len(reversePitch) {
			pitch = reversePitch[b.SpritePitch]
		}
	}
	img := int(v.ImageOffset + sp.BodyImageIndex(pitch, yaw, b.AnimationFrame, b.CargoFrame))
	bbOff, bbSize := bodyBounds(b, v, sp, originalYaw)
	spr := sprite(img)
	if spr == nil {
		return
	}
	pm := palMap(b.Colour[0], b.Colour[1])
	ps := s.Add4FD200(Image{Sprite: spr, Op: gfx8.OpFor(spr, pm != nil, false), PalMap: pm}, Pos3{0, 0, b.Z}, bbOff, bbSize)
	if ps != nil && sp.Flags&objects.BodyFlagHasBrakingLights != 0 && b.BrakeLights && !b.Ghost && !b.VehicleUnk5 {
		if br := sprite(int(v.ImageOffset + sp.BrakingImageIndex(pitch, yaw))); br != nil {
			s.AttachToPrevious(Image{Sprite: br, Op: gfx8.OpFor(br, false, false)}, 0, 0)
		}
	}
}

// Add4FD200 is addToPlotList4FD200: addToPlotListAsParent, then the bounding
// box's minimum corner pulled inside the render target's horizontal range
// (screen x = y − x at rotation 0, and the other rotations' forms).
func (s *Session) Add4FD200(img Image, off, bbOff, bbSize Pos3) *Struct {
	s.last = nil
	ps := s.create(img, off, bbOff, bbSize)
	if ps == nil {
		return nil
	}
	left, right := s.rtX, s.rtX+s.rtW
	m := ps.Mins
	switch s.Rotation & 3 {
	case 0:
		if m.Y-m.X < left {
			m.Y = m.X + left
		}
		if m.Y-m.X > right {
			m.X = m.Y - right
		}
	case 1:
		if -m.Y-m.X < left {
			m.X = -m.Y - left
		}
		if -m.Y-m.X > right {
			m.Y = -m.X - right
		}
	case 2:
		if -m.Y+m.X < left {
			m.Y = m.X - left
		}
		if -m.Y+m.X > right {
			m.X = m.Y + right
		}
	case 3:
		if m.Y+m.X < left {
			m.X = -m.Y + left
		}
		if m.Y+m.X > right {
			m.Y = -m.X + right
		}
	}
	ps.Mins = m
	s.last = ps
	s.addToQuadrant(ps)
	return ps
}
