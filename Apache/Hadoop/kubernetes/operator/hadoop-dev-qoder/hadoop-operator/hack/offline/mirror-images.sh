#!/bin/bash
# Offline Image Mirroring Script for Hadoop Operator
# This script helps mirror required images to a private registry for offline deployment

set -e

# Configuration
SOURCE_REGISTRY=""
TARGET_REGISTRY=""
HADOOP_VERSION="3.3.6"
ZOOKEEPER_VERSION="3.8"
OPERATOR_VERSION="latest"

# Images to mirror
IMAGES=(
    "apache/hadoop:${HADOOP_VERSION}"
    "zookeeper:${ZOOKEEPER_VERSION}"
    "hadoop-operator:${OPERATOR_VERSION}"
)

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --source-registry)
            SOURCE_REGISTRY="$2"
            shift 2
            ;;
        --target-registry)
            TARGET_REGISTRY="$2"
            shift 2
            ;;
        --hadoop-version)
            HADOOP_VERSION="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --source-registry REGISTRY    Source registry (optional, for proxy)"
            echo "  --target-registry REGISTRY    Target private registry (required)"
            echo "  --hadoop-version VERSION      Hadoop version (default: 3.3.6)"
            echo "  --help                        Show this help message"
            echo ""
            echo "Example:"
            echo "  $0 --target-registry myregistry.example.com:5000"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

if [ -z "$TARGET_REGISTRY" ]; then
    echo "Error: --target-registry is required"
    exit 1
fi

echo "=========================================="
echo "Hadoop Operator Image Mirroring Tool"
echo "=========================================="
echo "Target Registry: $TARGET_REGISTRY"
echo "Hadoop Version: $HADOOP_VERSION"
echo ""

# Function to mirror a single image
mirror_image() {
    local source_image="$1"
    local target_image="${TARGET_REGISTRY}/${source_image}"
    
    echo "Mirroring: $source_image -> $target_image"
    
    # Pull from source (with optional registry prefix)
    if [ -n "$SOURCE_REGISTRY" ]; then
        docker pull "${SOURCE_REGISTRY}/${source_image}"
        docker tag "${SOURCE_REGISTRY}/${source_image}" "$target_image"
    else
        docker pull "$source_image"
        docker tag "$source_image" "$target_image"
    fi
    
    # Push to target
    docker push "$target_image"
    
    # Cleanup
    docker rmi "$target_image" || true
    if [ -n "$SOURCE_REGISTRY" ]; then
        docker rmi "${SOURCE_REGISTRY}/${source_image}" || true
    else
        docker rmi "$source_image" || true
    fi
    
    echo "  ✓ Mirrored successfully"
}

# Mirror all images
echo "Starting image mirroring..."
echo ""

for image in "${IMAGES[@]}"; do
    mirror_image "$image"
done

echo ""
echo "=========================================="
echo "Image mirroring completed successfully!"
echo "=========================================="
echo ""
echo "To use the mirrored images, update your HadoopCluster CR:"
echo ""
cat <<EOF
apiVersion: hadoop.apache.org/v1
kind: HadoopCluster
metadata:
  name: my-cluster
spec:
  image:
    repository: ${TARGET_REGISTRY}/apache/hadoop
    tag: "${HADOOP_VERSION}"
    pullPolicy: IfNotPresent
    pullSecrets:
      - name: regcred  # If your registry requires authentication
  # ... rest of your configuration
EOF
