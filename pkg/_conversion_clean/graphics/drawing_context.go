package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Font.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/PaletteMap.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/Ui/Rect.hpp>
// #include <cstdint>
// namespace OpenLoco::Gfx
type RectInsetFlags int

const (
func init() {
	FillTransparent RectInsetFlags = 1 << 2
	BorderNone      RectInsetFlags = 1 << 3
	FillNone        RectInsetFlags = 1 << 4
	BorderInset     RectInsetFlags = 1 << 5
	FillDarker      RectInsetFlags = 1 << 6
	ColourLight     RectInsetFlags = 1 << 7
	None            RectInsetFlags = 0
}

)

// OPENLOCO_ENABLE_ENUM_OPERATORS(RectInsetFlags);
type RectFlags int

func init() {
	CrossHatching RectFlags = 1 << 24
	Transparent   RectFlags = 1 << 25
	SelectPattern RectFlags = 1 << 26
	G1Pattern     RectFlags = 1 << 27
	None          RectFlags = 0


// OPENLOCO_ENABLE_ENUM_OPERATORS(RectFlags);
type DrawingContext struct {
	// virtual ~DrawingContext() = default;
	// virtual void pushRenderTarget(const RenderTarget& rt) = 0;
	// virtual void popRenderTarget() = 0;
	// virtual const RenderTarget& currentRenderTarget() const = 0;
	// virtual void clear(uint32 fill) = 0;
	// virtual void clearSingle(uint8 paletteId) = 0;
	// virtual void fillRect(int16 left, int16 top, int16 right, int16 bottom, uint8 colour, RectFlags flags) = 0;
	// method: void fillRect(const Ui::Point& origin, const Ui::Size& size, uint8 colour, RectFlags flags)
	// fillRect(origin.x, origin.y, origin.x + size.width - 1, origin.y + size.height - 1, colour, flags);

// virtual void drawRect(int16 x, int16 y, uint16 dx, uint16 dy, uint8 colour, RectFlags flags) = 0;
// func DrawRect(origin Ui::Point, size Ui::Size, colour uint8, flags RectFlags)
// drawRect(origin.x, origin.y, size.width, size.height, colour, flags);
// virtual void fillRectInset(int16 left, int16 top, int16 right, int16 bottom, AdvancedColour colour, RectInsetFlags flags) = 0;
// func FillRectInset(origin Ui::Point, size Ui::Size, colour AdvancedColour, flags RectInsetFlags)
// fillRectInset(origin.x, origin.y, origin.x + size.width - 1, origin.y + size.height - 1, colour, flags);
// virtual void drawRectInset(int16 x, int16 y, uint16 dx, uint16 dy, AdvancedColour colour, RectInsetFlags flags) = 0;
// func DrawRectInset(origin Ui::Point, size Ui::Size, colour AdvancedColour, flags RectInsetFlags)
// drawRectInset(origin.x, origin.y, size.width, size.height, colour, flags);
// virtual void drawLine(const Ui::Point& a, const Ui::Point& b, PaletteIndex_t colour) = 0;
// virtual void drawImage(int16 x, int16 y, uint32 image) = 0;
// func DrawImage(pos Ui::Point, image uint32)
// drawImage(pos, ImageId::fromUInt32(image));
// virtual void drawImage(const Ui::Point& pos, const ImageId& image) = 0;
// virtual void drawImageMasked(const Ui::Point& pos, const ImageId& image, const ImageId& maskImage) = 0;
// virtual void drawImageSolid(const Ui::Point& pos, const ImageId& image, PaletteIndex_t paletteIndex) = 0;
// virtual void drawImagePaletteSet(const Ui::Point& pos, const ImageId& image, PaletteMap::View palette, const G1Element* noiseImage) = 0;