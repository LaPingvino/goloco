package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Colour.h"
// #include <cassert>
// namespace OpenLoco
type ImageIndex = uint32
// /**
// * Represents a specific image from a catalogue such as G1, object etc. with remap
// * colours and flags.
// *
// * This is currently all stored as a single 32-bit integer, but will allow easy
// * extension to 64-bits or higher so that more images can be used.
// */
type ImageId struct {
// // clang-format off
// // clang-format on
// // Noise mask can be used with NONE and PRIMARY
// // NONE = No remap
// // BLENDED = No source copy, remap destination only (glass)
// // PRIMARY | BLENDED = Destination is blended with source (water)
// // PRIMARY = Remap with palette id (first 32 are colour palettes)
// // PRIMARY | SECONDARY = Remap with primary and secondary colours
// uint32_t _index = kImageIndexUndefined;
// [[nodiscard]] static ImageId fromUInt32(uint32_t value)
	Result ImageId
// result._index = value;
	Result return
}
const ImageIdIndexUndefined uint32 = 0b00000000000001111111111111111111
const ImageIdMaskIndex uint32 = 0b00000000000001111111111111111111
const ImageIdMaskRemap uint32 = 0b00000011111110000000000000000000
const ImageIdMaskTranslucent uint32 = 0b00000111111110000000000000000000
const ImageIdMaskPrimary uint32 = 0b00000000111110000000000000000000
const ImageIdMaskSecondary uint32 = 0b00011111000000000000000000000000
const ImageIdMaskNoiseMask uint32 = 0b00011100000000000000000000000000
const ImageIdFlagPrimary uint32 = 0b00100000000000000000000000000000
const ImageIdFlagBlend uint32 = 0b01000000000000000000000000000000
const ImageIdFlagSecondary uint32 = 0b10000000000000000000000000000000
const ImageIdShiftRemap uint32 = 19
const ImageIdShiftPrimary uint32 = 19
const ImageIdShiftTranslucent uint32 = 19
const ImageIdShiftSecondary uint32 = 24
const ImageIdShiftNoiseMask uint32 = 26
const ImageIdValueUndefined uint32 = kIndexUndefined
const ImageIdImageIndexUndefined ImageIndex = std.numeric_limits<ImageIndex>.max()
// ImageId() = default;
// explicit constexpr ImageId(ImageIndex index)
// : _index(index == kIndexUndefined ? kImageIndexUndefined : index)
// func ImageId(index uint32, palette ExtColour) constexpr
// : ImageId(ImageId(index).withRemap(palette))
// func ImageId(index uint32, primaryColour Colour) constexpr
// : ImageId(ImageId(index).withPrimary(primaryColour))
// func ImageId(index uint32, primaryColour Colour, secondaryColour Colour) constexpr
// : ImageId(ImageId(index).withPrimary(primaryColour).withSecondary(secondaryColour))
// func ImageId(index uint32, scheme ColourScheme) constexpr
// : ImageId(index, scheme.primary, scheme.secondary)
// [[nodiscard]] constexpr uint32_t toUInt32() const
// orphan member: return _index;
// [[nodiscard]] constexpr bool hasValue() const
// func GetIndex() return
// [[nodiscard]] constexpr bool hasPrimary() const
// return _index & kFlagPrimary;
// [[nodiscard]] constexpr bool hasSecondary() const
// return _index & kFlagSecondary;
// [[nodiscard]] constexpr bool hasNoiseMask() const
// return !isBlended() && !hasSecondary() && (getNoiseMask() != 0);
// [[nodiscard]] constexpr bool isRemap() const
// func HasPrimary() return
// [[nodiscard]] constexpr bool isBlended() const
// return _index & kFlagBlend;
// [[nodiscard]] constexpr ImageIndex getIndex() const
// return _index & kMaskIndex;
// [[nodiscard]] constexpr ExtColour getTranslucency() const
// return static_cast<ExtColour>((_index & kMaskTranslucent) >> kShiftTranslucent);
// [[nodiscard]] constexpr ExtColour getRemap() const
// return static_cast<ExtColour>((_index & kMaskRemap) >> kShiftRemap);
// [[nodiscard]] constexpr Colour getPrimary() const
// return static_cast<Colour>((_index & kMaskPrimary) >> kShiftPrimary);
// [[nodiscard]] constexpr Colour getSecondary() const
// return static_cast<Colour>((_index & kMaskSecondary) >> kShiftSecondary);
// [[nodiscard]] constexpr uint8_t getNoiseMask() const
// return (_index & kMaskNoiseMask) >> kShiftNoiseMask;
// [[nodiscard]] constexpr ImageId withIndex(ImageIndex index) const
// ImageId result = *this;
// result._index &= ~kMaskIndex;
// result._index |= index;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withIndexOffset(int32_t offset) const
// ImageId result = *this;
// result._index += offset;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withRemap(ExtColour paletteId) const
// ImageId result = *this;
// assert(enumValue(paletteId) <= 0x7F); // If larger then it eats into noiseMask
// result._index &= ~(kMaskRemap | kFlagSecondary | kFlagBlend);
// result._index |= enumValue(paletteId) << kShiftRemap;
// result._index |= kFlagPrimary;
// orphan member: return result;
// // Can be used withRemap or withPrimary or with None
// [[nodiscard]] constexpr ImageId withNoiseMask(uint8_t noise) const
// ImageId result = *this;
// result._index &= ~(kMaskNoiseMask | kFlagSecondary | kFlagBlend);
// result._index |= noise << kShiftNoiseMask;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withPrimary(Colour colour) const
// ImageId result = *this;
// result._index &= ~kMaskPrimary;
// result._index |= enumValue(colour) << kShiftPrimary;
// result._index |= kFlagPrimary;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withSecondary(Colour colour) const
// ImageId result = *this;
// result._index &= ~kMaskSecondary;
// result._index |= enumValue(colour) << kShiftSecondary;
// result._index |= kFlagSecondary;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withTranslucency(ExtColour colour) const
// ImageId result = *this;
// result._index &= ~(kMaskPrimary | kMaskSecondary | kFlagSecondary | kFlagPrimary);
// result._index |= enumValue(colour) << kShiftRemap;
// result._index |= kFlagBlend;
// orphan member: return result;
// [[nodiscard]] constexpr ImageId withBlend(ExtColour colour) const
// ImageId result = *this;
// result._index &= ~(kMaskPrimary | kMaskSecondary | kFlagSecondary);
// result._index |= enumValue(colour) << kShiftRemap;
// result._index |= kFlagBlend | kFlagPrimary;
// orphan member: return result;
// static_assert(sizeof(ImageId) == 4);
