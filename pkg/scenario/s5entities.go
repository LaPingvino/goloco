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

	// EntityTableOffsetType2 is the entity table offset for GameStateType2.
	// GameStateType2 uses a smaller Company struct, shifting all subsequent
	// arrays down by 0x1C20 bytes relative to GameState.
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h
	//   struct GameStateType2 { Entity entities[...]; // 0x1B3CA4 }
	//   static_assert(sizeof(GameStateType2) == 0x49EA24)
	EntityTableOffsetType2 = 0x1B3CA4

	// GameStateSizeType2 is the decompressed size of a GameStateType2 chunk.
	// When the GameState chunk decompresses to exactly this size, use
	// EntityTableOffsetType2 instead of EntityTableOffset.
	GameStateSizeType2 = 0x49EA24

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
	// TileX/TileY are tile-aligned world coords in WorldUnits (same units as PosX/PosY).
	// TileX = PosX.Tile().World() — i.e., floor(PosX/32)*32.
	// Use PosX.Tile() / PosY.Tile() for the tile grid index.
	TileX     coords.WorldUnits // vhCommonTileX (0x30) — tile-aligned X in WorldUnits
	TileY     coords.WorldUnits // vhCommonTileY (0x32) — tile-aligned Y in WorldUnits
	TileBaseZ uint8             // vhCommonTileBaseZ (0x34) — height in SmallZ units
	NextCarID uint16            // vhCommonNextCarId (0x3A) — next entity in car chain (0xFFFF = end)
	// VehicleBase motion state (updateTrackMotion / updateRoadMotion).
	RemainingDistance       int32  // vhCommonRemainingDist (0x28)
	TrackAndDirection       uint16 // vhCommonTrackAndDir (0x2C)
	SubPosition             uint16 // vhCommonSubPosition (0x2E)
	TrackType               uint8  // vhCommonTrackType (0x35) — track/road object id
	RoutingHandle           uint16 // vhCommonRoutingHandle (0x36): vehicleRef·64 | index
	Mode                    uint8  // vhCommonMode (0x42) — TransportMode (0 rail, 1 road, 2 air, 3 water)
	HeadVar53               uint8  // VehicleHead var_53 (0x53): the train's required track mods
	HeadVar52               uint8  // var_52 (0x52): VehicleHead (1 = placement mode: no lookahead) / Vehicle1 (one-way signal countdown)
	OrderTableOffset        uint32 // VehicleHead orderTableOffset (0x46): its orders in GameState orders[]
	CurrentOrder            uint16 // VehicleHead currentOrder (0x4A): offset from orderTableOffset
	SizeOfOrderTable        uint16 // VehicleHead sizeOfOrderTable (0x4C)
	Var48                   uint8  // Vehicle1 var_48 (0x48): Flags48 (passSignal 1, expressMode 2)
	Var49                   uint8  // Vehicle1 var_49 (0x49): the train's query mods
	Var5C                   uint8  // VehicleHead var_5C (0x5C): pass-signal countdown
	StationID               uint16 // VehicleHead stationId (0x54); the same bytes are Vehicle2's maxSpeed
	TrainAcceptedCargoTypes uint32 // VehicleHead trainAcceptedCargoTypes (0x4E): cargo slots its cars carry
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
	TimeoutToBreakdown uint16 // VehicleBogie 0x68 (front bogies, days)
	BreakdownTimeout   uint8  // VehicleBogie / VehicleBody 0x6A
	CurrentSpeed       int32  // Vehicle2 currentSpeed (0x56): Speed32, mph << 16
	Var44              uint16 // VehicleBogie var_44 (0x44): wheel roll accumulator (updateRoll)
	WheelSlipping      uint8  // VehicleBogie wheelSlipping (0x5E): ticks, counts up
	// Driving state (offsets from the compiler over S5Entity.h).
	VehicleFlags     uint16   // EntityBase vehicleFlags (0x0C): unk_0 = 1, manualControl = 1<<6
	ManualPower      int8     // VehicleHead manualPower (0x6E)
	TargetSpeed      int16    // Vehicle1 targetSpeed (0x44): Speed16, mph
	TotalPower       uint16   // Vehicle2 totalPower (0x50)
	TotalWeight      uint16   // Vehicle2 totalWeight (0x52); VehicleBogie totalCarWeight (0x52, first bogie of a car)
	MotorState       uint8    // Vehicle2 motorState (0x5A)
	Var73            uint8    // Vehicle2 var_73 (0x73): bit 0 broken down
	Veh2Reliability  uint8    // Vehicle2 reliability (0x72), percent
	CrashedTimeout   uint8    // VehicleHead crashedTimeout (0x6D)
	CurMonthRevenue  int32    // Vehicle2 curMonthRevenue (0x5E)
	Profit           [4]int32 // Vehicle2 profit[4] (0x62), last 4 months net
	RefundCost       uint32   // bogie / body refundCost (0x62)
	TotalRefundCost  uint32   // VehicleHead totalRefundCost (0x69)
	Var3C            int32    // VehicleHead / Vehicle1 var_3C (0x3C): distance budget the moves draw down
	Var4F            int8     // Vehicle2 var_4F (0x4F): the road object it is on (-1 none)
	MaxSpeed         int16    // Vehicle2 maxSpeed (0x54): Speed16
	RackRailMaxSpeed int16    // Vehicle2 rackRailMaxSpeed (0x5C): Speed16

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
	// The entity's screen rectangle (EntityBase spriteLeft/Top/Right/Bottom),
	// for paintEntities' culling.
	SpriteLeft, SpriteTop, SpriteRight, SpriteBottom int16
}

// ParseVehicleEntities extracts VehicleHead and VehicleBody entities from a
// decompressed GameState chunk (as read from an .SV5 saved-game file).
//
// Detects GameStateType2 vs GameState by decompressed chunk size and chooses
// the correct entity table offset automatically.
//
// IMPORTANT: In GameStateType2 files (including title.dat), TileX and TileY
// are stored in sub-tile world units (same as PosX/PosY), NOT tile indices.
// Convert to tile index with TileX/32 or TileY/32.
//
// Returns nil without error when the chunk is too small to contain the entity
// table — this is expected for .SC5 scenario files, which only write a
// GeneralState chunk (0xB96C bytes) rather than the full GameState.
//
// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
func ParseVehicleEntities(gameState []byte) []VehicleEntity {
	// Choose the entity table offset based on the GameState variant.
	// GameStateType2 (sizeof==0x49EA24) uses offset 0x1B3CA4.
	// GameState       (sizeof==0x4A0644) uses offset 0x1B58C4.
	tableOffset := EntityTableOffset
	if len(gameState) == GameStateSizeType2 {
		tableOffset = EntityTableOffsetType2
	}

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
			CurrentOrder:            binary.LittleEndian.Uint16(e[0x4A:]),
			SizeOfOrderTable:        binary.LittleEndian.Uint16(e[0x4C:]),
			Var48:                   e[0x48],
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
			TimeAtSignal:     binary.LittleEndian.Uint16(e[0x46:]),
			CurrentSpeed:     int32(binary.LittleEndian.Uint32(e[0x56:])),
			Var44:            binary.LittleEndian.Uint16(e[0x44:]),
			WheelSlipping:    e[0x5E],
			VehicleFlags:     binary.LittleEndian.Uint16(e[0x0C:]),
			ManualPower:      int8(e[0x6E]),
			TargetSpeed:      int16(binary.LittleEndian.Uint16(e[0x44:])),
			TotalPower:       binary.LittleEndian.Uint16(e[0x50:]),
			TotalWeight:      binary.LittleEndian.Uint16(e[0x52:]),
			MotorState:       e[0x5A],
			Var73:            e[0x73],
			Veh2Reliability:  e[0x72],
			CrashedTimeout:   e[0x6D],
			CurMonthRevenue:  int32(binary.LittleEndian.Uint32(e[0x5E:])),
			Profit:           [4]int32{int32(binary.LittleEndian.Uint32(e[0x62:])), int32(binary.LittleEndian.Uint32(e[0x66:])), int32(binary.LittleEndian.Uint32(e[0x6A:])), int32(binary.LittleEndian.Uint32(e[0x6E:]))},
			RefundCost:       binary.LittleEndian.Uint32(e[0x62:]),
			TotalRefundCost:  binary.LittleEndian.Uint32(e[0x69:]),
			Var3C:            int32(binary.LittleEndian.Uint32(e[0x3C:])),
			Var4F:            int8(e[0x4F]),
			MaxSpeed:         int16(binary.LittleEndian.Uint16(e[0x54:])),
			RackRailMaxSpeed: int16(binary.LittleEndian.Uint16(e[0x5C:])),

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
	if len(gameState) == GameStateSizeType2 {
		tableOffset = EntityTableOffsetType2
	}
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
