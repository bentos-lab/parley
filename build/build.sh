#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR" # ensure we always run from the repo root
WEBUI_DIR="$ROOT_DIR/webui"
WEBUI_DIST="$WEBUI_DIR/dist"
REST_DIST="$ROOT_DIR/adapter/inbound/rest/dist"
DIST_DIR="$ROOT_DIR/dist"
ENTRYPOINT="./cmd/parley"
APP_NAME="parley"

PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

log() {
  # log prints a visual separator and the provided message.
  printf "\n%s\n" "$1"
}

ensure_command() {
  # ensure_command verifies the named tool is available on PATH.
  # Parameters: name is the binary to check. Returns: exits with error if missing.
  local name=$1
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "error: $name is required but not installed" >&2
    exit 1
  fi
}

ensure_command pnpm
ensure_command git
ensure_command go

VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
LDFLAGS="-s -w \
  -X main.version=${VERSION} \
  -X main.commit=${COMMIT}"
WINDOWS_LDFLAGS="-H=windowsgui"

log "📦 Preparing web UI"
pushd "$WEBUI_DIR" >/dev/null
pnpm install --frozen-lockfile >/dev/null
pnpm run build >/dev/null
popd >/dev/null

log "🧱 Syncing SPA to REST assets"
rm -rf "$REST_DIST"
mkdir -p "$REST_DIST"
cp -R "$WEBUI_DIST/"* "$REST_DIST/"

log "🚀 Building ${APP_NAME} (${VERSION} @ ${COMMIT})"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"
  BIN_NAME="$APP_NAME"
  ARCHIVE_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"

  if [ "$GOOS" = "windows" ]; then
    BIN_NAME="${BIN_NAME}.exe"
    ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.zip"
  else
    ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.tar.gz"
  fi

  log "🔧 Building ${GOOS}/${GOARCH}"
  TMP_DIR=$(mktemp -d)

  CURRENT_LDFLAGS="${LDFLAGS}"
  if [ "${GOOS}" = "windows" ]; then
    CURRENT_LDFLAGS="${CURRENT_LDFLAGS} ${WINDOWS_LDFLAGS}"
  fi

  env GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 \
    go build -ldflags="${CURRENT_LDFLAGS}" \
    -o "${TMP_DIR}/${BIN_NAME}" \
    ${ENTRYPOINT}

  if [ "$GOOS" = "windows" ]; then
    (cd "$TMP_DIR" && zip -q "${ARCHIVE_PATH}" "${BIN_NAME}")
  else
    (cd "$TMP_DIR" && tar -czf "${ARCHIVE_PATH}" "${BIN_NAME}")
  fi

  rm -rf "$TMP_DIR"
  log "✅ ${ARCHIVE_PATH}"

done

log "🎉 Done. Build artifacts are in ${DIST_DIR}"
