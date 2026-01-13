package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
type ClimateObject struct {
	Name StringId
	FirstSeason uint8
// uint8_t seasonLength[4]; // 0x03
	WinterSnowLine uint8
	SummerSnowLine uint8
	Pad_09 uint8
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}
const ClimateObjectObjectType any = ObjectType.climate
// static_assert(sizeof(ClimateObject) == 0xA);
