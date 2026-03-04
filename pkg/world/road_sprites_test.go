package world

import "testing"

// TestRoadSpriteTableValidity verifies basic invariants of the road sprite
// offset tables so regressions in kRoadPartsStyle0/1/2 are caught early.

// Road IDs and their expected piece counts (sequenceIndex count).
var roadPieceCounts = [10]int{
	1, // 0: Straight
	1, // 1: LeftCurveVerySmall
	1, // 2: RightCurveVerySmall
	4, // 3: LeftCurveSmall
	4, // 4: RightCurveSmall
	2, // 5: StraightSlopeUp
	2, // 6: StraightSlopeDown
	1, // 7: StraightSteepSlopeUp
	1, // 8: StraightSteepSlopeDown
	1, // 9: Turnaround
}

func TestStyle0PieceCounts(t *testing.T) {
	for id, want := range roadPieceCounts {
		if got := len(kRoadPartsStyle0[id]); got != want {
			t.Errorf("kRoadPartsStyle0[%d]: %d pieces, want %d", id, got, want)
		}
	}
}

func TestStyle2PieceCounts(t *testing.T) {
	for id, want := range roadPieceCounts {
		if got := len(kRoadPartsStyle2[id]); got != want {
			t.Errorf("kRoadPartsStyle2[%d]: %d pieces, want %d", id, got, want)
		}
	}
}

// TestStyle0StraightNonZero guards against the bug where Style0 straight
// roads had all-zero offsets (making them invisible).
// OpenLoco reference: src/OpenLoco/src/Objects/RoadObject.h Style0 namespace
//   kStraight0NE=34, kStraight0SE=35
func TestStyle0StraightNonZero(t *testing.T) {
	pieces := kRoadPartsStyle0[0]
	if len(pieces) == 0 {
		t.Fatal("Style0 straight has no pieces")
	}
	img := pieces[0].img
	for rot, v := range img {
		if v == 0 {
			t.Errorf("Style0 straight rotation %d: sprite offset=0 (road will be invisible)", rot)
		}
	}
}

// TestStyle0StraightOffsets verifies exact values: NE=34, SE=35 per RoadObject.h.
func TestStyle0StraightOffsets(t *testing.T) {
	img := kRoadPartsStyle0[0][0].img
	// NE=34, SE=35; SW/NW reuse NE/SE (rotationally symmetric for straight)
	want := [4]uint32{34, 35, 34, 35}
	if img != want {
		t.Errorf("Style0 straight offsets = %v, want %v", img, want)
	}
}

// TestStyle2StraightOffsets verifies exact values: NE=34, SE=35, SW=85, NW=86.
// OpenLoco reference: src/OpenLoco/src/Objects/RoadObject.h Style2 namespace
func TestStyle2StraightOffsets(t *testing.T) {
	img := kRoadPartsStyle2[0][0].img
	want := [4]uint32{34, 35, 85, 86}
	if img != want {
		t.Errorf("Style2 straight offsets = %v, want %v", img, want)
	}
}

// ---------------------------------------------------------------------------
// Multi-tile merge (junction) system
// OpenLoco reference: PaintRoad.cpp kExitsToMergeId, PaintRoadCommonData.h bridgeEdges
// ---------------------------------------------------------------------------

// TestRoadMergeExitsStraight verifies that a straight road contributes NE+SW for
// rotation 0 and SE+NW for rotation 1 (rotl4 of 0b0101).
func TestRoadMergeExitsStraight(t *testing.T) {
	cases := []struct {
		rotation int
		want     uint8
		desc     string
	}{
		{0, 0b0101, "rot=0: NE+SW"},
		{1, 0b1010, "rot=1: SE+NW"},
		{2, 0b0101, "rot=2: same as rot=0 (straight is symmetric)"},
		{3, 0b1010, "rot=3: same as rot=1"},
	}
	for _, c := range cases {
		got := kRoadMergeExits[0][c.rotation]
		if got != c.want {
			t.Errorf("kRoadMergeExits[straight][%d] = 0b%04b, want 0b%04b — %s",
				c.rotation, got, c.want, c.desc)
		}
	}
}

// TestExitsToMergeIdKeyValues checks the most important entries:
// straight, T-junctions, and crossroads.
func TestExitsToMergeIdKeyValues(t *testing.T) {
	cases := []struct {
		exits    uint8
		wantID   uint8
		desc     string
	}{
		{0b0101, 0, "NE+SW straight"},
		{0b1010, 1, "SE+NW straight"},
		{0b0110, 2, "SE+SW right curve"},
		{0b1100, 3, "SW+NW right curve rotated"},
		{0b1001, 4, "NW+NE right curve rotated"},
		{0b0011, 5, "NE+SE right curve rotated"},
		{0b1101, 6, "NE+SE+NW T-junction (missing SW)"},
		{0b1011, 7, "NE+SW+NW T-junction (missing SE)"},
		{0b0111, 8, "NE+SE+SW T-junction (missing NW)"},
		{0b1110, 9, "SE+SW+NW T-junction (missing NE)"},
		{0b1111, 10, "all 4 exits — crossroads"},
	}
	for _, c := range cases {
		got := kExitsToMergeId[c.exits]
		if got != c.wantID {
			t.Errorf("kExitsToMergeId[0b%04b] = %d, want %d — %s",
				c.exits, got, c.wantID, c.desc)
		}
	}
}

// TestMergeSpriteStraight verifies that a single straight road at rot=0 produces
// exits 0b0101 → mergeId=0 → sprite at roadObj.Image + 34 + 0 = roadObj.Image + 34.
// This ensures single-road tiles still render as plain straights.
func TestMergeSpriteStraight(t *testing.T) {
	exits := kRoadMergeExits[0][0] // straight, rot=0
	mergeID := kExitsToMergeId[exits]
	spriteOffset := kMergeBaseOffset + uint32(mergeID)
	const wantOffset = uint32(34) // kStraight0NE
	if spriteOffset != wantOffset {
		t.Errorf("single straight NE: spriteOffset=%d, want %d", spriteOffset, wantOffset)
	}
}

// TestMergeSpriteCrossroads verifies that two crossing straight roads produce
// exits 0b1111 → mergeId=10 → sprite at roadObj.Image + 34 + 10 = roadObj.Image + 44.
// OpenLoco reference: RoadObject.h Style0::kJunctionCrossroads0NE = 44
func TestMergeSpriteCrossroads(t *testing.T) {
	// Straight NE-SW at rot=0 AND straight SE-NW at rot=1
	exits := kRoadMergeExits[0][0] | kRoadMergeExits[0][1]
	if exits != 0b1111 {
		t.Errorf("crossroads exits = 0b%04b, want 0b1111", exits)
	}
	mergeID := kExitsToMergeId[exits]
	if mergeID != 10 {
		t.Errorf("crossroads mergeId = %d, want 10", mergeID)
	}
	spriteOffset := kMergeBaseOffset + uint32(mergeID)
	const wantOffset = uint32(44) // kJunctionCrossroads0NE
	if spriteOffset != wantOffset {
		t.Errorf("crossroads spriteOffset=%d, want %d", spriteOffset, wantOffset)
	}
}

// TestMergeSpritesTJunction verifies a T-junction formed by straight NE-SW plus a
// very-small-right-curve at rot=3 (NE+SE), producing exits 0b0111 → kJunctionLeft0SW.
func TestMergeSpritesTJunction(t *testing.T) {
	// exits: straight rot=0 (NE+SW = 0b0101) | right-curve rot=3 (NE+SE = 0b0011)
	exits := kRoadMergeExits[0][0] | kRoadMergeExits[2][3]
	if exits != 0b0111 {
		t.Errorf("T-junction exits = 0b%04b, want 0b0111", exits)
	}
	mergeID := kExitsToMergeId[exits]
	if mergeID != 8 {
		t.Errorf("T-junction mergeId = %d, want 8", mergeID)
	}
	spriteOffset := kMergeBaseOffset + uint32(mergeID)
	const wantOffset = uint32(42) // kJunctionLeft0SW
	if spriteOffset != wantOffset {
		t.Errorf("T-junction spriteOffset=%d, want %d", spriteOffset, wantOffset)
	}
}
