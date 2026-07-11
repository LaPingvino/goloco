#!/bin/bash
# Fetch the OpenTTD TTF fonts (OFL-licensed) into ../fonts/ for goloco's
# TTF text path. https://github.com/OpenTTD/OpenTTD-TTF
set -e
DEST="${1:-../fonts}"
mkdir -p "$DEST"
BASE="https://github.com/OpenTTD/OpenTTD-TTF/releases/latest/download"
for f in OpenTTD-Sans.ttf OpenTTD-Small.ttf OpenTTD-Mono.ttf; do
  echo "Fetching $f..."
  curl -fsSL "$BASE/$f" -o "$DEST/$f" || echo "  (failed: $f)"
done
ls -la "$DEST"
