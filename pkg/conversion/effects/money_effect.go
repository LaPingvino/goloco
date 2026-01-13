package effects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include "Effect.h"
// namespace OpenLoco
type MoneyEffect struct {
	EffectEntity // embedded
// union
	Frame uint16
	MoveDelay uint16
}
const MoneyEffectLifetime uint32 = 160
const MoneyEffectRedGreenLifetime uint32 = 55
// orphan member: uint16_t numMovements; // 0x28 Note: this is only used by redGreen money (RCT2 legacy)
// orphan member: int32_t amount;        // 0x2A - currency amount in British pounds - different currencies are probably getting recalculated
// orphan member: CompanyId var_2E;      // company colour?
// orphan member: int16_t offsetX;       // 0x44
// orphan member: uint16_t wiggle;       // 0x46
// func Update() 
// static MoneyEffect* create(const World::Pos3& loc, const CompanyId company, const currency32_t amount);
// static_assert(sizeof(MoneyEffect) <= sizeof(Entity));
