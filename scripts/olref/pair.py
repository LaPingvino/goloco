"""Put a goloco shot and an OpenLoco shot next to each other.

    python3 -I pair.py out.png goloco.png[@X,Y,W,H] openloco.png[@X,Y,W,H] [SCALE]

goloco on the left, OpenLoco on the right, each cropped to its @rectangle (a
window, say) and enlarged SCALE times (default 1, nearest-neighbour). When the
two crops are the same size a third panel marks every differing pixel in
magenta over a dimmed goloco, and the share of them is printed. For whole map
views of the same save use match.py, which finds the offset first.
"""
import sys

import numpy as np
from PIL import Image


def load(spec):
    path, _, rect = spec.partition("@")
    im = Image.open(path).convert("RGB")
    if rect:
        x, y, w, h = (int(v) for v in rect.split(","))
        im = im.crop((x, y, x + w, y + h))
    return im


out, g, o = sys.argv[1], load(sys.argv[2]), load(sys.argv[3])
scale = int(sys.argv[4]) if len(sys.argv) > 4 else 1
panels = [g, o]
if g.size == o.size:
    a, b = np.asarray(g).astype(int), np.asarray(o).astype(int)
    d = np.abs(a - b).sum(axis=2) > 0
    m = (a * 0.4).astype(np.uint8)
    m[d] = [255, 0, 255]
    panels.append(Image.fromarray(m))
    print(f"{g.size[0]}x{g.size[1]}: {100 * d.mean():.2f}% of pixels differ")
else:
    print(f"sizes differ: goloco {g.size[0]}x{g.size[1]}, OpenLoco {o.size[0]}x{o.size[1]}")
panels = [p.resize((p.width * scale, p.height * scale), Image.NEAREST) for p in panels]
gap = 8
sheet = Image.new("RGB", (sum(p.width for p in panels) + gap * (len(panels) - 1), max(p.height for p in panels)), "white")
x = 0
for p in panels:
    sheet.paste(p, (x, 0))
    x += p.width + gap
sheet.save(out)
