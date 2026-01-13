package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "DrawSprite.h"
// #include "DrawSpriteHelper.hpp"
// #include "Graphics/Gfx.h"
// #include "Graphics/RenderTarget.h"
// namespace OpenLoco::Gfx
// template<DrawBlendOp TBlendOp, uint8_t TZoomLevel>
// func DrawBMPSprite(rt RenderTarget, args DrawSpriteArgs) 
// const auto& g1 = args.sourceImage;
// const auto* src = g1.offset + ((static_cast<size_t>(g1.width) * args.srcPos.y) + args.srcPos.x);
// const auto& paletteMap = args.palMap;
// const int32_t width = args.size.width;
// int32_t height = args.size.height;
// const size_t srcLineWidth = g1.width << TZoomLevel;
// const size_t dstLineWidth = (static_cast<size_t>(rt.width) >> TZoomLevel) + rt.pitch;
// auto* dst = rt.bits;
// // Move the pointer to the start point of the destination
// dst += dstLineWidth * args.dstPos.y + args.dstPos.x;
var Zoom = 1 << TZoomLevel // auto
// func Constexpr(DrawBlendOp::noiseMask (TBlendOp) if
// const auto* noiseMask = args.noiseImage->offset + ((static_cast<size_t>(g1.width) * args.srcPos.y) + args.srcPos.x);
// for (; height > 0; height -= zoom)
// auto* nextSrc = src + srcLineWidth;
// auto* nextDst = dst + dstLineWidth;
// auto* nextNoiseMask = noiseMask + srcLineWidth;
// for (int32_t widthRemaining = width; widthRemaining > 0; widthRemaining -= zoom, src += zoom, noiseMask += zoom, dst++)
// blitPixel<TBlendOp>(*src, *dst, paletteMap, *noiseMask);
// src = nextSrc;
// dst = nextDst;
// noiseMask = nextNoiseMask;
// else
// for (; height > 0; height -= zoom)
// auto* nextSrc = src + srcLineWidth;
// auto* nextDst = dst + dstLineWidth;
// for (int32_t widthRemaining = width; widthRemaining > 0; widthRemaining -= zoom, src += zoom, dst++)
// blitPixel<TBlendOp>(*src, *dst, paletteMap, 0xFF);
// src = nextSrc;
// dst = nextDst;
