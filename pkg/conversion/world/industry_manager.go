package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Industry.h"
// #include "Objects/IndustryObject.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/LocoFixedVector.hpp>
// #include <array>
// #include <cstddef>
// namespace OpenLoco::IndustryManager
type Flags int

const (
	None Flags = 0
	DisallowIndustriesCloseDown Flags = 1 << 0
	DisallowIndustriesStartUp Flags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(Flags);
// func Reset() 
// func Industries() any /* FixedVector<Industry, Limits::kMaxIndustries> */ 
// Industry* get(IndustryId id);
// func GetFlags() Flags
// func HasFlags(flags Flags) bool
// func SetFlags(flags Flags) 
// func Update() 
// func UpdateDaily() 
// func GetMostCommonBuildingCargoType() uint8
// func UpdateMonthly() 
// func CapOfTypeOfIndustry(indObjId uint8, numIndustriesFactor uint8) int32
// func CreateNewIndustry(indObjId uint8, buildImmediately bool, numAttempts int32) 
// func CreateAllMapAnimations() 
// func IndustryNearPosition(position World::Pos2, flags IndustryObjectFlags) bool
// func UpdateProducedCargoStats() 
// func AllocateNewIndustry(type uint8, pos World::Pos2, prng Core::Prng, nearbyTown TownId) IndustryId
