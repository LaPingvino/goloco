package main

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// vehicle_paint.go — the Vehicle window's paint tool (Vehicle.cpp Details:
// paintBrush, paintColourPrimary / Secondary, paintTool*), repainting
// through GameCommands::vehicleRepaint (pkg/sim vehiclerepaint.go).

// paintToolOn is isPaintToolActive for the vehicle's window.
func (g *Game) paintToolOn(headID uint16) bool { return g.paintTool && g.paintHead == headID }

// companyMainColours is switchTab's colours for the Details tab: the window
// owner's mainColours (primary, secondary).
func (g *Game) companyMainColours(headID uint16) [2]uint8 {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return [2]uint8{}
	}
	if o := int(m.Entities[tr.Head].Owner); o < len(m.CompanyColourSchemes) {
		return m.CompanyColourSchemes[o].Main
	}
	return [2]uint8{}
}

// paintColours is getPaintToolColour: the window's two swatches.
func (g *Game) paintColours(headID uint16) sim.ColourScheme {
	c := g.companyMainColours(headID)
	if w := g.vehicleWin(headID); w != nil {
		c = w.PaintColours
	}
	return sim.ColourScheme{Primary: c[0], Secondary: c[1]}
}

// paintBrushOpen is Details::onMouseDown(paintBrush): a text dropdown at the
// button — Repaint tool, Repaint entire vehicle — the first selected and
// highlighted; onDropdown: the second (or shift) paints the whole train and
// takes the tool if it isn't on, the first (or none: a plain click) toggles the tool
// (paintToolBegin's toolSet).
func (g *Game) paintBrushOpen(headID uint16, x, y, w, h int) {
	items := []ui.DropdownItem{{Text: i18n.Label(i18n.VehicleRepaintTool)}, {Text: i18n.Label(i18n.VehicleRepaintEntireVehicle)}}
	items[0].Selected = true
	d := ui.NewDropdownMenu(x, y+h, items)
	d.Colour = g.vehicleWinColour(headID)
	d.Highlighted = 0
	// showText without flag 1<<7: the release on the button chooses
	// (dropdownRegisterSelection(kDropdownItemUndefined) → itemIndex <= 0) —
	// a click on the brush takes the tool at once.
	d.OpenedByPress(image.Rect(x, y, x+w, y+h))
	d.OnSelect = func(item int) {
		if item == 1 || g.hasKeyModifier(shortcuts.ModShift) {
			g.paintEntireTrain(headID)
			if !g.paintToolOn(headID) {
				g.paintToolBegin(headID)
			}
			return
		}
		if item <= 0 {
			g.paintToolBegin(headID)
		}
	}
	g.dropdown = d
}

// paintColourOpen is Details::onMouseDown(paintColourPrimary / Secondary):
// Dropdown::showColour with every colour (0x7FFFFFFF), the swatch's own
// selected; onDropdown sets the swatch to the colour picked.
func (g *Game) paintColourOpen(headID uint16, which, x, y, h int) {
	vw := g.vehicleWin(headID)
	if vw == nil || which < 0 || which > 1 {
		return
	}
	d := ui.NewColourDropdown(x, y, h, 0x7FFFFFFF, vw.PaintColours[which], vw.Colour, g.sw, g.sh)
	d.OnSelect = func(c uint8) {
		if w := g.vehicleWin(headID); w != nil {
			w.PaintColours[which] = c
		}
	}
	g.colourDropdown = d
}

// paintToolBegin is toolSet(self, paintBrush, brush): the tool another
// window had is cancelled; this window's own paint tool toggles off.
func (g *Game) paintToolBegin(headID uint16) {
	if g.paintToolOn(headID) {
		g.paintToolAbort()
		return
	}
	g.toolCancel()
	g.paintTool, g.paintHead = true, headID
	g.refreshVehicleWindow(headID, 1)
}

// paintToolAbort is paintToolAbort: the tool off, the button released.
func (g *Game) paintToolAbort() {
	if !g.paintTool {
		return
	}
	g.paintTool = false
	g.refreshVehicleWindow(g.paintHead, -1)
}

// repaint runs vehicleRepaint with the window's colours for every part the
// flags name (setColours), then redraws the window.
func (g *Game) repaint(headID, vehicle uint16, flags uint8) {
	c := g.paintColours(headID)
	a := sim.VehicleRepaint{Vehicle: vehicle, Flags: flags}
	for i, f := range [3]uint8{sim.RepaintBody, sim.RepaintFrontBogie, sim.RepaintBackBogie} {
		if flags&f != 0 {
			a.Colours[i] = c
		}
	}
	g.w.Map.VehicleRepaint(g.playerCmd(), sim.CmdApply, a)
	g.refreshVehicleWindow(headID, -1)
}

// paintEntireTrain is paintEntireTrain: paintFromVehicleUi | applyToEntireTrain on the head.
func (g *Game) paintEntireTrain(headID uint16) {
	g.repaint(headID, headID, sim.RepaintFromVehicleUi|sim.RepaintEntireTrain)
}

// paintToolDown is paintToolDown: the vehicle part under the cursor
// (getItemLeft) when it is this train's — the entire car (paintFromVehicleUi),
// control just that component (^ applyToEntireCar), shift the whole train.
func (g *Game) paintToolDown(sx, sy int) {
	headID := g.paintHead
	if g.vehicleWin(headID) == nil {
		return
	}
	vps, _ := g.viewportsAt(sx, sy)
	it, ok := g.w.ItemLeftAt(vps, sx, sy)
	if !ok || it.Kind != paint.ItemEntity || uint16(it.ID) != headID {
		return
	}
	flags := sim.RepaintFromVehicleUi
	if g.hasKeyModifier(shortcuts.ModControl) {
		flags ^= sim.RepaintEntireCar
	}
	if g.hasKeyModifier(shortcuts.ModShift) {
		flags |= sim.RepaintEntireTrain
	}
	g.repaint(headID, uint16(it.Part), flags)
}

// paintToolDownScroll is paintToolDownScroll: a click on a car in the list
// paints that car (its front bogie); shift the whole train; control (without
// shift) only the body under the cursor (^ applyToEntireCar), or the car
// when no body reaches it.
func (g *Game) paintToolDownScroll(headID uint16, row, scrollX int) {
	vw := g.vehicleWin(headID)
	if vw == nil || row < 0 || row >= len(vw.Cars) || len(vw.Cars[row].Components) == 0 {
		return
	}
	vehicle := vw.Cars[row].Components[0].Front.ID
	flags := sim.RepaintFromVehicleUi
	if g.hasKeyModifier(shortcuts.ModShift) {
		flags |= sim.RepaintEntireTrain
	}
	if g.hasKeyModifier(shortcuts.ModControl) && flags&sim.RepaintEntireTrain == 0 {
		if vw.Cars[row].Object == nil {
			return
		}
		if body, ok := vw.CarBodyAt(row, scrollX); ok {
			vehicle = body
			flags ^= sim.RepaintEntireCar
		}
	}
	g.repaint(headID, vehicle, flags)
}

// cancelVehicleWindowTools is toolSet's toolCancel as another window takes
// the tool (the construction window): a Vehicle window's pickup, route or
// paint tool goes.
func (g *Game) cancelVehicleWindowTools() {
	if g.pickupOn {
		head := g.pickupHead
		g.pickupToolAbort()
		g.pickupOn = false
		g.refreshVehicleWindow(head, 1)
	}
	g.routeTool = false
	g.paintToolAbort()
}
