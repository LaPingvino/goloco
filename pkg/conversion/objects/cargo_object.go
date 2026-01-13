package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
type CargoObjectFlags int

const (
	None CargoObjectFlags = 0
	Unk0 CargoObjectFlags = 1 << 0
	Refit CargoObjectFlags = 1 << 1
	Delivering CargoObjectFlags = 1 << 2
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(CargoObjectFlags);
// // TODO: Eventually perhaps make this a separate object type
type CargoCategory int

const (
	Grain CargoCategory = 1
	Coal
	IronOre
	Liquids
	Paper
	Steel
	Timber
	Goods
	Foods
	Livestock CargoCategory = 11
	Passengers
	Mail
// // Note: Mods may (and do) use other numbers not covered by this list
	Null CargoCategory = 0xFFFFU
)
type CargoObject struct {
	Name StringId
	UnitWeight uint16
	CargoTransferTime uint16
	UnitsAndCargoName StringId
	UnitNameSingular StringId
	UnitNamePlural StringId
	UnitInlineSprite uint32
	CargoCategory CargoCategory
	Flags CargoObjectFlags
	NumPlatformVariations uint8
	StationCargoDensity uint8
	PremiumDays uint8
	MaxNonPremiumDays uint8
	NonPremiumRate uint16
	PenaltyRate uint16
	PaymentFactor uint16
	PaymentIndex uint8
	UnitSize uint8
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: constexpr bool hasFlags(CargoObjectFlags flagsToTest) const
// return (flags & flagsToTest) != CargoObjectFlags::none;
}
const CargoObjectObjectType any = ObjectType.cargo
// static_assert(sizeof(CargoObject) == 0x1F);
// namespace Cargo::ImageIds
const InlineSprite uint32 = 0
// // There are numPlatformVariations sprites after this one
const StationPlatformBegin uint32 = 1
