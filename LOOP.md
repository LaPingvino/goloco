# GoLoco correction loop

Started 2026-11-01 (Joop: "set up a loop to correct all things previous you disregarded and got wrong").
Runs by day as a session cron (~15 min); overnight via the idle-loop Stop hook: .idle-loop.json here,
switched on with `.idle-loop-on` (content 'until HH:MM'; off after 24 h) in the session's working directory —
currently ~/Projects/empty (its .idle-loop.json points here; status goes to ~/Projects/empty/.idle-loop-status).
Each overnight tick ends by writing work/nowork to .idle-loop-status. Each tick: read this file, do ONE bounded chunk (~10 min),
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
- **Git:** commit each tick to main with a clear message and push to origin main (Joop, 2026-11-01:
  'feel free to push to Github'). Never commit data, reference screenshots or build output.
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
1. [x] **Private wasm test page.** DONE tick 2: published as a private Artifact (link in Joop's chat — never in the
   repo). Rebuild + republish: `go run ./cmd/webpack -out <scratchpad>/goloco-web -scenarios "..."`, then publish
   index.html with the other files from that folder to the same artifact. The host serves no gzip/binary types,
   so chunks ship as base64 text of gzip (.b64.txt, < 15 MB each; ~42 MB total of a 64 MB budget).
   Headless check: chromium --allow-file-access-from-files --use-angle=swiftshader --virtual-time-budget=300000
   on a copy with an autostart script; screenshots come out black (WebGL readback), so capture with goloco's own
   `--shots` (args in manifest.json) and print the PNG as base64 to the console. Wasm sprite preload is slow
   (11,057 sprites; still at 1% after 60 s under swiftshader) — check the speed in a real browser, and make the
   preload incremental if it is slow there too (→ Questions).
2. [x] **Clean slate.** DONE tick 3: the July Build Vehicles window was not a port of BuildVehicle.cpp (no
   transport/track tabs, raw cost factor as 'Cost', placed running trains on track instead of VehicleCreate
   into a depot) → parked on branch wip/build-vehicles-july (pushed), main clean. CLAUDE.md paths updated
   for this machine. Real port of the window queued after item 5 (needs depots + VehicleCreate).
3. [ ] **Inventory of deviations.** Walk the ~90 markers (approximat|simplif|fudge|hack|stub|TODO|
   FIXME|placeholder|not yet implemented) and every invented heuristic; write docs/deviations.md:
   file:line, what upstream does instead (file/function), severity, fix size. This becomes the queue.
4. [ ] **Split simulation from rendering** (APPROVED by Joop 2026-11-01: 'the split is perfectly okay. Free to make bold decisions.'). pkg/world imports
   Ebiten in world.go, construction.go, overlay.go; its tests panic without a display.
5. [ ] **Vehicle movement port.** Generate SubpositionData.cpp into Go like track_data_gen.go;
   port the movement step (VehicleBogie/VehicleHead/VehicleManager) in faithful slices with tests
   against upstream values; retire vehicle_sim.go's straight-segment approximation.
5b. [ ] **Build Vehicle window**, ported from Ui/Windows/BuildVehicle.cpp (tabs per transport type and
   track type, availability by year, VehicleCreate into a depot). Needs item 5. Reference: the parked
   branch wip/build-vehicles-july.
6. [ ] Then: deviations from docs/deviations.md, worst first (text measurement stub, terrain generation
   'simplified approximation', depth-sort approximation for vehicles, g1font width fudge, …).

## Questions for Joop
- Wasm test page: how long does it take to get from Start to the title screen in your browser? (Headless it was
  still preloading sprites after a minute.)

## Blocked
- 2026-11-01 tick 1: serving the test page from a local web server (python -m http.server) was refused by the auto-mode check ('Expose Local Services'). Test from file:// with chromium --allow-file-access-from-files instead.
- 2026-11-01: Xvfb is not installed on the VPS (`sudo pacman -S xorg-server-xvfb` — needs Joop).

## Tick log
- 2026-11-01 (setup) — LOOP.md written; Locomotion data extracted to ~/Projects/locomotion; native and
  wasm builds confirmed; tests pass except pkg/world (needs a display).
- 2026-11-01 (setup) — pushing allowed (Joop); overnight idle-loop configured in ~/Projects/empty, the session's directory.
- 2026-11-01 tick 1 (item 1) — wasm test page parts (a)–(c): in-memory fs shim for Go's js/wasm syscalls (open/read/fstat/stat/readdir…, O_DIRECTORY fixed after the first run), cmd/webpack (data pack in gzipped chunks < 15 MB, no music, refuses an output dir inside the repo), page with progress + Start + log drawer. Headless run: g1, 544 objects, title scenario load; canvas black → next tick. Split approved by Joop.
- 2026-11-01 tick 2 (item 1 done) — black screenshots were a WebGL readback artefact: goloco's own --shots capture shows its loading screen drawing correctly. Artifacts serve no gzip, so the pack ships as base64 text of gzip; published privately; verified headless (544 objects, init complete). Body padding no longer squeezes the canvas.
- 2026-11-01 tick 3 (item 2 done) — July WIP reviewed against upstream BuildVehicle.cpp: an invented list + track placement, not a port → parked on wip/build-vehicles-july; main clean; CLAUDE.md paths fixed (~/Projects/OpenLoco, ~/Projects/locomotion; no Xvfb yet, wasm page instead). Queued the real window as 5b.
