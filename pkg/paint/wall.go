package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Walls — a port of Paint/PaintWall.cpp paintWall (0x004C3D7C). PARTIAL: the
// glass second image (translucency Colours::getGlass) and the secondary colour
// (the wall element's secondary colour isn't parsed yet) are not drawn; ghost
// walls are not ported.

// WallEl is what paintWall reads from a wall element.
type WallEl struct {
	Rotation      uint8
	EdgeSlope     uint8 // EdgeSlope flags: 1 upwards, 2 downwards
	PrimaryColour uint8
	BaseZ         uint8 // SmallZ
}

// WallObj is what paintWall reads from a wall object.
type WallObj struct {
	Sprite     int
	Height     uint8 // SmallZ
	TwoSided   bool
	HasPrimary bool
	HasGlass   bool
}

var (
	wallOffsets   = [4]Pos3{{0, 0, 0}, {1, 31, 0}, {31, 0, 0}, {2, 1, 0}}
	wallBBOffsets = [4]Pos3{{1, 1, 1}, {2, 30, 1}, {30, 2, 1}, {1, 1, 1}}
	wallBBLengths = [4]Pos3{{1, 28, 0}, {29, 1, 0}, {1, 29, 0}, {28, 1, 0}}
	// kImageOffsets[isTwoSided][rotation][slopeIndex]
	wallImageOffsets = [2][4][3]int{
		{{WallSlopedNE, WallSlopedSW, WallFlatNE}, {WallSlopedSE, WallSlopedNW, WallFlatSE}, {WallSlopedSW, WallSlopedNE, WallFlatNE}, {WallSlopedNW, WallSlopedSE, WallFlatSE}},
		{{WallSlopedNE, WallSlopedSW, WallFlatNE}, {WallGlassSlopedSE, WallGlassSlopedNW, WallGlassFlatSE}, {WallGlassSlopedSW, WallGlassSlopedNE, WallGlassFlatNE}, {WallSlopedNW, WallSlopedSE, WallFlatSE}},
	}
)

// slopeFlagsToIndex: 0 downwards, 1 upwards, 2 no slope.
func slopeFlagsToIndex(f uint8) int {
	switch {
	case f&2 != 0:
		return 0
	case f&1 != 0:
		return 1
	}
	return 2
}

// PaintWall paints one wall element at the current sprite position. colourMap
// is the remap for the primary colour (used when the object has one).
func (s *Session) PaintWall(el WallEl, obj WallObj, colourMap []uint8, sprite func(int) *gfx8.Sprite) {
	rot := int((s.Rotation + el.Rotation) & 3)
	h := int(el.BaseZ) * 4
	off, bbOff, bbLen := wallOffsets[rot], wallBBOffsets[rot], wallBBLengths[rot]
	off.Z += h
	bbOff.Z += h
	bbLen.Z += 4*int(obj.Height) - 2
	two := 0
	if obj.TwoSided {
		two = 1
	}
	sp := sprite(obj.Sprite + wallImageOffsets[two][rot][slopeFlagsToIndex(el.EdgeSlope)])
	if sp == nil {
		return
	}
	img := Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}
	if obj.HasPrimary && colourMap != nil {
		img = Image{Sprite: sp, Op: gfx8.OpFor(sp, true, false), PalMap: colourMap}
	}
	s.AddAsParent(img, off, bbOff, bbLen)
}
