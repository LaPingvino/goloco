package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Colour.h"
// #include "Localisation/StringManager.h"
// #include "Window.h"
// #include "World/Company.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstdlib>
// #include <optional>
type Format_arg_type int

const (
	U16 Format_arg_type = iota
	U32
	Ptr
)
type Format_arg struct {
	Type format_arg_type
// union
	U16 uint16
	U32 uint32
	Ptr uintptr_t
}
// format_arg(uint16_t value)
// type = format_arg_type::u16;
// u16 = value;
// format_arg(uint32_t value)
// type = format_arg_type::u32;
// u32 = value;
// format_arg(char* value)
// type = format_arg_type::ptr;
// ptr = (uintptr_t)value;
// format_arg(const char* value)
// type = format_arg_type::ptr;
// ptr = (uintptr_t)value;
// namespace OpenLoco
// forward: class FormatArguments;
// namespace OpenLoco::Ui::Dropdown
type Flags int

const (
	None Flags = 0
	Unk1 Flags = 1 << 0
	Unk2 Flags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(Flags);
// func AddSeparator(index int) 
// func Add(index int, title StringId) 
// func Add(index int, title StringId, l any /* std::initializer_list<format_arg> */ ) 
// func Add(index int, title StringId, l format_arg) 
// func Add(index int, title StringId, fArgs FormatArguments) 
// func GetHighlightedItem() int16
// func SetItemDisabled(index int) 
// func SetHighlightedItem(index int) 
// func SetItemSelected(index int) 
// func Show(x int16, y int16, width int16, height int16, colour AdvancedColour, count int, itemHeight uint8, flags uint8) 
// func Show(x int16, y int16, width int16, height int16, colour AdvancedColour, count int, flags uint8) 
// func ShowImage(x int16, y int16, width int16, height int16, heightOffset int16, colour AdvancedColour, columnCount uint8, count uint8, 0 uint8_t flags =) 
// func ShowBelow(window Window, widgetIndex WidgetIndex_t, count int, flags uint8) 
// func ShowBelow(window Window, widgetIndex WidgetIndex_t, count int, itemHeight int8, flags uint8) 
// func ShowText(x int16, y int16, width int16, height int16, itemHeight uint8, colour AdvancedColour, count int, flags uint8) 
// func ShowText(x int16, y int16, width int16, height int16, colour AdvancedColour, count int, flags uint8) 
// func ShowText2(x int16, y int16, width int16, height int16, itemHeight uint8, colour AdvancedColour, count int, flags uint8) 
// func ShowText2(x int16, y int16, width int16, height int16, colour AdvancedColour, count int, flags uint8) 
// func ShowColour(window Window, widget Widget, availableColours uint32, selectedColour Colour, dropdownColour AdvancedColour) 
// func ForceCloseCompanySelect() 
// func PopulateCompanySelect(window Window, widget Widget) 
// func GetCompanyIdFromSelection(itemIndex int16) CompanyId
// func GetItemArgument(index uint8, argument uint8) uint16
// func GetItemsPerRow(itemCount uint8) uint16
type DropdownItemId = int32
type Builder struct {
// Window* _window{};
	WidgetIndex WidgetIndex_t
// std::vector<std::tuple<std::optional<DropdownItemId>, StringId>> _items;
// std::optional<DropdownItemId> _highlightedId;
// Builder& below(Window& window, WidgetIndex_t widgetIndex);
// Builder& item(DropdownItemId id, StringId text);
// Builder& separator();
// Builder& highlight(DropdownItemId id);
	// method: void show();
// template<typename T>
// Builder& item(T id, StringId text)
// item(static_cast<int32_t>(id), text);
// return *this;
}
// template<typename T>
// Builder& highlight(T id)
// highlight(static_cast<int32_t>(id));
// return *this;
// func Create() Builder
// std::optional<DropdownItemId> getSelectedItem(int32_t index);
// template<typename T>
// std::optional<T> getSelectedItem(int32_t index)
// auto result = getSelectedItem(index);
// if (result)
// return static_cast<T>(*result);
// return std::nullopt;
// std::optional<int> dropdownIndexFromPoint(Ui::Window* window, int x, int y);
// func HasFlags(flags Flags) bool
// func SetFlags(flags Flags) 
// func SetMenuOption(index int, value uint8) 
// func GetMenuOption(index int) uint8
