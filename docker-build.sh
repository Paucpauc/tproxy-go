#!/bin/bash

# Docker multi-architecture build script for tproxy
# Builds and pushes Docker images for amd64, arm64, and arm

set -e

# Default values
IMAGE_NAME="tproxy"
TAG=${1:-"latest"}
PUSH=${2:-"false"}

# Supported architectures
ARCHS=("amd64" "arm64" "arm")

echo "Building Docker images for $IMAGE_NAME:$TAG"
echo "Architectures: ${ARCHS[*]}"
echo "Push images: $PUSH"

# Build for each architecture
for ARCH in "${ARCHS[@]}"; do
    echo "=== Building for $ARCH ==="
    
    # Set TARGETVARIANT for ARM
    case $ARCH in
        "arm")
            TARGETVARIANT="7"  # For Mikrotik HAP AC2 (ARMv7)
            ;;
        *)
            TARGETVARIANT=""
            ;;
    esac
    
    # Build the Docker image
    docker build \
        --build-arg TARGETARCH=$ARCH \
        --build-arg TARGETVARIANT=$TARGETVARIANT \
        -t "$IMAGE_NAME:$TAG-$ARCH" .
    
    echo "✓ Built: $IMAGE_NAME:$TAG-$ARCH"
    
    # Create manifest if pushing
    if [ "$PUSH" = "true" ]; then
        docker tag "$IMAGE_NAME:$TAG-$ARCH" "$IMAGE_NAME:$TAG-$ARCH"
    fi
done

# Create and push multi-arch manifest if requested
if [ "$PUSH" = "true" ]; then
    echo ""
    echo "=== Creating multi-arch manifest ==="
    
    # Create manifest command
    MANIFEST_CMD="docker manifest create $IMAGE_NAME:$TAG"
    for ARCH in "${ARCHS[@]}"; do
        MANIFEST_CMD="$MANIFEST_CMD $IMAGE_NAME:$TAG-$ARCH"
    done
    
    # Execute manifest creation
    eval $MANIFEST_CMD
    
    # Annotate each architecture
    for ARCH in "${ARCHS[@]}"; do
        case $ARCH in
            "arm")
                OS_ARCH="linux/arm/v7"
                ;;
            *)
                OS_ARCH="linux/$ARCH"
                ;;
        esac
        
        docker manifest annotate "$IMAGE_NAME:$TAG" "$IMAGE_NAME:$TAG-$ARCH" --arch $ARCH --os linux
    done
    
    # Push the manifest
    docker manifest push "$IMAGE_NAME:$TAG"
    echo "✓ Pushed multi-arch manifest: $IMAGE_NAME:$TAG"
fi

echo ""
echo "Docker build completed successfully!"