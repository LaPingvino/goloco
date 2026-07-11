# GoLoco — agent handbook

Go reimplementation of Chris Sawyer's Locomotion. Most code was ported by AI
from OpenLoco (a C++ decompilation); porting bugs are common and usually
SYSTEMIC — the same few root causes explain most defects.

## Ground truth

- **Upstream reference**: `~/OpenLoco` (shallow clone). When goloco code or
  comments disagree with upstream, trust upstream — comments here may quote a
  stale or misread decompilation. Game logic: `~/OpenLoco/src/OpenLoco/src/`.
- **Game data**: `~/.local/share/Steam/steamapps/common/Locomotion`
  (`Data/g1.DAT`, `ObjData/*.DAT`, `Scenarios/*.SC5`). Without it the game
  falls back to placeholder rendering.
- Debug heuristic (from Joop): when stuck, fix every detail that looks off
  along the way — small oddities are evidence about the systemic cause.

## Recurring bug classes — check these FIRST

1. **Sprite anchor offsets.** Upstream `drawImage` always applies each G1
   element's baked x/y offsets. Any draw that ignores `GetSpriteInfo` offsets
   is misplaced (caused: loading-bar rails on top, bridge-deck band, vehicle
   drift). Never add ad-hoc `-16`-style fudges; find the missing offset/unit.
2. **Steam g1.DAT realignment.** `pkg/assets/g1.go LoadG1` shifts indices
   ≥3549 up by 2 and copies a font block to 3898 (upstream Gfx.cpp). All
   hard-coded sprite ids in this repo are DISC-layout ids (match upstream
   ImageIds.h).
3. **Units.** WorldUnits: 1 tile = 32 (int16). Heights: `baseZ` is SmallZ
   (×4 px at zoom 0); entity `PosZ` and piece z-offsets are WORLD z (int16,
   ×1 px); MicroZ = SmallZ/4 (slope corners). Mixing these caused floating
   trains and misplaced decks.
4. **DAT section walks.** Object loaders must mirror the upstream
   `Object::load` byte-for-byte: string tables (some objects have TWO —
   name + description), conditional headers, variable-length cargo lists.
   A skipped section corrupts every image table after it. Image-id fields in
   the DAT are placeholders — recompute like upstream (see vehicle.go).
5. **RLE formats.** Three distinct codecs, all with correct impls in
   `pkg/assets`: G1 sprite RLE (`decodeRLEElement`, line-offset chunks),
   Sawyer run-length single, and run-length multi (0xFF literal +
   back-reference). Never invent a decoder; reuse these.
6. **Remap palettes.** Company-coloured sprites need `GetSpriteColoured`
   (one colour) or `GetSpriteColoured2` (recolour2, e.g. loading train);
   raw draws show magenta/green remap ranges. Text glyphs colour ONLY
   palette index 0x01 (textRemap0); 0x02/0x03 are inset/outline styles.

## Verification loop (vision-based)

Build: `go build -o goloco ./cmd/goloco`. Always verify visually:

- `./goloco <scenario.SC5> --shots S,N` — save shot_NN.png every S seconds.
- `./goloco [scenario] --diag [TX,TY|water|track|station|train|industry|worst]`
  — 480×360 diag.png centred on the target, then quit.
- `GOLOCO_SCRIPT=file.txt ./goloco …` — scripted input playback
  (cmd/goloco/script.go: wait/move/click/rightclick/key/wheel/shot/quit),
  like OpenLoco tutorial playback. CAUTION: scripted input ORs with real
  hardware input — coordinate with the user before driving the visible
  window, their clicks will interleave.
- `GOLOCO_OPEN=newgame|track` — open a window / chain a test track at start.
- `GOLOCO_DEBUG_SPRITES=1` — per-object sprite ranges + decode stack traces.
- `GOLOCO_NO_SMOOTH/NO_TRACK/NO_STATION=1` — render-pass kill switches for
  visual bisection (find which pass draws an artefact).
- `GOLOCO_FONT=ttf` — force TTF text (default: authentic g1 bitmap glyphs).
- Sprite inspection: `go run ./cmd/extract_g1 -start N -end M [-info-start
  N -info-end M]`; read the PNGs with vision. `scripts/fetch-fonts.sh` gets
  the OpenTTD TTFs (Unicode fallback).

- ALWAYS run long headless sessions under Xvfb so the desktop/screen-lock
  cannot freeze the loop: `Xvfb :99 -screen 0 800x600x24 &` then
  `DISPLAY=:99 WAYLAND_DISPLAY= ./goloco …` (xorg-server-xvfb installed).
- `GOLOCO_OPEN=win` — autonomous scenario attempt: builds a shuttle beside
  a town (World.FindTownsideRun), stations, vehicle; the objective counter
  runs to the win banner. First victory: Weatherworld, 648/500 PASS.

Read captures with vision, judge against real Locomotion, fix, re-capture.
The scenario boot takes ~40-80s wall (software rendering ~10-20 fps).

## Data tables

Generated tables (e.g. `pkg/world/track_data_gen.go`: kTrackCoordinates,
kTrackPieceTiles) are extracted MECHANICALLY from upstream TrackData.cpp with
a small python script — never hand-transcribe tables; regenerate the same way
(see the commit that added the file for the script pattern).

## Current state / roadmap

Working: terrain (+edge smoothing, per-tile variation), water, roads/track
rendering incl. junctions, stations, signals, level crossings, bridges
(decks only), trees, buildings, vehicles rendered from save data, title
screen (logo/menu/music), loading screen, bitmap fonts, New Game browser,
RCT-style track construction (head + connected pieces, curve/slope window).

Next (task list): road construction via the same head system (extract
kRoadCoordinates/roadPiece tables), track/road TYPE picker in the
construction window (list ObjMgr.TrackObjects/RoadObjects by DisplayName),
station/signal tabs, vehicle purchase/placement, airports/docks, full
PaintBridge port (walls/supports/quarter decks), IndustryObject fields,
gameplay simulation (movement, economy).

## Conventions

- Commit style: `fix:`/`feat:`/`chore:` + why + OpenLoco reference paths.
  Push to origin main (SSH key configured).
- Comments cite upstream as `OpenLoco reference: src/.../File.cpp`.
- Text drawing: y is the TOP of 10px bitmap glyphs; centre in a box with
  `y + (h-10)/2`. Measure with ui.MeasureText.
- `go test ./pkg/...` must stay green; golden-vector tests preferred (see
  pkg/core/prng_test.go).
