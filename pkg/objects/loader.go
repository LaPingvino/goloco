package objects

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Object is the interface for all loaded objects
type Object interface {
	GetHeader() *ObjectHeader
}

// LoadedObject holds a loaded object with its data
type LoadedObject struct {
	Header       *ObjectHeader
	Data         []byte      // Decompressed object data
	Object       interface{} // Parsed object (VehicleObject, etc.)
	ImageOffset  uint32      // Base image offset in global sprite pool
	ImageCount   uint32      // Number of images in this object
	ImageData    []byte      // Raw sprite data for this object
}

// ObjectManager manages loaded objects
type ObjectManager struct {
	Objects     map[string]*LoadedObject
	Vehicles    []*VehicleObject
	ObjDataPath string
}

// NewObjectManager creates a new object manager
func NewObjectManager(objDataPath string) *ObjectManager {
	return &ObjectManager{
		Objects:     make(map[string]*LoadedObject),
		Vehicles:    make([]*VehicleObject, 0),
		ObjDataPath: objDataPath,
	}
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
	case ObjectTypeVehicle:
		vehicle, err := ParseVehicleObject(header, decompressed)
		if err != nil {
			return nil, fmt.Errorf("parsing vehicle: %w", err)
		}
		// Try to extract display name from string table
		vehicle.DisplayName = extractFirstString(decompressed)
		loaded.Object = vehicle
		m.Vehicles = append(m.Vehicles, vehicle)
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
	return nil
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

func isPrintable(b byte) bool {
	return b >= 0x20 && b < 0x7F
}

func cleanString(data []byte) string {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if b == 0xFF {
			continue // Skip FF bytes (format codes)
		}
		if b == 0 {
			break
		}
		if isPrintable(b) {
			result = append(result, b)
		}
	}
	return strings.TrimSpace(string(result))
}
