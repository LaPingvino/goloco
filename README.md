# GoLoco — OpenLoco Reimplementation in Go

A from-scratch reimplementation of Chris Sawyer's Locomotion in Go, using
[Ebitengine](https://ebitengine.org/) for rendering. Loads the original game's
data files (G1 sprites, DAT objects, SC5/SV5 scenarios).

## What Works

- Loads Locomotion game data: G1 sprites, object DAT files, packed objects from SC5/SV5 scenarios
- Isometric terrain rendering with height, cliffs, water, and growth-stage surfaces
- Track and road rendering, including junctions (ballast/sleeper/rail layering)
- Stations (train, bus stop), bridges, train signals, level crossings
- Title sequence with scripted camera cuts
- Window/UI system (drag, buttons, checkboxes) with 3D beveled widgets
- Audio/music playback
- Camera pan and zoom; O(visible) tile culling with a sprite atlas

## Not There Yet

- Gameplay systems: economy, companies, vehicle movement, routing
- Full paint/depth-sort parity with OpenLoco

## Quick Start

```bash
./build.sh        # or: go build -o goloco ./cmd/goloco
./goloco
```

Requires Locomotion game data in `../locomotion/Data` (falls back to
placeholders without it).

## Controls

- **Arrow keys / WASD:** pan camera
- **Q / E** or **mouse wheel:** zoom
- **Mouse:** drag windows by title bar, click widgets
- **Esc:** quit

## Layout

```
cmd/
├── goloco/        # Game entry point
├── diagnostics/   # Implementation-coverage report vs OpenLoco source
├── extract_g1/    # Dump G1 sprites to PNG
└── list_objects/  # List objects in game data
pkg/
├── assets/        # G1/DAT loading
├── audio/         # Audio/music
├── gfx, graphics/ # Drawing primitives
├── loco/          # Type definitions (economy, objects, s5, vehicles, worldmap)
├── objects/       # Game object parsing
├── render/        # Ebiten adapter, sprite atlas
├── scenario/      # S5 scenario loading
├── title/         # Title sequence
├── ui/            # Window system
└── world/         # World state and rendering
```

`OBJECT_TYPES.md` documents the Locomotion object type IDs.

## Development

```bash
go build ./...
go test ./...
```

CI runs the test suite on every push and PR.

## License

Reimplementation project; original game by Chris Sawyer. Requires an original
copy of Locomotion for game data.
