package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Trees — a port of Paint/PaintTree.cpp paintTree (0x004BAEDA). PARTIAL: ghost
// trees and coloured trees (no vanilla object uses colours) are not ported.
// The second (season-transition) image carries a noise mask: noise image
// 3478+n (ImageIds::noise_mask_1 = 3479), applied by gfx8 to non-RLE sprites.

// TreeEl is what paintTree reads from a tree element.
type TreeEl struct {
	Rotation, Quadrant, Growth, Season, Unk7l uint8
	HasSnow                                   bool
	BaseZ, ClearZ                             uint8
}

// TreeObj is what paintTree reads from a tree object.
type TreeObj struct {
	NumRotations      uint8
	Sprites           [6]uint32 // absolute image bases per season
	SnowSprites       [6]uint32
	ShadowImageOffset uint32
	HasShadow         bool
}

var (
	treeSeasonSwap   = [6]uint8{3, 0, 1, 2, 1, 4}                    // _50076A
	treeSeasonFirst  = [6]bool{true, true, false, false, true, true} // _500770
	treeQuadrantOffs = [4][2]int{{7, 7}, {7, 23}, {23, 23}, {23, 7}}
)

// PaintTree paints one tree at the current sprite position. shadowMap is the
// unk32 translucency map (ImageIds::paletteMapUnk32 = 2208).
func (s *Session) PaintTree(el TreeEl, obj TreeObj, sprite func(int) *gfx8.Sprite, shadowMap []uint8) {
	if obj.NumRotations == 0 {
		return
	}
	viewRot := (s.Rotation + el.Rotation) & 3
	frame := uint32(viewRot%obj.NumRotations) + uint32(el.Growth)*uint32(obj.NumRotations)
	seasons := func(snow bool) [6]uint32 {
		if snow {
			return obj.SnowSprites
		}
		return obj.Sprites
	}
	season := el.Season % 6
	hasImage2 := false
	var image2 uint32
	noiseMask := 0
	if el.Unk7l != 7 {
		hasImage2 = true
		season = treeSeasonSwap[season]
		noiseMask = int(el.Unk7l) + 1
		img2Season := el.Season % 6
		if treeSeasonFirst[img2Season] {
			img2Season = season
			season = el.Season % 6
			noiseMask = 8 - noiseMask
		}
		image2 = frame + seasons(el.HasSnow)[img2Season]
	}
	seasonBase := seasons(el.HasSnow)[season]
	q := (el.Quadrant + s.Rotation) % 4
	off := Pos3{treeQuadrantOffs[q][0], treeQuadrantOffs[q][1], int(el.BaseZ) * 4}
	bbz := min(int(el.ClearZ)-int(el.BaseZ), 32)*4 - 3

	if obj.HasShadow && shadowMap != nil {
		if sp := sprite(int(obj.ShadowImageOffset + frame + seasonBase)); sp != nil {
			s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpTransparent | gfx8.OpDst, PalMap: shadowMap}, off, off, Pos3{18, 18, 1})
		}
	}
	bbOff := Pos3{off.X, off.Y, off.Z + 2}
	if sp := sprite(int(frame + seasonBase)); sp != nil {
		s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, off, bbOff, Pos3{2, 2, bbz})
	}
	if hasImage2 {
		if sp := sprite(int(image2)); sp != nil {
			var noise *gfx8.Sprite
			if noiseMask > 0 {
				noise = sprite(3478 + noiseMask) // _noiseMaskImages[noiseMask]
			}
			s.AddAsChild(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false), Noise: noise}, off, bbOff, Pos3{2, 2, bbz})
		}
	}
}
