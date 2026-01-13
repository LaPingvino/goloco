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
	None TextDrawFlags = 0
	Inset TextDrawFlags = (1 << 0)
	Outline TextDrawFlags = (1 << 1)
	Dark TextDrawFlags = (1 << 2)
	ExtraDark TextDrawFlags = (1 << 3)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TextDrawFlags);
type TextRenderer struct {
// DrawingContext& _ctx;
	CurrentFontFlags TextDrawFlags
	CurrentFontSpriteBase Font
// TextRenderer(DrawingContext& ctx);
	// method: Font getCurrentFont() const;
	// method: void setCurrentFont(Font base);
	// method: int16_t clipString(int16_t width, char* string) const;
	// method: static int16_t clipString(Font font, int16_t width, char* string);
	// method: uint16_t getStringWidth(const char* buffer) const;
	// method: static uint16_t getStringWidth(Font base, const char* buffer);
	// method: uint16_t getMaxStringWidth(const char* buffer) const;
	// method: static uint16_t getMaxStringWidth(Font font, const char* buffer);
	// method: uint16_t getStringWidthNewLined(const char* buffer) const;
	// method: static uint16_t getStringWidthNewLined(Font font, const char* buffer);
// std::pair<uint16_t, uint16_t> wrapString(char* buffer, uint16_t stringWidth) const;
// static std::pair<uint16_t, uint16_t> wrapString(Font font, char* buffer, uint16_t stringWidth);
	// method: static uint16_t getLineHeight(Font font);
	// method: static uint16_t getSmallerLineHeight(Font font);
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
// uint16_t width,
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
// uint16_t width,
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
// uint16_t width,
// AdvancedColour colour,
// StringId stringId,
// FormatArgumentsView args = {});
// Ui::Point drawStringCentredRaw(
// Ui::Point origin,
// uint16_t linebreakCount,
// AdvancedColour colour,
// const char* wrappedStr);
// Ui::Point drawStringCentredWrapped(
// Ui::Point origin,
// uint16_t width,
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
	// method: void drawStringTicker(Ui::Point origin, StringId stringId, Colour colour, uint8_t numLinesToDisplay, uint16_t numCharactersToDisplay, uint16_t width);
}
