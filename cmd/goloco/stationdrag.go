package main

import "github.com/LaPingvino/goloco/pkg/sim"

// The Station tab's tool as upstream (StationTab.cpp): the press marks a tile
// (onToolDown: _toolPosInitial, the surface or water under the cursor); held,
// the mouse over another tile makes it a drag (onToolDrag: _toolPosDrag, the
// area shown as the map selection, the ghost gone — onToolUpdate shows none
// while dragging); the release builds — a drag on every tile of the area
// (onToolUpMultiple: constructPieceAtPosition, errors silent, the error window
// closed when anything was built), else the one under the cursor
// (onToolUpSingle).
type stationDrag struct {
	pressed, dragging bool
	from, to          [2]int // tiles
}

// stationToolDown is onToolDown.
func (g *Game) stationToolDown(sx, sy int) {
	g.stDrag = stationDrag{}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return
	}
	t := [2]int{int(px) >> 5, int(py) >> 5}
	g.stDrag = stationDrag{pressed: true, from: t, to: t}
}

// stationToolDrag is onToolDrag (each frame while the press is held).
func (g *Game) stationToolDrag(sx, sy int) {
	if !g.stDrag.pressed {
		return
	}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return
	}
	g.stDrag.to = [2]int{int(px) >> 5, int(py) >> 5}
	g.stDrag.dragging = g.stDrag.to != g.stDrag.from
	if !g.stDrag.dragging {
		g.w.SetMapSelection(false, 0, 0, 0, 0)
		return
	}
	g.removeStationGhost()
	a, b := g.stDrag.from, g.stDrag.to
	g.w.SetMapSelection(true, min(a[0], b[0]), min(a[1], b[1]), max(a[0], b[0]), max(a[1], b[1]))
}

// stationToolRelease is onToolUp: the area (a drag) or the tile under the
// cursor.
func (g *Game) stationToolRelease(sx, sy int) {
	d := g.stDrag
	g.stDrag = stationDrag{}
	if !d.pressed {
		return
	}
	if !d.dragging {
		g.stationToolUp(sx, sy) // onToolUpSingle
		return
	}
	g.w.SetMapSelection(false, 0, 0, 0, 0) // onToolUpMultiple
	g.removeStationGhost()
	dx, dy := 1, 1
	if d.to[0]-d.from[0] <= 0 {
		dx = -1
	}
	if d.to[1]-d.from[1] <= 0 {
		dy = -1
	}
	built := false
	for y := d.from[1]; y != d.to[1]+dy; y += dy {
		for x := d.from[0]; x != d.to[0]+dx; x += dx {
			built = g.constructStationAt(int32(x)*32, int32(y)*32) || built
		}
	}
	if built {
		g.errorWin = nil // WindowManager::close(WindowType::error)
	}
}

// constructStationAt is constructPieceAtPosition: the selected airport, dock,
// road stop or train station at a tile's world position, applied (no error
// window); true when it was built.
func (g *Game) constructStationAt(x, y int32) bool {
	om := g.r.ObjMgr
	if om == nil {
		return false
	}
	m := &g.w.Map
	c, g1 := g.stationCmd()
	do := func(f func(uint8) (int32, bool)) bool {
		_, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, f)
		return ok
	}
	switch {
	case g.buildMode == buildModeAirport:
		list := om.AvailableAirports(int(g.w.CurrentYear))
		if len(list) == 0 {
			return false
		}
		if a, ok := m.AirportPlacementAt(x, y, uint8(pickSlot(list, g.consStation)), g.airportRotation, om); ok {
			return do(func(f uint8) (int32, bool) { return m.CreateAirport(c, f, a, g1) })
		}
	case g.buildMode == buildModeDock:
		list := om.AvailableDocks(int(g.w.CurrentYear))
		if len(list) == 0 {
			return false
		}
		if a, ok := m.DockPlacementAt(x, y, uint8(pickSlot(list, g.consStation)), om); ok {
			return do(func(f uint8) (int32, bool) { return m.CreatePort(c, f, a, g1) })
		}
	case g.consIsRoad():
		list := om.AvailableCompatibleStations(g.buildObjID, true, int(g.w.CurrentYear))
		if len(list) == 0 {
			return false
		}
		if a, ok := g.w.RoadStationPlacementAtTile(x, y, uint8(pickSlot(list, g.consStation))); ok {
			return do(func(f uint8) (int32, bool) { return m.CreateRoadStation(c, f, a, g1) })
		}
	default:
		list := om.AvailableCompatibleStations(g.buildObjID, false, int(g.w.CurrentYear))
		if len(list) == 0 {
			return false
		}
		if a, ok := g.w.TrainStationPlacementAtTile(x, y, uint8(pickSlot(list, g.consStation))); ok {
			return do(func(f uint8) (int32, bool) { return m.CreateTrainStation(c, f, a, g1) })
		}
	}
	return false
}
