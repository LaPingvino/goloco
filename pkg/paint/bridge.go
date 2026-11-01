package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Bridges — a port of Paint/PaintBridge.cpp paintBridge, with its tables
// generated from the .cpp (bridgedata_gen.go) and its image ids from the
// compiler (bridgeids_gen.go).
//
// PARTIAL — ported so far: the early returns, the deck (k4F91DC bit 0) with
// its side walls (bits 2, 3) and deck-wall edges (bits 4-7). NOT YET: the
// quarter cases and the NE/SW span + support code (paintFlatSingleQuarter*,
// paintBridgeNE/SW, paintSupport*), the subType-0 span walls, and the shadow.
// Until then a bridge whose quarters/span branch would draw shows only its deck.

// BridgeObj is what paintBridge reads from a bridge object.
type BridgeObj struct {
	Image         int
	SpanLength    uint8
	PillarSpacing uint8
	HasRoof       bool
}

// PaintBridge is paintBridge: draws the session's bridge entry b. It reports
// whether anything was painted (upstream's return value).
func (s *Session) PaintBridge(b BridgeEntry, obj BridgeObj, sprite func(int) *gfx8.Sprite) bool {
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
	// PARTIAL: quarter cases, spans, supports and the shadow (see the file comment).
	return true
}
