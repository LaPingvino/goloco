package world

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// station.go — goloco's minimal station placement for the construction
// window: the track piece's origin element gains HasStation and a
// scenario.StationElement is appended so paintStation draws it.
//
// NOT A PORT (docs/deviations.md): upstream's createTrainStation game command
// (Station record, name, catchment, cargo) is not ported; the element's
// station id is a running count. The invented passenger supply and
// board→travel→deliver loop that used to live here are retired with the
// SimVehicles (item 5).

// PlaceStationAtHead attaches a train station to the track piece under the
// construction head. Returns false when the head is inactive or its tile has no
// track origin element to carry the station.
func (w *World) PlaceStationAtHead(stationObjID uint8) bool {
	h := w.ConsHead
	if !h.Active {
		return false
	}
	return w.PlaceStationOnTile(h.X, h.Y, stationObjID)
}

// PlaceStationOnTile marks the SequenceIndex-0 track element on tile (x,y) as a
// station, appends a matching scenario.StationElement (so paintStation renders
// it). Returns false if the tile has no track origin or already carries a
// station.
func (w *World) PlaceStationOnTile(x, y int, stationObjID uint8) bool {
	t := w.getTile(x, y)
	if t == nil {
		return false
	}
	var te *scenario.TrackElement
	for i := range t.Tracks {
		if t.Tracks[i].SequenceIndex == 0 {
			te = &t.Tracks[i]
			break
		}
	}
	if te == nil || te.HasStation {
		return false
	}
	te.HasStation = true
	t.Stations = append(t.Stations, scenario.StationElement{
		StationType:   0, // train station
		ObjectID:      stationObjID,
		Rotation:      te.Rotation & 3,
		SequenceIndex: 0,
		StationID:     uint16(w.placedStations), // no Station record: PARTIAL
		Owner:         te.Owner,
		BaseZ:         te.BaseZ,
		ClearZ:        te.BaseZ + 8,
	})
	t.InsertRef(scenario.KindStation, len(t.Stations)-1, te.BaseZ)

	w.placedStations++
	log.Printf("[Station] placed at (%d,%d) z%d objSlot=%d", x, y, te.BaseZ, stationObjID)
	return true
}
