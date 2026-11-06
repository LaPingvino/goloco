# A public goloco build on open graphics — gap analysis

Status: RESEARCH (2026-11-01, LOOP item 3d), for Joop. Nothing here ships the original Locomotion data,
and no public build happens before Joop approves a design.

## What exists

**OpenGraphics** (github.com/OpenLoco/OpenGraphics, **MIT licence**), release v0.1.8 `objects.zip` — the
same file OpenLoco's CMakeLists.txt fetches (sha256 658891d9…5013c7, checked). 152 files, 13 MB, all
object files (`OG_*.dat`) in the standard Locomotion object format, so goloco's existing object loader
reads them in principle.

| Object type | Open (OG) | Original | Note |
|---|---:|---:|---|
| vehicle | 66 | 165 | enough for every transport mode |
| cargo | 15 | 14 | complete |
| land | 7 | 5 | complete (terrain) |
| water, snow, hill shapes, cliff edges | 2 / 2 / 2 / 4 | 1 / 1 / 1 / 2 | complete |
| interface skin | 2 | 1 | complete |
| wall | 9 | 8 | complete |
| road, road station, road extra | 4 / 8 / 1 | 6 / 7 / 1 | usable |
| bridge, tunnel, street light, scaffolding | 5 / 3 / 2 / 2 | 5 / 2 / 1 / 1 | complete |
| track, track extra, train station, train signal, level crossing | 1 / 1 / 1 / 1 / 1 | 2 / 3 / 5 / 10 / 4 | one of each — enough to play |
| building | 1 | 76 | **towns would be one building repeated** |
| industry | 1 | 22 | **one industry: no supply chains** |
| tree | **0** | 39 | **no trees** — v0.1.8's Tree/ folder is empty (corrected 2026-11-06; this table said 1) |
| airport, dock | 1 / 1 | 4 / 1 | usable |
| climate, region, currency, town names, steam, competitor, scenario text | 1 each (climate **0**: its folder is empty, corrected 2026-11-06) | 6 / 3 / 12 / 6 / 5 / 38 / 45 | minimal set |
| sound | 1 | 42 | **almost no sound effects** |

## Inventory: what a fully free build needs (2026-11-02)

Measured, not estimated, where marked ✓: the world renderer was traced over title.dat and the SV5 at all four
zooms (a throwaway hook on every g1 read). Sizes: S one tick, M 2–4, L a series.

### 1. Original files goloco reads today

| File | Used for | Free replacement |
|---|---|---|
| Data/g1.DAT (4,122 images) | palette, colour remaps, ground masks, UI art, fonts | sections 2–3 below |
| ObjData/*.DAT | every object | OpenGraphics (MIT, 152 objects) — thin, section 4 |
| Data/title.dat | the title scene | a generated map, section 5 |
| Scenarios/*.SC5, saves | scenarios | generated / hand-made OG-only scenarios, section 5 |
| Data/css*.dat | music | open music or none, section 6 |
| sound objects (41 of 42 original) | effects | section 6 |
| data/language/*.yml | UI text | already free: OpenLoco's own (MIT) |

### 2. What the world renderer takes from g1.DAT ✓ (430 of 4,122 images)

| g1 ids | What | Free replacement | Size |
|---|---|---|---|
| palette + 2170–2304, 417–427 (via PaletteToG1) | the 256-colour palette and ~135 colour remap tables (company colours, shadows, water, translucency) | the palette must stay index-compatible with OpenGraphics (its objects are drawn in it); the remap tables are colour-ramp substitutions that can be generated from ramp definitions | M |
| 3746–3897 + 76–303 (their zoomed copies) | ground blending masks (smooth 1–3), snow coverage masks | geometric slope masks: generate them from the slope shapes | M |
| 3726–3745 | cliff edge masks | generated, as above | S |
| 3631–3716, 3617–3625 | bridge shadows | generated shadow shapes | S |
| 3479–3485 | noise masks (tree / ground dithering) | generated noise | S |

Everything else the world shows (terrain, trees, track, roads, buildings, vehicles…) comes from the objects.

### 3. What the UI takes from g1.DAT (estimate; a runtime trace of the UI is step 1 of the plan)

- 85 named images (pkg/sprites/imageids.go): ~25 toolbar icons, ~10 tabs, construction arrows / curves,
  progress-bar art, window resize handle, rubbish bin, step arrows …
- the four bitmap fonts: 4 × 224 glyphs (small, medium, medium bold, large).
- title art: logo, title-menu globe animation, intro logos (the intro can be skipped).
- window frames / colours come from the interface skin object — OpenGraphics has one.
Replacement: draw them (Joop allows that), or take OpenTTD's (GPL-2.0 — that part of the build becomes GPL;
keep source and licence together). For fonts: an open pixel font at Locomotion's sizes, or the TTF path
goloco already has (loses the bitmap look). Size: L for drawn art; S–M for a TTF-only first version.

### 4. Objects (OpenGraphics, MIT)

Usable today for a minimal game: vehicles 66, all cargo, terrain, water, snow, walls, roads, one track /
station / signal / level crossing, bridges, tunnels, airport, dock, interface skin, region, currency, town
names. **Thin**: 1 building (towns are one building repeated), 1 industry (no supply chains), 1 tree, 1 sound.
Variety means drawing objects — best contributed to OpenGraphics, which also helps OpenLoco. Size: L, art.

### 5. Code goloco still needs for a free start

| Work | Why | Size |
|---|---|---|
| map generator: towns (5.9f3, in progress), industries (5.9f4), misc buildings (5.9f5) | a generated map is the only free title scene / scenario source | L (in the queue) |
| start without g1.DAT / Data/ (an `-open` mode: asset loader that takes the free palette, masks, UI art; no title.dat; generated title) | today startup requires them | M |
| `-open` pack in cmd/webpack (OG objects + the free assets only) and a check (CI) that no original file is read | keeps the public build clean | S |
| a free title sequence (camera path over the generated map) | title.dat holds the original one | S |
| at least one OG-only scenario (generated, saved as .SC5; objectives) | something to play | M |
| licence / credits page; a public route without basic auth (Joop runs the Caddy side) | publishing | S |

### 6. Audio

Music: none, or open tracks (licence-checked). Effects: 41 sound objects to replace or leave silent at first.

### Order (proposal)

1. UI g1 trace → the exact UI list (S). 2. Generated masks + remaps + palette check (M) — then the world draws
without g1.DAT. 3. Finish the generator (5.9f3–f5). 4. `-open` start + pack + no-original-files check (M).
5. TTF-only UI as the first free UI (S–M), drawn UI art later (L). 6. Title sequence + one scenario (M).
7. Publish (Joop). Objects and audio variety grow alongside.

## What was missing for a fully open build (2026-11-01, superseded by the inventory above)


1. **g1.DAT** — the base sprite file: UI icons and widgets, cursors, the bitmap fonts, the title logo,
   the loading/progress art, toolbar images. OpenGraphics has no replacement; OpenLoco itself still needs
   the original g1.DAT. goloco uses 85 hard-coded g1 image ids (pkg/sprites/imageids.go) plus the g1 font
   glyphs. Options: (a) draw an open g1 subset for the ids goloco uses; (b) goloco-only UI art with TTF fonts
   (goloco already falls back to TTF text) — loses the authentic look in the public build only.
2. **Title scene and scenarios** — title.dat and every .SC5 are original data. A public build needs its own
   title scene and at least one scenario made only from OG objects (Joop: "we might need a custom
   level/title screen"). goloco can't edit scenarios yet; pkg/scenario/generate.go is a 'simplified
   approximation' of map generation (deviations A) — porting upstream's MapGenerator would give a
   faithful way to make one.
3. **Music and most sounds** — the music tracks (Data/*.dat) and 41 of 42 sound objects are original.
   A public build would be silent or need open audio.
4. **Variety** — one building and one industry make towns and economies flat; contributing to
   OpenGraphics is the path that also helps OpenLoco.

## Proposed steps (design — needs Joop's yes)

1. A `-open` mode in cmd/webpack: pack only OG objects (no Data/, no original ObjData, no scenarios), and
   a startup path in goloco that works without g1.DAT (TTF text, no g1 UI sprites) — shows how far it gets.
2. A generated open scenario (after the MapGenerator port) as the title scene and first level.
3. The g1 gap: Joop (2026-11-01) allows drawing open replacements ourselves, or borrowing from OpenTTD.
   Licences decide what can be mixed: OpenGraphics is MIT; OpenTTD's OpenGFX base set is GPL-2.0, so
   borrowed sprites make that part of the build GPL — keep their source and licence with them.
4. Only then a public link — under a separate route without basic auth, built only from MIT/open sources,
   with the OpenGraphics licence and credits on the page.

## The palette (2026-11-06)

OpenGraphics ships its palette, MIT: `palettes/locopal.{act,aco,jasc.pal,ms.pal,png}`. Compared index by index
with the palette goloco reads from the original g1.DAT (locally; nothing of the original is kept): **identical over
10–244** — the range sprites are drawn in — except 155–165, within 6 RGB steps (the animated water colours); the
20 that differ are the reserved ends, 1–9 and 245–255, which g1 leaves black. So the open build takes
`locopal.jasc.pal` as its palette: the Pages workflow fetches it pinned to OpenGraphics commit d19bf5c6 and
checks its sha256 (cf558101…3bfd23); webpack packs it at /w/opengraphics/locopal.jasc.pal. Next: the game
reads it when there is no g1.DAT (an empty sprite pool with this palette), then the colour remap tables (§2).

## Generated title map, and what OpenGraphics lacks (2026-11-06)

sim.NewGeneratedMap with the editor's defaults (EditorController::init) makes the open build's landscape:
hills and water from OpenGraphics' hill shapes. Two objects are missing from v0.1.8 — **tree** and **climate**
(their folders are empty): so no trees, and no snow line (goloco sets it out of reach rather than at 0, a
stand-in for the missing climate). Both are object-drawing work, best contributed to OpenGraphics.

## Colour remap tables (2026-11-06, research)

What the game's colour maps (g1 images 2170+, read by GetPaletteMap) change — measured locally against the
original, nothing of it copied: each colour's table is identity except the 12 remap slots (primaryRemap
7–9 and 246–254), which point into one 12-shade swatch of the palette (Colour.h's PaletteIndex: swatches are
contiguous, dark 0 to light B). The colours come in pairs on one swatch: the odd one maps the slots straight
onto the swatch's 12 shades; the even one onto a compressed / clamped run of the same swatch (repeated
shades, the dark end held). OpenLoco's source has no copy of these tables — they exist only in g1.DAT — so
the open build has to make them. Proposed (for Joop): odd colours exactly as above (the straight ramp is the
swatch itself — no judgement involved); even colours as a documented approximation (a compressed ramp by a
fixed rule), listed in docs/deviations.md as the open build's. Until then the open build draws company
colours and coloured windows with no remap (the GetPaletteMap error paths).

## UI images — the inventory (2026-11-06)

The g1 images goloco's interface names, from the code (static: every image constant in pkg/ui and
pkg/sprites/imageids.go; the UI has no raw numeric g1 indexing):

| g1 range | What | Count |
|---|---|---:|
| 2300–2399 | the UI icon block: folder / parent folder, close button, frame backgrounds, progress bar, red arrows, construction buttons (turnaround, remove, new position, rotate), rubbish bin, centre view, route flags and buttons, refit, the tabs (display, control, sound, misc, globe) | ~74 named in imageids.go + ~35 in pkg/ui (overlapping) |
| 2460–2469 | game speed buttons (pause … extra fast), random map, compass | 6 |
| 3543–3544 | slider track and thumb | 2 |
| 3625, 3700–3765 | the wide tab; a few late icons | ~7 |
| (glyph ranges) | the four bitmap fonts, 4 × 224 glyphs — goloco already has the TTF fallback | — |

Not on this list because they come from objects (OpenGraphics has them): the interface skin (window
colours, toolbar images), the track / road preview images, vehicle, station and building images.
So the open build's UI needs about **120 small icons drawn** (most 16–24 px, 2300–2399) plus a pixel font
at Locomotion's sizes — or the TTF it already uses. Next: a runtime check that nothing else is drawn
(an index-recording hook in a debug run), then draw them — art, Joop's call (§3: draw, or OpenTTD's GPL
icons with the licence kept).
