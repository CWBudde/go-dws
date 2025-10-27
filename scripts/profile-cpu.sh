#!/bin/bash
# CPU profiling script for go-dws
# Generates CPU profiles for lexer, parser, and interpreter

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
BENCHTIME="10s"
OUTPUT_DIR="profiles"
OPEN_WEB=false
PACKAGE=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--benchtime)
            BENCHTIME="$2"
            shift 2
            ;;
        -o|--output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -w|--web)
            OPEN_WEB=true
            shift
            ;;
        -p|--package)
            PACKAGE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -t, --benchtime TIME    Benchmark duration (default: 10s)"
            echo "  -o, --output-dir DIR    Output directory for profiles (default: profiles)"
            echo "  -w, --web               Open profile in web browser"
            echo "  -p, --package PKG       Profile specific package (lexer, parser, interp)"
            echo "  -h, --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                      # Profile all packages"
            echo "  $0 -p lexer -w          # Profile only lexer and open in browser"
            echo "  $0 -t 30s               # Profile for 30 seconds"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}go-dws CPU Profiling${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Settings:${NC}"
echo "  Benchmark time: $BENCHTIME"
echo "  Output directory: $OUTPUT_DIR"
echo ""

# Function to profile a package
profile_package() {
    local pkg=$1
    local name=$2
    local profile_file="$OUTPUT_DIR/${name}_cpu.prof"

    echo -e "${GREEN}Profiling $name...${NC}"

    if go test -bench=Benchmark -benchtime="$BENCHTIME" -cpuprofile="$profile_file" "./$pkg" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Profile saved to: $profile_file"

        # Show top functions
        echo -e "  ${BLUE}Top CPU consumers:${NC}"
        go tool pprof -top -nodecount=5 "$profile_file" 2>/dev/null | tail -n +4 || true

        if [ "$OPEN_WEB" = true ]; then
            echo -e "  ${YELLOW}Opening web UI...${NC}"
            go tool pprof -http=:8080 "$profile_file" &
            sleep 1
        fi
    else
        echo -e "  ${RED}✗${NC} Failed to generate profile"
    fi
    echo ""
}

# Profile packages based on selection
if [ -n "$PACKAGE" ]; then
    case $PACKAGE in
        lexer)
            profile_package "internal/lexer" "lexer"
            ;;
        parser)
            profile_package "internal/parser" "parser"
            ;;
        interp|interpreter)
            profile_package "internal/interp" "interpreter"
            ;;
        *)
            echo -e "${RED}Unknown package: $PACKAGE${NC}"
            echo "Valid packages: lexer, parser, interp"
            exit 1
            ;;
    esac
else
    # Profile all packages
    profile_package "internal/lexer" "lexer"
    profile_package "internal/parser" "parser"
    profile_package "internal/interp" "interpreter"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}CPU profiling complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "To analyze profiles:"
echo "  - View in terminal: go tool pprof -text $OUTPUT_DIR/<name>_cpu.prof"
echo "  - View top functions: go tool pprof -top $OUTPUT_DIR/<name>_cpu.prof"
echo "  - Interactive mode: go tool pprof $OUTPUT_DIR/<name>_cpu.prof"
echo "  - Web UI: go tool pprof -http=:8080 $OUTPUT_DIR/<name>_cpu.prof"
echo ""
echo "Common pprof commands:"
echo "  (pprof) top        - Show top CPU consumers"
echo "  (pprof) list func  - Show source code for function"
echo "  (pprof) web        - Generate call graph (requires graphviz)"
