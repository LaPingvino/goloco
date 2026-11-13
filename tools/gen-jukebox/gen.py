import re, sys, os
O = os.path.expanduser("~/Projects/OpenLoco/src/OpenLoco")
env_h = open(O + "/include/OpenLoco/Environment.h").read()
enum = re.search(r"enum class PathId[^{]*\{(.*?)\};", env_h, re.S).group(1)
ids = [re.sub(r"//.*", "", l).strip().rstrip(",") for l in enum.splitlines()]
ids = [i for i in ids if i]
env_c = open(O + "/src/Environment.cpp").read()
paths = re.findall(r'"([^"]*)"', re.search(r"kPaths = \{(.*?)\};", env_c, re.S).group(1))
path = dict(zip(ids, paths))
juke = open(O + "/src/Jukebox.cpp").read()
rows = re.findall(r"\{ PathId::(\w+), StringIds::(\w+), (\w+), (\w+) \}", re.search(r"kMusicInfo\[\] = \{(.*?)\};", juke, re.S).group(1))
cfg = open(O + "/include/OpenLoco/Config.h").read()
items = [l.strip().rstrip(",") for l in re.search(r"enum class PlaylistItem[^{]*\{(.*?)\};", cfg, re.S).group(1).splitlines() if l.strip()]
assert len(rows) == 30 == len(items), (len(rows), len(items))
camel = lambda s: "".join(p.capitalize() for p in s.split("_"))
yr = {"kBeginningOfTime": "BeginningOfTime", "kEndOfTime": "EndOfTime"}
out = ["// Code generated from OpenLoco's Jukebox.cpp (kMusicInfo), Environment.h /",
       "// Environment.cpp (PathId → sub path) and Config.h (PlaylistItem) by",
       "// tools/gen-jukebox/gen.py; DO NOT EDIT.", "",
       "package jukebox", "", 'import "github.com/LaPingvino/goloco/pkg/i18n"', "",
       "// musicInfo is kMusicInfo (0x004FE910): each track's file under the",
       "// Locomotion folder, its title and the years of its era; Key is its",
       "// PlaylistItem name, the key of openloco.yml's audio.customJukebox.",
       "var musicInfo = [NumMusicTracks]MusicInfo{"]
for (p, s, a, b), k in zip(rows, items):
    assert path[p].startswith("Data/")
    out.append('\t{File: "%s", Key: "%s", TitleID: i18n.%s, StartYear: %s, EndYear: %s},' % (path[p][5:], k, camel(s), yr.get(a, a), yr.get(b, b)))
out.append("}")
open(sys.argv[1], "w").write("\n".join(out) + "\n")
