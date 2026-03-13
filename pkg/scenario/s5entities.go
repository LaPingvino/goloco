package scenario

import "github.com/LaPingvino/goloco/pkg/coords"

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
	VehicleTypeHead  = uint8(0) // VehicleHead  — lead car; has status, orders, route
	VehicleTypeVeh1  = uint8(1) // Vehicle1     — front physics / income-tracking car
	VehicleTypeVeh2  = uint8(2) // Vehicle2     — speed / power / weight / sound car
	VehicleTypeBody  = uint8(3) // VehicleBody  — visible cargo/passenger body (has sprites)
	VehicleTypeBogie = uint8(4) // VehicleBogie — undercarriage / wheels (has sprites)
	VehicleTypeTail  = uint8(5) // VehicleTail  — rear car; sound output
)

// VehicleHead status values — VehicleHead byte 0x5D.
// OpenLoco reference: src/OpenLoco/src/Vehicles/Vehicle.h  Status
const (
	VehicleStatusUnplaced   = uint8(0) // not yet placed on the map
	VehicleStatusStopped    = uint8(1) // parked at a station
	VehicleStatusTravelling = uint8(2) // moving along track
	VehicleStatusWaiting    = uint8(3) // waiting at a signal
	VehicleStatusLoading    = uint8(4) // loading/unloading cargo
	VehicleStatusApproaching = uint8(5) // approaching station
	VehicleStatusBrokenDown = uint8(6)
	VehicleStatusCrashed    = uint8(7)
)

// EntityBase byte offsets — sizeof = 0x24.
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  struct EntityBase
const (
	ebBaseType            = 0x00 // uint8  — EntityBaseTypeVehicle / EntityBaseTypeFree
	ebType                = 0x01 // uint8  — VehicleType* or misc entity subtype
	ebNextQuadrantId      = 0x02 // uint16 — spatial hash next-entity for quadrant
	ebNextEntityId        = 0x04 // uint16 — intrusive linked-list next ID
	ebLLPreviousId        = 0x06 // uint16 — intrusive linked-list previous ID
	ebLinkedListOffset    = 0x08 // uint8  — which list this entity belongs to
	ebSpriteHeightNeg     = 0x09 // uint8  — negative sprite extent above position
	ebID                  = 0x0A // uint16 — this entity's own slot index
	ebVehicleFlags        = 0x0C // uint16 — vehicle-specific bit flags
	ebPosX                = 0x0E // int16  — world X in sub-tile units (1 tile = 32 units)
	ebPosY                = 0x10 // int16  — world Y in sub-tile units
	ebPosZ                = 0x12 // uint8  — world Z (SmallZ height units)
	ebSpriteWidth         = 0x14 // uint8  — sprite bounding box width
	ebSpriteHeightPos     = 0x15 // uint8  — sprite bounding box height (positive)
	ebSpriteLeft          = 0x16 // int16  — sprite bounding box left edge
	ebSpriteTop           = 0x18 // int16  — sprite bounding box top edge
	ebSpriteRight         = 0x1A // int16  — sprite bounding box right edge
	ebSpriteBottom        = 0x1C // int16  — sprite bounding box bottom edge
	ebSpriteYaw           = 0x1E // uint8  — rotation 0-63 (64 steps = full circle)
	ebSpritePitch         = 0x1F // uint8  — pitch (used by aircraft/boats)
	ebOwner               = 0x21 // uint8  — company index that owns this entity
	ebName                = 0x22 // uint16 — display name (string table ID)
)

// Common vehicle field offsets — shared by all 6 VehicleType* subtypes.
// These come after EntityBase (0x00–0x23) in every vehicle struct.
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  VehicleHead / VehicleBody / …
const (
	vhCommonHead             = 0x26 // uint16 — ID of the VehicleHead entity for this train
	vhCommonRemainingDist    = 0x28 // int32  — remaining movement distance (sub-tile units)
	vhCommonTrackAndDir      = 0x2C // uint16 — packed track ID + direction bits
	vhCommonSubPosition      = 0x2E // uint16 — sub-tile position within track segment
	vhCommonTileX            = 0x30 // int16  — tile-space X coordinate
	vhCommonTileY            = 0x32 // int16  — tile-space Y coordinate
	vhCommonTileBaseZ        = 0x34 // uint8  — tile-space height (SmallZ units)
	vhCommonTrackType        = 0x35 // uint8  — track object type index
	vhCommonRoutingHandle    = 0x36 // uint16 — order routing handle
	vhCommonNextCarId        = 0x3A // uint16 — ID of next entity in the car chain (0xFFFF = end)
	vhCommonMode             = 0x42 // uint8  — driving mode (same in all vehicle types)
)

// VehicleBody / VehicleBogie specific offsets (type = VehicleTypeBody or VehicleTypeBogie).
// OpenLoco reference: src/OpenLoco/src/S5/S5Entity.h  struct VehicleBody / VehicleBogie
const (
	vbColourPrimary   = 0x24 // uint8  — company primary colour
	vbColourSecondary = 0x25 // uint8  — company secondary colour
	vbObjectSpriteType = 0x39 // uint8  — body component index for sprite lookup
	vbObjectId        = 0x40 // uint16 — vehicle object slot index (→ ObjMgr.Vehicles[N])
	vbBodyIndex       = 0x54 // uint8  — body index within the vehicle's body list
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
	ID         uint16 // ebID (0x0A) — slot index in the entity array
	EntityType uint8  // ebType (0x01) — VehicleType* constant
	SpriteYaw  uint8  // ebSpriteYaw (0x1E) — 0-63; halve for 32-direction sprite index
	Owner      uint8  // ebOwner (0x21) — company index

	// Sub-tile world position (ebPosX/Y/Z, 0x0E-0x12).
	// 1 tile = 32 WorldUnits.  Call .Tile() for tile index, .SubOffset() for intra-tile.
	PosX coords.WorldUnits // world sub-tile X
	PosY coords.WorldUnits // world sub-tile Y
	PosZ uint8             // world Z (SmallZ height; multiply by 4 for pixels)

	// ---- Common vehicle fields ----
	HeadID uint16 // vhCommonHead (0x26) — ID of this train's VehicleHead entity
	// TileX/TileY are tile-aligned world coords in WorldUnits (same units as PosX/PosY).
	// TileX = PosX.Tile().World() — i.e., floor(PosX/32)*32.
	// Use PosX.Tile() / PosY.Tile() for the tile grid index.
	TileX     coords.WorldUnits // vhCommonTileX (0x30) — tile-aligned X in WorldUnits
	TileY     coords.WorldUnits // vhCommonTileY (0x32) — tile-aligned Y in WorldUnits
	TileBaseZ uint8             // vhCommonTileBaseZ (0x34) — height in SmallZ units
	NextCarID uint16 // vhCommonNextCarId (0x3A) — next entity in car chain (0xFFFF = end)

	// ---- VehicleBody / VehicleBogie sprite fields ----
	ObjectID         uint16 // vbObjectId (0x40) — vehicle object slot (ObjMgr.Vehicles[N])
	ObjectSpriteType uint8  // vbObjectSpriteType (0x39) — body component index for GetBodySpriteID
	ColourPrimary    uint8  // vbColourPrimary (0x24) — company primary colour index
	ColourSecondary  uint8  // vbColourSecondary (0x25) — company secondary colour index

	// ---- VehicleHead status (only meaningful when EntityType == VehicleTypeHead) ----
	Status uint8 // vHeadStatus (0x5D) — VehicleStatus* constant
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
		// We only extract Head (for status checks) and Body (for sprites).
		// Vehicle1/2/Bogie/Tail entities carry no information we need here.
		if entityType != VehicleTypeHead && entityType != VehicleTypeBody {
			continue
		}

		ent := VehicleEntity{
			ID:         uint16(e[ebID]) | uint16(e[ebID+1])<<8,
			EntityType: entityType,
			SpriteYaw:  e[ebSpriteYaw],
			Owner:      e[ebOwner],

			PosX: coords.WorldUnits(int16(e[ebPosX]) | int16(e[ebPosX+1])<<8),
			PosY: coords.WorldUnits(int16(e[ebPosY]) | int16(e[ebPosY+1])<<8),
			PosZ: e[ebPosZ],

			HeadID:    uint16(e[vhCommonHead]) | uint16(e[vhCommonHead+1])<<8,
			TileX:     coords.WorldUnits(int16(e[vhCommonTileX]) | int16(e[vhCommonTileX+1])<<8),
			TileY:     coords.WorldUnits(int16(e[vhCommonTileY]) | int16(e[vhCommonTileY+1])<<8),
			TileBaseZ: e[vhCommonTileBaseZ],
			NextCarID: uint16(e[vhCommonNextCarId]) | uint16(e[vhCommonNextCarId+1])<<8,

			ObjectID:         uint16(e[vbObjectId]) | uint16(e[vbObjectId+1])<<8,
			ObjectSpriteType: e[vbObjectSpriteType],
			ColourPrimary:    e[vbColourPrimary],
			ColourSecondary:  e[vbColourSecondary],
		}

		// VehicleHead status lives at a type-specific offset that overlaps with
		// padding in VehicleBody — only read it for Head entities.
		if entityType == VehicleTypeHead {
			ent.Status = e[vHeadStatus]
		}

		result = append(result, ent)
	}
	return result
}

