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
- **Exact on the CPU first, then optimise with proof (Joop, 2026-11-01: 'build 100% match on CPU first … The
  original game had no GPU anything at all'; 'optimizing is easier if you can check every step that the code
  remains equivalent').** Every rendering rule gets a CPU implementation that matches upstream exactly, with a
  test; a faster version (GPU, caching) only replaces it with an equivalence check against the CPU result.
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
- The test page is private: the Caddy route https://kiefte.eu/goloco/ (basic auth) or a private Artifact —
  never a public URL.
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
1. [x] **Private wasm test page.** Since 2026-11-01 (Joop: the artifact is slow; serve it from the VPS
   behind Caddy with HTTP auth): `go run ./cmd/webpack -local -out ~/Projects/goloco-web-local` and
   https://kiefte.eu/goloco/ (basic auth; set up once by Joop with `sudo bash scripts/caddy-goloco.sh`).
   Rebuild = `go run ./cmd/webpack -local -music -out ~/Projects/goloco-web-local` (music fetched per track on first
   open via the shim's lazy files); no Caddy change needed. Earlier: DONE tick 2: published as a private Artifact (link in Joop's chat — never in the
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
3. [x] **Inventory of deviations.** DONE tick 4: docs/deviations.md — 103 markers sorted into gameplay (A),
   rendering (B), text/UI (C), dead scaffolding (D: pkg/loco/*, pkg/tests, uncalled ui stubs — keep pkg/core
   and pkg/olmath, real ports), not implemented (E); with the upstream file for each and a fix order.
3b. [x] **Remove dead scaffolding** DONE tick 5: deleted pkg/loco/* (5 packages), pkg/tests, and the 11 uncalled
   functions/types in pkg/ui (window_manager.go: ProcessEvents, UpdateAnimations, ViewportsUpdatePosition, UiWidth,
   UiHeight, GfxRender, WindowEventList, UpdateWindows, RenderWindows, InitWindowManager; types.go: Window.Draw).
   Note: pkg/ui's GPT-generated Window/Widget system (window_manager.go header: 'batch_ai_impl … GPT-5-mini')
   is still LIVE via input.go and title_menu.go, next to the SimpleWindow system the game windows use — two
   window systems. Unify on a port of Ui/WindowManager.cpp later (→ deviations C).
3c. [x] **Startup speed** DONE tick 6: the blocking preload of every sprite before the first frame (11,057 for
   the title) is gone; sprites decode on first use (as upstream draws g1 directly) and the rest warm in the
   background within 4 ms per frame (prewarmTick). Headless capture at 3 s now shows the title screen (it showed
   the loading bar at 1% before). Startup timings logged ('First world frame …', 'Background prewarm complete
   …'). The progress window is now unreached — marked: upstream shows it during S5 loading (S5.cpp:728).
   Local test build now carries all 46 scenarios (631 files, 47 MB).
3d. [~] **Open-graphics build (Joop: a public link once possible).** Tick 7: research done →
   docs/open-build.md (OpenGraphics is MIT; objects for every type, but 1 building/1 industry/1 tree/1 sound;
   no g1.DAT replacement; title + scenarios must be our own). Steps there wait for Joop's yes (→ Questions). OpenLoco fetches OpenGraphics
   (github.com/OpenLoco/OpenGraphics, objects.zip v0.1.8 in OpenLoco's CMakeLists.txt): so far the interface
   skin, walls, tunnels, land, road stations, streetlights, hill shapes, some vehicles. Research step: list
   what goloco loads (g1.DAT base sprites, fonts, objects per type, title.dat, scenarios) against what
   OpenGraphics replaces; write docs/open-build.md with the gap. A public build also needs its own title
   scene and scenario made only from open objects (Joop: 'we might need a custom level/title screen') —
   design, wait for Joop before building. Never publish anything containing the original data.
3e. [~] **Joop's observations on the live build (2026-11-01):** (1) slopes and corners are wonky; (2) water is
   still slightly broken where it touches the coast; (3) roads up and down slopes don't work yet. Each: find the
   upstream paint code (Paint/PaintSurface.cpp for slopes/corners + water edges; Paint/PaintRoad.cpp and
   Map/Track/TrackData.cpp road slope pieces), capture the case with --diag, fix, compare. Interleave with item 4
   (one per tick when a split step is done).
   Water (tick 12): diag capture at the coast of Weatherworld. Two departures from Paint/PaintSurface.cpp
   1736–1775: (a) goloco animates waves on EVERY water tile with an invented frame; upstream draws frame +60
   only on tiles with flag 6, from WaveManager waves that spawn at random on water tiles surrounded by water;
   (b) goloco draws all water in a late separate pass ('pass 1.5'), upstream adds water to the paint sort at
   waterHeight with bbox {32,32,-1} — a late pass can overpaint neighbouring land at the shore. Done so far:
   WaveManager + tile-loop water branch ported to pkg/sim/wave.go with tests; tick 13: wired — World.Update runs
   UpdateTiles (visible = tile centre on the last drawn screen) + UpdateWaves; the painter draws +60+frame only on
   WaveFlag tiles (diag: the all-over wave flecks are gone). NOTE: goloco ticks once per frame (60/s), upstream's
   game tick is 25 ms (40/s) — separate deviation, affects every timed effect. Remaining water item: (b) the late
   water pass → with the paint-sort work (deviations B, vehicle depth).
   (c) FOUND tick 14 — the colour of the water itself. Reference (reference/gallery-screenshot-7.png): dark,
   textured water, the seabed showing through; goloco: flat pale teal. Upstream (Graphics/DrawSprite.cpp
   getDrawBlendOp): an image that is blended but has no primary colour is 'only a transparency mask — just colour
   the pixels using the palette map' → for each non-transparent pixel of water +35, dst = waterMap[dst], where
   waterMap = the water object's image +41 (WaterObject::kColourPalette; ObjectManager updateWaterPalette sets
   ExtColour::water to it). Then +30 is drawn normally on top. goloco instead remaps the +35 sprite into an
   opaque flat colour, and its pass-2 blend lerps toward one tint (worldBlendOp) — both approximations.
   NEXT TICK: make the blend pass exact — at zoom 0 every drawn pixel is a palette colour, so reverse-look-up the
   index (RGB→index table from the G1 palette), map it through the 256-entry map, write the mapped palette colour;
   route water +35 through it as a mask (and glass, which uses the same op); compare with gallery-screenshot-7.
   Tick 15: built it — pkg/world/blend.go (Kage shader: mask + destination sub-image + a mask-sized table of
   palette[i] and palette[map[i]]; exact match search over 256 entries), the dead CPU tint pass removed, water
   sprites (+30, waves, zoomed-out fallback) now drawn after the blend composite. BUT the shader's read of the
   worldBuf sub-image (imageSrc1At) returns black in headless swiftshader although the seabed is drawn there
   (debug variant returning the terrain colour: black). Probe (throwaway program): the water map +41 is fine
   (256 entries, land colours → indices 192–197, dark teal). So the exact path is behind `exactWaterBlend =
   false`; NEXT: a tiny Kage repro of sub-image sampling in pixel mode (origin mapping; maybe copy the region
   into a mask-sized temp image first and sample that at src), then switch it on and compare with the reference.
   Tick 16: DONE on the CPU instead (Joop: exact on CPU first) — pkg/palblend (pure, tested headless): read the
   frame once, recover each masked pixel's palette index from its exact colour, dst = map[dst] (skip if 0),
   write to blendBuf. The shader is gone. Bug found by counting: the mask came from DecodeSprite, which drops
   indices whose palette alpha is 0 — the water mask uses 1-4 → now DecodeSpriteMaskOnly. Diag now shows dark,
   textured water with the seabed through it, like reference/gallery-screenshot-7.png. Live build updated.
   PERF: one full-frame ReadPixels per frame with water on screen.
3i. [ ] **DESIGN (bold, Joop-approved direction): port the software renderer.** OpenLoco draws into an 8-bit
   palette-indexed RenderTarget with software sprite blits (Graphics/DrawSpriteHelper.hpp ops: none,
   transparent, src remap, dst blend, src+dst blend, noise mask) and a sorted paint session (Paint/Paint.cpp:
   addToPlotList, quadrant sort). Porting both gives exact blends without colour reverse-lookup, and fixes the
   draw-order deviations at once (water 'pass 1.5', vehicles 'post-tile depth approximation'). Plan: (1)
   pkg/gfx8: indexed framebuffer + drawSprite with every DrawBlendOp, tested against hand-made sprites; (2) a
   paint session port; (3) move painters onto it one element at a time, each checked against the current RGBA
   output before switching (equivalence per step). Ebiten then only shows the final RGBA frame.
3f. [ ] **Standing (Joop: 'go through all of it and fix by bits and pieces'):** walk the paint code element by
   element against upstream Paint/: surface (slopes, corners, cliffs, water, snow), track, road (incl. slopes),
   stations, signals, bridges, tunnels, trees, buildings, industries, walls, vehicles — one element per tick:
   read the upstream painter, diff goloco's, capture with --diag, fix, note it here.
3g. [ ] Music stutters in the browser (Joop): find where audio is decoded/streamed (pkg/audio, oto in js/wasm);
   likely decode on the main thread or too small a buffer.
3h. [ ] (later, Joop: 'first fix and put in context how the graphics should work, then replace it with a fully
   free version') A draft/comic-style g1 for testing: needs a g1.DAT writer (goloco only decodes) and drawings
   for the ids goloco uses — after the graphics are right, as the first step of docs/open-build.md.
4. [~] **Split simulation from rendering** (APPROVED by Joop 2026-11-01: 'the split is perfectly okay. Free to make bold decisions.').
   Plan: new pkg/sim (no ebiten/render imports — ebiten panics at init without a display, so headless
   tests need a package that never imports it) holding map state + rules; pkg/world keeps drawing and embeds
   it. Steps: (1) Tile → sim.Tile, fields exported [DONE tick 8; the write-only invented TileType
   (grass/dirt/water) and NewWorld's placeholder 'variety' map removed]; (2) sim.Map {Width, Height, Tiles,
   Towns, Entities} + TileAt, World embeds sim.Map by value [DONE tick 9; LoadFromScenario's data part still
   in world.go → move with step 4]; (3) construction logic
   (placedPiece, ConstructionHead, track/road data tables) [tick 10: tables moved; tick 11: sim.State = Map + ConsHead
   + undo stacks; construction rules (head, place/undo track+road, footprints → TrackFootprint/RoadFootprint)
   and Map.AddTrack/AddRoad now in pkg/sim; pkg/world/construction.go only draws. OWED: a visual check of the
   construction ghost/head marker (scripted session with GOLOCO_OPEN=track via go.env in the test page)]; (4) stations + vehicles + lists; (5) move their
   tests to pkg/sim and run them headless. Tooling: `gofmt -r 'a.old -> a.New'` is NOT type-aware — it also
   renames same-named fields of other types; check with `go build -gcflags=-e`. pkg/world imports
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
- (answered 2026-11-01) g1 gap: Joop allows drawing replacements myself, or borrowing from OpenTTD
  (check licences first: OpenTTD's OpenGFX is GPL-2.0; OpenGraphics is MIT).
- docs/open-build.md: go for step 1 (an '-open' pack with only OpenGraphics objects and a goloco start path
  without g1.DAT)? And for the g1 gap: (a) draw open replacements for the ~85 g1 ids goloco uses, or (b)
  goloco-only UI art with TTF fonts in the public build?

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
- 2026-11-01 tick 4 (item 3 done) — docs/deviations.md: 103 markers + invented heuristics, each with the upstream file it should follow, severity and size; found 5 dead packages (pkg/loco/*, pkg/tests) → new item 3b.
- 2026-11-01 (Joop: artifact too slow → Caddy with password) — webpack -local writes plain .gz (page decodes both forms); scripts/caddy-goloco.sh inserts a marked /goloco/ block with basic_auth into the kiefte.eu site, backs up, validates (restores on failure), reloads; dry run on a copy validated and replaced cleanly on a second run. Needs Joop's sudo run.
- 2026-11-01 tick 5 (item 3b done) — removed 6 dead packages and 11 uncalled pkg/ui stubs (−~1,400 lines incl. packages); build + vet green; tests as before (pkg/world needs a display → item 4). Found two coexisting window systems (GPT-generated Window + SimpleWindow) → noted.
- 2026-11-01 — Caddy route live (Joop ran the script): 401 without auth, 200 with; wasm served in 0.07 s → slowness is client-side; queued 3c (lazy sprite preload).
- 2026-11-01 tick 6 (item 3c done) — blocking sprite preload replaced by lazy decode + 4 ms/frame background prewarm; title visible at the first 3 s capture (was 1% loading bar); all scenarios in the local build; OpenGraphics noted for a future public build (3d).
- 2026-11-01 tick 7 (item 3d research) — OpenGraphics v0.1.8 (MIT) compared with the original data per object type; gap = g1.DAT (UI, fonts, logo), title/scenarios, sounds/music, variety → docs/open-build.md with a 4-step plan awaiting Joop.
- 2026-11-01 tick 8 (item 4 step 1) — pkg/sim created with sim.Tile (exported fields); world uses it via an alias; dropped the write-only invented TileType and the placeholder 'variety' map; build+vet green; title capture identical. Joop: may draw g1 replacements or borrow from OpenTTD (licence check).
- 2026-11-01 tick 9 (item 4 step 2) — sim.Map (size, tiles, towns, saved entities, TileAt) embedded in World by value so zero Worlds in tests stay usable; build+vet green; title capture matches (only the water animation frame differs).
- 2026-11-01 tick 10 (item 4 step 3a) — generated track/road coordinate tables moved to pkg/sim with exported names; build+vet green. Joop: add music to the private build → lazy per-file fetch planned (shim open() fetches on first use).
- 2026-11-01 (Joop: music in the private build) — webpack -music (local only): 30 music tracks as separate .gz files, registered as lazy files that the fs shim fetches on first open(); headless: css5.dat (title music) fetched and decoded. Live build 445 MB on disk, nothing extra downloaded until a track plays.
- 2026-11-01 tick 11 (item 4 step 3b) — sim.State (Map + construction head + undo stacks); construction rules and AddTrack/AddRoad moved to pkg/sim; World embeds sim.State; build+vet green; ghost drawing owes a visual check. Joop's three observations queued as 3e.
- 2026-11-01 tick 12 (item 3e water) — coast diag captured; found invented all-tile waves and a late water pass vs upstream's paint sort; ported WaveManager + the tile loop's water branch to pkg/sim (first headless tests in pkg/sim, all pass). Queued: 3f element-by-element paint audit (Joop), 3g music stutter, 3h draft g1 (after fidelity).
- 2026-11-01 tick 13 (item 3e water a) — waves wired: tile loop + wave update per tick, frames drawn only on wave tiles; diag shows plain water instead of flecks everywhere; live build updated. Noted the 60 vs 40 ticks/s deviation.
- 2026-11-01 tick 14 (item 3e) — reference gallery found in goloco/reference/; water colour cause traced to upstream's mask-blend (dst through the water palette map +41) vs goloco's opaque remap + tint lerp; plan recorded; no code change.
- 2026-11-01 tick 15 (item 3e water c) — exact palette-map blend implemented as a Kage shader; terrain read comes back black under headless WebGL, so it's switched off (exactWaterBlend=false) pending a repro; water looks as before; ripple/wave draw order now after the blend; dead CPU tint pass removed.
- 2026-11-01 tick 16 (item 3e water c done) — exact palette-map blend on the CPU (pkg/palblend + test); mask bug (DecodeSprite drops palette-alpha-0 indices) fixed; water now matches the reference look; live. New rule from Joop: exact CPU first, optimise with equivalence checks. Queued 3i: port OpenLoco's indexed software renderer + paint sort.
