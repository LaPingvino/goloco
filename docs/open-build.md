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
| tree | 1 | 39 | thin landscape |
| airport, dock | 1 / 1 | 4 / 1 | usable |
| climate, region, currency, town names, steam, competitor, scenario text | 1 each | 6 / 3 / 12 / 6 / 5 / 38 / 45 | minimal set |
| sound | 1 | 42 | **almost no sound effects** |

## What is missing for a fully open build

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
3. Decide (a) vs (b) for the g1 gap.
4. Only then a public link — under a separate route without basic auth, built only from MIT/open sources,
   with the OpenGraphics licence and credits on the page.
