#!/bin/bash
# Save Docker images to tar files for offline transport

set -e

OUTPUT_DIR="./offline-images"
HADOOP_VERSION="3.3.6"
ZOOKEEPER_VERSION="3.8"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --output-dir)
            OUTPUT_DIR="$2"
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
            echo "  --output-dir DIR         Output directory (default: ./offline-images)"
            echo "  --hadoop-version VERSION Hadoop version (default: 3.3.6)"
            echo "  --help                   Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

mkdir -p "$OUTPUT_DIR"

echo "=========================================="
echo "Saving Docker Images for Offline Use"
echo "=========================================="
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Images to save
IMAGES=(
    "apache/hadoop:${HADOOP_VERSION}"
    "zookeeper:${ZOOKEEPER_VERSION}"
)

for image in "${IMAGES[@]}"; do
    filename=$(echo "$image" | tr '/:' '_').tar
    echo "Saving: $image -> $filename"
    docker pull "$image"
    docker save -o "${OUTPUT_DIR}/${filename}" "$image"
    echo "  ✓ Saved"
done

echo ""
echo "=========================================="
echo "Images saved successfully!"
echo "=========================================="
echo ""
echo "To load these images on the target system:"
echo "  for file in ${OUTPUT_DIR}/*.tar; do docker load -i \"\$file\"; done"
