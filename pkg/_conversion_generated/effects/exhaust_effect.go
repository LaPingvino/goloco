package effects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Effect.h"
// namespace OpenLoco
// forward: struct SteamObject;
type Exhaust struct {
	EffectEntity // embedded
	FrameNum uint16
	StationaryProgress int16
	WindProgress uint16
	Var_34 int16
	Var_36 int16
	ObjectId uint8
// const SteamObject* getObject() const;
	// method: void update();
// static Exhaust* create(World::Pos3 loc, uint8_t type);
func (e *Exhaust) IsSubObjType1() bool {
}
// static_assert(sizeof(Exhaust) <= sizeof(Entity));
}
