#!/usr/bin/env bash
# caddy-goloco-first-time-setup.sh — ONE-TIME SETUP (done 2026-11-03): make Caddy
# serve the goloco wasm test build at https://kiefte.eu/goloco/ behind HTTP basic
# auth (the build holds the Locomotion game data, which may not be public).
#
# NOT the deploy. Redeploying is only a rebuild into the served folder, no sudo:
#   bash scripts/redeploy-test-page.sh
# Run this one again only to change the password or the served folder:
#   sudo bash scripts/caddy-goloco-first-time-setup.sh [build dir] [user]
#
# Inserts (or replaces) a marked block in the kiefte.eu site of /etc/caddy/Caddyfile,
# keeps a Caddyfile.bak.<epoch> backup, validates, and reloads Caddy. If validation
# fails, the backup is put back and Caddy is left untouched.
set -euo pipefail

WEB=${1:-/home/joop/Projects/goloco-web-local}
AUTH_USER=${2:-joop}
CADDYFILE=/etc/caddy/Caddyfile
SITE_RE='^kiefte[.]eu, www[.]kiefte[.]eu'

[[ $EUID -eq 0 ]] || { echo "run with sudo" >&2; exit 1; }
[[ -f $WEB/index.html ]] || { echo "no build in $WEB — run: go run ./cmd/webpack -local -out $WEB" >&2; exit 1; }
grep -qE "$SITE_RE" "$CADDYFILE" || { echo "kiefte.eu site block not found in $CADDYFILE" >&2; exit 1; }

read -rsp "Password for $AUTH_USER at /goloco/: " pw; echo
[[ -n $pw ]] || { echo "empty password" >&2; exit 1; }
hash=$(caddy hash-password --plaintext "$pw")
unset pw

# Caddy runs as its own user: let it read the build (the dirs above are already 755).
chmod -R o+rX "$WEB"

backup="$CADDYFILE.bak.$(date +%s)"
cp -p "$CADDYFILE" "$backup"

block=$(cat <<EOF
        # goloco-test BEGIN (managed by goloco/scripts/caddy-goloco-first-time-setup.sh)
        redir /goloco /goloco/
        handle_path /goloco/* {
                basic_auth {
                        $AUTH_USER $hash
                }
                root * $WEB
                file_server
        }
        # goloco-test END
EOF
)

# Drop an earlier copy of the block, then insert it right after the site's opening line.
tmp=$(mktemp)
awk -v block="$block" -v site="$SITE_RE" '
        /# goloco-test BEGIN/ { skip = 1 }
        skip { if (/# goloco-test END/) skip = 0; next }
        { print }
        $0 ~ site && !done { print block; done = 1 }
' "$CADDYFILE" > "$tmp"
cat "$tmp" > "$CADDYFILE"
rm -f "$tmp"

if ! caddy validate --config "$CADDYFILE" --adapter caddyfile >/dev/null 2>&1; then
        cp -p "$backup" "$CADDYFILE"
        echo "Caddyfile did not validate — restored $backup. Details:" >&2
        caddy validate --config "$CADDYFILE" --adapter caddyfile >&2 || true
        exit 1
fi
systemctl reload caddy
echo "Done: https://kiefte.eu/goloco/ (user $AUTH_USER). Backup: $backup"
