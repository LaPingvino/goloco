package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Crash effects: Effects/ExplosionEffect.cpp (ExplosionCloud),
// VehicleCrashEffect.cpp (VehicleCrashParticle) and SplashEffect.cpp, made by
// VehicleBase::explodeComponent (Vehicle.cpp 0x004AA5D4) and a crashed bogie
// landing in water. Entities of the misc list, as upstream; goloco keeps their
// fields by id beside the money effects (moneyeffect.go), and the scene paints
// them (PaintEffectEntity.cpp). PARTIAL: the crash and splash2 sounds (the sim
// has no sound queue).

const (
	effectTypeVehicleCrashParticle = 3 // EffectType::vehicleCrashParticle
	effectTypeExplosionCloud       = 4 // EffectType::explosionCloud
	effectTypeSplash               = 5 // EffectType::splash

	maxMiscEntities = 4000 // Limits::kMaxMiscEntities

	crashParticleMaxFrames = 12 * 256 // kMaxSplashFrames: 12 sprites, frame / 256
	crashParticleSpeed     = 85       // kAnimationSpeed
	crashParticleGravity   = 5041     // kGravity
)

// EffectType values the scene paints.
const (
	EffectVehicleCrashParticle = effectTypeVehicleCrashParticle
	EffectExplosionCloud       = effectTypeExplosionCloud
	EffectSplash               = effectTypeSplash
)

// Effect is a crash effect entity's fields: position, sprite extents
// (spriteWidth / spriteHeightNegative / spriteHeightPositive) and, for a crash
// particle, its colour scheme, life, sprite set (crashedSpriteBase 0–4),
// velocity (the 16-bit fractions carried between ticks, coord_t as upstream)
// and acceleration.
type Effect struct {
	X, Y, Z                     int32
	SubType                     uint8
	Width, HeightNeg, HeightPos uint8
	Frame                       uint16
	Colour                      [2]uint8
	TimeToLive                  uint16
	SpriteBase                  uint8
	VelX, VelY, VelZ            int16
	AccX, AccY, AccZ            int32
	// An exhaust's (exhaust.go): objectId (steam slot, 0x80 the second frame
	// run), stationaryProgress and windProgress; its frameNum is Frame.
	ObjectID   uint8
	Stationary uint8
	Wind       uint16
}

// createEntityMisc is EntityManager::createEntityMisc: none past
// kMaxMiscEntities; the free list's head, moved to misc.
func (l *EntityLists) createEntityMisc() (uint16, bool) {
	if l.Counts[EntityListMisc] >= maxMiscEntities || l.Counts[EntityListNull] == 0 {
		return 0, false
	}
	id := l.Heads[EntityListNull]
	l.moveEntityToList(id, EntityListMisc)
	l.Links[id].BaseType = scenario.EntityBaseTypeEffect
	return id, true
}

func (m *Map) newEffect(e Effect) (*Effect, bool) {
	if m.EntityLists == nil {
		return nil, false
	}
	id, ok := m.EntityLists.createEntityMisc()
	if !ok {
		return nil, false
	}
	if m.Effects == nil {
		m.Effects = map[uint16]*Effect{}
	}
	p := &e
	m.Effects[id] = p
	return p, true
}

func (m *Map) freeEffect(id uint16) {
	delete(m.Effects, id)
	m.EntityLists.freeEntity(id)
}

// explodeComponent is VehicleBase::explodeComponent: 22 above the part, an
// explosion cloud and up to seven crash particles (a quarter of its sprite
// width) in its colours.
func (m *Map) explodeComponent(i int) {
	e := &m.Entities[i]
	x, y, z := int32(e.PosX), int32(e.PosY), int32(e.PosZ)+22
	m.newEffect(Effect{X: x, Y: y, Z: z, SubType: effectTypeExplosionCloud, Width: 44, HeightNeg: 32, HeightPos: 34}) // ExplosionCloud::create
	for range min(int(e.SpriteWidth)/4, 7) {
		m.createCrashParticle(x, y, z, [2]uint8{e.ColourPrimary, e.ColourSecondary})
	}
}

// createCrashParticle is VehicleCrashParticle::create (0x00440B0A): one gPrng1
// draw sets its frame, life, sprite set and push.
func (m *Map) createCrashParticle(x, y, z int32, colour [2]uint8) {
	p, ok := m.newEffect(Effect{X: x, Y: y, Z: z, SubType: effectTypeVehicleCrashParticle, Width: 8, HeightNeg: 8, HeightPos: 8, Colour: colour})
	if !ok {
		return
	}
	r := m.Rng1.RandNext()
	p.Frame = uint16((r & 0xFF) * 12)
	p.TimeToLive = uint16(140 + (r & 0x7F))
	p.SpriteBase = uint8((((r >> 23) & 0xFF) * 5) >> 8)
	p.AccX = int32(int16(r&0xFFFF)) * 4
	p.AccY = int32(int16((r>>16)&0xFFFF)) * 4
	p.AccZ = int32(((r>>8)&0xFFFF)*4 + 0x10000)
}

// createSplash is Splash::create (0x00440C6B): off the map none; 3 further
// along y.
func (m *Map) createSplash(x, y, z int32) {
	if x < 0 || y < 0 || x >= int32(m.Width)*32 || y >= int32(m.Height)*32 { // validCoords
		return
	}
	m.newEffect(Effect{X: x, Y: y + 3, Z: z, SubType: effectTypeSplash, Width: 33, HeightNeg: 51, HeightPos: 16})
}

// updateEffect is the crash effects' update: ExplosionCloud (0x004407CC),
// Splash (0x004407E0) and VehicleCrashParticle (0x004406A0).
func (m *Map) updateEffect(om *objects.ObjectManager, id uint16, e *Effect) {
	switch e.SubType {
	case effectTypeExhaust:
		m.updateExhaust(om, id, e)
	case effectTypeExplosionCloud:
		e.Frame += 0x80
		if e.Frame >= 0x1200 {
			m.freeEffect(id)
		}
	case effectTypeSplash:
		e.Frame += 0x55
		if e.Frame >= 0x1C00 {
			m.freeEffect(id)
		}
	case effectTypeVehicleCrashParticle:
		m.updateCrashParticle(id, e)
	}
}

// updateCrashParticle is VehicleCrashParticle::update: gravity and air
// resistance on the push, the 16.16 step, a splash (and the end) where it
// falls into water, a bounce off the land.
func (m *Map) updateCrashParticle(id uint16, e *Effect) {
	e.TimeToLive--
	if e.TimeToLive == 0 {
		m.freeEffect(id)
		return
	}
	e.AccZ -= crashParticleGravity
	e.AccX -= e.AccX / 256
	e.AccY -= e.AccY / 256
	e.AccZ -= e.AccZ / 256
	vx, vy, vz := int32(e.VelX)+e.AccX, int32(e.VelY)+e.AccY, int32(e.VelZ)+e.AccZ
	nx, ny, nz := int32(int16(e.X+(vx>>16))), int32(int16(e.Y+(vy>>16))), int32(int16(e.Z+(vz>>16)))
	e.VelX, e.VelY, e.VelZ = int16(vx), int16(vy), int16(vz)
	land, water := m.GetHeight(nx, ny)
	if water != 0 && e.Z >= int32(water) && nz <= int32(water) {
		m.createSplash(e.X, e.Y, int32(water))
		m.freeEffect(id)
		return
	}
	if e.Z >= int32(land) && nz <= int32(land) {
		e.AccZ = -e.AccZ
		nz = int32(land)
	}
	e.X, e.Y, e.Z = nx, ny, nz
	e.Frame += crashParticleSpeed
	if e.Frame >= crashParticleMaxFrames {
		e.Frame = 0
	}
}
