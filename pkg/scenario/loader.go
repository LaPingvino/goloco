package scenario

import (
	"errors"
	"fmt"
	"log"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// DefaultMapSize is typical Locomotion map size
const DefaultMapSize = 384

// LoadScenarioData loads and returns scenario data (for main.go use)
func LoadScenarioData(filePath string) (*Scenario, error) {
	log.Println("[Scenario] Loading scenario from:", filePath)

	data, err := assets.LoadDatFile(filePath)
	if err != nil {
		log.Println("[Scenario] ERROR reading file:", err)
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}
	log.Println("[Scenario] File read successfully, size:", len(data), "bytes")

	scenario, err := ParseScenario(data, filePath)
	if err != nil {
		return nil, err
	}

	log.Printf("[Scenario] Scenario loaded: %dx%d map", scenario.MapWidth, scenario.MapHeight)
	return scenario, nil
}

// ParseScenario parses scenario data from memory
func ParseScenario(data []byte, filePath string) (*Scenario, error) {
	log.Println("[Scenario] Parsing scenario data, size:", len(data))

	if len(data) < 32 {
		return nil, errors.New("scenario file too small")
	}

	header, err := assets.ParseS5Header(data)
	if err != nil {
		log.Println("[Scenario] ERROR parsing S5 header:", err)
		return nil, fmt.Errorf("failed to parse S5 header: %w", err)
	}
	log.Printf("[Scenario] S5 Header: Type=%d, Flags=0x%X, NumPackedObjects=%d, Version=%d, Magic=0x%X",
		header.Type, header.Flags, header.NumPackedObjects, header.Version, header.Magic)

	scenario := &Scenario{
		Header:    header,
		FilePath:  filePath,
		MapWidth:  DefaultMapSize,
		MapHeight: DefaultMapSize,
	}

	compressedData := data[32:]
	log.Println("[Scenario] Decompressing single payload from", len(compressedData), "bytes")

	decompressed, err := assets.DecompressS5(compressedData, 1699535)
	if err != nil {
		log.Printf("[Scenario] Failed to decompress payload: %v", err)
		return scenario, fmt.Errorf("failed to decompress S5 data: %w", err)
	}
	log.Println("[Scenario] Decompressed", len(decompressed), "bytes")

	scenario.parseTileData(decompressed)

	if scenario.Tiles == nil {
		log.Println("[Scenario] No tiles loaded, creating placeholder map")
		scenario.initPlaceholderMap()
	}

	log.Printf("[Scenario] Scenario loaded: %dx%d map", scenario.MapWidth, scenario.MapHeight)
	return scenario, nil
}

// parseTileData attempts to parse raw tile data from a decompressed S5
// payload. Stub — the current mapping (one byte per tile, surface
// inferred from height thresholds) is a heuristic placeholder and does
// not match the real Locomotion tile layout.
//
// OpenLoco reference: src/OpenLoco/src/Scenario/Scenario.cpp
//   Scenario::load(const fs::path&)  — full scenario deserialisation
// Also: src/OpenLoco/src/Map/TileManager.cpp
//   TileManager — tile array layout and access helpers
//
// In OpenLoco each map tile is a multi-byte TileBase entry (surface type,
// height nibble, building/track flags, etc.).  The offset into the
// decompressed payload and the per-tile struct size must be reverse-
// engineered from Scenario::load before this function can be correct.
func (sc *Scenario) parseTileData(data []byte) {
	sc.Tiles = make([][]Tile, sc.MapHeight)
	for y := 0; y < sc.MapHeight; y++ {
		sc.Tiles[y] = make([]Tile, sc.MapWidth)
		for x := 0; x < sc.MapWidth; x++ {
			offset := y*sc.MapWidth + x
			if offset < len(data) {
				sc.Tiles[y][x] = Tile{
					Surface: surfaceFromHeight(data[offset]),
					Height:  data[offset] / 16,
				}
			} else {
				sc.Tiles[y][x] = Tile{
					Surface: SurfaceGrass,
					Height:  uint8((x + y) % 4),
				}
			}
		}
	}
}

// surfaceFromHeight determines surface type based on height value
func surfaceFromHeight(h uint8) SurfaceType {
	if h < 10 {
		return SurfaceWater
	} else if h < 50 {
		return SurfaceGrass
	} else if h < 100 {
		return SurfaceDirt
	} else if h < 150 {
		return SurfaceSand
	} else {
		return SurfaceRock
	}
}

// initPlaceholderMap creates a placeholder map with simple terrain.
// This is a development fallback used when parseTileData produces no
// usable tiles (e.g. because the S5 tile layout is not yet correct).
// It should be removed once parseTileData is fully implemented.
//
// OpenLoco reference: see parseTileData comment above —
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

// GetTile returns tile at the given coordinates
func (sc *Scenario) GetTile(x, y int) *Tile {
	if x < 0 || x >= sc.MapWidth || y < 0 || y >= sc.MapHeight {
		return nil
	}
	if sc.Tiles == nil {
		return nil
	}
	return &sc.Tiles[y][x]
}

// IsTitleSequence returns true if this is a title sequence file
func (sc *Scenario) IsTitleSequence() bool {
	return (sc.Header.Flags & 0x4) != 0
}
