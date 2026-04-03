#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

WEBUI_DIR="$ROOT_DIR/webui"
WEBUI_DIST="$WEBUI_DIR/dist"
REST_DIST="$ROOT_DIR/adapter/inbound/rest/dist"

log() {
  # log prints a header to make progress easier to scan.
  printf "\n%s\n" "$1"
}

ensure_command() {
  # ensure_command exits if the named binary is missing from PATH.
  local name=$1
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "error: $name is required but not installed" >&2
    exit 1
  fi
}

ensure_command pnpm

log "📦 Building web UI assets"
pushd "$WEBUI_DIR" >/dev/null
pnpm install --frozen-lockfile >/dev/null
pnpm run build >/dev/null
popd >/dev/null

log "🧱 Syncing SPA to REST assets"
rm -rf "$REST_DIST"
mkdir -p "$REST_DIST"
cp -R "$WEBUI_DIST/"* "$REST_DIST/"

log "✅ UI build sync complete"
