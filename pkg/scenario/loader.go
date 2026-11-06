package scenario

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// DefaultMapSize is the standard Locomotion map dimension (384×384).
const DefaultMapSize = 384

// tileElementSize is the fixed size of every TileElement on disk (8 bytes).
// OpenLoco reference: src/OpenLoco/src/Map/TileElement.h
const tileElementSize = 8

// Surface element byte-4 layout (the type-specific portion):
//   [4] slope          bits [3:0] = slope corners, bit [4] = double height, bits [7:5] = snow
//   [5] water          bits [4:0] = water level, bits [7:5] = update timer
//   [6] terrain        bits [4:0] = terrain type, bits [7:5] = growth stage
//   [7] variation/industry
//
// OpenLoco reference: src/OpenLoco/src/Map/SurfaceElement.h

const (
	// ElementType is stored in bits [5:2] of the first byte of a TileElement.
	// OpenLoco: TileElementBase::type() uses (_type & 0x3C) >> 2 — NOT 0xFC.
	// Bits 6-7 are element-specific flags (isConstructed, hasSignal, etc.)
	// and must NOT be included in the type mask.
	elementTypeMask  = 0x3C // bits 5-2 only
	elementTypeShift = 2

	// ElementFlags byte (second byte):
	elementFlagLast = 0x80 // last element on this tile

	// Known element type values (after >> 2)
	elementTypeSurface  = 0
	elementTypeTrack    = 1
	elementTypeStation  = 2
	elementTypeSignal   = 3
	elementTypeBuilding = 4
	elementTypeTree     = 5
	elementTypeWall     = 6
	elementTypeRoad     = 7
	elementTypeIndustry = 8
)

// terrainTypes maps the 5-bit terrain field in a surface element to our SurfaceType.
// OpenLoco reference: src/OpenLoco/src/Map/SurfaceElement.h  TerrainType enum
// Values 0-29 are defined; anything ≥30 is unknown.
var terrainTypes = [32]SurfaceType{
	0: SurfaceGrass, // grassland
	1: SurfaceDirt,  // sand / dirt
	2: SurfaceRock,  // rock
	3: SurfaceDirt,  // mud
	4: SurfaceGrass, // grass (variant)
	5: SurfaceSnow,  // snow
	6: SurfaceRock,  // gravel
	7: SurfaceDirt,  // dirt (variant)
	// 8-29: additional variants — default to grass
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// LoadScenarioData loads and returns scenario data (for main.go use).
func LoadScenarioData(filePath string) (*Scenario, error) {
	log.Println("[Scenario] Loading scenario from:", filePath)

	data, err := assets.LoadDatFile(filePath)
	if err != nil {
		log.Println("[Scenario] ERROR reading file:", err)
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}
	log.Println("[Scenario] File read successfully, size:", len(data), "bytes")

	sc, err := ParseScenario(data, filePath)
	if err != nil {
		return nil, err
	}

	log.Printf("[Scenario] Scenario loaded: %dx%d map", sc.MapWidth, sc.MapHeight)
	return sc, nil
}

// ParseScenario reads an S5 scenario file by walking its chunk stream.
//
// Chunk sequence for a scenario (S5Type == 1), per OpenLoco S5.cpp importSave():
//  1. Header          (rotate)
//  2. ScenarioOptions (rotate)          — only if type == scenario
//  3. PackedObjects                     — only if numPackedObjects > 0
//  4. RequiredObjects (rotate)
//  5. GeneralState    (runLengthSingle)
//  6. Towns           (runLengthSingle)
//  7. Animations      (runLengthSingle)
//  8. TileElements    (runLengthMulti)  — only if tileManagerLoaded flag set
//
// For saved games (S5Type == 0) chunks 5-7 are replaced by a single
// GameState chunk.  We handle both paths.
func ParseScenario(data []byte, filePath string) (*Scenario, error) {
	log.Println("[Scenario] Parsing scenario data, size:", len(data), "bytes")

	if len(data) < 32 {
		return nil, errors.New("scenario file too small")
	}

	reader := assets.NewS5ChunkReader(data)

	// --- Chunk 1: Header (rotate-encoded) ---
	headerBytes, err := reader.ReadChunk()
	if err != nil {
		return nil, fmt.Errorf("failed to read S5 header chunk: %w", err)
	}
	header, err := assets.ParseS5Header(headerBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S5 header: %w", err)
	}
	log.Printf("[Scenario] S5 Header: Type=%d Flags=0x%02X NumPackedObjects=%d Version=0x%X Magic=0x%X",
		header.Type, header.Flags, header.NumPackedObjects, header.Version, header.Magic)

	sc := &Scenario{
		Header:    header,
		FilePath:  filePath,
		MapWidth:  DefaultMapSize,
		MapHeight: DefaultMapSize,
	}

	// --- Chunk 2: SaveDetails or ScenarioOptions ---
	// SaveDetails comes first if hasSaveDetails flag is set (saved games).
	// ScenarioOptions come next if type == scenario.
	if header.Flags&assets.HeaderFlagHasSaveDetails != 0 {
		saveDetails, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read SaveDetails chunk: %w", err)
		}
		log.Println("[Scenario] Read SaveDetails,", len(saveDetails), "bytes")
	}
	if header.Type == assets.S5TypeScenario || header.Type == assets.S5TypeLandscape {
		scenarioOpts, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read ScenarioOptions chunk: %w", err)
		}
		log.Printf("[Scenario] Read ScenarioOptions, %d bytes", len(scenarioOpts))
		sc.Options = parseScenarioOptions(scenarioOpts)
	}

	// --- Packed objects (raw ObjectHeader + data chunk per object) ---
	if header.NumPackedObjects > 0 {
		log.Printf("[Scenario] Loading %d packed objects", header.NumPackedObjects)
		for i := uint16(0); i < header.NumPackedObjects; i++ {
			// Each packed object is: 16-byte ObjectHeader (raw, no chunk framing)
			// followed by one chunk of object data.
			if reader.Offset()+16 > len(data) {
				return sc, fmt.Errorf("truncated packed object header at index %d", i)
			}
			// Capture the raw 16-byte ObjectHeader before advancing.
			var rawHdr [16]byte
			copy(rawHdr[:], data[reader.Offset():reader.Offset()+16])
			advanceRaw(reader, 16)

			// Decode the object data chunk (run-length / rotate encoded).
			objData, err := reader.ReadChunk()
			if err != nil {
				return sc, fmt.Errorf("failed to read packed object %d data: %w", i, err)
			}

			sc.PackedObjectsRaw = append(sc.PackedObjectsRaw, PackedObjectRaw{
				HeaderBytes: rawHdr,
				Data:        objData,
			})
		}
		log.Printf("[Scenario] Decoded %d packed objects", len(sc.PackedObjectsRaw))
	}

	// --- Required objects (rotate-encoded) ---
	reqObjBytes, err := reader.ReadChunk()
	if err != nil {
		return sc, fmt.Errorf("failed to read RequiredObjects chunk: %w", err)
	}
	log.Println("[Scenario] Read RequiredObjects,", len(reqObjBytes), "bytes")
	sc.LandObjectOrder = parseLandObjectOrder(reqObjBytes)
	sc.CargoObjectOrder = parseObjectOrder(reqObjBytes, cargoSlotStart, cargoSlotCount, objectTypeCargo)
	sc.TownNamesObjectOrder = parseObjectOrder(reqObjBytes, 203, 1, 7)    // ObjectType::townNames (slot start from the compiler)
	sc.RegionObjectOrder = parseObjectOrder(reqObjBytes, 825, 1, 31)      // ObjectType::region (slot start from the compiler)
	sc.ClimateObjectOrder = parseObjectOrder(reqObjBytes, 678, 1, 26)     // ObjectType::climate (slot start from the compiler)
	sc.HillShapesObjectOrder = parseObjectOrder(reqObjBytes, 679, 1, 27)  // ObjectType::hillShapes (slot start from the compiler)
	sc.CompetitorObjectOrder = parseObjectOrder(reqObjBytes, 826, 32, 32) // ObjectType::competitor (slot start from the compiler)
	sc.TreeObjectOrder = parseTreeObjectOrder(reqObjBytes)
	sc.BuildingObjectOrder = parseBuildingObjectOrder(reqObjBytes)
	sc.WallObjectOrder = parseWallObjectOrder(reqObjBytes)
	sc.TrainSignalObjectOrder = parseTrainSignalObjectOrder(reqObjBytes)
	sc.IndustryObjectOrder = parseIndustryObjectOrder(reqObjBytes)
	// Track extras: 8 slots after the train stations (313 + 16), ObjectType 16.
	sc.TrackExtraOrder = parseObjectOrder(reqObjBytes, 329, 8, 16)
	// Road extras: 4 slots after the tracks (337 + 8) and road stations (345 + 16), ObjectType 19.
	sc.RoadExtraOrder = parseObjectOrder(reqObjBytes, 361, 4, 19)
	// Currency: one slot, after the interface skin (1) and the sounds (128).
	if o := parseObjectOrder(reqObjBytes, 129, 1, objectTypeCurrency); len(o) == 1 {
		sc.CurrencyObject = o[0]
	}
	sc.LevelCrossingObjectOrder = parseLevelCrossingObjectOrder(reqObjBytes)
	sc.BridgeObjectOrder = parseBridgeObjectOrder(reqObjBytes)
	sc.TrainStationObjectOrder = parseTrainStationObjectOrder(reqObjBytes)
	sc.RoadStationObjectOrder = parseRoadStationObjectOrder(reqObjBytes)
	sc.AirportObjectOrder = parseObjectOrder(reqObjBytes, airportSlotStart, airportSlotCount, objectTypeAirport)
	sc.DockObjectOrder = parseObjectOrder(reqObjBytes, dockSlotStart, dockSlotCount, objectTypeDock)
	sc.TrackObjectOrder = parseTrackObjectOrder(reqObjBytes)
	sc.RoadObjectOrder = parseRoadObjectOrder(reqObjBytes)
	sc.VehicleObjectOrder = parseVehicleObjectOrder(reqObjBytes)
	parseObjective := func(buf []byte) {
		// GameState objective block at 0x418.
		// OpenLoco reference: include/OpenLoco/S5/S5GameState.h 0x000418
		if len(buf) < 0x429 {
			return
		}
		sc.Objective = Objective{
			Type:                 buf[0x418],
			Flags:                buf[0x419],
			CompanyValue:         binary.LittleEndian.Uint32(buf[0x41A:]),
			MonthlyVehicleProfit: binary.LittleEndian.Uint32(buf[0x41E:]),
			PerformanceIndex:     buf[0x422],
			DeliveredCargoType:   buf[0x423],
			DeliveredCargoAmount: binary.LittleEndian.Uint32(buf[0x424:]),
			TimeLimitYears:       buf[0x428],
		}
		if len(buf) >= 0x42F { // scenarioObjectiveProgress (0x429)
			sc.ObjectiveTimeLimitUntilYear = binary.LittleEndian.Uint16(buf[0x429:])
			sc.ObjectiveMonthsInChallenge = binary.LittleEndian.Uint16(buf[0x42B:])
		}
		sc.HasObjective = true
		log.Printf("[Scenario] Objective: %s", sc.Objective.Describe())
	}

	// --- Game state chunks ---
	// For scenarios: three separate chunks (GeneralState, Towns, Animations).
	// For saved games: one big GameState chunk.
	// We only need the flags field from GeneralState to know if tiles follow.
	var generalStateFlags uint32
	if header.Type == assets.S5TypeScenario || header.Type == assets.S5TypeLandscape {
		generalState, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read GeneralState chunk: %w", err)
		}
		log.Println("[Scenario] Read GeneralState,", len(generalState), "bytes")
		generalStateFlags = readGeneralStateFlags(generalState)
		parseObjective(generalState)
		parseGeneralState(sc, generalState)
		animCount := 0 // numMapAnimations (GeneralState 0x154)
		if len(generalState) >= 0x156 {
			animCount = int(binary.LittleEndian.Uint16(generalState[0x154:]))
		}

		// Towns.  The chunk begins with kMaxTowns (80) Town records, followed by
		// the Industries and Stations arrays (the writer bundles all three under
		// one runLengthSingle chunk).  We only parse the leading town records.
		// OpenLoco reference: src/OpenLoco/src/S5/S5.cpp exportGameStateToFile()
		//   fs.writeChunk(runLengthSingle, gameState.towns, 0x123480)
		towns, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read Towns chunk: %w", err)
		}
		log.Println("[Scenario] Read Towns,", len(towns), "bytes")
		sc.Towns = parseTowns(towns)
		log.Printf("[Scenario] Parsed %d towns", len(sc.Towns))

		// Animations
		animations, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read Animations chunk: %w", err)
		}
		log.Println("[Scenario] Read Animations,", len(animations), "bytes")
		sc.Animations = parseAnimations(animations, animCount)
	} else {
		gameState, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read GameState chunk: %w", err)
		}
		log.Println("[Scenario] Read GameState,", len(gameState), "bytes")
		generalStateFlags = readGeneralStateFlags(gameState)
		parseObjective(gameState)
		parseGeneralState(sc, gameState)
		off := 0x4268C4 // GameState::animations (type 2: 0x424CA4)
		if len(gameState) == GameStateSizeType2 {
			off = 0x424CA4
		}
		if len(gameState) > 0x156 && len(gameState) >= off {
			sc.Animations = parseAnimations(gameState[off:], int(binary.LittleEndian.Uint16(gameState[0x154:])))
		}

		// Parse vehicle entities from the entity table embedded in the GameState.
		// This is only present in .SV5 saved-game files; ParseVehicleEntities
		// returns nil if the chunk is smaller than EntityTableOffset + table size.
		// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
		sc.Entities = ParseVehicleEntities(gameState)
		sc.EntityLists = ParseEntityLists(gameState)
		sc.VehicleList = ParseVehicleListEntries(gameState)
		sc.Orders = ParseOrders(gameState)
		sc.Routings = ParseRoutings(gameState)
		sc.Companies = parseCompanyEconomy(gameState)
		sc.Industries = parseIndustries(gameState)
		if sc.Entities != nil {
			log.Printf("[Scenario] Parsed %d vehicle entities (Head+Body) from GameState", len(sc.Entities))
		}

		sc.IndustryObjectIDs = parseIndustryObjectIDs(gameState)
		sc.CompanyColours = parseCompanyColours(gameState)
		sc.CompanyNames = parseCompanyNames(gameState)
		sc.CompanyFaces = parseCompanyFaces(gameState)
		sc.CompanyColourSchemes = parseCompanyColourSchemes(gameState)
		sc.UnlockedVehicles = parseUnlockedVehicles(gameState)
		sc.StationOwners = parseStationOwners(gameState)
		sc.UserStrings = parseUserStrings(gameState)
		sc.StationNames = parseStationNames(gameState)

		// Towns live inside the GameState chunk for saved games.
		sc.Towns = parseTownsFromGameState(gameState)
		if len(sc.Towns) > 0 {
			log.Printf("[Scenario] Parsed %d towns from GameState", len(sc.Towns))
		}
	}

	// --- Tile elements (runLengthMulti-encoded) ---
	// Only present if the tileManagerLoaded flag (bit 0) is set in general.flags.
	const gameFlagTileManagerLoaded = 1
	if generalStateFlags&gameFlagTileManagerLoaded != 0 {
		tileData, err := reader.ReadChunk()
		if err != nil {
			log.Printf("[Scenario] Warning: failed to read TileElements chunk: %v", err)
			sc.initPlaceholderMap()
			return sc, nil
		}
		log.Println("[Scenario] Read TileElements,", len(tileData), "bytes (", len(tileData)/tileElementSize, "elements)")
		sc.parseTileElements(tileData)
	} else {
		// No tile elements in file. For "randomly generated" scenarios
		// (landscapeGenerationDone flag clear) generate terrain from the
		// stored parameters; otherwise fall back to a flat placeholder.
		if sc.Options.Flags&ScenarioFlagLandscapeGenerationDone == 0 &&
			(sc.Header.Type == assets.S5TypeScenario || sc.Header.Type == assets.S5TypeLandscape) {
			log.Println("[Scenario] Generating landscape from scenario parameters")
			sc.initGeneratedLandscape() // stand-in until the game's Scenario::start runs sim.GenerateLandscape
			sc.LandscapeToGenerate = true
		} else {
			log.Println("[Scenario] tileManagerLoaded flag not set — using placeholder map")
			sc.initPlaceholderMap()
		}
	}

	if sc.Tiles == nil {
		log.Println("[Scenario] No tiles loaded, using placeholder map")
		sc.initPlaceholderMap()
	}

	return sc, nil
}

// ---------------------------------------------------------------------------
// Tile element parsing
// ---------------------------------------------------------------------------

// parseTileElements walks the flat byte array of 8-byte TileElement records
// and builds the Tile grid.  Each tile on the map has one or more elements;
// the last element for a tile has the "last" flag (bit 7 of byte 1) set.
// Elements are stored in tile order: (0,0), (1,0), … (383,0), (0,1), …
//
// OpenLoco reference: src/OpenLoco/src/Map/TileManager.cpp
//
//	TileManager::setElements()
//	TileManager::updateTilePointers()
func (sc *Scenario) parseTileElements(data []byte) {
	sc.Tiles = make([][]Tile, sc.MapHeight)
	for y := range sc.Tiles {
		sc.Tiles[y] = make([]Tile, sc.MapWidth)
		for x := range sc.Tiles[y] {
			sc.Tiles[y][x] = Tile{Surface: SurfaceGrass} // default
		}
	}

	x, y := 0, 0
	pos := 0
	elemCount := 0
	treeCount := 0
	buildingCount := 0
	wallCount := 0
	trackCount := 0
	roadCount := 0
	stationCount := 0
	signalCount := 0
	industryCount := 0

	for pos+tileElementSize <= len(data) {
		elem := data[pos : pos+tileElementSize]
		pos += tileElementSize
		elemCount++

		typeByte := elem[0]
		flagsByte := elem[1]
		baseZ := elem[2]
		// clearZ := elem[3]  // not needed for surface rendering yet

		elemType := (typeByte & elementTypeMask) >> elementTypeShift

		if x < sc.MapWidth && y < sc.MapHeight {
			switch elemType {
			case elementTypeSurface:
				// Surface element byte layout:
				//   byte 4: slope — bits [4:0] = slope type, bits [7:5] = snow coverage
				//   byte 5: water — bits [4:0] = water level (MicroZ)
				//   byte 6: terrain — bits [4:0] = LandObject slot, bits [7:5] = growth stage
				//   byte 7: variation — alternate terrain style (0 = default)
				//
				// OpenLoco reference: src/OpenLoco/src/Map/SurfaceElement.h
				terrainRaw := elem[6] & 0x1F // bits [4:0] → LandObject slot (0-31)
				growthStage := elem[6] >> 5  // bits [7:5] → growth/season stage (0-7)
				variation := elem[7]         // alternate terrain style

				surface := SurfaceGrass
				if int(terrainRaw) < len(terrainTypes) {
					surface = terrainTypes[terrainRaw]
				}

				// Water level in byte 5, bits [4:0] — MicroZ units (1 MicroZ = 16 px).
				// The field is set for ALL tiles (it's the global water table height).
				// A tile is submerged only when water surface is above terrain:
				//   waterLevel * 16 > baseZ * 4
				waterLevel := elem[5] & 0x1F
				if waterLevel > 0 && uint16(waterLevel)*16 > uint16(baseZ)*4 {
					surface = SurfaceWater
				}

				// Slope in byte 4, bits [4:0] (corner heights + double-height flag).
				slopeByte := elem[4]

				// Preserve any existing non-surface elements parsed before
				// the surface element (elements are not guaranteed to be ordered).
				existing := sc.Tiles[y][x]
				sc.Tiles[y][x] = Tile{
					Surface:      surface,
					Height:       baseZ,
					Water:        waterLevel,
					UpdateTimer:  elem[5] >> 5, // getUpdateTimer: an industrial field's
					TerrainIndex: terrainRaw,
					GrowthStage:  growthStage,
					Variation:    variation,
					IsIndustrial: typeByte&0x80 != 0,
					Type6Flag:    typeByte&0x40 != 0,
					Slope:        slopeByte,
					Trees:        existing.Trees,
					Buildings:    existing.Buildings,
					Walls:        existing.Walls,
					Tracks:       existing.Tracks,
					Roads:        existing.Roads,
					Stations:     existing.Stations,
					Signals:      existing.Signals,
					Industries:   existing.Industries,
					Order:        append(existing.Order, ElemRef{Kind: elementTypeSurface, Flags: flagsByte}),
				}

			case elementTypeTree:
				// TreeElement: byte layout from TreeElement.h
				te := TreeElement{
					TreeObjectID: elem[4],
					Rotation:     typeByte & 0x03,
					Quadrant:     (typeByte >> 6) & 0x03,
					Growth:       elem[5] & 0x0F,
					Season:       (elem[7] >> 3) & 0x07,
					Colour:       elem[6] & 0x1F,
					HasSnow:      elem[6]&0x40 != 0,
					BaseZ:        baseZ,
					ClearZ:       elem[3],
					Unk7l:        elem[7] & 0x07,
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeTree, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Trees))})
				sc.Tiles[y][x].Trees = append(sc.Tiles[y][x].Trees, te)
				treeCount++

			case elementTypeBuilding:
				be := BuildingElement{
					ObjectID:        elem[4],
					Rotation:        typeByte & 0x03,
					SequenceIndex:   elem[5] & 0x03,
					Variation:       ((elem[7] & 0x07) << 2) | ((elem[6] >> 6) & 0x03),
					Colour:          (elem[7] >> 3) & 0x1F,
					IsConstructed:   typeByte&0x80 != 0,
					IsMisc:          typeByte&0x40 != 0,
					BaseZ:           baseZ,
					ClearZ:          elem[3],
					Age:             elem[6] & 0x3F,
					SectionProgress: elem[5] >> 5,
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeBuilding, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Buildings))})
				sc.Tiles[y][x].Buildings = append(sc.Tiles[y][x].Buildings, be)
				buildingCount++

			case elementTypeWall:
				// WallElement: byte layout from WallElement.h
				we := WallElement{
					WallObjectID:    elem[4],
					Rotation:        typeByte & 0x03,
					EdgeSlope:       (typeByte >> 6) & 0x03,
					PrimaryColour:   elem[6] & 0x1F,
					SecondaryColour: (elem[6]>>5)&0x7 | (elem[1]>>2)&0x18,
					BaseZ:           baseZ,
					ClearZ:          elem[3],
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeWall, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Walls))})
				sc.Tiles[y][x].Walls = append(sc.Tiles[y][x].Walls, we)
				wallCount++

			case elementTypeTrack:
				// TrackElement: byte layout from TrackElement.h
				// bridge(): elem[6] >> 5 (bits [7:5]) — OpenLoco TrackElement.h
				te := TrackElement{
					TrackID:          elem[4] & 0x3F,
					IsLast:           elem[1]&0x40 != 0, // ElementFlags::flag_6
					HasLevelCrossing: elem[6]&0x10 != 0,
					Ghost:            elem[1]&0x10 != 0,
					AiAllocated:      elem[1]&0x20 != 0,
					HasBridge:        elem[4]&0x80 != 0,
					BridgeID:         elem[6] >> 5,
					SequenceIndex:    elem[5] & 0x0F,
					TrackObjectID:    (elem[5] >> 4) & 0x0F,
					Rotation:         typeByte & 0x03,
					HasSignal:        typeByte&0x40 != 0,
					HasStation:       typeByte&0x80 != 0,
					Owner:            elem[7] & 0x0F,
					Mods:             (elem[7] >> 4) & 0x0F,
					BaseZ:            baseZ,
					ClearZ:           elem[3],
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeTrack, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Tracks))})
				sc.Tiles[y][x].Tracks = append(sc.Tiles[y][x].Tracks, te)
				trackCount++

			case elementTypeRoad:
				// RoadElement: byte layout from RoadElement.h
				// BridgeID: byte 6 bits [7:5]; HasLevelCrossing: byte 7 bit 5;
				// LevelCrossingObjectID: byte 5 bits [3:2]; AnimFrame: byte 6 bits [3:0]
				re := RoadElement{
					RoadID:                elem[4] & 0x0F,
					HasBridge:             elem[4]&0x80 != 0,
					BridgeID:              elem[6] >> 5,
					SequenceIndex:         elem[5] & 0x03,
					LevelCrossingObjectID: (elem[5] >> 2) & 0x03,
					RoadObjectID:          (elem[5] >> 4) & 0x0F,
					AnimFrame:             elem[6] & 0x0F,
					HasLevelCrossing:      elem[7]&(1<<5) != 0,
					HasStationElement:     typeByte&0x80 != 0,
					HasSignalElement:      typeByte&0x40 != 0,
					Rotation:              typeByte & 0x03,
					Owner:                 elem[7] & 0x0F,
					Mods:                  (elem[7] >> 6) & 0x03,
					BaseZ:                 baseZ,
					ClearZ:                elem[3],
					IsLast:                elem[1]&0x40 != 0,
					Ghost:                 elem[1]&0x10 != 0,
					AiAllocated:           elem[1]&0x20 != 0,
					LaneOccupation:        (elem[4] >> 4) & 3,
					Unk7_10:               elem[7]&0x10 != 0,
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeRoad, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Roads))})
				sc.Tiles[y][x].Roads = append(sc.Tiles[y][x].Roads, re)
				roadCount++

			case elementTypeStation:
				// StationElement: byte layout from StationElement.h
				stationWord := uint16(elem[6]) | uint16(elem[7])<<8
				se := StationElement{
					StationType:   elem[5] >> 5,
					ObjectID:      elem[5] & 0x1F,
					Rotation:      typeByte & 0x03,
					SequenceIndex: (typeByte >> 6) & 0x03,
					StationID:     stationWord & 0x03FF,
					BuildingType:  uint8((stationWord >> 10) & 0x3F),
					Owner:         elem[4] & 0x0F,
					BaseZ:         baseZ,
					ClearZ:        elem[3],
					Ghost:         elem[1]&0x10 != 0,
					AiAllocated:   elem[1]&0x20 != 0,
					Flag6:         elem[1]&0x40 != 0,
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeStation, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Stations))})
				sc.Tiles[y][x].Stations = append(sc.Tiles[y][x].Stations, se)
				stationCount++

			case elementTypeSignal:
				// SignalElement: byte layout from SignalElement.h
				sig := SignalElement{
					SignalObjectID:      elem[4] & 0x0F,
					RightSignalObjectID: elem[6] & 0x0F,
					Rotation:            typeByte & 0x03,
					HasLeftSignal:       elem[4]&0x80 != 0,
					LeftFrame:           elem[5] & 0x0F,
					HasRightSignal:      elem[6]&0x80 != 0,
					RightFrame:          elem[7] & 0x0F,
					BaseZ:               baseZ,
					ClearZ:              elem[3],
					Ghost:               elem[1]&0x10 != 0,
					AiAllocated:         elem[1]&0x20 != 0,
					Sides:               [2]SignalSide{{elem[4], elem[5]}, {elem[6], elem[7]}},
					LeftGhost:           typeByte&0x80 != 0,
					RightGhost:          typeByte&0x40 != 0,
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeSignal, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Signals))})
				sc.Tiles[y][x].Signals = append(sc.Tiles[y][x].Signals, sig)
				signalCount++

			case elementTypeIndustry:
				// IndustryElement: byte layout from IndustryElement.h
				sectionWord := uint16(elem[6]) | uint16(elem[7])<<8
				ie := IndustryElement{
					IndustryID:      elem[4],
					BuildingType:    uint8((sectionWord >> 6) & 0x1F),
					Colour:          uint8((sectionWord >> 11) & 0x1F), // var_6_F800
					Var6Low:         uint8(sectionWord & 0x3F),         // var_6_003F
					SectionProgress: (elem[5] >> 5) & 0x7,
					SequenceIndex:   elem[5] & 0x3,
					Rotation:        typeByte & 0x03,
					IsConstructed:   typeByte&0x80 != 0,
					BaseZ:           baseZ,
					ClearZ:          elem[3],
				}
				sc.Tiles[y][x].Order = append(sc.Tiles[y][x].Order, ElemRef{Kind: elementTypeIndustry, Flags: flagsByte, Index: uint16(len(sc.Tiles[y][x].Industries))})
				sc.Tiles[y][x].Industries = append(sc.Tiles[y][x].Industries, ie)
				industryCount++
			}
		}

		// Advance to next tile when we see the "last" flag
		if flagsByte&elementFlagLast != 0 {
			x++
			if x >= sc.MapWidth {
				x = 0
				y++
			}
		}
	}

	log.Printf("[Scenario] Parsed %d tile elements covering %d×%d tiles (stopped at x=%d y=%d) — %d trees, %d buildings, %d walls, %d tracks, %d roads, %d stations, %d signals, %d industries",
		elemCount, sc.MapWidth, sc.MapHeight, x, y, treeCount, buildingCount, wallCount, trackCount, roadCount, stationCount, signalCount, industryCount)
}

// readGeneralStateFlags extracts the flags uint32 from a GeneralState or
// GameState blob.  The flags field is at a fixed offset within the struct.
//
// OpenLoco reference: src/OpenLoco/src/S5/GameState.h
//
//	GeneralState::flags  — offset 0x434 from start of GameState,
//	but GeneralState itself starts at offset 0 when read as a standalone chunk.
//
// For a scenario file the GeneralState chunk contains only the GeneralState
// sub-struct (not the full GameState), so flags is near the beginning.
// The flags field is at offset 0x00 within GeneralState per the OpenLoco
// layout; however empirically for the standalone GeneralState chunk it is at
// a small offset.  We scan for the tileManagerLoaded pattern: if we can't
// find it definitively, default to assuming tiles are present (which is the
// common case for title.dat and scenario files).
func readGeneralStateFlags(data []byte) uint32 {
	// GeneralState.flags is at byte offset 0x10 within the struct:
	//
	//   0x00  rng[0]    uint32
	//   0x04  rng[1]    uint32
	//   0x08  unkRng[0] uint32
	//   0x0C  unkRng[1] uint32
	//   0x10  flags     uint32  ← tileManagerLoaded = bit 0
	//
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GeneralState::flags
	const off = GeneralFlagsOffset // 0x10 = 16
	if len(data) >= off+4 {
		flags := uint32(data[off]) | uint32(data[off+1])<<8 |
			uint32(data[off+2])<<16 | uint32(data[off+3])<<24
		log.Printf("[Scenario] GeneralState flags = 0x%08X (offset 0x%02X)", flags, off)
		return flags
	}
	// Default: assume tile manager is loaded (safest for very short chunks)
	log.Println("[Scenario] GeneralState too short to read flags, assuming tileManagerLoaded")
	return 1
}

// advanceRaw moves the reader forward by n raw bytes (no chunk framing).
// This is used to skip ObjectHeader structs that are written raw (not as chunks).
// parseScenarioOptions extracts metadata from the raw ScenarioOptions (S5::Options) chunk.
//
// Field offsets from src/OpenLoco/src/S5/S5Options.h (packed, 1-byte alignment):
//
//	0x01        difficulty    uint8   — scenario category (0=Beginner … 4=Expert)
//	0x02-0x03  startYear     uint16  — start year
//	0x06-0x07  scenarioFlags uint16  — bit 0 = landscapeGenerationDone
//	0x2A-0x69  scenarioName  char[64]
//	0x6A-0x169 scenarioDetails char[256]
//	0x184      minLandHeight  uint8
//	0x185      topographyStyle uint8
//	0x186      hillDensity    uint8
//	0x18A-0x4189 preview[128][128] uint8 — palette-indexed preview image
//	0x418A     maxCompetingCompanies uint8
//	0x418C     objective.type uint8
func parseScenarioOptions(data []byte) Options {
	var opts Options
	if len(data) < 0x08 {
		return opts
	}
	opts.Category = ScenarioCategory(data[0x01])
	opts.StartYear = uint16(data[0x02]) | uint16(data[0x03])<<8
	opts.Flags = uint16(data[0x06]) | uint16(data[0x07])<<8
	opts.HasPreview = opts.Flags&ScenarioFlagLandscapeGenerationDone != 0

	if len(data) >= 0x6A {
		opts.Name = nullTermString(data[0x2A:0x6A])
	}
	if len(data) >= 0x6A+256 {
		opts.Description = nullTermString(data[0x6A : 0x6A+256])
	}
	// Generation parameters
	if len(data) >= 0x187 {
		opts.MinLandHeight = data[0x184]
		opts.TopographyStyle = data[0x185]
		opts.HillDensity = data[0x186]
		opts.NumberOfForests = binary.LittleEndian.Uint16(data[0x17A:]) // trees (offsets from the compiler)
		opts.MinForestRadius, opts.MaxForestRadius = data[0x17C], data[0x17D]
		opts.MinForestDensity, opts.MaxForestDensity = data[0x17E], data[0x17F]
		opts.NumberRandomTrees = binary.LittleEndian.Uint16(data[0x180:])
		opts.MinAltitudeForTrees, opts.MaxAltitudeForTrees = data[0x182], data[0x183]
		copy(opts.LandDistributionPatterns[:], data[0x0A:0x0A+32]) // landDistributionPatterns
		if len(data) > 0x41C3 {                                    // OpenLoco's new fields (zero — padding — in vanilla files)
			opts.Generator = data[0x41BD]
			opts.NumRiverbeds, opts.MinRiverWidth, opts.MaxRiverWidth = data[0x41BF], data[0x41C0], data[0x41C1]
			opts.RiverbankWidth, opts.RiverMeanderRate = data[0x41C2], data[0x41C3]
		}
	}
	// Preview image (128×128 palette-indexed bytes)
	const previewOffset = 0x18A
	const previewBytes = PreviewSize * PreviewSize
	if len(data) >= previewOffset+previewBytes {
		copy(opts.Preview[:], data[previewOffset:previewOffset+previewBytes])
	}
	// Competing companies
	if len(data) >= 0x418B {
		opts.MaxCompetitors = data[0x418A]
	}
	// Full objective block at 0x418C, then the delivered-cargo ObjectHeader
	// at 0x419D (name bytes 4-11).
	// OpenLoco reference: include/OpenLoco/S5/S5Options.h
	if len(data) >= 0x419D+16 {
		opts.ObjectiveType = data[0x418C]
		opts.Objective = Objective{
			Type:                 data[0x418C],
			Flags:                data[0x418D],
			CompanyValue:         binary.LittleEndian.Uint32(data[0x418E:]),
			MonthlyVehicleProfit: binary.LittleEndian.Uint32(data[0x4192:]),
			PerformanceIndex:     data[0x4196],
			DeliveredCargoType:   data[0x4197],
			DeliveredCargoAmount: binary.LittleEndian.Uint32(data[0x4198:]),
			TimeLimitYears:       data[0x419C],
		}
		opts.HasObjective = true
		opts.ObjectiveCargoName = nullTermString(data[0x419D+4 : 0x419D+12])
	}
	return opts
}

// nullTermString returns a Go string from a null-terminated (or space-padded) byte slice.
func nullTermString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return strings.TrimSpace(string(b[:i]))
		}
	}
	return strings.TrimSpace(string(b))
}

func advanceRaw(r *assets.S5ChunkReader, n int) {
	// S5ChunkReader doesn't expose a skip method, so we manipulate via the
	// exported Offset and reconstruct.  For now we use a simple wrapper that
	// re-slices internally.  This is only called for packed object headers
	// which are rare.
	r.AdvanceRaw(n)
}

// ---------------------------------------------------------------------------
// RequiredObjects → land-object order
// ---------------------------------------------------------------------------

// ObjectHeader layout (16 bytes, little-endian):
//
//	[0:4]  uint32 flags   — type in bits [5:0], sourceGame in bits [7:6]
//	[4:12] [8]byte name
//	[12:16] uint32 checksum
//
// OpenLoco reference: src/OpenLoco/src/Objects/Object.h  ObjectHeader

const (
	objectHeaderSize       = 16
	objectTypeMask         = 0x3F
	objectTypeCurrency     = 2  // ObjectType::currency
	objectTypeLand         = 6  // ObjectType::land
	objectTypeCargo        = 8  // ObjectType::cargo
	objectTypeWall         = 9  // ObjectType::wall
	objectTypeTrackStation = 15 // ObjectType::trackStation
	objectTypeTrack        = 17 // ObjectType::track
	objectTypeRoadStation  = 18 // ObjectType::roadStation
	objectTypeRoad         = 20 // ObjectType::road
	objectTypeTree         = 24 // ObjectType::tree
	objectTypeBuilding     = 28 // ObjectType::building

	// Offsets into the 859-element RequiredObjects array.
	// Cumulative counts: Interface(1)+Sound(128)+Currency(1)+Steam(32)+Rock(8)+Water(1)=171
	landSlotStart = 171
	landSlotCount = 32

	cargoSlotStart = 204
	cargoSlotCount = 32

	// Wall slots: after land(32)+TownNames(1)+Cargo(32) = 171+32+1+32 = 236
	wallSlotStart = 236
	wallSlotCount = 32

	// TrackSignal slots: Wall(32) = 236+32 = 268
	trackSignalSlotStart  = 268
	trackSignalSlotCount  = 16
	objectTypeTrackSignal = 10 // ObjectType::trackSignal

	// LevelCrossing slots: TrackSignal(16) = 268+16 = 284
	levelCrossingSlotStart  = 284
	levelCrossingSlotCount  = 4
	objectTypeLevelCrossing = 11 // ObjectType::levelCrossing

	// Bridge slots: TrackSignal(16)+LevelCrossing(4)+StreetLight(1)+Tunnel(16) = 268+16+4+1+16 = 305
	bridgeSlotStart  = 305
	bridgeSlotCount  = 8
	objectTypeBridge = 14 // ObjectType::bridge

	// TrainStation slots: Wall(32)+TrackSignal(16)+LevelCrossing(4)+StreetLight(1)+Tunnel(16)+Bridge(8) = 236+32+16+4+1+16+8 = 313
	trainStationSlotStart = 313
	trainStationSlotCount = 16

	// Track slots: TrainStation(16)+TrackExtra(8) = 313+16+8 = 337
	trackSlotStart = 337
	trackSlotCount = 8

	// RoadStation slots: Track(8) = 337+8 = 345
	roadStationSlotStart = 345
	roadStationSlotCount = 16

	// Road slots: RoadStation(16)+RoadExtra(4) = 345+16+4 = 365
	// Airports: ObjectType::airport (21), slots 373..380 (getMaxObjects, via the compiler).
	airportSlotStart  = 373
	airportSlotCount  = 8
	objectTypeAirport = 21
	// Docks: ObjectType::dock (22), slots 381..388.
	dockSlotStart  = 381
	dockSlotCount  = 8
	objectTypeDock = 22

	roadSlotStart = 365
	roadSlotCount = 8

	// Vehicle slots: Road(8)+Airport(8)+Dock(8) = 365+8+8+8 = 389
	vehicleSlotStart  = 389
	vehicleSlotCount  = 224
	objectTypeVehicle = 23 // ObjectType::vehicle

	// Tree slots: after Interface(1)+Sound(128)+Currency(1)+Steam(32)+Rock(8)+Water(1)+
	//   Surface(32)+TownNames(1)+Cargo(32)+Wall(32)+TrackSignal(16)+LevelCrossing(4)+
	//   StreetLight(1)+Tunnel(16)+Bridge(8)+TrainStation(16)+TrackExtra(8)+Track(8)+
	//   RoadStation(16)+RoadExtra(4)+Road(8)+Airport(8)+Dock(8)+Vehicle(224) = 613
	treeSlotStart = 613
	treeSlotCount = 64

	// Building slots: tree(64)+Snow(1)+Climate(1)+HillShapes(1) = 613+64+1+1+1 = 680
	buildingSlotStart = 680
	buildingSlotCount = 128

	// Industry: ObjectType::industry (30); slots from getMaxObjects summed over
	// the earlier types (computed with upstream's ObjectManager.h: 809, 16).
	industrySlotStart  = 809
	industrySlotCount  = 16
	objectTypeIndustry = 30
)

// parseObjectOrder extracts object names from a RequiredObjects blob for a
// given slot range and expected object type.
func parseObjectOrder(data []byte, slotStart, slotCount int, expectedType uint32) []string {
	order := make([]string, slotCount)
	for i := range slotCount {
		off := (slotStart + i) * objectHeaderSize
		if off+objectHeaderSize > len(data) {
			break
		}
		hdr := data[off : off+objectHeaderSize]

		// Check for empty header (flags == 0xFFFFFFFF)
		if hdr[0] == 0xFF && hdr[1] == 0xFF && hdr[2] == 0xFF && hdr[3] == 0xFF {
			continue
		}

		// Verify type field matches expected
		flags := uint32(hdr[0]) | uint32(hdr[1])<<8 | uint32(hdr[2])<<16 | uint32(hdr[3])<<24
		if flags&objectTypeMask != expectedType {
			continue
		}

		// Extract 8-byte name, trimming trailing nulls, 0xFF, and spaces.
		raw := hdr[4:12]
		nameEnd := 8
		for nameEnd > 0 && (raw[nameEnd-1] == 0 || raw[nameEnd-1] == 0xFF || raw[nameEnd-1] == ' ') {
			nameEnd--
		}
		order[i] = string(raw[:nameEnd])
	}
	return order
}

// parseLandObjectOrder extracts the 32 land-object names from a RequiredObjects
// blob in terrain-slot order.  Empty slots (all-0xFF headers) produce "".
func parseLandObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, landSlotStart, landSlotCount, objectTypeLand)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d land objects in slot order", count)
	return order
}

// parseTreeObjectOrder extracts the 64 tree-object names from RequiredObjects.
func parseTreeObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, treeSlotStart, treeSlotCount, objectTypeTree)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d tree objects in slot order", count)
	return order
}

// parseBuildingObjectOrder extracts the 128 building-object names from RequiredObjects.
// parseIndustryObjectOrder extracts the 16 industry-object names from RequiredObjects.
func parseIndustryObjectOrder(data []byte) []string {
	return parseObjectOrder(data, industrySlotStart, industrySlotCount, objectTypeIndustry)
}

func parseBuildingObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, buildingSlotStart, buildingSlotCount, objectTypeBuilding)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d building objects in slot order", count)
	return order
}

// parseWallObjectOrder extracts the 32 wall-object names from RequiredObjects.
func parseWallObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, wallSlotStart, wallSlotCount, objectTypeWall)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d wall objects in slot order", count)
	return order
}

// parseTrainStationObjectOrder extracts the 16 train-station-object names from RequiredObjects.
// Slots 313-328 (ObjectType::trackStation = 15).
func parseTrainStationObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, trainStationSlotStart, trainStationSlotCount, objectTypeTrackStation)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d train station objects in slot order", count)
	return order
}

// parseRoadStationObjectOrder extracts the 16 road-station-object names from RequiredObjects.
// Slots 345-360 (ObjectType::roadStation = 18).
func parseRoadStationObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, roadStationSlotStart, roadStationSlotCount, objectTypeRoadStation)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d road station objects in slot order", count)
	return order
}

// parseTrackObjectOrder extracts the 8 track-object names from RequiredObjects.
// Slots 337-344 (ObjectType::track = 17).
func parseTrackObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, trackSlotStart, trackSlotCount, objectTypeTrack)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d track objects in slot order", count)
	return order
}

// parseRoadObjectOrder extracts the 8 road-object names from RequiredObjects.
// Slots 365-372 (ObjectType::road = 20).
func parseRoadObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, roadSlotStart, roadSlotCount, objectTypeRoad)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d road objects in slot order", count)
	return order
}

// parseBridgeObjectOrder extracts the 8 bridge-object names from RequiredObjects.
// Slots 305-312 (ObjectType::bridge = 14).
func parseBridgeObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, bridgeSlotStart, bridgeSlotCount, objectTypeBridge)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d bridge objects in slot order", count)
	return order
}

// parseLevelCrossingObjectOrder extracts the 4 level-crossing-object names from RequiredObjects.
// Slots 284-287 (ObjectType::levelCrossing = 11).
func parseLevelCrossingObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, levelCrossingSlotStart, levelCrossingSlotCount, objectTypeLevelCrossing)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d level crossing objects in slot order", count)
	return order
}

// parseTrainSignalObjectOrder extracts the 16 train-signal-object names from RequiredObjects.
// Slots 268-283 (ObjectType::trackSignal = 10).
func parseTrainSignalObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, trackSignalSlotStart, trackSignalSlotCount, objectTypeTrackSignal)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d train signal objects in slot order", count)
	return order
}

// parseVehicleObjectOrder extracts the 224 vehicle-object names from RequiredObjects.
// Slots 389-612 (ObjectType::vehicle = 23).
func parseVehicleObjectOrder(data []byte) []string {
	order := parseObjectOrder(data, vehicleSlotStart, vehicleSlotCount, objectTypeVehicle)
	count := 0
	for _, n := range order {
		if n != "" {
			count++
		}
	}
	log.Printf("[Scenario] RequiredObjects: %d vehicle objects in slot order", count)
	return order
}

// ---------------------------------------------------------------------------
// Town parsing
// ---------------------------------------------------------------------------

// Town record layout within the Towns chunk / GameState towns array.
// OpenLoco reference: include/OpenLoco/S5/S5Town.h struct Town (packed).
const (
	townRecordSize         = 0x270 // sizeof(S5::Town)
	townNumStationsOffset  = 0x1A6 // offsetof(S5::Town, numStations), from the compiler
	townMaxCount           = 80    // S5::Limits::kMaxTowns
	townNameOffset         = 0x00  // uint16 StringId (0xFFFF = empty)
	townXOffset            = 0x02  // int16 world-units X
	townYOffset            = 0x04  // int16 world-units Y
	townPopulationOffset   = 0x30  // uint32
	townNumBuildingsOffset = 0x38  // int16
	townSizeOffset         = 0x5A  // uint8 TownSize

	// Offsets of the towns array inside a decompressed GameState chunk.
	// OpenLoco reference: include/OpenLoco/S5/S5GameState.h GameState::towns.
	townsGameStateOffset      = 0x092444
	townsGameStateOffsetType2 = 0x090824
)

// parseTowns reads up to kMaxTowns Town records from the start of buf, skipping
// empty slots (name == StringIds::null == 0xFFFF).
func parseTowns(buf []byte) []Town {
	var towns []Town
	for i := 0; i < townMaxCount; i++ {
		off := i * townRecordSize
		if off+townRecordSize > len(buf) {
			break
		}
		rec := buf[off : off+townRecordSize]
		name := binary.LittleEndian.Uint16(rec[townNameOffset:])
		if name == 0xFFFF {
			continue // empty slot (Town::empty)
		}
		towns = append(towns, Town{
			Name:         name,
			X:            int16(binary.LittleEndian.Uint16(rec[townXOffset:])),
			Y:            int16(binary.LittleEndian.Uint16(rec[townYOffset:])),
			Population:   binary.LittleEndian.Uint32(rec[townPopulationOffset:]),
			NumBuildings: int16(binary.LittleEndian.Uint16(rec[townNumBuildingsOffset:])),
			NumStations:  int16(binary.LittleEndian.Uint16(rec[townNumStationsOffset:])),
			ID:           uint16(i),
			LabelFrame:   ParseLabelFrame(rec[0x08:]),
			Var1A8:       binary.LittleEndian.Uint32(rec[0x1A8:]),
			Size:         rec[townSizeOffset],
		})
		for c := range 32 {
			towns[len(towns)-1].MonthlyCargoDelivered[c] = binary.LittleEndian.Uint16(rec[0x158+2*c:])
		}
		tw := &towns[len(towns)-1] // the rest of S5::Town (offsets from S5Town.h)
		tw.Flags = binary.LittleEndian.Uint16(rec[0x06:])
		tw.Prng = [2]uint32{binary.LittleEndian.Uint32(rec[0x28:]), binary.LittleEndian.Uint32(rec[0x2C:])}
		tw.PopulationCapacity = binary.LittleEndian.Uint32(rec[0x34:])
		for c := range 15 {
			tw.CompanyRatings[c] = int16(binary.LittleEndian.Uint16(rec[0x3A+2*c:]))
		}
		tw.CompaniesWithRating = binary.LittleEndian.Uint16(rec[0x58:])
		tw.HistorySize = rec[0x5B]
		copy(tw.History[:], rec[0x5C:0x5C+240])
		tw.HistoryMinPopulation = int32(binary.LittleEndian.Uint32(rec[0x14C:]))
		copy(tw.AmenityCounts[:], rec[0x150:0x158])
		tw.CargoInfluenceFlags = binary.LittleEndian.Uint32(rec[0x198:])
		for k := range 4 {
			tw.Var19C[k/2][k%2] = binary.LittleEndian.Uint16(rec[0x19C+2*k:])
		}
		tw.BuildSpeed, tw.NumberOfAirports = rec[0x1A4], rec[0x1A5]
	}
	return towns
}

// parseTownsFromGameState extracts town records from a full decompressed
// GameState chunk (saved games), choosing the array offset by chunk variant.
func parseTownsFromGameState(gs []byte) []Town {
	off := townsGameStateOffset
	if len(gs) == GameStateSizeType2 {
		off = townsGameStateOffsetType2
	}
	if len(gs) < off+townMaxCount*townRecordSize {
		return nil
	}
	return parseTowns(gs[off:])
}

// ---------------------------------------------------------------------------
// Placeholder fallback
// ---------------------------------------------------------------------------

// initPlaceholderMap creates a placeholder map with simple terrain.
// This is a development fallback used when tile elements are not present
// or cannot be parsed.  It should be removed once all scenario files load
// correctly.
//
// OpenLoco reference: see parseTileElements comment above —
//
//	src/OpenLoco/src/Scenario/Scenario.cpp  Scenario::load()
//	src/OpenLoco/src/Map/TileManager.cpp    TileManager
func (sc *Scenario) initPlaceholderMap() {
	sc.Tiles = make([][]Tile, sc.MapHeight)
	for y := 0; y < sc.MapHeight; y++ {
		sc.Tiles[y] = make([]Tile, sc.MapWidth)
		for x := 0; x < sc.MapWidth; x++ {
			surface := SurfaceGrass
			if x < 5 || x >= sc.MapWidth-5 || y < 5 || y >= sc.MapHeight-5 {
				surface = SurfaceWater
			}
			if (x+y)%7 == 0 {
				surface = SurfaceDirt
			}
			sc.Tiles[y][x] = Tile{
				Surface: surface,
				Height:  uint8((x + y) % 4),
				Order:   []ElemRef{{Kind: KindSurface}},
			}
		}
	}
}

// GetTile returns tile at the given coordinates.
func (sc *Scenario) GetTile(x, y int) *Tile {
	if x < 0 || x >= sc.MapWidth || y < 0 || y >= sc.MapHeight {
		return nil
	}
	if sc.Tiles == nil {
		return nil
	}
	return &sc.Tiles[y][x]
}

// IsTitleSequence returns true if this is a title sequence file.
func (sc *Scenario) IsTitleSequence() bool {
	return sc.Header.Flags&assets.HeaderFlagIsTitleSequence != 0
}

// Industry records in the GameState (S5/S5GameState.h, S5/S5Industry.h).
// Offsets and sizes as the compiler lays out upstream's structs (checked with
// offsetof/sizeof against OpenLoco b93e074): GameState::industries at 0x9E744,
// GameStateType2::industries at 0x9CB24; kMaxIndustries 128; sizeof(Industry)
// 1107 with objectId at byte 16.
const (
	industriesGameStateOffset      = 0x9E744
	industriesGameStateOffsetType2 = 0x9CB24
	industryMaxCount               = 128
	industryRecordSize             = 1107
	industryObjectIDOffset         = 16
)

// Companies in the GameState (S5/S5GameState.h, S5/S5Company.h; offsetof and
// sizeof from the compiler, OpenLoco b93e074): companies at 0xB96C in both
// GameState and GameStateType2; 15 records of 36776 bytes (CompanyType2:
// 36296); mainColours at 0x1A, primary first. Neutral (owner 15) is grey (1).
const (
	companiesGameStateOffset = 0xB96C
	companyMaxCount          = 15
	companyRecordSize        = 36776
	companyRecordSizeType2   = 36296
	companyMainColoursOffset = 0x1A
	companyNeutral           = 15
	colourGrey               = 1
)

// parseCompanyNames returns every company slot's name StringId (Company 0x00).
func parseCompanyNames(gs []byte) []uint16 {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	n := make([]uint16, companyMaxCount)
	for i := range n {
		o := companiesGameStateOffset + i*size
		n[i] = uint16(gs[o]) | uint16(gs[o+1])<<8
	}
	return n
}

// CompanyFace is what the owner's face and the company select order read of
// a company (S5::Company, offsets from the compiler, both layouts):
// competitorId 0x18, ownerEmotion 0x19, performanceIndex 0x16.
type CompanyFace struct {
	Competitor, Emotion uint8
	Performance         int16
	ChallengeFlags      uint32 // 0x04 CompanyFlags (increased/decreasedPerformance 1<<4 / 1<<5)
}

// parseCompanyFaces returns every company slot's CompanyFace.
func parseCompanyFaces(gs []byte) []CompanyFace {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	out := make([]CompanyFace, companyMaxCount)
	for i := range out {
		o := companiesGameStateOffset + i*size
		out[i] = CompanyFace{gs[o+0x18], gs[o+0x19], int16(binary.LittleEndian.Uint16(gs[o+0x16:])), binary.LittleEndian.Uint32(gs[o+0x04:])}
	}
	return out
}

// parseUnlockedVehicles returns every company's unlockedVehicles bitset
// (S5::Company 0x34, uint32_t[7]: one bit per vehicle object slot).
func parseUnlockedVehicles(gs []byte) [][7]uint32 {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	out := make([][7]uint32, companyMaxCount)
	for c := range out {
		o := companiesGameStateOffset + c*size + 0x34
		for w := range 7 {
			out[c][w] = binary.LittleEndian.Uint32(gs[o+4*w:])
		}
	}
	return out
}

// playerCompaniesOffset is GameState general.playerCompanies (uint8[2]; from
// the compiler). playerCompanies[0] is CompanyManager::getControllingId.
const playerCompaniesOffset = 0x24

// parseControllingCompany returns playerCompanies[0].
func parseControllingCompany(gs []byte) uint8 {
	if len(gs) <= playerCompaniesOffset {
		return 0
	}
	return gs[playerCompaniesOffset]
}

// parseCompanyColours is CompanyManager::updateColours: every company slot's
// mainColours.primary, then neutral = grey. nil if the chunk is too short.
func parseCompanyColours(gs []byte) []uint8 {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	c := make([]uint8, companyMaxCount+1)
	for i := range companyMaxCount {
		c[i] = gs[companiesGameStateOffset+i*size+companyMainColoursOffset]
	}
	c[companyNeutral] = colourGrey
	return c
}

// User strings (StringManager): StringIds 0x8000.. name entries of the
// GameState's userStrings table — 2048 × 32 bytes at 0x432A44 (GameStateType2:
// 0x430E24), offsets from the compiler. Town names live here.
const (
	userStringsOffset      = 0x432A44
	userStringsOffsetType2 = 0x430E24
	userStringsCount       = 2048
	userStringSize         = 32
	// UserStringsStart is StringManager::kUserStringsStart.
	UserStringsStart = 0x8000
)

// parseUserStrings returns the user strings (NUL-terminated, Latin-1 as the
// game stores them). nil if the chunk is too short.
func parseUserStrings(gs []byte) []string {
	off := userStringsOffset
	if len(gs) == GameStateSizeType2 {
		off = userStringsOffsetType2
	}
	if len(gs) < off+userStringsCount*userStringSize {
		return nil
	}
	out := make([]string, userStringsCount)
	for i := range out {
		rec := gs[off+i*userStringSize : off+(i+1)*userStringSize]
		n := 0
		for n < len(rec) && rec[n] != 0 {
			n++
		}
		r := make([]rune, n)
		for k := 0; k < n; k++ {
			r[k] = rune(rec[k])
		}
		out[i] = string(r)
	}
	return out
}

// UserString resolves a user-string StringId (ok false for other ids).
func (sc *Scenario) UserString(id uint16) (string, bool) {
	if id < UserStringsStart || int(id-UserStringsStart) >= len(sc.UserStrings) {
		return "", false
	}
	return sc.UserStrings[id-UserStringsStart], true
}

// Stations in the GameState (S5/S5Station.h; from the compiler): stations at
// 0xC10C4 (GameStateType2: 0xBF4A4), 1024 records of 978 bytes, owner at 0x28.
const (
	stationsGameStateOffset      = 0xC10C4
	stationsGameStateOffsetType2 = 0xBF4A4
	stationMaxCount              = 1024
	stationRecordSize            = 978
	stationOwnerOffset           = 0x28
)

// parseStationOwners returns, per station id, its owner (Station::owner). nil if
// the chunk is too short.
func parseStationOwners(gs []byte) []uint8 {
	off := stationsGameStateOffset
	if len(gs) == GameStateSizeType2 {
		off = stationsGameStateOffsetType2
	}
	if len(gs) < off+stationMaxCount*stationRecordSize {
		return nil
	}
	o := make([]uint8, stationMaxCount)
	for i := range o {
		o[i] = gs[off+i*stationRecordSize+stationOwnerOffset]
	}
	return o
}

// StationName is a station record's name StringId and town (S5::Station 0x00
// and 0x2C) and its cargo stats.
type StationName struct {
	Name, Town uint16
	Cargo      [32]StationCargo
	X, Y, Z    int16 // Station x, y, z (0x02, 0x04, 0x06): its position (world units)
	// VisitedTypes is var_3B2: a bit per VehicleType that has called here.
	VisitedTypes uint8
	Flags        uint16     // Station flags (0x2A)
	Owner        uint8      // owner (0x28): the company
	LabelFrame   LabelFrame // labelFrame (0x08): the name plaque's box per zoom

	Var3B1 uint8 // var_3B1 (0x3B1)
	// The airport (aircraft, AW.b): airportRotation (0x3B3), airportStartPos
	// (0x3B4: its first tile and height, world units) and
	// airportMovementOccupiedEdges (0x3BA: a bit per movement edge in use).
	AirportRotation      uint8
	AirportStartPos      [3]int16
	AirportOccupiedEdges uint32
	Var3B0               uint8  // var_3B0 (0x3B0)
	TileSize             uint16 // stationTileSize (0x1CE)
	NoTilesTimeout       uint8  // noTilesTimeout (0x29): days the station has had no tiles
	// Tiles is stationTiles (0x1D0): the first TileSize are the station's
	// tiles (z also holds the rotation: heightFloor it).
	Tiles [80][3]int16
}

// StationCargo is S5 StationCargoStats (13 bytes).
type StationCargo struct {
	Quantity       uint16
	Origin         uint16 // station the cargo came from
	Flags          uint8  // bit 0: acceptedForConsumer (isAccepted), bit 3: flag3
	Age            uint8
	Rating         uint8
	EnrouteAge     uint8
	VehicleSpeed   int16 // max speed of the vehicle that brought it
	VehicleAge     uint8
	IndustryID     uint8
	DensityPerTile uint8 // cargo visible per station tile (paintStationCargo)
}

// parseStationNames returns, per station id, its name and town. nil if short.
func parseStationNames(gs []byte) []StationName {
	off := stationsGameStateOffset
	if len(gs) == GameStateSizeType2 {
		off = stationsGameStateOffsetType2
	}
	if len(gs) < off+stationMaxCount*stationRecordSize {
		return nil
	}
	out := make([]StationName, stationMaxCount)
	for i := range out {
		rec := gs[off+i*stationRecordSize:]
		out[i] = StationName{Name: binary.LittleEndian.Uint16(rec[0x00:]), Town: binary.LittleEndian.Uint16(rec[0x2C:]),
			X: int16(binary.LittleEndian.Uint16(rec[0x02:])), Y: int16(binary.LittleEndian.Uint16(rec[0x04:])), Z: int16(binary.LittleEndian.Uint16(rec[0x06:])),
			VisitedTypes: rec[0x3B2], Flags: binary.LittleEndian.Uint16(rec[0x2A:]), Owner: rec[0x28], Var3B1: rec[0x3B1], Var3B0: rec[0x3B0],
			LabelFrame: ParseLabelFrame(rec[0x08:]),
			TileSize:   binary.LittleEndian.Uint16(rec[0x1CE:]), NoTilesTimeout: rec[0x29], AirportRotation: rec[0x3B3],
			AirportStartPos:      [3]int16{int16(binary.LittleEndian.Uint16(rec[0x3B4:])), int16(binary.LittleEndian.Uint16(rec[0x3B6:])), int16(binary.LittleEndian.Uint16(rec[0x3B8:]))},
			AirportOccupiedEdges: binary.LittleEndian.Uint32(rec[0x3BA:])}
		for k := range out[i].Tiles {
			for c := range 3 {
				out[i].Tiles[k][c] = int16(binary.LittleEndian.Uint16(rec[0x1D0+6*k+2*c:]))
			}
		}
		// cargoStats: 32 × 13 bytes at 0x2E (quantity @0, flags @4; from the compiler).
		for c := range out[i].Cargo {
			e := rec[0x2E+c*13:]
			out[i].Cargo[c] = StationCargo{Quantity: binary.LittleEndian.Uint16(e[0:]), Origin: binary.LittleEndian.Uint16(e[2:]),
				Flags: e[4], Age: e[5], Rating: e[6], EnrouteAge: e[7], VehicleSpeed: int16(binary.LittleEndian.Uint16(e[8:])),
				VehicleAge: e[10], IndustryID: e[11], DensityPerTile: e[12]}
		}
	}
	return out
}

// parseIndustryObjectIDs returns, per industry index, its industry object id
// (what IndustryElement.IndustryID refers to). nil if the chunk is too short.
func parseIndustryObjectIDs(gs []byte) []uint8 {
	off := industriesGameStateOffset
	if len(gs) == GameStateSizeType2 {
		off = industriesGameStateOffsetType2
	}
	if len(gs) < off+industryMaxCount*industryRecordSize {
		return nil
	}
	ids := make([]uint8, industryMaxCount)
	for i := range ids {
		ids[i] = gs[off+i*industryRecordSize+industryObjectIDOffset]
	}
	return ids
}

// CompanyEconomy is what the economy reads and writes of a company (S5
// Company / CompanyType2: the same offsets in both, from the compiler).
type CompanyEconomy struct {
	Cash                     int64         // 0x08 currency48_t
	CurrentLoan              int32         // 0x0E
	AvailableVehicles        uint16        // 0x50: a bit per VehicleType the company can buy
	Expenditures             [16][17]int32 // 0x58 [month][ExpenditureType]
	Var4A0                   uint32        // 0x4A0 cargo types delivered (bits)
	TransportTypeCount       [6]uint16     // 0x8BA4 vehicles by VehicleType (recalculateTransportCounts)
	Var49C                   uint32        // 0x49C cargo types loaded (bits)
	CargoUnitsTotalDelivered uint32        // 0x85F8
	CargoDelivered           [32]uint32    // 0x8BCE
	CargoUnitsTotalDistance  uint32        // 0x8C50
	CompanyValueHistory      [120]int64    // 0x88CE currency48_t
	VehicleProfit            int64         // 0x8B9E currency48_t
	ChallengeProgress        uint8         // 0x8C4E
	CurrentRating            uint8         // 0x25BF
	NumExpenditureYears      uint8         // 0x57
	HistorySize              uint16        // 0x88CC
	NumMonthsInTheRed        uint8         // 0x8C4F
	CargoUnitsDistanceHist   [120]uint32   // 0x8C54
	CargoUnitsDeliveredHist  [120]uint32   // 0x85FC
	PerformanceIndexHist     [120]int16    // 0x87DC
}

// parseCompanyEconomy reads every company slot's CompanyEconomy.
func parseCompanyEconomy(gs []byte) []CompanyEconomy {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	out := make([]CompanyEconomy, companyMaxCount)
	for i := range out {
		r := gs[companiesGameStateOffset+i*size:]
		c := &out[i]
		raw := uint64(binary.LittleEndian.Uint32(r[0x08:])) | uint64(binary.LittleEndian.Uint16(r[0x0C:]))<<32
		c.Cash = int64(raw<<16) >> 16 // sign-extend the 48 bits
		c.CurrentLoan = int32(binary.LittleEndian.Uint32(r[0x0E:]))
		c.AvailableVehicles = binary.LittleEndian.Uint16(r[0x50:])
		for k := range c.TransportTypeCount {
			c.TransportTypeCount[k] = binary.LittleEndian.Uint16(r[0x8BA4+2*k:])
		}
		for mo := range c.Expenditures {
			for t := range c.Expenditures[mo] {
				c.Expenditures[mo][t] = int32(binary.LittleEndian.Uint32(r[0x58+4*(mo*17+t):]))
			}
		}
		c.Var4A0 = binary.LittleEndian.Uint32(r[0x4A0:])
		c.Var49C = binary.LittleEndian.Uint32(r[0x49C:])
		c.CargoUnitsTotalDelivered = binary.LittleEndian.Uint32(r[0x85F8:])
		for k := range c.CargoDelivered {
			c.CargoDelivered[k] = binary.LittleEndian.Uint32(r[0x8BCE+4*k:])
		}
		c.CargoUnitsTotalDistance = binary.LittleEndian.Uint32(r[0x8C50:])
		for k := range 120 {
			c.CompanyValueHistory[k] = currency48(r[0x88CE+6*k:])
			c.CargoUnitsDistanceHist[k] = binary.LittleEndian.Uint32(r[0x8C54+4*k:])
			c.CargoUnitsDeliveredHist[k] = binary.LittleEndian.Uint32(r[0x85FC+4*k:])
			c.PerformanceIndexHist[k] = int16(binary.LittleEndian.Uint16(r[0x87DC+2*k:]))
		}
		c.VehicleProfit = currency48(r[0x8B9E:])
		c.ChallengeProgress = r[0x8C4E]
		c.CurrentRating = r[0x25BF]
		c.NumExpenditureYears = r[0x57]
		c.HistorySize = binary.LittleEndian.Uint16(r[0x88CC:])
		c.NumMonthsInTheRed = r[0x8C4F]
	}
	return out
}

// parseIndustries reads the non-empty industry slots (Industry::empty: name
// is StringIds::null).
func parseIndustries(gs []byte) []Industry {
	off := industriesGameStateOffset
	if len(gs) == GameStateSizeType2 {
		off = industriesGameStateOffsetType2
	}
	if len(gs) < off+industryMaxCount*industryRecordSize {
		return nil
	}
	var out []Industry
	for i := range industryMaxCount {
		r := gs[off+i*industryRecordSize:]
		if binary.LittleEndian.Uint16(r[0:]) == 0xFFFF {
			continue
		}
		ind := Industry{ID: uint8(i), ObjectID: r[0x10], Town: binary.LittleEndian.Uint16(r[0xD5:]),
			HistoryMinProduction: binary.LittleEndian.Uint32(r[0x38B:])}
		for k := range 3 {
			ind.ReceivedMonthlyTotal[k] = binary.LittleEndian.Uint16(r[0x18D+2*k:])
			ind.ReceivedDailyTotal[k] = binary.LittleEndian.Uint16(r[0x199+2*k:])
		}
		u16 := func(o int) uint16 { return binary.LittleEndian.Uint16(r[o:]) }
		ind.Flags, ind.UnderConstruction, ind.ProductionRate = u16(0x06), r[0x11], r[0xDF]
		for k := range 2 {
			for j := range 4 {
				ind.ProducedStations[k][j] = u16(0x161 + 8*k + 2*j)
			}
			ind.DailyProductionTarget[k] = u16(0x179 + 2*k)
			ind.DailyProduction[k] = u16(0x17D + 2*k)
			ind.OutputBuffer[k] = u16(0x181 + 2*k)
			ind.ProducedMonthlyTotal[k] = u16(0x185 + 2*k)
			ind.DeliveredMonthlyTotal[k] = u16(0x19F + 2*k)
			for j := range 4 {
				ind.ProducedRatings[k][j] = r[0x171+4*k+j]
			}
		}
		ind.X, ind.Y = int16(u16(0x02)), int16(u16(0x04))
		ind.Name, ind.FoundingYear, ind.Owner = u16(0x00), u16(0x12), r[0xE0]
		ind.TileLoop = [2]int16{int16(u16(0xD7)), int16(u16(0xD9))}
		ind.NumFarmTiles, ind.NumIdleFarmTiles = int16(u16(0xDB)), int16(u16(0xDD))
		for k := 0; k < int(r[0x14]) && k < 32; k++ {
			ind.Tiles = append(ind.Tiles, [3]int16{int16(u16(0x15 + 6*k)), int16(u16(0x17 + 6*k)), int16(u16(0x19 + 6*k))})
		}
		ind.Prng = [2]uint32{binary.LittleEndian.Uint32(r[0x08:]), binary.LittleEndian.Uint32(r[0x0C:])}
		for k := range 3 {
			ind.ReceivedPrevMonth[k] = u16(0x193 + 2*k)
		}
		for k := range 2 {
			ind.ProducedPrevMonth[k] = u16(0x189 + 2*k)
			ind.DeliveredPrevMonth[k] = u16(0x1A3 + 2*k)
			ind.PercentTransported[k] = r[0x1A7+k]
			ind.HistorySize[k] = r[0x1A9+k]
			copy(ind.History[k][:], r[0x1AB+0xF0*k:])
		}
		for k := range ind.StationsInRange {
			ind.StationsInRange[k] = binary.LittleEndian.Uint32(r[0xE1+4*k:])
		}
		out = append(out, ind)
	}
	return out
}

// currency48 reads a little-endian currency48_t (sign-extended).
func currency48(b []byte) int64 {
	raw := uint64(binary.LittleEndian.Uint32(b)) | uint64(binary.LittleEndian.Uint16(b[4:]))<<32
	return int64(raw<<16) >> 16
}

// parseGeneralState reads GameState's general part (everything before the
// companies, offsets from the compiler): it heads a saved game's one GameState
// chunk and is a scenario's whole first GameState chunk ("up to just before
// companies", S5.cpp importSave).
func parseGeneralState(sc *Scenario, gs []byte) {
	if len(gs) < 0x434 {
		return
	}
	sc.RoadAnyTypeCompatible = binary.LittleEndian.Uint32(gs[0x1A4:])
	sc.OrderTableLength = binary.LittleEndian.Uint32(gs[0x1A0:]) // GameState orderTableLength
	sc.Rng = [2]uint32{binary.LittleEndian.Uint32(gs[0:]), binary.LittleEndian.Uint32(gs[4:])}
	sc.ScenarioTicks = binary.LittleEndian.Uint32(gs[0x146:])
	sc.UnkRng = [2]uint32{binary.LittleEndian.Uint32(gs[8:]), binary.LittleEndian.Uint32(gs[12:])}
	sc.RoadUsableByAll = binary.LittleEndian.Uint32(gs[0x1A8:])
	sc.CurrentDay = ParseCurrentDay(gs)
	sc.DayCounter = binary.LittleEndian.Uint16(gs[0x18:]) // general.dayCounter
	sc.CurrentMonth = gs[0x1C]                            // general.currentMonth
	sc.CurrentDayOfMonth = gs[0x1D]                       // general.currentDayOfMonth
	sc.GameStateFlags = binary.LittleEndian.Uint32(gs[0x10:])
	sc.IndustryManagerFlags = gs[0x42F]
	sc.NumberOfIndustries = gs[0x415]         // numberOfIndustries: 0 low, 1 medium, 2 high
	if len(gs) > GeneralCurrentYearOffset+1 { // general.currentYear (from the compiler: 0x1A)
		sc.CurrentYear = binary.LittleEndian.Uint16(gs[GeneralCurrentYearOffset:])
	}
	// general.currencyMultiplicationFactor[32] (0x46, from the compiler): the
	// inflation-adjusted price of each cost index.
	if len(gs) >= 0x46+32*4 {
		for i := range sc.CurrencyMultiplicationFactor {
			sc.CurrencyMultiplicationFactor[i] = binary.LittleEndian.Uint32(gs[0x46+4*i:])
		}
	}
	sc.ControllingCompany = parseControllingCompany(gs)
	sc.PlayerCompanies = [2]uint8{gs[playerCompaniesOffset], gs[playerCompaniesOffset+1]}
	sc.LoanInterestRate = gs[0x1AE]                              // loanInterestRate
	sc.MaxLoanSize = binary.LittleEndian.Uint16(gs[0x402:])      // maxLoanSize
	sc.StartingLoanSize = binary.LittleEndian.Uint16(gs[0x400:]) // startingLoanSize
	sc.SeaLevel = uint8(binary.LittleEndian.Uint16(gs[0x19A:]))  // seaLevel
	sc.CurrentSnowLine, sc.CurrentSeason = gs[0x19C], gs[0x19D]
	sc.LastLandOption = gs[0x19E]                              // lastLandOption
	sc.PreferredAI = [3]uint8{gs[0x3FD], gs[0x3FE], gs[0x3FF]} // preferredAIIntelligence / Aggressiveness / Competitiveness
	sc.DefaultLevelCrossingType = gs[0x1AC]                    // currentDefaultLevelCrossingType
	sc.ForbiddenVehiclesPlayers = binary.LittleEndian.Uint16(gs[0x430:])
	sc.ForbiddenVehiclesCompetitors = binary.LittleEndian.Uint16(gs[0x432:])
	if len(gs) > 0x196 { // general.trafficHandedness (from the compiler)
		sc.TrafficHandedness = gs[0x196]
	}
	if len(gs) > 0x1AD { // S5GameState lastRailroadOption 0x192, lastRoadOption 0x193, lastTrackTypeOption 0x1AD
		sc.LastRailroadOption, sc.LastRoadOption, sc.LastTrackTypeOption = gs[0x192], gs[0x193], gs[0x1AD]
		sc.LastBuildVehiclesOption = gs[0x414] // GameState::lastBuildVehiclesOption
		sc.PickupDirection = gs[0x198]         // GameState::pickupDirection
	}
}

// CompanyColourScheme is a company's colours (S5::Company): mainColours
// (0x1A: primary, secondary), vehicleColours (0x1C: a scheme per vehicle
// colour type 1–10) and customVehicleColoursSet (0x30: which types use theirs).
type CompanyColourScheme struct {
	Main      [2]uint8
	Vehicle   [10][2]uint8
	CustomSet uint32
}

// parseCompanyColourSchemes reads every company slot's colour scheme.
func parseCompanyColourSchemes(gs []byte) []CompanyColourScheme {
	size := companyRecordSize
	if len(gs) == GameStateSizeType2 {
		size = companyRecordSizeType2
	}
	if len(gs) < companiesGameStateOffset+companyMaxCount*size {
		return nil
	}
	out := make([]CompanyColourScheme, companyMaxCount)
	for i := range out {
		o := companiesGameStateOffset + i*size
		c := &out[i]
		c.Main = [2]uint8{gs[o+0x1A], gs[o+0x1B]}
		for k := range c.Vehicle {
			c.Vehicle[k] = [2]uint8{gs[o+0x1C+2*k], gs[o+0x1C+2*k+1]}
		}
		c.CustomSet = binary.LittleEndian.Uint32(gs[o+0x30:])
	}
	return out
}

// parseAnimations reads n World::Animation records (6 bytes: baseZ, type,
// x, y) from data.
func parseAnimations(data []byte, n int) []Animation {
	n = min(n, len(data)/6)
	out := make([]Animation, n)
	for i := range out {
		r := data[6*i:]
		out[i] = Animation{BaseZ: r[0], Type: r[1], X: int16(binary.LittleEndian.Uint16(r[2:])), Y: int16(binary.LittleEndian.Uint16(r[4:]))}
	}
	return out
}
