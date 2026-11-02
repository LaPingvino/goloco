package sim

// Movement sub-positions — the accessors of Map/Track/SubpositionData.cpp
// over the generated tables (subposition_gen.go).

// TrackSubPosition is getTrackSubPositon(trackAndDirection).
func TrackSubPosition(trackAndDirection uint16) []MoveInfo {
	if int(trackAndDirection) >= len(TrackSubPositions) {
		return nil
	}
	return TrackSubPositions[trackAndDirection]
}

// RoadSubPosition is getRoadSubPositon: by the game's traffic handedness
// (right-hand when set), indexed [bit 8][bit 7][trackAndDirection & 0x7F].
func RoadSubPosition(trackAndDirection uint16, rightHand bool) []MoveInfo {
	data := RoadSubPositionsLeft
	if rightHand {
		data = RoadSubPositionsRight
	}
	a, b := 0, 0
	if trackAndDirection&(1<<8) != 0 {
		a = 1
	}
	if trackAndDirection&(1<<7) != 0 {
		b = 1
	}
	i := (a*2+b)*80 + int(trackAndDirection&0x7F) // [2][2][10·8]
	if i >= len(data) {
		return nil
	}
	return data[i]
}

// RoadPlacementSubPosition is getRoadPlacementSubPositon.
func RoadPlacementSubPosition(trackAndDirection uint16) []MoveInfo {
	if int(trackAndDirection) >= len(RoadPlacementSubPositions) {
		return nil
	}
	return RoadPlacementSubPositions[trackAndDirection]
}
