#!/bin/sh
# Runs goloco's whole test suite under Xvfb: pkg/... and cmd/goloco each on a
# display of its own. A combined run (go test ./pkg/... ./cmd/goloco under one
# xvfb-run) twice panicked in cmd/goloco with "glfw: The GLFW library is not
# initialized" and never reproduced alone (see LOOP.md); a failing run's full
# output is kept for its stack. GOLOCO_DATA must point at the game's data.
set -u
log=${TMPDIR:-/tmp}/goloco-test-$$.log
status=0
for target in "./pkg/..." "./cmd/goloco"; do
	if ! xvfb-run -a go test -count=1 "$target" >"$log" 2>&1; then
		status=1
		keep=${TMPDIR:-/tmp}/goloco-test-failed-$(date +%Y%m%d-%H%M%S).log
		cp "$log" "$keep"
		grep -E '^(--- FAIL|FAIL|panic)' "$log"
		echo "full output kept: $keep"
	else
		grep -E '^ok' "$log" | wc -l | xargs printf "%s: %s packages ok\n" "$target"
	fi
done
rm -f "$log"
exit $status
