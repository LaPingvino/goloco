package construct

// trackCostFactor and roadCostFactor are TrackData's getTrackMiscData /
// getRoadMiscData costFactor per track and road id (printed by compiling
// upstream's Map/Track/TrackData.cpp).
var trackCostFactor = [44]int32{256, 362, 201, 201, 603, 603, 1005, 1005, 703, 703, 703, 703, 896, 896, 544, 544,
	312, 312, 848, 848, 848, 848, 1055, 1055, 1055, 1055, 256, 256, 100, 301, 301, 100, 294, 294, 312, 312,
	312, 312, 276, 276, 276, 276, 603, 603}
var roadCostFactor = [10]int32{256, 201, 201, 603, 603, 544, 544, 312, 312, 603}

// InflationAdjustedCost is Economy::getInflationAdjustedCost: costFactor ×
// currencyMultiplicationFactor[costIndex] >> divisor.
func InflationAdjustedCost(costFactor int16, costIndex uint8, divisor uint, factors *[32]uint32) int32 {
	if int(costIndex) >= len(factors) {
		return 0
	}
	return int32(int64(costFactor) * int64(factors[costIndex]) >> divisor)
}

// TrackPieceCost is CreateTrack's track part of the cost: the track
// object's inflation-adjusted build cost (divisor 10) × the piece's cost
// factor / 256.
// PARTIAL: the mods (track extras), clearing buildings/trees, level
// crossings, bridges and tunnels that CreateTrack adds.
func TrackPieceCost(trackID int, buildCostFactor int16, costIndex uint8, factors *[32]uint32) int32 {
	if trackID < 0 || trackID >= len(trackCostFactor) {
		return 0
	}
	return InflationAdjustedCost(buildCostFactor, costIndex, 10, factors) * trackCostFactor[trackID] / 256
}

// RoadPieceCost is CreateRoad's road part of the cost (as TrackPieceCost).
func RoadPieceCost(roadID int, buildCostFactor int16, costIndex uint8, factors *[32]uint32) int32 {
	if roadID < 0 || roadID >= len(roadCostFactor) {
		return 0
	}
	return InflationAdjustedCost(buildCostFactor, costIndex, 10, factors) * roadCostFactor[roadID] / 256
}
