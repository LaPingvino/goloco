package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <array>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type TrainSignalObjectFlags int

const (
	None TrainSignalObjectFlags = 0
	IsLeft TrainSignalObjectFlags = 1 << 0
	HasLights TrainSignalObjectFlags = 1 << 1
	Unk2 TrainSignalObjectFlags = 1 << 2
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TrainSignalObjectFlags);
// namespace TrainSignal::ImageIds
const RedLights uint32 = 80
const RedLights2 uint32 = 88
const GreenLights uint32 = 96
const GreenLights2 uint32 = 104
type TrainSignalObject struct {
	Name StringId
	Flags TrainSignalObjectFlags
	AnimationSpeed uint8
	NumFrames uint8
	CostFactor int16
	SellCostFactor int16
	CostIndex uint8
	Var_0B uint8
	Description StringId
	Image uint32
	NumCompatible uint8
// uint8_t mods[7];       // 0x13
	DesignedYear uint16
	ObsoleteYear uint16
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: constexpr bool hasFlags(TrainSignalObjectFlags flagsToTest) const
// return (flags & flagsToTest) != TrainSignalObjectFlags::none;
}
const TrainSignalObjectObjectType any = ObjectType.trackSignal
// static_assert(sizeof(TrainSignalObject) == 0x1E);
var SignalFrames2State = std.to_array<uint8_t>({ 1, 2, 3, 3, 3, 3, 3, 3, 2, 1, 0, 0, 0, 0, 0 }) // auto
var SignalFrames3State = std.to_array<uint8_t>({ 1, 2, 3, 3, 3, 3, 3, 3, 3, 4, 5, 6, 6, 6, 6, 6, 6, 6, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0 }) // auto
var SignalFrames4State = std.to_array<uint8_t>({ 1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 5, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 8, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 }) // auto
// static constexpr auto signalFrames = std::to_array<std::span<const uint8_t>>({
// signalFrames2State,
// signalFrames3State,
// signalFrames4State,
// });
