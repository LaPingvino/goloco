"""Make Locomotion's title-screen game loadable as an ordinary save.

    python3 -I untitle.py ~/Projects/locomotion/Data/title.dat $SCRATCH/title.SV5

Data/title.dat is a saved game with years of history (stations, vehicles,
ratings, graphs) — goloco opens it as it is, the real OpenLoco refuses ("File
is a title sequence", S5.cpp). The only thing in the way is the header's
isTitleSequence flag: this clears it and corrects the file's checksum. The
copy is game data: keep it out of git.
"""
import struct
import sys

d = bytearray(open(sys.argv[1], "rb").read())
# chunk 0: encoding 3 (rotate), length 32, then the header; its second byte
# (flags) is stored rotated left by 3.
assert d[0] == 3 and struct.unpack_from("<I", d, 1)[0] == 32, "not an S5 file"
flags = (d[6] >> 3 | d[6] << 5) & 0xFF
new = flags & ~(1 << 2)
enc = (new << 3 | new >> 5) & 0xFF
(total,) = struct.unpack_from("<I", d, len(d) - 4)
struct.pack_into("<I", d, len(d) - 4, (total - d[6] + enc) & 0xFFFFFFFF)
d[6] = enc
open(sys.argv[2], "wb").write(d)
print(f"flags {flags:#04x} -> {new:#04x}")
