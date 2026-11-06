package objects

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// IndustryObject — a port of OpenLoco Objects/IndustryObject.h/.cpp: the fields
// the painter needs, read with load()'s exact section walk. sizeof(IndustryObject)
// is 0xF4 (static_assert upstream).
type IndustryObject struct {
	Header *ObjectHeader
	Name   string

	NumBuildingParts                uint8        // 0x1E
	NumBuildingVariations           uint8        // 0x1F
	MaxNumBuildings                 uint8        // 0xBD
	MinNumBuildings                 uint8        // 0xBC
	AvailableColours                uint32       // 0xC2: a bit per colour
	CostFactor                      int16        // 0xD0
	InitialProductionRate           [2][2]uint16 // 0xD6: min, max per produced cargo
	FarmTileGrowthStageNoProduction uint8        // 0xEA (0xFF: no farm tiles)
	FarmNumFields                   uint8        // 0xEB
	FarmTileNumGrowthStages         uint8        // 0xEC
	WallTypes                       [4]uint8     // 0xED
	BuildingWall                    uint8        // 0xF1 (0xFF none)
	BuildingWallEntrance            uint8        // 0xF2
	Buildings                       []uint8      // the building types, in placement order (getBuildings)
	BuildingSizeFlags               uint32       // 0xC6: bit n set = building type n is 2×2
	ScaffoldingSegmentType          uint8        // 0xD4
	ScaffoldingColour               uint8        // 0xD5
	Flags                           uint32       // 0xE4 (IndustryObjectFlags)
	ObsoleteYear                    uint16       // 0xCC
	DesignedYear                    uint16       // 0xCA
	TotalOfTypeInScenario           uint8        // 0xCE: 1–32
	CostIndex                       uint8        // 0xCF
	ClearCostFactor                 int16        // 0xD2
	MonthlyClosureChance            uint8        // 0xF3: out of 65535
	FarmTileNumImageAngles          uint8        // 0xE9

	PartHeights        []uint8
	PartAnimations     []BuildingPartAnimation
	AnimationSequences [4][]uint8 // each: size (power of 2), then data
	Variations         [][]uint8  // building variation → parts

	// Image bases (load(): shadowImageIds = table offset; buildings after the
	// shadows when hasShadows; fields after the building parts).
	ShadowImageIDs               uint32
	BuildingImageIDs             uint32
	FieldImageIDs                uint32
	NumImagesPerFieldGrowthStage uint32
	TableLength                  uint32 // image table length (for the exact-walk check)
	// ProducedCargoNames / RequiredCargoNames are the cargo DAT names
	// (producedCargoType / requiredCargoType); "" for none.
	ProducedCargoNames [2]string
	RequiredCargoNames [3]string
	RemainingAtTable   int // bytes left when the image table starts
}

// IndustryFlagHasShadows is IndustryObjectFlags::hasShadows.
const IndustryFlagHasShadows = 1 << 21

// IndustryFlagFarmTilesDrawAboveSnow is IndustryObjectFlags::farmTilesDrawAboveSnow
// (ski slopes; other farm tiles draw below snow).
const IndustryFlagFarmTilesDrawAboveSnow = 1 << 26

const industryObjectSize = 0xF4

// ParseIndustryObjectWithG1 parses an industry object (IndustryObject::load) and
// loads its images into the G1 pool when g1 is set.
func ParseIndustryObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*IndustryObject, error) {
	if len(data) < industryObjectSize {
		return nil, fmt.Errorf("industry object too small: %d bytes", len(data))
	}
	o := &IndustryObject{
		Header:                          header,
		NumBuildingParts:                data[0x1E],
		NumBuildingVariations:           data[0x1F],
		MaxNumBuildings:                 data[0xBD],
		MinNumBuildings:                 data[0xBC],
		AvailableColours:                binary.LittleEndian.Uint32(data[0xC2:]),
		CostFactor:                      int16(binary.LittleEndian.Uint16(data[0xD0:])),
		InitialProductionRate:           [2][2]uint16{{binary.LittleEndian.Uint16(data[0xD6:]), binary.LittleEndian.Uint16(data[0xD8:])}, {binary.LittleEndian.Uint16(data[0xDA:]), binary.LittleEndian.Uint16(data[0xDC:])}},
		FarmTileGrowthStageNoProduction: data[0xEA],
		FarmNumFields:                   data[0xEB],
		FarmTileNumGrowthStages:         data[0xEC],
		WallTypes:                       [4]uint8{data[0xED], data[0xEE], data[0xEF], data[0xF0]},
		BuildingWall:                    data[0xF1],
		BuildingWallEntrance:            data[0xF2],
		BuildingSizeFlags:               binary.LittleEndian.Uint32(data[0xC6:]),
		ScaffoldingSegmentType:          data[0xD4],
		ScaffoldingColour:               data[0xD5],
		Flags:                           binary.LittleEndian.Uint32(data[0xE4:]),
		ObsoleteYear:                    binary.LittleEndian.Uint16(data[0xCC:]),
		DesignedYear:                    binary.LittleEndian.Uint16(data[0xCA:]),
		TotalOfTypeInScenario:           data[0xCE],
		CostIndex:                       data[0xCF],
		ClearCostFactor:                 int16(binary.LittleEndian.Uint16(data[0xD2:])),
		MonthlyClosureChance:            data[0xF3],
		FarmTileNumImageAngles:          data[0xE9],
	}
	p := industryObjectSize
	need := func(n int) error {
		if p+n > len(data) {
			return fmt.Errorf("industry %q truncated at %#x (+%d)", o.Name, p, n)
		}
		return nil
	}
	// 8 string tables: name, var_02, unused, closingDown, upProduction, downProduction, singular, plural.
	for i := 0; i < 8; i++ {
		n, str := parseStringTable(data[p:])
		if i == 0 {
			o.Name = str
		}
		p += n
	}
	np := int(o.NumBuildingParts)
	if err := need(np * 3); err != nil {
		return nil, err
	}
	o.PartHeights = append([]uint8(nil), data[p:p+np]...)
	p += np
	o.PartAnimations = make([]BuildingPartAnimation, np)
	for i := range o.PartAnimations {
		o.PartAnimations[i] = BuildingPartAnimation{NumFrames: data[p+2*i], AnimationSpeed: data[p+2*i+1]}
	}
	p += np * 2
	for i := range o.AnimationSequences {
		if err := need(1); err != nil {
			return nil, err
		}
		n := int(data[p])
		if err := need(1 + n); err != nil {
			return nil, err
		}
		o.AnimationSequences[i] = append([]uint8(nil), data[p+1:p+1+n]...)
		p += 1 + n
	}
	for { // var_38 entries (2 bytes each) until 0xFF
		if err := need(1); err != nil {
			return nil, err
		}
		if data[p] == 0xFF {
			p++
			break
		}
		p += 2
	}
	for v := 0; v < int(o.NumBuildingVariations); v++ {
		var parts []uint8
		for {
			if err := need(1); err != nil {
				return nil, err
			}
			if data[p] == 0xFF {
				p++
				break
			}
			parts = append(parts, data[p])
			p++
		}
		o.Variations = append(o.Variations, parts)
	}
	o.Buildings = append([]uint8(nil), data[p:p+int(o.MaxNumBuildings)]...) // buildings (getBuildings)
	p += int(o.MaxNumBuildings)
	// produced (2) and required (3) cargo headers (load resolves them into
	// producedCargoType / requiredCargoType; 0xFF: none).
	for i := range o.ProducedCargoNames {
		if data[p+i*16] != 0xFF {
			o.ProducedCargoNames[i] = headerName(data, p+i*16)
		}
	}
	for i := range o.RequiredCargoNames {
		if data[p+(2+i)*16] != 0xFF {
			o.RequiredCargoNames[i] = headerName(data, p+(2+i)*16)
		}
	}
	p += (2 + 3 + 4 + 1 + 1) * 16 // produced, required cargo; wall types; buildingWall; buildingWallEntrance
	if err := need(0); err != nil {
		return nil, err
	}
	o.RemainingAtTable = len(data) - p
	if g1 != nil {
		res, err := g1.LoadImageTable(data[p:])
		if err != nil {
			return nil, fmt.Errorf("industry %q image table: %w", o.Name, err)
		}
		o.TableLength = res.TableLength
		o.ShadowImageIDs = res.ImageOffset
		o.BuildingImageIDs = o.ShadowImageIDs
		if o.Flags&IndustryFlagHasShadows != 0 {
			o.BuildingImageIDs += uint32(o.NumBuildingVariations) * 4
		}
		o.FieldImageIDs = uint32(o.NumBuildingParts)*4 + o.BuildingImageIDs
		o.NumImagesPerFieldGrowthStage = uint32(o.FarmTileNumImageAngles) * 21
	}
	return o, nil
}

// IndustryCargoIDs are the industry object's producedCargoType and
// requiredCargoType: the cargo names resolved to the scenario's cargo slots
// (0xFF for none or not loaded, as load leaves them).
func (m *ObjectManager) IndustryCargoIDs(o *IndustryObject) (produced [2]uint8, required [3]uint8) {
	slot := func(name string) uint8 {
		if name == "" {
			return 0xFF
		}
		for i, c := range m.CargoObjects {
			if c != nil && c.Header != nil && strings.EqualFold(strings.TrimRight(string(c.Header.Name[:]), " \x00"), name) {
				return uint8(i)
			}
		}
		return 0xFF
	}
	for i, n := range o.ProducedCargoNames {
		produced[i] = slot(n)
	}
	for i, n := range o.RequiredCargoNames {
		required[i] = slot(n)
	}
	return
}

// IndustryFlagRequiresAllCargo is IndustryObjectFlags::requiresAllCargo.
const IndustryFlagRequiresAllCargo = 1 << 17

// IndustryFlagCanIncreaseProduction / CanDecreaseProduction are
// IndustryObjectFlags::canIncreaseProduction / canDecreaseProduction.
const (
	IndustryFlagCanIncreaseProduction = 1 << 18
	IndustryFlagCanDecreaseProduction = 1 << 19
)
