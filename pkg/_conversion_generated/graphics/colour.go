package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// namespace OpenLoco
type PaletteIndex_t = uint8
type Colour int

const (
	Black Colour = 0
	Grey Colour = 1
	White Colour = 2
	MutedDarkPurple Colour = 3
	MutedPurple Colour = 4
	Purple Colour = 5
	DarkBlue Colour = 6
	Blue Colour = 7
	MutedDarkTeal Colour = 8
	MutedTeal Colour = 9
	DarkGreen Colour = 10
	MutedSeaGreen Colour = 11
	MutedGrassGreen Colour = 12
	Green Colour = 13
	MutedAvocadoGreen Colour = 14
	MutedOliveGreen Colour = 15
	Yellow Colour = 16
	DarkYellow Colour = 17
	Orange Colour = 18
	Amber Colour = 19
	DarkOrange Colour = 20
	MutedDarkYellow Colour = 21
	MutedYellow Colour = 22
	Brown Colour = 23
	MutedOrange Colour = 24
	MutedDarkRed Colour = 25
	DarkRed Colour = 26
	Red Colour = 27
	DarkPink Colour = 28
	Pink Colour = 29
	MutedRed Colour = 30
	Max
)
type ExtColour int

const (
	Black ExtColour = 0
	Grey ExtColour = 1
	White ExtColour = 2
	MutedDarkPurple ExtColour = 3
	MutedPurple ExtColour = 4
	Purple ExtColour = 5
	DarkBlue ExtColour = 6
	Blue ExtColour = 7
	MutedDarkTeal ExtColour = 8
	MutedTeal ExtColour = 9
	DarkGreen ExtColour = 10
	MutedSeaGreen ExtColour = 11
	MutedGrassGreen ExtColour = 12
	Green ExtColour = 13
	MutedAvocadoGreen ExtColour = 14
	MutedOliveGreen ExtColour = 15
	Yellow ExtColour = 16
	DarkYellow ExtColour = 17
	Orange ExtColour = 18
	Amber ExtColour = 19
	DarkOrange ExtColour = 20
	MutedDarkYellow ExtColour = 21
	MutedYellow ExtColour = 22
	Brown ExtColour = 23
	MutedOrange ExtColour = 24
	MutedDarkRed ExtColour = 25
	DarkRed ExtColour = 26
	Red ExtColour = 27
	DarkPink ExtColour = 28
	Pink ExtColour = 29
	MutedRed ExtColour = 30
// // First 30 are inherited from Colour
	Clear ExtColour = 31
	Water ExtColour = 32
	Unk21
	Unk22
	Unk23
	Unk24
	Unk25
	Unk26
	Unk27
	Unk28
	Unk29
	Unk2A
	Unk2B
	Unk2C
	Unk2D
	Unk2E
	Unk2F
	Unk30
	Unk31
	Unk32
	Unk33
	Unk34
	TranslucentGrey1
	TranslucentGrey2
	TranslucentGrey0
	TranslucentBlue1
	TranslucentBlue2
	TranslucentBlue0
	TranslucentMutedDarkRed1
	TranslucentMutedDarkRed2
	TranslucentMutedDarkRed0
	TranslucentMutedSeaGreen1
	TranslucentMutedSeaGreen2
	TranslucentMutedSeaGreen0
	TranslucentMutedPurple1
	TranslucentMutedPurple2
	TranslucentMutedPurple0
	TranslucentMutedOliveGreen1
	TranslucentMutedOliveGreen2
	TranslucentMutedOliveGreen0
	TranslucentMutedYellow1
	TranslucentMutedYellow2
	TranslucentMutedYellow0
	TranslucentYellow1
	TranslucentYellow2
	TranslucentYellow0
	TranslucentMutedGrassGreen1
	TranslucentMutedGrassGreen2
	TranslucentMutedGrassGreen0
	TranslucentMutedAvocadoGreen1
	TranslucentMutedAvocadoGreen2
	TranslucentMutedAvocadoGreen0
	TranslucentGreen1
	TranslucentGreen2
	TranslucentGreen0
	TranslucentMutedOrange1
	TranslucentMutedOrange2
	TranslucentMutedOrange0
	TranslucentPurple1
	TranslucentPurple2
	TranslucentPurple0
	TranslucentRed1
	TranslucentRed2
	TranslucentRed0
	TranslucentOrange1
	TranslucentOrange2
	TranslucentOrange0
	TranslucentMutedTeal1
	TranslucentMutedTeal2
	TranslucentMutedTeal0
	TranslucentPink1
	TranslucentPink2
	TranslucentPink0
	TranslucentBrown1
	TranslucentBrown2
	TranslucentBrown0
	TranslucentMutedRed1
	TranslucentMutedRed2
	TranslucentMutedRed0
	TranslucentWhite1
	TranslucentWhite2
	TranslucentWhite0
	TranslucentAmber1
	TranslucentAmber2
	TranslucentAmber0
	Unk74
	Unk75
	Unk76
	Unk77
	Unk78
	Unk79
	Unk7A
	Unk7B
	Unk7C
	Unk7D
	Unk7E
	Unk7F
	Unk80
	Unk81
	Unk82
	Unk83
	Unk84
	Unk85
	Unk86
	Unk87
	Unk88
	Unk89
	Unk8A
	Unk8B
	Unk8C
	Unk8D
	Unk8E
	Unk8F
	Unk90
	Unk91
	Unk92
	Max
)
type AdvancedColour struct {
// Colour _c = Colour::black;
	// method: constexpr AdvancedColour() = default;
	// method: constexpr AdvancedColour(const Colour c)
// : _c(c)
}
const AdvancedColourOutlineFlag uint8 = 1 << 5
const AdvancedColourInsetFlag uint8 = 1 << 6
const AdvancedColourTranslucentFlag uint8 = 1 << 7
const AdvancedColourFd uint8 = 0xFD
const AdvancedColourFe uint8 = 0xFE
const AdvancedColourFf uint8 = 0xFF
// constexpr explicit operator Colour() const { return static_cast<Colour>(enumValue(_c) & ~(outlineFlag | insetFlag | translucentFlag)); }
// // Returns the Colour without any additional flags set.
// [[nodiscard]] constexpr Colour c() const { return static_cast<Colour>(*this); }
// constexpr explicit operator uint8_t() const { return enumValue(_c); }
// [[nodiscard]] constexpr uint8_t u8() const { return static_cast<uint8_t>(*this); }
// [[nodiscard]] constexpr AdvancedColour outline() const
// return { static_cast<Colour>(enumValue(_c) | outlineFlag) };
// [[nodiscard]] constexpr bool isOutline() const { return enumValue(_c) & outlineFlag; }
// [[nodiscard]] constexpr AdvancedColour inset() const
// return { static_cast<Colour>(enumValue(_c) | insetFlag) };
// [[nodiscard]] constexpr bool isInset() const { return enumValue(_c) & insetFlag; }
// [[nodiscard]] constexpr AdvancedColour translucent() const
// return { static_cast<Colour>(enumValue(_c) | translucentFlag) };
// [[nodiscard]] constexpr bool isTranslucent() const { return enumValue(_c) & translucentFlag; }
// [[nodiscard]] constexpr AdvancedColour opaque() const
// return { static_cast<Colour>(enumValue(_c) & ~translucentFlag) };
// [[nodiscard]] constexpr bool isOpaque() const { return !isTranslucent(); }
// [[nodiscard]] constexpr AdvancedColour clearInset() const
// return { static_cast<Colour>(enumValue(_c) & ~insetFlag) };
// [[nodiscard]] constexpr AdvancedColour clearOutline() const
// return { static_cast<Colour>(enumValue(_c) & ~outlineFlag) };
// [[nodiscard]] static constexpr AdvancedColour FF()
// return { static_cast<Colour>(ff) };
// [[nodiscard]] constexpr bool isFF() const { return enumValue(_c) == ff; }
// [[nodiscard]] static constexpr AdvancedColour FE()
// return { static_cast<Colour>(fe) };
// [[nodiscard]] constexpr bool isFE() const { return enumValue(_c) == fe; }
// [[nodiscard]] static constexpr AdvancedColour FD()
// return { static_cast<Colour>(fd) };
// [[nodiscard]] constexpr bool isFD() const { return enumValue(_c) == fd; }
// static_assert(sizeof(AdvancedColour) == 1);
// namespace Colours
// func InitColourMap() 
// func GetShade(colour Colour, shade uint8) PaletteIndex_t
// func GetTranslucent(colour Colour) ExtColour
// func GetTranslucent(colour Colour, shade uint8) ExtColour
// func GetShadow(colour Colour) ExtColour
// func GetGlass(colour Colour) ExtColour
func ToExt(c Colour) ExtColour {
}
// namespace PaletteIndex
// // All indexes are of the form {swatch colour name}{shade}
// // Shades are from dark 0x0 to light 0xB
// // Most swatches are 0xB in length apart from:
// //   transparent (1),
// //   grey (4),
// //   brightYellow (3),
// //   textRemap (6)
// //
// // There are 3 unused indexes:
// //   textRemap4,
// //   textRemap5,
// //   hotPinkUnused
// // 0xFF is technically unused as well
const Transparent PaletteIndex_t = 0
const TextRemap0 PaletteIndex_t = 0x01
const TextRemap1 PaletteIndex_t = 0x02
const TextRemap2 PaletteIndex_t = 0x03
const TextRemap3 PaletteIndex_t = 0x04
const TextRemap4 PaletteIndex_t = 0x05
const TextRemap5 PaletteIndex_t = 0x06
const PrimaryRemap0 PaletteIndex_t = 0x07
const PrimaryRemap1 PaletteIndex_t = 0x08
const PrimaryRemap2 PaletteIndex_t = 0x09
const Black0 PaletteIndex_t = 0x0A
const Black1 PaletteIndex_t = 0x0B
const Black2 PaletteIndex_t = 0x0C
const Black3 PaletteIndex_t = 0x0D
const Black4 PaletteIndex_t = 0x0E
const Black5 PaletteIndex_t = 0x0F
const Black6 PaletteIndex_t = 0x10
const Black7 PaletteIndex_t = 0x11
const Black8 PaletteIndex_t = 0x12
const Black9 PaletteIndex_t = 0x13
const BlackA PaletteIndex_t = 0x14
const BlackB PaletteIndex_t = 0x15
const MutedOliveGreen0 PaletteIndex_t = 0x16
const MutedOliveGreen1 PaletteIndex_t = 0x17
const MutedOliveGreen2 PaletteIndex_t = 0x18
const MutedOliveGreen3 PaletteIndex_t = 0x19
const MutedOliveGreen4 PaletteIndex_t = 0x1A
const MutedOliveGreen5 PaletteIndex_t = 0x1B
const MutedOliveGreen6 PaletteIndex_t = 0x1C
const MutedOliveGreen7 PaletteIndex_t = 0x1D
const MutedOliveGreen8 PaletteIndex_t = 0x1E
const MutedOliveGreen9 PaletteIndex_t = 0x1F
const MutedOliveGreenA PaletteIndex_t = 0x20
const MutedOliveGreenB PaletteIndex_t = 0x21
const MutedDarkYellow0 PaletteIndex_t = 0x22
const MutedDarkYellow1 PaletteIndex_t = 0x23
const MutedDarkYellow2 PaletteIndex_t = 0x24
const MutedDarkYellow3 PaletteIndex_t = 0x25
const MutedDarkYellow4 PaletteIndex_t = 0x26
const MutedDarkYellow5 PaletteIndex_t = 0x27
const MutedDarkYellow6 PaletteIndex_t = 0x28
const MutedDarkYellow7 PaletteIndex_t = 0x29
const MutedDarkYellow8 PaletteIndex_t = 0x2A
const MutedDarkYellow9 PaletteIndex_t = 0x2B
const MutedDarkYellowA PaletteIndex_t = 0x2C
const MutedDarkYellowB PaletteIndex_t = 0x2D
const Yellow0 PaletteIndex_t = 0x2E
const Yellow1 PaletteIndex_t = 0x2F
const Yellow2 PaletteIndex_t = 0x30
const Yellow3 PaletteIndex_t = 0x31
const Yellow4 PaletteIndex_t = 0x32
const Yellow5 PaletteIndex_t = 0x33
const Yellow6 PaletteIndex_t = 0x34
const Yellow7 PaletteIndex_t = 0x35
const Yellow8 PaletteIndex_t = 0x36
const Yellow9 PaletteIndex_t = 0x37
const YellowA PaletteIndex_t = 0x38
const YellowB PaletteIndex_t = 0x39
const MutedDarkRed0 PaletteIndex_t = 0x3A
const MutedDarkRed1 PaletteIndex_t = 0x3B
const MutedDarkRed2 PaletteIndex_t = 0x3C
const MutedDarkRed3 PaletteIndex_t = 0x3D
const MutedDarkRed4 PaletteIndex_t = 0x3E
const MutedDarkRed5 PaletteIndex_t = 0x3F
const MutedDarkRed6 PaletteIndex_t = 0x40
const MutedDarkRed7 PaletteIndex_t = 0x41
const MutedDarkRed8 PaletteIndex_t = 0x42
const MutedDarkRed9 PaletteIndex_t = 0x43
const MutedDarkRedA PaletteIndex_t = 0x44
const MutedDarkRedB PaletteIndex_t = 0x45
const MutedGrassGreen0 PaletteIndex_t = 0x46
const MutedGrassGreen1 PaletteIndex_t = 0x47
const MutedGrassGreen2 PaletteIndex_t = 0x48
const MutedGrassGreen3 PaletteIndex_t = 0x49
const MutedGrassGreen4 PaletteIndex_t = 0x4A
const MutedGrassGreen5 PaletteIndex_t = 0x4B
const MutedGrassGreen6 PaletteIndex_t = 0x4C
const MutedGrassGreen7 PaletteIndex_t = 0x4D
const MutedGrassGreen8 PaletteIndex_t = 0x4E
const MutedGrassGreen9 PaletteIndex_t = 0x4F
const MutedGrassGreenA PaletteIndex_t = 0x50
const MutedGrassGreenB PaletteIndex_t = 0x51
const MutedAvocadoGreen0 PaletteIndex_t = 0x52
const MutedAvocadoGreen1 PaletteIndex_t = 0x53
const MutedAvocadoGreen2 PaletteIndex_t = 0x54
const MutedAvocadoGreen3 PaletteIndex_t = 0x55
const MutedAvocadoGreen4 PaletteIndex_t = 0x56
const MutedAvocadoGreen5 PaletteIndex_t = 0x57
const MutedAvocadoGreen6 PaletteIndex_t = 0x58
const MutedAvocadoGreen7 PaletteIndex_t = 0x59
const MutedAvocadoGreen8 PaletteIndex_t = 0x5A
const MutedAvocadoGreen9 PaletteIndex_t = 0x5B
const MutedAvocadoGreenA PaletteIndex_t = 0x5C
const MutedAvocadoGreenB PaletteIndex_t = 0x5D
const Green0 PaletteIndex_t = 0x5E
const Green1 PaletteIndex_t = 0x5F
const Green2 PaletteIndex_t = 0x60
const Green3 PaletteIndex_t = 0x61
const Green4 PaletteIndex_t = 0x62
const Green5 PaletteIndex_t = 0x63
const Green6 PaletteIndex_t = 0x64
const Green7 PaletteIndex_t = 0x65
const Green8 PaletteIndex_t = 0x66
const Green9 PaletteIndex_t = 0x67
const GreenA PaletteIndex_t = 0x68
const GreenB PaletteIndex_t = 0x69
const MutedOrange0 PaletteIndex_t = 0x6A
const MutedOrange1 PaletteIndex_t = 0x6B
const MutedOrange2 PaletteIndex_t = 0x6C
const MutedOrange3 PaletteIndex_t = 0x6D
const MutedOrange4 PaletteIndex_t = 0x6E
const MutedOrange5 PaletteIndex_t = 0x6F
const MutedOrange6 PaletteIndex_t = 0x70
const MutedOrange7 PaletteIndex_t = 0x71
const MutedOrange8 PaletteIndex_t = 0x72
const MutedOrange9 PaletteIndex_t = 0x73
const MutedOrangeA PaletteIndex_t = 0x74
const MutedOrangeB PaletteIndex_t = 0x75
const MutedPurple0 PaletteIndex_t = 0x76
const MutedPurple1 PaletteIndex_t = 0x77
const MutedPurple2 PaletteIndex_t = 0x78
const MutedPurple3 PaletteIndex_t = 0x79
const MutedPurple4 PaletteIndex_t = 0x7A
const MutedPurple5 PaletteIndex_t = 0x7B
const MutedPurple6 PaletteIndex_t = 0x7C
const MutedPurple7 PaletteIndex_t = 0x7D
const MutedPurple8 PaletteIndex_t = 0x7E
const MutedPurple9 PaletteIndex_t = 0x7F
const MutedPurpleA PaletteIndex_t = 0x80
const MutedPurpleB PaletteIndex_t = 0x81
const Blue0 PaletteIndex_t = 0x82
const Blue1 PaletteIndex_t = 0x83
const Blue2 PaletteIndex_t = 0x84
const Blue3 PaletteIndex_t = 0x85
const Blue4 PaletteIndex_t = 0x86
const Blue5 PaletteIndex_t = 0x87
const Blue6 PaletteIndex_t = 0x88
const Blue7 PaletteIndex_t = 0x89
const Blue8 PaletteIndex_t = 0x8A
const Blue9 PaletteIndex_t = 0x8B
const BlueA PaletteIndex_t = 0x8C
const BlueB PaletteIndex_t = 0x8D
const MutedSeaGreen0 PaletteIndex_t = 0x8E
const MutedSeaGreen1 PaletteIndex_t = 0x8F
const MutedSeaGreen2 PaletteIndex_t = 0x90
const MutedSeaGreen3 PaletteIndex_t = 0x91
const MutedSeaGreen4 PaletteIndex_t = 0x92
const MutedSeaGreen5 PaletteIndex_t = 0x93
const MutedSeaGreen6 PaletteIndex_t = 0x94
const MutedSeaGreen7 PaletteIndex_t = 0x95
const MutedSeaGreen8 PaletteIndex_t = 0x96
const MutedSeaGreen9 PaletteIndex_t = 0x97
const MutedSeaGreenA PaletteIndex_t = 0x98
const MutedSeaGreenB PaletteIndex_t = 0x99
const Purple0 PaletteIndex_t = 0x9A
const Purple1 PaletteIndex_t = 0x9B
const Purple2 PaletteIndex_t = 0x9C
const Purple3 PaletteIndex_t = 0x9D
const Purple4 PaletteIndex_t = 0x9E
const Purple5 PaletteIndex_t = 0x9F
const Purple6 PaletteIndex_t = 0xA0
const Purple7 PaletteIndex_t = 0xA1
const Purple8 PaletteIndex_t = 0xA2
const Purple9 PaletteIndex_t = 0xA3
const PurpleA PaletteIndex_t = 0xA4
const PurpleB PaletteIndex_t = 0xA5
const Red0 PaletteIndex_t = 0xA6
const Red1 PaletteIndex_t = 0xA7
const Red2 PaletteIndex_t = 0xA8
const Red3 PaletteIndex_t = 0xA9
const Red4 PaletteIndex_t = 0xAA
const Red5 PaletteIndex_t = 0xAB
const Red6 PaletteIndex_t = 0xAC
const Red7 PaletteIndex_t = 0xAD
const Red8 PaletteIndex_t = 0xAE
const Red9 PaletteIndex_t = 0xAF
const RedA PaletteIndex_t = 0xB0
const RedB PaletteIndex_t = 0xB1
const Orange0 PaletteIndex_t = 0xB2
const Orange1 PaletteIndex_t = 0xB3
const Orange2 PaletteIndex_t = 0xB4
const Orange3 PaletteIndex_t = 0xB5
const Orange4 PaletteIndex_t = 0xB6
const Orange5 PaletteIndex_t = 0xB7
const Orange6 PaletteIndex_t = 0xB8
const Orange7 PaletteIndex_t = 0xB9
const Orange8 PaletteIndex_t = 0xBA
const Orange9 PaletteIndex_t = 0xBB
const OrangeA PaletteIndex_t = 0xBC
const OrangeB PaletteIndex_t = 0xBD
const MutedDarkTeal0 PaletteIndex_t = 0xBE
const MutedDarkTeal1 PaletteIndex_t = 0xBF
const MutedDarkTeal2 PaletteIndex_t = 0xC0
const MutedDarkTeal3 PaletteIndex_t = 0xC1
const MutedDarkTeal4 PaletteIndex_t = 0xC2
const MutedDarkTeal5 PaletteIndex_t = 0xC3
const MutedDarkTeal6 PaletteIndex_t = 0xC4
const MutedDarkTeal7 PaletteIndex_t = 0xC5
const MutedDarkTeal8 PaletteIndex_t = 0xC6
const MutedDarkTeal9 PaletteIndex_t = 0xC7
const MutedDarkTealA PaletteIndex_t = 0xC8
const MutedDarkTealB PaletteIndex_t = 0xC9
const Pink0 PaletteIndex_t = 0xCA
const Pink1 PaletteIndex_t = 0xCB
const Pink2 PaletteIndex_t = 0xCC
const Pink3 PaletteIndex_t = 0xCD
const Pink4 PaletteIndex_t = 0xCE
const Pink5 PaletteIndex_t = 0xCF
const Pink6 PaletteIndex_t = 0xD0
const Pink7 PaletteIndex_t = 0xD1
const Pink8 PaletteIndex_t = 0xD2
const Pink9 PaletteIndex_t = 0xD3
const PinkA PaletteIndex_t = 0xD4
const PinkB PaletteIndex_t = 0xD5
const SecondaryRemap0 PaletteIndex_t = 0xCA
const SecondaryRemap1 PaletteIndex_t = 0xCB
const SecondaryRemap2 PaletteIndex_t = 0xCC
const SecondaryRemap3 PaletteIndex_t = 0xCD
const SecondaryRemap4 PaletteIndex_t = 0xCE
const SecondaryRemap5 PaletteIndex_t = 0xCF
const SecondaryRemap6 PaletteIndex_t = 0xD0
const SecondaryRemap7 PaletteIndex_t = 0xD1
const SecondaryRemap8 PaletteIndex_t = 0xD2
const SecondaryRemap9 PaletteIndex_t = 0xD3
const SecondaryRemapA PaletteIndex_t = 0xD4
const SecondaryRemapB PaletteIndex_t = 0xD5
const Brown0 PaletteIndex_t = 0xD6
const Brown1 PaletteIndex_t = 0xD7
const Brown2 PaletteIndex_t = 0xD8
const Brown3 PaletteIndex_t = 0xD9
const Brown4 PaletteIndex_t = 0xDA
const Brown5 PaletteIndex_t = 0xDB
const Brown6 PaletteIndex_t = 0xDC
const Brown7 PaletteIndex_t = 0xDD
const Brown8 PaletteIndex_t = 0xDE
const Brown9 PaletteIndex_t = 0xDF
const BrownA PaletteIndex_t = 0xE0
const BrownB PaletteIndex_t = 0xE1
const Grey0 PaletteIndex_t = 0xE2
const BrightYellow0 PaletteIndex_t = 0xE3
const BrightYellow1 PaletteIndex_t = 0xE4
const BrightYellow2 PaletteIndex_t = 0xE5
const Amber0 PaletteIndex_t = 0xE6
const Amber1 PaletteIndex_t = 0xE7
const Amber2 PaletteIndex_t = 0xE8
const Amber3 PaletteIndex_t = 0xE9
const Amber4 PaletteIndex_t = 0xEA
const Amber5 PaletteIndex_t = 0xEB
const Amber6 PaletteIndex_t = 0xEC
const Amber7 PaletteIndex_t = 0xED
const Amber8 PaletteIndex_t = 0xEE
const Amber9 PaletteIndex_t = 0xEF
const Grey1 PaletteIndex_t = 0xF0
const Grey2 PaletteIndex_t = 0xF1
const Grey3 PaletteIndex_t = 0xF2
const AmberA PaletteIndex_t = 0xF3
const AmberB PaletteIndex_t = 0xF4
const UnusedHotPink PaletteIndex_t = 0xF5
const PrimaryRemap3 PaletteIndex_t = 0xF6
const PrimaryRemap4 PaletteIndex_t = 0xF7
const PrimaryRemap5 PaletteIndex_t = 0xF8
const PrimaryRemap6 PaletteIndex_t = 0xF9
const PrimaryRemap7 PaletteIndex_t = 0xFA
const PrimaryRemap8 PaletteIndex_t = 0xFB
const PrimaryRemap9 PaletteIndex_t = 0xFC
const PrimaryRemapA PaletteIndex_t = 0xFD
const PrimaryRemapB PaletteIndex_t = 0xFE
const Index_FF PaletteIndex_t = 0xFF
