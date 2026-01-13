package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "DrawSprite.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/PaletteMap.h"
// namespace OpenLoco::Gfx
// func Blend(paletteMap PaletteMap::View, src uint8, dst uint8) uint8
// // src = 0 would be transparent so there is no blend palette for that, hence src - 1
// assert(src != 0);
// // src is treated as a row in the palette map, validate its in range.
// const auto row = src - 1u;
// assert(row < (paletteMap.size() / PaletteMap::kDefaultSize));
// const auto idx = (row * PaletteMap::kDefaultSize) + dst;
// assert(idx < paletteMap.size());
// return paletteMap[idx];
// template<DrawBlendOp TBlendOp>
// func BlitPixel(src uint8, dst uint8, paletteMap [[maybe_unused]] PaletteMap::View, noiseMask uint8) bool
// func Constexpr(DrawBlendOp::noiseMask (TBlendOp) if
// // noiseMask is either 0 or 0xFF
// src &= noiseMask;
// func Constexpr(DrawBlendOp::transparent (TBlendOp) if
// // Ignore transparent pixels
// if (src == PaletteIndex::transparent)
// orphan member: return false;
// func Constexpr(DrawBlendOp::src ((TBlendOp) if
// auto pixel = blend(paletteMap, src, dst);
// func Constexpr(DrawBlendOp::transparent (TBlendOp) if
// if (pixel == PaletteIndex::transparent)
// orphan member: return false;
// dst = pixel;
// orphan member: return true;
// else if constexpr ((TBlendOp & DrawBlendOp::src) != DrawBlendOp::none)
// auto pixel = paletteMap[src];
// func Constexpr(DrawBlendOp::transparent (TBlendOp) if
// if (pixel == PaletteIndex::transparent)
// orphan member: return false;
// dst = pixel;
// orphan member: return true;
// else if constexpr ((TBlendOp & DrawBlendOp::dst) != DrawBlendOp::none)
// auto pixel = paletteMap[dst];
// func Constexpr(DrawBlendOp::transparent (TBlendOp) if
// if (pixel == PaletteIndex::transparent)
// orphan member: return false;
// dst = pixel;
// orphan member: return true;
// else
// dst = src;
// orphan member: return true;
