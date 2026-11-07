package main

import (
	"bufio"
	"fmt"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
//	click <x> <y>       move + left press (release follows next frame)
//	rightclick <x> <y>  move + right press
//	press <x> <y>       move + left press, held until release
//	release <x> <y>     move + left release (after press)
//	drag <x1> <y1> <x2> <y2>  press, five moves, release (a window by its title)
//	key <name>          tap a key: R X ESC Q E ENTER SPACE ...
//	wheel <dy>          mouse wheel step at current cursor
//	shot <file.png>     save a full-frame screenshot
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
	rightReleaseNext bool
	rightJustRel     bool
	wheelDY          float64
	keysJust         map[ebiten.Key]bool

	shotRequest string
	// cameraReq: "camera X Y" — the main view centred on tile (X, Y), so a
	// playthrough starts where it means to whatever the scenario's saved view
	// or the time passed (scripted runs only).
	cameraReq *[2]float64
	// scenarioReq: "scenario NAME" — that scenario started (as the New Game
	// list's Start does), by name rather than by where its row happens to be.
	scenarioReq string
	// seed: "seed N" — the time loadAndStart mixes into the new game's rng,
	// fixed (a generated landscape, the company … the same every run).
	seed *uint32
	quit bool
}

var scriptKeyNames = map[string]ebiten.Key{
	"R": ebiten.KeyR, "X": ebiten.KeyX, "Q": ebiten.KeyQ, "E": ebiten.KeyE,
	"ESC": ebiten.KeyEscape, "ESCAPE": ebiten.KeyEscape,
	"ENTER": ebiten.KeyEnter, "SPACE": ebiten.KeySpace,
	"W": ebiten.KeyW, "A": ebiten.KeyA, "S": ebiten.KeyS, "D": ebiten.KeyD,
	"F12": ebiten.KeyF12,
}

func loadInputScript(path string) *inputScript {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("[Script] cannot open %s: %v", path, err)
		return nil
	}
	defer f.Close()
	s := &inputScript{keysJust: map[ebiten.Key]bool{}}
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
	// shotRequest is NOT cleared here: multiple Updates can run between two
	// Draws at low FPS; only scriptShotName (called from Draw) consumes it.
	for k := range s.keysJust {
		delete(s.keysJust, k)
	}

	if time.Now().Before(s.waitUntil) {
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
		s.waitUntil = time.Now().Add(time.Duration(sec * float64(time.Second)))
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
	case "key":
		if k, ok := scriptKeyNames[strings.ToUpper(fields[1])]; ok {
			s.keysJust[k] = true
		} else {
			log.Printf("[Script] unknown key %q", fields[1])
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
	case "scenario":
		s.scenarioReq = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "scenario"))
	case "camera":
		if len(fields) >= 3 {
			x, _ := strconv.ParseFloat(fields[1], 64)
			y, _ := strconv.ParseFloat(fields[2], 64)
			s.cameraReq = &[2]float64{x, y}
		}
	case "shot":
		s.shotRequest = fields[1]
		s.waitUntil = time.Now().Add(300 * time.Millisecond) // let Draw consume it
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
	if g.script != nil && btn == ebiten.MouseButtonLeft && g.script.leftHeld {
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

func (g *Game) inKeyJustPressed(k ebiten.Key) bool {
	if g.script != nil && g.script.keysJust[k] {
		return true
	}
	return inpututil.IsKeyJustPressed(k)
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

var errScriptQuit = fmt.Errorf("script quit")

// scriptCamera applies a script's "camera X Y" (once).
func (g *Game) scriptCamera() {
	if g.script == nil || g.script.cameraReq == nil || g.w == nil {
		return
	}
	c := g.script.cameraReq
	g.script.cameraReq = nil
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
			g.scenarioSelect = nil
			if err := g.loadScenario(m.FilePath); err != nil {
				log.Printf("[Script] scenario %q: %v", name, err)
			}
			return
		}
	}
	log.Printf("[Script] no scenario named %q", name)
}
