package sim

import (
	"encoding/binary"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// ReorderObjectsForScenario puts every object type the map refers to by index in
// the scenario's slot order (its RequiredObjects chunk), as the object
// manager's slots are upstream. One list for the game (title map and loaded
// games) and the tests, so no type can be missed on one path.
func ReorderObjectsForScenario(m *objects.ObjectManager, sc *scenario.Scenario) {
	for _, r := range []struct {
		order []string
		apply func([]string)
	}{
		{sc.LandObjectOrder, m.ReorderLandObjects},
		{sc.CargoObjectOrder, m.ReorderCargoObjects},
		{sc.CompetitorObjectOrder, m.ReorderCompetitorObjects},
		{sc.TownNamesObjectOrder, m.ReorderTownNames},
		{sc.RegionObjectOrder, m.ReorderRegion},
		{sc.ClimateObjectOrder, m.ReorderClimate},
		{sc.HillShapesObjectOrder, m.ReorderHillShapes},
		{sc.TreeObjectOrder, m.ReorderTreeObjects},
		{sc.BuildingObjectOrder, m.ReorderBuildingObjects},
		{sc.WallObjectOrder, m.ReorderWallObjects},
		{sc.TrainSignalObjectOrder, m.ReorderTrainSignalObjects},
		{sc.LevelCrossingObjectOrder, m.ReorderLevelCrossingObjects},
		{sc.BridgeObjectOrder, m.ReorderBridgeObjects},
		{sc.TrainStationObjectOrder, m.ReorderTrainStationObjects},
		{sc.RoadStationObjectOrder, m.ReorderRoadStationObjects},
		{sc.TrackObjectOrder, m.ReorderTrackObjects},
		{sc.RoadObjectOrder, m.ReorderRoadObjects},
		{sc.VehicleObjectOrder, m.ReorderVehicleObjects},
		{sc.IndustryObjectOrder, m.ReorderIndustryObjects},
		{sc.AirportObjectOrder, m.ReorderAirportObjects},
		{sc.DockObjectOrder, m.ReorderDockObjects},
		{sc.SoundObjectOrder, m.ReorderSoundObjects},
	} {
		if len(r.order) > 0 {
			r.apply(r.order)
		}
	}
	m.TrackExtraOrder = sc.TrackExtraOrder
	m.RoadExtraOrder = sc.RoadExtraOrder
}

// Object list slots the game changes after the load (ObjectManager.h's
// slot order): the currency (one slot at 129: loadPreferredCurrency), the
// competitors (32 from 826: createCompany's selectNewCompetitor loads one).
const (
	objSlotCurrency    = 129
	objSlotCompetitors = 826
	objHeaderSize      = 16
)

// RequiredObjects is ObjectManager::getHeaders for a save: the loaded file's
// object list (base) with the slots the game has changed since — the
// currency in use and every competitor slot — as the objects now loaded.
func RequiredObjects(base []byte, om *objects.ObjectManager) []byte {
	req := append([]byte(nil), base...)
	if om == nil {
		return req
	}
	put := func(slot int, lo *objects.LoadedObject) {
		if (slot+1)*objHeaderSize > len(req) {
			return
		}
		h := req[slot*objHeaderSize : (slot+1)*objHeaderSize]
		if lo == nil || lo.Header == nil {
			return // nothing goloco loaded there: the file's own entry stays
		}
		binary.LittleEndian.PutUint32(h[0:], lo.Header.Flags)
		copy(h[4:12], lo.Header.Name[:])
		binary.LittleEndian.PutUint32(h[12:], lo.Header.Checksum)
	}
	if om.CurrencySlot != "" {
		if lo := om.GetObject(om.CurrencySlot); lo != nil {
			put(objSlotCurrency, lo)
		}
	}
	byObject := map[any]*objects.LoadedObject{}
	for _, lo := range om.Objects {
		if c, ok := lo.Object.(*objects.CompetitorObject); ok {
			byObject[c] = lo
		}
	}
	for i := range 32 {
		var lo *objects.LoadedObject
		if i < len(om.CompetitorObjects) && om.CompetitorObjects[i] != nil {
			lo = byObject[om.CompetitorObjects[i]]
		}
		put(objSlotCompetitors+i, lo)
	}
	return req
}
