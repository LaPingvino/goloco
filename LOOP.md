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
3j. [~] **Switch the live renderer to gfx8+paint.** World.Draw builds a paint session from the visible tiles with
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
