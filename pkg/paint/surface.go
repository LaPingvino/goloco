package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Surface cliff edges — a port of PaintSurface.cpp paintSurfaceCliffEdge /
// paintSurfaceCliffEdgeImpl / paintEdgeSection and getTileDescriptor's edge
// heights. PARTIAL: rotation 0 only; tunnels (track/road entering a slope)
// and the underground view are not ported yet.

// Surface is what the cliff code needs of a tile's surface element.
type Surface struct {
	BaseZ uint8 // SmallZ (×4 world units)
	Slope uint8 // 5-bit slope (corner flags + steep bit 0x10)
}

const (
	microZStep        = 16 // World::kMicroZStep
	microToSmallZStep = 4  // World::kMicroToSmallZStep
)

// cornerHeights is kCornerHeights: top, right, bottom, left per slope.
var cornerHeights = [32][4]uint8{
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1}, {1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 1},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 1, 1}, {1, 1, 0, 0}, {1, 1, 1, 0}, {1, 1, 0, 1}, {1, 1, 1, 1},
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1}, {1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 2},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 2, 1}, {1, 1, 0, 0}, {1, 2, 1, 0}, {2, 1, 0, 1}, {1, 1, 1, 1},
}

const (
	cTop = iota
	cRight
	cBottom
	cLeft
)

// NeighbourOffsets are kNeighbourOffsets[0]: SW, SE, NW, NE (world units).
var NeighbourOffsets = [4][2]int{{32, 0}, {0, 32}, {0, -32}, {-32, 0}}

var (
	edgeFactorOffset = [4]int{0, 16, 16, 0}
	edgeImageOffset  = [4]Pos3{{30, 0, 0}, {0, 30, 0}, {0, -2, 1}, {-2, 0, 1}}
	edgeBBoxSize     = [4]Pos3{{0, 30, 15}, {30, 0, 15}, {30, 0, 15}, {0, 30, 15}}
	// kEdgeMaskImageFromSlope[not underground]: cliffEdge0MaskSlope0 = 3726,
	// cliffEdge1MaskSlope0 = 3731 (ImageIds.h, disc layout) + 0..4.
	edgeMaskBase = [4]int{3726, 3731, 3731, 3726}
)

// edgeHeight is EdgeHeight (MicroZ).
type edgeHeight struct{ self0, neighbour0, self1, neighbour1 int }

func corners(s Surface) [4]int {
	mz := int(s.BaseZ) / microToSmallZStep
	ch := cornerHeights[s.Slope&0x1F]
	return [4]int{mz + int(ch[0]), mz + int(ch[1]), mz + int(ch[2]), mz + int(ch[3])}
}

// edgeHeightFor is getTileDescriptor<dir>'s edgeHeight.
func edgeHeightFor(self [4]int, nb Surface, dir int) edgeHeight {
	n := corners(nb)
	switch dir {
	case 0: // SW
		return edgeHeight{self[cLeft], n[cTop], self[cBottom], n[cRight]}
	case 1: // SE
		return edgeHeight{self[cRight], n[cTop], self[cBottom], n[cLeft]}
	case 2: // NW
		return edgeHeight{self[cTop], n[cRight], self[cLeft], n[cBottom]}
	}
	// NE
	return edgeHeight{self[cTop], n[cLeft], self[cRight], n[cBottom]}
}

func (s *Session) paintEdgeSection(sprite func(int) *gfx8.Sprite, cliffBase, factor, height, edge, maskID int) {
	img := sprite(cliffBase + factor + (height & 0xF))
	mask := sprite(maskID)
	if img == nil || mask == nil {
		return
	}
	off := edgeImageOffset[edge]
	off.Z += height * microZStep
	s.AddAsParent(Image{Sprite: img, Mask: mask}, off, off, edgeBBoxSize[edge])
}

// paintCliffEdgeImpl is paintSurfaceCliffEdgeImpl (without tunnels/underground).
func (s *Session) paintCliffEdgeImpl(sprite func(int) *gfx8.Sprite, edge int, eh edgeHeight, cliffBase int) {
	if eh.self0 <= eh.neighbour0 && eh.self1 <= eh.neighbour1 {
		return
	}
	factor := ((s.spriteX ^ s.spriteY) & 0x20) + edgeFactorOffset[edge]
	mask := func(i int) int { return edgeMaskBase[edge] + i }
	h := eh.neighbour1
	if h != eh.neighbour0 {
		unk := 3
		if h >= eh.neighbour0 {
			unk = 4
			h = eh.neighbour0
		}
		if h != eh.self0 && h != eh.self1 {
			s.paintEdgeSection(sprite, cliffBase, factor, h, edge, mask(unk))
			h++
		}
	}
	for h < eh.self0 && h < eh.self1 {
		s.paintEdgeSection(sprite, cliffBase, factor, h, edge, mask(0))
		h++
	}
	unk := 1
	if h >= eh.self0 {
		unk = 2
		if h >= eh.self1 {
			return
		}
	}
	s.paintEdgeSection(sprite, cliffBase, factor, h, edge, mask(unk))
}

// PaintCliffEdges paints the cliff faces of the tile at the current sprite
// position against its neighbours (nil = no surface), in upstream's edge order
// 2, 3, 0, 1. cliffBase is the land object's cliffEdgeImage.
func (s *Session) PaintCliffEdges(self Surface, neighbours [4]*Surface, cliffBase int, sprite func(int) *gfx8.Sprite) {
	sc := corners(self)
	for _, edge := range [4]int{2, 3, 0, 1} {
		nb := neighbours[edge]
		if nb == nil {
			continue
		}
		s.paintCliffEdgeImpl(sprite, edge, edgeHeightFor(sc, *nb, edge), cliffBase)
	}
}
