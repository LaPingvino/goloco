package vehicles

// Vehicle system type definitions

import "github.com/LaPingvino/goloco/pkg/loco/worldmap"

// Vehicle constants
const (
	MaxRoadVehicleLength       = 176
	WheelSlippingDuration      = 64
	AirportMovementNodeNull    = 0xFF
	AirportMovementNoValidEdge = 0xFE
)

// Flags38 are vehicle flags (purpose unclear from name)
type Flags38 uint8

const (
	Flags38None                 Flags38 = 0
	Flags38Unk0                 Flags38 = 1 << 0
	Flags38IsReversed           Flags38 = 1 << 1
	Flags38Unk2                 Flags38 = 1 << 2
	Flags38JacobsBogieAvailable Flags38 = 1 << 3
	Flags38IsGhost              Flags38 = 1 << 4
	Flags38FasterAroundCurves   Flags38 = 1 << 5
)

// SoundFlags control vehicle sound playback
type SoundFlags uint8

const (
	SoundFlagNone SoundFlags = 0
	SoundFlag0    SoundFlags = 1 << 0
	SoundFlag1    SoundFlags = 1 << 1
)

// UpdateVar1136114Flags are update-related flags
type UpdateVar1136114Flags uint16

const (
	UpdateFlagNone                     UpdateVar1136114Flags = 0
	UpdateFlagUnkM00                   UpdateVar1136114Flags = 1 << 0
	UpdateFlagNoRouteFound             UpdateVar1136114Flags = 1 << 1
	UpdateFlagCrashed                  UpdateVar1136114Flags = 1 << 2
	UpdateFlagUnkM03                   UpdateVar1136114Flags = 1 << 3
	UpdateFlagApproachingGradeCrossing UpdateVar1136114Flags = 1 << 4
	UpdateFlagUnkM15                   UpdateVar1136114Flags = 1 << 15
)

// BreakdownFlags track vehicle breakdown state
type BreakdownFlags uint8

const (
	BreakdownNone           BreakdownFlags = 0
	BreakdownUnk0           BreakdownFlags = 1 << 0
	BreakdownPending        BreakdownFlags = 1 << 1
	BreakdownBrokenDown     BreakdownFlags = 1 << 2
	BreakdownJourneyStarted BreakdownFlags = 1 << 3
)

// VehicleFlags are general vehicle flags
type VehicleFlags uint8

const (
	VehicleFlagNone          VehicleFlags = 0
	VehicleFlagUnk0          VehicleFlags = 1 << 0
	VehicleFlagCommandStop   VehicleFlags = 1 << 1
	VehicleFlagUnk2          VehicleFlags = 1 << 2
	VehicleFlagSorted        VehicleFlags = 1 << 3
	VehicleFlagUnk5          VehicleFlags = 1 << 5
	VehicleFlagManualControl VehicleFlags = 1 << 6
	VehicleFlagShuntCheat    VehicleFlags = 1 << 7
)

// VehicleEntityType represents the type of vehicle entity
type VehicleEntityType uint8

const (
	VehicleEntityHead          VehicleEntityType = 0
	VehicleEntityVehicle1      VehicleEntityType = 1
	VehicleEntityVehicle2      VehicleEntityType = 2
	VehicleEntityBogie         VehicleEntityType = 3
	VehicleEntityBodyStart     VehicleEntityType = 4
	VehicleEntityBodyContinued VehicleEntityType = 5
	VehicleEntityTail          VehicleEntityType = 6
)

// TrackAndDirection combines track ID and direction
type TrackAndDirection struct {
	Data uint16
}

// NewTrackAndDirection creates a track and direction from ID and direction
func NewTrackAndDirection(id, direction uint8) TrackAndDirection {
	return TrackAndDirection{Data: uint16(id) | (uint16(direction) << 8)}
}

// GetTrackId extracts the track ID
func (tad *TrackAndDirection) GetTrackId() uint8 {
	return uint8(tad.Data & 0xFF)
}

// GetDirection extracts the direction
func (tad *TrackAndDirection) GetDirection() uint8 {
	return uint8((tad.Data >> 8) & 0xFF)
}

// Vehicle represents a game vehicle. Stub — only a minimal set of fields
// are defined; the full entity layout from OpenLoco is not yet mapped.
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/Vehicle.h
//
//	Vehicles::Vehicle — base entity struct embedded in all vehicle sub-types.
//	Fields include: x, y, z, owner, flags, breakdownFlags, flags38,
//	subType (entity type), among many others.
type Vehicle struct {
	EntityType VehicleEntityType
	Flags      VehicleFlags
	Position   worldmap.Pos3
	Breakdown  BreakdownFlags
	Flags38    Flags38
}

// VehicleHead represents the head entity of a vehicle. Stub — no
// head-specific fields have been added yet.
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleHead.h
//
//	Vehicles::VehicleHead
//	Key fields: status (Status enum), currentSpeed, targetSpeed,
//	  routeIndex, routeHeads, orderStatus, daysAge, ...
//	Key methods: update(), updateBreakdown(), updateDaily(), updateMonthly()
type VehicleHead struct {
	Vehicle
}

// VehicleTail represents the tail entity of a vehicle. Stub — no
// tail-specific fields have been added yet.
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleTail.h
//
//	Vehicles::VehicleTail
//	Key method: update()
type VehicleTail struct {
	Vehicle
}

// VehicleBogie represents a bogie (wheel assembly) entity. Stub — no
// bogie-specific fields have been added yet.
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleBogie.h
//
//	Vehicles::VehicleBogie
//	Key fields: bogieFrame, trackAndDirection, animation
//	Key methods: update(), updateSegmentCrashed(), isOnRackRail()
type VehicleBogie struct {
	Vehicle
}

// VehicleBody represents a vehicle body entity. Stub — no body-specific
// fields have been added yet.
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleBody.h
//
//	Vehicles::VehicleBody
//	Key fields: cargoAmount, cargoType, cargoCapacity, trackAndDirection
//	Key methods: update(const CarUpdateState&), updateCargoSprite()
type VehicleBody struct {
	Vehicle
}
