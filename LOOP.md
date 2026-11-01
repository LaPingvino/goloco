# GoLoco correction loop

Started 2026-11-01 (Joop: "set up a loop to correct all things previous you disregarded and got wrong").
A session cron fires a tick every ~15 min. Each tick: read this file, do ONE bounded chunk (~10 min),
verify it, commit, update this file (queue + one tick-log line), stop. Don't start two items.

## Non-negotiables
- **Upstream is the truth.** `~/Projects/OpenLoco` (src/OpenLoco/src/…). Every correction cites the
  upstream file/function it mirrors. Where goloco's code or comments disagree with upstream, upstream
  wins — unless the code deliberately keeps a vanilla quirk (0x004XXXXX address comments).
- **Port, don't approximate.** No new "simplified", "approximate" or fudge-factor code. When a full port
  doesn't fit in a tick, port a faithful slice (with a test) rather than a whole approximate feature.
- **Never delete a warning to make a number go up.** A TODO / "approximation" / "stub" note goes only
  when the thing it warns about is fixed. cmd/diagnostics rewards "TODO-free" files — ignore that score.
- **Verify every change**: `go test ./...` (headless-safe packages), plus a visual check for anything
  that draws (native under Xvfb when installed; otherwise the private wasm test page, item 1).
  Compare against the original (reference/ screenshots, or the OpenLoco build when it runs here).
- **Game data is not ours.** ~/Projects/locomotion (from ~/locodata.tar.gz) is never committed and
  never published publicly. The wasm test page is a PRIVATE artifact for Joop only — never make it
  public, never share its link; public builds wait until the data is properly substituted.
- **Git:** commit each tick to main locally with a clear message. Pushing to origin waits for Joop's
  go (asked 2026-11-01). Never commit data, reference screenshots or build output.
- Design questions (architecture splits beyond the queue, gameplay choices) → write them under
  "Questions for Joop" and move to the next item.

## Where things are
- Repo: ~/Projects/goloco (github.com/LaPingvino/goloco). Handbook: CLAUDE.md (its paths say ~/OpenLoco
  and the Steam folder — fix in item 2).
- Data: ~/Projects/locomotion (Data/, ObjData/, Scenarios/) — found via ../locomotion/Data.
- Build: `GOCACHE=$TMPDIR/gocache GOMODCACHE=$TMPDIR/gomod GOFLAGS=-modcacherw go build ./cmd/goloco`
  (module downloads need proxy.golang.org, storage.googleapis.com, sum.golang.org allowed).
  Wasm: `GOOS=js GOARCH=wasm go build -ldflags='-s -w'` → 21.4 MB, 5.1 MB gzipped.
- Writes under ~/Projects/goloco need the sandbox bypass. If a tool call is denied, log it under
  Blocked and move on — don't work around it.

## Queue (top = next)
1. [ ] **Private wasm test page.** Goal: Joop can click a private Artifact link and see/play the current
   build. Pieces: (a) web/ — index.html + wasm_exec.js from the Go toolchain + a JS in-memory fs shim
   (Go js/wasm routes os.Open/ReadDir/Stat through globalThis.fs: implement open/read/fstat/close/
   stat/lstat/readdir over fetched buffers, so goloco's loaders run unchanged); (b) a data pack
   script: g1.DAT, language/interface files, ObjData, a few scenarios — no music — packed into
   chunks < 15 MB (artifact limits: 255 files, 64 MB/version); (c) gzipped wasm, unpacked in the
   page with DecompressionStream; (d) publish privately, check it loads, log misses from the shim.
   Files goloco opens: cmd/goloco/main.go 130/204/217/1365/1951, pkg/scenario/meta.go,
   pkg/objects/loader.go, pkg/assets/datloader.go, pkg/i18n, pkg/audio, pkg/render/atlas.go, pkg/ui/font.go.
2. [ ] **Clean slate.** The July session left uncommitted work (Build Vehicles window
   cmd/goloco/buildvehicle.go; SpawnVehicleAt refactor in pkg/world/vehicle_sim.go; WinRun fleet of 3).
   Review it against upstream BuildVehicle.cpp, finish or trim to a correct slice, commit. Update
   CLAUDE.md paths for this machine.
3. [ ] **Inventory of deviations.** Walk the ~90 markers (approximat|simplif|fudge|hack|stub|TODO|
   FIXME|placeholder|not yet implemented) and every invented heuristic; write docs/deviations.md:
   file:line, what upstream does instead (file/function), severity, fix size. This becomes the queue.
4. [ ] **Split simulation from rendering** (needs Joop's yes — see Questions). pkg/world imports
   Ebiten in world.go, construction.go, overlay.go; its tests panic without a display.
5. [ ] **Vehicle movement port.** Generate SubpositionData.cpp into Go like track_data_gen.go;
   port the movement step (VehicleBogie/VehicleHead/VehicleManager) in faithful slices with tests
   against upstream values; retire vehicle_sim.go's straight-segment approximation.
6. [ ] Then: deviations from docs/deviations.md, worst first (text measurement stub, terrain generation
   'simplified approximation', depth-sort approximation for vehicles, g1font width fudge, …).

## Questions for Joop
- Push each tick's commits to origin main, or keep them local until you review?
- Item 4 (splitting pkg/world into simulation and rendering): go?

## Blocked
- 2026-11-01: Xvfb is not installed on the VPS (`sudo pacman -S xorg-server-xvfb` — needs Joop).

## Tick log
- 2026-11-01 (setup) — LOOP.md written; Locomotion data extracted to ~/Projects/locomotion; native and
  wasm builds confirmed; tests pass except pkg/world (needs a display).
