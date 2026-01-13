package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdint>
// namespace OpenLoco
// forward: struct Entity;
// namespace OpenLoco::S5
type EntityBase struct {
	BaseType uint8
	Type uint8
	NextQuadrantId uint16
	NextEntityId uint16
	LlPreviousId uint16
	LinkedListOffset uint8
	SpriteHeightNegative uint8
	Id uint16
	VehicleFlags uint16
// World::Pos3 position;         // 0x0E
	SpriteWidth uint8
	SpriteHeightPositive uint8
	SpriteLeft int16
	SpriteTop int16
	SpriteRight int16
	SpriteBottom int16
	SpriteYaw uint8
	SpritePitch uint8
	Pad_20 uint8
	Owner uint8
	Name uint16
}
// static_assert(sizeof(EntityBase) == 0x24);
type Exhaust struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
	FrameNum uint16
	StationaryProgress int16
// uint8_t pad_2A[0x32 - 0x2A];
	WindProgress uint16
	Var_34 int16
	Var_36 int16
// uint8_t pad_38[0x49 - 0x38];
	ObjectId uint8
}
type ExplosionCloud struct {
	Base EntityBase
// uint8_t pad_24[0x28 - 0x24];
	Frame uint16
}
type ExplosionSmoke struct {
	Base EntityBase
// uint8_t pad_24[0x28 - 0x24];
	Frame uint16
}
type Fireball struct {
	Base EntityBase
// uint8_t pad_24[0x28 - 0x24];
	Frame uint16
}
type MoneyEffect struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
// union
	Frame uint16
	MoveDelay uint16
}
// orphan member: uint16_t numMovements; // 0x28 Note: this is only used by redGreen money (RCT2 legacy)
// orphan member: int32_t amount;        // 0x2A - currency amount in British pounds - different currencies are probably getting recalculated
// orphan member: uint8_t var_2E;        // company colour?
// uint8_t pad_2F[0x44 - 0x2F];
// orphan member: int16_t offsetX; // 0x44
// orphan member: uint16_t wiggle; // 0x46
type Smoke struct {
	Base EntityBase
// uint8_t pad_24[0x28 - 0x24];
	Frame uint16
}
type Splash struct {
	Base EntityBase
// uint8_t pad_24[0x28 - 0x24];
	Frame uint16
}
type VehicleCrashParticle struct {
	Base EntityBase
// uint8_t pad_24[0x02];          // 0x24
	TimeToLive uint16
	Frame uint16
// uint8_t pad_2A[0x04];          // 0x2A
	ColourSchemePrimary uint8
	ColourSchemeSecondary uint8
	CrashedSpriteBase uint16
// World::Pos3 velocity;          // 0x32
	AccelerationX int32
	AccelerationY int32
	AccelerationZ int32
}
type VehicleCargo struct {
	AcceptedTypes uint32
	Type uint8
	MaxQty uint8
	TownFrom uint16
	NumDays uint8
	Qty uint8
}
// static_assert(sizeof(VehicleCargo) == 0xA);
type VehicleHead struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;    // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	Pad_39 uint8
	NextCarId uint16
	Var_3C int32
// uint8_t pad_40[0x2]; // 0x40
	Mode uint8
	Pad_43 uint8
	OrdinalNumber int16
	OrderTableOffset uint32
	CurrentOrder uint16
	SizeOfOrderTable uint16
	TrainAcceptedCargoTypes uint32
	Var_52 uint8
	Var_53 uint8
	StationId uint16
	CargoTransferTimeout uint16
	Var_58 uint32
	Var_5C uint8
	Status uint8
	VehicleType uint8
	BreakdownFlags uint8
	AiThoughtId uint8
// World::Pos2 aiPlacementPos;        // 0x61
	AiPlacementTaD uint16
	AiPlacementBaseZ uint8
	AirportMovementEdge uint8
	TotalRefundCost uint32
	CrashedTimeout uint8
	ManualPower int8
// World::Pos2 journeyStartPos;       // 0x6F journey start position
	JourneyStartTicks uint32
	LastAverageSpeed int16
	RestartStoppedCarsTimeout uint8
}
// static_assert(sizeof(VehicleHead) == 0x7A); // Can't use offset_of change this to last field if more found
type IncomeStats struct {
	Day int32
// uint8_t cargoTypes[4];        // 0x57
// uint16_t cargoQtys[4];        // 0x5B
// uint16_t cargoDistances[4];   // 0x63
// uint8_t cargoAges[4];         // 0x6B
// currency32_t cargoProfits[4]; // 0x6F
}
// static_assert(sizeof(IncomeStats) == 0x2C);
type Vehicle1 struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;    // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	Pad_39 uint8
	NextCarId uint16
	Var_3C int32
// uint8_t pad_40[0x2]; // 0x40
	Mode uint8
	Pad_43 uint8
	TargetSpeed int16
	TimeAtSignal uint16
	Var_48 uint8
	Var_49 uint8
	DayCreated uint32
	Var_4E uint16
	Var_50 uint16
	Var_52 uint8
	LastIncome IncomeStats
}
// static_assert(sizeof(Vehicle1) == 0x7F); // Can't use offset_of change this to last field if more found
type Vehicle2 struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;    // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	Pad_39 uint8
	NextCarId uint16
	Var_3C int32
// uint8_t pad_40[0x42 - 0x40]; // 0x40
	Mode uint8
	Pad_43 uint8
	DrivingSoundId uint8
	DrivingSoundVolume uint8
	DrivingSoundFrequency uint16
	ObjectId uint16
	SoundFlags uint16
	SoundWindowNumber uint16
	SoundWindowType uint8
	Var_4F int8
	TotalPower uint16
	TotalWeight uint16
	MaxSpeed int16
	CurrentSpeed int32
	MotorState uint8
	BrakeLightTimeout uint8
	RackRailMaxSpeed int16
	CurMonthRevenue currency32_t
// currency32_t profit[4];       // 0x62 last 4 months net profit
	Reliability uint8
	Var_73 uint8
}
// static_assert(sizeof(Vehicle2) == 0x74); // Can't use offset_of change this to last field if more found
type VehicleBody struct {
	Base EntityBase
	ColourSchemePrimary uint8
	ColourSchemeSecondary uint8
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;       // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	ObjectSpriteType uint8
	NextCarId uint16
	Var_3C int32
	ObjectId uint16
	Mode uint8
	Pad_43 uint8
	Var_44 int16
	AnimationFrame uint8
	CargoFrame uint8
	PrimaryCargo VehicleCargo
// uint8_t pad_52[0x54 - 0x52];
	BodyIndex uint8
	ChuffSoundIndex int8
	CreationDay uint32
	Var_5A uint32
	WheelSlipping uint8
	BreakdownFlags uint8
	Pad_60 uint16
	RefundCost uint32
// uint8_t pad_66[0x6A - 0x66];
	BreakdownTimeout uint8
}
// static_assert(sizeof(VehicleBody) == 0x6B); // Can't use offset_of change this to last field if more found
type VehicleBogie struct {
	Base EntityBase
	ColourSchemePrimary uint8
	ColourSchemeSecondary uint8
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;       // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	ObjectSpriteType uint8
	NextCarId uint16
	Var_3C int32
	ObjectId uint16
	Mode uint8
	Pad_43 uint8
	Var_44 uint16
	AnimationIndex uint8
	Var_47 uint8
	SecondaryCargo VehicleCargo
	TotalCarWeight uint16
	BodyIndex uint8
	Pad_55 uint8
	CreationDay uint32
	Var_5A uint32
	WheelSlipping uint8
	BreakdownFlags uint8
	Var_60 uint8
	Var_61 uint8
	RefundCost uint32
	Reliability uint16
	TimeoutToBreakdown uint16
	BreakdownTimeout uint8
}
// static_assert(sizeof(VehicleBogie) == 0x6B); // Can't use offset_of change this to last field if more found
type VehicleTail struct {
	Base EntityBase
// uint8_t pad_24[0x26 - 0x24];
	Head uint16
	RemainingDistance int32
	TrackAndDirection uint16
	SubPosition uint16
	TileX int16
	TileY int16
// World::SmallZ tileBaseZ;    // 0x34
	TrackType uint8
	RoutingHandle uint16
	Var_38 uint8
	Pad_39 uint8
	NextCarId uint16
	Var_3C int32
// uint8_t pad_40[0x42 - 0x40]; // 0x40
	Mode uint8
	Pad_43 uint8
	DrivingSoundId uint8
	DrivingSoundVolume uint8
	DrivingSoundFrequency uint16
	ObjectId uint16
	SoundFlags uint16
	SoundWindowNumber uint16
	SoundWindowType uint8
	TrainDanglingTimeout uint16
}
// static_assert(sizeof(VehicleTail) == 0x51); // Can't use offset_of change this to last field if more found
type Entity struct {
	Base EntityBase
// uint8_t pad_24[0x80 - 0x24]; // Type specific data
}
// static_assert(sizeof(Entity) == 0x80);
// S5::Entity exportEntity(const OpenLoco::Entity& src);
// OpenLoco::Entity importEntity(const S5::Entity& src);
