package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The generated addition tables cover every track piece's sequence, like the
// track pieces themselves (TrackParts).
func TestTrackAdditionTablesShape(t *testing.T) {
	if len(TrackAdditionParts0) != len(TrackParts) || len(TrackAdditionParts1) != len(TrackParts) {
		t.Fatalf("track ids: %d / %d / %d", len(TrackAdditionParts0), len(TrackAdditionParts1), len(TrackParts))
	}
	for id := range TrackParts {
		if len(TrackAdditionParts0[id]) != len(TrackParts[id]) || len(TrackAdditionParts1[id]) != len(TrackParts[id]) {
			t.Errorf("trackId %d: %d / %d additions for %d pieces", id, len(TrackAdditionParts0[id]), len(TrackAdditionParts1[id]), len(TrackParts[id]))
		}
	}
}

// A mod on straight track (id 0, sequence 0) adds Style0's piece image for
// the rotation: into track/road priority 4 when mergeable, else as a child.
func TestPaintTrackModsStraight(t *testing.T) {
	p := TrackAdditionParts0[0][0]
	var asked []int
	sprite := func(id int) *gfx8.Sprite {
		asked = append(asked, id)
		return &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{1}}
	}
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	s.PaintTrackMods(TrackEl{TrackID: 0, SequenceIndex: 0, Rotation: 1, BaseZ: 8}, [4]TrackMod{{}, {Has: true, Image: 1000, PaintStyle: 0}}, sprite)
	if len(asked) != 1 || asked[0] != 1000+int(p.ImageIDs[1]) {
		t.Fatalf("images %v, want [%d]", asked, 1000+int(p.ImageIDs[1]))
	}
	if p.IsMergeable && s.trackRoad[4] == nil {
		t.Error("a mergeable piece should be in track/road priority 4")
	}
	asked = nil
	s.Zoom = 1
	s.PaintTrackMods(TrackEl{}, [4]TrackMod{{Has: true, Image: 1000}}, sprite)
	if len(asked) != 0 {
		t.Error("track mods are drawn at zoom 0 only")
	}
}
