#!/bin/bash
# Load Docker images from tar files and optionally push to private registry

set -e

INPUT_DIR="./offline-images"
TARGET_REGISTRY=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --input-dir)
            INPUT_DIR="$2"
            shift 2
            ;;
        --target-registry)
            TARGET_REGISTRY="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --input-dir DIR          Input directory (default: ./offline-images)"
            echo "  --target-registry REG    Push to private registry after loading"
            echo "  --help                   Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "=========================================="
echo "Loading Docker Images"
echo "=========================================="
echo "Input Directory: $INPUT_DIR"
if [ -n "$TARGET_REGISTRY" ]; then
    echo "Target Registry: $TARGET_REGISTRY"
fi
echo ""

# Load images
for file in "${INPUT_DIR}"/*.tar; do
    if [ -f "$file" ]; then
        echo "Loading: $(basename "$file")"
        docker load -i "$file"
        echo "  ✓ Loaded"
    fi
done

# Push to private registry if specified
if [ -n "$TARGET_REGISTRY" ]; then
    echo ""
    echo "Pushing images to $TARGET_REGISTRY..."
    
    # Get list of loaded images
    images=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep -E "(hadoop|zookeeper)" || true)
    
    for image in $images; do
        target_image="${TARGET_REGISTRY}/${image}"
        echo "Pushing: $image -> $target_image"
        docker tag "$image" "$target_image"
        docker push "$target_image"
        echo "  ✓ Pushed"
    done
fi

echo ""
echo "=========================================="
echo "Images loaded successfully!"
echo "=========================================="
