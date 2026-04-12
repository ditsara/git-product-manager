#!/bin/bash
set -e

# Determine version
if [ -n "$1" ]; then
  VERSION="$1"
else
  VERSION=$(git tag --points-at HEAD 2>/dev/null | grep -E '^v[0-9]' | head -1)
  if [ -z "$VERSION" ]; then
    echo "Error: no version tag found on current commit. Pass a version as an argument or tag the commit first." >&2
    echo "  Usage: $0 [version]" >&2
    echo "  Example: $0 v0.1.0-beta" >&2
    exit 1
  fi
fi

BINARY="pm"
OUTPUT_DIR="bin"
PACKAGE="./cmd/pm"

PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

mkdir -p "$OUTPUT_DIR"

echo "Building $BINARY $VERSION for all platforms..."
echo ""

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  OUTPUT_NAME="${BINARY}_${VERSION}_${GOOS}_${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
  fi

  printf "  %-40s" "${GOOS}/${GOARCH}"

  GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
    "$PACKAGE"

  echo "→ ${OUTPUT_DIR}/${OUTPUT_NAME}"
done

echo ""
echo "✓ Done. Binaries in ${OUTPUT_DIR}/:"
ls -lh "${OUTPUT_DIR}/${BINARY}_${VERSION}_"* 2>/dev/null
