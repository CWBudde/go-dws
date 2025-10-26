#!/usr/bin/env bash
# Optimize DWScript WASM binary using wasm-opt (Binaryen)
#
# Usage:
#   ./optimize.sh <wasm-file>
#
# Requires:
#   wasm-opt (install with: npm install -g binaryen)

set -e

if [ $# -eq 0 ]; then
    echo "Usage: $0 <wasm-file>"
    exit 1
fi

WASM_FILE="$1"

if [ ! -f "$WASM_FILE" ]; then
    echo "Error: File not found: $WASM_FILE"
    exit 1
fi

if ! command -v wasm-opt &> /dev/null; then
    echo "Error: wasm-opt not found"
    echo "Install with: npm install -g binaryen"
    exit 1
fi

echo "=== WASM Optimization ==="
echo "Input: $WASM_FILE"

# Get original size
ORIGINAL_SIZE=$(stat -f%z "$WASM_FILE" 2>/dev/null || stat -c%s "$WASM_FILE")
ORIGINAL_MB=$(echo "scale=2; $ORIGINAL_SIZE / 1048576" | bc)
echo "Original size: $ORIGINAL_MB MB"

# Create optimized file
OPTIMIZED_FILE="${WASM_FILE%.wasm}.opt.wasm"

echo ""
echo "Running wasm-opt with -O3..."
wasm-opt -O3 -o "$OPTIMIZED_FILE" "$WASM_FILE"

# Get optimized size
OPTIMIZED_SIZE=$(stat -f%z "$OPTIMIZED_FILE" 2>/dev/null || stat -c%s "$OPTIMIZED_FILE")
OPTIMIZED_MB=$(echo "scale=2; $OPTIMIZED_SIZE / 1048576" | bc)

# Calculate savings
SAVINGS=$((ORIGINAL_SIZE - OPTIMIZED_SIZE))
SAVINGS_PERCENT=$(echo "scale=1; ($SAVINGS * 100) / $ORIGINAL_SIZE" | bc)

echo ""
echo "✓ Optimization complete!"
echo "  Optimized size: $OPTIMIZED_MB MB"
echo "  Savings: $SAVINGS_PERCENT% ($SAVINGS bytes)"
echo ""
echo "Output: $OPTIMIZED_FILE"

# Replace original with optimized if it's smaller
if [ $OPTIMIZED_SIZE -lt $ORIGINAL_SIZE ]; then
    echo ""
    echo "Replacing original with optimized version..."
    mv "$OPTIMIZED_FILE" "$WASM_FILE"
    echo "✓ Done!"
else
    echo ""
    echo "Warning: Optimized file is not smaller, keeping original"
    rm "$OPTIMIZED_FILE"
fi
