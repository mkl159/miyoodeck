#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
#  Onion Web Deck - Build Script
#  Builds the Go server binary + Svelte frontend, then packages
#  everything ready to copy onto the Miyoo Mini SD card.
# ═══════════════════════════════════════════════════════════════
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_DIR="$SCRIPT_DIR/package/App/WebDeck"
WWW_DIR="$PKG_DIR/www"
BINARY="$PKG_DIR/webdeck"

echo ""
echo "╔══════════════════════════════════════╗"
echo "║   Onion Web Deck - Build             ║"
echo "╚══════════════════════════════════════╝"
echo ""

# ── Step 1: Build Go binary using Docker ─────────────────────────────────────
echo "▶ [1/3] Building Go server (ARM Cortex-A7)..."

cd "$SCRIPT_DIR"

# Generate go.sum if missing
if [ ! -f server/go.sum ]; then
    echo "  → Generating go.sum via Docker..."
    docker run --rm \
        -v "$SCRIPT_DIR/server:/app" \
        -w /app \
        golang:1.21-alpine \
        sh -c "go mod tidy"
fi

# Build ARM binary
docker build \
    --output type=local,dest="$SCRIPT_DIR/build-output" \
    --target export \
    -f Dockerfile \
    .

cp "$SCRIPT_DIR/build-output/webdeck" "$BINARY"
chmod +x "$BINARY"
echo "  ✓ Binary: $BINARY ($(du -sh "$BINARY" | cut -f1))"

# ── Step 2: Build Svelte frontend ────────────────────────────────────────────
echo ""
echo "▶ [2/3] Building Svelte frontend..."

cd "$SCRIPT_DIR/frontend"

if [ ! -d node_modules ]; then
    echo "  → Installing npm dependencies..."
    npm install --silent
fi

npm run build

echo "  ✓ Frontend built → $WWW_DIR"

# ── Step 3: Show final package size ──────────────────────────────────────────
echo ""
echo "▶ [3/3] Package ready!"
echo ""
echo "  Package contents:"
find "$PKG_DIR" -type f | while read f; do
    size=$(du -sh "$f" | cut -f1)
    rel="${f#$SCRIPT_DIR/package/}"
    printf "    %-40s %s\n" "$rel" "$size"
done
echo ""
echo "  Total: $(du -sh "$PKG_DIR" | cut -f1)"
echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║  INSTALLATION                                        ║"
echo "║                                                      ║"
echo "║  Copy to your Miyoo Mini SD card:                    ║"
echo "║                                                      ║"
echo "║    cp -r package/App/WebDeck /Volumes/SDCARD/App/    ║"
echo "║                                                      ║"
echo "║  Then launch 'Web Deck' from the Onion App menu.     ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""
