package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include "Objects/Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
type ExpenditureType int

const (
type GameSpeed int

const (
type LoadOrQuitMode int

const (
)
// namespace OpenLoco::World
// forward: struct TileElement;
// forward: struct WallElement;
// namespace OpenLoco::World::TileManager
type ElementPositionFlags int

const (
)
// namespace OpenLoco::Vehicles
// forward: struct VehicleHead;
// namespace OpenLoco::GameCommands
// namespace Flags
const Apply uint8 = 1 << 0
const PreventBuildingClearing uint8 = 1 << 1
const AllowNegativeCashFlow uint8 = 1 << 2
const NoErrorWindow uint8 = 1 << 3
const AiAllocated uint8 = 1 << 4
const NoPayment uint8 = 1 << 5
const Ghost uint8 = 1 << 6
const Flag_7 uint8 = 1 << 7
type GameCommand int

const (
	VehicleRearrange GameCommand = 0
	VehiclePlace GameCommand = 1
	VehiclePickup GameCommand = 2
	VehicleReverse GameCommand = 3
	VehiclePassSignal GameCommand = 4
	VehicleCreate GameCommand = 5
	VehicleSell GameCommand = 6
	CreateTrack GameCommand = 7
	RemoveTrack GameCommand = 8
	ChangeLoan GameCommand = 9
	VehicleRename GameCommand = 10
	ChangeStationName GameCommand = 11
	VehicleChangeRunningMode GameCommand = 12
	CreateSignal GameCommand = 13
	RemoveSignal GameCommand = 14
	CreateTrainStation GameCommand = 15
	RemoveTrainStation GameCommand = 16
	CreateTrackMod GameCommand = 17
	RemoveTrackMod GameCommand = 18
	ChangeCompanyColourScheme GameCommand = 19
	PauseGame GameCommand = 20
	LoadSaveQuitGame GameCommand = 21
	RemoveTree GameCommand = 22
	CreateTree GameCommand = 23
	ChangeLandMaterial GameCommand = 24
	RaiseLand GameCommand = 25
	LowerLand GameCommand = 26
	LowerRaiseLandMountain GameCommand = 27
	RaiseWater GameCommand = 28
	LowerWater GameCommand = 29
	ChangeCompanyName GameCommand = 30
	ChangeCompanyOwnerName GameCommand = 31
	CreateWall GameCommand = 32
	RemoveWall GameCommand = 33
	Gc_unk_34 GameCommand = 34
	VehicleOrderInsert GameCommand = 35
	VehicleOrderDelete GameCommand = 36
	VehicleOrderSkip GameCommand = 37
	CreateRoad GameCommand = 38
	RemoveRoad GameCommand = 39
	CreateRoadMod GameCommand = 40
	RemoveRoadMod GameCommand = 41
	CreateRoadStation GameCommand = 42
	RemoveRoadStation GameCommand = 43
	CreateBuilding GameCommand = 44
	RemoveBuilding GameCommand = 45
	RenameTown GameCommand = 46
	CreateIndustry GameCommand = 47
	RemoveIndustry GameCommand = 48
	CreateTown GameCommand = 49
	RemoveTown GameCommand = 50
	AiCreateTrackAndStation GameCommand = 51
	AiTrackReplacement GameCommand = 52
	AiCreateRoadAndStation GameCommand = 53
	BuildCompanyHeadquarters GameCommand = 54
	RemoveCompanyHeadquarters GameCommand = 55
	CreateAirport GameCommand = 56
	RemoveAirport GameCommand = 57
	VehiclePlaceAir GameCommand = 58
	VehiclePickupAir GameCommand = 59
	CreatePort GameCommand = 60
	RemovePort GameCommand = 61
	VehiclePlaceWater GameCommand = 62
	VehiclePickupWater GameCommand = 63
	VehicleRefit GameCommand = 64
	ChangeCompanyFace GameCommand = 65
	ClearLand GameCommand = 66
	LoadMultiplayerMap GameCommand = 67
	Gc_unk_68 GameCommand = 68
	Gc_unk_69 GameCommand = 69
	Gc_unk_70 GameCommand = 70
	SendChatMessage GameCommand = 71
	MultiplayerSave GameCommand = 72
	UpdateOwnerStatus GameCommand = 73
	VehicleSpeedControl GameCommand = 74
	VehicleOrderUp GameCommand = 75
	VehicleOrderDown GameCommand = 76
	VehicleApplyShuntCheat GameCommand = 77
	ApplyFreeCashCheat GameCommand = 78
	RenameIndustry GameCommand = 79
	VehicleClone GameCommand = 80
	Cheat GameCommand = 81
	SetGameSpeed GameCommand = 82
	VehicleOrderReverse GameCommand = 83
	VehicleRepaint GameCommand = 84
)
const DefaultRegValue int32 = 0xCCCCCCCC
// /**
// * x86 register structure, only used for easy interop to Locomotion code.
// */
type Registers struct {
// union
	Eax int32
	Ax int16
// struct
	Al int8
	Ah int8
}
// union
// orphan member: int32_t ebx{ kDefaultRegValue };
// orphan member: int16_t bx;
// struct
// orphan member: int8_t bl;
// orphan member: int8_t bh;
// union
// orphan member: int32_t ecx{ kDefaultRegValue };
// orphan member: int16_t cx;
// struct
// orphan member: int8_t cl;
// orphan member: int8_t ch;
// union
// orphan member: int32_t edx{ kDefaultRegValue };
// orphan member: int16_t dx;
// struct
// orphan member: int8_t dl;
// orphan member: int8_t dh;
// union
// orphan member: int32_t esi{ kDefaultRegValue };
// orphan member: int16_t si;
// union
// orphan member: int32_t edi{ kDefaultRegValue };
// orphan member: int16_t di;
// union
// orphan member: int32_t ebp{ kDefaultRegValue };
// orphan member: int16_t bp;
// static_assert(sizeof(registers) == 7 * 4);
const FAILURE uint32 = 0x80000000
// func DoCommand(command GameCommand, registers registers) uint32
// func DoCommandForReal(command GameCommand, company CompanyId, registers registers) uint32
// func Sub_431E6A(company CompanyId, nullptr World::TileElement tile =) bool
// template<typename T>
// func DoCommand(args T, flags uint8) uint32
// registers regs = registers(args);
// regs.bl = flags;
// func DoCommand(T::command, regs) return
// // Load multiplayer map
// func Do_67(filename [[maybe_unused]] char) 
// orphan member: registers regs;
// regs.bl = Flags::apply;
// // This is commented out as it will not work on 64-bit builds
// // regs.ebp = reinterpret_cast<uint32_t>(filename);
// doCommand(GameCommand::loadMultiplayerMap, regs);
// // Multiplayer-related
// func Do_69() 
// orphan member: registers regs;
// regs.bl = Flags::apply;
// doCommand(GameCommand::gc_unk_69, regs);
// // Multiplayer-related
// func Do_70() 
// orphan member: registers regs;
// regs.bl = Flags::apply;
// doCommand(GameCommand::gc_unk_70, regs);
// // Send chat message
// func Do_71(ax int32, string byte) 
// orphan member: registers regs;
// regs.bl = Flags::apply;
// regs.ax = ax;
// memcpy(&regs.ecx, &string[0], 4);
// memcpy(&regs.edx, &string[4], 4);
// memcpy(&regs.ebp, &string[8], 4);
// memcpy(&regs.edi, &string[12], 4);
// doCommand(GameCommand::sendChatMessage, regs);
// // Multiplayer save
// func Do_72() 
// orphan member: registers regs;
// regs.bl = Flags::apply;
// doCommand(GameCommand::multiplayerSave, regs);
// const World::Pos3& getPosition();
// func SetPosition(pos World::Pos3) 
// func SetErrorSound(state bool) 
// func SetErrorText(message StringId) 
// func GetErrorText() StringId
// func SetErrorTitle(title StringId) 
// func GetExpenditureType() ExpenditureType
// func SetExpenditureType(type ExpenditureType) 
// func GetUpdatingCompanyId() CompanyId
// func SetUpdatingCompanyId(companyId CompanyId) 
// func GetCommandNestLevel() uint8
// func ResetCommandNestLevel() 
// // TODO: rework these
type LegacyReturnState struct {
// World::TileManager::ElementPositionFlags flags_1136072; // 0x01136072
	Flags_1136073 uint8
// World::MicroZ byte_1136074;                             // 0x01136074
	Byte_1136075 uint8
	LastPlacedTrackRoadStationId StationId
	LastConstructedAdjoiningStation StationId
	LastPlacedAirport StationId
	LastPlacedDock StationId
// World::Pos2 lastConstructedAdjoiningStationPos;         // 0x0112C792 centre pos
	LastPlacedIndustryId IndustryId
// World::WallElement* lastPlacedWall;                     // 0x01136470
	LastCreatedVehicleId EntityId
	AlternateRoadObjectId uint8
}
// // Note: this is deliberately a mutable ref
// LegacyReturnState& getLegacyReturnState();
// func PlayConstructionPlacementSound(pos World::Pos3) 
// func ShouldInvalidateTile(flags uint8) bool
