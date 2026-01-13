package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Font.h"
// #include "ImageId.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/Ui/Rect.hpp>
// #include <array>
// #include <cstdint>
// #include <optional>
// #include <span>
// namespace OpenLoco
type PaletteIndex_t = uint8
// forward: struct AdvancedColour;
type ExtColour int

const (
)
// namespace OpenLoco::Gfx
// forward: class SoftwareDrawingEngine;
// namespace OpenLoco::Gfx
// forward: struct RenderTarget;
// namespace G1ExpectedCount
const Disc uint32 = 0x101A
const Steam uint32 = 0x0F38
const Objects uint32 = 0x40000
type G1Header struct {
	NumEntries uint32
	TotalSize uint32
}
type G1ElementFlags int

const (
	None G1ElementFlags = 0
	HasTransparency G1ElementFlags = 1 << 0
	Unk1 G1ElementFlags = 1 << 1
	IsRLECompressed G1ElementFlags = 1 << 2
	IsR8G8B8Palette G1ElementFlags = 1 << 3
	HasZoomSprites G1ElementFlags = 1 << 4
	NoZoomDraw G1ElementFlags = 1 << 5
	DuplicatePrevious G1ElementFlags = 1 << 6
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(G1ElementFlags);
type G1Element32 struct {
	Offset uint32
	Width int16
	Height int16
	XOffset int16
	YOffset int16
	Flags G1ElementFlags
	ZoomOffset int16
}
// // A version that can be 64-bit when ready...
type G1Element struct {
// uint8_t* offset = nullptr;
// int16_t width = 0;
// int16_t height = 0;
// int16_t xOffset = 0;
// int16_t yOffset = 0;
// G1ElementFlags flags = G1ElementFlags::none;
// int16_t zoomOffset = 0;
// G1Element() = default;
// G1Element(const G1Element32& src)
// : offset(reinterpret_cast<uint8_t*>(static_cast<uintptr_t>(src.offset)))
// , width(src.width)
// , height(src.height)
// , xOffset(src.xOffset)
// , yOffset(src.yOffset)
// , flags(src.flags)
// , zoomOffset(src.zoomOffset)
}
// func HasFlags(flagsToTest G1ElementFlags) bool
// return (flags & flagsToTest) != G1ElementFlags::none;
type PaletteEntry struct {
	B uint8
	G uint8
	R uint8
	A uint8
}
type ImageExtents struct {
	Width uint8
	HeightNegative uint8
	HeightPositive uint8
}
// namespace ImageIdFlags
const Remap uint32 = 1 << 29
const Translucent uint32 = 1 << 30
const Remap2 uint32 = 1 << 31
// func LoadG1() 
// func Initialise() 
// // TODO: Move the recolour functions into Colour.h
// [[nodiscard]] constexpr uint32_t recolour(uint32_t image)
// return ImageIdFlags::remap | image;
// [[nodiscard]] constexpr uint32_t recolour(uint32_t image, ExtColour colour)
// return ImageIdFlags::remap | (enumValue(colour) << 19) | image;
// [[nodiscard]] constexpr uint32_t recolour(uint32_t image, Colour colour)
// func Recolour(image, Colours::toExt(colour) return
// [[nodiscard]] constexpr uint32_t recolour2(uint32_t image, Colour colour1, Colour colour2)
// return ImageIdFlags::remap | ImageIdFlags::remap2 | (enumValue(colour1) << 19) | (enumValue(colour2) << 24) | image;
// [[nodiscard]] constexpr uint32_t recolour2(uint32_t image, ColourScheme colourScheme)
// func Recolour2(image, colourScheme.primary, colourScheme.secondary) return
// [[nodiscard]] constexpr uint32_t recolourTranslucent(uint32_t image, ExtColour colour)
// return ImageIdFlags::translucent | (enumValue(colour) << 19) | image;
// [[nodiscard]] ImageId applyGhostToImage(uint32_t imageIndex);
// [[nodiscard]] constexpr uint32_t getImageIndex(uint32_t imageId) { return imageId & 0x7FFFF; }
// // Invalidates the entire screen.
// func InvalidateScreen() 
// // Invalidate a region of the screen.
// func InvalidateRegion(left int32, top int32, right int32, bottom int32) 
// // Renders all invalidated regions the next frame.
// func Render() 
// // Renders a region of the screen.
// func Render(rect Ui::Rect) 
// func Render(left int16, top int16, right int16, bottom int16) 
// // Moves the pixels on screen.
// func MovePixelsOnScreen(dstX int16, dstY int16, width int16, height int16, srcX int16, srcY int16) 
// // Renders all invalidated regions and processes new messages.
// func RenderAndUpdate() 
// G1Element* getG1Element(uint32_t id);
// Gfx::SoftwareDrawingEngine& getDrawingEngine();
// func LoadCurrency() 
// func LoadDefaultPalette() 
// func LoadPalette(imageIndex uint32, modifier uint8) 
// std::span<const PaletteEntry> getRgbaPalette();
// func GetImagesMaxExtent(baseImageId ImageId, numImages int) ImageExtents
// func GetCharacterWidth(font Font, character char32_t) int16
// func SetCharacterWidth(font Font, character char32_t, width int16) 
// func GetImageForCharacter(font Font, character char32_t) ImageId
