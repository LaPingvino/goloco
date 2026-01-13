package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/ImageId.h"
// #include "Graphics/PaletteMap.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/Ui/Point.hpp>
// #include <OpenLoco/Engine/Ui/Size.hpp>
// namespace OpenLoco::Gfx
// forward: struct G1Element;
// forward: struct RenderTarget;
type DrawSpritePosArgs struct {
// Ui::Point srcPos;
// Ui::Point dstPos;
// Ui::Size size;
}
type DrawSpriteArgs struct {
// const PaletteMap::View palMap;
// const G1Element& sourceImage;
// Ui::Point srcPos;
// Ui::Point dstPos;
// Ui::Size size;
// const G1Element* noiseImage;
// DrawSpriteArgs(
// const PaletteMap::View _palMap, const G1Element& _sourceImage, const Ui::Point& _srcPos, const Ui::Point& _dstPos, const Ui::Size& _size, const G1Element* _noiseImage)
// : palMap(_palMap)
// , sourceImage(_sourceImage)
// , srcPos(_srcPos)
// , dstPos(_dstPos)
// , size(_size)
// , noiseImage(_noiseImage)
}
type DrawBlendOp int

const (
	None DrawBlendOp = 0
// /**
// * Only supported by BITMAP. RLE images always encode transparency via the encoding.
// * Pixel value of 0 represents transparent.
// */
	Transparent DrawBlendOp = 1 << 0
// /**
// * Whether to use the pixel value from the source image.
// * This is usually only unset for glass images where there the src is only a transparency mask.
// */
	Src DrawBlendOp = 1 << 1
// /**
// * Whether to use the pixel value of the destination image for blending.
// * This is used for any image that filters the target image, e.g. glass or water.
// */
	Dst DrawBlendOp = 1 << 2
// /**
// * Whether to use the noise image to prevent draws on certain parts of the image.
// */
	NoiseMask DrawBlendOp = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(DrawBlendOp);
// func GetDrawBlendOp(image ImageId, args DrawSpriteArgs) DrawBlendOp
// template<uint8_t TZoomLevel, bool TIsRLE>
// func DrawSpriteToBuffer(rt RenderTarget, args DrawSpriteArgs, op DrawBlendOp) 
