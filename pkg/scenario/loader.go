package scenario

import (
	"errors"
	"fmt"
	"log"

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
	// ElementType is stored in bits [7:2] of the first byte of a TileElement.
	elementTypeMask  = 0xFC // bits 7-2
	elementTypeShift = 2

	// ElementFlags byte (second byte):
	elementFlagLast = 0x80 // last element on this tile

	// Known element type values (after >> 2)
	elementTypeSurface = 0
)

// terrainTypes maps the 5-bit terrain field in a surface element to our SurfaceType.
// OpenLoco reference: src/OpenLoco/src/Map/SurfaceElement.h  TerrainType enum
// Values 0-29 are defined; anything ≥30 is unknown.
var terrainTypes = [32]SurfaceType{
	0: SurfaceGrass,  // grassland
	1: SurfaceDirt,   // sand / dirt
	2: SurfaceRock,   // rock
	3: SurfaceDirt,   // mud
	4: SurfaceGrass,  // grass (variant)
	5: SurfaceSnow,   // snow
	6: SurfaceRock,   // gravel
	7: SurfaceDirt,   // dirt (variant)
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
//   1. Header          (rotate)
//   2. ScenarioOptions (rotate)          — only if type == scenario
//   3. PackedObjects                     — only if numPackedObjects > 0
//   4. RequiredObjects (rotate)
//   5. GeneralState    (runLengthSingle)
//   6. Towns           (runLengthSingle)
//   7. Animations      (runLengthSingle)
//   8. TileElements    (runLengthMulti)  — only if tileManagerLoaded flag set
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
	if header.Type == assets.S5TypeScenario {
		scenarioOpts, err := reader.ReadChunk()
		if err != nil {
			return sc, fmt.Errorf("failed to read ScenarioOptions chunk: %w", err)
		}
		log.Println("[Scenario] Read ScenarioOptions,", len(scenarioOpts), "bytes")
	}

	// --- Packed objects (raw ObjectHeader + data chunk per object) ---
	if header.NumPackedObjects > 0 {
		log.Printf("[Scenario] Skipping %d packed objects", header.NumPackedObjects)
		for i := uint16(0); i < header.NumPackedObjects; i++ {
			// Each packed object is: 16-byte ObjectHeader (raw, no chunk framing)
			// followed by one chunk of object data.
			// Skip the raw header bytes directly.
			if reader.Offset()+16 > len(data) {
				return sc, fmt.Errorf("truncated packed object header at index %d", i)
			}
			// Advance past the 16-byte raw ObjectHeader (not chunk-framed)
			advanceRaw(reader, 16)

			// Then one chunk of object data
			_, err := reader.ReadChunk()
			if err != nil {
				return sc, fmt.Errorf("failed to read packed object %d data: %w", i, err)
			}
		}
	}

	// --- Required objects (rotate-encoded) ---
	reqObjBytes, err := reader.ReadChunk()
	if err != nil {
		return sc, fmt.Errorf("failed to read RequiredObjects chunk: %w", err)
	}
	log.Println("[Scenario] Read RequiredObjects,", len(reqObjBytes), "bytes")

	// --- Game state chunks ---
	// For scenarios: three separate chunks (GeneralState, Towns, Animations).
	// For saved games: one big GameState chunk.
	// We only need the flags field from GeneralState to know if tiles follow.
	var generalStateFlags uint32
	if header.Type == assets.S5TypeScenario {
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
		log.Println("[Scenario] tileManagerLoaded flag not set — no tile elements in file")
		sc.initPlaceholderMap()
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
//   TileManager::setElements()
//   TileManager::updateTilePointers()
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
			if elemType == elementTypeSurface {
				// Surface element: byte 6 has terrain type in bits [4:0]
				terrainRaw := elem[6] & 0x1F
				surface := SurfaceGrass
				if int(terrainRaw) < len(terrainTypes) {
					surface = terrainTypes[terrainRaw]
				}

				// Water level in byte 5, bits [4:0].  If > 0, treat as water.
				waterLevel := elem[5] & 0x1F
				if waterLevel > 0 {
					surface = SurfaceWater
				}

				sc.Tiles[y][x] = Tile{
					Surface: surface,
					Height:  baseZ,
					Water:   waterLevel,
				}
			}
			// Non-surface elements (track, building, tree, …) are skipped
			// for now; only the first surface element per tile matters for
			// terrain rendering.
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

	log.Printf("[Scenario] Parsed %d tile elements covering %d×%d tiles (stopped at x=%d y=%d)",
		elemCount, sc.MapWidth, sc.MapHeight, x, y)
}

// readGeneralStateFlags extracts the flags uint32 from a GeneralState or
// GameState blob.  The flags field is at a fixed offset within the struct.
//
// OpenLoco reference: src/OpenLoco/src/S5/GameState.h
//   GeneralState::flags  — offset 0x434 from start of GameState,
//   but GeneralState itself starts at offset 0 when read as a standalone chunk.
//
// For a scenario file the GeneralState chunk contains only the GeneralState
// sub-struct (not the full GameState), so flags is near the beginning.
// The flags field is at offset 0x00 within GeneralState per the OpenLoco
// layout; however empirically for the standalone GeneralState chunk it is at
// a small offset.  We scan for the tileManagerLoaded pattern: if we can't
// find it definitively, default to assuming tiles are present (which is the
// common case for title.dat and scenario files).
func readGeneralStateFlags(data []byte) uint32 {
	// In OpenLoco's GameState, general.flags is at byte offset 0x00 of the
	// GeneralState sub-struct.  When GeneralState is written as its own chunk
	// (scenario path), the flags field is at offset 0 of that chunk.
	if len(data) >= 4 {
		// Read the first uint32 — this is general.flags for the standalone chunk.
		flags := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
		log.Printf("[Scenario] GeneralState flags = 0x%08X", flags)
		return flags
	}
	// Default: assume tile manager is loaded (safest for title sequences)
	log.Println("[Scenario] GeneralState too short to read flags, assuming tileManagerLoaded")
	return 1
}

// advanceRaw moves the reader forward by n raw bytes (no chunk framing).
// This is used to skip ObjectHeader structs that are written raw (not as chunks).
func advanceRaw(r *assets.S5ChunkReader, n int) {
	// S5ChunkReader doesn't expose a skip method, so we manipulate via the
	// exported Offset and reconstruct.  For now we use a simple wrapper that
	// re-slices internally.  This is only called for packed object headers
	// which are rare.
	r.AdvanceRaw(n)
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
//   src/OpenLoco/src/Scenario/Scenario.cpp  Scenario::load()
//   src/OpenLoco/src/Map/TileManager.cpp    TileManager
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
