package sim

import (
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
	} {
		if len(r.order) > 0 {
			r.apply(r.order)
		}
	}
	m.TrackExtraOrder = sc.TrackExtraOrder
	m.RoadExtraOrder = sc.RoadExtraOrder
}
