package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Colour.h"
// #include "Graphics/Gfx.h"
// #include "Map/Tile.h"
// #include "Objects/Object.h"
// #include "Types.hpp"
// #include "Ui.h"
// #include "Ui/ScrollFlags.hpp"
// #include "Ui/Widget.h"
// #include "Ui/WindowType.h"
// #include "Viewport.hpp"
// #include "World/Company.h"
// #include "ZoomLevel.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <algorithm>
// #include <optional>
// #include <sfl/small_vector.hpp>
// namespace OpenLoco::Gfx
// forward: class DrawingContext;
// namespace OpenLoco::Ui
type WindowNumber_t = uint16
// forward: struct Window;
// forward: struct Viewport;
type WindowColour int

const (
	Primary WindowColour = iota
	Secondary
	Tertiary
	Quaternary
	Count
)
type WindowFlags int

const (
	None WindowFlags = 0
	StickToBack WindowFlags = 1 << 0
	StickToFront WindowFlags = 1 << 1
	ViewportNoScrolling WindowFlags = 1 << 2
	ScrollingToLocation WindowFlags = 1 << 3
	Transparent WindowFlags = 1 << 4
	NoBackground WindowFlags = 1 << 5
	FramedWidgets WindowFlags = 1 << 6
	IgnoreInFindAt WindowFlags = 1 << 7
	ViewportNoShiftPixels WindowFlags = 1 << 8
	Resizable WindowFlags = 1 << 9
	NoAutoClose WindowFlags = 1 << 10
	LighterFrame WindowFlags = 1 << 11
	PlaySoundOnOpen WindowFlags = 1 << 12
	OpenQuietly WindowFlags = 1 << 13
	NotScrollView WindowFlags = 1 << 14
	FinishedResize WindowFlags = 1 << 15
	BeingResized WindowFlags = 1 << 16
	WhiteBorderOne WindowFlags = 1 << 17
	WhiteBorderMask WindowFlags = whiteBorderOne | (1 << 18)
	HasStoredState WindowFlags = 1 << 31
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(WindowFlags);
type WindowEventList struct {
// void (*onClose)(Window&) = nullptr;
// void (*onMouseUp)(Window&, WidgetIndex_t, WidgetId) = nullptr;
// void (*onResize)(Window&) = nullptr;
// void (*onMouseHover)(Window&, WidgetIndex_t, WidgetId) = nullptr;
// void (*onMouseDown)(Window&, WidgetIndex_t, WidgetId) = nullptr;
// void (*onDropdown)(Window&, WidgetIndex_t, WidgetId, int16_t) = nullptr;
// void (*onPeriodicUpdate)(Window&) = nullptr;
// void (*onUpdate)(Window&) = nullptr;
// void (*event_08)(Window&) = nullptr;
// void (*event_09)(Window&) = nullptr;
// void (*onToolUpdate)(Window&, const WidgetIndex_t, WidgetId, const int16_t, const int16_t) = nullptr;
// void (*onToolDown)(Window&, const WidgetIndex_t, WidgetId, const int16_t, const int16_t) = nullptr;
// void (*toolDrag)(Window&, const WidgetIndex_t, WidgetId, const int16_t, const int16_t) = nullptr;
// void (*toolUp)(Window&, const WidgetIndex_t, WidgetId, const int16_t, const int16_t) = nullptr;
// void (*onToolAbort)(Window&, const WidgetIndex_t, WidgetId) = nullptr;
// Ui::CursorId (*toolCursor)(Window&, const int16_t x, const int16_t y, const Ui::CursorId, bool&) = nullptr;
// void (*getScrollSize)(Window&, uint32_t scrollIndex, int32_t& scrollWidth, int32_t& scrollHeight) = nullptr;
// void (*scrollMouseDown)(Ui::Window&, int16_t x, int16_t y, uint8_t scrollIndex) = nullptr;
// void (*scrollMouseDrag)(Ui::Window&, int16_t x, int16_t y, uint8_t scrollIndex) = nullptr;
// void (*scrollMouseOver)(Ui::Window& window, int16_t x, int16_t y, uint8_t scrollIndex) = nullptr;
// void (*textInput)(Window&, WidgetIndex_t, WidgetId, const char*) = nullptr;
// void (*viewportRotate)(Window&) = nullptr;
	Event_22 uint32
// std::optional<FormatArguments> (*tooltip)(Window&, WidgetIndex_t, WidgetId) = nullptr;
// Ui::CursorId (*cursor)(Window&, WidgetIndex_t, WidgetId, int16_t, int16_t, Ui::CursorId) = nullptr;
// void (*onMove)(Window&, const int16_t x, const int16_t y) = nullptr;
// void (*prepareDraw)(Window&) = nullptr;
// void (*draw)(Window&, Gfx::DrawingContext&) = nullptr;
// void (*drawScroll)(Window&, Gfx::DrawingContext&, const uint32_t scrollIndex) = nullptr;
// bool (*keyUp)(Window&, uint32_t charCode, uint32_t keyCode) = nullptr;
}
type SavedViewSimple struct {
	ViewX coord_t
	ViewY coord_t
	ZoomLevel ZoomLevel
	Rotation int8
}
type SavedView struct {
	MapX coord_t
	MapY coord_t
	EntityId EntityId
	Flags uint16
	ZoomLevel ZoomLevel
	Rotation int8
	SurfaceZ int16
	// method: constexpr SavedView() = default;
	// method: constexpr SavedView(coord_t mapX, coord_t mapY, ZoomLevel zoomLevel, int8_t rotation, coord_t surfaceZ)
// : mapX(mapX)
// , mapY(mapY)
// , zoomLevel(zoomLevel)
// , rotation(rotation)
// , surfaceZ(surfaceZ) {};
	// method: constexpr SavedView(EntityId entityId, uint16_t flags, ZoomLevel zoomLevel, int8_t rotation, coord_t surfaceZ)
// : entityId(entityId)
// , flags(flags)
// , zoomLevel(zoomLevel)
// , rotation(rotation)
// , surfaceZ(surfaceZ) {};
	// method: constexpr bool isEmpty() const
// return mapX == -1 && mapY == -1 && entityId == EntityId::null;
}
// func IsEntityView() bool
// return (flags & (1 << 15)) != 0;
// constexpr World::Pos3 getPos() const
// if (isEntityView())
// return {};
// return { mapX, mapY, surfaceZ };
// func Clear() 
// mapX = -1;
// mapY = -1;
// entityId = EntityId::null;
// auto operator<=>(const SavedView& other) const = default;
type Window struct {
// sfl::small_vector<Widget, 16> widgets;
// const WindowEventList* eventHandlers;
// uint64_t disabledWidgets = 0;
// uint64_t activatedWidgets = 0;
// uint64_t holdableWidgets = 0;
// union
// std::byte* object;
// struct
	Var_85A int16
	Var_85C int16
}
const WindowMaxScrollAreas int = 2
// orphan member: uintptr_t info;
// Ui::Viewport* viewports[2] = { nullptr, nullptr };
// orphan member: SavedView savedView;
// int16_t expandContentCounter = 0; // Used to delay content expand when hovering over expandable scroll content
// bool showTownNames = false;       // Map window only
// orphan member: WindowFlags flags;
// WindowNumber_t number = 0;
// orphan member: int32_t x;
// orphan member: int32_t y;
// orphan member: int32_t width;
// orphan member: int32_t height;
// orphan member: int32_t minWidth;
// orphan member: int32_t maxWidth;
// orphan member: int32_t minHeight;
// orphan member: int32_t maxHeight;
// ScrollArea scrollAreas[kMaxScrollAreas];
// int16_t rowInfo[1000];
// orphan member: uint16_t rowCount;
// orphan member: uint16_t var_83C;
// orphan member: uint16_t rowHeight;
// int16_t rowHover = -1;
// union
// int16_t orderTableIndex = -1;
// orphan member: int16_t selectedTileIndex;
// orphan member: uint16_t sortMode;
// uint16_t var_846 = 0;
// uint16_t var_850 = 0; // vehicle list filter
// uint16_t var_852 = 0; // vehicle list cargo; news ticks
// uint16_t var_854 = 0; // used to limit updates
// union
// orphan member: uint16_t filterLevel;     // ObjectSelectionWindow
// orphan member: uint16_t numTicksVisible; // TimePanel
// uint16_t var_858 = 0;
// uint16_t currentTab = 0;
// uint16_t frameNo = 0;
// uint16_t currentSecondaryTab = 0;
// orphan member: int16_t var_88A; // map window orig width
// orphan member: int16_t var_88C; // map window orig height
// ViewportConfig viewportConfigurations[2];
// orphan member: WindowType type;
// CompanyId owner = CompanyId::null;
// uint8_t var_885 = 0xFF;
// AdvancedColour colours[enumValue(WindowColour::count)];
// Window(Ui::Point position, Ui::Size size);
// // TODO: Remove this once position is a member.
// constexpr Ui::Point position() const
// return { x, y };
// // TODO: Remove this once size is a member.
// constexpr Ui::Size size() const
// return { width, height };
// func SetWidgets(newWidgets any /* std::span<Widget> */ ) 
// widgets.clear();
// widgets.insert(widgets.end(), newWidgets.begin(), newWidgets.end());
// func SetSize(minSize Ui::Size, maxSize Ui::Size) bool
// bool hasResized = false;
// minWidth = minSize.width;
// minHeight = minSize.height;
// maxWidth = maxSize.width;
// maxHeight = maxSize.height;
// if (width < minWidth)
// width = minWidth;
// invalidate();
// hasResized = true;
// func If(maxWidth width >) else
// width = maxWidth;
// invalidate();
// hasResized = true;
// if (height < minHeight)
// height = minHeight;
// invalidate();
// hasResized = true;
// func If(maxHeight height >) else
// height = maxHeight;
// invalidate();
// hasResized = true;
// orphan member: return hasResized;
// func SetSize(size Ui::Size) 
// setSize(size, size);
// func GetColour(index WindowColour) AdvancedColour
// if (index >= WindowColour::primary && index < WindowColour::count)
// return colours[enumValue(index)];
// return colours[enumValue(WindowColour::primary)];
// func SetColour(index WindowColour, colour AdvancedColour) 
// if (index >= WindowColour::primary && index < WindowColour::count)
// colours[enumValue(index)] = colour;
// func HasFlags(flagsToTest WindowFlags) bool
// return (flags & flagsToTest) != WindowFlags::none;
// func IsVisible() bool
// func IsTranslucent() bool
// func IsEnabled(widgetIndex WidgetIndex_t) bool
// func IsDisabled(widgetIndex WidgetIndex_t) bool
// func IsActivated(index WidgetIndex_t) bool
// func IsHoldable(index WidgetIndex_t) bool
// func CanResize() bool
// func CapSize(minWidth int32, minHeight int32, maxWidth int32, maxHeight int32) 
// func ViewportsUpdatePosition() 
// func InvalidatePressedImageButtons() 
// func Invalidate() 
// func UpdateScrollWidgets() 
// func InitScrollWidgets() 
// func GetScrollDataIndex(index WidgetIndex_t) int8
// func SetDisabledWidgetsAndInvalidate(_disabledWidgets uint32) 
// func ViewportCentreMain() 
// func ViewportSetUndergroundFlag(underground bool, vp Ui::Viewport) 
// func ViewportGetMapCoordsByCursor(mapX int16, mapY int16, offsetX int16, offsetY int16) 
// func MoveWindowToLocation(pos viewport_pos) 
// func ViewportCentreOnTile(loc World::Pos3) 
// func ViewportCentreTileAroundCursor(mapX int16, mapY int16, offsetX int16, offsetY int16) 
// func ViewportZoomSet(zoomLevel int8, toCursor bool) 
// func ViewportZoomIn(toCursor bool) 
// func ViewportZoomOut(toCursor bool) 
// func ViewportRotateRight() 
// func ViewportRotateLeft() 
// func ViewportRotate(directionRight bool) 
// func ViewportRemove(viewportId uint8) 
// func ViewportFromSavedView(savedView SavedViewSimple) 
// func Move(dx int16, dy int16) bool
// func MoveInsideScreenEdges() 
// func MoveToCentre() bool
// func FindWidgetAt(xPos int16, yPos int16) WidgetIndex_t
// func Draw(drawingCtx Gfx::DrawingContext) 
// func CallClose() 
// func CallOnMouseUp(widgetIndex WidgetIndex_t, id WidgetId) 
// Ui::Window* callOnResize();                                                                                       // 2
// func CallOnMouseHover(widgetIndex WidgetIndex_t, id WidgetId) 
// func CallOnMouseDown(widgetIndex WidgetIndex_t, id WidgetId) 
// func CallOnDropdown(widgetIndex WidgetIndex_t, id WidgetId, itemIndex int16) 
// func CallOnPeriodicUpdate() 
// func CallUpdate() 
// func Call_8() 
// func Call_9() 
// func CallToolUpdate(widgetIndex WidgetIndex_t, id WidgetId, xPos int16, yPos int16) 
// func CallToolDown(widgetIndex WidgetIndex_t, id WidgetId, xPos int16, yPos int16) 
// func CallToolDrag(widgetIndex WidgetIndex_t, id WidgetId, xPos int16, yPos int16) 
// func CallToolUp(widgetIndex WidgetIndex_t, id WidgetId, xPos int16, yPos int16) 
// func CallToolAbort(widgetIndex WidgetIndex_t, id WidgetId) 
// Ui::CursorId callToolCursor(int16_t xPos, int16_t yPos, Ui::CursorId fallback, bool* out);                        // 15
// func CallGetScrollSize(scrollIndex uint32, scrollWidth int32, scrollHeight int32) 
// func CallScrollMouseDown(x int16, y int16, scrollIndex uint8) 
// func CallScrollMouseDrag(x int16, y int16, scrollIndex uint8) 
// func CallScrollMouseOver(x int16, y int16, scrollIndex uint8) 
// func CallTextInput(caller WidgetIndex_t, id WidgetId, buffer byte) 
// func CallViewportRotate() 
// std::optional<FormatArguments> callTooltip(WidgetIndex_t widgetIndex, WidgetId id);                               // 23
// Ui::CursorId callCursor(WidgetIndex_t widgetIdx, WidgetId id, int16_t xPos, int16_t yPos, Ui::CursorId fallback); // 24
// func CallOnMove(xPos int16, yPos int16) 
// func CallPrepareDraw() 
// func CallDraw(ctx Gfx::DrawingContext) 
// func CallDrawScroll(drawingCtx Gfx::DrawingContext, scrollIndex uint32) 
// func CallKeyUp(charCode uint32, keyCode uint32) bool
// func FirstActivatedWidgetInRange(minIndex WidgetIndex_t, maxIndex WidgetIndex_t) WidgetIndex_t
// func PrevAvailableWidgetInRange(minIndex WidgetIndex_t, maxIndex WidgetIndex_t) WidgetIndex_t
// func NextAvailableWidgetInRange(minIndex WidgetIndex_t, maxIndex WidgetIndex_t) WidgetIndex_t
// World::Pos2 viewportCoordToMapCoord(int16_t x, int16_t y, int16_t z, int32_t rotation);
// std::optional<World::Pos2> screenGetMapXyWithZ(const Point& mouse, const int16_t z);
