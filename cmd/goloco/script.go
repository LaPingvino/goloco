package main

import (
	"bufio"
	"fmt"
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
	rightJust        bool
	rightReleaseNext bool
	rightJustRel     bool
	wheelDY          float64
	keysJust         map[ebiten.Key]bool

	shotRequest string
	quit        bool
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
