package input

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// namespace OpenLoco::Input
type Shortcut int

const (
	CloseTopmostWindow Shortcut = iota
	CloseAllFloatingWindows
	CancelConstructionMode
	PauseUnpauseGame
	ZoomViewOut
	ZoomViewIn
	RotateView
	RotateConstructionObject
	ToggleUndergroundView
	ToggleSeeThroughTracks
	ToggleSeeThroughRoads
	ToggleSeeThroughTrees
	ToggleSeeThroughBuildings
	ToggleSeeThroughScenery
	ToggleSeeThroughBridges
	ToggleHeightMarksOnLand
	ToggleHeightMarksOnTracks
	ToggleDirArrowsonTracks
	AdjustLand
	AdjustWater
	PlantTrees
	BulldozeArea
	BuildTracks
	BuildRoads
	BuildAirports
	BuildShipPorts
	BuildNewVehicles
	ShowVehiclesList
	ShowStationsList
	ShowTownsList
	ShowIndustriesList
	ShowMap
	ShowCompaniesList
	ShowCompanyInformation
	ShowFinances
	ShowAnnouncementsList
	ShowOptionsWindow
	Screenshot
	ToggleLastAnnouncement
	SendMessage
	ConstructionPreviousTab
	ConstructionNextTab
	ConstructionPreviousTrackPiece
	ConstructionNextTrackPiece
	ConstructionPreviousSlope
	ConstructionNextSlope
	ConstructionBuildAtCurrentPos
	ConstructionRemoveAtCurrentPos
	ConstructionSelectPosition
	GameSpeedNormal
	GameSpeedFastForward
	GameSpeedExtraFastForward
	OpenDebugWindow
)
// namespace Shortcuts
// func Initialize() 
