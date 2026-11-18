package scenario

import (
	"encoding/binary"

	"github.com/LaPingvino/goloco/pkg/coords"
)

// s5entities.go — Go types mirroring the S5 entity table binary format.
//
// The entity table is embedded in the GameState chunk of an .SV5 save file.
// For .SC5 scenarios the GeneralState chunk is written separately (and is
// much smaller); the entity table is not present in scenario files.
//
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h
//                     src/OpenLoco/src/S5/S5GameState.h  (struct GameState)
//                     src/OpenLoco/src/S5/Limits.h       (kMaxEntities = 20000)

// Entity table layout within the decompressed GameState chunk.
// The GameState is laid out as:
//
//	[0x000000] GeneralState  (0xB96C bytes)
//	[0x00B96C] Companies     (15 × Company)
//	[0x092444] Towns         (80 × Town)
//	[0x09E744] Industries    (128 × Industry)
//	[0x0C10C4] Stations      (1024 × Station)
//	[0x1B58C4] Entities      (20000 × 0x80 bytes)   ← this constant
//	[0x4268C4] Animations    …
//
// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
const (
	// EntityTableOffset is the byte offset of the entity array within the
	// full decompressed GameState chunk for the original GameState format.
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h
	//   struct GameState { Entity entities[...]; // 0x1B58C4 }
	//   static_assert(sizeof(GameState) == 0x4A0644)
	EntityTableOffset = 0x1B58C4

	// EntitySize is the size of a single entity slot in bytes.
	// OpenLoco reference: static_assert(sizeof(Entity) == 0x80)
	EntitySize = 0x80

	// MaxEntities is the total number of entity slots in the table.
	// OpenLoco reference: S5::Limits::kMaxEntities = 20000
	MaxEntities = 20_000
)

// GeneralState field offsets (within the GameState / GeneralState chunk).
// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  struct GeneralState
const (
	// GeneralFlagsOffset is the offset of the uint32 flags field.
	// Bit 0 = tileManagerLoaded (tile elements follow in the file).
	GeneralFlagsOffset = 0x10

	// GeneralCurrentYearOffset is the offset of the uint16 current year.
	GeneralCurrentYearOffset = 0x1A

	// GeneralCurrentMonthOffset is the offset of the uint8 current month (0-11).
	GeneralCurrentMonthOffset = 0x1C

	// GeneralCurrentDayOffset is the offset of the uint8 current day-of-month (1-31).
	GeneralCurrentDayOffset = 0x1D

	// GeneralEntityListHeadsOffset is the offset of the array
	//   uint16_t entityListHeads[7]
	// Each element is the entity-slot ID of the head of one linked list.
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  entityListHeads
	GeneralEntityListHeadsOffset = 0x26

	// GeneralEntityListCountsOffset is the offset of the array
	//   uint16_t entityListCounts[7]
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  entityListCounts
	GeneralEntityListCountsOffset = 0x34

	// NumEntityLists is the number of linked entity lists.
	// OpenLoco reference: S5::Limits::kNumEntityLists = 7
	NumEntityLists = 7
)

// Entity base-type values — EntityBase byte 0x00.
// A free/unoccupied slot has baseType == 0xFF.
// OpenLoco reference: src/OpenLoco/src/Entities/EntityManager.h
const (
	EntityBaseTypeVehicle = uint8(0x00) // vehicle (all 6 subtypes)
	EntityBaseTypeFree    = uint8(0xFF) // free / unoccupied slot
	EntityBaseTypeEffect  = uint8(0x01) // effect (money, exhaust, …)
)

// Vehicle entity sub-type values — EntityBase byte 0x01.
// Valid only when EntityBase.baseType == EntityBaseTypeVehicle.
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleBase.h  VehicleEntityType
const (
	VehicleTypeHead          = uint8(0) // VehicleHead  — lead car; has status, orders, route
	VehicleTypeVeh1          = uint8(1) // Vehicle1     — front physics / income-tracking car
	VehicleTypeVeh2          = uint8(2) // Vehicle2     — speed / power / weight / sound car
	VehicleTypeBogie         = uint8(3) // VehicleBogie — undercarriage / wheels (has sprites)
	VehicleTypeBodyStart     = uint8(4) // VehicleBody  — the first body of a car (has sprites)
	VehicleTypeBodyContinued = uint8(5) // VehicleBody  — a car's further component bodies
	VehicleTypeTail          = uint8(6) // VehicleTail  — rear car; sound output
)

// IsVehicleBody reports a body entity (body_start or body_continued).
func IsVehicleBody(entityType uint8) bool {
	return entityType == VehicleTypeBodyStart || entityType == VehicleTypeBodyContinued
}

// VehicleHead status values — VehicleHead byte 0x5D.
// OpenLoco reference: include/OpenLoco/Vehicles/VehicleHead.h  Status
// (FIXED 2026-11-02: these were renumbered from 4 on; world/vehiclestatus.go
// reads them).
const (
	VehicleStatusUnplaced        = uint8(0) // unk_0: no position (not placed)
	VehicleStatusStopped         = uint8(1)
	VehicleStatusTravelling      = uint8(2)
	VehicleStatusWaitingAtSignal = uint8(3)
	VehicleStatusApproaching     = uint8(4)
	VehicleStatusUnloading       = uint8(5)
	VehicleStatusLoading         = uint8(6)
	VehicleStatusBrokenDown      = uint8(7)
	VehicleStatusCrashed         = uint8(8)
	VehicleStatusStuck           = uint8(9)
	VehicleStatusLanding         = uint8(10)
	VehicleStatusTaxiing1        = uint8(11)
	VehicleStatusTaxiing2        = uint8(12)
	VehicleStatusTakingOff       = uint8(13)
)

// EntityBase byte offsets — sizeof = 0x24.
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  struct EntityBase
const (
	ebBaseType         = 0x00 // uint8  — EntityBaseTypeVehicle / EntityBaseTypeFree
	ebType             = 0x01 // uint8  — VehicleType* or misc entity subtype
	ebNextQuadrantId   = 0x02 // uint16 — spatial hash next-entity for quadrant
	ebNextEntityId     = 0x04 // uint16 — intrusive linked-list next ID
	ebLLPreviousId     = 0x06 // uint16 — intrusive linked-list previous ID
	ebLinkedListOffset = 0x08 // uint8  — which list this entity belongs to
	ebSpriteHeightNeg  = 0x09 // uint8  — negative sprite extent above position
	ebID               = 0x0A // uint16 — this entity's own slot index
	ebVehicleFlags     = 0x0C // uint16 — vehicle-specific bit flags
	ebPosX             = 0x0E // int16  — world X in sub-tile units (1 tile = 32 units)
	ebPosY             = 0x10 // int16  — world Y in sub-tile units
	ebPosZ             = 0x12 // int16  — world Z (1 unit = 1 px at zoom 0; = SmallZ*4)
	ebSpriteWidth      = 0x14 // uint8  — sprite bounding box width
	ebSpriteHeightPos  = 0x15 // uint8  — sprite bounding box height (positive)
	ebSpriteLeft       = 0x16 // int16  — sprite bounding box left edge
	ebSpriteTop        = 0x18 // int16  — sprite bounding box top edge
	ebSpriteRight      = 0x1A // int16  — sprite bounding box right edge
	ebSpriteBottom     = 0x1C // int16  — sprite bounding box bottom edge
	ebSpriteYaw        = 0x1E // uint8  — rotation 0-63 (64 steps = full circle)
	ebSpritePitch      = 0x1F // uint8  — pitch (used by aircraft/boats)
	ebOwner            = 0x21 // uint8  — company index that owns this entity
	ebName             = 0x22 // uint16 — display name (string table ID)
)

// Common vehicle field offsets — shared by all 6 VehicleType* subtypes.
// These come after EntityBase (0x00–0x23) in every vehicle struct.
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  VehicleHead / VehicleBody / …
const (
	vhCommonHead          = 0x26 // uint16 — ID of the VehicleHead entity for this train
	vhCommonRemainingDist = 0x28 // int32  — remaining movement distance (sub-tile units)
	vhCommonTrackAndDir   = 0x2C // uint16 — packed track ID + direction bits
	vhCommonSubPosition   = 0x2E // uint16 — sub-tile position within track segment
	vhCommonTileX         = 0x30 // int16  — tile-space X coordinate
	vhCommonTileY         = 0x32 // int16  — tile-space Y coordinate
	vhCommonTileBaseZ     = 0x34 // uint8  — tile-space height (SmallZ units)
	vhCommonTrackType     = 0x35 // uint8  — track object type index
	vhCommonRoutingHandle = 0x36 // uint16 — order routing handle
	vhCommonNextCarId     = 0x3A // uint16 — ID of next entity in the car chain (0xFFFF = end)
	vhCommonMode          = 0x42 // uint8  — driving mode (same in all vehicle types)
)

// VehicleBody / VehicleBogie specific offsets (type = VehicleTypeBogie, VehicleTypeBodyStart or VehicleTypeBodyContinued).
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  struct VehicleBody / VehicleBogie
const (
	vbColourPrimary     = 0x24 // uint8  — company primary colour
	vbColourSecondary   = 0x25 // uint8  — company secondary colour
	vbObjectSpriteType  = 0x39 // uint8  — body component index for sprite lookup
	vbObjectId          = 0x40 // uint16 — vehicle object slot index (→ ObjMgr.Vehicles[N])
	vbFlags38           = 0x38 // uint8  — Flags38 (bit 1 isReversed); bogie and body
	vbAnimation         = 0x46 // uint8  — bogie animationIndex / body animationFrame
	vbCargoFrame        = 0x47 // uint8  — body cargoFrame
	v2BrakeLightTimeout = 0x5B // uint8  — Vehicle2 brakeLightTimeout
	vbBodyIndex         = 0x54 // uint8  — body index within the vehicle's body list
)

// VehicleHead specific offsets (type = VehicleTypeHead).
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  struct VehicleHead
const (
	vHeadStatus      = 0x5D // uint8  — VehicleStatus* (unplaced/stopped/travelling/…)
	vHeadVehicleType = 0x5E // uint8  — VehicleType (train/road/air/water)
)

// VehicleEntity is a rendering-oriented, flattened representation of an S5
// entity slot.  Only the fields needed for world-space rendering are extracted.
//
// Raw byte offsets from S5Entity.h are recorded above as named constants so
// that the mapping between Go fields and the binary format can be verified
// against the OpenLoco source.
//
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h
type VehicleEntity struct {
	// Raw is the slot as read (nil for a component made in play): the save
	// writes the modelled fields over it (putVehicle).
	Raw []byte
	// ---- EntityBase ----
	ID          uint16 // ebID (0x0A) — slot index in the entity array
	EntityType  uint8  // ebType (0x01) — VehicleType* constant
	SpriteYaw   uint8  // ebSpriteYaw (0x1E) — 0-63
	SpritePitch uint8  // ebSpritePitch (0x1F) — Entity.h Pitch
	Owner       uint8  // ebOwner (0x21) — company index

	// Sub-tile world position (ebPosX/Y/Z, 0x0E-0x12).
	// 1 tile = 32 WorldUnits.  Call .Tile() for tile index, .SubOffset() for intra-tile.
	PosX coords.WorldUnits // world sub-tile X
	PosY coords.WorldUnits // world sub-tile Y
	PosZ int16             // world Z (1 unit = 1 px at zoom 0; = SmallZ*4)

	// ---- Common vehicle fields ----
	HeadID uint16 // vhCommonHead (0x26) — ID of this train's VehicleHead entity
	// Name (EntityBase 0x22: a StringId, combined with the head's ordinal
	// number — "Train 3") and OrdinalNumber (VehicleHead 0x44; other types
	// read other fields there).
	Name          uint16
	OrdinalNumber int16
	// SpriteWidth (0x14), SpriteHeightNegative (0x09), SpriteHeightPositive
	// (0x15): the sprite's extent (EntityBase).
	SpriteWidth, SpriteHeightNegative, SpriteHeightPositive uint8
	// TileX/TileY are tile-aligned world coords in WorldUnits (same units as PosX/PosY).
	// TileX = PosX.Tile().World() — i.e., floor(PosX/32)*32.
	// Use PosX.Tile() / PosY.Tile() for the tile grid index.
	TileX     coords.WorldUnits // vhCommonTileX (0x30) — tile-aligned X in WorldUnits
	TileY     coords.WorldUnits // vhCommonTileY (0x32) — tile-aligned Y in WorldUnits
	TileBaseZ uint8             // vhCommonTileBaseZ (0x34) — height in SmallZ units
	NextCarID uint16            // vhCommonNextCarId (0x3A) — next entity in car chain (0xFFFF = end)
	// VehicleBase motion state (updateTrackMotion / updateRoadMotion).
	RemainingDistance       int32       // vhCommonRemainingDist (0x28)
	TrackAndDirection       uint16      // vhCommonTrackAndDir (0x2C)
	SubPosition             uint16      // vhCommonSubPosition (0x2E)
	TrackType               uint8       // vhCommonTrackType (0x35) — track/road object id
	RoutingHandle           uint16      // vhCommonRoutingHandle (0x36): vehicleRef·64 | index
	Mode                    uint8       // vhCommonMode (0x42) — TransportMode (0 rail, 1 road, 2 air, 3 water)
	HeadVar53               uint8       // VehicleHead var_53 (0x53): the train's required track mods
	HeadVar52               uint8       // var_52 (0x52): VehicleHead (1 = placement mode: no lookahead) / Vehicle1 (one-way signal countdown)
	OrderTableOffset        uint32      // VehicleHead orderTableOffset (0x46): its orders in GameState orders[]
	CurrentOrder            uint16      // VehicleHead currentOrder (0x4A): offset from orderTableOffset
	SizeOfOrderTable        uint16      // VehicleHead sizeOfOrderTable (0x4C)
	Var48                   uint8       // Vehicle1 var_48 (0x48): Flags48 (passSignal 1, expressMode 2)
	LastIncome              IncomeStats // Vehicle1 lastIncome (0x53): the vehicle window's Finances tab
	Var49                   uint8       // Vehicle1 var_49 (0x49): the train's query mods
	Var5C                   uint8       // VehicleHead var_5C (0x5C): pass-signal countdown
	StationID               uint16      // VehicleHead stationId (0x54); the same bytes are Vehicle2's maxSpeed
	TrainAcceptedCargoTypes uint32      // VehicleHead trainAcceptedCargoTypes (0x4E): cargo slots its cars carry
	// Arrival / journey state (offsets from the compiler).
	BreakdownFlags       uint8  // 0x5F (head and components): awaitingCargoTransfer 1, journeyStarted 8
	CargoTransferTimeout uint16 // VehicleHead 0x56 (the same bytes are Vehicle2's currentSpeed)
	Var58                uint32 // VehicleHead var_58 (0x58)
	JourneyStartX        int16  // VehicleHead journeyStartPos (0x6F)
	JourneyStartY        int16
	JourneyStartTicks    uint32 // VehicleHead 0x73
	LastAverageSpeed     int16  // VehicleHead 0x77
	// Cargo is VehicleCargo at 0x48: VehicleBody primaryCargo / VehicleBogie
	// secondaryCargo (offsets from the compiler).
	Cargo              VehicleCargo
	CreationDay        uint32 // VehicleBogie creationDay (0x56; the same bytes are Vehicle2 currentSpeed)
	Reliability        uint16 // VehicleBogie 0x66 (front bogies)
	TimeAtSignal       uint16 // Vehicle1 timeAtSignal (0x46)
	DayCreated         uint32 // Vehicle1 dayCreated (0x4A)
	TimeoutToBreakdown uint16 // VehicleBogie 0x68 (front bogies, days)
	BreakdownTimeout   uint8  // VehicleBogie / VehicleBody 0x6A
	CurrentSpeed       int32  // Vehicle2 currentSpeed (0x56): Speed32, mph << 16
	Var44              uint16 // VehicleBogie var_44 (0x44): wheel roll accumulator (updateRoll)
	WheelSlipping      uint8  // VehicleBogie wheelSlipping (0x5E): ticks, counts up
	// Driving state (offsets from the compiler over S5Entity.h).
	VehicleFlags    uint16 // EntityBase vehicleFlags (0x0C): unk_0 = 1, manualControl = 1<<6
	ManualPower     int8   // VehicleHead manualPower (0x6E)
	TargetSpeed     int16  // Vehicle1 targetSpeed (0x44): Speed16, mph
	TotalPower      uint16 // Vehicle2 totalPower (0x50)
	TotalWeight     uint16 // Vehicle2 totalWeight (0x52); VehicleBogie totalCarWeight (0x52, first bogie of a car)
	MotorState      uint8  // Vehicle2 motorState (0x5A)
	Var73           uint8  // Vehicle2 var_73 (0x73): bit 0 broken down
	Veh2Reliability uint8  // Vehicle2 reliability (0x72), percent
	CrashedTimeout  uint8  // VehicleHead crashedTimeout (0x6D)
	// AirportMovementEdge is VehicleHead airportMovementEdge (0x68): the
	// movement edge of its airport an aircraft is on, 0xFF none.
	AirportMovementEdge uint8
	CurMonthRevenue     int32    // Vehicle2 curMonthRevenue (0x5E)
	Profit              [4]int32 // Vehicle2 profit[4] (0x62), last 4 months net
	RefundCost          uint32   // bogie / body refundCost (0x62)
	Var5A               uint32   // bogie / body var_5A (0x5A): the crash's speed, then its fall (bit 31: exploding)
	TotalRefundCost     uint32   // VehicleHead totalRefundCost (0x69)
	Var3C               int32    // VehicleHead / Vehicle1 var_3C (0x3C): distance budget the moves draw down
	Var4F               int8     // Vehicle2 var_4F (0x4F): the road object it is on (-1 none)
	MaxSpeed            int16    // Vehicle2 maxSpeed (0x54): Speed16
	RackRailMaxSpeed    int16    // Vehicle2 rackRailMaxSpeed (0x5C): Speed16
	// Vehicle2 / VehicleTail sound (VehicleSound, 0x44): the driving sound's
	// sound object slot (0xFF none), volume, frequency, and the vehicle object
	// it sounds for (0xFFFF none), soundFlags. soundWindowNumber and
	// soundWindowType (0x4C–0x4E) stay as read.
	DrivingSoundID        uint8
	DrivingSoundVolume    uint8
	DrivingSoundFrequency uint16
	SoundObjectID         uint16
	SoundFlags            uint16 // 0x4A
	// SoundHandle and ActiveSoundID are VehicleSound's audioHandle and
	// activeSoundId: the game's audio state, not saved (0: no handle).
	SoundHandle   uint32
	ActiveSoundID uint8

	// ---- VehicleBody / VehicleBogie sprite fields ----
	ObjectID         uint16 // vbObjectId (0x40) — vehicle object slot (ObjMgr.Vehicles[N])
	ObjectSpriteType uint8  // vbObjectSpriteType (0x39) — body component index for GetBodySpriteID
	ColourPrimary    uint8  // vbColourPrimary (0x24) — company primary colour index
	ColourSecondary  uint8  // vbColourSecondary (0x25) — company secondary colour index

	// ---- VehicleHead status (only meaningful when EntityType == VehicleTypeHead) ----
	Status      uint8 // vHeadStatus (0x5D) — VehicleStatus* constant
	VehicleType uint8 // vHeadVehicleType (0x5E) — VehicleType (train, bus, truck, tram, aircraft, ship)

	// Drawing state (S5Entity.h, offsets from the compiler): Flags38
	// (isReversed = bit 1), the bogie's animationIndex / body's animationFrame,
	// the body's cargoFrame, and veh2's brakeLightTimeout.
	Flags38           uint8
	Animation         uint8
	CargoFrame        uint8
	BrakeLightTimeout uint8
	BodyIndex         uint8
	ChuffSoundIndex   uint8 // VehicleBody chuffSoundIndex (0x55): the steam puff's chuff count
	// The entity's screen rectangle (EntityBase spriteLeft/Top/Right/Bottom),
	// for paintEntities' culling.
	SpriteLeft, SpriteTop, SpriteRight, SpriteBottom int16
}

// ParseVehicleEntities extracts VehicleHead and VehicleBody entities from a
// decompressed GameState chunk (as read from an .SV5 saved-game file).
//
// TileX and TileY are stored in world units (as PosX/PosY), not tile
// indices: the tile is TileX/32, TileY/32.
//
// Returns nil without error when the chunk is too small to contain the entity
// table — this is expected for .SC5 scenario files, which only write a
// GeneralState chunk (0xB96C bytes) rather than the full GameState.
//
// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
func ParseVehicleEntities(gameState []byte) []VehicleEntity {
	tableOffset := EntityTableOffset

	minSize := tableOffset + MaxEntities*EntitySize
	if len(gameState) < minSize {
		// Not a full saved-game GameState — scenario files are much smaller.
		return nil
	}

	var result []VehicleEntity
	for i := 0; i < MaxEntities; i++ {
		off := tableOffset + i*EntitySize
		e := gameState[off : off+EntitySize]

		baseType := e[ebBaseType]
		if baseType != EntityBaseTypeVehicle {
			// 0xFF = free slot; other non-zero values = particle/effect entity.
			continue
		}

		entityType := e[ebType]
		// Every vehicle entity is kept: the head (status), bogies and bodies
		// (sprites), veh1/veh2/tail for walking a train's chain (NextCarID).
		// The bogie/body fields read garbage for the other types.

		ent := VehicleEntity{
			ID:                uint16(e[ebID]) | uint16(e[ebID+1])<<8,
			EntityType:        entityType,
			SpriteYaw:         e[ebSpriteYaw],
			SpritePitch:       e[ebSpritePitch],
			Flags38:           e[vbFlags38],
			Animation:         e[vbAnimation],
			CargoFrame:        e[vbCargoFrame],
			BrakeLightTimeout: e[v2BrakeLightTimeout],
			BodyIndex:         e[vbBodyIndex],
			ChuffSoundIndex:   e[vbBodyIndex+1],
			SpriteLeft:        int16(uint16(e[ebSpriteLeft]) | uint16(e[ebSpriteLeft+1])<<8),
			SpriteTop:         int16(uint16(e[ebSpriteTop]) | uint16(e[ebSpriteTop+1])<<8),
			SpriteRight:       int16(uint16(e[ebSpriteRight]) | uint16(e[ebSpriteRight+1])<<8),
			SpriteBottom:      int16(uint16(e[ebSpriteBottom]) | uint16(e[ebSpriteBottom+1])<<8),
			Owner:             e[ebOwner],

			PosX: coords.WorldUnits(int16(e[ebPosX]) | int16(e[ebPosX+1])<<8),
			PosY: coords.WorldUnits(int16(e[ebPosY]) | int16(e[ebPosY+1])<<8),
			PosZ: int16(e[ebPosZ]) | int16(e[ebPosZ+1])<<8,

			HeadID:    uint16(e[vhCommonHead]) | uint16(e[vhCommonHead+1])<<8,
			TileX:     coords.WorldUnits(int16(e[vhCommonTileX]) | int16(e[vhCommonTileX+1])<<8),
			TileY:     coords.WorldUnits(int16(e[vhCommonTileY]) | int16(e[vhCommonTileY+1])<<8),
			TileBaseZ: e[vhCommonTileBaseZ],
			NextCarID: uint16(e[vhCommonNextCarId]) | uint16(e[vhCommonNextCarId+1])<<8,

			RemainingDistance:       int32(binary.LittleEndian.Uint32(e[vhCommonRemainingDist:])),
			TrackAndDirection:       binary.LittleEndian.Uint16(e[vhCommonTrackAndDir:]),
			SubPosition:             binary.LittleEndian.Uint16(e[vhCommonSubPosition:]),
			TrackType:               e[vhCommonTrackType],
			RoutingHandle:           binary.LittleEndian.Uint16(e[vhCommonRoutingHandle:]),
			Mode:                    e[vhCommonMode],
			HeadVar53:               e[0x53],
			HeadVar52:               e[0x52],
			OrderTableOffset:        binary.LittleEndian.Uint32(e[0x46:]),
			Name:                    binary.LittleEndian.Uint16(e[0x22:]),
			SpriteWidth:             e[0x14],
			SpriteHeightNegative:    e[0x09],
			SpriteHeightPositive:    e[0x15],
			OrdinalNumber:           int16(binary.LittleEndian.Uint16(e[0x44:])),
			CurrentOrder:            binary.LittleEndian.Uint16(e[0x4A:]),
			SizeOfOrderTable:        binary.LittleEndian.Uint16(e[0x4C:]),
			Var48:                   e[0x48],
			LastIncome:              parseIncomeStats(e[0x53:]),
			Var49:                   e[0x49],
			Var5C:                   e[0x5C],
			StationID:               binary.LittleEndian.Uint16(e[0x54:]),
			TrainAcceptedCargoTypes: binary.LittleEndian.Uint32(e[0x4E:]),
			BreakdownFlags:          e[0x5F],
			CargoTransferTimeout:    binary.LittleEndian.Uint16(e[0x56:]),
			Var58:                   binary.LittleEndian.Uint32(e[0x58:]),
			JourneyStartX:           int16(binary.LittleEndian.Uint16(e[0x6F:])),
			JourneyStartY:           int16(binary.LittleEndian.Uint16(e[0x71:])),
			JourneyStartTicks:       binary.LittleEndian.Uint32(e[0x73:]),
			LastAverageSpeed:        int16(binary.LittleEndian.Uint16(e[0x77:])),
			Cargo: VehicleCargo{AcceptedTypes: binary.LittleEndian.Uint32(e[0x48:]), Type: e[0x4C], MaxQty: e[0x4D],
				TownFrom: binary.LittleEndian.Uint16(e[0x4E:]), NumDays: e[0x50], Qty: e[0x51]},
			CreationDay: binary.LittleEndian.Uint32(e[0x56:]),
			Reliability: binary.LittleEndian.Uint16(e[0x66:]), TimeoutToBreakdown: binary.LittleEndian.Uint16(e[0x68:]), BreakdownTimeout: e[0x6A],
			TimeAtSignal:        binary.LittleEndian.Uint16(e[0x46:]),
			DayCreated:          binary.LittleEndian.Uint32(e[0x4A:]),
			CurrentSpeed:        int32(binary.LittleEndian.Uint32(e[0x56:])),
			Var44:               binary.LittleEndian.Uint16(e[0x44:]),
			WheelSlipping:       e[0x5E],
			VehicleFlags:        binary.LittleEndian.Uint16(e[0x0C:]),
			ManualPower:         int8(e[0x6E]),
			TargetSpeed:         int16(binary.LittleEndian.Uint16(e[0x44:])),
			TotalPower:          binary.LittleEndian.Uint16(e[0x50:]),
			TotalWeight:         binary.LittleEndian.Uint16(e[0x52:]),
			MotorState:          e[0x5A],
			Var73:               e[0x73],
			Veh2Reliability:     e[0x72],
			CrashedTimeout:      e[0x6D],
			AirportMovementEdge: e[0x68],
			CurMonthRevenue:     int32(binary.LittleEndian.Uint32(e[0x5E:])),
			Profit:              [4]int32{int32(binary.LittleEndian.Uint32(e[0x62:])), int32(binary.LittleEndian.Uint32(e[0x66:])), int32(binary.LittleEndian.Uint32(e[0x6A:])), int32(binary.LittleEndian.Uint32(e[0x6E:]))},
			RefundCost:          binary.LittleEndian.Uint32(e[0x62:]),
			Var5A:               binary.LittleEndian.Uint32(e[0x5A:]),
			TotalRefundCost:     binary.LittleEndian.Uint32(e[0x69:]),
			Var3C:               int32(binary.LittleEndian.Uint32(e[0x3C:])),
			Var4F:               int8(e[0x4F]),
			MaxSpeed:            int16(binary.LittleEndian.Uint16(e[0x54:])),
			RackRailMaxSpeed:    int16(binary.LittleEndian.Uint16(e[0x5C:])),
			DrivingSoundID:      e[0x44], DrivingSoundVolume: e[0x45],
			DrivingSoundFrequency: binary.LittleEndian.Uint16(e[0x46:]), SoundObjectID: binary.LittleEndian.Uint16(e[0x48:]),
			SoundFlags: binary.LittleEndian.Uint16(e[0x4A:]),

			Raw:              append([]byte(nil), e...),
			ObjectID:         uint16(e[vbObjectId]) | uint16(e[vbObjectId+1])<<8,
			ObjectSpriteType: e[vbObjectSpriteType],
			ColourPrimary:    e[vbColourPrimary],
			ColourSecondary:  e[vbColourSecondary],
		}

		// VehicleHead status lives at a type-specific offset that overlaps with
		// padding in VehicleBody — only read it for Head entities.
		if entityType == VehicleTypeHead {
			ent.Status = e[vHeadStatus]
			ent.VehicleType = e[vHeadVehicleType]
		}

		result = append(result, ent)
	}
	return result
}

// VehicleCargo is S5 VehicleCargo (0xA bytes): townFrom is the station the
// cargo was loaded at.
type VehicleCargo struct {
	AcceptedTypes uint32
	Type          uint8 // 0xFF: none
	MaxQty        uint8
	TownFrom      uint16
	NumDays       uint8
	Qty           uint8
}

// EntityLink is an entity slot's place in EntityManager's linked lists
// (S5Entity.h EntityBase: baseType 0x00, nextEntityId 0x04, llPreviousId 0x06,
// linkedListOffset 0x08 — the list index × 2): which list (free, vehicles, misc …)
// and its neighbours there. Their order decides which slot a new entity gets.
type EntityLink struct {
	BaseType   uint8
	ListOffset uint8
	Next, Prev uint16
}

// EntityLists is GameState's entityListHeads / entityListCounts (0x26 / 0x34)
// and every slot's EntityLink — only in a saved game's full GameState (nil
// otherwise: a .SC5 starts from EntityManager::reset).
type EntityLists struct {
	Heads, Counts [NumEntityLists]uint16
	Links         []EntityLink
}

// ParseEntityLists reads the entity lists from a saved game's GameState.
func ParseEntityLists(gameState []byte) *EntityLists {
	tableOffset := EntityTableOffset
	if len(gameState) < tableOffset+MaxEntities*EntitySize {
		return nil
	}
	u16 := func(b []byte, o int) uint16 { return uint16(b[o]) | uint16(b[o+1])<<8 }
	l := &EntityLists{Links: make([]EntityLink, MaxEntities)}
	for i := range NumEntityLists {
		l.Heads[i] = u16(gameState, GeneralEntityListHeadsOffset+2*i)
		l.Counts[i] = u16(gameState, GeneralEntityListCountsOffset+2*i)
	}
	for i := range MaxEntities {
		e := gameState[tableOffset+i*EntitySize:]
		l.Links[i] = EntityLink{BaseType: e[ebBaseType], ListOffset: e[ebLinkedListOffset], Next: u16(e, ebNextEntityId), Prev: u16(e, ebLLPreviousId)}
	}
	return l
}

// IncomeStats is Vehicle1's lastIncome (S5Entity.h, 0x2C bytes at 0x53): the
// day of the last delivery run and up to four cargo lines (type 0xFF: none).
type IncomeStats struct {
	Day            int32
	CargoTypes     [4]uint8
	CargoQtys      [4]uint16
	CargoDistances [4]uint16
	CargoAges      [4]uint8
	CargoProfits   [4]int32
}

func parseIncomeStats(b []byte) IncomeStats {
	s := IncomeStats{Day: int32(binary.LittleEndian.Uint32(b))}
	for i := 0; i < 4; i++ {
		s.CargoTypes[i] = b[0x04+i]
		s.CargoQtys[i] = binary.LittleEndian.Uint16(b[0x08+2*i:])
		s.CargoDistances[i] = binary.LittleEndian.Uint16(b[0x10+2*i:])
		s.CargoAges[i] = b[0x18+i]
		s.CargoProfits[i] = int32(binary.LittleEndian.Uint32(b[0x1C+4*i:]))
	}
	return s
}

// MiscEntity is an entity of the misc list (EntityBaseType::effect): the
// effects of S5Entity.h, every subtype's fields in one struct (each subtype
// reads its own). Raw is the slot as read (nil for one made in play).
type MiscEntity struct {
	ID                                            uint16
	SubType                                       uint8 // EffectType
	X, Y, Z                                       int16
	SpriteWidth, SpriteHeightNeg, SpriteHeightPos uint8
	Field26, Field28                              uint16 // exhaust frameNum / money frame·moveDelay / particle timeToLive; frame (numMovements, stationaryProgress)
	WindProgress                                  uint16 // exhaust 0x32
	ObjectID                                      uint8  // exhaust 0x49
	Amount                                        int32  // money 0x2A
	Var2E                                         uint8  // money company / particle colour primary
	Colour2                                       uint8  // particle colour secondary (0x2F)
	CrashedSpriteBase                             uint16 // particle 0x30
	Velocity                                      [3]int16
	Acceleration                                  [3]int32
	Wiggle                                        uint16 // money 0x46
	Raw                                           []byte
}

// Effect subtypes (EffectType).
const (
	EffectExhaust              = 0
	EffectRedGreenCurrency     = 1
	EffectWindowCurrency       = 2
	EffectVehicleCrashParticle = 3
	EffectExplosionCloud       = 4
	EffectSplash               = 5
	EffectFireball             = 6
	EffectExplosionSmoke       = 7
	EffectSmoke                = 8
)

// ParseMiscEntities reads every effect entity of a saved game.
func ParseMiscEntities(gs []byte) []MiscEntity {
	if len(gs) < EntityTableOffset+MaxEntities*EntitySize {
		return nil
	}
	var out []MiscEntity
	for i := range MaxEntities {
		e := gs[EntityTableOffset+i*EntitySize : EntityTableOffset+(i+1)*EntitySize]
		if e[ebBaseType] != EntityBaseTypeEffect {
			continue
		}
		u16 := func(o int) uint16 { return binary.LittleEndian.Uint16(e[o:]) }
		m := MiscEntity{ID: u16(ebID), SubType: e[ebType],
			X: int16(u16(ebPosX)), Y: int16(u16(ebPosY)), Z: int16(u16(ebPosZ)),
			SpriteWidth: e[ebSpriteWidth], SpriteHeightNeg: e[ebSpriteHeightNeg], SpriteHeightPos: e[ebSpriteHeightPos],
			Field26: u16(0x26), Field28: u16(0x28), WindProgress: u16(0x32), ObjectID: e[0x49],
			Amount: int32(binary.LittleEndian.Uint32(e[0x2A:])), Var2E: e[0x2E], Colour2: e[0x2F],
			CrashedSpriteBase: u16(0x30), Wiggle: u16(0x46),
			Raw: append([]byte(nil), e...)}
		for k := range 3 {
			m.Velocity[k] = int16(u16(0x32 + 2*k))
			m.Acceleration[k] = int32(binary.LittleEndian.Uint32(e[0x38+4*k:]))
		}
		out = append(out, m)
	}
	return out
}

// putEntityBase is exportEntityBase: the base fields (the list links are the
// caller's), the sprite bounds Location::null.
func putEntityBase(e []byte, typ uint8, id uint16, x, y, z int16, w, hNeg, hPos uint8) {
	e[ebType] = typ
	e[ebSpriteHeightNeg] = hNeg
	put16(e, ebID, id)
	put16(e, ebPosX, uint16(x))
	put16(e, ebPosY, uint16(y))
	put16(e, ebPosZ, uint16(z))
	e[ebSpriteWidth], e[ebSpriteHeightPos] = w, hPos
	for _, o := range []int{ebSpriteLeft, ebSpriteTop, ebSpriteRight, ebSpriteBottom} {
		put16(e, o, 0x8000)
	}
}

// putMiscEntity is exportEntity for an effect, by subtype.
func putMiscEntity(e []byte, m *MiscEntity) {
	putEntityBase(e, m.SubType, m.ID, m.X, m.Y, m.Z, m.SpriteWidth, m.SpriteHeightNeg, m.SpriteHeightPos)
	switch m.SubType {
	case EffectExhaust:
		put16(e, 0x26, m.Field26)
		put16(e, 0x28, m.Field28)
		put16(e, 0x32, m.WindProgress)
		e[0x49] = m.ObjectID
	case EffectRedGreenCurrency, EffectWindowCurrency:
		put16(e, 0x26, m.Field26)
		put16(e, 0x28, m.Field28)
		put32(e, 0x2A, uint32(m.Amount))
		e[0x2E] = m.Var2E
		put16(e, 0x46, m.Wiggle)
	case EffectVehicleCrashParticle:
		put16(e, 0x26, m.Field26)
		put16(e, 0x28, m.Field28)
		e[0x2E], e[0x2F] = m.Var2E, m.Colour2
		put16(e, 0x30, m.CrashedSpriteBase)
		for k := range 3 {
			put16(e, 0x32+2*k, uint16(m.Velocity[k]))
			put32(e, 0x38+4*k, uint32(m.Acceleration[k]))
		}
	default: // the cloud, splash, fireball, smokes: their frame
		put16(e, 0x28, m.Field28)
	}
}

// putVehicle is exportEntity for a vehicle component: the base and the
// fields every vehicle has, then its type's own (S5Entity.h's VehicleHead …
// VehicleTail: one offset means another field in each).
// putVehicleSound writes Vehicle2's / VehicleTail's VehicleSound (0x44–0x4B).
func putVehicleSound(e []byte, v *VehicleEntity) {
	e[0x44], e[0x45] = v.DrivingSoundID, v.DrivingSoundVolume
	put16(e, 0x46, v.DrivingSoundFrequency)
	put16(e, 0x48, v.SoundObjectID)
	put16(e, 0x4A, v.SoundFlags)
}

func putVehicle(e []byte, v *VehicleEntity) {
	putEntityBase(e, v.EntityType, v.ID, int16(v.PosX), int16(v.PosY), v.PosZ, v.SpriteWidth, v.SpriteHeightNegative, v.SpriteHeightPositive)
	put16(e, ebVehicleFlags, v.VehicleFlags)
	e[ebSpriteYaw], e[ebSpritePitch], e[ebOwner] = v.SpriteYaw, v.SpritePitch, v.Owner
	put16(e, ebName, v.Name)
	e[vbColourPrimary], e[vbColourSecondary] = v.ColourPrimary, v.ColourSecondary
	put16(e, vhCommonHead, v.HeadID)
	put32(e, vhCommonRemainingDist, uint32(v.RemainingDistance))
	put16(e, vhCommonTrackAndDir, v.TrackAndDirection)
	put16(e, vhCommonSubPosition, v.SubPosition)
	put16(e, vhCommonTileX, uint16(v.TileX))
	put16(e, vhCommonTileY, uint16(v.TileY))
	e[vhCommonTileBaseZ], e[vhCommonTrackType] = v.TileBaseZ, v.TrackType
	put16(e, vhCommonRoutingHandle, v.RoutingHandle)
	e[vbFlags38], e[vbObjectSpriteType] = v.Flags38, v.ObjectSpriteType
	put16(e, vhCommonNextCarId, v.NextCarID)
	put32(e, 0x3C, uint32(v.Var3C))
	put16(e, vbObjectId, v.ObjectID)
	e[vhCommonMode] = v.Mode
	switch v.EntityType {
	case VehicleTypeHead:
		put16(e, 0x44, uint16(v.OrdinalNumber))
		put32(e, 0x46, v.OrderTableOffset)
		put16(e, 0x4A, v.CurrentOrder)
		put16(e, 0x4C, v.SizeOfOrderTable)
		put32(e, 0x4E, v.TrainAcceptedCargoTypes)
		e[0x52], e[0x53] = v.HeadVar52, v.HeadVar53
		put16(e, 0x54, v.StationID)
		put16(e, 0x56, v.CargoTransferTimeout)
		put32(e, 0x58, v.Var58)
		e[0x5C], e[vHeadStatus], e[vHeadVehicleType], e[0x5F] = v.Var5C, v.Status, v.VehicleType, v.BreakdownFlags
		e[0x68] = v.AirportMovementEdge
		put32(e, 0x69, v.TotalRefundCost)
		e[0x6D], e[0x6E] = v.CrashedTimeout, uint8(v.ManualPower)
		put16(e, 0x6F, uint16(v.JourneyStartX))
		put16(e, 0x71, uint16(v.JourneyStartY))
		put32(e, 0x73, v.JourneyStartTicks)
		put16(e, 0x77, uint16(v.LastAverageSpeed))
	case VehicleTypeVeh1:
		put16(e, 0x44, uint16(v.TargetSpeed))
		put16(e, 0x46, v.TimeAtSignal)
		e[0x48], e[0x49] = v.Var48, v.Var49
		put32(e, 0x4A, v.DayCreated)
		put32(e, 0x4E, v.TrainAcceptedCargoTypes) // var_4E, var_50: a boat's sub-pixel position (sim watermotion.go)
		e[0x52] = v.HeadVar52
		s := v.LastIncome
		put32(e, 0x53, uint32(s.Day))
		for k := range 4 {
			e[0x57+k] = s.CargoTypes[k]
			put16(e, 0x5B+2*k, s.CargoQtys[k])
			put16(e, 0x63+2*k, s.CargoDistances[k])
			e[0x6B+k] = s.CargoAges[k]
			put32(e, 0x6F+4*k, uint32(s.CargoProfits[k]))
		}
	case VehicleTypeVeh2:
		putVehicleSound(e, v)
		e[0x4F] = uint8(v.Var4F)
		put16(e, 0x50, v.TotalPower)
		put16(e, 0x52, v.TotalWeight)
		put16(e, 0x54, uint16(v.MaxSpeed))
		put32(e, 0x56, uint32(v.CurrentSpeed))
		e[0x5A], e[v2BrakeLightTimeout] = v.MotorState, v.BrakeLightTimeout
		put16(e, 0x5C, uint16(v.RackRailMaxSpeed))
		put32(e, 0x5E, uint32(v.CurMonthRevenue))
		for k, p := range v.Profit {
			put32(e, 0x62+4*k, uint32(p))
		}
		e[0x72], e[0x73] = v.Veh2Reliability, v.Var73
	case VehicleTypeTail:
		putVehicleSound(e, v)
	case VehicleTypeBogie, VehicleTypeBodyStart, VehicleTypeBodyContinued:
		put16(e, 0x44, v.Var44)
		e[vbAnimation], e[vbCargoFrame] = v.Animation, v.CargoFrame
		c := v.Cargo
		put32(e, 0x48, c.AcceptedTypes)
		e[0x4C], e[0x4D] = c.Type, c.MaxQty
		put16(e, 0x4E, c.TownFrom)
		e[0x50], e[0x51] = c.NumDays, c.Qty
		e[vbBodyIndex], e[vbBodyIndex+1] = v.BodyIndex, v.ChuffSoundIndex
		put32(e, 0x56, v.CreationDay)
		put32(e, 0x5A, v.Var5A)
		e[0x5E], e[0x5F] = v.WheelSlipping, v.BreakdownFlags
		put32(e, 0x62, v.RefundCost)
		e[0x6A] = v.BreakdownTimeout
		if v.EntityType == VehicleTypeBogie {
			put16(e, 0x52, v.TotalWeight)
			put16(e, 0x66, v.Reliability)
			put16(e, 0x68, v.TimeoutToBreakdown)
		}
	}
}

// initCreatedVehicle: the fields of a component made in play that goloco
// doesn't model, as CreateVehicle.cpp sets them (createHead: aiThoughtId
// 0xFF, aiPlacementPos.x −1; createVehicle2 / createVehicleTail: the sound's
// drivingSoundId null and objectId 0xFFFF).
func initCreatedVehicle(e []byte, typ uint8) {
	switch typ {
	case VehicleTypeHead:
		e[0x60] = 0xFF
		put16(e, 0x61, 0xFFFF)
	case VehicleTypeVeh2, VehicleTypeTail:
		e[0x44] = 0xFF
		put16(e, 0x48, 0xFFFF)
	}
}

func (sc *Scenario) putEntities(gs []byte) {
	l := sc.EntityLists
	if l == nil {
		return // no entities modelled: the table as read
	}
	for i := range NumEntityLists {
		put16(gs, GeneralEntityListHeadsOffset+2*i, l.Heads[i])
		put16(gs, GeneralEntityListCountsOffset+2*i, l.Counts[i])
	}
	vehicles := make(map[uint16]*VehicleEntity, len(sc.Entities))
	for i := range sc.Entities {
		vehicles[sc.Entities[i].ID] = &sc.Entities[i]
	}
	misc := make(map[uint16]*MiscEntity, len(sc.MiscEntities))
	for i := range sc.MiscEntities {
		misc[sc.MiscEntities[i].ID] = &sc.MiscEntities[i]
	}
	for id := range MaxEntities {
		e := gs[EntityTableOffset+id*EntitySize : EntityTableOffset+(id+1)*EntitySize]
		ln := l.Links[id]
		switch {
		case ln.BaseType == EntityBaseTypeVehicle && vehicles[uint16(id)] != nil:
			v := vehicles[uint16(id)]
			if v.Raw != nil {
				copy(e, v.Raw)
			} else {
				clear(e) // createEntity: zeroEntity
				initCreatedVehicle(e, v.EntityType)
			}
			putVehicle(e, v)
		case ln.BaseType == EntityBaseTypeEffect && misc[uint16(id)] != nil:
			m := misc[uint16(id)]
			if m.Raw != nil {
				copy(e, m.Raw)
			} else {
				clear(e)
			}
			putMiscEntity(e, m)
		case ln.BaseType == EntityBaseTypeFree:
			clear(e) // EntityManager::zeroUnused (zeroEntity)
			put16(e, ebID, uint16(id))
		}
		// else: an entity goloco doesn't model — its slot as read.
		e[ebBaseType] = ln.BaseType
		put16(e, ebNextEntityId, ln.Next)
		put16(e, ebLLPreviousId, ln.Prev)
		e[ebLinkedListOffset] = ln.ListOffset
	}
}
