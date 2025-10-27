#!/bin/bash
# Benchmark script for go-dws
# Runs all benchmarks in the project with standard settings

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
BENCHTIME="3s"
COUNT=5
OUTPUT_FILE=""
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--benchtime)
            BENCHTIME="$2"
            shift 2
            ;;
        -n|--count)
            COUNT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -t, --benchtime TIME    Benchmark duration (default: 3s)"
            echo "  -n, --count N           Run each benchmark N times (default: 5)"
            echo "  -o, --output FILE       Save results to FILE"
            echo "  -v, --verbose           Verbose output"
            echo "  -h, --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                      # Run with defaults"
            echo "  $0 -t 5s -n 10          # Run for 5s, 10 iterations"
            echo "  $0 -o results.txt       # Save results to file"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}go-dws Performance Benchmarks${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Settings:${NC}"
echo "  Benchmark time: $BENCHTIME"
echo "  Iterations: $COUNT"
if [ -n "$OUTPUT_FILE" ]; then
    echo "  Output file: $OUTPUT_FILE"
fi
echo ""

# Build the benchmark command
# -run=^$ ensures no regular tests are run, only benchmarks
BENCH_CMD="go test -bench=. -benchmem -benchtime=$BENCHTIME -count=$COUNT -run=^$"

if [ "$VERBOSE" = true ]; then
    BENCH_CMD="$BENCH_CMD -v"
fi

BENCH_CMD="$BENCH_CMD ./..."

# Run benchmarks
echo -e "${GREEN}Running benchmarks...${NC}"
echo ""

if [ -n "$OUTPUT_FILE" ]; then
    # Run and save to file
    $BENCH_CMD | tee "$OUTPUT_FILE"
    echo ""
    echo -e "${GREEN}Results saved to: $OUTPUT_FILE${NC}"
else
    # Just run
    $BENCH_CMD
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Benchmarks complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "To analyze results:"
echo "  - Compare with previous runs using benchstat"
echo "  - Profile with: ./scripts/profile-cpu.sh"
echo "  - Memory profile: ./scripts/profile-mem.sh"
