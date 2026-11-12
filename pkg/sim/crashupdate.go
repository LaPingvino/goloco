package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A crashed train's parts: VehicleHead::landCrashedUpdate (VehicleHead.cpp
// 0x004AA625) with VehicleBogie::updateSegmentCrashed (VehicleBogie.cpp
// 0x004AA68E) and VehicleBody::updateSegmentCrashed (VehicleBody.cpp
// 0x004AA904). An exploding bogie slides on along its heading, slowing by
// 1/64 a tick, falls (tileBaseZ, reused as its fall counter, grows), and
// bounces off the ground, tile elements and other vehicles (half its speed,
// turned a little); one that isn't exploding rolls on along the track until
// it stops or runs out of track. Each body follows its bogies and explodes
// when one of them has.
//
// PARTIAL: explodeComponent's sound and effect entities, the splash, and the
// body's sub_4AAB0B (animation; not ported for land at all).

// bogieRoll is the bogie's roll state, as its normal update reads it.
func (m *Map) bogieRoll(i int, om *objects.ObjectManager) BogieRoll {
	e := &m.Entities[i]
	roll := BogieRoll{Var44: e.Var44, WheelSlipping: e.WheelSlipping, Reversed: e.Flags38&2 != 0, AnimationIndex: e.Animation}
	if veh := vehicleObject(om, e.ObjectID); veh != nil && e.ObjectSpriteType != 0xFF {
		roll.NumAnimationFrames = veh.BogieSprites[e.ObjectSpriteType&1].NumAnimationFrames
	}
	return roll
}

// landCrashedUpdate moves a crashed train's parts one tick.
func (m *Map) landCrashedUpdate(tr Train, om *objects.ObjectManager) {
	for _, car := range tr.Cars {
		for _, cc := range car {
			// (carState.hasBogieMoved only feeds the body's animations, PARTIAL)
			m.bogieCrashed(cc.Front, om)
			m.bogieCrashed(cc.Back, om)
			m.bodyCrashed(cc, om)
		}
	}
}

// bogieCrashed is VehicleBogie::updateSegmentCrashed.
func (m *Map) bogieCrashed(i int, om *objects.ObjectManager) {
	e := &m.Entities[i]
	speed := int32(e.Var5A & 0x7FFFFFFF)
	destroyed := e.Var5A&(1<<31) != 0
	speed -= speed / 64
	if speed <= 2*mph {
		speed = 0
	}
	e.Var5A = uint32(speed)
	if destroyed {
		e.Var5A |= 1 << 31
	}
	dist := speed / 128 // unkDistance1 = unkDistance2
	roll := m.bogieRoll(i, om)
	roll.UpdateRoll(dist)
	if e.Animation != roll.AnimationIndex {
		m.invalidateSprite(e)
	}
	e.Var44, e.Animation = roll.Var44, roll.AnimationIndex

	if !destroyed {
		if e.Mode != 1 { // not road: on along the track
			c := m.component(i)
			r := c.UpdateMotion(dist, func(c *MotionComponent, f *uint32) bool { return m.TrackNewPiece(c, f, 0, false, 0) })
			m.store(i, &c)
			if r.Flags&(MotionUnkM00|MotionNoRouteFound) != 0 {
				e.Var5A |= 1 << 31
			}
		}
		return
	}
	if speed < 1*mph {
		return
	}
	xv, yv := olmath.ComputeXYVector(speed/16, e.SpriteYaw)
	// The whole part of the step, as upstream's uint32 → int16 cast rounds it;
	// the fraction is kept in tileX / tileY, which hold no tile any more here.
	dx, dy := int32(int16(uint32(xv)>>16)), int32(int16(uint32(yv)>>16))
	tx := int32(uint16(e.TileX)) + int32(uint16(xv))
	if tx >= 0x10000 {
		dx++
	}
	e.TileX = coords.WorldUnits(int16(uint16(tx)))
	ty := int32(uint16(e.TileY)) + int32(uint16(yv))
	if ty >= 0x10000 {
		dy++
	}
	e.TileY = coords.WorldUnits(int16(uint16(ty)))
	e.TileBaseZ++ // the fall: one more each tick
	px, py, pz := int32(e.PosX), int32(e.PosY), int32(e.PosZ)
	nx, ny, nz := px+dx, py+dy, pz-int32(e.TileBaseZ/32)
	bounce := func() { // off something: turned, exploded if it hadn't, half the speed
		if speed >= 5*mph {
			m.rotateAndExplode(e, true)
		}
		e.Var5A = uint32(speed/2) | 1<<31
	}
	if !m.destroyedBogieHits(i, px, py, pz) {
		if m.destroyedBogieHits(i, nx, ny, pz) {
			nx, ny = px, py
			bounce()
		}
		if m.destroyedBogieHits(i, nx, ny, nz) {
			nz = pz
			if e.TileBaseZ >= 0x0A {
				m.rotateAndExplode(e, false)
			}
			e.TileBaseZ = 0
		}
	}
	land, water := m.GetHeight(nx, ny)
	if int32(land) >= pz || int32(land) >= nz {
		if int32(land) < pz && int32(land) >= nz {
			nz = pz
			e.TileBaseZ = 0
		}
		nx, ny = px, py
		bounce()
	}
	if water != 0 && int32(water) < pz && int32(water) >= nz {
		e.SpriteYaw = (e.SpriteYaw + 4) & 0x3F
		e.VehicleFlags |= vehicleFlagUnk5 // PARTIAL: the splash and its sound
	}
	m.moveTo(e, nx, ny, nz)
}

// rotateAndExplode is rotateAndExplodeIfNotAlreadyExploded: a quarter turn of
// the yaw circle's 64 steps by 4, and the part explodes once (unk_5).
func (m *Map) rotateAndExplode(e *scenario.VehicleEntity, clockwise bool) {
	if clockwise {
		e.SpriteYaw = (e.SpriteYaw + 4) & 0x3F
	} else {
		e.SpriteYaw = (e.SpriteYaw - 4) & 0x3F
	}
	e.VehicleFlags |= vehicleFlagUnk5 // explodeComponent: PARTIAL (no effect entities)
}

// destroyedBogieHits is destroyedBogieCheckForCollision: a tile element in
// the way (checkForTileCollision, 0x00462893) or another vehicle.
func (m *Map) destroyedBogieHits(bogie int, x, y, z int32) bool {
	return m.tileElementAt(x, y, z) || m.CheckForCollisions(bogie, x, y, z) != 0xFFFF
}

// tileElementAt is checkForTileCollision: an element of the tile whose height
// span holds z and which occupies the point's quarter.
func (m *Map) tileElementAt(x, y, z int32) bool {
	xn, yn := x&0x1F, y&0x1F
	var q uint8
	if xn < 16 {
		q = 1 << 2
		if yn >= 16 {
			q = 1 << 3
		}
	} else {
		q = 1 << 0
		if yn < 16 {
			q = 1 << 1
		}
	}
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil {
		return false
	}
	for _, r := range t.Order {
		if z < int32(t.baseZOf(r))*smallZStep || z >= int32(t.clearZOf(r))*smallZStep {
			continue
		}
		if r.Flags&q != 0 {
			return true
		}
	}
	return false
}

// bodyCrashed is VehicleBody::updateSegmentCrashed: the body follows its
// bogies (sub_4AC255) and explodes once one of them has. (animationUpdate
// does nothing for a crashed train; sub_4AAB0B is PARTIAL.)
func (m *Map) bodyCrashed(cc CarComponent, om *objects.ObjectManager) {
	f, b, body := &m.Entities[cc.Front], &m.Entities[cc.Back], &m.Entities[cc.Body]
	m.invalidateSprite(body)
	p := PlaceBody(int32(f.PosX), int32(f.PosY), int32(f.PosZ), int32(b.PosX), int32(b.PosY), int32(b.PosZ), body.ObjectSpriteType, vehicleObject(om, body.ObjectID))
	m.moveTo(body, p.X, p.Y, p.Z)
	if p.Turned {
		body.SpritePitch, body.SpriteYaw = p.Pitch, p.Yaw
	}
	m.invalidateSprite(body)
	if body.VehicleFlags&vehicleFlagUnk5 == 0 && (f.VehicleFlags&vehicleFlagUnk5 != 0 || b.VehicleFlags&vehicleFlagUnk5 != 0) {
		body.VehicleFlags |= vehicleFlagUnk5 // explodeComponent: PARTIAL
	}
}
