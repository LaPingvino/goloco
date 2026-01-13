package localisation

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "FormatArguments.hpp"
// #include "StringManager.h"
// #include "Types.hpp"
// #include <cstddef>
// #include <cstdint>
// #include <string_view>
// #include <utility>
// #error "small is defined, likely by windows.h"
// namespace OpenLoco
// namespace ControlCodes
// // Arguments (1 byte): uint8_t
const MoveX uint8 = 1
const AdjustPalette uint8 = 2
// // 3-4 Not used
const OneArgBegin uint8 = moveX
const OneArgEnd uint8 = 4 + 1
// // Arguments: none
const Newline uint8 = 5
const NewlineSmaller uint8 = 6
// namespace Font
const Small uint8 = 7
const Large uint8 = 8
const Bold uint8 = 9
const Regular uint8 = 10
const Outline uint8 = 11
const OutlineOff uint8 = 12
const WindowColour1 uint8 = 13
const WindowColour2 uint8 = 14
const WindowColour3 uint8 = 15
const WindowColour4 uint8 = 16
const NoArgBegin uint8 = newline
const NoArgEnd uint8 = windowColour4 + 1
// // Arguments (2 bytes): int8_t, int8_t
const NewlineXY uint8 = 17
// // 18-22 Not used
const TwoArgBegin uint8 = newlineXY
const TwoArgEnd uint8 = 22 + 1
// // Arguments (4 bytes): uint32_t
const InlineSpriteStr uint8 = 23
// // 24-31 Not used
const FourArgBegin uint8 = inlineSpriteStr
const FourArgEnd uint8 = 31 + 1
// // Arguments in Args buffer
// // Note:
// // Pre formatString:
// //     ControlCodes valid args in args buffer.
// // Post formatString:
// //     ControlCodes are invalid
// //     inlineSpriteArgs replaced with inlineSpriteStr, arg is in string
const Int32_grouped uint8 = 123 + 0
const Int32_ungrouped uint8 = 123 + 1
const Int16_decimals uint8 = 123 + 2
const Int32_decimals uint8 = 123 + 3
const Int16_grouped uint8 = 123 + 4
const Uint16_ungrouped uint8 = 123 + 5
const Currency32 uint8 = 123 + 6
const Currency48 uint8 = 123 + 7
const StringidArgs uint8 = 123 + 8
const StringidStr uint8 = 123 + 9
const String_ptr uint8 = 123 + 10
const Date uint8 = 123 + 11
const Velocity uint8 = 123 + 12
const Pop16 uint8 = 123 + 13
const Push16 uint8 = 123 + 14
const TimeMS uint8 = 123 + 15
const TimeHM uint8 = 123 + 16
const Distance uint8 = 123 + 17
const Height uint8 = 123 + 18
const Power uint8 = 123 + 19
const InlineSpriteArgs uint8 = 123 + 20
// namespace Colour
// // Arguments: none
const Black uint8 = 144
const Grey uint8 = 145
const White uint8 = 146
const Red uint8 = 147
const Green uint8 = 148
const Yellow uint8 = 149
const Topaz uint8 = 150
const Celadon uint8 = 151
const BabyBlue uint8 = 152
const PaleLavender uint8 = 153
const PaleGold uint8 = 154
const LightPink uint8 = 155
const PearlAqua uint8 = 156
const PaleSilver uint8 = 157
// namespace DateModifier
const Dmy_full uint8 = 0
const My_full uint8 = 4
const My_abbr uint8 = 5
const Raw_my_abbr uint8 = 8
// namespace OpenLoco
type MonthId int

const (
)
// namespace OpenLoco::StringManager
// char* formatString(char* buffer, StringId id);
// char* formatString(char* buffer, size_t bufferLen, StringId id);
// char* formatString(char* buffer, StringId id, FormatArgumentsView args);
// char* formatString(char* buffer, size_t bufferLen, StringId id, FormatArgumentsView args);
// // TODO: Move this somewhere more sensible, the string manager should have no idea about the meaning of strings
// func IsTownName(stringId StringId) StringId
// func ToTownName(stringId StringId) StringId
// func FromTownName(stringId StringId) StringId
// std::pair<StringId, StringId> monthToString(MonthId month);
// func InternalLengthToComma1DP(length int32) int32
// func LocoStrlen(buffer byte) int
// func LocoStrlenS(buffer byte, size std::size_t) int
// char* locoStrcpy(char* dest, const char* src);
// char* locoStrcpyS(char* dest, std::size_t destSize, const char* src, std::size_t srcSize);
