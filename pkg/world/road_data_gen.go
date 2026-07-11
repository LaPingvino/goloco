package world

// Code generated from OpenLoco src/OpenLoco/src/Map/Track/TrackData.cpp
// (kRoadCoordinates 0x4F6F8C and roadPiece0..9). Do not edit by hand.
//
// Road equivalent of track_data_gen.go. Reuses trackCoord (same layout:
// begin/end rotation + end position delta in world units, 1 tile = 32).
// Indexed by roadAndDirection = roadID<<3 | reversedBit<<2 | rotation.

var kRoadCoordinates = [80]trackCoord{
	// roadID 0 — straight
	{0, 0, -32, 0, 0},
	{1, 1, 0, 32, 0},
	{2, 2, 32, 0, 0},
	{3, 3, 0, -32, 0},
	{2, 2, 32, 0, 0},
	{3, 3, 0, -32, 0},
	{0, 0, -32, 0, 0},
	{1, 1, 0, 32, 0},
	// roadID 1 — leftCurveVerySmall
	{0, 3, 0, -32, 0},
	{1, 0, -32, 0, 0},
	{2, 1, 0, 32, 0},
	{3, 2, 32, 0, 0},
	{1, 2, 32, 0, 0},
	{2, 3, 0, -32, 0},
	{3, 0, -32, 0, 0},
	{0, 1, 0, 32, 0},
	// roadID 2 — rightCurveVerySmall
	{0, 1, 0, 32, 0},
	{1, 2, 32, 0, 0},
	{2, 3, 0, -32, 0},
	{3, 0, -32, 0, 0},
	{3, 2, 32, 0, 0},
	{0, 3, 0, -32, 0},
	{1, 0, -32, 0, 0},
	{2, 1, 0, 32, 0},
	// roadID 3 — leftCurveSmall
	{0, 3, -32, -64, 0},
	{1, 0, -64, 32, 0},
	{2, 1, 32, 64, 0},
	{3, 2, 64, -32, 0},
	{1, 2, 64, 32, 0},
	{2, 3, 32, -64, 0},
	{3, 0, -64, -32, 0},
	{0, 1, -32, 64, 0},
	// roadID 4 — rightCurveSmall
	{0, 1, -32, 64, 0},
	{1, 2, 64, 32, 0},
	{2, 3, 32, -64, 0},
	{3, 0, -64, -32, 0},
	{3, 2, 64, -32, 0},
	{0, 3, -32, -64, 0},
	{1, 0, -64, 32, 0},
	{2, 1, 32, 64, 0},
	// roadID 5 — straightSlopeUp
	{0, 0, -64, 0, 16},
	{1, 1, 0, 64, 16},
	{2, 2, 64, 0, 16},
	{3, 3, 0, -64, 16},
	{2, 2, 64, 0, -16},
	{3, 3, 0, -64, -16},
	{0, 0, -64, 0, -16},
	{1, 1, 0, 64, -16},
	// roadID 6 — straightSlopeDown
	{0, 0, -64, 0, -16},
	{1, 1, 0, 64, -16},
	{2, 2, 64, 0, -16},
	{3, 3, 0, -64, -16},
	{2, 2, 64, 0, 16},
	{3, 3, 0, -64, 16},
	{0, 0, -64, 0, 16},
	{1, 1, 0, 64, 16},
	// roadID 7 — straightSteepSlopeUp
	{0, 0, -32, 0, 16},
	{1, 1, 0, 32, 16},
	{2, 2, 32, 0, 16},
	{3, 3, 0, -32, 16},
	{2, 2, 32, 0, -16},
	{3, 3, 0, -32, -16},
	{0, 0, -32, 0, -16},
	{1, 1, 0, 32, -16},
	// roadID 8 — straightSteepSlopeDown
	{0, 0, -32, 0, -16},
	{1, 1, 0, 32, -16},
	{2, 2, 32, 0, -16},
	{3, 3, 0, -32, -16},
	{2, 2, 32, 0, 16},
	{3, 3, 0, -32, 16},
	{0, 0, -32, 0, 16},
	{1, 1, 0, 32, 16},
	// roadID 9 — turnaround
	{0, 2, 32, 0, 0},
	{1, 3, 0, -32, 0},
	{2, 0, -32, 0, 0},
	{3, 1, 0, 32, 0},
	{0, 2, 32, 0, 0},
	{1, 3, 0, -32, 0},
	{2, 0, -32, 0, 0},
	{3, 1, 0, 32, 0},
}

// kRoadPieceTiles[roadID] lists each sequence tile's offset from the piece
// origin at rotation 0, in world units (x, y) and world z (int16).
var kRoadPieceTiles = [10][][3]int16{
	0: {{0, 0, 0}},
	1: {{0, 0, 0}},
	2: {{0, 0, 0}},
	3: {{0, 0, 0}, {0, -32, 0}, {-32, 0, 0}, {-32, -32, 0}},
	4: {{0, 0, 0}, {0, 32, 0}, {-32, 0, 0}, {-32, 32, 0}},
	5: {{0, 0, 0}, {-32, 0, 0}},
	6: {{0, 0, -16}, {-32, 0, -16}},
	7: {{0, 0, 0}},
	8: {{0, 0, -16}},
	9: {{0, 0, 0}},
}
