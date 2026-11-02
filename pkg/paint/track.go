package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Tracks — a port of Paint/PaintTrack.cpp paintTrack / paintTrackPP with the
// table generated from upstream (trackdata_gen.go). PARTIAL: company colour
// remap only via TrackEl.Remap (the owner's colour, from the scenario), bridges and tunnels (BridgeEntry,
// insertTunnels), track extras (mods), height markers and ghosts.

// TrackEl is what paintTrack reads from a track element.
type TrackEl struct {
	TrackID, SequenceIndex, Rotation uint8
	BaseZ                            uint8   // SmallZ
	Remap                            []uint8 // company colour of elTrack.owner() (nil: none)
	Tunnel                           int     // trackObj->tunnel's image base (-1: none)
	HasBridge                        bool
	Bridge                           uint8 // elTrack.bridge(): bridge object slot
}

// PaintTrack paints one track element at the current sprite position; image
// is the track object's image base.
func (s *Session) PaintTrack(el TrackEl, image int, sprite func(int) *gfx8.Sprite) {
	tid, seq := int(el.TrackID), int(el.SequenceIndex)
	if tid >= len(TrackParts) || seq >= len(TrackParts[tid]) {
		return
	}
	tpp := TrackParts[tid][seq]
	rot := int((s.Rotation + el.Rotation) & 3)
	height := int(el.BaseZ) * 4
	ho := Pos3{0, 0, height}
	bbOff := tpp.BBOffsets[rot]
	bbOff.Z += height
	img := func(off uint32) Image {
		return Colour(sprite(image+int(off)), el.Remap)
	}
	if el.HasBridge { // both paintTrackPPMergeable and paintTrackPPStandard
		s.setBridge(height, tpp.BridgeType[rot], tpp.BridgeEdges[rot], tpp.BridgeQuarters[rot], el.Bridge, el.Remap)
	}
	if tpp.IsMergeable {
		// paintTrackPPMergeable: ballast, sleepers, rails into priority lists 0, 1, 3.
		for layer, prio := range [3]int{0, 1, 3} {
			if i := img(tpp.ImageIndexOffsets[rot][layer]); i.Sprite != nil {
				s.AddTrackRoad(i, prio, ho, bbOff, tpp.BBSizes[rot])
			}
		}
		s.InsertTunnels(tpp.TunnelHeights[rot], height, el.Tunnel)
		s.AddPendingSegments(tpp.Segments[rot])
		return
	}
	// paintTrackPPStandard: one image via addToPlotList4FD150 (= AsParent).
	if i := img(tpp.ImageIndexOffsets[rot][0]); i.Sprite != nil {
		s.AddAsParent(i, ho, bbOff, tpp.BBSizes[rot])
	}
	s.InsertTunnels(tpp.TunnelHeights[rot], height, el.Tunnel)
	s.AddPendingSegments(tpp.Segments[rot])
}

// TrackMod is one of the track object's mods (trackObj->mods[i]) as
// paintTrack reads it: the track extra's image and paint style; Has is
// elTrack.hasMod(i).
type TrackMod struct {
	Has        bool
	Image      int
	PaintStyle uint8
}

// PaintTrackMods is paintTrack's track extras (zoom 0 only): for each mod the
// element has, the Style0/Style1 addition piece for its track id and sequence
// index (paintTrackAdditionPP) in the track's colour — mergeable pieces into
// track/road priority 4, the others as a child of the last struct. PARTIAL:
// Style1's supports (setAdditionSupport / paintSupports) and ghost mods.
func (s *Session) PaintTrackMods(el TrackEl, mods [4]TrackMod, sprite func(int) *gfx8.Sprite) {
	if s.Zoom > 0 {
		return
	}
	rot := int((s.Rotation + el.Rotation) & 3)
	h := Pos3{0, 0, int(el.BaseZ) * 4}
	tid, seq := int(el.TrackID), int(el.SequenceIndex)
	for _, m := range mods {
		if !m.Has {
			continue
		}
		var parts [][]TrackAdditionPiece
		switch m.PaintStyle {
		case 0:
			parts = TrackAdditionParts0[:]
		case 1:
			parts = TrackAdditionParts1[:]
		default:
			continue
		}
		if tid >= len(parts) || seq >= len(parts[tid]) {
			continue
		}
		p := parts[tid][seq]
		sp := sprite(m.Image + int(p.ImageIDs[rot]))
		if sp == nil {
			continue
		}
		img := Colour(sp, el.Remap)
		bbOff := Pos3{p.BBOffsets[rot].X, p.BBOffsets[rot].Y, p.BBOffsets[rot].Z + h.Z}
		if p.IsMergeable {
			s.AddTrackRoad(img, 4, h, bbOff, p.BBSizes[rot])
		} else {
			s.AddAsChild(img, h, bbOff, p.BBSizes[rot])
		}
		if sup := p.Supports; sup != nil && m.PaintStyle == 1 { // Style1::paintSupport
			s.setAdditionSupport(h.Z+int(sup.Height), sup.Segments[rot], sup.ImageIDs[rot][:], false, sup.Frequencies[rot], m.Image, el.Remap)
		}
	}
}
