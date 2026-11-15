# Port plans for the big items

Written 2026-11-14 so the agents doing these ports start with the map instead of exploring. Each plan names
upstream's files (~/Projects/OpenLoco, 09d74ee), what goloco already has, an order of work in steps that each end
green and committed, and how to verify. Re-check a plan's facts briefly before starting: the code moves.

General rules (CLAUDE.md has the rest):
- Port, don't approximate. Each step leaves `scripts/test-all.sh` and `scripts/playthrough/run.sh` green
  (private TMPDIR, sandbox off), with a commit.
- Compare against the real OpenLoco with scripts/olref, using your own OLREF_DISPLAY and OLREF_HOME.
- Never touch the user's openloco.yml; point XDG_CONFIG_HOME at a scratch copy.
- Budget: these are week-sized jobs. Stop at a green, committed step, and write what is left in LOOP.md.

## A. The map window (LOOP N34 showMap) — Ui/Windows/MapWindow.cpp, 2504 lines

goloco has none. The toolbar's map entries and the `showMap` shortcut (M) do nothing, and nothing in pkg/ or cmd/
draws a minimap.

Upstream's structure:
- **Pixel buffer:** `_mapPixels` (two 768×768 frames, flipped each update) is filled a few rows per tick:
  - `setMapPixelsOverall` (l. 258): surface colours by height and land type, water, track, road, buildings, trees;
  - `setMapPixelsVehicles` (l. 447), `setMapPixelsIndustries` (l. 549), `setMapPixelsRoutes` (l. 671);
  - the ownership tab (companies' track, road and stations).
- **Tabs:** overall, vehicles, industries, routes, ownership. Each has a graph key on the right (`drawGraphKey*`,
  l. 1337–1526) whose rows filter what is drawn.
- **drawScroll (l. 2169):** the buffer rotated 45° for the view's rotation, then vehicles (`drawVehiclesOnMap`),
  routes, town names (`drawTownNames`) and the main view's rectangle (`drawViewportPosition`).
- **Input:** a click centres the main view; scroll and resize; the window opens centred on the main view
  (`centerOnViewPoint`, also called by ToolbarTop's rotate).

Order of work:
1. Window shell: open/close from the toolbar's Map entry and the M key, frame, the five tabs, a resizable scroll
   view, the overall tab's pixels filled all at once (correct first, incremental later). Test the pixel colour of a
   few known tiles against upstream's table.
2. Upstream's incremental fill: rows per tick, two buffers.
3. Overlays: the view rectangle, vehicles, town names; a click and drag centre the main view.
4. The other tabs' fills and their graph keys with filtering.
5. Compare with OpenLoco's map window on the title game: tabs, colours, overlays; the rotation in all four views.

Reuse:
- goloco's scroll views (`drawVScrollViewPressed`, ui.MusicSelectionWindow as a pattern);
- surface data in sim.Map.Tiles;
- vehicle positions in sim's entity lists;
- the land objects' map colours: check LandObject for `mapPixelColour`; it may need parsing.

## B. Sounds (LOOP N9) — Audio/Audio.cpp 576, VehicleAudio.cpp 254, AmbientAudio.cpp 178, Audio/src/AudioEngine.cpp 598

goloco plays only music: pkg/audio Manager (Ebitengine audio context, music, volume, sfx volume) and pkg/jukebox.
There are no sound effects yet. Step 1 is done (2026-11-15): pkg/objects/sound.go `ParseSoundObject` (the 42
SND*.DAT objects: 22050 Hz 16-bit mono PCM) and pkg/audio/css.go `ReadCSS` (css1.dat: 32 samples, SoundId
clickDown…breakdown6). Next: step 2, the mixer.

Upstream's structure:
- **Samples:** `css1.dat` (Environment::PathId::css1) holds the UI samples, read by `loadSoundsFromCSS`
  (Audio.cpp:54). The vehicle and ambient sounds are SoundObjects (`getSoundObject`, Audio.cpp:348) with their own
  sample data.
- **`playSound(SoundId, pos | pan)`:** a sound at a world position is given volume and pan from the main view
  (distance, zoom). 79 call sites; the most are in:
  - ViewportInteraction (demolish);
  - Vehicle and VehicleHead (horns, crashes, splash);
  - TerraForm, MouseInput (clicks), Keyboard, the construction tabs (construct).
- **Vehicle audio:** per-vehicle running sounds on a fixed set of channels, chosen by distance to the view
  (VehicleAudio.cpp).
- **Ambient:** sea, forest, town and so on, from the tiles around the view (AmbientAudio.cpp).

Order of work:
1. A SoundObject loader (Object::load layout byte for byte, its PCM data), and the css1.dat reader. Unit-test the
   sample counts and lengths against upstream's.
2. An effects mixer on the existing audio.Context: `playSound` with volume and pan, and the sfx volume option.
   Wire the UI sounds (clicks, the construct and demolish sounds) at upstream's call sites; grep
   `Audio::playSound(` and port each. Every "PARTIAL: the … sound" note in goloco marks a call site
   (grep "sound" PARTIAL).
3. Vehicle audio channels.
4. Ambient audio.

Testing without a sound card: ALSA_CONFIG_PATH null device (CLAUDE.md). Assert in tests which sounds were
queued, through a test hook on the mixer.

## C. 2× / 4× zoom (LOOP N30) — upstream 9850f2d, 124 files

The core is ZoomLevel.hpp (negative levels: −1 = 2×, −2 = 4×), RenderTarget, the software drawing context, and
DrawSpriteBMP/RLE (scaling sprites up). The rest is mostly mechanical `zoom` → `ZoomLevel` call-site changes.
goloco's views clamp zoom to 0…3 (pkg/world/world.go ZoomIn/ZoomInAt); its renderer is pkg/scene with gfx8
sprites. Zoom-dependent things to find:
- label frames (Town/Station LabelFrame per zoom);
- vehiclesMinScale / stationNamesMinScale;
- the paint session's zoom checks (`s.Zoom <= …`);
- the money effects (zoom ≤ 1);
- map tooltips and hit tests (`<< zoom`).

Order of work:
1. gfx8: draw a sprite scaled ×2 / ×4, nearest neighbour as upstream's DrawSprite*.hpp do; golden-test against
   their pixel logic.
2. world/scene: negative zoom in the view maths. Replace the `<< zoom` shifts with a helper that handles negative
   levels, because shifts by a negative count panic or misbehave in Go.
3. Labels, hit tests, tooltips, the money effects at the new levels.
4. Compare with OpenLoco at both new levels on the title game.

## E. Gamepad and focus navigation — docs/GAMEPAD.md

The design (general, meant as a reference for other sim games too) and goloco's six steps are in docs/GAMEPAD.md.
Design agreed in outline by Joop 2026-11-15 (shoulders tab through windows, buttons per context); the details are open.

## D. Smaller follow-ups ready to pick up

- **N35 save leftovers:**
  - write last_save_path;
  - "Save this before loading?" when loading during play;
  - ~~the Options autosave widgets read the config~~ (done 2026-11-14);
  - re-pack every custom object, not only the loaded file's.
- **N7:** a refused road's reason line, the bridge and alternate-road retries, shift's downward search.
- **N6:** toolbar and panels as stack windows; drag by frame; vehicle list column widths.
- **N32:** see-through checks still not compared with OpenLoco; the unexplained dark face at 186,143 underground.
