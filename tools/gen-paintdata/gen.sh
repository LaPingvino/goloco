#!/usr/bin/env bash
# gen.sh — regenerate pkg/paint/roaddata_gen.go from an OpenLoco checkout.
#   tools/gen-paintdata/gen.sh [~/Projects/OpenLoco]
set -euo pipefail
OL=${1:-$HOME/Projects/OpenLoco}
here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../.." && pwd)
inc=(-I"$OL/src/OpenLoco/include" -I"$OL/src/OpenLoco/src" -I"$OL/src/OpenLoco/include/OpenLoco" -I"$OL/src/Audio/include"
     -I"$OL/src/Core/include" -I"$OL/build/_deps/fmt-src/include" -I"$OL/build/_deps/sfl-src/include" -I"$OL/src/Diagnostics/include"
     -I"$OL/src/Platform/include" -I"$OL/src/Utility/include" -I"$OL/src/Math/include" -I"$OL/src/Engine/include" -I"$OL/src/Gfx/include")
tmp=$(mktemp -d)
g++ -std=c++20 -fno-char8_t -DDEBUG=0 "${inc[@]}" -o "$tmp/gen" "$here/gen.cpp"
g++ -std=c++20 -fno-char8_t -DDEBUG=0 "${inc[@]}" -o "$tmp/gen_track" "$here/gen_track.cpp"
"$tmp/gen" > "$repo/pkg/paint/roaddata_gen.go"
"$tmp/gen_track" > "$repo/pkg/paint/trackdata_gen.go"
gofmt -w "$repo/pkg/paint/roaddata_gen.go" "$repo/pkg/paint/trackdata_gen.go"
rm -rf "$tmp"
echo "wrote pkg/paint/roaddata_gen.go + trackdata_gen.go (OpenLoco $(git -C "$OL" rev-parse --short HEAD 2>/dev/null || echo '?'))"
