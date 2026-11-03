#!/usr/bin/env python3
"""Writes pkg/sim/testdata/roadjunctions.txt from upstream's own code: the
junction loop of createRoad (CreateRoad.cpp 0x47656B) cut out verbatim, its
placeAddition swapped for a recorder (rotation & 3, as setRotation keeps),
run over every combination of roadIdUnk[0..2]'s rotation bits.
Usage: gen_roadjunctions.py [OpenLoco dir]  (needs g++)"""
import os, subprocess, sys, tempfile
ol = sys.argv[1] if len(sys.argv) > 1 else os.path.expanduser("~/Projects/OpenLoco")
repo = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))
src = open(os.path.join(ol, "src/OpenLoco/src/GameCommands/Road/CreateRoad.cpp")).read()
start = src.index("            for (auto i = 0U; i < 4; ++i)\n            {\n                const auto rot0 = i;")
end = src.index("        if ((returnState.flags_1136073 & (1U << 0)) && !(returnState.flags_1136073 & (1U << 6)))")
body = src[start:end]
body = body[:body.rstrip().rindex("}")]
pa = body.index("auto placeAddition =")
pe = body.index("};", body.index("setAiAllocated")) + 2
body = body[:pa] + 'auto placeAddition = [&](uint8_t rotation, uint8_t roadId) { std::printf(" %u:%u", (unsigned)(rotation & 3), (unsigned)roadId); };' + body[pe:]
prog = """#include <array>
#include <cstdint>
#include <cstdio>
int main() {
    for (unsigned in = 0; in < 4096; ++in) {
        std::array<uint8_t, 16> roadIdUnk = {};
        roadIdUnk[0] = (in & 0xF) | 0x80; roadIdUnk[1] = (in >> 4) & 0xF; roadIdUnk[2] = (in >> 8) & 0xF;
        std::printf("%u", in);
""" + body + """
        std::printf(" | %u %u %u\\n", (unsigned)roadIdUnk[0], (unsigned)roadIdUnk[1], (unsigned)roadIdUnk[2]);
    }
}
"""
with tempfile.TemporaryDirectory() as tmp:
    open(os.path.join(tmp, "h.cpp"), "w").write(prog)
    subprocess.run(["g++", "-std=c++20", "-o", os.path.join(tmp, "h"), os.path.join(tmp, "h.cpp")], check=True)
    out = subprocess.run([os.path.join(tmp, "h")], check=True, capture_output=True).stdout
open(os.path.join(repo, "pkg/sim/testdata/roadjunctions.txt"), "wb").write(out)
