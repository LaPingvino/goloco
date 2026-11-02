package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Buildings — a port of Paint/PaintBuilding.cpp paintBuilding /
// paintBuildingBuilding, with the shadow (translucency
// Colours::getShadow(orange)) and the under-construction stages (built
// sections by age, the newest noise-masked). PARTIAL: scaffolding and ghost
// buildings are not ported.

// BuildingEl is what paintBuilding reads from a building element.
type BuildingEl struct {
	Rotation, SequenceIndex, Variation uint8
	BaseZ, ClearZ                      uint8 // SmallZ
	// Under construction (!IsConstructed): Age sections are built, the next
	// one at SectionProgress (unk5u).
	IsConstructed        bool
	Age, SectionProgress uint8
}

// BuildingObj is what paintBuilding reads from a building object.
type BuildingObj struct {
	Image             int
	NumParts          int
	Parts             []uint8 // the element's variation
	PartHeights       []uint8
	PartAnimFrames    []uint8
	PartAnimSpeed     []uint8
	ElevatorSequences [][]uint8
	LargeTile         bool // BuildingObjectFlags::largeTile (2×2)
	Shadow            Shadow
}

// Shadow is the object's shadow image (hasShadows, zoom <= 1): Image is its base
// before + rotation; Pal the palette map of its translucency (Colours::getShadow);
// Primary when the image also carries a primary colour (industries), which
// makes the blend src+dst as getDrawBlendOp. Pal nil: no shadow.
type Shadow struct {
	Image   int
	Pal     []uint8
	Primary bool
}

// paintShadow adds the shadow as a child, as upstream does right after
// resetLastPS (before the 2×2 front-tile test, so every tile draws it).
func (s *Session) paintShadow(sh Shadow, rotation uint8, imgOff, bbOff, bbSz Pos3, sprite func(int) *gfx8.Sprite) {
	if sh.Pal == nil {
		return
	}
	if sp := sprite(sh.Image + int(rotation)); sp != nil {
		s.AddAsChild(Image{Sprite: sp, Op: gfx8.OpFor(sp, sh.Primary, true), PalMap: sh.Pal}, imgOff, bbOff, bbSz)
	}
}

var (
	bldImgOff1x1, bldImgOff2x2 = Pos3{16, 16, 0}, Pos3{0, 0, 0}
	bldBBOff1x1, bldBBOff2x2   = Pos3{3, 3, 0}, Pos3{-8, -8, 0}
	bldBBSz1x1, bldBBSz2x2     = Pos3{26, 26, 0}, Pos3{38, 38, 0}
)

// PaintBuilding paints a building element at the current sprite position.
// colourMap is the remap for the element's colour (ImageId(0, colour)); tileX,
// tileY is its tile (for position-modified animations); ticks the scenario ticks.
func (s *Session) PaintBuilding(el BuildingEl, obj BuildingObj, colourMap []uint8, tileX, tileY int, ticks uint32, sprite func(int) *gfx8.Sprite) {
	rotation := (s.Rotation + el.Rotation) & 3
	baseHeight := int(el.BaseZ) * 4
	bbz := min(int(el.ClearZ)*4-baseHeight, 128) - 2
	imgOff, bbOff, bbSz := bldImgOff1x1, bldBBOff1x1, bldBBSz1x1
	if obj.LargeTile {
		imgOff, bbOff, bbSz = bldImgOff2x2, bldBBOff2x2, bldBBSz2x2
	}
	imgOff.Z += baseHeight
	bbOff.Z += baseHeight
	bbSz.Z += bbz
	s.last = nil // session.resetLastPS()
	s.paintShadow(obj.Shadow, rotation, imgOff, bbOff, bbSz, sprite)
	if obj.LargeTile {
		// After the 2×2 front-tile test, every tile of the footprint:
		// setGeneralSupportHeight(0xFFFF, 0).
		defer s.blockSupports()
	}
	if obj.LargeTile && int(el.SequenceIndex^2) != int(-int(s.Rotation))&3 {
		return // only the front tile of a 2×2 draws; images overlap the others
	}
	img := func(idx int) Image {
		sp := sprite(idx)
		op := gfx8.OpTransparent
		if colourMap != nil {
			op |= gfx8.OpSrc
		}
		return Image{Sprite: sp, Op: op, PalMap: colourMap}
	}
	// paintBuildingBuilding: under construction the clock stops and only the
	// built sections (age) draw, the newest with the noise mask of its progress;
	// constructed, numSections = 0xF0 (int8 −16: every part).
	numSections, sectionProgress := 0xF0, 0
	if !el.IsConstructed {
		ticks = 0
		numSections, sectionProgress = int(el.Age), int(el.SectionProgress)
	}
	totalSectionHeight := 0 // uint8: the built height (≥ 1 once counting stops)
	if numSections != 0xF0 {
		sc := int(int8(uint8(numSections)))
		for _, part := range obj.Parts {
			if int(part) < len(obj.PartHeights) {
				totalSectionHeight = (totalSectionHeight + int(obj.PartHeights[part])) & 0xFF
			}
			sc--
			if sc == -1 {
				totalSectionHeight = max(1, totalSectionHeight)
				break
			}
		}
	}
	// PARTIAL: the scaffolding (back before the parts, front and roof after)
	// is not drawn yet.
	sectionCount := int(int8(uint8(numSections)))
	off := imgOff
	for _, part := range obj.Parts {
		if sectionCount == -1 {
			break
		}
		p := int(part)
		if p < len(obj.PartAnimFrames) {
			frameMask := int(obj.PartAnimFrames[p]) - 1
			speed := obj.PartAnimSpeed[p]
			tick := int(ticks >> (speed & 0x7F))
			if speed&0x80 != 0 {
				tick += tileX*5 + tileY*3
			}
			p += frameMask & tick
		}
		if i := p*4 + obj.Image + int(rotation); sprite(i) != nil {
			im := img(i)
			if sectionCount == 0 { // withNoiseMask((sectionProgress + 1) & 7); 0 = none
				if n := (sectionProgress + 1) & 7; n != 0 {
					im.Noise = sprite(3478 + n)
				}
			}
			s.AddAsChild(im, off, bbOff, bbSz)
		}
		if p < len(obj.PartHeights) {
			off.Z += int(obj.PartHeights[p])
		}
		sectionCount--
	}
	// Elevators, only once nothing is left to build (totalSectionHeight == 0).
	if totalSectionHeight != 0 {
		return
	}
	for a, seq := range obj.ElevatorSequences {
		if len(seq) == 0 {
			continue
		}
		t := int(ticks/2) + tileX*8 + tileY*8
		h := int(seq[t&(len(seq)-1)])
		i := obj.Image + obj.NumParts*4 + a*4 + int(rotation)
		if sprite(i) != nil {
			s.AddAsChild(img(i), Pos3{imgOff.X, imgOff.Y, imgOff.Z + h}, bbOff, bbSz)
		}
	}
}
