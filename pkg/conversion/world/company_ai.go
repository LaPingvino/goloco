package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
type AiThinkState int

const (
	Unk0 AiThinkState = iota
	Unk1
	Unk2
	Unk3
	Unk4
	Unk5
	Unk6
	Unk7
	Unk8
	Unk9
	EndCompany
)
type AiPlaceVehicleState int

const (
	Begin AiPlaceVehicleState = iota
	ResetList
	Place
	Restart
)
type AiThoughtType int

const (
	Unk0 AiThoughtType = iota
	Unk1
	Unk2
	Unk3
	Unk4
	Unk5
	Unk6
	Unk7
	Unk8
	Unk9
	Unk10
	Unk11
	Unk12
	Unk13
	Unk14
	Unk15
	Unk16
	Unk17
	Unk18
	Unk19
	Null AiThoughtType = 0xFF
)
var AiThoughtTypeCount = 20 // auto
type AiThoughtStationFlags int

const (
	None AiThoughtStationFlags = 0
	AiAllocated AiThoughtStationFlags = 1 << 0
	Operational AiThoughtStationFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(AiThoughtStationFlags);
type AiPurchaseFlags int

const (
	None AiPurchaseFlags = 0
	Unk0 AiPurchaseFlags = 1 << 0
	Unk1 AiPurchaseFlags = 1 << 1
	Unk2 AiPurchaseFlags = 1 << 2
	RequiresMods AiPurchaseFlags = 1 << 3
	Unk4 AiPurchaseFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(AiPurchaseFlags);
var MaxAiThoughts = 60 // auto
var AiThoughtIdNull = 0xFFU // auto
type AiThought struct {
}

type AiThoughtStation struct {
	Id StationId
	Var_02 AiThoughtStationFlags
	Rotation uint8
// World::Pos2 pos;              // 0x4
	BaseZ uint8
	Var_9 uint8
	Var_A uint8
	Var_B uint8
	Var_C uint8
func (a *AiThought) HasFlags(flags AiThoughtStationFlags) bool {
}
	Type AiThoughtType
	DestinationA uint8
	DestinationB uint8
	NumStations uint8
	StationLength uint8
// Station stations[4];           // 0x06 0x4AE Will lists stations created that vehicles will route to
	TrackObjId uint8
	RackRailType uint8
	Mods uint16
	CargoType uint8
	Var_43 uint8
	NumVehicles uint8
	Var_45 uint8
// uint16_t var_46[16];           // 0x4EF array of uint16_t object id
// EntityId vehicles[8];          // 0x66 0x50E see also numVehicles for current size
	Var_76 currency32_t
	Var_7C currency32_t
	Var_80 currency32_t
	Var_84 currency32_t
	Var_88 uint8
	StationObjId uint8
	SignalObjId uint8
	PurchaseFlags AiPurchaseFlags
func (a *AiThought) HasPurchaseFlags(flags AiPurchaseFlags) bool {
	// // Converts the TownId or IndustryId of destinationA into the center position of the destination.
	// World::Pos2 getDestinationPositionA() const;
	// // Converts the TownId or IndustryId of destinationB into the center position of the destination.
	// World::Pos2 getDestinationPositionB() const;
}
	// method: void aiThink(CompanyId id);
	// method: void setAiObservation(CompanyId id);
	// method: void removeEntityFromThought(AiThought& thought, EntityId id);
}
