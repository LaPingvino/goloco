"""Align a goloco shot with an OpenLoco shot of the same save, then diff them.

    python3 -I match.py goloco.png openloco_window.png diff.png

Both 800x600 (crop OpenLoco's window out of the Xvfb screen first). Finds the
offset by exact-matching small goloco patches anywhere in the OpenLoco frame
(the worlds are pixel-identical where goloco is right, so the best patch scores
every pixel), then writes goloco | OpenLoco | the differing pixels in magenta.
"""
import sys

import numpy as np
from numpy.lib.stride_tricks import sliding_window_view
from PIL import Image


def enc(a):
    return (a[..., 0] << 16) | (a[..., 1] << 8) | a[..., 2]


g = np.asarray(Image.open(sys.argv[1]).convert("RGB")).astype(np.int64)
o = np.asarray(Image.open(sys.argv[2]).convert("RGB")).astype(np.int64)
gc, oc = enc(g), enc(o)
best = (0, 0, 0)
for cx in range(100, 700, 150):
    for cy in range(100, 500, 120):
        p = gc[cy - 12:cy + 12, cx - 24:cx + 24]
        s = (sliding_window_view(oc[30:570], (24, 48)) == p).sum(axis=(2, 3))
        y, x = np.unravel_index(s.argmax(), s.shape)
        if s.max() > best[0]:
            best = (int(s.max()), int(x + 24 - cx), int(y + 30 + 12 - cy))
n, dx, dy = best
print(f"offset {dx},{dy} (OpenLoco = goloco + offset); patch match {n}/1152")
gy0, gy1 = max(30, 30 - dy), min(570, 570 - dy)
gx0, gx1 = max(0, -dx), min(800, 800 - dx)
gg = g[gy0:gy1, gx0:gx1]
oo = o[gy0 + dy:gy1 + dy, gx0 + dx:gx1 + dx]
d = np.abs(gg - oo).sum(axis=2) > 0
print(f"overlap {gg.shape[1]}x{gg.shape[0]}: {100 * d.mean():.2f}% of pixels differ")
m = (gg * 0.4).astype(np.uint8)
m[d] = [255, 0, 255]
out = Image.new("RGB", (gg.shape[1], gg.shape[0] * 3 + 20), "white")
for i, a in enumerate((gg, oo, m)):
    out.paste(Image.fromarray(a.astype(np.uint8)), (0, i * (gg.shape[0] + 10)))
out.save(sys.argv[3])
