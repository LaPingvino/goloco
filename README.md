# GoLoco — OpenLoco Reimplementation in Go

> ⚠️ **Project Status:** See [STATUS.md](../STATUS.md) for actual verified state.
> Previous documentation was over-optimistic. This README reflects reality.

## What Actually Works

- ✅ Compiles and runs without crashes
- ✅ Loads Locomotion game data (G1 sprites, objects)
- ✅ Basic isometric tile rendering (placeholder maps)
- ✅ Simple UI system (windows, drag, click)
- ✅ Audio/music playback
- ✅ Camera controls (pan, zoom)

## What Doesn't Work (Known Issues)

- ❌ Scenario S5 decompression fails (creates placeholder maps)
- ❌ Paint system disabled (can't render tracks/roads/buildings properly)
- ❌ Excessive debug logging (performance issue)
- ❌ No gameplay systems (economy, companies, vehicles, stations)
- ❌ No unit tests

See [STATUS.md](../STATUS.md) for detailed issue list.

## Quick Start

```bash
cd /home/joop/goloco-project/goloco

# Build
go build ./...

# Run
./goloco
```

**Note:** Requires Locomotion game data in `../locomotion/Data` or falls back to placeholders.

## Documentation

| Document | Purpose |
|----------|---------|
| [STATUS.md](../STATUS.md) | Actual verified state of all systems |
| [TODO.md](../TODO.md) | Realistic task list with estimates |
| [DEVELOPMENT.md](../DEVELOPMENT.md) | Guidelines to avoid past mistakes |
| [BUILD_NOTES.md](../BUILD_NOTES.md) | Build, run, and testing reference |

**Read STATUS.md before starting any work.**

## Project Structure

```
goloco/
├── cmd/
│   └── goloco/main.go       # Game entry point
├── pkg/
│   ├── assets/               # G1/DAT loading
│   ├── audio/                # Audio/music
│   ├── graphics/             # Drawing primitives
│   ├── objects/              # Game objects
│   ├── paint/                # Depth-sorted rendering (partial, mostly disabled)
│   ├── render/               # Ebiten adapter
│   ├── scenario/             # Scenario loading (S5 parser)
│   ├── ui/                   # Window system
│   ├── world/                # World rendering
│   └── loco/                 # Clean type definitions
├── STATUS.md (↑)           # Actual project status
├── TODO.md (↑)             # Realistic task list
├── DEVELOPMENT.md (↑)       # Development guidelines
└── BUILD_NOTES.md (↑)      # Build/run reference
```

## Controls

- **Arrow keys / WASD:** Pan camera
- **Q / E:** Zoom in/out
- **Mouse drag (edges):** Pan camera
- **Mouse wheel:** Zoom
- **Mouse click:** Interact with UI

## Known Issues

See [STATUS.md](../STATUS.md) for complete list. Critical issues:

1. **S5 decompression fails** - Can't load real maps/scenarios
2. **Paint system disabled** - Can't render tracks/roads/buildings
3. **Excessive logging** - Performance degradation
4. **No gameplay systems** - Economy, companies, vehicles missing

## Next Steps

See [TODO.md](../TODO.md) for prioritized task list.

Immediate priorities:
1. Fix S5 chunk decompression
2. Remove excessive debug logging
3. Enable and debug paint system
4. Add basic unit tests

## Contributing

Before contributing:
1. Read [DEVELOPMENT.md](../DEVELOPMENT.md) for guidelines
2. Check [STATUS.md](../STATUS.md) for actual project state
3. Review [TODO.md](../TODO.md) for current priorities
4. Test thoroughly before claiming anything works

## License

This is a reimplementation project. Original game by Chris Sawyer.

---

**Last Updated:** 2026-01-31

**Important:** Do not trust older documentation. Read STATUS.md for actual state.
