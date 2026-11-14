package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Surface cliff edges — a port of PaintSurface.cpp paintSurfaceCliffEdge /
// paintSurfaceCliffEdgeImpl / paintEdgeSection and getTileDescriptor's edge
// heights, with the tunnel portals (track/road entering a slope, inserted by
// the elements painted before the surface). The caller hands in the slopes as
// the view shows them (getRotatedSlope) and the neighbours by
// NeighbourOffsets[rotation]. PARTIAL: the underground view; a cliff's top
// section hidden under a bridge (isCliffEdgeTopCoveredByBridge).

// Surface is what the cliff code needs of a tile's surface element.
type Surface struct {
	BaseZ uint8 // SmallZ (×4 world units)
	Slope uint8 // 5-bit slope (corner flags + steep bit 0x10)
	Water uint8 // water level in MicroZ (×16 world units); 0 = none
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

// NeighbourOffsets is kNeighbourOffsets (0x004FD99E): by rotation, the SW, SE,
// NW and NE neighbours (as the view shows them) from the tile's sprite
// position — which is the tile's corner kUnkOffsets[rotation], so the offsets
// take that back (world units).
var NeighbourOffsets = [4][4][2]int{
	{{32, 0}, {0, 32}, {0, -32}, {-32, 0}},
	{{-32, 32}, {-64, 0}, {0, 0}, {-32, -32}},
	{{-64, -32}, {-32, -64}, {-32, 0}, {0, -32}},
	{{0, -64}, {32, -32}, {-32, -32}, {0, 0}},
}

var (
	edgeFactorOffset = [4]int{0, 16, 16, 0}
	edgeImageOffset  = [4]Pos3{{30, 0, 0}, {0, 30, 0}, {0, -2, 1}, {-2, 0, 1}}
	edgeBBoxSize     = [4]Pos3{{0, 30, 15}, {30, 0, 15}, {30, 0, 15}, {0, 30, 15}}
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

// kCliffEdgeToTunnelEdge (PaintSurface.cpp).
var cliffEdgeToTunnelEdge = [4]int{2, 1, 3, 0}

// paintCliffEdgeImpl is paintSurfaceCliffEdgeImpl, including the tunnel portals
// (tunnels inserted by the track/road elements painted before the surface).
// baseHeight is the surface's (or the water's) height the edge is painted
// from. In the underground view the far edges (2 and 3) first attach the
// cliff's underground image — the ground's cut face seen from inside — to the
// previous struct, and every section takes the underground masks.
func (s *Session) paintCliffEdgeImpl(sprite func(int) *gfx8.Sprite, edge, baseHeight int, eh edgeHeight, cliffBase int) {
	if eh.self0 <= eh.neighbour0 && eh.self1 <= eh.neighbour1 {
		return
	}
	underground := s.ViewFlags.Has(ViewUnderground)
	factor := ((s.spriteX ^ s.spriteY) & 0x20) + edgeFactorOffset[edge]
	if uo := EdgeUndergroundOffset[edge]; uo != 0 && underground {
		offset := eh.self1 - eh.self0 + 1
		yOffset := (eh.self0-baseHeight/microToSmallZStep)*microZStep + 1
		if sp := sprite(cliffBase + offset + uo); sp != nil {
			s.AttachToPrevious(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, 0, -yOffset)
		}
	}
	masks := &EdgeMaskImageFromSlope[b2i(underground)][edge]
	mask := func(i int) int { return int(masks[i]) }
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
	tunnelNum := 0
	for h < eh.self0 && h < eh.self1 {
		te := cliffEdgeToTunnelEdge[edge]
		for h > s.tunnelAt(te, tunnelNum).height {
			tunnelNum++
		}
		if tn := s.tunnelAt(te, tunnelNum); h == tn.height {
			z := h * microZStep
			switch edge {
			case 0:
				s.addPortal(sprite, tn.tunnel, Pos3{30, 0, z}, Pos3{29, 0, 6 + z}, Pos3{2, 2, 25})
				s.addPortal(sprite, tn.tunnel+1, Pos3{30, 0, z}, Pos3{30, 30, z}, Pos3{1, 1, 31})
			case 1:
				s.addPortal(sprite, tn.tunnel+2, Pos3{0, 30, z}, Pos3{0, 29, 6 + z}, Pos3{2, 2, 25})
				s.addPortal(sprite, tn.tunnel+3, Pos3{0, 30, z}, Pos3{30, 30, z}, Pos3{1, 1, 31})
			}
			h += 2
			tunnelNum++
			continue
		}
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

func (s *Session) addPortal(sprite func(int) *gfx8.Sprite, img int, off, bbOff, bbSz Pos3) {
	if sp := sprite(img); sp != nil {
		s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, off, bbOff, bbSz)
	}
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
		s.paintCliffEdgeImpl(sprite, edge, int(self.BaseZ)*microToSmallZStep, edgeHeightFor(sc, *nb, edge), cliffBase)
	}
}

// Water — PaintSurface.cpp paintSurface water block (~1736-1775), the waves
// included: wave (≥ 0) is the running wave's frame on a flag-6 tile, drawn at
// zoom 0 as image + 60 + frame attached to the water.

// WaterImages are the water object's image base and its palette map (+41,
// WaterObject::kColourPalette; ExtColour::water).
type WaterImages struct {
	Base   int
	PalMap []uint8
}

// PaintWater paints the water surface of the tile at the current sprite position:
// image +35+shape as a destination mask through the water palette map, +30+shape
// attached on top, then the water cliff edges (edges 2,3,0,1). slopeToWaterShape
// is kSlopeToWaterShape.
func (s *Session) PaintWater(self Surface, neighbours [4]*Surface, w WaterImages, slopeToWaterShape []uint8, cliffBase int, wave int, sprite func(int) *gfx8.Sprite) {
	if self.Water == 0 {
		return
	}
	waterHeight := int(self.Water) * microZStep
	baseHeight := int(self.BaseZ) * microToSmallZStep
	shape := 0
	if waterHeight <= baseHeight+microZStep {
		shape = int(slopeToWaterShape[self.Slope&0xF])
	}
	blend := sprite(w.Base + shape + 35)
	if blend != nil {
		off := Pos3{0, 0, waterHeight}
		s.AddAsParent(Image{Sprite: blend, Op: gfx8.OpTransparent | gfx8.OpDst, PalMap: w.PalMap}, off, off, Pos3{32, 32, -1})
		if top := sprite(w.Base + shape + 30); top != nil {
			s.AttachToPrevious(Image{Sprite: top, Op: gfx8.OpFor(top, false, false)}, 0, 0)
		}
		if wave >= 0 && s.Zoom == 0 { // Draw waves
			if ws := sprite(w.Base + wave + 60); ws != nil {
				s.AttachToPrevious(Image{Sprite: ws, Op: gfx8.OpFor(ws, false, false)}, 0, 0)
			}
		}
	}
	if cliffBase <= 0 {
		return
	}
	sc := corners(self)
	for _, edge := range [4]int{2, 3, 0, 1} {
		nb := neighbours[edge]
		if nb == nil || int(nb.Water)*microZStep == waterHeight {
			continue
		}
		eh := edgeHeightFor(sc, *nb, edge)
		wz := waterHeight / microZStep
		s.paintCliffEdgeImpl(sprite, edge, waterHeight, edgeHeight{wz, eh.neighbour0, wz, eh.neighbour1}, cliffBase)
	}
}

// SmoothTile is what paintSurfaceSmoothenEdge reads of a tile descriptor:
// the surface, its land object (0xFF for an industry field), display slope,
// snow coverage and growth stage. Present is false for no neighbour surface.
type SmoothTile struct {
	Surface
	Present      bool
	Land         uint8
	DisplaySlope int
	Snow, Growth uint8
}

// SmoothLand is what the smoothing reads of a land object.
type SmoothLand struct {
	Image, ImagesPerGrowthStage int
	SharpSlopeTransition        bool // LandObjectFlags::hasSharpSlopeTransition
	DisableSmoothTransition     bool // LandObjectFlags::disableSmoothTileTransition
}

// PaintSurfaceSmoothing is paintSurface's landscape smoothing (zoom 0,
// Config::landscapeSmoothing, not underground): paintSurfaceSmoothenEdge for
// edges 2, 3, 0, 1, each attached to the surface.
func (s *Session) PaintSurfaceSmoothing(self SmoothTile, nbs [4]SmoothTile, land func(id uint8) (SmoothLand, bool), snowImage int, sprite func(int) *gfx8.Sprite) {
	sc := corners(self.Surface)
	for _, edge := range [4]int{2, 3, 0, 1} {
		s.smoothenEdge(edge, self, sc, nbs[edge], land, snowImage, sprite)
	}
}

// smoothenEdge is paintSurfaceSmoothenEdge.
func (s *Session) smoothenEdge(edge int, self SmoothTile, selfCorners [4]int, nb SmoothTile, land func(uint8) (SmoothLand, bool), snowImage int, sprite func(int) *gfx8.Sprite) {
	if !nb.Present {
		return
	}
	eh := edgeHeightFor(selfCorners, nb.Surface, edge)
	if eh.neighbour1 != eh.self1 || eh.neighbour0 != eh.self0 {
		return
	}
	if nb.Land == 0xFF || self.Land == 0xFF { // either is industrial
		return
	}
	if self.Snow == 3 {
		return
	}
	var dh, cl uint8
	switch edge {
	case 0:
		dh, cl = Surfacek4FDA5E[self.DisplaySlope], Surfacek4FDA84[nb.DisplaySlope]
	case 1:
		dh, cl = Surfacek4FDA97[self.DisplaySlope], Surfacek4FDA71[nb.DisplaySlope]
	case 2:
		dh, cl = Surfacek4FDA71[self.DisplaySlope], Surfacek4FDA97[nb.DisplaySlope]
	case 3:
		dh, cl = Surfacek4FDA84[self.DisplaySlope], Surfacek4FDA5E[nb.DisplaySlope]
	}
	selfObj, nbObj := self.Land, nb.Land
	if self.Snow >= 4 {
		selfObj = 0xFF
	}
	if nb.Snow >= 4 {
		nbObj = 0xFF
	}
	selfLand, ok1 := land(self.Land)
	nbLand, ok2 := land(nb.Land)
	if !ok1 || !ok2 {
		return
	}
	if self.Growth == nb.Growth && selfObj == nbObj { // same tint
		if cl == dh || selfLand.SharpSlopeTransition {
			return
		}
	} else if selfLand.DisableSmoothTransition || nbLand.DisableSmoothTransition {
		return
	}
	base := sprite(int(SurfaceSmoothFromSlope[edge][self.DisplaySlope]))
	var mask *gfx8.Sprite
	if nbObj == 0xFF { // the neighbour is snow
		mask = sprite(snowImage + 38 + int(cl))
	} else {
		mask = sprite(nbLand.Image + nbLand.ImagesPerGrowthStage*int(nb.Growth) + 19 + int(cl))
	}
	if base == nil || mask == nil {
		return
	}
	s.AttachToPrevious(Image{Sprite: base, Mask: mask}, 0, 0)
}
