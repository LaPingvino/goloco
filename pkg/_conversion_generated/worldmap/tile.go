package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "TileElement.h"
// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// #include <array>
// #include <cassert>
// #include <cstddef>
// #include <cstdint>
// #include <limits>
// #include <tuple>
// namespace OpenLoco::Ui
// forward: struct viewport_pos;
// namespace OpenLoco::World
type TileHeight struct {
	LandHeight coord_t
	WaterHeight coord_t
// explicit operator coord_t() const
// return waterHeight == 0 ? landHeight : waterHeight;
}
// // 0x004F9296, 0x004F9298
var Offsets = [4]Pos2{
// Pos2{ 0, 0 },
// Pos2{ 0, 32 },
// Pos2{ 32, 32 },
// Pos2{ 32, 0 },
// // 0x00503C6C, 0x00503C6E
var RotationOffset = [16]Pos2{
// Pos2{ -32, 0 },
// Pos2{ 0, 32 },
// Pos2{ 32, 0 },
// Pos2{ 0, -32 },
// Pos2{ -32, 0 },
// Pos2{ 0, 32 },
// Pos2{ 32, 0 },
// Pos2{ 0, -32 },
// Pos2{ -32, 0 },
// Pos2{ 0, 32 },
// Pos2{ 32, 0 },
// Pos2{ 0, -32 },
// Pos2{ -32, 32 },
// Pos2{ 32, 32 },
// Pos2{ 32, -32 },
// Pos2{ -32, -32 },
// // 0x00503CAC
var ReverseRotation = [16]uint8{
// 2,
// 3,
// 0,
// 1,
// 10,
// 11,
// 8,
// 9,
// 6,
// 7,
// 4,
// 5,
// 14,
// 15,
// 12,
// 13,
type LessThanPos3 struct {
	// bool operator()(World::Pos3 const& lhs, World::Pos3 const& rhs) const
// return std::tie(lhs.x, lhs.y, lhs.z) < std::tie(rhs.x, rhs.y, rhs.z);
}
// Ui::viewport_pos gameToScreen(const Pos3& loc, int rotation);
// forward: struct SurfaceElement;
// forward: struct StationElement;
type Tile struct {
}

type TileIterator struct {
type Iterator_concept = std::forward_iterator_tag
type Value_type = TileElement
type Difference_type = std::ptrdiff_t
type Pointer = TileElement
type Reference = TileElement
// TileElement* _current{};
	// method: constexpr Iterator() = default;
	// method: constexpr Iterator(TileElement* current)
// : _current(current)
}
	// constexpr TileElement& operator*() const
// return *_current;
}
// constexpr TileElement* operator->() const
// orphan member: return _current;
// constexpr Iterator& operator++()
// if (_current == nullptr)
// return *this;
// if (_current->isLast())
// _current = nullptr;
// else
// _current = _current->next();
// return *this;
// constexpr Iterator operator++(int)
// Iterator result = *this;
// ++(*this);
// orphan member: return result;
// constexpr auto operator<=>(const Iterator& other) const = default;
// private:
// TileElement* const _data;
// public:
const Npos int = std.numeric_limits<size_t>().max()
// const TilePos2 pos;
// Tile(const TilePos2& tPos, TileElement* data);
// func IsNull() bool
// func Begin() Iterator
// func Begin() Iterator
// func End() Iterator
// func End() Iterator
// func Size() int
// TileElement* operator[](size_t i);
// func IndexOf(element TileElementBase) int
// SurfaceElement* surface() const;
// StationElement* trainStation(uint8_t trackId, uint8_t direction, uint8_t baseZ) const;
// StationElement* roadStation(uint8_t roadId, uint8_t direction, uint8_t baseZ) const;
