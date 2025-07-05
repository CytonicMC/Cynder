#!/bin/bash

# Exit on any error
set -e

# Variables
IMAGE_NAME="cynder"
GHCR_USERNAME_LOWER=$(echo "$GHCR_USERNAME" | tr '[:upper:]' '[:lower:]')
VERSION="$1"         # Version tag passed as first argument

# Check if environment variables are set
if [ -z "$GHCR_USERNAME" ] || [ -z "$GHCR_TOKEN" ]; then
    echo "Error: GHCR_USERNAME and GHCR_TOKEN environment variables must be set"
    exit 1
fi

# If version is not provided, use latest git tag
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "latest")
fi

# Login to GitHub Container Registry
echo "$GHCR_TOKEN" | docker login ghcr.io --username "$GHCR_USERNAME" --password-stdin

# Create and use a new builder instance with multi-architecture support
docker buildx create --use --name multi-arch-builder

# Build and push multi-architecture images
GHCR_IMAGE="ghcr.io/$GHCR_USERNAME_LOWER/$IMAGE_NAME"
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag "$GHCR_IMAGE:$VERSION" \
    --tag "$GHCR_IMAGE:latest" \
    --push \
    .

# Clean up the builder
docker buildx rm multi-arch-builder

echo "Successfully pushed multi-arch images $GHCR_IMAGE:$VERSION and $GHCR_IMAGE:latest"