#!/bin/bash
set -euo pipefail

# Build snipt.app bundle for macOS with app icon.
# Usage: ./build/build-macos.sh

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP="$ROOT/build/bin/snipt.app"
ICON_SRC="$ROOT/build/appicon/icon.png"

# Build frontend
echo "Building frontend..."
(cd "$ROOT/src/frontend" && npm run build)

# Build Go binary
echo "Building Go binary..."
mkdir -p "$APP/Contents/MacOS"
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
go build -ldflags "-s -w -X main.version=$VERSION" -o "$APP/Contents/MacOS/snipt" ./src/cmd/snipt

# Copy Info.plist
cp "$ROOT/build/appicon/Info.plist" "$APP/Contents/Info.plist"

# Generate .icns from source PNG
echo "Generating app icon..."
TMPDIR_ICON=$(mktemp -d)
ICONSET="$TMPDIR_ICON/snipt.iconset"
mkdir -p "$ICONSET"

sips -z   16   16 "$ICON_SRC" --out "$ICONSET/icon_16x16.png"      > /dev/null
sips -z   32   32 "$ICON_SRC" --out "$ICONSET/icon_16x16@2x.png"   > /dev/null
sips -z   32   32 "$ICON_SRC" --out "$ICONSET/icon_32x32.png"      > /dev/null
sips -z   64   64 "$ICON_SRC" --out "$ICONSET/icon_32x32@2x.png"   > /dev/null
sips -z  128  128 "$ICON_SRC" --out "$ICONSET/icon_128x128.png"    > /dev/null
sips -z  256  256 "$ICON_SRC" --out "$ICONSET/icon_128x128@2x.png" > /dev/null
sips -z  256  256 "$ICON_SRC" --out "$ICONSET/icon_256x256.png"    > /dev/null
sips -z  512  512 "$ICON_SRC" --out "$ICONSET/icon_256x256@2x.png" > /dev/null
sips -z  512  512 "$ICON_SRC" --out "$ICONSET/icon_512x512.png"    > /dev/null
sips -z 1024 1024 "$ICON_SRC" --out "$ICONSET/icon_512x512@2x.png" > /dev/null

mkdir -p "$APP/Contents/Resources"
iconutil -c icns "$ICONSET" -o "$APP/Contents/Resources/iconfile.icns"
rm -rf "$TMPDIR_ICON"

echo "Done: $APP"
