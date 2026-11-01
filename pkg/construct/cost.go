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
// PARTIAL: clearing buildings/trees, level crossings, bridges and tunnels
// that CreateTrack adds (the mods: ModsCost).
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

// trackCompatibleFlags and roadCompatibleFlags are getTrackMiscData /
// getRoadMiscData compatibleFlags per id (printed from TrackData.cpp): the
// piece kind an extra's trackPieces/roadPieces must cover.
var trackCompatibleFlags = [44]uint16{0, 1, 16, 16, 8, 8, 4, 4, 2, 2, 2, 2, 512, 512, 32, 32, 64, 64, 296, 296, 296, 296,
	328, 328, 328, 328, 128, 128, 144, 144, 144, 144, 128, 128, 192, 192, 192, 192, 128, 128, 128, 128, 128, 128}
var roadCompatibleFlags = [10]uint16{0, 2, 2, 1, 1, 4, 4, 8, 8, 32}

// Mod is a track or road extra as the cost needs it.
type Mod struct {
	Pieces          uint16
	BuildCostFactor int16
	CostIndex       uint8
}

// ModsCost is CreateTrack's (road: CreateRoad's) mods part: for each chosen
// mod (bit i of mods, mods[i] its object, nil if none) that suits the piece
// (validMods: its pieces cover the piece's compatibleFlags), its
// inflation-adjusted build cost × the piece's cost factor / 256.
func ModsCost(road bool, id int, mods uint8, objs []*Mod, factors *[32]uint32) int32 {
	compat, factor := trackCompatibleFlags[:], trackCostFactor[:]
	if road {
		compat, factor = roadCompatibleFlags[:], roadCostFactor[:]
	}
	if id < 0 || id >= len(compat) {
		return 0
	}
	var total int32
	for i, m := range objs {
		if m == nil || mods&(1<<i) == 0 || m.Pieces&compat[id] != compat[id] {
			continue
		}
		total += InflationAdjustedCost(m.BuildCostFactor, m.CostIndex, 10, factors) * factor[id] / 256
	}
	return total
}
