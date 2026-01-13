package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/Track/TrackModSection.h"
// #include "Ui/Widgets/CaptionWidget.h"
// #include "Ui/Widgets/FrameWidget.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/Widgets/TabWidget.h"
// #include "Ui/Window.h"
// #include <sfl/static_vector.hpp>
// namespace OpenLoco::World
// forward: struct RoadElement;
// forward: struct TrackElement;
// namespace OpenLoco::Ui::Windows::Construction
type GhostVisibilityFlags int

const (
	None GhostVisibilityFlags = 0
	ConstructArrow GhostVisibilityFlags = 1 << 0
	Track GhostVisibilityFlags = 1 << 1
	Signal GhostVisibilityFlags = 1 << 2
	Station GhostVisibilityFlags = 1 << 3
	Overhead GhostVisibilityFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(GhostVisibilityFlags);
type ConstructionState struct {
	TrackCost uint32
	RoadCost uint32
	ModCost uint32
	SignalCost uint32
	StationCost uint32
	ConstructingStationId StationId
	ConstructingStationAcceptedCargoTypes uint32
	ConstructingStationProducedCargoTypes uint32
// World::Pos2 stationMinPos;                      // 0x01135F7C
// World::Pos2 stationMaxPos;                      // 0x01135F80
	ViewportFlags ViewportFlags
	X uint16
	Y uint16
	ConstructionZ uint16
// World::Pos3 ghostTrackPos;        // 0x01135FBA
// World::Pos3 ghostRemovalTrackPos; // 0x01135FC0
// World::Pos3 nextTile;             // 0x01135FC6
	NextTileRotation uint16
// World::Pos3 previousTile;         // 0x01135FCE
	PreviousTileRotation uint16
	PreviewTrackPieceId uint16
	PreviewMods uint16
	LastSelectedMods uint16
// World::Pos3 stationGhostPos;    // 0x01135FE6
	StationGhostType uint16
// World::Pos3 modGhostPos;        // 0x01135FF8
	ConstructionHeight uint16
	ConstructionMaxHeight int16
	SignalGhostSides uint16
// World::Pos3 signalGhostPos;     // 0x01136004
	SignalGhostTrackObjId uint16
	ModGhostTrackObjId uint8
// uint8_t signalList[17];          // 0x0113601D
	LastSelectedSignal uint8
	IsSignalBothDirections uint8
// uint8_t bridgeList[9];           // 0x01136030
	LastSelectedBridge uint8
	Byte_113603A uint8
// uint8_t stationList[17];         // 0x0113603B
	LastSelectedStationType uint8
	SignalGhostRotation uint8
	SignalGhostTrackId uint8
	SignalGhostTileIndex uint8
// uint8_t modList[4];        // 0x01136054
	ModGhostRotation uint8
	ModGhostTrackId uint8
	ModGhostTileIndex uint8
	MakeJunction uint8
	ConstructionHover bool
	TrackType uint8
	Byte_1136063 uint8
	ConstructionRotation uint8
	PlaceTrackPieceId uint8
	ConstructionArrowFrameNum uint8
	LastSelectedTrackPiece uint8
	LastSelectedTrackGradient uint8
	GhostRemovalTrackRotation uint8
	GhostRemovalTrackId uint8
	StationGhostRotation uint8
	StationGhostTrackId uint8
	StationGhostTileIndex uint8
// World::Track::ModSection lastSelectedTrackModSection; // 0x0113606E
	Byte_1136076 uint8
	PreviewTrackType uint8
	PreviewRotation uint8
	LastSelectedTrackPieceId uint8
	Byte_113607E uint8
	StationGhostTypeDockAirport uint8
// World::TileElement backupTileElement; // 0x01136090
}
// namespace Common
type Widx int

const (
	Frame Widx = iota
	Caption
	Close_button
	Panel
	Tab_construction
	Tab_station
	Tab_signal
	Tab_overhead
)
// func MakeCommonWidgets(frameWidth int32, frameHeight int32, windowCaptionId StringId) any
// return makeWidgets(
// Widgets::Frame({ 0, 0 }, { frameWidth, frameHeight }, WindowColour::primary),
// Widgets::Caption({ 1, 1 }, { frameWidth - 2, 13 }, Widgets::Caption::Style::colourText, WindowColour::primary, windowCaptionId),
// Widgets::ImageButton({ frameWidth - 15, 2 }, { 13, 13 }, WindowColour::primary, ImageIds::close_button, StringIds::tooltip_close_window),
// Widgets::Panel({ 0, 41 }, { frameWidth, frameHeight - 41 }, WindowColour::secondary),
// Widgets::Tab({ 3, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_track_road_construction),
// Widgets::Tab({ 34, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_station_construction),
// Widgets::Tab({ 65, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_signal_construction),
// Widgets::Tab({ 96, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tab_electrification_construction));
// func PrepareDraw(self Window) 
// func ResetWindow(self Window, tabWidgetIndex WidgetIndex_t) 
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// func RepositionTabs(self Window) 
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// func OnClose(self Window) 
// func OnUpdate(self Window, flag GhostVisibilityFlags) 
// func Sub_4CD454() 
// func SetTrackOptions(trackType uint8) 
// func SetDisabledWidgets(self Window) 
// func CreateConstructionWindow() 
// func Sub_4A3A50() 
// func SetNextAndPreviousTrackTile(elTrack World::TrackElement, pos World::Pos2) 
// func SetNextAndPreviousRoadTile(elRoad World::RoadElement, pos World::Pos2) 
// func IsPointCloserToNextOrPreviousTile(point Point, viewport Viewport) bool
// func PreviousTab(self Window) 
// func NextTab(self Window) 
// [[nodiscard]] bool hasGhostVisibilityFlag(GhostVisibilityFlags flags);
// func SetGhostVisibilityFlag(flag GhostVisibilityFlags) 
// func ToggleGhostVisibilityFlag(flag GhostVisibilityFlags) 
// func UnsetGhostVisibilityFlag(flag GhostVisibilityFlags) 
// template<size_t NewCapacity, size_t LegacyCapacity>
// func CopyToLegacyList(any /* sfl::static_vector<uint8_t */ , sflType NewCapacity>, (legacyList uint8) 
// static_assert(LegacyCapacity > NewCapacity);
// std::copy(sflType.begin(), sflType.end(), legacyList);
// legacyList[sflType.size()] = 0xFFU;
// namespace Construction
// static constexpr Ui::Size kWindowSize = { 138, 276 };
type TrackRoadPreviewFlags int

const (
	None TrackRoadPreviewFlags = 0
	SkipTrackRoadSurfaces TrackRoadPreviewFlags = 1 << 0
	DisplayConstructionArrow TrackRoadPreviewFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TrackRoadPreviewFlags);
type Widx int

const (
	Left_hand_curve_very_small Widx = 8
	Left_hand_curve_small
	Left_hand_curve
	Left_hand_curve_large
	Right_hand_curve_large
	Right_hand_curve
	Right_hand_curve_small
	Right_hand_curve_very_small
	S_bend_dual_track_left
	S_bend_left
	Straight
	S_bend_right
	S_bend_dual_track_right
	Steep_slope_down
	Slope_down
	Level
	Slope_up
	Steep_slope_up
	Bridge
	Bridge_dropdown
	Construct
	Remove
	Rotate_90
)
// // clang-format off
// constexpr uint64_t allTrack = {
// (1ULL << widx::left_hand_curve_very_small) |
// (1ULL << widx::left_hand_curve_small) |
// (1ULL << widx::left_hand_curve) |
// (1ULL << widx::left_hand_curve_large) |
// (1ULL << widx::right_hand_curve_large) |
// (1ULL << widx::right_hand_curve) |
// (1ULL << widx::right_hand_curve_small) |
// (1ULL << widx::right_hand_curve_very_small ) |
// (1ULL << widx::s_bend_dual_track_left) |
// (1ULL << widx::s_bend_left) |
// (1ULL << widx::straight) |
// (1ULL << widx::s_bend_right) |
// (1ULL << widx::s_bend_dual_track_right) |
// (1ULL << widx::steep_slope_down) |
// (1ULL << widx::slope_down) |
// (1ULL << widx::level) |
// (1ULL << widx::slope_up) |
// (1ULL << widx::steep_slope_up)
// constexpr uint64_t allConstruction = {
// allTrack |
// (1ULL << widx::bridge) |
// (1ULL << widx::bridge_dropdown) |
// (1ULL << widx::construct) |
// (1ULL << widx::remove) |
// (1ULL << widx::rotate_90)
// //clang-format on
// std::span<const Widget> getWidgets();
// func Reset() 
// func ActivateSelectedConstructionWidgets() 
// func TabReset(self Window) 
// func DrawTrack(pos World::Pos3, selectedMods uint16, trackType uint8, trackPieceId uint8, rotation uint8, flags TrackRoadPreviewFlags, drawingCtx Gfx::DrawingContext) 
// func DrawRoad(pos World::Pos3, selectedMods uint16, trackType uint8, trackPieceId uint8, rotation uint8, flags TrackRoadPreviewFlags, drawingCtx Gfx::DrawingContext) 
// func RemoveTrackGhosts() 
// func PreviousTrackPiece(self Window) 
// func NextTrackPiece(self Window) 
// func PreviousSlope(self Window) 
// func NextSlope(self Window) 
// func BuildAtCurrentPos(self Window) 
// func RemoveAtCurrentPos(self Window) 
// func SelectPosition(self Window) 
// const WindowEventList& getEvents();
// namespace Station
type Widx int

const (
	Station Widx = 8
	Station_dropdown
	Image
	Rotate
)
// std::span<const Widget> getWidgets();
// func TabReset(self Window) 
// func RemoveStationGhost() 
// const WindowEventList& getEvents();
// func Sub_49E1F1(id StationId) 
// namespace Signal
type Widx int

const (
	Signal Widx = 8
	Signal_dropdown
	Both_directions
	Single_direction
)
// std::span<const Widget> getWidgets();
// func TabReset(self Window) 
// func RemoveSignalGhost() 
// const WindowEventList& getEvents();
// namespace Overhead
type Widx int

const (
	Checkbox_1 Widx = 8
	Checkbox_2
	Checkbox_3
	Checkbox_4
	Image
	Track
	Track_dropdown
)
// std::span<const Widget> getWidgets();
// func TabReset(self Window) 
// func RemoveTrackModsGhost() 
// const WindowEventList& getEvents();
// ConstructionState& getConstructionState();
