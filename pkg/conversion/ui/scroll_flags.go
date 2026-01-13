package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <type_traits>
// namespace OpenLoco::Ui
type ScrollPart int

const (
	None ScrollPart = -1
	View ScrollPart = 0
	HscrollbarButtonLeft ScrollPart = 1
	HscrollbarButtonRight ScrollPart = 2
	HscrollbarTrackLeft ScrollPart = 3
	HscrollbarTrackRight ScrollPart = 4
	HscrollbarThumb ScrollPart = 5
	VscrollbarButtonTop ScrollPart = 6
	VscrollbarButtonBottom ScrollPart = 7
	VscrollbarTrackTop ScrollPart = 8
	VscrollbarTrackBottom ScrollPart = 9
	VscrollbarThumb ScrollPart = 10
)
// namespace Scrollbars
const None uint8 = 0
const Horizontal uint8 = (1 << 0)
const Vertical uint8 = (1 << 1)
const Both uint8 = horizontal | vertical
type ScrollFlags int

const (
	None ScrollFlags = 0
	HscrollbarVisible ScrollFlags = 1 << 0
	HscrollbarThumbPressed ScrollFlags = 1 << 1
	HscrollbarLeftPressed ScrollFlags = 1 << 2
	HscrollbarRightPressed ScrollFlags = 1 << 3
	VscrollbarVisible ScrollFlags = 1 << 4
	VscrollbarThumbPressed ScrollFlags = 1 << 5
	VscrollbarUpPressed ScrollFlags = 1 << 6
	VscrollbarDownPressed ScrollFlags = 1 << 7
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(ScrollFlags);
type ScrollArea struct {
	Flags ScrollFlags
	ContentOffsetX int32
	ContentWidth int32
	HThumbLeft int32
	HThumbRight int32
	ContentOffsetY int32
	ContentHeight int32
	VThumbTop int32
	VThumbBottom int32
	// method: constexpr bool hasFlags(ScrollFlags flagsToTest) const
// return (flags & flagsToTest) != ScrollFlags::none;
}
