package entities

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Location.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdint>
// #include <limits>
// namespace OpenLoco
type EntityBaseType int

const (
	Vehicle EntityBaseType = 0
	Effect
	Null EntityBaseType = 0xFF
)
type Pitch int

const (
// //                               actual angle (for trig)
	Flat Pitch = 0
	Up6deg Pitch = 1
	Up12deg Pitch = 2
	Up18deg Pitch = 3
	Up25deg Pitch = 4
	Down6deg Pitch = 5
	Down12deg Pitch = 6
	Down18deg Pitch = 7
	Down25deg Pitch = 8
	Up10deg Pitch = 9
	Down10deg Pitch = 10
	Up20deg Pitch = 11
	Down20deg Pitch = 12
)
type EntityBase struct {
	BaseType EntityBaseType
	LinkedListOffset uint8
	Id EntityId
	NextQuadrantId EntityId
	NextEntityId EntityId
	LlPreviousId EntityId
// World::Pos3 position;
	Owner CompanyId
	SpriteHeightNegative uint8
	SpriteHeightPositive uint8
	SpriteWidth uint8
	SpriteYaw uint8
	SpritePitch Pitch
	SpriteLeft int16
	SpriteTop int16
	SpriteRight int16
	SpriteBottom int16
	Name StringId
	// method: void moveTo(const World::Pos3& loc);
	// method: void invalidateSprite();
// template<typename T>
	// method: bool isBase() const
// return baseType == T::kBaseType;
}
// template<typename BaseType>
// BaseType* asBase()
// return isBase<BaseType>() ? reinterpret_cast<BaseType*>(this) : nullptr;
// template<typename BaseType>
// const BaseType* asBase() const
// return isBase<BaseType>() ? reinterpret_cast<const BaseType*>(this) : nullptr;
// func Empty() bool
// // Max size of a Entity. Use when needing to know Entity size
type Entity struct {
	EntityBase // embedded
// uint8_t pad_24[0x80 - 0x20];
}
// static_assert(sizeof(Entity) == 0x80);
