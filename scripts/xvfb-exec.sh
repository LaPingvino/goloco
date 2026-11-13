#!/bin/sh
# go test -exec helper: runs one test binary on an Xvfb display of its own.
# Ebitengine opens the window system in its package init, so every test
# binary that imports it (pkg/ui, pkg/world, cmd/goloco) needs a display.
# -displayfd lets Xvfb pick a free display and report it, so test binaries
# running in parallel never race for one (xvfb-run -a can).
#
#   go test -exec "$PWD/scripts/xvfb-exec.sh" ./...
set -eu
dir=$(mktemp -d)
mkfifo "$dir/fd"
Xvfb -displayfd 3 -screen 0 800x600x24 -nolisten tcp 3>"$dir/fd" 2>/dev/null &
xpid=$!
read -r num <"$dir/fd"
rm -r "$dir"
trap 'kill $xpid 2>/dev/null; wait $xpid 2>/dev/null' EXIT INT TERM
DISPLAY=:$num SDL_AUDIODRIVER=dummy "$@"
