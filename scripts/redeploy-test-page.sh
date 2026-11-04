#!/usr/bin/env bash
# redeploy-test-page.sh — THE DEPLOY of the private test page https://kiefte.eu/goloco/
# (basic auth). No sudo: Caddy already serves ~/Projects/goloco-web-local (set up once by
# caddy-goloco-first-time-setup.sh), so a rebuild into that folder is the whole deploy.
# The page fetches with cache: "no-cache", so the new build shows at once.
#
# Builds what is committed (uncommitted work is stashed for the build and put back).
#   bash scripts/redeploy-test-page.sh [build dir]
set -euo pipefail
cd "$(dirname "$0")/.."
WEB=${1:-$HOME/Projects/goloco-web-local}
stashed=0
if [[ -n $(git status --porcelain --untracked-files=no) ]]; then
	git stash -q && stashed=1
fi
trap '[[ $stashed == 1 ]] && git stash pop -q' EXIT
go run ./cmd/webpack -local -music -out "$WEB"
chmod -R o+rX "$WEB" # Caddy runs as its own user
echo "Deployed $(git log --oneline -1) → https://kiefte.eu/goloco/"
