package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Localisation/StringIds.h"
// #include "Object.h"
// #include "Speed.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type TransportMode int

const (
	Rail TransportMode = 0
	Road TransportMode = 1
	Air TransportMode = 2
	Water TransportMode = 3
)
type EmitterAnimationType int

const (
	None EmitterAnimationType = 0
	Steam_puff1
	Steam_puff2
	Steam_puff3
	Diesel_exhaust1
	Electric_spark1
	Electric_spark2
	Diesel_exhaust2
	Ship_wake
)
// namespace SpriteIndex
const Null uint8 = 0xFF
const IsReversed uint8 = (1 << 7)
type VehicleObjectFrictionSound struct {
	SoundObjectId uint8
	MinSpeed Speed32
	SpeedFreqFactor uint8
	BaseFrequency uint16
	SpeedVolumeFactor uint8
	BaseVolume uint8
	MaxVolume uint8
}
// static_assert(sizeof(VehicleObjectFrictionSound) == 0xB);
type VehicleSimpleMotorSound struct {
	SoundObjectId uint8
	IdleFrequency uint16
	IdleVolume uint8
	CoastingFrequency uint16
	CoastingVolume uint8
	AccelerationBaseFreq uint16
	AcclerationVolume uint8
	FreqIncreaseStep uint16
	FreqDecreaseStep uint16
	VolumeIncreaseStep uint8
	VolumeDecreaseStep uint8
	SpeedFreqFactor uint8
}
// static_assert(sizeof(VehicleSimpleMotorSound) == 0x11);
type VehicleGearboxMotorSound struct {
	SoundObjectId uint8
	IdleFrequency uint16
	IdleVolume uint8
	FirstGearFrequency uint16
	FirstGearSpeed Speed16
	SecondGearFreqOffset uint16
	SecondGearSpeed Speed16
	ThirdGearFreqOffset uint16
	ThirdGearSpeed Speed16
	FourthGearFreqOffset uint16
	CoastingVolume uint8
	AcceleratingVolume uint8
	FreqIncreaseStep uint16
	FreqDecreaseStep uint16
	VolumeIncreaseStep uint8
	VolumeDecreaseStep uint8
	SpeedFreqFactor uint8
}
// static_assert(sizeof(VehicleGearboxMotorSound) == 0x1B);
type VehicleObjectEmitterAnimation struct {
	ObjectId uint8
	EmitterVerticalPos uint8
	Type EmitterAnimationType
}
// static_assert(sizeof(VehicleObjectEmitterAnimation) == 0x3);
type VehicleObjectCar struct {
	FrontBogiePosition uint8
	BackBogiePosition uint8
	FrontBogieSpriteInd uint8
	BackBogieSpriteInd uint8
	BodySpriteInd uint8
	EmitterHorizontalPos uint8
}
// static_assert(sizeof(VehicleObjectCar) == 0x6);
type BogieSpriteFlags int

const (
	None BogieSpriteFlags = 0
	HasSprites BogieSpriteFlags = 1 << 0
	RotationalSymmetry BogieSpriteFlags = 1 << 1
	HasGentleSprites BogieSpriteFlags = 1 << 2
	HasSteepSprites BogieSpriteFlags = 1 << 3
	LargerBoundingBox BogieSpriteFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(BogieSpriteFlags);
type VehicleObjectBogieSprite struct {
	NumAnimationFrames uint8
	Flags BogieSpriteFlags
	Width uint8
	HeightNegative uint8
	HeightPositive uint8
	NumFramesPerRotation uint8
	FlatImageIds uint32
	GentleImageIds uint32
	SteepImageIds uint32
	// method: constexpr bool hasFlags(BogieSpriteFlags flagsToTest) const
// return (flags & flagsToTest) != BogieSpriteFlags::none;
}
// static_assert(sizeof(VehicleObjectBogieSprite) == 0x12);
type BodySpriteFlags int

const (
	None BodySpriteFlags = 0
	HasSprites BodySpriteFlags = 1 << 0
	RotationalSymmetry BodySpriteFlags = 1 << 1
	Flag02Deprecated BodySpriteFlags = 1 << 2
	HasGentleSprites BodySpriteFlags = 1 << 3
	HasSteepSprites BodySpriteFlags = 1 << 4
	HasBrakingLights BodySpriteFlags = 1 << 5
	HasSpeedAnimation BodySpriteFlags = 1 << 6
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(BodySpriteFlags);
type VehicleObjectBodySprite struct {
	NumFlatRotationFrames uint8
	NumSlopedRotationFrames uint8
	NumAnimationFrames uint8
	NumCargoLoadFrames uint8
	NumCargoFrames uint8
	NumRollFrames uint8
	HalfLength uint8
	Flags BodySpriteFlags
	Width uint8
	HeightNegative uint8
	HeightPositive uint8
	FlatYawAccuracy uint8
	SlopedYawAccuracy uint8
	NumFramesPerRotation uint8
	FlatImageId uint32
	UnkImageId uint32
	GentleImageId uint32
	SteepImageId uint32
	// method: constexpr bool hasFlags(BodySpriteFlags flagsToTest) const
// return (flags & flagsToTest) != BodySpriteFlags::none;
}
type VehicleObjectFlags int

const (
// // See github issue https://github.com/OpenLoco/OpenLoco/issues/2877 for discussion on unnamed flags
	None VehicleObjectFlags = 0
	AlternatingDirection VehicleObjectFlags = 1 << 0
	TopAndTailPosition VehicleObjectFlags = 1 << 1
	JacobsBogieFront VehicleObjectFlags = 1 << 2
	JacobsBogieRear VehicleObjectFlags = 1 << 3
	Flag_04 VehicleObjectFlags = 1 << 4
	CenterPosition VehicleObjectFlags = 1 << 5
	RackRail VehicleObjectFlags = 1 << 6
// // Alternates between sprite 0 and sprite 1 for each vehicle of this type in a train
// // NOTE: This is for vehicles and not vehicle components (which can also do similar)
	AlternatingCarSprite VehicleObjectFlags = 1 << 7
	Flag_08 VehicleObjectFlags = 1 << 8
	AircraftIsTaildragger VehicleObjectFlags = 1 << 8
	AnyRoadType VehicleObjectFlags = 1 << 9
	Flag_10 VehicleObjectFlags = 1 << 10
	CannotCoupleToSelf VehicleObjectFlags = 1 << 11
	AircraftFlaresLanding VehicleObjectFlags = 1 << 11
	MustHavePair VehicleObjectFlags = 1 << 12
	CanWheelslip VehicleObjectFlags = 1 << 13
	AircraftIsHelicopter VehicleObjectFlags = 1 << 13
	Refittable VehicleObjectFlags = 1 << 14
	QuietInvention VehicleObjectFlags = 1 << 15
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(VehicleObjectFlags);
type DrivingSoundType int

const (
	None DrivingSoundType = iota
	Friction
	SimpleMotor
	GearboxMotor
)
// namespace NumStartSounds
const HasCrossingWhistle uint8 = 1 << 7
const Mask uint8 = 0x7F
type VehicleObject struct {
	Name StringId
	Mode TransportMode
	Type VehicleType
	NumCarComponents uint8
	TrackType uint8
	NumTrackExtras uint8
	CostIndex uint8
	CostFactor int16
	Reliability uint8
	RunCostIndex uint8
	RunCostFactor int16
	ColourType uint8
	NumCompatibleVehicles uint8
// uint16_t compatibleVehicles[8];                       // 0x10 array of compatible vehicle_types
// uint8_t requiredTrackExtras[4];                       // 0x20
// VehicleObjectCar carComponents[kMaxCarComponents];    // 0x24
// VehicleObjectBodySprite bodySprites[kMaxBodySprites]; // 0x3C
// VehicleObjectBogieSprite bogieSprites[2];             // 0xB4
	Power uint16
	Speed Speed16
	RackSpeed Speed16
	Weight uint16
	Flags VehicleObjectFlags
// uint8_t maxCargo[2];                                  // 0xE2 size is relative to the first compatibleCargoCategories
// uint32_t compatibleCargoCategories[2];                // 0xE4
// uint8_t cargoTypeSpriteOffsets[32];                   // 0xEC
	NumSimultaneousCargoTypes uint8
// VehicleObjectEmitterAnimation animation[2];           // 0x10D
	ShipWakeSpacing uint8
	Designed uint16
	Obsolete uint16
	RackRailType uint8
	DrivingSoundType DrivingSoundType
// union
	Friction VehicleObjectFrictionSound
	SimpleMotor VehicleSimpleMotorSound
	GearboxMotor VehicleGearboxMotorSound
// } sound;
// uint8_t pad_135[0x15A - 0x135];
	NumStartSounds uint8
// SoundObjectId_t startSounds[kMaxStartSounds]; // 0x15B sound array length numStartSounds highest sound is the crossing whistle
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y, const int16_t width) const;
	// method: void getCargoString(char* buffer) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
	// method: uint32_t getLength() const;
	// method: constexpr bool hasFlags(VehicleObjectFlags flagsToTest) const
// return (flags & flagsToTest) != VehicleObjectFlags::none;
}
const VehicleObjectObjectType any = ObjectType.vehicle
const VehicleObjectMaxBodySprites any = 4
const VehicleObjectMaxCarComponents any = 4
const VehicleObjectMaxStartSounds any = 3
// static_assert(sizeof(VehicleObject) == 0x15E);
// namespace StringIds
// func GetVehicleType(type VehicleType) StringId
// switch (type)
// case VehicleType::train:
// return StringIds::train;
// case VehicleType::bus:
// return StringIds::bus;
// case VehicleType::truck:
// return StringIds::truck;
// case VehicleType::tram:
// return StringIds::tram;
// case VehicleType::aircraft:
// return StringIds::aircraft;
// case VehicleType::ship:
// return StringIds::ship;
// return StringIds::empty;
