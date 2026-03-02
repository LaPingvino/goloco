package objects

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// G1Loader interface for loading sprites into the G1 pool
type G1Loader interface {
	LoadImageTable(data []byte) (ImageTableResult, error)
}

// ImageTableResult holds the result of loading an image table
type ImageTableResult struct {
	ImageOffset uint32
	TableLength uint32
}

// Object is the interface for all loaded objects
type Object interface {
	GetHeader() *ObjectHeader
}

// LoadedObject holds a loaded object with its data
type LoadedObject struct {
	Header      *ObjectHeader
	Data        []byte      // Decompressed object data
	Object      interface{} // Parsed object (VehicleObject, etc.)
	ImageOffset uint32      // Base image offset in global sprite pool
	ImageCount  uint32      // Number of images in this object
	ImageData   []byte      // Raw sprite data for this object
}

// ObjectManager manages loaded objects
type ObjectManager struct {
	Objects          map[string]*LoadedObject
	Vehicles         []*VehicleObject
	LandObjects      []*LandObject
	TreeObjects      []*TreeObject              // all loaded tree objects
	BuildingObjects  []*BuildingObject           // all loaded building objects
	WallObjects      []*WallObject              // all loaded wall objects
	CliffEdgeObjs    map[string]*CliffEdgeObject // keyed by upper-case name
	InterfaceSkin    *InterfaceSkinObject
	ObjDataPath      string

	// Sprite pool for object sprites
	NextSpriteIndex uint32
	G1File          interface{} // Reference to assets.G1File (using interface{} to avoid circular import)
}

// NewObjectManager creates a new object manager
func NewObjectManager(objDataPath string) *ObjectManager {
	return &ObjectManager{
		Objects:         make(map[string]*LoadedObject),
		Vehicles:        make([]*VehicleObject, 0),
		LandObjects:     make([]*LandObject, 0),
		TreeObjects:     make([]*TreeObject, 0),
		BuildingObjects: make([]*BuildingObject, 0),
		WallObjects:     make([]*WallObject, 0),
		CliffEdgeObjs:   make(map[string]*CliffEdgeObject),
		ObjDataPath:     objDataPath,
		NextSpriteIndex: 0, // Will be set after G1 is loaded
	}
}

// SetBaseSpriteIndex sets the starting index for object sprites
// This should be called after loading G1.DAT with the count of G1 sprites
func (m *ObjectManager) SetBaseSpriteIndex(baseIndex uint32) {
	m.NextSpriteIndex = baseIndex
}

// LoadObject loads a single DAT file
func (m *ObjectManager) LoadObject(path string) (*LoadedObject, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	return m.LoadObjectFromReader(f, filepath.Base(path))
}

// LoadObjectFromReader loads an object from a reader
func (m *ObjectManager) LoadObjectFromReader(r io.Reader, name string) (*LoadedObject, error) {
	// Read the full file into memory for easier processing
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	if len(data) < HeaderSize {
		return nil, fmt.Errorf("file too small: %d bytes", len(data))
	}

	// Parse header
	headerReader := bytes.NewReader(data[:HeaderSize])
	header, err := ReadHeader(headerReader)
	if err != nil {
		return nil, err
	}

	// The rest of the file after the header is the Sawyer-encoded chunk
	chunkReader := bytes.NewReader(data[HeaderSize:])
	decompressed, err := DecodeSawyerChunk(chunkReader)
	if err != nil {
		return nil, fmt.Errorf("decompressing: %w", err)
	}

	loaded := &LoadedObject{
		Header: header,
		Data:   decompressed,
	}

	// Parse type-specific data
	switch header.GetType() {
	case ObjectTypeInterfaceSkin:
		skin, err := ParseInterfaceSkinObject(header, decompressed)
		if err != nil {
			return nil, fmt.Errorf("parsing interface skin: %w", err)
		}
		// Assign sprite indices
		skin.ImageOffset = m.NextSpriteIndex
		loaded.ImageOffset = m.NextSpriteIndex
		loaded.ImageCount = uint32(len(skin.Sprites))
		m.NextSpriteIndex += uint32(len(skin.Sprites))

		loaded.Object = skin
		// Store the first (typically only) InterfaceSkin as the active one
		if m.InterfaceSkin == nil {
			m.InterfaceSkin = skin
			m.InterfaceSkin.Img = m.NextSpriteIndex - uint32(len(skin.Sprites))
		}

	case ObjectTypeVehicle:
		vehicle, err := ParseVehicleObject(header, decompressed)
		if err != nil {
			return nil, fmt.Errorf("parsing vehicle: %w", err)
		}
		// Try to extract display name from string table
		vehicle.DisplayName = extractFirstString(decompressed)
		loaded.Object = vehicle
		m.Vehicles = append(m.Vehicles, vehicle)

	case ObjectTypeLand:
		// Use G1 dynamic loading if available
		var land *LandObject
		var err error
		if m.G1File != nil {
			// Cast to G1Loader interface
			if g1Loader, ok := m.G1File.(G1Loader); ok {
				land, err = ParseLandObjectWithG1(header, decompressed, g1Loader)
			} else {
				land, err = ParseLandObject(header, decompressed)
			}
		} else {
			land, err = ParseLandObject(header, decompressed)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing land: %w", err)
		}

		// If using G1 dynamic loading, ImageOffset is already set
		// Otherwise, assign sprite indices for embedded sprites
		if land.ImageOffset == 0 && len(land.Sprites) > 0 {
			land.ImageOffset = m.NextSpriteIndex
			loaded.ImageOffset = m.NextSpriteIndex
			loaded.ImageCount = uint32(len(land.Sprites))
			m.NextSpriteIndex += uint32(len(land.Sprites))
		} else {
			loaded.ImageOffset = land.ImageOffset
			loaded.ImageCount = 0 // Sprites are in G1, not embedded
		}

		loaded.Object = land
		m.LandObjects = append(m.LandObjects, land)

	case ObjectTypeCliffEdge:
		var cliff *CliffEdgeObject
		var err error
		if m.G1File != nil {
			if g1Loader, ok := m.G1File.(G1Loader); ok {
				cliff, err = ParseCliffEdgeObjectWithG1(header, decompressed, g1Loader)
			} else {
				cliff, err = ParseCliffEdgeObject(header, decompressed)
			}
		} else {
			cliff, err = ParseCliffEdgeObject(header, decompressed)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing cliff edge: %w", err)
		}

		loaded.Object = cliff
		loaded.ImageOffset = cliff.Image
		m.CliffEdgeObjs[strings.ToUpper(header.GetName())] = cliff

	case ObjectTypeTree:
		var tree *TreeObject
		var err error
		if m.G1File != nil {
			if g1Loader, ok := m.G1File.(G1Loader); ok {
				tree, err = ParseTreeObjectWithG1(header, decompressed, g1Loader)
			} else {
				tree, err = ParseTreeObject(header, decompressed)
			}
		} else {
			tree, err = ParseTreeObject(header, decompressed)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing tree: %w", err)
		}

		loaded.Object = tree
		loaded.ImageOffset = tree.ImageOffset
		m.TreeObjects = append(m.TreeObjects, tree)

	case ObjectTypeBuilding:
		var bldg *BuildingObject
		var err error
		if m.G1File != nil {
			if g1Loader, ok := m.G1File.(G1Loader); ok {
				bldg, err = ParseBuildingObjectWithG1(header, decompressed, g1Loader)
			} else {
				bldg, err = ParseBuildingObject(header, decompressed)
			}
		} else {
			bldg, err = ParseBuildingObject(header, decompressed)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing building: %w", err)
		}

		loaded.Object = bldg
		loaded.ImageOffset = bldg.ImageOffset
		m.BuildingObjects = append(m.BuildingObjects, bldg)

	case ObjectTypeWall:
		var wall *WallObject
		var err error
		if m.G1File != nil {
			if g1Loader, ok := m.G1File.(G1Loader); ok {
				wall, err = ParseWallObjectWithG1(header, decompressed, g1Loader)
			} else {
				wall, err = ParseWallObject(header, decompressed)
			}
		} else {
			wall, err = ParseWallObject(header, decompressed)
		}
		if err != nil {
			return nil, fmt.Errorf("parsing wall: %w", err)
		}

		loaded.Object = wall
		loaded.ImageOffset = wall.Sprite
		m.WallObjects = append(m.WallObjects, wall)
	}

	// Store in map
	key := strings.ToUpper(header.GetName())
	m.Objects[key] = loaded

	return loaded, nil
}

// LoadAllObjects loads all DAT files from the ObjData directory
func (m *ObjectManager) LoadAllObjects() error {
	if m.ObjDataPath == "" {
		return fmt.Errorf("ObjDataPath not set")
	}

	entries, err := os.ReadDir(m.ObjDataPath)
	if err != nil {
		return fmt.Errorf("reading ObjData directory: %w", err)
	}

	loaded := 0
	failed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToUpper(name), ".DAT") {
			continue
		}

		path := filepath.Join(m.ObjDataPath, name)
		_, err := m.LoadObject(path)
		if err != nil {
			// Log but continue
			failed++
			continue
		}
		loaded++
	}

	fmt.Printf("Loaded %d objects (%d failed)\n", loaded, failed)

	// Resolve cliff edge references: link each LandObject to its CliffEdgeObject
	m.ResolveCliffEdges()

	return nil
}

// ResolveCliffEdges links LandObjects to their CliffEdgeObjects.
// Each LandObject's DAT file embeds an ObjectHeader naming the cliff edge
// object to use. After all objects are loaded, this method looks up the
// CliffEdgeObject by name and copies its Image (G1 sprite offset) into
// the LandObject's CliffEdgeImage field.
//
// OpenLoco reference: src/OpenLoco/src/Objects/LandObject.cpp
//   cliffEdgeImage = cliffObj->image;
func (m *ObjectManager) ResolveCliffEdges() {
	resolved := 0
	for _, land := range m.LandObjects {
		if land == nil || land.CliffEdgeName == "" {
			continue
		}
		if cliff, ok := m.CliffEdgeObjs[strings.ToUpper(land.CliffEdgeName)]; ok {
			land.CliffEdgeImage = cliff.Image
			resolved++
		}
	}
	if resolved > 0 {
		fmt.Printf("Resolved %d cliff edge references (%d cliff edge objects loaded)\n",
			resolved, len(m.CliffEdgeObjs))
	}
}

// LoadVehicles loads only vehicle objects
func (m *ObjectManager) LoadVehicles() error {
	if m.ObjDataPath == "" {
		return fmt.Errorf("ObjDataPath not set")
	}

	entries, err := os.ReadDir(m.ObjDataPath)
	if err != nil {
		return fmt.Errorf("reading ObjData directory: %w", err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToUpper(name), ".DAT") {
			continue
		}

		path := filepath.Join(m.ObjDataPath, name)
		obj, err := m.LoadObject(path)
		if err != nil {
			continue
		}

		// Only count vehicles
		if obj.Header.GetType() == ObjectTypeVehicle {
			loaded++
		}
	}

	fmt.Printf("Loaded %d vehicles\n", loaded)
	return nil
}

// GetObject returns a loaded object by name
func (m *ObjectManager) GetObject(name string) *LoadedObject {
	return m.Objects[strings.ToUpper(name)]
}

// GetVehicle returns a vehicle by name
func (m *ObjectManager) GetVehicle(name string) *VehicleObject {
	obj := m.GetObject(name)
	if obj == nil {
		return nil
	}
	if v, ok := obj.Object.(*VehicleObject); ok {
		return v
	}
	return nil
}

// GetLandObject returns a land object by name
func (m *ObjectManager) GetLandObject(name string) *LandObject {
	obj := m.GetObject(name)
	if obj == nil {
		return nil
	}
	if l, ok := obj.Object.(*LandObject); ok {
		return l
	}
	return nil
}

// GetDefaultLandObject returns the first loaded land object (typically GRASS1)
func (m *ObjectManager) GetDefaultLandObject() *LandObject {
	if len(m.LandObjects) > 0 {
		return m.LandObjects[0]
	}
	return nil
}

// GetLandObjectByIndex returns the LandObject at the given terrain slot index,
// or nil if the index is out of range.  The index corresponds to the 5-bit
// terrain field in a surface element (0–29 in OpenLoco).
func (m *ObjectManager) GetLandObjectByIndex(index int) *LandObject {
	if index < 0 || index >= len(m.LandObjects) {
		return nil
	}
	return m.LandObjects[index]
}

// GetTreeObjectByIndex returns the TreeObject at the given index (from tile
// element byte 4), or nil if the index is out of range.
func (m *ObjectManager) GetTreeObjectByIndex(index int) *TreeObject {
	if index < 0 || index >= len(m.TreeObjects) {
		return nil
	}
	return m.TreeObjects[index]
}

// GetBuildingObjectByIndex returns the BuildingObject at the given index,
// or nil if the index is out of range.
func (m *ObjectManager) GetBuildingObjectByIndex(index int) *BuildingObject {
	if index < 0 || index >= len(m.BuildingObjects) {
		return nil
	}
	return m.BuildingObjects[index]
}

// ReorderBuildingObjects rebuilds the BuildingObjects slice so that each
// building slot index maps to the BuildingObject whose name matches order[i].
func (m *ObjectManager) ReorderBuildingObjects(order []string) {
	reordered := make([]*BuildingObject, len(order))
	matched := 0
	for i, name := range order {
		if name == "" {
			continue
		}
		if obj := m.GetObject(strings.ToUpper(name)); obj != nil {
			if bldg, ok := obj.Object.(*BuildingObject); ok {
				reordered[i] = bldg
				matched++
			}
		}
	}
	m.BuildingObjects = reordered
	fmt.Printf("Reordered building objects: %d matched out of %d slots\n", matched, len(order))
}

// GetWallObjectByIndex returns the wall object at the given slot index
func (m *ObjectManager) GetWallObjectByIndex(index int) *WallObject {
	if index < 0 || index >= len(m.WallObjects) {
		return nil
	}
	return m.WallObjects[index]
}

// ReorderWallObjects rebuilds the WallObjects slice so that each
// wall slot index maps to the WallObject whose name matches order[i].
func (m *ObjectManager) ReorderWallObjects(order []string) {
	reordered := make([]*WallObject, len(order))
	matched := 0
	for i, name := range order {
		if name == "" {
			continue
		}
		if obj := m.GetObject(strings.ToUpper(name)); obj != nil {
			if wall, ok := obj.Object.(*WallObject); ok {
				reordered[i] = wall
				matched++
			}
		}
	}
	m.WallObjects = reordered
	fmt.Printf("Reordered wall objects: %d matched out of %d slots\n", matched, len(order))
}

// ReorderTreeObjects rebuilds the TreeObjects slice so that each tree
// slot index maps to the TreeObject whose name matches order[i].
func (m *ObjectManager) ReorderTreeObjects(order []string) {
	reordered := make([]*TreeObject, len(order))
	matched := 0
	for i, name := range order {
		if name == "" {
			continue
		}
		if obj := m.GetObject(strings.ToUpper(name)); obj != nil {
			if tree, ok := obj.Object.(*TreeObject); ok {
				reordered[i] = tree
				matched++
			}
		}
	}
	m.TreeObjects = reordered
	fmt.Printf("Reordered tree objects: %d matched out of %d slots\n", matched, len(order))
}

// ReorderLandObjects rebuilds the LandObjects slice so that each terrain
// slot index maps to the LandObject whose name matches order[i].  Slots
// whose name is "" or whose object was not loaded stay nil.
//
// This must be called after LoadAllObjects and after the scenario's
// RequiredObjects chunk has been parsed to obtain the correct slot order.
func (m *ObjectManager) ReorderLandObjects(order []string) {
	reordered := make([]*LandObject, len(order))
	for i, name := range order {
		if name == "" {
			continue
		}
		// Objects map is keyed by upper-case name
		if obj := m.GetObject(strings.ToUpper(name)); obj != nil {
			if land, ok := obj.Object.(*LandObject); ok {
				reordered[i] = land
			}
		}
	}
	m.LandObjects = reordered
}

// GetLandSprite returns a sprite from a land object
func (m *ObjectManager) GetLandSprite(land *LandObject, localIndex int) *SpriteElement {
	if land == nil || localIndex < 0 || localIndex >= len(land.Sprites) {
		return nil
	}
	return land.Sprites[localIndex]
}

// extractFirstString tries to extract the first string from object data
// Object strings are typically after the fixed structure, prefixed with length
func extractFirstString(data []byte) string {
	// Look for printable ASCII sequences
	// Strings in Loco objects often appear after the fixed structure
	// and can be identified by a sequence of printable characters

	// Skip the fixed structure (around 0x15E for vehicles)
	if len(data) < 0x160 {
		return ""
	}

	// Scan for a string starting point
	for i := 0x80; i < len(data)-10; i++ {
		// Look for a potential string start
		if isPrintable(data[i]) {
			// Check if we have a reasonable string
			end := i
			for end < len(data) && end-i < 100 {
				if data[end] == 0 {
					break
				}
				if !isPrintable(data[end]) && data[end] != 0xFF {
					break
				}
				end++
			}
			strLen := end - i
			if strLen >= 5 && strLen <= 60 {
				// Found a potential name
				str := cleanString(data[i:end])
				if len(str) >= 4 {
					return str
				}
			}
		}
	}
	return ""
}

// Note: isPrintable and cleanString are now in stringtable.go
