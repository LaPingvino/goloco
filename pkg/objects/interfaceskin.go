package objects

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// InterfaceSkinObject represents a loaded InterfaceSkin DAT object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/InterfaceSkinObject.h
//
//	The struct is exactly 0x18 (24) bytes.
//	After the fixed struct comes a string table (localized name) and
//	then an image table in G1 format.
type InterfaceSkinObject struct {
	Header *ObjectHeader

	// Fixed structure fields (24 bytes total)
	Name                       uint16 // StringId - reference to localized name
	Img                        uint32 // Base sprite offset in global pool
	MapTooltipObjectColour     uint8  // Colour palette index
	MapTooltipCargoColour      uint8
	TooltipColour              uint8
	ErrorColour                uint8
	WindowPlayerColor          uint8
	WindowTitlebarColour       uint8
	WindowColour               uint8
	WindowConstructionColour   uint8
	WindowTerraFormColour      uint8
	WindowMapColour            uint8
	WindowOptionsColour        uint8
	Colour11                   uint8
	TopToolbarPrimaryColour    uint8
	TopToolbarSecondaryColour  uint8
	TopToolbarTertiaryColour   uint8
	TopToolbarQuaternaryColour uint8
	PlayerInfoToolbarColour    uint8
	TimeToolbarColour          uint8

	// Loaded data
	DisplayName string           // Extracted from string table
	ImageOffset uint32           // Global sprite index where these sprites start
	ImageCount  uint32           // Number of sprites in this skin
	Sprites     []*SpriteElement // G1-format sprite elements
}

// InterfaceSkin image ID constants
//
// OpenLoco reference: src/OpenLoco/src/Objects/InterfaceSkinObject.h
//
//	namespace InterfaceSkin::ImageIds
//
// These are offsets from the InterfaceSkinObject.Img base.
// Usage: interface.Img + ImageIds.ToolbarLoadsave
const (
	ISPreviewImage                     = 0
	ISToolbarPause                     = 1
	ISToolbarPauseHover                = 2
	ISToolbarLoadsave                  = 3
	ISToolbarLoadsaveHover             = 4
	ISToolbarZoom                      = 5
	ISToolbarZoomHover                 = 6
	ISToolbarRotate                    = 7
	ISToolbarRotateHover               = 8
	ISToolbarTerraform                 = 9
	ISToolbarTerraformHover            = 10
	ISToolbarAudioActive               = 11
	ISToolbarAudioActiveHover          = 12
	ISToolbarAudioInactive             = 13
	ISToolbarAudioInactiveHover        = 14
	ISToolbarView                      = 15
	ISToolbarViewHover                 = 16
	ISToolbarTowns                     = 17
	ISToolbarTownsHover                = 18
	ISToolbarEmptyOpaque               = 19
	ISToolbarEmptyOpaqueHover          = 20
	ISToolbarEmptyTransparent          = 21
	ISToolbarEmptyTransparentHover     = 22
	ISToolbarIndustries                = 23
	ISToolbarIndustriesHover           = 24
	ISToolbarAirports                  = 25
	ISToolbarAirportsHover             = 26
	ISToolbarPorts                     = 27
	ISToolbarPortsHover                = 28
	ISToolbarCogwheels                 = 29
	ISToolbarCogwheelsHover            = 30
	ISToolbarBuildVehicleTrain         = 31
	ISToolbarBuildVehicleTrainHover    = 32
	ISToolbarBuildVehicleBus           = 33
	ISToolbarBuildVehicleBusHover      = 34
	ISToolbarBuildVehicleTruck         = 35
	ISToolbarBuildVehicleTruckHover    = 36
	ISToolbarBuildVehicleTram          = 37
	ISToolbarBuildVehicleTramHover     = 38
	ISToolbarBuildVehicleAirplane      = 39
	ISToolbarBuildVehicleAirplaneHover = 40
	ISToolbarBuildVehicleBoat          = 41
	ISToolbarBuildVehicleBoatHover     = 42
	ISToolbarStations                  = 43
	ISToolbarStationsHover             = 44
)

// ParseInterfaceSkinObject parses an InterfaceSkin DAT file payload.
//
// OpenLoco reference: src/OpenLoco/src/Objects/InterfaceSkinObject.cpp
//
//	InterfaceSkinObject::load() at address 0x0043C82D
//
// Layout:
//
//	[24 bytes] fixed InterfaceSkinObject struct
//	[variable] string table (localized name)
//	[variable] image table (G1 format sprite data)
func ParseInterfaceSkinObject(header *ObjectHeader, data []byte) (*InterfaceSkinObject, error) {
	if len(data) < 24 {
		return nil, fmt.Errorf("interfaceskin data too small: %d bytes", len(data))
	}

	obj := &InterfaceSkinObject{
		Header: header,
	}

	// Parse fixed struct (24 bytes)
	r := bytes.NewReader(data[:24])
	binary.Read(r, binary.LittleEndian, &obj.Name)
	binary.Read(r, binary.LittleEndian, &obj.Img)
	binary.Read(r, binary.LittleEndian, &obj.MapTooltipObjectColour)
	binary.Read(r, binary.LittleEndian, &obj.MapTooltipCargoColour)
	binary.Read(r, binary.LittleEndian, &obj.TooltipColour)
	binary.Read(r, binary.LittleEndian, &obj.ErrorColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowPlayerColor)
	binary.Read(r, binary.LittleEndian, &obj.WindowTitlebarColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowConstructionColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowTerraFormColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowMapColour)
	binary.Read(r, binary.LittleEndian, &obj.WindowOptionsColour)
	binary.Read(r, binary.LittleEndian, &obj.Colour11)
	binary.Read(r, binary.LittleEndian, &obj.TopToolbarPrimaryColour)
	binary.Read(r, binary.LittleEndian, &obj.TopToolbarSecondaryColour)
	binary.Read(r, binary.LittleEndian, &obj.TopToolbarTertiaryColour)
	binary.Read(r, binary.LittleEndian, &obj.TopToolbarQuaternaryColour)
	binary.Read(r, binary.LittleEndian, &obj.PlayerInfoToolbarColour)
	binary.Read(r, binary.LittleEndian, &obj.TimeToolbarColour)

	remainingData := data[24:]

	// Parse string table
	// OpenLoco: ObjectManager::loadStringTable() at 0x00471FF5
	// Format: [uint16 numStrings] [uint16 offsets...] [string data...]
	stringTableLen, displayName := parseStringTable(remainingData)
	obj.DisplayName = displayName
	remainingData = remainingData[stringTableLen:]

	// Parse image table (G1 format)
	sprites, err := parseImageTable(remainingData)
	if err != nil {
		return nil, fmt.Errorf("parsing image table: %w", err)
	}
	obj.Sprites = sprites
	obj.ImageCount = uint32(len(sprites))

	return obj, nil
}
