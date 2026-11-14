package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Scripted input playback: GOLOCO_SCRIPT=file.txt drives the game through the
// same input paths as a human, so interactions can be exercised and verified
// headlessly. Conceptually the playback half of OpenLoco's tutorial system
// (recorded input streams replayed into the game loop).
//
// Script commands, one per line ('#' comments):
//
//	wait <seconds>      pause before the next command (game keeps running)
//	move <x> <y>        move the cursor
//	click <x> <y>       move + left press (release follows next frame, with
//	                    the next command: put a wait after a click whose
//	                    release must land where it pressed, e.g. a caption)
//	rightclick <x> <y>  move + right press (release follows next frame)
//	rightpress <x> <y>  move + right press, held until rightrelease
//	rightrelease <x> <y>  move + right release (after rightpress)
//	press <x> <y>       move + left press, held until release
//	release <x> <y>     move + left release (after press)
//	drag <x1> <y1> <x2> <y2>  press, five moves, release (a window by its title)
//	key <binding>       tap a key, as the config file names a shortcut's: Z, 1, PageUp, Space,
//	                    Left Ctrl+S, Left Shift+Backspace (SDL's key names, any case; ESC and
//	                    ENTER are Escape and Return)
//	keydown <key>       hold a key (the arrow keys' scrolling, Insert for a cheat word) until
//	keyup <key>         lets it go
//	type <text>         type the rest of the line: each character its key pressed (into a
//	                    text input; a cheat word while Insert is held)
//	wheel <dy>          mouse wheel step at current cursor
//	shot <file.png>     save a full-frame screenshot
//	expect <key> <op> <n>  check the game: key cash, stations, vehicles, placed,
//	                    delivered, rotation or townwindows (the controlling company's cash, station
//	                    and vehicle counts, its vehicles on the map, cargo units delivered — this month
//	                    and the kept history; the view's rotation 0–3; 1 while a Town window is open; maptip: 1
//	                    while the map tooltip shows, its text logged; conswindow, consbuilding, constab: the
//	                    construction window open, building on from a position, its tab; camx, camy: the view's
//	                    position; townlist, stationlist, vehiclelist, industrylist, companywindow: 1 while one
//	                    is open; terraformwindow, terraformtab, paused, speed, zoom, constype, conspiece,
//	                    consgradient, consrotation, drivercheat, jailed, windows, textinput (script.go
//	                    scriptExpect says what each is); window:NAME: 1 while a window of that type is open; viewflags: the main viewport's flags (ViewportFlags' bits); maptip~TEXT: 1 while the map tooltip shows text holding TEXT, _ for a space),
//	                    op one of
//	                    > >= == <= <; a failed check makes the run exit 1
//	news <type> [a [b [c]]] [date <day>]  post a message of that MessageType
//	                    with those subjects for the player's company (first
//	                    moving the game's day): news shots on demand
//	quit                exit the game
type inputScript struct {
	cmds []string
	idx  int

	waitUntil time.Time // wall-clock deadline before the next command runs

	// synthesized per-frame state
	mouseX, mouseY   int
	hasMouse         bool
	leftJust         bool
	leftReleaseNext  bool
	leftJustReleased bool
	leftHeld         bool // between press and release
	rightJust        bool
	rightHeld        bool // between rightpress and rightrelease
	rightReleaseNext bool
	rightJustRel     bool
	wheelDY          float64
	keyTaps          []keyEvent      // "key", "type": this frame's key presses (keyboard.go readKeys)
	tapHeld          map[uint32]bool // the modifier keys held with this frame's "key"
	keysHeld         map[uint32]bool // "keydown" … "keyup"

	shotRequest string
	// cameraReq: "camera X Y [ZOOM [ROTATION]]" — the main view centred on
	// tile (X, Y) at that zoom and rotation (0 unless given), so a playthrough
	// starts where it means to whatever the scenario's saved view (its place,
	// zoom and rotation) or the time passed (scripted runs only).
	cameraReq *[4]float64
	// newsReq: "news TYPE …" — a message posted (scriptNews).
	newsReq []string
	// scenarioReq: "scenario NAME" — that scenario started (as the New Game
	// list's Start does), by name rather than by where its row happens to be.
	scenarioReq string
	// seed: "seed N" — the time loadAndStart mixes into the new game's rng,
	// fixed (a generated landscape, the company … the same every run).
	seed *uint32
	// expectReq: "expect KEY OP N" — a check of the game's state, run by the
	// game (scriptExpect); failed: a failed check, the run exits non-zero.
	expectReq []string
	failed    bool
	quit      bool
}

func loadInputScript(path string) *inputScript {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[Script] cannot open %s: %v", path, err)
		return nil
	}
	defer f.Close()
	s := &inputScript{tapHeld: map[uint32]bool{}, keysHeld: map[uint32]bool{}}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s.cmds = append(s.cmds, line)
	}
	log.Printf("[Script] loaded %d commands from %s", len(s.cmds), path)
	return s
}

// step advances the script by one Update tick, synthesizing this frame's
// input state.
func (s *inputScript) step() {
	// releases trail presses by one frame
	s.leftJustReleased = s.leftReleaseNext
	s.leftReleaseNext = false
	s.rightJustRel = s.rightReleaseNext
	s.rightReleaseNext = false
	s.leftJust = false
	s.rightJust = false
	s.wheelDY = 0
	s.keyTaps = nil
	// shotRequest is NOT cleared here: multiple Updates can run between two
	// Draws at low FPS; only scriptShotName (called from Draw) consumes it.
	clear(s.tapHeld)

	if gameNow().Before(s.waitUntil) {
		return
	}
	if s.idx >= len(s.cmds) {
		return
	}
	line := s.cmds[s.idx]
	s.idx++
	fields := strings.Fields(line)
	arg := func(i int) int { v, _ := strconv.Atoi(fields[i]); return v }
	switch strings.ToLower(fields[0]) {
	case "wait":
		sec, _ := strconv.ParseFloat(fields[1], 64)
		s.waitUntil = gameNow().Add(time.Duration(sec * float64(time.Second)))
	case "move":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
	case "click":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.leftJust = true
		s.leftReleaseNext = true
	case "press":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.leftJust, s.leftHeld = true, true
	case "release":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.leftHeld = false
		s.leftReleaseNext = true
	case "drag": // expands in place: press, moves a frame apart, release
		x1, y1, x2, y2 := arg(1), arg(2), arg(3), arg(4)
		steps := []string{fmt.Sprintf("press %d %d", x1, y1)}
		for i := 1; i <= 5; i++ {
			steps = append(steps, fmt.Sprintf("move %d %d", x1+(x2-x1)*i/5, y1+(y2-y1)*i/5))
		}
		steps = append(steps, fmt.Sprintf("release %d %d", x2, y2))
		s.cmds = append(s.cmds[:s.idx], append(steps, s.cmds[s.idx:]...)...)
	case "rightclick":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.rightJust = true
		s.rightReleaseNext = true
	case "rightpress":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.rightJust, s.rightHeld = true, true
	case "rightrelease":
		s.mouseX, s.mouseY, s.hasMouse = arg(1), arg(2), true
		s.rightHeld = false
		s.rightReleaseNext = true
	case "type":
		for _, r := range strings.TrimPrefix(strings.TrimSpace(line), fields[0]+" ") {
			s.keyTaps = append(s.keyTaps, keyEvent{keyOfChar(r), r})
		}
	case "key", "keydown", "keyup":
		name, _, _ := strings.Cut(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), fields[0])), " #")
		ev, held, ok := scriptKey(strings.TrimSpace(name))
		switch {
		case !ok:
			log.Printf("[Script] unknown key %q", name)
			s.failed = true
		case strings.EqualFold(fields[0], "keydown"):
			s.keysHeld[ev.keyCode] = true
			s.keyTaps = append(s.keyTaps, ev) // the press itself
		case strings.EqualFold(fields[0], "keyup"):
			delete(s.keysHeld, ev.keyCode)
		default:
			s.keyTaps = append(s.keyTaps, ev)
			for _, h := range held {
				s.tapHeld[h] = true
			}
		}
	case "wheel":
		dy, _ := strconv.ParseFloat(fields[1], 64)
		s.wheelDY = dy
	case "seed":
		if len(fields) >= 2 {
			if n, err := strconv.ParseUint(fields[1], 10, 32); err == nil {
				v := uint32(n)
				s.seed = &v
			}
		}
	case "expect":
		s.expectReq = fields[1:]
	case "scenario":
		s.scenarioReq = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "scenario"))
	case "camera":
		if len(fields) >= 3 {
			var c [4]float64 // x, y, then the optional zoom and rotation (0)
			for i := range c {
				if i+1 < len(fields) {
					c[i], _ = strconv.ParseFloat(fields[i+1], 64)
				}
			}
			s.cameraReq = &c
		}
	case "news": // goloco only: post a message — news TYPE [SUBJECT …] [date DAY]
		s.newsReq = fields[1:]
	case "shot":
		s.shotRequest = fields[1]
		s.waitUntil = gameNow().Add(300 * time.Millisecond) // let Draw consume it
	case "quit":
		s.quit = true
	default:
		log.Printf("[Script] unknown command %q", line)
	}
	if fields[0] != "wait" {
		log.Printf("[Script] %s", line)
	}
}

// --- Game input accessors: script overlays real hardware input ------------

func (g *Game) inMouseJustPressed(btn ebiten.MouseButton) bool {
	if g.script != nil {
		switch btn {
		case ebiten.MouseButtonLeft:
			if g.script.leftJust {
				return true
			}
		case ebiten.MouseButtonRight:
			if g.script.rightJust {
				return true
			}
		}
	}
	return inpututil.IsMouseButtonJustPressed(btn)
}

// inMousePressed is ebiten.IsMouseButtonPressed with a script's held press.
func (g *Game) inMousePressed(btn ebiten.MouseButton) bool {
	if g.script != nil && (btn == ebiten.MouseButtonLeft && g.script.leftHeld || btn == ebiten.MouseButtonRight && g.script.rightHeld) {
		return true
	}
	return ebiten.IsMouseButtonPressed(btn)
}

func (g *Game) inMouseJustReleased(btn ebiten.MouseButton) bool {
	if g.script != nil {
		switch btn {
		case ebiten.MouseButtonLeft:
			if g.script.leftJustReleased {
				return true
			}
		case ebiten.MouseButtonRight:
			if g.script.rightJustRel {
				return true
			}
		}
	}
	return inpututil.IsMouseButtonJustReleased(btn)
}

func (g *Game) inWheel() (float64, float64) {
	dx, dy := ebiten.Wheel()
	if g.script != nil && g.script.wheelDY != 0 {
		dy = g.script.wheelDY
	}
	return dx, dy
}

// scriptShotName returns a pending screenshot request (and clears it).
func (g *Game) scriptShotName() string {
	if g.script == nil {
		return ""
	}
	n := g.script.shotRequest
	g.script.shotRequest = ""
	return n
}

// scriptNews applies a script's "news TYPE [A [B [C]]] [date DAY]" (once):
// MessageManager::post of that message type for the player's company with
// those subjects (0xFFFF when left out) — "date DAY" moves the game's day
// first (the newspaper's style goes by the message's year). For the shots
// of news a played game doesn't bring up on demand.
func (g *Game) scriptNews() {
	if g.script == nil || g.script.newsReq == nil {
		return
	}
	req := g.script.newsReq
	g.script.newsReq = nil
	m := &g.w.Map
	vals := []uint16{0, 0xFFFF, 0xFFFF, 0xFFFF}
	n := 0
	for i := 0; i < len(req); i++ {
		if req[i] == "date" && i+1 < len(req) {
			if d, err := strconv.ParseUint(req[i+1], 10, 32); err == nil {
				m.CurrentDay = uint32(d)
			}
			i++
			continue
		}
		if v, err := strconv.ParseUint(req[i], 10, 16); err == nil && n < len(vals) {
			vals[n] = uint16(v)
			n++
		}
	}
	m.PostMessage(g.r.ObjMgr, sim.MessageType(vals[0]), m.ControllingCompany, vals[1], vals[2], vals[3])
}

// scriptCamera applies a script's "camera X Y [ZOOM [ROTATION]]" (once).
func (g *Game) scriptCamera() {
	if g.script == nil || g.script.cameraReq == nil {
		return
	}
	c := g.script.cameraReq
	g.script.cameraReq = nil
	g.w.SetZoom(int(c[2]))
	if rot := uint8(c[3]) & 3; rot != g.w.Rotation() {
		g.w.SetRotation(rot)
	}
	g.w.SetCamera(c[0], c[1])
}

// scriptScenario starts a script's "scenario NAME" (once): the scenario of
// that name in the game's Scenarios folder, as the New Game list starts it.
func (g *Game) scriptScenario() {
	if g.script == nil || g.script.scenarioReq == "" {
		return
	}
	name := g.script.scenarioReq
	g.script.scenarioReq = ""
	metas, _ := scenario.ScanScenarios(g.dataDir + "/../Scenarios")
	if len(metas) == 0 {
		metas, _ = scenario.ScanScenarios(g.dataDir)
	}
	for _, m := range metas {
		if strings.EqualFold(m.Name, name) {
			g.wins.CloseType(ui.WindowTypeScenarioSelect, ui.AnyNumber)
			if err := g.loadScenario(m.FilePath); err != nil {
				log.Printf("[Script] scenario %q: %v", name, err)
			}
			return
		}
	}
	log.Printf("[Script] no scenario named %q", name)
}

// scriptExpect runs a script's "expect KEY OP N" against the controlling
// company: logged either way, a failure marks the run failed.
func (g *Game) scriptExpect() {
	if g.script == nil || g.script.expectReq == nil {
		return
	}
	req := g.script.expectReq
	g.script.expectReq = nil
	fail := func(why string) {
		log.Printf("[Script] EXPECT FAILED: %s — %s", strings.Join(req, " "), why)
		g.script.failed = true
	}
	if len(req) != 3 {
		fail("want: expect KEY OP N")
		return
	}
	want, err := strconv.ParseInt(req[2], 10, 64)
	if err != nil {
		fail(err.Error())
		return
	}
	company, _ := g.w.ControllingCompany()
	var got int64
	switch req[0] {
	case "cash":
		got, _ = g.w.ControllingCompanyCash()
	case "stations":
		for _, s := range g.w.StationList() {
			if s.Owner == company && s.Flags&(1<<5) == 0 {
				got++
			}
		}
	case "vehicles":
		for _, v := range g.w.VehicleListRows() {
			if v.Owner == company {
				got++
			}
		}
	case "placed": // the company's vehicles on the map (head tileX != -1)
		if trains, err := g.w.Map.Trains(); err == nil {
			for _, tr := range trains {
				if h := &g.w.Map.Entities[tr.Head]; int(h.Owner) == company && h.TileX != -1 {
					got++
				}
			}
		}
	case "townwindows": // 1 while a Town window is open
		if g.wins.Find(ui.WindowTypeTown, ui.AnyNumber) != nil {
			got = 1
		}
	case "rotation": // the view's rotation (WindowManager::getCurrentRotation)
		got = int64(g.w.Rotation())
	case "delivered": // this month's (cargoUnitsTotalDelivered, reset monthly) and the history's
		if company < len(g.w.Map.Companies) {
			c := &g.w.Map.Companies[company]
			got = int64(c.CargoUnitsTotalDelivered)
			for _, n := range c.CargoUnitsDeliveredHist {
				got += int64(n)
			}
		}
	case "maptip": // 1 while the map tooltip shows (its text logged)
		if t := &g.mapTip; t.open && t.tip.Text != "" {
			got = 1
			log.Printf("[Script] map tooltip: %q", strings.NewReplacer("\x05", " / ", string(ui.MapTipColour3), "").Replace(t.tip.Text))
		}
	case "conswindow": // 1 while the construction window is open
		if g.buildMode != buildModeNone && g.consWin() != nil {
			got = 1
		}
	case "consbuilding": // 1 while the construction window builds on from a position (not hovering)
		if g.consHasHead() && g.consWin() != nil && !g.w.Cons.Hover {
			got = 1
		}
	case "constab": // the construction window's tab (0 construction, 1 station, 2 signal, 3 overhead)
		got = int64(g.consTab)
	case "camx", "camy": // the main view's position (viewport pixels)
		cx, cy := g.w.CameraPos()
		got = int64(cx)
		if req[0] == "camy" {
			got = int64(cy)
		}
	case "townlist", "stationlist", "vehiclelist", "industrylist", "companywindow": // 1 while such a window is open
		typ := map[string]ui.WindowType{"townlist": ui.WindowTypeTownList, "stationlist": ui.WindowTypeStationList,
			"vehiclelist": ui.WindowTypeVehicleList, "industrylist": ui.WindowTypeIndustryList, "companywindow": ui.WindowTypeCompany}[req[0]]
		if g.wins.Find(typ, ui.AnyNumber) != nil {
			got = 1
		}
	case "viewflags": // the main viewport's flags (Ui::ViewportFlags' bits)
		got = int64(g.w.ViewFlags())
	case "terraformwindow": // 1 while the Terraform window is open
		if g.terraformWin() != nil {
			got = 1
		}
	case "terraformtab": // the Terraform window's tab (0 clear, 1 land, 2 water, 3 trees, 4 walls; -1 closed)
		got = -1
		if w := g.terraformWin(); w != nil {
			got = int64(w.Tab)
		}
	case "paused": // 1 while the game is paused
		if g.isPaused {
			got = 1
		}
	case "speed": // the game speed's multiplier (1, 2, 4)
		got = int64(g.speedMult)
	case "zoom": // the main view's zoom (0 closest … 3)
		got = int64(g.w.GetZoom())
	case "constype": // the construction window's track type (a road's with bit 7; -1: no track or road window)
		got = -1
		if t, ok := g.consCurrentTrackType(); ok && g.consHasHead() {
			got = int64(t)
		}
	case "conspiece", "consgradient": // the construction state's selected piece / gradient
		got = int64(g.w.Cons.Piece)
		if req[0] == "consgradient" {
			got = int64(g.w.Cons.Gradient)
		}
	case "consrotation": // the construction state's rotation
		got = int64(g.w.Cons.Rotation)
	case "drivercheat": // 1 once the DRIVER cheat is on
		if g.driverCheat {
			got = 1
		}
	case "jailed": // the controlling company's days in jail (the FREECASH cheat: 30)
		if company < len(g.w.Map.Companies) {
			got = int64(g.w.Map.Companies[company].JailStatus)
		}
	case "windows": // how many windows are open that a close-all closes (not the news or other sticky ones)
		for _, w := range g.wins.Windows {
			if !w.Base().HasFlags(ui.WindowFlagStickToBack | ui.WindowFlagStickToFront) {
				got++
			}
		}
	case "textinput": // 1 while a text input (a rename) is open
		if g.textInput != nil {
			got = 1
		}
	default:
		if name, ok := strings.CutPrefix(req[0], "window:"); ok { // 1 while a window of that WindowType is open
			typ, known := ui.WindowTypeNames[name]
			if !known {
				fail("no window type " + name)
				return
			}
			if g.wins.Find(typ, ui.AnyNumber) != nil {
				got = 1
			}
			break
		}
		if sub, ok := strings.CutPrefix(req[0], "maptip~"); ok { // 1 while the map tooltip shows text holding this (_ for a space)
			if t := &g.mapTip; t.open && strings.Contains(t.tip.Text, strings.ReplaceAll(sub, "_", " ")) {
				got = 1
			}
			break
		}
		// KEY@X,Y: a tile's surface height (baseZ, in SmallZ), its slope bits,
		// its water level, its trees and walls (ghosts not counted).
		key, at, _ := strings.Cut(req[0], "@")
		var tx, ty int
		t := (*sim.Tile)(nil)
		if n, _ := fmt.Sscanf(at, "%d,%d", &tx, &ty); n == 2 {
			t = g.w.Map.TileAt(tx, ty)
		}
		if t == nil {
			fail("unknown key")
			return
		}
		count := func(kind uint8) (n int64) {
			for _, r := range t.Order {
				if r.Kind == kind && r.Flags&(1<<4) == 0 {
					n++
				}
			}
			return n
		}
		switch key {
		case "height":
			got = int64(t.BaseZ)
		case "slope":
			got = int64(t.Slope & 0x1F)
		case "water":
			got = int64(t.WaterLevel)
		case "trees":
			got = count(scenario.KindTree)
		case "walls":
			got = count(scenario.KindWall)
		default:
			fail("unknown key")
			return
		}
	}
	ok := map[string]bool{">": got > want, ">=": got >= want, "==": got == want, "<=": got <= want, "<": got < want}
	res, known := ok[req[1]]
	switch {
	case !known:
		fail("unknown op")
	case !res:
		fail(fmt.Sprintf("it is %d", got))
	default:
		log.Printf("[Script] expect %s: ok (%d)", strings.Join(req, " "), got)
	}
}
