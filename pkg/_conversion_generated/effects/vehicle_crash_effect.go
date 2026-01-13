package effects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Effect.h"
// namespace OpenLoco
type VehicleCrashParticle struct {
	EffectEntity // embedded
	TimeToLive uint16
	Frame uint16
	ColourScheme ColourScheme
	CrashedSpriteBase uint16
// World::Pos3 velocity;       // 0x32
// // TODO: Convert this to World::Pos3 once we can change the coord type to int32_t.
	AccelerationX int32
	AccelerationY int32
	AccelerationZ int32
	// method: void update();
// static VehicleCrashParticle* create(const World::Pos3& loc, const ColourScheme colourScheme);
}
// static_assert(sizeof(VehicleCrashParticle) <= sizeof(Entity));
