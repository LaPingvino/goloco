# Where OpenLoco's drawing can be split for the GPU

Notes from GoLoco, a Go port of OpenLoco that draws the world with OpenLoco's own software renderer (ported)
and the interface with Ebitengine on the GPU. Joop's question: OpenLoco is hard to speed up with a GPU as it
stands; does the port show where a meaningful split could go? These are measurements, not a proposal; every
number is from GoLoco (Go, 1280×800 unless said), so only the proportions carry over.

## 1. What a frame costs, by part

Measured with frame statistics in the game (avg / worst over 2 s) and CPU profiles:

| Part | Cost | Notes |
|---|---|---|
| World paint (generate → arrange → draw, per column) | 45–75 ms one core, ~40 ms on 8 | the bulk; OpenLoco already paints columns in parallel |
| 8-bit → RGBA conversion of the world | a few ms | a palette lookup per pixel |
| UI windows (translucency) — with a GPU readback | 50–110 ms a frame | reading the pixels under a window back from the GPU |
| UI windows — with a CPU copy of the indices | 0.1–2 ms | see §3 |

## 2. The natural cut: world (8-bit, CPU) | everything else (GPU)

The world is the part that has to stay in OpenLoco's paint pipeline: the sort (arrangeStructs) and the
overdraw rules define what the game looks like, and they work on palette indices. Everything drawn after the
world — windows, text, the cursor, labels — can go to the GPU as ordinary textured quads, with the world as
one more texture underneath.

What makes the cut work cheaply:

- **Columns are independent.** Viewport::paint's 32-pixel columns each get their own session; we checked that
  painting any subset of columns gives exactly those columns of the full frame (zoom 0–3). So a frame's world
  can be painted in parts — a third of the columns a frame while a window is open, say — and only that part
  converted and uploaded. In GoLoco this took the worst frame with a window open from 74 ms to 21 ms on one
  core, without changing a pixel of the result.
- **Upload only what changed.** Converting and uploading only the repainted columns (WritePixels on a
  sub-rectangle) is most of the gain once the paint itself is spread.

## 3. The catch: translucent windows need the palette indices under them

OpenLoco's translucent window backgrounds are palette remaps of the pixels beneath — an operation on
indices, not colours. With the world on the GPU, the naive way reads the screen back from the GPU and maps
RGB back to indices: that readback was nearly all of GoLoco's UI cost (50–110 ms a frame for two windows).

What fixed it: keep a CPU copy of the screen as palette indices — the world's 8-bit frame, updated by each
translucent window as it is drawn — and take the indices from there. Only where something else was drawn
(labels, the toolbar) does it fall back to a readback. Result: 0.1–2 ms; 0 differing pixels against the
readback over ~18 million compared.

For OpenLoco this suggests: if the final composition moves to the GPU, keep the 8-bit world buffer on the
CPU side as the source of truth for remap-style effects, or do the remaps on the GPU as a palette-index
texture plus a remap lookup texture (one texture fetch each) — never read the screen back.

## 4. What stays on the CPU, and why that's fine

- The paint sort and overdraw: the look of the game. Parallel by column already.
- Per-tile animation state, the tile walk (generateTilesAndEntities): cheap once per column.
- Palette conversion could move to the GPU (an index texture and a palette texture, converted in a shader),
  which would remove the per-pixel CPU conversion and halve the upload (1 byte a pixel instead of 4).

## 5. Things we found on the way (port deviations, now fixed or noted)

- We had painted the world in one session over the whole view instead of upstream's columns: 329–1551
  pixels of the title view differed (sort order). Upstream's column walk (generateParameters,
  paintTileElements2, paintEntities2) also turned out ~4× cheaper per column than a rectangle walk.
- Upstream clears each column to brown2 (black underground); a renderer that clears to 0 shows black at the
  map's edge.
