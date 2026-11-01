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
   (d) Tick 17: submerged terrain — goloco flattened it (slope 0) and lifted it to the water height, citing
   PaintSurface ~676-688; that is getWaterSurface, used only for selection boxes. Upstream always paints the
   surface at baseHeight with its displaySlope (paintMainSurface, ~1545-1569). Fixed: the seabed's real slope
   now shows through the blend — shallow bands along the coast, as in the reference. Open: the light-green tile
   at the waterline in the Weatherworld diag (growth stage or variation?) — check against PaintSurface next.
   Tick 18: that tile is (251,137), slope 0x6 → display slope 9, a light-facing facet; data consistent (growth 4
   land, 0 under water, water level = next land step) and the reference shows the same bright coast slopes →
   not a bug. Roads: a headless test (pkg/sim) places slope pieces 5-8 in all rotations and moves the head by DZ
   — construction is fine; the problem must be drawing. Real slope pieces exist in Weatherworld at (228,23)…
   (233,28) (road 8, rot 3). Their capture showed a WORSE issue: water blended OVER the town in front of the sea
   — the blend recolours every pixel in a water tile's footprint, buildings drawn later included (the old opaque
   'pass 1.5' covered the same buildings). NEXT TICK (interim, until 3i): draw tracks/roads/buildings/trees/
   walls/stations/signals also into an object-mask buffer and make palblend skip pixels an object covers; then
   capture (231,25) again to judge the road slopes.
3i. [~] **DESIGN (bold, Joop-approved direction): port the software renderer.** OpenLoco draws into an 8-bit
   palette-indexed RenderTarget with software sprite blits (Graphics/DrawSpriteHelper.hpp ops: none,
   transparent, src remap, dst blend, src+dst blend, noise mask) and a sorted paint session (Paint/Paint.cpp:
   addToPlotList, quadrant sort). Porting both gives exact blends without colour reverse-lookup, and fixes the
   draw-order deviations at once (water 'pass 1.5', vehicles 'post-tile depth approximation'). Plan: (1)
   pkg/gfx8: indexed framebuffer + drawSprite with every DrawBlendOp, tested against hand-made sprites; (2) a
   paint session port; (3) move painters onto it one element at a time, each checked against the current RGBA
   output before switching (equivalence per step). Ebiten then only shows the final RGBA frame.
   Tick 19 — step (1) DONE: pkg/gfx8 — RenderTarget (8-bit), Decode (raw + RLE → indices), blitPixel exactly as
   DrawSpriteHelper.hpp (transparent / src / dst / src+dst rows / noise mask), OpFor = getDrawBlendOp, clipped
   DrawSprite with offsets. Tests: every op on hand-made sprites; equivalence: all 4,118 g1.DAT sprites decode
   to indices whose palette colours equal the existing RGBA decoder's. Why now (instead of an interim mask): the
   water blend must hit only what was drawn BEFORE the water tile, which needs in-order blending — what an 8-bit
   target gives for free. NEXT: (2) the paint session (Paint/Paint.cpp: addToPlotList, bounding boxes, quadrant
   sort), then (3) surfaces first, compared pixel-for-pixel with the current output on dry land.
   Tick 20 — step (2) core DONE: pkg/paint — createNormalPaintStruct (offset/bbox rotation, gameToScreen,
   imageWithinRT), addToPlotListAsParent/AsChild, attachToPrevious, quadrant bucketing, arrangeStructs with its
   helper ported line for line (all 4 rotations of checkBoundingBox), drawStructs (children + attachments).
   Tests: back-to-front tile order; an object on a tile draws after its ground even when added first; children
   and attachments. NOT ported yet: track/road priority lists (finaliseTrackRoadOrdering), string structs,
   culling, interaction lookup, masked images. NEXT: (3) a surface painter on gfx8+paint (PaintSurface main
   surface only) rendering a test region, compared pixel-for-pixel with the current RGBA output on dry land.
   Tick 21 — step (3a): cmd/render8 renders a region through gfx8+paint to PNG with no graphics library
   (headless CPU reference). Main surface only, same sprite choice as world.go, paintMainSurface box
   (offset 0,0,baseHeight; size 32,32,-1). Title map at 196,190: the terrain matches goloco's diag of that spot;
   cliffs are black gaps (not painted yet). NEXT: (3b) cliff edges (paintSurfaceCliffEdge) in render8; (3c) an
   aligned pixel diff against the wasm --diag crop (same camera maths) to prove equivalence on dry land; also
   check goloco's surface image base (ImageOffset + GetFlatTerrainSpriteIndex) against upstream LandObject::load
   'image' — render8 copies goloco's formula for now, so both could share an error.
   Tick 22: surface image base CHECKED equal to upstream LandObject::load (image = angles × growthStages × 57 +
   offset; perStage = angles × 25; mapPixelImage = angles × growthStages × 82 + offset). Masked images added
   (gfx8.DrawSpriteMasked = drawImageMaskedZoom<0>: dst = mask & image where non-zero, image-local mask reads;
   paint.Image.Mask) with a test — needed by cliff edges, smoothing, snow. NEXT: port paintSurfaceCliffEdge +
   paintEdgeSection (PaintSurface.cpp 1198–1345) into a surface painter package used by render8.
   Tick 23: DONE — pkg/paint/surface.go: kCornerHeights, getTileDescriptor edge heights, paintEdgeSection,
   paintSurfaceCliffEdgeImpl (edges 2,3,0,1; masked images 3726/3731 + 0..4), rotation 0; tunnels and the
   underground view NOT ported. render8 now fills the cliff gaps with rock faces sorted against the surfaces.
   NEXT: water in render8 (main water +35 as a dst mask through the +41 map — gfx8 OpDst — +30 attached, waves,
   water cliff edges), then the aligned pixel diff against the wasm --diag crop.
   Tick 24: water DONE in render8 (paint.PaintWater: +35+shape as gfx8 OpDst through the +41 map at waterHeight,
   bbox 32,32,-1; +30+shape attached; water cliff edges; no waves in a still frame). At Weatherworld (231,25) —
   where the Ebiten renderer blends water over the town — the water now stops cleanly at the coast; the land in
   front is untouched (in-order blending). NEXT: objects in render8, in the order the deviations list ranks them
   (trees + buildings first: they are what water was painted over), each from its upstream painter.
   Tick 25: trees ported (pkg/paint/tree.go, paintTree: season swap tables, quadrant offsets, bbox, shadow through
   the unk32 map 2208) and drawn by render8. WRONG STILL: the second (season-transition) image draws in full —
   tree sprites are NOT RLE (flags 0x1, checked), so upstream DOES apply the noise mask (images 3479-3485,
   128×256, _noiseMaskImages[unk7l+1 or 8-…]). NEXT: gfx8 noise op — read how drawBMPSprite indexes the noise
   image for a sprite pixel (DrawSprite.cpp / DrawSpriteHelper), port it, pass the noise sprite from PaintTree.
   Also verify goloco's tree season sprite remap against upstream TreeObject::load.
   Tick 26: noise mask DONE — gfx8.DrawSpriteNoise (non-RLE only; noise byte at noise[sprite.Width*y + x], the
   sprite's width as stride, as DrawSpriteBMP.hpp), paint.Image.Noise, PaintTree passes noise image 3478+n
   (n = unk7l+1, or 8-n when the season order flips). Transition trees now dither their two seasons, as upstream.
   Tick 27: tree season remap VERIFIED equal to TreeObject::load (var3C assignment, the five fallbacks, snow at
   +numPrimaryImages, shadow offset). Only difference: goloco pre-fills SnowSprites with the base offset where
   upstream leaves them unset — matters only for a snowy tree whose object has no snow variation (not seen).
   NEXT: buildings (Paint/PaintBuilding.cpp) in render8 — deviations B lists 'first variation only' and
   'under construction skipped' for the current renderer; port the real painter.
   Tick 28: buildings DONE in render8 (pkg/paint/building.go: paintBuilding + paintBuildingBuilding — the tile's
   variation, all parts, part animation frames, colour remap, 2×2 front-tile rule, elevators). The building loader
   now keeps PartAnimations, ElevatorSequences, scaffolding type/colour (its section walk was already equal to
   BuildingObject::load). Weatherworld (231,25): the town stands in front of the sea, water nowhere over it.
   NOT YET: scaffolding/under-construction (BuildingElement age + unk5u aren't parsed), building shadows
   (Colours::getShadow(orange) translucency), ghosts. NEXT: roads (PaintRoad.cpp) — then capture the road slope
   pieces at (228,23)…(233,28) for Joop's 'up and down roads' observation.
   Tick 29 — roads, survey: PaintRoad.cpp (719 lines) + data headers include/OpenLoco/Paint/PaintRoadCommonData.h
   (576), PaintRoadStyle0/1/2Data.h (233/205/339), PaintRoadAdditionsData.h (458). paintRoad picks
   kRoadPaintCommonParts[roadId][seq] and Style{0,1,2}::kRoadPaintParts by roadObj->paintStyle, then paintRoadPP
   (standard or multi-tile merge) + paintRoadPCP (tunnels/supports), level crossings, streetlights, road extras.
   goloco's road_sprites.go is a hand-copied partial version — the likely source of the slope bugs. PLAN:
   (a) cmd/gen-paintdata: generate Go tables from those headers (like track_data_gen.go — exact, no hand
   copying); (b) port paintRoadPP/PCP for styles 0/2 + 1; (c) addToPlotListTrackRoad + finaliseTrackRoadOrdering
   (the priority lists the paint port deferred); (d) roads in render8; (e) capture the slope pieces.
   Tick 30: (a) DONE — tools/gen-paintdata (gen.cpp compiled against upstream's headers with upstream's own
   include line; gen.sh) writes pkg/paint/roaddata_gen.go: RoadCommonParts, RoadStyle0/1/2Parts — values evaluated
   by the C++ compiler, constexpr rotations included. Road additions not yet. Comparing goloco's hand table
   (pkg/world/road_sprites.go) with it: style2 roadId 3 (left curve small) had pieces 0 and 3 SWAPPED → fixed.
   Roads 0-2 differ by design (merge sprites instead of hit-test images). The SLOPE pieces (5-8) MATCH — so Joop's
   'up and down roads' bug is not in the sprite tables but in placement (heights/bboxes) → check during (b).
   Tick 31: (b)+(c)+(d) core DONE — pkg/paint/road.go (Style02 standard + multi-tile merge, Style1 layers 0/1/3,
   finalisePaintRoad with kExitsToMergeId and the merge bboxes; merge base images generated), paint.Session
   AddTrackRoad + FinaliseTrackRoadOrdering (priority lists → one struct with getMinMaxXYBounding), render8 calls
   finalise at the end of each same-baseZ run (PaintTile.cpp paintTileElementsEndLoop). Weatherworld town: streets
   and junctions render between the buildings. NOT YET: company colour remap (saved company colours not parsed),
   streetlights, bridges/tunnels (paintRoadPCP), level crossings, road extras. NEXT: (e) a close crop of the slope
   pieces (228..233, 23..28) in render8 vs goloco's current renderer — to find the placement bug Joop saw.
   Tick 32: close crops of the town (render8 vs the wasm --diag) — roads match apart from the known water spill;
   the steep pieces there sit behind buildings. Joop's 'up and down roads don't work' most likely means BUILDING
   them: the UI maps slope buttons to 5/6/7/8 correctly and placement passes the sim test, but goloco's current
   paintRoads draws pieces directly in tile order with no depth sort, so a sloped piece can be painted over by
   the next tile's terrain — the class of bug the paint session fixes. Also found in the current paintRoads:
   (1) style-1 roads 0–2 are wrongly merged (upstream: Style1 never merges); (2) an invented left/right lane exits
   split for style 2 (upstream: one exits mask per tile, merge base from the element's own MergeType).
   DECISION: the payoff now is switching the game's World.Draw onto gfx8+paint (render8's pipeline), then
   retiring the per-tile painters one by one. → new item 3j.
3j. [~] **Switch the live renderer to gfx8+paint.**
   PERF (tick 46): pkg/scene BenchmarkRenderTitleView (800×600, title farm view, native, with Whisper using 6 cores):
   28.7 ms → 17.6 ms/frame after clipping the sprite rectangle once and a fast path for plain transparent copies.
   EQUIVALENCE PROCEDURE for every optimisation: render8 the three reference views (title 169,150; Weatherworld
   231,25; Bottleneck Blues 237,12) before and after, `cmp` the PNGs — must be byte-identical (kept in the
   scratchpad, never committed: they contain game graphics). Next candidates (profile): 28k allocations/frame (paint
   structs → a reusable slab), ToRGBA allocating each frame, sprite cache map lookups.
   Tick 47: paint session reused per frame (Reset; structs/attachments from slabs), render target and RGBA kept
   → 2.8 MB → 190 KB allocated per frame, byte-identical on all three views. Allocation COUNT still ~26k: the
   remaining ones are per-element (GetPaletteMap copies 256 bytes per call; frames/speeds slices per building/
   industry) → cache palette maps by colour and keep the animation arrays on the objects. Timing is noisy while the
   Whisper + OCR jobs share the CPU — re-measure when they finish. World.Draw builds a paint session from the visible tiles with
   the ported painters (surface, cliffs, water, trees, buildings, roads), draws into an 8-bit target, converts
   through the palette into one Ebiten image. Keep the old path behind a flag (GOLOCO_OLD_RENDER=1) for comparison
   until tracks, stations, signals, walls, bridges and vehicles are ported; elements not ported yet are drawn by
   the old painters on top meanwhile (marked). Check: title capture + Weatherworld town + a slope road built in
   the game (scripted), vs the old path.
   Tick 33 — step 1 DONE: pkg/scene (Renderer.Render(sim.Map, View) → 8-bit target; ToRGBA) holds render8's
   painting, reading sim.Map instead of the raw scenario; sim.Map.LoadScenario does the scenario→map copy (World
   uses it); tile walk = the box of tiles that can reach the view (or an explicit window). Water now in upstream's
   place (right after the surface). EQUIVALENCE: render8 via pkg/scene is pixel-identical to the pre-refactor
   output on both reference views (title 196,190 and Weatherworld 231,25). NEXT: World.Draw → scene (behind a flag
   first), old painters on top for the elements not ported yet.
   Tick 34 — step 2 DONE: World.Draw → pkg/scene behind GOLOCO_SCENE=1 (zoom 0; pkg/world/scene.go), vehicles and
   labels on top; the test page passes env from manifest.json ("env") into go.env. Title capture in scene mode:
   terrain, cliffs, trees with shadows, water, UI — all through the CPU renderer. MISSING in scene mode: walls,
   tracks, stations, signals, bridges, industries (next: port walls (PaintWall.cpp, small), then tracks). PERF:
   one full-frame CPU render + upload per frame — measure in a real browser before making it the default.
   Tick 35: walls DONE (pkg/paint/wall.go = paintWall: offsets/bboxes per rotation, kImageOffsets[twoSided][rot]
   [slope], primary colour remap; WallObj image ids generated by tools/gen-paintdata). Stone walls follow the
   slopes on the title map. NOT YET: glass second image, secondary colour (not parsed), ghosts. NEXT: tracks
   (PaintTrack.cpp + PaintTrackData.h via the generator — the biggest remaining piece).
   Tick 36: generator split (gen.cpp roads/walls, gen_track.cpp tracks — upstream compiles them in separate
   units, they redefine the same helpers); pkg/paint/trackdata_gen.go = kTrackPaintParts (91 pieces). goloco's
   hand table (pkg/world/track_sprites.go) checked against it: 0 differences. Track painter (paintTrack: mergeable →
   priority lists 0/1/3, standard → AsParent) next.
   Tick 37: tracks DONE (pkg/paint/track.go = paintTrack/paintTrackPP: mergeable → priority lists 0/1/3, standard
   → AsParent). Scene: tracks and roads per baseZ group (ascending), one finalise per group. APPROXIMATION: the tile
   keeps tracks and roads in separate slices, so their order within one baseZ is lost (tracks first) — fix when
   the tile keeps element order. Title map: the railway renders. NOT YET: company colours, bridges, tunnels, track
   extras. NEXT: signals + stations, then make scene mode the default once nothing visible is missing.
   Tick 38: signals DONE (pkg/paint/signal.go = paintSignal/paintSignalSide; offset tables, image rotation,
   frames; TrackSignalHeightOffset generated from TrackData.cpp getTrackMiscData — gen_track now links that
   file). PARSE FIXES (scenario): SignalElement.RightSignalObjectID (byte 6 bits 0-3 — the right side was read with
   the left side's object), TrackElement.IsLast (element flags byte 1, flag_6). Signals pair with the track of the
   same baseZ (APPROXIMATION: upstream uses the element just before). Title map (87,180): signal renders, sorted
   behind trees. NEXT: stations (PaintStation/PaintTrainStation/PaintRoadStation).
   Tick 39: train stations (style 0) DONE — pkg/paint/station.go (straight: back/front platform + canopy into the
   additions list; diagonal sequences 0-3 per rotation), session gains addToPlotListTrackRoadAddition and its
   finalise; station image ids generated (TrainStationObject.h); goloco's SequenceImageOffsets match load()
   (image + 2 previews + i × style total). Title map (89,90): platform + arched canopy render. NOT YET: the
   translucent canopy glazing (needs the company colour → Colours::getTranslucent → palette map), platform cargo
   (station cargo stats), road stations, airports, docks.
   Tick 40 — industries step 1: goloco had no industry records and no IndustryObject. Parsed the GameState
   industries table (offsets/sizes taken from the compiler: GameState 0x9E744, Type2 0x9CB24, 128 × 1107 bytes,
   objectId @16) → Scenario.IndustryObjectIDs; title map: tiles use object ids 0–14 (15 types, plausible).
   NEXT: (2) IndustryObject loader (port IndustryObject::load's section walk: building parts, heights,
   animations, sequences, buildingImageIds/shadowImageIds, buildingSizeFlags, scaffolding) with a test that every
   ObjData industry parses to the end; (3) paintIndustry.
   Tick 41 — step (2) DONE: pkg/objects/industry.go (IndustryObject::load's walk: 8 string tables, part heights,
   part animations, 4 animation sequences, var_38 pairs, variations, buildings, 2+3 cargo + 4+1+1 wall headers,
   image table; image bases as load()). Test (pkg/scene): all 22 ObjData industries walk EXACTLY to their image
   table (upstream's assert). NEXT: (3) paintIndustry + ReorderIndustryObjects (the scenario's industry slot order).
   Tick 42 — step (3) DONE: pkg/paint/industry.go (paintIndustry/paintIndustryBuilding: parts, part animations or
   the animation sequence, under-construction sections with the noise-masked top, colour remap, 2×2 front-tile
   rule). Industry slot order parsed (slots 809×16 — computed from upstream getMaxObjects, which also reproduces
   goloco's building 680×128), ReorderIndustryObjects, industry element colour/var_6/progress/sequence fields.
   Title map (169,150): farm silos/barns and a refinery render among tracks, station, river. NOT YET: shadows,
   scaffolding, farm FIELDS (PaintSurface industrial branch → next), ghosts.
   Tick 43: farm fields DONE (scene: industrial surfaces from the industry object — field image per growth stage +
   display slope, tree image +19 at SurfaceTreeHeight[ds]; byte 7 = industry id). k4FD30C extracted from
   PaintSurface.cpp by gen.sh into surfacedata_gen.go (literal integer table, no hand copy). Title farm: crop rows
   render. FOUND: goloco doesn't parse SNOW COVERAGE at all — both the land and field branches draw snow
   (SnowObject image + kSnowCoverageSlopeToMask) → item 3k.
3n. [ ] **UI (Joop 2026-11-01, HIGH): 'the menus are severely misaligned and we still have a fair bit of improvised
   dialogs instead of the game style'.** Port, don't patch: windows and widgets from upstream Ui/ (Widget.cpp
   drawing and layout, WindowManager, the dropdown and the title/toolbar menus) and each improvised goloco
   dialog replaced by the port of its upstream window (Ui/Windows/*.cpp). Start: inventory every goloco
   menu/dialog → its upstream window; fix the shared layer first (widget positions, font metrics — deviations C:
   g1font width fudge, DrawStringCentered len×7), since misalignment everywhere points there.
   Tick 76 INVENTORY (goloco → upstream): toolbar + dropdowns (pkg/ui/toolbar.go, dropdown.go) → ToolbarTop.cpp +
   Dropdown.cpp; New Game (cmd/goloco openNewGameWindow) → ScenarioSelect.cpp; Options → Options.cpp (tabs);
   Load/Save (openFileWindow) → PromptBrowse.cpp; Vehicles/Stations/Towns lists (listwindows.go) → VehicleList/
   StationList/TownList.cpp; progress (progress.go) → ProgressBar.cpp; widgets (pkg/ui *_widget.go) → Widget.cpp
   + Gfx drawRectInset (3D bevels from Colours::getShade). Title screen checked: upstream-faithful (TitleLogo IS at
   0,0; menu centred, 25 px margin) except goloco's invented subtitle line. g1 font widths = upstream (not a fudge).
   JOOP 2026-11-01 (later): 'The objective overlaps, I think we need to start implementing the bottom bar and the
   New Game window is still old style.' → PRIORITY after the window chrome: (a) ToolbarBottom (player info + time
   panels; the objective belongs there / in its window, not as an overlapping overlay), (b) New Game → ScenarioSelect.cpp.
   ORDER: (1) drawRectInset + getShade + window frame/caption/button widget drawing in the 8-bit pipeline;
   (2) Dropdown.cpp (Joop named 'menus'); (3) ToolbarTop; (4) each dialog → its upstream window.
3m. [~] **Zoom 1-3 in the CPU renderer (blocks making scene mode the default).** Upstream
   (Graphics/SoftwareDrawingContext.cpp drawImagePaletteSet + getDrawImagePosArgs, DrawSpriteBMP.hpp /
   DrawSpriteRLE.hpp, 248 lines): an element with hasZoomSprites (G1 flag 1<<4) at zoom>0 recurses at half
   coordinates into image - zoomOffset (rt x/y/w/h >>1, zoom-1); noZoomDraw (1<<5) skips at zoom>0; positions are
   aligned to the zoom grid (BMP masks y, RLE does not; RLE pre-shifts by ~zoomMask), then BMP samples every
   2^zoom pixel and RLE walks runs writing x>>zoom on rows i>>zoom. goloco has the G1 flags and ZoomOffset
   (pkg/assets) but gfx8.Sprite drops them and gfx8 only does zoom 0. Steps: (a) Sprite keeps flags+zoomOffset,
   port getDrawImagePosArgs+BMP/RLE zoom with tests (zoom 0 must stay byte-identical); (b) RenderTarget zoom +
   the recursion; (c) paint session/scene at zoom (viewport scale, culling, zoom-dependent branches:
   trees/shadows zoom<=1, snow zoom<=2, gridlines/station cargo zoom 0…); (d) world.drawScene for zoom>0.
   Also missing for default: vehicles in the paint sort (with item 5), track extras + paintSupports, rotation.
3l. [x] **Keep tile element order (blocks tunnels; removes 3 APPROXIMATIONs).** Upstream paints a tile's
   elements in stored (height) order: an underground tunnel track comes BEFORE the surface, registers
   insertTunnel, and the surface's cliff edges then draw the portal (PaintSurface.cpp 1248-1301); signals and
   stations paint against `entry.prev()`; tracks/roads interleave per baseZ. goloco keeps typed slices only.
   Plan: scenario.Tile gets `Order []ElemRef{Kind, Index}` appended in parseTileElements (each case; kept through
   the surface rebuild at loader.go ~396); sim.Tile copies it (map.go ~88) and EVERY mutation keeps it in step:
   sim/map.go AddTrack (42) / AddRoad (61) insert by baseZ like TileManager::insertElement; sim/construction.go
   181/294 removals; world/world.go 1033/1043 undo pops; world/station.go 82. Then scene walks Order and
   dispatches like paintTileElements, with prev() available — signal/station/track-road APPROXIMATIONs go, and
   tunnels (insertTunnel from the track/road tunnel tables + the cliff-edge portal code) can be ported.
3k. [x] **Snow cover.** DONE tick 44 for land tiles (coverage = slope byte bits 5-7, already parsed; SnowObject loader;
   full cover → snow image, 1-4 → snow attached masked by kSnowCoverageSlopeToMask — table names extracted from
   PaintSurface.cpp and resolved by the compiler via gen.sh). Bottleneck Blues renders snow. NOT YET: industry
   fields' snow (farmTilesDrawAboveSnow), the SurfaceElement::update snow step (coverage drifting to target). Parse SurfaceElement snowCoverage (SurfaceElement.h), load the SnowObject, port the
   snow branch of paintSurface (coverage 5 → snow image; 1-4 → surface + snow image masked by
   kSnowCoverageSlopeToMask) and the SurfaceElement::update snow step (coverage moving toward the target).
3f. [ ] **Standing (Joop: 'go through all of it and fix by bits and pieces'):** walk the paint code element by
   element against upstream Paint/: surface (slopes, corners, cliffs, water, snow), track, road (incl. slopes),
   stations, signals, bridges, tunnels, trees, buildings, industries, walls, vehicles — one element per tick:
   read the upstream painter, diff goloco's, capture with --diag, fix, note it here.
3g. [x] Music stutters in the browser (Joop): find where audio is decoded/streamed (pkg/audio, oto in js/wasm);
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
- (answered 2026-11-01, Joop: 'the top bar mouseover text menus are misaligned', 'text doesn't fit boxes') UI: which menus look most misaligned?
- Bridges (scene mode, title map river crossing near tile 90,99): do they look like the original? A reference screenshot from OpenLoco/Locomotion at the same spot would let me check pixel-exactness.
- Does the music still stutter on the test page (rebuilt tick 75: music now played by the browser's <audio> element) (rebuilt 2026-11-01 tick 63, audio now at the music's native 22050 Hz)?
- (answered 2026-11-01: Joop — 'looks a lot better and isn't as slow as I expected; use that as the baseline') Try the test page with 'Use the ported renderer' ticked: does it look right, and what frame time does the console log
  ('[Scene] CPU renderer: … per frame')? That decides how soon it can become the default.
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
- 2026-11-01 tick 17 (item 3e water d / slopes) — submerged terrain is no longer flattened and lifted (a misread of getWaterSurface); seabed slopes show through the water like the reference; slope→display table checked equal to upstream. Live build updated.
- 2026-11-01 tick 18 (item 3e) — waterline tile explained (light-facing slope, not a bug); road slope placement proven by a headless test; found water blended over buildings near the coast (draw order) → interim object-mask fix queued, then the road slope capture.
- 2026-11-01 tick 19 (item 3i step 1) — pkg/gfx8: 8-bit render target + exact blitPixel port + index decoder; unit tests for every blend op; equivalence test: 4,118/4,118 g1 sprites identical to the RGBA decoder. Chosen over an interim water mask that would still be wrong for terrain in front.
- 2026-11-01 tick 20 (item 3i step 2) — pkg/paint: core of OpenLoco's paint session (structs, quadrants, arrangeStructs, draw order) ported line for line, 4 tests pass headless.
- 2026-11-01 tick 21 (item 3i step 3a) — cmd/render8: headless CPU render of land surfaces via gfx8+paint; title-map region matches goloco's diag structure; cliffs next, then an aligned pixel diff.
- 2026-11-01 tick 22 (item 3i) — land image formula verified against LandObject::load; gfx8 masked draw (bitwise AND as upstream) + paint support, tested; cliffs next.
- 2026-11-01 tick 23 (item 3i) — cliff edges ported (pkg/paint/surface.go) and drawn by render8; gaps filled.
- 2026-11-01 tick 24 (item 3i) — water ported into the CPU renderer (exact dst blend in paint order); at the town-by-the-sea spot the water no longer spills onto the land in front.
- 2026-11-01 tick 25 (item 3i) — trees in the CPU renderer; transition image needs the noise mask (tree sprites are BMP, so upstream applies it) → next.
- 2026-11-01 tick 26 (item 3i) — noise-mask blit ported (sprite-width stride quirk kept, RLE excluded) with a test; trees in transition dither like the original. (Paused briefly for Joop's video-subtitle task: drop folder ~/Projects/videosubs, Whisper large-v3 pre-downloading.)
- 2026-11-01 tick 27 (item 3i) — tree season sprite remap checked equal to upstream; buildings next.
- 2026-11-01 tick 28 (item 3i) — buildings in the CPU renderer (variation, parts, animations, colours, 2×2, elevators); the town-by-the-sea renders with water behind it, as the original. Roads next.
- 2026-11-01 tick 29 (item 3i) — road painter surveyed (code + ~1,800 lines of data headers); plan: generate the tables from upstream headers instead of goloco's hand-copied road_sprites.go, then port paintRoad.
- 2026-11-01 tick 30 (item 3i roads a) — road paint tables generated from upstream via the C++ compiler; found and fixed a swapped left-curve piece in goloco's hand table; slope sprites verified equal (bug is in placement).
- 2026-11-01 tick 31 (item 3i roads b-d) — road painter + track/road priority ordering ported; town streets and junctions render in the CPU renderer.
- 2026-11-01 tick 32 (item 3i roads e) — slope comparison: tables and placement fine; the current renderer's unsorted per-tile drawing is the likely cause; found two more invented behaviours in its road merge. New item 3j: switch the live renderer to gfx8+paint.
- 2026-11-01 tick 33 (item 3j step 1) — pkg/scene: the CPU renderer as a reusable package reading sim.Map; sim.Map.LoadScenario; refactor proven pixel-identical on both reference views.
- 2026-11-01 tick 34 (item 3j step 2) — the game can draw through the ported CPU renderer (GOLOCO_SCENE=1); title renders correctly; walls/tracks/stations still to port. (Videos: audio copied from Crostini, Whisper transcription running in a background agent.)
- 2026-11-01 tick 35 (item 3j) — walls ported (image ids generated); title-map walls render in the CPU renderer.
- 2026-11-01 tick 36 (item 3j) — track paint table generated from upstream; goloco's hand track table verified identical (0 diffs). (Video SRT: OCR of the burned-in captions planned as a second source, per Joop.)
- 2026-11-01 tick 37 (item 3j) — tracks in the CPU renderer (priority layers, sorted with roads per height group); the title railway renders.
- 2026-11-01 tick 38 (item 3j) — signals ported; fixed two parse gaps (right-side signal object, track flag_6); signal offsets generated from upstream TrackData.
- 2026-11-01 tick 39 (item 3j) — train stations (style 0) in the CPU renderer; additions priority list ported; station image ids generated.
- 2026-11-01 tick 40 (item 3j) — industries step 1: GameState industry records parsed (offsets checked with the compiler). Video: Whisper at ~3.5 min of 71 after 8 min (slow on CPU); caption strips 7,356/8,560.
- 2026-11-01 tick 41 (item 3j) — IndustryObject loader ported; all 22 industry objects walk exactly to their image tables. (OCR of the caption strips running in a second agent.)
- 2026-11-01 tick 42 (item 3j) — industries render in the CPU renderer (farm, refinery); slot order and element fields parsed; farm fields next.
- 2026-11-01 tick 43 (item 3j) — farm fields render (industrial surface branch); k4FD30C extracted from upstream source; snow cover found missing → 3k.
- 2026-11-01 tick 44 (item 3k) — snow cover on land in the CPU renderer; SnowObject loader; mask table generated (names from the .cpp, values from the compiler).
- 2026-11-01 tick 45 (item 3j) — test page: checkbox 'Use the ported renderer' (passes GOLOCO_SCENE=1, remembered per browser); scene mode logs its average CPU frame time every 120 frames to the console. Live at kiefte.eu/goloco for Joop to compare.
- 2026-11-01 tick 46 (item 3j perf) — scene benchmark added; blit clipping + transparent fast path: 28.7 → 17.6 ms/frame, byte-identical on all three reference views.
- 2026-11-01 tick 47 (item 3j perf) — paint/scene buffers reused across frames (2.8 MB → 190 KB per frame), byte-identical; next: per-element allocations.
- 2026-11-01 tick 48 (item 3j perf) — palette maps cached per colour, part-animation arrays per object: 190 → 82 KB/frame, byte-identical; allocation count (~25.6k) comes from elsewhere — profile next.
- 2026-11-01 tick 49 (item 3j perf) — the 25.6k allocations/frame were the 4 neighbour Surface structs per tile escaping to the heap; renderer scratch fixes it: 1 alloc, 160 B per warm frame, 14.7 ms (CPU shared with Whisper). TestRenderSteadyStateAllocs guards it. Byte-identical on all three views.
- 2026-11-01 tick 50 (item 3j) — road stations (PaintRoadStation style 0, imageOffsets[1] as upstream; constants checked against RoadStationObject.h); title map (71,57) renders a truck station at a road end. Same PARTIAL set as train stations.
- 2026-11-01 tick 51 (item 3j) — airports: AirportObject loader (layout from the compiler: layout_gen.go via gen.sh; slots 373..380 from getMaxObjects), PaintAirport ported incl. the unchecked numFrames-1 quirk; title map (124,332) airport renders. PARTIAL: shadows, company colour. Video: speakers are in Malay with burned-in English — the OCR captions are the primary source; Whisper (forced English) only cross-checks narration.
- 2026-11-01 tick 52 (item 3j) — docks: DockObject loader (layout from the compiler), PaintDocks; airport/dock share paintStationBuilding (airport render byte-identical after the refactor). SHIPST1 parses (2 parts, 1 variation). NO VISUAL CHECK YET: no scenario or title map has a dock — check once a save with one exists (or after the construction tools can place one).
- 2026-11-01 tick 53 (item 3j) — company colours: parsed from the GameState companies (0xB96C, 36776/36296-byte records, mainColours@0x1A — all from the compiler; neutral=grey as updateColours); applied to track, road (incl. the merge image), train stations, airports, docks by the element's owner as upstream. Title station canopy now in its company colour. PARTIAL: road stations (upstream uses station->owner: station table not parsed), translucent canopy layer.
- 2026-11-01 tick 54 (item 3j) — station table owners parsed (stations 0xC10C4/0xBF4A4, 978-byte records, owner@0x28, from the compiler; all 515 title station tiles agree with their element owner); road stations take station->owner's colour (roof trim now company pink).
- 2026-11-01 tick 55 (item 3j) — FIX: GetPaletteMap assumed palette maps at 2170+colour; true for colours 0..31 only — ExtColour 32..42 are g1 417..427 and later ones are 11 lower. Now from kDefaultPaletteToG1Offset (generated, pkg/assets/palettedata_gen.go). Translucent/shadow colour tables generated too (pkg/paint/colourdata_gen.go). No visible change today (all callers pass 0..31; render8 byte-identical) — needed for the translucent canopy (next) and shadows.
- 2026-11-01 tick 56 (item 3j) — translucent station canopies (train straight + diagonal NE1/SE3, road stations): getTranslucent(companyColour) or unk2E when not recolourable; station flag offsets now from the compiler (agreed with the hand values). Title station's glass canopy renders tinted.
- 2026-11-01 tick 57 (item 3j) — shadows: buildings ((parts+elevators+variation)*4, getShadow(orange)), industries (getShadow(own colour), primary → src+dst blend), airports, docks; added after resetLastPS as upstream (before the 2×2 front test). Original data: only BLDPYLON/BLDTRANS have shadows (Bottleneck Blues 246,54: shadow patch under trees; reference views unchanged, as they contain no shadowed objects). gfx8 blend: out-of-range src row now leaves dst (upstream asserts; Go would panic).
- 2026-11-01 tick 58 (item 3l, design) — tunnels need the tile's element order (tunnel track precedes the surface); goloco loses it in typed slices. Plan + every mutation site recorded under 3l; no code yet (switched to Joop's video request).
- 2026-11-01 tick 59 (item 3l step 1) — Tile.Order (ElemRef kind+index, stored order) parsed and copied into sim; insertRef (TileManager::insertElement placement: after all baseZ <= new) / removeRef kept in step at every mutation (AddTrack/AddRoad, removals, undo pops, station placement). Tests: title map Order complete on all tiles (270 tiles have elements below the surface — tunnels); sim mutations keep it sorted and renumbered. NEXT: scene walks Order (step 2).
- 2026-11-01 tick 60 (item 3l step 2) — scene paints each tile's elements in stored order (paintTileElements: paintElement switch + paintTileElementsEndLoop finalise at the last element / baseZ change); signals and train/road stations use entry.prev(). All three APPROXIMATION notes in scene.go removed (fixed, not deleted). Reference views (title, town, station, snow) byte-identical. NEXT (3l step 3): tunnels — insertTunnel from track/road + cliff-edge portals.
- 2026-11-01 tick 61 (item 3l step 3, research) — tunnels: TunnelHeights already in the generated track/road tables; insertTunnel/insertTunnels (Paint.cpp 442-472) per edge with dedupe of the last entry; trackObj->tunnel / roadObj->tunnel = res->id of the embedded tunnel ObjectHeader (TrackObject.cpp: after compatible+mods+signals; RoadObject.cpp: after compatible+mods) → resolve by that header's name. Needs: TunnelObject loader (string table + image table), header capture in track.go/road.go, session tunnel lists reset per tile, portal code in PaintCliffEdges. No code yet (video work for Joop).
- 2026-11-01 tick 62 (item 3l step 3, DONE) — tunnels: TunnelObject loader; track/road loaders keep the embedded tunnel header name (resolved by name = upstream's res->id object); session tunnel lists (resetTunnels per tile, insertTunnels with last-entry dedupe, 0xFF terminator); paintSurfaceCliffEdgeImpl portal branch (edge 0: image+0/1, edge 1: image+2/3, skip 2 heights; edges 2/3 skip only). Title map: portals at real mouths (101,280; 107,232 road tunnels) where rock covered the road before. Zero-alloc guard holds (tunnel image cached per name).
- 2026-11-01 tick 63 (item 3g) — music stutter: the audio context ran at 44100 while Locomotion's music is 22050 Hz stereo, so Ebitengine resampled while playing — measured 64 ms CPU per second of music natively (13.3 s for one 208 s track) vs 5 ms for the whole track at 22050. Context now 22050 (native, as OpenLoco); 250 ms player buffer in the browser. Test page rebuilt — Joop to confirm by ear.
- 2026-11-01 tick 64 (item 3j bridges, step 1) — bridge data audit: element HasBridge/BridgeID, slot order (305..312) and BridgeObject loader already existed; loader offsets now come from the compiler (layout_gen.go — all matched the hand values). NEXT: port PaintBridge.cpp (1418 lines; BridgeEntry set in paintTrackPPStandard/paintRoadPCP, painted in paintTileElementsEndLoop) — several ticks.
- 2026-11-01 tick 65 (item 3j bridges, step 2) — all PaintBridge.cpp tables generated (gen_bridge.cpp includes the .cpp itself; unused painting code dropped by --gc-sections, so the tables are exactly upstream's): deck/span image sets, support-edge images and offsets per slope, slope→display slope, bridge shadows, twos→ones. NEXT: port paintBridge + BridgeEntry from track/road (session state) + call in the end-loop.
- 2026-11-01 tick 66 (item 3j bridges, step 3) — BridgeEntry on the session: set by track (both PP paths) and road PCP with edges/quarters ORed, owner colour kept; taken and cleared in the end-loop after the finalisers. paintBridge is an explicit PARTIAL no-op hook (marked in scene.go). Output byte-identical. NEXT: port paintBridge body.
- 2026-11-01 tick 67 (item 3j bridges, step 4) — support-height state paintBridge needs: general support (reset 0xFFFF per tile; surface sets baseHeight+slope; 2×2 buildings/industries/airports and docks set 0xFFFF on every footprint tile; after a bridge, lowered to the bridge height with slope 0x20), surface height/slope and waterHeight2 for the shadow. Byte-identical (nothing reads it yet). NEXT: paintBridge body (deck/walls/edges part first, then NE/SW spans + supports, then shadow).
- 2026-11-01 tick 68 (item 3j bridges, step 5) — paintBridge part 1 (pkg/paint/bridge.go): early returns, support-length test (ceil2 general support), deck + side walls + deck-wall edges with owner colour; every Bridge::ImageIds name PaintBridge.cpp uses generated (bridgeids_gen.go, 123). Title river bridge (90,99): curved pieces show brick deck+walls; straight span pieces (subType 0) and supports not yet — looks unfinished in scene mode until step 6. NEXT: quarter cases + paintBridgeNE/SW + supports, then shadow.
- 2026-11-01 tick 69 (item 3j bridges, step 6a) — segment support state (needed by paintBridgeNE/SW pillars): 9 segments (setSegmentsSupportHeight, slope kept on 0xFFFF), reset per tile, surface sets kSlopeToSegmentsSupportHeights (generated by including PaintSurface.cpp), 2×2 objects block all, track/road/stations OR into the pending mask (525CF8) flushed at the end-loop after the bridge block; waterHeight (not reset per tile, as upstream). Byte-identical. NEXT: paintBridgeNE/SW + paintSupportNE/SW/Diagonal + span walls.
- 2026-11-01 tick 70 (item 3j bridges, step 6b) — paintBridge nearly complete: quarter cases 1/2/4, NE/SW spans as one function parametrised by direction (segments, slope tables, edge images, box orientation — paintBridgeNE/SW + paintSupportNE/SW), span tile from the tile position masked by spanLength, subType-0 span walls, ground shadow (unk32). Title river bridge: brick arched piers on the riverbed + company-pink steel span. PARTIAL: quarter case 8 (front supports). UNVERIFIED against an OpenLoco reference render (plausible by eye). Question for Joop added.
- 2026-11-01 tick 71 (item 3j bridges, DONE) — quarter case 8 (paintFlatSingleQuarterSupportFront + paintSupportDiagonal) ported literally, incl. upstream's re-test of unkHeight and its swapped Lhs/Rhs image names. Bridge PARTIAL notes gone; the end-loop note now names only paintSupports (track/road addition supports) as missing.
- 2026-11-01 tick 72 (assessment) — what blocks scene mode as default: zoom 1-3 (new item 3m, plan + upstream refs), vehicles in the sort (item 5), track extras/supports, rotation. Vehicles currently drawn by the old post-pass over the scene.
- 2026-11-01 tick 73 (item 3m step a) — gfx8 zoom: getDrawImagePosArgs + drawBMPSprite/drawRLESprite for zoom 1-3 ported (zoom.go); Sprite keeps hasZoomSprites/noZoomDraw/zoomOffset and the RLE runs. Property test: 2^z-upscaled sprite at zoom z == original at zoom 0 (BMP and RLE, 300 random cases each); noZoomDraw skipped. Joop reported misaligned menus + improvised dialogs → new item 3n (HIGH), next.
- 2026-11-01 tick 74 (item 3j / 3n) — CPU renderer is now the DEFAULT (Joop: 'use that as the baseline'); GOLOCO_SCENE=0 or the test page's 'Use the old renderer' box selects the old one; zoom>0 still falls back to the old renderer (3m). Test page redeployed. 3n: g1 font widthFudge is upstream's own table (deviations C entry was wrong); title screen captured headless — logo top-left, menu buttons without their window, hand-placed texts; upstream positions read (TitleLogo 298×170, TitleMenu centred at height-kWH-25, TitleVersion (8, h-kHeight), TitleOptions 60×15 top-right, TitleExit 40×28 bottom-right) → next: port the title windows.
- 2026-11-01 tick 75 (Joop: 'the biggest issue is the sound … separate goroutine for desktop, piggyback on browser features for the web') — js/wasm music now plays through an HTMLAudioElement (Blob URL of the WAV; the browser decodes on its own audio thread, so slow wasm frames can't starve it); desktop keeps Ebitengine's player, which already runs on its own goroutine/thread (and since tick 63 at the native 22050 Hz). Headless check: title track css5.dat (12.9 MB) plays via <audio>, no autoplay refusal. Test page redeployed. Title windows (3n) paused mid-read: upstream TitleLogo/TitleMenu sizes noted in tick 74.
- 2026-11-01 tick 76 (item 3n) — UI inventory: every improvised goloco window/menu mapped to its upstream Ui/Windows file; title screen found upstream-faithful (logo top-left is upstream) apart from the invented subtitle; order set: widget drawing (drawRectInset/getShade) → Dropdown → ToolbarTop → dialogs.
- 2026-11-01 tick 77 (item 3n step 1) — gfx8/rect.go: drawRectImpl (fill + translucent remap), fillRectInset/drawRectInset (all flag paths, translucent and opaque), getShade via PaletteIndex::primaryRemap (generated from the compiler: 7,8,9,246..254). Visual: raised/sunken/light bevels in grey, brown, orange, blue match Locomotion's window style; unit test pins edge shades. NEXT: widget drawing (frame, caption, button, panel) on top, then Dropdown.cpp; the UI then draws into the 8-bit frame like upstream.
- 2026-11-01 Joop: toolbar dropdown menus misaligned, text doesn't fit. Cause found in pkg/ui/dropdown.go: width = len(text)*6 (guessed glyph width), itemHeight 16 (invented), flat ebitenutil colours. NEXT TICK: port Ui/Dropdown.cpp (measured string widths, upstream item height/padding, inset bevels via gfx8 rect.go) and the toolbar's dropdown placement (ToolbarTop.cpp).
- 2026-11-01 tick 78 (item 3n step 2) — Dropdown.cpp ported for text dropdowns: width = widest item in medium bold + 3 (+3 +1 window), rows 10 px, text at (2, 1+row·10), highlight = unk2E translucent at (2, 2+row·10), separators shade 3/7, hit test minus the 2 px top padding; frame = Wt3 panel via gfx8.FillRectInset. Headless shot (GOLOCO_DIAG_DROPDOWN=1): File menu items fit their box. PARTIAL: translucent toolbar dropdown colour (needs the UI in the 8-bit frame), off-screen flip placement, text colour via the text renderer, inset text for disabled items. Test page redeployed.
- 2026-11-01 tick 79 (item 3n) — 'mouseover text' may mean the toolbar TOOLTIPS: pkg/ui/toolbar.go:208-211 draws them as bare DrawText with no window. NEXT: port the upstream tooltip window (Ui/…Tooltip…: frame, wrapped text, placement below the cursor). (Tick otherwise spent finishing Joop's video SRTs: en + eo uploaded to Crostini.)
- 2026-11-01 tick 80 (item 3n step 3) — tooltips: ToolTip.cpp ported (medium bold, max 196 px wrapped, width+3 × lines·10+4, 26 px below the cursor / flipped up, clamped; body unk2D + glass(unk74+skin tooltipColour), unk2E edges and rounded corners, black centred text). New ui.draw8: reads a screen region back to palette indices so gfx8 translucent remaps work on the RGBA UI (exact on palette colours) — also unblocks translucent dropdowns. Headless shot (GOLOCO_DIAG_TOOLTIP=1) OK. PARTIAL: wrapString not ported (space-based wrap). Test page redeployed.
- 2026-11-01 tick 81 (item 3n step 4) — toolbar dropdowns translucent as upstream: showBelow colour = kDropdownColourTable[button's window colour] (generated; the skin's topToolbar{Primary..Quaternary}Colour by the button's colour group), drawn through draw8 with fillRectInset's translucent branch, unk2E highlight and getTranslucent+1/+2 separators. Headless shot: blue glass menu over the title. Test page redeployed. Still PARTIAL: text colour, disabled inset text, off-screen flip.
- 2026-11-01 tick 82 (item 3n) — dropdown text colours per TextRenderer: plain colour = getShade(c, 9) (so the earlier 'PARTIAL: shade 9 stands in' was in fact upstream — note replaced by the rule), highlighted = Colour::white's shade 9 (was RGB white), disabled = inset text (shade 6, then shade 4 at +1,+1). Headless shot shows the embossed disabled item. Remaining dropdown PARTIAL: off-screen flip placement.
- 2026-11-01 tick 83 (item 3n) — dropdown placement ported (showText: below, else above, else beside at the top; x clamped) in new pkg/uilayout (Ebitengine-free, so testable here — pkg/ui tests can't start without a display: no Xvfb, and Ebitengine's js build needs a DOM); 5-case test passes. Toolbar dropdowns use it with the screen size. All dropdown PARTIALs resolved. NEXT (3n): window frame + caption widgets (Widget.cpp), then the dialogs.
- 2026-11-01 tick 84 (item 3n) — window chrome: FrameWidget (background style: frame_background_image(_alt) 2322/2323 recoloured, tiled from the right, clipped to 41 px, right shadow shade 3/1) and CaptionWidget drawSimple whiteText (centred on 2+(w-18)/2, clip w-26, white + outline black0); outlined text = glyph pixels 0x02/0x03 in black0 (new DrawTextOutlined). Headless shot OK (GOLOCO_DIAG_WINDOW=1). PARTIAL: resize handle, clipString '...'. Joop: menus improved; next priorities bottom bar (objective overlaps) and New Game → ScenarioSelect.
- 2026-11-01 tick 85 (item 3n, Joop's bottom-bar request) — TimePanel.cpp ported: 140×27 translucent window bottom-right in the skin's timeToolbarColour (unk34 dim, Wt3 panels, inset border), date 'DMY full' centred (white on hover), four speed buttons (active image variant, recoloured), map globe; speed buttons clickable (pause toggles; normal ×1, fast ×2, extra ×4). The objective overlay under the top toolbar is GONE: it is shown as the date's tooltip, as upstream (tooltip_daymonthyear_challenge). Layout + hit tests + date in pkg/uilayout (tested). The left part of the old grey strip remains, marked IMPROVISED, until PlayerInfoPanel.cpp (next). PARTIAL: Paused blink, map/chat dropdown, button hover draw, exact challenge string. Test page redeployed.
- 2026-11-01 tick 86 (item 3n, bottom bar) — PlayerInfoPanel.cpp ported (translucent bottom-left window in playerInfoToolbarColour, cash {CURRENCY48} centred at frame.w/2+12, red when negative, 'Bankrupt!'); the improvised grey status strip is GONE (both bottom windows are ports now). PARTIAL: owner face (no competitor/company record yet), performance index row (not computed — game logic), currency object (£ prefix fixed). Test page redeployed. NEXT: New Game → ScenarioSelect.cpp.
- 2026-11-01 tick 87 (item 3n, Joop's New Game request) — ScenarioSelect.cpp ported (pkg/ui/scenarioselect.go) and the improvised 285-line New Game window removed: 610×412 centred, black/mutedSeaGreen, frame + extended solid band, caption, close button, five wide_tab tabs (active +1) with centred captions, list cleared to shade 4 with 24 px rows (selected: unk30 + window-colour text), details pane (name, inset preview outline + 128×128 preview + compass / random-map watermark, description, Challenge, Start date, Competing companies). Click row selects, second click starts; wheel scrolls. Headless shot (GOLOCO_DIAG_NEWGAME=1) matches the original layout. PARTIAL: scrollbar widget, text_palette colours, wrapString, completed tick, delayed-competitor line. Test page redeployed.
- 2026-11-01 tick 88 (item 3n) — ScrollView ported for drawing: inset frame (borderInset|fillDarker), 1 px content inset, vertical bar when content > widget (hatched track via crossHatching, the four track lines, up/down buttons with glyphs 0xA0/0xAA, thumb from calculateThumbSizeAndPosition incl. uint16 truncation, min 20 — tested in uilayout). ScenarioSelect's list now sits in it (row positions 1 px in, as upstream). No shipped category exceeds 13 rows, so the bar shows only in the GOLOCO_DIAG_NEWGAME=many shot — renders correctly. PARTIAL: scrollbar mouse interaction. Test page redeployed.
- 2026-11-01 tick 89 (item 3n) — Load window = PromptBrowseWindow.cpp port (load mode): 500×380 centred, black/mutedSeaGreen, caption, close, panel, parent-folder button, 'Folder:' line, ScrollView list (11 px rows, folders first with icon_folder, then .sv5/.sc5/.goloco; select, second click opens/loads), selected name in the right pane. The improvised 285-line file window is removed. Headless shot OK (GOLOCO_DIAG_BROWSE=1). PARTIAL: save mode (text box + OK), save/landscape preview pane, folder '...' shortening, scrollbar mouse. Test page redeployed. (Joop also handed a secondary task: Stelo Ebitengine remake — research agent running.)
- 2026-11-01 tick 90 (item 3n, research) — Options: goloco's improvised window has sound/music on/off; upstream counterpart = Options.cpp AudioTab (366 wide; GroupBox 'Sound' at y49 with device dropdown + 'Play title music' checkbox; GroupBox 'Volume' with 6 sliders in 2 columns, 26 px rows) inside the 7-tab common frame (makeCommonWidgets, tab images from the interface skin tab_colour_scheme frames). NEXT: port the common frame + Audio tab (GroupBox, Checkbox, Slider widgets), other tabs later. (Tick time also went to Joop's Stelo task: original 2018 builds found in ~/Projects/stelo; research agent running.)
- 2026-11-01 tick 91 (item 3n) — Options window = Options.cpp common frame + Audio tab (366×201, skin windowTitlebarColour/windowOptionsColour; caption, close, panel, seven 31 px tabs with their icons, GroupBox 'Sound' (device field, 'Play title screen music' checkbox) and 'Volume' (six sliders: track/thumb images at 10 + pct·80/100)). Wired to goloco: checkbox = music on/off, Music slider = music volume (click). Improvised options window removed. Headless shot matches the original. PARTIAL: other tabs, tab animation, device dropdown, the five channels goloco lacks (inert at 100%), slider dragging. Test page redeployed.
- 2026-11-01 tick 92 (item 3n) — Towns window = TownList.cpp list tab (600×197 at (sw-600, 30), skin titlebar/window colours, four tabs, Name/Type/Population/Stations headers, ScrollView of 10 px rows cleared to shade 3, unk30 selection, 'N towns' status). Town numStations parsed (S5::Town 0x1A6, from the compiler). Headless shot on the title map: 32 towns, scrollbar, correct columns. FOUND: town NAMES are goloco placeholders 'Town N' — the town name StringIds (town-names object / user strings) are not resolved → deviations. PARTIAL: other tabs, tab icons, sorting, resize. Test page redeployed.
- 2026-11-01 tick 93 (deviation fix) — town names: GameState userStrings parsed (2048×32 at 0x432A44 / Type2 0x430E24, from the compiler); Town::name 0x8000+ resolves through them (title map: Chicago, Detroit, Toledo, Milwaukee, Grand Rapids…). Town list shows real names; deviations.md entry marked FIXED (placeholder kept only for non-user StringIds). Test page redeployed.
- 2026-11-01 tick 94 (item 3n prep) — station names: S5::Station name (0x00) and town (0x2C) parsed; names formatted as formatStringImpl does (kTownNamesStart 0x9EE7 + template id → en-GB template generated from en-GB.yml with {STRINGID} = the town's name; kTownNamesEnd → town name; user strings for renamed stations). Title map: Columbus, Kansas City South, St Louis Transfer, Greensbro North. The (still improvised) Stations window shows them. NEXT: StationList.cpp window port (needs cargo waiting/accepted per station: S5::Station cargoStats).
- 2026-11-01 tick 95 (item 3n) — Stations window = StationList.cpp (shared table-list window with the town list): All Stations, five tabs, Name/Status/Total waiting/Accepts, status = getStatusString ('N tonnes of X waiting' joined ' + ', or 'Nothing waiting'), units total, accepted cargo names; cargoStats parsed (32×13 at 0x2E, from the compiler). FOUND+FIXED: object strings carry control code 123 (int32_grouped) as a raw byte — now formatted (the raw byte had also forced the TTF fallback font, drawing 'ghost' lines). Station name templates regenerated as the contiguous id range 281..327 (Upper/Lower were missed by name). PARTIAL: company name in the caption, transport icons, per-mode tabs, sorting. Test page redeployed.
- 2026-11-01 tick 96 (item 3n) — title version text = TitleVersion.cpp (white outline, version + platform at height-30); invented subtitle and fps counter removed. FOUND+FIXED (big): the game never reordered industry/airport/dock objects (render8 and the bench did), so every title-map industry and farm field drew another industry's images (magenta unremapped buildings, black field tiles) — pre-existing since the CPU renderer became default. Both reorder lists in main.go replaced by one reorderScenarioObjects covering all 16 types (title path also lacked cargo). Headless title verified; test page redeployed. NEXT: VehicleList.cpp (550×213, six tabs, rows with drawTrainInline).
- 2026-11-01 tick 97 (item 3n) — Vehicles window = VehicleList.cpp: 550×213, six type tabs (clickable, refilter), Name/Monthly Profit/Age/Reliability, row_heights per tab (28/28/28/28/48/36), outlined WINDOW_2 text, red for a loss, solid shade-0 hover row. Data: head name+ordinal ('Train 2'; en-GB StringIds 4..9 or user string), veh2 profit[4] sum/4 and reliability, veh1 dayCreated vs GameState currentDay (offsets from the compiler). Table window generalised (size, row height, list rect, outline, cell colour, tabs). PARTIAL: drawTrainInline, getStatus after the name, owner filter (all companies listed), caption company, filter dropdowns, sorting. All three list windows now ported.
- 2026-11-01 tick 98 (item 3n) — Vehicle and station lists open for CompanyManager::getControllingId() (GameState playerCompanies[0] @0x24, from the compiler) and show only its vehicles/stations, caption '{company} - Trains' / '- All Stations' (company name StringId @Company+0; user string, or 'New Company' for StringId 2). Title map: company 0 is the unnamed 'New Company' with nothing, so both lists are empty there, as upstream. Fixed: hover row drawn past the last row. Test page redeployed.
- 2026-11-01 tick 99 (item 3n) — ProgressBar.cpp ported to the 8-bit UI (pkg/ui/progressbar.go): 350×47 centred (y≥28), lighter black frame, white 'Loading...' caption (StringIds::loading), black panel, clipRenderTarget(2,17,w-5,h-19), track 2330 + train frame (value/4)%4 at value-255 in both styles. New gfx8.TwoColourMap = PaletteMap::getForImage with a secondary colour (recolour2: secondary remap 0xCA..0xD5 from the second colour's primary entries), unit-tested. Replaces the improvised hand-coloured window (and its invented % line). Still NOT REACHED: saves load within one frame (note kept). Headless check via GOLOCO_DIAG_PROGRESS. NEXT 3n: Options' other tabs, PromptBrowse save mode.
- 2026-11-01 tick 100 (item 3n) — Options tabs switch (Common::tab, switchTab keeps the position): each tab's kWindowSize and caption (Display 400×151, Rendering 400×218, Regional 366×167, Controls 366×114, Company 420×187, Misc 420×266), current tab drawn raised. Controls tab ported: three checkboxes at (10, 49+15i) and the Customize Keys button (26,94,160×12; Label centred x+(w-1)/2 - width/2); edge scrolling wired to Config::edgeScrolling (default on), zoomToCursor/invert kept in the window only (goloco has neither). PARTIAL: Display/Rendering/Regional/Company/Misc contents. Headless check (GOLOCO_DIAG_OPTIONS=controls).
