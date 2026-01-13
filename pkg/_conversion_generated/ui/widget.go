package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Gfx.h"
// #include "Graphics/ImageIds.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringIds.h"
// #include "Localisation/StringManager.h"
// #include <array>
// #include <cstdint>
// namespace OpenLoco::Gfx
type RectInsetFlags int

const (
// forward: class DrawingContext;
)
// namespace OpenLoco::Ui
type WidgetIndex_t = int16
const WidgetIndexNull WidgetIndex_t = -1
var ResizeHandleSize = 18 // auto
// namespace Detail
// // TODO: Move this to a more appropriate location.
// func GetHashFNV1a(s byte, int) uint64
// // FNV-1a hash
var Prime = 0x00000100000001B3 // auto
var OffsetBasis = 0xCBF29CE484222325 // auto
// auto res = kOffsetBasis;
// for (size_t i = 0; s[i] != '\0'; i++)
// res ^= s[i];
// res *= kPrime;
// orphan member: return res;
type WidgetId struct {
type ValueType int

const (
	None ValueType = 0
)
	Value ValueType
	// method: constexpr WidgetId() = default;
	// method: constexpr WidgetId(ValueType value)
// : _value{ value }
}
const WidgetIdNone any = ValueType.none
// template<size_t TSize>
// func WidgetId((str byte) constexpr
// : _value{ static_cast<ValueType>(Detail::getHashFNV1a(str, TSize)) }
// func WidgetId(str byte) constexpr
// : _value{ static_cast<ValueType>(Detail::getHashFNV1a(str, 0)) }
// constexpr auto operator<=>(const WidgetId&) const = default;
// // This makes switch statements work.
// func Uint64_t() operator
// return static_cast<uint64_t>(_value);
// forward: struct Window;
type WindowColour int

const (
type ContentAlign int

const (
	Left ContentAlign = 0
	Center
	Right
)
type WidgetType int

const (
	Empty WidgetType = 0
	Panel
	NewsPanel
	Frame
	Wt_3
	Slider
	Wt_6
	ToolbarTab
	Tab
	ButtonWithImage
	ButtonWithColour
	Button
	Label
	ButtonTableHeader
	Groupbox
	Textbox
	Combobox
	Viewport
	Caption
	Scrollview
	Checkbox
)
type WidgetState struct {
// Window* window{};
// Gfx::RectInsetFlags flags{};
	Colour AdvancedColour
	Disabled bool
	Activated bool
	Hovered bool
	ScrollviewIndex int
}
// forward: struct Widget;
type WidgetEventsList struct {
// void (*draw)(Gfx::DrawingContext&, const Widget&, const WidgetState&) = nullptr;
}
type Widget struct {
// // Indicates that the imageId has a colour set and not to replace it with the window colour
// // This reuses the ImageIdFlags::translucent flag for use in widget draw
// // Flag *MUST* be removed before passing to drawingCtx.drawImage functions
	// method: constexpr Widget(WidgetId widgetId, Ui::Point origin, Ui::Size size, WidgetType widgetType, WindowColour colour, uint32_t content = Widget::kContentNull, StringId tooltip = StringIds::null)
// : id{ widgetId }
// , content{ content }
// , left{ static_cast<int16_t>(origin.x) }
// , right{ static_cast<int16_t>(origin.x + size.width - 1) }
// , top{ static_cast<int16_t>(origin.y) }
// , bottom{ static_cast<int16_t>(origin.y + size.height - 1) }
// , tooltip{ tooltip }
// , type{ widgetType }
// , windowColour{ colour }
}
const WidgetImageIdColourSet uint32 = (1 << 30)
const WidgetContentNull uint32 = 0xFFFFFFFFU
const WidgetContentUnk uint32 = 0xFFFFFFFEU
// func Widget(widgetId WidgetId, origin Ui::Point, size Ui::Size, widgetType WidgetType, colour WindowColour, content StringId, StringIds::null StringId tooltip =) constexpr
// : id{ widgetId }
// , text{ content }
// , left{ static_cast<int16_t>(origin.x) }
// , right{ static_cast<int16_t>(origin.x + size.width - 1) }
// , top{ static_cast<int16_t>(origin.y) }
// , bottom{ static_cast<int16_t>(origin.y + size.height - 1) }
// , tooltip{ tooltip }
// , type{ widgetType }
// , windowColour{ colour }
// func Widget() constexpr
// orphan member: WidgetId id{ WidgetId::none };
// orphan member: FormatArgumentsBuffer textArgs;
// orphan member: WidgetEventsList events;
// union
// orphan member: uint32_t image{ ImageIds::null };
// orphan member: StringId text;
// orphan member: uint32_t content;
// orphan member: int32_t left{};
// orphan member: int32_t right{};
// orphan member: int32_t top{};
// orphan member: int32_t bottom{};
// Gfx::Font font{ Gfx::Font::medium_bold };
// orphan member: StringId tooltip{ StringIds::null };
// orphan member: WidgetType type{};
// orphan member: ContentAlign contentAlign{ ContentAlign::left };
// orphan member: WindowColour windowColour{};
// func MidX() int32
// func MidY() int32
// func Width() int32
// func Height() int32
// // Custom widget attributes.
// orphan member: uint32_t styleData{};
// // Widget state.
// bool disabled : 1 {};
// bool activated : 1 {};
// bool hidden : 1 {};
// // TODO: Remove this once position is a member.
// Ui::Point position() const
// return { left, top };
// // TODO: Remove this once size is a member.
// Ui::Size size() const
// return { width(), height() };
// // TODO: Make tabs actual widgets.
// func DrawTab(w Window, drawingCtx Gfx::DrawingContext, imageId uint32, index WidgetIndex_t) 
// // typical tab width, to be used in most (all?) cases
const DefaultTabWidth uint16 = 30
// func LeftAlignTabs(window Window, firstTabIndex uint8, lastTabIndex uint8, kDefaultTabWidth uint16_t tabWidth =) 
// func Draw(drawingCtx Gfx::DrawingContext, window Window, pressedWidgets uint64, toolWidgets uint64, hoveredWidgets uint64, scrollviewIndex uint8) 
// func MakeWidget(origin Ui::Point, size Ui::Size, type WidgetType, colour WindowColour, Widget::kContentNull uint32_t content =, StringIds::null StringId tooltip =) Widget
// orphan member: Widget out{ WidgetId::none, origin, size, type, colour, content, tooltip };
// orphan member: return out;
// namespace Detail
// template<typename T, typename Enable = void>
type WidgetsCount struct {
}
const WidgetsCountCount int = 0
// template<typename T>
type WidgetsCount struct {
}
const WidgetsCountCount int = 1
// template<typename T, std::size_t N>
type WidgetsCount struct {
}
const WidgetsCountCount int = N
// template<typename T>
type IsWidgetsArray struct {
	Std // embedded
}
// template<std::size_t N>
type IsWidgetsArray struct {
}
// template<typename... TArgs>
// func MakeWidgets(args TArgs...) any
// constexpr auto totalCount = [&]() {
// size_t count = 0;
// ((count += Detail::WidgetsCount<std::decay_t<decltype(args)>>::count), ...);
// orphan member: return count;
// }();
// std::array<Widget, totalCount> res{};
// size_t index = 0;
// const auto append = [&](auto&& val) {
// func Constexpr(any /* Detail::IsWidgetsArray<std::decay_t<decltype(val */ ) if
// for (auto&& widget : val)
// res[index] = std::move(widget);
// index++;
// else
// res[index] = std::move(val);
// index++;
// ((append(args)), ...);
// orphan member: return res;
