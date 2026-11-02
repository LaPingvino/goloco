package sim

// Vehicle motion along track and road — a port of Vehicle.cpp's
// updateTrackMotion (0x004B15FF, rail) and updateRoadMotion (0x0047C7FA):
// a component advances sub-position by sub-position along its piece's
// TrackData::MoveInfo path while it has distance to spend.
//
// The step onto the next piece is a hook: Map.TrackNewPiece is
// updateTrackMotionNewTrackPiece. PARTIAL: updateRoadMotionNewRoadPiece and
// the bogie collision checks (checkForCollisions) are not ported yet.

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
	RoutingHandle     RoutingHandle
	Owner             uint8 // company
	TrackType         uint8 // track (or road) object id
}

// UpdateVar1136114Flags (Vehicles/Vehicle.h).
const (
	MotionUnkM00                   uint32 = 1 << 0 // stopped: no next piece
	MotionNoRouteFound             uint32 = 1 << 1
	MotionCrashed                  uint32 = 1 << 2
	MotionUnkM03                   uint32 = 1 << 3
	MotionApproachingGradeCrossing uint32 = 1 << 4
)

// MotionResult is UpdateMotionResult: Flags are UpdateVar1136114Flags
// (unk_m00: stopped with no next piece; RemainingDistance is what was left
// over then).
type MotionResult struct {
	RemainingDistance int32
	Flags             uint32
}

// Blocked is Flags & unk_m00.
func (r MotionResult) Blocked() bool { return r.Flags&MotionUnkM00 != 0 }

func (c *MotionComponent) path() []MoveInfo {
	if c.Road {
		return RoadSubPosition(c.TrackAndDirection, c.RightHand)
	}
	return TrackSubPosition(c.TrackAndDirection)
}

// UpdateMotion is updateTrackMotion / updateRoadMotion. nextPiece is the
// new-piece step: it moves c onto the following piece (TrackAndDirection,
// TileX/Y, TileBaseZ), may add flags, and reports whether there is one; nil
// means none.
func (c *MotionComponent) UpdateMotion(distance int32, nextPiece func(c *MotionComponent, flags *uint32) bool) MotionResult {
	var res MotionResult
	c.RemainingDistance += distance
	moved := false
	ix, iy, iz := c.X, c.Y, c.Z
	for c.RemainingDistance >= 0x368A {
		moved = true
		sub := c.SubPosition + 1
		if int(sub) >= len(c.path()) {
			if nextPiece == nil || !nextPiece(c, &res.Flags) {
				res.RemainingDistance = c.RemainingDistance - 0x3689
				c.RemainingDistance = 0x3689
				res.Flags |= MotionUnkM00
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

// TrackNewPiece is updateTrackMotionNewTrackPiece: c moves onto the piece its
// next routing names, if that routing is set and the track there really
// connects (getTrackConnections). requiredMods is the head's var_53;
// isVeh2UnkM15 with veh1Handle stops vehicle 2 where vehicle 1's routing is.
func (m *Map) TrackNewPiece(c *MotionComponent, flags *uint32, requiredMods uint8, isVeh2UnkM15 bool, veh1Handle RoutingHandle) bool {
	next := c.RoutingHandle.WithIndex(c.RoutingHandle.Index() + 1)
	routing := m.Routing(next)
	if routing == RoutingAllocatedButFree {
		return false
	}
	if isVeh2UnkM15 && veh1Handle == c.RoutingHandle {
		*flags |= MotionUnkM03
		return false
	}
	x, y, z := c.TileX, c.TileY, int32(c.TileBaseZ)*4
	e, ok := TrackConnectionEnd(x, y, z, c.TrackAndDirection)
	if !ok {
		return false
	}
	tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, c.Owner, c.TrackType, requiredMods, 0)
	if tc.HasLevelCrossing {
		*flags |= MotionApproachingGradeCrossing
	}
	found := false
	for _, conn := range tc.Connections {
		if conn&TaDBasicMask == routing&TaDBasicMask {
			found = true
			break
		}
	}
	if !found {
		*flags |= MotionNoRouteFound
		return false
	}
	c.RoutingHandle = next
	old := TrackCoordinates[c.TrackAndDirection]
	c.TrackAndDirection = routing & TaDBasicMask
	c.TileX, c.TileY = x+int32(old.DX), y+int32(old.DY)
	c.TileBaseZ = uint8((z + int32(old.DZ)) / 4)
	return true
}
