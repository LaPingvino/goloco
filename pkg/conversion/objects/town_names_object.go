package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
type TownNamesObject struct {
}

type TownNamesObjectCategory struct {
	Count uint8
	Bias uint8
	Offset uint16
}
	Name StringId
// Category categories[6]; // 0x02
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}
const TownNamesObjectObjectType any = ObjectType.townNames
const TownNamesObjectMinNumNameCombinations any = 80
// static_assert(sizeof(TownNamesObject) == 0x1A);
