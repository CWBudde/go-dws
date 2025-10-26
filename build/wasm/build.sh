#!/usr/bin/env bash
# Build script for DWScript WebAssembly binary
#
# Usage:
#   ./build.sh [mode]
#
# Modes:
#   monolithic (default) - Single WASM binary with full interpreter
#   modular              - Core WASM + optional modules
#   hybrid               - WASM core + JavaScript glue
#
# Environment variables:
#   OUTPUT_DIR - Directory for build output (default: build/wasm/dist)
#   OPTIMIZE   - Run wasm-opt after build (default: false)

set -e

# Configuration
MODE="${1:-monolithic}"
OUTPUT_DIR="${OUTPUT_DIR:-build/wasm/dist}"
WASM_FILE="dwscript.wasm"
GO_VERSION=$(go version | awk '{print $3}')

echo "=== DWScript WASM Build ==="
echo "Mode: $MODE"
echo "Go version: $GO_VERSION"
echo "Output directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build flags based on mode
BUILD_FLAGS="-ldflags=-s"
case "$MODE" in
    monolithic)
        echo "Building monolithic WASM (single file, all features)"
        BUILD_FLAGS="$BUILD_FLAGS -tags=wasm_monolithic"
        ;;
    modular)
        echo "Building modular WASM (core + loadable modules)"
        BUILD_FLAGS="$BUILD_FLAGS -tags=wasm_modular"
        ;;
    hybrid)
        echo "Building hybrid WASM (core + JS glue)"
        BUILD_FLAGS="$BUILD_FLAGS -tags=wasm_hybrid"
        ;;
    *)
        echo "Error: Unknown mode '$MODE'"
        echo "Valid modes: monolithic, modular, hybrid"
        exit 1
        ;;
esac

# Build WASM binary
echo "Building WASM binary..."
GOOS=js GOARCH=wasm go build $BUILD_FLAGS -o "$OUTPUT_DIR/$WASM_FILE" ./cmd/dwscript-wasm

# Check build result
if [ ! -f "$OUTPUT_DIR/$WASM_FILE" ]; then
    echo "Error: Build failed, WASM file not created"
    exit 1
fi

# Get file size
SIZE=$(stat -f%z "$OUTPUT_DIR/$WASM_FILE" 2>/dev/null || stat -c%s "$OUTPUT_DIR/$WASM_FILE")
SIZE_MB=$(echo "scale=2; $SIZE / 1048576" | bc)

echo "✓ Build successful!"
echo "  Size: $SIZE_MB MB ($SIZE bytes)"

# Copy wasm_exec.js
echo ""
echo "Copying wasm_exec.js..."
GOROOT=$(go env GOROOT)

# Try multiple possible locations (Go versions have different paths)
WASM_EXEC_PATHS=(
    "$GOROOT/go/lib/wasm/wasm_exec.js"
    "$GOROOT/lib/wasm/wasm_exec.js"
    "$GOROOT/misc/wasm/wasm_exec.js"
)

WASM_EXEC_JS=""
for path in "${WASM_EXEC_PATHS[@]}"; do
    if [ -f "$path" ]; then
        WASM_EXEC_JS="$path"
        break
    fi
done

if [ -n "$WASM_EXEC_JS" ]; then
    cp "$WASM_EXEC_JS" "$OUTPUT_DIR/"
    echo "✓ Copied wasm_exec.js from $WASM_EXEC_JS"
else
    echo "⚠️  Warning: wasm_exec.js not found in GOROOT"
    echo "  Searched: ${WASM_EXEC_PATHS[*]}"
fi

# Run optimization if requested
if [ "$OPTIMIZE" = "true" ]; then
    echo ""
    echo "Running optimization..."
    if command -v wasm-opt &> /dev/null; then
        ./build/wasm/optimize.sh "$OUTPUT_DIR/$WASM_FILE"
    else
        echo "Warning: wasm-opt not found, skipping optimization"
        echo "Install with: npm install -g binaryen"
    fi
fi

# Size check
echo ""
MAX_SIZE=$((3 * 1024 * 1024))  # 3MB
if [ $SIZE -gt $MAX_SIZE ]; then
    SIZE_LIMIT_MB=$(echo "scale=2; $MAX_SIZE / 1048576" | bc)
    echo "⚠️  Warning: WASM binary is larger than $SIZE_LIMIT_MB MB"
    echo "  Consider running with OPTIMIZE=true or using modular mode"
else
    echo "✓ Size check passed (< 3MB)"
fi

echo ""
echo "Build output: $OUTPUT_DIR"
echo "  $WASM_FILE ($SIZE_MB MB)"
echo "  wasm_exec.js"
echo ""
echo "Done!"
