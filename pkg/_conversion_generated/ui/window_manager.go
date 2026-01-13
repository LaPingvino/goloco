package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringManager.h"
// #include "Window.h"
// #include <Map/Track/TrackModSection.h>
// #include <OpenLoco/Engine/World.hpp>
// #include <cstddef>
// #include <functional>
// #include <string_view>
// namespace OpenLoco
type LoadOrQuitMode int

const (
type ObjectType int

const (
)
// namespace OpenLoco::Gfx
// forward: struct RenderTarget;
// namespace OpenLoco::Ui
// forward: struct Viewport;
// forward: struct Window;
// namespace OpenLoco::World
// forward: struct TrackElement;
// forward: struct RoadElement;
// forward: struct TreeElement;
// namespace OpenLoco::Ui::WindowManager
type ViewportVisibility int

const (
	Reset ViewportVisibility = iota
	UndergroundView
	HeightMarksOnTrack
	OvergroundView
)
// func Init() 
// func SetWindowColours(slot WindowColour, colour AdvancedColour) 
// func GetWindowColour(slot WindowColour) AdvancedColour
// func GetCurrentModalType() WindowType
// func SetCurrentModalType(type WindowType) 
// func ResetThousandthTickCounter() 
// Window* get(size_t index);
// func IndexOf(pWindow Window) int
// func Count() int
// func UpdateViewports() 
// func Update() 
// func UpdateDaily() 
// Window* getMainWindow();
// Viewport* getMainViewport();
// Window* find(WindowType type);
// Window* find(WindowType type, WindowNumber_t number);
// Window* findAt(int32_t x, int32_t y);
// Window* findAt(Ui::Point point);
// Window* findAtAlt(int32_t x, int32_t y);
// Window* bringToFront(Window& window);
// Window* bringToFront(WindowType type, WindowNumber_t id = 0);
// func Invalidate(type WindowType) 
// func Invalidate(type WindowType, number WindowNumber_t) 
// func InvalidateWidget(type WindowType, number WindowNumber_t, widgetIndex WidgetIndex_t) 
// func InvalidateAllWindowsAfterInput() 
// func Close(type WindowType) 
// func Close(type WindowType, id WindowNumber_t) 
// func Close(window Window) 
// Window* createWindow(WindowType type, Ui::Size size, WindowFlags flags, const WindowEventList& events);
// Window* createWindow(WindowType type, Ui::Point origin, Ui::Size size, WindowFlags flags, const WindowEventList& events);
// Window* createWindowCentred(WindowType type, Ui::Size size, WindowFlags flags, const WindowEventList& events);
// Window* createWindow(WindowType type, Ui::Size size, WindowFlags flags, const WindowEventList& events);
// func DispatchUpdateAll() 
// func CallEvent8OnAllWindows() 
// func CallEvent9OnAllWindows() 
// func CallViewportRotateEventOnAllWindows() 
// func CallKeyUpEventBackToFront(charCode uint32, keyCode uint32) bool
// func RelocateWindows() 
// func MoveOtherWindowsDown(self Window) 
// func InvalidateOrderPageByVehicleNumber(number WindowNumber_t) 
// func CloseConstructionWindows() 
// func CloseTopmost() 
// func WheelInput(wheel int) 
// func IsInFront(w Ui::Window) bool
// func IsInFrontAlt(w Ui::Window) bool
// Ui::Window* findWindowShowing(const viewport_pos& position);
// func CloseAllFloatingWindows() 
// func GetCurrentRotation() int32
// func SetCurrentRotation(value int32) 
// func ViewportShiftPixels(window Ui::Window, viewport Ui::Viewport, dX int32, dY int32) 
// func ViewportSetVisibility(flags ViewportVisibility) 
// // 0x0052622E
// func GetVehiclePreviewRotationFrame() uint16
// func SetVehiclePreviewRotationFrame(uint16) 
// func GetVehiclePreviewRotationFrameYaw() uint8
// func GetVehiclePreviewRotationFrameRoll() uint8
// func Render(ctx Gfx::DrawingContext, rect Rect) 
// namespace OpenLoco::Vehicles
// forward: struct VehicleBase;
// forward: struct VehicleBogie;
// forward: struct Car;
// namespace OpenLoco::Ui::Windows
// namespace About
// func Open() 
// namespace AboutMusic
// func Open() 
// namespace BuildVehicle
// Window* openByVehicleId(EntityId vehicleId);
// Window* openByType(VehicleType vehicleType);
// Window* openByVehicleObjectId(uint16_t vehicleObjectId);
// func Sub_4B92A5(window Ui::Window) 
// namespace Cheats
// Window* open();
// namespace CompanyFaceSelection
// func Open(id CompanyId, callingWindowType WindowType) 
// namespace CompanyList
// func OpenPerformanceIndexes() 
// Window* open();
// func RemoveCompany(id CompanyId) 
// namespace CompanyWindow
// Window* open(CompanyId companyId);
// Window* openAndSetName();
// Window* openChallenge(CompanyId companyId);
// Window* openFinances(CompanyId companyId);
// func Rotate(self Window) bool
// namespace Construction
// Window* openWithFlags(uint32_t flags);
// Window* openAtTrack(const Window& main, World::TrackElement* track, const World::Pos2 pos);
// Window* openAtRoad(const Window& main, World::RoadElement* track, const World::Pos2 pos);
// func UpdateAvailableRoadAndRailOptions() 
// func UpdateAvailableAirportAndDockOptions() 
// func Sub_4A6FAC() 
// func IsStationTabOpen() bool
// func IsOverheadTabOpen() bool
// func IsSignalTabOpen() bool
// func Rotate(self Window) bool
// func RemoveConstructionGhosts() 
// func ResetGhostVisibilityFlags() 
// func GetLastSelectedMods() uint16
// World::Track::ModSection getLastSelectedTrackModSection();
// namespace DragVehiclePart
// func Open(car Vehicles::Car) 
// Vehicles::VehicleBogie* getDragCarComponent();
// namespace EditKeyboardShortcut
// Window* open(uint8_t shortcutIndex);
// namespace Error
// func Open(title StringId, StringIds::null StringId message =) 
// func OpenQuiet(title StringId, StringIds::null StringId message =) 
// func OpenWithCompetitor(title StringId, message StringId, competitorId CompanyId) 
// namespace Industry
// Window* open(IndustryId id);
// namespace IndustryList
// Window* open();
// func Reset() 
// func RemoveIndustry(id IndustryId) 
// namespace KeyboardShortcuts
// Window* open();
// namespace LandscapeGeneration
// Window* open();
// namespace Main
// func Open() 
// func ShowGridlines() 
// func HideGridlines() 
// func ShowDirectionArrows() 
// func HideDirectionArrows() 
// func ViewportFocusOnEntity(main Window, targetEntity EntityId) 
// func ViewportIsFocusedOnEntity(main Window, targetEntity EntityId) bool
// func ViewportIsFocusedOnAnyEntity(main Window) bool
// func ViewportUnfocusFromEntity(main Window) 
// namespace MapToolTip
// func Open() 
// func SetOwner(company CompanyId) 
// func GetTooltipTimeout() uint16
// func Reset() 
// namespace MapWindow
// func Open() 
// func CenterOnViewPoint() 
// namespace MessageWindow
// func Open() 
// namespace MusicSelection
// Window* open();
// namespace NetworkStatus
type CloseCallback = any /* std::function<void()> */ 
// Window* open(std::string_view text, CloseCallback cbClose);
// func SetText(text string) 
// func SetText(text string, cbClose CloseCallback) 
// func Close() 
// namespace NewsWindow
// func Open(messageIndex MessageId) 
// func OpenLastMessage() 
// func Close(window Ui::Window) 
// namespace ObjectLoadError
// Window* open(const std::vector<ObjectHeader>& list);
// namespace ObjectSelectionWindow
// Window* open();
// Window& openInTab(ObjectType objectType);
// func TryCloseWindow() bool
// namespace Options
// Window* open();
// Window* openMusicSettings();
const TabOffsetMusic uint8 = 2
// namespace PlayerInfoPanel
// Window* open();
// func InvalidateFrame() 
// namespace ProgressBar
// Window* open(std::string_view captionString);
// func SetProgress(value uint8) 
// func Close() 
// namespace PromptBrowse
type Browse_type int

const (
	Load Browse_type = 1
	Save Browse_type = 2
)
// std::optional<std::string> open(browse_type type, std::string_view path, const char* filter, StringId titleId);
// namespace PromptOkCancel
// func Open(captionId StringId, descriptionId StringId, descriptionArgs FormatArguments, okButtonStringId StringId) bool
// namespace PromptSaveWindow
// Window* open(LoadOrQuitMode savePromptType);
// namespace ScenarioOptions
// Window* open();
// namespace ScenarioSelect
// Window* open();
// namespace Station
// Window* open(StationId id);
// func Reset() 
// func ShowStationCatchment(id StationId) 
// func Sub_491BC6() 
// namespace VehiclesStopping
// func RemoveTrainFromList(self Window, head EntityId) 
// namespace StationList
// Window* open(CompanyId companyId);
// Window* open(CompanyId companyId, uint8_t type);
// func RemoveStationFromList(stationId StationId) 
// namespace Terraform
// Window* open();
// func OpenClearArea() 
// func OpenAdjustLand() 
// func OpenAdjustWater() 
// func OpenPlantTrees() 
// func OpenBuildWalls() 
// func Rotate(Window) bool
// func SetAdjustLandToolSize(size uint8) 
// func SetAdjustWaterToolSize(size uint8) 
// func SetClearAreaToolSize(size uint8) 
// func SetLastPlacedTree(elTree World::TreeElement) 
// func ResetLastSelections() 
// namespace TextInput
// func OpenTextInput(w Ui::Window, title StringId, message StringId, value StringId, callingWidget int, valueArgs FormatArgumentsView, 1 uint32_t inputSize = StringManager::kUserStringSize -) 
// func Sub_4CE6C9(type WindowType, number WindowNumber_t) 
// func Cancel() 
// func Sub_4CE6FF() 
// namespace TileInspector
// Window* open();
// namespace TimePanel
// Window* open();
// func InvalidateFrame() 
// func BeginSendChatMessage(self Window) 
// namespace TitleExit
// Window* open();
// namespace TitleLogo
// Window* open();
// namespace TitleMenu
// Window* open();
// func BeginSendChatMessage(self Window) 
// namespace TitleOptions
// Window* open();
// namespace TitleVersion
// Window* open();
// namespace ToolbarBottom::Editor
// func Open() 
// namespace ToolbarTop::Game
// func Open() 
// namespace ToolbarTop::Editor
// func Open() 
// namespace ToolTip
// func Open(window Ui::Window, widgetIndex int32, x int32, y int32) 
// func Update(window Ui::Window, widgetIndex int32, stringId StringId, x int32, y int32) 
// func CloseAndReset() 
// namespace Town
// Window* open(uint16_t townId);
// namespace TownList
// Window* open();
// func RemoveTown(TownId) 
// func Reset() 
// func Rotate(self Window) bool
// namespace Tutorial
// Window* open();
// namespace Vehicle
// namespace Main
// Window* open(const Vehicles::VehicleBase* vehicle);
// namespace Details
// Window* open(const Vehicles::VehicleBase* vehicle);
// func ScrollDrag(pos Ui::Point) 
// func ScrollDragEnd(pos Ui::Point) 
// func Rotate() bool
// func CancelVehicleTools() bool
// namespace VehicleList
// Window* open(CompanyId companyId, VehicleType type);
// func RemoveTrainFromList(self Window, head EntityId) 
// namespace Debug
// Window* open();
