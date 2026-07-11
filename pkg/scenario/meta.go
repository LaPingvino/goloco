package scenario

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// ScenarioMeta holds metadata extracted from a .SC5 file without loading the full map.
// It is obtained cheaply by reading only the header + ScenarioOptions chunks.
type ScenarioMeta struct {
	FilePath    string
	Name        string           // scenarioName (offset 0x2A in options chunk)
	Description string           // scenarioDetails (offset 0x6A)
	Category    ScenarioCategory // difficulty (offset 0x01)
	StartYear   uint16
	HasPreview  bool             // landscapeGenerationDone flag set → hand-crafted map
	IsGenerated bool             // !HasPreview → map generated from parameters on load
	// Preview is the raw 128×128 palette-indexed image.
	// Zero-filled when IsGenerated is true (no preview stored).
	Preview        [PreviewSize * PreviewSize]byte
	MaxCompetitors uint8
	ObjectiveType  uint8

	Objective          Objective
	HasObjective       bool
	ObjectiveCargoName string
}

// ScanScenario reads the header and ScenarioOptions chunks from a .SC5 file
// and returns metadata without loading the tile map.
func ScanScenario(filePath string) (*ScenarioMeta, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	reader := assets.NewS5ChunkReader(data)

	// Chunk 1: header
	headerBytes, err := reader.ReadChunk()
	if err != nil {
		return nil, err
	}
	header, err := assets.ParseS5Header(headerBytes)
	if err != nil {
		return nil, err
	}
	if header.Type != assets.S5TypeScenario && header.Type != assets.S5TypeLandscape {
		return nil, nil // not a scenario or landscape, skip silently
	}

	// Skip SaveDetails if present
	if header.Flags&assets.HeaderFlagHasSaveDetails != 0 {
		if _, err := reader.ReadChunk(); err != nil {
			return nil, err
		}
	}

	// Chunk 2: ScenarioOptions
	optsBytes, err := reader.ReadChunk()
	if err != nil {
		return nil, err
	}
	opts := parseScenarioOptions(optsBytes)

	name := opts.Name
	if name == "" {
		// Fall back to filename without extension
		name = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	return &ScenarioMeta{
		FilePath:       filePath,
		Name:           name,
		Description:    opts.Description,
		Category:       opts.Category,
		StartYear:      opts.StartYear,
		HasPreview:     opts.HasPreview,
		IsGenerated:    !opts.HasPreview,
		Preview:        opts.Preview,
		MaxCompetitors: opts.MaxCompetitors,
		ObjectiveType:  opts.ObjectiveType,

		Objective:          opts.Objective,
		HasObjective:       opts.HasObjective,
		ObjectiveCargoName: opts.ObjectiveCargoName,
	}, nil
}

// ScanScenarios scans dir (and its immediate subdirectories) for .SC5 files,
// reads their metadata, and returns them sorted by category then name.
func ScanScenarios(dir string) ([]*ScenarioMeta, error) {
	var metas []*ScenarioMeta

	scanDir := func(d string) {
		entries, err := os.ReadDir(d)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if strings.ToLower(filepath.Ext(e.Name())) != ".sc5" {
				continue
			}
			m, err := ScanScenario(filepath.Join(d, e.Name()))
			if err != nil || m == nil {
				continue
			}
			metas = append(metas, m)
		}
	}

	// Scan the directory itself
	scanDir(dir)

	// Also scan one level of subdirectories (e.g., Scenarios/Landscapes/)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() {
			scanDir(filepath.Join(dir, e.Name()))
		}
	}

	// Sort: category ascending, then name alphabetically
	sort.Slice(metas, func(i, j int) bool {
		if metas[i].Category != metas[j].Category {
			return metas[i].Category < metas[j].Category
		}
		return strings.ToLower(metas[i].Name) < strings.ToLower(metas[j].Name)
	})

	return metas, nil
}
