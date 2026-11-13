#!/bin/sh
# Runs goloco's whole test suite, each test binary on an Xvfb display of its
# own (scripts/xvfb-exec.sh). Sharing one xvfb-run display between binaries
# twice panicked in cmd/goloco ("glfw: The GLFW library is not initialized").
# GOLOCO_DATA must point at the game's data for the data-backed tests.
here=$(dirname "$(readlink -f "$0")")
exec go -C "$here/.." test -count=1 -exec "$here/xvfb-exec.sh" "${@:-./...}"
