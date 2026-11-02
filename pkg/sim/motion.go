package sim

// Vehicle motion along track and road — a port of Vehicle.cpp's
// updateTrackMotion (0x004B15FF, rail) and updateRoadMotion (0x0047C7FA):
// a component advances sub-position by sub-position along its piece's
// TrackData::MoveInfo path while it has distance to spend.
//
// PARTIAL: the step onto the next piece (updateTrackMotionNewTrackPiece /
// updateRoadMotionNewRoadPiece) is a caller-supplied hook, not ported yet;
// bogie collision checks (checkForCollisions) are not ported yet.

// MovementNibble is getMovementNibble: bit 0 x changed, bit 1 y, bit 2 z.
func MovementNibble(ax, ay, az, bx, by, bz int32) uint8 {
	var n uint8
	if ax != bx {
		n |= 1
	}
	if ay != by {
		n |= 2
	}
	if az != bz {
		n |= 4
	}
	return n
}

// MovementNibbleToDistance is kMovementNibbleToDistance (0x00500120).
var MovementNibbleToDistance = [8]int32{0, 0x220C, 0x220C, 0x3027, 0x199A, 0x2A99, 0x2A99, 0x3689}

// MotionComponent is the part of a VehicleBase the motion step reads and writes.
type MotionComponent struct {
	Road              bool // TransportMode::road (else rail)
	RightHand         bool // getGameState().trafficHandedness, for road
	TrackAndDirection uint16
	SubPosition       uint16
	TileX, TileY      int32 // world coordinates of the piece's tile
	TileBaseZ         uint8 // SmallZ
	RemainingDistance int32
	X, Y, Z           int32 // position (moveTo)
	SpriteYaw         uint8
	SpritePitch       uint8
}

// MotionResult is UpdateMotionResult: Blocked is UpdateVar1136114Flags::unk_m00
// (no next piece), RemainingDistance what was left over then.
type MotionResult struct {
	RemainingDistance int32
	Blocked           bool
}

func (c *MotionComponent) path() []MoveInfo {
	if c.Road {
		return RoadSubPosition(c.TrackAndDirection, c.RightHand)
	}
	return TrackSubPosition(c.TrackAndDirection)
}

// UpdateMotion is updateTrackMotion / updateRoadMotion. nextPiece is the
// new-piece step: it moves c onto the following piece (TrackAndDirection,
// TileX/Y, TileBaseZ) and reports whether there is one; nil means none.
func (c *MotionComponent) UpdateMotion(distance int32, nextPiece func(*MotionComponent) bool) MotionResult {
	var res MotionResult
	c.RemainingDistance += distance
	moved := false
	ix, iy, iz := c.X, c.Y, c.Z
	for c.RemainingDistance >= 0x368A {
		moved = true
		sub := c.SubPosition + 1
		if int(sub) >= len(c.path()) {
			if nextPiece == nil || !nextPiece(c) {
				res.RemainingDistance = c.RemainingDistance - 0x3689
				c.RemainingDistance = 0x3689
				res.Blocked = true
				break
			}
			sub = 0
		}
		c.SubPosition = sub
		m := c.path()[sub]
		nx, ny, nz := int32(m.X)+c.TileX, int32(m.Y)+c.TileY, int32(m.Z)+int32(c.TileBaseZ)*4
		c.RemainingDistance -= MovementNibbleToDistance[MovementNibble(ix, iy, iz, nx, ny, nz)]
		ix, iy, iz = nx, ny, nz
		c.SpriteYaw, c.SpritePitch = m.Yaw, m.Pitch
	}
	if moved {
		c.X, c.Y, c.Z = ix, iy, iz
	}
	return res
}
