package main

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// reorderScenarioObjects puts every object type the map refers to by index in
// the scenario's slot order (its RequiredObjects chunk), as the object
// manager's slots are upstream. One list for the title map and loaded games,
// so no type can be missed on one path (industry, airport and dock were, which
// gave title-map industries and their fields another industry's images).
func reorderScenarioObjects(om *objects.ObjectManager, sc *scenario.Scenario) {
	for _, r := range []struct {
		order []string
		apply func([]string)
	}{
		{sc.LandObjectOrder, om.ReorderLandObjects},
		{sc.CargoObjectOrder, om.ReorderCargoObjects},
		{sc.TreeObjectOrder, om.ReorderTreeObjects},
		{sc.BuildingObjectOrder, om.ReorderBuildingObjects},
		{sc.WallObjectOrder, om.ReorderWallObjects},
		{sc.TrainSignalObjectOrder, om.ReorderTrainSignalObjects},
		{sc.LevelCrossingObjectOrder, om.ReorderLevelCrossingObjects},
		{sc.BridgeObjectOrder, om.ReorderBridgeObjects},
		{sc.TrainStationObjectOrder, om.ReorderTrainStationObjects},
		{sc.RoadStationObjectOrder, om.ReorderRoadStationObjects},
		{sc.TrackObjectOrder, om.ReorderTrackObjects},
		{sc.RoadObjectOrder, om.ReorderRoadObjects},
		{sc.VehicleObjectOrder, om.ReorderVehicleObjects},
		{sc.IndustryObjectOrder, om.ReorderIndustryObjects},
		{sc.AirportObjectOrder, om.ReorderAirportObjects},
		{sc.DockObjectOrder, om.ReorderDockObjects},
	} {
		if len(r.order) > 0 {
			r.apply(r.order)
		}
	}
}
