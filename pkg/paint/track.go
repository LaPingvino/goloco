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
	if tpp.IsMergeable {
		// paintTrackPPMergeable: ballast, sleepers, rails into priority lists 0, 1, 3.
		for layer, prio := range [3]int{0, 1, 3} {
			if i := img(tpp.ImageIndexOffsets[rot][layer]); i.Sprite != nil {
				s.AddTrackRoad(i, prio, ho, bbOff, tpp.BBSizes[rot])
			}
		}
		return
	}
	// paintTrackPPStandard: one image via addToPlotList4FD150 (= AsParent).
	if i := img(tpp.ImageIndexOffsets[rot][0]); i.Sprite != nil {
		s.AddAsParent(i, ho, bbOff, tpp.BBSizes[rot])
	}
}
