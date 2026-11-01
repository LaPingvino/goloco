package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Industries — a port of Paint/PaintIndustry.cpp paintIndustry /
// paintIndustryBuilding. PARTIAL: shadows (translucency Colours::getShadow),
// scaffolding while under construction, ghosts and industry fields (the
// surface painter's industrial branch) are not ported yet.

// IndustryEl is what paintIndustry reads from an industry element.
type IndustryEl struct {
	Rotation, SequenceIndex, BuildingType uint8
	Var6Low, SectionProgress              uint8
	IsConstructed                         bool
	BaseZ, ClearZ                         uint8 // SmallZ
}

// IndustryObj is what paintIndustry reads from an industry object.
type IndustryObj struct {
	BuildingImageIDs   int
	Parts              []uint8 // the element's building type
	PartHeights        []uint8
	PartAnimFrames     []uint8
	PartAnimSpeed      []uint8
	AnimationSequences [4][]uint8
	LargeTile          bool // buildingSizeFlags & (1 << buildingType)
}

// PaintIndustry paints an industry element at the current sprite position.
func (s *Session) PaintIndustry(el IndustryEl, obj IndustryObj, colourMap []uint8, tileX, tileY int, ticks uint32, sprite func(int) *gfx8.Sprite) {
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
	if obj.LargeTile && int(el.SequenceIndex^2) != int(-int(s.Rotation))&3 {
		return
	}
	// paintIndustryBuilding
	numSections := 0xF0 // all sections completed
	sectionProgress := 0
	if !el.IsConstructed {
		ticks = 0
		numSections = int(el.Var6Low)
		sectionProgress = int(el.SectionProgress)
	}
	var seq []uint8
	if el.Var6Low&(1<<5) != 0 && el.Var6Low&(1<<4) != 0 {
		seq = obj.AnimationSequences[el.Var6Low&3]
	}
	// int8_t sectionCount = numSections, as upstream: 0xF0 becomes -16 and the
	// -1 stop is never reached for a completed building.
	sectionCount := int(int8(uint8(numSections)))
	off := imgOff
	for _, part := range obj.Parts {
		if sectionCount == -1 {
			break
		}
		p := int(part)
		if p < len(obj.PartAnimFrames) {
			if obj.PartAnimFrames[p] != 0 {
				frameMask := int(obj.PartAnimFrames[p]) - 1
				speed := obj.PartAnimSpeed[p]
				tick := int(ticks >> (speed & 0x7F))
				if speed&0x80 != 0 {
					tick += tileX*5 + tileY*3
				}
				p += frameMask & tick
			} else if len(seq) > 0 {
				p += int(seq[int(ticks>>obj.PartAnimSpeed[p])&(len(seq)-1)])
			}
		}
		img := Image{Sprite: sprite(p*4 + obj.BuildingImageIDs + int(rotation)), Op: gfx8.OpTransparent}
		if colourMap != nil {
			img.Op |= gfx8.OpSrc
			img.PalMap = colourMap
		}
		if sectionCount == 0 {
			if n := (sectionProgress + 1) & 7; n != 0 {
				img.Noise = sprite(3478 + n)
			}
		}
		if img.Sprite != nil {
			s.AddAsChild(img, off, bbOff, bbSz)
		}
		if p < len(obj.PartHeights) {
			off.Z += int(obj.PartHeights[p])
		}
		sectionCount--
	}
}

// AirportObj is what paintAirport reads from an airport object.
type AirportObj struct {
	BuildingImage  int
	Parts          []uint8 // the element's building type (variation)
	PartHeights    []uint8
	PartAnimFrames []uint8
	PartAnimSpeed  []uint8
	LargeTile      bool    // largeTiles & (1 << variation)
	Remap          []uint8 // company colour of elStation.owner()
}

// PaintAirport — a port of Paint/PaintAirport.cpp paintAirport and
// paintAirportBuilding. PARTIAL: shadows, company colour, ghosts.
func (s *Session) PaintAirport(rotationEl, sequenceIndex, baseZ uint8, obj AirportObj, tileX, tileY int, ticks uint32, sprite func(int) *gfx8.Sprite) {
	rotation := (s.Rotation + rotationEl) & 3
	// "Odd... different to industry, building, dock": the clear height is the
	// sum of the variation's part heights, rounded up to 4.
	clearHeight := 0
	for _, p := range obj.Parts {
		if int(p) < len(obj.PartHeights) {
			clearHeight += int(obj.PartHeights[p])
		}
	}
	clearHeight = (clearHeight + 3) &^ 3
	bbz := min(clearHeight, 128) - 2
	baseHeight := int(baseZ) * 4
	imgOff, bbOff, bbSz := bldImgOff1x1, bldBBOff1x1, bldBBSz1x1
	if obj.LargeTile {
		imgOff, bbOff, bbSz = bldImgOff2x2, bldBBOff2x2, bldBBSz2x2
	}
	imgOff.Z += baseHeight
	bbOff.Z += baseHeight
	bbSz.Z += bbz
	s.last = nil // session.resetLastPS()
	if obj.LargeTile && int(sequenceIndex^2) != int(-int(s.Rotation))&3 {
		return // only the front tile of a 2×2 draws
	}
	s.paintStationBuilding(obj, imgOff, bbOff, bbSz, rotation, tileX, tileY, ticks, sprite)
}

// paintStationBuilding is paintAirportBuilding / paintDocksBuilding (the same
// code upstream): parts stacked by height, each advanced by its animation.
func (s *Session) paintStationBuilding(obj AirportObj, off, bbOff, bbSz Pos3, rotation uint8, tileX, tileY int, ticks uint32, sprite func(int) *gfx8.Sprite) {
	for _, part := range obj.Parts {
		// Upstream applies numFrames-1 unchecked (no numFrames==0 test as in
		// buildings): uint8 - 1 → int -1 → all bits.
		frameMask := uint32(int32(obj.PartAnimFrames[part]) - 1)
		speed := obj.PartAnimSpeed[part]
		tick := ticks >> (speed & 0x7F)
		if speed&0x80 != 0 {
			tick += uint32(tileX*5 + tileY*3)
		}
		adj := int(uint32(part) + (frameMask & tick))
		if adj >= len(obj.PartHeights) {
			break // out of the object's part table: upstream would read past it
		}
		if sp := sprite(adj*4 + obj.BuildingImage + int(rotation)); sp != nil {
			s.AddAsChild(Colour(sp, obj.Remap), off, bbOff, bbSz)
		}
		off.Z += int(obj.PartHeights[adj])
	}
}

// PaintDocks — a port of Paint/PaintDocks.cpp paintDocks (always 2×2; bbox
// height from the element's clear height). PARTIAL: shadows, company colour, ghosts.
func (s *Session) PaintDocks(rotationEl, sequenceIndex, baseZ, clearZ uint8, obj AirportObj, tileX, tileY int, ticks uint32, sprite func(int) *gfx8.Sprite) {
	rotation := (s.Rotation + rotationEl) & 3
	baseHeight := int(baseZ) * 4
	bbz := min(int(clearZ)*4-baseHeight, 128) - 2
	imgOff, bbOff, bbSz := bldImgOff2x2, bldBBOff2x2, bldBBSz2x2
	imgOff.Z += baseHeight
	bbOff.Z += baseHeight
	bbSz.Z += bbz
	s.last = nil // session.resetLastPS()
	if int(sequenceIndex^2) == int(-int(s.Rotation))&3 {
		s.paintStationBuilding(obj, imgOff, bbOff, bbSz, rotation, tileX, tileY, ticks, sprite)
	}
}
