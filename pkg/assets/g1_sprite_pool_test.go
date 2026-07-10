package assets

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// TestG1SpritePoolDynamicLoading verifies that the dynamic sprite pool system works correctly
func TestG1SpritePoolDynamicLoading(t *testing.T) {
	// Create a minimal G1File
	g1 := &G1File{
		Header: G1Header{
			NumEntries: 10,
			TotalSize:  0,
		},
		Elements: make([]G1Element, 10),
	}
	g1.SetTotalSprites(10)

	// Create a minimal image table with 3 sprites
	// Format: [G1Header][G1Element32 array][pixel data]
	imageTable := []byte{
		// G1Header (8 bytes)
		3, 0, 0, 0, // numEntries = 3
		60, 0, 0, 0, // totalSize = 60 bytes (3 sprites * 20 bytes each)

		// G1Element32[0] - 16 bytes
		0, 0, 0, 0, // offset = 0
		10, 0, // width = 10
		10, 0, // height = 10
		0, 0, // xOffset = 0
		0, 0, // yOffset = 0
		0, 0, // flags = 0
		0, 0, // zoomOffset = 0

		// G1Element32[1] - 16 bytes
		20, 0, 0, 0, // offset = 20
		5, 0, // width = 5
		5, 0, // height = 5
		2, 0, // xOffset = 2
		3, 0, // yOffset = 3
		0, 0, // flags = 0
		0, 0, // zoomOffset = 0

		// G1Element32[2] - 16 bytes
		40, 0, 0, 0, // offset = 40
		8, 0, // width = 8
		8, 0, // height = 8
		1, 0, // xOffset = 1
		1, 0, // yOffset = 1
		0, 0, // flags = 0
		0, 0, // zoomOffset = 0
	}

	// Add pixel data (60 bytes total)
	pixelData := make([]byte, 60)
	for i := range pixelData {
		pixelData[i] = byte(i % 256)
	}
	imageTable = append(imageTable, pixelData...)

	// Load the image table
	result, err := g1.LoadImageTable(imageTable)
	if err != nil {
		t.Fatalf("LoadImageTable failed: %v", err)
	}

	// Verify result
	if result.ImageOffset != 10 {
		t.Errorf("Expected ImageOffset=10, got %d", result.ImageOffset)
	}

	expectedTableLength := uint32(8 + 3*16 + 60) // header + elements + data
	if result.TableLength != expectedTableLength {
		t.Errorf("Expected TableLength=%d, got %d", expectedTableLength, result.TableLength)
	}

	// Verify total sprite count increased
	if g1.GetTotalSprites() != 13 {
		t.Errorf("Expected totalSprites=13 (10+3), got %d", g1.GetTotalSprites())
	}

	// Verify elements were appended
	if len(g1.Elements) != 13 {
		t.Errorf("Expected 13 elements, got %d", len(g1.Elements))
	}

	// Verify first loaded sprite (index 10)
	sprite10 := &g1.Elements[10]
	if sprite10.Width != 10 || sprite10.Height != 10 {
		t.Errorf("Sprite 10: expected 10x10, got %dx%d", sprite10.Width, sprite10.Height)
	}
	if len(sprite10.Data) != 20 {
		t.Errorf("Sprite 10: expected 20 bytes of data, got %d", len(sprite10.Data))
	}

	// Verify second loaded sprite (index 11)
	sprite11 := &g1.Elements[11]
	if sprite11.Width != 5 || sprite11.Height != 5 {
		t.Errorf("Sprite 11: expected 5x5, got %dx%d", sprite11.Width, sprite11.Height)
	}
	if sprite11.XOffset != 2 || sprite11.YOffset != 3 {
		t.Errorf("Sprite 11: expected offset (2,3), got (%d,%d)", sprite11.XOffset, sprite11.YOffset)
	}

	// Verify third loaded sprite (index 12)
	sprite12 := &g1.Elements[12]
	if sprite12.Width != 8 || sprite12.Height != 8 {
		t.Errorf("Sprite 12: expected 8x8, got %dx%d", sprite12.Width, sprite12.Height)
	}
}

// TestG1SpritePoolMultipleLoads verifies multiple LoadImageTable calls work correctly
func TestG1SpritePoolMultipleLoads(t *testing.T) {
	g1 := &G1File{
		Header:   G1Header{NumEntries: 5, TotalSize: 0},
		Elements: make([]G1Element, 5),
	}
	g1.SetTotalSprites(5)

	// Create first image table with 2 sprites
	table1 := createMinimalImageTable(2, 10, 10)
	result1, err := g1.LoadImageTable(table1)
	if err != nil {
		t.Fatalf("First LoadImageTable failed: %v", err)
	}

	if result1.ImageOffset != 5 {
		t.Errorf("First load: expected offset=5, got %d", result1.ImageOffset)
	}
	if g1.GetTotalSprites() != 7 {
		t.Errorf("After first load: expected 7 sprites, got %d", g1.GetTotalSprites())
	}

	// Create second image table with 3 sprites
	table2 := createMinimalImageTable(3, 5, 5)
	result2, err := g1.LoadImageTable(table2)
	if err != nil {
		t.Fatalf("Second LoadImageTable failed: %v", err)
	}

	if result2.ImageOffset != 7 {
		t.Errorf("Second load: expected offset=7, got %d", result2.ImageOffset)
	}
	if g1.GetTotalSprites() != 10 {
		t.Errorf("After second load: expected 10 sprites, got %d", g1.GetTotalSprites())
	}

	// Verify all elements exist
	if len(g1.Elements) != 10 {
		t.Errorf("Expected 10 total elements, got %d", len(g1.Elements))
	}
}

// TestG1LoaderInterface verifies G1File implements the objects.G1Loader interface
func TestG1LoaderInterface(t *testing.T) {
	// Compile-time check that G1File implements G1Loader.
	var _ objects.G1Loader = &G1File{}
}

// TestSpriteDataNonEmpty verifies loaded sprites have realistic data
func TestSpriteDataNonEmpty(t *testing.T) {
	g1 := &G1File{
		Elements: make([]G1Element, 0),
	}
	g1.SetTotalSprites(0)

	// Create image table with non-empty pixel data
	numSprites := uint32(3)
	pixelsPerSprite := 100 // 10x10
	totalDataSize := numSprites * uint32(pixelsPerSprite)

	table := []byte{
		// Header
		byte(numSprites), 0, 0, 0,
		byte(totalDataSize), byte(totalDataSize >> 8), 0, 0,
	}

	// Add elements
	for i := uint32(0); i < numSprites; i++ {
		offset := i * uint32(pixelsPerSprite)
		table = append(table,
			byte(offset), byte(offset>>8), byte(offset>>16), byte(offset>>24), // offset
			10, 0, // width
			10, 0, // height
			0, 0, 0, 0, 0, 0, 0, 0, // offsets and flags
		)
	}

	// Add pixel data with recognizable pattern
	for i := uint32(0); i < totalDataSize; i++ {
		table = append(table, byte(i%256))
	}

	result, err := g1.LoadImageTable(table)
	if err != nil {
		t.Fatalf("LoadImageTable failed: %v", err)
	}

	// Verify all sprites have data
	for i := uint32(0); i < numSprites; i++ {
		spriteIdx := result.ImageOffset + i
		sprite := &g1.Elements[spriteIdx]

		if len(sprite.Data) == 0 {
			t.Errorf("Sprite %d has empty data", spriteIdx)
		}

		// Check data is not all zeros
		allZero := true
		for _, b := range sprite.Data {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero && len(sprite.Data) > 0 {
			t.Errorf("Sprite %d has all-zero data (should have pattern)", spriteIdx)
		}

		// Verify dimensions are realistic
		if sprite.Width <= 0 || sprite.Width > 1000 {
			t.Errorf("Sprite %d has unrealistic width: %d", spriteIdx, sprite.Width)
		}
		if sprite.Height <= 0 || sprite.Height > 1000 {
			t.Errorf("Sprite %d has unrealistic height: %d", spriteIdx, sprite.Height)
		}
	}
}

// Helper function to create a minimal valid image table
func createMinimalImageTable(numSprites int, width, height int16) []byte {
	dataPerSprite := int(width) * int(height)
	totalDataSize := numSprites * dataPerSprite

	table := []byte{
		// Header
		byte(numSprites), 0, 0, 0,
		byte(totalDataSize), byte(totalDataSize >> 8), byte(totalDataSize >> 16), byte(totalDataSize >> 24),
	}

	// Add elements
	for i := 0; i < numSprites; i++ {
		offset := i * dataPerSprite
		table = append(table,
			byte(offset), byte(offset>>8), byte(offset>>16), byte(offset>>24),
			byte(width), byte(width>>8),
			byte(height), byte(height>>8),
			0, 0, 0, 0, 0, 0, 0, 0,
		)
	}

	// Add pixel data
	pixelData := make([]byte, totalDataSize)
	for i := range pixelData {
		pixelData[i] = byte(i % 256)
	}
	table = append(table, pixelData...)

	return table
}
