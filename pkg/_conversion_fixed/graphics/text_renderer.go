package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "DrawingContext.h"
// #include "Font.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/PaletteMap.h"
// #include "Types.hpp"
// #include <OpenLoco/Engine/Ui/Rect.hpp>
// #include <cstdint>
// namespace OpenLoco::Gfx
// forward: class DrawingContext;
// // Make this maybe public?
type TextDrawFlags int

const (
func init() {
	None      TextDrawFlags = 0
	Inset     TextDrawFlags = (1 << 0)
	Outline   TextDrawFlags = (1 << 1)
	Dark      TextDrawFlags = (1 << 2)
	ExtraDark TextDrawFlags = (1 << 3)
}

)

// OPENLOCO_ENABLE_ENUM_OPERATORS(TextDrawFlags);
type TextRenderer struct {
	// DrawingContext& _ctx;
	CurrentFontFlags      TextDrawFlags
	CurrentFontSpriteBase Font
	// TextRenderer(DrawingContext& ctx);
	// method: Font getCurrentFont() const;
	// method: void setCurrentFont(Font base);
	// method: int16 clipString(int16 width, char* string) const;
	// method: static int16 clipString(Font font, int16 width, char* string);
	// method: uint16 getStringWidth(const char* buffer) const;
	// method: static uint16 getStringWidth(Font base, const char* buffer);
	// method: uint16 getMaxStringWidth(const char* buffer) const;
	// method: static uint16 getMaxStringWidth(Font font, const char* buffer);
	// method: uint16 getStringWidthNewLined(const char* buffer) const;
	// method: static uint16 getStringWidthNewLined(Font font, const char* buffer);
	// std::pair<uint16, uint16> wrapString(char* buffer, uint16 stringWidth) const;
	// static std::pair<uint16, uint16> wrapString(Font font, char* buffer, uint16 stringWidth);
	// method: static uint16 getLineHeight(Font font);
	// method: static uint16 getSmallerLineHeight(Font font);
	// Ui::Point drawString(
	// Ui::Point origin,
	// AdvancedColour colour,
	// const char* str);
	// Ui::Point drawStringLeft(
	// Ui::Point origin,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringLeftClipped(
	// Ui::Point origin,
	// uint16 width,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringLeftUnderline(
	// Ui::Point origin,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringLeftWrapped(
	// Ui::Point origin,
	// uint16 width,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringCentred(
	// Ui::Point origin,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringCentredClipped(
	// Ui::Point origin,
	// uint16 width,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringCentredRaw(
	// Ui::Point origin,
	// uint16 linebreakCount,
	// AdvancedColour colour,
	// const char* wrappedStr);
	// Ui::Point drawStringCentredWrapped(
	// Ui::Point origin,
	// uint16 width,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringRight(
	// Ui::Point origin,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// Ui::Point drawStringRightUnderline(
	// Ui::Point origin,
	// AdvancedColour colour,
	// StringId stringId,
	// FormatArgumentsView args = {});
	// method: void drawStringYOffsets(Ui::Point loc, AdvancedColour colour, const char* str, const int8_t* yOffsets);
	// method: void drawStringTicker(Ui::Point origin, StringId stringId, Colour colour, uint8 numLinesToDisplay, uint16 numCharactersToDisplay, uint16 width);
