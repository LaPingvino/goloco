package sim

// VehicleBogie::update (0x004AA008) and updateRoll (0x004AAC02). Its
// collisions and the train's destruction are in crash.go.

// WheelSlippingDuration is kWheelSlippingDuration (ticks).
const WheelSlippingDuration = 64

// BogieRoll is the bogie state updateRoll reads and writes.
type BogieRoll struct {
	Var44          uint16
	WheelSlipping  uint8
	Reversed       bool  // Flags38::isReversed
	AnimationIndex uint8 // shown wheel frame
	// NumAnimationFrames is bogieSprites[objectSpriteType].numAnimationFrames;
	// 0 when objectSpriteType is 0xFF (no bogie sprite).
	NumAnimationFrames uint8
}

// UpdateRoll is updateRoll: the wheels turn by distance/8 (backwards when
// reversed); the frame is (numFrames-1) & (var_44 / 4096).
func (b *BogieRoll) UpdateRoll(distance int32) {
	unk := distance / 8
	if b.Reversed {
		unk = -unk
	}
	b.Var44 = uint16(int32(b.Var44) + unk)
	if b.NumAnimationFrames != 0 {
		b.AnimationIndex = (b.NumAnimationFrames - 1) & uint8(b.Var44/4096)
	}
}

// BogieUpdate is VehicleBogie::update for land vehicles: move by
// unkDistance1, then roll the wheels by unkDistance2 (or, while slipping,
// 500 + 320·t with t mirrored after half the slip). The caller acts on the
// result: noRouteFound destroys the train, crashed is a collision (c.Collide
// set: CheckForCollisions at each step).
func (m *Map) BogieUpdate(c *MotionComponent, roll *BogieRoll, unkDistance1, unkDistance2 int32, requiredMods uint8) MotionResult {
	next := m.TrackNewPiece
	if c.Road {
		next = m.RoadNewPiece
	}
	r := c.UpdateMotion(unkDistance1, func(c *MotionComponent, f *uint32) bool {
		return next(c, f, requiredMods, false, 0)
	})
	d := unkDistance2
	if roll.WheelSlipping != 0 {
		u := int32(roll.WheelSlipping)
		if u > WheelSlippingDuration/2 {
			u = WheelSlippingDuration - u
		}
		d = 500 + u*320
	}
	roll.UpdateRoll(d)
	return r
}
