package sim

// Track and road connection ends — Map/Track/Track.cpp getTrackConnectionEnd /
// getRoadConnectionEnd over TrackData::getUnkTrack / getUnkRoad
// (TrackCoordinates / RoadCoordinates, track_data_gen.go / road_data_gen.go).

// ConnectionEnd is Track::ConnectionEnd: the position (world units) of the
// tile the piece's end leads into, and the rotation it arrives with.
type ConnectionEnd struct {
	X, Y, Z      int32
	NextRotation uint8
}

func connectionEnd(c TrackCoord, x, y, z int32) ConnectionEnd {
	return ConnectionEnd{x + int32(c.DX), y + int32(c.DY), z + int32(c.DZ), c.RotEnd}
}

// TrackConnectionEnd is getTrackConnectionEnd(pos, trackAndDirection).
func TrackConnectionEnd(x, y, z int32, trackAndDirection uint16) (ConnectionEnd, bool) {
	if int(trackAndDirection) >= len(TrackCoordinates) {
		return ConnectionEnd{}, false
	}
	return connectionEnd(TrackCoordinates[trackAndDirection], x, y, z), true
}

// RoadConnectionEnd is getRoadConnectionEnd(pos, trackAndDirection).
func RoadConnectionEnd(x, y, z int32, trackAndDirection uint16) (ConnectionEnd, bool) {
	if int(trackAndDirection) >= len(RoadCoordinates) {
		return ConnectionEnd{}, false
	}
	return connectionEnd(RoadCoordinates[trackAndDirection], x, y, z), true
}
