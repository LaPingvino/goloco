package scenario

import (
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
	sc.TreeObjectOrder = parseTreeObjectOrder(reqObjBytes)
	sc.BuildingObjectOrder = parseBuildingObjectOrder(reqObjBytes)
	sc.WallObjectOrder = parseWallObjectOrder(reqObjBytes)
	sc.TrainSignalObjectOrder = parseTrainSignalObjectOrder(reqObjBytes)
	sc.LevelCrossingObjectOrder = parseLevelCrossingObjectOrder(reqObjBytes)
	sc.BridgeObjectOrder = parseBridgeObjectOrder(reqObjBytes)
	sc.TrainStationObjectOrder = parseTrainStationObjectOrder(reqObjBytes)
	sc.RoadStationObjectOrder = parseRoadStationObjectOrder(reqObjBytes)
	sc.TrackObjectOrder = parseTrackObjectOrder(reqObjBytes)
	sc.RoadObjectOrder = parseRoadObjectOrder(reqObjBytes)
	sc.VehicleObjectOrder = parseVehicleObjectOrder(reqObjBytes)

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

		// Towns
		towns, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read Towns chunk: %w", err)
		}
		log.Println("[Scenario] Read Towns,", len(towns), "bytes")

		// Animations
		animations, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read Animations chunk: %w", err)
		}
		log.Println("[Scenario] Read Animations,", len(animations), "bytes")
	} else {
		gameState, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read GameState chunk: %w", err)
		}
		log.Println("[Scenario] Read GameState,", len(gameState), "bytes")
		generalStateFlags = readGeneralStateFlags(gameState)

		// Parse vehicle entities from the entity table embedded in the GameState.
		// This is only present in .SV5 saved-game files; ParseVehicleEntities
		// returns nil if the chunk is smaller than EntityTableOffset + table size.
		// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
		sc.Entities = ParseVehicleEntities(gameState)
		if sc.Entities != nil {
			log.Printf("[Scenario] Parsed %d vehicle entities (Head+Body) from GameState", len(sc.Entities))
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
			sc.initGeneratedLandscape()
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
					TerrainIndex: terrainRaw,
					GrowthStage:  growthStage,
					Variation:    variation,
					IsIndustrial: typeByte&0x80 != 0,
					Slope:        slopeByte,
					Trees:        existing.Trees,
					Buildings:    existing.Buildings,
					Walls:        existing.Walls,
					Tracks:       existing.Tracks,
					Roads:        existing.Roads,
					Stations:     existing.Stations,
					Signals:      existing.Signals,
					Industries:   existing.Industries,
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
				sc.Tiles[y][x].Trees = append(sc.Tiles[y][x].Trees, te)
				treeCount++

			case elementTypeBuilding:
				be := BuildingElement{
					ObjectID:      elem[4],
					Rotation:      typeByte & 0x03,
					SequenceIndex: elem[5] & 0x03,
					Variation:     ((elem[7] & 0x07) << 2) | ((elem[6] >> 6) & 0x03),
					Colour:        (elem[7] >> 3) & 0x1F,
					IsConstructed: typeByte&0x80 != 0,
					BaseZ:         baseZ,
					ClearZ:        elem[3],
				}
				sc.Tiles[y][x].Buildings = append(sc.Tiles[y][x].Buildings, be)
				buildingCount++

			case elementTypeWall:
				// WallElement: byte layout from WallElement.h
				we := WallElement{
					WallObjectID:  elem[4],
					Rotation:      typeByte & 0x03,
					EdgeSlope:     (typeByte >> 6) & 0x03,
					PrimaryColour: elem[6] & 0x1F,
					BaseZ:         baseZ,
					ClearZ:        elem[3],
				}
				sc.Tiles[y][x].Walls = append(sc.Tiles[y][x].Walls, we)
				wallCount++

			case elementTypeTrack:
				// TrackElement: byte layout from TrackElement.h
				// bridge(): elem[6] >> 5 (bits [7:5]) — OpenLoco TrackElement.h
				te := TrackElement{
					TrackID:       elem[4] & 0x3F,
					HasBridge:     elem[4]&0x80 != 0,
					BridgeID:      elem[6] >> 5,
					SequenceIndex: elem[5] & 0x0F,
					TrackObjectID: (elem[5] >> 4) & 0x0F,
					Rotation:      typeByte & 0x03,
					HasSignal:     typeByte&0x40 != 0,
					HasStation:    typeByte&0x80 != 0,
					Owner:         elem[7] & 0x0F,
					Mods:          (elem[7] >> 4) & 0x0F,
					BaseZ:         baseZ,
					ClearZ:        elem[3],
				}
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
					Rotation:              typeByte & 0x03,
					Owner:                 elem[7] & 0x0F,
					Mods:                  (elem[7] >> 6) & 0x03,
					BaseZ:                 baseZ,
					ClearZ:                elem[3],
				}
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
				}
				sc.Tiles[y][x].Stations = append(sc.Tiles[y][x].Stations, se)
				stationCount++

			case elementTypeSignal:
				// SignalElement: byte layout from SignalElement.h
				sig := SignalElement{
					SignalObjectID: elem[4] & 0x0F,
					Rotation:       typeByte & 0x03,
					HasLeftSignal:  elem[4]&0x80 != 0,
					LeftFrame:      elem[5] & 0x0F,
					HasRightSignal: elem[6]&0x80 != 0,
					RightFrame:     elem[7] & 0x0F,
					BaseZ:          baseZ,
					ClearZ:         elem[3],
				}
				sc.Tiles[y][x].Signals = append(sc.Tiles[y][x].Signals, sig)
				signalCount++

			case elementTypeIndustry:
				// IndustryElement: byte layout from IndustryElement.h
				sectionWord := uint16(elem[6]) | uint16(elem[7])<<8
				ie := IndustryElement{
					IndustryID:    elem[4],
					BuildingType:  uint8((sectionWord >> 6) & 0x1F),
					Rotation:      typeByte & 0x03,
					IsConstructed: typeByte&0x80 != 0,
					BaseZ:         baseZ,
					ClearZ:        elem[3],
				}
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
	// Objective type
	if len(data) >= 0x418D {
		opts.ObjectiveType = data[0x418C]
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
	objectHeaderSize = 16
	objectTypeMask   = 0x3F
	objectTypeLand         = 6  // ObjectType::land
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

	// Wall slots: after land(32)+TownNames(1)+Cargo(32) = 171+32+1+32 = 236
	wallSlotStart = 236
	wallSlotCount = 32

	// TrackSignal slots: Wall(32) = 236+32 = 268
	trackSignalSlotStart = 268
	trackSignalSlotCount = 16
	objectTypeTrackSignal = 10 // ObjectType::trackSignal

	// LevelCrossing slots: TrackSignal(16) = 268+16 = 284
	levelCrossingSlotStart = 284
	levelCrossingSlotCount = 4
	objectTypeLevelCrossing = 11 // ObjectType::levelCrossing

	// Bridge slots: TrackSignal(16)+LevelCrossing(4)+StreetLight(1)+Tunnel(16) = 268+16+4+1+16 = 305
	bridgeSlotStart = 305
	bridgeSlotCount = 8
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
	roadSlotStart = 365
	roadSlotCount = 8

	// Vehicle slots: Road(8)+Airport(8)+Dock(8) = 365+8+8+8 = 389
	vehicleSlotStart = 389
	vehicleSlotCount = 224
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
