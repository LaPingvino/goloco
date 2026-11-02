package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Track and road addition supports (e.g. catenary poles) — a port of
// PaintTile.cpp paintSupports and the session's TrackRoadAdditionSupports,
// filled by the Style1 addition pieces (setAdditionSupport).

// additionSupports is TrackRoadAdditionSupports: height 0 = none; per segment
// (kSegmentOffsets: segment i is flag 1 << i) an image (0 = none, with its
// colour remap) and a frequency.
type additionSupports struct {
	height    int
	occupied  uint16
	images    [9]int
	remaps    [9][]uint8
	frequency [9]uint8
}

// SetOccupiedAdditionSupportSegments is setOccupiedAdditionSupportSegments
// (stations block supports: SegmentsAll).
func (s *Session) SetOccupiedAdditionSupportSegments(seg uint16) { s.occupiedSupports = seg }

// setAdditionSupport fills the session's supports from an addition piece:
// track pieces use their first segment's image 0, road pieces one image per
// segment (images 0, 2 …).
func (s *Session) setAdditionSupport(height int, segments uint16, images []uint32, perSegment bool, frequency uint8, base int, remap []uint8) {
	sup := additionSupports{height: height, occupied: s.occupiedSupports}
	i := 0
	for seg := range 9 {
		if segments&(1<<seg) == 0 {
			continue
		}
		k := 0
		if perSegment {
			k = i * 2
		}
		if k < len(images) {
			sup.images[seg], sup.remaps[seg], sup.frequency[seg] = base+int(images[k]), remap, frequency
		}
		i++
		if !perSegment {
			break // bitScanForward: the first segment only
		}
	}
	s.supports = sup
}

var (
	supportBBOffsets = [2][9]Pos3{
		{{2, 2, 6}, {28, 2, 6}, {2, 28, 6}, {28, 28, 6}, {15, 15, 6}, {15, 2, 6}, {2, 15, 6}, {28, 15, 6}, {15, 28, 6}},
		{{2, 2, 28}, {28, 2, 28}, {2, 28, 28}, {28, 28, 28}, {15, 15, 28}, {15, 2, 28}, {2, 15, 28}, {28, 15, 28}, {15, 28, 28}},
	}
	supportBBLengths     = [2]Pos3{{1, 1, 17}, {1, 1, 1}}
	frequencyRotationMap = [4][4]uint8{{1 << 0, 1 << 2, 1 << 1, 1 << 3}, {1 << 3, 1 << 0, 1 << 2, 1 << 1}, {1 << 1, 1 << 3, 1 << 0, 1 << 2}, {1 << 2, 1 << 1, 1 << 3, 1 << 0}}
)

// PaintSupports is paintSupports (0x0046748F), run at a tile's end loop when
// supports were set: nothing under a roofed bridge; per segment with an
// image, not occupied and not skipped by its frequency (by the tile's x / y
// parity), the support's two images at the support height.
func (s *Session) PaintSupports(bridgeHasRoof func(objectID uint8) bool, sprite func(int) *gfx8.Sprite) {
	sup := s.supports
	s.supports = additionSupports{}
	if sup.height == 0 {
		return
	}
	if s.bridge.Height != -1 && bridgeHasRoof != nil && bridgeHasRoof(s.bridge.ObjectID) {
		return
	}
	px, py := s.spriteX, s.spriteY
	line := frequencyRotationMap[s.Rotation&3]
	for i := range 9 {
		if sup.images[i] == 0 || sup.occupied&(1<<i) != 0 {
			continue
		}
		f := sup.frequency[i]
		if (f&line[0] != 0 && px&0x20 == 0) || (f&line[1] != 0 && py&0x20 == 0) ||
			(f&line[2] != 0 && px&0x20 != 0) || (f&line[3] != 0 && py&0x20 != 0) {
			continue
		}
		h := Pos3{0, 0, sup.height}
		for j := range 2 {
			sp := sprite(sup.images[i] + j)
			if sp == nil {
				continue
			}
			o := supportBBOffsets[j][i]
			s.AddAsParent(Colour(sp, sup.remaps[i]), h, Pos3{o.X, o.Y, o.Z + sup.height}, supportBBLengths[j])
		}
	}
}
