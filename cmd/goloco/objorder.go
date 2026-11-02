package main

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
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
	om.TrackExtraOrder = sc.TrackExtraOrder
	om.RoadExtraOrder = sc.RoadExtraOrder
	useCurrency(om, sc.CurrencyObject)
}

// useCurrency makes dat the loaded currency object: its slot, then
// Gfx::loadCurrency and formatCurrency's object (scenario load, and Options'
// currencyDropdown).
func useCurrency(om *objects.ObjectManager, dat string) {
	om.CurrencySlot = dat
	if c := om.GetCurrencyObject(dat); c != nil {
		ui.LoadCurrency(c.ObjectIcon)
		uilayout.SetCurrency(c.Prefix, c.Suffix, c.Separator, c.Factor)
	}
}

// loadPreferredCurrency is Scenario.cpp loadPreferredCurrency: the preferred
// currency becomes the loaded one, unless it already is or it can't be
// loaded (then the old one stays).
func (g *Game) loadPreferredCurrency() {
	if g.objMgr == nil || g.preferredCurrency == "" || g.preferredCurrency == g.objMgr.CurrencySlot {
		return
	}
	if g.objMgr.GetCurrencyObject(g.preferredCurrency) == nil {
		return
	}
	useCurrency(g.objMgr, g.preferredCurrency)
}

// loadPreferredCurrencyAlways: only with usePreferredCurrencyAlways.
func (g *Game) loadPreferredCurrencyAlways() {
	if g.preferredCurrencyFlags[1] {
		g.loadPreferredCurrency()
	}
}

// loadPreferredCurrencyNewGame: with usePreferredCurrencyForNewGames, else
// as loadPreferredCurrencyAlways.
func (g *Game) loadPreferredCurrencyNewGame() {
	if !g.preferredCurrencyFlags[0] {
		g.loadPreferredCurrencyAlways()
		return
	}
	g.loadPreferredCurrency()
}
