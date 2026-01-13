package effects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Entities/Entity.h"
// #include <cstdint>
// namespace OpenLoco
// forward: struct Exhaust;
// forward: struct MoneyEffect;
// forward: struct VehicleCrashParticle;
// forward: struct ExplosionCloud;
// forward: struct Splash;
// forward: struct Fireball;
// forward: struct ExplosionSmoke;
// forward: struct Smoke;
type EffectType int

const (
	Exhaust EffectType = 0
	RedGreenCurrency EffectType = 1
	WindowCurrency EffectType = 2
	VehicleCrashParticle EffectType = 3
	ExplosionCloud EffectType = 4
	Splash EffectType = 5
	Fireball EffectType = 6
	ExplosionSmoke EffectType = 7
	Smoke EffectType = 8 // Smoke from broken down train
)
type EffectEntity struct {
	EntityBase // embedded
// template<typename TType, EffectType TClass>
// TType* as() const
// return subType == TClass ? (TType*)this : nullptr;
}
const EffectEntityBaseType any = EntityBaseType.effect
// public:
// orphan member: EffectType subType;
func GetSubType() EffectType {
	// void setSubType(const EffectType newType) { subType = newType; }
	// Exhaust* asExhaust() const { return as<Exhaust, EffectType::exhaust>(); }
	// MoneyEffect* asRedGreenCurrency() const { return as<MoneyEffect, EffectType::redGreenCurrency>(); }
	// MoneyEffect* asWindowCurrency() const { return as<MoneyEffect, EffectType::windowCurrency>(); }
	// VehicleCrashParticle* asVehicleCrashParticle() const { return as<VehicleCrashParticle, EffectType::vehicleCrashParticle>(); }
	// ExplosionCloud* asExplosionCloud() const { return as<ExplosionCloud, EffectType::explosionCloud>(); }
	// Splash* asSplash() const { return as<Splash, EffectType::splash>(); }
	// Fireball* asFireball() const { return as<Fireball, EffectType::fireball>(); }
	// ExplosionSmoke* asExplosionSmoke() const { return as<ExplosionSmoke, EffectType::explosionSmoke>(); }
	// Smoke* asSmoke() const { return as<Smoke, EffectType::smoke>(); }
	// void update();
}
// static_assert(sizeof(EffectEntity) <= sizeof(Entity));
