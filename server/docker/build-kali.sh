#!/bin/bash

# Build script for Kali Linux security tools container
# Usage: ./build-kali.sh [--no-cache] [--quiet]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_NAME="gogogadgeto/kali-tools"
TAG="latest"
DOCKERFILE="$SCRIPT_DIR/Dockerfile.kali"

# Parse command line arguments
NO_CACHE=""
QUIET=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        --quiet)
            QUIET="--quiet"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--no-cache] [--quiet]"
            exit 1
            ;;
    esac
done

echo "🐳 Building Kali Linux security tools container..."
echo "📁 Dockerfile: $DOCKERFILE"
echo "🏷️  Image: $IMAGE_NAME:$TAG"
echo ""

# Check if Dockerfile exists
if [[ ! -f "$DOCKERFILE" ]]; then
    echo "❌ Error: Dockerfile not found at $DOCKERFILE"
    exit 1
fi

# Build the Docker image
echo "🔨 Starting Docker build..."
BUILD_START=$(date +%s)

docker build \
    $NO_CACHE \
    $QUIET \
    -t "$IMAGE_NAME:$TAG" \
    -f "$DOCKERFILE" \
    "$SCRIPT_DIR"

BUILD_END=$(date +%s)
BUILD_TIME=$((BUILD_END - BUILD_START))

echo ""
echo "✅ Build completed successfully!"
echo "⏱️  Build time: ${BUILD_TIME} seconds"
echo "🏷️  Image: $IMAGE_NAME:$TAG"
echo ""

# Show image size
echo "📊 Image information:"
docker images "$IMAGE_NAME:$TAG" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

echo ""
echo "🧪 Testing image..."
if docker run --rm "$IMAGE_NAME:$TAG" nmap --version >/dev/null 2>&1; then
    echo "✅ Image test passed - nmap is working"
else
    echo "❌ Image test failed - nmap not working"
    exit 1
fi

echo ""
echo "🎉 Kali container build complete!"
echo ""
echo "Usage in Go code:"
echo "  Image: \"$IMAGE_NAME:$TAG\""
echo ""
echo "To rebuild with no cache:"
echo "  ./build-kali.sh --no-cache" 