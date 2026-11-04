package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Bridges — a port of Paint/PaintBridge.cpp paintBridge, with its tables
// generated from the .cpp (bridgedata_gen.go) and its image ids from the
// compiler (bridgeids_gen.go).
//
// All of paintBridge is ported: deck, walls and edges, the quarter cases
// (including 8, the front supports), the NE/SW spans with their pillars, the
// subType-0 span walls and the shadow. Rotation is 0 here (the scene renderer
// has no rotation yet).

// BridgeObj is what paintBridge reads from a bridge object.
type BridgeObj struct {
	Image         int
	SpanLength    uint8
	PillarSpacing uint8
	HasRoof       bool
	DeckDepth     int
	TileX, TileY  int     // the tile being painted (session unkPosition, rotation 0)
	ShadowPal     []uint8 // palette map of ExtColour::unk32 (the bridge shadow)
}

// PaintBridge is paintBridge: draws the session's bridge entry b. It reports
// whether anything was painted (upstream's return value).
func (s *Session) PaintBridge(b BridgeEntry, obj BridgeObj, sprite func(int) *gfx8.Sprite) bool {
	s.ItemType = ItemBridge // setItemType (PaintBridge.cpp:1191; the inner noInteraction resets not yet ported)
	if b.Height == -1 {
		return false
	}
	genHeight := (s.support.Height + 15) &^ 15 // Numerics::ceil2(height, 16)
	supportLength := b.Height - genHeight
	if supportLength < 0 {
		return false
	}
	base := Pos3{0, 0, b.Height}
	at := func(p Pos3) Pos3 { return Pos3{p.X + base.X, p.Y + base.Y, p.Z + base.Z} }
	add := func(img uint32, off, bbOff, bbLen Pos3) {
		if sp := sprite(obj.Image + int(img)); sp != nil {
			s.AddAsParent(Colour(sp, b.Remap), off, bbOff, bbLen) // addToPlotList4FD150
		}
	}
	flags := Bridge4F91DC[b.SubType]
	if flags&(1<<0) != 0 {
		deck := BridgeDeckImagesBySubType[b.SubType-1]
		add(deck.Deck, base, at(Pos3{2, 2, 0}), Pos3{28, 28, 1})
		if flags&(1<<2) != 0 {
			if b.EdgesQuarters&(1<<7) == 0 {
				add(deck.WallLhs, base, at(Pos3{2, 0, 8}), Pos3{28, 1, 30})
			}
			if b.EdgesQuarters&(1<<5) == 0 {
				add(deck.WallRhs, base, at(Pos3{1, 30, 8}), Pos3{29, 1, 30})
			}
		}
		if flags&(1<<3) != 0 {
			if b.EdgesQuarters&(1<<4) == 0 {
				add(deck.WallLhs, base, at(Pos3{0, 2, 8}), Pos3{1, 28, 30})
			}
			if b.EdgesQuarters&(1<<6) == 0 {
				add(deck.WallRhs, base, at(Pos3{30, 1, 8}), Pos3{1, 29, 30})
			}
		}
		off2 := Pos3{base.X, base.Y, base.Z + 8}
		at2 := func(p Pos3) Pos3 { return Pos3{p.X + off2.X, p.Y + off2.Y, p.Z + off2.Z} }
		if flags&(1<<4) != 0 {
			add(BridgeImgDeckWallEdge0, off2, at2(Pos3{22, 24, 0}), Pos3{2, 2, 26})
		}
		if flags&(1<<5) != 0 {
			add(BridgeImgDeckWallEdge1, off2, at2(Pos3{7, 7, 0}), Pos3{2, 2, 26})
		}
		if flags&(1<<6) != 0 {
			add(BridgeImgDeckWallEdge2, off2, at2(Pos3{24, 22, 0}), Pos3{2, 2, 26})
		}
		if flags&(1<<7) != 0 {
			add(BridgeImgDeckWallEdge3, off2, at2(Pos3{17, 17, 2}), Pos3{2, 2, 24})
		}
	}
	if s.support.Slope&(1<<5) == 0 && supportLength == 0 {
		return true // no shadow for this
	}
	slope := s.support.Slope & 0x1F
	shadowType := 0
	switch b.EdgesQuarters & 0xF {
	case 1:
		shadowType = 4
		s.bridgeQuarterNoSupport(b, obj, add, [3]uint32{BridgeImgRoofEdge0, BridgeImgDeckBaseNoSupportEdge0, BridgeImgDeckWallEdge0}, Pos3{22, 24, 0})
	case 2:
		shadowType = 2
		s.bridgeQuarterNoSupport(b, obj, add, [3]uint32{BridgeImgRoofEdge1, BridgeImgDeckBaseNoSupportEdge1, BridgeImgDeckWallEdge1}, Pos3{7, 7, 0})
	case 4:
		shadowType = 3
		s.bridgeQuarterNoSupport(b, obj, add, [3]uint32{BridgeImgRoofEdge2, BridgeImgDeckBaseNoSupportEdge2, BridgeImgDeckWallEdge2}, Pos3{24, 22, 0})
	case 8:
		shadowType = 1
		s.bridgeQuarterSupportFront(b, obj, add, supportLength, slope)
	default:
		// unkTile = toTileSpace(rotate(unkPosition, 0)) + kTwosToOnesCompliment[0], masked by the span.
		ux := (obj.TileX + BridgeTwosToOnesCompliment[0][0]) & (int(obj.SpanLength) - 1)
		uy := (obj.TileY + BridgeTwosToOnesCompliment[0][1]) & (int(obj.SpanLength) - 1)
		xPillar := uint8(int(obj.PillarSpacing) / (1 << (ux * 2)))
		yPillar := uint8(int(obj.PillarSpacing) / (1 << (uy * 2)))
		xSpan, ySpan := BridgeSpanImageSets[0][ux], BridgeSpanImageSets[1][uy]
		// bridgeRotation, rotation 0/2.
		if b.EdgesQuarters&(1<<5) == 0 || b.EdgesQuarters&(1<<7) == 0 {
			s.bridgeSpan(b, obj, add, supportLength, slope, xPillar, xSpan, &bridgeDirNE)
		} else {
			s.bridgeSpan(b, obj, add, supportLength, slope, yPillar, ySpan, &bridgeDirSW)
		}
		if b.SubType == 0 {
			if b.EdgesQuarters&(1<<7) == 0 {
				add(xSpan.WallLhs, base, at(Pos3{2, 0, 0}), Pos3{28, 1, 26})
			}
			if b.EdgesQuarters&(1<<4) == 0 {
				add(ySpan.WallLhs, base, at(Pos3{0, 2, 0}), Pos3{1, 28, 26})
			}
			if b.EdgesQuarters&(1<<5) == 0 {
				add(xSpan.WallRhs, base, at(Pos3{0, 30, 2}), Pos3{28, 1, 24})
			}
			if b.EdgesQuarters&(1<<6) == 0 {
				add(ySpan.WallRhs, base, at(Pos3{30, 2, 2}), Pos3{1, 28, 24})
			}
		}
	}
	// Shadow (zoom <= 1): on the water, else on the surface by its display slope.
	if obj.ShadowPal != nil && s.Zoom <= 1 {
		displaySlope := 0
		height := s.waterHeight2
		if height == 0 {
			height = s.surfaceHeight
			displaySlope = int(BridgeSlopeToDisplaySlope[s.surfaceSlope&0x1F])
		}
		if sp := sprite(int(BridgeSlopeToShadow[shadowType][displaySlope])); sp != nil {
			ho := Pos3{0, 0, height}
			s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, true), PalMap: obj.ShadowPal}, ho, Pos3{15, 15, 1 + height}, Pos3{2, 2, 1})
		}
	}
	return true
}

// bridgeQuarterNoSupport is paintFlatSingleQuarterNoSupport.
func (s *Session) bridgeQuarterNoSupport(b BridgeEntry, obj BridgeObj, add func(uint32, Pos3, Pos3, Pos3), imgs [3]uint32, wallBBOff Pos3) {
	base := Pos3{0, 0, b.Height}
	if obj.HasRoof {
		add(imgs[0], base, Pos3{0, 0, 30 + b.Height}, Pos3{32, 32, 0})
	}
	off := Pos3{0, 0, b.Height - 16}
	add(imgs[1], off, Pos3{0, 0, 14 + off.Z}, Pos3{32, 32, 1})
	if b.SubType == 0 {
		add(imgs[2], base, Pos3{wallBBOff.X, wallBBOff.Y, wallBBOff.Z + b.Height}, Pos3{2, 2, 26})
	}
}

// bridgeDir holds what differs between paintBridgeNE and paintBridgeSW (and
// paintSupportNE / paintSupportSW): the segments a pillar side needs free, the
// slope offset tables, the support edge images and the box orientation.
type bridgeDir struct {
	lhsSegs, rhsSegs   [3]int // pillarSpacing bit 0 (lhs) / bit 1 (rhs)
	lhsSlope, rhsSlope *[32]int16
	edgeLhs, edgeRhs   *[32]uint32
	seg16, seg32       [2]uint32 // [lhs, rhs] support segment images
	thin, rhsOff       Pos3      // a 2-wide box along the span; offset of the rhs side
}

var (
	bridgeDirNE = bridgeDir{
		lhsSegs: [3]int{0, 2, 6}, rhsSegs: [3]int{1, 3, 7},
		lhsSlope: &Bridge4F8F9C, rhsSlope: &Bridge4F8F5C,
		edgeLhs: &BridgeSlopeToSupportEdgeImageLhsNE, edgeRhs: &BridgeSlopeToSupportEdgeImageRhsNE,
		seg16: [2]uint32{BridgeImgSupportSegmentLhs16NE, BridgeImgSupportSegmentRhs16NE},
		seg32: [2]uint32{BridgeImgSupportSegmentLhs32NE, BridgeImgSupportSegmentRhs32NE},
		thin:  Pos3{2, 32, 0}, rhsOff: Pos3{30, 0, 0},
	}
	bridgeDirSW = bridgeDir{
		lhsSegs: [3]int{0, 1, 5}, rhsSegs: [3]int{2, 3, 8},
		lhsSlope: &Bridge4F8FDC, rhsSlope: &Bridge4F901C,
		edgeLhs: &BridgeSlopeToSupportEdgeImageLhsSW, edgeRhs: &BridgeSlopeToSupportEdgeImageRhsSW,
		seg16: [2]uint32{BridgeImgSupportSegmentLhs16SW, BridgeImgSupportSegmentRhs16SW},
		seg32: [2]uint32{BridgeImgSupportSegmentLhs32SW, BridgeImgSupportSegmentRhs32SW},
		thin:  Pos3{32, 2, 0}, rhsOff: Pos3{0, 30, 0},
	}
)

// bridgeSpan is paintBridgeNE / paintBridgeSW (by d), with paintSupportNE/SW.
func (s *Session) bridgeSpan(b BridgeEntry, obj BridgeObj, add func(uint32, Pos3, Pos3, Pos3), supportLength int, slope, pillarSpacing uint8, span BridgeSpanImages, d *bridgeDir) {
	base := Pos3{0, 0, b.Height}
	full := b.EdgesQuarters&0xF0 == 0xF0
	if obj.HasRoof {
		roof := span.Roof
		if full {
			roof = BridgeImgRoofFullTile
		}
		add(roof, base, Pos3{0, 0, 30 + b.Height}, Pos3{32, 32, 0})
	}
	lhs, rhs, ok := func() (int, int, bool) {
		if b.EdgesQuarters&(1<<6|1<<5) == 1<<6|1<<5 {
			return 0, 0, false
		}
		blocked := func(segs [3]int) bool {
			for _, i := range segs {
				if s.segments[i].Height == 0xFFFF {
					return true
				}
			}
			return false
		}
		if pillarSpacing&(1<<1) != 0 && blocked(d.rhsSegs) {
			return 0, 0, false
		}
		if pillarSpacing&(1<<0) != 0 && blocked(d.lhsSegs) {
			return 0, 0, false
		}
		unk := supportLength - obj.DeckDepth
		if unk < 0 {
			return 0, 0, false
		}
		if obj.DeckDepth == 32 {
			unk = b.Height - 16
			if unk == s.waterHeight {
				return 0, 0, false
			}
		}
		r := unk - int(d.rhsSlope[slope])
		if r < 0 {
			return 0, 0, false
		}
		l := unk - int(d.lhsSlope[slope])
		if l < 0 {
			return 0, 0, false
		}
		return l, r, true
	}()
	if !ok {
		deck := span.Deck
		if full {
			deck = BridgeImgDeckBaseNoSupport
		}
		off := Pos3{0, 0, b.Height - 16}
		add(deck, off, Pos3{0, 0, 14 + off.Z}, Pos3{32, 32, 1})
		return
	}
	off := Pos3{0, 0, b.Height - obj.DeckDepth}
	add(span.Deck, off, Pos3{0, 0, obj.DeckDepth - 2 + off.Z}, Pos3{32, 32, 1})
	hdr := Pos3{d.thin.X, d.thin.Y, obj.DeckDepth - 3}
	add(span.SupportHeaderLhs, off, Pos3{0, 0, off.Z}, hdr)
	add(span.SupportHeaderRhs, off, Pos3{d.rhsOff.X, d.rhsOff.Y, off.Z}, hdr)
	// paintSupportNE / paintSupportSW
	supportHeight := b.Height - obj.DeckDepth
	side := func(bit uint8, length int, which int, bbOff Pos3, edge uint32) {
		if pillarSpacing&bit == 0 {
			return
		}
		h := supportHeight
		for length >= 16 {
			is16 := length == 16 || h-16 == s.waterHeight
			sec, img, bz := 32, d.seg32[which], 31
			if is16 {
				sec, img, bz = 16, d.seg16[which], 15
			}
			length -= sec
			h -= sec
			add(img, Pos3{0, 0, h}, Pos3{bbOff.X, bbOff.Y, h}, Pos3{d.thin.X, d.thin.Y, bz})
		}
		if edge != 0 {
			h -= 16
			add(edge, Pos3{0, 0, h}, Pos3{bbOff.X, bbOff.Y, h}, Pos3{d.thin.X, d.thin.Y, 14})
		}
	}
	side(1<<0, lhs, 0, Pos3{}, d.edgeLhs[slope])
	side(1<<1, rhs, 1, d.rhsOff, d.edgeRhs[slope])
}

// bridgeQuarterSupportFront is paintFlatSingleQuarterSupportFront (quarter 8:
// "needs to do the front supports to ground as well").
func (s *Session) bridgeQuarterSupportFront(b BridgeEntry, obj BridgeObj, add func(uint32, Pos3, Pos3, Pos3), supportLength int, slope uint8) {
	base := Pos3{0, 0, b.Height}
	if obj.HasRoof {
		add(BridgeImgRoofEdge3, base, Pos3{0, 0, 30 + b.Height}, Pos3{32, 32, 0})
	}
	lhs, rhs, ok := func() (int, int, bool) {
		if s.segments[1].Height == 0xFFFF || s.segments[2].Height == 0xFFFF {
			return 0, 0, false
		}
		unk := supportLength - obj.DeckDepth
		if unk < 0 {
			return 0, 0, false
		}
		if obj.DeckDepth == 32 {
			unk = b.Height - 16
			if unk == s.waterHeight {
				return 0, 0, false
			}
		}
		r := unk - int(Bridge4F915C[slope])
		if r < 0 {
			return 0, 0, false
		}
		l := unk - int(Bridge4F919C[slope])
		if unk < 0 { // sic: upstream tests unkHeight here, not the lhs length
			return 0, 0, false
		}
		return l, r, true
	}()
	if ok {
		off := Pos3{0, 0, b.Height - obj.DeckDepth}
		add(BridgeImgDeckBaseWithSupportHeaderEdge3, off, Pos3{0, 0, 14 + off.Z}, Pos3{32, 32, 1})
		s.bridgeSupportDiagonal(add, lhs, rhs, b.Height-obj.DeckDepth)
	} else {
		off := Pos3{0, 0, b.Height - 16}
		add(BridgeImgDeckBaseNoSupportEdge3, off, Pos3{0, 0, 14 + off.Z}, Pos3{32, 32, 1})
	}
	if b.SubType == 0 {
		add(BridgeImgDeckWallEdge3, base, Pos3{17, 17, 2 + b.Height}, Pos3{2, 2, 24})
	}
}

// bridgeSupportDiagonal is paintSupportDiagonal. As upstream, the rhs loop
// draws the image named Lhs (and vice versa).
func (s *Session) bridgeSupportDiagonal(add func(uint32, Pos3, Pos3, Pos3), lhsLen, rhsLen, supportHeight int) {
	loop := func(length int, img16, img32 uint32, bx, by int) {
		h := supportHeight
		for length >= 16 {
			is16 := length == 16 || h-16 == s.waterHeight
			sec, img, bz := 32, img32, 31
			if is16 {
				sec, img, bz = 16, img16, 15
			}
			h -= sec
			length -= sec
			add(img, Pos3{0, 0, h}, Pos3{bx, by, h}, Pos3{2, 2, bz})
		}
	}
	loop(rhsLen, BridgeImgSupportSegmentEdge3Lhs16SW, BridgeImgSupportSegmentEdge3Lhs32SW, 30, 0)
	loop(lhsLen, BridgeImgSupportSegmentEdge3Rhs16SW, BridgeImgSupportSegmentEdge3Rhs32SW, 0, 30)
}
