package main

import (
	"image"
	"log"
	"os"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// toolbar.go — the top toolbar's buttons and their menus (ToolbarTop.cpp,
// ToolbarTopCommon.cpp).

// toolbarDropdown is Dropdown::showBelow for a toolbar button: below the
// button, in the button's window colour (the interface skin's top-toolbar
// colour for its group), translucent.
func (g *Game) toolbarDropdown(btn *ui.ToolbarButton, items []ui.DropdownItem) *ui.DropdownMenu {
	d := ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
	d.Place(btn.X, btn.Y, btn.Width, btn.Height, g.sw, g.sh)
	g.setToolbarColour(d, btn)
	d.OpenedByPress(image.Rect(btn.X, btn.Y, btn.X+btn.Width, btn.Y+btn.Height)) // as toolbarImageMenu
	return d
}

// toolbarImageMenu is the toolbar's menu of images with text (Dropdown's
// showBelow with 25 px items, flag 1 << 6), in the button's colour.
func (g *Game) toolbarImageMenu(btn *ui.ToolbarButton, items []ui.DropdownItem) *ui.DropdownMenu {
	d := ui.NewDropdownShowText(btn.X, btn.Y, btn.Width, btn.Height, ui.DropdownColour, items, 25, g.sw, g.sh)
	g.setToolbarColour(d, btn)
	// showBelow without flag 1<<7: the release of the opening press on the
	// button takes the highlighted item (MouseInput.cpp 0x4C7BC7)
	d.OpenedByPress(image.Rect(btn.X, btn.Y, btn.X+btn.Width, btn.Y+btn.Height))
	return d
}

// setToolbarColour gives a toolbar button's menu the interface skin's
// top-toolbar colour for the button's group.
func (g *Game) setToolbarColour(d *ui.DropdownMenu, btn *ui.ToolbarButton) {
	if sk := g.skin(); sk != nil && btn.ColorGroup >= 0 && btn.ColorGroup < 4 {
		cols := [4]uint8{sk.TopToolbarPrimaryColour, sk.TopToolbarSecondaryColour, sk.TopToolbarTertiaryColour, sk.TopToolbarQuaternaryColour}
		d.SetToolbarColour(cols[btn.ColorGroup])
	}
}

// handleToolbarButton handles clicks on toolbar buttons, matching the new
// OpenLoco-based button layout.
//
// Button order (indices):
//
//	0: Load/Save, 1: Audio, 2: (reserved/hidden)
//	3: Zoom, 4: Rotate, 5: View
//	6: Terraform, 7: Railroad, 8: Road, 9: Port/Airport, 10: Build Vehicles
//	11: Vehicles, 12: Stations, 13: Towns
//
// clickable(ToolbarTop.cpp): build_vehicles_menu port_menu vehicles_menu stations_menu
// clickable(ToolbarTop.cpp): railroad_menu
// clickable(ToolbarTopCommon.cpp): zoom_menu towns_menu
// clickable-partial(ToolbarTop.cpp): loadsave_menu audio_menu
// clickable(ToolbarTopCommon.cpp): railroad_menu road_menu
// (not yet: rotate_menu and view_menu only log; terraform_menu, port_menu, cheats_menu nothing)
func (g *Game) handleToolbarButton(idx int) {
	btn := &g.toolbar.Buttons[idx]
	tooltip := btn.Tooltip

	// Create appropriate dropdown or window based on button tooltip
	switch tooltip {
	case "Load/Save":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp:117-154
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuLoadGame), Action: func() { g.openLoadGameWindow() }},
			{Text: i18n.Label(i18n.MenuSaveGame), Action: func() { g.openSaveWindow() }},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuScreenshot), Action: func() {
				g.diagPending = true
				g.diagQuit = false
				g.diagTileX = -1
				g.diagTileY = -1
			}},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuAbout), Action: func() { log.Println("[Game] About selected") }},
			{Text: i18n.Label(i18n.Options), Action: func() { g.openOptionsWindow() }},
			{Separator: true},
			{Text: i18n.Label(i18n.MenuQuitToMenu), Action: func() { g.quitToMenu() }},
			{Text: i18n.Label(i18n.MenuExitOpenloco), Action: func() { os.Exit(0) }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)

	case "Audio":
		items := []ui.DropdownItem{
			{Text: "Stop Music", Action: func() {
				if g.audioMgr != nil {
					g.audioMgr.StopMusic()
					log.Println("[Game] Music stopped")
				}
			}},
		}
		g.dropdown = g.toolbarDropdown(btn, items)

	case "Zoom":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:95-132
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuZoomIn), Action: func() { g.w.ZoomIn() }},
			{Text: i18n.Label(i18n.MenuZoomOut), Action: func() { g.w.ZoomOut() }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)

	case "Rotate":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:135-143
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuRotateClockwise), Action: func() { log.Println("[Game] Rotate CW") }},
			{Text: i18n.Label(i18n.MenuRotateAntiClockwise), Action: func() { log.Println("[Game] Rotate CCW") }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)

	case "View":
		items := []ui.DropdownItem{
			{Text: i18n.Label(i18n.MenuUndergroundView), Action: func() { log.Println("[Game] Underground view") }},
			{Text: "See Through Tracks", Action: func() { log.Println("[Game] See through tracks") }},
		}
		g.dropdown = g.toolbarDropdown(btn, items)
	case "Terraform":
		log.Println("[Game] Terraform menu (not yet implemented)")
	case "Railroad":
		g.constructionMenu(btn, true) // the company's track types, the last one highlighted

	case "Road":
		g.constructionMenu(btn, false)
	case "Port/Airport":
		g.portMenu(btn)
	case "Build Vehicles":
		g.buildVehiclesMenu(btn)
	case "Vehicles":
		g.vehiclesMenu(btn) // by type, with counts (ToolbarTop.cpp vehiclesMenuMouseDown)
	case "Stations":
		g.stationsMenu(btn) // by station type (ToolbarTop.cpp stationsMenuMouseDown)
	case "Towns":
		g.townsMenu(btn) // Towns / Industries (ToolbarTopCommon.cpp townsMenuMouseDown)
	default:
		// Not ported yet: no window (an invented placeholder window would not
		// be the game's own).
		log.Printf("[Game] %s: window not ported yet", tooltip)
	}
}

// constructionMenu is the toolbar's railroad menu (ToolbarTop.cpp
// railroadMenuMouseDown, 0x0043A0EF) or road menu (ToolbarTopCommon.cpp
// roadMenuMouseDown, 0x0043A19F): the company's available track or road
// types — each its object's image recoloured to the player's colour and its
// name — with the last one chosen highlighted, so a click without moving
// takes it (Dropdown's highlighted item on release). Choosing one is
// Construction::openWithFlags (openConstructionWithFlags). Upstream's
// railroad menu asks for GameCommands::getUpdatingCompanyId ("This looks
// odd"); outside a command that is the controlling company, as here.
func (g *Game) constructionMenu(btn *ui.ToolbarButton, rail bool) {
	if g.r.ObjMgr == nil {
		return
	}
	m, om := &g.w.Map, g.r.ObjMgr
	ctrl, _ := g.w.ControllingCompany()
	objs, last := m.CompanyAvailableRoads(uint8(ctrl), om), m.LastRoadOption
	if rail {
		objs, last = m.CompanyAvailableRailTracks(uint8(ctrl), om), m.LastRailroadOption
	}
	if len(objs) == 0 {
		return // "Sanity check: any objects available?"
	}
	colour := g.companyColours(ctrl).Main[0] // getPlayerCompanyColour: the main colour's primary
	var items []ui.DropdownItem
	highlighted := 0
	for i, obj := range objs {
		name, image := "", 0
		if obj&(1<<7) != 0 {
			if r := om.GetRoadObjectByIndex(int(obj &^ (1 << 7))); r != nil {
				name, image = r.Name, int(r.Image)
			}
		} else if t := om.GetTrackObjectByIndex(int(obj)); t != nil {
			name, image = t.Name, int(t.Image) // + TrackObj::ImageIds::kUiPreviewImage0 (0)
		}
		obj := obj
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: image, SpriteColour: colour, Text: name, Action: func() { g.openConstructionWithFlags(obj) }})
		if obj == last {
			highlighted = i
		}
	}
	d := g.toolbarImageMenu(btn, items)
	d.Highlighted = highlighted
	g.dropdown = d
}

// openConstructionWithFlags is Construction::openWithFlags for a track or
// road type (bit 7: a road object): the construction window opens with it,
// the piece and gradient reset, and the toolbar's last option follows
// (setTrackOptions). PARTIAL: upstream's keep-the-window case (a road over
// an any-road-type-compatible road keeps the open window and its head).
func (g *Game) openConstructionWithFlags(flags uint8) {
	if flags&(1<<7) != 0 {
		g.buildMode, g.buildObjID = buildModeRoad, int(flags&^(1<<7))
	} else {
		g.buildMode, g.buildObjID = buildModeTrack, int(flags)
	}
	g.consCurve, g.consSlope = 0, 0
	g.w.Map.SetTrackOptions(flags, g.r.ObjMgr)
	g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
}

// portMenu is ToolbarTop's portMenuMouseDown (0x0043A9C0): the airport item
// when the game state has a last airport, the ship port item when it has a
// dock (toolbar_menu_airport / _ship_port images, menu_airport / _ship_port),
// highlighted at the last option; none, no menu. Choosing one is
// portMenuDropdown: the option remembered, Construction::openWithFlags.
// PARTIAL: the airport and dock construction windows (logged for now).
func (g *Game) portMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil {
		return
	}
	m := &g.w.Map
	m.UpdateAvailableAirportAndDockOptions(g.r.ObjMgr) // upstream runs it on load / company changes; the state it leaves is what the menu reads
	skin := int(g.skin().ImageOffset)
	var items []ui.DropdownItem
	var options []int
	add := func(option, img, label int) {
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: skin + img, Text: i18n.Label(label), Action: func() {
			g.lastPortOption = option
			g.openPortConstruction(option)
		}})
		options = append(options, option)
	}
	if m.LastAirport != 0xFF {
		add(0, objects.ISToolbarMenuAirport, i18n.MenuAirport)
	}
	if m.LastShipPort != 0xFF {
		add(1, objects.ISToolbarMenuShipPort, i18n.MenuShipPort)
	}
	if len(items) == 0 {
		return
	}
	d := g.toolbarImageMenu(btn, items)
	if g.lastPortOption != options[0] && len(items) > 1 { // the second item when the last option isn't the first
		d.Highlighted = 1
	} else {
		d.Highlighted = 0
	}
	g.dropdown = d
}

// openPortConstruction is Construction::openWithFlags(1 << 31: airport,
// 1 << 30: dock). PARTIAL: the construction windows themselves.
func (g *Game) openPortConstruction(option int) {
	m := &g.w.Map
	if option == 0 { // createAirportConstructionWindow: lastSelectedStationType = lastAirport; rotation 0
		g.removeStationGhost()
		g.buildMode, g.consTab, g.consStation, g.airportRotation = buildModeAirport, 1, int(m.LastAirport), 0
		g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
		return
	}
	// createDockConstructionWindow: lastSelectedStationType = lastShipPort,
	// the Station tab alone (nonTrackWindow). PARTIAL: the dock tool (hover
	// ghost, placing) — next.
	g.removeStationGhost()
	g.buildMode, g.consTab, g.consStation = buildModeDock, 1, int(m.LastShipPort)
	g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
}

// buildVehiclesMenu is ToolbarTop.cpp buildVehiclesMenuMouseDown (0x0043AD1F):
// one item per vehicle type the controlling company may build
// (availableVehicles), its build image recoloured to the company colour with
// build_trains … (menu_sprite_stringid), 25 px items (showBelow, flag 1<<6);
// the highlighted item is lastBuildVehiclesOption taken as an item index, as
// upstream does (an enum value, not the item's place — kept). Choosing one is
// buildVehiclesMenuDropdown: lastBuildVehiclesOption = the type, then
// BuildVehicle::openByType. The button face shows the last type
// (pkg/ui toolbar Draw, kBuildVehicleImages).
func (g *Game) buildVehiclesMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil {
		return
	}
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	var avail uint16
	if company >= 0 && company < len(m.Companies) {
		avail = m.Companies[company].AvailableVehicles
	}
	var colour uint8
	if company >= 0 && company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	skin := int(g.skin().ImageOffset)
	var items []ui.DropdownItem
	for vt := 0; vt < 6; vt++ {
		if avail&(1<<vt) == 0 {
			continue
		}
		t := vt
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: skin + objects.ISBuildVehicleFrame0[vt], SpriteColour: colour,
			Text: i18n.Label(i18n.BuildTrains + t), Action: func() {
				m.LastBuildVehiclesOption = uint8(t)
				g.openBuildVehicle(t)
			}})
	}
	if len(items) == 0 {
		return
	}
	d := g.toolbarImageMenu(btn, items)
	if h := int(m.LastBuildVehiclesOption); h < len(items) {
		d.Highlighted = h
	}
	g.dropdown = d
}

// vehiclesMenu is ToolbarTop.cpp vehiclesMenuMouseDown (0x0043ABCB): one item
// per vehicle type the controlling company may have (availableVehicles), its
// image recoloured to the company colour and how many it owns
// ("{COMMA16} train" / "trains" — num_*_singular / _plural, by VehicleType),
// not counting ghosts; lastVehicleType highlighted (as an item index, as
// upstream). Choosing one is vehiclesMenuDropdown: lastVehicleType = the type,
// then VehicleList::open(the controlling company, type).
func (g *Game) vehiclesMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil {
		return
	}
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	var avail uint16
	if company >= 0 && company < len(m.Companies) {
		avail = m.Companies[company].AvailableVehicles
	}
	var colour uint8
	if company >= 0 && company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	var counts [6]int
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == scenario.VehicleTypeHead && int(e.Owner) == company && e.Flags38&0x10 == 0 && e.VehicleType < 6 { // Flags38::isGhost
			counts[e.VehicleType]++
		}
	}
	faces := [6]int{objects.ISVehicleTrainFrame0, objects.ISVehicleBusesFrame0, objects.ISVehicleTrucksFrame0,
		objects.ISVehicleTramsFrame0, objects.ISVehicleAircraftFrame0, objects.ISVehicleShipsFrame0}
	skin := int(g.skin().ImageOffset)
	var items []ui.DropdownItem
	for vt := 0; vt < 6; vt++ {
		if avail&(1<<vt) == 0 {
			continue
		}
		t := vt
		label := i18n.NumTrainsPlural + t
		if counts[t] == 1 {
			label = i18n.NumTrainsSingular + t
		}
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: skin + faces[t], SpriteColour: colour,
			Text: ui.FillArgs(i18n.Label(label), ui.Comma16(counts[t])), Action: func() {
				m.LastVehicleType = uint8(t)
				g.openVehicleList(t)
			}})
	}
	if len(items) == 0 {
		return
	}
	d := g.toolbarImageMenu(btn, items)
	if h := int(m.LastVehicleType); h < len(items) {
		d.Highlighted = h
	}
	g.dropdown = d
}

// stationsMenu is ToolbarTop.cpp stationsMenuMouseDown (0x0043A4E9): all
// stations, rail stations, road stations, airports, ship ports — the skin's
// images recoloured to the player's colour —, "All stations" highlighted.
// Choosing one is stationsMenuDropdown: StationList::open(the controlling
// company, that type).
func (g *Game) stationsMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil {
		return
	}
	company, _ := g.w.ControllingCompany()
	var colour uint8
	if company >= 0 && company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	skin := int(g.skin().ImageOffset)
	var items []ui.DropdownItem
	for t := 0; t < 5; t++ {
		tab := t
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: skin + objects.ISAllStations + t, SpriteColour: colour,
			Text: i18n.Label(i18n.AllStations + t), Action: func() { c, _ := g.w.ControllingCompany(); g.openStationList(c, tab) }})
	}
	d := g.toolbarImageMenu(btn, items)
	d.Highlighted = 0
	g.dropdown = d
}

// townsMenu is ToolbarTopCommon.cpp townsMenuMouseDown (0x0043A8CE): Towns
// and Industries (the skin's toolbar_menu_towns / _industries), the last one
// chosen highlighted (_lastTownOption, kept for the session as upstream).
// Choosing is townsMenuDropdown: TownList::open or IndustryList::open.
func (g *Game) townsMenu(btn *ui.ToolbarButton) {
	if g.skin() == nil {
		return
	}
	skin := int(g.skin().ImageOffset)
	items := []ui.DropdownItem{
		{MenuSprite: true, Sprite: skin + objects.ISToolbarMenuTowns, Text: i18n.Label(i18n.MenuTowns), Action: func() {
			g.lastTownOption = 0
			g.openTownsWindow()
		}},
		{MenuSprite: true, Sprite: skin + objects.ISToolbarMenuIndustries, Text: i18n.Label(i18n.MenuIndustries), Action: func() {
			g.lastTownOption = 1
			g.openIndustryList()
		}},
	}
	d := g.toolbarImageMenu(btn, items)
	d.Highlighted = g.lastTownOption
	g.dropdown = d
}
