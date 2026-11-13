// Command setview writes a copy of a saved game that opens on another view:
// its GameState savedViewX / Y / Zoom / Rotation set to a tile, a zoom and a
// rotation. Both OpenLoco and goloco open a save on its saved view
// (Window::viewportFromSavedView), so the copy shows the same place, at the
// same rotation, in both — for scripts/olref comparisons.
//
//	go run ./scripts/olref/setview -tile 196,190 -rotation 1 -zoom 0 in.SV5 out.SV5
//
// The GameState chunk is rewritten uncompressed (SawyerEncoding 0) and the
// file's checksum (the sum of its bytes) redone. The copy is game data: keep
// it out of git.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

func main() {
	tile := flag.String("tile", "", "tile x,y the view centres on (default: the saved view kept)")
	rotation := flag.Int("rotation", -1, "view rotation 0-3 (default: kept)")
	zoom := flag.Int("zoom", -1, "zoom 0-3 (default: kept)")
	flag.Parse()
	if flag.NArg() != 2 {
		log.Fatal("usage: setview [-tile x,y] [-rotation r] [-zoom z] in.SV5 out.SV5")
	}
	data, err := os.ReadFile(flag.Arg(0))
	must(err)
	if len(data) < 4 {
		log.Fatal("setview: not an S5 file")
	}
	body := data[:len(data)-4] // the checksum follows

	// The chunks before the GameState: header, save details, packed objects
	// (a raw 16-byte ObjectHeader before each), required objects.
	r := assets.NewS5ChunkReader(body)
	hb, err := r.ReadChunk()
	must(err)
	h, err := assets.ParseS5Header(hb)
	must(err)
	if h.Type == assets.S5TypeScenario || h.Type == assets.S5TypeLandscape {
		log.Fatal("setview: a saved game, not a scenario (its state is in several chunks)")
	}
	if h.Flags&assets.HeaderFlagHasSaveDetails != 0 {
		_, err = r.ReadChunk()
		must(err)
	}
	for i := uint16(0); i < h.NumPackedObjects; i++ {
		r.AdvanceRaw(16)
		_, err = r.ReadChunk()
		must(err)
	}
	_, err = r.ReadChunk()
	must(err)
	start := r.Offset()
	gs, err := r.ReadChunk()
	must(err)
	end := r.Offset()

	rot := gs[0x23] & 3
	if *rotation >= 0 {
		rot = uint8(*rotation) & 3
	}
	if *zoom >= 0 {
		gs[0x22] = uint8(*zoom)
	}
	gs[0x23] = rot
	if *tile != "" {
		var tx, ty int32
		if _, err := fmt.Sscanf(*tile, "%d,%d", &tx, &ty); err != nil {
			log.Fatalf("setview: -tile: %v", err)
		}
		sc, err := scenario.LoadScenarioData(flag.Arg(0))
		must(err)
		var m sim.Map
		m.LoadScenario(sc)
		land, _ := m.GetHeight(tx*32+16, ty*32+16)
		vx, vy := sim.GameToScreen(tx*32+16, ty*32+16, int32(land), rot) // Viewport::centre2dCoordinates' centre
		binary.LittleEndian.PutUint16(gs[0x1E:], uint16(int16(vx)))
		binary.LittleEndian.PutUint16(gs[0x20:], uint16(int16(vy)))
	} else if *rotation >= 0 {
		log.Print("setview: the rotation alone — the saved centre is a viewport position, so the view shows another place")
	}

	out := append([]byte{}, body[:start]...)
	out = append(out, byte(assets.EncodingUncompressed))
	out = binary.LittleEndian.AppendUint32(out, uint32(len(gs)))
	out = append(out, gs...)
	out = append(out, body[end:]...)
	var sum uint32
	for _, b := range out {
		sum += uint32(b)
	}
	out = binary.LittleEndian.AppendUint32(out, sum)
	must(os.WriteFile(flag.Arg(1), out, 0o600))
	fmt.Printf("view %d,%d zoom %d rotation %d → %s\n", int16(binary.LittleEndian.Uint16(gs[0x1E:])), int16(binary.LittleEndian.Uint16(gs[0x20:])), gs[0x22], gs[0x23], flag.Arg(1))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
